package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"vcv/internal/config"
	vcverrors "vcv/internal/errors"
	"vcv/internal/httputil"
	"vcv/internal/logger"
	"vcv/internal/middleware"
	"vcv/internal/vault"
)

const adminCookieName string = "vcv_admin_session"
const adminUsername string = "admin"
const adminMaxSessions int = 1024

type adminSessionStore struct {
	mu            sync.Mutex
	password      string
	sessions      map[string]time.Time
	sessionTTL    time.Duration
	secureCookies bool
	limiter       *adminLoginLimiter
}

func newAdminSessionStore(password string, secureCookies bool) *adminSessionStore {
	ttl := 12 * time.Hour
	if secureCookies {
		ttl = 4 * time.Hour
	}
	return &adminSessionStore{
		password:      password,
		sessions:      make(map[string]time.Time),
		sessionTTL:    ttl,
		secureCookies: secureCookies,
		limiter:       newAdminLoginLimiter(5, 3*time.Minute),
	}
}

type adminLoginLimiter struct {
	mu          sync.Mutex
	maxAttempts int
	window      time.Duration
	entries     map[string]adminLoginLimiterEntry
}

type adminLoginLimiterEntry struct {
	count   int
	resetAt time.Time
}

func newAdminLoginLimiter(maxAttempts int, window time.Duration) *adminLoginLimiter {
	return &adminLoginLimiter{maxAttempts: maxAttempts, window: window, entries: make(map[string]adminLoginLimiterEntry)}
}

func (l *adminLoginLimiter) allow(now time.Time, key string) bool {
	if key == "" {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	entry := l.entries[key]
	if entry.resetAt.IsZero() || now.After(entry.resetAt) {
		entry = adminLoginLimiterEntry{count: 0, resetAt: now.Add(l.window)}
	}
	entry.count++
	l.entries[key] = entry
	return entry.count <= l.maxAttempts
}

func (s *adminSessionStore) allowLoginAttempt(r *http.Request) bool {
	if s.limiter == nil {
		return true
	}
	return s.limiter.allow(time.Now(), httputil.ClientIP(r, s.secureCookies))
}

func (s *adminSessionStore) pruneSessions(now time.Time) {
	for token, expiresAt := range s.sessions {
		if now.After(expiresAt) {
			delete(s.sessions, token)
		}
	}
	if len(s.sessions) <= adminMaxSessions {
		return
	}
	for len(s.sessions) > adminMaxSessions {
		var oldestToken string
		oldestExpiry := now.Add(365 * 24 * time.Hour)
		for token, expiresAt := range s.sessions {
			if expiresAt.Before(oldestExpiry) {
				oldestToken = token
				oldestExpiry = expiresAt
			}
		}
		if oldestToken == "" {
			break
		}
		delete(s.sessions, oldestToken)
	}
}

func (s *adminSessionStore) createToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (s *adminSessionStore) verify(username string, password string) bool {
	if username != adminUsername {
		return false
	}
	stored := strings.TrimSpace(s.password)
	if stored == "" {
		return false
	}
	if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) == nil
	}
	return false
}

func (s *adminSessionStore) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     adminCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.secureCookies,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func (s *adminSessionStore) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(adminCookieName)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		token := strings.TrimSpace(cookie.Value)
		if token == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		s.mu.Lock()
		s.pruneSessions(time.Now())
		expiresAt, ok := s.sessions[token]
		if !ok || time.Now().After(expiresAt) {
			delete(s.sessions, token)
			s.mu.Unlock()
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		s.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func (s *adminSessionStore) isAuthed(r *http.Request) bool {
	cookie, err := r.Cookie(adminCookieName)
	if err != nil {
		return false
	}
	token := strings.TrimSpace(cookie.Value)
	if token == "" {
		return false
	}
	s.mu.Lock()
	expiresAt, ok := s.sessions[token]
	if !ok || time.Now().After(expiresAt) {
		delete(s.sessions, token)
		s.mu.Unlock()
		return false
	}
	s.mu.Unlock()
	return true
}

type adminSettingsStore struct {
	mu          sync.Mutex
	path        string
	defaultEnv  config.Environment
	defaultPort int
}

func newAdminSettingsStore(path string, env config.Environment) *adminSettingsStore {
	return &adminSettingsStore{path: path, defaultEnv: env, defaultPort: 52000}
}

func (s *adminSettingsStore) load() (config.SettingsFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return config.SettingsFile{App: config.AppSettings{Env: string(s.defaultEnv), Port: s.defaultPort}}, nil
		}
		return config.SettingsFile{}, err
	}
	var settings config.SettingsFile
	if jsonErr := json.Unmarshal(data, &settings); jsonErr != nil {
		return config.SettingsFile{}, jsonErr
	}
	if strings.TrimSpace(settings.App.Env) == "" {
		settings.App.Env = string(s.defaultEnv)
	}
	if settings.App.Port == 0 {
		settings.App.Port = s.defaultPort
	}
	return settings, nil
}

