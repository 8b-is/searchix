package importer

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/nix"

	"github.com/bcicen/jstream"
	"github.com/pkg/errors"
)

func ValueTypeToString(valueType jstream.ValueType) string {
	switch valueType {
	case jstream.Unknown:
		return "unknown"
	case jstream.Null:
		return "null"
	case jstream.String:
		return "string"
	case jstream.Number:
		return "number"
	case jstream.Boolean:
		return "boolean"
	case jstream.Array:
		return "array"
	case jstream.Object:
		return "object"
	}

	return "very strange"
}

func makeRepoURL(repo config.Repository, subPath string, line string) (string, error) {
	switch repo.Type {
	case config.GitHub:
		ref := repo.Revision
		if ref == "" {
			ref = "master"
		}
		url, _ := url.JoinPath("https://github.com/", repo.Owner, repo.Repo, "blob", ref, subPath)
		if line != "" {
			url = url + "#L" + line
		}

		return url, nil
	default:
		return "", errors.Errorf(
			"don't know how to generate a repository URL for %s",
			repo.Type.String(),
		)
	}
}

func MakeChannelLink(repo config.Repository, subPath string) (*nix.Link, error) {
	url, err := makeRepoURL(repo, subPath, "")
	if err != nil {
		return nil, err
	}

	return &nix.Link{
		Name: fmt.Sprintf("<%s/%s>", repo.Repo, subPath),
		URL:  url,
	}, nil
}

func setRepoRevision(file io.ReadCloser, source *config.Source) error {
	if file != nil {
		defer file.Close()
		var str strings.Builder
		_, err := io.Copy(&str, file)
		if err != nil {
			return errors.WithMessagef(
				err,
				"unable to read revision file",
			)
		}

		source.Repo.Revision = strings.TrimSpace(str.String())
	}

	return nil
}
