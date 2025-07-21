package server

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/manpages"
	"alin.ovh/x/log"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
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

type Options struct {
	ReadIndex      *index.ReadIndex
	ManpagesURLMap *manpages.URLMap
}

func New(
	conf *config.Config,
	options *Options,
	log *log.Logger,
	liveReload bool,
) (*Server, error) {
	mux, err := NewMux(conf, options, log, liveReload)
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

func (s *Server) Start() error {
	listenAddress := net.JoinHostPort(s.cfg.Web.ListenAddress, strconv.Itoa(s.cfg.Web.Port))
	l, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return fault.Wrap(
			err, fmsg.Withf("could not listen on address %s and port %d",
				s.cfg.Web.ListenAddress,
				s.cfg.Web.Port,
			))
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
		return fault.Wrap(err, fmsg.With("could not start server"))
	}

	return nil
}

func (s *Server) Stop() chan struct{} {
	s.log.Debug("stop called")

	idleConnsClosed := make(chan struct{})
	if s.cfg.Web.Environment == "development" {
		shutdownTimeout = 1 * time.Millisecond
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
