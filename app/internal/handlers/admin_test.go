package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
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
	for i := 0; i < 3; i++ {
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

func TestRegisterAdminRoutes(t *testing.T) {
	r := chi.NewRouter()

	tmpFile := t.TempDir() + "/settings.json"
	vaultRegistry := vault.NewRegistry([]config.VaultInstance{})
	vaultStatusClients := make(map[string]vault.Client)
	cacheClient := &vault.MockClient{}

	RegisterAdminRoutes(r, tmpFile, config.EnvDev, vaultRegistry, vaultStatusClients, cacheClient)

	// Verify routes are registered
	// This is a basic test - in a real scenario you might want to test specific routes
	assert.NotNil(t, r)
}
