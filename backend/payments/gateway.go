// Package payments provides comprehensive payment gateway integration
// supporting deposits, withdrawals, and multi-provider reconciliation.
package payments

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// PaymentMethod represents supported payment methods
type PaymentMethod string

const (
	MethodCard         PaymentMethod = "card"
	MethodBankTransfer PaymentMethod = "bank_transfer"
	MethodACH          PaymentMethod = "ach"
	MethodSEPA         PaymentMethod = "sepa"
	MethodWire         PaymentMethod = "wire"
	MethodPayPal       PaymentMethod = "paypal"
	MethodSkrill       PaymentMethod = "skrill"
	MethodNeteller     PaymentMethod = "neteller"
	MethodBitcoin      PaymentMethod = "bitcoin"
	MethodEthereum     PaymentMethod = "ethereum"
	MethodUSDT         PaymentMethod = "usdt"
	MethodLocalPayment PaymentMethod = "local_payment"
)

// PaymentProvider represents payment gateway providers
type PaymentProvider string

const (
	ProviderStripe    PaymentProvider = "stripe"
	ProviderBraintree PaymentProvider = "braintree"
	ProviderPayPal    PaymentProvider = "paypal"
	ProviderCoinbase  PaymentProvider = "coinbase"
	ProviderBitPay    PaymentProvider = "bitpay"
	ProviderCircle    PaymentProvider = "circle"
	ProviderWise      PaymentProvider = "wise"
	ProviderInternal  PaymentProvider = "internal"
)

// TransactionStatus represents payment transaction states
type TransactionStatus string

const (
	StatusPending    TransactionStatus = "pending"
	StatusProcessing TransactionStatus = "processing"
	StatusCompleted  TransactionStatus = "completed"
	StatusFailed     TransactionStatus = "failed"
	StatusCancelled  TransactionStatus = "cancelled"
	StatusRefunded   TransactionStatus = "refunded"
	StatusDisputed   TransactionStatus = "disputed"
)

// TransactionType represents transaction direction
type TransactionType string

const (
	TypeDeposit    TransactionType = "deposit"
	TypeWithdrawal TransactionType = "withdrawal"
	TypeRefund     TransactionType = "refund"
	TypeChargeback TransactionType = "chargeback"
)

// Transaction represents a payment transaction
type Transaction struct {
	ID               string            `json:"id"`
	UserID           string            `json:"user_id"`
	Type             TransactionType   `json:"type"`
	Method           PaymentMethod     `json:"method"`
	Provider         PaymentProvider   `json:"provider"`
	Status           TransactionStatus `json:"status"`
	Amount           float64           `json:"amount"`
	Currency         string            `json:"currency"`
	Fee              float64           `json:"fee"`
	NetAmount        float64           `json:"net_amount"`
	ExchangeRate     float64           `json:"exchange_rate,omitempty"`
	ProviderTxID     string            `json:"provider_tx_id,omitempty"`
	PaymentDetails   map[string]string `json:"payment_details,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	IPAddress        string            `json:"ip_address"`
	DeviceID         string            `json:"device_id,omitempty"`
	Country          string            `json:"country"`
	FailureReason    string            `json:"failure_reason,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	CompletedAt      *time.Time        `json:"completed_at,omitempty"`
	ConfirmationsReq int               `json:"confirmations_required,omitempty"`
	ConfirmationsRcv int               `json:"confirmations_received,omitempty"`
}

// PaymentLimits represents limits for a payment method
type PaymentLimits struct {
	Method               PaymentMethod `json:"method"`
	MinAmount            float64       `json:"min_amount"`
	MaxAmount            float64       `json:"max_amount"`
	DailyLimit           float64       `json:"daily_limit"`
	WeeklyLimit          float64       `json:"weekly_limit"`
	MonthlyLimit         float64       `json:"monthly_limit"`
	RequiresVerification bool          `json:"requires_verification"`
}

