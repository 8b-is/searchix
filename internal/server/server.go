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
	"net/url"
	"os"
	"path"
	"slices"
	"strconv"
	"time"

	cfg "searchix/internal/config"
	"searchix/internal/importer"
	"searchix/internal/options"
	"searchix/internal/search"

	"github.com/blevesearch/bleve/v2"
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
	Production    bool       `conf:"default:false"`
	InDevServer   bool       `conf:"default:false"`
	LiveReload    bool       `conf:"default:false,flag:live"`
	Root          string     `conf:"default:website"`
	ListenAddress string     `conf:"default:localhost"`
	Port          string     `conf:"default:3000,short:p"`
	BaseURL       cfg.URL    `conf:"default:http://localhost:3000,short:b"`
	ConfigFile    string     `conf:"short:c"`
	LogLevel      slog.Level `conf:"default:INFO"`
	IndexPath     string     `conf:"default:data/index.bleve"`
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
	LiveReload   template.HTML
	Sources      map[string]*importer.Source
	Source       importer.Source
	Query        string
	Results      bool
	SourceResult *bleve.SearchResult
}

type ResultData[T options.NixOption] struct {
	TemplateData
	Query          string
	ResultsPerPage int
	Results        *search.Result
	Prev           string
	Next           string
}

func applyDevModeOverrides(config *cfg.Config) {
	config.CSP.ScriptSrc = slices.Insert(config.CSP.ScriptSrc, 0, "'unsafe-inline'")
	config.CSP.ConnectSrc = slices.Insert(config.CSP.ConnectSrc, 0, "'self'")
}

func New(runtimeConfig *Config) (*Server, error) {
	var err error
	config, err = cfg.GetConfig(runtimeConfig.ConfigFile)
	if err != nil {
		return nil, errors.WithMessage(err, "error parsing configuration file")
	}

	slog.Debug("loading index")
	index, err := search.Open(runtimeConfig.IndexPath)
	slog.Debug("loaded index")
	if err != nil {
		log.Fatalf("could not open search index, error: %#v", err)
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

	templates, err := loadTemplates()
	if err != nil {
		log.Panicf("could not load templates: %v", err)
	}

	top := http.NewServeMux()
	mux := http.NewServeMux()
	indexData := TemplateData{
		LiveReload: jsSnippet,
		Sources:    config.Sources,
	}
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, _ *http.Request) {
		err := templates["index"].ExecuteTemplate(w, "index.gotmpl", indexData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	const getSourceTimeout = 1 * time.Second
	mux.HandleFunc("/options/{source}/search", func(w http.ResponseWriter, r *http.Request) {
		sourceKey := r.PathValue("source")

		source := config.Sources[sourceKey]
		if source == nil {
			http.Error(w, "Source not found", http.StatusNotFound)

			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), getSourceTimeout)
		defer cancel()

		sourceResult, err := index.GetSource(ctx, sourceKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		err = templates["search"].Execute(w, TemplateData{
			LiveReload:   jsSnippet,
			Source:       *source,
			SourceResult: sourceResult,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	})

	timeout := 1 * time.Second
	mux.HandleFunc("/options/{source}/results", func(w http.ResponseWriter, r *http.Request) {
		sourceKey := r.PathValue("source")
		ctx, cancel := context.WithTimeoutCause(r.Context(), timeout, errors.New("timeout"))
		defer cancel()
		source := config.Sources[sourceKey]
		if source == nil {
			http.Error(w, "Unknown source", http.StatusNotFound)

			return
		}

		qs := r.URL.Query().Get("query")
		pg := r.URL.Query().Get("page")
		var page uint64 = 1
		if pg != "" {
			page, err = strconv.ParseUint(pg, 10, 64)
			if err != nil || page == 0 {
				http.Error(w, "Bad query string", http.StatusBadRequest)
			}
		}
		results, err := index.Search(ctx, sourceKey, qs, (page-1)*search.ResultsPerPage)
		if err != nil {
			if err == context.DeadlineExceeded {
				http.Error(w, "Search timed out", http.StatusInternalServerError)

				return
			}
			slog.Error("search error", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		tdata := ResultData[options.NixOption]{
			TemplateData: TemplateData{
				LiveReload: jsSnippet,
				Source:     *source,
			},
			ResultsPerPage: search.ResultsPerPage,
			Query:          qs,
			Results:        results,
		}

		hits := uint64(len(results.Hits))
		if results.Total > hits {
			q, err := url.ParseQuery(r.URL.RawQuery)
			if err != nil {
				http.Error(w, "Query string error", http.StatusBadRequest)

				return
			}

			if page*search.ResultsPerPage > results.Total {
				http.Error(w, "Not found", http.StatusNotFound)

				return
			}

			if page*search.ResultsPerPage < results.Total {
				q.Set("page", strconv.FormatUint(page+1, 10))
				tdata.Next = "results?" + q.Encode()
			}

			if page > 1 {
				p := page - 1
				if p == 1 {
					q.Del("page")
				} else {
					q.Set("page", strconv.FormatUint(p, 10))
				}
				tdata.Prev = "results?" + q.Encode()
			}
		}

		if r.Header.Get("Fetch") == "true" {
			w.Header().Add("Content-Type", "text/html; charset=utf-8")
			err = templates["options"].ExecuteTemplate(w, "options.gotmpl", tdata)
		} else {
			err = templates["options"].ExecuteTemplate(w, "index.gotmpl", tdata)
		}
		if err != nil {
			slog.Error("template error", "template", "options", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.Handle(
		"/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("frontend/static"))),
	)

	if runtimeConfig.LiveReload {
		applyDevModeOverrides(config)
		liveReload := livereload.New()
		liveReload.Start()
		top.Handle("/livereload", liveReload)
		fw, err := NewFileWatcher()
		if err != nil {
			return nil, errors.WithMessage(err, "could not create file watcher")
		}
		err = fw.AddRecursive(path.Join("frontend"))
		if err != nil {
			return nil, errors.WithMessage(err, "could not add directory to file watcher")
		}
		go fw.Start(func(filename string) {
			slog.Debug(fmt.Sprintf("got filename %s", filename))
			if path.Ext(filename) == ".gotmpl" {
				templates, err = loadTemplates()
				if err != nil {
					slog.Error(fmt.Sprintf("could not reload templates: %v", err))
				}
			}
			liveReload.Reload()
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
