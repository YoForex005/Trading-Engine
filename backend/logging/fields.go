package logging

import "context"

// Field represents a log field that can be added to a log entry
type Field interface {
	Apply(entry *LogEntry)
}

// fieldFunc wraps a function as a Field
type fieldFunc func(*LogEntry)

func (f fieldFunc) Apply(entry *LogEntry) {
	f(entry)
}

// Common field constructors

func RequestID(id string) Field {
	return fieldFunc(func(e *LogEntry) {
		e.RequestID = id
	})
}

func UserID(id string) Field {
	return fieldFunc(func(e *LogEntry) {
		e.UserID = id
	})
}

func AccountID(id string) Field {
	return fieldFunc(func(e *LogEntry) {
		e.AccountID = id
	})
}

func TradeID(id string) Field {
	return fieldFunc(func(e *LogEntry) {
		e.TradeID = id
	})
}

func OrderID(id string) Field {
	return fieldFunc(func(e *LogEntry) {
		e.OrderID = id
	})
}

func Symbol(symbol string) Field {
	return fieldFunc(func(e *LogEntry) {
		e.Symbol = symbol
	})
}

func Component(component string) Field {
	return fieldFunc(func(e *LogEntry) {
		e.Component = component
	})
}

func Duration(ms float64) Field {
	return fieldFunc(func(e *LogEntry) {
		e.Duration = ms
	})
}

func String(key, value string) Field {
	return fieldFunc(func(e *LogEntry) {
		if e.Extra == nil {
			e.Extra = make(map[string]interface{})
		}
		e.Extra[key] = value
	})
}

func Int(key string, value int) Field {
	return fieldFunc(func(e *LogEntry) {
		if e.Extra == nil {
			e.Extra = make(map[string]interface{})
		}
		e.Extra[key] = value
	})
}

func Int64(key string, value int64) Field {
	return fieldFunc(func(e *LogEntry) {
		if e.Extra == nil {
			e.Extra = make(map[string]interface{})
		}
		e.Extra[key] = value
	})
}

func Float64(key string, value float64) Field {
	return fieldFunc(func(e *LogEntry) {
		if e.Extra == nil {
			e.Extra = make(map[string]interface{})
		}
		e.Extra[key] = value
	})
}

func Bool(key string, value bool) Field {
	return fieldFunc(func(e *LogEntry) {
		if e.Extra == nil {
			e.Extra = make(map[string]interface{})
		}
		e.Extra[key] = value
	})
}

func Any(key string, value interface{}) Field {
	return fieldFunc(func(e *LogEntry) {
		if e.Extra == nil {
			e.Extra = make(map[string]interface{})
		}
		e.Extra[key] = value
	})
}

// Context keys for storing values in context
type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
	accountIDKey contextKey = "account_id"
)

// Context helpers

func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func ContextWithAccountID(ctx context.Context, accountID string) context.Context {
	return context.WithValue(ctx, accountIDKey, accountID)
}

func FieldsFromContext(ctx context.Context) []Field {
	var fields []Field

	if requestID, ok := ctx.Value(requestIDKey).(string); ok && requestID != "" {
		fields = append(fields, RequestID(requestID))
	}

	if userID, ok := ctx.Value(userIDKey).(string); ok && userID != "" {
		fields = append(fields, UserID(userID))
	}

	if accountID, ok := ctx.Value(accountIDKey).(string); ok && accountID != "" {
		fields = append(fields, AccountID(accountID))
	}

	return fields
}
