package server

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	"searchix/frontend"
	"searchix/internal/config"
	search "searchix/internal/index"

	"github.com/blevesearch/bleve/v2"
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

type TemplateData struct {
	Sources       map[string]*config.Source
	Source        config.Source
	Query         string
	Results       bool
	SourceResult  *bleve.SearchResult
	ExtraHeadHTML template.HTML
	Code          int
	Message       string
	Assets        *frontend.AssetCollection
}

type ResultData struct {
	TemplateData
	Query          string
	ResultsPerPage int
	Results        *search.Result
	Prev           string
	Next           string
}

var templates TemplateCollection

func applyDevModeOverrides(cfg *config.Config) {
	if len(cfg.Web.ContentSecurityPolicy.ScriptSrc) == 0 {
		cfg.Web.ContentSecurityPolicy.ScriptSrc = cfg.Web.ContentSecurityPolicy.DefaultSrc
	}
	cfg.Web.ContentSecurityPolicy.ScriptSrc = append(
		cfg.Web.ContentSecurityPolicy.ScriptSrc,
		"'unsafe-inline'",
	)
}

func NewMux(
	cfg *config.Config,
	index *search.ReadIndex,
	liveReload bool,
) (*http.ServeMux, error) {
	var err error
	if cfg == nil {
		return nil, errors.New("cfg is nil")
	}
	if index == nil {
		return nil, errors.New("index is nil")
	}
	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})
	templates, err = loadTemplates()
	if err != nil {
		log.Panicf("could not load templates: %v", err)
	}

	errorHandler := createErrorHandler(cfg)

	top := http.NewServeMux()
	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		indexData := TemplateData{
			ExtraHeadHTML: cfg.Web.ExtraHeadHTML,
			Sources:       cfg.Importer.Sources,
			Assets:        frontend.Assets,
		}
		w.Header().Add("Cache-Control", "max-age=86400")
		err := templates["index"].Execute(w, indexData)
		if err != nil {
			errorHandler(w, r, err.Error(), http.StatusInternalServerError)
		}
	})

	const searchTimeout = 1 * time.Second
	createSearchHandler := func(importerType config.ImporterType) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			var err error
			source := cfg.Importer.Sources[r.PathValue("source")]
			if source == nil || importerType != source.Importer {
				errorHandler(w, r, http.StatusText(http.StatusNotFound), http.StatusNotFound)

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
						errorHandler(w, r, "Bad query string", http.StatusBadRequest)
					}
				}
				results, err := index.Search(ctx, source, qs, (page-1)*search.ResultsPerPage)
				if err != nil {
					if err == context.DeadlineExceeded {
						errorHandler(w, r, "Search timed out", http.StatusInternalServerError)

						return
					}
					slog.Error("search error", "error", err)
					errorHandler(w, r, err.Error(), http.StatusInternalServerError)
				}

				tdata := ResultData{
					TemplateData: TemplateData{
						ExtraHeadHTML: cfg.Web.ExtraHeadHTML,
						Source:        *source,
						Sources:       cfg.Importer.Sources,
						Assets:        frontend.Assets,
					},
					ResultsPerPage: search.ResultsPerPage,
					Query:          qs,
					Results:        results,
				}

				hits := uint64(len(results.Hits))
				if results.Total > hits {
					q, err := url.ParseQuery(r.URL.RawQuery)
					if err != nil {
						errorHandler(w, r, "Query string error", http.StatusBadRequest)

						return
					}

					if page > uint64(math.Ceil(float64(results.Total)/search.ResultsPerPage)) {
						errorHandler(w, r, "Not found", http.StatusNotFound)

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

				w.Header().Add("Cache-Control", "max-age=300")
				w.Header().Add("Vary", "Fetch")
				if r.Header.Get("Fetch") == "true" {
					w.Header().Add("Content-Type", "text/html; charset=utf-8")
					err = templates[importerType.String()].ExecuteTemplate(w, "results", tdata)
				} else {
					err = templates[importerType.String()].Execute(w, tdata)
				}
				if err != nil {
					slog.Error("template error", "template", importerType, "error", err)
					errorHandler(w, r, err.Error(), http.StatusInternalServerError)
				}
			} else {
				sourceResult, err := index.GetSource(ctx, source.Key)
				if err != nil {
					errorHandler(w, r, err.Error(), http.StatusInternalServerError)

					return
				}

				w.Header().Add("Cache-Control", "max-age=14400")
				err = templates["search"].Execute(w, TemplateData{
					ExtraHeadHTML: cfg.Web.ExtraHeadHTML,
					Sources:       cfg.Importer.Sources,
					Source:        *source,
					SourceResult:  sourceResult,
					Assets:        frontend.Assets,
				})
				if err != nil {
					errorHandler(w, r, err.Error(), http.StatusInternalServerError)

					return
				}
			}
		}
	}

	mux.HandleFunc("/options/{source}/search", createSearchHandler(config.Options))
	mux.HandleFunc("/packages/{source}/search", createSearchHandler(config.Packages))

	createOpenSearchXMLHandler := func(importerType config.ImporterType) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			type openSearchData struct {
				BaseURL string
				Source  *config.Source
			}

			source := cfg.Importer.Sources[r.PathValue("source")]
			if source == nil || importerType != source.Importer {
				errorHandler(w, r, http.StatusText(http.StatusNotFound), http.StatusNotFound)

				return
			}

			w.Header().Add("Cache-Control", "max-age=604800")
			w.Header().Set("Content-Type", "application/opensearchdescription+xml")
			err := templates["opensearch.xml"].ExecuteTemplate(
				w,
				"opensearch.xml.gotmpl",
				openSearchData{
					BaseURL: cfg.Web.BaseURL.String(),
					Source:  source,
				},
			)
			if err != nil {
				// no errorHandler; HTML does not make sense here
				http.Error(
					w,
					fmt.Sprintf("Template render error: %v", err),
					http.StatusInternalServerError,
				)
			}
		}
	}

	mux.HandleFunc("/options/{source}/opensearch.xml", createOpenSearchXMLHandler(config.Options))
	mux.HandleFunc("/packages/{source}/opensearch.xml", createOpenSearchXMLHandler(config.Packages))

	fs := http.FileServer(http.FS(frontend.Files))
	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "max-age=86400")
		fs.ServeHTTP(w, r)
	})

	if liveReload {
		applyDevModeOverrides(cfg)
		cfg.Web.ExtraHeadHTML = jsSnippet
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
			if match, _ := path.Match("frontend/static/*", filename); match {
				err := frontend.Rehash()
				if err != nil {
					slog.Error("failed to re-hash frontend assets", "error", err)
				}
			}
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
	if cfg.Web.Environment == "production" {
		logWriter = law.NewWriteAsyncer(os.Stdout, nil)
	} else {
		logWriter = os.Stdout
	}
	top.Handle("/",
		AddHeadersMiddleware(
			sentryHandler.Handle(
				wrapHandlerWithLogging(mux, wrappedHandlerOptions{
					defaultHostname: cfg.Web.BaseURL.Hostname(),
					logger:          logWriter,
				}),
			),
			cfg,
		),
	)
	// no logging, no sentry
	top.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return top, nil
}
