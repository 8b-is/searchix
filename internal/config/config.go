package config

import (
	"log/slog"
	"maps"
	"net/url"
	"os"
	"searchix/internal/importer"
	"time"

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
		"nixos": {
			Name:          "NixOS",
			Enable:        true,
			Type:          importer.Channel,
			Channel:       "nixpkgs",
			ImportPath:    "nixos/release.nix",
			Attribute:     "options",
			OutputPath:    "share/doc/nixos/options.json",
			FetchTimeout:  5 * time.Minute,
			ImportTimeout: 15 * time.Minute,
			Repo: importer.Repository{
				Type:  "github",
				Owner: "NixOS",
				Repo:  "nixpkgs",
			},
		},
		"darwin": {
			Name:          "darwin",
			Enable:        false,
			Type:          importer.Channel,
			Channel:       "nix-darwin",
			ImportPath:    "release.nix",
			Attribute:     "options",
			OutputPath:    "share/doc/darwin/options.json",
			FetchTimeout:  5 * time.Minute,
			ImportTimeout: 15 * time.Minute,
			Repo: importer.Repository{
				Type:  "github",
				Owner: "LnL7",
				Repo:  "nix-darwin",
			},
		},
		"home-manager": {
			Name:          "home-manager",
			Enable:        false,
			Type:          importer.Channel,
			ImportPath:    "default.nix",
			Attribute:     "docs.json",
			OutputPath:    "share/doc/home-manager/options.json",
			FetchTimeout:  5 * time.Minute,
			ImportTimeout: 15 * time.Minute,
			Repo: importer.Repository{
				Type:  "github",
				Owner: "nix-community",
				Repo:  "home-manager",
			},
		},
	},
}

func GetConfig(filename string) (*Config, error) {
	config := defaultConfig
	if filename != "" {
		slog.Debug("reading config", "filename", filename)
		f, err := os.Open(filename)
		if err != nil {
			return nil, errors.Wrap(err, "reading config failed")
		}
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
