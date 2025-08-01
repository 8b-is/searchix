package web

import (
	"context"
	"time"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/server"
	"alin.ovh/x/log"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/getsentry/sentry-go"
)

type Server struct {
	sv        *server.Server
	cfg       *config.Config
	log       *log.Logger
	sentryHub *sentry.Hub
	options   *server.Options
}

func New(cfg *config.Config, log *log.Logger, options *server.Options) (*Server, error) {
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

func (s *Server) Start(ctx context.Context, liveReload bool) error {
	var err error
	s.sv, err = server.New(s.cfg, s.options, s.log.Named("server"), liveReload)
	if err != nil {
		return fault.Wrap(err, fmsg.With("error setting up server"))
	}

	err = s.sv.Start(ctx)
	if err != nil {
		return fault.Wrap(err, fmsg.With("error starting server"))
	}

	return nil
}

func (s *Server) Stop() {
	<-s.sv.Stop()
	if s.cfg.Web.Environment == "production" {
		s.sentryHub.Flush(2 * time.Second)
	}
}
