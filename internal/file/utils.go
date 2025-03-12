package file

import (
	"io"
	"io/fs"
	"os"

	"gitlab.com/tozd/go/errors"
)

func Mkdirp(dir string) errors.E {
	err := os.MkdirAll(dir, os.ModeDir|os.ModePerm)
	if err != nil {
		return errors.WithMessagef(err, "could not create directory %s", dir)
	}

	return nil
}

func NeedNotExist(err error) errors.E {
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return errors.WithStack(err)
	}

	return nil
}

func StatIfExists(file string) (fs.FileInfo, errors.E) {
	stat, err := os.Stat(file)

	return stat, NeedNotExist(err)
}

func Exists(file string) (bool, errors.E) {
	stat, err := StatIfExists(file)

	return stat != nil, err
}

func WriteToFile(path string, body io.Reader) errors.E {
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
