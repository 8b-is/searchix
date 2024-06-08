package index

import (
	"bytes"
	"context"
	"encoding/gob"
	"searchix/internal/config"
	"searchix/internal/nix"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/standard"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
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

func setField(
	q query.FieldableQuery,
	field string,
) query.FieldableQuery {
	q.SetField(field)

	return q
}

func (index *ReadIndex) search(
	ctx context.Context,
	request *bleve.SearchRequest,
) (*Result, error) {
	request.Fields = []string{"_data"}

	bleveResult, err := index.index.SearchInContext(ctx, request)
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

func (index *ReadIndex) Search(
	ctx context.Context,
	source *config.Source,
	keyword string,
	from uint64,
) (*Result, error) {
	query := bleve.NewBooleanQuery()

	// match the user's query in any field ...
	userQuery := bleve.NewMatchQuery(keyword)
	userQuery.Analyzer = "option_name"

	query.AddMust(
		setField(bleve.NewTermQuery(source.Key), "Source"),
		userQuery,
	)

	switch source.Importer {
	case config.Packages:
		// ...and boost it if it matches any of these
		query.AddShould(
			setField(bleve.NewMatchQuery(keyword), "MainProgram"),
			setField(bleve.NewMatchQuery(keyword), "Name"),
			setField(bleve.NewMatchQuery(keyword), "Attribute"),
		)
	case config.Options:
		query.AddShould(
			setField(bleve.NewMatchQuery(keyword), "Loc"),
			setField(bleve.NewMatchQuery(keyword), "Name"),
		)
		nameLiteralQuery := bleve.NewMatchQuery(keyword)
		nameLiteralQuery.SetField("Name")
		nameLiteralQuery.Analyzer = standard.Name
		query.AddShould(
			nameLiteralQuery,
		)
	default:
		return nil, errors.Errorf(
			"unsupported source type for search: %s",
			source.Importer.String(),
		)
	}

	search := bleve.NewSearchRequest(query)
	search.Size = ResultsPerPage

	if from != 0 {
		search.From = int(from)
	}

	return index.search(ctx, search)
}

func (index *ReadIndex) GetDocument(
	ctx context.Context,
	source *config.Source,
	id string,
) (*nix.Importable, error) {
	key := nix.MakeKey(source, id)
	query := bleve.NewDocIDQuery([]string{key})
	search := bleve.NewSearchRequest(query)

	result, err := index.search(ctx, search)
	if err != nil {
		return nil, err
	}

	return &result.Hits[0].Data, err
}
