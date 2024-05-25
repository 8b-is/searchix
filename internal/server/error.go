package server

import (
	"log/slog"
	"net/http"
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
		indexData := TemplateData{
			ExtraHeadHTML: config.Web.ExtraHeadHTML,
			Sources:       config.Importer.Sources,
			Code:          code,
			Message:       message,
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Del("Vary")
		w.WriteHeader(code)
		if r.Header.Get("Fetch") == "true" {
			err = templates["error"].ExecuteTemplate(w, "main", indexData)
		} else {
			err = templates["error"].Execute(w, indexData)
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
