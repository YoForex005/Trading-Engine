package decimal

import (
	"fmt"
	"testing"

	"github.com/govalues/decimal"
)

// AssertDecimalEqual asserts two decimals are equal.
// Uses decimal.Cmp for exact comparison (not string comparison which could have formatting differences).
func AssertDecimalEqual(t *testing.T, expected, actual decimal.Decimal, msgAndArgs ...interface{}) {
	t.Helper()
	if expected.Cmp(actual) != 0 {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("Expected %s, got %s%s", expected.String(), actual.String(), msg)
	}
}

// AssertDecimalNear asserts two decimals are equal within epsilon (for calculations where rounding acceptable).
// Example: AssertDecimalNear(t, MustParse("10.00"), calculated, MustParse("0.01")) allows ±0.01 variance.
func AssertDecimalNear(t *testing.T, expected, actual, epsilon decimal.Decimal, msgAndArgs ...interface{}) {
	t.Helper()
	diff, err := expected.Sub(actual)
	if err != nil {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("Error calculating difference: %v%s", err, msg)
		return
	}
	if diff.Cmp(Zero()) < 0 {
		diff = diff.Neg() // absolute value
	}
	if diff.Cmp(epsilon) > 0 {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("Expected %s ±%s, got %s (diff: %s)%s",
			expected.String(), epsilon.String(), actual.String(), diff.String(), msg)
	}
}

// AssertDecimalZero asserts decimal equals zero.
func AssertDecimalZero(t *testing.T, actual decimal.Decimal, msgAndArgs ...interface{}) {
	t.Helper()
	AssertDecimalEqual(t, Zero(), actual, msgAndArgs...)
}

// AssertDecimalPositive asserts decimal > 0.
func AssertDecimalPositive(t *testing.T, actual decimal.Decimal, msgAndArgs ...interface{}) {
	t.Helper()
	if !IsPositive(actual) {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("Expected positive decimal, got %s%s", actual.String(), msg)
	}
}

// AssertDecimalNegative asserts decimal < 0.
func AssertDecimalNegative(t *testing.T, actual decimal.Decimal, msgAndArgs ...interface{}) {
	t.Helper()
	if !IsNegative(actual) {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("Expected negative decimal, got %s%s", actual.String(), msg)
	}
}

func formatMessage(msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 {
		return ""
	}
	if len(msgAndArgs) == 1 {
		return ": " + msgAndArgs[0].(string)
	}
	format := msgAndArgs[0].(string)
	args := msgAndArgs[1:]
	return ": " + fmt.Sprintf(format, args...)
}
