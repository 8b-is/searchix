package server

import (
	"net/http"

	cfg "go.alanpearce.eu/searchix/internal/config"
)

func AddHeadersMiddleware(next http.Handler, config *cfg.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for h, v := range config.Web.Headers {
			w.Header().Add(h, v)
		}
		w.Header().Add("Content-Security-Policy", config.Web.ContentSecurityPolicy.String())
		w.Header().Add("Server", "go.alanpearce.eu/searchix/"+cfg.Version)

		next.ServeHTTP(w, r)
	})
}
