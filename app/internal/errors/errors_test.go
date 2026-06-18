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
		{"ErrVaultIDEmpty", ErrVaultIDEmpty, "vault id is empty"},
		{"ErrDuplicateVaultID", ErrDuplicateVaultID, "duplicate vault id"},
		{"ErrInvalidAddress", ErrInvalidAddress, "invalid vault address"},
		{"ErrInvalidToken", ErrInvalidToken, "invalid vault token"},
		{"ErrInvalidThreshold", ErrInvalidThreshold, "invalid expiration threshold"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.errVar)
			assert.Equal(t, tt.expected, tt.errVar.Error())
		})
	}
}

func TestErrorUniqueness(t *testing.T) {
	errorVars := []error{
		ErrVaultIDEmpty,
		ErrDuplicateVaultID,
		ErrInvalidAddress,
		ErrInvalidToken,
		ErrInvalidThreshold,
	}

	seen := make(map[string]bool)
	for _, err := range errorVars {
		msg := err.Error()
		assert.False(t, seen[msg], "Error message '%s' is duplicated", msg)
		seen[msg] = true
	}
	assert.Equal(t, len(errorVars), len(seen))
}

func TestErrorComparison(t *testing.T) {
	assert.True(t, errors.Is(ErrVaultIDEmpty, ErrVaultIDEmpty))
	assert.False(t, errors.Is(ErrVaultIDEmpty, ErrInvalidAddress))

	wrappedErr := fmt.Errorf("wrapped: %w", ErrVaultIDEmpty)
	assert.True(t, errors.Is(wrappedErr, ErrVaultIDEmpty))
	assert.Equal(t, ErrVaultIDEmpty, errors.Unwrap(wrappedErr))
}
