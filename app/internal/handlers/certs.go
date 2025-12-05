package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"vcv/internal/logger"
	"vcv/internal/vault"
	"vcv/middleware"
)

func RegisterCertRoutes(r chi.Router, vaultClient vault.Client) {
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
		requestID := middleware.GetRequestID(req.Context())
		logger.HTTPEvent(req.Method, req.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Int("count", len(certificates)).
			Msg("listed certificates")
	})

	r.Get("/api/certs/{id}/details", func(w http.ResponseWriter, req *http.Request) {
		serialNumber := chi.URLParam(req, "id")
		if serialNumber == "" {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusBadRequest, nil).
				Str("request_id", requestID).
				Msg("missing certificate id in path")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		details, err := vaultClient.GetCertificateDetails(req.Context(), serialNumber)
		if err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Str("serial_number", serialNumber).
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
			Str("serial_number", serialNumber).
			Msg("fetched certificate details")
	})

	r.Get("/api/certs/{id}/pem", func(w http.ResponseWriter, req *http.Request) {
		serialNumber := chi.URLParam(req, "id")
		if serialNumber == "" {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusBadRequest, nil).
				Str("request_id", requestID).
				Msg("missing certificate id in path")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		pemResponse, err := vaultClient.GetCertificatePEM(req.Context(), serialNumber)
		if err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Str("serial_number", serialNumber).
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
			Str("serial_number", serialNumber).
			Msg("served certificate PEM")
	})

	r.Post("/api/crl/rotate", func(w http.ResponseWriter, req *http.Request) {
		if err := vaultClient.RotateCRL(req.Context()); err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to rotate CRL")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		requestID := middleware.GetRequestID(req.Context())
		logger.HTTPEvent(req.Method, req.URL.Path, http.StatusNoContent, 0).
			Str("request_id", requestID).
			Msg("rotated CRL")
	})

	r.Get("/api/crl/download", func(w http.ResponseWriter, req *http.Request) {
		crlData, err := vaultClient.GetCRL(req.Context())
		if err != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("Failed to download CRL")
			http.Error(w, "Failed to download CRL", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/x-pem-file")
		w.Header().Set("Content-Disposition", "attachment; filename=crl.pem")
		w.WriteHeader(http.StatusOK)
		if _, writeErr := w.Write(crlData); writeErr != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, writeErr).
				Str("request_id", requestID).
				Msg("failed to write CRL response")
			return
		}
		requestID := middleware.GetRequestID(req.Context())
		logger.HTTPEvent(req.Method, req.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Int("bytes", len(crlData)).
			Msg("downloaded CRL")
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
