package importer

import (
	"context"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"time"

	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/fetcher"
	"go.alanpearce.eu/searchix/internal/index"
	"go.alanpearce.eu/x/log"

	"github.com/pkg/errors"
)

func createSourceImporter(
	parent context.Context,
	log *log.Logger,
	meta *index.Meta,
	indexer *index.WriteIndex,
	forceUpdate bool,
) func(*config.Source) error {
	return func(source *config.Source) error {
		logger := log.With(
			"name",
			source.Key,
			"fetcher",
			source.Fetcher.String(),
			"timeout",
			source.Timeout.Duration,
		)
		logger.Debug("starting fetcher")

		fetcher, err := fetcher.New(source, logger)
		if err != nil {
			return errors.WithMessage(err, "error creating fetcher")
		}

		sourceMeta := meta.GetSourceMeta(source.Key)
		if forceUpdate {
			sourceMeta.Updated = time.Time{}
		}
		previousUpdate := sourceMeta.Updated
		ctx, cancel := context.WithTimeout(parent, source.Timeout.Duration)
		defer cancel()
		files, err := fetcher.FetchIfNeeded(ctx, &sourceMeta)

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
				}
			}

			return errors.WithMessage(err, "importer fetch failed")
		}
		logger.Info(
			"importer fetch succeeded",
			"previous",
			previousUpdate,
			"current",
			sourceMeta.Updated,
			"is_updated",
			sourceMeta.Updated.After(previousUpdate),
			"update_force",
			forceUpdate,
		)

		if sourceMeta.Updated.After(previousUpdate) || forceUpdate {

			err = setRepoRevision(files.Revision, source)
			if err != nil {
				logger.Warn("could not set source repo revision", "error", err)
			}

			var processor Processor
			logger.Debug(
				"creating processor",
				"importer_type",
				source.Importer,
				"revision",
				source.Repo.Revision,
			)
			switch source.Importer {
			case config.Options:
				logger.Debug("processor created", "file", fmt.Sprintf("%T", files.Options))
				processor, err = NewOptionProcessor(
					files.Options,
					source,
					logger.Named("processor"),
				)
			case config.Packages:
				processor, err = NewPackageProcessor(
					files.Packages,
					source,
					logger.Named("processor"),
				)
			}
			if err != nil {
				return errors.WithMessagef(err, "failed to create processor")
			}

			hadWarnings, err := process(ctx, indexer, processor, logger)
			if err != nil {
				return errors.WithMessagef(err, "failed to process source")
			}

			if hadWarnings {
				logger.Warn("importer succeeded, but with warnings/errors")
			} else {
				logger.Info("importer succeeded")
			}
		}

		sourceMeta.Rev = source.Repo.Revision
		meta.SetSourceMeta(source.Key, sourceMeta)

		return nil
	}
}

func Start(
	cfg *config.Config,
	log *log.Logger,
	indexer *index.WriteIndex,
	forceUpdate bool,
	onlyUpdateSources *[]string,
) error {
	if len(cfg.Importer.Sources) == 0 {
		log.Info("No sources enabled")

		return nil
	}

	log.Debug("starting importer", "timeout", cfg.Importer.Timeout.Duration)
	importCtx, cancelImport := context.WithTimeout(
		context.Background(),
		cfg.Importer.Timeout.Duration,
	)
	defer cancelImport()

	forceUpdate = forceUpdate || (onlyUpdateSources != nil && len(*onlyUpdateSources) > 0)

	meta := indexer.Meta

	importSource := createSourceImporter(importCtx, log, meta, indexer, forceUpdate)
	for name, source := range cfg.Importer.Sources {
		if onlyUpdateSources != nil && len(*onlyUpdateSources) > 0 {
			if !slices.Contains(*onlyUpdateSources, name) {
				continue
			}
		}
		err := importSource(source)
		if err != nil {
			log.Error("import failed", "source", name, "error", err)
		}
	}

	err := indexer.SaveMeta()
	if err != nil {
		return errors.Wrap(err, "failed to save metadata")
	}

	return nil
}
