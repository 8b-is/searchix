package web

import (
	"context"
	"math"
	"sync"
	"time"

	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/importer"
	"go.alanpearce.eu/searchix/internal/index"
	"go.alanpearce.eu/searchix/internal/server"
	"go.alanpearce.eu/x/log"

	"github.com/getsentry/sentry-go"
	"gitlab.com/tozd/go/errors"
)

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

type Server struct {
	sv         *server.Server
	wg         *sync.WaitGroup
	cfg        *config.Config
	log        *log.Logger
	sentryHub  *sentry.Hub
	readIndex  *index.ReadIndex
	writeIndex *index.WriteIndex
}

func New(cfg *config.Config, log *log.Logger) (*Server, errors.E) {
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
		var err errors.E
		scope.SetContext("monitor", sentry.Context{"slug": monitorSlug})
		monitorConfig := &sentry.MonitorConfig{
			Schedule:      sentry.IntervalSchedule(1, sentry.MonitorScheduleUnitDay),
			MaxRuntime:    int64(math.Ceil(cfg.Importer.Timeout.Minutes())),
			CheckInMargin: 5,
			Timezone:      time.Local.String(),
		}

		s.wg.Add(1)
		nextRun := nextUTCOccurrenceOfTime(s.cfg.Importer.UpdateAt)
		importer.SetNextRun(nextRun)
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
			importer.MarkIndexingStarted()

			imp := importer.New(s.cfg, s.log.Named("importer"), s.writeIndex)
			err = imp.Start(ctx, false, nil)
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
			importer.MarkIndexingFinished(nextRun)
		}
	})
}

func (s *Server) Start(ctx context.Context, liveReload bool) errors.E {
	var err errors.E
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
