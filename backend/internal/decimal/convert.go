package decimal

import (
	"fmt"

	"github.com/govalues/decimal"
)

// MustParse parses string to decimal, panics on error.
// Use for constant values and test data where failure is programming error.
func MustParse(s string) decimal.Decimal {
	d, err := decimal.Parse(s)
	if err != nil {
		panic("decimal.MustParse: invalid decimal string: " + s)
	}
	return d
}

// Parse parses string to decimal, returns error.
// Use for user input, database values where failure is possible.
func Parse(s string) (decimal.Decimal, error) {
	return decimal.Parse(s)
}

// ToString converts decimal to string with full precision.
func ToString(d decimal.Decimal) string {
	return d.String()
}

// ToStringFixed converts decimal to string with fixed decimal places.
// Common for display: ToStringFixed(value, 2) for currency display.
func ToStringFixed(d decimal.Decimal, scale int) string {
	return d.Round(scale).String()
}

// Zero returns decimal zero value.
func Zero() decimal.Decimal {
	return decimal.MustParse("0")
}

// NewFromInt64 creates decimal from int64.
func NewFromInt64(i int64) decimal.Decimal {
	return decimal.MustParse(fmt.Sprintf("%d", i))
}

// NewFromFloat64 creates decimal from float64.
// WARNING: Use ONLY for migration from existing float64 data.
// Prefer string input for new data to avoid float precision issues.
func NewFromFloat64(f float64) decimal.Decimal {
	// Format with enough precision to capture float value
	return decimal.MustParse(fmt.Sprintf("%.15g", f))
}

// Min returns smaller of two decimals.
func Min(a, b decimal.Decimal) decimal.Decimal {
	if a.Cmp(b) <= 0 {
		return a
	}
	return b
}

// Max returns larger of two decimals.
func Max(a, b decimal.Decimal) decimal.Decimal {
	if a.Cmp(b) >= 0 {
		return a
	}
	return b
}

// IsZero returns true if decimal equals zero.
func IsZero(d decimal.Decimal) bool {
	return d.Cmp(Zero()) == 0
}

// IsPositive returns true if decimal > 0.
func IsPositive(d decimal.Decimal) bool {
	return d.Cmp(Zero()) > 0
}

// IsNegative returns true if decimal < 0.
func IsNegative(d decimal.Decimal) bool {
	return d.Cmp(Zero()) < 0
}

// Add performs decimal addition and panics on overflow.
// For financial calculations where overflow is a programming error.
func Add(a, b decimal.Decimal) decimal.Decimal {
	result, err := a.Add(b)
	if err != nil {
		panic(fmt.Sprintf("decimal.Add overflow: %v + %v", a, b))
	}
	return result
}

// Sub performs decimal subtraction and panics on overflow.
// For financial calculations where overflow is a programming error.
func Sub(a, b decimal.Decimal) decimal.Decimal {
	result, err := a.Sub(b)
	if err != nil {
		panic(fmt.Sprintf("decimal.Sub overflow: %v - %v", a, b))
	}
	return result
}

// Mul performs decimal multiplication and panics on overflow.
// For financial calculations where overflow is a programming error.
func Mul(a, b decimal.Decimal) decimal.Decimal {
	result, err := a.Mul(b)
	if err != nil {
		panic(fmt.Sprintf("decimal.Mul overflow: %v * %v", a, b))
	}
	return result
}

// Div performs decimal division and panics on overflow or division by zero.
// For financial calculations where these are programming errors.
func Div(a, b decimal.Decimal) decimal.Decimal {
	result, err := a.Quo(b)
	if err != nil {
		panic(fmt.Sprintf("decimal.Div error: %v / %v: %v", a, b, err))
	}
	return result
}
