package fetcher

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"path"
	"searchix/internal/config"
	"searchix/internal/file"

	"github.com/pkg/errors"
)

type NixpkgsChannelFetcher struct {
	DataPath string
	Source   *config.Source
	Logger   *slog.Logger
}

func makeChannelURL(channel string, subPath string) (string, error) {
	url, err := url.JoinPath("https://channels.nixos.org/", channel, subPath)

	return url, errors.WithMessagef(err, "error creating URL")
}

func NewNixpkgsChannelFetcher(
	source *config.Source,
	dataPath string,
	logger *slog.Logger,
) (*NixpkgsChannelFetcher, error) {
	switch source.Importer {
	case config.Options, config.Packages:
		return &NixpkgsChannelFetcher{
			DataPath: dataPath,
			Source:   source,
			Logger:   logger,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported importer type %s", source.Importer)
	}
}

const (
	revisionFilename = "git-revision"
	optionsFilename  = "options.json.br"
	packagesFileName = "packages.json.br"
)

func (i *NixpkgsChannelFetcher) FetchIfNeeded(
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

		outPath := path.Join(root, filename)

		i.Logger.Debug("attempting to fetch file", "url", fetchURL, "outPath", outPath)
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
		Revision: path.Join(root, "git-revision"),
		Options:  path.Join(root, "options.json.br"),
		Packages: path.Join(root, "packages.json.br"),
	}

	return
}
