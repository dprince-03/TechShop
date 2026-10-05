// Package logging builds structured JSON loggers (log/slog).
package logging

import (
	"log/slog"
	"os"
	"strings"
)

// New returns a JSON logger on stdout at the given level (debug, info, warn, error; default info).
func New(level string) *slog.Logger {
	var l slog.Level
	switch strings.ToLower(level) {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: l}))
}
