package fetcher

import (
	"context"
	"io"
	"os"

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

func Open(
	source *config.Source,
	opts *Options,
) (*FetchedFiles, error) {
	root := opts.Root

	f := &FetchedFiles{}

	rev, err := openIfExists(root, source.JoinPath("revision"))
	if err != nil {
		opts.Logger.Warn("failed to open revision file", "error", err)
	}
	if rev != nil {
		f.Revision = rev
	}

	switch source.Importer {
	case config.Options:
		f.Options, err = root.Open(source.JoinPath("options.json"))
	case config.Packages:
		f.Packages, err = root.Open(source.JoinPath("packages.json"))
	default:
		err = fault.Newf("unsupported importer type %s", source.Importer.String())
	}
	if err != nil {
		return nil, err
	}

	return f, nil
}

func openIfExists(root *file.Root, filename string) (*os.File, error) {
	if exists, err := root.Exists(filename); err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to check if file exists"))
	} else if !exists {
		return nil, nil
	}

	f, err := root.Open(filename)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.Withf("failed to open file %s", filename))
	}

	return f, nil
}
