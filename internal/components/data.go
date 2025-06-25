package components

import (
	"alin.ovh/searchix/frontend"
	"alin.ovh/searchix/internal/config"
	search "alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/nix"
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

func convertMatch[I nix.Importable](m nix.Importable) *I {
	i, ok := m.(I)
	if !ok {
		return nil
	}

	return &i
}
