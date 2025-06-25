package fetcher

import (
	"context"
	"net/url"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/fetcher/http"
	"alin.ovh/searchix/internal/index"

	"alin.ovh/x/log"
	"gitlab.com/tozd/go/errors"
)

type DownloadFetcher struct {
	Source     *config.Source
	SourceFile string
	Logger     *log.Logger
}

func NewDownloadFetcher(
	source *config.Source,
	logger *log.Logger,
) (*DownloadFetcher, errors.E) {
	switch source.Importer {
	case config.Options, config.Packages:
		return &DownloadFetcher{
			Source: source,
			Logger: logger,
		}, nil
	default:
		return nil, errors.Errorf("unsupported importer type %s", source.Importer)
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
) (*FetchedFiles, errors.E) {
	f := &FetchedFiles{}
	sourceUpdated := sourceMeta.Updated

	filesToFetch := make([]string, 2)
	filesToFetch[0] = files["revision"]
	switch i.Source.Importer {
	case config.Packages:
		filesToFetch[1] = files["packages"]
	case config.Options:
		filesToFetch[1] = files["options"]
	}

	for _, filename := range filesToFetch {
		fetchURL, baseErr := url.JoinPath(i.Source.URL, filename)
		if baseErr != nil {
			return nil, errors.WithMessagef(
				baseErr,
				"could not build URL with elements %s and %s",
				i.Source.URL,
				filename,
			)
		}

		i.Logger.Debug("preparing to fetch URL", "url", fetchURL)

		body, mtime, err := http.FetchFileIfNeeded(ctx, i.Logger, sourceUpdated, fetchURL)
		if err != nil {
			i.Logger.Warn("failed to fetch file", "url", fetchURL, "error", err)

			return nil, err
		}
		// don't bother to issue requests for the later files
		if mtime.Before(sourceUpdated) {
			break
		}
		sourceMeta.Updated = mtime

		switch filename {
		case files["revision"]:
			f.Revision = body
		case files["options"]:
			f.Options = body
		case files["packages"]:
			f.Packages = body
		default:
			return f, errors.Errorf("unknown filename %s", filename)
		}
	}

	return f, nil
}
