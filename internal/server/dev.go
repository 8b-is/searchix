package server

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher *fsnotify.Watcher
	log     *log.Logger
}

func NewFileWatcher(log *log.Logger) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("could not create watcher"))
	}

	return &FileWatcher{
		watcher,
		log,
	}, nil
}

func (i FileWatcher) AddRecursive(from string) error {
	i.log.Debug(fmt.Sprintf("watching files under %s", from))
	err := filepath.WalkDir(from, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fault.Wrap(err, fmsg.Withf("could not walk directory %s", path))
		}
		if entry.IsDir() {
			i.log.Debug(fmt.Sprintf("adding directory %s to watcher", path))
			if err = i.watcher.Add(path); err != nil {
				return fault.Wrap(err, fmsg.Withf("could not add directory %s to watcher", path))
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
				f, err := os.Stat(event.Name)
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
