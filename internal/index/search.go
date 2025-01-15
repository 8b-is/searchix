package index

import (
	"bytes"
	"context"
	"encoding/gob"

	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/nix"
	"go.alanpearce.eu/x/log"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/pkg/errors"
)

const DefaultPageSize = 100

type DocumentMatch struct {
	*search.DocumentMatch
	Data nix.Importable
}

type Result struct {
	*bleve.SearchResult
	Hits []DocumentMatch
}

type ReadIndex struct {
	index bleve.Index
	log   *log.Logger
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

func (index *ReadIndex) GetSource(
	ctx context.Context,
	source *config.Source,
) (*bleve.SearchResult, error) {
	var query query.Query
	if source == nil {
		query = bleve.NewMatchAllQuery()
	} else {
		tq := bleve.NewTermQuery(source.Name)
		tq.SetField("Source")
		query = tq
	}
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
				source,
			)
		}
	}

	return result, nil
}

func setField[T query.FieldableQuery](
	q T,
	field string,
) T {
	q.SetField(field)

	return q
}

func (index *ReadIndex) search(
	ctx context.Context,
	request *bleve.SearchRequest,
) (*Result, error) {
	request.Fields = []string{"_data", "Source"}

	bleveResult, err := index.index.SearchInContext(ctx, request)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if err != nil {
			return nil, errors.WithMessage(err, "failed to execute search query")
		}

		results := make([]DocumentMatch, bleveResult.Hits.Len())
		var buf bytes.Buffer
		for i, result := range bleveResult.Hits {
			results[i].DocumentMatch = bleveResult.Hits[i]
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
	from int,
	pageSize int,
) (*Result, error) {
	query := bleve.NewBooleanQuery()

	// match the user's query in any field ...
	userQuery := bleve.NewMatchQuery(keyword)
	userQuery.Analyzer = "option_name"

	userQueryLoc := bleve.NewMatchQuery(keyword)
	userQueryLoc.Analyzer = "loc"
	userQueryLoc.SetBoost(2)

	query.AddMust(bleve.NewDisjunctionQuery(
		userQuery,
		userQueryLoc,
	))

	if source != nil {
		query.AddMust(
			setField(bleve.NewTermQuery(source.Key), "Source"),
		)
	} else {
		q := bleve.NewTermQuery("nixpkgs")
		q.SetField("Source")
		q.SetBoost(-150)

		query.AddShould(q)
	}

	programsQuery := bleve.NewMatchQuery(keyword)
	programsQuery.SetField("Programs")
	programsQuery.Analyzer = "keyword_single"
	query.AddShould(programsQuery)

	attrQuery := bleve.NewMatchQuery(keyword)
	attrQuery.SetField("Attribute")
	attrQuery.Analyzer = "keyword_single"
	query.AddShould(attrQuery)

	nameLiteralQuery := bleve.NewMatchQuery(keyword)
	nameLiteralQuery.SetField("Name")
	nameLiteralQuery.Analyzer = "keyword_single"
	query.AddShould(nameLiteralQuery)

	search := bleve.NewSearchRequest(query)
	search.Explain = config.DevMode
	search.Size = pageSize

	if from != 0 {
		search.From = from
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

	if result.Total == 0 {
		return nil, nil
	}

	return &result.Hits[0].Data, err
}
