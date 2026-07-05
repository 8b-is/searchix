package server

import (
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/Southclaws/fault/ftag"

	"alin.ovh/searchix/internal/components"
	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/opensearch"
)

type SourceHandler struct {
	global *GlobalHandler
	source config.Source
	mux    *http.ServeMux
}

func (h *SourceHandler) getTitle(doc index.Indexable) string {
	title := ""
	if doc != nil {
		title = doc.GetName() + " &hyphen; "
	}

	title += h.source.Name + " &hyphen; "

	return title
}

func (h *SourceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *SourceHandler) searchSource(r *http.Request) config.Source {
	if r.URL.Query().Has("scoped") {
		return h.source
	}

	return config.Source{Importer: config.All}
}

func (h *SourceHandler) closeURL(r *http.Request) string {
	q := r.URL.Query()
	scoped := q.Has("scoped")
	q.Del("scoped")

	if scoped {
		return "/" + h.source.Importer.String() + "/" + h.source.Key + "/search?" + q.Encode()
	}

	return "/?" + q.Encode()
}

func (h *SourceHandler) Search(w http.ResponseWriter, r *http.Request) {
	h.global.Search(h.source, w, r)
}

func (h *SourceHandler) Detail(w http.ResponseWriter, r *http.Request) {
	doc, err := h.global.index.GetDocument(r.Context(), h.source, r.PathValue("id"))
	if err != nil {
		if ftag.Get(err) == ftag.NotFound {
			h.global.log.Warn("document not found", "source", h.source.Key, "id", r.PathValue("id"))
			h.global.errorHandler(w, r, http.StatusText(http.StatusNotFound), http.StatusNotFound)

			return
		}

		h.global.log.Error(
			"failed to get document",
			"source",
			h.source.Key,
			"id",
			r.PathValue("id"),
			"error",
			err,
		)
		h.global.errorHandler(
			w,
			r,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	if doc == nil {
		h.global.errorHandler(w, r, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	var baseErr error
	if r.Header.Get("Fetch") == "true" {
		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		w.Header().Add("Title", h.getTitle(doc))
		baseErr = components.Detail(doc).Render(w)
	} else if searchResult, _, searchErr := h.global.ExecuteSearch(h.searchSource(r), r); searchErr == nil &&
		searchResult != nil {
		searchTitle := searchResult.Title
		searchResult.Title = h.getTitle(doc)
		searchResult.CanonicalURL = "/" + index.GetKey(doc)
		searchResult.Description = components.MetaDescription(doc)
		searchResult.DetailContent = components.Detail(doc)
		searchResult.SearchURL = h.closeURL(r)
		searchResult.SearchTitle = searchTitle
		baseErr = components.DetailResultsPage(*searchResult).Render(w)
	} else {
		tdata := components.TemplateData{
			ExtraHeadHTML: h.global.cfg.Web.ExtraHeadHTML,
			Source:        h.source,
			Sources:       sources,
			Assets:        h.global.assets,
			Title:         h.getTitle(doc),
			Description:   components.MetaDescription(doc),
			BaseURL:       h.global.cfg.Web.BaseURL,
		}
		baseErr = components.DetailPage(tdata, doc).Render(w)
	}
	if baseErr != nil {
		h.global.log.Error("template error", "template", h.source.Importer.Singular(), "error", baseErr)
		h.global.errorHandler(w, r, baseErr.Error(), http.StatusInternalServerError)
	}
}

func (h *SourceHandler) OpenSearchXML(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Cache-Control", "max-age=604800")
	w.Header().Set("Content-Type", "application/opensearchdescription+xml")
	osd := &opensearch.Description{
		ShortName:   fmt.Sprintf("Searchix %s", h.source),
		LongName:    fmt.Sprintf("Search %s with Searchix", h.source),
		Description: fmt.Sprintf("Search %s", h.source),
		SearchForm: h.global.cfg.Web.BaseURL.JoinPath(
			h.source.Importer.String(),
			h.source.Key,
			"search",
		),
		Image: opensearch.Image{
			Height:  32,
			Width:   32,
			Type:    "image/x-icon",
			Content: h.global.cfg.Web.BaseURL.JoinPath("favicon.ico").String(),
		},
		URL: opensearch.URL{
			Method: "get",
			Type:   "text/html",
			Template: h.global.cfg.Web.BaseURL.JoinPath(
				h.source.Importer.String(),
				h.source.Key,
				"search",
			).AddRawQuery("query", "{searchTerms}"),
		},
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "    ")
	err := enc.Encode(osd)
	if err != nil {
		// no h.global.errorHandler; HTML does not make sense here
		http.Error(
			w,
			fmt.Sprintf("OpenSearch XML encoding error: %v", err),
			http.StatusInternalServerError,
		)
	}
}
