package payments

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Repository defines payment repository interface
type Repository interface {
	SaveTransaction(ctx context.Context, tx *Transaction) error
	UpdateTransaction(ctx context.Context, tx *Transaction) error
	GetTransaction(ctx context.Context, txID string) (*Transaction, error)
	ListByUser(userID string) ([]*Transaction, error)
	UpdateStatus(txID string, status TransactionStatus) error
	UpdateProviderTxID(txID string, providerTxID string) error
	GetUserTransactions(ctx context.Context, userID string, limit, offset int) ([]*Transaction, error)
	GetTransactionsByTimeRange(ctx context.Context, from, to time.Time) ([]*Transaction, error)
	GetProviderTransactions(ctx context.Context, provider PaymentProvider, from, to time.Time) ([]*Transaction, error)
	GetCompletedTransactions(ctx context.Context, from, to time.Time) ([]*Transaction, error)
	CountUserTransactions(ctx context.Context, userID string, from, to time.Time) (int, error)
	GetUserBalance(ctx context.Context, userID string, currency string) (float64, error)
	CreditUserBalance(ctx context.Context, userID string, amount float64, currency string, txID string) error
	DebitUserBalance(ctx context.Context, userID string, amount float64, currency string, txID string) error
	ReserveUserBalance(ctx context.Context, userID string, amount float64, currency string, txID string) error
	UnreserveUserBalance(ctx context.Context, userID string, amount float64, currency string, txID string) error
	DebitReservedBalance(ctx context.Context, userID string, amount float64, currency string, txID string) error
	GetUserVerificationLevel(ctx context.Context, userID string) (int, error)
	GetUserCreatedAt(ctx context.Context, userID string) (time.Time, error)
	GetUserLastIP(ctx context.Context, userID string) (string, error)
	GetUserAverageTransactionAmount(ctx context.Context, userID string) (float64, error)
	GetUserTotalDeposits(ctx context.Context, userID string) (float64, error)
	GetUserTotalWithdrawals(ctx context.Context, userID string) (float64, error)
	GetUserDepositMethods(ctx context.Context, userID string) ([]PaymentMethod, error)
	HasPendingWithdrawal(ctx context.Context, userID string) (bool, error)
	HasDepositedWithMethod(ctx context.Context, userID string, method PaymentMethod) (bool, error)
	GetDeviceFailedTransactionCount(ctx context.Context, deviceID string) (int, error)
	GetExchangeRate(ctx context.Context, from, to string) (float64, error)
	SaveExchangeRate(ctx context.Context, from, to string, rate float64) error
	GetUserVerificationLevelByID(userID string) (int, error)
}

// InMemoryPaymentRepository implements Repository with in-memory storage
// Thread-safe with mutex protection
type InMemoryPaymentRepository struct {
	mu           sync.RWMutex
	transactions map[string]*Transaction // txID -> Transaction
	userIndex    map[string][]string     // userID -> []txID
}

// NewInMemoryPaymentRepository creates a new in-memory repository
func NewInMemoryPaymentRepository() *InMemoryPaymentRepository {
	return &InMemoryPaymentRepository{
		transactions: make(map[string]*Transaction),
		userIndex:    make(map[string][]string),
	}
}

