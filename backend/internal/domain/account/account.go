package account

import (
	"errors"
	"time"

	"github.com/govalues/decimal"
)

// Account represents a trading account with pure business logic
type Account struct {
	ID            int64
	AccountNumber string
	UserID        string
	Username      string
	Password      string
	Balance       float64
	Equity        float64
	Margin        float64
	FreeMargin    float64
	MarginLevel   float64
	Leverage      float64
	MarginMode    string // HEDGING or NETTING
	Currency      string
	Status        string // ACTIVE, DISABLED
	IsDemo        bool
	CreatedAt     int64
}

// CanOpenPosition checks if account has sufficient funds to open a position
func (a *Account) CanOpenPosition(requiredMargin float64) bool {
	return a.FreeMargin >= requiredMargin
}

// UpdateBalance adjusts account balance by the given amount
func (a *Account) UpdateBalance(amount float64) {
	a.Balance += amount
}

// UpdateEquity recalculates equity based on unrealized P&L
func (a *Account) UpdateEquity(unrealizedPnL float64) {
	a.Equity = a.Balance + unrealizedPnL
}

// UpdateMargin sets the used margin
func (a *Account) UpdateMargin(usedMargin float64) {
	a.Margin = usedMargin
	a.FreeMargin = a.Equity - usedMargin

	// Calculate margin level (avoid division by zero)
	if usedMargin > 0 {
		a.MarginLevel = (a.Equity / usedMargin) * 100
	} else {
		a.MarginLevel = 0
	}
}

// SetLeverage updates the account leverage
func (a *Account) SetLeverage(leverage float64) error {
	if leverage <= 0 {
		return errors.New("leverage must be positive")
	}
	a.Leverage = leverage
	return nil
}

// SetMarginMode updates the margin mode
func (a *Account) SetMarginMode(mode string) error {
	if mode != "HEDGING" && mode != "NETTING" {
		return errors.New("margin mode must be HEDGING or NETTING")
	}
	a.MarginMode = mode
	return nil
}

// Disable disables the account
func (a *Account) Disable() {
	a.Status = "DISABLED"
}

// Enable enables the account
func (a *Account) Enable() {
	a.Status = "ACTIVE"
}

// IsActive returns true if the account is active
func (a *Account) IsActive() bool {
	return a.Status == "ACTIVE"
}

// Summary represents computed account data for API responses
type Summary struct {
	AccountID     int64
	AccountNumber string
	Currency      string
	Balance       float64
	Equity        float64
	Margin        float64
	FreeMargin    float64
	MarginLevel   float64
	UnrealizedPnL float64
	Leverage      float64
	MarginMode    string
	OpenPositions int
}

// GetSummary creates a summary for API responses
func (a *Account) GetSummary(unrealizedPnL float64, openPositions int) *Summary {
	return &Summary{
		AccountID:     a.ID,
		AccountNumber: a.AccountNumber,
		Currency:      a.Currency,
		Balance:       a.Balance,
		Equity:        a.Equity,
		Margin:        a.Margin,
		FreeMargin:    a.FreeMargin,
		MarginLevel:   a.MarginLevel,
		UnrealizedPnL: unrealizedPnL,
		Leverage:      a.Leverage,
		MarginMode:    a.MarginMode,
		OpenPositions: openPositions,
	}
}

// WithDecimalBalance is a helper for future decimal migration
type DecimalAccount struct {
	ID            int64
	AccountNumber string
	UserID        string
	Username      string
	Balance       decimal.Decimal
	Equity        decimal.Decimal
	Margin        decimal.Decimal
	FreeMargin    decimal.Decimal
	MarginLevel   decimal.Decimal
	Leverage      float64
	MarginMode    string
	Currency      string
	Status        string
	IsDemo        bool
	CreatedAt     time.Time
}
