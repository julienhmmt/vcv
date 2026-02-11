package errors

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrVaultIDEmpty        = errors.New("vault id is empty")
	ErrDuplicateVaultID    = errors.New("duplicate vault id")
	ErrInvalidAddress      = errors.New("invalid vault address")
	ErrInvalidToken        = errors.New("invalid vault token")
	ErrNoPKIMounts         = errors.New("no PKI mounts configured")
	ErrInvalidThreshold    = errors.New("invalid expiration threshold")
	ErrSettingsNotFound    = errors.New("settings file not found")
	ErrInvalidSettings     = errors.New("invalid settings format")
	ErrPasswordRequired    = errors.New("admin password required")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrSessionExpired      = errors.New("session expired")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrCertificateNotFound = errors.New("certificate not found")
	ErrInvalidCertID       = errors.New("invalid certificate id")
)

// IsVaultError returns true if the error is vault-related
func IsVaultError(err error) bool {
	if err == nil {
		return false
	}

	vaultErrors := []error{
		ErrVaultIDEmpty,
		ErrDuplicateVaultID,
		ErrInvalidAddress,
		ErrInvalidToken,
		ErrNoPKIMounts,
	}

	for _, vaultErr := range vaultErrors {
		if err == vaultErr || err.Error() == vaultErr.Error() {
			return true
		}
	}
	return false
}

// IsConfigError returns true if the error is configuration-related
func IsConfigError(err error) bool {
	if err == nil {
		return false
	}

	configErrors := []error{
		ErrInvalidThreshold,
		ErrSettingsNotFound,
		ErrInvalidSettings,
	}

	for _, configErr := range configErrors {
		if err == configErr || err.Error() == configErr.Error() {
			return true
		}
	}
	return false
}

// IsAuthError returns true if the error is authentication-related
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}

	authErrors := []error{
		ErrPasswordRequired,
		ErrInvalidCredentials,
		ErrSessionExpired,
		ErrUnauthorized,
	}

	for _, authErr := range authErrors {
		if err == authErr || err.Error() == authErr.Error() {
			return true
		}
	}
	return false
}

// IsCertificateError returns true if the error is certificate-related
func IsCertificateError(err error) bool {
	if err == nil {
		return false
	}

	certErrors := []error{
		ErrCertificateNotFound,
		ErrInvalidCertID,
	}

	for _, certErr := range certErrors {
		if err == certErr || err.Error() == certErr.Error() {
			return true
		}
	}
	return false
}

// GetErrorCategory returns a category string for the error
func GetErrorCategory(err error) string {
	if err == nil {
		return "unknown"
	}

	if IsVaultError(err) {
		return "vault"
	}
	if IsConfigError(err) {
		return "config"
	}
	if IsAuthError(err) {
		return "auth"
	}
	if IsCertificateError(err) {
		return "certificate"
	}

	return "unknown"
}

// Wrap wraps an error with context
func Wrap(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// Wrapf wraps an error with formatted context
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

// IsRetryable returns true if the error might be resolved by retrying
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Network and temporary errors are typically retryable
	msg := strings.ToLower(err.Error())
	retryableKeywords := []string{
		"connection refused",
		"timeout",
		"temporary",
		"network",
		"unavailable",
	}

	for _, keyword := range retryableKeywords {
		if strings.Contains(msg, keyword) {
			return true
		}
	}

	// Configuration and authentication errors are typically not retryable
	if IsConfigError(err) || IsAuthError(err) {
		return false
	}

	return false
}
