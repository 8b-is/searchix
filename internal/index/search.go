package index

import (
	"bytes"
	"context"
	"encoding/gob"
	"iter"
	"net/url"
	"strings"
	"time"

	"alin.ovh/x/log"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/nix"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
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
	index  bleve.Index
	config *config.Config
	log    *log.Logger
	exists bool
	meta   *Meta
}

func (index *ReadIndex) Exists() bool {
	return index.exists
}

func (index *ReadIndex) LastUpdated() time.Time {
	return index.meta.LastUpdated()
}

func (index *ReadIndex) GetEnabledSources() ([]string, error) {
	facet := bleve.NewFacetRequest("Source", len(index.config.Importer.Sources))
	query := bleve.NewMatchAllQuery()
	search := bleve.NewSearchRequest(query)
	search.AddFacet("Source", facet)

	results, err := index.index.Search(search)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("could not get list of enabled sources from index"))
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
) (*Result, error) {
	request.Fields = []string{"_data", "Source"}

	bleveResult, err := index.index.SearchInContext(ctx, request)
	select {
	case <-ctx.Done():
		return nil, fault.Wrap(ctx.Err(), fctx.With(ctx))
	default:
		if err != nil {
			return nil, fault.Wrap(err,
				fctx.With(ctx),
				fmsg.Withf("failed to execute search query"),
			)
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
	facets url.Values,
) (*Result, error) {
	query := bleve.NewBooleanQuery()

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

	if keyword != "" {
		if strings.ContainsAny(keyword, "+-=&|<>!(){}[]^\"~*?:\\/") {
			query.AddMust(bleve.NewQueryStringQuery(keyword))
		} else {
			// match the user's query in any field ...
			query.AddMust(bleve.NewDisjunctionQuery(
				setBoost(bleve.NewTermQuery(keyword), 50),
				setBoost(bleve.NewPrefixQuery(keyword), 25),
				setBoost(bleve.NewMatchPhraseQuery(keyword), 25),
				bleve.NewMatchQuery(keyword),
			))
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
	}

	search := bleve.NewSearchRequest(query)
	search.Explain = config.DevMode
	search.Size = pageSize

	if len(facets) > 0 {
		for name, values := range facets {
			or := bleve.NewDisjunctionQuery()
			for _, value := range values {
				q := bleve.NewTermQuery(value)
				q.SetField(name)

				or.AddQuery(q)
			}
			query.AddMust(or)
		}

		if keyword == "" {
			search.SortBy([]string{"_id"})
		}
	}

	if source == nil || source.Importer == config.Packages {
		search.AddFacet("Package set", bleve.NewFacetRequest("PackageSet", 10))
		search.AddFacet("Platform", bleve.NewFacetRequest("Platforms", 10))
	}
	if source == nil || source.Importer == config.Options {
		search.AddFacet("Option set", bleve.NewFacetRequest("Parents", 10))
	}

	if from != 0 {
		search.From = from
	}

	return index.search(ctx, search)
}

func (index *ReadIndex) ImportedBefore(
	cutoff time.Time,
	source *config.Source,
) (*bleve.SearchResult, error) {
	cutoffQuery := bleve.NewDateRangeQuery(time.UnixMilli(0), cutoff)
	cutoffQuery.SetField("ImportedAt")

	all := bleve.NewConjunctionQuery(cutoffQuery)

	if source != nil {
		sourceQuery := bleve.NewTermQuery(source.Key)
		sourceQuery.SetField("Source")

		all.AddQuery(sourceQuery)
	}

	req := bleve.NewSearchRequest(all)
	req.Size = 10_000
	req.SortBy([]string{"_id"})

	res, err := index.index.Search(req)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("could not query old documents"))
	}

	return res, nil
}

func (index *ReadIndex) Count(source *config.Source) (uint64, error) {
	if source == nil {
		count, err := index.index.DocCount()
		if err != nil {
			return 0, fault.Wrap(err)
		}

		return count, nil
	}

	query := setField(bleve.NewTermQuery(source.Key), "Source")
	req := bleve.NewSearchRequest(query)
	res, err := index.index.Search(req)
	if err != nil {
		return 0, fault.Wrap(err)
	}

	return res.Total, nil
}

func (index *ReadIndex) Close() error {
	err := index.index.Close()
	if err != nil {
		return fault.Wrap(err)
	}

	return nil
}

func (index *ReadIndex) GetDocument(
	ctx context.Context,
	source *config.Source,
	id string,
) (nix.Importable, error) {
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
			return hit.Data, err
		}
	}

	return nil, err
}
