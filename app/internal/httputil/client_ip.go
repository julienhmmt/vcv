package httputil

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP extracts the client IP address from the request.
// When trustProxy is true, X-Forwarded-For and X-Real-IP headers are checked first.
// X-Forwarded-For uses the right-most entry, the one a trusted proxy appended. This
// assumes exactly one trusted reverse proxy that appends to X-Forwarded-For.
func ClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if value := lastForwardedFor(r); value != "" && net.ParseIP(value) != nil {
			return value
		}
		realIP := strings.TrimSpace(r.Header.Get("X-Real-IP"))
		if realIP != "" && net.ParseIP(realIP) != nil {
			return realIP
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

// lastForwardedFor returns the right-most non-empty X-Forwarded-For entry across all header lines.
func lastForwardedFor(r *http.Request) string {
	parts := strings.Split(strings.Join(r.Header.Values("X-Forwarded-For"), ","), ",")
	for i := len(parts) - 1; i >= 0; i-- {
		if value := strings.TrimSpace(parts[i]); value != "" {
			return value
		}
	}
	return ""
}