// PaymentRequest represents a payment initiation request
type PaymentRequest struct {
	UserID         string            `json:"user_id"`
	Type           TransactionType   `json:"type"`
	Method         PaymentMethod     `json:"method"`
	Amount         float64           `json:"amount"`
	Currency       string            `json:"currency"`
	PaymentDetails map[string]string `json:"payment_details"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	IPAddress      string            `json:"ip_address"`
	DeviceID       string            `json:"device_id,omitempty"`
	Country        string            `json:"country,omitempty"`
	ReturnURL      string            `json:"return_url,omitempty"`
	WebhookURL     string            `json:"webhook_url,omitempty"`
}

// PaymentResponse represents a payment processing response
type PaymentResponse struct {
	TransactionID  string            `json:"transaction_id"`
	Status         TransactionStatus `json:"status"`
	RedirectURL    string            `json:"redirect_url,omitempty"`
	RequiresAction bool              `json:"requires_action"`
	ActionType     string            `json:"action_type,omitempty"`
	ActionData     map[string]string `json:"action_data,omitempty"`
	EstimatedTime  string            `json:"estimated_time,omitempty"`
	Message        string            `json:"message,omitempty"`
}

// Gateway defines the payment gateway interface
type Gateway interface {
	// Deposit operations
	ProcessDeposit(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
	VerifyDeposit(ctx context.Context, txID string) (*Transaction, error)

	// Withdrawal operations
	ProcessWithdrawal(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
	VerifyWithdrawal(ctx context.Context, txID string) (*Transaction, error)
	CancelWithdrawal(ctx context.Context, txID string) error

	// Transaction management
	GetTransaction(ctx context.Context, txID string) (*Transaction, error)
	GetUserTransactions(ctx context.Context, userID string, limit, offset int) ([]*Transaction, error)

	// Limits and validation
	GetLimits(ctx context.Context, userID string, method PaymentMethod) (*PaymentLimits, error)
	ValidateTransaction(ctx context.Context, req *PaymentRequest) error

	// Reconciliation
	ReconcileTransactions(ctx context.Context, from, to time.Time) ([]ReconciliationResult, error)

	// Webhooks
	HandleWebhook(ctx context.Context, provider PaymentProvider, payload []byte) error
}

// Provider defines the payment provider interface
type Provider interface {
	Name() PaymentProvider
	SupportedMethods() []PaymentMethod
	SupportedCurrencies() []string

	// Deposit operations
	InitiateDeposit(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
	VerifyDeposit(ctx context.Context, providerTxID string) (*Transaction, error)

	// Withdrawal operations
	InitiateWithdrawal(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
	VerifyWithdrawal(ctx context.Context, providerTxID string) (*Transaction, error)
	CancelWithdrawal(ctx context.Context, providerTxID string) error

	// Webhook handling
	ParseWebhook(ctx context.Context, payload []byte) (*WebhookEvent, error)
	VerifyWebhookSignature(ctx context.Context, payload, signature []byte) error
}

// WebhookEvent represents a webhook event from a provider
type WebhookEvent struct {
	Provider      PaymentProvider   `json:"provider"`
	EventType     string            `json:"event_type"`
	TransactionID string            `json:"transaction_id"`
	ProviderTxID  string            `json:"provider_tx_id"`
	Status        TransactionStatus `json:"status"`
	Amount        float64           `json:"amount"`
	Currency      string            `json:"currency"`
	Timestamp     time.Time         `json:"timestamp"`
	Data          map[string]any    `json:"data"`
}

// ReconciliationResult represents a reconciliation check result
type ReconciliationResult struct {
	TransactionID  string            `json:"transaction_id"`
	ProviderTxID   string            `json:"provider_tx_id"`
	OurStatus      TransactionStatus `json:"our_status"`
	ProviderStatus TransactionStatus `json:"provider_status"`
	Matched        bool              `json:"matched"`
	Discrepancy    string            `json:"discrepancy,omitempty"`
}

// FraudCheck represents fraud detection result
type FraudCheck struct {
	TransactionID string            `json:"transaction_id"`
	RiskScore     float64           `json:"risk_score"` // 0-100
	RiskLevel     string            `json:"risk_level"` // low, medium, high, critical
	Flags         []string          `json:"flags"`
	Blocked       bool              `json:"blocked"`
	Reason        string            `json:"reason,omitempty"`
	Checks        map[string]string `json:"checks"`
}

// Common errors
var (
	ErrInvalidAmount        = errors.New("invalid transaction amount")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrLimitExceeded        = errors.New("transaction limit exceeded")
	ErrMethodNotSupported   = errors.New("payment method not supported")
	ErrProviderNotAvailable = errors.New("payment provider not available")
	ErrTransactionNotFound  = errors.New("transaction not found")
	ErrTransactionFailed    = errors.New("transaction failed")
	ErrFraudDetected        = errors.New("fraud detected")
	ErrVerificationRequired = errors.New("additional verification required")
	ErrSameMethodRequired   = errors.New("same method required for withdrawal")
	ErrPendingWithdrawal    = errors.New("pending withdrawal exists")
)

// NewTransactionID generates a new transaction ID
func NewTransactionID(txType TransactionType) string {
	prefix := "DEP"
	if txType == TypeWithdrawal {
		prefix = "WTH"
	} else if txType == TypeRefund {
		prefix = "REF"
	} else if txType == TypeChargeback {
		prefix = "CHB"
	}
	return prefix + "-" + uuid.New().String()
}
