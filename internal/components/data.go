package components

import (
	"time"

	"go.alanpearce.eu/searchix/frontend"
	"go.alanpearce.eu/searchix/internal/config"
	search "go.alanpearce.eu/searchix/internal/index"
	"go.alanpearce.eu/searchix/internal/nix"
)

var Indexing struct {
	InProgress bool
	StartedAt  time.Time
	FinishedAt time.Time
	NextRun    time.Time
}

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

func SetNextRun(nextRun time.Time) {
	Indexing.NextRun = nextRun
}

func SetLastUpdated(last time.Time) {
	Indexing.FinishedAt = last
}

func MarkIndexingStarted() {
	Indexing.StartedAt = time.Now()
	Indexing.InProgress = true
}

func MarkIndexingFinished(nextRun time.Time) {
	Indexing.FinishedAt = time.Now()
	Indexing.InProgress = false
	Indexing.NextRun = nextRun
}
