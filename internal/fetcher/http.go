package fetcher

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"searchix/internal/config"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
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
	return errors.Wrap(r.src.Close(), "failed to call close on underlying reader")
}

func fetchFileIfNeeded(
	ctx context.Context,
	mtime time.Time,
	url string,
) (body io.ReadCloser, newMtime time.Time, err error) {
	var ifModifiedSince string
	if !mtime.IsZero() {
		ifModifiedSince = strings.Replace(mtime.UTC().Format(time.RFC1123), "UTC", "GMT", 1)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		err = errors.WithMessagef(err, "could not create HTTP request for %s", url)

		return
	}

	req.Header.Set("User-Agent", fmt.Sprintf("Searchix %s", config.ShortSHA))

	if ifModifiedSince != "" {
		req.Header.Set("If-Modified-Since", ifModifiedSince)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		err = errors.WithMessagef(err, "could not make HTTP request to %s", url)

		return
	}

	switch res.StatusCode {
	case http.StatusNotModified:
		newMtime = mtime

		return
	case http.StatusOK:
		newMtime, err = time.Parse(time.RFC1123, res.Header.Get("Last-Modified"))
		if err != nil {
			slog.Warn(
				"could not parse Last-Modified header from response",
				"value",
				res.Header.Get("Last-Modified"),
			)
			newMtime = time.Now()
		}

		switch ce := res.Header.Get("Content-Encoding"); ce {
		case "br":
			slog.Debug("using brotli encoding")
			body = newBrotliReader(res.Body)
		case "", "identity", "gzip":
			body = res.Body
		default:
			err = fmt.Errorf("cannot handle a body with content-encoding %s", ce)
		}
	default:
		err = fmt.Errorf("got response code %d, don't know what to do", res.StatusCode)
	}

	return
}
