package importer

import (
	"context"
	"io"
	"iter"
	"reflect"
	"strings"
	"time"

	"alin.ovh/x/log"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/nix"
	"alin.ovh/searchix/internal/nixdocs"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/bcicen/jstream"
)

func (i *OptionIngester) convertDocsValue(docMap map[string]any) *nix.Docs {
	if docMap == nil {
		return nil
	}

	var docType string
	if t, ok := docMap["_type"].(string); ok {
		docType = t
	}

	var text string
	if t, ok := docMap["text"].(string); ok {
		text = t
	}

	switch docType {
	case "", "literalExpression":
		return &nix.Docs{
			Plain: text,
		}
	case "literalMD":
		return &nix.Docs{
			Markdown: nixdocs.Markdown(text),
		}
	default:
		i.log.Warn("got unexpected docs type", "type", docType, "text", text)

		return nil
	}
}

type OptionIngester struct {
	dec    *jstream.Decoder
	log    *log.Logger
	infile io.ReadCloser
	source config.Source
	errs   []error
}

func NewOptionProcessor(
	infile io.ReadCloser,
	source config.Source,
	log *log.Logger,
) *OptionIngester {
	return &OptionIngester{
		dec:    jstream.NewDecoder(infile, source.JSONDepth).EmitKV(),
		log:    log,
		infile: infile,
		source: source,
	}
}

func (i *OptionIngester) Process(
	ctx context.Context,
) iter.Seq[index.Indexable] {
	return func(yield func(index.Indexable) bool) {
		defer i.infile.Close()

		for mv := range i.dec.Stream() {
			if ctx.Err() != nil {
				break
			}
			if err := i.dec.Err(); err != nil {
				i.errs = append(i.errs, fault.Wrap(err, fmsg.With("could not decode JSON")))

				continue
			}
			if mv.ValueType != jstream.Object {
				i.errs = append(
					i.errs,
					fault.Newf("unexpected object type %s", ValueTypeToString(mv.ValueType)),
				)

				continue
			}
			kv := mv.Value.(jstream.KV)
			x := kv.Value.(map[string]any)

			var decls []nix.Link
			for _, decl := range x["declarations"].([]any) {
				switch decl := decl.(type) {
				case string:
					s := decl
					link, err := MakeChannelLink(i.source.Repo, s)
					if err != nil {
						i.errs = append(
							i.errs,
							fault.Wrap(
								err,
								fmsg.Withf(
									"could not make a channel link for channel %s, revision %s and subpath %s",
									i.source.Channel,
									i.source.Repo.Revision,
									s,
								),
							),
						)

						continue
					}
					decls = append(decls, *link)
				case map[string]any:
					v := decl
					link := nix.Link{
						Name: v["name"].(string),
						URL:  v["url"].(string),
					}
					decls = append(decls, link)
				default:
					i.errs = append(
						i.errs,
						fault.Newf("unexpected declaration type %s", reflect.TypeOf(decl).String()),
					)

					continue
				}
			}
			var description string
			if v, ok := x["description"].(string); ok {
				description = v
			}

			var optType string
			if v, ok := x["type"].(string); ok {
				optType = v
			}

			var relatedPackages string
			if v, ok := x["relatedPackages"].(string); ok {
				relatedPackages = v
			}

			var loc []string
			if v, ok := x["loc"].([]any); ok {
				loc = make([]string, len(v))
				for i, item := range v {
					if s, ok := item.(string); ok {
						loc[i] = s
					}
				}
			}

			var defaultDocs *nix.Docs
			if v, ok := x["default"].(map[string]any); ok {
				defaultDocs = i.convertDocsValue(v)
			}

			var exampleDocs *nix.Docs
			if v, ok := x["example"].(map[string]any); ok {
				exampleDocs = i.convertDocsValue(v)
			}

			var parents string
			if len(loc) > 1 {
				parents = strings.Join(loc[:len(loc)-1], ".")
			}

			// log.Debug("sending option", "name", kv.Key)
			if !yield(nix.Option{
				Name:            kv.Key,
				Source:          i.source.Key,
				Declarations:    decls,
				Default:         defaultDocs,
				Description:     nixdocs.Markdown(description),
				Example:         exampleDocs,
				RelatedPackages: nixdocs.Markdown(relatedPackages),
				Loc:             loc,
				Parents:         parents,
				Type:            optType,
				ImportedAt:      time.Now(),
			}) {
				return
			}
		}
	}
}

func (i *OptionIngester) Err() error {
	errs := i.errs
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return i.errs[0]
	}

	i.errs = []error{}

	return fault.Newf("encountered %d errors during processing", len(errs))
}
