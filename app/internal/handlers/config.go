package handlers

import (
	"encoding/json"
	"net/http"

	"vcv/config"
	"vcv/internal/logger"
	"vcv/middleware"
)

// ConfigResponse holds the public configuration exposed to the frontend.
type ConfigResponse struct {
	ExpirationThresholds struct {
		Critical int `json:"critical"`
		Warning  int `json:"warning"`
	} `json:"expirationThresholds"`
	PKIMounts []string              `json:"pkiMounts"`
	Vaults    []VaultConfigResponse `json:"vaults"`
}

type VaultConfigResponse struct {
	ID          string   `json:"id"`
	DisplayName string   `json:"displayName"`
	PKIMounts   []string `json:"pkiMounts"`
}

// GetConfig returns the application configuration.
func GetConfig(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetRequestID(r.Context())

		resp := ConfigResponse{}
		resp.ExpirationThresholds.Critical = cfg.ExpirationThresholds.Critical
		resp.ExpirationThresholds.Warning = cfg.ExpirationThresholds.Warning
		resp.PKIMounts = cfg.Vault.PKIMounts
		if resp.PKIMounts == nil {
			resp.PKIMounts = []string{}
		}
		resp.Vaults = make([]VaultConfigResponse, 0, len(cfg.Vaults))
		for _, instance := range cfg.Vaults {
			vaultID := instance.ID
			if vaultID == "" {
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
		w.WriteHeader(http.StatusOK)

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
