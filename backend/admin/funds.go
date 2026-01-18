package admin

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
)

// FundManagementService handles fund operations
type FundManagementService struct {
	mu         sync.RWMutex
	engine     *core.Engine
	auditLog   *AuditLog
	operations map[int64]*FundOperation
	nextOpID   int64
}

// NewFundManagementService creates a new fund management service
func NewFundManagementService(engine *core.Engine, auditLog *AuditLog) *FundManagementService {
	return &FundManagementService{
		engine:     engine,
		auditLog:   auditLog,
		operations: make(map[int64]*FundOperation),
		nextOpID:   1,
	}
}

// Deposit adds funds to a user account
func (s *FundManagementService) Deposit(accountID int64, amount float64, method, reference, description, reason string, admin *Admin, ipAddress string) (*FundOperation, error) {
	if amount <= 0 {
		return nil, errors.New("deposit amount must be positive")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify account exists
	account, ok := s.engine.GetAccount(accountID)
	if !ok {
		return nil, errors.New("account not found")
	}

	if account.Status != "ACTIVE" {
		return nil, errors.New("account is not active")
	}

	// Execute deposit via ledger
	entry, err := s.engine.GetLedger().Deposit(accountID, amount, method, reference, description, admin.Username)
	if err != nil {
		s.auditLog.Log(admin.ID, admin.Username, "FUND_DEPOSIT", "FUND", accountID, map[string]interface{}{
			"amount":      amount,
			"method":      method,
			"reference":   reference,
			"description": description,
			"reason":      reason,
		}, reason, ipAddress, "", "FAILED", err.Error())
		return nil, fmt.Errorf("deposit failed: %w", err)
	}

	// Update account balance
	account.Balance = entry.BalanceAfter

	// Create operation record
	now := time.Now()
	operation := &FundOperation{
		ID:          s.nextOpID,
		AccountID:   accountID,
		Type:        "DEPOSIT",
		Amount:      amount,
		Method:      method,
		Reference:   reference,
		Description: description,
		Reason:      reason,
		AdminID:     admin.ID,
		AdminName:   admin.Username,
		Status:      "COMPLETED",
		CreatedAt:   now,
		CompletedAt: &now,
	}
	s.nextOpID++
	s.operations[operation.ID] = operation

	// Log audit
	s.auditLog.Log(admin.ID, admin.Username, "FUND_DEPOSIT", "FUND", accountID, map[string]interface{}{
		"amount":        amount,
		"method":        method,
		"reference":     reference,
		"description":   description,
		"reason":        reason,
		"balanceBefore": entry.BalanceAfter - amount,
		"balanceAfter":  entry.BalanceAfter,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[FundMgmt] DEPOSIT: Account %s +%.2f %s by %s | Balance: %.2f",
		account.AccountNumber, amount, method, admin.Username, entry.BalanceAfter)

	return operation, nil
}

// Withdraw removes funds from a user account
func (s *FundManagementService) Withdraw(accountID int64, amount float64, method, reference, description, reason string, admin *Admin, ipAddress string) (*FundOperation, error) {
	if amount <= 0 {
		return nil, errors.New("withdrawal amount must be positive")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify account exists
	account, ok := s.engine.GetAccount(accountID)
	if !ok {
		return nil, errors.New("account not found")
	}

	if account.Status != "ACTIVE" {
		return nil, errors.New("account is not active")
	}

	// Check if user has open positions
	positions := s.engine.GetPositions(accountID)
	if len(positions) > 0 {
		return nil, fmt.Errorf("cannot withdraw while user has %d open positions", len(positions))
	}

	// Execute withdrawal via ledger
	entry, err := s.engine.GetLedger().Withdraw(accountID, amount, method, reference, description, admin.Username)
	if err != nil {
		s.auditLog.Log(admin.ID, admin.Username, "FUND_WITHDRAW", "FUND", accountID, map[string]interface{}{
			"amount":      amount,
			"method":      method,
			"reference":   reference,
			"description": description,
			"reason":      reason,
		}, reason, ipAddress, "", "FAILED", err.Error())
		return nil, fmt.Errorf("withdrawal failed: %w", err)
	}

	// Update account balance
	account.Balance = entry.BalanceAfter

	// Create operation record
	now := time.Now()
	operation := &FundOperation{
		ID:          s.nextOpID,
		AccountID:   accountID,
		Type:        "WITHDRAW",
		Amount:      amount,
		Method:      method,
		Reference:   reference,
		Description: description,
		Reason:      reason,
		AdminID:     admin.ID,
		AdminName:   admin.Username,
		Status:      "COMPLETED",
		CreatedAt:   now,
		CompletedAt: &now,
	}
	s.nextOpID++
	s.operations[operation.ID] = operation

	// Log audit
	s.auditLog.Log(admin.ID, admin.Username, "FUND_WITHDRAW", "FUND", accountID, map[string]interface{}{
		"amount":        amount,
		"method":        method,
		"reference":     reference,
		"description":   description,
		"reason":        reason,
		"balanceBefore": entry.BalanceAfter + amount,
		"balanceAfter":  entry.BalanceAfter,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[FundMgmt] WITHDRAW: Account %s -%.2f %s by %s | Balance: %.2f",
		account.AccountNumber, amount, method, admin.Username, entry.BalanceAfter)

	return operation, nil
}

// Adjust makes a manual balance adjustment
func (s *FundManagementService) Adjust(accountID int64, amount float64, description, reason string, admin *Admin, ipAddress string) (*FundOperation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify account exists
	account, ok := s.engine.GetAccount(accountID)
	if !ok {
		return nil, errors.New("account not found")
	}

	// Execute adjustment via ledger
	entry, err := s.engine.GetLedger().Adjust(accountID, amount, description, admin.Username)
	if err != nil {
		s.auditLog.Log(admin.ID, admin.Username, "FUND_ADJUST", "FUND", accountID, map[string]interface{}{
			"amount":      amount,
			"description": description,
			"reason":      reason,
		}, reason, ipAddress, "", "FAILED", err.Error())
		return nil, fmt.Errorf("adjustment failed: %w", err)
	}

	// Update account balance
	account.Balance = entry.BalanceAfter

	// Create operation record
	now := time.Now()
	operation := &FundOperation{
		ID:          s.nextOpID,
		AccountID:   accountID,
		Type:        "ADJUSTMENT",
		Amount:      amount,
		Method:      "MANUAL",
		Description: description,
		Reason:      reason,
		AdminID:     admin.ID,
		AdminName:   admin.Username,
		Status:      "COMPLETED",
		CreatedAt:   now,
		CompletedAt: &now,
	}
	s.nextOpID++
	s.operations[operation.ID] = operation

	// Log audit
	s.auditLog.Log(admin.ID, admin.Username, "FUND_ADJUST", "FUND", accountID, map[string]interface{}{
		"amount":        amount,
		"description":   description,
		"reason":        reason,
		"balanceBefore": entry.BalanceAfter - amount,
		"balanceAfter":  entry.BalanceAfter,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[FundMgmt] ADJUSTMENT: Account %s %+.2f by %s: %s | Balance: %.2f",
		account.AccountNumber, amount, admin.Username, description, entry.BalanceAfter)

	return operation, nil
}

// AddBonus adds a bonus to a user account
func (s *FundManagementService) AddBonus(accountID int64, amount float64, description, reason string, admin *Admin, ipAddress string) (*FundOperation, error) {
	if amount <= 0 {
		return nil, errors.New("bonus amount must be positive")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify account exists
	account, ok := s.engine.GetAccount(accountID)
	if !ok {
		return nil, errors.New("account not found")
	}

	if account.Status != "ACTIVE" {
		return nil, errors.New("account is not active")
	}

	// Execute bonus via ledger
	entry, err := s.engine.GetLedger().AddBonus(accountID, amount, description, admin.Username)
	if err != nil {
		s.auditLog.Log(admin.ID, admin.Username, "FUND_BONUS", "FUND", accountID, map[string]interface{}{
			"amount":      amount,
			"description": description,
			"reason":      reason,
		}, reason, ipAddress, "", "FAILED", err.Error())
		return nil, fmt.Errorf("bonus failed: %w", err)
	}

	// Update account balance
	account.Balance = entry.BalanceAfter

	// Create operation record
	now := time.Now()
	operation := &FundOperation{
		ID:          s.nextOpID,
		AccountID:   accountID,
		Type:        "BONUS",
		Amount:      amount,
		Method:      "BONUS",
		Description: description,
		Reason:      reason,
		AdminID:     admin.ID,
		AdminName:   admin.Username,
		Status:      "COMPLETED",
		CreatedAt:   now,
		CompletedAt: &now,
	}
	s.nextOpID++
	s.operations[operation.ID] = operation

	// Log audit
	s.auditLog.Log(admin.ID, admin.Username, "FUND_BONUS", "FUND", accountID, map[string]interface{}{
		"amount":        amount,
		"description":   description,
		"reason":        reason,
		"balanceBefore": entry.BalanceAfter - amount,
		"balanceAfter":  entry.BalanceAfter,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[FundMgmt] BONUS: Account %s +%.2f by %s: %s | Balance: %.2f",
		account.AccountNumber, amount, admin.Username, description, entry.BalanceAfter)

	return operation, nil
}

// GetOperations returns fund operations with optional filters
func (s *FundManagementService) GetOperations(accountID *int64, opType *string, limit int) []*FundOperation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var ops []*FundOperation
	for _, op := range s.operations {
		// Apply filters
		if accountID != nil && op.AccountID != *accountID {
			continue
		}
		if opType != nil && op.Type != *opType {
			continue
		}
		ops = append(ops, op)
	}

	// Sort by created_at descending
	for i := 0; i < len(ops)-1; i++ {
		for j := i + 1; j < len(ops); j++ {
			if ops[i].CreatedAt.Before(ops[j].CreatedAt) {
				ops[i], ops[j] = ops[j], ops[i]
			}
		}
	}

	if limit > 0 && limit < len(ops) {
		return ops[:limit]
	}

	return ops
}

// GetOperation returns a specific operation
func (s *FundManagementService) GetOperation(operationID int64) (*FundOperation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	op, exists := s.operations[operationID]
	if !exists {
		return nil, errors.New("operation not found")
	}

	return op, nil
}
