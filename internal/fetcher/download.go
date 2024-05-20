package fetcher

import (
	"context"
	"log/slog"
	"net/url"
	"path"
	"searchix/internal/config"
	"searchix/internal/file"

	"github.com/pkg/errors"
)

type DownloadFetcher struct {
	DataPath   string
	Source     *config.Source
	SourceFile string
	Logger     *slog.Logger
}

var files = map[string]string{
	"revision": "revision",
	"options":  "options.json",
}

func (i *DownloadFetcher) FetchIfNeeded(
	parent context.Context,
) (f FetchedFiles, updated bool, err error) {
	ctx, cancel := context.WithTimeout(parent, i.Source.FetchTimeout.Duration)
	defer cancel()

	root := i.DataPath

	err = file.Mkdirp(root)
	if err != nil {
		err = errors.WithMessagef(err, "error creating directory for data: %s", root)

		return
	}

	var fetchURL string
	for _, filename := range files {
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

		outPath := path.Join(root, filename)

		i.Logger.Debug("preparing to fetch URL", "url", fetchURL, "path", outPath)

		updated, err = fetchFileIfNeeded(ctx, outPath, fetchURL)
		if err != nil {
			return
		}
		// don't bother to issue requests for the later files
		if !updated {
			break
		}
	}

	f = FetchedFiles{
		Revision: path.Join(root, "revision"),
		Options:  path.Join(root, "options.json"),
	}

	return
}
