package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorVariables(t *testing.T) {
	tests := []struct {
		name     string
		errVar   error
		expected string
	}{
		{
			name:     "ErrVaultIDEmpty",
			errVar:   ErrVaultIDEmpty,
			expected: "vault id is empty",
		},
		{
			name:     "ErrDuplicateVaultID",
			errVar:   ErrDuplicateVaultID,
			expected: "duplicate vault id",
		},
		{
			name:     "ErrInvalidAddress",
			errVar:   ErrInvalidAddress,
			expected: "invalid vault address",
		},
		{
			name:     "ErrInvalidToken",
			errVar:   ErrInvalidToken,
			expected: "invalid vault token",
		},
		{
			name:     "ErrNoPKIMounts",
			errVar:   ErrNoPKIMounts,
			expected: "no PKI mounts configured",
		},
		{
			name:     "ErrInvalidThreshold",
			errVar:   ErrInvalidThreshold,
			expected: "invalid expiration threshold",
		},
		{
			name:     "ErrSettingsNotFound",
			errVar:   ErrSettingsNotFound,
			expected: "settings file not found",
		},
		{
			name:     "ErrInvalidSettings",
			errVar:   ErrInvalidSettings,
			expected: "invalid settings format",
		},
		{
			name:     "ErrPasswordRequired",
			errVar:   ErrPasswordRequired,
			expected: "admin password required",
		},
		{
			name:     "ErrInvalidCredentials",
			errVar:   ErrInvalidCredentials,
			expected: "invalid credentials",
		},
		{
			name:     "ErrSessionExpired",
			errVar:   ErrSessionExpired,
			expected: "session expired",
		},
		{
			name:     "ErrUnauthorized",
			errVar:   ErrUnauthorized,
			expected: "unauthorized access",
		},
		{
			name:     "ErrCertificateNotFound",
			errVar:   ErrCertificateNotFound,
			expected: "certificate not found",
		},
		{
			name:     "ErrInvalidCertID",
			errVar:   ErrInvalidCertID,
			expected: "invalid certificate id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify error is not nil
			assert.NotNil(t, tt.errVar)

			// Verify error message
			assert.Equal(t, tt.expected, tt.errVar.Error())

			// Verify it's a proper error type
			var err error = tt.errVar
			assert.Error(t, err)
			assert.Equal(t, tt.expected, err.Error())
		})
	}
}

func TestErrorUniqueness(t *testing.T) {
	// Ensure all error variables are unique
	errorVars := []error{
		ErrVaultIDEmpty,
		ErrDuplicateVaultID,
		ErrInvalidAddress,
		ErrInvalidToken,
		ErrNoPKIMounts,
		ErrInvalidThreshold,
		ErrSettingsNotFound,
		ErrInvalidSettings,
		ErrPasswordRequired,
		ErrInvalidCredentials,
		ErrSessionExpired,
		ErrUnauthorized,
		ErrCertificateNotFound,
		ErrInvalidCertID,
	}

	// Check that all errors are unique
	seen := make(map[string]bool)
	for _, err := range errorVars {
		msg := err.Error()
		assert.False(t, seen[msg], "Error message '%s' is duplicated", msg)
		seen[msg] = true
	}

	// Verify we have the expected number of unique errors
	assert.Equal(t, 14, len(seen), "Expected 14 unique error messages")
}

func TestErrorComparison(t *testing.T) {
	// Test that errors.Is works correctly
	customErr := errors.New("vault id is empty")
	assert.True(t, errors.Is(ErrVaultIDEmpty, ErrVaultIDEmpty))
	assert.False(t, errors.Is(ErrVaultIDEmpty, ErrInvalidAddress))
	assert.False(t, errors.Is(ErrVaultIDEmpty, customErr))

	// Test error wrapping scenarios
	wrappedErr := fmt.Errorf("wrapped: %w", ErrVaultIDEmpty)
	assert.True(t, errors.Is(wrappedErr, ErrVaultIDEmpty))
	assert.True(t, errors.Unwrap(wrappedErr) == ErrVaultIDEmpty)
}

