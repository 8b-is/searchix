package index

import (
	"context"
	"math"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/index/nixattr"
	"alin.ovh/searchix/internal/nix"
	"alin.ovh/searchix/internal/storage"

	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/web"
	"github.com/blevesearch/bleve/v2/analysis/token/camelcase"
	"github.com/blevesearch/bleve/v2/analysis/token/ngram"
	"github.com/blevesearch/bleve/v2/analysis/token/porter"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/letter"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/unicode"
	"github.com/blevesearch/bleve/v2/mapping"
	"go.uber.org/zap"
)

type Options struct {
	Force     bool
	LowMemory bool
	BatchSize int
	Store     *storage.Store
	Logger    *log.Logger
	Root      *file.Root
	Config    *config.Config
}

type WriteIndex struct {
	batchSize int
	index     bleve.Index
	log       *log.Logger
	store     *storage.Store
	exists    bool
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

	var idx bleve.Index
	var meta *Meta
	if !exists || options.Force {
		if options.Force {
			err = deleteIndex(root)
			if err != nil {
				return nil, nil, err
			}
		}
		idx, err = createIndex(root, options)
		if err != nil {
			return nil, nil, err
		}

		meta, err = createMeta(root, options.Logger)
		if err != nil {
			return nil, nil, err
		}

	} else {
		var baseErr error
		//nolint:forbidigo // external package
		indexPath := root.JoinPath(indexBaseName)
		idx, baseErr = bleve.Open(indexPath)
		if baseErr != nil {
			return nil, nil, fault.Wrap(baseErr, fmsg.Withf("could not open index at path %s", indexPath))
		}

		meta, err = openMeta(root, options.Logger)
		if err != nil {
			return nil, nil, err
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
			store:  options.Store,
			exists: exists,
			index:  idx,
			meta:   meta,
		},
		&WriteIndex{
			exists:    exists,
			index:     idx,
			batchSize: options.BatchSize,
			log:       options.Logger,
			store:     options.Store,
			Meta:      meta,
		},
		nil
}

func (i *WriteIndex) Exists() bool {
	return i.exists
}

func (i *WriteIndex) SaveMeta() error {
	return i.Meta.Save()
}

func (i *WriteIndex) Import(
	ctx context.Context,
	objects <-chan nix.Importable,
) <-chan error {
	return i.WithBatch(ctx, objects, func(batch *bleve.Batch, obj nix.Importable) error {
		if err := batch.Index(nix.GetKey(obj), obj); err != nil {
			return fault.Wrap(err, fmsg.Withf("could not index object %s", obj.GetName()))
		}

		return nil
	})
}

func (i *WriteIndex) WithBatch(
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
