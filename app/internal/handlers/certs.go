package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"

	"vcv/internal/certs"
	"vcv/internal/logger"
	"vcv/internal/vault"
	"vcv/middleware"
)

const mountsAllSentinel = "__all__"

func RegisterCertRoutes(r chi.Router, vaultClient vault.Client) {
	r.Get("/api/certs", func(w http.ResponseWriter, req *http.Request) {
		// Parse mount filter from query parameters
		selectedMounts := parseMountsQueryParam(req.URL.Query())

		certificates, err := vaultClient.ListCertificates(req.Context())
		if err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to list certificates")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Filter certificates by selected mounts
		filteredCertificates := filterCertificatesByMounts(certificates, selectedMounts)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(filteredCertificates); err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to encode certificates response")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		requestID := middleware.GetRequestID(req.Context())
		logger.HTTPEvent(req.Method, req.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Int("count", len(filteredCertificates)).
			Strs("mounts", selectedMounts).
			Msg("listed certificates")
	})

	r.Get("/api/certs/{id}/details", func(w http.ResponseWriter, req *http.Request) {
		certificateIDParam := chi.URLParam(req, "id")
		if certificateIDParam == "" {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusBadRequest, nil).
				Str("request_id", requestID).
				Msg("missing certificate id in path")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		certificateID, err := url.PathUnescape(certificateIDParam)
		if err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusBadRequest, err).
				Str("request_id", requestID).
				Msg("failed to decode certificate id")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		details, err := vaultClient.GetCertificateDetails(req.Context(), certificateID)
		if err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Str("serial_number", certificateID).
				Msg("failed to get certificate details")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(details); err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to encode certificate details response")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		requestID := middleware.GetRequestID(req.Context())
		logger.HTTPEvent(req.Method, req.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Str("serial_number", certificateID).
			Msg("fetched certificate details")
	})

	r.Get("/api/certs/{id}/pem", func(w http.ResponseWriter, req *http.Request) {
		certificateIDParam := chi.URLParam(req, "id")
		if certificateIDParam == "" {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusBadRequest, nil).
				Str("request_id", requestID).
				Msg("missing certificate id in path")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		certificateID, err := url.PathUnescape(certificateIDParam)
		if err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusBadRequest, err).
				Str("request_id", requestID).
				Msg("failed to decode certificate id")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		pemResponse, err := vaultClient.GetCertificatePEM(req.Context(), certificateID)
		if err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Str("serial_number", certificateID).
				Msg("failed to get certificate PEM")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(pemResponse); err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to encode certificate PEM response")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		requestID := middleware.GetRequestID(req.Context())
		logger.HTTPEvent(req.Method, req.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Str("serial_number", certificateID).
			Msg("served certificate PEM")
	})

	r.Get("/api/certs/{id}/pem/download", func(w http.ResponseWriter, req *http.Request) {
		certificateIDParam := chi.URLParam(req, "id")
		if certificateIDParam == "" {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusBadRequest, nil).
				Str("request_id", requestID).
				Msg("missing certificate id in path")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		certificateID, err := url.PathUnescape(certificateIDParam)
		if err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusBadRequest, err).
				Str("request_id", requestID).
				Msg("failed to decode certificate id")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		pemResponse, err := vaultClient.GetCertificatePEM(req.Context(), certificateID)
		if err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Str("serial_number", certificateID).
				Msg("failed to get certificate PEM")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		filename := buildPEMDownloadFilename(pemResponse.SerialNumber)
		w.Header().Set("Content-Type", "application/x-pem-file")
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		w.WriteHeader(http.StatusOK)
		if _, writeErr := w.Write([]byte(pemResponse.PEM)); writeErr != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, writeErr).
				Str("request_id", requestID).
				Msg("failed to write certificate PEM download")
			return
		}
		requestID := middleware.GetRequestID(req.Context())
		logger.HTTPEvent(req.Method, req.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Str("serial_number", certificateID).
			Msg("downloaded certificate PEM")
	})

	r.Post("/api/cache/invalidate", func(w http.ResponseWriter, req *http.Request) {
		vaultClient.InvalidateCache()
		w.WriteHeader(http.StatusNoContent)
		requestID := middleware.GetRequestID(req.Context())
		logger.HTTPEvent(req.Method, req.URL.Path, http.StatusNoContent, 0).
			Str("request_id", requestID).
			Msg("invalidated cache")
	})
}

func parseMountsQueryParam(query url.Values) []string {
	_, present := query["mounts"]
	if !present {
		return nil
	}
	raw := strings.TrimSpace(query.Get("mounts"))
	if raw == mountsAllSentinel {
		return nil
	}
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	mounts := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		mounts = append(mounts, trimmed)
	}
	return mounts
}

// filterCertificatesByMounts filters certificates by the specified mounts
func filterCertificatesByMounts(certificates []certs.Certificate, selectedMounts []string) []certs.Certificate {
	if selectedMounts == nil {
		return certificates
	}
	if len(selectedMounts) == 0 {
		return []certs.Certificate{}
	}

	selectedSet := make(map[string]struct{}, len(selectedMounts))
	selectedLegacyMounts := make(map[string]struct{}, len(selectedMounts))
	for _, selectedMount := range selectedMounts {
		trimmed := strings.TrimSpace(selectedMount)
		if trimmed == "" {
			continue
		}
		selectedSet[trimmed] = struct{}{}
		if !strings.Contains(trimmed, "|") {
			selectedLegacyMounts[trimmed] = struct{}{}
		}
	}

	filtered := make([]certs.Certificate, 0, len(certificates))
	for _, certificate := range certificates {
		vaultMountKey, mountName := extractVaultMountFromCertificateID(certificate.ID)
		if vaultMountKey == "" && mountName == "" {
			continue
		}
		if _, ok := selectedSet[vaultMountKey]; ok {
			filtered = append(filtered, certificate)
			continue
		}
		if _, ok := selectedLegacyMounts[mountName]; ok {
			filtered = append(filtered, certificate)
		}
	}
	return filtered
}

func extractVaultMountFromCertificateID(value string) (string, string) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", ""
	}
	vaultID := ""
	mountSerial := trimmed
	if parts := strings.SplitN(trimmed, "|", 2); len(parts) == 2 {
		vaultID = strings.TrimSpace(parts[0])
		mountSerial = strings.TrimSpace(parts[1])
	}
	parts := strings.SplitN(mountSerial, ":", 2)
	if len(parts) < 2 {
		return "", ""
	}
	mountName := strings.TrimSpace(parts[0])
	if mountName == "" {
		return "", ""
	}
	if vaultID == "" {
		return mountName, mountName
	}
	return vaultID + "|" + mountName, mountName
}

func buildPEMDownloadFilename(serialNumber string) string {
	replacer := strings.NewReplacer(":", "-", "/", "-", "\\", "-", "..", "-")
	safe := replacer.Replace(serialNumber)
	if safe == "" {
		return "certificate.pem"
	}
	return "certificate-" + safe + ".pem"
}
