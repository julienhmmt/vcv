package errors

import "errors"

var (
	ErrVaultIDEmpty     = errors.New("vault id is empty")
	ErrDuplicateVaultID = errors.New("duplicate vault id")
	ErrInvalidAddress   = errors.New("invalid vault address")
	ErrInvalidToken     = errors.New("invalid vault token")
	ErrInvalidThreshold = errors.New("invalid expiration threshold")
)
