package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBodyLimit(t *testing.T) {
	handler := BodyLimit(100)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			t.Logf("write error: %v", err)
		}
	}))

	t.Run("nil body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("body within limit", func(t *testing.T) {
		body := strings.NewReader("small body")
		req := httptest.NewRequest(http.MethodPost, "/", body)
		req.ContentLength = int64(len("small body"))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("body exceeds limit", func(t *testing.T) {
		body := strings.NewReader(strings.Repeat("a", 200))
		req := httptest.NewRequest(http.MethodPost, "/", body)
		req.ContentLength = 200
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("expected status 413, got %d", rec.Code)
		}
	})
}

func TestRateLimit_Prune(t *testing.T) {
	config := RateLimitConfig{
		MaxRequests: 10,
		Window:      100 * time.Millisecond,
		MaxEntries:  2,
	}
	limiter := &rateLimiter{
		config:  config,
		entries: make(map[string]rateLimiterEntry),
	}

	// Add entries that will expire
	now := time.Now()
	limiter.entries["ip1"] = rateLimiterEntry{count: 1, resetAt: now.Add(-1 * time.Second)}
	limiter.entries["ip2"] = rateLimiterEntry{count: 1, resetAt: now.Add(-1 * time.Second)}
	limiter.entries["ip3"] = rateLimiterEntry{count: 1, resetAt: now.Add(1 * time.Hour)}

	// Prune should remove expired entries
	limiter.prune(now)

	if len(limiter.entries) != 1 {
		t.Errorf("expected 1 entry after prune, got %d", len(limiter.entries))
	}
	if _, exists := limiter.entries["ip3"]; !exists {
		t.Error("expected ip3 to remain after prune")
	}
}

func TestRateLimit_MaxEntries(t *testing.T) {
	config := RateLimitConfig{
		MaxRequests: 100,
		Window:      1 * time.Minute,
		MaxEntries:  3,
	}
	handler := RateLimit(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create requests from different IPs to trigger pruning
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.0.2." + string(rune('1'+i)) + ":1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func TestShouldSkipRateLimit(t *testing.T) {
	config := RateLimitConfig{
		ExemptPaths:        []string{"/health", "/ready"},
		ExemptPathPrefixes: []string{"/api/public/"},
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"/health", true},
		{"/ready", true},
		{"/api/public/data", true},
		{"/api/public/users", true},
		{"/api/private/data", false},
		{"/other", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			got := shouldSkipRateLimit(req, config)
			if got != tt.expected {
				t.Errorf("shouldSkipRateLimit(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
	ip := clientIP(req)
	if ip != "203.0.113.1" {
		t.Errorf("clientIP with X-Forwarded-For = %q, want %q", ip, "203.0.113.1")
	}
}

func TestClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "203.0.113.2")
	ip := clientIP(req)
	if ip != "203.0.113.2" {
		t.Errorf("clientIP with X-Real-IP = %q, want %q", ip, "203.0.113.2")
	}
}

func TestClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.3:1234"
	ip := clientIP(req)
	if ip != "203.0.113.3" {
		t.Errorf("clientIP with RemoteAddr = %q, want %q", ip, "203.0.113.3")
	}
}

func TestRateLimit_ExceedsLimit(t *testing.T) {
	config := RateLimitConfig{
		MaxRequests: 2,
		Window:      1 * time.Second,
		MaxEntries:  100,
	}
	handler := RateLimit(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First two requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.0.2.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i+1, rec.Code)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", rec.Code)
	}
}

func TestRateLimit_OptionsMethod(t *testing.T) {
	config := DefaultRateLimitConfig()
	handler := RateLimit(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("OPTIONS request should not be rate limited, got status %d", rec.Code)
	}
}

func TestCSRFProtection_SameOrigin(t *testing.T) {
	handler := CSRFProtection(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "https://example.com/api", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("same origin should be allowed, got status %d", rec.Code)
	}
}

func TestCSRFProtection_DifferentOrigin(t *testing.T) {
	handler := CSRFProtection(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "https://example.com/api", nil)
	req.Header.Set("Origin", "https://evil.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("different origin should be forbidden, got status %d", rec.Code)
	}
}

func TestCSRFProtection_SafeMethods(t *testing.T) {
	handler := CSRFProtection(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	safeMethods := []string{http.MethodGet, http.MethodHead, http.MethodOptions}
	for _, method := range safeMethods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("%s should not require CSRF check, got status %d", method, rec.Code)
			}
		})
	}
}
