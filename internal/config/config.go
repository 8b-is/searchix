package config

import (
	"log/slog"
	"maps"
	"net/url"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
)

var (
	CommitSHA string
	ShortSHA  string
)

type URL struct {
	*url.URL
}

func (u *URL) MarshalText() ([]byte, error) {
	return []byte(u.URL.String()), nil
}

func (u *URL) UnmarshalText(text []byte) (err error) {
	u.URL, err = url.Parse(string(text))

	if err != nil {
		return errors.WithMessagef(err, "could not parse URL %s", string(text))
	}

	return nil
}

type Duration struct {
	time.Duration
}

func (d *Duration) MarshalText() ([]byte, error) {
	return []byte(d.Duration.String()), nil
}

func (d *Duration) UnmarshalText(text []byte) (err error) {
	d.Duration, err = time.ParseDuration(string(text))
	if err != nil {
		return errors.WithMessagef(err, "could not parse duration %s", string(text))
	}

	return nil
}

func mustURL(in string) (u URL) {
	var err error
	u.URL, err = url.Parse(in)
	if err != nil {
		panic(errors.Errorf("URL cannot be parsed: %s", in))
	}

	return u
}

// this type is necessary as nix's `fromTOML` doesn't support TOML date/time formats
type LocalTime struct {
	toml.LocalTime
}

func (t *LocalTime) MarshalText() ([]byte, error) {
	b, err := t.LocalTime.MarshalText()
	if err != nil {
		return nil, errors.WithMessage(err, "could not marshal time value")
	}

	return b, nil
}

func (t *LocalTime) UnmarshalText(in []byte) (err error) {
	err = t.LocalTime.UnmarshalText(in)
	if err != nil {
		return errors.WithMessage(err, "could not parse time value")
	}

	return nil
}

func mustLocalTime(in string) (time LocalTime) {
	err := time.UnmarshalText([]byte(in))
	if err != nil {
		panic(errors.Errorf("Could not parse time: %s", in))
	}

	return
}

var nixpkgs = Repository{
	Type:  "github",
	Owner: "NixOS",
	Repo:  "nixpkgs",
}

var defaultConfig = Config{
	DataPath: "./data",
	Web: &Web{
		ListenAddress: "localhost",
		Port:          3000,
		BaseURL:       mustURL("http://localhost:3000"),
		Environment:   "development",
		ContentSecurityPolicy: CSP{
			DefaultSrc: []string{"'self'"},
		},
		Headers: map[string]string{
			"x-content-type-options": "nosniff",
		},
	},
	Importer: &Importer{
		Timeout:  Duration{30 * time.Minute},
		UpdateAt: mustLocalTime("04:00:00"),
		Sources: map[string]*Source{
			"nixos": {
				Name:          "NixOS",
				Key:           "nixos",
				Enable:        true,
				Importer:      Options,
				Fetcher:       Channel,
				Channel:       "nixpkgs",
				URL:           "https://channels.nixos.org/nixos-unstable/nixexprs.tar.xz",
				ImportPath:    "nixos/release.nix",
				Attribute:     "options",
				OutputPath:    "share/doc/nixos",
				FetchTimeout:  Duration{5 * time.Minute},
				ImportTimeout: Duration{15 * time.Minute},
				Repo:          nixpkgs,
			},
			"darwin": {
				Name:          "Darwin",
				Key:           "darwin",
				Enable:        false,
				Importer:      Options,
				Fetcher:       Channel,
				Channel:       "darwin",
				URL:           "https://github.com/LnL7/nix-darwin/archive/master.tar.gz",
				ImportPath:    "release.nix",
				Attribute:     "options",
				OutputPath:    "share/doc/darwin",
				FetchTimeout:  Duration{5 * time.Minute},
				ImportTimeout: Duration{15 * time.Minute},
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
				Importer:      Options,
				Channel:       "home-manager",
				URL:           "https://github.com/nix-community/home-manager/archive/master.tar.gz",
				Fetcher:       Channel,
				ImportPath:    "default.nix",
				Attribute:     "docs.json",
				OutputPath:    "share/doc/home-manager",
				FetchTimeout:  Duration{5 * time.Minute},
				ImportTimeout: Duration{15 * time.Minute},
				Repo: Repository{
					Type:  "github",
					Owner: "nix-community",
					Repo:  "home-manager",
				},
			},
			"nixpkgs": {
				Name:          "Nix Packages",
				Key:           "nixpkgs",
				Enable:        true,
				Importer:      Packages,
				Fetcher:       ChannelNixpkgs,
				Channel:       "nixos-unstable",
				OutputPath:    "packages.json.br",
				FetchTimeout:  Duration{5 * time.Minute},
				ImportTimeout: Duration{15 * time.Minute},
				Repo:          nixpkgs,
			},
		},
	},
}

func GetDefaultConfig() string {
	out, err := toml.Marshal(&defaultConfig)
	if err != nil {
		panic("could not read default configuration")
	}

	return string(out)
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
		dec.DisallowUnknownFields()
		err = dec.Decode(&config)
		if err != nil {
			var tomlError *toml.DecodeError
			if errors.As(err, &tomlError) {
				return nil, errors.WithMessage(err, tomlError.Error())
			}
			var missingConfigError *toml.StrictMissingError
			if errors.As(err, &missingConfigError) {
				return nil, errors.Errorf("unexpected config: %s", missingConfigError.String())
			}

			return nil, errors.Wrap(err, "config error")
		}
	}

	maps.DeleteFunc(config.Importer.Sources, func(_ string, v *Source) bool {
		return !v.Enable
	})

	return &config, nil
}
