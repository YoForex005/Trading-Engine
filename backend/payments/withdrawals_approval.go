package payments

import (
	"errors"
	"sync"
	"time"
)

// ApprovalStatus represents withdrawal approval status
type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "PENDING"
	ApprovalApproved ApprovalStatus = "APPROVED"
	ApprovalRejected ApprovalStatus = "REJECTED"
)

// WithdrawalRequest represents a withdrawal pending approval
type WithdrawalRequest struct {
	TransactionID   string         `json:"transactionId"`
	UserID          string         `json:"userId"`
	Amount          float64        `json:"amount"`
	Currency        string         `json:"currency"`
	Method          PaymentMethod  `json:"method"`
	RiskScore       int            `json:"riskScore"`
	Flags           []string       `json:"flags"`
	Status          ApprovalStatus `json:"status"`
	AssignedTo      string         `json:"assignedTo,omitempty"`
	ApprovedBy      string         `json:"approvedBy,omitempty"`
	ApprovedAt      *time.Time     `json:"approvedAt,omitempty"`
	RejectedBy      string         `json:"rejectedBy,omitempty"`
	RejectedAt      *time.Time     `json:"rejectedAt,omitempty"`
	RejectionReason string         `json:"rejectionReason,omitempty"`
	CreatedAt       time.Time      `json:"createdAt"`
	
	// User context (populated on retrieval)
	UserDepositHistory    []*Transaction `json:"userDepositHistory,omitempty"`
	UserWithdrawalHistory []*Transaction `json:"userWithdrawalHistory,omitempty"`
	AccountAge            int            `json:"accountAge,omitempty"` // days since first transaction
}

// WithdrawalApprovalService manages withdrawal approval workflow
type WithdrawalApprovalService struct {
	mu         sync.RWMutex
	queue      map[string]*WithdrawalRequest // txID -> request
	repository Repository                     // for user history queries
	gateway    *Gateway                       // for processing approved withdrawals
}

// NewWithdrawalApprovalService creates a new approval service
func NewWithdrawalApprovalService(repository Repository) *WithdrawalApprovalService {
	return &WithdrawalApprovalService{
		queue:      make(map[string]*WithdrawalRequest),
		repository: repository,
	}
}

// SetGateway sets the payment gateway reference (for processing approved withdrawals)
func (s *WithdrawalApprovalService) SetGateway(gateway *Gateway) {
	s.gateway = gateway
}

// SubmitForApproval submits a withdrawal for approval
// Auto-approves if riskScore < 60 AND amount < 10000
// Otherwise queues for manual approval
func (s *WithdrawalApprovalService) SubmitForApproval(tx *Transaction, riskScore int, flags []string) error {
	if tx == nil {
		return errors.New("transaction cannot be nil")
	}

	request := &WithdrawalRequest{
		TransactionID: tx.ID,
		UserID:        tx.UserID,
		Amount:        tx.Amount,
		Currency:      tx.Currency,
		Method:        tx.Method,
		RiskScore:     riskScore,
		Flags:         flags,
		Status:        ApprovalPending,
		CreatedAt:     time.Now(),
	}

	// Auto-approve if low risk AND low amount
	if riskScore < 60 && tx.Amount < 10000 {
		// Auto-approve
		now := time.Now()
		request.Status = ApprovalApproved
		request.ApprovedBy = "SYSTEM_AUTO"
		request.ApprovedAt = &now

		// Store in queue for audit trail
		s.mu.Lock()
		s.queue[tx.ID] = request
		s.mu.Unlock()

		// Process withdrawal immediately
		// (Gateway will handle actual payout)
		return nil
	}

	// Queue for manual approval
	request.Status = ApprovalPending

	s.mu.Lock()
	s.queue[tx.ID] = request
	s.mu.Unlock()

	return nil
}

// GetPendingRequests returns all pending approval requests
func (s *WithdrawalApprovalService) GetPendingRequests() []*WithdrawalRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var pending []*WithdrawalRequest
	for _, req := range s.queue {
		if req.Status == ApprovalPending {
			// Enrich with user context
			enriched := s.enrichWithUserContext(req)
			pending = append(pending, enriched)
		}
	}

	return pending
}

