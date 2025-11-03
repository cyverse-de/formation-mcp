package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHumanReadableHandler(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func(*slog.Logger)
		contains []string
	}{
		{
			name: "simple info message",
			logFunc: func(l *slog.Logger) {
				l.Info("test message")
			},
			contains: []string{
				"INFO ",
				"test message",
			},
		},
		{
			name: "with string attributes",
			logFunc: func(l *slog.Logger) {
				l.Info("api call", "endpoint", "/apps", "method", "GET")
			},
			contains: []string{
				"INFO ",
				"api call",
				"endpoint=/apps",
				"method=GET",
			},
		},
		{
			name: "with quoted string (contains spaces)",
			logFunc: func(l *slog.Logger) {
				l.Info("message", "description", "this has spaces")
			},
			contains: []string{
				"INFO ",
				"message",
				`description="this has spaces"`,
			},
		},
		{
			name: "with int attributes",
			logFunc: func(l *slog.Logger) {
				l.Info("message", "count", 42, "status", 200)
			},
			contains: []string{
				"INFO ",
				"message",
				"count=42",
				"status=200",
			},
		},
		{
			name: "with duration",
			logFunc: func(l *slog.Logger) {
				l.Info("message", "duration", 123*time.Millisecond)
			},
			contains: []string{
				"INFO ",
				"message",
				"duration=123ms",
			},
		},
		{
			name: "error level",
			logFunc: func(l *slog.Logger) {
				l.Error("error occurred", "error", "something failed")
			},
			contains: []string{
				"ERROR",
				"error occurred",
				"error=\"something failed\"",
			},
		},
		{
			name: "debug level",
			logFunc: func(l *slog.Logger) {
				l.Debug("debug info", "detail", "verbose")
			},
			contains: []string{
				"DEBUG",
				"debug info",
				"detail=verbose",
			},
		},
		{
			name: "warn level",
			logFunc: func(l *slog.Logger) {
				l.Warn("warning", "issue", "deprecated")
			},
			contains: []string{
				"WARN ",
				"warning",
				"issue=deprecated",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			handler := NewHumanReadableHandler(&buf, slog.LevelDebug)
			logger := slog.New(handler)

			tt.logFunc(logger)

			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}

			// Verify timestamp format
			assert.True(t, strings.Contains(output, "T"))
			assert.True(t, strings.HasSuffix(output, "\n"))
		})
	}
}

func TestHumanReadableHandlerWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	handler := NewHumanReadableHandler(&buf, slog.LevelInfo)
	logger := slog.New(handler)

	// Create a logger with pre-attached attributes
	loggerWithAttrs := logger.With("request_id", "abc123", "user", "testuser")
	loggerWithAttrs.Info("operation completed", "status", "success")

	output := buf.String()
	assert.Contains(t, output, "request_id=abc123")
	assert.Contains(t, output, "user=testuser")
	assert.Contains(t, output, "status=success")
	assert.Contains(t, output, "operation completed")
}

func TestHumanReadableHandlerWithGroup(t *testing.T) {
	var buf bytes.Buffer
	handler := NewHumanReadableHandler(&buf, slog.LevelInfo)
	logger := slog.New(handler)

	// Create a logger with a group
	logger.WithGroup("http").Info("request", "method", "GET", "path", "/api")

	output := buf.String()
	assert.Contains(t, output, "http.method=GET")
	assert.Contains(t, output, "http.path=/api")
}

func TestHumanReadableHandlerLevel(t *testing.T) {
	tests := []struct {
		name         string
		handlerLevel slog.Level
		logLevel     slog.Level
		shouldLog    bool
	}{
		{
			name:         "debug handler logs debug",
			handlerLevel: slog.LevelDebug,
			logLevel:     slog.LevelDebug,
			shouldLog:    true,
		},
		{
			name:         "info handler filters debug",
			handlerLevel: slog.LevelInfo,
			logLevel:     slog.LevelDebug,
			shouldLog:    false,
		},
		{
			name:         "info handler logs info",
			handlerLevel: slog.LevelInfo,
			logLevel:     slog.LevelInfo,
			shouldLog:    true,
		},
		{
			name:         "warn handler filters info",
			handlerLevel: slog.LevelWarn,
			logLevel:     slog.LevelInfo,
			shouldLog:    false,
		},
		{
			name:         "warn handler logs error",
			handlerLevel: slog.LevelWarn,
			logLevel:     slog.LevelError,
			shouldLog:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			handler := NewHumanReadableHandler(&buf, tt.handlerLevel)
			logger := slog.New(handler)

			switch tt.logLevel {
			case slog.LevelDebug:
				logger.Debug("test")
			case slog.LevelInfo:
				logger.Info("test")
			case slog.LevelWarn:
				logger.Warn("test")
			case slog.LevelError:
				logger.Error("test")
			}

			if tt.shouldLog {
				assert.NotEmpty(t, buf.String())
			} else {
				assert.Empty(t, buf.String())
			}
		})
	}
}

func TestSetupJSONHandler(t *testing.T) {
	var buf bytes.Buffer
	logger := Setup(&buf, "info", true)

	logger.Info("test message", "key", "value")

	// Verify it's valid JSON
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "test message", logEntry["msg"])
	assert.Equal(t, "value", logEntry["key"])
	assert.Equal(t, "INFO", logEntry["level"])
}

func TestSetupHumanReadableHandler(t *testing.T) {
	var buf bytes.Buffer
	logger := Setup(&buf, "info", false)

	logger.Info("test message", "key", "value")

	output := buf.String()
	assert.Contains(t, output, "INFO ")
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "key=value")

	// Should NOT be JSON
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.Error(t, err) // Should fail to parse as JSON
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"invalid", slog.LevelInfo}, // Default to info
		{"", slog.LevelInfo},         // Default to info
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level := parseLevel(tt.input)
			assert.Equal(t, tt.expected, level)
		})
	}
}

func TestSetupWithDifferentLevels(t *testing.T) {
	tests := []struct {
		level          string
		debugShouldLog bool
		infoShouldLog  bool
		warnShouldLog  bool
		errorShouldLog bool
	}{
		{"debug", true, true, true, true},
		{"info", false, true, true, true},
		{"warn", false, false, true, true},
		{"error", false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			var buf bytes.Buffer
			logger := Setup(&buf, tt.level, false)

			buf.Reset()
			logger.Debug("debug")
			assert.Equal(t, tt.debugShouldLog, buf.Len() > 0)

			buf.Reset()
			logger.Info("info")
			assert.Equal(t, tt.infoShouldLog, buf.Len() > 0)

			buf.Reset()
			logger.Warn("warn")
			assert.Equal(t, tt.warnShouldLog, buf.Len() > 0)

			buf.Reset()
			logger.Error("error")
			assert.Equal(t, tt.errorShouldLog, buf.Len() > 0)
		})
	}
}
