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

type NixpkgsChannelFetcher struct {
	DataPath string
	Source   *config.Source
	Logger   *slog.Logger
}

func makeChannelURL(channel string, subPath string) (string, error) {
	url, err := url.JoinPath("https://channels.nixos.org/", channel, subPath)

	return url, errors.WithMessagef(err, "error creating URL")
}

var filesToFetch = map[string]string{
	"revision": "git-revision",
	"options":  "options.json.br",
	"packages": "packages.json.br",
}

func (i *NixpkgsChannelFetcher) FetchIfNeeded(
	parent context.Context,
) (f FetchedFiles, updated bool, err error) {
	ctx, cancel := context.WithTimeout(parent, i.Source.FetchTimeout)
	defer cancel()

	root := i.DataPath

	err = file.Mkdirp(root)
	if err != nil {
		err = errors.WithMessagef(err, "error creating directory for data: %s", root)

		return
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
