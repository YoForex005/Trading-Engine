package logging

import (
	"regexp"
	"strings"
)

// SensitiveDataMasker masks sensitive data in logs
type SensitiveDataMasker struct {
	patterns map[string]*regexp.Regexp
}

// NewSensitiveDataMasker creates a new data masker
func NewSensitiveDataMasker() *SensitiveDataMasker {
	return &SensitiveDataMasker{
		patterns: map[string]*regexp.Regexp{
			"email":        regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
			"phone":        regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`),
			"ssn":          regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			"credit_card":  regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),
			"api_key":      regexp.MustCompile(`(?i)(api[_-]?key|apikey|access[_-]?token)[\s:="']+([a-zA-Z0-9_\-]{20,})`),
			"password":     regexp.MustCompile(`(?i)(password|passwd|pwd)[\s:="']+([^\s"']+)`),
			"bearer_token": regexp.MustCompile(`(?i)Bearer\s+([a-zA-Z0-9_\-\.]{20,})`),
			"jwt":          regexp.MustCompile(`eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*`),
		},
	}
}

// Mask masks sensitive data in a string
func (m *SensitiveDataMasker) Mask(input string) string {
	result := input

	// Mask emails
	result = m.patterns["email"].ReplaceAllStringFunc(result, func(match string) string {
		parts := strings.Split(match, "@")
		if len(parts) == 2 {
			return maskString(parts[0]) + "@" + parts[1]
		}
		return maskString(match)
	})

	// Mask phone numbers
	result = m.patterns["phone"].ReplaceAllString(result, "XXX-XXX-XXXX")

	// Mask SSN
	result = m.patterns["ssn"].ReplaceAllString(result, "XXX-XX-XXXX")

	// Mask credit cards
	result = m.patterns["credit_card"].ReplaceAllStringFunc(result, func(match string) string {
		// Keep last 4 digits
		cleaned := strings.ReplaceAll(strings.ReplaceAll(match, " ", ""), "-", "")
		if len(cleaned) >= 4 {
			return "XXXX-XXXX-XXXX-" + cleaned[len(cleaned)-4:]
		}
		return "XXXX-XXXX-XXXX-XXXX"
	})

	// Mask API keys
	result = m.patterns["api_key"].ReplaceAllString(result, "$1=[REDACTED]")

	// Mask passwords
	result = m.patterns["password"].ReplaceAllString(result, "$1=[REDACTED]")

	// Mask bearer tokens
	result = m.patterns["bearer_token"].ReplaceAllString(result, "Bearer [REDACTED]")

	// Mask JWTs
	result = m.patterns["jwt"].ReplaceAllString(result, "[JWT_REDACTED]")

	return result
}

// MaskJSON masks sensitive data in JSON strings
func (m *SensitiveDataMasker) MaskJSON(input string) string {
	// First apply standard masking
	result := m.Mask(input)

	// Additional JSON-specific patterns
	sensitiveKeys := []string{
		"password", "passwd", "pwd", "secret", "token", "api_key", "apiKey",
		"accessToken", "refreshToken", "privateKey", "private_key",
		"credit_card", "creditCard", "cvv", "ssn", "pin",
	}

	for _, key := range sensitiveKeys {
		// Match "key": "value" or 'key': 'value'
		pattern := regexp.MustCompile(`"` + key + `"\s*:\s*"[^"]*"`)
		result = pattern.ReplaceAllString(result, `"`+key+`":"[REDACTED]"`)

		pattern = regexp.MustCompile(`'` + key + `'\s*:\s*'[^']*'`)
		result = pattern.ReplaceAllString(result, `'`+key+`':'[REDACTED]'`)
	}

	return result
}

// MaskMap masks sensitive data in a map
func (m *SensitiveDataMasker) MaskMap(input map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	sensitiveKeys := map[string]bool{
		"password":      true,
		"passwd":        true,
		"pwd":           true,
		"secret":        true,
		"token":         true,
		"api_key":       true,
		"apiKey":        true,
		"apikey":        true,
		"access_token":  true,
		"accessToken":   true,
		"refresh_token": true,
		"refreshToken":  true,
		"private_key":   true,
		"privateKey":    true,
		"credit_card":   true,
		"creditCard":    true,
		"cvv":           true,
		"ssn":           true,
		"pin":           true,
	}

	for key, value := range input {
		if sensitiveKeys[key] || sensitiveKeys[strings.ToLower(key)] {
			result[key] = "[REDACTED]"
		} else {
			// Recursively mask nested maps
			if nestedMap, ok := value.(map[string]interface{}); ok {
				result[key] = m.MaskMap(nestedMap)
			} else if strValue, ok := value.(string); ok {
				result[key] = m.Mask(strValue)
			} else {
				result[key] = value
			}
		}
	}

	return result
}

// maskString masks a string keeping first and last character
func maskString(s string) string {
	if len(s) <= 2 {
		return strings.Repeat("*", len(s))
	}
	return string(s[0]) + strings.Repeat("*", len(s)-2) + string(s[len(s)-1])
}

// Global masker instance
var globalMasker = NewSensitiveDataMasker()

// MaskSensitiveData masks sensitive data using the global masker
func MaskSensitiveData(input string) string {
	return globalMasker.Mask(input)
}

// MaskSensitiveJSON masks sensitive data in JSON using the global masker
func MaskSensitiveJSON(input string) string {
	return globalMasker.MaskJSON(input)
}

// MaskSensitiveMap masks sensitive data in a map using the global masker
func MaskSensitiveMap(input map[string]interface{}) map[string]interface{} {
	return globalMasker.MaskMap(input)
}
