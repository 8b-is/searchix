package nix

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
