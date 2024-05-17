package index

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
	"searchix/internal/nix"
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
	indexAPI "github.com/blevesearch/bleve_index_api"
	"github.com/pkg/errors"
)

type WriteIndex struct {
	index bleve.Index
	meta  *Meta
}

const batchSize = 10_000

func createIndexMapping() (mapping.IndexMapping, error) {
	indexMapping := bleve.NewIndexMapping()
	indexMapping.StoreDynamic = false
	indexMapping.IndexDynamic = false

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

	packageMapping := bleve.NewDocumentStaticMapping()

	packageMapping.AddFieldMappingsAt("Name", keywordFieldMapping)
	packageMapping.AddFieldMappingsAt("Source", keywordFieldMapping)
	packageMapping.AddFieldMappingsAt("Description", textFieldMapping)
	packageMapping.AddFieldMappingsAt("MainProgram", keywordFieldMapping)
	packageMapping.AddFieldMappingsAt("PackageSet", keywordFieldMapping)
	packageMapping.AddFieldMappingsAt("Platforms", keywordFieldMapping)

	indexMapping.AddDocumentMapping("option", optionMapping)
	indexMapping.AddDocumentMapping("package", packageMapping)

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

func OpenOrCreate(dataRoot string, force bool) (*ReadIndex, *WriteIndex, bool, error) {
	var err error
	bleve.SetLog(log.Default())

	indexPath := path.Join(dataRoot, indexBaseName)
	metaPath := path.Join(dataRoot, metaBaseName)

	exists, err := file.Exists(indexPath)
	if err != nil {
		return nil, nil, exists, errors.WithMessagef(
			err,
			"could not check if index exists at path %s",
			indexPath,
		)
	}

	var idx bleve.Index
	var meta *Meta
	if !exists || force {
		if force {
			err = deleteIndex(dataRoot)
			if err != nil {
				return nil, nil, false, err
			}
		}
		idx, err = createIndex(indexPath)
		if err != nil {
			return nil, nil, false, err
		}

		meta, err = createMeta(metaPath)
		if err != nil {
			return nil, nil, false, err
		}

		err = meta.Save()
		if err != nil {
			return nil, nil, false, err
		}
	} else {
		idx, err = bleve.Open(indexPath)
		if err != nil {
			return nil, nil, exists, errors.WithMessagef(err, "could not open index at path %s", indexPath)
		}

		meta, err = openMeta(metaPath)
		if err != nil {
			return nil, nil, exists, err
		}

	}

	return &ReadIndex{
			idx,
			meta,
		},
		&WriteIndex{
			idx,
			meta,
		},
		exists,
		nil
}

func (i *WriteIndex) Import(
	ctx context.Context,
	objects <-chan nix.Importable,
) <-chan error {
	var err error
	errs := make(chan error)

	go func() {
		defer close(errs)
		k := 0
		batch := i.index.NewBatch()
		indexMapping := i.index.Mapping()

	outer:
		for obj := range objects {
			select {
			case <-ctx.Done():
				slog.Debug("context cancelled")

				break outer
			default:
			}

			doc := document.NewDocument(nix.GetKey(obj))
			err = indexMapping.MapDocument(doc, obj)
			if err != nil {
				errs <- errors.WithMessagef(err, "could not map document for object: %s", obj.GetName())

				continue
			}

			var data bytes.Buffer
			enc := gob.NewEncoder(&data)
			err = enc.Encode(&obj)
			if err != nil {
				errs <- errors.WithMessage(err, "could not store object in search index")

				continue
			}
			field := document.NewTextFieldWithIndexingOptions("_data", nil, data.Bytes(), indexAPI.StoreField)
			newDoc := doc.AddField(field)

			// slog.Debug("adding object to index", "name", opt.Name)
			err = batch.IndexAdvanced(newDoc)

			if err != nil {
				errs <- errors.WithMessagef(err, "could not index object %s", obj.GetName())

				continue
			}

			if k++; k%batchSize == 0 {
				err = i.Flush(batch)
				if err != nil {
					errs <- err
				}
			}
		}

		err := i.Flush(batch)
		if err != nil {
			errs <- err
		}
	}()

	return errs
}

func (i *WriteIndex) Flush(batch *bleve.Batch) error {
	size := batch.Size()
	slog.Debug("flushing batch", "size", size)

	err := i.index.Batch(batch)
	if err != nil {
		return errors.WithMessagef(err, "could not flush batch")
	}

	batch.Reset()

	return nil
}

func (i *WriteIndex) Close() error {
	err := i.index.Close()
	if err != nil {
		return errors.WithMessagef(err, "could not close index")
	}

	return nil
}
