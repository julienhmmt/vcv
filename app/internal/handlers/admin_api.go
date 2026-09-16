package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"vcv/internal/config"
	"vcv/internal/docs"
	vcverrors "vcv/internal/errors"
	"vcv/internal/logger"
	"vcv/internal/middleware"
	"vcv/internal/vault"
)

type adminDocsResponse struct {
	HTML string `json:"html"`
}

type adminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type adminSessionResponse struct {
	Authenticated bool `json:"authenticated"`
}

type adminVaultStatus struct {
	ID        string `json:"id"`
	Enabled   bool   `json:"enabled"`
	Connected bool   `json:"connected"`
}

type adminSettingsResponse struct {
	Settings      config.SettingsFile `json:"settings"`
	VaultStatuses []adminVaultStatus  `json:"vault_statuses"`
}

type adminVaultAddedResponse struct {
	Key   string               `json:"key"`
	Vault config.VaultInstance `json:"vault"`
}

const (
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json"
	// errLoadSettings is the sanitized message returned when the settings
	// file cannot be read or parsed.
	errLoadSettings = "failed to load settings"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(status)
	if body == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(body)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (s *adminSessionStore) loginFromJSON(w http.ResponseWriter, r *http.Request, body adminLoginRequest) (bool, string) {
	if !s.allowLoginAttempt(r) {
		return false, "Too many attempts"
	}
	username := strings.TrimSpace(body.Username)
	if !s.verify(username, body.Password) {
		return false, "Invalid credentials"
	}
	token, err := s.createToken()
	if err != nil {
		return false, "Invalid credentials"
	}
	expiresAt := time.Now().Add(s.sessionTTL)
	s.mu.Lock()
	if oldCookie, cookieErr := r.Cookie(adminCookieName); cookieErr == nil && oldCookie.Value != "" {
		delete(s.sessions, oldCookie.Value)
	}
	s.pruneSessions(time.Now())
	s.sessions[token] = expiresAt
	s.mu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name:     adminCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   s.secureCookies,
		Expires:  expiresAt,
	})
	return true, ""
}

// registerAdminAPIRoutes mounts JSON admin API endpoints for the Svelte admin panel.
// Shares the session and settings stores with the rest of the admin routes.
func registerAdminAPIRoutes(
	router chi.Router,
	sessions *adminSessionStore,
	store *adminSettingsStore,
	vaultStatusClients map[string]vault.Client,
	refreshRegistry func(),
	auditor adminAuditor,
) {
	router.Get("/api/admin/session", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, adminSessionResponse{Authenticated: sessions.isAuthed(r)})
	})

	router.Post("/api/admin/login", func(w http.ResponseWriter, r *http.Request) {
		var body adminLoginRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			auditor.log(r, "admin.login", false, map[string]string{"reason": "invalid request body"})
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		ok, message := sessions.loginFromJSON(w, r, body)
		if !ok {
			auditor.log(r, "admin.login", false, map[string]string{"username": strings.TrimSpace(body.Username), "reason": message})
			writeJSONError(w, http.StatusUnauthorized, message)
			return
		}
		auditor.log(r, "admin.login", true, map[string]string{"username": strings.TrimSpace(body.Username)})
		writeJSON(w, http.StatusOK, adminSessionResponse{Authenticated: true})
	})

	router.Post("/api/admin/logout", func(w http.ResponseWriter, r *http.Request) {
		session := "none"
		if sessions.isAuthed(r) {
			session = "active"
		}
		sessions.logout(w, r)
		auditor.log(r, "admin.logout", true, map[string]string{"session": session})
		w.WriteHeader(http.StatusNoContent)
	})

	router.Group(func(r chi.Router) {
		r.Use(sessions.requireAuth)

		r.Get("/api/admin/docs", func(w http.ResponseWriter, req *http.Request) {
			writeJSON(w, http.StatusOK, adminDocsResponse{HTML: docs.AdminHTML()})
		})

		r.Get("/api/admin/settings", func(w http.ResponseWriter, req *http.Request) {
			settings, err := store.load()
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, errLoadSettings)
				return
			}
			statuses := computeVaultStatuses(req.Context(), settings.Vaults, vaultStatusClients)
			writeJSON(w, http.StatusOK, adminSettingsResponse{Settings: maskSecrets(settings), VaultStatuses: statuses})
		})

		r.Put("/api/admin/settings", adminSettingsPut(store, vaultStatusClients, refreshRegistry, auditor))

		r.Post("/api/admin/vault", func(w http.ResponseWriter, req *http.Request) {
			key, err := newVaultKey()
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "failed to allocate vault key")
				return
			}
			enabled := true
			vault := config.VaultInstance{
				ID:          "",
				Address:     "",
				Token:       "",
				PKIMount:    "pki",
				PKIMounts:   []string{"pki"},
				DisplayName: "",
				Enabled:     &enabled,
			}
			writeJSON(w, http.StatusOK, adminVaultAddedResponse{Key: key, Vault: vault})
		})

		r.Delete("/api/admin/vault/{id}", adminVaultDelete(store, refreshRegistry, auditor))
	})
}

