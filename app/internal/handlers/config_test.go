package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"vcv/config"
)

// failingResponseWriter is a ResponseWriter that always fails on Write
type failingResponseWriter struct {
	header http.Header
}

func (w *failingResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *failingResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func (w *failingResponseWriter) WriteHeader(statusCode int) {
}

func TestGetConfig_Success(t *testing.T) {
	cfg := config.Config{
		ExpirationThresholds: config.ExpirationThresholds{
			Critical: 7,
			Warning:  30,
		},
	}

	handler := GetConfig(cfg)
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	var resp ConfigResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ExpirationThresholds.Critical != 7 {
		t.Errorf("expected critical threshold 7, got %d", resp.ExpirationThresholds.Critical)
	}
	if resp.ExpirationThresholds.Warning != 30 {
		t.Errorf("expected warning threshold 30, got %d", resp.ExpirationThresholds.Warning)
	}
}

func TestGetConfig_CustomValues(t *testing.T) {
	cfg := config.Config{
		ExpirationThresholds: config.ExpirationThresholds{
			Critical: 14,
			Warning:  60,
		},
	}

	handler := GetConfig(cfg)
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp ConfigResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ExpirationThresholds.Critical != 14 {
		t.Errorf("expected critical threshold 14, got %d", resp.ExpirationThresholds.Critical)
	}
	if resp.ExpirationThresholds.Warning != 60 {
		t.Errorf("expected warning threshold 60, got %d", resp.ExpirationThresholds.Warning)
	}
}

func TestGetConfig_EncodingError(t *testing.T) {
	cfg := config.Config{
		ExpirationThresholds: config.ExpirationThresholds{Critical: 7, Warning: 30},
		Vault:                config.VaultConfig{PKIMounts: []string{"pki"}},
	}

	handler := GetConfig(cfg)

	// Create a response writer that will fail on write
	w := &failingResponseWriter{}
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)

	// This should not panic
	handler(w, req)
}
