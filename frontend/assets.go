package frontend

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"io"
	"io/fs"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
)

type Asset struct {
	URL          string
	ETag         string
	Filename     string
	Base64SHA256 string
}

type AssetCollection struct {
	Scripts     []*Asset
	Stylesheets []*Asset
	ByPath      map[string]*Asset
}

func New() (*AssetCollection, error) {
	a := &AssetCollection{
		Scripts:     []*Asset{},
		Stylesheets: []*Asset{},
		ByPath:      make(map[string]*Asset),
	}

	err := a.Rehash()
	if err != nil {
		return nil, err
	}

	return a, nil
}

func newAsset(filename string) (*Asset, error) {
	file, err := Files.Open(filename)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.Withf("could not open file %s", filename))
	}
	defer file.Close()

	shasum := sha256.New()
	hash := fnv.New64a()
	if _, err := io.Copy(io.MultiWriter(shasum, hash), file); err != nil {
		return nil, fault.Wrap(err, fmsg.Withf("could not hash file %s", filename))
	}

	return &Asset{
		URL:          "/" + filename,
		ETag:         fmt.Sprintf(`W/"%x"`, hash.Sum(nil)),
		Filename:     filename,
		Base64SHA256: base64.StdEncoding.EncodeToString(shasum.Sum(nil)),
	}, nil
}

func (a *AssetCollection) hashScripts() error {
	scripts, err := fs.Glob(Files, "static/**.js")
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not glob files"))
	}
	for _, filename := range scripts {
		asset, err := newAsset(filename)
		if err != nil {
			return err
		}
		a.Scripts = append(a.Scripts, asset)
		a.ByPath[asset.URL] = asset
	}

	return nil
}

func (a *AssetCollection) hashStyles() error {
	styles, err := fs.Glob(Files, "static/**.css")
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not glob files"))
	}
	for _, filename := range styles {
		asset, err := newAsset(filename)
		if err != nil {
			return err
		}
		a.Stylesheets = append(a.Stylesheets, asset)
		a.ByPath[asset.URL] = asset
	}

	return nil
}

func (a *AssetCollection) Rehash() (err error) {
	a.Scripts = []*Asset{}
	err = a.hashScripts()
	if err != nil {
		return err
	}
	a.Stylesheets = []*Asset{}
	err = a.hashStyles()
	if err != nil {
		return err
	}

	return nil
}
