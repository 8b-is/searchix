package importer

import (
	"context"
	"log/slog"
	"os/exec"
	"path"
	"searchix/internal/config"
	"searchix/internal/fetcher"
	"searchix/internal/index"
	"strings"

	"github.com/pkg/errors"
)

func Start(cfg *config.Config, indexer *index.WriteIndex, replace bool) error {
	if len(cfg.Importer.Sources) == 0 {
		slog.Info("No sources enabled")

		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Importer.Timeout.Duration)
	defer cancel()

	for name, source := range cfg.Importer.Sources {
		logger := slog.With("name", name, "fetcher", source.Fetcher.String())
		logger.Debug("starting fetcher")

		fetcherDataPath := path.Join(cfg.DataPath, "sources", source.Key)

		fetcher, err := fetcher.New(source, fetcherDataPath, logger)
		if err != nil {
			logger.Warn("error creating fetcher", "error", err)

			continue
		}

		files, updated, err := fetcher.FetchIfNeeded(ctx)

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

			continue
		}
		logger.Info("importer fetch succeeded", "updated", updated)

		if updated || replace {
			err = setRepoRevision(files.Revision, source)
			if err != nil {
				logger.Warn("could not set source repo revision", "error", err)
			}

			var processor Processor
			switch source.Importer {
			case config.Options:
				logger.Debug(
					"creating processor",
					"filename",
					files.Options,
					"revision",
					source.Repo.Revision,
				)
				processor, err = NewOptionProcessor(files.Options, source)
			case config.Packages:
				logger.Debug(
					"creating processor",
					"filename",
					files.Packages,
					"revision",
					source.Repo.Revision,
				)
				processor, err = NewPackageProcessor(files.Packages, source)
			}
			if err != nil {
				logger.Warn("failed to create processor", "type", source.Importer, "error", err)
			}

			hadWarnings := process(ctx, indexer, processor, logger)

			if hadWarnings {
				logger.Warn("importer succeeded, but with warnings/errors")
			} else {
				logger.Info("importer succeeded")
			}
		}
	}

	return nil
}
