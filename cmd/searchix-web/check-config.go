package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Southclaws/fault"
	"github.com/pelletier/go-toml/v2"

	"alin.ovh/searchix/internal/config"
)

type CheckConfig struct{}

func (*CheckConfig) Execute(_ []string) error {
	err := config.Load(string(globalOptions.Config), &config.Config{})
	if err != nil {
		var details *toml.StrictMissingError
		if errors.As(err, &details) {
			return fmt.Errorf("%s", strings.ReplaceAll(details.String(), "missing", "unknown"))
		}

		return fault.Wrap(err)
	}

	return nil
}

func init() {
	var opts CheckConfig

	_, err := parser.AddCommand(
		"check-config",
		"validate configuration",
		"Validate the configuration file",
		&opts,
	)
	if err != nil {
		panic(err)
	}
}
