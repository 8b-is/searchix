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

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/fetcher"
	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/manpages"
	"alin.ovh/searchix/internal/programs"
	"alin.ovh/searchix/internal/storage"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
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

var Job struct {
	StartedAt time.Time
	LastRun   struct {
		StartedAt  time.Time
		FinishedAt time.Time
	}
	NextRun time.Time
}

func MarkImportStarted() {
	Job.StartedAt = time.Now()
}

func MarkImportFinished() {
	Job.LastRun.StartedAt = Job.StartedAt
	Job.LastRun.FinishedAt = time.Now()
	Job.StartedAt = time.Time{}
}

func MarkLastFetch(meta *index.Meta) {
	meta.LastImport = Job.LastRun
}

func SetNextRun(nextRun time.Time) {
	Job.NextRun = nextRun
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

type Importer struct {
	config  *config.Config
	options *Options
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

	MarkLastFetch(meta)
	err := imp.options.WriteIndex.SaveMeta()
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to save metadata"))
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

func (imp *Importer) StartUpdateTimer(
	parentCtx context.Context,
) {
	var nextRun time.Time
	switch {
	case Job.LastRun.FinishedAt.Before(time.Now().Add(-24 * time.Hour)):
		imp.options.Logger.Info(
			"indexing last ran more than 24 hours ago, scheduling immediate update",
		)
		nextRun = time.Now()
	case imp.options.WriteIndex.Meta.IsSchemaOutdated():
		imp.options.Logger.Info(
			"indexing schema version is out of date, scheduling immediate update",
		)
		nextRun = time.Now()
	default:
		nextRun = nextUTCOccurrenceOfTime(imp.config.Importer.UpdateAt)
	}
	SetNextRun(nextRun)
	for {
		select {
		case <-parentCtx.Done():
			imp.options.Logger.Debug("stopping scheduler")

			return
		case <-time.After(time.Until(nextRun)):
		}
		imp.options.Logger.Info("updating index")

		MarkImportStarted()
		ctx, cancel := context.WithTimeout(parentCtx, imp.config.Importer.Timeout.Duration)

		err := imp.Fetch(ctx, false, false, nil)
		if err != nil {
			imp.options.Logger.Warn("error fetching update", "error", err)
		} else {
			imp.options.Logger.Info("update complete")
		}

		err = imp.Index(ctx)
		if err != nil {
			imp.options.Logger.Warn("error indexing update", "error", err)
		} else {
			imp.options.Logger.Info("indexing complete")
		}

		cancel()
		MarkImportFinished()

		nextRun = nextUTCOccurrenceOfTime(imp.config.Importer.UpdateAt)
		SetNextRun(nextRun)

		imp.options.Logger.Info(
			"scheduling next run",
			"next-run",
			nextRun.Format(time.DateTime),
		)
	}
}

func nextUTCOccurrenceOfTime(dayTime config.LocalTime) time.Time {
	now := time.Now()
	nextRun := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		dayTime.Hour,
		dayTime.Minute,
		dayTime.Second,
		0,
		time.UTC,
	)
	if nextRun.Before(now) {
		return nextRun.AddDate(0, 0, 1)
	}

	return nextRun
}
