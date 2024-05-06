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
	Option string

	Declarations    []Link
	Default         *NixValue `json:",omitempty"`
	Description     Markdown
	Example         *NixValue `json:",omitempty"`
	Loc             []string
	RelatedPackages Markdown `json:",omitempty"`
	Type            string
}

type NixOptions []NixOption
