package risk

import (
	"errors"
	"sync"
)

// Account represents a trading account with risk metrics
type Account struct {
	ID         string  `json:"id"`
	UserID     string  `json:"userId"`
	Balance    float64 `json:"balance"`
	Equity     float64 `json:"equity"`
	Margin     float64 `json:"margin"`
	FreeMargin float64 `json:"freeMargin"`
	MarginLevel float64 `json:"marginLevel"` // Equity/Margin as percentage
	Leverage   int     `json:"leverage"`
	Currency   string  `json:"currency"`
}

// Engine handles risk calculations and checks
type Engine struct {
	accounts map[string]*Account
	mu       sync.RWMutex
}

func NewEngine() *Engine {
	// Initialize with a demo account
	return &Engine{
		accounts: map[string]*Account{
			"demo_001": {
				ID:         "demo_001",
				UserID:     "user_001",
				Balance:    10000.00,
				Equity:     10000.00,
				Margin:     0,
				FreeMargin: 10000.00,
				MarginLevel: 0,
				Leverage:   100,
				Currency:   "USD",
			},
		},
	}
}

// PreTradeCheck validates if an order can be placed
func (e *Engine) PreTradeCheck(accountID string, symbol string, volume float64, price float64) error {
	e.mu.RLock()
	account, ok := e.accounts[accountID]
	e.mu.RUnlock()

	if !ok {
		return errors.New("account not found")
	}

	// Calculate required margin
	contractSize := 100000.0 // Standard lot
	requiredMargin := (volume * contractSize * price) / float64(account.Leverage)

	if requiredMargin > account.FreeMargin {
		return errors.New("insufficient margin")
	}

	return nil
}

// UpdateMargin updates the margin used for an account
func (e *Engine) UpdateMargin(accountID string, marginDelta float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	account, ok := e.accounts[accountID]
	if !ok {
		return errors.New("account not found")
	}

	account.Margin += marginDelta
	account.FreeMargin = account.Equity - account.Margin
	
	if account.Margin > 0 {
		account.MarginLevel = (account.Equity / account.Margin) * 100
	} else {
		account.MarginLevel = 0
	}

	return nil
}

// UpdateEquity updates the equity based on floating PnL
func (e *Engine) UpdateEquity(accountID string, floatingPnL float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	account, ok := e.accounts[accountID]
	if !ok {
		return errors.New("account not found")
	}

	account.Equity = account.Balance + floatingPnL
	account.FreeMargin = account.Equity - account.Margin

	if account.Margin > 0 {
		account.MarginLevel = (account.Equity / account.Margin) * 100
	}

	// Check for margin call / stop out
	if account.MarginLevel > 0 && account.MarginLevel < 50 {
		// Would trigger liquidation in production
		return errors.New("margin call triggered")
	}

	return nil
}

// GetAccount returns account info
func (e *Engine) GetAccount(accountID string) (*Account, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	account, ok := e.accounts[accountID]
	if !ok {
		return nil, errors.New("account not found")
	}

	return account, nil
}

// CreateAccount creates a new trading account
func (e *Engine) CreateAccount(userID string, balance float64, leverage int) *Account {
	e.mu.Lock()
	defer e.mu.Unlock()

	account := &Account{
		ID:         "acc_" + userID,
		UserID:     userID,
		Balance:    balance,
		Equity:     balance,
		Margin:     0,
		FreeMargin: balance,
		MarginLevel: 0,
		Leverage:   leverage,
		Currency:   "USD",
	}

	e.accounts[account.ID] = account
	return account
}
