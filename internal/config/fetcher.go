package config

import (
	"fmt"

	"github.com/stoewer/go-strcase"
	"gitlab.com/tozd/go/errors"
)

type Fetcher int

const (
	UnknownFetcher = iota
	Channel
	ChannelNixpkgs
	Download
)

func (f Fetcher) String() string {
	switch f {
	case Channel:
		return "channel"
	case ChannelNixpkgs:
		return "channel-nixpkgs"
	case Download:
		return "download"
	}

	return fmt.Sprintf("Fetcher(%d)", f)
}

func ParseFetcher(name string) (Fetcher, error) {
	switch strcase.KebabCase(name) {
	case "channel":
		return Channel, nil
	case "channel-nixpkgs":
		return ChannelNixpkgs, nil
	case "download":
		return Download, nil
	default:
		return UnknownFetcher, errors.Errorf("unsupported fetcher %s", name)
	}
}

func (f *Fetcher) UnmarshalText(text []byte) error {
	var err error
	*f, err = ParseFetcher(string(text))

	return err
}

func (f *Fetcher) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}
