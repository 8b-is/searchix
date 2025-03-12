package index

import (
	"bytes"
	"context"
	"encoding/gob"
	"io/fs"
	"math"
	"os"
	"path"
	"slices"

	"go.alanpearce.eu/searchix/internal/file"
	"go.alanpearce.eu/searchix/internal/nix"
	"go.alanpearce.eu/x/log"
	"go.uber.org/zap"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/web"
	"github.com/blevesearch/bleve/v2/analysis/lang/en"
	"github.com/blevesearch/bleve/v2/analysis/token/camelcase"
	"github.com/blevesearch/bleve/v2/analysis/token/porter"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/letter"
	"github.com/blevesearch/bleve/v2/document"
	"github.com/blevesearch/bleve/v2/mapping"
	indexAPI "github.com/blevesearch/bleve_index_api"
	"gitlab.com/tozd/go/errors"
)

type WriteIndex struct {
	index bleve.Index
	log   *log.Logger
	Meta  *Meta
}

type BatchError struct {
	errors.E
}

func (e *BatchError) Error() string {
	return e.E.Error()
}

var batchSize = 10_000

func createIndexMapping() (mapping.IndexMapping, errors.E) {
	indexMapping := bleve.NewIndexMapping()
	indexMapping.StoreDynamic = false
	indexMapping.IndexDynamic = false

	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Store = false

	descriptionFieldMapping := bleve.NewTextFieldMapping()
	descriptionFieldMapping.Store = false
	descriptionFieldMapping.Analyzer = web.Name

	var err error
	err = indexMapping.AddCustomAnalyzer("loc", map[string]any{
		"type":      keyword.Name,
		"tokenizer": letter.Name,
		"token_filters": []string{
			camelcase.Name,
			porter.Name,
		},
	})
	if err != nil {
		return nil, errors.WithMessage(err, "could not add custom analyser")
	}
	err = indexMapping.AddCustomAnalyzer("keyword_single", map[string]any{
		"type":      keyword.Name,
		"tokenizer": letter.Name,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "could not add custom analyser")
	}

	identityFieldMapping := bleve.NewKeywordFieldMapping()

	keywordFieldMapping := bleve.NewKeywordFieldMapping()
	keywordFieldMapping.Analyzer = simple.Name

	nameMapping := bleve.NewTextFieldMapping()
	nameMapping.Analyzer = en.AnalyzerName
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

	optionMapping.AddFieldMappingsAt("Name", nameMapping)
	optionMapping.AddFieldMappingsAt("Source", identityFieldMapping)
	optionMapping.AddFieldMappingsAt("Loc", locFieldMapping)
	optionMapping.AddFieldMappingsAt("RelatedPackages", textFieldMapping)
	optionMapping.AddFieldMappingsAt("Description", descriptionFieldMapping)

	optionMapping.AddSubDocumentMapping("Default", nixValueMapping)
	optionMapping.AddSubDocumentMapping("Example", nixValueMapping)

	packageMapping := bleve.NewDocumentStaticMapping()

	packageMapping.AddFieldMappingsAt("Name", nameMapping)
	packageMapping.AddFieldMappingsAt("Attribute", keywordFieldMapping)
	packageMapping.AddFieldMappingsAt("Source", keywordFieldMapping)
	packageMapping.AddFieldMappingsAt("Description", descriptionFieldMapping)
	packageMapping.AddFieldMappingsAt("MainProgram", keywordFieldMapping)
	packageMapping.AddFieldMappingsAt("PackageSet", keywordFieldMapping)
	packageMapping.AddFieldMappingsAt("Platforms", keywordFieldMapping)
	packageMapping.AddFieldMappingsAt("Programs", keywordFieldMapping)

	indexMapping.AddDocumentMapping("option", optionMapping)
	indexMapping.AddDocumentMapping("package", packageMapping)

	return indexMapping, nil
}

