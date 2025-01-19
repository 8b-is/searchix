package sentryhttp

import (
	"net/http"

	sentryhttp "github.com/getsentry/sentry-go/http"
)

type ServeMux struct {
	sentryHandler *sentryhttp.Handler
	*http.ServeMux
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		sentryHandler: sentryhttp.New(sentryhttp.Options{
			Repanic: true,
		}),
		ServeMux: http.NewServeMux(),
	}
}

func (sm *ServeMux) Handle(pattern string, handler http.Handler) {
	sm.ServeMux.Handle(pattern, sm.sentryHandler.Handle(handler))
}

func (sm *ServeMux) HandleFunc(pattern string, handler http.HandlerFunc) {
	sm.ServeMux.HandleFunc(pattern, sm.sentryHandler.HandleFunc(handler))
}
