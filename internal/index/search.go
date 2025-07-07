package index

import (
	"bytes"
	"context"
	"encoding/gob"
	"iter"
	"time"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/nix"
	"alin.ovh/x/log"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
	"gitlab.com/tozd/go/errors"
)

const DefaultPageSize = 100

type DocumentMatch struct {
	*search.DocumentMatch
	Data nix.Importable
}

type Result struct {
	*bleve.SearchResult
	Hits iter.Seq[DocumentMatch]
}

type ReadIndex struct {
	index bleve.Index
	log   *log.Logger
	meta  *Meta
}

func (index *ReadIndex) LastUpdated() time.Time {
	return index.meta.LastUpdated()
}

func (index *ReadIndex) GetEnabledSources() ([]string, errors.E) {
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

func setField[T query.FieldableQuery](
	q T,
	field string,
) T {
	q.SetField(field)

	return q
}

func setBoost[T query.BoostableQuery](q T, boost float64) T {
	q.SetBoost(boost)

	return q
}

func (index *ReadIndex) search(
	ctx context.Context,
	request *bleve.SearchRequest,
) (*Result, errors.E) {
	request.Fields = []string{"_data", "Source"}

	bleveResult, err := index.index.SearchInContext(ctx, request)
	select {
	case <-ctx.Done():
		return nil, errors.WithStack(ctx.Err())
	default:
		if err != nil {
			return nil, errors.WithMessage(err, "failed to execute search query")
		}

		hits := func(yield func(DocumentMatch) bool) {
			var buf bytes.Buffer
			for _, match := range bleveResult.Hits {
				hit := DocumentMatch{
					DocumentMatch: match,
					Data:          nil,
				}
				_, err := buf.WriteString(match.Fields["_data"].(string))
				if err != nil {
					index.log.Warn("error fetching result data", "error", err)
				}
				err = gob.NewDecoder(&buf).Decode(&hit.Data)
				if err != nil {
					index.log.Warn("error decoding gob data", "error", err, "data", buf.String())
				}
				buf.Reset()
				if !yield(hit) {
					return
				}
			}
		}

		return &Result{
			SearchResult: bleveResult,
			Hits:         hits,
		}, nil
	}
}

func (index *ReadIndex) Search(
	ctx context.Context,
	source *config.Source,
	keyword string,
	from int,
	pageSize int,
) (*Result, errors.E) {
	query := bleve.NewBooleanQuery()

	// match the user's query in any field ...
	query.AddMust(bleve.NewDisjunctionQuery(
		setBoost(bleve.NewTermQuery(keyword), 50),
		setBoost(bleve.NewPrefixQuery(keyword), 25),
		setBoost(bleve.NewMatchPhraseQuery(keyword), 25),
		bleve.NewMatchQuery(keyword),
	))

	if source != nil {
		query.AddMust(
			setField(bleve.NewTermQuery(source.Key), "Source"),
		)
	} else {
		q := bleve.NewDisjunctionQuery(
			setBoost(setField(bleve.NewTermQuery("nixpkgs"), "Source"), -1000),
			setBoost(setField(bleve.NewTermQuery("nur"), "Source"), -5000),
		)

		query.AddShould(q)
	}

	idQuery := bleve.NewMatchQuery(keyword)
	idQuery.SetField("_id")
	idQuery.SetBoost(8)
	query.AddShould(idQuery)

	mainProgramQuery := bleve.NewMatchQuery(keyword)
	mainProgramQuery.SetField("MainProgram")
	mainProgramQuery.SetBoost(10)
	query.AddShould(mainProgramQuery)

	mainProgramLiteralQuery := bleve.NewTermQuery(keyword)
	mainProgramLiteralQuery.SetField("MainProgram")
	mainProgramLiteralQuery.SetBoost(50)
	query.AddShould(mainProgramLiteralQuery)

	programsQuery := bleve.NewMatchQuery(keyword)
	programsQuery.SetField("Programs")
	programsQuery.SetBoost(2)
	query.AddShould(programsQuery)

	programsLiteralQuery := bleve.NewTermQuery(keyword)
	programsLiteralQuery.SetField("Programs")
	programsLiteralQuery.SetBoost(100)
	query.AddShould(programsLiteralQuery)

	attrQuery := bleve.NewMatchQuery(keyword)
	attrQuery.SetField("Attribute")
	attrQuery.SetBoost(2)
	query.AddShould(attrQuery)

	attrPrefixQuery := bleve.NewPrefixQuery(keyword)
	attrPrefixQuery.SetField("Attribute")
	attrPrefixQuery.SetBoost(10)
	query.AddShould(attrPrefixQuery)

	attrTermQuery := bleve.NewTermQuery(keyword)
	attrTermQuery.SetField("Attribute")
	attrTermQuery.SetBoost(600)
	query.AddShould(attrTermQuery)

	nameQuery := bleve.NewMatchQuery(keyword)
	nameQuery.SetField("Name")
	nameQuery.SetBoost(2)
	query.AddShould(nameQuery)

	nameNGramQuery := bleve.NewMatchQuery(keyword)
	nameNGramQuery.SetField("NameNGram")
	nameNGramQuery.SetBoost(2)
	query.AddShould(nameNGramQuery)

	namePrefixQuery := bleve.NewPrefixQuery(keyword)
	namePrefixQuery.SetField("Name")
	namePrefixQuery.SetBoost(10)
	query.AddShould(namePrefixQuery)

	nameTermQuery := bleve.NewTermQuery(keyword)
	nameTermQuery.SetField("Name")
	nameTermQuery.SetBoost(100)
	query.AddShould(nameTermQuery)

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
) (*nix.Importable, errors.E) {
	key := nix.MakeKey(source, id)
	query := bleve.NewDocIDQuery([]string{key})
	search := bleve.NewSearchRequest(query)
	search.Size = 1

	result, err := index.search(ctx, search)
	if err != nil {
		return nil, err
	}

	if result.Total == 0 {
		return nil, nil
	}

	for hit := range result.Hits {
		if hit.ID == key {
			return &hit.Data, err
		}
	}

	return nil, err
}

func (index *ReadIndex) Close() error {
	err := index.index.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
