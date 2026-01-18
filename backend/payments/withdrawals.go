package payments

import (
	"context"
	"fmt"
	"time"
)

// WithdrawalService handles withdrawal processing
type WithdrawalService struct {
	gateway    Gateway
	providers  map[PaymentProvider]Provider
	fraud      FraudDetector
	limits     LimitsChecker
	repository Repository
	verifier   WithdrawalVerifier
}

// NewWithdrawalService creates a new withdrawal service
func NewWithdrawalService(
	gateway Gateway,
	providers map[PaymentProvider]Provider,
	fraud FraudDetector,
	limits LimitsChecker,
	repository Repository,
	verifier WithdrawalVerifier,
) *WithdrawalService {
	return &WithdrawalService{
		gateway:    gateway,
		providers:  providers,
		fraud:      fraud,
		limits:     limits,
		repository: repository,
		verifier:   verifier,
	}
}

// ProcessWithdrawal processes a withdrawal request
func (s *WithdrawalService) ProcessWithdrawal(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// Step 1: Validate request
	if err := s.validateWithdrawalRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Step 2: Verify user (2FA, email confirmation)
	if err := s.verifier.VerifyWithdrawalRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("verification failed: %w", err)
	}

	// Step 3: Check for pending withdrawals
	hasPending, err := s.repository.HasPendingWithdrawal(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, ErrPendingWithdrawal
	}

	// Step 4: Check balance
	balance, err := s.repository.GetUserBalance(ctx, req.UserID, req.Currency)
	if err != nil {
		return nil, err
	}
	if balance < req.Amount {
		return nil, ErrInsufficientFunds
	}

	// Step 5: Check withdrawal limits
	if err := s.limits.CheckWithdrawalLimits(ctx, req.UserID, req.Method, req.Amount); err != nil {
		return nil, fmt.Errorf("limit check failed: %w", err)
	}

	// Step 6: Same-method withdrawal check (AML compliance)
	if err := s.checkSameMethodRule(ctx, req); err != nil {
		return nil, err
	}

	// Step 7: Fraud detection
	fraudCheck, err := s.fraud.CheckWithdrawal(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fraud check failed: %w", err)
	}
	if fraudCheck.Blocked {
		return nil, fmt.Errorf("%w: %s", ErrFraudDetected, fraudCheck.Reason)
	}

	// Step 8: Determine if manual approval needed
	requiresApproval := s.requiresManualApproval(req.Amount, fraudCheck.RiskScore)

	// Step 9: Select provider
	provider, err := s.selectProvider(req.Method)
	if err != nil {
		return nil, err
	}

	// Step 10: Calculate fees
	fee := s.calculateWithdrawalFee(req.Method, req.Amount)
	netAmount := req.Amount - fee

	// Step 11: Create transaction record
	tx := &Transaction{
		ID:             NewTransactionID(TypeWithdrawal),
		UserID:         req.UserID,
		Type:           TypeWithdrawal,
		Method:         req.Method,
		Provider:       provider.Name(),
		Status:         StatusPending,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Fee:            fee,
		NetAmount:      netAmount,
		PaymentDetails: req.PaymentDetails,
		Metadata:       req.Metadata,
		IPAddress:      req.IPAddress,
		DeviceID:       req.DeviceID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Add approval flag
	if requiresApproval {
		if tx.Metadata == nil {
			tx.Metadata = make(map[string]string)
		}
		tx.Metadata["requires_approval"] = "true"
		tx.Metadata["risk_score"] = fmt.Sprintf("%.2f", fraudCheck.RiskScore)
	}

	// Save transaction
	if err := s.repository.SaveTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	// Step 12: Reserve funds (debit from available balance)
	if err := s.repository.ReserveUserBalance(ctx, req.UserID, req.Amount, req.Currency, tx.ID); err != nil {
		return nil, fmt.Errorf("failed to reserve funds: %w", err)
	}

	// Step 13: Process or queue for approval
	var resp *PaymentResponse
	if requiresApproval {
		resp = &PaymentResponse{
			TransactionID:  tx.ID,
			Status:         StatusPending,
			RequiresAction: false,
			Message:        "Withdrawal is pending manual approval",
			EstimatedTime:  "1-24 hours",
		}
	} else {
		// Process immediately
		resp, err = s.processWithdrawalWithProvider(ctx, provider, req, tx)
		if err != nil {
			// Unblock funds on failure
			s.repository.UnreserveUserBalance(ctx, req.UserID, req.Amount, req.Currency, tx.ID)
			tx.Status = StatusFailed
			tx.FailureReason = err.Error()
			tx.UpdatedAt = time.Now()
			s.repository.UpdateTransaction(ctx, tx)
			return nil, err
		}
	}

	resp.TransactionID = tx.ID
	return resp, nil
}

