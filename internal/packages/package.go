package packages

type Package struct {
	Name    string
	Meta    Meta
	Version string
}

type Meta struct {
	Broken          bool
	Description     string
	LongDescription string
	Homepages       []string
	Licenses        []License
	MainProgram     string
	Maintainers     []Maintainer
	Platforms       []string
	Position        string
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

func (*Package) BleveType() string {
	return "package"
}
