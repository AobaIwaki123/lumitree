// Package logger provides structured logging initialization using log/slog.
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// New creates a configured *slog.Logger.
func New(levelStr, formatStr string, out io.Writer) *slog.Logger {
	if out == nil {
		out = os.Stdout
	}

	var level slog.Level
	switch strings.ToLower(levelStr) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if strings.ToLower(formatStr) == "json" {
		handler = slog.NewJSONHandler(out, opts)
	} else {
		handler = slog.NewTextHandler(out, opts)
	}

	return slog.New(handler)
}

// Init sets the default global logger.
func Init(levelStr, formatStr string) *slog.Logger {
	l := New(levelStr, formatStr, os.Stdout)
	slog.SetDefault(l)
	return l
}
