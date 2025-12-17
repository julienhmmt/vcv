package logger_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// TestHTTPEvent verifies that HTTPEvent logs HTTP request events correctly.
func TestHTTPEvent(t *testing.T) {
	logger.Init("debug")

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.HTTPEvent("GET", "/api/test", 200, 150).Msg("")

	output := buf.String()
	if !strings.Contains(output, `"method":"GET"`) {
		t.Errorf("Expected log to contain HTTP method 'GET', got: %s", output)
	}
	if !strings.Contains(output, `"path":"/api/test"`) {
		t.Errorf("Expected log to contain path '/api/test', got: %s", output)
	}
	if !strings.Contains(output, `"status":200`) {
		t.Errorf("Expected log to contain status code '200', got: %s", output)
	}
	if !strings.Contains(output, `"duration_ms":150`) {
		t.Errorf("Expected log to contain duration '150', got: %s", output)
	}
}

// TestHTTPError verifies that HTTPError logs HTTP errors correctly.
func TestHTTPError(t *testing.T) {
	logger.Init("debug")

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	err := fmt.Errorf("test error")
	logger.HTTPError("POST", "/api/error", 500, err).Msg("")

	output := buf.String()
	if !strings.Contains(output, `"method":"POST"`) {
		t.Errorf("Expected log to contain HTTP method 'POST', got: %s", output)
	}
	if !strings.Contains(output, `"path":"/api/error"`) {
		t.Errorf("Expected log to contain path '/api/error', got: %s", output)
	}
	if !strings.Contains(output, `"status":500`) {
		t.Errorf("Expected log to contain status code '500', got: %s", output)
	}
	if !strings.Contains(output, `"error":"test error"`) {
		t.Errorf("Expected log to contain error message, got: %s", output)
	}
}

// TestPanicEvent verifies that PanicEvent logs panic events correctly.
func TestPanicEvent(t *testing.T) {
	logger.Init("debug")

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.PanicEvent("panic message", "stack trace").Msg("")

	output := buf.String()
	if !strings.Contains(output, `"error":"panic message"`) {
		t.Errorf("Expected log to contain panic message, got: %s", output)
	}
	if !strings.Contains(output, `"stack":"stack trace"`) {
		t.Errorf("Expected log to contain stack trace, got: %s", output)
	}
}

func TestInit_FileOutput_WritesToFile(t *testing.T) {
	setEnvHelper(t, "LOG_OUTPUT", "file")
	setEnvHelper(t, "LOG_FORMAT", "json")
	logFilePath := filepath.Join(t.TempDir(), "app.log")
	setEnvHelper(t, "LOG_FILE_PATH", logFilePath)
	t.Cleanup(func() {
		unsetEnvHelper(t, "LOG_OUTPUT")
		unsetEnvHelper(t, "LOG_FORMAT")
		unsetEnvHelper(t, "LOG_FILE_PATH")
	})
	logger.Init("info")
	logger.Get().Info().Msg("file output test")
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}
	if !strings.Contains(string(content), "file output test") {
		t.Fatalf("expected log file to contain message")
	}
}

func TestInit_FileOutput_MissingPath_DoesNotPanic(t *testing.T) {
	setEnvHelper(t, "LOG_OUTPUT", "file")
	unsetEnvHelper(t, "LOG_FILE_PATH")
	unsetEnvHelper(t, "LOG_FORMAT")
	t.Cleanup(func() {
		unsetEnvHelper(t, "LOG_OUTPUT")
		unsetEnvHelper(t, "LOG_FILE_PATH")
		unsetEnvHelper(t, "LOG_FORMAT")
	})
	logger.Init("info")
	logger.Get().Info().Msg("fallback output test")
}

func TestInit_InvalidOutput_FallsBackToStdout(t *testing.T) {
	setEnvHelper(t, "LOG_OUTPUT", "nope")
	unsetEnvHelper(t, "LOG_FILE_PATH")
	setEnvHelper(t, "LOG_FORMAT", "console")
	t.Cleanup(func() {
		unsetEnvHelper(t, "LOG_OUTPUT")
		unsetEnvHelper(t, "LOG_FILE_PATH")
		unsetEnvHelper(t, "LOG_FORMAT")
	})
	logger.Init("info")
	logger.Get().Info().Msg("invalid output fallback")
}

func TestInit_BothOutput_InvalidFilePath_DoesNotPanic(t *testing.T) {
	setEnvHelper(t, "LOG_OUTPUT", "both")
	setEnvHelper(t, "LOG_FORMAT", "json")
	setEnvHelper(t, "LOG_FILE_PATH", t.TempDir())
	t.Cleanup(func() {
		unsetEnvHelper(t, "LOG_OUTPUT")
		unsetEnvHelper(t, "LOG_FORMAT")
		unsetEnvHelper(t, "LOG_FILE_PATH")
	})
	logger.Init("info")
	logger.Get().Info().Msg("both output fallback")
}
