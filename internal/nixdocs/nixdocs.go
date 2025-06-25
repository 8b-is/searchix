package nixdocs

import (
	fences "github.com/stefanfritsch/goldmark-fences"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"

	"alin.ovh/searchix/internal/nixdocs/command"
	"alin.ovh/searchix/internal/nixdocs/envvar"
	"alin.ovh/searchix/internal/nixdocs/filepath"
	"alin.ovh/searchix/internal/nixdocs/manpage"
	"alin.ovh/searchix/internal/nixdocs/option"
	"alin.ovh/searchix/internal/nixdocs/optlink"
	"alin.ovh/searchix/internal/nixdocs/variable"
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
		filepath.New(),
		manpage.New(),
	)
}
