package handlers

import (
	"net/http"

	"vcv/internal/httputil"
	"vcv/internal/logger"
	"vcv/internal/middleware"
)

// adminAuditor emits structured audit lines (event_category=audit) for admin
// security events: logins, logouts, and state-changing admin calls. It
// carries trustProxy so client IPs resolve the same way the login rate
// limiter sees them. Usernames are safe to log; passwords and tokens
// must never appear in audit fields.
type adminAuditor struct {
	trustProxy bool
}

// log records one admin audit event. ok reports the outcome; fields carries
// small event-specific attributes such as username, reason, or vault id.
func (a adminAuditor) log(r *http.Request, action string, ok bool, fields map[string]string) {
	event := logger.Get().Info().
		Str("event_category", "audit").
		Str("event", action).
		Str("client_ip", httputil.ClientIP(r, a.trustProxy)).
		Str("request_id", middleware.GetRequestID(r.Context())).
		Bool("success", ok)
	for key, value := range fields {
		event = event.Str(key, value)
	}
	event.Msg("admin " + action)
}
