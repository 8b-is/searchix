package server

import (
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"log/slog"
	"path"
	"regexp"
	"searchix/frontend"
	"searchix/internal/config"
	"searchix/internal/nix"
	"strings"

	"github.com/pkg/errors"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type TemplateCollection map[string]*template.Template

var (
	md = goldmark.New(
		goldmark.WithExtensions(extension.NewLinkify()),
	)
	firstSentenceRegexp = regexp.MustCompile(`^.*?\.[[:space:]]`)
)
var templateFuncs = template.FuncMap{
	"firstSentence": func(input nix.Markdown) nix.Markdown {
		if fs := firstSentenceRegexp.FindString(string(input)); fs != "" {
			return nix.Markdown(fs)
		}

		return input
	},
	"markdown": func(input nix.Markdown) template.HTML {
		var out strings.Builder
		err := md.Convert([]byte(input), &out)
		if err != nil {
			slog.Warn(fmt.Sprintf("markdown conversion failed: %v", err))

			return template.HTML(html.EscapeString(err.Error())) // #nosec G203 -- duh?
		}

		return template.HTML(out.String()) // #nosec G203
	},
	"sourceNameAndType": func(source config.Source) (string, error) {
		switch source.Importer {
		case config.Options:
			return source.Name + " " + source.Importer.String(), nil
		case config.Packages:
			return source.Name, nil
		default:
			return "", errors.Errorf("unknown source importer type %s", source.Importer.String())
		}
	},
	"sourceName": func(input string) string {
		switch input {
		case "nixos":
			return "NixOS"
		case "darwin":
			return "Darwin"
		case "home-manager":
			return "Home Manager"
		}

		return input
	},
}

func loadTemplate(
	layoutFile string,
	filenames ...string,
) (*template.Template, error) {
	tpl := template.New("index.gotmpl")

	tpl.Funcs(templateFuncs)
	_, err := tpl.ParseFS(frontend.Files, layoutFile)
	if err != nil {
		return nil, errors.WithMessage(err, "could not parse layout template")
	}

	if len(filenames) > 0 {
		_, err = tpl.ParseFS(frontend.Files, filenames...)
		if err != nil {
			return nil, errors.WithMessage(err, "could not parse template")
		}
	}

	return tpl, nil
}

func loadTemplates() (TemplateCollection, error) {
	templateDir := "templates"
	templates := make(TemplateCollection, 0)

	glob := path.Join(templateDir, "*.gotmpl")
	templatePaths, err := fs.Glob(frontend.Files, glob)
	if err != nil {
		return nil, errors.WithMessage(err, "could not glob main templates")
	}
	for _, fullname := range templatePaths {
		tpl, err := loadTemplate(fullname)
		if err != nil {
			return nil, err
		}
		name, _ := strings.CutSuffix(path.Base(fullname), ".gotmpl")
		templates[name] = tpl
	}

	return templates, nil
}
