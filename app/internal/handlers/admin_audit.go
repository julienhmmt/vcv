package handlers

import (
	"net/http"
	"strings"

	"vcv/internal/httputil"
	"vcv/internal/logger"
	"vcv/internal/middleware"
)

// ProxyConfig carries reverse-proxy trust settings (app.trust_proxy and
// app.trusted_auth_header). A trusted auth header is honored only when
// TrustProxy is true; otherwise it is attacker-controlled input and ignored.
type ProxyConfig struct {
	TrustProxy        bool
	TrustedAuthHeader string
}

// adminAuditor emits structured audit lines (event_category=audit) for admin
// security events: logins, logouts, and state-changing admin calls. It
// carries proxy trust settings so client IPs resolve the same way the login
// rate limiter sees them, and proxy identities are only ever read from a
// trusted header. Usernames are safe to log; passwords and tokens
// must never appear in audit fields.
type adminAuditor struct {
	proxy ProxyConfig
}

// proxyUser returns the reverse-proxy authenticated identity for the request,
// or "" when unconfigured, untrusted, or absent. Values are trimmed and
// capped so a hostile upstream cannot blow up log lines.
func (a adminAuditor) proxyUser(r *http.Request) string {
	if !a.proxy.TrustProxy || a.proxy.TrustedAuthHeader == "" {
		return ""
	}
	user := strings.TrimSpace(r.Header.Get(a.proxy.TrustedAuthHeader))
	if len(user) > 256 {
		user = user[:256]
	}
	return user
}

// log records one admin audit event. ok reports the outcome; fields carries
// small event-specific attributes such as username, reason, or vault id.
func (a adminAuditor) log(r *http.Request, action string, ok bool, fields map[string]string) {
	event := logger.Get().Info().
		Str("event_category", "audit").
		Str("event", action).
		Str("client_ip", httputil.ClientIP(r, a.proxy.TrustProxy)).
		Str("request_id", middleware.GetRequestID(r.Context())).
		Bool("success", ok)
	if user := a.proxyUser(r); user != "" {
		event = event.Str("proxy_user", user)
	}
	for key, value := range fields {
		event = event.Str(key, value)
	}
	event.Msg("admin " + action)
}
