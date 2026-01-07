package core

import (
	"errors"
	"log"
	"sync"
	"time"
)

// LedgerEntry represents a transaction in the ledger
type LedgerEntry struct {
	ID            int64     `json:"id"`
	AccountID     int64     `json:"accountId"`
	Type          string    `json:"type"` // DEPOSIT/WITHDRAW/REALIZED_PNL/COMMISSION/SWAP/ADJUSTMENT/BONUS
	Amount        float64   `json:"amount"`
	BalanceAfter  float64   `json:"balanceAfter"`
	Currency      string    `json:"currency"`
	Description   string    `json:"description"`
	RefType       string    `json:"refType,omitempty"` // TRADE/POSITION/ADMIN/SYSTEM
	RefID         int64     `json:"refId,omitempty"`
	AdminID       string    `json:"adminId,omitempty"`
	PaymentMethod string    `json:"paymentMethod,omitempty"` // BANK/CRYPTO/CARD/MANUAL/BONUS
	PaymentRef    string    `json:"paymentRef,omitempty"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
}

// Ledger manages account transactions
type Ledger struct {
	mu       sync.RWMutex
	entries  map[int64][]LedgerEntry // accountID -> entries
	nextID   int64
	balances map[int64]float64 // accountID -> balance cache
}

// NewLedger creates a new ledger
func NewLedger() *Ledger {
	return &Ledger{
		entries:  make(map[int64][]LedgerEntry),
		balances: make(map[int64]float64),
		nextID:   1,
	}
}

// Deposit adds funds to an account
func (l *Ledger) Deposit(accountID int64, amount float64, method, ref, description, adminID string) (*LedgerEntry, error) {
	if amount <= 0 {
		return nil, errors.New("deposit amount must be positive")
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	currentBalance := l.balances[accountID]
	newBalance := currentBalance + amount
	l.balances[accountID] = newBalance

	entry := LedgerEntry{
		ID:            l.nextID,
		AccountID:     accountID,
		Type:          "DEPOSIT",
		Amount:        amount,
		BalanceAfter:  newBalance,
		Currency:      "USD",
		Description:   description,
		RefType:       "ADMIN",
		AdminID:       adminID,
		PaymentMethod: method,
		PaymentRef:    ref,
		Status:        "COMPLETED",
		CreatedAt:     time.Now(),
	}
	l.nextID++

	l.entries[accountID] = append(l.entries[accountID], entry)

	log.Printf("[Ledger] DEPOSIT: Account #%d +%.2f via %s | Balance: %.2f", accountID, amount, method, newBalance)
	return &entry, nil
}

// Withdraw removes funds from an account
func (l *Ledger) Withdraw(accountID int64, amount float64, method, ref, description, adminID string) (*LedgerEntry, error) {
	if amount <= 0 {
		return nil, errors.New("withdrawal amount must be positive")
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	currentBalance := l.balances[accountID]
	if currentBalance < amount {
		return nil, errors.New("insufficient balance for withdrawal")
	}

	newBalance := currentBalance - amount
	l.balances[accountID] = newBalance

	entry := LedgerEntry{
		ID:            l.nextID,
		AccountID:     accountID,
		Type:          "WITHDRAW",
		Amount:        -amount, // Negative for debit
		BalanceAfter:  newBalance,
		Currency:      "USD",
		Description:   description,
		RefType:       "ADMIN",
		AdminID:       adminID,
		PaymentMethod: method,
		PaymentRef:    ref,
		Status:        "COMPLETED",
		CreatedAt:     time.Now(),
	}
	l.nextID++

	l.entries[accountID] = append(l.entries[accountID], entry)

	log.Printf("[Ledger] WITHDRAW: Account #%d -%.2f via %s | Balance: %.2f", accountID, amount, method, newBalance)
	return &entry, nil
}

// Adjust makes a manual balance adjustment
func (l *Ledger) Adjust(accountID int64, amount float64, description, adminID string) (*LedgerEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	currentBalance := l.balances[accountID]
	newBalance := currentBalance + amount

	if newBalance < 0 {
		return nil, errors.New("adjustment would result in negative balance")
	}

	l.balances[accountID] = newBalance

	entry := LedgerEntry{
		ID:            l.nextID,
		AccountID:     accountID,
		Type:          "ADJUSTMENT",
		Amount:        amount,
		BalanceAfter:  newBalance,
		Currency:      "USD",
		Description:   description,
		RefType:       "ADMIN",
		AdminID:       adminID,
		PaymentMethod: "MANUAL",
		Status:        "COMPLETED",
		CreatedAt:     time.Now(),
	}
	l.nextID++

	l.entries[accountID] = append(l.entries[accountID], entry)

	log.Printf("[Ledger] ADJUSTMENT: Account #%d %+.2f | Balance: %.2f", accountID, amount, newBalance)
	return &entry, nil
}

// RecordRealizedPnL records realized profit/loss from a closed trade
func (l *Ledger) RecordRealizedPnL(accountID int64, amount float64, tradeID int64) *LedgerEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	currentBalance := l.balances[accountID]
	newBalance := currentBalance + amount
	l.balances[accountID] = newBalance

	description := "Trading P/L"
	if amount >= 0 {
		description = "Trading Profit"
	} else {
		description = "Trading Loss"
	}

	entry := LedgerEntry{
		ID:           l.nextID,
		AccountID:    accountID,
		Type:         "REALIZED_PNL",
		Amount:       amount,
		BalanceAfter: newBalance,
		Currency:     "USD",
		Description:  description,
		RefType:      "TRADE",
		RefID:        tradeID,
		Status:       "COMPLETED",
		CreatedAt:    time.Now(),
	}
	l.nextID++

	l.entries[accountID] = append(l.entries[accountID], entry)
	return &entry
}

// RecordCommission records commission deduction
func (l *Ledger) RecordCommission(accountID int64, amount float64, tradeID int64) *LedgerEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	currentBalance := l.balances[accountID]
	newBalance := currentBalance + amount // amount is negative
	l.balances[accountID] = newBalance

	entry := LedgerEntry{
		ID:           l.nextID,
		AccountID:    accountID,
		Type:         "COMMISSION",
		Amount:       amount,
		BalanceAfter: newBalance,
		Currency:     "USD",
		Description:  "Trading Commission",
		RefType:      "TRADE",
		RefID:        tradeID,
		Status:       "COMPLETED",
		CreatedAt:    time.Now(),
	}
	l.nextID++

	l.entries[accountID] = append(l.entries[accountID], entry)
	return &entry
}

// RecordSwap records overnight swap
func (l *Ledger) RecordSwap(accountID int64, amount float64, positionID int64) *LedgerEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	currentBalance := l.balances[accountID]
	newBalance := currentBalance + amount
	l.balances[accountID] = newBalance

	entry := LedgerEntry{
		ID:           l.nextID,
		AccountID:    accountID,
		Type:         "SWAP",
		Amount:       amount,
		BalanceAfter: newBalance,
		Currency:     "USD",
		Description:  "Overnight Swap",
		RefType:      "POSITION",
		RefID:        positionID,
		Status:       "COMPLETED",
		CreatedAt:    time.Now(),
	}
	l.nextID++

	l.entries[accountID] = append(l.entries[accountID], entry)
	return &entry
}

// AddBonus adds a bonus to account
func (l *Ledger) AddBonus(accountID int64, amount float64, description, adminID string) (*LedgerEntry, error) {
	if amount <= 0 {
		return nil, errors.New("bonus amount must be positive")
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	currentBalance := l.balances[accountID]
	newBalance := currentBalance + amount
	l.balances[accountID] = newBalance

	entry := LedgerEntry{
		ID:            l.nextID,
		AccountID:     accountID,
		Type:          "BONUS",
		Amount:        amount,
		BalanceAfter:  newBalance,
		Currency:      "USD",
		Description:   description,
		RefType:       "ADMIN",
		AdminID:       adminID,
		PaymentMethod: "BONUS",
		Status:        "COMPLETED",
		CreatedAt:     time.Now(),
	}
	l.nextID++

	l.entries[accountID] = append(l.entries[accountID], entry)

	log.Printf("[Ledger] BONUS: Account #%d +%.2f | Balance: %.2f", accountID, amount, newBalance)
	return &entry, nil
}

// GetHistory returns ledger history for an account
func (l *Ledger) GetHistory(accountID int64, limit int) []LedgerEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	entries := l.entries[accountID]
	if limit <= 0 || limit > len(entries) {
		limit = len(entries)
	}

	// Return most recent first
	result := make([]LedgerEntry, limit)
	for i := 0; i < limit; i++ {
		result[i] = entries[len(entries)-1-i]
	}
	return result
}

// GetBalance returns the cached balance
func (l *Ledger) GetBalance(accountID int64) float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.balances[accountID]
}

// SetBalance sets the balance (for initialization)
func (l *Ledger) SetBalance(accountID int64, balance float64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.balances[accountID] = balance
}

// GetAllEntries returns all ledger entries (for admin)
func (l *Ledger) GetAllEntries(limit int) []LedgerEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var all []LedgerEntry
	for _, entries := range l.entries {
		all = append(all, entries...)
	}

	// Sort by created_at descending (newest first)
	// Simple bubble sort for now
	for i := 0; i < len(all)-1; i++ {
		for j := i + 1; j < len(all); j++ {
			if all[i].CreatedAt.Before(all[j].CreatedAt) {
				all[i], all[j] = all[j], all[i]
			}
		}
	}

	if limit > 0 && limit < len(all) {
		return all[:limit]
	}
	return all
}

// GetEntriesByType returns entries filtered by type
func (l *Ledger) GetEntriesByType(entryType string, limit int) []LedgerEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var filtered []LedgerEntry
	for _, entries := range l.entries {
		for _, e := range entries {
			if e.Type == entryType {
				filtered = append(filtered, e)
			}
		}
	}

	if limit > 0 && limit < len(filtered) {
		return filtered[:limit]
	}
	return filtered
}
