//go:build !embed

package frontend

import (
	"io/fs"
	"os"
	"path/filepath"

	"alin.ovh/searchix/internal/file"
)

var Files fs.FS

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	//nolint:forbidigo // own source code
	root, err := file.OpenRoot(filepath.Join(wd, "frontend"))
	if err != nil {
		panic(err)
	}

	Files = root.FS()
}
