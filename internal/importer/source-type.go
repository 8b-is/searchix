package importer

import (
	"fmt"
	"time"

	"github.com/stoewer/go-strcase"
)

type Type int

const (
	Unknown = iota
	Channel
	ChannelNixpkgs
)

func (f Type) String() string {
	switch f {
	case Channel:
		return "channel"
	case ChannelNixpkgs:
		return "channel-nixpkgs"
	}

	return fmt.Sprintf("Fetcher(%d)", f)
}

func parseType(name string) (Type, error) {
	switch strcase.KebabCase(name) {
	case "channel":
		return Channel, nil
	case "channel-nixpkgs":
		return ChannelNixpkgs, nil
	default:
		return Unknown, fmt.Errorf("unsupported fetcher %s", name)
	}
}

func (f *Type) UnmarshalText(text []byte) error {
	var err error
	*f, err = parseType(string(text))

	return err
}

type Source struct {
	Name          string
	Key           string
	Enable        bool
	Type          Type
	Channel       string
	URL           string
	Attribute     string
	ImportPath    string        `toml:"import-path"`
	FetchTimeout  time.Duration `toml:"fetch-timeout"`
	ImportTimeout time.Duration `toml:"import-timeout"`
	OutputPath    string        `toml:"output-path"`
	Repo          Repository
}
