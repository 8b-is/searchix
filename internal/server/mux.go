package server

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	"searchix/frontend"
	"searchix/internal/config"
	search "searchix/internal/index"
	"searchix/internal/options"

	"github.com/blevesearch/bleve/v2"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/osdevisnot/sorvor/pkg/livereload"
	"github.com/pkg/errors"
	"github.com/shengyanli1982/law"
)

type HTTPError struct {
	Error   error
	Message string
	Code    int
}

const jsSnippet = template.HTML(livereload.JsSnippet) // #nosec G203

type VersionInfo struct {
	ShortSHA  string
	CommitSHA string
}

type TemplateData struct {
	Sources       map[string]*config.Source
	Source        config.Source
	Query         string
	Results       bool
	SourceResult  *bleve.SearchResult
	ExtraBodyHTML template.HTML
	Version       VersionInfo
}

type ResultData[T options.NixOption] struct {
	TemplateData
	Query          string
	ResultsPerPage int
	Results        *search.Result
	Prev           string
	Next           string
}

var versionInfo = &VersionInfo{
	ShortSHA:  config.ShortSHA,
	CommitSHA: config.CommitSHA,
}

func applyDevModeOverrides(config *config.Config) {
	if len(config.Web.ContentSecurityPolicy.ScriptSrc) == 0 {
		config.Web.ContentSecurityPolicy.ScriptSrc = config.Web.ContentSecurityPolicy.DefaultSrc
	}
	config.Web.ContentSecurityPolicy.ScriptSrc = append(
		config.Web.ContentSecurityPolicy.ScriptSrc,
		"'unsafe-inline'",
	)
}

func NewMux(
	config *config.Config,
	index *search.ReadIndex,
	liveReload bool,
) (*http.ServeMux, error) {
	err := sentry.Init(sentry.ClientOptions{
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		Dsn:              config.Web.SentryDSN,
		Environment:      config.Web.Environment,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "could not set up sentry")
	}
	defer sentry.Flush(2 * time.Second)
	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})

	templates, err := loadTemplates()
	if err != nil {
		log.Panicf("could not load templates: %v", err)
	}

	top := http.NewServeMux()
	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, _ *http.Request) {
		indexData := TemplateData{
			ExtraBodyHTML: config.Web.ExtraBodyHTML,
			Sources:       config.Importer.Sources,
			Version:       *versionInfo,
		}
		err := templates["index"].ExecuteTemplate(w, "index.gotmpl", indexData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	const searchTimeout = 1 * time.Second
	mux.HandleFunc("/options/{source}/search", func(w http.ResponseWriter, r *http.Request) {
		sourceKey := r.PathValue("source")

		source := config.Importer.Sources[sourceKey]
		if source == nil {
			http.Error(w, "Source not found", http.StatusNotFound)

			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), searchTimeout)
		defer cancel()

		if r.URL.Query().Has("query") {
			qs := r.URL.Query().Get("query")
			pg := r.URL.Query().Get("page")
			var page uint64 = 1
			if pg != "" {
				page, err = strconv.ParseUint(pg, 10, 64)
				if err != nil || page == 0 {
					http.Error(w, "Bad query string", http.StatusBadRequest)
				}
			}
			results, err := index.Search(ctx, sourceKey, qs, (page-1)*search.ResultsPerPage)
			if err != nil {
				if err == context.DeadlineExceeded {
					http.Error(w, "Search timed out", http.StatusInternalServerError)

					return
				}
				slog.Error("search error", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			tdata := ResultData[options.NixOption]{
				TemplateData: TemplateData{
					ExtraBodyHTML: config.Web.ExtraBodyHTML,
					Source:        *source,
					Sources:       config.Importer.Sources,
					Version:       *versionInfo,
				},
				ResultsPerPage: search.ResultsPerPage,
				Query:          qs,
				Results:        results,
			}

			hits := uint64(len(results.Hits))
			if results.Total > hits {
				q, err := url.ParseQuery(r.URL.RawQuery)
				if err != nil {
					http.Error(w, "Query string error", http.StatusBadRequest)

					return
				}

				if page*search.ResultsPerPage > results.Total {
					http.Error(w, "Not found", http.StatusNotFound)

					return
				}

				if page*search.ResultsPerPage < results.Total {
					q.Set("page", strconv.FormatUint(page+1, 10))
					tdata.Next = "search?" + q.Encode()
				}

				if page > 1 {
					p := page - 1
					if p == 1 {
						q.Del("page")
					} else {
						q.Set("page", strconv.FormatUint(p, 10))
					}
					tdata.Prev = "search?" + q.Encode()
				}
			}

			if r.Header.Get("Fetch") == "true" {
				w.Header().Add("Content-Type", "text/html; charset=utf-8")
				err = templates["options"].ExecuteTemplate(w, "options.gotmpl", tdata)
			} else {
				err = templates["options"].ExecuteTemplate(w, "index.gotmpl", tdata)
			}
			if err != nil {
				slog.Error("template error", "template", "options", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			sourceResult, err := index.GetSource(ctx, sourceKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return
			}

			err = templates["search"].Execute(w, TemplateData{
				ExtraBodyHTML: config.Web.ExtraBodyHTML,
				Sources:       config.Importer.Sources,
				Source:        *source,
				SourceResult:  sourceResult,
				Version:       *versionInfo,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return
			}
		}
	})

	mux.Handle("/static/", http.FileServer(http.FS(frontend.Files)))

	if liveReload {
		applyDevModeOverrides(config)
		config.Web.ExtraBodyHTML = jsSnippet
		liveReload := livereload.New()
		liveReload.Start()
		top.Handle("/livereload", liveReload)
		fw, err := NewFileWatcher()
		if err != nil {
			return nil, errors.WithMessage(err, "could not create file watcher")
		}
		err = fw.AddRecursive(path.Join("frontend"))
		if err != nil {
			return nil, errors.WithMessage(err, "could not add directory to file watcher")
		}
		go fw.Start(func(filename string) {
			slog.Debug(fmt.Sprintf("got filename %s", filename))
			if path.Ext(filename) == ".gotmpl" {
				templates, err = loadTemplates()
				if err != nil {
					slog.Error(fmt.Sprintf("could not reload templates: %v", err))
				}
			}
			liveReload.Reload()
		})
	}

	var logWriter io.Writer
	if config.Web.Environment == "production" {
		logWriter = law.NewWriteAsyncer(os.Stdout, nil)
	} else {
		logWriter = os.Stdout
	}
	top.Handle("/",
		AddHeadersMiddleware(
			sentryHandler.Handle(
				wrapHandlerWithLogging(mux, wrappedHandlerOptions{
					defaultHostname: config.Web.BaseURL.Hostname(),
					logger:          logWriter,
				}),
			),
			config,
		),
	)
	// no logging, no sentry
	top.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return top, nil
}
