package fetcher

import (
	"context"
	"net/url"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/fetcher/http"
	"alin.ovh/searchix/internal/index/meta"
)

type DownloadFetcher struct {
	Source  config.Source
	fetcher *http.Fetcher
	*Options
}

func NewDownloadFetcher(
	source config.Source,
	options *Options,
) (*DownloadFetcher, error) {
	switch source.Importer {
	case config.Options, config.Packages:
		return &DownloadFetcher{
			Source:  source,
			Options: options,
			fetcher: http.NewFetcher(&http.Options{
				Logger: options.Logger.Named("http"),
				Root:   options.Root,
			}),
		}, nil
	default:
		return nil, fault.Newf("unsupported importer type %s", source.Importer)
	}
}

const revisionFileName = "revision"

var files = map[config.ImporterType]string{
	config.Options:  "options.json",
	config.Packages: "packages.json",
}

func (i *DownloadFetcher) FetchIfNeeded(
	ctx context.Context,
	sourceMeta *meta.SourceMeta,
) (*FetchedFiles, error) {
	f := &FetchedFiles{}

	filesToFetch := []string{
		revisionFileName,
		files[i.Source.Importer],
	}

	for _, basename := range filesToFetch {
		target := i.Source.JoinFilePath(basename)
		fetchURL, baseErr := url.JoinPath(i.Source.URL, basename)
		if baseErr != nil {
			return nil, fault.Wrap(
				baseErr,
				fmsg.Withf(
					"could not build URL with elements %s and %s",
					i.Source.URL,
					basename,
				),
			)
		}

		body, err := i.fetcher.FetchFileIfNeeded(ctx, target, fetchURL)
		if err != nil {
			i.Logger.Warn("failed to fetch file", "url", fetchURL, "error", err)

			return nil, fault.Wrap(err, fmsg.Withf("could not fetch file %s", basename))
		}

		stat, err := i.Root.Stat(target)
		if err != nil {
			return nil, fault.Wrap(err, fmsg.Withf("could not stat file %s", target))
		}
		sourceMeta.UpdatedAt = stat.ModTime()

		switch basename {
		case revisionFileName:
			f.Revision = body
		case files[config.Options]:
			f.Options = body
		case files[config.Packages]:
			f.Packages = body
		default:
			return f, fault.Newf("unknown filename %s", basename)
		}
	}

	return f, nil
}
