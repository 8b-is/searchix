package fetcher

import (
	"context"
	"log/slog"
	"searchix/internal/config"

	"github.com/pkg/errors"
)

type FetchedFiles struct {
	Revision string
	Options  string
	Packages string
}

type Fetcher interface {
	FetchIfNeeded(context.Context) (FetchedFiles, bool, error)
}

func NewNixpkgsChannelFetcher(
	source *config.Source,
	dataPath string,
	logger *slog.Logger,
) *NixpkgsChannelFetcher {
	return &NixpkgsChannelFetcher{
		DataPath: dataPath,
		Source:   source,
		Logger:   logger,
	}
}

func NewChannelFetcher(
	source *config.Source,
	dataPath string,
	logger *slog.Logger,
) *ChannelFetcher {
	return &ChannelFetcher{
		DataPath: dataPath,
		Source:   source,
		Logger:   logger,
	}
}

func NewDownloadFetcher(
	source *config.Source,
	dataPath string,
	logger *slog.Logger,
) *DownloadFetcher {
	return &DownloadFetcher{
		DataPath: dataPath,
		Source:   source,
		Logger:   logger,
	}
}

func New(
	source *config.Source,
	fetcherDataPath string,
	logger *slog.Logger,
) (fetcher Fetcher, err error) {
	switch source.Fetcher {
	case config.ChannelNixpkgs:
		fetcher = NewNixpkgsChannelFetcher(source, fetcherDataPath, logger)
	case config.Channel:
		fetcher = NewChannelFetcher(source, fetcherDataPath, logger)
	case config.Download:
		fetcher = NewDownloadFetcher(source, fetcherDataPath, logger)
	default:
		err = errors.Errorf("unsupported fetcher type %s", source.Fetcher.String())
	}

	return
}
