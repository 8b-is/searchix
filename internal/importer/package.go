package importer

import (
	"context"
	"encoding/json"
	"io"
	"net/url"
	"reflect"
	"strings"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/nix"
	"alin.ovh/searchix/internal/programs"
	"alin.ovh/x/log"

	"github.com/bcicen/jstream"
	"github.com/mitchellh/mapstructure"
	"gitlab.com/tozd/go/errors"
)

type packageJSON struct {
	Name    string `mapstructure:"pname"`
	Meta    metaJSON
	Version string
}

type metaJSON struct {
	Broken          bool
	Description     string
	LongDescription string
	Homepages       []string `mapstructure:"homepage"`
	MainProgram     string
	Maintainers     []maintainerJSON
	Platforms       []string
	Position        string
}

type maintainerJSON struct {
	Github string
	Name   string
}

type PackageIngester struct {
	dec      *jstream.Decoder
	ms       *mapstructure.Decoder
	log      *log.Logger
	pkg      packageJSON
	infile   io.ReadCloser
	source   *config.Source
	programs *programs.DB
}

func makeAdhocLicense(name string) nix.License {
	return nix.License{
		FullName: name,
	}
}

func makeAdhocPlatform(v any) string {
	s, err := json.Marshal(v)
	if err != nil {
		panic("can't convert json back to json?")
	}

	return string(s)
}

func NewPackageProcessor(
	infile io.ReadCloser,
	source *config.Source,
	log *log.Logger,
	programsDB *programs.DB,
) (*PackageIngester, errors.E) {
	i := &PackageIngester{
		dec:      jstream.NewDecoder(infile, source.JSONDepth).EmitKV(),
		log:      log,
		pkg:      packageJSON{},
		infile:   infile,
		source:   source,
		programs: programsDB,
	}

	ms, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ZeroFields: true,
		Result:     &i.pkg,
		Squash:     true,
		DecodeHook: mapstructure.TextUnmarshallerHookFunc(),
	})
	if err != nil {
		defer infile.Close()

		return nil, errors.WithMessage(err, "could not create mapstructure decoder")
	}
	i.ms = ms

	return i, nil
}

func convertToLicense(in map[string]any) *nix.License {
	l := &nix.License{}
	if v, found := in["shortName"]; found {
		l.Name = v.(string)
	}
	if v, found := in["fullName"]; found {
		l.FullName = v.(string)
	}
	if v, found := in["appendixUrl"]; found {
		l.AppendixURL = v.(string)
	}
	if v, found := in["spdxId"]; found {
		l.SPDXId = v.(string)
	}
	if v, found := in["url"]; found {
		l.URL = v.(string)
	}

	return l
}