// ApproveWithdrawal manually approves a pending withdrawal
func (s *WithdrawalService) ApproveWithdrawal(ctx context.Context, txID string, approverID string) error {
	tx, err := s.repository.GetTransaction(ctx, txID)
	if err != nil {
		return err
	}

	if tx.Type != TypeWithdrawal {
		return fmt.Errorf("transaction is not a withdrawal")
	}

	if tx.Status != StatusPending {
		return fmt.Errorf("withdrawal is not pending approval")
	}

	// Get provider
	provider := s.providers[tx.Provider]
	if provider == nil {
		return ErrProviderNotAvailable
	}

	// Build payment request from transaction
	req := &PaymentRequest{
		UserID:         tx.UserID,
		Type:           TypeWithdrawal,
		Method:         tx.Method,
		Amount:         tx.Amount,
		Currency:       tx.Currency,
		PaymentDetails: tx.PaymentDetails,
		Metadata:       tx.Metadata,
	}

	// Process with provider
	resp, err := s.processWithdrawalWithProvider(ctx, provider, req, tx)
	if err != nil {
		// Unblock funds on failure
		s.repository.UnreserveUserBalance(ctx, tx.UserID, tx.Amount, tx.Currency, tx.ID)
		tx.Status = StatusFailed
		tx.FailureReason = err.Error()
		tx.UpdatedAt = time.Now()
		s.repository.UpdateTransaction(ctx, tx)
		return err
	}

	// Update metadata
	if tx.Metadata == nil {
		tx.Metadata = make(map[string]string)
	}
	tx.Metadata["approved_by"] = approverID
	tx.Metadata["approved_at"] = time.Now().Format(time.RFC3339)

	tx.Status = resp.Status
	tx.UpdatedAt = time.Now()

	return s.repository.UpdateTransaction(ctx, tx)
}

// RejectWithdrawal manually rejects a pending withdrawal
func (s *WithdrawalService) RejectWithdrawal(ctx context.Context, txID string, approverID string, reason string) error {
	tx, err := s.repository.GetTransaction(ctx, txID)
	if err != nil {
		return err
	}

	if tx.Type != TypeWithdrawal {
		return fmt.Errorf("transaction is not a withdrawal")
	}

	if tx.Status != StatusPending {
		return fmt.Errorf("withdrawal is not pending approval")
	}

	// Unblock funds
	if err := s.repository.UnreserveUserBalance(ctx, tx.UserID, tx.Amount, tx.Currency, tx.ID); err != nil {
		return err
	}

	// Update transaction
	tx.Status = StatusCancelled
	tx.FailureReason = reason
	tx.UpdatedAt = time.Now()

	if tx.Metadata == nil {
		tx.Metadata = make(map[string]string)
	}
	tx.Metadata["rejected_by"] = approverID
	tx.Metadata["rejected_at"] = time.Now().Format(time.RFC3339)

	return s.repository.UpdateTransaction(ctx, tx)
}

// CancelWithdrawal cancels a pending withdrawal (user-initiated)
func (s *WithdrawalService) CancelWithdrawal(ctx context.Context, txID string, userID string) error {
	tx, err := s.repository.GetTransaction(ctx, txID)
	if err != nil {
		return err
	}

	if tx.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	if tx.Type != TypeWithdrawal {
		return fmt.Errorf("transaction is not a withdrawal")
	}

	if tx.Status != StatusPending {
		return fmt.Errorf("withdrawal cannot be cancelled")
	}

	// Try to cancel with provider if already initiated
	if tx.ProviderTxID != "" {
		provider := s.providers[tx.Provider]
		if provider != nil {
			provider.CancelWithdrawal(ctx, tx.ProviderTxID)
		}
	}

	// Unblock funds
	if err := s.repository.UnreserveUserBalance(ctx, tx.UserID, tx.Amount, tx.Currency, tx.ID); err != nil {
		return err
	}

	tx.Status = StatusCancelled
	tx.UpdatedAt = time.Now()

	return s.repository.UpdateTransaction(ctx, tx)
}

// VerifyWithdrawal verifies a withdrawal status
func (s *WithdrawalService) VerifyWithdrawal(ctx context.Context, txID string) (*Transaction, error) {
	tx, err := s.repository.GetTransaction(ctx, txID)
	if err != nil {
		return nil, err
	}

	if tx.Type != TypeWithdrawal {
		return nil, fmt.Errorf("transaction is not a withdrawal")
	}

	// Check with provider if processing
	if tx.Status == StatusProcessing && tx.ProviderTxID != "" {
		provider := s.providers[tx.Provider]
		if provider != nil {
			providerTx, err := provider.VerifyWithdrawal(ctx, tx.ProviderTxID)
			if err == nil {
				tx.Status = providerTx.Status
				tx.UpdatedAt = time.Now()

				if providerTx.Status == StatusCompleted {
					now := time.Now()
					tx.CompletedAt = &now
					// Debit reserved funds
					s.repository.DebitReservedBalance(ctx, tx.UserID, tx.Amount, tx.Currency, tx.ID)
				} else if providerTx.Status == StatusFailed {
					// Unblock funds on failure
					s.repository.UnreserveUserBalance(ctx, tx.UserID, tx.Amount, tx.Currency, tx.ID)
				}

				s.repository.UpdateTransaction(ctx, tx)
			}
		}
	}

	return tx, nil
}

