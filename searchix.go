package searchix

import (
	"context"
	"log"
	"log/slog"
	"slices"
	"sync"
	"time"

	"searchix/internal/config"
	"searchix/internal/importer"
	"searchix/internal/index"
	"searchix/internal/server"

	"github.com/getsentry/sentry-go"
	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
)

func nextOccurrenceOfLocalTime(t toml.LocalTime) time.Time {
	now := time.Now()
	dayTime := t
	nextRun := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		dayTime.Hour,
		dayTime.Minute,
		dayTime.Second,
		0,
		time.Local,
	)
	if nextRun.Before(now) {
		return nextRun.AddDate(0, 0, 1)
	}

	return nextRun
}

func (s *Server) SetupIndex(update bool, replace bool) error {
	var i uint
	cfgEnabledSources := make([]string, len(s.cfg.Importer.Sources))
	for key := range s.cfg.Importer.Sources {
		cfgEnabledSources[i] = key
		i++
	}
	slices.Sort(cfgEnabledSources)

	read, write, exists, err := index.OpenOrCreate(s.cfg.DataPath, replace)
	if err != nil {
		return errors.Wrap(err, "Failed to open or create index")
	}
	s.readIndex = read
	s.writeIndex = write

	if !exists || replace || update {
		slog.Info(
			"Starting build job",
			"new",
			!exists,
			"replace",
			replace,
			"update",
			update,
		)
		err = importer.Start(s.cfg, write, replace, nil)
		if err != nil {
			return errors.Wrap(err, "Failed to build index")
		}
		if replace || update {
			return nil
		}
	} else {
		indexedSources, err := read.GetEnabledSources()
		if err != nil {
			return errors.Wrap(err, "Failed to get enabled sources from index")
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
				slog.Info("adding new sources", "sources", newSources)
				err := importer.Start(s.cfg, write, false, &newSources)
				if err != nil {
					return errors.Wrap(err, "Failed to update index with new sources")
				}
			}
			if len(retiredSources) > 0 {
				slog.Info("removing retired sources", "sources", retiredSources)
				for _, s := range retiredSources {
					err := write.DeleteBySource(s)
					if err != nil {
						return errors.Wrapf(err, "Failed to remove retired source %s", s)
					}
				}
			}
		}
	}

	return nil
}

type Server struct {
	sv         *server.Server
	wg         *sync.WaitGroup
	cfg        *config.Config
	sentryHub  *sentry.Hub
	readIndex  *index.ReadIndex
	writeIndex *index.WriteIndex
}

func New(cfg *config.Config) (*Server, error) {
	slog.SetLogLoggerLevel(cfg.LogLevel)
	if cfg.Web.Environment == "production" {
		log.SetFlags(0)
	} else {
		log.SetFlags(log.LstdFlags)
	}

	err := sentry.Init(sentry.ClientOptions{
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		Dsn:              cfg.Web.SentryDSN,
		Environment:      cfg.Web.Environment,
	})
	if err != nil {
		slog.Warn("could not initialise sentry", "error", err)
	}

	return &Server{
		cfg:       cfg,
		sentryHub: sentry.CurrentHub(),
	}, nil
}

func (s *Server) startUpdateTimer(
	ctx context.Context,
	localHub *sentry.Hub,
) {
	const monitorSlug = "import"
	localHub.WithScope(func(scope *sentry.Scope) {
		var err error
		scope.SetContext("monitor", sentry.Context{"slug": monitorSlug})
		monitorConfig := &sentry.MonitorConfig{
			Schedule: sentry.IntervalSchedule(1, sentry.MonitorScheduleUnitDay),
			// minutes
			MaxRuntime:    10,
			CheckInMargin: 5,
			Timezone:      time.Local.String(),
		}

		s.wg.Add(1)
		nextRun := nextOccurrenceOfLocalTime(s.cfg.Importer.UpdateAt.LocalTime)
		for {
			slog.Debug("scheduling next run", "next-run", nextRun)
			select {
			case <-ctx.Done():
				slog.Debug("stopping scheduler")
				s.wg.Done()

				return
			case <-time.After(time.Until(nextRun)):
			}
			s.wg.Add(1)
			slog.Info("updating index")

			eventID := localHub.CaptureCheckIn(&sentry.CheckIn{
				MonitorSlug: monitorSlug,
				Status:      sentry.CheckInStatusInProgress,
			}, monitorConfig)

			err = importer.Start(s.cfg, s.writeIndex, false, nil)
			s.wg.Done()
			if err != nil {
				slog.Warn("error updating index", "error", err)

				localHub.CaptureException(err)
				localHub.CaptureCheckIn(&sentry.CheckIn{
					ID:          *eventID,
					MonitorSlug: monitorSlug,
					Status:      sentry.CheckInStatusError,
				}, monitorConfig)
			} else {
				slog.Info("update complete")

				localHub.CaptureCheckIn(&sentry.CheckIn{
					ID:          *eventID,
					MonitorSlug: monitorSlug,
					Status:      sentry.CheckInStatusOK,
				}, monitorConfig)
			}
			nextRun = nextRun.AddDate(0, 0, 1)
		}
	})
}

func (s *Server) Start(ctx context.Context, liveReload bool) error {
	var err error
	s.sv, err = server.New(s.cfg, s.readIndex, liveReload)
	if err != nil {
		return errors.Wrap(err, "error setting up server")
	}

	s.wg = &sync.WaitGroup{}
	go s.startUpdateTimer(ctx, sentry.CurrentHub().Clone())

	s.wg.Add(1)
	err = s.sv.Start()
	if err != nil {
		s.wg.Done()

		return errors.Wrap(err, "error starting server")
	}

	return nil
}

func (s *Server) Stop() {
	<-s.sv.Stop()
	defer s.wg.Done()
	s.sentryHub.Flush(2 * time.Second)
}
