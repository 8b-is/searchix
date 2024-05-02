package server

import (
	"fmt"
	"io"
	"net/http"
)

type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	// avoids warning: superfluous response.WriteHeader call
	if lrw.statusCode != http.StatusOK {
		lrw.ResponseWriter.WriteHeader(code)
	}
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{w, http.StatusOK}
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
		host := r.Header.Get("Host")
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
