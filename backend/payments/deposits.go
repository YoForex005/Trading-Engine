package payments

import (
	"context"
	"fmt"
	"time"
)

// DepositService handles deposit processing
type DepositService struct {
	gateway    Gateway
	providers  map[PaymentProvider]Provider
	fraud      FraudDetector
	limits     LimitsChecker
	repository Repository
}

// NewDepositService creates a new deposit service
func NewDepositService(
	gateway Gateway,
	providers map[PaymentProvider]Provider,
	fraud FraudDetector,
	limits LimitsChecker,
	repository Repository,
) *DepositService {
	return &DepositService{
		gateway:    gateway,
		providers:  providers,
		fraud:      fraud,
		limits:     limits,
		repository: repository,
	}
}

// ProcessDeposit processes a deposit request
func (s *DepositService) ProcessDeposit(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// Step 1: Validate request
	if err := s.validateDepositRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Step 2: Check limits
	if err := s.limits.CheckDepositLimits(ctx, req.UserID, req.Method, req.Amount); err != nil {
		return nil, fmt.Errorf("limit check failed: %w", err)
	}

	// Step 3: Fraud detection
	fraudCheck, err := s.fraud.CheckDeposit(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fraud check failed: %w", err)
	}
	if fraudCheck.Blocked {
		return nil, fmt.Errorf("%w: %s", ErrFraudDetected, fraudCheck.Reason)
	}

	// Step 4: Select provider
	provider, err := s.selectProvider(req.Method)
	if err != nil {
		return nil, err
	}

	// Step 5: Create transaction record
	tx := &Transaction{
		ID:             NewTransactionID(TypeDeposit),
		UserID:         req.UserID,
		Type:           TypeDeposit,
		Method:         req.Method,
		Provider:       provider.Name(),
		Status:         StatusPending,
		Amount:         req.Amount,
		Currency:       req.Currency,
		PaymentDetails: req.PaymentDetails,
		Metadata:       req.Metadata,
		IPAddress:      req.IPAddress,
		DeviceID:       req.DeviceID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Calculate fees
	tx.Fee = s.calculateDepositFee(req.Method, req.Amount)
	tx.NetAmount = req.Amount - tx.Fee

	// Save transaction
	if err := s.repository.SaveTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	// Step 6: Initiate with provider
	resp, err := provider.InitiateDeposit(ctx, req)
	if err != nil {
		// Update transaction status
		tx.Status = StatusFailed
		tx.FailureReason = err.Error()
		tx.UpdatedAt = time.Now()
		s.repository.UpdateTransaction(ctx, tx)
		return nil, fmt.Errorf("provider initiation failed: %w", err)
	}

	// Step 7: Update transaction with provider details
	tx.ProviderTxID = resp.TransactionID
	tx.Status = resp.Status
	tx.UpdatedAt = time.Now()

	if resp.Status == StatusCompleted {
		now := time.Now()
		tx.CompletedAt = &now
		// Credit user account
		if err := s.creditUserAccount(ctx, tx); err != nil {
			return nil, fmt.Errorf("failed to credit account: %w", err)
		}
	}

	if err := s.repository.UpdateTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	resp.TransactionID = tx.ID
	return resp, nil
}

// ProcessInstantCardDeposit handles instant card deposits (Stripe, Braintree)
func (s *DepositService) ProcessInstantCardDeposit(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	req.Method = MethodCard

	// Add 3D Secure requirement for cards
	if req.PaymentDetails == nil {
		req.PaymentDetails = make(map[string]string)
	}
	req.PaymentDetails["require_3ds"] = "true"

	return s.ProcessDeposit(ctx, req)
}

// ProcessBankTransferDeposit handles bank transfer deposits (ACH, SEPA, Wire)
func (s *DepositService) ProcessBankTransferDeposit(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// Bank transfers take 1-3 days
	resp, err := s.ProcessDeposit(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.EstimatedTime == "" {
		resp.EstimatedTime = "1-3 business days"
	}

	return resp, nil
}

// ProcessCryptoDeposit handles cryptocurrency deposits
func (s *DepositService) ProcessCryptoDeposit(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	resp, err := s.ProcessDeposit(ctx, req)
	if err != nil {
		return nil, err
	}

	// Get transaction for confirmations
	tx, err := s.repository.GetTransaction(ctx, resp.TransactionID)
	if err != nil {
		return resp, nil
	}

	// Set required confirmations based on crypto
	switch req.Method {
	case MethodBitcoin:
		tx.ConfirmationsReq = 3
	case MethodEthereum:
		tx.ConfirmationsReq = 12
	case MethodUSDT:
		tx.ConfirmationsReq = 12
	}

	s.repository.UpdateTransaction(ctx, tx)

	resp.EstimatedTime = fmt.Sprintf("~%d confirmations required", tx.ConfirmationsReq)
	return resp, nil
}

// MonitorCryptoDeposit monitors crypto deposit confirmations
func (s *DepositService) MonitorCryptoDeposit(ctx context.Context, txID string) error {
	tx, err := s.repository.GetTransaction(ctx, txID)
	if err != nil {
		return err
	}

	provider := s.providers[tx.Provider]
	if provider == nil {
		return ErrProviderNotAvailable
	}

	// Verify with provider
	providerTx, err := provider.VerifyDeposit(ctx, tx.ProviderTxID)
	if err != nil {
		return err
	}

	// Update confirmations
	tx.ConfirmationsRcv = providerTx.ConfirmationsRcv
	tx.UpdatedAt = time.Now()

	// Credit account when confirmations met
	if tx.ConfirmationsRcv >= tx.ConfirmationsReq && tx.Status != StatusCompleted {
		tx.Status = StatusCompleted
		now := time.Now()
		tx.CompletedAt = &now

		if err := s.creditUserAccount(ctx, tx); err != nil {
			return err
		}
	}

	return s.repository.UpdateTransaction(ctx, tx)
}

// VerifyDeposit verifies a deposit status
func (s *DepositService) VerifyDeposit(ctx context.Context, txID string) (*Transaction, error) {
	tx, err := s.repository.GetTransaction(ctx, txID)
	if err != nil {
		return nil, err
	}

	if tx.Type != TypeDeposit {
		return nil, fmt.Errorf("transaction is not a deposit")
	}

	// For crypto deposits, check confirmations
	if s.isCryptoMethod(tx.Method) && tx.Status == StatusProcessing {
		if err := s.MonitorCryptoDeposit(ctx, txID); err != nil {
			return nil, err
		}
		// Reload transaction
		tx, err = s.repository.GetTransaction(ctx, txID)
		if err != nil {
			return nil, err
		}
	}

	return tx, nil
}

// GetDepositMethods returns available deposit methods for a user
func (s *DepositService) GetDepositMethods(ctx context.Context, userID string) ([]DepositMethod, error) {
	methods := []DepositMethod{}

	// Check user verification level
	verificationLevel, err := s.repository.GetUserVerificationLevel(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, provider := range s.providers {
		for _, method := range provider.SupportedMethods() {
			limits, err := s.limits.GetLimits(ctx, userID, method)
			if err != nil {
				continue
			}

			// Skip if verification required but user not verified
			if limits.RequiresVerification && verificationLevel < 2 {
				continue
			}

			dm := DepositMethod{
				Method:      method,
				Provider:    provider.Name(),
				MinAmount:   limits.MinAmount,
				MaxAmount:   limits.MaxAmount,
				Fee:         s.calculateDepositFee(method, 0),
				ProcessingTime: s.getProcessingTime(method),
				RequiresVerification: limits.RequiresVerification,
			}

			methods = append(methods, dm)
		}
	}

	return methods, nil
}

// Helper functions

func (s *DepositService) validateDepositRequest(req *PaymentRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id required")
	}
	if req.Amount <= 0 {
		return ErrInvalidAmount
	}
	if req.Currency == "" {
		return fmt.Errorf("currency required")
	}
	if req.Method == "" {
		return fmt.Errorf("payment method required")
	}
	return nil
}

func (s *DepositService) selectProvider(method PaymentMethod) (Provider, error) {
	// Select best provider for the method
	for _, provider := range s.providers {
		for _, supported := range provider.SupportedMethods() {
			if supported == method {
				return provider, nil
			}
		}
	}
	return nil, ErrMethodNotSupported
}

func (s *DepositService) calculateDepositFee(method PaymentMethod, amount float64) float64 {
	// Fee calculation logic
	switch method {
	case MethodCard:
		return amount * 0.029 // 2.9% for cards
	case MethodPayPal:
		return amount * 0.035 // 3.5% for PayPal
	case MethodBankTransfer, MethodACH, MethodSEPA:
		return 0 // Free for bank transfers
	case MethodBitcoin, MethodEthereum, MethodUSDT:
		return 0 // Crypto fees paid by user on-chain
	default:
		return 0
	}
}

func (s *DepositService) getProcessingTime(method PaymentMethod) string {
	switch method {
	case MethodCard:
		return "Instant"
	case MethodPayPal, MethodSkrill, MethodNeteller:
		return "Instant"
	case MethodACH:
		return "1-3 business days"
	case MethodSEPA:
		return "1-2 business days"
	case MethodWire:
		return "1-3 business days"
	case MethodBitcoin:
		return "~30 minutes (3 confirmations)"
	case MethodEthereum, MethodUSDT:
		return "~5 minutes (12 confirmations)"
	default:
		return "Unknown"
	}
}

func (s *DepositService) isCryptoMethod(method PaymentMethod) bool {
	return method == MethodBitcoin || method == MethodEthereum || method == MethodUSDT
}

func (s *DepositService) creditUserAccount(ctx context.Context, tx *Transaction) error {
	// Credit user's account balance
	return s.repository.CreditUserBalance(ctx, tx.UserID, tx.NetAmount, tx.Currency, tx.ID)
}

// DepositMethod represents an available deposit method
type DepositMethod struct {
	Method               PaymentMethod `json:"method"`
	Provider             PaymentProvider `json:"provider"`
	MinAmount            float64       `json:"min_amount"`
	MaxAmount            float64       `json:"max_amount"`
	Fee                  float64       `json:"fee"`
	ProcessingTime       string        `json:"processing_time"`
	RequiresVerification bool          `json:"requires_verification"`
}
