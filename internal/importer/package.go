package importer

import (
	"context"
	"encoding/json"
	"io"
	"net/url"
	"reflect"
	"strings"
	"time"

	"alin.ovh/x/log"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/nix"
	"alin.ovh/searchix/internal/nixdocs"
	"alin.ovh/searchix/internal/programs"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/bcicen/jstream"
)

type PackageIngester struct {
	dec      *jstream.Decoder
	log      *log.Logger
	infile   io.ReadCloser
	source   config.Source
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
	source config.Source,
	log *log.Logger,
	programsDB *programs.DB,
) *PackageIngester {
	return &PackageIngester{
		dec:      jstream.NewDecoder(infile, source.JSONDepth).EmitKV(),
		log:      log,
		infile:   infile,
		source:   source,
		programs: programsDB,
	}
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

func (i *PackageIngester) Process(
	ctx context.Context,
	results chan<- index.Indexable,
	errs chan<- error,
) {
	if i.programs != nil {
		err := i.programs.Open(ctx)
		if err != nil {
			errs <- fault.Wrap(err, fmsg.With("could not open programs database"))
			i.programs = nil
		}
	}

	defer i.infile.Close()

	if i.programs != nil {
		defer i.programs.Close()
	}

outer:
	for mv := range i.dec.Stream() {
		var err error
		var programs []string
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

		meta := x["meta"].(map[string]any)

		var licenses []nix.License
		if meta["license"] != nil {
			switch v := reflect.ValueOf(meta["license"]); v.Kind() {
			case reflect.Map:
				licenses = append(licenses, *convertToLicense(v.Interface().(map[string]any)))
			case reflect.Array, reflect.Slice:
				licenses = make([]nix.License, v.Len())
				for i, v := range v.Interface().([]any) {
					switch v := reflect.ValueOf(v); v.Kind() {
					case reflect.String:
						licenses[i] = makeAdhocLicense(v.String())
					case reflect.Map:
						licenses[i] = *convertToLicense(v.Interface().(map[string]any))
					default:
						errs <- fault.Newf(
							"don't know how to handle sublicense of type %s: %v",
							v.Kind().String(),
							v,
						)
					}
				}
			case reflect.String:
				licenses = append(licenses, makeAdhocLicense(v.String()))
			default:
				errs <- fault.Newf(
					"don't know how to handle license of type %s: %v",
					v.Kind().String(),
					meta["license"],
				)
			}
			delete(meta, "license")
		}

		if meta["platforms"] != nil {
			plats := make([]any, len(meta["platforms"].([]any)))
			i := 0
			for _, plat := range meta["platforms"].([]any) {
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
					errs <- fault.Newf(
						"don't know how to convert platform type %s: %v",
						v.Kind().String(),
						v.Interface(),
					)
				}
				i++
			}
			meta["platforms"] = plats
		}

		if meta["homepage"] == nil && meta["homePage"] != nil {
			meta["homepage"] = meta["homePage"]
		}
		if meta["homepage"] != nil {
			switch v := reflect.ValueOf(meta["homepage"]); v.Kind() {
			case reflect.String:
				meta["homepage"] = []string{v.String()}
			case reflect.Slice:
			// already fine
			default:
				errs <- fault.Newf(
					"don't know how to interpret homepage type %s'",
					v.Kind().String(),
				)
			}
		}

		var maints []nix.Maintainer
		if meta["maintainers"] != nil {
			switch maint := reflect.ValueOf(meta["maintainers"]); maint.Kind() {
			case reflect.String:
				maints = []nix.Maintainer{{Name: maint.String(), Github: maint.String()}}
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
						errs <- fault.Newf(
							"don't know how to handle maintainer entry of type %s: %v",
							v.Kind().String(),
							v,
						)
					}
				}
			default:
				errs <- fault.Newf(
					"don't know how to interpret maintainers type %s'",
					maint.Kind().String(),
				)
			}
			meta["maintainers"] = maints
		}

		// Extract package name
		var pkgName string
		if pname, ok := x["pname"].(string); ok {
			pkgName = pname
		}

		// Extract version
		var version string
		if v, ok := x["version"].(string); ok {
			version = v
		}

		// Extract meta fields
		var broken bool
		if v, ok := meta["broken"].(bool); ok {
			broken = v
		}

		var description string
		if v, ok := meta["description"].(string); ok {
			description = v
		}

		var longDescription string
		if v, ok := meta["longDescription"].(string); ok {
			longDescription = v
		}

		var homepages []string
		if v, ok := meta["homepage"].([]string); ok {
			homepages = v
		} else if v, ok := meta["homepage"].([]any); ok {
			homepages = make([]string, len(v))
			for i, h := range v {
				if s, ok := h.(string); ok {
					homepages[i] = s
				}
			}
		}

		var mainProgram string
		if v, ok := meta["mainProgram"].(string); ok {
			mainProgram = v
		}

		var platforms []string
		if v, ok := meta["platforms"].([]any); ok {
			platforms = make([]string, len(v))
			for i, p := range v {
				if s, ok := p.(string); ok {
					platforms[i] = s
				}
			}
		}

		var position string
		if v, ok := meta["position"].(string); ok {
			position = v
		}

		if i.source.Programs.Enable {
			programs, err = i.programs.GetPackagePrograms(ctx, kv.Key)
			if err != nil {
				errs <- fault.Wrap(err, fmsg.Withf("failed to get programs for package %s", pkgName))
			}
		}

		pkgSet, _, found := strings.Cut(kv.Key, ".")
		if !found {
			pkgSet = ""
		}

		var definition string
		if position != "" {
			defURL, err := url.Parse(position)
			if err != nil {
				errs <- fault.Wrap(err, fmsg.Withf("failed to parse source URL %s", definition))
			}
			if defURL.IsAbs() {
				definition = position
			} else {
				subpath, line, _ := strings.Cut(position, ":")
				definition, err = i.source.Repo.GetFileURL(subpath, line)
				if err != nil {
					errs <- fault.Wrap(err, fmsg.Withf("failed to make repo URL for package %s", pkgName))
				}
			}
		}

		results <- nix.Package{
			Name:            pkgName,
			Attribute:       strings.TrimPrefix(kv.Key, "nur.repos."),
			Source:          i.source.Key,
			PackageSet:      pkgSet,
			Version:         version,
			Broken:          broken,
			Description:     description,
			LongDescription: nixdocs.Markdown(longDescription),
			Homepages:       homepages,
			MainProgram:     mainProgram,
			Platforms:       platforms,
			Licenses:        licenses,
			Maintainers:     maints,
			Definition:      definition,
			Programs:        programs,
			ImportedAt:      time.Now(),
		}
	}
}
