package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"vcv/internal/handlers"
)

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
	}{
		{"returns 200 OK", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			rec := httptest.NewRecorder()

			handlers.HealthCheck(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}
