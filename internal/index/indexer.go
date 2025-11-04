package index

import (
	"bytes"
	"context"
	"encoding/gob"
	"math"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/index/meta"
	"alin.ovh/searchix/internal/index/nixattr"
	"alin.ovh/searchix/internal/nix"

	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/web"
	"github.com/blevesearch/bleve/v2/analysis/token/camelcase"
	"github.com/blevesearch/bleve/v2/analysis/token/ngram"
	"github.com/blevesearch/bleve/v2/analysis/token/porter"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/letter"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/unicode"
	"github.com/blevesearch/bleve/v2/document"
	"github.com/blevesearch/bleve/v2/mapping"
	index "github.com/blevesearch/bleve_index_api"
	"go.uber.org/zap"
)

type Options struct {
	Force     bool
	LowMemory bool
	BatchSize int
	Logger    *log.Logger
	Root      *file.Root
	Config    *config.Config
}

type WriteIndex struct {
	batchSize int
	index     bleve.Index
	log       *log.Logger
	exists    bool
	Meta      *meta.Meta
}

type BatchError struct {
	error
}

var idAnalyzer analysis.Analyzer

func createIndexMapping() (mapping.IndexMapping, error) {
	indexMapping := bleve.NewIndexMapping()
	indexMapping.StoreDynamic = false
	indexMapping.IndexDynamic = false

	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Store = false

	descriptionFieldMapping := bleve.NewTextFieldMapping()
	descriptionFieldMapping.Store = false
	descriptionFieldMapping.Analyzer = web.Name

	var err error

	err = indexMapping.AddCustomTokenFilter("ngram", map[string]any{
		"type": ngram.Name,
		"min":  3.0,
		"max":  25.0,
	})
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to add ngram token filter"))
	}

	err = indexMapping.AddCustomAnalyzer("c_name", map[string]any{
		"type":      custom.Name,
		"tokenizer": unicode.Name,
		"token_filters": []string{
			nixattr.Name,
			"ngram",
		},
	})
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("could not add custom analyser"))
	}

	err = indexMapping.AddCustomAnalyzer("loc", map[string]any{
		"type":      keyword.Name,
		"tokenizer": letter.Name,
		"token_filters": []string{
			camelcase.Name,
			porter.Name,
		},
	})
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("could not add custom analyser"))
	}

	err = indexMapping.AddCustomAnalyzer("dotted_keyword", map[string]any{
		"type":      custom.Name,
		"tokenizer": unicode.Name,
		"token_filters": []string{
			nixattr.Name,
		},
	})
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("could not add custom analyser"))
	}

	identityFieldMapping := bleve.NewKeywordFieldMapping()
	identityFieldMapping.Store = false

	attributeFieldMapping := bleve.NewKeywordFieldMapping()
	attributeFieldMapping.Analyzer = "dotted_keyword"
	attributeFieldMapping.Store = true

	keywordFieldMapping := bleve.NewKeywordFieldMapping()
	keywordFieldMapping.Analyzer = simple.Name
	keywordFieldMapping.Store = false

	nameNGramMapping := bleve.NewTextFieldMapping()
	nameNGramMapping.Analyzer = "c_name"
	nameNGramMapping.IncludeTermVectors = true

	nixDocMapping := bleve.NewDocumentStaticMapping()
	nixDocMapping.AddFieldMappingsAt("Text", textFieldMapping)
	nixDocMapping.AddFieldMappingsAt("Markdown", textFieldMapping)

	locFieldMapping := bleve.NewKeywordFieldMapping()
	locFieldMapping.Analyzer = "loc"
	locFieldMapping.IncludeTermVectors = true
	locFieldMapping.Store = false

	optionMapping := bleve.NewDocumentStaticMapping()

	optionMapping.AddFieldMappingsAt(
		"Name",
		attributeFieldMapping,
		locFieldMapping,
		nameNGramMapping,
	)
	optionMapping.AddFieldMappingsAt("Source", identityFieldMapping)
	optionMapping.AddFieldMappingsAt("Loc", locFieldMapping)
	optionMapping.AddFieldMappingsAt("Parents", locFieldMapping)
	optionMapping.AddFieldMappingsAt("RelatedPackages", textFieldMapping)
	optionMapping.AddFieldMappingsAt("Description", descriptionFieldMapping)

	optionMapping.AddSubDocumentMapping("Default", nixDocMapping)
	optionMapping.AddSubDocumentMapping("Example", nixDocMapping)

	packageMapping := bleve.NewDocumentStaticMapping()

	packageMapping.AddFieldMappingsAt(
		"Name",
		keywordFieldMapping,
		locFieldMapping,
		nameNGramMapping,
	)
	packageMapping.AddFieldMappingsAt("Attribute", attributeFieldMapping, nameNGramMapping)
	packageMapping.AddFieldMappingsAt("Source", keywordFieldMapping)
	packageMapping.AddFieldMappingsAt("Description", descriptionFieldMapping)
	packageMapping.AddFieldMappingsAt("Homepages", keywordFieldMapping)
	packageMapping.AddFieldMappingsAt("MainProgram", identityFieldMapping)
	packageMapping.AddFieldMappingsAt("PackageSet", identityFieldMapping)
	packageMapping.AddFieldMappingsAt("Platforms", identityFieldMapping)
	packageMapping.AddFieldMappingsAt("Programs", identityFieldMapping)

	indexMapping.AddDocumentMapping("option", optionMapping)
	indexMapping.AddDocumentMapping("package", packageMapping)

	return indexMapping, nil
}

