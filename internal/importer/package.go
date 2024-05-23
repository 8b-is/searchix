package importer

import (
	"context"
	"encoding/json"
	"io"
	"reflect"
	"searchix/internal/config"
	"searchix/internal/nix"
	"strings"

	"github.com/bcicen/jstream"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
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
	dec    *jstream.Decoder
	ms     *mapstructure.Decoder
	pkg    packageJSON
	infile io.ReadCloser
	source *config.Source
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

func NewPackageProcessor(infile io.ReadCloser, source *config.Source) (*PackageIngester, error) {
	i := &PackageIngester{
		dec:    jstream.NewDecoder(infile, 2).EmitKV(),
		pkg:    packageJSON{},
		infile: infile,
		source: source,
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

func (i *PackageIngester) Process(ctx context.Context) (<-chan nix.Importable, <-chan error) {
	results := make(chan nix.Importable)
	errs := make(chan error)

	go func() {
		defer i.infile.Close()
		defer close(results)
		defer close(errs)

	outer:
		for mv := range i.dec.Stream() {
			var err error
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
				for i, plat := range meta["platforms"].([]interface{}) {
					switch v := reflect.ValueOf(plat); v.Kind() {
					case reflect.String:
						plats[i] = v.String()
					case reflect.Map:
						plats[i] = makeAdhocPlatform(v.Interface())
					default:
						errs <- errors.Errorf(
							"don't know how to convert platform type %s",
							v.Kind().String(),
						)
					}
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

			i.pkg = packageJSON{}
			err = i.ms.Decode(x) // stores in i.pkg
			if err != nil {
				errs <- errors.WithMessagef(err, "failed to decode package %#v", x)

				continue
			}

			maintainers := make([]nix.Maintainer, len(i.pkg.Meta.Maintainers))
			for i, m := range i.pkg.Meta.Maintainers {
				maintainers[i] = nix.Maintainer{
					Name:   m.Name,
					Github: m.Github,
				}
			}

			subpath, line, _ := strings.Cut(i.pkg.Meta.Position, ":")

			pkgSet, _, found := strings.Cut(kv.Key, ".")
			if !found {
				pkgSet = ""
			}

			url, err := makeRepoURL(i.source.Repo, subpath, line)
			if err != nil {
				errs <- err
			}
			results <- &nix.Package{
				Name:            i.pkg.Name,
				Attribute:       kv.Key,
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
				Definition:      url,
			}
		}
	}()

	return results, errs
}
