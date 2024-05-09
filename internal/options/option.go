package options

type Markdown string

type NixValue struct {
	Text     string   `json:",omitempty"`
	Markdown Markdown `json:",omitempty"`
}

type Link struct {
	Name string
	URL  string
}

type NixOption struct {
	Name   string
	Source string

	Declarations    []Link
	Default         *NixValue `json:",omitempty"`
	Description     Markdown
	Example         *NixValue `json:",omitempty"`
	Loc             []string
	RelatedPackages Markdown `json:",omitempty"`
	Type            string
}

func (*NixOption) BleveType() string {
	return "option"
}
