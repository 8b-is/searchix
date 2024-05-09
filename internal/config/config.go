package config

import (
	"log/slog"
	"maps"
	"net/url"
	"os"
	"searchix/internal/file"
	"searchix/internal/importer"

	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
)

type URL struct {
	*url.URL
}

func (u *URL) UnmarshalText(text []byte) (err error) {
	u.URL, err = url.Parse(string(text))

	if err != nil {
		return errors.WithMessage(err, "could not parse URL")
	}

	return nil
}

type Config struct {
	DataPath string `toml:"data-path"`
	CSP      CSP    `toml:"content-security-policy"`
	Headers  map[string]string
	Sources  map[string]importer.Source
}

var defaultConfig = Config{
	DataPath: "./data",
	CSP: CSP{
		DefaultSrc: []string{"'self'"},
	},
	Headers: map[string]string{
		"x-content-type-options": "nosniff",
	},
	Sources: map[string]importer.Source{
		"nixos": importer.Source{
			Name:       "NixOS",
			Enable:     true,
			Type:       importer.Channel,
			Channel:    "nixos-unstable",
			ImportPath: "nixos/release.nix",
			Attribute:  "options",
			OutputPath: "share/doc/nixos/options.json",
			Repo: importer.Repository{
				Type:  "github",
				Owner: "NixOS",
				Repo:  "nixpkgs",
			},
		},
	},
}

func GetConfig() (*Config, error) {
	config := defaultConfig
	slog.Debug("reading config.toml")
	f, err := os.Open("config.toml")
	if err := file.NeedNotExist(err); err != nil {
		return nil, errors.Wrap(err, "reading config.toml failed")
	}
	if f != nil {
		defer f.Close()
		dec := toml.NewDecoder(f)
		err = dec.Decode(&config)
		if err != nil {
			var tomlError toml.DecodeError
			if errors.As(err, &tomlError) {
				return nil, errors.WithMessage(err, tomlError.Error())
			}

			return nil, errors.Wrap(err, "config error")
		}
	}
	maps.DeleteFunc(config.Sources, func(_ string, v importer.Source) bool {
		return !v.Enable
	})

	return &config, nil
}
