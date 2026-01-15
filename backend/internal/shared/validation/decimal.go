package validation

import (
	"fmt"

	"github.com/govalues/decimal"
)

// ValidatePositive checks if decimal value is greater than zero
func ValidatePositive(value decimal.Decimal, fieldName string) error {
	zero := decimal.MustNew(0, 0)
	if value.Cmp(zero) <= 0 {
		return fmt.Errorf("%s must be positive", fieldName)
	}
	return nil
}

// ValidateNonNegative checks if decimal value is greater than or equal to zero
func ValidateNonNegative(value decimal.Decimal, fieldName string) error {
	zero := decimal.MustNew(0, 0)
	if value.Cmp(zero) < 0 {
		return fmt.Errorf("%s must be non-negative", fieldName)
	}
	return nil
}

// ValidateRange checks if decimal value is within min/max bounds (inclusive)
func ValidateRange(value, min, max decimal.Decimal, fieldName string) error {
	if value.Cmp(min) < 0 || value.Cmp(max) > 0 {
		return fmt.Errorf("%s must be between %s and %s", fieldName, min, max)
	}
	return nil
}

// ValidateMinimum checks if decimal value is greater than or equal to minimum
func ValidateMinimum(value, min decimal.Decimal, fieldName string) error {
	if value.Cmp(min) < 0 {
		return fmt.Errorf("%s must be at least %s", fieldName, min)
	}
	return nil
}

// ValidateMaximum checks if decimal value is less than or equal to maximum
func ValidateMaximum(value, max decimal.Decimal, fieldName string) error {
	if value.Cmp(max) > 0 {
		return fmt.Errorf("%s must be at most %s", fieldName, max)
	}
	return nil
}
