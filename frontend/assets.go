package frontend

import (
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
)

type Asset struct {
	ETag          string
	Filename      string
	ImmutablePath string
	StaticURL     string
}

type AssetCollection struct {
	Scripts         []*Asset
	Stylesheets     []*Asset
	ByImmutablePath map[string]*Asset
	ByPath          map[string]*Asset
}

func New() (*AssetCollection, error) {
	a := &AssetCollection{
		Scripts:         []*Asset{},
		Stylesheets:     []*Asset{},
		ByImmutablePath: make(map[string]*Asset),
		ByPath:          make(map[string]*Asset),
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

	hasher := fnv.New64a()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, fault.Wrap(err, fmsg.Withf("could not hash file %s", filename))
	}

	rel, err := filepath.Rel("static", filename)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.Withf("could not get relative path for %s", filename))
	}

	hash := hex.EncodeToString(hasher.Sum(nil))

	return &Asset{
		ETag:          fmt.Sprintf(`W/"%s"`, hash),
		Filename:      filename,
		ImmutablePath: makeImmutablePath(rel, hash),
		StaticURL:     "/" + rel,
	}, nil
}

func makeImmutablePath(filename string, hash string) string {
	ext := filepath.Ext(filename)

	return path.Join("/", "assets", strings.Replace(filename, ext, "."+hash+ext, 1))
}

func (a *AssetCollection) Rehash() (err error) {
	a.Scripts = nil
	a.Stylesheets = nil
	clear(a.ByImmutablePath)
	clear(a.ByPath)

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
		a.ByImmutablePath[asset.ImmutablePath] = asset
		a.ByPath[asset.StaticURL] = asset
	}

	return nil
}
