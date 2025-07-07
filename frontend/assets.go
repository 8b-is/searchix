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

var Assets = &AssetCollection{
	Scripts:     []*Asset{},
	Stylesheets: []*Asset{},
	ByPath:      make(map[string]*Asset),
}

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

func hashScripts() error {
	scripts, err := fs.Glob(Files, "static/**.js")
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not glob files"))
	}
	for _, filename := range scripts {
		asset, err := newAsset(filename)
		if err != nil {
			return err
		}
		Assets.Scripts = append(Assets.Scripts, asset)
		Assets.ByPath[asset.URL] = asset
	}

	return nil
}

func hashStyles() error {
	styles, err := fs.Glob(Files, "static/**.css")
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not glob files"))
	}
	for _, filename := range styles {
		asset, err := newAsset(filename)
		if err != nil {
			return err
		}
		Assets.Stylesheets = append(Assets.Stylesheets, asset)
		Assets.ByPath[asset.URL] = asset
	}

	return nil
}

func Rehash() (err error) {
	Assets.Scripts = []*Asset{}
	err = hashScripts()
	if err != nil {
		return err
	}
	Assets.Stylesheets = []*Asset{}
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
