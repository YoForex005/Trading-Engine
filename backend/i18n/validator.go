package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// ValidationResult represents the result of translation validation
type ValidationResult struct {
	TotalKeys       int                 `json:"totalKeys"`
	MissingKeys     map[string][]string `json:"missingKeys"`
	InvalidKeys     map[string][]string `json:"invalidKeys"`
	CoveragePercent map[string]float64  `json:"coveragePercent"`
	IsValid         bool                `json:"isValid"`
}

// Validator validates translation files
type Validator struct {
	baseLanguage string
	basePath     string
}

// NewValidator creates a new translation validator
func NewValidator(baseLanguage, basePath string) *Validator {
	return &Validator{
		baseLanguage: baseLanguage,
		basePath:     basePath,
	}
}

// Validate validates all translation files
func (v *Validator) Validate() (*ValidationResult, error) {
	result := &ValidationResult{
		MissingKeys:     make(map[string][]string),
		InvalidKeys:     make(map[string][]string),
		CoveragePercent: make(map[string]float64),
		IsValid:         true,
	}

	// Load base language keys
	baseKeys, err := v.loadLanguageKeys(v.baseLanguage)
	if err != nil {
		return nil, fmt.Errorf("failed to load base language: %w", err)
	}

	result.TotalKeys = len(baseKeys)

	// Validate each language
	for langCode := range SupportedLanguages {
		if langCode == v.baseLanguage {
			result.CoveragePercent[langCode] = 100.0
			continue
		}

		langKeys, err := v.loadLanguageKeys(langCode)
		if err != nil {
			result.InvalidKeys[langCode] = []string{fmt.Sprintf("failed to load: %v", err)}
			result.IsValid = false
			continue
		}

		// Find missing keys
		missing := v.findMissingKeys(baseKeys, langKeys)
		if len(missing) > 0 {
			result.MissingKeys[langCode] = missing
			result.IsValid = false
		}

		// Calculate coverage
		coverage := float64(len(langKeys)-len(missing)) / float64(len(baseKeys)) * 100
		result.CoveragePercent[langCode] = coverage
	}

	return result, nil
}

// loadLanguageKeys loads all keys for a language
func (v *Validator) loadLanguageKeys(lang string) (map[string]bool, error) {
	keys := make(map[string]bool)

	langPath := filepath.Join(v.basePath, lang)
	err := filepath.WalkDir(langPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}

		// Read file
		data, err := fs.ReadFile(localesFS, strings.TrimPrefix(path, v.basePath+"/"))
		if err != nil {
			return err
		}

		// Parse JSON
		var translations map[string]interface{}
		if err := json.Unmarshal(data, &translations); err != nil {
			return err
		}

		// Extract keys
		namespace := strings.TrimSuffix(d.Name(), ".json")
		v.extractKeys(namespace, "", translations, keys)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return keys, nil
}

// extractKeys recursively extracts all translation keys
func (v *Validator) extractKeys(namespace, prefix string, obj map[string]interface{}, keys map[string]bool) {
	for key, value := range obj {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		fullKeyWithNS := namespace + "." + fullKey

		switch val := value.(type) {
		case map[string]interface{}:
			v.extractKeys(namespace, fullKey, val, keys)
		case string:
			keys[fullKeyWithNS] = true
		}
	}
}

// findMissingKeys finds keys present in base but missing in target
func (v *Validator) findMissingKeys(base, target map[string]bool) []string {
	var missing []string

	for key := range base {
		if !target[key] {
			missing = append(missing, key)
		}
	}

	return missing
}

// GenerateMissingTranslations generates a template for missing translations
func (v *Validator) GenerateMissingTranslations(lang string) (map[string]interface{}, error) {
	result := &ValidationResult{
		MissingKeys: make(map[string][]string),
	}

	baseKeys, err := v.loadLanguageKeys(v.baseLanguage)
	if err != nil {
		return nil, err
	}

	langKeys, err := v.loadLanguageKeys(lang)
	if err != nil {
		langKeys = make(map[string]bool)
	}

	missing := v.findMissingKeys(baseKeys, langKeys)
	result.MissingKeys[lang] = missing

	// Organize by namespace
	byNamespace := make(map[string][]string)
	for _, key := range missing {
		parts := strings.SplitN(key, ".", 2)
		if len(parts) == 2 {
			namespace := parts[0]
			byNamespace[namespace] = append(byNamespace[namespace], parts[1])
		}
	}

	return map[string]interface{}{
		"language": lang,
		"missing":  byNamespace,
	}, nil
}
