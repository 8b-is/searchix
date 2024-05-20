package importer

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"searchix/internal/config"
	"searchix/internal/nix"

	"github.com/andybalholm/brotli"
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

func setRepoRevision(filename string, source *config.Source) error {
	if filename != "" {
		bits, err := os.ReadFile(filename)
		if err != nil {
			return errors.WithMessagef(
				err,
				"unable to read revision file at %s",
				filename,
			)
		}

		source.Repo.Revision = string(bytes.TrimSpace(bits))

	}

	return nil
}

type brotliReadCloser struct {
	src io.ReadCloser
	*brotli.Reader
}

func newBrotliReader(src io.ReadCloser) *brotliReadCloser {
	return &brotliReadCloser{
		src:    src,
		Reader: brotli.NewReader(src),
	}
}

func (r *brotliReadCloser) Close() error {
	return errors.Wrap(r.src.Close(), "failed to call close on underlying reader")
}

func openFileDecoded(filename string) (io.ReadCloser, error) {
	var reader io.ReadCloser
	var err error
	ext := path.Ext(filename)
	reader, err = os.Open(filename)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to open file %s", filename)
	}
	switch ext {
	case ".json":
	// nothing to do
	case ".br":
		reader = newBrotliReader(reader)
	default:
		reader.Close()

		return nil, errors.Errorf("invalid file extension %s", ext)
	}

	return reader, nil
}
