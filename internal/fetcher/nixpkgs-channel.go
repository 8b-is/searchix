package fetcher

import (
	"context"
	"net/url"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/fetcher/http"
	"alin.ovh/searchix/internal/index"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
)

type NixpkgsChannelFetcher struct {
	Source config.Source
	*Options
}

func makeChannelURL(channel string, subPath string) (string, error) {
	url, err := url.JoinPath("https://channels.nixos.org/", channel, subPath)

	return url, fault.Wrap(err, fmsg.Withf("error creating URL"))
}

func NewNixpkgsChannelFetcher(
	source config.Source,
	options *Options,
) (*NixpkgsChannelFetcher, error) {
	switch source.Importer {
	case config.Options, config.Packages:
		return &NixpkgsChannelFetcher{
			Source:  source,
			Options: options,
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

	filesToFetch := make(map[string]string, 2)
	filesToFetch[revisionFilename] = "revision"
	switch i.Source.Importer {
	case config.Packages:
		filesToFetch[packagesFileName] = "packages.json"
	case config.Options:
		filesToFetch[optionsFilename] = "options.json"
	}

	fetcher := http.NewFetcher(&http.Options{
		Logger: i.Logger.Named("http"),
		Root:   i.Root,
	})

	var fetchURL string
	for urlname, filename := range filesToFetch {
		target := i.Source.JoinPath(filename)
		fetchURL, err = makeChannelURL(i.Source.Channel, urlname)
		if err != nil {
			return
		}

		i.Logger.Debug("attempting to fetch file", "url", fetchURL)
		body, err := fetcher.FetchFileIfNeeded(ctx, target, fetchURL)
		if err != nil {
			return f, fault.Wrap(err, fmsg.Withf("failed to fetch file with url %s", fetchURL))
		}

		stat, err := i.Root.Stat(target)
		if err != nil {
			return nil, fault.Wrap(err, fmsg.Withf("could not stat file %s", target))
		}
		sourceMeta.UpdatedAt = stat.ModTime()

		switch urlname {
		case revisionFilename:
			f.Revision = body
		case optionsFilename:
			f.Options = body
		case packagesFileName:
			f.Packages = body
		default:
			return f, fault.Newf("unknown file kind %s", urlname)
		}
	}

	return
}