func createIndex(indexPath string, options *Options) (bleve.Index, errors.E) {
	indexMapping, err := createIndexMapping()
	if err != nil {
		return nil, err
	}
	kvconfig := make(map[string]any)
	if options.LowMemory {
		kvconfig = map[string]any{
			"PersisterNapTimeMSec":      1000,
			"PersisterNapUnderNumFiles": 500,
		}
	}
	idx, baseErr := bleve.NewUsing(
		indexPath,
		indexMapping,
		bleve.Config.DefaultIndexType,
		bleve.Config.DefaultKVStore,
		kvconfig,
	)
	if baseErr != nil {
		return nil, errors.WithMessagef(baseErr, "unable to create index at path %s", indexPath)
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
	"nixpkgs-programs.db",
}

func deleteIndex(dataRoot string) errors.E {
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

type Options struct {
	LowMemory bool
	Logger    *log.Logger
}

func OpenOrCreate(
	dataRoot string,
	force bool,
	options *Options,
) (*ReadIndex, *WriteIndex, bool, errors.E) {
	var err errors.E
	bleve.SetLog(zap.NewStdLog(options.Logger.Named("bleve").GetLogger()))

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
		idx, err = createIndex(indexPath, options)
		if err != nil {
			return nil, nil, false, err
		}

		meta, err = createMeta(metaPath, options.Logger)
		if err != nil {
			return nil, nil, false, err
		}

	} else {
		var baseErr error
		idx, baseErr = bleve.Open(indexPath)
		if baseErr != nil {
			return nil, nil, exists, errors.WithMessagef(baseErr, "could not open index at path %s", indexPath)
		}

		meta, err = openMeta(metaPath, options.Logger)
		if err != nil {
			return nil, nil, exists, err
		}
	}

	if options.LowMemory {
		batchSize = 1_000
	}

	return &ReadIndex{
			index: idx,
			log:   options.Logger,
			meta:  meta,
		},
		&WriteIndex{
			index: idx,
			log:   options.Logger,
			Meta:  meta,
		},
		exists,
		nil
}

func (i *WriteIndex) SaveMeta() errors.E {
	return i.Meta.Save()
}

func (i *WriteIndex) Import(
	ctx context.Context,
	objects <-chan nix.Importable,
) <-chan errors.E {
	var err errors.E
	errs := make(chan errors.E)

	go func() {
		defer close(errs)
		k := 0
		batch := i.index.NewBatch()
		indexMapping := i.index.Mapping()

	outer:
		for obj := range objects {
			select {
			case <-ctx.Done():
				i.log.Warn("import aborted")

				break outer
			default:
			}

			doc := document.NewDocument(nix.GetKey(obj))
			if err := indexMapping.MapDocument(doc, obj); err != nil {
				errs <- errors.WithMessagef(err, "could not map document for object: %s", obj.GetName())

				continue
			}

			var data bytes.Buffer
			enc := gob.NewEncoder(&data)
			if err := enc.Encode(&obj); err != nil {
				errs <- errors.WithMessage(err, "could not store object in search index")

				continue
			}
			field := document.NewTextFieldWithIndexingOptions("_data", nil, data.Bytes(), indexAPI.StoreField)
			newDoc := doc.AddField(field)

			// log.Debug("adding object to index", "name", opt.Name)
			if err := batch.IndexAdvanced(newDoc); err != nil {
				errs <- errors.WithMessagef(err, "could not index object %s", obj.GetName())

				continue
			}

			if k++; k%batchSize == 0 {
				err = i.Flush(batch)
				if err != nil {
					errs <- err

					return
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

func (i *WriteIndex) Flush(batch *bleve.Batch) errors.E {
	size := batch.Size()
	if size == 0 {
		return &BatchError{
			E: errors.New("no documents to flush"),
		}
	}
	i.log.Debug("flushing batch", "size", size)

	err := i.index.Batch(batch)
	if err != nil {
		return &BatchError{
			E: errors.WithMessagef(err, "could not flush batch"),
		}
	}

	batch.Reset()

	return nil
}

func (i *WriteIndex) Close() (err errors.E) {
	if e := i.Meta.Save(); e != nil {
		// index needs to be closed anyway
		err = errors.WithMessage(e, "could not save metadata")
	}

	if e := i.index.Close(); e != nil {
		err = errors.WithMessagef(e, "could not close index")
	}

	return err
}

func (i *WriteIndex) DeleteBySource(source string) errors.E {
	query := bleve.NewTermQuery(source)
	search := bleve.NewSearchRequest(query)
	search.Size = math.MaxInt
	search.Fields = []string{"_id"}

	results, err := i.index.Search(search)
	if err != nil {
		return errors.WithMessagef(err, "failed to query documents of retired index %s", source)
	}

	batch := i.index.NewBatch()
	var k int
	for _, hit := range results.Hits {
		batch.Delete(hit.ID)
		if k++; k%batchSize == 0 {
			err := i.Flush(batch)
			if err != nil {
				return err
			}
		}
	}
	err = i.Flush(batch)
	if err != nil {
		return errors.WithStack(err)
	}

	if uint64(search.Size) < results.Total {
		return i.DeleteBySource(source) // unlikely :^)
	}

	return nil
}
