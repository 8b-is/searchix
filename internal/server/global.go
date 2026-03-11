package server

import (
	"context"
	"encoding/xml"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"alin.ovh/x/log"
	"github.com/Southclaws/fault/fctx"

	"alin.ovh/searchix/frontend"
	"alin.ovh/searchix/internal/components"
	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/manpages"
	"alin.ovh/searchix/internal/opensearch"
	"alin.ovh/searchix/internal/pagination"
)

type GlobalHandler struct {
	assets *frontend.AssetCollection
	cfg    *config.Config
	index  *index.ReadIndex
	log    *log.Logger
	mdb    *manpages.URLMap

	errorHandler func(w http.ResponseWriter, r *http.Request, message string, statusCode int)
}

func (g *GlobalHandler) getTitle(source config.Source, query string) (title string) {
	if query != "" {
		title = query + " &hyphen; "
	}
	if source.Importer != config.All {
		title += source.Name + " &hyphen; "
	}

	return
}

func (g *GlobalHandler) NewSourceHandler(source config.Source) *SourceHandler {
	h := &SourceHandler{
		global: g,
		source: source,
		mux:    http.NewServeMux(),
	}

	// JoinPath would escape braces, breaking parameters
	h.mux.HandleFunc(source.LocalURL().Path+"search", h.Search)
	h.mux.HandleFunc(source.LocalURL().Path+"{id}", h.Detail)
	h.mux.HandleFunc(source.LocalURL().Path+"opensearch.xml", h.OpenSearchXML)

	return h
}

func (g *GlobalHandler) ExecuteSearch(
	source config.Source,
	r *http.Request,
) (*components.ResultData, int, error) {
	facets := r.URL.Query()
	facets.Del("query")
	facets.Del("page")
	facets.Del("scoped")

	if !r.URL.Query().Has("query") && len(facets) == 0 {
		return nil, 0, nil
	}

	qs := r.URL.Query().Get("query")
	ctx := r.Context()
	if qs != "" {
		ctx = fctx.WithMeta(ctx, "search_query", qs)
	}

	if len(qs) < 2 && len(facets) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("query too short")
	}

	pageSize := index.DefaultPageSize
	pageNumber := 1
	if pg := r.URL.Query().Get("page"); pg != "" {
		var err error
		pageNumber, err = strconv.Atoi(pg)
		if err != nil || pageNumber > math.MaxInt {
			return nil, http.StatusBadRequest, fmt.Errorf("bad query string")
		}
		if pageNumber == 0 {
			pageNumber = 1
			pageSize = config.MaxResultsShowAll
		}
	}
	page := pagination.New(pageNumber, pageSize)

	ctx, cancel := context.WithTimeout(ctx, g.cfg.Web.SearchTimeout.Duration)
	results, err := g.index.Search(ctx, index.SearchRequest{
		Source:   source,
		Keyword:  qs,
		From:     page.From,
		PageSize: page.Size,
		Facets:   facets,
	})
	cancel()

	if err != nil {
		if err == context.DeadlineExceeded {
			return nil, http.StatusInternalServerError, fmt.Errorf("Search timed out")
		}
		meta := fctx.Unwrap(err)
		g.log.Error("search error", "error", err, "search_query", meta["search_query"])

		return nil, http.StatusInternalServerError, fmt.Errorf("search error: %v", err)
	}
	if pageSize == config.MaxResultsShowAll &&
		results.Total > config.MaxResultsShowAll {
		return nil, http.StatusBadRequest, fmt.Errorf("too many results, use pagination")
	}
	page.SetResults(results.Total)

	searchDesc := fmt.Sprintf("%d results for %s", results.Total, qs)
	if source.Importer != config.All {
		searchDesc += " in " + source.Name
	}

	searchURL := *source.LocalURL()
	searchURL.Path += "search"
	searchQuery := r.URL.Query()
	searchQuery.Del("scoped")
	searchURL.RawQuery = searchQuery.Encode()

	tdata := &components.ResultData{
		TemplateData: components.TemplateData{
			ExtraHeadHTML: g.cfg.Web.ExtraHeadHTML,
			Source:        source,
			Sources:       sources,
			Assets:        g.assets,
			Query:         qs,
			Title:         g.getTitle(source, qs),
			Description:   searchDesc,
			BaseURL:       g.cfg.Web.BaseURL,
			SearchNav:     components.NewSearchNav(searchURL).WithPagination(page),
		},
		Query:        qs,
		FacetQueries: facets,
		Results:      results,
	}

	return tdata, 0, nil
}

