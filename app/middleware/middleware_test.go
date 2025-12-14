package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"vcv/middleware"
)

func TestSecurityHeaders(t *testing.T) {
	handler := middleware.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	tests := []struct {
		header   string
		expected string
	}{
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := rec.Header().Get(tt.header)
			if got != tt.expected {
				t.Errorf("expected %s=%q, got %q", tt.header, tt.expected, got)
			}
		})
	}

	csp := rec.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Error("expected Content-Security-Policy header to be set")
	}
}

func TestRequestID_GeneratesID(t *testing.T) {
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetRequestID(r.Context())
		if requestID == "" {
			t.Error("expected request ID in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header to be set")
	}
}

func TestRequestID_UsesProvidedID(t *testing.T) {
	providedID := "test-request-id-123"
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetRequestID(r.Context())
		if requestID != providedID {
			t.Errorf("expected request ID %q, got %q", providedID, requestID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", providedID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") != providedID {
		t.Errorf("expected X-Request-ID header %q, got %q", providedID, rec.Header().Get("X-Request-ID"))
	}
}

func TestRecoverer_HandlesNormalRequest(t *testing.T) {
	handler := middleware.Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRecoverer_HandlesPanic(t *testing.T) {
	handler := middleware.Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestCORS_NoOrigin(t *testing.T) {
	config := middleware.DefaultCORSConfig()
	handler := middleware.CORS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("expected no CORS headers when Origin is not set")
	}
}

func TestCORS_WithOrigin(t *testing.T) {
	config := middleware.DefaultCORSConfig()
	handler := middleware.CORS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Errorf("expected Access-Control-Allow-Origin=http://example.com, got %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_Preflight(t *testing.T) {
	config := middleware.DefaultCORSConfig()
	handler := middleware.CORS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d for preflight, got %d", http.StatusNoContent, rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected Access-Control-Allow-Methods header for preflight")
	}
}

func TestLogger_LogsRequest(t *testing.T) {
	// Create a handler that will be wrapped by Logger
	var writeErr error
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("response"))
		writeErr = err
	})

	handler := middleware.Logger(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if writeErr != nil {
		t.Fatalf("expected no write error, got %v", writeErr)
	}

	// The Logger middleware should pass through the request
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Add a new assertion to check the response body
	if rec.Body.String() != "response" {
		t.Errorf("expected response body %q, got %q", "response", rec.Body.String())
	}
}