// adminSettingsPut returns the handler for PUT /api/admin/settings: it merges
// the incoming settings over the stored ones (preserving secrets) and saves.
func adminSettingsPut(store *adminSettingsStore, vaultStatusClients map[string]vault.Client, refreshRegistry func(), auditor adminAuditor) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var incoming config.SettingsFile
		if err := json.NewDecoder(req.Body).Decode(&incoming); err != nil {
			auditor.log(req, "admin.settings_put", false, map[string]string{"reason": "invalid settings payload"})
			writeJSONError(w, http.StatusBadRequest, "invalid settings payload")
			return
		}
		current, err := store.load()
		if err != nil {
			auditor.log(req, "admin.settings_put", false, map[string]string{"reason": errLoadSettings})
			writeJSONError(w, http.StatusInternalServerError, errLoadSettings)
			return
		}
		merged := mergeAdminSettings(current, incoming)
		if saveErr := store.save(merged); saveErr != nil {
			auditor.log(req, "admin.settings_put", false, map[string]string{"reason": saveErr.Error()})
			writeJSONError(w, validationErrorStatus(saveErr), saveErr.Error())
			return
		}
		refreshRegistry()
		updated, err := store.load()
		if err != nil {
			auditor.log(req, "admin.settings_put", false, map[string]string{"reason": "failed to reload settings"})
			writeJSONError(w, http.StatusInternalServerError, "failed to reload settings")
			return
		}
		auditor.log(req, "admin.settings_put", true, nil)
		statuses := computeVaultStatuses(req.Context(), updated.Vaults, vaultStatusClients)
		writeJSON(w, http.StatusOK, adminSettingsResponse{Settings: maskSecrets(updated), VaultStatuses: statuses})
	}
}

// validationErrorStatus maps settings validation errors to 400 and everything
// else to 500.
func validationErrorStatus(saveErr error) int {
	if errors.Is(saveErr, vcverrors.ErrInvalidAddress) ||
		errors.Is(saveErr, vcverrors.ErrInvalidToken) ||
		errors.Is(saveErr, vcverrors.ErrInvalidThreshold) ||
		errors.Is(saveErr, vcverrors.ErrInvalidWebhookURL) ||
		errors.Is(saveErr, vcverrors.ErrInvalidWebhookTarget) ||
		errors.Is(saveErr, vcverrors.ErrVaultIDEmpty) ||
		errors.Is(saveErr, vcverrors.ErrDuplicateVaultID) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

// adminVaultDelete returns the handler for DELETE /api/admin/vault/{id}.
func adminVaultDelete(store *adminSettingsStore, refreshRegistry func(), auditor adminAuditor) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vaultID := strings.TrimSpace(chi.URLParam(req, "id"))
		if vaultID == "" {
			writeJSONError(w, http.StatusBadRequest, "vault id required")
			return
		}
		settings, err := store.load()
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, errLoadSettings)
			return
		}
		updatedVaults := make([]config.VaultInstance, 0, len(settings.Vaults))
		removed := false
		for _, vault := range settings.Vaults {
			if strings.TrimSpace(vault.ID) == vaultID {
				removed = true
				continue
			}
			updatedVaults = append(updatedVaults, vault)
		}
		if !removed {
			writeJSONError(w, http.StatusNotFound, "vault not found")
			return
		}
		settings.Vaults = updatedVaults
		if saveErr := store.save(settings); saveErr != nil {
			requestID := middleware.GetRequestID(req.Context())
			logger.HTTPError(req.Method, req.URL.Path, http.StatusInternalServerError, saveErr).
				Str("request_id", requestID).
				Msg("failed to save settings after vault removal")
			auditor.log(req, "admin.vault_delete", false, map[string]string{"vault_id": vaultID, "reason": "failed to save settings"})
			writeJSONError(w, http.StatusInternalServerError, "failed to save settings")
			return
		}
		refreshRegistry()
		auditor.log(req, "admin.vault_delete", true, map[string]string{"vault_id": vaultID})
		w.WriteHeader(http.StatusNoContent)
	}
}

func mergeAdminSettings(current, incoming config.SettingsFile) config.SettingsFile {
	merged := current
	merged.Certificates.ExpirationThresholds = incoming.Certificates.ExpirationThresholds
	merged.Metrics.PerCertificate = incoming.Metrics.PerCertificate
	merged.Metrics.EnhancedMetrics = incoming.Metrics.EnhancedMetrics
	merged.Metrics.PinnedCertificates = incoming.Metrics.PinnedCertificates
	merged.CORS.AllowedOrigins = incoming.CORS.AllowedOrigins
	merged.Notifications.WebhookURL = mergeSecret(incoming.Notifications.WebhookURL, current.Notifications.WebhookURL)
	merged.Notifications.Webhooks = mergeWebhookTargets(current.Notifications.Webhooks, incoming.Notifications.Webhooks)
	merged.Vaults = mergeVaultTokens(incoming.Vaults, current.Vaults)
	return merged
}

