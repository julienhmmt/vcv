package logger_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"vcv/internal/logger"
)

// setEnvHelper sets an environment variable and fails the test if it cannot.
func setEnvHelper(t *testing.T, key, value string) {
	t.Helper()
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed to set env %s: %v", key, err)
	}
}

// unsetEnvHelper unsets an environment variable and fails the test if it cannot.
func unsetEnvHelper(t *testing.T, key string) {
	t.Helper()
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("failed to unset env %s: %v", key, err)
	}
}

// TestInit verifies that the logger initializes correctly with various log levels.
func TestInit(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		logLevel string // level to log at
		wantLog  bool   // whether we expect the message to appear
	}{
		{"debug level logs debug", "debug", "debug", true},
		{"debug level logs info", "debug", "info", true},
		{"info level logs info", "info", "info", true},
		{"info level skips debug", "info", "debug", false},
		{"warn level logs warn", "warn", "warn", true},
		{"warn level skips info", "warn", "info", false},
		{"error level logs error", "error", "error", true},
		{"error level skips warn", "error", "warn", false},
		{"invalid level defaults to info", "invalid", "info", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset environment
			unsetEnvHelper(t, "LOG_OUTPUT")
			unsetEnvHelper(t, "LOG_FILE_PATH")
			unsetEnvHelper(t, "LOG_FORMAT")

			// Capture output
			var buf bytes.Buffer
			logger.Init(tt.level)
			logger.SetOutput(&buf)

			// Log a message at the specified level
			switch tt.logLevel {
			case "debug":
				logger.Get().Debug().Msg("test message")
			case "info":
				logger.Get().Info().Msg("test message")
			case "warn":
				logger.Get().Warn().Msg("test message")
			case "error":
				logger.Get().Error().Msg("test message")
			}

			hasMessage := strings.Contains(buf.String(), "test message")
			if tt.wantLog && !hasMessage {
				t.Errorf("Expected log output to contain 'test message', got: %s", buf.String())
			}
			if !tt.wantLog && hasMessage {
				t.Errorf("Expected log output NOT to contain 'test message', got: %s", buf.String())
			}
		})
	}
}

// TestLoggerGet verifies that Get returns a non-nil logger.
func TestLoggerGet(t *testing.T) {
	logger.Init("info")
	log := logger.Get()
	if log == nil {
		t.Error("Expected Get() to return a non-nil logger")
	}
}

// TestSetOutput verifies that SetOutput changes the log destination.
func TestSetOutput(t *testing.T) {
	logger.Init("info")

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.Get().Info().Msg("custom output test")

	if !strings.Contains(buf.String(), "custom output test") {
		t.Errorf("Expected log output to contain 'custom output test', got: %s", buf.String())
	}
}

// TestJSONFormat verifies that JSON format produces valid JSON output.
func TestJSONFormat(t *testing.T) {
	setEnvHelper(t, "LOG_OUTPUT", "stdout")
	setEnvHelper(t, "LOG_FORMAT", "json")
	defer func() {
		unsetEnvHelper(t, "LOG_OUTPUT")
		unsetEnvHelper(t, "LOG_FORMAT")
	}()

	logger.Init("info")

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.Get().Info().Str("key", "value").Msg("json test")

	// Parse the JSON output
	output := strings.TrimSpace(buf.String())
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v, output: %s", err, output)
	}

	// Verify expected fields
	if result["message"] != "json test" {
		t.Errorf("Expected message 'json test', got: %v", result["message"])
	}
	if result["key"] != "value" {
		t.Errorf("Expected key 'value', got: %v", result["key"])
	}
}
