package payments

import (
	"context"
	"time"
)

// Repository defines the payment data persistence interface
type Repository interface {
	// Transaction operations
	SaveTransaction(ctx context.Context, tx *Transaction) error
	UpdateTransaction(ctx context.Context, tx *Transaction) error
	GetTransaction(ctx context.Context, txID string) (*Transaction, error)
	GetUserTransactions(ctx context.Context, userID string, limit, offset int) ([]*Transaction, error)
	GetTransactionsByTimeRange(ctx context.Context, from, to time.Time) ([]*Transaction, error)
	GetProviderTransactions(ctx context.Context, provider PaymentProvider, from, to time.Time) ([]*Transaction, error)
	GetCompletedTransactions(ctx context.Context, from, to time.Time) ([]*Transaction, error)
	CountUserTransactions(ctx context.Context, userID string, from, to time.Time) (int, error)

	// User balance operations
	GetUserBalance(ctx context.Context, userID string, currency string) (float64, error)
	CreditUserBalance(ctx context.Context, userID string, amount float64, currency string, txID string) error
	DebitUserBalance(ctx context.Context, userID string, amount float64, currency string, txID string) error
	ReserveUserBalance(ctx context.Context, userID string, amount float64, currency string, txID string) error
	UnreserveUserBalance(ctx context.Context, userID string, amount float64, currency string, txID string) error
	DebitReservedBalance(ctx context.Context, userID string, amount float64, currency string, txID string) error

	// User information
	GetUserVerificationLevel(ctx context.Context, userID string) (int, error)
	GetUserCreatedAt(ctx context.Context, userID string) (time.Time, error)
	GetUserLastIP(ctx context.Context, userID string) (string, error)
	GetUserAverageTransactionAmount(ctx context.Context, userID string) (float64, error)
	GetUserTotalDeposits(ctx context.Context, userID string) (float64, error)
	GetUserTotalWithdrawals(ctx context.Context, userID string) (float64, error)
	GetUserDepositMethods(ctx context.Context, userID string) ([]PaymentMethod, error)

	// Transaction checks
	HasPendingWithdrawal(ctx context.Context, userID string) (bool, error)
	HasDepositedWithMethod(ctx context.Context, userID string, method PaymentMethod) (bool, error)

	// Device tracking
	GetDeviceFailedTransactionCount(ctx context.Context, deviceID string) (int, error)

	// Exchange rates
	GetExchangeRate(ctx context.Context, from, to string) (float64, error)
	SaveExchangeRate(ctx context.Context, from, to string, rate float64) error
}
