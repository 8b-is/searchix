package importer

import (
	"context"
	"io"
	"log/slog"
	"reflect"

	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/nix"

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

func convertValue(nj *nixValueJSON) *nix.Value {
	if nj == nil {
		return nil
	}
	switch nj.Type {
	case "", "literalExpression":
		return &nix.Value{
			Text: nj.Text,
		}
	case "literalMD":
		return &nix.Value{
			Markdown: nix.Markdown(nj.Text),
		}
	default:
		slog.Warn("got unexpected Value type", "type", nj.Type, "text", nj.Text)

		return nil
	}
}

type OptionIngester struct {
	dec     *jstream.Decoder
	ms      *mapstructure.Decoder
	optJSON nixOptionJSON
	infile  io.ReadCloser
	source  *config.Source
}

func NewOptionProcessor(infile io.ReadCloser, source *config.Source) (*OptionIngester, error) {
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

func (i *OptionIngester) Process(ctx context.Context) (<-chan nix.Importable, <-chan error) {
	results := make(chan nix.Importable)
	errs := make(chan error)

	go func() {
		defer i.infile.Close()
		defer close(results)
		defer close(errs)

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

			var decls []*nix.Link
			for _, decl := range x["declarations"].([]interface{}) {

				switch decl := reflect.ValueOf(decl); decl.Kind() {
				case reflect.String:
					s := decl.String()
					link, err := MakeChannelLink(i.source.Repo, s)
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
					link := nix.Link{
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

			i.optJSON = nixOptionJSON{}
			err := i.ms.Decode(x) // stores in optJSON
			if err != nil {
				errs <- errors.WithMessagef(err, "failed to decode option %#v", x)

				continue
			}

			var decs = make([]nix.Link, len(i.optJSON.Declarations))
			for i, d := range i.optJSON.Declarations {
				decs[i] = nix.Link(d)
			}

			// slog.Debug("sending option", "name", kv.Key)
			results <- nix.Option{
				Name:            kv.Key,
				Source:          i.source.Key,
				Declarations:    decs,
				Default:         convertValue(i.optJSON.Default),
				Description:     nix.Markdown(i.optJSON.Description),
				Example:         convertValue(i.optJSON.Example),
				RelatedPackages: nix.Markdown(i.optJSON.RelatedPackages),
				Loc:             i.optJSON.Loc,
				Type:            i.optJSON.Type,
			}
		}
	}()

	return results, errs
}
