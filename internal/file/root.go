//nolint:forbidigo // wrappers for os.File functions go here
package file

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
)

type Root struct {
	*os.Root
}

func CreateAndOpenRoot(name string) (*Root, error) {
	exists, err := exists(name)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.Withf("failed to check data root existence %s", name))
	}
	if !exists {
		err := mkdirp(name)
		if err != nil {
			return nil, fault.Wrap(err, fmsg.Withf("failed to create data root %s", name))
		}
	}

	return OpenRoot(name)
}

func OpenRoot(name string) (*Root, error) {
	r, err := os.OpenRoot(name)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.Withf("failed to open data root directory %s", name))
	}

	return &Root{
		Root: r,
	}, nil
}

func (r *Root) JoinPath(path string) string {
	return filepath.Join(r.Name(), path)
}

func (r *Root) StatIfExists(name string) (fs.FileInfo, error) {
	stat, err := r.Stat(name)

	return stat, needNotExist(err)
}

func (r *Root) Exists(file string) (bool, error) {
	stat, err := r.StatIfExists(file)

	return stat != nil, err
}

func (r *Root) CopyTo(source, target string) error {
	src, err := os.Open(source)
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("failed to open source file %s", source))
	}
	defer src.Close()

	dst, err := r.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o640)
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("failed to create target file %s", target))
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fault.Wrap(err, fmsg.Withf("failed to copy file from %s to %s", source, target))
	}

	err = dst.Sync()
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("failed to sync target file %s", target))
	}

	return nil
}

func (r *Root) Chtimes(name string, atime time.Time, mtime time.Time) error {
	err := os.Chtimes(r.JoinPath(name), atime, mtime)
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("failed to change times of %s", name))
	}

	return nil
}

func (r *Root) MkdirAll(name string) error {
	err := os.MkdirAll(r.JoinPath(name), 0o750)
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("failed to create directory %s", name))
	}

	return nil
}

func (r *Root) ReadFile(name string) ([]byte, error) {
	b, err := fs.ReadFile(r.FS(), name)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.Withf("failed to read file %s", name))
	}

	return b, nil
}

func (r *Root) WriteFile(name string, data []byte, perm fs.FileMode) error {
	f, err := r.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("failed to write file %s", name))
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("failed to write file %s", name))
	}

	return nil
}

func (r *Root) RemoveAll(name string) error {
	err := os.RemoveAll(r.JoinPath(name))
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("failed to remove path %s", r.JoinPath(name)))
	}

	return nil
}
