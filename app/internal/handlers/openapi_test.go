package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vcv/internal/version"
)

func TestOpenAPISpec_ServesDocument(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/openapi.json", nil)
	w := httptest.NewRecorder()

	OpenAPISpec(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var doc map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &doc))

	assert.Equal(t, "3.0.3", doc["openapi"])
	info, ok := doc["info"].(map[string]any)
	require.True(t, ok, "info object present")
	assert.Equal(t, version.Version, info["version"])

	paths, ok := doc["paths"].(map[string]any)
	require.True(t, ok, "paths object present")
	for _, p := range []string{
		"/api/certs",
		"/api/certs/{id}/details",
		"/api/certs/{id}/ca",
		"/api/certs/{id}/pem",
		"/api/status",
		"/api/config",
		"/api/version",
		"/api/openapi.json",
		"/api/health",
		"/api/ready",
		"/api/i18n",
		"/metrics",
		"/api/admin/session",
		"/api/admin/login",
		"/api/admin/logout",
		"/api/admin/settings",
	} {
		assert.Contains(t, paths, p, "spec documents %s", p)
	}
}