// mergeSecret returns incoming unless it's blank or a UI mask sentinel, in
// which case the previously stored value is preserved. Used for any
// settings field whose GET response must never round-trip a masked value
// back into storage (webhook URL may carry an auth token in its path, same
// concern as vault tokens).
func mergeSecret(incoming, existing string) string {
	if isBlankOrMaskedSecret(incoming) {
		return existing
	}
	return incoming
}

// mergeWebhookTargets merges the routed webhook list positionally: URLs are
// secret-like (they may embed auth tokens), so a blank or masked incoming
// URL preserves the stored one at the same position while levels are taken
// as-is. A nil incoming list means the field was absent and preserves the
// stored list; an explicitly empty list clears it.
func mergeWebhookTargets(current, incoming []config.WebhookTarget) []config.WebhookTarget {
	if incoming == nil {
		return current
	}
	merged := make([]config.WebhookTarget, 0, len(incoming))
	for i, target := range incoming {
		url := strings.TrimSpace(target.URL)
		if isBlankOrMaskedSecret(url) && i < len(current) {
			url = current[i].URL
		}
		merged = append(merged, config.WebhookTarget{URL: url, Levels: target.Levels})
	}
	return merged
}

// maskSecrets returns a copy of settings with every vault's Token and the
// webhook URL blanked, so cleartext secrets never reach the browser. Stored
// values are preserved on save by mergeVaultTokens/mergeSecret when the
// incoming field is empty, so the round-trip still works with masked
// responses.
func maskSecrets(s config.SettingsFile) config.SettingsFile {
	out := s
	out.Vaults = make([]config.VaultInstance, len(s.Vaults))
	for i, v := range s.Vaults {
		v.Token = ""
		out.Vaults[i] = v
	}
	out.Notifications.WebhookURL = ""
	masked := make([]config.WebhookTarget, len(s.Notifications.Webhooks))
	for i, target := range s.Notifications.Webhooks {
		target.URL = ""
		masked[i] = target
	}
	out.Notifications.Webhooks = masked
	return out
}

// isBlankOrMaskedSecret reports whether value is empty or a common UI mask
// sentinel that must not overwrite a stored secret (vault token, webhook
// URL) on admin PUT.
func isBlankOrMaskedSecret(token string) bool {
	t := strings.TrimSpace(token)
	if t == "" {
		return true
	}
	switch t {
	case "***", "********", "••••••••", "__MASKED__", "[masked]":
		return true
	}
	if strings.Trim(t, "*") == "" && len(t) >= 3 {
		return true
	}
	return false
}

// restoreVaultToken preserves the stored token when the incoming one is
// blank or masked (vaults are keyed by OriginalID, falling back to ID).
func restoreVaultToken(v *config.VaultInstance, tokens map[string]string) {
	if !isBlankOrMaskedSecret(v.Token) {
		return
	}
	lookupKey := v.OriginalID
	if lookupKey == "" {
		lookupKey = v.ID
	}
	if prior, ok := tokens[lookupKey]; ok {
		v.Token = prior
	}
}

// normalizePKIMounts keeps the PKIMount and PKIMounts fields consistent.
func normalizePKIMounts(v *config.VaultInstance) {
	if len(v.PKIMounts) == 0 && strings.TrimSpace(v.PKIMount) != "" {
		v.PKIMounts = []string{strings.TrimSpace(v.PKIMount)}
	}
	if strings.TrimSpace(v.PKIMount) == "" && len(v.PKIMounts) > 0 {
		v.PKIMount = v.PKIMounts[0]
	}
}

func mergeVaultTokens(incoming, existing []config.VaultInstance) []config.VaultInstance {
	tokens := make(map[string]string, len(existing))
	for _, v := range existing {
		tokens[v.ID] = v.Token
	}
	merged := make([]config.VaultInstance, 0, len(incoming))
	for _, v := range incoming {
		restoreVaultToken(&v, tokens)
		v.OriginalID = ""
		normalizePKIMounts(&v)
		merged = append(merged, v)
	}
	return merged
}

func computeVaultStatuses(ctx context.Context, vaults []config.VaultInstance, clients map[string]vault.Client) []adminVaultStatus {
	ordered := make([]string, 0, len(vaults))
	enabledByID := make(map[string]bool, len(vaults))
	for _, v := range vaults {
		ordered = append(ordered, v.ID)
		enabledByID[v.ID] = config.IsVaultEnabled(v)
	}
	checkClients := clients
	if checkClients == nil {
		checkClients = map[string]vault.Client{}
	}
	// Only check enabled vaults; pass nil for disabled so CheckInstances skips work.
	filtered := make(map[string]vault.Client, len(checkClients))
	for id, client := range checkClients {
		if enabledByID[id] {
			filtered[id] = client
		}
	}
	checked := vault.CheckInstances(ctx, ordered, filtered, 5*time.Second)
	statuses := make([]adminVaultStatus, len(vaults))
	for i, v := range vaults {
		connected := false
		if i < len(checked) {
			connected = checked[i].Connected
		}
		statuses[i] = adminVaultStatus{ID: v.ID, Enabled: enabledByID[v.ID], Connected: connected}
	}
	return statuses
}
