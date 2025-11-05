package nix

import (
	"time"

	"alin.ovh/searchix/internal/config"
)

type Package struct {
	Name            string
	Attribute       string
	Source          string
	Broken          bool
	Definition      string
	Description     string
	Homepages       []string
	Licenses        []License
	LongDescription Markdown
	MainProgram     string
	Programs        []string
	Maintainers     []Maintainer
	PackageSet      string
	Platforms       []string
	Version         string
	ImportedAt      time.Time
}

type License struct {
	FullName    string
	Name        string
	SPDXId      string
	URL         string
	AppendixURL string
}

type Maintainer struct {
	Github string
	Name   string
}

func (Package) BleveType() string {
	return config.Packages.Singular()
}

func (Package) ImporterType() string {
	return config.Packages.String()
}

func (p Package) GetName() string {
	return p.Attribute
}

func (p Package) GetSource() string {
	return p.Source
}
