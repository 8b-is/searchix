package nixdocs

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"

	"go.alanpearce.eu/searchix/internal/nixdocs/optlink"
)

func WithNixDocsExtensions() goldmark.Option {
	return goldmark.WithExtensions(
		extension.NewLinkify(),
		optlink.New(),
	)
}
