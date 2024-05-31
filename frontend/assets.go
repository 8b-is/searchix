package frontend

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"hash/fnv"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

var Assets = &AssetCollection{
	Scripts:         make(map[string]*Asset),
	Stylesheets:     make(map[string]*Asset),
	ByImmutablePath: make(map[string]*Asset),
}

type Asset struct {
	URL          string
	Filename     string
	Base64SHA256 string
}

type AssetCollection struct {
	Scripts         map[string]*Asset
	Stylesheets     map[string]*Asset
	ByImmutablePath map[string]*Asset
}

func newAsset(filename string) (*Asset, error) {
	file, err := Files.Open(filename)
	if err != nil {
		return nil, errors.WithMessagef(err, "could not open file %s", filename)
	}
	defer file.Close()

	shasum := sha256.New()
	hash := fnv.New64a()
	if _, err := io.Copy(io.MultiWriter(shasum, hash), file); err != nil {
		return nil, errors.WithMessagef(err, "could not hash file %s", filename)
	}

	return &Asset{
		URL:          makeImmutablePath(filename, hex.EncodeToString(hash.Sum(nil))),
		Filename:     filename,
		Base64SHA256: base64.StdEncoding.EncodeToString(shasum.Sum(nil)),
	}, nil
}

func makeImmutablePath(filename string, hash string) string {
	ext := filepath.Ext(filename)

	return "/" + strings.Replace(filename, ext, "."+hash+ext, 1)
}

func hashScripts() error {
	scripts, err := fs.Glob(Files, "static/**.js")
	if err != nil {
		return errors.WithMessage(err, "could not glob files")
	}
	for _, filename := range scripts {
		asset, err := newAsset(filename)
		if err != nil {
			return err
		}
		Assets.Scripts[filename] = asset
		Assets.ByImmutablePath[asset.URL] = asset
	}

	return nil
}

func hashStyles() error {
	styles, err := fs.Glob(Files, "static/**.css")
	if err != nil {
		return errors.WithMessage(err, "could not glob files")
	}
	for _, filename := range styles {
		asset, err := newAsset(filename)
		if err != nil {
			return err
		}
		Assets.Stylesheets[filename] = asset
		Assets.ByImmutablePath[asset.URL] = asset
	}

	return nil
}

func Rehash() (err error) {
	err = hashScripts()
	if err != nil {
		return err
	}
	err = hashStyles()
	if err != nil {
		return err
	}

	return nil
}

func init() {
	err := Rehash()
	if err != nil {
		panic(err)
	}
}
