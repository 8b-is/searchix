package importer

import (
	"context"
	"errors"
	"maps"
	"os/exec"
	"slices"
	"strings"
	"time"

	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/blevesearch/bleve/v2"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/fetcher"
	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/manpages"
	"alin.ovh/searchix/internal/programs"
)

type Options struct {
	LowMemory  bool
	Offline    bool
	Logger     *log.Logger
	ReadIndex  *index.ReadIndex
	WriteIndex *index.WriteIndex
	Manpages   *manpages.URLMap
	Root       *file.Root
}

type Importer struct {
	config  *config.Config
	options *Options
}

func New(
	cfg *config.Config,
	options *Options,
) (*Importer, error) {
	Job.LastRun = options.WriteIndex.Meta.LastImport

	return &Importer{
		config:  cfg,
		options: options,
	}, nil
}

func (imp *Importer) Fetch(
	ctx context.Context,
	forceUpdate bool,
	onlyUpdateSources []string,
) error {
	if len(imp.config.Importer.Sources) == 0 {
		imp.options.Logger.Info("No sources enabled")

		return nil
	}

	imp.options.Logger.Debug("starting importer", "timeout", imp.config.Importer.Timeout.Duration)
	importCtx, cancelImport := context.WithTimeout(
		ctx,
		imp.config.Importer.Timeout.Duration,
	)
	defer cancelImport()

	forceUpdate = forceUpdate || (len(onlyUpdateSources) > 0)

	meta := imp.options.WriteIndex.Meta

	importSource := imp.createSourceFetcher(importCtx, meta, forceUpdate)
	for name, source := range imp.config.Importer.Sources {
		if len(onlyUpdateSources) > 0 {
			if !slices.Contains(onlyUpdateSources, name) {
				continue
			}
		}
		err := importSource(source)
		if err != nil {
			imp.options.Logger.Error("import failed", "source", name, "error", err)
		}
	}

	return nil
}

func (imp *Importer) Index(ctx context.Context, onlyUpdateSources []string) error {
	if len(imp.config.Importer.Sources) == 0 {
		imp.options.Logger.Info("No sources enabled")

		return nil
	}

	imp.options.Logger.Debug("starting importer", "timeout", imp.config.Importer.Timeout.Duration)
	importCtx, cancelImport := context.WithTimeout(
		ctx,
		imp.config.Importer.Timeout.Duration,
	)
	defer cancelImport()

	meta := imp.options.WriteIndex.Meta

	importSource := imp.createSourceImporter(importCtx, meta)
	for name, source := range imp.config.Importer.Sources {
		if len(onlyUpdateSources) > 0 {
			if !slices.Contains(onlyUpdateSources, name) {
				continue
			}
		}
		err := importSource(source)
		if err != nil {
			imp.options.Logger.Error("import failed", "source", name, "error", err)
		}
	}

	err := meta.Save()
	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to save metadata"))
	}

	return nil
}

func (imp *Importer) EnsureSourcesIndexed(
	ctx context.Context,
	read *index.ReadIndex,
) error {
	cfgEnabledSources := slices.Collect(maps.Keys(imp.config.Importer.Sources))
	slices.Sort(cfgEnabledSources)

	indexedSources, err := read.GetEnabledSources()
	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to get enabled sources from index"))
	}
	slices.Sort(indexedSources)
	if !slices.Equal(cfgEnabledSources, indexedSources) {
		newSources := slices.DeleteFunc(slices.Clone(cfgEnabledSources), func(s string) bool {
			return slices.Contains(indexedSources, s)
		})
		retiredSources := slices.DeleteFunc(slices.Clone(indexedSources), func(s string) bool {
			return slices.Contains(cfgEnabledSources, s)
		})
		if len(newSources) > 0 {
			imp.options.Logger.Info("adding new sources", "sources", newSources)
			err := imp.Fetch(
				ctx,
				false,
				newSources,
			)
			if err != nil {
				return fault.Wrap(err, fmsg.With("Failed to update index with new sources"))
			}
		}
		if len(retiredSources) > 0 {
			imp.options.Logger.Info("removing retired sources", "sources", retiredSources)
			for _, s := range retiredSources {
				err := imp.options.WriteIndex.DeleteBySource(s)
				if err != nil {
					return fault.Wrap(err, fmsg.Withf("Failed to remove retired source %s", s))
				}
			}
		}
	}

	return nil
}

func (imp *Importer) Prune(ctx context.Context) error {
	for _, source := range imp.config.Importer.Sources {
		err := imp.PruneSource(ctx, source)
		if err != nil {
			imp.options.Logger.Error("Failed to prune source", "source", source.Key, "error", err)
		}
	}

	return nil
}

