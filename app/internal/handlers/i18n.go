package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"

	"vcv/internal/i18n"
	"vcv/internal/logger"
	"vcv/middleware"

	"github.com/go-chi/chi/v5"
)

// RegisterI18nRoutes exposes a small JSON API for UI translations.
func RegisterI18nRoutes(router chi.Router) {
	router.Get("/api/i18n", func(writer http.ResponseWriter, request *http.Request) {
		language := resolveLanguage(request)
		payload := i18n.Response{
			Language: language,
			Messages: i18n.MessagesForLanguage(language),
		}
		writer.Header().Set("Content-Type", "application/json")
		encodeError := json.NewEncoder(writer).Encode(payload)
		if encodeError != nil {
			requestID := middleware.GetRequestID(request.Context())
			logger.HTTPError(request.Method, request.URL.Path, http.StatusInternalServerError, encodeError).
				Str("request_id", requestID).
				Msg("failed to encode i18n response")
			writer.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func resolveLanguage(request *http.Request) i18n.Language {
	queryLanguage := request.URL.Query().Get("lang")
	if queryLanguage != "" {
		language, ok := i18n.FromQueryLanguage(queryLanguage)
		if ok {
			return language
		}
	}
	currentURL := request.Header.Get("HX-Current-URL")
	if currentURL != "" {
		parsed, err := url.Parse(currentURL)
		if err == nil {
			headerLanguage := parsed.Query().Get("lang")
			if headerLanguage != "" {
				language, ok := i18n.FromQueryLanguage(headerLanguage)
				if ok {
					return language
				}
			}
		}
	}
	return i18n.FromAcceptLanguage(request.Header.Get("Accept-Language"))
}
