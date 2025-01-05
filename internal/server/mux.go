package server

import (
	"context"
	"encoding/xml"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"go.alanpearce.eu/searchix/frontend"
	"go.alanpearce.eu/searchix/internal/components"
	"go.alanpearce.eu/searchix/internal/config"
	search "go.alanpearce.eu/searchix/internal/index"
	"go.alanpearce.eu/searchix/internal/opensearch"
	"go.alanpearce.eu/searchix/internal/pagination"
	"go.alanpearce.eu/x/log"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/osdevisnot/sorvor/pkg/livereload"
	"github.com/pkg/errors"
)

type HTTPError struct {
	Error   error
	Message string
	Code    int
}

var (
	sources []*config.Source
)

func applyDevModeOverrides(cfg *config.Config) {
	if len(cfg.Web.ContentSecurityPolicy.ScriptSrc) == 0 {
		cfg.Web.ContentSecurityPolicy.ScriptSrc = cfg.Web.ContentSecurityPolicy.DefaultSrc
	}
	cfg.Web.ContentSecurityPolicy.ScriptSrc = append(
		cfg.Web.ContentSecurityPolicy.ScriptSrc,
		"http://localhost:7331",
		"'unsafe-inline'",
	)
}

func sortSources(ss map[string]*config.Source) {
	sources = make([]*config.Source, len(ss))
	for _, v := range ss {
		sources[v.Order] = v
	}
}

