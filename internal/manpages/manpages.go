package manpages

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gitlab.com/tozd/go/errors"
	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/fetcher/http"
	"go.alanpearce.eu/x/log"
)

const basename = "manpage-urls.json"

type URLMap struct {
	path string

	mtime  time.Time
	logger *log.Logger
	urlMap map[string]string
}

func New(cfg *config.Config, logger *log.Logger) *URLMap {
	return &URLMap{
		path:   filepath.Join(cfg.DataPath, basename),
		logger: logger,
	}
}

func (m *URLMap) Update(
	ctx context.Context,
	source *config.Source,
) errors.E {
	if !source.Manpages.Enable {
		return errors.New("manpages not enabled for this source")
	}

	if source.Manpages.Path == "" {
		return errors.New("manpages repo source path not configured")
	}

	url, err := makeManpageURL(source)
	if err != nil {
		return errors.WithMessage(err, "failed to join manpages URL")
	}

	m.logger.Debug("fetching manpages URL map", "url", url)
	r, mtime, err := http.FetchFileIfNeeded(ctx, m.logger.Named("http"), m.mtime, url)
	if err != nil {
		return errors.WithMessage(err, "failed to fetch manpages")
	}
	defer r.Close()

	if err := m.save(r); err != nil {
		return errors.WithMessage(err, "failed to save manpages")
	}

	m.mtime = mtime

	return nil
}

// Open loads the manpage URLs from the JSON file
func (m *URLMap) Open() errors.E {
	m.logger.Debug("opening manpages file", "path", m.path)

	stat, err := os.Stat(m.path)
	if err != nil {
		return errors.WithMessagef(err, "failed to stat manpages file: %s", m.path)
	}

	data, err := os.ReadFile(m.path)
	if err != nil {
		return errors.WithMessage(err, "failed to read manpages file")
	}

	m.mtime = stat.ModTime()

	m.urlMap = make(map[string]string)
	if err := json.Unmarshal(data, &m.urlMap); err != nil {
		return errors.WithMessage(err, "failed to parse manpages JSON")
	}

	m.logger.Debug("loaded manpages data", "urls", len(m.urlMap))

	return nil
}

func (m *URLMap) save(r io.Reader) errors.E {
	m.logger.Debug("saving manpages file", "path", m.path)

	f, err := os.Create(m.path)
	if err != nil {
		return errors.WithMessage(err, "failed to create manpages file")
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return errors.WithMessage(err, "failed to write manpages file")
	}

	return nil
}

func (m *URLMap) Get(section string, page string) (string, bool) {
	key := fmt.Sprintf("%s(%s)", page, section)

	m.logger.Debug("getting manpage URL", "key", key)
	url, ok := m.urlMap[key]
	if !ok {
		return "", false
	}

	return url, true
}

func makeManpageURL(source *config.Source) (string, errors.E) {
	url, err := source.Repo.GetRawFileURL(source.Manpages.Path)
	if err != nil {
		return "", errors.WithMessage(err, "failed to join manpage URL")
	}

	return url, nil
}