// GetWithdrawalMethods returns available withdrawal methods for a user
func (s *WithdrawalService) GetWithdrawalMethods(ctx context.Context, userID string) ([]WithdrawalMethod, error) {
	methods := []WithdrawalMethod{}

	// Get user's deposit methods (same-method rule)
	depositMethods, err := s.repository.GetUserDepositMethods(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, provider := range s.providers {
		for _, method := range provider.SupportedMethods() {
			// Check if user has deposited with this method
			hasDeposited := false
			for _, dm := range depositMethods {
				if dm == method {
					hasDeposited = true
					break
				}
			}

			limits, err := s.limits.GetLimits(ctx, userID, method)
			if err != nil {
				continue
			}

			wm := WithdrawalMethod{
				Method:        method,
				Provider:      provider.Name(),
				MinAmount:     limits.MinAmount,
				MaxAmount:     limits.MaxAmount,
				Fee:           s.calculateWithdrawalFee(method, 0),
				ProcessingTime: s.getProcessingTime(method),
				SameMethodOnly: !hasDeposited,
			}

			methods = append(methods, wm)
		}
	}

	return methods, nil
}

// Helper functions

func (s *WithdrawalService) validateWithdrawalRequest(req *PaymentRequest) error {
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

func (s *WithdrawalService) checkSameMethodRule(ctx context.Context, req *PaymentRequest) error {
	// AML regulation: withdrawals must use same method as deposits
	hasDeposited, err := s.repository.HasDepositedWithMethod(ctx, req.UserID, req.Method)
	if err != nil {
		return err
	}

	if !hasDeposited {
		return ErrSameMethodRequired
	}

	return nil
}

func (s *WithdrawalService) requiresManualApproval(amount float64, riskScore float64) bool {
	// Large amounts require manual approval
	if amount >= 10000 {
		return true
	}

	// High risk scores require manual approval
	if riskScore >= 70 {
		return true
	}

	return false
}

func (s *WithdrawalService) processWithdrawalWithProvider(
	ctx context.Context,
	provider Provider,
	req *PaymentRequest,
	tx *Transaction,
) (*PaymentResponse, error) {
	resp, err := provider.InitiateWithdrawal(ctx, req)
	if err != nil {
		return nil, err
	}

	// Update transaction
	tx.ProviderTxID = resp.TransactionID
	tx.Status = resp.Status
	tx.UpdatedAt = time.Now()

	if resp.Status == StatusCompleted {
		now := time.Now()
		tx.CompletedAt = &now
		// Debit reserved funds
		s.repository.DebitReservedBalance(ctx, tx.UserID, tx.Amount, tx.Currency, tx.ID)
	}

	s.repository.UpdateTransaction(ctx, tx)

	return resp, nil
}

func (s *WithdrawalService) selectProvider(method PaymentMethod) (Provider, error) {
	for _, provider := range s.providers {
		for _, supported := range provider.SupportedMethods() {
			if supported == method {
				return provider, nil
			}
		}
	}
	return nil, ErrMethodNotSupported
}

func (s *WithdrawalService) calculateWithdrawalFee(method PaymentMethod, amount float64) float64 {
	switch method {
	case MethodBankTransfer, MethodACH, MethodSEPA:
		return 0 // Free for bank transfers
	case MethodWire:
		return 25.0 // Flat fee for wire transfers
	case MethodPayPal, MethodSkrill, MethodNeteller:
		return amount * 0.02 // 2% for e-wallets
	case MethodBitcoin:
		return 0.0005 // BTC network fee estimate
	case MethodEthereum, MethodUSDT:
		return 0.01 // ETH gas fee estimate
	default:
		return 0
	}
}

func (s *WithdrawalService) getProcessingTime(method PaymentMethod) string {
	switch method {
	case MethodBitcoin, MethodEthereum, MethodUSDT:
		return "Within 30 minutes"
	case MethodPayPal, MethodSkrill, MethodNeteller:
		return "1-2 business days"
	case MethodACH:
		return "2-3 business days"
	case MethodSEPA:
		return "1-2 business days"
	case MethodWire:
		return "1-3 business days"
	default:
		return "Unknown"
	}
}

// WithdrawalMethod represents an available withdrawal method
type WithdrawalMethod struct {
	Method         PaymentMethod   `json:"method"`
	Provider       PaymentProvider `json:"provider"`
	MinAmount      float64         `json:"min_amount"`
	MaxAmount      float64         `json:"max_amount"`
	Fee            float64         `json:"fee"`
	ProcessingTime string          `json:"processing_time"`
	SameMethodOnly bool            `json:"same_method_only"`
}

// WithdrawalVerifier handles withdrawal verification (2FA, email)
type WithdrawalVerifier interface {
	VerifyWithdrawalRequest(ctx context.Context, req *PaymentRequest) error
}
