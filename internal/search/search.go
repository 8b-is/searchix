package search

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"
	"os"
	"path"
	"searchix/internal/options"

	"github.com/bcicen/jstream"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/token/camelcase"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/letter"
	"github.com/blevesearch/bleve/v2/document"
	"github.com/blevesearch/bleve/v2/search"
	index "github.com/blevesearch/bleve_index_api"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

const ResultsPerPage = 20

type DocumentMatch[T options.NixOption] struct {
	search.DocumentMatch
	Data T
}

type Result[T options.NixOption] struct {
	*bleve.SearchResult
	Hits []DocumentMatch[T]
}

type Index[T options.NixOption] struct {
	index bleve.Index
}

func New[T options.NixOption](kind string) (*Index[T], error) {
	var err error
	bleve.SetLog(log.Default())

	indexMapping := bleve.NewIndexMapping()

	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Store = false

	descriptionFieldMapping := bleve.NewTextFieldMapping()
	descriptionFieldMapping.Store = false
	descriptionFieldMapping.Analyzer = "web"

	err = indexMapping.AddCustomAnalyzer("option_name", map[string]interface{}{
		"type":      custom.Name,
		"tokenizer": letter.Name,
		"token_filters": []string{
			camelcase.Name,
		},
	})
	if err != nil {
		return nil, errors.WithMessage(err, "could not add custom analyser")
	}
	nameMapping := bleve.NewTextFieldMapping()
	nameMapping.Analyzer = "option_name"
	nameMapping.IncludeTermVectors = true
	nameMapping.Store = false

	nixValueMapping := bleve.NewDocumentStaticMapping()
	nixValueMapping.AddFieldMappingsAt("Text", textFieldMapping)
	nixValueMapping.AddFieldMappingsAt("Markdown", textFieldMapping)

	optionMapping := bleve.NewDocumentStaticMapping()

	optionMapping.AddFieldMappingsAt("Option", nameMapping)
	optionMapping.AddFieldMappingsAt("Loc", bleve.NewKeywordFieldMapping())
	optionMapping.AddFieldMappingsAt("RelatedPackages", textFieldMapping)
	optionMapping.AddFieldMappingsAt("Description", textFieldMapping)

	optionMapping.AddSubDocumentMapping("Default", nixValueMapping)
	optionMapping.AddSubDocumentMapping("Example", nixValueMapping)

	indexMapping.AddDocumentMapping("option", optionMapping)

	idx, err := bleve.NewMemOnly(indexMapping)
	// index, err = bleve.New(path.Join(cfg.DataPath, const indexFilename = "index.bleve"), indexMapping)

	if err != nil {
		return nil, errors.WithMessage(err, "error opening index")
	}
	batch := idx.NewBatch()

	jsonFile, err := os.Open(path.Join("data", "processed", kind+".json"))
	if err != nil {
		return nil, errors.WithMessage(err, "error opening json file")
	}

	dec := jstream.NewDecoder(jsonFile, 1)
	var opt options.NixOption
	ms, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		ZeroFields:  true,
		Result:      &opt,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "could not create struct decoder")
	}
	for mv := range dec.Stream() {
		opt = options.NixOption{}
		orig := mv.Value.(map[string]interface{})
		err := ms.Decode(orig) // stores in opt
		if err != nil {
			return nil, errors.WithMessagef(err, "could not decode value: %+v", orig)
		}

		doc := document.NewDocument(opt.Option)
		err = indexMapping.MapDocument(doc, opt)
		if err != nil {
			return nil, errors.WithMessagef(err, "could not map document for option: %s", opt.Option)
		}

		var data bytes.Buffer
		enc := gob.NewEncoder(&data)
		err = enc.Encode(opt)
		if err != nil {
			return nil, errors.WithMessage(err, "could not store option in search index")
		}
		field := document.NewTextFieldWithIndexingOptions("data", nil, data.Bytes(), index.StoreField)
		newDoc := doc.AddField(field)

		err = batch.IndexAdvanced(newDoc)

		if err != nil {
			return nil, errors.WithMessagef(err, "could not index option %s", opt.Option)
		}
	}
	err = idx.Batch(batch)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to run batch index operation")
	}

	return &Index[T]{
		idx,
	}, nil
}

func (index *Index[T]) Search(ctx context.Context, keyword string, from uint64) (*Result[T], error) {
	query := bleve.NewMatchQuery(keyword)
	query.Analyzer = "option_name"
	search := bleve.NewSearchRequest(query)
	search.Size = ResultsPerPage
	search.Fields = []string{"data"}
	search.Explain = true

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

		results := make([]DocumentMatch[T], min(ResultsPerPage, bleveResult.Total))
		var buf bytes.Buffer
		for i, result := range bleveResult.Hits {
			_, err = buf.WriteString(result.Fields["data"].(string))
			if err != nil {
				return nil, errors.WithMessage(err, "error fetching result data")
			}
			err = gob.NewDecoder(&buf).Decode(&results[i].Data)
			if err != nil {
				return nil, errors.WithMessagef(err, "error decoding gob data: %s", buf.String())
			}
			buf.Reset()
		}

		return &Result[T]{
			SearchResult: bleveResult,
			Hits:         results,
		}, nil
	}
}
