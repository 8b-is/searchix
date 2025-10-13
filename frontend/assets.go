package frontend

import (
	"fmt"
	"hash/fnv"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
)

type Asset struct {
	StaticURL string
	URL       string
	ETag      string
	Filename  string
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

	hash := fnv.New64a()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, fault.Wrap(err, fmsg.Withf("could not hash file %s", filename))
	}

	rel, err := filepath.Rel("static", filename)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.Withf("could not get relative path for %s", filename))
	}

	return &Asset{
		StaticURL: "/" + filename,
		URL:       "/" + rel,
		ETag:      fmt.Sprintf(`W/"%x"`, hash.Sum(nil)),
		Filename:  filename,
	}, nil
}

func (a *AssetCollection) Rehash() (err error) {
	a.Scripts = []*Asset{}
	a.Stylesheets = []*Asset{}

	files, err := fs.Glob(Files, "static/**")
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not glob files"))
	}

	for _, filename := range files {
		asset, err := newAsset(filename)
		if err != nil {
			return err
		}
		switch filepath.Ext(filename) {
		case ".js":
			a.Scripts = append(a.Scripts, asset)
		case ".css":
			a.Stylesheets = append(a.Stylesheets, asset)
		}
		a.ByPath[asset.URL] = asset
		a.ByPath[asset.StaticURL] = asset
	}

	return nil
}
