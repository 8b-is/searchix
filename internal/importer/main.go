package importer

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math"
	"os/exec"
	"slices"
	"strings"
	"time"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/fetcher"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/manpages"
	"alin.ovh/searchix/internal/programs"
	"alin.ovh/x/log"
	"github.com/getsentry/sentry-go"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
)

type Options struct {
	LowMemory  bool
	Logger     *log.Logger
	WriteIndex *index.WriteIndex
	Manpages   *manpages.URLMap
}

var Job struct {
	InProgress bool
	StartedAt  time.Time
	FinishedAt time.Time
	NextRun    time.Time
}

func SetNextRun(nextRun time.Time) {
	Job.NextRun = nextRun
}

func SetLastUpdated(last time.Time) {
	Job.FinishedAt = last
}

func MarkIndexingStarted() {
	Job.StartedAt = time.Now()
	Job.InProgress = true
}

func MarkIndexingFinished(nextRun time.Time) {
	Job.FinishedAt = time.Now()
	Job.InProgress = false
	Job.NextRun = nextRun
}

func (imp *Importer) createSourceImporter(
	parent context.Context,
	meta *index.Meta,
	forceUpdate bool,
) func(*config.Source) error {
	return func(source *config.Source) error {
		logger := imp.options.Logger.With("name", source.Key)
		logger.Debug("starting fetcher")

		fetcher, err := fetcher.New(source, logger)
		if err != nil {
			return fault.Wrap(err, fmsg.With("error creating fetcher"))
		}

		sourceMeta := meta.GetSourceMeta(source.Key)
		if forceUpdate {
			sourceMeta.Updated = time.Unix(0, 0)
		}
		previousUpdate := sourceMeta.Updated
		ctx, cancel := context.WithTimeout(parent, source.Timeout.Duration)
		defer cancel()
		files, err := fetcher.FetchIfNeeded(ctx, sourceMeta)
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
			previousUpdate,
			"current",
			sourceMeta.Updated.Format(time.DateTime),
			"is_updated",
			sourceMeta.Updated.After(previousUpdate),
			"update_force",
			forceUpdate,
		)

		if sourceMeta.Updated.After(previousUpdate) || forceUpdate {
			var pdb *programs.DB

			if source.Programs.Enable {
				pdb, err = programs.Instantiate(ctx, source, imp.options.Logger.Named("programs"))
				if err != nil {
					logger.Warn("programs database instantiation failed", "error", err)
				}
			}

			err = setRepoRevision(files.Revision, source)
			if err != nil {
				logger.Warn("could not set source repo revision", "error", err)
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
				logger.Debug("processor created", "file", fmt.Sprintf("%T", files.Options))
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

			hadWarnings, err := process(ctx, imp.options.WriteIndex, processor, logger)
			if err != nil {
				return fault.Wrap(err, fmsg.Withf("failed to process source"))
			}

			if source.Manpages.Enable {
				err = imp.options.Manpages.Update(ctx, source)
				if err != nil {
					logger.Warn("manpages database update failed", "error", err)
				}
			}

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

func (imp *Importer) Start(
	ctx context.Context,
	forceUpdate bool,
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

	importSource := imp.createSourceImporter(importCtx, meta, forceUpdate)
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

	err := imp.options.WriteIndex.SaveMeta()
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to save metadata"))
	}

	SetLastUpdated(time.Now())

	return nil
}

func New(
	cfg *config.Config,
	options *Options,
) (*Importer, error) {
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
			err := imp.Start(
				ctx,
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
	localHub *sentry.Hub,
) {
	const monitorSlug = "import"
	localHub.WithScope(func(scope *sentry.Scope) {
		var err error
		scope.SetContext("monitor", sentry.Context{"slug": monitorSlug})
		monitorConfig := &sentry.MonitorConfig{
			Schedule:      sentry.IntervalSchedule(1, sentry.MonitorScheduleUnitDay),
			MaxRuntime:    int64(math.Ceil(imp.config.Importer.Timeout.Minutes())),
			CheckInMargin: 5,
			Timezone:      time.Local.String(),
		}

		Job.FinishedAt = imp.options.WriteIndex.Meta.LastUpdated()

		var nextRun time.Time
		switch {
		case Job.FinishedAt.Before(time.Now().Add(-24 * time.Hour)):
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

			eventID := localHub.CaptureCheckIn(&sentry.CheckIn{
				MonitorSlug: monitorSlug,
				Status:      sentry.CheckInStatusInProgress,
			}, monitorConfig)
			MarkIndexingStarted()

			ctx, cancel := context.WithTimeout(parentCtx, imp.config.Importer.Timeout.Duration)
			err = imp.Start(ctx, false, nil)
			cancel()

			if err != nil {
				imp.options.Logger.Warn("error updating index", "error", err)

				localHub.CaptureException(err)
				localHub.CaptureCheckIn(&sentry.CheckIn{
					ID:          *eventID,
					MonitorSlug: monitorSlug,
					Status:      sentry.CheckInStatusError,
				}, monitorConfig)
			} else {
				imp.options.Logger.Info("update complete")

				localHub.CaptureCheckIn(&sentry.CheckIn{
					ID:          *eventID,
					MonitorSlug: monitorSlug,
					Status:      sentry.CheckInStatusOK,
				}, monitorConfig)
			}
			nextRun = nextUTCOccurrenceOfTime(imp.config.Importer.UpdateAt)
			MarkIndexingFinished(nextRun)
			imp.options.Logger.Info(
				"scheduling next run",
				"next-run",
				nextRun.Format(time.DateTime),
			)
		}
	})
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
