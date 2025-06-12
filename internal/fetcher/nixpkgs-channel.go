package fetcher

import (
	"context"
	"net/url"

	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/fetcher/http"
	"go.alanpearce.eu/searchix/internal/index"

	"gitlab.com/tozd/go/errors"
	"go.alanpearce.eu/x/log"
)

type NixpkgsChannelFetcher struct {
	Source *config.Source
	Logger *log.Logger
}

func makeChannelURL(channel string, subPath string) (string, errors.E) {
	url, err := url.JoinPath("https://channels.nixos.org/", channel, subPath)

	return url, errors.WithMessagef(err, "error creating URL")
}

func NewNixpkgsChannelFetcher(
	source *config.Source,
	logger *log.Logger,
) (*NixpkgsChannelFetcher, errors.E) {
	switch source.Importer {
	case config.Options, config.Packages:
		return &NixpkgsChannelFetcher{
			Source: source,
			Logger: logger,
		}, nil
	default:
		return nil, errors.Errorf("unsupported importer type %s", source.Importer)
	}
}

const (
	revisionFilename = "git-revision"
	optionsFilename  = "options.json.br"
	packagesFileName = "packages.json.br"
)

func (i *NixpkgsChannelFetcher) FetchIfNeeded(
	ctx context.Context,
	sourceMeta *index.SourceMeta,
) (f *FetchedFiles, err errors.E) {
	f = &FetchedFiles{}

	filesToFetch := make([]string, 2)
	filesToFetch[0] = revisionFilename
	switch i.Source.Importer {
	case config.Packages:
		filesToFetch[1] = packagesFileName
	case config.Options:
		filesToFetch[1] = optionsFilename
	}

	var fetchURL string
	for _, filename := range filesToFetch {
		fetchURL, err = makeChannelURL(i.Source.Channel, filename)
		if err != nil {
			return
		}

		i.Logger.Debug("attempting to fetch file", "url", fetchURL)
		body, mtime, err := http.FetchFileIfNeeded(ctx, i.Logger, sourceMeta.Updated, fetchURL)
		if err != nil {
			i.Logger.Warn("failed to fetch file", "url", fetchURL, "error", err)

			return f, err
		}
		// don't bother to issue requests for the later files
		if mtime.Before(sourceMeta.Updated) {
			break
		}
		sourceMeta.Updated = mtime

		switch filename {
		case revisionFilename:
			f.Revision = body
		case optionsFilename:
			f.Options = body
		case packagesFileName:
			f.Packages = body
		default:
			return f, errors.Errorf("unknown file kind %s", filename)
		}
	}

	return
}
