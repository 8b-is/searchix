package manpages

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/fetcher/http"
	"alin.ovh/searchix/internal/file"
)

const basename = "manpage-urls.json"

type URLMap struct {
	root *file.Root

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

func (m *URLMap) Update(
	ctx context.Context,
	source *config.Source,
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

	url, err := makeManpageURL(source)
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to join manpages URL"))
	}

	m.logger.Debug("fetching manpages URL map", "url", url)
	r, err := fetcher.FetchFileIfNeeded(ctx, basename, url)
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to fetch manpages"))
	}
	defer r.Close()

	return nil
}

// Open loads the manpage URLs from the JSON file
func (m *URLMap) Open() error {
	stat, err := m.root.StatIfExists(basename)
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("failed to stat manpages file: %s", basename))
	}

	if stat == nil {
		return nil
	}

	data, err := m.root.ReadFile(basename)
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to read manpages file"))
	}

	m.mtime = stat.ModTime()

	if err := json.Unmarshal(data, &m.urlMap); err != nil {
		return fault.Wrap(err, fmsg.With("failed to parse manpages JSON"))
	}

	m.logger.Debug("loaded manpages data", "urls", len(m.urlMap))

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

func makeManpageURL(source *config.Source) (string, error) {
	url, err := source.Repo.GetRawFileURL(source.Manpages.Path)
	if err != nil {
		return "", fault.Wrap(err, fmsg.With("failed to join manpage URL"))
	}

	return url, nil
}
