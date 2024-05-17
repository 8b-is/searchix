package index

import (
	"bytes"
	"context"
	"encoding/gob"
	"searchix/internal/config"
	"searchix/internal/nix"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/pkg/errors"
)

const ResultsPerPage = 20

type DocumentMatch struct {
	search.DocumentMatch
	Data nix.Importable
}

type Result struct {
	*bleve.SearchResult
	Hits []DocumentMatch
}

type ReadIndex struct {
	index bleve.Index
	meta  *Meta
}

func (index *ReadIndex) GetEnabledSources() ([]string, error) {
	facet := bleve.NewFacetRequest("Source", 100)
	query := bleve.NewMatchAllQuery()
	search := bleve.NewSearchRequest(query)
	search.AddFacet("Source", facet)

	results, err := index.index.Search(search)
	if err != nil {
		return nil, errors.WithMessage(err, "could not get list of enabled sources from index")
	}

	enabledSources := make([]string, results.Facets["Source"].Terms.Len())
	for i, term := range results.Facets["Source"].Terms.Terms() {
		enabledSources[i] = term.Term
	}

	return enabledSources, nil
}

func (index *ReadIndex) GetSource(ctx context.Context, name string) (*bleve.SearchResult, error) {
	query := bleve.NewTermQuery(name)
	query.SetField("Source")
	search := bleve.NewSearchRequest(query)

	result, err := index.index.SearchInContext(ctx, search)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if err != nil {
			return nil, errors.WithMessagef(
				err,
				"failed to execute search to find source %s in index",
				name,
			)
		}
	}

	return result, nil
}

func (index *ReadIndex) Search(
	ctx context.Context,
	source *config.Source,
	keyword string,
	from uint64,
) (*Result, error) {
	sourceQuery := bleve.NewTermQuery(source.Key)
	sourceQuery.SetField("Source")
	userQuery := bleve.NewMatchQuery(keyword)
	userQuery.Analyzer = "option_name"

	query := bleve.NewConjunctionQuery(sourceQuery, userQuery)

	search := bleve.NewSearchRequest(query)
	search.Size = ResultsPerPage
	search.Fields = []string{"_data"}

	if from != 0 {
		search.From = int(from)
	}

	bleveResult, err := index.index.SearchInContext(ctx, search)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if err != nil {
			return nil, errors.WithMessage(err, "failed to execute search query")
		}

		results := make([]DocumentMatch, min(ResultsPerPage, bleveResult.Total))
		var buf bytes.Buffer
		for i, result := range bleveResult.Hits {
			_, err = buf.WriteString(result.Fields["_data"].(string))
			if err != nil {
				return nil, errors.WithMessage(err, "error fetching result data")
			}
			err = gob.NewDecoder(&buf).Decode(&results[i].Data)
			if err != nil {
				return nil, errors.WithMessagef(err, "error decoding gob data: %s", buf.String())
			}
			buf.Reset()
		}

		return &Result{
			SearchResult: bleveResult,
			Hits:         results,
		}, nil
	}
}
