package server

import (
	"net/http"

	"alin.ovh/x/log"
	"gitlab.com/tozd/go/errors"
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
	logger          *log.Logger
	enabled         bool
}

func wrapHandlerWithLogging(wrappedHandler http.Handler, opts wrappedHandlerOptions) http.Handler {
	if !opts.enabled {
		return wrappedHandler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme := r.Header.Get("X-Forwarded-Proto")
		if scheme == "" {
			scheme = "http"
		}
		lw := NewLoggingResponseWriter(w)
		wrappedHandler.ServeHTTP(lw, r)
		statusCode := lw.statusCode
		opts.logger.Info(
			"http request",
			"scheme",
			scheme,
			"method",
			r.Method,
			"code",
			statusCode,
			"path",
			r.URL.Path,
			"location",
			lw.Header().Get("Location"),
		)
	})
}
