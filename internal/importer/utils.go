package importer

import (
	"fmt"
	"io"
	"strings"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/nix"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/bcicen/jstream"
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

func MakeChannelLink(repo config.Repository, subPath string) (*nix.Link, error) {
	url, err := repo.GetFileURL(subPath)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.Withf("failed to generate channel link for %s", subPath))
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
			return fault.Wrap(
				err, fmsg.Withf("unable to read revision file"))
		}

		source.Repo.Revision = strings.TrimSpace(str.String())
	}

	return nil
}
