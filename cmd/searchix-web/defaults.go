package main

import (
	"fmt"

	"github.com/Southclaws/fault"

	"alin.ovh/searchix/internal/config"
)

type PrintDefaultsOptions struct{}

func (*PrintDefaultsOptions) Execute(_ []string) error {
	_, err := fmt.Print(config.GetDefaultConfig())
	if err != nil {
		return fault.Wrap(err)
	}

	return nil
}

func init() {
	var opts PrintDefaultsOptions

	_, err := parser.AddCommand(
		"defaults",
		"print defaults",
		"print default configuration",
		&opts,
	)
	if err != nil {
		panic(err)
	}
}
