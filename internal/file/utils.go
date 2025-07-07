package file

import (
	"errors"
	"io"
	"io/fs"
	"os"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
)

func Mkdirp(dir string) error {
	err := os.MkdirAll(dir, os.ModeDir|os.ModePerm)
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("could not create directory %s", dir))
	}

	return nil
}

func NeedNotExist(err error) error {
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fault.Wrap(err)
	}

	return nil
}

func StatIfExists(file string) (fs.FileInfo, error) {
	stat, err := os.Stat(file)

	return stat, NeedNotExist(err)
}

func Exists(file string) (bool, error) {
	stat, err := StatIfExists(file)

	return stat != nil, err
}

func WriteToFile(path string, body io.Reader) error {
	file, err := os.Create(path)
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("error creating file at %s", path))
	}
	defer file.Close()

	_, err = io.Copy(file, body)
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("error downloading file %s", path))
	}

	return nil
}