// GetHistory returns processed withdrawal requests (approved/rejected)
func (s *WithdrawalApprovalService) GetHistory(status ApprovalStatus, limit, offset int) []*WithdrawalRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var history []*WithdrawalRequest
	for _, req := range s.queue {
		if status == "" || req.Status == status {
			if req.Status != ApprovalPending {
				history = append(history, req)
			}
		}
	}

	// Sort by created date descending (newest first)
	for i := 0; i < len(history)-1; i++ {
		for j := i + 1; j < len(history); j++ {
			if history[i].CreatedAt.Before(history[j].CreatedAt) {
				history[i], history[j] = history[j], history[i]
			}
		}
	}

	// Apply pagination
	if offset >= len(history) {
		return []*WithdrawalRequest{}
	}

	end := offset + limit
	if end > len(history) {
		end = len(history)
	}

	return history[offset:end]
}

// GetRequest retrieves a specific withdrawal request with user context
func (s *WithdrawalApprovalService) GetRequest(txID string) (*WithdrawalRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	req, exists := s.queue[txID]
	if !exists {
		return nil, ErrTransactionNotFound
	}

	// Enrich with user context
	enriched := s.enrichWithUserContext(req)
	return enriched, nil
}

// Approve approves a withdrawal request
func (s *WithdrawalApprovalService) Approve(txID string, adminID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	req, exists := s.queue[txID]
	if !exists {
		return ErrTransactionNotFound
	}

	if req.Status != ApprovalPending {
		return errors.New("withdrawal is not pending approval")
	}

	now := time.Now()
	req.Status = ApprovalApproved
	req.ApprovedBy = adminID
	req.ApprovedAt = &now

	// NOTE: Actual payout processing should be handled by the gateway
	// The caller (handler) should retrieve the transaction and process it

	return nil
}

// Reject rejects a withdrawal request
func (s *WithdrawalApprovalService) Reject(txID string, adminID string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	req, exists := s.queue[txID]
	if !exists {
		return ErrTransactionNotFound
	}

	if req.Status != ApprovalPending {
		return errors.New("withdrawal is not pending approval")
	}

	now := time.Now()
	req.Status = ApprovalRejected
	req.RejectedBy = adminID
	req.RejectedAt = &now
	req.RejectionReason = reason

	return nil
}

// enrichWithUserContext adds user deposit/withdrawal history to request
func (s *WithdrawalApprovalService) enrichWithUserContext(req *WithdrawalRequest) *WithdrawalRequest {
	// Make a copy to avoid modifying the original
	enriched := *req

	// Get all user transactions
	transactions, err := s.repository.ListByUser(req.UserID)
	if err != nil {
		return &enriched
	}

	var deposits []*Transaction
	var withdrawals []*Transaction
	var firstTxTime time.Time

	for _, tx := range transactions {
		if tx.Type == TypeDeposit && tx.Status == StatusCompleted {
			deposits = append(deposits, tx)
		} else if tx.Type == TypeWithdrawal {
			withdrawals = append(withdrawals, tx)
		}

		// Track earliest transaction for account age
		if firstTxTime.IsZero() || tx.CreatedAt.Before(firstTxTime) {
			firstTxTime = tx.CreatedAt
		}
	}

	enriched.UserDepositHistory = deposits
	enriched.UserWithdrawalHistory = withdrawals

	// Calculate account age in days
	if !firstTxTime.IsZero() {
		enriched.AccountAge = int(time.Since(firstTxTime).Hours() / 24)
	}

	return &enriched
}

// IsAutoApproved checks if a withdrawal was auto-approved
func (s *WithdrawalApprovalService) IsAutoApproved(txID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	req, exists := s.queue[txID]
	if !exists {
		return false
	}

	return req.Status == ApprovalApproved && req.ApprovedBy == "SYSTEM_AUTO"
}

// GetStats returns approval queue statistics
func (s *WithdrawalApprovalService) GetStats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]int{
		"pending":  0,
		"approved": 0,
		"rejected": 0,
		"total":    len(s.queue),
	}

	for _, req := range s.queue {
		switch req.Status {
		case ApprovalPending:
			stats["pending"]++
		case ApprovalApproved:
			stats["approved"]++
		case ApprovalRejected:
			stats["rejected"]++
		}
	}

	return stats
}