func TestErrorTypes(t *testing.T) {
	// Verify all error variables are of type error
	errorVars := []struct {
		name string
		err  error
	}{
		{"ErrVaultIDEmpty", ErrVaultIDEmpty},
		{"ErrDuplicateVaultID", ErrDuplicateVaultID},
		{"ErrInvalidAddress", ErrInvalidAddress},
		{"ErrInvalidToken", ErrInvalidToken},
		{"ErrNoPKIMounts", ErrNoPKIMounts},
		{"ErrInvalidThreshold", ErrInvalidThreshold},
		{"ErrSettingsNotFound", ErrSettingsNotFound},
		{"ErrInvalidSettings", ErrInvalidSettings},
		{"ErrPasswordRequired", ErrPasswordRequired},
		{"ErrInvalidCredentials", ErrInvalidCredentials},
		{"ErrSessionExpired", ErrSessionExpired},
		{"ErrUnauthorized", ErrUnauthorized},
		{"ErrCertificateNotFound", ErrCertificateNotFound},
		{"ErrInvalidCertID", ErrInvalidCertID},
	}

	for _, tt := range errorVars {
		t.Run(tt.name+" type check", func(t *testing.T) {
			assert.Implements(t, (*error)(nil), tt.err)
			assert.IsType(t, errors.New(""), tt.err)
		})
	}
}

func TestErrorMessages(t *testing.T) {
	// Test that error messages are descriptive and follow expected patterns
	tests := []struct {
		err      error
		expected string
	}{
		{ErrVaultIDEmpty, "vault id is empty"},
		{ErrDuplicateVaultID, "duplicate vault id"},
		{ErrInvalidAddress, "invalid vault address"},
		{ErrInvalidToken, "invalid vault token"},
		{ErrNoPKIMounts, "no PKI mounts configured"},
		{ErrInvalidThreshold, "invalid expiration threshold"},
		{ErrSettingsNotFound, "settings file not found"},
		{ErrInvalidSettings, "invalid settings format"},
		{ErrPasswordRequired, "admin password required"},
		{ErrInvalidCredentials, "invalid credentials"},
		{ErrSessionExpired, "session expired"},
		{ErrUnauthorized, "unauthorized access"},
		{ErrCertificateNotFound, "certificate not found"},
		{ErrInvalidCertID, "invalid certificate id"},
	}

	for _, tt := range tests {
		t.Run("message: "+tt.expected, func(t *testing.T) {
			msg := tt.err.Error()
			assert.Equal(t, tt.expected, msg)
			assert.NotEmpty(t, msg)
			assert.True(t, len(msg) > 0, "Error message should not be empty")
		})
	}
}

func TestErrorCategorization(t *testing.T) {
	// Group errors by category for testing
	vaultErrors := []error{
		ErrVaultIDEmpty,
		ErrDuplicateVaultID,
		ErrInvalidAddress,
		ErrInvalidToken,
		ErrNoPKIMounts,
	}

	configErrors := []error{
		ErrInvalidThreshold,
		ErrSettingsNotFound,
		ErrInvalidSettings,
	}

	authErrors := []error{
		ErrPasswordRequired,
		ErrInvalidCredentials,
		ErrSessionExpired,
		ErrUnauthorized,
	}

	certErrors := []error{
		ErrCertificateNotFound,
		ErrInvalidCertID,
	}

	t.Run("vault errors", func(t *testing.T) {
		for _, err := range vaultErrors {
			assert.Error(t, err)
			assert.NotEmpty(t, err.Error())
		}
		assert.Equal(t, 5, len(vaultErrors))
	})

	t.Run("config errors", func(t *testing.T) {
		for _, err := range configErrors {
			assert.Error(t, err)
			assert.NotEmpty(t, err.Error())
		}
		assert.Equal(t, 3, len(configErrors))
	})

	t.Run("auth errors", func(t *testing.T) {
		for _, err := range authErrors {
			assert.Error(t, err)
			assert.NotEmpty(t, err.Error())
		}
		assert.Equal(t, 4, len(authErrors))
	})

	t.Run("certificate errors", func(t *testing.T) {
		for _, err := range certErrors {
			assert.Error(t, err)
			assert.NotEmpty(t, err.Error())
		}
		assert.Equal(t, 2, len(certErrors))
	})
}

func TestIsVaultError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"vault id empty", ErrVaultIDEmpty, true},
		{"duplicate vault id", ErrDuplicateVaultID, true},
		{"invalid address", ErrInvalidAddress, true},
		{"invalid token", ErrInvalidToken, true},
		{"no PKI mounts", ErrNoPKIMounts, true},
		{"invalid threshold", ErrInvalidThreshold, false},
		{"settings not found", ErrSettingsNotFound, false},
		{"password required", ErrPasswordRequired, false},
		{"certificate not found", ErrCertificateNotFound, false},
		{"nil error", nil, false},
		{"custom vault error", errors.New("vault id is empty"), true}, // Same message
		{"custom non-vault error", errors.New("some other error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsVaultError(tt.err))
		})
	}
}

