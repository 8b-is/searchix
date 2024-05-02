package server

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path"
	"slices"
	"time"

	cfg "searchix/internal/config"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/osdevisnot/sorvor/pkg/livereload"
	"github.com/pkg/errors"
	"github.com/shengyanli1982/law"
)

var config *cfg.Config

var (
	CommitSHA string
	ShortSHA  string
)

type Config struct {
	Production    bool    `conf:"default:false"`
	InDevServer   bool    `conf:"default:false"`
	LiveReload    bool    `conf:"default:false,flag:live"`
	Root          string  `conf:"default:website"`
	ListenAddress string  `conf:"default:localhost"`
	Port          string  `conf:"default:3000,short:p"`
	BaseURL       cfg.URL `conf:"default:http://localhost:3000,short:b"`
}

type HTTPError struct {
	Error   error
	Message string
	Code    int
}

type Server struct {
	*http.Server
}

const jsSnippet = template.HTML(livereload.JsSnippet) // #nosec G203

type TemplateData struct {
	LiveReload template.HTML
	Data       map[string]interface{}
}

func applyDevModeOverrides(config *cfg.Config) {
	config.CSP.ScriptSrc = slices.Insert(config.CSP.ScriptSrc, 0, "'unsafe-inline'")
	config.CSP.ConnectSrc = slices.Insert(config.CSP.ConnectSrc, 0, "'self'")
}

func New(runtimeConfig *Config) (*Server, error) {
	var err error
	config, err = cfg.GetConfig()
	if err != nil {
		return nil, errors.WithMessage(err, "error parsing configuration file")
	}

	env := "development"
	if runtimeConfig.Production {
		env = "production"
	}
	err = sentry.Init(sentry.ClientOptions{
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		Dsn:              os.Getenv("SENTRY_DSN"),
		Release:          CommitSHA,
		Environment:      env,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "could not set up sentry")
	}
	defer sentry.Flush(2 * time.Second)
	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})

	templatePaths := path.Join("frontend", "templates", "*.tmpl")
	tpl := template.Must(template.ParseGlob(templatePaths))

	top := http.NewServeMux()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		tdata := TemplateData{
			LiveReload: jsSnippet,
			Data:       make(map[string]interface{}),
		}
		err := tpl.Execute(w, tdata)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend/static"))))

	if runtimeConfig.LiveReload {
		applyDevModeOverrides(config)
		liveReload := livereload.New()
		liveReload.Start()
		top.Handle("/livereload", liveReload)
		fw, err := NewFileWatcher()
		if err != nil {
			return nil, errors.WithMessage(err, "could not create file watcher")
		}
		err = fw.AddRecursive("frontend")
		if err != nil {
			return nil, errors.WithMessage(err, "could not add directory to file watcher")
		}
		go fw.Start(func() {
			t, err := template.ParseGlob(path.Join("frontend", "templates", "*.tmpl"))
			if err != nil {
				slog.Error(fmt.Sprintf("could not parse template: %v", err))
			} else {
				tpl = t
				liveReload.Reload()
			}
		})
	}

	var logWriter io.Writer
	if runtimeConfig.Production {
		logWriter = law.NewWriteAsyncer(os.Stdout, nil)
	} else {
		logWriter = os.Stdout
	}
	top.Handle("/",
		AddHeadersMiddleware(
			sentryHandler.Handle(
				wrapHandlerWithLogging(mux, wrappedHandlerOptions{
					defaultHostname: runtimeConfig.BaseURL.Hostname(),
					logger:          logWriter,
				}),
			),
			config,
		),
	)
	// no logging, no sentry
	top.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	listenAddress := net.JoinHostPort(runtimeConfig.ListenAddress, runtimeConfig.Port)

	return &Server{
		&http.Server{
			Addr:              listenAddress,
			Handler:           top,
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
