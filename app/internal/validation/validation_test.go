package validation

import (
	"testing"

	vcverrors "vcv/internal/errors"
)

func TestValidateVaultID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "valid id",
			id:      "vault-1",
			wantErr: nil,
		},
		{
			name:    "empty id",
			id:      "",
			wantErr: vcverrors.ErrVaultIDEmpty,
		},
		{
			name:    "whitespace only",
			id:      "   ",
			wantErr: vcverrors.ErrVaultIDEmpty,
		},
		{
			name:    "valid id with spaces",
			id:      "  vault-prod  ",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVaultID(tt.id)
			if err != tt.wantErr {
				t.Errorf("ValidateVaultID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateVaultAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr error
	}{
		{
			name:    "valid https address",
			address: "https://vault.example.com",
			wantErr: nil,
		},
		{
			name:    "valid http address",
			address: "http://localhost:8200",
			wantErr: nil,
		},
		{
			name:    "empty address",
			address: "",
			wantErr: vcverrors.ErrInvalidAddress,
		},
		{
			name:    "whitespace only",
			address: "   ",
			wantErr: vcverrors.ErrInvalidAddress,
		},
		{
			name:    "invalid address",
			address: "not-a-url",
			wantErr: vcverrors.ErrInvalidAddress,
		},
		{
			name:    "valid address with spaces",
			address: "  https://vault.example.com  ",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVaultAddress(tt.address)
			if err != tt.wantErr {
				t.Errorf("ValidateVaultAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateVaultToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{
			name:    "valid token",
			token:   "s.1234567890",
			wantErr: nil,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: vcverrors.ErrInvalidToken,
		},
		{
			name:    "whitespace only",
			token:   "   ",
			wantErr: vcverrors.ErrInvalidToken,
		},
		{
			name:    "valid token with spaces",
			token:   "  s.abcdef  ",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVaultToken(tt.token)
			if err != tt.wantErr {
				t.Errorf("ValidateVaultToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePKIMounts(t *testing.T) {
	tests := []struct {
		name          string
		mounts        []string
		fallbackMount string
		wantMounts    []string
		wantErr       error
	}{
		{
			name:          "valid mounts",
			mounts:        []string{"pki", "pki_prod"},
			fallbackMount: "",
			wantMounts:    []string{"pki", "pki_prod"},
			wantErr:       nil,
		},
		{
			name:          "empty mounts with fallback",
			mounts:        []string{},
			fallbackMount: "pki",
			wantMounts:    []string{"pki"},
			wantErr:       nil,
		},
		{
			name:          "empty mounts without fallback",
			mounts:        []string{},
			fallbackMount: "",
			wantMounts:    nil,
			wantErr:       vcverrors.ErrNoPKIMounts,
		},
		{
			name:          "mounts with empty strings",
			mounts:        []string{"pki", "", "pki_prod", "  "},
			fallbackMount: "",
			wantMounts:    []string{"pki", "pki_prod"},
			wantErr:       nil,
		},
		{
			name:          "all empty mounts",
			mounts:        []string{"", "  ", "   "},
			fallbackMount: "",
			wantMounts:    nil,
			wantErr:       vcverrors.ErrNoPKIMounts,
		},
		{
			name:          "mounts with spaces",
			mounts:        []string{"  pki  ", "  pki_prod  "},
			fallbackMount: "",
			wantMounts:    []string{"pki", "pki_prod"},
			wantErr:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMounts, err := ValidatePKIMounts(tt.mounts, tt.fallbackMount)
			if err != tt.wantErr {
				t.Errorf("ValidatePKIMounts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotMounts) != len(tt.wantMounts) {
				t.Errorf("ValidatePKIMounts() gotMounts length = %v, want %v", len(gotMounts), len(tt.wantMounts))
				return
			}
			for i := range gotMounts {
				if gotMounts[i] != tt.wantMounts[i] {
					t.Errorf("ValidatePKIMounts() gotMounts[%d] = %v, want %v", i, gotMounts[i], tt.wantMounts[i])
				}
			}
		})
	}
}

func TestValidateExpirationThreshold(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		wantErr error
	}{
		{
			name:    "valid threshold 7",
			value:   7,
			wantErr: nil,
		},
		{
			name:    "valid threshold 30",
			value:   30,
			wantErr: nil,
		},
		{
			name:    "valid threshold 365",
			value:   365,
			wantErr: nil,
		},
		{
			name:    "valid threshold 0",
			value:   0,
			wantErr: nil,
		},
		{
			name:    "negative threshold",
			value:   -1,
			wantErr: vcverrors.ErrInvalidThreshold,
		},
		{
			name:    "threshold too large",
			value:   366,
			wantErr: vcverrors.ErrInvalidThreshold,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExpirationThreshold(tt.value)
			if err != tt.wantErr {
				t.Errorf("ValidateExpirationThreshold() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
