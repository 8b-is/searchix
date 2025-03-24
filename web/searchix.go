package web

import (
	"time"

	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/index"
	"go.alanpearce.eu/searchix/internal/server"
	"go.alanpearce.eu/x/log"

	"github.com/getsentry/sentry-go"
	"gitlab.com/tozd/go/errors"
)

type Server struct {
	sv        *server.Server
	cfg       *config.Config
	log       *log.Logger
	sentryHub *sentry.Hub
	readIndex *index.ReadIndex
}

func New(cfg *config.Config, log *log.Logger, read *index.ReadIndex) (*Server, errors.E) {
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
		readIndex: read,
	}, nil
}

func (s *Server) Start(liveReload bool) errors.E {
	var err errors.E
	s.sv, err = server.New(s.cfg, s.readIndex, s.log.Named("server"), liveReload)
	if err != nil {
		return errors.Wrap(err, "error setting up server")
	}

	err = s.sv.Start()
	if err != nil {
		return errors.Wrap(err, "error starting server")
	}

	return nil
}

func (s *Server) Stop() {
	<-s.sv.Stop()
	s.sentryHub.Flush(2 * time.Second)
}
