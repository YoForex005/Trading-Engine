package i18n

import "context"

type contextKey string

const languageKey contextKey = "language"

// WithLanguage adds language to context
func WithLanguage(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, languageKey, lang)
}

// FromContext retrieves language from context
func FromContext(ctx context.Context) string {
	if lang, ok := ctx.Value(languageKey).(string); ok {
		return lang
	}
	return "en-US" // Default
}

// TranslateContext translates a key using language from context
func TranslateContext(ctx context.Context, key string, params ...interface{}) string {
	lang := FromContext(ctx)
	return T(lang, key, params...)
}

// FormatContext creates a formatter using language from context
func FormatContext(ctx context.Context) *Formatter {
	lang := FromContext(ctx)
	return NewFormatter(lang)
}
