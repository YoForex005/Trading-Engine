package testutil

import (
	"testing"

	"github.com/govalues/decimal"
)

// AssertDecimalEqual verifies two decimal values are equal
func AssertDecimalEqual(t *testing.T, got, want decimal.Decimal, msg string) {
	t.Helper()
	if !got.Equal(want) {
		t.Errorf("%s: got %s, want %s", msg, got, want)
	}
}

// MustParseDecimal parses a decimal string or fails the test
func MustParseDecimal(t *testing.T, s string) decimal.Decimal {
	t.Helper()
	d, err := decimal.Parse(s)
	if err != nil {
		t.Fatalf("failed to parse decimal %q: %v", s, err)
	}
	return d
}

// NewTestAccount creates a test account with default values
func NewTestAccount(accountNumber string, balance string) map[string]interface{} {
	return map[string]interface{}{
		"account_number": accountNumber,
		"balance":        balance,
		"equity":         balance,
		"margin":         "0",
		"free_margin":    balance,
		"margin_level":   "0",
		"currency":       "USD",
		"leverage":       100,
	}
}

// NewTestPosition creates a test position with default values
func NewTestPosition(symbol string, lots float64, entryPrice string) map[string]interface{} {
	return map[string]interface{}{
		"symbol":      symbol,
		"volume":      lots,
		"entry_price": entryPrice,
		"side":        "buy",
		"profit":      "0",
		"swap":        "0",
	}
}

// NewTestQuote creates a test market quote
func NewTestQuote(symbol string, bid, ask string) map[string]interface{} {
	return map[string]interface{}{
		"symbol": symbol,
		"bid":    bid,
		"ask":    ask,
	}
}