// SaveTransaction saves a transaction
func (r *InMemoryPaymentRepository) SaveTransaction(_ context.Context, tx *Transaction) error {
	if tx == nil {
		return errors.New("transaction cannot be nil")
	}
	if tx.ID == "" {
		return errors.New("transaction ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Store transaction
	r.transactions[tx.ID] = tx

	// Update user index
	if _, exists := r.userIndex[tx.UserID]; !exists {
		r.userIndex[tx.UserID] = []string{}
	}

	// Check if already indexed
	alreadyIndexed := false
	for _, id := range r.userIndex[tx.UserID] {
		if id == tx.ID {
			alreadyIndexed = true
			break
		}
	}

	if !alreadyIndexed {
		r.userIndex[tx.UserID] = append(r.userIndex[tx.UserID], tx.ID)
	}

	return nil
}

// UpdateTransaction updates a transaction
func (r *InMemoryPaymentRepository) UpdateTransaction(_ context.Context, tx *Transaction) error {
	if tx == nil {
		return errors.New("transaction cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.transactions[tx.ID]; !exists {
		return ErrTransactionNotFound
	}

	r.transactions[tx.ID] = tx
	return nil
}

// GetTransaction retrieves a transaction by ID
func (r *InMemoryPaymentRepository) GetTransaction(_ context.Context, txID string) (*Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tx, exists := r.transactions[txID]
	if !exists {
		return nil, ErrTransactionNotFound
	}

	// Return a copy to prevent external modification
	txCopy := *tx
	return &txCopy, nil
}

// ListByUser retrieves all transactions for a user
func (r *InMemoryPaymentRepository) ListByUser(userID string) ([]*Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	txIDs, exists := r.userIndex[userID]
	if !exists {
		return []*Transaction{}, nil // Empty slice, not error
	}

	result := make([]*Transaction, 0, len(txIDs))
	for _, txID := range txIDs {
		if tx, exists := r.transactions[txID]; exists {
			// Return copies
			txCopy := *tx
			result = append(result, &txCopy)
		}
	}

	return result, nil
}

// UpdateStatus updates transaction status
func (r *InMemoryPaymentRepository) UpdateStatus(txID string, status TransactionStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx, exists := r.transactions[txID]
	if !exists {
		return ErrTransactionNotFound
	}

	tx.Status = status
	return nil
}

// UpdateProviderTxID updates the provider transaction ID
func (r *InMemoryPaymentRepository) UpdateProviderTxID(txID string, providerTxID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx, exists := r.transactions[txID]
	if !exists {
		return ErrTransactionNotFound
	}

	tx.ProviderTxID = providerTxID
	return nil
}

// GetAllTransactions returns all transactions (for admin)
func (r *InMemoryPaymentRepository) GetAllTransactions() []*Transaction {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Transaction, 0, len(r.transactions))
	for _, tx := range r.transactions {
		txCopy := *tx
		result = append(result, &txCopy)
	}

	return result
}

// Count returns total transaction count
func (r *InMemoryPaymentRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.transactions)
}

// CountByUser returns transaction count for a user
func (r *InMemoryPaymentRepository) CountByUser(userID string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if txIDs, exists := r.userIndex[userID]; exists {
		return len(txIDs)
	}
	return 0
}

// GetUserTransactions returns user transactions with pagination
func (r *InMemoryPaymentRepository) GetUserTransactions(_ context.Context, userID string, _, _ int) ([]*Transaction, error) {
	return r.ListByUser(userID)
}

// GetTransactionsByTimeRange returns transactions in a time range
func (r *InMemoryPaymentRepository) GetTransactionsByTimeRange(_ context.Context, from, to time.Time) ([]*Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Transaction
	for _, tx := range r.transactions {
		if (tx.CreatedAt.Equal(from) || tx.CreatedAt.After(from)) && tx.CreatedAt.Before(to) {
			txCopy := *tx
			result = append(result, &txCopy)
		}
	}
	return result, nil
}

// GetProviderTransactions returns transactions for a specific provider in a time range
func (r *InMemoryPaymentRepository) GetProviderTransactions(_ context.Context, provider PaymentProvider, from, to time.Time) ([]*Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Transaction
	for _, tx := range r.transactions {
		if tx.Provider == provider && (tx.CreatedAt.Equal(from) || tx.CreatedAt.After(from)) && tx.CreatedAt.Before(to) {
			txCopy := *tx
			result = append(result, &txCopy)
		}
	}
	return result, nil
}

// GetCompletedTransactions returns completed transactions in a time range
func (r *InMemoryPaymentRepository) GetCompletedTransactions(_ context.Context, from, to time.Time) ([]*Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Transaction
	for _, tx := range r.transactions {
		if tx.Status == StatusCompleted && (tx.CreatedAt.Equal(from) || tx.CreatedAt.After(from)) && tx.CreatedAt.Before(to) {
			txCopy := *tx
			result = append(result, &txCopy)
		}
	}
	return result, nil
}

// CountUserTransactions counts transactions for a user in a time range
func (r *InMemoryPaymentRepository) CountUserTransactions(_ context.Context, userID string, from, to time.Time) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, tx := range r.transactions {
		if tx.UserID == userID && (tx.CreatedAt.Equal(from) || tx.CreatedAt.After(from)) && tx.CreatedAt.Before(to) {
			count++
		}
	}
	return count, nil
}

