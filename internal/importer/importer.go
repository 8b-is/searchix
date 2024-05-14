package importer

import (
	"context"
	"log/slog"
	"searchix/internal/config"
	"searchix/internal/index"
	"sync"
)

type Importer interface {
	FetchIfNeeded(context.Context) (bool, error)
	Import(context.Context, *index.WriteIndex) (bool, error)
}

func NewNixpkgsChannelImporter(
	source *config.Source,
	dataPath string,
	logger *slog.Logger,
) *NixpkgsChannelImporter {
	return &NixpkgsChannelImporter{
		DataPath: dataPath,
		Source:   source,
		Logger:   logger,
	}
}

func NewChannelImporter(
	source *config.Source,
	dataPath string,
	logger *slog.Logger,
) *ChannelImporter {
	return &ChannelImporter{
		DataPath: dataPath,
		Source:   source,
		Logger:   logger,
	}
}

func NewDownloadOptionsImporter(
	source *config.Source,
	dataPath string,
	logger *slog.Logger,
) *DownloadOptionsImporter {
	return &DownloadOptionsImporter{
		DataPath: dataPath,
		Source:   source,
		Logger:   logger,
	}
}

type importConfig struct {
	Filename string
	Source   *config.Source
	Logger   *slog.Logger
}

func processOptions(
	parent context.Context,
	indexer *index.WriteIndex,
	conf *importConfig,
) (bool, error) {
	ctx, cancel := context.WithTimeout(parent, conf.Source.ImportTimeout)
	defer cancel()

	conf.Logger.Debug("creating option processor", "filename", conf.Filename)
	processor, err := NewOptionProcessor(conf.Filename, conf.Source)
	if err != nil {
		return true, err
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	options, pErrs := processor.Process(ctx)

	wg.Add(1)
	iErrs := indexer.ImportOptions(ctx, options)

	var hadErrors bool
	go func() {
		for {
			select {
			case err, running := <-iErrs:
				if !running {
					wg.Done()
					iErrs = nil
					conf.Logger.Info("ingest completed")

					continue
				}
				hadErrors = true
				conf.Logger.Warn("error ingesting option", "error", err)
			case err, running := <-pErrs:
				if !running {
					wg.Done()
					pErrs = nil
					conf.Logger.Debug("processing completed")

					continue
				}
				hadErrors = true
				conf.Logger.Warn("error processing option", "error", err)
			}
		}
	}()

	conf.Logger.Debug("options processing", "state", "waiting")
	wg.Wait()
	conf.Logger.Debug("options processing", "state", "complete")

	return hadErrors, nil
}
