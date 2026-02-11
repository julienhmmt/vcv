package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadMetricsConfig(t *testing.T) {
	tests := []struct {
		name                   string
		envPerCertificate      string
		envEnhanced            string
		expectedPerCertificate bool
		expectedEnhanced       bool
	}{
		{
			name:                   "defaults",
			envPerCertificate:      "",
			envEnhanced:            "",
			expectedPerCertificate: false,
			expectedEnhanced:       true,
		},
		{
			name:                   "per certificate enabled",
			envPerCertificate:      "true",
			envEnhanced:            "",
			expectedPerCertificate: true,
			expectedEnhanced:       true,
		},
		{
			name:                   "enhanced disabled",
			envPerCertificate:      "",
			envEnhanced:            "false",
			expectedPerCertificate: false,
			expectedEnhanced:       false,
		},
		{
			name:                   "both enabled",
			envPerCertificate:      "1",
			envEnhanced:            "yes",
			expectedPerCertificate: true,
			expectedEnhanced:       true,
		},
		{
			name:                   "both disabled",
			envPerCertificate:      "0",
			envEnhanced:            "no",
			expectedPerCertificate: false,
			expectedEnhanced:       false,
		},
		{
			name:                   "invalid values fallback to defaults",
			envPerCertificate:      "invalid",
			envEnhanced:            "maybe",
			expectedPerCertificate: false,
			expectedEnhanced:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			os.Unsetenv("VCV_METRICS_PER_CERTIFICATE")
			os.Unsetenv("VCV_METRICS_ENHANCED")

			// Set test environment variables
			if tt.envPerCertificate != "" {
				os.Setenv("VCV_METRICS_PER_CERTIFICATE", tt.envPerCertificate)
			}
			if tt.envEnhanced != "" {
				os.Setenv("VCV_METRICS_ENHANCED", tt.envEnhanced)
			}

			// Test loadMetricsConfig function
			config := loadMetricsConfig()
			assert.Equal(t, tt.expectedPerCertificate, config.PerCertificate, "PerCertificate mismatch")
			assert.Equal(t, tt.expectedEnhanced, config.EnhancedMetrics, "EnhancedMetrics mismatch")

			// Clean up
			os.Unsetenv("VCV_METRICS_PER_CERTIFICATE")
			os.Unsetenv("VCV_METRICS_ENHANCED")
		})
	}
}

func TestMetricsConfigFromSettings(t *testing.T) {
	tests := []struct {
		name                   string
		settingsFile           string
		expectedPerCertificate bool
		expectedEnhanced       bool
	}{
		{
			name:                   "settings with metrics enabled",
			settingsFile:           `{"app":{"env":"dev"},"metrics":{"per_certificate":true,"enhanced_metrics":true}}`,
			expectedPerCertificate: true,
			expectedEnhanced:       true,
		},
		{
			name:                   "settings with metrics disabled",
			settingsFile:           `{"app":{"env":"dev"},"metrics":{"per_certificate":false,"enhanced_metrics":false}}`,
			expectedPerCertificate: false,
			expectedEnhanced:       false,
		},
		{
			name:                   "settings without metrics section",
			settingsFile:           `{"app":{"env":"dev"}}`,
			expectedPerCertificate: false,
			expectedEnhanced:       true, // defaults
		},
		{
			name:                   "settings with partial metrics",
			settingsFile:           `{"app":{"env":"dev"},"metrics":{"per_certificate":true}}`,
			expectedPerCertificate: true,
			expectedEnhanced:       true, // defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary settings file
			tmpFile := t.TempDir() + "/settings.json"
			err := os.WriteFile(tmpFile, []byte(tt.settingsFile), 0644)
			require.NoError(t, err)

			// Set settings path environment
			os.Setenv("SETTINGS_PATH", tmpFile)
			defer os.Unsetenv("SETTINGS_PATH")

			// Clear environment variables that might interfere
			os.Unsetenv("VCV_METRICS_PER_CERTIFICATE")
			os.Unsetenv("VCV_METRICS_ENHANCED")

			// Load configuration
			cfg, err := Load()
			require.NoError(t, err)

			// Verify metrics configuration
			assert.Equal(t, tt.expectedPerCertificate, cfg.Metrics.PerCertificate, "PerCertificate mismatch")
			assert.Equal(t, tt.expectedEnhanced, cfg.Metrics.EnhancedMetrics, "EnhancedMetrics mismatch")
		})
	}
}

func TestParseBoolEnv(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		fallback bool
		expected bool
	}{
		{
			name:     "empty string returns fallback",
			value:    "",
			fallback: true,
			expected: true,
		},
		{
			name:     "true variants",
			value:    "true",
			fallback: false,
			expected: true,
		},
		{
			name:     "1 returns true",
			value:    "1",
			fallback: false,
			expected: true,
		},
		{
			name:     "yes returns true",
			value:    "yes",
			fallback: false,
			expected: true,
		},
		{
			name:     "y returns true",
			value:    "y",
			fallback: false,
			expected: true,
		},
		{
			name:     "on returns true",
			value:    "on",
			fallback: false,
			expected: true,
		},
		{
			name:     "false variants",
			value:    "false",
			fallback: true,
			expected: false,
		},
		{
			name:     "0 returns false",
			value:    "0",
			fallback: true,
			expected: false,
		},
		{
			name:     "no returns false",
			value:    "no",
			fallback: true,
			expected: false,
		},
		{
			name:     "n returns false",
			value:    "n",
			fallback: true,
			expected: false,
		},
		{
			name:     "off returns false",
			value:    "off",
			fallback: true,
			expected: false,
		},
		{
			name:     "case insensitive true",
			value:    "TRUE",
			fallback: false,
			expected: true,
		},
		{
			name:     "case insensitive false",
			value:    "FALSE",
			fallback: true,
			expected: false,
		},
		{
			name:     "whitespace trimmed",
			value:    " true ",
			fallback: false,
			expected: true,
		},
		{
			name:     "invalid returns fallback",
			value:    "invalid",
			fallback: true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TEST_ENV_VAR", tt.value)
			result := parseBoolEnv("TEST_ENV_VAR", tt.fallback)
			assert.Equal(t, tt.expected, result)
			os.Unsetenv("TEST_ENV_VAR")
		})
	}
}
