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
	"github.com/asdine/storm/v3/q"
	"github.com/blevesearch/bleve/v2"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/fetcher"
	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/manpages"
	"alin.ovh/searchix/internal/nix"
	"alin.ovh/searchix/internal/programs"
	"alin.ovh/searchix/internal/storage"
)

type Options struct {
	LowMemory  bool
	Offline    bool
	Logger     *log.Logger
	WriteIndex *index.WriteIndex
	Manpages   *manpages.URLMap
	Root       *file.Root
	Storage    *storage.Store
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
	fetchOnly bool,
	onlyUpdateSources *[]string,
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

	forceUpdate = forceUpdate || (onlyUpdateSources != nil && len(*onlyUpdateSources) > 0)

	meta := imp.options.WriteIndex.Meta

	importSource := imp.createSourceImporter(importCtx, meta, forceUpdate, fetchOnly)
	for name, source := range imp.config.Importer.Sources {
		if onlyUpdateSources != nil && len(*onlyUpdateSources) > 0 {
			if !slices.Contains(*onlyUpdateSources, name) {
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

func (imp *Importer) Index(ctx context.Context) error {
	for name, source := range imp.config.Importer.Sources {
		hadErrors, err := imp.indexSource(ctx, source)
		if err != nil {
			return fault.Wrap(err, fmsg.Withf("Failed to import source %s", name))
		}

		if hadErrors {
			imp.options.Logger.Warn("Imported source encountered errors", "source", source.Name)
		}
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
				false,
				&newSources,
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

func (imp *Importer) PruneSource(ctx context.Context, source *config.Source) error {
	store := imp.options.Storage
	write := imp.options.WriteIndex

	tx, err := store.WithBatch(true).From(source.Key).Begin(true)
	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to begin transaction"))
	}
	defer tx.Rollback()

	cutoff := write.Meta.LastImport.StartedAt
	imp.options.Logger.Debug("searching for old entities", "cutoff", cutoff.Format(time.RFC3339))

	query := tx.Select(q.Lt("ImportedAt", cutoff))

	var obj nix.Importable
	switch source.Importer {
	case config.Options:
		obj = new(nix.Option)
	case config.Packages:
		obj = new(nix.Package)
	}
	count, err := query.Count(obj)
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to retrieve entities for pruning"))
	}

	imp.options.Logger.Debug(
		"pruning",
		"type",
		source.Importer.String(),
		"source",
		source.Key,
		"count",
		count,
	)

	if count == 0 {
		return nil
	}

	objs := make(chan nix.Importable, 1)
	errs := write.WithBatch(ctx, objs, func(batch *bleve.Batch, obj nix.Importable) error {
		batch.Delete(obj.GetName())

		return nil
	})

	go func() {
		for err := range errs {
			imp.options.Logger.Error("failed to prune old entities", "error", err)
		}
	}()

	err = query.Each(obj, func(record any) error {
		objs <- record.(nix.Importable)

		return nil
	})
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to prune old entities from index"))
	}

	err = query.Delete(obj)
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to prune old entities from storage"))
	}

	err = tx.Commit()
	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to commit transaction"))
	}

	return nil
}

func (imp *Importer) createSourceImporter(
	parent context.Context,
	meta *index.Meta,
	forceUpdate bool,
	fetchOnly bool,
) func(*config.Source) error {
	return func(source *config.Source) error {
		var files *fetcher.FetchedFiles

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

		if imp.options.Offline {
			logger.Debug("skipping fetch; in offline mode")

			files, err = fetcher.Open(source, fopts)
			if err != nil {
				return fault.Wrap(err, fmsg.With("error opening fetched files"))
			}
		} else {
			logger.Debug("starting fetcher")

			fetcher, err := fetcher.New(source, fopts)
			if err != nil {
				return fault.Wrap(err, fmsg.With("error creating fetcher"))
			}

			files, err = fetcher.FetchIfNeeded(ctx, sourceMeta)
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
				"fetch_only",
				fetchOnly,
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
		}

		if !fetchOnly &&
			(!sourceMeta.UpdatedAt.After(sourceMeta.StoredAt) || sourceMeta.StoredAt.IsZero() || forceUpdate) {

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

			hadWarnings, err := pipe(
				ctx,
				logger.Named("importer"),
				processor.Process,
				imp.options.Storage.MakeSourceImporter(source),
			)
			if err != nil {
				return fault.Wrap(err, fmsg.Withf("failed to process source"))
			}

			sourceMeta.StoredAt = time.Now()

			if hadWarnings {
				logger.Warn("importer succeeded, but with warnings/errors")
			} else {
				logger.Info("importer succeeded")
			}
		}

		sourceMeta.Rev = source.Repo.Revision
		meta.SetSourceMeta(source.Key, sourceMeta)

		return nil
	}
}
