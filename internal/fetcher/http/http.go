package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"alin.ovh/searchix/internal/config"

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

func FetchFileIfNeeded(
	ctx context.Context,
	log *log.Logger,
	mtime time.Time,
	url string,
) (io.ReadCloser, time.Time, error) {
	var newMtime time.Time
	var ifModifiedSince string
	if !mtime.IsZero() {
		ifModifiedSince = strings.Replace(mtime.UTC().Format(time.RFC1123), "UTC", "GMT", 1)
	}

	req, baseErr := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if baseErr != nil {
		return nil, newMtime, fault.Wrap(
			baseErr,
			fmsg.Withf("could not create HTTP request for %s", url),
		)
	}

	req.Header.Set("User-Agent", fmt.Sprintf("Searchix %s", config.Version))

	if ifModifiedSince != "" {
		req.Header.Set("If-Modified-Since", ifModifiedSince)
	}
	res, baseErr := http.DefaultClient.Do(req)
	if baseErr != nil {
		return nil, newMtime, fault.Wrap(
			baseErr,
			fmsg.Withf("could not make HTTP request to %s", url),
		)
	}

	var body io.ReadCloser
	var err error
	switch res.StatusCode {
	case http.StatusNotModified:
		newMtime = mtime

		return nil, newMtime, nil
	case http.StatusOK:
		var baseErr error
		newMtime, baseErr = time.Parse(time.RFC1123, res.Header.Get("Last-Modified"))
		if baseErr != nil {
			log.Warn(
				"could not parse Last-Modified header from response",
				"value",
				res.Header.Get("Last-Modified"),
			)
			newMtime = time.Now()
		}

		switch ce := res.Header.Get("Content-Encoding"); ce {
		case "br":
			log.Debug("using brotli encoding")
			body = newBrotliReader(res.Body)
		case "", "identity", "gzip":
			body = res.Body
		default:
			err = fault.Newf("cannot handle a body with content-encoding %s", ce)
		}
	default:
		err = fault.Newf("got response code %d, don't know what to do", res.StatusCode)
	}

	return NewReadCloser(body), newMtime, err
}
