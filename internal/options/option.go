package options

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type NixValue struct {
	Type string `json:"_type" mapstructure:"_type"`
	Text string `json:"text"`
}

type HTML struct {
	HTML string
}

var md = goldmark.New(
	goldmark.WithExtensions(extension.NewLinkify()),
)

func (html *HTML) UnmarshalText(text []byte) error {
	var out strings.Builder
	err := md.Convert(text, &out)
	if err != nil {
		return errors.WithMessage(err, "failed to convert markdown to HTML")
	}

	html.HTML = out.String()

	return nil
}

func (html *HTML) UnmarshalJSON(raw []byte) error {
	var v struct {
		HTML string
	}
	err := json.Unmarshal(raw, &v)
	if err != nil {
		return errors.WithMessage(err, "error unmarshaling json")
	}
	html.HTML = v.HTML

	return nil
}

type Link struct {
	Name string
	URL  string `json:"url"`
}

type NixOption struct {
	Option          string
	Declarations    []Link
	Default         NixValue
	Description     HTML
	Example         NixValue
	ReadOnly        bool
	Type            string
	Loc             []string
	RelatedPackages HTML
}

type NixOptions []NixOption
