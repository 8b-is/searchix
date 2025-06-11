package nixdocs

import (
	fences "github.com/stefanfritsch/goldmark-fences"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"

	"go.alanpearce.eu/searchix/internal/nixdocs/command"
	"go.alanpearce.eu/searchix/internal/nixdocs/envvar"
	"go.alanpearce.eu/searchix/internal/nixdocs/option"
	"go.alanpearce.eu/searchix/internal/nixdocs/optlink"
	"go.alanpearce.eu/searchix/internal/nixdocs/variable"
)

func WithNixDocsExtensions() goldmark.Option {
	return goldmark.WithExtensions(
		extension.NewLinkify(),
		&fences.Extender{},
		optlink.New(),
		option.New(),
		command.New(),
		envvar.New(),
		variable.New(),
	)
}