func (s *adminSettingsStore) save(settings config.SettingsFile) error {
	if err := validateSettings(settings); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	payload, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
		return mkErr
	}
	tmp, err := os.CreateTemp(dir, "settings-*.json")
	if err != nil {
		return fallbackWriteSettings(s.path, payload, err)
	}
	tmpPath := tmp.Name()
	closeErr := tmp.Close()
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return fallbackWriteSettings(s.path, payload, closeErr)
	}
	if writeErr := os.WriteFile(tmpPath, payload, 0o600); writeErr != nil {
		_ = os.Remove(tmpPath)
		return fallbackWriteSettings(s.path, payload, writeErr)
	}
	if renameErr := os.Rename(tmpPath, s.path); renameErr != nil {
		_ = os.Remove(tmpPath)
		return fallbackWriteSettings(s.path, payload, renameErr)
	}
	return nil
}

func fallbackWriteSettings(path string, payload []byte, originalErr error) error {
	if !shouldFallbackToDirectWrite(originalErr) {
		return originalErr
	}
	if writeErr := os.WriteFile(path, payload, 0o600); writeErr != nil {
		return originalErr
	}
	return nil
}

func shouldFallbackToDirectWrite(err error) bool {
	if err == nil {
		return false
	}
	if os.IsPermission(err) {
		return true
	}
	if errors.Is(err, syscall.EROFS) || errors.Is(err, syscall.EPERM) || errors.Is(err, syscall.EACCES) {
		return true
	}
	return false
}

func validateSettings(settings config.SettingsFile) error {
	normalizedVaults := make([]config.VaultInstance, len(settings.Vaults))
	copy(normalizedVaults, settings.Vaults)
	seen := make(map[string]struct{})
	for i, vault := range normalizedVaults {
		id := strings.TrimSpace(vault.ID)
		address := strings.TrimSpace(vault.Address)
		token := strings.TrimSpace(vault.Token)
		if id == "" {
			return vcverrors.ErrVaultIDEmpty
		}
		if _, ok := seen[id]; ok {
			return vcverrors.ErrDuplicateVaultID
		}
		seen[id] = struct{}{}
		if address == "" {
			return errors.New("vault address is empty")
		}
		if _, err := url.ParseRequestURI(address); err != nil {
			return vcverrors.ErrInvalidAddress
		}
		if token == "" {
			return vcverrors.ErrInvalidToken
		}
		if len(vault.PKIMounts) == 0 {
			if strings.TrimSpace(vault.PKIMount) == "" {
				normalizedVaults[i].PKIMount = "pki"
			}
			normalizedVaults[i].PKIMounts = []string{normalizedVaults[i].PKIMount}
		}
	}
	return nil
}

func newVaultKey() (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// RegisterAdminRoutes wires the JSON admin API. HTMX routes have been removed
// in favor of the Svelte admin panel that talks to /api/admin/*.
func RegisterAdminRoutes(router chi.Router, settingsPath string, env config.Environment, vaultRegistry *vault.Registry, vaultStatusClients map[string]vault.Client, cacheClient vault.Client) {
	settingsStore := newAdminSettingsStore(settingsPath, env)
	settings, err := settingsStore.load()
	if err != nil {
		return
	}

	password := strings.TrimSpace(settings.Admin.Password)
	if password == "" {
		return
	}
	if !strings.HasPrefix(password, "$2a$") && !strings.HasPrefix(password, "$2b$") && !strings.HasPrefix(password, "$2y$") {
		return
	}

	secureCookies := env == config.EnvProd
	sessions := newAdminSessionStore(password, secureCookies)
	store := settingsStore
	refreshRegistry := func() {
		if vaultRegistry == nil {
			return
		}
		if s, loadErr := store.load(); loadErr == nil {
			vaultRegistry.Update(s.Vaults)
		}
	}

	registerAdminAPIRoutes(router, sessions, store, vaultStatusClients, refreshRegistry)

	router.Group(func(r chi.Router) {
		r.Use(sessions.requireAuth)
		r.Post("/api/cache/invalidate", func(w http.ResponseWriter, r *http.Request) {
			if cacheClient == nil {
				http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
				return
			}
			cacheClient.InvalidateCache()
			w.WriteHeader(http.StatusNoContent)
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPEvent(r.Method, r.URL.Path, http.StatusNoContent, 0).
				Str("request_id", requestID).
				Msg("invalidated cache")
		})
	})
}
