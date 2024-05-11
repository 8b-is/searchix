package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"searchix/internal/config"
	"searchix/internal/importer"
	"searchix/internal/search"
	"strings"
	"time"

	"github.com/ardanlabs/conf/v3"
)

type Config struct {
	ConfigFile string        `conf:"short:c"`
	LogLevel   slog.Level    `conf:"default:INFO"`
	Timeout    time.Duration `conf:"default:30m,help:maximum time to wait for all fetchers and importers combined"`
	Replace    bool          `conf:"default:false,help:whether to remove existing database, if exists"`
}

func main() {
	if _, found := os.LookupEnv("DEBUG"); found {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	var runtimeConfig Config
	help, err := conf.Parse("", &runtimeConfig)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			log.Println(help)
			os.Exit(1)
		}
		log.Panicf("parsing runtime configuration: %v", err)
	}
	slog.SetLogLoggerLevel(runtimeConfig.LogLevel)

	cfg, err := config.GetConfig(runtimeConfig.ConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	if len(cfg.Sources) == 0 {
		slog.Info("No sources enabled")

		return
	}

	indexer, err := search.NewIndexer(cfg.DataPath, runtimeConfig.Replace)
	if err != nil {
		log.Fatalf("Failed to create indexer: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), runtimeConfig.Timeout)
	defer cancel()

	var imp importer.Importer
	var hadErrors bool
	for name, source := range cfg.Sources {
		logger := slog.With("name", name, "importer", source.Type.String())
		logger.Debug("starting importer")

		switch source.Type {
		case importer.ChannelNixpkgs:
			imp = importer.NewNixpkgsChannelImporter(source, cfg.DataPath, logger)
		case importer.Channel:
			imp = importer.NewChannelImporter(source, cfg.DataPath, logger)
		default:
			log.Printf("unsupported importer type %s", source.Type.String())

			continue
		}

		updated, err := imp.FetchIfNeeded(ctx)

		if err != nil {
			var exerr *exec.ExitError
			if errors.As(err, &exerr) {
				lines := strings.Split(strings.TrimSpace(string(exerr.Stderr)), "\n")
				for _, line := range lines {
					logger.Warn("importer fetch failed", "stderr", line, "status", exerr.ExitCode())
				}
			} else {
				logger.Warn("importer fetch failed", "error", err)
			}
			hadErrors = true

			continue
		}
		logger.Info("importer fetch succeeded", "updated", updated)

		if updated || runtimeConfig.Replace {
			hadWarnings, err := imp.Import(ctx, indexer)

			if err != nil {
				msg := err.Error()
				for _, line := range strings.Split(strings.TrimSpace(msg), "\n") {
					logger.Error("importer init failed", "error", line)
				}

				continue
			}
			if hadWarnings {
				logger.Warn("importer succeeded, but with warnings/errors")
			} else {
				logger.Info("importer succeeded")
			}
		}
	}

	err = indexer.Close()
	if err != nil {
		slog.Error("error closing indexer", "error", err)
	}

	if hadErrors {
		os.RemoveAll(path.Join(cfg.DataPath, "index.bleve"))
		defer os.Exit(1)
	}
}
