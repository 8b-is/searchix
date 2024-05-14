package importer

import (
	"bytes"
	"context"
	"log/slog"
	"net/url"
	"os"
	"path"
	"searchix/internal/config"
	"searchix/internal/file"
	"searchix/internal/index"

	"github.com/pkg/errors"
)

type DownloadOptionsImporter struct {
	DataPath   string
	Source     *config.Source
	SourceFile string
	Logger     *slog.Logger
}

var optionsFiles = map[string]string{
	"revision": "revision",
	"options":  "options.json",
}

func (i *DownloadOptionsImporter) FetchIfNeeded(parent context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(parent, i.Source.FetchTimeout)
	defer cancel()

	root := i.DataPath

	err := file.Mkdirp(root)
	if err != nil {
		return false, errors.WithMessagef(err, "error creating directory for data: %s", root)
	}

	var updated bool
	for _, filename := range optionsFiles {
		url, err := url.JoinPath(i.Source.URL, filename)
		if err != nil {
			return false, errors.WithMessagef(
				err,
				"could not build URL with elements %s and %s",
				i.Source.URL,
				filename,
			)
		}

		path := path.Join(root, filename)

		i.Logger.Debug("preparing to fetch URL", "url", url, "path", path)

		updated, err = fetchFileIfNeeded(ctx, path, url)
		if err != nil {
			return false, err
		}
		// don't bother to issue requests for the later files
		if !updated {
			return false, err
		}
	}

	return updated, nil
}

func (i *DownloadOptionsImporter) Import(
	parent context.Context,
	indexer *index.WriteIndex,
) (bool, error) {
	filename := path.Join(i.DataPath, optionsFiles["options"])
	revFilename := path.Join(i.DataPath, optionsFiles["revision"])
	bits, err := os.ReadFile(revFilename)
	if err != nil {
		return false, errors.WithMessagef(err, "unable to read revision file at %s", revFilename)
	}
	i.Source.Repo.Revision = string(bytes.TrimSpace(bits))
	i.Logger.Debug("preparing import run", "revision", i.Source.Repo.Revision, "filename", filename)

	return processOptions(parent, indexer, &importConfig{
		Source:   i.Source,
		Filename: filename,
		Logger:   i.Logger,
	})
}
