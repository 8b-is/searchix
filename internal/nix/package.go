package nix

type Package struct {
	Name            string
	Source          string
	Broken          bool
	Definition      string
	Description     string
	Homepages       []string
	Licenses        []License
	LongDescription string
	MainProgram     string
	Maintainers     []Maintainer
	Platforms       []string
	Version         string
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
	return "package"
}

func (p Package) GetName() string {
	return p.Name
}

func (p Package) GetSource() string {
	return p.Source
}
