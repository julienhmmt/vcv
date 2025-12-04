package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"vcv/internal/logger"
	"vcv/internal/vault"
	"vcv/middleware"
)

type revokeRequest struct {
	WriteToken string `json:"writeToken"`
}

func RegisterCertRoutes(r chi.Router, vaultClient vault.Client, enableRevoke bool) {
	r.Get("/api/certs", func(w http.ResponseWriter, req *http.Request) {
		certificates, err := vaultClient.ListCertificates(req.Context())
		if err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to list certificates")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(certificates); err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to encode certificates response")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})

	r.Post("/api/certs/{id}/revoke", func(w http.ResponseWriter, req *http.Request) {
		if !enableRevoke {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusForbidden, nil).
				Str("request_id", requestID).
				Msg("revocation disabled by configuration")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		serialNumber := chi.URLParam(req, "id")
		if serialNumber == "" {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusBadRequest, nil).
				Str("request_id", requestID).
				Msg("missing certificate id in path")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var body revokeRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusBadRequest, err).
				Str("request_id", requestID).
				Msg("invalid revoke request payload")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if body.WriteToken == "" {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusBadRequest, nil).
				Str("request_id", requestID).
				Msg("missing write token in revoke request")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err := vaultClient.RevokeCertificate(req.Context(), serialNumber, body.WriteToken); err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to revoke certificate")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}