func (imp *Importer) PruneSource(
	_ context.Context,
	source config.Source,
) error {
	read := imp.options.ReadIndex
	write := imp.options.WriteIndex

	if read == nil {
		imp.options.Logger.DPanic("read index is not available")
	}

	cutoff := write.Meta.LastImport.StartedAt
	imp.options.Logger.Debug("searching for old entities", "cutoff", cutoff.Format(time.RFC3339))

	maxCount, err := read.Count(source)
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to retrieve entities for pruning"))
	}

	res, err := read.ImportedBefore(cutoff, source)
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to retrieve entities for pruning"))
	}

	if res.Total == 0 {
		imp.options.Logger.Debug("no old entities found")

		return nil
	}

	imp.options.Logger.Debug("found old entities", "count", res.Total)
	if float64(res.Total) > (0.9 * float64(maxCount)) {
		return fault.Newf("too many entities to prune: %d/%d (threshold: 90%%)", res.Total, maxCount)
	}

	err = write.WithBatch(func(batch *bleve.Batch) {
		for _, dm := range res.Hits {
			batch.Delete(dm.ID)
		}
	})
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to prune entities"))
	}

	imp.options.Logger.Info(
		"pruned old entities",
		"type",
		source.Importer.String(),
		"source",
		source.Key,
		"count",
		res.Total,
	)

	return nil
}

func (imp *Importer) createSourceFetcher(
	parent context.Context,
	meta *index.Meta,
	forceUpdate bool,
) func(config.Source) error {
	return func(source config.Source) error {
		logger := imp.options.Logger.With("name", source.Key)
		pdb, err := programs.New(source, &programs.Options{
			Logger: logger,
			Root:   imp.options.Root,
		})
		if err != nil {
			return fault.Wrap(err, fmsg.With("error creating program database"))
		}

		sourceMeta := meta.GetSourceMeta(source.Key)
		previousUpdate := sourceMeta.UpdatedAt

		fopts := &fetcher.Options{
			Logger: logger,
			Root:   imp.options.Root,
		}

		ctx, cancel := context.WithTimeout(parent, source.Timeout.Duration)
		defer cancel()

		logger.Debug("starting fetcher")

		fetcher, err := fetcher.New(source, fopts)
		if err != nil {
			return fault.Wrap(err, fmsg.With("error creating fetcher"))
		}

		_, err = fetcher.FetchIfNeeded(ctx, sourceMeta)
		if err != nil {
			var exerr *exec.ExitError
			if errors.As(err, &exerr) {
				lines := strings.SplitSeq(strings.TrimSpace(string(exerr.Stderr)), "\n")
				for line := range lines {
					logger.Error(
						"importer fetch failed",
						"fetcher",
						source.Fetcher.String(),
						"stderr",
						line,
						"status",
						exerr.ExitCode(),
					)
				}
			}

			return fault.Wrap(err, fmsg.With("importer fetch failed"))
		}
		logger.Info(
			"importer fetch succeeded",
			"previous",
			previousUpdate.Format(time.DateTime),
			"current",
			sourceMeta.UpdatedAt.Format(time.DateTime),
			"is_updated",
			sourceMeta.UpdatedAt.After(previousUpdate),
			"update_force",
			forceUpdate,
		)

		if source.Programs.Enable {
			err = pdb.Instantiate(ctx)
			if err != nil {
				logger.Warn("programs database instantiation failed", "error", err)
			}
		}

		if source.Manpages.Enable {
			err = imp.options.Manpages.Update(ctx, source)
			if err != nil {
				logger.Warn("manpages database update failed", "error", err)
			}
		}

		return nil
	}
}

func (imp *Importer) createSourceImporter(
	parent context.Context,
	meta *index.Meta,
) func(config.Source) error {
	return func(source config.Source) error {
		logger := imp.options.Logger.With("name", source.Key)
		pdb, err := programs.New(source, &programs.Options{
			Logger: logger,
			Root:   imp.options.Root,
		})
		if err != nil {
			return fault.Wrap(err, fmsg.With("error creating program database"))
		}

		sourceMeta := meta.GetSourceMeta(source.Key)

		fopts := &fetcher.Options{
			Logger: logger,
			Root:   imp.options.Root,
		}

		ctx, cancel := context.WithTimeout(parent, source.Timeout.Duration)
		defer cancel()

		files, err := fetcher.Open(source, fopts)
		if err != nil {
			return fault.Wrap(err, fmsg.With("error opening fetched files"))
		}

		if source.Programs.Enable {
			err = pdb.Instantiate(ctx)
			if err != nil {
				logger.Warn("programs database instantiation failed", "error", err)
			}
		}

		if files.Revision != nil {
			err = setRepoRevision(files.Revision, source)
			if err != nil {
				logger.Warn("could not set source repo revision", "error", err)
			}
		}

		var processor Processor
		logger.Debug(
			"creating processor",
			"importer_type",
			source.Importer,
			"revision",
			source.Repo.Revision,
		)
		switch source.Importer {
		case config.Options:
			processor, err = NewOptionProcessor(
				files.Options,
				source,
				logger.Named("processor"),
			)
		case config.Packages:
			processor, err = NewPackageProcessor(
				files.Packages,
				source,
				logger.Named("processor"),
				pdb,
			)
		}
		if err != nil {
			return fault.Wrap(err, fmsg.Withf("failed to create processor"))
		}

		hadWarnings, err := imp.process(ctx, processor)
		if err != nil {
			return fault.Wrap(err, fmsg.Withf("failed to process source"))
		}

		sourceMeta.StoredAt = time.Now()

		if hadWarnings {
			logger.Warn("importer succeeded, but with warnings/errors")
		} else {
			logger.Info("importer succeeded")
		}

		sourceMeta.Rev = source.Repo.Revision
		meta.SetSourceMeta(source.Key, sourceMeta)

		return nil
	}
}
