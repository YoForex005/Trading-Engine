package logging

import (
	"log/slog"
	"os"
)

// NewLogger creates a structured JSON logger for production use.
// Logs are written to stdout (container best practice) with INFO level.
// The logger outputs JSON format for log aggregation and searchability.
//
// Example usage:
//
//	logger := logging.NewLogger()
//	slog.SetDefault(logger)
//	slog.Info("server starting", "port", 8080)
//	slog.Error("database error", "error", err, "database", "trading_engine")
func NewLogger() *slog.Logger {
	// Create JSON handler with INFO level for production
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return logger
}
