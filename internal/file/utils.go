package file

import (
	"io"
	"io/fs"
	"os"

	"github.com/pkg/errors"
)

func Mkdirp(dir string) error {
	err := os.MkdirAll(dir, os.ModeDir|os.ModePerm)
	if err != nil {
		return errors.WithMessagef(err, "could not create directory %s", dir)
	}

	return nil
}

func NeedNotExist(err error) error {
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
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
		return errors.WithMessagef(err, "error creating file at %s", path)
	}
	defer file.Close()

	_, err = io.Copy(file, body)
	if err != nil {
		return errors.WithMessagef(err, "error downloading file %s", path)
	}

	return nil
}
