package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type LoggingResponseWriter struct {
	wroteHeader bool
	http.ResponseWriter
	statusCode int
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	if !lrw.wroteHeader {
		lrw.ResponseWriter.WriteHeader(code)
		lrw.wroteHeader = true
	}
}

func (lrw *LoggingResponseWriter) Write(b []byte) (int, error) {
	if !lrw.wroteHeader {
		lrw.statusCode = http.StatusOK
		lrw.wroteHeader = true
	}

	count, err := lrw.ResponseWriter.Write(b)
	if err != nil {
		return count, errors.Wrap(err, "failed to write response")
	}

	return count, nil
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{false, w, http.StatusOK}
}

type wrappedHandlerOptions struct {
	defaultHostname string
	logger          io.Writer
}

func wrapHandlerWithLogging(wrappedHandler http.Handler, opts wrappedHandlerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme := r.Header.Get("X-Forwarded-Proto")
		if scheme == "" {
			scheme = "http"
		}
		host := r.Host
		if host == "" {
			host = opts.defaultHostname
		}
		lw := NewLoggingResponseWriter(w)
		wrappedHandler.ServeHTTP(lw, r)
		statusCode := lw.statusCode
		fmt.Fprintf(
			opts.logger,
			"%s %s %d %s %s %s\n",
			scheme,
			r.Method,
			statusCode,
			host,
			r.URL.Path,
			lw.Header().Get("Location"),
		)
	})
}