func (i *PackageIngester) Process(ctx context.Context) (<-chan nix.Importable, <-chan errors.E) {
	results := make(chan nix.Importable)
	errs := make(chan errors.E)

	if i.programs != nil {
		err := i.programs.Open()
		if err != nil {
			errs <- errors.WithMessage(err, "could not open programs database")
			i.programs = nil
		}
	}

	go func() {
		defer i.infile.Close()
		defer close(results)
		defer close(errs)

		if i.programs != nil {
			defer i.programs.Close()
		}

	outer:
		for mv := range i.dec.Stream() {
			var err errors.E
			var programs []string
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

			meta := x["meta"].(map[string]interface{})

			var licenses []nix.License
			if meta["license"] != nil {
				switch v := reflect.ValueOf(meta["license"]); v.Kind() {
				case reflect.Map:
					licenses = append(licenses, *convertToLicense(v.Interface().(map[string]interface{})))
				case reflect.Array, reflect.Slice:
					licenses = make([]nix.License, v.Len())
					for i, v := range v.Interface().([]interface{}) {
						switch v := reflect.ValueOf(v); v.Kind() {
						case reflect.String:
							licenses[i] = makeAdhocLicense(v.String())
						case reflect.Map:
							licenses[i] = *convertToLicense(v.Interface().(map[string]interface{}))
						default:
							errs <- errors.Errorf(
								"don't know how to handle sublicense of type %s: %v",
								v.Kind().String(),
								v,
							)
						}
					}
				case reflect.String:
					licenses = append(licenses, makeAdhocLicense(v.String()))
				default:
					errs <- errors.Errorf(
						"don't know how to handle license of type %s: %v",
						v.Kind().String(),
						meta["license"],
					)
				}
				delete(meta, "license")
			}

			if meta["platforms"] != nil {
				var plats = make([]any, len(meta["platforms"].([]any)))
				i := 0
				for _, plat := range meta["platforms"].([]interface{}) {
					switch v := reflect.ValueOf(plat); v.Kind() {
					case reflect.String:
						plats[i] = v.String()
					case reflect.Map:
						plats[i] = makeAdhocPlatform(v.Interface())
					case reflect.Slice:
						ps := make([]any, v.Len())
						for j, item := range v.Slice(0, v.Len()).Interface().([]any) {
							ps[j] = item.(string)
						}
						plats = append(plats, ps...)
					default:
						errs <- errors.Errorf(
							"don't know how to convert platform type %s: %v",
							v.Kind().String(),
							v.Interface(),
						)
					}
					i++
				}
				meta["platforms"] = plats
			}
			if meta["homepage"] != nil {
				switch v := reflect.ValueOf(meta["homepage"]); v.Kind() {
				case reflect.String:
					meta["homepage"] = []string{v.String()}
				case reflect.Slice:
				// already fine
				default:
					errs <- errors.Errorf(
						"don't know how to interpret homepage type %s'",
						v.Kind().String(),
					)
				}
			}

			var maints []nix.Maintainer
			if meta["maintainers"] != nil {
				switch maint := reflect.ValueOf(meta["maintainers"]); maint.Kind() {
				case reflect.String:
					maints = []nix.Maintainer{nix.Maintainer{Name: maint.String(), Github: maint.String()}}
				case reflect.Slice, reflect.Array:
					maints = make([]nix.Maintainer, maint.Len())
					for i, val := range maint.Slice(0, maint.Len()).Interface().([]any) {
						switch v := reflect.ValueOf(val); v.Kind() {
						case reflect.String:
							maints[i] = nix.Maintainer{Name: v.String(), Github: v.String()}
						case reflect.Map:
							m := v.Interface().(map[string]any)
							maints[i] = nix.Maintainer{}
							if m["name"] != nil && m["name"].(string) != "" {
								maints[i].Name = m["name"].(string)
							}
							if m["github"] != nil && m["github"].(string) != "" {
								maints[i].Github = m["github"].(string)
							}
						default:
							errs <- errors.Errorf(
								"don't know how to handle maintainer entry of type %s: %v",
								v.Kind().String(),
								v,
							)
						}
					}
				default:
					errs <- errors.Errorf(
						"don't know how to interpret maintainers type %s'",
						maint.Kind().String(),
					)
				}
				meta["maintainers"] = maints
			}

			i.pkg = packageJSON{}
			if err := i.ms.Decode(x); err != nil { // stores in i.pkg
				errs <- errors.WithMessagef(err, "failed to decode package %#v", x)

				continue
			}

			if i.programs != nil {
				programs, err = i.programs.GetPackagePrograms(ctx, kv.Key)
				if err != nil {
					errs <- errors.WithMessagef(err, "failed to get programs for package %s", i.pkg.Name)
				}
			}

			maintainers := make([]nix.Maintainer, len(i.pkg.Meta.Maintainers))
			for i, m := range i.pkg.Meta.Maintainers {
				maintainers[i] = nix.Maintainer{
					Name:   m.Name,
					Github: m.Github,
				}
			}

			pkgSet, _, found := strings.Cut(kv.Key, ".")
			if !found {
				pkgSet = ""
			}

			var definition string
			if i.pkg.Meta.Position != "" {
				defURL, err := url.Parse(i.pkg.Meta.Position)
				if err != nil {
					errs <- errors.WithMessagef(err, "failed to parse source URL %s", definition)
				}
				if defURL.IsAbs() {
					definition = i.pkg.Meta.Position
				} else {
					subpath, line, _ := strings.Cut(i.pkg.Meta.Position, ":")
					definition, err = i.source.Repo.GetFileURL(subpath, line)
					if err != nil {
						errs <- errors.WithMessagef(err, "failed to make repo URL for package %s", i.pkg.Name)
					}
				}
			}

			results <- &nix.Package{
				Name:            i.pkg.Name,
				Attribute:       strings.TrimPrefix(kv.Key, "nur.repos."),
				Source:          i.source.Key,
				PackageSet:      pkgSet,
				Version:         i.pkg.Version,
				Broken:          i.pkg.Meta.Broken,
				Description:     i.pkg.Meta.Description,
				LongDescription: nix.Markdown(i.pkg.Meta.LongDescription),
				Homepages:       i.pkg.Meta.Homepages,
				MainProgram:     i.pkg.Meta.MainProgram,
				Platforms:       i.pkg.Meta.Platforms,
				Licenses:        licenses,
				Maintainers:     maintainers,
				Definition:      definition,
				Programs:        programs,
			}
		}
	}()

	return results, errs
}
