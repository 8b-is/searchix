package importer

import (
	"context"
	"io"
	"reflect"
	"time"

	"alin.ovh/x/log"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/nix"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/bcicen/jstream"
	"github.com/mitchellh/mapstructure"
)

type nixDocJSON struct {
	Type string `mapstructure:"_type"`
	Text string
}

type linkJSON struct {
	Name string
	URL  string `json:"url"`
}

type nixOptionJSON struct {
	Declarations    []linkJSON
	Default         *nixDocJSON
	Description     string
	Example         *nixDocJSON
	Loc             []string
	ReadOnly        bool
	RelatedPackages string
	Type            string
}

func (i *OptionIngester) convertDocsValue(nj *nixDocJSON) *nix.Docs {
	if nj == nil {
		return nil
	}
	switch nj.Type {
	case "", "literalExpression":
		return &nix.Docs{
			Plain: nj.Text,
		}
	case "literalMD":
		return &nix.Docs{
			Markdown: nix.Markdown(nj.Text),
		}
	default:
		i.log.Warn("got unexpected docs type", "type", nj.Type, "text", nj.Text)

		return nil
	}
}

type OptionIngester struct {
	dec     *jstream.Decoder
	ms      *mapstructure.Decoder
	log     *log.Logger
	optJSON nixOptionJSON
	infile  io.ReadCloser
	source  *config.Source
}

func NewOptionProcessor(
	infile io.ReadCloser,
	source *config.Source,
	log *log.Logger,
) (*OptionIngester, error) {
	i := OptionIngester{
		dec:     jstream.NewDecoder(infile, source.JSONDepth).EmitKV(),
		log:     log,
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

		return nil, fault.Wrap(err, fmsg.With("could not create mapstructure decoder"))
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
				errs <- fault.Wrap(err, fmsg.With("could not decode JSON"))

				continue
			}
			if mv.ValueType != jstream.Object {
				errs <- fault.Newf("unexpected object type %s", ValueTypeToString(mv.ValueType))

				continue
			}
			kv := mv.Value.(jstream.KV)
			x := kv.Value.(map[string]any)

			var decls []*nix.Link
			for _, decl := range x["declarations"].([]any) {
				switch decl := reflect.ValueOf(decl); decl.Kind() {
				case reflect.String:
					s := decl.String()
					link, err := MakeChannelLink(i.source.Repo, s)
					if err != nil {
						errs <- fault.Wrap(err, fmsg.Withf("could not make a channel link for channel %s, revision %s and subpath %s",
							i.source.Channel, i.source.Repo.Revision, s,
						))

						continue
					}
					decls = append(decls, link)
				case reflect.Map:
					v := decl.Interface().(map[string]any)
					link := nix.Link{
						Name: v["name"].(string),
						URL:  v["url"].(string),
					}
					decls = append(decls, &link)
				default:
					errs <- fault.Newf("unexpected declaration type %s", decl.Kind().String())

					continue
				}
			}
			if len(decls) > 0 {
				x["declarations"] = decls
			}

			i.optJSON = nixOptionJSON{}
			err := i.ms.Decode(x) // stores in optJSON
			if err != nil {
				errs <- fault.Wrap(err, fmsg.Withf("failed to decode option %#v", x))

				continue
			}

			decs := make([]nix.Link, len(i.optJSON.Declarations))
			for i, d := range i.optJSON.Declarations {
				decs[i] = nix.Link(d)
			}

			// log.Debug("sending option", "name", kv.Key)
			results <- nix.Option{
				Name:            kv.Key,
				Source:          i.source.Key,
				Declarations:    decs,
				Default:         i.convertDocsValue(i.optJSON.Default),
				Description:     nix.Markdown(i.optJSON.Description),
				Example:         i.convertDocsValue(i.optJSON.Example),
				RelatedPackages: nix.Markdown(i.optJSON.RelatedPackages),
				Loc:             i.optJSON.Loc,
				Type:            i.optJSON.Type,
				ImportedAt:      time.Now(),
			}
		}
	}()

	return results, errs
}
