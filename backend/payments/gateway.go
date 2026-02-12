package payments

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PaymentMethod represents supported payment methods
type PaymentMethod string

const (
	MethodCard         PaymentMethod = "CARD"
	MethodPayPal       PaymentMethod = "PAYPAL"
	MethodBankTransfer PaymentMethod = "BANK_TRANSFER"
	MethodACH          PaymentMethod = "ACH"
	MethodSEPA         PaymentMethod = "SEPA"
	MethodWire         PaymentMethod = "WIRE"
	MethodBitcoin      PaymentMethod = "BITCOIN"
	MethodEthereum     PaymentMethod = "ETHEREUM"
	MethodUSDT         PaymentMethod = "USDT"
	MethodSkrill       PaymentMethod = "SKRILL"
	MethodNeteller     PaymentMethod = "NETELLER"
)

// PaymentProvider represents a payment provider identifier
type PaymentProvider = string

const (
	ProviderStripe PaymentProvider = "stripe"
	ProviderCrypto PaymentProvider = "crypto"
	ProviderWise   PaymentProvider = "wise"
)

// TransactionType represents deposit or withdrawal
type TransactionType string

const (
	TypeDeposit    TransactionType = "DEPOSIT"
	TypeWithdrawal TransactionType = "WITHDRAWAL"
	TypeRefund     TransactionType = "REFUND"
	TypeChargeback TransactionType = "CHARGEBACK"
)

// TransactionStatus represents transaction lifecycle
type TransactionStatus string

const (
	StatusPending    TransactionStatus = "PENDING"
	StatusProcessing TransactionStatus = "PROCESSING"
	StatusCompleted  TransactionStatus = "COMPLETED"
	StatusFailed     TransactionStatus = "FAILED"
	StatusCancelled  TransactionStatus = "CANCELLED"
	StatusDisputed   TransactionStatus = "DISPUTED"
)

// Errors
var (
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrUnsupportedMethod    = errors.New("unsupported payment method")
	ErrProviderNotAvailable = errors.New("provider not available")
	ErrTransactionNotFound  = errors.New("transaction not found")
	ErrHighRiskTransaction  = errors.New("high risk transaction blocked")
	ErrFraudDetected        = errors.New("fraud detected")
	ErrMethodNotSupported   = errors.New("payment method not supported")
	ErrSameMethodRequired   = errors.New("withdrawal method must match a previous deposit method")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrPendingWithdrawal    = errors.New("pending withdrawal already exists")
	ErrLimitExceeded        = errors.New("transaction limit exceeded")
)

// PaymentGateway defines the payment gateway interface
type PaymentGateway interface {
	InitiateDeposit(userID string, amount float64, currency string, method PaymentMethod) (*PaymentSession, error)
	InitiateWithdrawal(userID string, amount float64, currency string, method PaymentMethod, details *WithdrawalDetails) (*Transaction, error)
	HandleWebhook(provider string, payload []byte, signature string) error
	GetTransactionStatus(txID string) (*Transaction, error)
	GetUserTransactions(userID string) ([]*Transaction, error)
}

// PaymentSession represents a payment session (for redirects)
type PaymentSession struct {
	ID          string    `json:"id"`
	ProviderURL string    `json:"providerUrl"` // Redirect URL for user
	ExpiresAt   time.Time `json:"expiresAt"`
	Provider    string    `json:"provider"`
	Method      string    `json:"method"`
}

// Transaction represents a payment transaction
type Transaction struct {
	ID               string            `json:"id"`
	UserID           string            `json:"userId"`
	Type             TransactionType   `json:"type"`
	Method           PaymentMethod     `json:"method"`
	Provider         PaymentProvider   `json:"provider"`
	ProviderTxID     string            `json:"providerTxId,omitempty"`
	Amount           float64           `json:"amount"`
	Currency         string            `json:"currency"`
	Fee              float64           `json:"fee"`
	NetAmount        float64           `json:"netAmount"`
	Status           TransactionStatus `json:"status"`
	PaymentDetails   map[string]string `json:"paymentDetails,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	FailureReason    string            `json:"failureReason,omitempty"`
	IPAddress        string            `json:"ipAddress,omitempty"`
	DeviceID         string            `json:"deviceId,omitempty"`
	Country          string            `json:"country,omitempty"`
	ConfirmationsReq int               `json:"confirmationsReq,omitempty"`
	ConfirmationsRcv int               `json:"confirmationsRcv,omitempty"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
	CompletedAt      *time.Time        `json:"completedAt,omitempty"`
}

