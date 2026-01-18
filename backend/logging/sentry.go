package logging

import (
	"time"

	"github.com/getsentry/sentry-go"
)

// SentryHook integrates Sentry for error tracking and alerting
type SentryHook struct {
	levels      []LogLevel
	environment string
}

// NewSentryHook creates a new Sentry hook
func NewSentryHook(dsn, environment string) (*SentryHook, error) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      environment,
		TracesSampleRate: 1.0,
		AttachStacktrace: true,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Mask sensitive data before sending
			return maskSentryEvent(event)
		},
	})
	if err != nil {
		return nil, err
	}

	return &SentryHook{
		levels:      []LogLevel{ERROR, FATAL},
		environment: environment,
	}, nil
}

// Fire sends the log entry to Sentry
func (h *SentryHook) Fire(entry *LogEntry) error {
	// Set Sentry context
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(mapLogLevelToSentry(entry.Level))
		scope.SetTag("component", entry.Component)
		scope.SetTag("environment", entry.Environment)

		if entry.RequestID != "" {
			scope.SetTag("request_id", entry.RequestID)
		}

		if entry.UserID != "" {
			scope.SetUser(sentry.User{
				ID:        entry.UserID,
				IPAddress: "",
			})
		}

		if entry.AccountID != "" {
			scope.SetTag("account_id", entry.AccountID)
		}

		if entry.Symbol != "" {
			scope.SetTag("symbol", entry.Symbol)
		}

		if entry.OrderID != "" {
			scope.SetTag("order_id", entry.OrderID)
		}

		if entry.TradeID != "" {
			scope.SetTag("trade_id", entry.TradeID)
		}

		// Add extra context
		for key, value := range entry.Extra {
			scope.SetExtra(key, value)
		}
	})

	// Capture the error
	if entry.Error != "" {
		event := sentry.NewEvent()
		event.Message = entry.Message
		event.Level = mapLogLevelToSentry(entry.Level)
		event.Timestamp = entry.Timestamp

		// Add exception
		event.Exception = []sentry.Exception{
			{
				Type:       "Error",
				Value:      entry.Error,
				Stacktrace: parseStackTrace(entry.StackTrace),
			},
		}

		sentry.CaptureEvent(event)
	} else {
		// Capture as message
		sentry.CaptureMessage(entry.Message)
	}

	// Flush to ensure delivery for FATAL errors
	if entry.Level == "FATAL" {
		sentry.Flush(2 * time.Second)
	}

	return nil
}

// Levels returns the log levels this hook handles
func (h *SentryHook) Levels() []LogLevel {
	return h.levels
}

// mapLogLevelToSentry converts our log level to Sentry level
func mapLogLevelToSentry(level string) sentry.Level {
	switch level {
	case "DEBUG":
		return sentry.LevelDebug
	case "INFO":
		return sentry.LevelInfo
	case "WARN":
		return sentry.LevelWarning
	case "ERROR":
		return sentry.LevelError
	case "FATAL":
		return sentry.LevelFatal
	default:
		return sentry.LevelInfo
	}
}

// parseStackTrace converts string stack trace to Sentry format
func parseStackTrace(stackTrace string) *sentry.Stacktrace {
	if stackTrace == "" {
		return nil
	}

	// Simple stack trace parsing
	// In production, use a more robust parser
	return &sentry.Stacktrace{
		Frames: []sentry.Frame{
			{
				Filename: "unknown",
				Function: "unknown",
				Lineno:   0,
			},
		},
	}
}

// maskSentryEvent removes sensitive data from Sentry events
func maskSentryEvent(event *sentry.Event) *sentry.Event {
	// Remove sensitive keys from extra data
	sensitiveKeys := []string{"password", "api_key", "token", "secret", "credit_card"}

	for key := range event.Extra {
		for _, sensitive := range sensitiveKeys {
			if containsIgnoreCase(key, sensitive) {
				event.Extra[key] = "[REDACTED]"
			}
		}
	}

	return event
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return contains(s, substr)
}

func toLower(s string) string {
	// Simple lowercase conversion
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && indexString(s, substr) >= 0)
}

func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