func NewMux(
	cfg *config.Config,
	index *search.ReadIndex,
	log *log.Logger,
	liveReload bool,
) (*http.ServeMux, error) {
	if cfg == nil {
		return nil, errors.New("cfg is nil")
	}
	if index == nil {
		return nil, errors.New("index is nil")
	}
	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})
	sortSources(cfg.Importer.Sources)

	errorHandler := createErrorHandler(cfg, log)

	top := http.NewServeMux()
	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		indexData := components.TemplateData{
			ExtraHeadHTML: cfg.Web.ExtraHeadHTML,
			Sources:       sources,
			Assets:        frontend.Assets,
		}
		w.Header().Add("Cache-Control", "max-age=86400")
		err := components.Homepage(indexData).Render(r.Context(), w)
		if err != nil {
			errorHandler(w, r, err.Error(), http.StatusInternalServerError)
		}
	})

	const searchTimeout = 1 * time.Second
	createSearchHandler := func(importerType config.ImporterType) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			var err error
			var source *config.Source
			if importerType != config.All {
				source = cfg.Importer.Sources[r.PathValue("source")]
				if source == nil || importerType != source.Importer {
					errorHandler(w, r, http.StatusText(http.StatusNotFound), http.StatusNotFound)

					return
				}
			}

			ctx, cancel := context.WithTimeout(r.Context(), searchTimeout)
			defer cancel()

			if r.URL.Query().Has("query") {
				qs := r.URL.Query().Get("query")

				var pageSize int = search.DefaultPageSize
				var pageNumber = 1
				if pg := r.URL.Query().Get("page"); pg != "" {
					pageNumber, err = strconv.Atoi(pg)
					if err != nil || pageNumber > math.MaxInt {
						errorHandler(w, r, "Bad query string", http.StatusBadRequest)

						return
					}
					if pageNumber == 0 {
						pageNumber = 1
						pageSize = math.MaxInt
					}
				}
				page := pagination.New(pageNumber, pageSize)
				results, err := index.Search(ctx, source, qs, page.From, page.Size)
				if err != nil {
					if err == context.DeadlineExceeded {
						errorHandler(w, r, "Search timed out", http.StatusInternalServerError)

						return
					}
					log.Error("search error", "error", err)
					errorHandler(w, r, err.Error(), http.StatusInternalServerError)

					return
				}
				if pageSize == math.MaxInt && results.Total > config.MaxResultsShowAll {
					errorHandler(w, r, "Too many results, use pagination", http.StatusBadRequest)
				}

				tdata := components.ResultData{
					TemplateData: components.TemplateData{
						ExtraHeadHTML: cfg.Web.ExtraHeadHTML,
						Source:        source,
						Sources:       sources,
						Assets:        frontend.Assets,
						Query:         qs,
					},
					Query:   qs,
					Results: results,
				}

				page.SetResults(results.Total)
				if page.Needed {
					q, err := url.ParseQuery(r.URL.RawQuery)
					if err != nil {
						errorHandler(w, r, "Query string error", http.StatusBadRequest)

						return
					}

					if page.Next != 0 {
						q.Set("page", strconv.Itoa(page.Next))
						tdata.Next = "search?" + q.Encode()
					}

					if page.Prev != 0 {
						if page.Prev == 1 {
							q.Del("page")
						} else {
							q.Set("page", strconv.Itoa(page.Prev))
						}
						tdata.Prev = "search?" + q.Encode()
					}

					q.Set("page", "0")
					tdata.All = "search?" + q.Encode()
				}

				w.Header().Add("Cache-Control", "max-age=300")
				w.Header().Add("Vary", "Fetch")
				if r.Header.Get("Fetch") == "true" {
					w.Header().Add("Content-Type", "text/html; charset=utf-8")
					err = components.Results(tdata).Render(r.Context(), w)
				} else {
					err = components.ResultsPage(tdata).Render(r.Context(), w)
				}
				if err != nil {
					log.Error("template error", "template", importerType, "error", err)
					errorHandler(w, r, err.Error(), http.StatusInternalServerError)
				}
			} else {
				sourceResult, err := index.GetSource(ctx, source)
				if err != nil {
					errorHandler(w, r, err.Error(), http.StatusInternalServerError)

					return
				}
				w.Header().Add("Cache-Control", "max-age=14400")
				err = components.SearchPage(
					components.TemplateData{
						ExtraHeadHTML: cfg.Web.ExtraHeadHTML,
						Sources:       sources,
						Source:        source,
						SourceResult:  sourceResult,
						Assets:        frontend.Assets,
					},
					components.ResultData{},
				).Render(r.Context(), w)
				if err != nil {
					errorHandler(w, r, err.Error(), http.StatusInternalServerError)

					return
				}
			}
		}
	}

	mux.HandleFunc("/all/search", createSearchHandler(config.All))
	mux.HandleFunc("/options/{source}/search", createSearchHandler(config.Options))
	mux.HandleFunc("/packages/{source}/search", createSearchHandler(config.Packages))

	createSourceIDHandler := func(importerType config.ImporterType) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			source := cfg.Importer.Sources[r.PathValue("source")]
			if source == nil || source.Importer != importerType {
				errorHandler(w, r, http.StatusText(http.StatusNotFound), http.StatusNotFound)

				return
			}
			importerSingular := importerType.Singular()

			ctx, cancel := context.WithTimeout(r.Context(), searchTimeout)
			defer cancel()

			doc, err := index.GetDocument(ctx, source, r.PathValue("id"))
			if err != nil {
				errorHandler(
					w,
					r,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)

				return
			}

			if doc == nil {
				errorHandler(w, r, http.StatusText(http.StatusNotFound), http.StatusNotFound)

				return
			}

			tdata := components.TemplateData{
				ExtraHeadHTML: cfg.Web.ExtraHeadHTML,
				Source:        source,
				Sources:       sources,
				Assets:        frontend.Assets,
			}
			if r.Header.Get("Fetch") == "true" {
				w.Header().Add("Content-Type", "text/html; charset=utf-8")
				err = components.Detail(*doc).Render(r.Context(), w)
			} else {
				err = components.DetailPage(tdata, *doc).Render(r.Context(), w)
			}
			if err != nil {
				log.Error("template error", "template", importerSingular, "error", err)
				errorHandler(w, r, err.Error(), http.StatusInternalServerError)
			}
		}
	}
	mux.HandleFunc("/options/{source}/{id}", createSourceIDHandler(config.Options))
	mux.HandleFunc("/packages/{source}/{id}", createSourceIDHandler(config.Packages))
	mux.HandleFunc("/option/{source}/{id}", createSourceIDHandler(config.Options))
	mux.HandleFunc("/package/{source}/{id}", createSourceIDHandler(config.Packages))

	createOpenSearchXMLHandler := func(importerType config.ImporterType) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			source := cfg.Importer.Sources[r.PathValue("source")]
			if source == nil || importerType != source.Importer {
				errorHandler(w, r, http.StatusText(http.StatusNotFound), http.StatusNotFound)

				return
			}

			w.Header().Add("Cache-Control", "max-age=604800")
			w.Header().Set("Content-Type", "application/opensearchdescription+xml")
			osd := &opensearch.Description{
				ShortName:   fmt.Sprintf("Searchix %s", source),
				LongName:    fmt.Sprintf("Search %s with Searchix", source),
				Description: fmt.Sprintf("Search %s", source),
				SearchForm: cfg.Web.BaseURL.JoinPath(
					source.Importer.String(),
					source.Key,
					"search",
				),
				URL: opensearch.URL{
					Method: "get",
					Type:   "text/html",
					Template: cfg.Web.BaseURL.JoinPath(
						source.Importer.String(),
						source.Key,
						"search",
					).AddRawQuery("query", "{searchTerms}"),
				},
			}
			enc := xml.NewEncoder(w)
			enc.Indent("", "    ")
			err := enc.Encode(osd)
			if err != nil {
				// no errorHandler; HTML does not make sense here
				http.Error(
					w,
					fmt.Sprintf("OpenSearch XML encoding error: %v", err),
					http.StatusInternalServerError,
				)
			}
		}
	}

	mux.HandleFunc("/options/{source}/opensearch.xml", createOpenSearchXMLHandler(config.Options))
	mux.HandleFunc("/packages/{source}/opensearch.xml", createOpenSearchXMLHandler(config.Packages))
	mux.HandleFunc("/all/opensearch.xml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Cache-Control", "max-age=604800")
		w.Header().Set("Content-Type", "application/opensearchdescription+xml")
		osd := &opensearch.Description{
			ShortName:   "Searchix Combined",
			LongName:    "Search nix options and packages with Searchix",
			Description: "Search nix options and packages with Searchix",
			SearchForm:  cfg.Web.BaseURL.JoinPath("all/search"),
			URL: opensearch.URL{
				Method: "get",
				Type:   "text/html",
				Template: cfg.Web.BaseURL.JoinPath("all/search").
					AddRawQuery("query", "{searchTerms}"),
			},
		}
		enc := xml.NewEncoder(w)
		enc.Indent("", "    ")
		err := enc.Encode(osd)
		if err != nil {
			// no errorHandler; HTML does not make sense here
			http.Error(
				w,
				fmt.Sprintf("OpenSearch XML encoding error: %v", err),
				http.StatusInternalServerError,
			)
		}
	})

	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		asset, found := frontend.Assets.ByPath[r.URL.Path]
		if !found {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

			return
		}
		// optimisation for HTTP/3: first header sent as byte(41), not the string
		w.Header().Add("Cache-Control", "public, max-age=86400")
		w.Header().Add("Cache-Control", "stale-while-revalidate")
		http.ServeFileFS(w, r, frontend.Files, asset.Filename)
	})

	if liveReload {
		applyDevModeOverrides(cfg)
		cfg.Web.ExtraHeadHTML = livereload.JsSnippet
		liveReload := livereload.New()
		liveReload.Start()
		top.Handle("/livereload", liveReload)
		fw, err := NewFileWatcher(log.Named("watcher"))
		if err != nil {
			return nil, errors.WithMessage(err, "could not create file watcher")
		}
		err = fw.AddRecursive(path.Join("frontend"))
		if err != nil {
			return nil, errors.WithMessage(err, "could not add directory to file watcher")
		}
		go fw.Start(func(filename string) {
			log.Debug(fmt.Sprintf("got filename %s", filename))
			if match, _ := path.Match("frontend/static/*", filename); match {
				err := frontend.Rehash()
				if err != nil {
					log.Error("failed to re-hash frontend assets", "error", err)
				}
			}
			liveReload.Reload()
		})
	}

	top.Handle("/",
		AddHeadersMiddleware(
			sentryHandler.Handle(
				wrapHandlerWithLogging(mux, wrappedHandlerOptions{
					defaultHostname: cfg.Web.BaseURL.Hostname(),
					logger:          log,
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