// PaymentRequest represents a payment request
type PaymentRequest struct {
	UserID         string            `json:"userId"`
	Type           TransactionType   `json:"type"`
	Method         PaymentMethod     `json:"method"`
	Amount         float64           `json:"amount"`
	Currency       string            `json:"currency"`
	PaymentDetails map[string]string `json:"paymentDetails,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	IPAddress      string            `json:"ipAddress,omitempty"`
	DeviceID       string            `json:"deviceId,omitempty"`
	Country        string            `json:"country,omitempty"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	TransactionID  string            `json:"transactionId"`
	Status         TransactionStatus `json:"status"`
	ProviderURL    string            `json:"providerUrl,omitempty"`
	RequiresAction bool              `json:"requiresAction"`
	Message        string            `json:"message,omitempty"`
	EstimatedTime  string            `json:"estimatedTime,omitempty"`
}

// FraudCheck represents the result of a fraud check
type FraudCheck struct {
	TransactionID string            `json:"transactionId"`
	RiskScore     float64           `json:"riskScore"`
	RiskLevel     string            `json:"riskLevel"`
	Flags         []string          `json:"flags"`
	Blocked       bool              `json:"blocked"`
	Reason        string            `json:"reason,omitempty"`
	Checks        map[string]string `json:"checks,omitempty"`
}

// PaymentLimits represents payment limits for a method
type PaymentLimits struct {
	Method               PaymentMethod `json:"method"`
	MinAmount            float64       `json:"minAmount"`
	MaxAmount            float64       `json:"maxAmount"`
	DailyLimit           float64       `json:"dailyLimit"`
	WeeklyLimit          float64       `json:"weeklyLimit"`
	MonthlyLimit         float64       `json:"monthlyLimit"`
	RequiresVerification bool          `json:"requiresVerification"`
}

// ReconciliationResult represents the result of reconciling a transaction
type ReconciliationResult struct {
	TransactionID  string            `json:"transactionId"`
	ProviderTxID   string            `json:"providerTxId"`
	OurStatus      TransactionStatus `json:"ourStatus"`
	ProviderStatus TransactionStatus `json:"providerStatus"`
	Matched        bool              `json:"matched"`
	Discrepancy    string            `json:"discrepancy,omitempty"`
}

// WithdrawalDetails contains withdrawal destination details
type WithdrawalDetails struct {
	BankName      string `json:"bankName,omitempty"`
	AccountNumber string `json:"accountNumber,omitempty"`
	RoutingNumber string `json:"routingNumber,omitempty"`
	IBAN          string `json:"iban,omitempty"`
	SWIFT         string `json:"swift,omitempty"`
	CryptoAddress string `json:"cryptoAddress,omitempty"`
	CryptoNetwork string `json:"cryptoNetwork,omitempty"` // BTC, ETH, TRC20, ERC20
}

// Gateway implements PaymentGateway interface
type Gateway struct {
	mu              sync.RWMutex
	providers       map[string]Provider
	repository      Repository
	fraud           *FraudChecker
	approvalService *WithdrawalApprovalService
}

// NewGateway creates a new payment gateway
func NewGateway(repository Repository, fraud *FraudChecker) *Gateway {
	return &Gateway{
		providers:  make(map[string]Provider),
		repository: repository,
		fraud:      fraud,
	}
}

// RegisterProvider registers a payment provider
func (g *Gateway) RegisterProvider(name string, provider Provider) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.providers[name] = provider
}

// SetApprovalService sets the withdrawal approval service
func (g *Gateway) SetApprovalService(service *WithdrawalApprovalService) {
	g.approvalService = service
}

