package components

import (
	"go.alanpearce.eu/searchix/frontend"
	"go.alanpearce.eu/searchix/internal/config"
	search "go.alanpearce.eu/searchix/internal/index"
	"go.alanpearce.eu/searchix/internal/nix"

	"github.com/blevesearch/bleve/v2"
)

type TemplateData struct {
	Sources       []*config.Source
	Source        *config.Source
	Query         string
	SourceResult  *bleve.SearchResult
	ExtraHeadHTML string
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

type DocumentData struct {
	TemplateData
	Document *nix.Importable
	Children *search.Result
}
