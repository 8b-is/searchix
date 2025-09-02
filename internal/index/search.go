package index

import (
	"context"
	"errors"
	"iter"
	"strings"
	"time"

	"alin.ovh/x/log"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/nix"
	"alin.ovh/searchix/internal/storage"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/asdine/storm/v3"
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
	store  *storage.Store
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
	facet := bleve.NewFacetRequest("Source", 100)
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
	source *config.Source,
	request *bleve.SearchRequest,
) (*Result, error) {
	request.Fields = []string{"Source"}

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
			for _, match := range bleveResult.Hits {
				hit := DocumentMatch{
					DocumentMatch: match,
					Data:          nil,
				}

				parts := strings.SplitN(match.ID, "/", 3)
				sourceName := parts[1]
				id := parts[2]

				src := source
				if src == nil {
					var ok bool
					src, ok = index.config.Importer.Sources[sourceName]
					if !ok {
						continue
					}
				}

				doc, err := index.store.GetDocument(src, id)
				if err != nil {
					if errors.Is(err, storm.ErrNotFound) {
						index.log.Warn("document not found", "source", sourceName, "id", id)
					} else {
						index.log.Error("error getting document", "error", err)
					}

					continue
				}

				hit.Data = doc

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
) (*Result, error) {
	query := bleve.NewBooleanQuery()

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

	return index.search(ctx, source, search)
}

func (index *ReadIndex) Close() error {
	err := index.index.Close()
	if err != nil {
		return fault.Wrap(err)
	}

	return nil
}
