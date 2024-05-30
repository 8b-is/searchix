package frontend

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"io/fs"
	"net/url"

	"github.com/pkg/errors"
)

var Assets = AssetCollection{
	Scripts:     make(map[string]Asset),
	Stylesheets: make(map[string]Asset),
}

type Asset struct {
	URL          string
	Base64SHA256 string
}

type AssetCollection struct {
	Scripts     map[string]Asset
	Stylesheets map[string]Asset
}

func hashFile(filename string) ([]byte, error) {
	file, err := Files.Open(filename)
	if err != nil {
		return nil, errors.WithMessagef(err, "could not open file %s", filename)
	}
	sum := sha256.New()
	defer file.Close()
	if _, err := io.Copy(sum, file); err != nil {
		return nil, errors.WithMessagef(err, "could not hash file %s", filename)
	}

	return sum.Sum(nil), nil
}

func newAsset(filename string, hash []byte) Asset {
	u, err := url.JoinPath("/", filename)
	if err != nil {
		panic(err)
	}

	return Asset{
		URL:          u,
		Base64SHA256: base64.StdEncoding.EncodeToString(hash),
	}
}

func hashScripts() error {
	scripts, err := fs.Glob(Files, "static/**.js")
	if err != nil {
		return errors.WithMessage(err, "could not glob files")
	}
	for _, filename := range scripts {
		hash, err := hashFile(filename)
		if err != nil {
			return err
		}
		Assets.Scripts[filename] = newAsset(filename, hash)
	}

	return nil
}

func hashStyles() error {
	styles, err := fs.Glob(Files, "static/**.css")
	if err != nil {
		return errors.WithMessage(err, "could not glob files")
	}
	for _, filename := range styles {
		hash, err := hashFile(filename)
		if err != nil {
			return err
		}
		Assets.Stylesheets[filename] = newAsset(filename, hash)
	}

	return nil
}

func init() {
	err := hashScripts()
	if err != nil {
		panic(err)
	}
	err = hashStyles()
	if err != nil {
		panic(err)
	}
}
