package server

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

type FileWatcher struct {
	*fsnotify.Watcher
}

func NewFileWatcher() (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.WithMessage(err, "could not create watcher")
	}

	return &FileWatcher{watcher}, nil
}

func (watcher FileWatcher) AddRecursive(from string) error {
	slog.Debug(fmt.Sprintf("watching files under %s", from))
	err := filepath.WalkDir(from, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return errors.WithMessagef(err, "could not walk directory %s", path)
		}
		if entry.IsDir() {
			slog.Debug(fmt.Sprintf("adding directory %s to watcher", path))
			if err = watcher.Add(path); err != nil {
				return errors.WithMessagef(err, "could not add directory %s to watcher", path)
			}
		}

		return nil
	})

	return errors.WithMessage(err, "error walking directory tree")
}

func (watcher FileWatcher) Start(callback func()) {
	for {
		select {
		case event := <-watcher.Events:
			slog.Debug(fmt.Sprintf("event received: %s", event))
			if event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) {
				f, err := os.Stat(event.Name)
				if err != nil {
					slog.Error(fmt.Sprintf("error handling %s event: %v", event.Op.String(), err))
				} else if f.IsDir() {
					err = watcher.Add(event.Name)
					if err != nil {
						slog.Error(fmt.Sprintf("error adding new folder to watcher: %v", err))
					}
				}
			}
			if event.Has(fsnotify.Rename) || event.Has(fsnotify.Write) {
				callback()
				time.Sleep(500 * time.Millisecond)
			}
		case err := <-watcher.Errors:
			slog.Error(fmt.Sprintf("error in watcher: %v", err))
		}
	}
}
