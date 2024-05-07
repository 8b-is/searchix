package server

import (
	"fmt"
	"html"
	"html/template"
	"log/slog"
	"os"
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

func loadTemplate(filename string) (*template.Template, error) {
	text, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.WithMessage(err, "could not read template")
	}
	name, _ := strings.CutSuffix(path.Base(filename), ".gotmpl")
	tpl := template.New(name)
	tpl.Funcs(templateFuncs)
	_, err = tpl.Parse(string(text))
	if err != nil {
		return nil, errors.WithMessage(err, "could not parse template")
	}

	return tpl, nil
}

func loadTemplates() (TemplateCollection, error) {
	templateDir := path.Join("frontend", "templates")
	templates := make(TemplateCollection, 0)

	index, err := loadTemplate(path.Join(templateDir, "index.gotmpl"))
	if err != nil {
		return nil, err
	}
	templates["index"] = index

	templatePaths, err := filepath.Glob(path.Join(templateDir, "blocks", "*.gotmpl"))
	if err != nil {
		return nil, errors.WithMessage(err, "could not glob block templates")
	}
	for _, fullname := range templatePaths {
		tpl, err := loadTemplate(fullname)
		if err != nil {
			return nil, err
		}
		_, err = tpl.AddParseTree("index", index.Tree)
		if err != nil {
			return nil, errors.WithMessage(err, "could not add index template")
		}
		templates[tpl.Name()] = tpl
	}

	return templates, nil
}
