package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"vcv/config"
	"vcv/internal/logger"
	"vcv/middleware"
)

const adminCookieName string = "vcv_admin_session"
const adminUsername string = "admin"

type adminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type adminSessionStore struct {
	mu            sync.Mutex
	password      string
	sessions      map[string]time.Time
	sessionTTL    time.Duration
	secureCookies bool
}

func newAdminSessionStore(password string, secureCookies bool) *adminSessionStore {
	return &adminSessionStore{password: password, sessions: make(map[string]time.Time), sessionTTL: 12 * time.Hour, secureCookies: secureCookies}
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
	return subtleConstantTimeEquals(stored, password)
}

func subtleConstantTimeEquals(left string, right string) bool {
	leftBytes := []byte(left)
	rightBytes := []byte(right)
	if len(leftBytes) != len(rightBytes) {
		return false
	}
	result := byte(0)
	for i := 0; i < len(leftBytes); i++ {
		result |= leftBytes[i] ^ rightBytes[i]
	}
	return result == 0
}

func (s *adminSessionStore) login(w http.ResponseWriter, r *http.Request) {
	var payload adminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if !s.verify(strings.TrimSpace(payload.Username), payload.Password) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	token, err := s.createToken()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	expiresAt := time.Now().Add(s.sessionTTL)
	s.mu.Lock()
	s.sessions[token] = expiresAt
	s.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: adminCookieName, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: s.secureCookies, Expires: expiresAt})
	w.WriteHeader(http.StatusNoContent)
}

func (s *adminSessionStore) logout(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: adminCookieName, Value: "", Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: s.secureCookies, Expires: time.Unix(0, 0), MaxAge: -1})
	w.WriteHeader(http.StatusNoContent)
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
		return err
	}
	tmpPath := tmp.Name()
	closeErr := tmp.Close()
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return closeErr
	}
	if writeErr := os.WriteFile(tmpPath, payload, 0o600); writeErr != nil {
		_ = os.Remove(tmpPath)
		return writeErr
	}
	if renameErr := os.Rename(tmpPath, s.path); renameErr != nil {
		_ = os.Remove(tmpPath)
		return renameErr
	}
	return nil
}

func validateSettings(settings config.SettingsFile) error {
	seen := make(map[string]struct{})
	for _, vault := range settings.Vaults {
		id := strings.TrimSpace(vault.ID)
		address := strings.TrimSpace(vault.Address)
		token := strings.TrimSpace(vault.Token)
		if id == "" {
			return errors.New("vault id is empty")
		}
		if _, ok := seen[id]; ok {
			return errors.New("duplicate vault id")
		}
		seen[id] = struct{}{}
		if address == "" {
			return errors.New("vault address is empty")
		}
		if _, err := url.ParseRequestURI(address); err != nil {
			return errors.New("invalid vault address")
		}
		if token == "" {
			return errors.New("vault token is empty")
		}
		if len(vault.PKIMounts) == 0 {
			if strings.TrimSpace(vault.PKIMount) == "" {
				vault.PKIMount = "pki"
			}
			vault.PKIMounts = []string{vault.PKIMount}
		}
	}
	return nil
}

func (s *adminSettingsStore) getSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.load()
	if err != nil {
		requestID := middleware.GetRequestID(r.Context())
		logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
			Str("request_id", requestID).
			Msg("failed to load admin settings")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(settings)
}

func (s *adminSettingsStore) putSettings(w http.ResponseWriter, r *http.Request) {
	var settings config.SettingsFile
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err := s.save(settings); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func RegisterAdminRoutes(router chi.Router, webFS fs.FS, settingsPath string, env config.Environment) {
	password := strings.TrimSpace(os.Getenv("VCV_ADMIN_PASSWORD"))
	if password == "" {
		return
	}
	secureCookies := env == config.EnvProd
	sessions := newAdminSessionStore(password, secureCookies)
	store := newAdminSettingsStore(settingsPath, env)
	router.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(webFS, "admin.html")
		if err != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to read admin html")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})
	router.Post("/api/admin/login", sessions.login)
	router.Post("/api/admin/logout", sessions.logout)
	router.Group(func(r chi.Router) {
		r.Use(sessions.requireAuth)
		r.Get("/api/admin/settings", store.getSettings)
		r.Put("/api/admin/settings", store.putSettings)
	})
}
