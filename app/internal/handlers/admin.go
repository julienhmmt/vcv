package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
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

type adminLoginTemplateData struct {
	ErrorText string
}

type adminVaultViewData struct {
	Enabled     bool
	Key         string
	MountsText  string
	Open        bool
	TLSInsecure bool
	Vault       config.VaultInstance
}

type adminPanelTemplateData struct {
	CorsOriginsText string
	ErrorText       string
	Settings        config.SettingsFile
	SuccessText     string
	VaultViews      []adminVaultViewData
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

func (s *adminSessionStore) loginFromForm(w http.ResponseWriter, r *http.Request) bool {
	if err := r.ParseForm(); err != nil {
		return false
	}
	username := strings.TrimSpace(r.PostForm.Get("username"))
	password := r.PostForm.Get("password")
	if !s.verify(username, password) {
		return false
	}
	token, err := s.createToken()
	if err != nil {
		return false
	}
	expiresAt := time.Now().Add(s.sessionTTL)
	s.mu.Lock()
	s.sessions[token] = expiresAt
	s.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: adminCookieName, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: s.secureCookies, Expires: expiresAt})
	return true
}

func (s *adminSessionStore) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: adminCookieName, Value: "", Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: s.secureCookies, Expires: time.Unix(0, 0), MaxAge: -1})
}

func (s *adminSessionStore) logoutJSON(w http.ResponseWriter, _ *http.Request) {
	s.clearCookie(w)
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
	writeErr := os.WriteFile(path, payload, 0o600)
	if writeErr != nil {
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
	if errors.Is(err, syscall.EROFS) {
		return true
	}
	if errors.Is(err, syscall.EPERM) {
		return true
	}
	if errors.Is(err, syscall.EACCES) {
		return true
	}
	return false
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

func parseTemplates(webFS fs.FS) (*template.Template, error) {
	return template.ParseFS(webFS, "templates/*.html")
}

func renderAdminTemplate(w http.ResponseWriter, templates *template.Template, name string, data interface{}) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return templates.ExecuteTemplate(w, name, data)
}

func buildAdminPanelData(settings config.SettingsFile, successText string, errorText string) adminPanelTemplateData {
	corsOriginsText := strings.Join(settings.CORS.AllowedOrigins, ",")
	views := make([]adminVaultViewData, 0, len(settings.Vaults))
	for i, vault := range settings.Vaults {
		mounts := vault.PKIMounts
		if len(mounts) == 0 {
			mount := strings.TrimSpace(vault.PKIMount)
			if mount != "" {
				mounts = []string{mount}
			}
		}
		mountsText := strings.Join(mounts, ",")
		key := fmt.Sprintf("%d", i)
		views = append(views, adminVaultViewData{Enabled: config.IsVaultEnabled(vault), Key: key, MountsText: mountsText, Open: false, TLSInsecure: vault.TLSInsecure, Vault: vault})
	}
	return adminPanelTemplateData{CorsOriginsText: corsOriginsText, ErrorText: errorText, Settings: settings, SuccessText: successText, VaultViews: views}
}

func parseSettingsUpdateForm(r *http.Request, existing config.SettingsFile) (config.SettingsFile, error) {
	if err := r.ParseForm(); err != nil {
		return config.SettingsFile{}, err
	}
	updated := existing
	criticalText := strings.TrimSpace(r.PostForm.Get("expire_critical"))
	warningText := strings.TrimSpace(r.PostForm.Get("expire_warning"))
	critical, err := strconv.Atoi(defaultString(criticalText, "0"))
	if err != nil {
		return config.SettingsFile{}, errors.New("invalid critical threshold")
	}
	warning, err := strconv.Atoi(defaultString(warningText, "0"))
	if err != nil {
		return config.SettingsFile{}, errors.New("invalid warning threshold")
	}
	updated.Certificates.ExpirationThresholds.Critical = critical
	updated.Certificates.ExpirationThresholds.Warning = warning
	updated.CORS.AllowedOrigins = splitAndTrim(r.PostForm.Get("cors_origins"))
	updated.Vaults = parseVaultsFromForm(r.PostForm)
	return updated, nil
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	trimmed := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		trimmed = append(trimmed, item)
	}
	return trimmed
}

func parseVaultsFromForm(form url.Values) []config.VaultInstance {
	keys := extractVaultKeys(form)
	vaults := make([]config.VaultInstance, 0, len(keys))
	for _, key := range keys {
		id := strings.TrimSpace(form.Get("vault_id_" + key))
		displayName := strings.TrimSpace(form.Get("vault_display_" + key))
		address := strings.TrimSpace(form.Get("vault_address_" + key))
		token := form.Get("vault_token_" + key)
		mounts := splitAndTrim(form.Get("vault_mounts_" + key))
		pkiMount := "pki"
		if len(mounts) > 0 {
			pkiMount = mounts[0]
		}
		tlsInsecure := form.Get("vault_tls_"+key) != ""
		enabledValue := form.Get("vault_enabled_"+key) != ""
		enabled := enabledValue
		vault := config.VaultInstance{ID: id, Address: address, Token: token, PKIMount: pkiMount, PKIMounts: mounts, DisplayName: displayName, TLSInsecure: tlsInsecure, Enabled: &enabled}
		vaults = append(vaults, vault)
	}
	return vaults
}