// GetUserBalance returns user balance (stub)
func (r *InMemoryPaymentRepository) GetUserBalance(_ context.Context, _ string, _ string) (float64, error) {
	return 0, nil
}

// CreditUserBalance credits user balance (stub)
func (r *InMemoryPaymentRepository) CreditUserBalance(_ context.Context, _, _ string, _ float64, _ string) error {
	return nil
}

// DebitUserBalance debits user balance (stub)
func (r *InMemoryPaymentRepository) DebitUserBalance(_ context.Context, _, _ string, _ float64, _ string) error {
	return nil
}

// ReserveUserBalance reserves user balance (stub)
func (r *InMemoryPaymentRepository) ReserveUserBalance(_ context.Context, _, _ string, _ float64, _ string) error {
	return nil
}

// UnreserveUserBalance unreserves user balance (stub)
func (r *InMemoryPaymentRepository) UnreserveUserBalance(_ context.Context, _, _ string, _ float64, _ string) error {
	return nil
}

// DebitReservedBalance debits reserved balance (stub)
func (r *InMemoryPaymentRepository) DebitReservedBalance(_ context.Context, _, _ string, _ float64, _ string) error {
	return nil
}

// GetUserVerificationLevel returns user verification level (stub)
func (r *InMemoryPaymentRepository) GetUserVerificationLevel(_ context.Context, _ string) (int, error) {
	return 0, nil
}

// GetUserCreatedAt returns user creation time (stub)
func (r *InMemoryPaymentRepository) GetUserCreatedAt(_ context.Context, _ string) (time.Time, error) {
	return time.Now().Add(-30 * 24 * time.Hour), nil
}

// GetUserLastIP returns user's last IP (stub)
func (r *InMemoryPaymentRepository) GetUserLastIP(_ context.Context, _ string) (string, error) {
	return "", nil
}

// GetUserAverageTransactionAmount returns average transaction amount (stub)
func (r *InMemoryPaymentRepository) GetUserAverageTransactionAmount(_ context.Context, _ string) (float64, error) {
	return 0, nil
}

// GetUserTotalDeposits returns total deposits (stub)
func (r *InMemoryPaymentRepository) GetUserTotalDeposits(_ context.Context, _ string) (float64, error) {
	return 0, nil
}

// GetUserTotalWithdrawals returns total withdrawals (stub)
func (r *InMemoryPaymentRepository) GetUserTotalWithdrawals(_ context.Context, _ string) (float64, error) {
	return 0, nil
}

// GetUserDepositMethods returns user's deposit methods (stub)
func (r *InMemoryPaymentRepository) GetUserDepositMethods(_ context.Context, _ string) ([]PaymentMethod, error) {
	return []PaymentMethod{}, nil
}

// HasPendingWithdrawal checks if user has a pending withdrawal
func (r *InMemoryPaymentRepository) HasPendingWithdrawal(_ context.Context, userID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, tx := range r.transactions {
		if tx.UserID == userID && tx.Type == TypeWithdrawal && tx.Status == StatusPending {
			return true, nil
		}
	}
	return false, nil
}

// HasDepositedWithMethod checks if user deposited with a specific method (stub)
func (r *InMemoryPaymentRepository) HasDepositedWithMethod(_ context.Context, _ string, _ PaymentMethod) (bool, error) {
	return true, nil
}

// GetDeviceFailedTransactionCount returns failed transaction count for a device (stub)
func (r *InMemoryPaymentRepository) GetDeviceFailedTransactionCount(_ context.Context, _ string) (int, error) {
	return 0, nil
}

// GetExchangeRate returns exchange rate (stub)
func (r *InMemoryPaymentRepository) GetExchangeRate(_ context.Context, _, _ string) (float64, error) {
	return 1.0, nil
}

// SaveExchangeRate saves exchange rate (stub)
func (r *InMemoryPaymentRepository) SaveExchangeRate(_ context.Context, _, _ string, _ float64) error {
	return nil
}

// GetUserVerificationLevelByID returns user verification level by ID (stub)
func (r *InMemoryPaymentRepository) GetUserVerificationLevelByID(_ string) (int, error) {
	return 0, nil
}
