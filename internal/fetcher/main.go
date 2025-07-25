package fetcher

import (
	"context"
	"io"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
)

type Options struct {
	Logger *log.Logger
	Root   *file.Root
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
	target := source.JoinPath("")
	exists, err := opts.Root.Exists(target)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to check if directory exists"))
	}

	if !exists {
		err = opts.Root.MkdirAll(target)
		if err != nil {
			return nil, fault.Wrap(err, fmsg.With("failed to create directory"))
		}
	}

	switch source.Fetcher {
	case config.ChannelNixpkgs:
		fetcher, err = NewNixpkgsChannelFetcher(source, opts)
	case config.Channel:
		fetcher, err = NewChannelFetcher(source, opts)
	case config.Download:
		fetcher, err = NewDownloadFetcher(source, opts)
	default:
		err = fault.Newf("unsupported fetcher type %s", source.Fetcher.String())
	}

	return
}
