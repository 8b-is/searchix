package nix

import (
	"io"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"gitlab.com/tozd/go/errors"
)

type Markdown string

type Value struct {
	Text     string   `json:",omitempty"`
	Markdown Markdown `json:",omitempty"`
}

type Link struct {
	Name string
	URL  string
}

type Option struct {
	Name            string
	NameNGram       string
	Source          string
	Declarations    []Link
	Default         *Value `json:",omitempty"`
	Description     Markdown
	Example         *Value `json:",omitempty"`
	Loc             []string
	RelatedPackages Markdown `json:",omitempty"`
	Type            string
}

func (Option) BleveType() string {
	return "option"
}

func (p Option) GetName() string {
	return p.Name
}

func (p Option) GetSource() string {
	return p.Source
}

var md = goldmark.New(
	goldmark.WithExtensions(extension.NewLinkify()),
)

// implements gomponent.Node
func (text Markdown) Render(w io.Writer) error {
	return errors.WithStack(md.Convert([]byte(text), w))
}
