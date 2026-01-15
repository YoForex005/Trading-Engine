package logging

import (
	"log/slog"
	"os"
)

// Default is the global logger instance used across the application.
// It is set during Init() and can be accessed directly.
var Default *slog.Logger

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

// Init initializes the global logger with the specified log level.
// It supports configurable log levels from environment variables.
//
// Example:
//
//	logging.Init(slog.LevelInfo)  // Production
//	logging.Init(slog.LevelDebug) // Development
func Init(level slog.Level) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
		// AddSource can be enabled for debugging (shows file:line)
		AddSource: false,
	})

	Default = slog.New(handler)
	slog.SetDefault(Default)
}

// WithContext creates a logger with additional context fields.
// This is useful for request-scoped logging with trace IDs, user IDs, etc.
//
// Example:
//
//	reqLogger := logging.WithContext("request_id", reqID, "ip", clientIP)
//	reqLogger.Info("processing request")
func WithContext(fields ...any) *slog.Logger {
	if Default == nil {
		// Fallback if Init hasn't been called
		Default = NewLogger()
	}
	return Default.With(fields...)
}
