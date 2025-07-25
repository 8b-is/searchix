package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/file"

	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/andybalholm/brotli"
)

type brotliReadCloser struct {
	src io.ReadCloser
	*brotli.Reader
}

func newBrotliReader(src io.ReadCloser) *brotliReadCloser {
	return &brotliReadCloser{
		src:    src,
		Reader: brotli.NewReader(src),
	}
}

func (r *brotliReadCloser) Close() error {
	return fault.Wrap(r.src.Close(), fmsg.With("failed to call close on underlying reader"))
}

type Options struct {
	Logger *log.Logger
	Root   *file.Root
}

type Fetcher struct {
	logger *log.Logger
	root   *file.Root
}

func NewFetcher(options *Options) *Fetcher {
	return &Fetcher{
		logger: options.Logger,
		root:   options.Root,
	}
}

func (h *Fetcher) FetchFileIfNeeded(
	ctx context.Context,
	filename string,
	url string,
) (io.ReadCloser, error) {
	stat, err := h.root.StatIfExists(filename)
	if err != nil {
		return nil, fault.Wrap(
			err,
			fmsg.Withf("failed to stat file %s", filename),
		)
	}
	var ifModifiedSince string
	if stat != nil && stat.Size() > 0 && !stat.ModTime().IsZero() {
		ifModifiedSince = strings.Replace(
			stat.ModTime().UTC().Format(time.RFC1123),
			"UTC",
			"GMT",
			1,
		)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fault.Wrap(
			err,
			fmsg.Withf("could not create HTTP request for %s", url),
		)
	}

	req.Header.Set("User-Agent", fmt.Sprintf("Searchix %s", config.Version))

	if ifModifiedSince != "" {
		req.Header.Set("If-Modified-Since", ifModifiedSince)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fault.Wrap(
			err,
			fmsg.Withf("could not make HTTP request to %s", url),
		)
	}

	var body io.ReadCloser
	var newMtime time.Time
	encoding := res.Header.Get("Content-Encoding")
	switch res.StatusCode {
	case http.StatusNotModified:
	case http.StatusOK:
		var baseErr error
		newMtime, baseErr = time.Parse(time.RFC1123, res.Header.Get("Last-Modified"))
		if baseErr != nil {
			h.logger.Warn(
				"could not parse Last-Modified header from response",
				"value",
				res.Header.Get("Last-Modified"),
			)
			newMtime = time.Now()
		}

		switch encoding {
		case "br":
			body = newBrotliReader(res.Body)
		case "", "identity", "gzip":
			body = res.Body
		default:
			return nil, fault.Newf("cannot handle a body with content-encoding %s", encoding)
		}

		writer, err := h.root.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
		if err != nil {
			return nil, fault.Wrap(err, fmsg.Withf("failed to open file %s", filename))
		}

		_, err = io.Copy(writer, body)
		if err != nil {
			return nil, fault.Wrap(
				err,
				fmsg.Withf("failed to copy response body to file %s", filename),
			)
		}

		err = writer.Sync()
		if err != nil {
			return nil, fault.Wrap(err, fmsg.Withf("failed to sync file %s", filename))
		}

		err = writer.Close()
		if err != nil {
			return nil, fault.Wrap(err, fmsg.Withf("failed to close file %s", filename))
		}

		err = h.root.Chtimes(filename, time.Time{}, newMtime)
		if err != nil {
			return nil, fault.Wrap(err, fmsg.Withf("failed to update file times %s", filename))
		}

	default:
		return nil, fault.Newf("got response code %d, don't know what to do", res.StatusCode)
	}

	reader, err := h.root.Open(filename)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.Withf("failed to open file %s", filename))
	}

	return reader, nil
}
