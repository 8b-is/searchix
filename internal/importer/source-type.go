package importer

import (
	"fmt"

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
