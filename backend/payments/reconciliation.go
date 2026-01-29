package payments

import (
	"context"
	"fmt"
	"time"
)

// ReconciliationService handles payment reconciliation
type ReconciliationService struct {
	providers  map[PaymentProvider]Provider
	repository Repository
}

// NewReconciliationService creates a new reconciliation service
func NewReconciliationService(
	providers map[PaymentProvider]Provider,
	repository Repository,
) *ReconciliationService {
	return &ReconciliationService{
		providers:  providers,
		repository: repository,
	}
}

// ReconcileTransactions reconciles all transactions in a time period
func (s *ReconciliationService) ReconcileTransactions(ctx context.Context, from, to time.Time) ([]ReconciliationResult, error) {
	results := []ReconciliationResult{}

	// Get all transactions in the time period
	transactions, err := s.repository.GetTransactionsByTimeRange(ctx, from, to)
	if err != nil {
		return nil, err
	}

	// Reconcile each transaction
	for _, tx := range transactions {
		result, err := s.reconcileTransaction(ctx, tx)
		if err != nil {
			// Log error but continue
			fmt.Printf("Error reconciling transaction %s: %v\n", tx.ID, err)
			continue
		}
		results = append(results, *result)
	}

	return results, nil
}

// ReconcileProvider reconciles all transactions for a specific provider
func (s *ReconciliationService) ReconcileProvider(ctx context.Context, provider PaymentProvider, from, to time.Time) ([]ReconciliationResult, error) {
	results := []ReconciliationResult{}

	// Get provider transactions
	transactions, err := s.repository.GetProviderTransactions(ctx, provider, from, to)
	if err != nil {
		return nil, err
	}

	// Get provider
	p := s.providers[provider]
	if p == nil {
		return nil, ErrProviderNotAvailable
	}

	// Reconcile each transaction
	for _, tx := range transactions {
		result, err := s.reconcileTransaction(ctx, tx)
		if err != nil {
			fmt.Printf("Error reconciling transaction %s: %v\n", tx.ID, err)
			continue
		}
		results = append(results, *result)
	}

	return results, nil
}

// ReconcileTransaction reconciles a single transaction
func (s *ReconciliationService) ReconcileTransaction(ctx context.Context, txID string) (*ReconciliationResult, error) {
	tx, err := s.repository.GetTransaction(ctx, txID)
	if err != nil {
		return nil, err
	}

	return s.reconcileTransaction(ctx, tx)
}

// GenerateSettlementReport generates a settlement report for a time period
func (s *ReconciliationService) GenerateSettlementReport(ctx context.Context, from, to time.Time) (*SettlementReport, error) {
	report := &SettlementReport{
		From:      from,
		To:        to,
		Generated: time.Now(),
		Providers: make(map[PaymentProvider]*ProviderSettlement),
	}

	// Get all completed transactions
	transactions, err := s.repository.GetCompletedTransactions(ctx, from, to)
	if err != nil {
		return nil, err
	}

	// Group by provider
	for _, tx := range transactions {
		if report.Providers[tx.Provider] == nil {
			report.Providers[tx.Provider] = &ProviderSettlement{
				Provider:     tx.Provider,
				Deposits:     &TransactionSummary{},
				Withdrawals:  &TransactionSummary{},
			}
		}

		ps := report.Providers[tx.Provider]

		if tx.Type == TypeDeposit {
			ps.Deposits.Count++
			ps.Deposits.TotalAmount += tx.Amount
			ps.Deposits.TotalFees += tx.Fee
			ps.Deposits.NetAmount += tx.NetAmount
		} else if tx.Type == TypeWithdrawal {
			ps.Withdrawals.Count++
			ps.Withdrawals.TotalAmount += tx.Amount
			ps.Withdrawals.TotalFees += tx.Fee
			ps.Withdrawals.NetAmount += tx.NetAmount
		}
	}

	// Calculate totals
	for _, ps := range report.Providers {
		report.TotalDeposits += ps.Deposits.TotalAmount
		report.TotalWithdrawals += ps.Withdrawals.TotalAmount
		report.TotalFees += ps.Deposits.TotalFees + ps.Withdrawals.TotalFees
	}

	report.NetSettlement = report.TotalDeposits - report.TotalWithdrawals - report.TotalFees

	return report, nil
}

