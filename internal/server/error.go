package server

import (
	"log/slog"
	"net/http"

	"searchix/internal/components"
	"searchix/internal/config"
)

func createErrorHandler(
	config *config.Config,
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
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Del("Vary")
		w.WriteHeader(code)
		if r.Header.Get("Fetch") == "true" {
			err = components.Error(indexData).Render(r.Context(), w)
		} else {
			err = components.ErrorPage(indexData).Render(r.Context(), w)
		}
		if err != nil {
			slog.Error(
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
