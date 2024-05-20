package config

import (
	"fmt"

	"github.com/stoewer/go-strcase"
)

type ImporterType int

const (
	UnknownType = iota
	Packages
	Options
)

func (i ImporterType) String() string {
	switch i {
	case Packages:
		return "packages"
	case Options:
		return "options"
	}

	return fmt.Sprintf("Type(%d)", i)
}

func ParseImporterType(name string) (ImporterType, error) {
	switch strcase.KebabCase(name) {
	case "packages":
		return Packages, nil
	case "options":
		return Options, nil
	default:
		return UnknownType, fmt.Errorf("unsupported importer %s", name)
	}
}

func (i *ImporterType) UnmarshalText(text []byte) error {
	var err error
	*i, err = ParseImporterType(string(text))

	return err
}

func (i *ImporterType) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}