// HandleChargeback handles a chargeback event
func (s *ReconciliationService) HandleChargeback(ctx context.Context, txID string, reason string) error {
	tx, err := s.repository.GetTransaction(ctx, txID)
	if err != nil {
		return err
	}

	if tx.Type != TypeDeposit {
		return fmt.Errorf("only deposits can be charged back")
	}

	if tx.Status != StatusCompleted {
		return fmt.Errorf("only completed deposits can be charged back")
	}

	// Create chargeback transaction
	chargeback := &Transaction{
		ID:            NewTransactionID(TypeChargeback),
		UserID:        tx.UserID,
		Type:          TypeChargeback,
		Method:        tx.Method,
		Provider:      tx.Provider,
		Status:        StatusCompleted,
		Amount:        tx.Amount,
		Currency:      tx.Currency,
		Fee:           0,
		NetAmount:     tx.Amount,
		ProviderTxID:  tx.ProviderTxID,
		FailureReason: reason,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	now := time.Now()
	chargeback.CompletedAt = &now

	if chargeback.Metadata == nil {
		chargeback.Metadata = make(map[string]string)
	}
	chargeback.Metadata["original_tx_id"] = tx.ID

	// Save chargeback transaction
	if err := s.repository.SaveTransaction(ctx, chargeback); err != nil {
		return err
	}

	// Debit user account
	if err := s.repository.DebitUserBalance(ctx, tx.UserID, tx.Amount, tx.Currency, chargeback.ID); err != nil {
		return err
	}

	// Update original transaction
	tx.Status = StatusDisputed
	tx.UpdatedAt = time.Now()
	if tx.Metadata == nil {
		tx.Metadata = make(map[string]string)
	}
	tx.Metadata["chargeback_tx_id"] = chargeback.ID
	tx.Metadata["chargeback_reason"] = reason

	return s.repository.UpdateTransaction(ctx, tx)
}

// HandleDispute handles a dispute on a transaction
func (s *ReconciliationService) HandleDispute(ctx context.Context, txID string, reason string, evidence map[string]string) error {
	tx, err := s.repository.GetTransaction(ctx, txID)
	if err != nil {
		return err
	}

	tx.Status = StatusDisputed
	tx.FailureReason = reason
	tx.UpdatedAt = time.Now()

	if tx.Metadata == nil {
		tx.Metadata = make(map[string]string)
	}
	tx.Metadata["dispute_reason"] = reason
	tx.Metadata["dispute_opened_at"] = time.Now().Format(time.RFC3339)

	// Store evidence
	for k, v := range evidence {
		tx.Metadata["evidence_"+k] = v
	}

	return s.repository.UpdateTransaction(ctx, tx)
}

// ProcessRefund processes a refund for a transaction
func (s *ReconciliationService) ProcessRefund(ctx context.Context, txID string, amount float64, reason string) error {
	tx, err := s.repository.GetTransaction(ctx, txID)
	if err != nil {
		return err
	}

	if tx.Type != TypeDeposit {
		return fmt.Errorf("only deposits can be refunded")
	}

	if tx.Status != StatusCompleted {
		return fmt.Errorf("only completed deposits can be refunded")
	}

	if amount <= 0 || amount > tx.Amount {
		return ErrInvalidAmount
	}

	// Create refund transaction
	refund := &Transaction{
		ID:           NewTransactionID(TypeRefund),
		UserID:       tx.UserID,
		Type:         TypeRefund,
		Method:       tx.Method,
		Provider:     tx.Provider,
		Status:       StatusPending,
		Amount:       amount,
		Currency:     tx.Currency,
		Fee:          0,
		NetAmount:    amount,
		ProviderTxID: tx.ProviderTxID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if refund.Metadata == nil {
		refund.Metadata = make(map[string]string)
	}
	refund.Metadata["original_tx_id"] = tx.ID
	refund.Metadata["refund_reason"] = reason

	// Save refund transaction
	if err := s.repository.SaveTransaction(ctx, refund); err != nil {
		return err
	}

	// Process with provider
	provider := s.providers[tx.Provider]
	if provider == nil {
		return ErrProviderNotAvailable
	}

	// Initiate refund with provider (implementation depends on provider)
	// For now, mark as completed and debit user account
	refund.Status = StatusCompleted
	now := time.Now()
	refund.CompletedAt = &now

	if err := s.repository.UpdateTransaction(ctx, refund); err != nil {
		return err
	}

	// Debit user account
	return s.repository.DebitUserBalance(ctx, tx.UserID, amount, tx.Currency, refund.ID)
}

// RetryFailedTransaction retries a failed transaction
func (s *ReconciliationService) RetryFailedTransaction(ctx context.Context, txID string) error {
	tx, err := s.repository.GetTransaction(ctx, txID)
	if err != nil {
		return err
	}

	if tx.Status != StatusFailed {
		return fmt.Errorf("transaction is not failed")
	}

	// Check retry count
	retryCount := 0
	if tx.Metadata != nil {
		if countStr, ok := tx.Metadata["retry_count"]; ok {
			fmt.Sscanf(countStr, "%d", &retryCount)
		}
	}

	if retryCount >= 3 {
		return fmt.Errorf("max retry attempts reached")
	}

	// Get provider
	provider := s.providers[tx.Provider]
	if provider == nil {
		return ErrProviderNotAvailable
	}

	// Build payment request
	req := &PaymentRequest{
		UserID:         tx.UserID,
		Type:           tx.Type,
		Method:         tx.Method,
		Amount:         tx.Amount,
		Currency:       tx.Currency,
		PaymentDetails: tx.PaymentDetails,
		Metadata:       tx.Metadata,
	}

	// Retry based on type
	var resp *PaymentResponse
	if tx.Type == TypeDeposit {
		resp, err = provider.InitiateDeposit(ctx, req)
	} else if tx.Type == TypeWithdrawal {
		resp, err = provider.InitiateWithdrawal(ctx, req)
	} else {
		return fmt.Errorf("invalid transaction type for retry")
	}

	if err != nil {
		// Increment retry count
		retryCount++
		if tx.Metadata == nil {
			tx.Metadata = make(map[string]string)
		}
		tx.Metadata["retry_count"] = fmt.Sprintf("%d", retryCount)
		tx.Metadata["last_retry_at"] = time.Now().Format(time.RFC3339)
		tx.UpdatedAt = time.Now()
		s.repository.UpdateTransaction(ctx, tx)
		return err
	}

	// Update transaction
	tx.Status = resp.Status
	tx.ProviderTxID = resp.TransactionID
	tx.UpdatedAt = time.Now()

	if tx.Metadata == nil {
		tx.Metadata = make(map[string]string)
	}
	tx.Metadata["retry_count"] = fmt.Sprintf("%d", retryCount+1)
	tx.Metadata["retry_succeeded_at"] = time.Now().Format(time.RFC3339)

	return s.repository.UpdateTransaction(ctx, tx)
}

// Helper functions

func (s *ReconciliationService) reconcileTransaction(ctx context.Context, tx *Transaction) (*ReconciliationResult, error) {
	result := ReconciliationResult{
		TransactionID:  tx.ID,
		ProviderTxID:   tx.ProviderTxID,
		OurStatus:      tx.Status,
		Matched:        true,
	}

	// Skip if no provider transaction ID
	if tx.ProviderTxID == "" {
		result.Matched = true
		return &result, nil
	}

	// Get provider
	provider := s.providers[tx.Provider]
	if provider == nil {
		result.Matched = false
		result.Discrepancy = "provider not available"
		return &result, nil
	}

	// Verify with provider
	var providerTx *Transaction
	var err error

	if tx.Type == TypeDeposit {
		providerTx, err = provider.VerifyDeposit(ctx, tx.ProviderTxID)
	} else if tx.Type == TypeWithdrawal {
		providerTx, err = provider.VerifyWithdrawal(ctx, tx.ProviderTxID)
	} else {
		result.Matched = true
		return &result, nil
	}

	if err != nil {
		result.Matched = false
		result.Discrepancy = fmt.Sprintf("provider verification failed: %v", err)
		return &result, nil
	}

	// Compare statuses
	result.ProviderStatus = providerTx.Status

	if tx.Status != providerTx.Status {
		result.Matched = false
		result.Discrepancy = fmt.Sprintf("status mismatch: ours=%s, provider=%s", tx.Status, providerTx.Status)

		// Update our status if provider is more recent
		if providerTx.UpdatedAt.After(tx.UpdatedAt) {
			tx.Status = providerTx.Status
			tx.UpdatedAt = time.Now()

			if providerTx.Status == StatusCompleted && tx.CompletedAt == nil {
				now := time.Now()
				tx.CompletedAt = &now
			}

			s.repository.UpdateTransaction(ctx, tx)
		}
	}

	return &result, nil
}

// SettlementReport represents a settlement report
type SettlementReport struct {
	From              time.Time                          `json:"from"`
	To                time.Time                          `json:"to"`
	Generated         time.Time                          `json:"generated"`
	Providers         map[PaymentProvider]*ProviderSettlement `json:"providers"`
	TotalDeposits     float64                            `json:"total_deposits"`
	TotalWithdrawals  float64                            `json:"total_withdrawals"`
	TotalFees         float64                            `json:"total_fees"`
	NetSettlement     float64                            `json:"net_settlement"`
}

// ProviderSettlement represents settlement for a provider
type ProviderSettlement struct {
	Provider    PaymentProvider     `json:"provider"`
	Deposits    *TransactionSummary `json:"deposits"`
	Withdrawals *TransactionSummary `json:"withdrawals"`
}

// TransactionSummary summarizes transactions
type TransactionSummary struct {
	Count       int     `json:"count"`
	TotalAmount float64 `json:"total_amount"`
	TotalFees   float64 `json:"total_fees"`
	NetAmount   float64 `json:"net_amount"`
}
