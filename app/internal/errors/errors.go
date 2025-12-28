package errors

import "errors"

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
