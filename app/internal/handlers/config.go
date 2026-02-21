package handlers

import (
	"encoding/json"
	"net/http"

	"vcv/internal/config"
	"vcv/internal/logger"
	"vcv/internal/middleware"
	"vcv/internal/vault"
)

// ConfigResponse holds the public configuration exposed to the frontend.
type ConfigResponse struct {
	ExpirationThresholds struct {
		Critical int `json:"critical"`
		Warning  int `json:"warning"`
	} `json:"expirationThresholds"`
	Metrics struct {
		PerCertificate  bool `json:"per_certificate"`
		EnhancedMetrics bool `json:"enhanced_metrics"`
	} `json:"metrics"`
	PKIMounts []string              `json:"pkiMounts"`
	Vaults    []VaultConfigResponse `json:"vaults"`
}

type VaultConfigResponse struct {
	ID          string   `json:"id"`
	DisplayName string   `json:"displayName"`
	PKIMounts   []string `json:"pkiMounts"`
}

// GetConfig returns the application configuration.
func GetConfig(cfg config.Config, vaultRegistry *vault.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetRequestID(r.Context())

		resp := ConfigResponse{}
		resp.ExpirationThresholds.Critical = cfg.ExpirationThresholds.Critical
		resp.ExpirationThresholds.Warning = cfg.ExpirationThresholds.Warning
		resp.Metrics.PerCertificate = cfg.Metrics.PerCertificate
		resp.Metrics.EnhancedMetrics = cfg.Metrics.EnhancedMetrics
		resp.PKIMounts = cfg.Vault.PKIMounts
		if resp.PKIMounts == nil {
			resp.PKIMounts = []string{}
		}
		allVaults := cfg.AllVaults
		if len(allVaults) == 0 {
			allVaults = cfg.Vaults
		}
		resp.Vaults = make([]VaultConfigResponse, 0, len(allVaults))
		for _, instance := range allVaults {
			vaultID := instance.ID
			if vaultID == "" {
				continue
			}
			if vaultRegistry != nil && !vaultRegistry.IsEnabled(vaultID) {
				continue
			}
			displayName := instance.DisplayName
			if displayName == "" {
				displayName = vaultID
			}
			pkiMounts := instance.PKIMounts
			if pkiMounts == nil {
				pkiMounts = []string{}
			}
			resp.Vaults = append(resp.Vaults, VaultConfigResponse{ID: vaultID, DisplayName: displayName, PKIMounts: pkiMounts})
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to encode config response")
			return
		}

		logger.HTTPEvent(r.Method, r.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Msg("config retrieved")
	}
}
