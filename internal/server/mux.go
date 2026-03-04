package server

import (
	"cmp"
	"maps"
	"net/http"
	"path"
	"slices"
	"strings"
	"time"

	"alin.ovh/searchix/frontend"
	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/sentryhttp"

	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/osdevisnot/sorvor/pkg/livereload"
)

type HTTPError struct {
	Error   error
	Message string
	Code    int
}

var sources []config.Source

func applyDevModeOverrides(cfg *config.Config) {
	if len(cfg.Web.ContentSecurityPolicy.ScriptSrc) == 0 {
		cfg.Web.ContentSecurityPolicy.ScriptSrc = cfg.Web.ContentSecurityPolicy.DefaultSrc
	}
	cfg.Web.ContentSecurityPolicy.ScriptSrc = append(
		cfg.Web.ContentSecurityPolicy.ScriptSrc,
		"'unsafe-inline'",
	)
	cfg.Web.GracefulShutdownTimeout.Duration = 200 * time.Millisecond
}

func sortSources(ss map[string]config.Source) {
	sources = slices.SortedFunc(maps.Values(ss), func(a, b config.Source) int {
		return cmp.Or(
			cmp.Compare(a.Order, b.Order),
			strings.Compare(a.Key, b.Key),
			strings.Compare(a.Name, b.Name),
		)
	})
}

func NewMux(
	cfg *config.Config,
	options *Options,
	log *log.Logger,
	liveReload bool,
) (*http.ServeMux, error) {
	if cfg == nil {
		return nil, fault.New("cfg is nil")
	}
	if options.ReadIndex == nil {
		return nil, fault.New("read index is nil")
	}
	sortSources(cfg.Importer.Sources)
	assets, err := frontend.New()
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("could not create frontend asset collection"))
	}

	handler := &GlobalHandler{
		assets:       assets,
		cfg:          cfg,
		log:          log,
		index:        options.ReadIndex,
		mdb:          options.ManpagesURLMap,
		errorHandler: createErrorHandler(cfg, assets, log),
	}

	mux := sentryhttp.NewServeMux()

	// Register global routes
	mux.HandleFunc("/{$}", handler.CombinedSearch)
	mux.HandleFunc("/opensearch.xml", handler.RootOpenSearch)

	mux.HandleFunc("/", handler.StaticFile)
	mux.HandleFunc("/assets/", handler.Asset)
	mux.HandleFunc("/man/{section}/{page}", handler.ManualPage)

	mux.Handle("/all/search", http.RedirectHandler("/", http.StatusFound))
	mux.Handle("/all/opensearch.xml", http.RedirectHandler("/opensearch.xml", http.StatusFound))

	for _, it := range []config.ImporterType{config.Options, config.Packages} {
		singularPath := path.Join("/", it.Singular())
		mux.HandleFunc(singularPath+"/", func(w http.ResponseWriter, r *http.Request) {
			newPath := path.Join("/", it.String(), strings.TrimPrefix(r.URL.Path, singularPath))
			http.Redirect(w, r, newPath, http.StatusFound)
		})
	}

	for _, source := range cfg.Importer.Sources {
		mux.Handle(source.LocalURL().String(), handler.NewSourceHandler(source))
	}

	top := http.NewServeMux()
	if liveReload {
		applyDevModeOverrides(cfg)
		cfg.Web.ExtraHeadHTML = livereload.JsSnippet
		liveReload := livereload.New()
		liveReload.Start()
		top.Handle("/livereload", liveReload)
		fw, err := NewFileWatcher(log.Named("watcher"), "frontend")
		if err != nil {
			return nil, fault.Wrap(err, fmsg.With("could not create file watcher"))
		}
		err = fw.AddRecursive(".")
		if err != nil {
			return nil, fault.Wrap(err, fmsg.With("could not add directory to file watcher"))
		}
		go fw.Start(func(filename string) {
			if match, _ := path.Match("frontend/static/*", filename); match {
				err := assets.Rehash()
				if err != nil {
					log.Error("failed to re-hash frontend assets", "error", err)
				}
			}
			liveReload.Reload()
		})
	}

	top.Handle("/",
		AddHeadersMiddleware(
			wrapHandlerWithLogging(mux, wrappedHandlerOptions{
				defaultHostname: cfg.Web.BaseURL.Hostname(),
				logger:          log,
				enabled:         cfg.Web.LogRequests,
			}),
			cfg,
		),
	)
	// no logging, no sentry
	top.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return top, nil
}