func extractVaultKeys(form url.Values) []string {
	set := make(map[string]struct{})
	for name := range form {
		if !strings.HasPrefix(name, "vault_id_") {
			continue
		}
		suffix := strings.TrimPrefix(name, "vault_id_")
		if suffix == "" {
			continue
		}
		set[suffix] = struct{}{}
	}
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func newVaultKey() (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func RegisterAdminRoutes(router chi.Router, webFS fs.FS, settingsPath string, env config.Environment) {
	password := strings.TrimSpace(os.Getenv("VCV_ADMIN_PASSWORD"))
	if password == "" {
		return
	}
	secureCookies := env == config.EnvProd
	sessions := newAdminSessionStore(password, secureCookies)
	store := newAdminSettingsStore(settingsPath, env)
	templates, templatesErr := parseTemplates(webFS)
	if templatesErr != nil {
		panic(templatesErr)
	}
	router.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		if err := renderAdminTemplate(w, templates, "admin-page.html", nil); err != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to render admin page")
			return
		}
	})
	router.Get("/admin/panel", func(w http.ResponseWriter, r *http.Request) {
		if !sessions.isAuthed(r) {
			if err := renderAdminTemplate(w, templates, "admin-login-fragment.html", adminLoginTemplateData{}); err != nil {
				requestID := middleware.GetRequestID(r.Context())
				logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
					Str("request_id", requestID).
					Msg("failed to render admin login")
				return
			}
			return
		}
		settings, err := store.load()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		data := buildAdminPanelData(settings, "", "")
		if err := renderAdminTemplate(w, templates, "admin-panel-fragment.html", data); err != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to render admin panel")
			return
		}
	})
	router.Post("/admin/login", func(w http.ResponseWriter, r *http.Request) {
		if sessions.loginFromForm(w, r) {
			settings, err := store.load()
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			data := buildAdminPanelData(settings, "", "")
			_ = renderAdminTemplate(w, templates, "admin-panel-fragment.html", data)
			return
		}
		_ = renderAdminTemplate(w, templates, "admin-login-fragment.html", adminLoginTemplateData{ErrorText: "Invalid credentials"})
	})
	router.Post("/admin/logout", func(w http.ResponseWriter, r *http.Request) {
		sessions.clearCookie(w)
		_ = renderAdminTemplate(w, templates, "admin-login-fragment.html", adminLoginTemplateData{})
	})
	router.Group(func(r chi.Router) {
		r.Use(sessions.requireAuth)
		r.Post("/admin/settings", func(w http.ResponseWriter, r *http.Request) {
			current, err := store.load()
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			updated, err := parseSettingsUpdateForm(r, current)
			if err != nil {
				data := buildAdminPanelData(current, "", err.Error())
				_ = renderAdminTemplate(w, templates, "admin-panel-fragment.html", data)
				return
			}
			if err := store.save(updated); err != nil {
				data := buildAdminPanelData(updated, "", err.Error())
				_ = renderAdminTemplate(w, templates, "admin-panel-fragment.html", data)
				return
			}
			settings, err := store.load()
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			data := buildAdminPanelData(settings, "Settings saved", "")
			_ = renderAdminTemplate(w, templates, "admin-panel-fragment.html", data)
		})
		r.Post("/admin/vault/add", func(w http.ResponseWriter, r *http.Request) {
			key, err := newVaultKey()
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			vault := config.VaultInstance{ID: "", Address: "", Token: "", PKIMount: "pki", PKIMounts: []string{"pki"}, DisplayName: "", TLSInsecure: false}
			data := adminVaultViewData{Enabled: true, Key: key, MountsText: "pki", Open: true, TLSInsecure: false, Vault: vault}
			w.Header().Set("HX-Trigger-After-Swap", fmt.Sprintf(`{"adminVaultAdded":{"key":"%s"}}`, key))
			_ = renderAdminTemplate(w, templates, "admin-vault-item.html", data)
		})
		r.Post("/admin/vault/remove", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			vaultID := strings.TrimSpace(r.PostForm.Get("vaultId"))
			if vaultID == "" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				return
			}
			settings, err := store.load()
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			updatedVaults := make([]config.VaultInstance, 0, len(settings.Vaults))
			for _, vault := range settings.Vaults {
				if strings.TrimSpace(vault.ID) == vaultID {
					continue
				}
				updatedVaults = append(updatedVaults, vault)
			}
			settings.Vaults = updatedVaults
			if err := store.save(settings); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		})
	})
	router.Post("/api/admin/login", sessions.login)
	router.Post("/api/admin/logout", sessions.logoutJSON)
	router.Group(func(r chi.Router) {
		r.Use(sessions.requireAuth)
		r.Get("/api/admin/settings", store.getSettings)
		r.Put("/api/admin/settings", store.putSettings)
	})
}
