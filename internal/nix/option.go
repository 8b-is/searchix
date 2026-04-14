package nix

import (
	"time"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/nixdocs"
)

type Docs struct {
	Plain    string           `json:",omitempty"`
	Markdown nixdocs.Markdown `json:",omitempty"`
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
	Description     nixdocs.Markdown
	Example         *Docs `json:",omitempty"`
	Loc             []string
	OptionSet       string
	Parents         string
	RelatedPackages nixdocs.Markdown `json:",omitempty"`
	Type            string
	ImportedAt      time.Time
}

func (Option) BleveType() string {
	return config.Options.Singular()
}

func (Option) ImporterType() string {
	return config.Options.String()
}

func (p Option) GetName() string {
	return p.Name
}

func (p Option) GetSource() string {
	return p.Source
}
