package server

import (
	"cmp"
	"context"
	"encoding/xml"
	"fmt"
	"maps"
	"math"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"go.alanpearce.eu/searchix/frontend"
	"go.alanpearce.eu/searchix/internal/components"
	"go.alanpearce.eu/searchix/internal/config"
	search "go.alanpearce.eu/searchix/internal/index"
	"go.alanpearce.eu/searchix/internal/opensearch"
	"go.alanpearce.eu/searchix/internal/pagination"
	"go.alanpearce.eu/searchix/internal/sentryhttp"
	"go.alanpearce.eu/x/log"

	"github.com/osdevisnot/sorvor/pkg/livereload"
	"gitlab.com/tozd/go/errors"
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
		"'unsafe-inline'",
	)
}

func sortSources(ss map[string]*config.Source) {
	sources = slices.SortedFunc(maps.Values(ss), func(a, b *config.Source) int {
		return cmp.Or(
			cmp.Compare(a.Order, b.Order),
			strings.Compare(a.Key, b.Key),
			strings.Compare(a.Name, b.Name),
		)
	})
}

func NewMux(
	cfg *config.Config,
	index *search.ReadIndex,
	log *log.Logger,
	liveReload bool,
) (*http.ServeMux, errors.E) {
	if cfg == nil {
		return nil, errors.New("cfg is nil")
	}
	if index == nil {
		return nil, errors.New("index is nil")
	}
	sortSources(cfg.Importer.Sources)

	errorHandler := createErrorHandler(cfg, log)

	top := http.NewServeMux()
	mux := sentryhttp.NewServeMux()

	const searchTimeout = 1 * time.Second
	createSearchHandler := func(importerType config.ImporterType) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			var err error
			var source *config.Source
			var urlBase string
			if importerType != config.All {
				urlBase = "search"
				source = cfg.Importer.Sources[r.PathValue("source")]
				if source == nil || importerType != source.Importer {
					errorHandler(w, r, http.StatusText(http.StatusNotFound), http.StatusNotFound)

					return
				}
			}

			if r.URL.Query().Has("query") {
				qs := r.URL.Query().Get("query")

				if len(qs) < 2 {
					errorHandler(w, r, "Query too short", http.StatusBadRequest)

					return
				}

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
						pageSize = config.MaxResultsShowAll
					}
				}
				page := pagination.New(pageNumber, pageSize)

				ctx, cancel := context.WithTimeout(r.Context(), searchTimeout)
				results, err := index.Search(ctx, source, qs, page.From, page.Size)
				cancel()

				if err != nil {
					if err == context.DeadlineExceeded {
						errorHandler(w, r, "Search timed out", http.StatusInternalServerError)

						return
					}
					log.Error("search error", "error", err)
					errorHandler(w, r, err.Error(), http.StatusInternalServerError)

					return
				}
				if pageSize == config.MaxResultsShowAll &&
					results.Total > config.MaxResultsShowAll {
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
						tdata.Next = urlBase + "?" + q.Encode()
					}

					if page.Prev != 0 {
						if page.Prev == 1 {
							q.Del("page")
						} else {
							q.Set("page", strconv.Itoa(page.Prev))
						}
						tdata.Prev = urlBase + "?" + q.Encode()
					}

					q.Set("page", "0")
					tdata.All = urlBase + "?" + q.Encode()
				}

				w.Header().Add("Cache-Control", "max-age=300")
				w.Header().Add("Vary", "Fetch")
				var baseErr error
				if r.Header.Get("Fetch") == "true" {
					w.Header().Add("Content-Type", "text/html; charset=utf-8")
					baseErr = components.Results(tdata).Render(w)
				} else {
					baseErr = components.ResultsPage(tdata).Render(w)
				}
				if baseErr != nil {
					log.Error("template error", "template", importerType, "error", baseErr)
					errorHandler(w, r, baseErr.Error(), http.StatusInternalServerError)
				}
			} else {
				w.Header().Add("Cache-Control", "max-age=14400")
				err = components.SearchPage(
					components.TemplateData{
						ExtraHeadHTML: cfg.Web.ExtraHeadHTML,
						Sources:       sources,
						Source:        source,
						Assets:        frontend.Assets,
					},
					components.ResultData{},
				).Render(w)
				if err != nil {
					errorHandler(w, r, err.Error(), http.StatusInternalServerError)

					return
				}
			}
		}
	}

	mux.HandleFunc("/{$}", createSearchHandler(config.All))
	mux.HandleFunc("/options/{source}/search", createSearchHandler(config.Options))
	mux.HandleFunc("/packages/{source}/search", createSearchHandler(config.Packages))

	mux.Handle("/all/search", http.RedirectHandler("/", http.StatusFound))

	createSourceIDHandler := func(importerType config.ImporterType) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			source := cfg.Importer.Sources[r.PathValue("source")]
			if source == nil || source.Importer != importerType {
				errorHandler(w, r, http.StatusText(http.StatusNotFound), http.StatusNotFound)

				return
			}
			importerSingular := importerType.Singular()

			ctx, cancel := context.WithTimeout(r.Context(), searchTimeout)
			doc, err := index.GetDocument(ctx, source, r.PathValue("id"))
			cancel()

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
			var baseErr error
			if r.Header.Get("Fetch") == "true" {
				w.Header().Add("Content-Type", "text/html; charset=utf-8")
				baseErr = components.Detail(*doc).Render(w)
			} else {
				baseErr = components.DetailPage(tdata, *doc).Render(w)
			}
			if baseErr != nil {
				log.Error("template error", "template", importerSingular, "error", baseErr)
				errorHandler(w, r, baseErr.Error(), http.StatusInternalServerError)
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
	mux.Handle("/all/opensearch.xml", http.RedirectHandler("/opensearch.xml", http.StatusFound))
	mux.HandleFunc("/opensearch.xml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Cache-Control", "max-age=604800")
		w.Header().Set("Content-Type", "application/opensearchdescription+xml")
		osd := &opensearch.Description{
			ShortName:   "Searchix Combined",
			LongName:    "Search nix options and packages with Searchix",
			Description: "Search nix options and packages with Searchix",
			SearchForm:  cfg.Web.BaseURL.JoinPath(),
			URL: opensearch.URL{
				Method: "get",
				Type:   "text/html",
				Template: cfg.Web.BaseURL.JoinPath().
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
