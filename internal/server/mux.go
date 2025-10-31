package server

import (
	"cmp"
	"context"
	"encoding/xml"
	"fmt"
	"maps"
	"math"
	"net/http"
	"path"
	"slices"
	"strconv"
	"strings"

	"alin.ovh/searchix/frontend"
	"alin.ovh/searchix/internal/components"
	"alin.ovh/searchix/internal/config"
	search "alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/opensearch"
	"alin.ovh/searchix/internal/pagination"
	"alin.ovh/searchix/internal/sentryhttp"

	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	"github.com/osdevisnot/sorvor/pkg/livereload"
)

type HTTPError struct {
	Error   error
	Message string
	Code    int
}

var sources []*config.Source

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
	index := options.ReadIndex
	sortSources(cfg.Importer.Sources)
	assets, err := frontend.New()
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("could not create frontend asset collection"))
	}

	errorHandler := createErrorHandler(cfg, assets, log)

	top := http.NewServeMux()
	mux := sentryhttp.NewServeMux()

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

			facets := r.URL.Query()
			facets.Del("query")
			facets.Del("page")

			if r.URL.Query().Has("query") || len(facets) > 0 {
				qs := r.URL.Query().Get("query")
				ctx := r.Context()
				if qs != "" {
					ctx = fctx.WithMeta(ctx, "search_query", qs)
				}

				if len(qs) < 2 && len(facets) == 0 {
					errorHandler(w, r, "Query too short", http.StatusBadRequest)

					return
				}

				pageSize := search.DefaultPageSize
				pageNumber := 1
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

				ctx, cancel := context.WithTimeout(ctx, cfg.Web.SearchTimeout.Duration)
				results, err := index.Search(ctx, source, qs, page.From, page.Size, facets)
				cancel()

				if err != nil {
					if err == context.DeadlineExceeded {
						errorHandler(w, r, "Search timed out", http.StatusInternalServerError)

						return
					}
					meta := fctx.Unwrap(err)
					log.Error("search error", "error", err, "search_query", meta["search_query"])
					errorHandler(w, r, err.Error(), http.StatusInternalServerError)

					return
				}
				if pageSize == config.MaxResultsShowAll &&
					results.Total > config.MaxResultsShowAll {
					errorHandler(w, r, "Too many results, use pagination", http.StatusBadRequest)
				}
				page.SetResults(results.Total)

				tdata := components.ResultData{
					TemplateData: components.TemplateData{
						ExtraHeadHTML: cfg.Web.ExtraHeadHTML,
						Source:        source,
						Sources:       sources,
						Assets:        assets,
						Query:         qs,
						SearchNav:     components.NewSearchNav(*r.URL).WithPagination(page),
					},
					Query:        qs,
					FacetQueries: facets,
					Results:      results,
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
						Assets:        assets,
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

			doc, err := index.GetDocument(r.Context(), source, r.PathValue("id"))
			if err != nil {
				if ftag.Get(err) == ftag.NotFound {
					log.Warn("document not found", "source", source.Key, "id", r.PathValue("id"))
					errorHandler(w, r, http.StatusText(http.StatusNotFound), http.StatusNotFound)

					return
				}

				log.Error(
					"failed to get document",
					"source",
					source.Key,
					"id",
					r.PathValue("id"),
					"error",
					err,
				)
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
				Assets:        assets,
			}
			var baseErr error
			if r.Header.Get("Fetch") == "true" {
				w.Header().Add("Content-Type", "text/html; charset=utf-8")
				baseErr = components.Detail(doc).Render(w)
			} else {
				baseErr = components.DetailPage(tdata, doc).Render(w)
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
				Image: opensearch.Image{
					Height:  32,
					Width:   32,
					Type:    "image/x-icon",
					Content: cfg.Web.BaseURL.JoinPath("favicon.ico").String(),
				},
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
			Image: opensearch.Image{
				Height:  32,
				Width:   32,
				Type:    "image/x-icon",
				Content: cfg.Web.BaseURL.JoinPath("favicon.ico").String(),
			},
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

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		asset, found := assets.ByPath[r.URL.Path]
		if !found {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

			return
		}
		// optimisation for HTTP/3: first header sent as byte(41), not the string
		w.Header().Add("Cache-Control", "public, max-age=86400")
		http.ServeFileFS(w, r, frontend.Files, asset.Filename)
	})

	mdb := options.ManpagesURLMap
	if err := mdb.Open(); err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to open manpages URL map"))
	}
	mux.HandleFunc("/man/{section}/{page}", func(w http.ResponseWriter, r *http.Request) {
		section := r.PathValue("section")
		page := r.PathValue("page")

		url, ok := mdb.Get(section, page)
		if !ok {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

			return
		}

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})

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
			log.Debug(fmt.Sprintf("got filename %s", filename))
			if match, _ := path.Match("static/*", filename); match {
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
