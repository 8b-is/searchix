package fetcher

import (
	"context"
	"io"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
)

type Options struct {
	Logger *log.Logger
}

type FetchedFiles struct {
	Revision io.ReadCloser
	Options  io.ReadCloser
	Packages io.ReadCloser
}

type Fetcher interface {
	FetchIfNeeded(context.Context, *index.SourceMeta) (*FetchedFiles, error)
}

func New(
	source *config.Source,
	opts *Options,
) (fetcher Fetcher, err error) {
	switch source.Fetcher {
	case config.ChannelNixpkgs:
		fetcher, err = NewNixpkgsChannelFetcher(source, opts.Logger)
	case config.Channel:
		fetcher, err = NewChannelFetcher(source, opts.Logger)
	case config.Download:
		fetcher, err = NewDownloadFetcher(source, opts.Logger)
	default:
		err = fault.Newf("unsupported fetcher type %s", source.Fetcher.String())
	}

	return
}
