package fetcher

import (
	"context"
	"io"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/x/log"

	"gitlab.com/tozd/go/errors"
)

type FetchedFiles struct {
	Revision io.ReadCloser
	Options  io.ReadCloser
	Packages io.ReadCloser
}

type Fetcher interface {
	FetchIfNeeded(context.Context, *index.SourceMeta) (*FetchedFiles, errors.E)
}

func New(
	source *config.Source,
	logger *log.Logger,
) (fetcher Fetcher, err errors.E) {
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
