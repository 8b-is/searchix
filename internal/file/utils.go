//nolint:forbidigo // wrappers for os.File functions go here
package file

import (
	"errors"
	"io/fs"
	"os"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
)

func mkdirp(dir string) error {
	err := os.MkdirAll(dir, os.ModeDir|os.ModePerm)
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("could not create directory %s", dir))
	}

	return nil
}

func needNotExist(err error) error {
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fault.Wrap(err)
	}

	return nil
}

func statIfExists(file string) (fs.FileInfo, error) {
	stat, err := os.Stat(file)

	return stat, needNotExist(err)
}

func exists(file string) (bool, error) {
	stat, err := statIfExists(file)

	return stat != nil, err
}
