package main

import (
	"os"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"

	"alin.ovh/searchix/frontend"
	"alin.ovh/searchix/internal/components"
	"alin.ovh/searchix/internal/config"
)

type GenerateErrorPageOptions struct{}

func (*GenerateErrorPageOptions) Execute(_ []string) error {
	assets, err := frontend.New()
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not create frontend"))
	}

	err = components.ErrorTemplate(components.TemplateData{
		Source: config.Source{
			Importer: config.All,
		},
		Sources:       []config.Source{},
		Query:         "",
		ExtraHeadHTML: "",
		Code:          0,
		Message:       `{{placeholder "http.error.status_code"}} {{placeholder "http.error.status_text"}}`,
		Assets:        assets,
	}).Render(os.Stdout)
	if err != nil {
		panic("failed to render error template: " + err.Error())
	}

	return nil
}

func init() {
	var opts GenerateErrorPageOptions

	_, err := parser.AddCommand(
		"generate-error-page",
		"generate error page template",
		"generate error page template for use with a reverse proxy",
		&opts,
	)
	if err != nil {
		panic(err)
	}
}
