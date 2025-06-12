package nixdocs

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

func WithNixDocsExtensions() goldmark.Option {
	return goldmark.WithExtensions(
		extension.NewLinkify(),
	)
}
