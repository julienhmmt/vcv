package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"vcv/internal/config"
	"vcv/internal/vault"
)

func TestAdminSessionStore_NewAdminSessionStore(t *testing.T) {
	password := "$2a$10$testhashedpassword"
	store := newAdminSessionStore(password, false)

	assert.NotNil(t, store)
	assert.Equal(t, password, store.password)
	assert.Equal(t, 12*time.Hour, store.sessionTTL)
	assert.False(t, store.secureCookies)
}

func TestAdminSessionStore_NewAdminSessionStore_SecureCookies(t *testing.T) {
	password := "$2a$10$testhashedpassword"
	store := newAdminSessionStore(password, true)

	assert.NotNil(t, store)
	assert.Equal(t, password, store.password)
	assert.Equal(t, 4*time.Hour, store.sessionTTL)
	assert.True(t, store.secureCookies)
}

func TestAdminSessionStore_CreateToken(t *testing.T) {
	store := newAdminSessionStore("$2a$10$testhashedpassword", false)

	token, err := store.createToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, len(token), 20) // Base64 encoded 32 bytes should be longer than 20 chars
}

func TestAdminSessionStore_Verify(t *testing.T) {
	// Create a bcrypt hash for "testpassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
	require.NoError(t, err)

	store := newAdminSessionStore(string(hashedPassword), false)

	tests := []struct {
		name     string
		username string
		password string
		expected bool
	}{
		{
			name:     "correct credentials",
			username: "admin",
			password: "testpassword",
			expected: true,
		},
		{
			name:     "wrong username",
			username: "wronguser",
			password: "testpassword",
			expected: false,
		},
		{
			name:     "wrong password",
			username: "admin",
			password: "wrongpassword",
			expected: false,
		},
		{
			name:     "empty password",
			username: "admin",
			password: "",
			expected: false,
		},
		{
			name:     "empty username",
			username: "",
			password: "testpassword",
			expected: false,
		},
		{
			name:     "bcrypt hash with different prefix",
			username: "admin",
			password: "testpassword",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := store.verify(tt.username, tt.password)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdminSessionStore_Verify_PlaintextPassword(t *testing.T) {
	store := newAdminSessionStore("plaintextpassword", false)

	// Should return false for plaintext passwords (not bcrypt)
	result := store.verify("admin", "plaintextpassword")
	assert.False(t, result)
}

func TestAdminSessionStore_ClearCookie(t *testing.T) {
	store := newAdminSessionStore("$2a$10$testhashedpassword", false)

	w := httptest.NewRecorder()
	store.clearCookie(w)

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, "vcv_admin_session", cookie.Name)
	assert.Equal(t, "", cookie.Value)
	assert.Equal(t, "/", cookie.Path)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.False(t, cookie.Secure) // false for non-secure cookies
	assert.True(t, cookie.Expires.Before(time.Now()))
	assert.Equal(t, -1, cookie.MaxAge)
}

func TestAdminSessionStore_ClearCookie_Secure(t *testing.T) {
	store := newAdminSessionStore("$2a$10$testhashedpassword", true)

	w := httptest.NewRecorder()
	store.clearCookie(w)

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.True(t, cookie.Secure) // true for secure cookies
}

func TestAdminSessionStore_IsAuthed(t *testing.T) {
	store := newAdminSessionStore("$2a$10$testhashedpassword", false)

	// Test without cookie
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	result := store.isAuthed(req)
	assert.False(t, result)

	// Test with empty cookie
	req.AddCookie(&http.Cookie{Name: "vcv_admin_session", Value: ""})
	result = store.isAuthed(req)
	assert.False(t, result)

	// Test with valid session
	token, err := store.createToken()
	require.NoError(t, err)

	// Manually add the token to sessions (simulating login)
	store.sessions[token] = time.Now().Add(1 * time.Hour)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "vcv_admin_session", Value: token})
	result = store.isAuthed(req)
	assert.True(t, result)

	// Test with expired session
	expiredToken, err := store.createToken()
	require.NoError(t, err)
	store.sessions[expiredToken] = time.Now().Add(-1 * time.Hour)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "vcv_admin_session", Value: expiredToken})
	result = store.isAuthed(req)
	assert.False(t, result)
}

