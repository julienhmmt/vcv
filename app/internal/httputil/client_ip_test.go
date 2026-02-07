package httputil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"vcv/internal/httputil"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		forwarded  string
		realIP     string
		remoteAddr string
		trustProxy bool
		expected   string
	}{
		{name: "XForwardedFor_trusted", forwarded: "10.0.0.1, 10.0.0.2", trustProxy: true, remoteAddr: "192.168.1.1:9999", expected: "10.0.0.1"},
		{name: "XForwardedFor_untrusted", forwarded: "10.0.0.1", trustProxy: false, remoteAddr: "192.168.1.1:9999", expected: "192.168.1.1"},
		{name: "XRealIP_trusted", realIP: "10.0.0.3", trustProxy: true, remoteAddr: "192.168.1.1:9999", expected: "10.0.0.3"},
		{name: "XRealIP_untrusted", realIP: "10.0.0.3", trustProxy: false, remoteAddr: "192.168.1.1:9999", expected: "192.168.1.1"},
		{name: "RemoteAddr_with_port", remoteAddr: "203.0.113.5:4321", trustProxy: true, expected: "203.0.113.5"},
		{name: "RemoteAddr_without_port", remoteAddr: "203.0.113.5", trustProxy: true, expected: "203.0.113.5"},
		{name: "empty_forwarded_falls_through", forwarded: "  ", trustProxy: true, remoteAddr: "1.2.3.4:80", expected: "1.2.3.4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.forwarded != "" {
				req.Header.Set("X-Forwarded-For", tt.forwarded)
			}
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}
			got := httputil.ClientIP(req, tt.trustProxy)
			if got != tt.expected {
				t.Errorf("ClientIP() = %q, want %q", got, tt.expected)
			}
		})
	}
}
