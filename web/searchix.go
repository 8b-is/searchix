package web

import (
	"time"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/server"
	"alin.ovh/x/log"

	"github.com/getsentry/sentry-go"
	"gitlab.com/tozd/go/errors"
)

type Server struct {
	sv        *server.Server
	cfg       *config.Config
	log       *log.Logger
	sentryHub *sentry.Hub
	options   *server.Options
}

func New(cfg *config.Config, log *log.Logger, options *server.Options) (*Server, errors.E) {
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
		options:   options,
	}, nil
}

func (s *Server) Start(liveReload bool) errors.E {
	var err errors.E
	s.sv, err = server.New(s.cfg, s.options, s.log.Named("server"), liveReload)
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
