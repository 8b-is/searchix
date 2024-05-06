package server

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	cfg "searchix/internal/config"
	"searchix/internal/options"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/osdevisnot/sorvor/pkg/livereload"
	"github.com/pkg/errors"
	"github.com/shengyanli1982/law"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
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
	Query      string
}

type OptionResultData struct {
	TemplateData
	Query   string
	Results options.NixOptions
}

type TemplateCollection map[string]*template.Template

func applyDevModeOverrides(config *cfg.Config) {
	config.CSP.ScriptSrc = slices.Insert(config.CSP.ScriptSrc, 0, "'unsafe-inline'")
	config.CSP.ConnectSrc = slices.Insert(config.CSP.ConnectSrc, 0, "'self'")
}

var md = goldmark.New(
	goldmark.WithExtensions(extension.NewLinkify()),
)
var templateFuncs = template.FuncMap{
	"markdown": func(input options.Markdown) template.HTML {
		var out strings.Builder
		err := md.Convert([]byte(input), &out)
		if err != nil {
			slog.Warn(fmt.Sprintf("markdown conversion failed: %v", err))

			return template.HTML(html.EscapeString(err.Error())) // #nosec G203 -- duh?
		}

		return template.HTML(out.String()) // #nosec G203
	},
}

func loadTemplate(filename string) (*template.Template, error) {
	text, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.WithMessage(err, "could not read template")
	}
	name, _ := strings.CutSuffix(path.Base(filename), ".gotmpl")
	tpl := template.New(name)
	tpl.Funcs(templateFuncs)
	_, err = tpl.Parse(string(text))
	if err != nil {
		return nil, errors.WithMessage(err, "could not parse template")
	}

	return tpl, nil
}

func loadTemplates() (TemplateCollection, error) {
	templateDir := path.Join("frontend", "templates")
	templates := make(TemplateCollection, 0)

	index, err := loadTemplate(path.Join(templateDir, "index.gotmpl"))
	if err != nil {
		return nil, err
	}
	templates["index"] = index

	templatePaths, err := filepath.Glob(path.Join(templateDir, "blocks", "*.gotmpl"))
	if err != nil {
		return nil, errors.WithMessage(err, "could not glob block templates")
	}
	for _, fullname := range templatePaths {
		tpl, err := loadTemplate(fullname)
		if err != nil {
			return nil, err
		}
		_, err = tpl.AddParseTree("index", index.Tree)
		if err != nil {
			return nil, errors.WithMessage(err, "could not add index template")
		}
		templates[tpl.Name()] = tpl
	}

	return templates, nil
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

	templates, err := loadTemplates()
	if err != nil {
		log.Panicf("could not load templates: %v", err)
	}

	top := http.NewServeMux()
	mux := http.NewServeMux()
	indexData := TemplateData{
		LiveReload: jsSnippet,
	}
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, _ *http.Request) {
		err := templates["index"].Execute(w, indexData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	var nixosOptions = options.NixOptions{}
	jsonFile, err := os.ReadFile(path.Join("data", "test.json"))
	if err != nil {
		slog.Error(fmt.Sprintf("error reading json file: %v", err))
	}
	err = json.Unmarshal(jsonFile, &nixosOptions)
	if err != nil {
		slog.Error(fmt.Sprintf("error parsing json file: %v", err))
	}

	mux.HandleFunc("/options/results", func(w http.ResponseWriter, r *http.Request) {
		tdata := OptionResultData{
			TemplateData: indexData,
			Query:        r.URL.Query().Get("query"),
			Results:      nixosOptions,
		}
		var err error
		if r.Header.Get("Fetch") == "true" {
			w.Header().Add("Content-Type", "text/html; charset=utf-8")
			err = templates["options"].ExecuteTemplate(w, "results", tdata)
		} else {
			err = templates["options"].ExecuteTemplate(w, "index", tdata)
		}
		if err != nil {
			slog.Error(fmt.Sprintf("template error: %v", err))
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