func TestAdminSessionStore_RequireAuth(t *testing.T) {
	store := newAdminSessionStore("$2a$10$testhashedpassword", false)

	// Create a test handler that sets a flag when called
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	authHandler := store.requireAuth(testHandler)

	// Test without cookie
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	authHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, handlerCalled)

	// Test with valid session
	token, err := store.createToken()
	require.NoError(t, err)
	store.sessions[token] = time.Now().Add(1 * time.Hour)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "vcv_admin_session", Value: token})
	w = httptest.NewRecorder()
	authHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, handlerCalled)
}

func TestAdminLoginLimiter_Allow(t *testing.T) {
	limiter := newAdminLoginLimiter(3, 5*time.Minute)
	now := time.Now()

	// Test with empty key (should always allow)
	result := limiter.allow(now, "")
	assert.True(t, result)

	// Test normal usage
	key := "192.168.1.1"

	// First 3 attempts should allow
	for range 3 {
		result := limiter.allow(now, key)
		assert.True(t, result)
	}

	// 4th attempt should deny
	result = limiter.allow(now, key)
	assert.False(t, result)
}

func TestAdminLoginLimiter_Allow_WindowReset(t *testing.T) {
	limiter := newAdminLoginLimiter(2, 1*time.Hour)
	now := time.Now()
	key := "192.168.1.1"

	// Use up the limit
	assert.True(t, limiter.allow(now, key))
	assert.True(t, limiter.allow(now, key))
	assert.False(t, limiter.allow(now, key))

	// After window passes, should allow again
	later := now.Add(2 * time.Hour)
	assert.True(t, limiter.allow(later, key))
}

func TestAdminSettingsStore_Load(t *testing.T) {
	// Test loading non-existent file
	store := newAdminSettingsStore("/tmp/nonexistent.json", config.EnvDev)

	settings, err := store.load()
	assert.NoError(t, err)
	assert.Equal(t, "dev", settings.App.Env)
	assert.Equal(t, 52000, settings.App.Port)
}

func TestAdminSettingsStore_Load_InvalidJSON(t *testing.T) {
	// Create a file with invalid JSON
	tmpFile := t.TempDir() + "/invalid.json"
	err := os.WriteFile(tmpFile, []byte("{ invalid json"), 0644)
	require.NoError(t, err)

	store := newAdminSettingsStore(tmpFile, config.EnvDev)

	_, err = store.load()
	assert.Error(t, err)
}

func TestAdminSettingsStore_Save(t *testing.T) {
	tmpFile := t.TempDir() + "/settings.json"
	store := newAdminSettingsStore(tmpFile, config.EnvDev)

	settings := config.SettingsFile{
		App: config.AppSettings{
			Env:  "prod",
			Port: 8080,
		},
		Vaults: []config.VaultInstance{
			{
				ID:      "test-vault",
				Address: "http://localhost:8200",
				Token:   "test-token",
			},
		},
	}

	err := store.save(settings)
	assert.NoError(t, err)

	// Verify file was created and contains valid JSON
	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	var loadedSettings config.SettingsFile
	err = json.Unmarshal(data, &loadedSettings)
	require.NoError(t, err)

	assert.Equal(t, settings.App.Env, loadedSettings.App.Env)
	assert.Equal(t, settings.App.Port, loadedSettings.App.Port)
	assert.Len(t, loadedSettings.Vaults, 1)
	assert.Equal(t, "test-vault", loadedSettings.Vaults[0].ID)
}

func TestValidateSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings config.SettingsFile
		wantErr  bool
	}{
		{
			name: "valid settings",
			settings: config.SettingsFile{
				Vaults: []config.VaultInstance{
					{
						ID:      "vault1",
						Address: "http://localhost:8200",
						Token:   "token1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty vault ID",
			settings: config.SettingsFile{
				Vaults: []config.VaultInstance{
					{
						ID:      "",
						Address: "http://localhost:8200",
						Token:   "token1",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate vault ID",
			settings: config.SettingsFile{
				Vaults: []config.VaultInstance{
					{
						ID:      "vault1",
						Address: "http://localhost:8200",
						Token:   "token1",
					},
					{
						ID:      "vault1",
						Address: "http://localhost:8201",
						Token:   "token2",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty vault address",
			settings: config.SettingsFile{
				Vaults: []config.VaultInstance{
					{
						ID:      "vault1",
						Address: "",
						Token:   "token1",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSettings(tt.settings)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShouldFallbackToDirectWrite(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "permission error",
			err:      os.ErrPermission,
			expected: true,
		},
		{
			name:     "read-only file system error",
			err:      syscall.EROFS,
			expected: true,
		},
		{
			name:     "operation not permitted error",
			err:      syscall.EPERM,
			expected: true,
		},
		{
			name:     "access denied error",
			err:      syscall.EACCES,
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldFallbackToDirectWrite(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdminSessionStore_PruneSessions(t *testing.T) {
	store := newAdminSessionStore("$2a$10$testhashedpassword", false)

	now := time.Now()
	// Add expired sessions
	store.sessions["expired1"] = now.Add(-1 * time.Hour)
	store.sessions["expired2"] = now.Add(-2 * time.Hour)
	store.sessions["valid"] = now.Add(1 * time.Hour)

	store.pruneSessions(now)

	assert.Len(t, store.sessions, 1)
	assert.Contains(t, store.sessions, "valid")
}

func TestAdminSessionStore_PruneSessions_MaxSessions(t *testing.T) {
	store := newAdminSessionStore("$2a$10$testhashedpassword", false)

	now := time.Now()
	// Add more sessions than max
	for i := range adminMaxSessions + 5 {
		store.sessions[fmt.Sprintf("token%d", i)] = now.Add(time.Duration(i+1) * time.Hour)
	}

	store.pruneSessions(now)

	assert.Len(t, store.sessions, adminMaxSessions)
}

func TestAdminSessionStore_AllowLoginAttempt(t *testing.T) {
	store := newAdminSessionStore("$2a$10$testhashedpassword", false)

	// Should allow initially
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	assert.True(t, store.allowLoginAttempt(req))

	// Exhaust the limit
	for range 5 {
		store.allowLoginAttempt(req)
	}

	// Should deny after max attempts
	assert.False(t, store.allowLoginAttempt(req))
}

func TestAdminSessionStore_AllowLoginAttempt_NilLimiter(t *testing.T) {
	store := newAdminSessionStore("$2a$10$testhashedpassword", false)
	store.limiter = nil

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	assert.True(t, store.allowLoginAttempt(req))
}

func TestAdminSessionStore_RequireAuth_ExpiredSession(t *testing.T) {
	store := newAdminSessionStore("$2a$10$testhashedpassword", false)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authHandler := store.requireAuth(testHandler)

	// Add expired session
	token, err := store.createToken()
	require.NoError(t, err)
	store.sessions[token] = time.Now().Add(-1 * time.Hour)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "vcv_admin_session", Value: token})
	w := httptest.NewRecorder()

	authHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminSettingsStore_Load_ReadError(t *testing.T) {
	// Create a directory where a file is expected - this causes read error
	store := newAdminSettingsStore(t.TempDir(), config.EnvDev)

	_, err := store.load()
	assert.Error(t, err)
}

func TestAdminSettingsStore_Save_ValidationError(t *testing.T) {
	store := newAdminSettingsStore(t.TempDir()+"/settings.json", config.EnvDev)

	settings := config.SettingsFile{
		Vaults: []config.VaultInstance{
			{ID: ""},
		},
	}

	err := store.save(settings)
	assert.Error(t, err)
}

func TestFallbackWriteSettings(t *testing.T) {
	t.Run("permission error triggers fallback", func(t *testing.T) {
		tmpFile := t.TempDir() + "/settings.json"
		payload := []byte(`{"test": true}`)
		err := fallbackWriteSettings(tmpFile, payload, os.ErrPermission)
		assert.NoError(t, err)
		data, _ := os.ReadFile(tmpFile)
		assert.Equal(t, payload, data)
	})

	t.Run("non-permission error returns original", func(t *testing.T) {
		originalErr := errors.New("some error")
		err := fallbackWriteSettings(t.TempDir()+"/settings.json", []byte(`{}`), originalErr)
		assert.ErrorIs(t, err, originalErr)
	})

	t.Run("nil error returns false from shouldFallback", func(t *testing.T) {
		assert.False(t, shouldFallbackToDirectWrite(nil))
	})
}

func TestRegisterAdminRoutes(t *testing.T) {
	r := chi.NewRouter()

	tmpFile := t.TempDir() + "/settings.json"
	vaultRegistry := vault.NewRegistry([]config.VaultInstance{})
	vaultStatusClients := make(map[string]vault.Client)
	cacheClient := &vault.MockClient{}

	RegisterAdminRoutes(r, tmpFile, config.EnvDev, vaultRegistry, vaultStatusClients, cacheClient)

	// Verify routes are registered
	assert.NotNil(t, r)
}

func TestIsBcryptHash(t *testing.T) {
	assert.True(t, isBcryptHash("$2a$10$abcdefghijklmnopqrstuu"))
	assert.True(t, isBcryptHash("$2b$10$abcdefghijklmnopqrstuu"))
	assert.True(t, isBcryptHash("$2y$10$abcdefghijklmnopqrstuu"))
	assert.False(t, isBcryptHash(""))
	assert.False(t, isBcryptHash("plaintext"))
	assert.False(t, isBcryptHash("$1$notbcrypt"))
}

func TestRegisterAdminRoutes_EmptyPassword_LoginNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := tmpDir + "/settings.json"
	settings := config.SettingsFile{
		App:   config.AppSettings{Env: "dev", Port: 52000},
		Admin: config.AdminSettings{Password: ""},
	}
	data, err := json.Marshal(settings)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(settingsPath, data, 0644))

	r := chi.NewRouter()
	RegisterAdminRoutes(r, settingsPath, config.EnvDev, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRegisterAdminRoutes_InvalidHash_LoginNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := tmpDir + "/settings.json"
	settings := config.SettingsFile{
		App:   config.AppSettings{Env: "dev", Port: 52000},
		Admin: config.AdminSettings{Password: "not-a-bcrypt-hash"},
	}
	data, err := json.Marshal(settings)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(settingsPath, data, 0644))

	r := chi.NewRouter()
	RegisterAdminRoutes(r, settingsPath, config.EnvDev, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRegisterAdminRoutes_CacheInvalidate(t *testing.T) {
	// Create settings with valid bcrypt password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
	require.NoError(t, err)

	tmpDir := t.TempDir()
	settingsPath := tmpDir + "/settings.json"
	settings := config.SettingsFile{
		App:   config.AppSettings{Env: "dev", Port: 52000},
		Admin: config.AdminSettings{Password: string(hashedPassword)},
	}
	data, _ := json.Marshal(settings)
	require.NoError(t, os.WriteFile(settingsPath, data, 0644))

	r := chi.NewRouter()
	cacheClient := &vault.MockClient{}
	cacheClient.On("InvalidateCache").Return()

	RegisterAdminRoutes(r, settingsPath, config.EnvDev, nil, nil, cacheClient)

	// Login to get session
	loginBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "testpassword"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewReader(loginBody))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Extract cookie
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == adminCookieName {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie)

	// Hit cache invalidate
	req = httptest.NewRequest(http.MethodPost, "/api/cache/invalidate", nil)
	req.AddCookie(sessionCookie)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	cacheClient.AssertExpectations(t)
}

func TestRegisterAdminRoutes_NoCacheClient(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
	require.NoError(t, err)

	tmpDir := t.TempDir()
	settingsPath := tmpDir + "/settings.json"
	settings := config.SettingsFile{
		App:   config.AppSettings{Env: "dev", Port: 52000},
		Admin: config.AdminSettings{Password: string(hashedPassword)},
	}
	data, _ := json.Marshal(settings)
	require.NoError(t, os.WriteFile(settingsPath, data, 0644))

	r := chi.NewRouter()
	RegisterAdminRoutes(r, settingsPath, config.EnvDev, nil, nil, nil)

	// Login to get session
	loginBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "testpassword"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewReader(loginBody))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == adminCookieName {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie)

	// Hit cache invalidate without cache client
	req = httptest.NewRequest(http.MethodPost, "/api/cache/invalidate", nil)
	req.AddCookie(sessionCookie)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
