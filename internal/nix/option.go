package nix

import (
	"io"
	"time"

	"github.com/Southclaws/fault"
	"github.com/yuin/goldmark"

	"alin.ovh/searchix/internal/nixdocs"
)

type Markdown string

type Docs struct {
	Plain    string   `json:",omitempty"`
	Markdown Markdown `json:",omitempty"`
}

type Link struct {
	Name string
	URL  string
}

type Option struct {
	Name            string
	Source          string
	Declarations    []Link
	Default         *Docs `json:",omitempty"`
	Description     Markdown
	Example         *Docs `json:",omitempty"`
	Loc             []string
	Parents         string
	RelatedPackages Markdown `json:",omitempty"`
	Type            string
	ImportedAt      time.Time
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
	nixdocs.WithNixDocsExtensions(),
)

// implements gomponent.Node
func (text Markdown) Render(w io.Writer) error {
	return fault.Wrap(md.Convert([]byte(text), w))
}
