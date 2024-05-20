package importer

import (
	"context"
	"io"
	"log/slog"
	"os/exec"
	"path"
	"searchix/internal/config"
	"searchix/internal/fetcher"
	"searchix/internal/index"
	"slices"
	"strings"

	"github.com/pkg/errors"
)

func Start(
	cfg *config.Config,
	indexer *index.WriteIndex,
	forceUpdate bool,
	onlyUpdateSources *[]string,
) error {
	if len(cfg.Importer.Sources) == 0 {
		slog.Info("No sources enabled")

		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Importer.Timeout.Duration)
	defer cancel()

	forceUpdate = forceUpdate || (onlyUpdateSources != nil && len(*onlyUpdateSources) > 0)

	for name, source := range cfg.Importer.Sources {
		if onlyUpdateSources != nil && len(*onlyUpdateSources) > 0 {
			if !slices.Contains(*onlyUpdateSources, name) {
				continue
			}
		}

		logger := slog.With("name", name, "fetcher", source.Fetcher.String())
		logger.Debug("starting fetcher")

		fetcherDataPath := path.Join(cfg.DataPath, "sources", source.Key)

		fetcher, err := fetcher.New(source, fetcherDataPath, logger)
		if err != nil {
			logger.Error("error creating fetcher", "error", err)

			continue
		}

		files, updated, err := fetcher.FetchIfNeeded(ctx)

		if err != nil {
			var exerr *exec.ExitError
			if errors.As(err, &exerr) {
				lines := strings.Split(strings.TrimSpace(string(exerr.Stderr)), "\n")
				for _, line := range lines {
					logger.Error(
						"importer fetch failed",
						"stderr",
						line,
						"status",
						exerr.ExitCode(),
					)

					continue
				}
			} else {
				logger.Error("importer fetch failed", "error", err)

				continue
			}

			continue
		}
		logger.Info("importer fetch succeeded", "updated", updated)

		if updated || forceUpdate {
			err = setRepoRevision(files.Revision, source)
			if err != nil {
				logger.Warn("could not set source repo revision", "error", err)
			}

			var file io.ReadCloser
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
				file, err = openFileDecoded(files.Options)
				if err != nil {
					logger.Error("could not open file", "filename", files.Options, "error", err)

					continue
				}
				processor, err = NewOptionProcessor(file, source)
			case config.Packages:
				logger.Debug(
					"creating processor",
					"filename",
					files.Packages,
					"revision",
					source.Repo.Revision,
				)
				file, err = openFileDecoded(files.Packages)
				if err != nil {
					logger.Error("could not open file", "filename", files.Packages, "error", err)

					continue
				}
				processor, err = NewPackageProcessor(file, source)
			}
			if err != nil {
				logger.Error("failed to create processor", "type", source.Importer, "error", err)

				continue
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
