package importer

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"reflect"
	"searchix/internal/options"
	"strings"

	"github.com/bcicen/jstream"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type nixValueJSON struct {
	Type string `mapstructure:"_type"`
	Text string
}

type linkJSON struct {
	Name string
	URL  string `json:"url"`
}

type nixOptionJSON struct {
	Declarations    []linkJSON
	Default         *nixValueJSON
	Description     string
	Example         *nixValueJSON
	Loc             []string
	ReadOnly        bool
	RelatedPackages string
	Type            string
}

func ValueTypeToString(valueType jstream.ValueType) string {
	switch valueType {
	case jstream.Unknown:
		return "unknown"
	case jstream.Null:
		return "null"
	case jstream.String:
		return "string"
	case jstream.Number:
		return "number"
	case jstream.Boolean:
		return "boolean"
	case jstream.Array:
		return "array"
	case jstream.Object:
		return "object"
	}

	return "very strange"
}

func makeGitHubFileURL(userRepo string, ref string, subPath string) string {
	url, _ := url.JoinPath("https://github.com/", userRepo, "blob", ref, subPath)

	return url
}

// make configurable?
var channelRepoMap = map[string]string{
	"nixpkgs":      "NixOS/nixpkgs",
	"nix-darwin":   "LnL7/nix-darwin",
	"home-manager": "nix-community/home-manager",
}

func MakeChannelLink(channel string, ref string, subPath string) (*options.Link, error) {
	if channelRepoMap[channel] == "" {
		return nil, fmt.Errorf("don't know what repository relates to channel <%s>", channel)
	}

	return &options.Link{
		Name: fmt.Sprintf("<%s/%s>", channel, subPath),
		URL:  makeGitHubFileURL(channelRepoMap[channel], ref, subPath),
	}, nil
}

func convertNixValue(nj *nixValueJSON) *options.NixValue {
	if nj == nil {
		return nil
	}
	switch nj.Type {
	case "", "literalExpression":
		return &options.NixValue{
			Text: nj.Text,
		}
	case "literalMD":
		return &options.NixValue{
			Markdown: options.Markdown(nj.Text),
		}
	default:
		slog.Warn("got unexpected NixValue type", "type", nj.Type, "text", nj.Text)

		return nil
	}
}

type OptionIngester struct {
	dec     *jstream.Decoder
	ms      *mapstructure.Decoder
	optJSON nixOptionJSON
	infile  *os.File
	source  *Source
}

type Ingester[T options.NixOption] interface {
	Process() (<-chan *T, <-chan error)
}

func NewOptionProcessor(inpath string, source *Source) (*OptionIngester, error) {
	infile, err := os.Open(inpath)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to open input file %s", inpath)
	}
	i := OptionIngester{
		dec:     jstream.NewDecoder(infile, 1).EmitKV(),
		optJSON: nixOptionJSON{},
		infile:  infile,
		source:  source,
	}

	ms, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		ZeroFields:  true,
		Result:      &i.optJSON,
		Squash:      true,
		DecodeHook:  mapstructure.TextUnmarshallerHookFunc(),
	})
	if err != nil {
		defer infile.Close()

		return nil, errors.WithMessage(err, "could not create mapstructure decoder")
	}
	i.ms = ms

	return &i, nil
}

func (i *OptionIngester) Process(ctx context.Context) (<-chan *options.NixOption, <-chan error) {
	results := make(chan *options.NixOption)
	errs := make(chan error)

	go func() {
		defer i.infile.Close()
		defer close(results)
		defer close(errs)

		slog.Debug("starting decoder stream")
	outer:
		for mv := range i.dec.Stream() {
			select {
			case <-ctx.Done():
				break outer
			default:
			}
			if err := i.dec.Err(); err != nil {
				errs <- errors.WithMessage(err, "could not decode JSON")

				continue
			}
			if mv.ValueType != jstream.Object {
				errs <- errors.Errorf("unexpected object type %s", ValueTypeToString(mv.ValueType))

				continue
			}
			kv := mv.Value.(jstream.KV)
			x := kv.Value.(map[string]interface{})

			var decls []*options.Link
			for _, decl := range x["declarations"].([]interface{}) {
				i.optJSON = nixOptionJSON{}

				switch decl := reflect.ValueOf(decl); decl.Kind() {
				case reflect.String:
					s := decl.String()
					link, err := MakeChannelLink(i.source.Channel, i.source.Repo.Revision, s)
					if err != nil {
						errs <- errors.WithMessagef(err,
							"could not make a channel link for channel %s, revision %s and subpath %s",
							i.source.Channel, i.source.Repo.Revision, s,
						)

						continue
					}
					decls = append(decls, link)
				case reflect.Map:
					v := decl.Interface().(map[string]interface{})
					link := options.Link{
						Name: v["name"].(string),
						URL:  v["url"].(string),
					}
					decls = append(decls, &link)
				default:
					errs <- errors.Errorf("unexpected declaration type %s", decl.Kind().String())

					continue
				}
			}
			if len(decls) > 0 {
				x["declarations"] = decls
			}

			err := i.ms.Decode(x) // stores in optJSON
			if err != nil {
				errs <- errors.WithMessagef(err, "failed to decode option %#v", x)

				continue
			}

			var decs = make([]options.Link, len(i.optJSON.Declarations))
			for i, d := range i.optJSON.Declarations {
				decs[i] = options.Link(d)
			}

			// slog.Debug("sending option", "name", kv.Key)
			results <- &options.NixOption{
				Name:            kv.Key,
				Source:          strings.ToLower(i.source.Name),
				Declarations:    decs,
				Default:         convertNixValue(i.optJSON.Default),
				Description:     options.Markdown(i.optJSON.Description),
				Example:         convertNixValue(i.optJSON.Example),
				RelatedPackages: options.Markdown(i.optJSON.RelatedPackages),
				Loc:             i.optJSON.Loc,
				Type:            i.optJSON.Type,
			}
		}
	}()

	return results, errs
}
