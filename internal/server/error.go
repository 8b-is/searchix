package server

import (
	"net/http"

	"alin.ovh/searchix/frontend"
	"alin.ovh/searchix/internal/components"
	"alin.ovh/searchix/internal/config"
	"alin.ovh/x/log"
)

func createErrorHandler(
	config *config.Config,
	assets *frontend.AssetCollection,
	log *log.Logger,
) func(http.ResponseWriter, *http.Request, string, int) {
	return func(w http.ResponseWriter, r *http.Request, message string, code int) {
		var err error
		if message == "" {
			message = http.StatusText(code)
		}
		indexData := components.TemplateData{
			ExtraHeadHTML: config.Web.ExtraHeadHTML,
			Sources:       sources,
			Code:          code,
			Message:       message,
			Assets:        assets,
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Del("Vary")
		w.WriteHeader(code)
		if r.Header.Get("Fetch") == "true" {
			err = components.Error(indexData).Render(w)
		} else {
			err = components.ErrorPage(indexData).Render(w)
		}
		if err != nil {
			log.Error(
				"error rendering error page template",
				"error",
				err,
				"code",
				code,
				"message",
				message,
			)
			http.Error(w, message, code)
		}
	}
}
