package httputil

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP extracts the client IP address from the request.
// When trustProxy is true, X-Forwarded-For and X-Real-IP headers are checked first.
func ClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
		if forwarded != "" {
			parts := strings.Split(forwarded, ",")
			if len(parts) > 0 {
				value := strings.TrimSpace(parts[0])
				if value != "" {
					return value
				}
			}
		}
		realIP := strings.TrimSpace(r.Header.Get("X-Real-IP"))
		if realIP != "" {
			return realIP
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}
