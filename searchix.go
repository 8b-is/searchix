package searchix

import (
	"context"
	"math"
	"slices"
	"sync"
	"time"

	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/importer"
	"go.alanpearce.eu/searchix/internal/index"
	"go.alanpearce.eu/searchix/internal/server"
	"go.alanpearce.eu/x/log"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

func nextUTCOccurrenceOfTime(t config.LocalTime) time.Time {
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
		time.UTC,
	)
	if nextRun.Before(now) {
		return nextRun.AddDate(0, 0, 1)
	}

	return nextRun
}

type IndexOptions struct {
	Update    bool
	Replace   bool
	LowMemory bool
	Logger    *log.Logger
}

func (s *Server) SetupIndex(ctx context.Context, options *IndexOptions) error {
	var i uint
	cfgEnabledSources := make([]string, len(s.cfg.Importer.Sources))
	for key := range s.cfg.Importer.Sources {
		cfgEnabledSources[i] = key
		i++
	}
	slices.Sort(cfgEnabledSources)

	read, write, exists, err := index.OpenOrCreate(
		s.cfg.DataPath,
		options.Replace,
		&index.Options{
			LowMemory: options.LowMemory,
			Logger:    options.Logger.Named("index"),
		},
	)
	if err != nil {
		return errors.Wrap(err, "Failed to open or create index")
	}
	s.readIndex = read
	s.writeIndex = write

	if !exists || options.Replace || options.Update {
		s.log.Info(
			"Starting build job",
			"new",
			!exists,
			"replace",
			options.Replace,
			"update",
			options.Update,
		)
		err = importer.Start(
			ctx,
			s.cfg,
			s.log.Named("importer"),
			write,
			options.Replace || options.Update,
			nil,
		)
		if err != nil {
			return errors.Wrap(err, "Failed to build index")
		}
		if options.Replace || options.Update {
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
				s.log.Info("adding new sources", "sources", newSources)
				err := importer.Start(
					ctx,
					s.cfg,
					options.Logger.Named("importer"),
					write,
					false,
					&newSources,
				)
				if err != nil {
					return errors.Wrap(err, "Failed to update index with new sources")
				}
			}
			if len(retiredSources) > 0 {
				s.log.Info("removing retired sources", "sources", retiredSources)
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
	log        *log.Logger
	sentryHub  *sentry.Hub
	readIndex  *index.ReadIndex
	writeIndex *index.WriteIndex
}

func New(cfg *config.Config, log *log.Logger) (*Server, error) {
	err := sentry.Init(sentry.ClientOptions{
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		Dsn:              cfg.Web.SentryDSN,
		Environment:      cfg.Web.Environment,
	})
	if err != nil {
		log.Warn("could not initialise sentry", "error", err)
	}

	return &Server{
		cfg:       cfg,
		log:       log,
		sentryHub: sentry.CurrentHub(),
	}, nil
}

func (s *Server) startUpdateTimer(
	ctx context.Context,
	cfg *config.Config,
	localHub *sentry.Hub,
) {
	const monitorSlug = "import"
	localHub.WithScope(func(scope *sentry.Scope) {
		var err error
		scope.SetContext("monitor", sentry.Context{"slug": monitorSlug})
		monitorConfig := &sentry.MonitorConfig{
			Schedule:      sentry.IntervalSchedule(1, sentry.MonitorScheduleUnitDay),
			MaxRuntime:    int64(math.Ceil(cfg.Importer.Timeout.Minutes())),
			CheckInMargin: 5,
			Timezone:      time.Local.String(),
		}

		s.wg.Add(1)
		nextRun := nextUTCOccurrenceOfTime(s.cfg.Importer.UpdateAt)
		for {
			s.log.Debug("scheduling next run", "next-run", nextRun)
			select {
			case <-ctx.Done():
				s.log.Debug("stopping scheduler")
				s.wg.Done()

				return
			case <-time.After(time.Until(nextRun)):
			}
			s.wg.Add(1)
			s.log.Info("updating index")

			eventID := localHub.CaptureCheckIn(&sentry.CheckIn{
				MonitorSlug: monitorSlug,
				Status:      sentry.CheckInStatusInProgress,
			}, monitorConfig)

			err = importer.Start(ctx, s.cfg, s.log.Named("importer"), s.writeIndex, false, nil)
			s.wg.Done()
			if err != nil {
				s.log.Warn("error updating index", "error", err)

				localHub.CaptureException(err)
				localHub.CaptureCheckIn(&sentry.CheckIn{
					ID:          *eventID,
					MonitorSlug: monitorSlug,
					Status:      sentry.CheckInStatusError,
				}, monitorConfig)
			} else {
				s.log.Info("update complete")

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
	s.sv, err = server.New(s.cfg, s.readIndex, s.log.Named("server"), liveReload)
	if err != nil {
		return errors.Wrap(err, "error setting up server")
	}

	s.wg = &sync.WaitGroup{}
	go s.startUpdateTimer(ctx, s.cfg, sentry.CurrentHub().Clone())

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
