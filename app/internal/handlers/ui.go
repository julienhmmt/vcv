package handlers

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"

	"vcv/internal/certs"
	"vcv/internal/i18n"
	"vcv/internal/logger"
	"vcv/internal/vault"
	"vcv/internal/version"
	"vcv/middleware"
)

type certDetailsTemplateData struct {
	Certificate   certs.DetailedCertificate
	Messages      i18n.Messages
	KeySummary    string
	UsageSummary  string
	Language      i18n.Language
	CertificateID string
}

type footerStatusTemplateData struct {
	VersionText string
	VaultText   string
	VaultClass  string
	VaultTitle  string
}

func RegisterUIRoutes(router chi.Router, vaultClient vault.Client, webFS fs.FS) {
	templates, err := template.ParseFS(webFS, "templates/*.html")
	if err != nil {
		panic(err)
	}
	router.Get("/ui/status", func(w http.ResponseWriter, r *http.Request) {
		language := resolveLanguage(r)
		messages := i18n.MessagesForLanguage(language)
		vaultClass := "vcv-footer-pill"
		vaultTitle := ""
		vaultText := messages.FooterVaultLoading
		if vaultErr := vaultClient.CheckConnection(r.Context()); vaultErr != nil {
			vaultClass = "vcv-footer-pill vcv-footer-pill-error"
			vaultText = messages.FooterVaultDisconnected
			vaultTitle = vaultErr.Error()
		} else {
			vaultClass = "vcv-footer-pill vcv-footer-pill-ok"
			vaultText = messages.FooterVaultConnected
		}
		data := footerStatusTemplateData{
			VersionText: interpolatePlaceholder(messages.FooterVersion, "version", version.Version),
			VaultText:   vaultText,
			VaultClass:  vaultClass,
			VaultTitle:  vaultTitle,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if err := templates.ExecuteTemplate(w, "footer-status.html", data); err != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to render footer status template")
			return
		}
		requestID := middleware.GetRequestID(r.Context())
		logger.HTTPEvent(r.Method, r.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Msg("rendered footer status")
	})
	router.Get("/ui/certs/{id}/details", func(w http.ResponseWriter, r *http.Request) {
		certificateIDParam := chi.URLParam(r, "id")
		if certificateIDParam == "" {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusBadRequest, nil).
				Str("request_id", requestID).
				Msg("missing certificate id in path")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		certificateID, err := url.PathUnescape(certificateIDParam)
		if err != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusBadRequest, err).
				Str("request_id", requestID).
				Msg("failed to decode certificate id")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		details, err := vaultClient.GetCertificateDetails(r.Context(), certificateID)
		if err != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Str("serial_number", certificateID).
				Msg("failed to get certificate details")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		language := resolveLanguage(r)
		messages := i18n.MessagesForLanguage(language)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		data := certDetailsTemplateData{
			Certificate:   details,
			Messages:      messages,
			KeySummary:    buildKeySummary(details),
			UsageSummary:  buildUsageSummary(details.Usage),
			Language:      language,
			CertificateID: certificateID,
		}
		if err := templates.ExecuteTemplate(w, "cert-details.html", data); err != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to render certificate details template")
			return
		}
		requestID := middleware.GetRequestID(r.Context())
		logger.HTTPEvent(r.Method, r.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Str("serial_number", certificateID).
			Msg("rendered certificate details")
	})
}

func buildKeySummary(details certs.DetailedCertificate) string {
	if details.KeyAlgorithm == "" && details.KeySize == 0 {
		return "—"
	}
	if details.KeySize == 0 {
		return details.KeyAlgorithm
	}
	if details.KeyAlgorithm == "" {
		return fmt.Sprintf("%d", details.KeySize)
	}
	return fmt.Sprintf("%s %d", details.KeyAlgorithm, details.KeySize)
}

func buildUsageSummary(usages []string) string {
	trimmed := make([]string, 0, len(usages))
	for _, usage := range usages {
		value := strings.TrimSpace(usage)
		if value == "" {
			continue
		}
		trimmed = append(trimmed, value)
	}
	if len(trimmed) == 0 {
		return "—"
	}
	return strings.Join(trimmed, ", ")
}

func interpolatePlaceholder(templateValue, key, value string) string {
	replaced := strings.ReplaceAll(templateValue, "{{"+key+"}}", value)
	return strings.ReplaceAll(replaced, "{{ "+key+" }}", value)
}
