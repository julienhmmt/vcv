package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger is the application-wide logger type, aliased to zerolog.Logger.
// This allows other packages to depend only on vcv/internal/logger instead of importing zerolog directly.
type Logger = zerolog.Logger

// Init initializes the logger with the specified log level
func Init(level string) {
	// Set time format
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Read logging configuration from environment
	outputMode := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_OUTPUT")))
	if outputMode == "" {
		outputMode = "stdout"
	}

	format := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_FORMAT")))
	if format == "" {
		format = "console"
	}

	logFilePath := strings.TrimSpace(os.Getenv("LOG_FILE_PATH"))

	stdoutEnabled := outputMode == "stdout" || outputMode == "both"
	fileEnabled := outputMode == "file" || outputMode == "both"

	writers := make([]io.Writer, 0, 2)
	deferredWarnings := make([]string, 0, 2)

	// Configure stdout writer
	if stdoutEnabled {
		if format == "json" {
			writers = append(writers, os.Stdout)
		} else {
			consoleWriter := zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: "2006-01-02 15:04:05",
			}
			writers = append(writers, consoleWriter)
		}
	}

	// Configure file writer if requested
	if fileEnabled {
		if logFilePath == "" {
			deferredWarnings = append(deferredWarnings, "LOG_OUTPUT requires a file but LOG_FILE_PATH is not set; disabling file logging")
		} else {
			file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				deferredWarnings = append(deferredWarnings, fmt.Sprintf("Failed to open log file '%s', disabling file logging: %v", logFilePath, err))
			} else {
				if format == "json" {
					writers = append(writers, file)
				} else {
					consoleWriter := zerolog.ConsoleWriter{
						Out:        file,
						TimeFormat: "2006-01-02 15:04:05",
					}
					writers = append(writers, consoleWriter)
				}
			}
		}
	}

	// Fallback: if no writers configured, use stdout console
	if len(writers) == 0 {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
		}
		writers = append(writers, consoleWriter)
		deferredWarnings = append(deferredWarnings, "No valid log output configured, falling back to stdout console")
		stdoutEnabled = true
		fileEnabled = false
		logFilePath = ""
	}

	var output io.Writer
	if len(writers) == 1 {
		output = writers[0]
	} else {
		output = zerolog.MultiLevelWriter(writers...)
	}

	// Set log level, defaulting to InfoLevel if parsing fails.
	lvl, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		lvl = zerolog.InfoLevel
		log.Warn().Str("log_level_in", level).Msg("Invalid log level, defaulting to 'info'")
	}
	zerolog.SetGlobalLevel(lvl)

	// Log any deferred warnings about configuration
	for _, msg := range deferredWarnings {
		log.Warn().Msg(msg)
	}

	// Re-create logger with correct level after setting global level
	log.Logger = log.Output(output).Level(lvl)

	log.Info().
		Str("level", zerolog.GlobalLevel().String()).
		Str("output_mode", outputMode).
		Str("format", format).
		Bool("stdout_enabled", stdoutEnabled).
		Bool("file_enabled", fileEnabled).
		Str("log_file_path", logFilePath).
		Msg("Logger initialized")
}

// Get returns a pointer to the configured logger instance
func Get() *zerolog.Logger {
	return &log.Logger
}

// SetOutput changes the destination for log output.
// This is useful for redirecting logs to a file or a buffer during testing.
func SetOutput(w io.Writer) {
	log.Logger = log.Output(w)
}

// Event is an alias for zerolog.Event to allow building log entries without importing zerolog.
type Event = zerolog.Event

// HTTPEvent logs HTTP request events with standardized fields.
func HTTPEvent(method, path string, status int, durationMs float64) *zerolog.Event {
	return log.Info().
		Str("event_category", "http").
		Str("method", method).
		Str("path", path).
		Int("status", status).
		Float64("duration_ms", durationMs)
}

// HTTPError logs HTTP error events.
func HTTPError(method, path string, status int, err error) *zerolog.Event {
	return log.Error().
		Str("event_category", "http").
		Str("method", method).
		Str("path", path).
		Int("status", status).
		Err(err)
}

// PanicEvent logs panic recovery events.
func PanicEvent(err interface{}, stack string) *zerolog.Event {
	return log.Error().
		Str("event_category", "panic").
		Interface("error", err).
		Str("stack", stack)
}
