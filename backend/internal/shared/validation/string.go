package validation

import (
	"fmt"
	"strings"
)

// ValidateRequired checks if string value is not empty
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateMinLength checks if string meets minimum length requirement
func ValidateMinLength(value string, minLen int, fieldName string) error {
	if len(value) < minLen {
		return fmt.Errorf("%s must be at least %d characters", fieldName, minLen)
	}
	return nil
}

// ValidateMaxLength checks if string does not exceed maximum length
func ValidateMaxLength(value string, maxLen int, fieldName string) error {
	if len(value) > maxLen {
		return fmt.Errorf("%s must be at most %d characters", fieldName, maxLen)
	}
	return nil
}

// ValidateLength checks if string is within length bounds
func ValidateLength(value string, minLen, maxLen int, fieldName string) error {
	length := len(value)
	if length < minLen || length > maxLen {
		return fmt.Errorf("%s must be between %d and %d characters", fieldName, minLen, maxLen)
	}
	return nil
}

// ValidateOneOf checks if value is in the allowed list
func ValidateOneOf(value string, allowed []string, fieldName string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of: %s", fieldName, strings.Join(allowed, ", "))
}
