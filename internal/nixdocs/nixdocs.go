package nixdocs

import (
	"bytes"
	"io"

	"github.com/Southclaws/fault"
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

type Markdown string

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.NewLinkify(),
		&fences.Extender{},
		optlink.New(),
		option.New(),
		command.New(),
		envvar.New(),
		variable.New(),
		filepath.New(),
		manpage.New(),
	),
)

// implements gomponent.Node
func (text Markdown) Render(w io.Writer) error {
	return fault.Wrap(md.Convert([]byte(text), w))
}

// implements io.WriterTo
func (text Markdown) WriteTo(w io.Writer) (int64, error) {
	var buf bytes.Buffer
	err := md.Convert([]byte(text), &buf)
	if err != nil {
		return 0, fault.Wrap(err)
	}

	n, err := w.Write(buf.Bytes())

	return int64(n), fault.Wrap(err)
}
