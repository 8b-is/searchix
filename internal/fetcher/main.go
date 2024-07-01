package fetcher

import (
	"context"
	"io"

	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/index"
	"go.alanpearce.eu/x/log"

	"github.com/pkg/errors"
)

type FetchedFiles struct {
	Revision io.ReadCloser
	Options  io.ReadCloser
	Packages io.ReadCloser
}

type Fetcher interface {
	FetchIfNeeded(context.Context, *index.SourceMeta) (FetchedFiles, error)
}

func New(
	source *config.Source,
	logger *log.Logger,
) (fetcher Fetcher, err error) {
	switch source.Fetcher {
	case config.ChannelNixpkgs:
		fetcher, err = NewNixpkgsChannelFetcher(source, logger)
	case config.Channel:
		fetcher, err = NewChannelFetcher(source, logger)
	case config.Download:
		fetcher, err = NewDownloadFetcher(source, logger)
	default:
		err = errors.Errorf("unsupported fetcher type %s", source.Fetcher.String())
	}

	return
}
