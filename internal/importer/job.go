package importer

import (
	"context"
	"time"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/index"
)

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
