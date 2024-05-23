//nolint:lll
package config

// keep config structs here so that lll ignores the long lines (go doesn't support multi-line struct tags)

import (
	"html/template"
	"log/slog"
)

type Config struct {
	DataPath string     `comment:"Path to store index data."`
	LogLevel slog.Level `comment:"How much information to log, one of 'debug', 'info', 'warn', 'error'."`
	Web      *Web       `comment:"Settings for the web server"`
	Importer *Importer  `comment:"Settings for the import job"`
}

type Web struct {
	ContentSecurityPolicy CSP               `comment:"Content-Security-Policy header to send with requests. Should only need changing if ExtraHeadHTML is used."`
	ListenAddress         string            `comment:"Which address or hostname to listen on. IPv6 addresses need square brackets."`
	Port                  int               `comment:"Port number to listen on."`
	BaseURL               URL               `comment:"Absolute URL to this instance, useful if behind a reverse proxy"`
	SentryDSN             string            `comment:"If set, will send server errors to Sentry"`
	Environment           string            `comment:"Affects logging parameters. One of 'development' or 'production'"`
	ExtraHeadHTML         template.HTML     `comment:"Content to add to HTML <head>. Can be used to override styling, add scripts, etc."`
	Headers               map[string]string `comment:"Extra headers to send with HTTP requests"`
}

type Importer struct {
	Sources  map[string]*Source
	Timeout  Duration  `comment:"Abort fetch and import process for all jobs if it takes longer than this value."`
	UpdateAt LocalTime `comment:"Local time of day to run fetch/import process"`
}

type Source struct {
	Name       string       `comment:"Human-readable name of source for generating links"`
	Key        string       `comment:"Machine-readable name of source. Must be URL- and path-safe."`
	Enable     bool         `comment:"Controls whether to show in the web interface and to run fetch/import jobs."`
	Fetcher    Fetcher      `comment:"How to fetch options.json. One of 'channel', 'channel-nixpkgs' or 'download'."`
	Importer   ImporterType `comment:"Kind of data available from source. Currently supports 'packages' and 'options'."`
	Channel    string       `comment:"(Fetcher=channel) Local name for channel, (Fetcher=channel-nixpkgs) Remote name of channel."`
	URL        string       `comment:"(Fetcher=channel) Remote URL for channel, (Fetcher=download) Path containing files named 'revision' and 'options.json'."`
	Attribute  string       `comment:"(Fetcher=channel) Nix attribute name (i.e. nix-build -A) that builds an {options,packages}.json"`
	ImportPath string       `comment:"(Fetcher=channel) Sub-path of imported channel which contains the attribute above, e.g. release.nix"`
	Timeout    Duration     `comment:"Abort import if it takes longer than this."`
	OutputPath string       `comment:"(Fetcher=channel) Path under ./result symlink to folder containing {options,packages}.json."`
	Repo       Repository   `comment:"Used to generate declaration/definition links"`
}
