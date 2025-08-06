package main

import (
	"fmt"

	"github.com/Southclaws/fault"

	"alin.ovh/searchix/internal/config"
)

type Version struct{}

func (*Version) Execute(_ []string) error {
	_, err := fmt.Printf("searchix %s\n", config.Version)
	if err != nil {
		return fault.Wrap(err)
	}

	return nil
}

func init() {
	var version Version

	_, err := parser.AddCommand("version", "print version", "print version", &version)
	if err != nil {
		panic(err)
	}
}
