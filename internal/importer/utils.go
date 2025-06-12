package importer

import (
	"fmt"
	"io"
	"strings"

	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/nix"

	"github.com/bcicen/jstream"
	"gitlab.com/tozd/go/errors"
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

func MakeChannelLink(repo config.Repository, subPath string) (*nix.Link, errors.E) {
	url, err := repo.GetFileURL(subPath)
	if err != nil {
		return nil, err
	}

	return &nix.Link{
		Name: fmt.Sprintf("<%s/%s>", repo.Repo, subPath),
		URL:  url,
	}, nil
}

func setRepoRevision(file io.ReadCloser, source *config.Source) errors.E {
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
