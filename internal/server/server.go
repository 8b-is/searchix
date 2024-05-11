package server

import (
	"context"
	"log"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type Server struct {
	*http.Server
}

func New(runtimeConfig *Config) (*Server, error) {
	mux, err := NewMux(runtimeConfig)
	if err != nil {
		return nil, err
	}
	listenAddress := net.JoinHostPort(runtimeConfig.ListenAddress, runtimeConfig.Port)

	return &Server{
		&http.Server{
			Addr:              listenAddress,
			Handler:           mux,
			ReadHeaderTimeout: 20 * time.Second,
		},
	}, nil
}

func (s *Server) Start() error {
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		return errors.WithMessage(err, "could not start server")
	}

	return nil
}

func (s *Server) Stop() chan struct{} {
	slog.Debug("stop called")

	idleConnsClosed := make(chan struct{})

	go func() {
		slog.Debug("shutting down server")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := s.Server.Shutdown(ctx)
		slog.Debug("server shut down")
		if err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	return idleConnsClosed
}
