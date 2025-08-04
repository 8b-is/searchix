package components

import (
	"alin.ovh/searchix/frontend"
	"alin.ovh/searchix/internal/config"
	search "alin.ovh/searchix/internal/index"
)

type TemplateData struct {
	Sources       []*config.Source
	Source        *config.Source
	Query         string
	ExtraHeadHTML string
	Code          int
	Message       string
	Assets        *frontend.AssetCollection
}

type ResultData struct {
	TemplateData
	Query   string
	Results *search.Result
	Prev    string
	Next    string
	All     string
}
