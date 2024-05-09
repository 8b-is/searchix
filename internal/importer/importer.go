package importer

import (
	"context"
	"log/slog"
	"path"
	"searchix/internal/search"
	"sync"
	"time"
)

type Source struct {
	Name          string
	Enable        bool
	Type          Type
	Channel       string
	Attribute     string
	ImportPath    string        `toml:"import-path"`
	FetchTimeout  time.Duration `toml:"fetch-timeout"`
	ImportTimeout time.Duration `toml:"import-timeout"`
	OutputPath    string        `toml:"output-path"`
	Repo          Repository
}

type Importer interface {
	FetchIfNeeded(context.Context) (bool, error)
	Import(context.Context, *search.WriteIndex) (bool, error)
}

func NewNixpkgsChannelImporter(
	source Source,
	dataPath string,
	logger *slog.Logger,
) *NixpkgsChannelImporter {
	indexPath := dataPath
	fullpath := path.Join(dataPath, source.Channel)

	return &NixpkgsChannelImporter{
		DataPath:  fullpath,
		Source:    source,
		Logger:    logger,
		indexPath: indexPath,
	}
}

func NewChannelImporter(source Source, dataPath string, logger *slog.Logger) *ChannelImporter {
	indexPath := dataPath
	fullpath := path.Join(dataPath, source.Channel)

	return &ChannelImporter{
		DataPath:  fullpath,
		Source:    source,
		Logger:    logger,
		indexPath: indexPath,
	}
}

type importConfig struct {
	IndexPath string
	Filename  string
	Source    Source
	Logger    *slog.Logger
}

func processOptions(
	parent context.Context,
	indexer *search.WriteIndex,
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
					slog.Info("ingest completed")

					continue
				}
				hadErrors = true
				conf.Logger.Warn("error ingesting option", "error", err)
			case err, running := <-pErrs:
				if !running {
					wg.Done()
					pErrs = nil
					slog.Info("processing completed")

					continue
				}
				hadErrors = true
				conf.Logger.Warn("error processing option", "error", err)
			}
		}
	}()

	slog.Debug("options processing", "state", "waiting")
	wg.Wait()
	slog.Debug("options processing", "state", "complete")

	return hadErrors, nil
}
