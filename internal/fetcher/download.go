package fetcher

import (
	"context"
	"net/url"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/fetcher/http"
	"alin.ovh/searchix/internal/index"
)

type DownloadFetcher struct {
	Source     config.Source
	SourceFile string
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
		}, nil
	default:
		return nil, fault.Newf("unsupported importer type %s", source.Importer)
	}
}

var files = map[string]string{
	"revision": "revision",
	"options":  "options.json",
	"packages": "packages.json",
}

func (i *DownloadFetcher) FetchIfNeeded(
	ctx context.Context,
	sourceMeta *index.SourceMeta,
) (*FetchedFiles, error) {
	f := &FetchedFiles{}

	filesToFetch := make([]string, 2)
	filesToFetch[0] = files["revision"]
	switch i.Source.Importer {
	case config.Packages:
		filesToFetch[1] = files["packages"]
	case config.Options:
		filesToFetch[1] = files["options"]
	}

	fetcher := http.NewFetcher(&http.Options{
		Logger: i.Logger.Named("http"),
		Root:   i.Root,
	})

	for _, basename := range filesToFetch {
		target := i.Source.JoinPath(basename)
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

		i.Logger.Debug("preparing to fetch URL", "url", fetchURL, "target", target)

		body, err := fetcher.FetchFileIfNeeded(ctx, target, fetchURL)
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
		case files["revision"]:
			f.Revision = body
		case files["options"]:
			f.Options = body
		case files["packages"]:
			f.Packages = body
		default:
			return f, fault.Newf("unknown filename %s", basename)
		}
	}

	return f, nil
}
