package server

import (
	"fmt"
	"io/fs"
	"time"

	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/fsnotify/fsnotify"

	"alin.ovh/searchix/internal/file"
)

type FileWatcher struct {
	watcher *fsnotify.Watcher
	root    *file.Root
	rootDir string
	log     *log.Logger
}

func NewFileWatcher(log *log.Logger, rootDir string) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("could not create watcher"))
	}

	root, err := file.OpenRoot(rootDir)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("could not open root directory"))
	}

	return &FileWatcher{
		watcher: watcher,
		root:    root,
		rootDir: rootDir,
		log:     log,
	}, nil
}

func (i FileWatcher) Add(path string) error {
	//nolint:forbidigo // external package
	fullpath := i.root.JoinPath(path)
	i.log.Debug(fmt.Sprintf("adding file %s to watcher", fullpath))
	if err := i.watcher.Add(fullpath); err != nil {
		return fault.Wrap(err, fmsg.Withf("could not add file %s to watcher", path))
	}

	return nil
}

func (i FileWatcher) AddRecursive(from string) error {
	err := fs.WalkDir(i.root.FS(), from, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fault.Wrap(err, fmsg.Withf("could not walk directory %s", path))
		}
		if entry.IsDir() {
			if err = i.Add(path); err != nil {
				return err
			}
		}

		return nil
	})

	return fault.Wrap(err, fmsg.With("error walking directory tree"))
}

func (i FileWatcher) Start(callback func(string)) {
	for {
		select {
		case event := <-i.watcher.Events:
			if event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) {
				f, err := i.root.Stat(event.Name)
				if err != nil {
					i.log.Error(fmt.Sprintf("error handling %s event: %v", event.Op.String(), err))
				} else if f.IsDir() {
					err = i.watcher.Add(event.Name)
					if err != nil {
						i.log.Error(fmt.Sprintf("error adding new folder to watcher: %v", err))
					}
				}
			}
			if event.Has(fsnotify.Rename) || event.Has(fsnotify.Write) {
				callback(event.Name)
				time.Sleep(500 * time.Millisecond)
			}
		case err := <-i.watcher.Errors:
			i.log.Error(fmt.Sprintf("error in watcher: %v", err))
		}
	}
}