func createIndex(root *file.Root, kvconfig map[string]any) (bleve.Index, error) {
	indexMapping, err := createIndexMapping()
	if err != nil {
		return nil, err
	}

	//nolint:forbidigo // external package
	indexPath := root.JoinPath(indexBaseName)
	idx, baseErr := bleve.NewUsing(
		indexPath,
		indexMapping,
		bleve.Config.DefaultIndexType,
		bleve.Config.DefaultKVStore,
		kvconfig,
	)
	if baseErr != nil {
		return nil, fault.Wrap(baseErr, fmsg.Withf("unable to create index at path %s", indexPath))
	}

	return idx, nil
}

const (
	indexBaseName = "index.bleve"
)

var expectedDataFiles = []string{
	meta.BaseName,
	indexBaseName,
}

func deleteIndex(root *file.Root) error {
	for _, file := range expectedDataFiles {
		err := root.RemoveAll(file)
		if err != nil {
			return fault.Wrap(err, fmsg.Withf("could not remove file %s", file))
		}
	}

	return nil
}

func OpenOrCreate(options *Options) (*ReadIndex, *WriteIndex, error) {
	var err error
	bleve.SetLog(zap.NewStdLog(options.Logger.Named("bleve").GetLogger()))

	root := options.Root
	exists, err := root.Exists(indexBaseName)
	if err != nil {
		return nil, nil, fault.Wrap(
			err, fmsg.Withf("could not check if index exists at path %s", indexBaseName))
	}

	kvconfig := map[string]any{
		"unsafe_batch": true,
		"scorchPersisterOptions": map[string]any{
			"NumPersisterWorkers":           8,
			"MaxSizeInMemoryMergePerWorker": 128 * 1024 * 1024,
		},
	}
	if options.LowMemory {
		kvconfig = map[string]any{
			"PersisterNapTimeMSec":      1000,
			"PersisterNapUnderNumFiles": 500,
		}
	}

	var idx bleve.Index
	var m *meta.Meta
	if !exists || options.Force {
		if options.Force {
			err = deleteIndex(root)
			if err != nil {
				return nil, nil, err
			}
		}
		idx, err = createIndex(root, kvconfig)
		if err != nil {
			return nil, nil, err
		}

		m, err = meta.Create(root, options.Logger)
		if err != nil {
			return nil, nil, fault.Wrap(err, fmsg.With("could not create index metadata file"))
		}

	} else {
		var baseErr error
		//nolint:forbidigo // external package
		indexPath := root.JoinPath(indexBaseName)
		idx, baseErr = bleve.OpenUsing(indexPath, kvconfig)
		if baseErr != nil {
			return nil, nil, fault.Wrap(baseErr, fmsg.Withf("could not open index at path %s", indexPath))
		}

		m, err = meta.Open(root, options.Logger)
		if err != nil {
			return nil, nil, fault.Wrap(err, fmsg.With("could not open index metadata file"))
		}
	}

	if options.BatchSize == 0 {
		options.BatchSize = config.DefaultConfig.Importer.BatchSize
	}

	if options.LowMemory && options.BatchSize == config.DefaultConfig.Importer.BatchSize {
		options.BatchSize = 1_000
	}

	return &ReadIndex{
			config: options.Config,
			log:    options.Logger,
			exists: exists,
			index:  idx,
			meta:   m,
		},
		&WriteIndex{
			exists:    exists,
			index:     idx,
			batchSize: options.BatchSize,
			log:       options.Logger,
			Meta:      m,
		},
		nil
}

