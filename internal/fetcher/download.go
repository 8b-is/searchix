package fetcher

import (
	"context"
	"fmt"
	"net/url"

	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/index"

	"github.com/pkg/errors"
	"go.alanpearce.eu/x/log"
)

type DownloadFetcher struct {
	Source     *config.Source
	SourceFile string
	Logger     *log.Logger
}

func NewDownloadFetcher(
	source *config.Source,
	logger *log.Logger,
) (*DownloadFetcher, error) {
	switch source.Importer {
	case config.Options:
		return &DownloadFetcher{
			Source: source,
			Logger: logger,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported importer type %s", source.Importer)
	}
}

var files = map[string]string{
	"revision": "revision",
	"options":  "options.json",
}

func (i *DownloadFetcher) FetchIfNeeded(
	ctx context.Context,
	sourceMeta *index.SourceMeta,
) (f FetchedFiles, err error) {
	var fetchURL string
	sourceUpdated := sourceMeta.Updated
	for key, filename := range files {
		fetchURL, err = url.JoinPath(i.Source.URL, filename)
		if err != nil {
			err = errors.WithMessagef(
				err,
				"could not build URL with elements %s and %s",
				i.Source.URL,
				filename,
			)

			return
		}

		i.Logger.Debug("preparing to fetch URL", "url", fetchURL)

		body, mtime, err := fetchFileIfNeeded(ctx, i.Logger, sourceUpdated, fetchURL)
		if err != nil {
			i.Logger.Warn("failed to fetch file", "url", fetchURL, "error", err)

			return f, err
		}
		// don't bother to issue requests for the later files
		if mtime.Before(sourceUpdated) {
			break
		}
		sourceMeta.Updated = mtime

		switch key {
		case "revision":
			f.Revision = body
		case "options":
			f.Options = body
		default:
			return f, errors.Errorf("unknown file kind %s", key)
		}
	}

	return
}
