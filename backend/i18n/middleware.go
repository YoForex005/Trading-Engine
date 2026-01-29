package i18n

import (
	"net/http"
	"strings"

	"golang.org/x/text/language"
)

// LanguageDetector detects user's preferred language
type LanguageDetector struct {
	defaultLang string
	supported   []language.Tag
	matcher     language.Matcher
}

// NewLanguageDetector creates a new language detector
func NewLanguageDetector(defaultLang string) *LanguageDetector {
	supported := make([]language.Tag, 0, len(SupportedLanguages))
	for code := range SupportedLanguages {
		supported = append(supported, language.MustParse(code))
	}

	return &LanguageDetector{
		defaultLang: defaultLang,
		supported:   supported,
		matcher:     language.NewMatcher(supported),
	}
}

// DetectLanguage detects language from HTTP request
func (ld *LanguageDetector) DetectLanguage(r *http.Request) string {
	// 1. Check query parameter
	if lang := r.URL.Query().Get("lang"); lang != "" {
		if ld.isSupported(lang) {
			return lang
		}
	}

	// 2. Check cookie
	if cookie, err := r.Cookie("language"); err == nil {
		if ld.isSupported(cookie.Value) {
			return cookie.Value
		}
	}

	// 3. Check Accept-Language header
	if acceptLang := r.Header.Get("Accept-Language"); acceptLang != "" {
		tags, _, _ := language.ParseAcceptLanguage(acceptLang)
		if len(tags) > 0 {
			_, index, _ := ld.matcher.Match(tags...)
			return ld.supported[index].String()
		}
	}

	// 4. Fallback to default
	return ld.defaultLang
}

// isSupported checks if a language is supported
func (ld *LanguageDetector) isSupported(lang string) bool {
	_, ok := SupportedLanguages[lang]
	return ok
}

// Middleware returns an HTTP middleware for language detection
func (ld *LanguageDetector) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := ld.DetectLanguage(r)

		// Add language to request context
		ctx := r.Context()
		ctx = WithLanguage(ctx, lang)
		r = r.WithContext(ctx)

		// Set Content-Language header
		w.Header().Set("Content-Language", lang)

		next.ServeHTTP(w, r)
	})
}

// GetLanguageFromPath extracts language from URL path
func GetLanguageFromPath(path string) (string, string) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) > 0 {
		lang := parts[0]
		if _, ok := SupportedLanguages[lang]; ok {
			remainingPath := "/" + strings.Join(parts[1:], "/")
			return lang, remainingPath
		}
	}
	return "", path
}
