package search

import (
	"bytes"
	"context"
	"encoding/gob"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"path"
	"searchix/internal/file"
	"searchix/internal/options"
	"slices"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/web"
	"github.com/blevesearch/bleve/v2/analysis/token/camelcase"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/letter"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/single"
	"github.com/blevesearch/bleve/v2/document"
	"github.com/blevesearch/bleve/v2/mapping"
	index "github.com/blevesearch/bleve_index_api"
	"github.com/pkg/errors"
)

type WriteIndex struct {
	index bleve.Index
	meta  *IndexMeta
}

func createIndexMapping() (mapping.IndexMapping, error) {
	indexMapping := bleve.NewIndexMapping()
	indexMapping.StoreDynamic = false
	indexMapping.IndexDynamic = false
	indexMapping.TypeField = "BleveType"

	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Store = false

	descriptionFieldMapping := bleve.NewTextFieldMapping()
	descriptionFieldMapping.Store = false
	descriptionFieldMapping.Analyzer = web.Name

	err := indexMapping.AddCustomAnalyzer("option_name", map[string]interface{}{
		"type":      custom.Name,
		"tokenizer": letter.Name,
		"token_filters": []string{
			camelcase.Name,
		},
	})
	if err != nil {
		return nil, errors.WithMessage(err, "could not add custom analyser")
	}
	err = indexMapping.AddCustomAnalyzer("loc", map[string]interface{}{
		"type":      keyword.Name,
		"tokenizer": letter.Name,
		"token_filters": []string{
			camelcase.Name,
		},
	})
	if err != nil {
		return nil, errors.WithMessage(err, "could not add custom analyser")
	}
	err = indexMapping.AddCustomAnalyzer("keyword_single", map[string]interface{}{
		"type":      keyword.Name,
		"tokenizer": single.Name,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "could not add custom analyser")
	}

	keywordFieldMapping := bleve.NewKeywordFieldMapping()
	keywordFieldMapping.Analyzer = "keyword_single"

	nameMapping := bleve.NewTextFieldMapping()
	nameMapping.Analyzer = "option_name"
	nameMapping.IncludeTermVectors = true
	nameMapping.Store = false

	nixValueMapping := bleve.NewDocumentStaticMapping()
	nixValueMapping.AddFieldMappingsAt("Text", textFieldMapping)
	nixValueMapping.AddFieldMappingsAt("Markdown", textFieldMapping)

	locFieldMapping := bleve.NewKeywordFieldMapping()
	locFieldMapping.Analyzer = "loc"
	locFieldMapping.IncludeTermVectors = true
	locFieldMapping.Store = false

	optionMapping := bleve.NewDocumentStaticMapping()

	optionMapping.AddFieldMappingsAt("Name", keywordFieldMapping)
	optionMapping.AddFieldMappingsAt("Source", keywordFieldMapping)
	optionMapping.AddFieldMappingsAt("Loc", locFieldMapping)
	optionMapping.AddFieldMappingsAt("RelatedPackages", textFieldMapping)
	optionMapping.AddFieldMappingsAt("Description", textFieldMapping)

	optionMapping.AddSubDocumentMapping("Default", nixValueMapping)
	optionMapping.AddSubDocumentMapping("Example", nixValueMapping)

	indexMapping.AddDocumentMapping("option", optionMapping)

	return indexMapping, nil
}

func createIndex(indexPath string) (bleve.Index, error) {
	indexMapping, err := createIndexMapping()
	if err != nil {
		return nil, err
	}
	idx, err := bleve.NewUsing(
		indexPath,
		indexMapping,
		bleve.Config.DefaultIndexType,
		bleve.Config.DefaultKVStore,
		map[string]interface{}{
			"nosync": true,
		},
	)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to create index at path %s", indexPath)
	}

	return idx, nil
}

const (
	indexBaseName = "index.bleve"
	metaBaseName  = "meta.json"
)

var expectedDataFiles = []string{
	metaBaseName,
	indexBaseName,
	"sources",
}

func deleteIndex(dataRoot string) error {
	dir, err := os.ReadDir(dataRoot)
	if err != nil {
		return errors.WithMessagef(err, "could not read data directory %s", dataRoot)
	}
	remainingFiles := slices.DeleteFunc(dir, func(e fs.DirEntry) bool {
		return slices.Contains(expectedDataFiles, e.Name())
	})
	if len(remainingFiles) > 0 {
		return errors.Errorf(
			"cowardly refusing to remove data directory %s as it contains unknown files: %v",
			dataRoot,
			remainingFiles,
		)
	}

	err = os.RemoveAll(dataRoot)
	if err != nil {
		return errors.WithMessagef(err, "could not remove data directory %s", dataRoot)
	}

	return nil
}

func NewIndexer(dataRoot string, force bool) (*WriteIndex, error) {
	var err error
	bleve.SetLog(log.Default())

	indexPath := path.Join(dataRoot, indexBaseName)
	metaPath := path.Join(dataRoot, metaBaseName)

	exists, err := file.Exists(indexPath)
	if err != nil {
		return nil, errors.WithMessagef(
			err,
			"could not check if index exists at path %s",
			indexPath,
		)
	}

	var idx bleve.Index
	var meta *IndexMeta
	if !exists || force {
		if force {
			err = deleteIndex(dataRoot)
			if err != nil {
				return nil, err
			}
		}
		idx, err = createIndex(indexPath)
		if err != nil {
			return nil, err
		}

		meta, err = createMeta(metaPath)
		if err != nil {
			return nil, err
		}

		err = meta.Save()
		if err != nil {
			return nil, err
		}
	} else {
		idx, err = bleve.Open(indexPath)
		if err != nil {
			return nil, errors.WithMessagef(err, "could not open index at path %s", indexPath)
		}

		meta, err = openMeta(metaPath)
		if err != nil {
			return nil, err
		}

	}

	return &WriteIndex{
		idx,
		meta,
	}, nil
}

func (i *WriteIndex) ImportOptions(
	ctx context.Context,
	objects <-chan *options.NixOption,
) <-chan error {
	var err error
	errs := make(chan error)

	go func() {
		defer close(errs)
		batch := i.index.NewBatch()
		indexMapping := i.index.Mapping()

	outer:
		for opt := range objects {
			select {
			case <-ctx.Done():
				slog.Debug("context cancelled")

				break outer
			default:
			}

			doc := document.NewDocument(opt.Source + "/" + opt.Name)
			err = indexMapping.MapDocument(doc, opt)
			if err != nil {
				errs <- errors.WithMessagef(err, "could not map document for option: %s", opt.Name)

				continue
			}

			var data bytes.Buffer
			enc := gob.NewEncoder(&data)
			err = enc.Encode(opt)
			if err != nil {
				errs <- errors.WithMessage(err, "could not store option in search index")

				continue
			}
			field := document.NewTextFieldWithIndexingOptions("_data", nil, data.Bytes(), index.StoreField)
			newDoc := doc.AddField(field)

			// slog.Debug("adding option to index", "name", opt.Name)
			err = batch.IndexAdvanced(newDoc)

			if err != nil {
				errs <- errors.WithMessagef(err, "could not index option %s", opt.Name)

				continue
			}
		}

		size := batch.Size()
		slog.Debug("flushing batch", "size", size)

		err := i.index.Batch(batch)
		if err != nil {
			errs <- errors.WithMessagef(err, "could not flush batch")
		}
	}()

	return errs
}

func (i *WriteIndex) Close() error {
	err := i.index.Close()
	if err != nil {
		return errors.WithMessagef(err, "could not close index")
	}

	return nil
}
