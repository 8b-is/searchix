package index

import (
	"bytes"
	"context"
	"encoding/gob"
	"math"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/index/nixattr"
	"alin.ovh/searchix/internal/nix"
	"alin.ovh/x/log"
	"go.uber.org/zap"

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
)

var idAnalyzer analysis.Analyzer

type Options struct {
	LowMemory bool
	BatchSize int
	Logger    *log.Logger
}

type WriteIndex struct {
	batchSize int
	index     bleve.Index
	log       *log.Logger
	Meta      *Meta
}

type BatchError struct {
	error
}

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

	attributeFieldMapping := bleve.NewKeywordFieldMapping()
	attributeFieldMapping.Analyzer = "dotted_keyword"

	keywordFieldMapping := bleve.NewKeywordFieldMapping()
	keywordFieldMapping.Analyzer = simple.Name

	nameNGramMapping := bleve.NewTextFieldMapping()
	nameNGramMapping.Analyzer = "c_name"
	nameNGramMapping.IncludeTermVectors = true
	nameNGramMapping.Store = false

	nixValueMapping := bleve.NewDocumentStaticMapping()
	nixValueMapping.AddFieldMappingsAt("Text", textFieldMapping)
	nixValueMapping.AddFieldMappingsAt("Markdown", textFieldMapping)

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
	optionMapping.AddFieldMappingsAt("RelatedPackages", textFieldMapping)
	optionMapping.AddFieldMappingsAt("Description", descriptionFieldMapping)

	optionMapping.AddSubDocumentMapping("Default", nixValueMapping)
	optionMapping.AddSubDocumentMapping("Example", nixValueMapping)

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

	idAnalyzer = indexMapping.AnalyzerNamed("c_name")

	return indexMapping, nil
}

func createIndex(root *file.Root, options *Options) (bleve.Index, error) {
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
	metaBaseName  = "meta.json"
)

var expectedDataFiles = []string{
	metaBaseName,
	indexBaseName,
	"manpage-urls.json",
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

func OpenOrCreate(
	root *file.Root,
	force bool,
	options *Options,
) (*ReadIndex, *WriteIndex, bool, error) {
	var err error
	bleve.SetLog(zap.NewStdLog(options.Logger.Named("bleve").GetLogger()))

	exists, err := root.Exists(indexBaseName)
	if err != nil {
		return nil, nil, exists, fault.Wrap(
			err, fmsg.Withf("could not check if index exists at path %s", indexBaseName))
	}

	var idx bleve.Index
	var meta *Meta
	if !exists || force {
		if force {
			err = deleteIndex(root)
			if err != nil {
				return nil, nil, false, err
			}
		}
		idx, err = createIndex(root, options)
		if err != nil {
			return nil, nil, false, err
		}

		meta, err = createMeta(root, options.Logger)
		if err != nil {
			return nil, nil, false, err
		}

	} else {
		var baseErr error
		//nolint:forbidigo // external package
		indexPath := root.JoinPath(indexBaseName)
		idx, baseErr = bleve.Open(indexPath)
		if baseErr != nil {
			return nil, nil, exists, fault.Wrap(baseErr, fmsg.Withf("could not open index at path %s", indexPath))
		}

		meta, err = openMeta(root, options.Logger)
		if err != nil {
			return nil, nil, exists, err
		}
	}

	if options.BatchSize == 0 {
		options.BatchSize = config.DefaultConfig.Importer.BatchSize
	}

	if options.LowMemory && options.BatchSize == config.DefaultConfig.Importer.BatchSize {
		options.BatchSize = 1_000
	}

	return &ReadIndex{
			index: idx,
			log:   options.Logger,
			meta:  meta,
		},
		&WriteIndex{
			index:     idx,
			batchSize: options.BatchSize,
			log:       options.Logger,
			Meta:      meta,
		},
		exists,
		nil
}

func (i *WriteIndex) SaveMeta() error {
	return i.Meta.Save()
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
				i.log.Warn("import aborted")

				break outer
			default:
			}

			doc := document.NewDocument(nix.GetKey(obj))
			if err := indexMapping.MapDocument(doc, obj); err != nil {
				errs <- fault.Wrap(err, fmsg.Withf("could not map document for object: %s", obj.GetName()))

				continue
			}

			var data bytes.Buffer
			enc := gob.NewEncoder(&data)
			if err := enc.Encode(&obj); err != nil {
				errs <- fault.Wrap(err, fmsg.With("could not store object in search index"))

				continue
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
				errs <- fault.Wrap(err, fmsg.Withf("could not index object %s", obj.GetName()))

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