func TestIsConfigError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"invalid threshold", ErrInvalidThreshold, true},
		{"settings not found", ErrSettingsNotFound, true},
		{"invalid settings", ErrInvalidSettings, true},
		{"vault id empty", ErrVaultIDEmpty, false},
		{"password required", ErrPasswordRequired, false},
		{"certificate not found", ErrCertificateNotFound, false},
		{"nil error", nil, false},
		{"custom config error", errors.New("invalid settings format"), true}, // Same message
		{"custom non-config error", errors.New("some other error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsConfigError(tt.err))
		})
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"password required", ErrPasswordRequired, true},
		{"invalid credentials", ErrInvalidCredentials, true},
		{"session expired", ErrSessionExpired, true},
		{"unauthorized", ErrUnauthorized, true},
		{"vault id empty", ErrVaultIDEmpty, false},
		{"invalid threshold", ErrInvalidThreshold, false},
		{"certificate not found", ErrCertificateNotFound, false},
		{"nil error", nil, false},
		{"custom auth error", errors.New("invalid credentials"), true}, // Same message
		{"custom non-auth error", errors.New("some other error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsAuthError(tt.err))
		})
	}
}

func TestIsCertificateError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"certificate not found", ErrCertificateNotFound, true},
		{"invalid cert id", ErrInvalidCertID, true},
		{"vault id empty", ErrVaultIDEmpty, false},
		{"password required", ErrPasswordRequired, false},
		{"invalid threshold", ErrInvalidThreshold, false},
		{"nil error", nil, false},
		{"custom cert error", errors.New("certificate not found"), true}, // Same message
		{"custom non-cert error", errors.New("some other error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsCertificateError(tt.err))
		})
	}
}

func TestGetErrorCategory(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"vault error", ErrVaultIDEmpty, "vault"},
		{"config error", ErrInvalidThreshold, "config"},
		{"auth error", ErrPasswordRequired, "auth"},
		{"certificate error", ErrCertificateNotFound, "certificate"},
		{"nil error", nil, "unknown"},
		{"custom vault error", errors.New("vault id is empty"), "vault"},
		{"custom unknown error", errors.New("completely unknown error"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetErrorCategory(tt.err))
		})
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		context  string
		expected string
	}{
		{
			name:     "wrap vault error",
			err:      ErrVaultIDEmpty,
			context:  "failed to validate vault",
			expected: "failed to validate vault: vault id is empty",
		},
		{
			name:     "wrap nil error",
			err:      nil,
			context:  "some context",
			expected: "",
		},
		{
			name:     "wrap with empty context",
			err:      ErrInvalidToken,
			context:  "",
			expected: ": invalid vault token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Wrap(tt.err, tt.context)

			if tt.err == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected, result.Error())
				assert.True(t, errors.Is(result, tt.err))
			}
		})
	}
}

func TestWrapf(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "wrapf with formatting",
			err:      ErrCertificateNotFound,
			format:   "failed to find certificate %s in vault %s",
			args:     []interface{}{"cert-123", "vault-1"},
			expected: "failed to find certificate cert-123 in vault vault-1: certificate not found",
		},
		{
			name:     "wrapf nil error",
			err:      nil,
			format:   "some format %s",
			args:     []interface{}{"arg"},
			expected: "",
		},
		{
			name:     "wrapf without args",
			err:      ErrInvalidAddress,
			format:   "validation failed",
			args:     []interface{}{},
			expected: "validation failed: invalid vault address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Wrapf(tt.err, tt.format, tt.args...)

			if tt.err == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected, result.Error())
				assert.True(t, errors.Is(result, tt.err))
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"connection refused error", errors.New("connection refused"), true},
		{"timeout error", errors.New("request timeout"), true},
		{"temporary error", errors.New("temporary failure"), true},
		{"network error", errors.New("network unreachable"), true},
		{"unavailable error", errors.New("service unavailable"), true},
		{"config error", ErrInvalidThreshold, false},
		{"auth error", ErrInvalidCredentials, false},
		{"vault error", ErrVaultIDEmpty, false},
		{"certificate error", ErrCertificateNotFound, false},
		{"nil error", nil, false},
		{"unknown error", errors.New("some random error"), false},
		{"mixed case network", errors.New("Network error"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsRetryable(tt.err))
		})
	}
}
