package config

import (
	"strconv"
	"time"

	"github.com/pelletier/go-toml/v2"
)

var nixpkgs = Repository{
	Type:  GitHub,
	Owner: "NixOS",
	Repo:  "nixpkgs",
}

const none = "'none'"
const self = "'self'"

const maxAge = (1 * 365 * 24 * time.Hour)

var defaultConfig = Config{
	DataPath: "./data",
	Web: &Web{
		ListenAddress: "localhost",
		Port:          3000,
		BaseURL:       mustURL("http://localhost:3000"),
		Environment:   "development",
		ContentSecurityPolicy: CSP{
			DefaultSrc: []string{none},
			BaseURI:    []string{none},
			ImgSrc:     []string{self},
			StyleSrc:   []string{self},
			// added dynamically based on final value of BaseURL
			ScriptSrc:  []string{},
			FormAction: []string{self},
			ConnectSrc: []string{self},
		},
		Headers: map[string]string{
			"strict-transport-security": "max-age=" + strconv.FormatFloat(
				maxAge.Seconds(),
				'f',
				0,
				64,
			),
			"x-content-type-options": "nosniff",
			"x-frame-options":        "DENY",
		},
	},
	Importer: &Importer{
		Timeout:  Duration{30 * time.Minute},
		UpdateAt: mustLocalTime("04:00:00"),
		Sources: map[string]*Source{
			"nixos": {
				Name:       "NixOS",
				Key:        "nixos",
				Enable:     true,
				Importer:   Options,
				Fetcher:    Channel,
				Channel:    "nixpkgs",
				URL:        "https://channels.nixos.org/nixos-unstable/nixexprs.tar.xz",
				ImportPath: "nixos/release.nix",
				Attribute:  "options",
				OutputPath: "share/doc/nixos",
				Timeout:    Duration{5 * time.Minute},
				Repo:       nixpkgs,
			},
			"darwin": {
				Name:       "Darwin",
				Key:        "darwin",
				Enable:     false,
				Importer:   Options,
				Fetcher:    Channel,
				Channel:    "darwin",
				URL:        "https://github.com/LnL7/nix-darwin/archive/master.tar.gz",
				ImportPath: "release.nix",
				Attribute:  "options",
				OutputPath: "share/doc/darwin",
				Timeout:    Duration{5 * time.Minute},
				Repo: Repository{
					Type:  GitHub,
					Owner: "LnL7",
					Repo:  "nix-darwin",
				},
			},
			"home-manager": {
				Name:       "Home Manager",
				Key:        "home-manager",
				Enable:     false,
				Importer:   Options,
				Channel:    "home-manager",
				URL:        "https://github.com/nix-community/home-manager/archive/master.tar.gz",
				Fetcher:    Channel,
				ImportPath: "default.nix",
				Attribute:  "docs.json",
				OutputPath: "share/doc/home-manager",
				Timeout:    Duration{5 * time.Minute},
				Repo: Repository{
					Type:  GitHub,
					Owner: "nix-community",
					Repo:  "home-manager",
				},
			},
			"nixpkgs": {
				Name:       "Nix Packages",
				Key:        "nixpkgs",
				Enable:     true,
				Importer:   Packages,
				Fetcher:    ChannelNixpkgs,
				Channel:    "nixos-unstable",
				OutputPath: "packages.json.br",
				Timeout:    Duration{5 * time.Minute},
				Repo:       nixpkgs,
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
