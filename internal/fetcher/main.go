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

func New(
	source *config.Source,
	fetcherDataPath string,
	logger *slog.Logger,
) (fetcher Fetcher, err error) {
	switch source.Fetcher {
	case config.ChannelNixpkgs:
		fetcher, err = NewNixpkgsChannelFetcher(source, fetcherDataPath, logger)
	case config.Channel:
		fetcher, err = NewChannelFetcher(source, fetcherDataPath, logger)
	case config.Download:
		fetcher, err = NewDownloadFetcher(source, fetcherDataPath, logger)
	default:
		err = errors.Errorf("unsupported fetcher type %s", source.Fetcher.String())
	}

	return
}
