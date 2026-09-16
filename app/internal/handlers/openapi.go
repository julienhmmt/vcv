package handlers

import (
	"encoding/json"
	"net/http"

	"vcv/internal/version"
)

// OpenAPISpec serves a curated OpenAPI 3.0 document describing the vcv JSON
// API. It is generated from code (not from reflection) so it stays readable
// and never leaks internal types: update it alongside route changes.
// Public inventory endpoints are unauthenticated by design (private-network
// threat model, see app/README.md); admin endpoints require the session
// cookie set by POST /api/admin/login.
func OpenAPISpec(w http.ResponseWriter, _ *http.Request) {
	spec := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "VaultCertsViewer API",
			"version":     version.Version,
			"description": "Read-only certificate inventory over Vault/OpenBao PKI mounts, plus an optional session-authenticated admin API. Serve the UI from / and /admin; talk JSON everywhere else.",
		},
		"paths": openAPIPaths(),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(spec)
}

func openAPIPaths() map[string]any {
	get := func(summary, description string) map[string]any {
		return map[string]any{
			"get": map[string]any{
				"summary":     summary,
				"description": description,
				"responses": map[string]any{
					"200": map[string]any{"description": "OK"},
				},
			},
		}
	}
	return map[string]any{
		"/api/certs": map[string]any{
			"get": map[string]any{
				"summary":     "List certificates",
				"description": "Partial-success envelope: certificates plus per-vault errors. Supports search/filter/sort/pagination; without query params returns the full list. Response adds total, page, page_size, total_pages, and status counts.",
				"parameters": []any{
					map[string]any{"name": "mounts", "in": "query", "schema": map[string]any{"type": "string"}, "description": "Comma-separated PKI mounts to include."},
					map[string]any{"name": "search", "in": "query", "schema": map[string]any{"type": "string"}, "description": "Case-insensitive substring over common name, serial, and SANs."},
					map[string]any{"name": "status", "in": "query", "schema": map[string]any{"type": "string"}, "description": "Comma-separated tiers: valid,warning,critical,expired,revoked."},
					map[string]any{"name": "cert_type", "in": "query", "schema": map[string]any{"type": "string"}, "description": "One of all,machine,user,both,unknown."},
					map[string]any{"name": "sort", "in": "query", "schema": map[string]any{"type": "string"}, "description": "One of commonName,expiresAt,vault,pki."},
					map[string]any{"name": "order", "in": "query", "schema": map[string]any{"type": "string"}, "description": "asc or desc (default asc)."},
					map[string]any{"name": "page", "in": "query", "schema": map[string]any{"type": "integer"}, "description": "1-based page number."},
					map[string]any{"name": "page_size", "in": "query", "schema": map[string]any{"type": "string"}, "description": "Items per page (numeric, capped at 1000) or all."},
				},
				"responses": map[string]any{"200": map[string]any{"description": "OK"}, "304": map[string]any{"description": "Not Modified (ETag match)"}, "400": map[string]any{"description": "Invalid query parameter"}},
			},
		},
		"/api/certs/{id}/details": get("Certificate details", "Detailed view for one certificate. The id path segment must be URL-encoded."),
		"/api/certs/{id}/ca":      get("Signing authority", "Intermediate/root CA chain for one certificate."),
		"/api/certs/{id}/pem":     get("Certificate PEM", "Public X.509 PEM as JSON. Private keys are never returned."),
		"/api/status":             get("Vault status", "Per-vault connectivity with sanitized error strings, plus version and admin_api_enabled."),
		"/api/config":             get("Public configuration", "Expiration thresholds, mounts, and other public settings."),
		"/api/version":            get("Version info", "Build version. Served for LAN inventory clients; block at the reverse proxy if you prefer."),
		"/api/openapi.json":       get("API specification", "This document."),
		"/api/health":             get("Liveness probe", "Always 200 when the process is alive."),
		"/api/ready":              get("Readiness probe", "200 when ready to serve. Stays green when only the admin API is disabled."),
		"/api/i18n": map[string]any{
			"get": map[string]any{
				"summary":     "UI translations",
				"description": "Message bundle for the requested language.",
				"parameters": []any{
					map[string]any{"name": "lang", "in": "query", "schema": map[string]any{"type": "string"}},
				},
				"responses": map[string]any{"200": map[string]any{"description": "OK"}},
			},
		},
		"/metrics":           get("Prometheus metrics", "Prometheus text exposition format. Scrape on private networks only."),
		"/api/admin/session": get("Admin session status", "Whether the request carries an authenticated admin session."),
		"/api/admin/login": map[string]any{
			"post": map[string]any{
				"summary":     "Admin login",
				"description": "JSON credentials; sets the vcv_admin_session cookie. Rate-limited per client IP.",
				"responses":   map[string]any{"200": map[string]any{"description": "Authenticated"}, "401": map[string]any{"description": "Invalid credentials"}, "429": map[string]any{"description": "Too many attempts"}},
			},
		},
		"/api/admin/logout": map[string]any{
			"post": map[string]any{
				"summary": "Admin logout", "description": "Clears the session cookie.",
				"responses": map[string]any{"200": map[string]any{"description": "OK"}},
			},
		},
		"/api/admin/docs":     get("Admin documentation", "Embedded operational docs HTML. Requires admin session."),
		"/api/admin/settings": get("Admin settings", "Masked settings JSON (tokens masked). Requires admin session. PUT accepts the same shape; blank/masked tokens preserve the stored value."),
		"/api/admin/vault": map[string]any{
			"post": map[string]any{
				"summary": "Add Vault instance", "description": "Requires admin session.",
				"responses": map[string]any{"200": map[string]any{"description": "OK"}},
			},
		},
		"/api/admin/vault/{id}": map[string]any{
			"delete": map[string]any{
				"summary": "Remove Vault instance", "description": "Requires admin session.",
				"responses": map[string]any{"200": map[string]any{"description": "OK"}},
			},
		},
		"/api/cache/invalidate": map[string]any{
			"post": map[string]any{
				"summary": "Invalidate cache", "description": "Requires admin session.",
				"responses": map[string]any{"204": map[string]any{"description": "Invalidated"}, "503": map[string]any{"description": "No cache client"}},
			},
		},
	}
}
