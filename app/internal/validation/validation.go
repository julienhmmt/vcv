package validation

import (
	"net/url"
	"strings"

	vcverrors "vcv/internal/errors"
)

func ValidateVaultID(id string) error {
	if strings.TrimSpace(id) == "" {
		return vcverrors.ErrVaultIDEmpty
	}
	return nil
}

func ValidateVaultAddress(address string) error {
	trimmed := strings.TrimSpace(address)
	if trimmed == "" {
		return vcverrors.ErrInvalidAddress
	}
	if _, err := url.ParseRequestURI(trimmed); err != nil {
		return vcverrors.ErrInvalidAddress
	}
	return nil
}

func ValidateVaultToken(token string) error {
	if strings.TrimSpace(token) == "" {
		return vcverrors.ErrInvalidToken
	}
	return nil
}

func ValidatePKIMounts(mounts []string, fallbackMount string) ([]string, error) {
	if len(mounts) == 0 {
		if strings.TrimSpace(fallbackMount) == "" {
			return nil, vcverrors.ErrNoPKIMounts
		}
		return []string{strings.TrimSpace(fallbackMount)}, nil
	}

	result := make([]string, 0, len(mounts))
	for _, mount := range mounts {
		trimmed := strings.TrimSpace(mount)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return nil, vcverrors.ErrNoPKIMounts
	}

	return result, nil
}

func ValidateExpirationThreshold(value int) error {
	if value < 0 || value > 365 {
		return vcverrors.ErrInvalidThreshold
	}
	return nil
}