func (g *GlobalHandler) Search(source config.Source, w http.ResponseWriter, r *http.Request) {
	tdata, statusCode, err := g.ExecuteSearch(source, r)
	if err != nil {
		g.errorHandler(w, r, err.Error(), statusCode)

		return
	}

	if tdata != nil {
		w.Header().Add("Cache-Control", "max-age=300")
		w.Header().Add("Vary", "Fetch")
		var baseErr error
		if r.Header.Get("Fetch") == "true" {
			w.Header().Add("Content-Type", "text/html; charset=utf-8")
			w.Header().Add("Title", g.getTitle(source, tdata.Query))
			baseErr = components.Results(*tdata).Render(w)
		} else {
			baseErr = components.ResultsPage(*tdata).Render(w)
		}
		if baseErr != nil {
			g.log.Error("template error", "template", source.Importer, "error", baseErr)
			g.errorHandler(w, r, baseErr.Error(), http.StatusInternalServerError)
		}
	} else {
		w.Header().Add("Cache-Control", "max-age=14400")
		err = components.SearchPage(
			components.TemplateData{
				ExtraHeadHTML: g.cfg.Web.ExtraHeadHTML,
				Sources:       sources,
				Source:        source,
				Title:         g.getTitle(source, ""),
				Description:   "Search Nix packages and options",
				BaseURL:       g.cfg.Web.BaseURL,
				Assets:        g.assets,
			},
			components.ResultData{},
		).Render(w)
		if err != nil {
			g.errorHandler(w, r, err.Error(), http.StatusInternalServerError)

			return
		}
	}
}

func (g *GlobalHandler) CombinedSearch(w http.ResponseWriter, r *http.Request) {
	g.Search(config.Source{
		Importer: config.All,
	}, w, r)
}

func (g *GlobalHandler) RootOpenSearch(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Cache-Control", "max-age=604800")
	w.Header().Set("Content-Type", "application/opensearchdescription+xml")
	osd := &opensearch.Description{
		ShortName:   "Searchix Combined",
		LongName:    "Search nix options and packages with Searchix",
		Description: "Search nix options and packages with Searchix",
		SearchForm:  g.cfg.Web.BaseURL.JoinPath(),
		Image: opensearch.Image{
			Height:  32,
			Width:   32,
			Type:    "image/x-icon",
			Content: g.cfg.Web.BaseURL.JoinPath("favicon.ico").String(),
		},
		URL: opensearch.URL{
			Method: "get",
			Type:   "text/html",
			Template: g.cfg.Web.BaseURL.JoinPath().
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
}

func (g *GlobalHandler) StaticFile(w http.ResponseWriter, r *http.Request) {
	asset, found := g.assets.ByPath[r.URL.Path]
	if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}
	// optimisation for HTTP/3: first header sent as byte(38), not the string
	// see https://datatracker.ietf.org/doc/html/rfc9204#appendix-A for values
	w.Header().Add("Cache-Control", "max-age=604800")
	w.Header().Add("ETag", asset.ETag)
	http.ServeFileFS(w, r, frontend.Files, asset.Filename)
}

func (g *GlobalHandler) Asset(w http.ResponseWriter, r *http.Request) {
	asset, found := g.assets.ByImmutablePath[r.URL.Path]
	if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}
	// optimisation for HTTP/3: first header sent as byte(41), not the string
	// see https://datatracker.ietf.org/doc/html/rfc9204#appendix-A for values
	w.Header().Add("Cache-Control", "public, max-age=31536000")
	w.Header().Add("Cache-Control", "immutable")
	http.ServeFileFS(w, r, frontend.Files, asset.Filename)
}

func (g *GlobalHandler) ManualPage(w http.ResponseWriter, r *http.Request) {
	section := r.PathValue("section")
	page := r.PathValue("page")

	url, ok := g.mdb.Get(section, page)
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
