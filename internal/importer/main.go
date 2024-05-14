package importer

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os/exec"
	"path"
	"searchix/internal/config"
	"searchix/internal/index"
	"strings"
)

func Start(cfg *config.Config, indexer *index.WriteIndex, replace bool) error {
	if len(cfg.Importer.Sources) == 0 {
		slog.Info("No sources enabled")

		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Importer.Timeout.Duration)
	defer cancel()

	var imp Importer
	for name, source := range cfg.Importer.Sources {
		logger := slog.With("name", name, "importer", source.Type.String())
		logger.Debug("starting importer")

		importerDataPath := path.Join(cfg.DataPath, "sources", source.Channel)

		switch source.Type {
		case config.ChannelNixpkgs:
			imp = NewNixpkgsChannelImporter(source, importerDataPath, logger)
		case config.Channel:
			imp = NewChannelImporter(source, importerDataPath, logger)
		case config.DownloadOptions:
			imp = NewDownloadOptionsImporter(source, importerDataPath, logger)
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

			continue
		}
		logger.Info("importer fetch succeeded", "updated", updated)

		if updated || replace {
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

	return nil
}
