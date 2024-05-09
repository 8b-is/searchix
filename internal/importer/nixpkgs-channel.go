package importer

import (
	"bytes"
	"context"
	"log/slog"
	"net/url"
	"os"
	"path"
	"searchix/internal/file"
	"searchix/internal/search"

	"github.com/pkg/errors"
)

type NixpkgsChannelImporter struct {
	DataPath  string
	Source    Source
	Logger    *slog.Logger
	indexPath string
}

func makeChannelURL(channel string, subPath string) (string, error) {
	url, err := url.JoinPath("https://channels.nixos.org/", channel, subPath)

	return url, errors.WithMessagef(err, "error creating URL")
}

var filesToFetch = map[string]string{
	"revision": "git-revision",
	"options":  "options.json.br",
}

func (i *NixpkgsChannelImporter) FetchIfNeeded(parent context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(parent, i.Source.FetchTimeout)
	defer cancel()

	root := i.DataPath

	err := file.Mkdirp(root)
	if err != nil {
		return false, errors.WithMessagef(err, "error creating directory for data: %s", root)
	}

	for _, filename := range filesToFetch {
		url, err := makeChannelURL(i.Source.Channel, filename)
		if err != nil {
			return false, err
		}

		path := path.Join(root, filename)

		updated, err := fetchFileIfNeeded(ctx, path, url)
		if err != nil {
			return false, err
		}
		// don't bother to issue requests for the later files
		if !updated {
			return false, err
		}
	}

	return true, nil
}

func (i *NixpkgsChannelImporter) Import(parent context.Context, indexer *search.WriteIndex) (bool, error) {
	filename := path.Join(i.DataPath, filesToFetch["options"])
	revFilename := path.Join(i.DataPath, filesToFetch["revision"])
	bits, err := os.ReadFile(revFilename)
	if err != nil {
		return false, errors.WithMessagef(err, "unable to read revision file at %s", revFilename)
	}
	i.Source.Repo.Revision = string(bytes.TrimSpace(bits))
	i.Logger.Debug("preparing import run", "revision", i.Source.Repo.Revision, "filename", filename)

	return processOptions(parent, indexer, &importConfig{
		IndexPath: i.indexPath,
		Source:    i.Source,
		Filename:  filename,
		Logger:    i.Logger,
	})
}
