// Package logging provides structured logging with both human-readable and JSON formats.
package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"
)

// HumanReadableHandler is a custom slog.Handler that outputs human-readable,
// yet still parseable, structured logs in the format:
// 2025-11-03T10:15:30.123 INFO message key1=value1 key2=value2
type HumanReadableHandler struct {
	out   io.Writer
	level slog.Level
	attrs []slog.Attr
	group string
}

// NewHumanReadableHandler creates a new HumanReadableHandler.
func NewHumanReadableHandler(w io.Writer, level slog.Level) *HumanReadableHandler {
	return &HumanReadableHandler{
		out:   w,
		level: level,
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *HumanReadableHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle handles the record.
func (h *HumanReadableHandler) Handle(_ context.Context, r slog.Record) error {
	// Format: 2025-11-03T10:15:30.123 INFO message key1=value1 key2=value2
	buf := &strings.Builder{}

	// Timestamp in ISO8601 format with milliseconds
	buf.WriteString(r.Time.Format("2006-01-02T15:04:05.000"))
	buf.WriteString(" ")

	// Level (uppercase, padded to 5 chars for alignment)
	level := r.Level.String()
	buf.WriteString(strings.ToUpper(level))
	if len(level) < 5 {
		buf.WriteString(strings.Repeat(" ", 5-len(level)))
	}
	buf.WriteString(" ")

	// Message
	buf.WriteString(r.Message)

	// Handler attributes (from WithAttrs)
	for _, attr := range h.attrs {
		buf.WriteString(" ")
		h.formatAttr(buf, attr)
	}

	// Record attributes
	r.Attrs(func(a slog.Attr) bool {
		buf.WriteString(" ")
		h.formatAttr(buf, a)
		return true
	})

	buf.WriteString("\n")
	_, err := h.out.Write([]byte(buf.String()))
	return err
}

// formatAttr formats an attribute as key=value
func (h *HumanReadableHandler) formatAttr(buf *strings.Builder, attr slog.Attr) {
	key := attr.Key
	if h.group != "" {
		key = h.group + "." + key
	}

	buf.WriteString(key)
	buf.WriteString("=")

	value := attr.Value
	switch value.Kind() {
	case slog.KindString:
		str := value.String()
		// Quote strings that contain spaces
		if strings.Contains(str, " ") {
			buf.WriteString(`"`)
			buf.WriteString(str)
			buf.WriteString(`"`)
		} else {
			buf.WriteString(str)
		}
	case slog.KindTime:
		buf.WriteString(value.Time().Format(time.RFC3339))
	case slog.KindDuration:
		buf.WriteString(value.Duration().String())
	case slog.KindGroup:
		// Format group attributes inline
		attrs := value.Group()
		if len(attrs) > 0 {
			buf.WriteString("{")
			for i, a := range attrs {
				if i > 0 {
					buf.WriteString(" ")
				}
				oldGroup := h.group
				h.group = ""
				h.formatAttr(buf, a)
				h.group = oldGroup
			}
			buf.WriteString("}")
		}
	default:
		buf.WriteString(fmt.Sprint(value.Any()))
	}
}

// WithAttrs returns a new handler with the given attributes added.
func (h *HumanReadableHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &HumanReadableHandler{
		out:   h.out,
		level: h.level,
		attrs: newAttrs,
		group: h.group,
	}
}

// WithGroup returns a new handler with the given group added.
func (h *HumanReadableHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	group := name
	if h.group != "" {
		group = h.group + "." + name
	}

	return &HumanReadableHandler{
		out:   h.out,
		level: h.level,
		attrs: h.attrs,
		group: group,
	}
}

// Setup configures the global slog logger based on the provided configuration.
func Setup(w io.Writer, logLevel string, useJSON bool) *slog.Logger {
	level := parseLevel(logLevel)

	var handler slog.Handler
	if useJSON {
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = NewHumanReadableHandler(w, level)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

// parseLevel converts a string log level to slog.Level
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
