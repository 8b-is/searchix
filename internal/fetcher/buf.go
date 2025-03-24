package fetcher

import (
	"bufio"
	"io"

	"gitlab.com/tozd/go/errors"
)

type Reader struct {
	closer io.ReadCloser
	*bufio.Reader
}

func NewReadCloser(body io.ReadCloser) *Reader {
	return &Reader{
		closer: body,
		Reader: bufio.NewReader(body),
	}
}

func (r *Reader) Close() error {
	return errors.WithStack(r.closer.Close())
}
