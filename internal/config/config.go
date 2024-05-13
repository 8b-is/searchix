package config

import (
	"html/template"
	"log/slog"
	"maps"
	"net/url"
	"os"
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
	DataPath      string        `toml:"data-path"`
	CSP           CSP           `toml:"content-security-policy"`
	ExtraBodyHTML template.HTML `toml:"extra-body-html"`
	Headers       map[string]string
	Sources       map[string]*Source
}

var defaultConfig = Config{
	DataPath: "./data",
	CSP: CSP{
		DefaultSrc: []string{"'self'"},
	},
	Headers: map[string]string{
		"x-content-type-options": "nosniff",
	},
	Sources: map[string]*Source{
		"nixos": {
			Name:          "NixOS",
			Key:           "nixos",
			Enable:        true,
			Type:          Channel,
			Channel:       "nixpkgs",
			URL:           "https://channels.nixos.org/nixos-unstable/nixexprs.tar.xz",
			ImportPath:    "nixos/release.nix",
			Attribute:     "options",
			OutputPath:    "share/doc/nixos/options.json",
			FetchTimeout:  5 * time.Minute,
			ImportTimeout: 15 * time.Minute,
			Repo: Repository{
				Type:  "github",
				Owner: "NixOS",
				Repo:  "nixpkgs",
			},
		},
		"darwin": {
			Name:          "Darwin",
			Key:           "darwin",
			Enable:        false,
			Type:          Channel,
			Channel:       "darwin",
			URL:           "https://github.com/LnL7/nix-darwin/archive/master.tar.gz",
			ImportPath:    "release.nix",
			Attribute:     "options",
			OutputPath:    "share/doc/darwin/options.json",
			FetchTimeout:  5 * time.Minute,
			ImportTimeout: 15 * time.Minute,
			Repo: Repository{
				Type:  "github",
				Owner: "LnL7",
				Repo:  "nix-darwin",
			},
		},
		"home-manager": {
			Name:          "Home Manager",
			Key:           "home-manager",
			Enable:        false,
			Channel:       "home-manager",
			URL:           "https://github.com/nix-community/home-manager/archive/master.tar.gz",
			Type:          Channel,
			ImportPath:    "default.nix",
			Attribute:     "docs.json",
			OutputPath:    "share/doc/home-manager/options.json",
			FetchTimeout:  5 * time.Minute,
			ImportTimeout: 15 * time.Minute,
			Repo: Repository{
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
			var tomlError *toml.DecodeError
			if errors.As(err, &tomlError) {
				return nil, errors.WithMessage(err, tomlError.Error())
			}

			return nil, errors.Wrap(err, "config error")
		}
	}

	maps.DeleteFunc(config.Sources, func(_ string, v *Source) bool {
		return !v.Enable
	})

	return &config, nil
}
