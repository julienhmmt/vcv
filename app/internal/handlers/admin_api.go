package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
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

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
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
		SameSite: http.SameSiteLaxMode,
		Secure:   s.secureCookies,
		Expires:  expiresAt,
	})
	return true, ""
}

// registerAdminAPIRoutes mounts JSON admin endpoints alongside the existing
// HTMX form routes. Shares the session and settings stores so both UIs see
// the same state.
func registerAdminAPIRoutes(
	router chi.Router,
	sessions *adminSessionStore,
	store *adminSettingsStore,
	vaultStatusClients map[string]vault.Client,
	refreshRegistry func(),
) {
	router.Get("/api/admin/session", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, adminSessionResponse{Authenticated: sessions.isAuthed(r)})
	})

	router.Post("/api/admin/login", func(w http.ResponseWriter, r *http.Request) {
		var body adminLoginRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		ok, message := sessions.loginFromJSON(w, r, body)
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, message)
			return
		}
		writeJSON(w, http.StatusOK, adminSessionResponse{Authenticated: true})
	})

	router.Post("/api/admin/logout", func(w http.ResponseWriter, r *http.Request) {
		sessions.clearCookie(w)
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
				writeJSONError(w, http.StatusInternalServerError, "failed to load settings")
				return
			}
			statuses := computeVaultStatuses(req.Context(), settings.Vaults, vaultStatusClients)
			writeJSON(w, http.StatusOK, adminSettingsResponse{Settings: maskVaultTokens(settings), VaultStatuses: statuses})
		})

		r.Put("/api/admin/settings", func(w http.ResponseWriter, req *http.Request) {
			var incoming config.SettingsFile
			if err := json.NewDecoder(req.Body).Decode(&incoming); err != nil {
				writeJSONError(w, http.StatusBadRequest, "invalid settings payload")
				return
			}
			current, err := store.load()
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "failed to load settings")
				return
			}
			merged := mergeAdminSettings(current, incoming)
			if saveErr := store.save(merged); saveErr != nil {
				status := http.StatusBadRequest
				if !errors.Is(saveErr, vcverrors.ErrInvalidAddress) &&
					!errors.Is(saveErr, vcverrors.ErrInvalidToken) &&
					!errors.Is(saveErr, vcverrors.ErrInvalidThreshold) &&
					!errors.Is(saveErr, vcverrors.ErrVaultIDEmpty) &&
					!errors.Is(saveErr, vcverrors.ErrDuplicateVaultID) {
					status = http.StatusInternalServerError
				}
				writeJSONError(w, status, saveErr.Error())
				return
			}
			refreshRegistry()
			updated, err := store.load()
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "failed to reload settings")
				return
			}
			statuses := computeVaultStatuses(req.Context(), updated.Vaults, vaultStatusClients)
			writeJSON(w, http.StatusOK, adminSettingsResponse{Settings: maskVaultTokens(updated), VaultStatuses: statuses})
		})

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

		r.Delete("/api/admin/vault/{id}", func(w http.ResponseWriter, req *http.Request) {
			vaultID := strings.TrimSpace(chi.URLParam(req, "id"))
			if vaultID == "" {
				writeJSONError(w, http.StatusBadRequest, "vault id required")
				return
			}
			settings, err := store.load()
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "failed to load settings")
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
				writeJSONError(w, http.StatusInternalServerError, "failed to save settings")
				return
			}
			refreshRegistry()
			w.WriteHeader(http.StatusNoContent)
		})
	})
}

func mergeAdminSettings(current, incoming config.SettingsFile) config.SettingsFile {
	merged := current
	merged.Certificates.ExpirationThresholds = incoming.Certificates.ExpirationThresholds
	merged.Metrics.PerCertificate = incoming.Metrics.PerCertificate
	merged.Metrics.EnhancedMetrics = incoming.Metrics.EnhancedMetrics
	merged.Metrics.PinnedCertificates = incoming.Metrics.PinnedCertificates
	merged.CORS.AllowedOrigins = incoming.CORS.AllowedOrigins
	merged.Vaults = mergeVaultTokens(incoming.Vaults, current.Vaults)
	return merged
}

// maskVaultTokens returns a copy of settings with every vault's Token blanked
// so cleartext tokens never reach the browser. Stored tokens are preserved on
// save by mergeVaultTokens when the incoming Token is empty, so the round-trip
// still works with masked responses.
func maskVaultTokens(s config.SettingsFile) config.SettingsFile {
	out := s
	out.Vaults = make([]config.VaultInstance, len(s.Vaults))
	for i, v := range s.Vaults {
		v.Token = ""
		out.Vaults[i] = v
	}
	return out
}

func mergeVaultTokens(incoming, existing []config.VaultInstance) []config.VaultInstance {
	tokens := make(map[string]string, len(existing))
	for _, v := range existing {
		tokens[v.ID] = v.Token
	}
	merged := make([]config.VaultInstance, 0, len(incoming))
	for _, v := range incoming {
		if strings.TrimSpace(v.Token) == "" {
			lookupKey := v.OriginalID
			if lookupKey == "" {
				lookupKey = v.ID
			}
			if prior, ok := tokens[lookupKey]; ok {
				v.Token = prior
			}
		}
		v.OriginalID = ""
		if len(v.PKIMounts) == 0 && strings.TrimSpace(v.PKIMount) != "" {
			v.PKIMounts = []string{strings.TrimSpace(v.PKIMount)}
		}
		if strings.TrimSpace(v.PKIMount) == "" && len(v.PKIMounts) > 0 {
			v.PKIMount = v.PKIMounts[0]
		}
		merged = append(merged, v)
	}
	return merged
}

func computeVaultStatuses(ctx context.Context, vaults []config.VaultInstance, clients map[string]vault.Client) []adminVaultStatus {
	statuses := make([]adminVaultStatus, len(vaults))
	var wg sync.WaitGroup
	for i, v := range vaults {
		enabled := config.IsVaultEnabled(v)
		statuses[i] = adminVaultStatus{ID: v.ID, Enabled: enabled, Connected: false}
		if !enabled || clients == nil {
			continue
		}
		client, ok := clients[v.ID]
		if !ok || client == nil {
			continue
		}
		wg.Add(1)
		go func(idx int, client vault.Client) {
			defer wg.Done()
			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := client.CheckConnection(checkCtx); err == nil {
				statuses[idx].Connected = true
			}
		}(i, client)
	}
	wg.Wait()
	return statuses
}
