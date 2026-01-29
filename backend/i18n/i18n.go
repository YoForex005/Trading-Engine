package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

//go:embed locales/*/*.json
var localesFS embed.FS

// SupportedLanguage represents a supported language
type SupportedLanguage struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"nativeName"`
	RTL        bool   `json:"rtl"`
}

// SupportedLanguages lists all supported languages
var SupportedLanguages = map[string]SupportedLanguage{
	"en-US": {Code: "en-US", Name: "English", NativeName: "English", RTL: false},
	"es-ES": {Code: "es-ES", Name: "Spanish", NativeName: "Español", RTL: false},
	"fr-FR": {Code: "fr-FR", Name: "French", NativeName: "Français", RTL: false},
	"de-DE": {Code: "de-DE", Name: "German", NativeName: "Deutsch", RTL: false},
	"ja-JP": {Code: "ja-JP", Name: "Japanese", NativeName: "日本語", RTL: false},
	"zh-CN": {Code: "zh-CN", Name: "Chinese (Simplified)", NativeName: "简体中文", RTL: false},
	"ar-SA": {Code: "ar-SA", Name: "Arabic", NativeName: "العربية", RTL: true},
	"ru-RU": {Code: "ru-RU", Name: "Russian", NativeName: "Русский", RTL: false},
	"pt-BR": {Code: "pt-BR", Name: "Portuguese (Brazil)", NativeName: "Português (Brasil)", RTL: false},
	"it-IT": {Code: "it-IT", Name: "Italian", NativeName: "Italiano", RTL: false},
}

// Translator handles internationalization
type Translator struct {
	defaultLanguage language.Tag
	translations    map[string]map[string]map[string]interface{}
	printers        map[string]*message.Printer
	mu              sync.RWMutex
}

// NewTranslator creates a new translator instance
func NewTranslator(defaultLang string) (*Translator, error) {
	t := &Translator{
		defaultLanguage: language.MustParse(defaultLang),
		translations:    make(map[string]map[string]map[string]interface{}),
		printers:        make(map[string]*message.Printer),
	}

	// Load all translation files
	if err := t.loadTranslations(); err != nil {
		return nil, fmt.Errorf("failed to load translations: %w", err)
	}

	return t, nil
}

// loadTranslations loads all translation files from embedded filesystem
func (t *Translator) loadTranslations() error {
	for langCode := range SupportedLanguages {
		namespaces := []string{"common", "trading", "errors", "notifications", "legal"}

		for _, ns := range namespaces {
			path := fmt.Sprintf("locales/%s/%s.json", langCode, ns)
			data, err := localesFS.ReadFile(path)
			if err != nil {
				// Skip if file doesn't exist
				continue
			}

			var translations map[string]interface{}
			if err := json.Unmarshal(data, &translations); err != nil {
				return fmt.Errorf("failed to unmarshal %s: %w", path, err)
			}

			if t.translations[langCode] == nil {
				t.translations[langCode] = make(map[string]map[string]interface{})
			}
			t.translations[langCode][ns] = translations
		}

		// Create printer for this language
		tag := language.MustParse(langCode)
		t.printers[langCode] = message.NewPrinter(tag)
	}

	return nil
}

// T translates a key with optional parameters
func (t *Translator) T(lang, key string, params ...interface{}) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Parse key (format: namespace.path.to.key)
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return key
	}

	namespace := parts[0]
	keyPath := parts[1:]

	// Try requested language
	if translation := t.getTranslation(lang, namespace, keyPath); translation != "" {
		return t.format(lang, translation, params...)
	}

	// Fallback to default language
	defaultLang := t.defaultLanguage.String()
	if translation := t.getTranslation(defaultLang, namespace, keyPath); translation != "" {
		return t.format(defaultLang, translation, params...)
	}

	// Return key if no translation found
	return key
}

// getTranslation retrieves a translation from the translations map
func (t *Translator) getTranslation(lang, namespace string, keyPath []string) string {
	langTranslations, ok := t.translations[lang]
	if !ok {
		return ""
	}

	nsTranslations, ok := langTranslations[namespace]
	if !ok {
		return ""
	}

	// Navigate through the key path
	current := nsTranslations
	for i, key := range keyPath {
		value, ok := current[key]
		if !ok {
			return ""
		}

		// Last key should be a string
		if i == len(keyPath)-1 {
			if str, ok := value.(string); ok {
				return str
			}
			return ""
		}

		// Intermediate keys should be objects
		if obj, ok := value.(map[string]interface{}); ok {
			current = obj
		} else {
			return ""
		}
	}

	return ""
}

// format applies formatting to translation strings
func (t *Translator) format(lang, template string, params ...interface{}) string {
	if len(params) == 0 {
		return template
	}

	printer, ok := t.printers[lang]
	if !ok {
		printer = t.printers[t.defaultLanguage.String()]
	}

	return printer.Sprintf(template, params...)
}

// GetLanguageTag returns the language tag for a language code
func (t *Translator) GetLanguageTag(lang string) language.Tag {
	tag, err := language.Parse(lang)
	if err != nil {
		return t.defaultLanguage
	}
	return tag
}

// IsRTL checks if a language is right-to-left
func (t *Translator) IsRTL(lang string) bool {
	if info, ok := SupportedLanguages[lang]; ok {
		return info.RTL
	}
	return false
}

// Global translator instance
var globalTranslator *Translator
var once sync.Once

// Init initializes the global translator
func Init(defaultLang string) error {
	var err error
	once.Do(func() {
		globalTranslator, err = NewTranslator(defaultLang)
	})
	return err
}

// T is a convenience function for translating with the global translator
func T(lang, key string, params ...interface{}) string {
	if globalTranslator == nil {
		return key
	}
	return globalTranslator.T(lang, key, params...)
}

// IsRTL is a convenience function for checking RTL with the global translator
func IsRTL(lang string) bool {
	if globalTranslator == nil {
		return false
	}
	return globalTranslator.IsRTL(lang)
}
