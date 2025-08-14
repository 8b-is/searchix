//nolint:lll
package config

// keep config structs here so that lll ignores the long lines (go doesn't support multi-line struct tags)

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/creasty/defaults"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	DataPath string        `comment:"Path to store index data."`
	LogLevel zapcore.Level `comment:"How much information to log, one of 'debug', 'info', 'warn', 'error', 'panic', 'fatal'."`
	Web      *Web          `comment:"Settings for the web server"`
	Importer *Importer     `comment:"Settings for the import job"`
}

type Web struct {
	ContentSecurityPolicy CSP               `comment:"Content-Security-Policy header to send with requests. Should only need changing if ExtraHeadHTML is used."`
	ListenAddress         string            `comment:"Which address or hostname to listen on. IPv6 addresses need square brackets."`
	Port                  int               `comment:"Port number to listen on."`
	BaseURL               URL               `comment:"Absolute URL to this instance, useful if behind a reverse proxy"`
	SentryDSN             string            `comment:"If set, will send server errors to Sentry"`
	Environment           string            `comment:"Affects logging parameters. One of 'development' or 'production'"`
	ExtraHeadHTML         string            `comment:"Content to add to HTML <head>. Can be used to override styling, add scripts, etc."`
	Headers               map[string]string `comment:"Extra headers to send with HTTP requests"`
	LogRequests           bool              `comment:"Whether to log incoming HTTP requests"`
	SearchTimeout         Duration          `comment:"Timeout for search requests"`
}

type Importer struct {
	Sources   map[string]*Source
	LowMemory bool      `comment:"Use less memory at the expense of import performance"`
	BatchSize int       `comment:"Number of items to process in each batch (affects memory usage)."`
	Timeout   Duration  `comment:"Abort fetch and import process for all jobs if it takes longer than this value."`
	UpdateAt  LocalTime `comment:"Time of day (UTC) to run fetch/import process"`
}

type Source struct {
	Name       string       `default:"-"               comment:"Human-readable name of source for generating links"`
	Order      uint         `                          comment:"Order in which to show source in web interface."`
	Key        string       `                          comment:"Machine-readable name of source. Must be URL- and path-safe."`
	Enable     bool         `                          comment:"Controls whether to show in the web interface and to run fetch/import jobs."`
	Fetcher    Fetcher      `                          comment:"How to fetch options.json. One of 'channel', 'channel-nixpkgs' or 'download'."`
	Importer   ImporterType `default:"options"         comment:"Kind of data available from source. Currently supports 'packages' and 'options'."`
	Channel    string       `                          comment:"(Fetcher=channel) Local name for channel, (Fetcher=channel-nixpkgs) Remote name of channel."`
	URL        string       `                          comment:"(Fetcher=channel) Remote URL for channel, (Fetcher=download) Path containing files named 'revision' and 'options.json'."`
	Attribute  string       `default:"options"         comment:"(Fetcher=channel) Nix attribute name (i.e. nix-build -A) that builds an {options,packages}.json"`
	ImportPath string       `                          comment:"(Fetcher=channel) Sub-path of imported channel which contains the attribute above, e.g. release.nix"`
	Timeout    Duration     `default:"5m"              comment:"Abort import if it takes longer than this."`
	OutputPath string       `default:"share/doc/nixos" comment:"(Fetcher=channel) Path under ./result symlink to folder containing {options,packages}.json."`
	Repo       Repository   `                          comment:"Used to generate declaration/definition links"`
	Programs   ProgramsDB   `                          comment:"Used to enable searching for programs in multi-program packages"`
	Manpages   Manpages     `                          comment:"Used to enable searching for manpages"`
	JSONDepth  int          `default:"1"               comment:"Depth at which packages/object object is to be found"`
}

type ProgramsDB struct {
	Enable    bool   `comment:"Enable searching for programs in multi-program packages"`
	Attribute string `comment:"Nix attribute name (i.e. nix-instantiate) that builds a programs.sqlite file"`
}

type Manpages struct {
	Enable bool   `comment:"Enable searching for manpages"`
	Path   string `comment:"Path to the manpage-urls.json file from repository root"`
}

func (source *Source) String() string {
	switch source.Importer {
	case Options:
		return source.Name + " " + source.Importer.String()
	case Packages:
		return source.Name
	default:
		return fmt.Sprintf("Source(%s)", source.Name)
	}
}

func (source *Source) SetDefaults() {
	if defaults.CanUpdate(source.Name) {
		source.Name = strings.ToTitle(source.Key[0:1]) + source.Key[1:]
	}
}

func (source *Source) JoinPath(name string) string {
	//nolint:forbidigo // is not absolute
	return filepath.Join("sources", source.Key, name)
}
