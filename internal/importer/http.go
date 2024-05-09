package importer

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"searchix/internal/file"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func fetchFileIfNeeded(ctx context.Context, path string, url string) (needed bool, err error) {
	stat, err := file.StatIfExists(path)
	if err != nil {
		return false, errors.WithMessagef(err, "could not stat file %s", path)
	}

	var mtime string
	if stat != nil {
		mtime = strings.Replace(stat.ModTime().UTC().Format(time.RFC1123), "UTC", "GMT", 1)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return false, errors.WithMessagef(err, "could not create HTTP request for %s", url)
	}

	if mtime != "" {
		req.Header.Set("If-Modified-Since", mtime)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, errors.WithMessagef(err, "could not make HTTP request to %s", url)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusNotModified:
		needed = false
	case http.StatusOK:
		newMtime, err := time.Parse(time.RFC1123, res.Header.Get("Last-Modified"))
		if err != nil {
			slog.Warn(
				"could not parse Last-Modified header from response",
				"value",
				res.Header.Get("Last-Modified"),
			)
		}
		err = file.WriteToFile(path, res.Body)
		if err != nil {
			return false, errors.WithMessagef(err, "could not write response body to file %s", path)
		}
		err = os.Chtimes(path, time.Time{}, newMtime)
		if err != nil {
			slog.Warn("could not update mtime on file", "file", path)
		}
		needed = true
	default:
		return false, fmt.Errorf("got response code %d, don't know what to do", res.StatusCode)
	}

	return needed, nil
}