// InitiateDeposit initiates a deposit transaction
func (g *Gateway) InitiateDeposit(userID string, amount float64, currency string, method PaymentMethod) (*PaymentSession, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Check fraud score
	riskScore := g.fraud.CalculateRiskScore(userID, amount)
	if riskScore >= 80 {
		return nil, ErrHighRiskTransaction
	}

	// Calculate fee
	fee := g.calculateDepositFee(amount, method)
	netAmount := amount - fee

	// Create transaction
	tx := &Transaction{
		ID:        NewTransactionID(TypeDeposit),
		UserID:    userID,
		Type:      TypeDeposit,
		Method:    method,
		Amount:    amount,
		Currency:  currency,
		Fee:       fee,
		NetAmount: netAmount,
		Status:    StatusPending,
		Metadata:  map[string]string{"risk_score": fmt.Sprintf("%d", riskScore)},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Select provider
	providerName := g.selectProvider(method)
	tx.Provider = providerName

	g.mu.RLock()
	provider, exists := g.providers[providerName]
	g.mu.RUnlock()

	if !exists {
		return nil, ErrProviderNotAvailable
	}

	// Save transaction
	if err := g.repository.SaveTransaction(context.Background(), tx); err != nil {
		return nil, err
	}

	// Initiate charge with provider
	session, err := provider.Charge(tx)
	if err != nil {
		tx.Status = StatusFailed
		tx.FailureReason = err.Error()
		g.repository.UpdateStatus(tx.ID, StatusFailed)
		return nil, err
	}

	return session, nil
}

// InitiateWithdrawal initiates a withdrawal transaction
func (g *Gateway) InitiateWithdrawal(userID string, amount float64, currency string, method PaymentMethod, details *WithdrawalDetails) (*Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// STEP 4: Same-method check (AML compliance)
	if !g.checkSameMethod(userID, method) {
		return nil, errors.New("withdrawal method must match a previous deposit method (AML compliance)")
	}

	// Check fraud score
	riskScore := g.fraud.CalculateRiskScore(userID, amount)

	// Calculate fee
	fee := g.calculateWithdrawalFee(amount, method)
	netAmount := amount - fee

	// Create transaction
	tx := &Transaction{
		ID:        NewTransactionID(TypeWithdrawal),
		UserID:    userID,
		Type:      TypeWithdrawal,
		Method:    method,
		Amount:    amount,
		Currency:  currency,
		Fee:       fee,
		NetAmount: netAmount,
		Status:    StatusPending,
		Metadata:  map[string]string{"risk_score": fmt.Sprintf("%d", riskScore)},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Select provider
	providerName := g.selectProvider(method)
	tx.Provider = providerName

	// Save transaction
	if err := g.repository.SaveTransaction(context.Background(), tx); err != nil {
		return nil, err
	}

	// STEP 3: Check if approval required (risk >= 60 OR amount >= 10000)
	requiresApproval := riskScore >= 60 || amount >= 10000

	if requiresApproval && g.approvalService != nil {
		// Build flags for admin review
		var flags []string
		if riskScore >= 60 {
			flags = append(flags, fmt.Sprintf("High risk score: %d", riskScore))
		}
		if amount >= 10000 {
			flags = append(flags, fmt.Sprintf("Large amount: $%.2f", amount))
		}

		// Submit to approval queue
		if err := g.approvalService.SubmitForApproval(tx, riskScore, flags); err != nil {
			return nil, err
		}

		// Check if auto-approved by service
		if g.approvalService.IsAutoApproved(tx.ID) {
			// Auto-approved - process immediately
			tx.Status = StatusPending
			g.repository.UpdateStatus(tx.ID, StatusPending)

			g.mu.RLock()
			provider, exists := g.providers[providerName]
			g.mu.RUnlock()

			if exists {
				if err := provider.Payout(tx, details); err != nil {
					tx.Status = StatusFailed
					tx.FailureReason = err.Error()
					g.repository.UpdateStatus(tx.ID, StatusFailed)
					return tx, err
				}
			}
		} else {
			// Queued for manual approval
			tx.Status = TransactionStatus("PENDING_APPROVAL")
			g.repository.UpdateStatus(tx.ID, tx.Status)
			tx.Metadata["requires_approval"] = "true"
		}
	} else {
		// Auto-approved - process with provider
		g.mu.RLock()
		provider, exists := g.providers[providerName]
		g.mu.RUnlock()

		if exists {
			if err := provider.Payout(tx, details); err != nil {
				tx.Status = StatusFailed
				tx.FailureReason = err.Error()
				g.repository.UpdateStatus(tx.ID, StatusFailed)
				return tx, err
			}
		}
	}

	return tx, nil
}

// HandleWebhook processes provider webhook
func (g *Gateway) HandleWebhook(providerName string, payload []byte, signature string) error {
	g.mu.RLock()
	provider, exists := g.providers[providerName]
	g.mu.RUnlock()

	if !exists {
		return ErrProviderNotAvailable
	}

	// Verify webhook signature
	if !provider.VerifyWebhook(payload, signature) {
		return errors.New("invalid webhook signature")
	}

	// Extract status from payload
	txID, status, err := provider.GetStatus(payload)
	if err != nil {
		return err
	}

	// Update transaction
	tx, err := g.repository.GetTransaction(context.Background(), txID)
	if err != nil {
		return err
	}

	var newStatus TransactionStatus
	switch status {
	case "succeeded", "completed":
		newStatus = StatusCompleted
		now := time.Now()
		tx.CompletedAt = &now
	case "failed":
		newStatus = StatusFailed
	case "canceled":
		newStatus = StatusCancelled
	default:
		newStatus = StatusPending
	}

	tx.Status = newStatus
	tx.UpdatedAt = time.Now()

	return g.repository.UpdateStatus(tx.ID, newStatus)
}

// GetTransactionStatus retrieves transaction status
func (g *Gateway) GetTransactionStatus(txID string) (*Transaction, error) {
	return g.repository.GetTransaction(context.Background(), txID)
}

// GetUserTransactions retrieves all user transactions
func (g *Gateway) GetUserTransactions(userID string) ([]*Transaction, error) {
	return g.repository.ListByUser(userID)
}

// Helper methods

func (g *Gateway) calculateDepositFee(amount float64, method PaymentMethod) float64 {
	switch method {
	case MethodCard:
		return amount * 0.029 // 2.9%
	case MethodPayPal:
		return amount * 0.035 // 3.5%
	case MethodBankTransfer, MethodACH, MethodSEPA:
		return 0.0 // Free
	case MethodWire:
		return 25.0 // $25 flat
	case MethodBitcoin, MethodEthereum, MethodUSDT:
		return 0.0 // Network fees handled separately
	default:
		return 0.0
	}
}

func (g *Gateway) calculateWithdrawalFee(amount float64, method PaymentMethod) float64 {
	switch method {
	case MethodCard, MethodPayPal:
		return amount * 0.02 // 2%
	case MethodBankTransfer, MethodACH, MethodSEPA:
		return 0.0 // Free
	case MethodWire:
		return 25.0 // $25 flat
	case MethodBitcoin:
		return 0.0005 * amount // Network fee approximation
	case MethodEthereum:
		return 0.001 * amount
	case MethodUSDT:
		return 1.0 // $1 flat for USDT
	default:
		return 0.0
	}
}

func (g *Gateway) selectProvider(method PaymentMethod) string {
	switch method {
	case MethodCard, MethodPayPal:
		return "stripe"
	case MethodBitcoin, MethodEthereum, MethodUSDT:
		return "crypto"
	case MethodBankTransfer, MethodACH, MethodSEPA, MethodWire:
		return "wise"
	default:
		return "stripe"
	}
}

// checkSameMethod checks if withdrawal method matches a previous deposit method (AML compliance)
// STEP 4: Same-method check
func (g *Gateway) checkSameMethod(userID string, method PaymentMethod) bool {
	// Get user's transaction history
	transactions, err := g.repository.ListByUser(userID)
	if err != nil {
		return false
	}

	// Check if user has made deposits with this method
	for _, tx := range transactions {
		if tx.Type == TypeDeposit && tx.Status == StatusCompleted && tx.Method == method {
			return true // Found a completed deposit with this method
		}
	}

	// No deposit found with this method
	return false
}

// NewTransactionID generates a unique transaction ID
func NewTransactionID(txType TransactionType) string {
	prefix := "DEP"
	if txType == TypeWithdrawal {
		prefix = "WTH"
	} else if txType == TypeRefund {
		prefix = "REF"
	} else if txType == TypeChargeback {
		prefix = "CHB"
	}
	return fmt.Sprintf("%s_%s", prefix, uuid.New().String())
}
