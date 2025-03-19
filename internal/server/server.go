package server

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/index"
	"go.alanpearce.eu/x/log"

	"gitlab.com/tozd/go/errors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var shutdownTimeout = 10 * time.Second

type Server struct {
	cfg      *config.Config
	log      *log.Logger
	server   *http.Server
	listener net.Listener
}

func New(
	conf *config.Config,
	index *index.ReadIndex,
	log *log.Logger,
	liveReload bool,
) (*Server, errors.E) {
	mux, err := NewMux(conf, index, log, liveReload)
	if err != nil {
		return nil, err
	}

	return &Server{
		cfg: conf,
		log: log,
		server: &http.Server{
			Handler: http.MaxBytesHandler(
				h2c.NewHandler(mux, &http2.Server{
					IdleTimeout: 5 * time.Minute,
				}),
				1024*1024,
			),
			ReadHeaderTimeout: 20 * time.Second,
		},
	}, nil
}

func (s *Server) Start() errors.E {
	listenAddress := net.JoinHostPort(s.cfg.Web.ListenAddress, strconv.Itoa(s.cfg.Web.Port))
	l, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return errors.WithMessagef(
			err,
			"could not listen on address %s and port %d",
			s.cfg.Web.ListenAddress,
			s.cfg.Web.Port,
		)
	}
	s.listener = l

	if s.cfg.Web.Environment == "development" {
		s.log.Info(
			"server listening on",
			"url",
			s.cfg.Web.BaseURL.String(),
		)
	}

	if err := s.server.Serve(l); err != nil && err != http.ErrServerClosed {
		return errors.WithMessage(err, "could not start server")
	}

	return nil
}

func (s *Server) Stop() chan struct{} {
	s.log.Debug("stop called")

	idleConnsClosed := make(chan struct{})
	if s.cfg.Web.Environment == "development" {
		shutdownTimeout = 1 * time.Second
	}

	go func() {
		s.log.Debug("shutting down server")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		err := s.server.Shutdown(ctx)
		s.log.Debug("server shut down")
		if err != nil {
			// Error from closing listeners, or context timeout:
			s.log.Error("error shutting down server", "error", err)
		}
		s.listener.Close()
		close(idleConnsClosed)
	}()

	return idleConnsClosed
}
