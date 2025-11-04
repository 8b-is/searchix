package config

import (
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

var nixpkgs = Repository{
	Type:  GitHub,
	Owner: "NixOS",
	Repo:  "nixpkgs",
}

const (
	none = "'none'"
	self = "'self'"
)

const maxAge = (1 * 365 * 24 * time.Hour)

var DefaultConfig = Config{
	DataPath: "./data",
	Web: Web{
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
		LogRequests:   true,
		SearchTimeout: Duration{1 * time.Second},
	},
	Importer: Importer{
		LowMemory: false,
		BatchSize: 10_000,
		Timeout:   Duration{30 * time.Minute},
		UpdateAt:  mustLocalTime("03:00:00"),
		Sources: map[string]Source{
			"nixos": {
				Name:       "NixOS",
				Order:      0,
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
				Manpages: Manpages{
					Enable: false,
					Path:   "",
				},
				JSONDepth: 1,
			},
			"darwin": {
				Name:       "Darwin",
				Order:      1,
				Key:        "darwin",
				Enable:     false,
				Importer:   Options,
				Fetcher:    Channel,
				Channel:    "darwin",
				URL:        "https://github.com/LnL7/nix-darwin/archive/master.tar.gz",
				ImportPath: "release.nix",
				Attribute:  "docs.optionsJSON",
				OutputPath: "share/doc/darwin",
				Timeout:    Duration{5 * time.Minute},
				Repo: Repository{
					Type:  GitHub,
					Owner: "LnL7",
					Repo:  "nix-darwin",
				},
				Manpages: Manpages{
					Enable: false,
					Path:   "",
				},
				JSONDepth: 1,
			},
			"home-manager": {
				Name:       "Home Manager",
				Order:      2,
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
				Manpages: Manpages{
					Enable: false,
					Path:   "",
				},
				JSONDepth: 1,
			},
			"nixpkgs": {
				Name:       "Nix Packages",
				Order:      3,
				Key:        "nixpkgs",
				Enable:     true,
				Importer:   Packages,
				Fetcher:    ChannelNixpkgs,
				Channel:    "nixos-unstable",
				OutputPath: "packages.json.br",
				Timeout:    Duration{15 * time.Minute},
				Repo:       nixpkgs,
				Programs: ProgramsDB{
					Enable:    true,
					Attribute: "programs.sqlite",
				},
				Manpages: Manpages{
					Enable: true,
					Path:   "/doc/manpage-urls.json",
				},
				JSONDepth: 2,
			},
			"nur": {
				Name:     "NUR",
				Order:    4,
				Key:      "nur",
				Enable:   false,
				Importer: Packages,
				Fetcher:  Download,
				URL:      "https://alanpearce.github.io/nix-options/nur",
				Timeout:  Duration{5 * time.Minute},
				Repo: Repository{
					Type:  GitHub,
					Owner: "nix-community",
					Repo:  "nur",
				},
				Manpages: Manpages{
					Enable: false,
					Path:   "",
				},
				JSONDepth: 1,
			},
		},
	},
}

func GetDefaultConfig() string {
	var out strings.Builder

	enc := toml.NewEncoder(&out)

	err := enc.Encode(DefaultConfig)
	if err != nil {
		panic("could not read default configuration")
	}

	return out.String()
}
