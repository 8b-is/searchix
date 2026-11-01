package manpages

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"net/url"
	"strings"
	"sync"
	"time"

	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/fetcher/http"
	"alin.ovh/searchix/internal/file"
)

const (
	basename          = "manpage-urls.json"
	fallbackManBase   = "https://www.mankier.com"
)

type URLMap struct {
	root *file.Root

	mu     sync.RWMutex
	mtime  time.Time
	logger *log.Logger
	urlMap map[string]string
}

type Options struct {
	Logger *log.Logger
	Root   *file.Root
}

func New(opts *Options) *URLMap {
	return &URLMap{
		logger: opts.Logger,
		root:   opts.Root,
		urlMap: make(map[string]string),
	}
}

func sourceFilename(sourceKey string) string {
	if sourceKey == "" || sourceKey == "nixpkgs" {
		return basename
	}

	return fmt.Sprintf("manpage-urls-%s.json", sourceKey)
}

func (m *URLMap) Update(
	ctx context.Context,
	source config.Source,
) error {
	if !source.Manpages.Enable {
		return fault.New("manpages not enabled for this source")
	}

	if source.Manpages.Path == "" {
		return fault.New("manpages repo source path not configured")
	}

	fetcher := http.NewFetcher(&http.Options{
		Logger: m.logger.Named("http"),
		Root:   m.root,
	})

	rawURL, err := makeManpageURL(source)
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to join manpages URL"))
	}

	targetFile := sourceFilename(source.Key)
	m.logger.Debug("fetching manpages URL map", "url", rawURL, "file", targetFile)
	r, err := fetcher.FetchFileIfNeeded(ctx, targetFile, rawURL)
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to fetch manpages"))
	}
	defer r.Close()

	return m.loadFile(targetFile)
}

func (m *URLMap) loadFile(filename string) error {
	stat, err := m.root.StatIfExists(filename)
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("failed to stat manpages file: %s", filename))
	}

	if stat == nil {
		return nil
	}

	data, err := m.root.ReadFile(filename)
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("failed to read manpages file: %s", filename))
	}

	var parsed map[string]string
	if err := json.Unmarshal(data, &parsed); err != nil {
		return fault.Wrap(err, fmsg.Withf("failed to parse manpages JSON: %s", filename))
	}

	m.mu.Lock()
	if m.urlMap == nil {
		m.urlMap = make(map[string]string, len(parsed))
	}
	maps.Copy(m.urlMap, parsed)
	if stat.ModTime().After(m.mtime) {
		m.mtime = stat.ModTime()
	}
	total := len(m.urlMap)
	m.mu.Unlock()

	m.logger.Debug("loaded manpages data", "file", filename, "entries", len(parsed), "total", total)

	return nil
}

// Open loads the manpage URLs from all manpage-urls*.json files in the data directory.
func (m *URLMap) Open() error {
	if m.root == nil {
		return nil
	}

	matches, err := fs.Glob(m.root.FS(), "manpage-urls*.json")
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to glob manpages files"))
	}

	if len(matches) == 0 {
		return m.loadFile(basename)
	}

	for _, name := range matches {
		if err := m.loadFile(name); err != nil {
			return err
		}
	}

	return nil
}

func (m *URLMap) Get(section string, page string) (string, bool) {
	section = strings.TrimSpace(section)
	page = strings.TrimSpace(page)
	if page == "" {
		return "", false
	}

	m.mu.RLock()
	if section != "" {
		key := fmt.Sprintf("%s(%s)", page, section)
		m.logger.Debug("getting manpage URL", "key", key)
		if target, ok := m.urlMap[key]; ok && target != "" {
			m.mu.RUnlock()

			return target, true
		}
	} else {
		if target, ok := m.urlMap[page]; ok && target != "" {
			m.mu.RUnlock()

			return target, true
		}
		for _, sec := range []string{"1", "5", "8", "7", "3", "2", "4", "6", "9"} {
			key := fmt.Sprintf("%s(%s)", page, sec)
			if target, ok := m.urlMap[key]; ok && target != "" {
				m.mu.RUnlock()

				return target, true
			}
		}
	}
	m.mu.RUnlock()

	// Fallback to Mankier when a manpage (e.g. wofi(5) from home-manager) is not in Nixpkgs's manpage-urls.json
	if section != "" {
		return fmt.Sprintf("%s/%s/%s", fallbackManBase, url.PathEscape(section), url.PathEscape(page)), true
	}

	return fmt.Sprintf("%s/%s", fallbackManBase, url.PathEscape(page)), true
}

func makeManpageURL(source config.Source) (string, error) {
	u, err := source.Repo.GetRawFileURL(source.Manpages.Path)
	if err != nil {
		return "", fault.Wrap(err, fmsg.With("failed to join manpage URL"))
	}

	return u, nil
}
