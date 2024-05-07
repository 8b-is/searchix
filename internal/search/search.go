package search

import (
	"context"
	"log"
	"os"
	"path"
	"sync"

	"searchix/internal/options"

	"github.com/bcicen/jstream"
	"github.com/blevesearch/bleve/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

const maxResults = 10

var index bleve.Index
var docs sync.Map

type Result[T options.NixOption] struct {
	*bleve.SearchResult
	Results []T
}

func Index() error {
	bleve.SetLog(log.Default())
	textFieldMapping := bleve.NewTextFieldMapping()
	nameMapping := bleve.NewTextFieldMapping()
	// something special for tokenisation?
	// nameMapping.
	optionMapping := bleve.NewDocumentStaticMapping()

	optionMapping.AddFieldMappingsAt("Option", nameMapping)
	optionMapping.AddFieldMappingsAt("RelatedPackages", textFieldMapping)
	optionMapping.AddFieldMappingsAt("Description", textFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("option", optionMapping)

	var err error
	index, err = bleve.NewMemOnly(indexMapping)
	// index, err = bleve.New(path.Join(cfg.DataPath, const indexFilename = "index.bleve"), indexMapping)

	if err != nil {
		return errors.WithMessage(err, "error opening index")
	}
	batch := index.NewBatch()
	_ = batch

	jsonFile, err := os.Open(path.Join("data", "processed", "darwin-options.json"))
	if err != nil {
		return errors.WithMessage(err, "error opening json file")
	}

	dec := jstream.NewDecoder(jsonFile, 1)
	var opt options.NixOption
	ms, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		ZeroFields:  true,
		Result:      &opt,
	})
	if err != nil {
		return errors.WithMessage(err, "could not create struct decoder")
	}
	for mv := range dec.Stream() {
		err := ms.Decode(mv.Value) // stores in opt

		if err != nil {
			return errors.WithMessagef(err, "could not decode object into option, object: %#v", mv.Value)
		}

		docs.Store(opt.Option, opt)

		err = batch.Index(opt.Option, opt)
		if err != nil {
			return errors.WithMessagef(err, "could not index option %s", opt.Option)
		}
	}
	err = index.Batch(batch)
	if err != nil {
		return errors.WithMessage(err, "failed to run batch index operation")
	}

	return nil
}

func Search[T options.NixOption](ctx context.Context, keyword string) (*Result[T], error) {
	query := bleve.NewMatchQuery(keyword)
	search := bleve.NewSearchRequest(query)

	bleveResult, err := index.SearchInContext(ctx, search)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if err != nil {
			return nil, errors.WithMessage(err, "failed to execute search query")
		}

		results := make([]T, min(maxResults, bleveResult.Total))
		for i, result := range bleveResult.Hits {
			doc, _ := docs.Load(result.ID)
			results[i] = doc.(T)
			if i > maxResults {
				break
			}
		}

		return &Result[T]{
			bleveResult,
			results,
		}, nil
	}
}

func Load(name string) any {
	doc, _ := docs.Load(name)

	return doc
}
