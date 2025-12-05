package handlers

import (
	"net/http"

	"vcv/internal/logger"
	"vcv/middleware"
)

func ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	logger.HTTPEvent(r.Method, r.URL.Path, http.StatusOK, 0).
		Str("request_id", requestID).
		Msg("readiness check")
	w.WriteHeader(http.StatusOK)
}