func (i *WriteIndex) Exists() bool {
	return i.exists
}

func (i *WriteIndex) SaveMeta() error {
	err := i.Meta.Save()
	if err != nil {
		return fault.Wrap(err)
	}

	return nil
}

func (i *WriteIndex) Import(
	ctx context.Context,
	objects <-chan nix.Importable,
) <-chan error {
	indexMapping := i.index.Mapping()

	return i.WithBatchObjects(ctx, objects, func(batch *bleve.Batch, obj nix.Importable) error {
		doc := document.NewDocument(nix.GetKey(obj))
		if err := indexMapping.MapDocument(doc, obj); err != nil {
			return fault.Wrap(err, fmsg.Withf("could not map document for object: %s", obj.GetName()))
		}

		var data bytes.Buffer
		enc := gob.NewEncoder(&data)
		if err := enc.Encode(&obj); err != nil {
			return fault.Wrap(err, fmsg.With("could not store object in search index"))
		}
		field := document.NewTextFieldWithIndexingOptions("_data", nil, data.Bytes(), index.StoreField)
		doc.AddField(field)
		idField := document.NewTextFieldCustom(
			"_id", nil, []byte(doc.ID()),
			index.IndexField|index.StoreField|index.IncludeTermVectors,
			idAnalyzer,
		)
		doc.AddField(idField)

		// log.Debug("adding object to index", "name", opt.Name)
		if err := batch.IndexAdvanced(doc); err != nil {
			return fault.Wrap(err, fmsg.Withf("could not index object %s", obj.GetName()))
		}

		return nil
	})
}

func (i *WriteIndex) GetBatchSize() int {
	return i.batchSize
}

func (i *WriteIndex) WithBatch(fn func(batch *bleve.Batch)) error {
	batch := i.index.NewBatch()
	fn(batch)

	err := i.Flush(batch)
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not flush batch"))
	}

	return nil
}

func (i *WriteIndex) WithBatchObjects(
	ctx context.Context,
	objects <-chan nix.Importable,
	processor func(batch *bleve.Batch, obj nix.Importable) error,
) <-chan error {
	var err error
	errs := make(chan error)

	go func() {
		defer close(errs)
		k := 0
		batch := i.index.NewBatch()

	outer:
		for obj := range objects {
			select {
			case <-ctx.Done():
				i.log.Warn("batch process aborted")

				break outer
			default:
			}

			if err := processor(batch, obj); err != nil {
				errs <- fault.Wrap(err, fmsg.With("could not process object"))

				continue
			}

			if k++; k%i.batchSize == 0 {
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

func (i *WriteIndex) Flush(batch *bleve.Batch) error {
	size := batch.Size()
	if size == 0 {
		return &BatchError{
			fault.New("no documents to flush"),
		}
	}
	i.log.Debug("flushing batch", "size", size)

	err := i.index.Batch(batch)
	if err != nil {
		return &BatchError{
			fault.Wrap(err, fmsg.Withf("could not flush batch")),
		}
	}

	batch.Reset()

	return nil
}

func (i *WriteIndex) Close() (err error) {
	if e := i.Meta.Save(); e != nil {
		// index needs to be closed anyway
		err = fault.Wrap(e, fmsg.With("could not save metadata"))
	}

	if e := i.index.Close(); e != nil {
		err = fault.Wrap(e, fmsg.Withf("could not close index"))
	}

	return err
}

func (i *WriteIndex) DeleteBySource(source string) error {
	query := bleve.NewTermQuery(source)
	search := bleve.NewSearchRequest(query)
	search.Size = math.MaxInt
	search.Fields = []string{"_id"}

	results, err := i.index.Search(search)
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("failed to query documents of retired index %s", source))
	}

	batch := i.index.NewBatch()
	var k int
	for _, hit := range results.Hits {
		batch.Delete(hit.ID)
		if k++; k%i.batchSize == 0 {
			err := i.Flush(batch)
			if err != nil {
				return err
			}
		}
	}
	err = i.Flush(batch)
	if err != nil {
		return fault.Wrap(err)
	}

	if uint64(search.Size) < results.Total {
		return i.DeleteBySource(source) // unlikely :^)
	}

	return nil
}
