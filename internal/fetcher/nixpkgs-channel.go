package fetcher

import (
	"context"
	"net/url"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/fetcher/http"
	"alin.ovh/searchix/internal/index"

	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
)

type NixpkgsChannelFetcher struct {
	Source *config.Source
	Logger *log.Logger
}

func makeChannelURL(channel string, subPath string) (string, error) {
	url, err := url.JoinPath("https://channels.nixos.org/", channel, subPath)

	return url, fault.Wrap(err, fmsg.Withf("error creating URL"))
}

func NewNixpkgsChannelFetcher(
	source *config.Source,
	logger *log.Logger,
) (*NixpkgsChannelFetcher, error) {
	switch source.Importer {
	case config.Options, config.Packages:
		return &NixpkgsChannelFetcher{
			Source: source,
			Logger: logger,
		}, nil
	default:
		return nil, fault.Newf("unsupported importer type %s", source.Importer)
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
) (f *FetchedFiles, err error) {
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
			return f, fault.Wrap(err, fmsg.Withf("failed to fetch file with url %s", fetchURL))
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
			return f, fault.Newf("unknown file kind %s", filename)
		}
	}

	return
}
