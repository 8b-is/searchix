package server

import (
	"fmt"
	"html"
	"html/template"
	"log/slog"
	"path"
	"path/filepath"
	"searchix/internal/options"
	"strings"

	"github.com/pkg/errors"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type TemplateCollection map[string]*template.Template

var md = goldmark.New(
	goldmark.WithExtensions(extension.NewLinkify()),
)
var templateFuncs = template.FuncMap{
	"markdown": func(input options.Markdown) template.HTML {
		var out strings.Builder
		err := md.Convert([]byte(input), &out)
		if err != nil {
			slog.Warn(fmt.Sprintf("markdown conversion failed: %v", err))

			return template.HTML(html.EscapeString(err.Error())) // #nosec G203 -- duh?
		}

		return template.HTML(out.String()) // #nosec G203
	},
}

func loadTemplate(filenames ...string) (*template.Template, error) {
	tpl := template.New(path.Base(filenames[0]))

	tpl.Funcs(templateFuncs)

	_, err := tpl.ParseFiles(filenames...)
	if err != nil {
		return nil, errors.WithMessage(err, "could not parse template")
	}

	return tpl, nil
}

func loadTemplates() (TemplateCollection, error) {
	templateDir := path.Join("frontend", "templates")
	templates := make(TemplateCollection, 0)

	layoutFile := path.Join(templateDir, "index.gotmpl")

	index, err := loadTemplate(layoutFile)
	if err != nil {
		return nil, err
	}
	templates["index"] = index

	templatePaths, err := filepath.Glob(path.Join(templateDir, "blocks", "*.gotmpl"))
	if err != nil {
		return nil, errors.WithMessage(err, "could not glob block templates")
	}
	for _, fullname := range templatePaths {
		tpl, err := loadTemplate(fullname, layoutFile)
		if err != nil {
			return nil, err
		}
		name, _ := strings.CutSuffix(path.Base(fullname), ".gotmpl")
		templates[name] = tpl
	}

	return templates, nil
}
