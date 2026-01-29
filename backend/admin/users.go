package admin

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
)

// UserManagementService handles admin user operations
type UserManagementService struct {
	mu          sync.RWMutex
	engine      *core.Engine
	authService *AuthService
	auditLog    *AuditLog
	userGroups  map[int64]int64 // accountID -> groupID
	userEmails  map[int64]string
	lastLogins  map[int64]time.Time
}

// NewUserManagementService creates a new user management service
func NewUserManagementService(engine *core.Engine, authService *AuthService, auditLog *AuditLog) *UserManagementService {
	return &UserManagementService{
		engine:      engine,
		authService: authService,
		auditLog:    auditLog,
		userGroups:  make(map[int64]int64),
		userEmails:  make(map[int64]string),
		lastLogins:  make(map[int64]time.Time),
	}
}

// GetAllUsers returns all user accounts with detailed info
func (s *UserManagementService) GetAllUsers() ([]*UserAccountInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var users []*UserAccountInfo

	// Iterate through accounts (assuming IDs 1-1000)
	for i := int64(1); i <= 1000; i++ {
		account, ok := s.engine.GetAccount(i)
		if !ok {
			continue
		}

		summary, err := s.engine.GetAccountSummary(i)
		if err != nil {
			log.Printf("[UserMgmt] Failed to get summary for account %d: %v", i, err)
			continue
		}

		// Get trades to calculate stats
		trades := s.engine.GetTrades(i)
		totalVolume := 0.0
		totalPnL := 0.0
		for _, trade := range trades {
			totalVolume += trade.Volume
			totalPnL += trade.RealizedPnL
		}

		userInfo := &UserAccountInfo{
			ID:            account.ID,
			AccountNumber: account.AccountNumber,
			UserID:        account.UserID,
			Username:      account.Username,
			Email:         s.userEmails[account.ID],
			Balance:       summary.Balance,
			Equity:        summary.Equity,
			Margin:        summary.Margin,
			FreeMargin:    summary.FreeMargin,
			Leverage:      account.Leverage,
			GroupID:       s.userGroups[account.ID],
			Status:        account.Status,
			IsDemo:        account.IsDemo,
			CreatedAt:     time.Unix(account.CreatedAt, 0),
			OpenPositions: summary.OpenPositions,
			TotalVolume:   totalVolume,
			TotalPnL:      totalPnL,
		}

		if lastLogin, ok := s.lastLogins[account.ID]; ok {
			userInfo.LastLogin = &lastLogin
		}

		users = append(users, userInfo)
	}

	return users, nil
}

// GetUser returns detailed info for a specific user
func (s *UserManagementService) GetUser(accountID int64) (*UserAccountInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, ok := s.engine.GetAccount(accountID)
	if !ok {
		return nil, errors.New("account not found")
	}

	summary, err := s.engine.GetAccountSummary(accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account summary: %w", err)
	}

	trades := s.engine.GetTrades(accountID)
	totalVolume := 0.0
	totalPnL := 0.0
	for _, trade := range trades {
		totalVolume += trade.Volume
		totalPnL += trade.RealizedPnL
	}

	userInfo := &UserAccountInfo{
		ID:            account.ID,
		AccountNumber: account.AccountNumber,
		UserID:        account.UserID,
		Username:      account.Username,
		Email:         s.userEmails[account.ID],
		Balance:       summary.Balance,
		Equity:        summary.Equity,
		Margin:        summary.Margin,
		FreeMargin:    summary.FreeMargin,
		Leverage:      account.Leverage,
		GroupID:       s.userGroups[account.ID],
		Status:        account.Status,
		IsDemo:        account.IsDemo,
		CreatedAt:     time.Unix(account.CreatedAt, 0),
		OpenPositions: summary.OpenPositions,
		TotalVolume:   totalVolume,
		TotalPnL:      totalPnL,
	}

	if lastLogin, ok := s.lastLogins[account.ID]; ok {
		userInfo.LastLogin = &lastLogin
	}

	return userInfo, nil
}

// UpdateUserAccount updates user account settings
func (s *UserManagementService) UpdateUserAccount(accountID int64, leverage *float64, marginMode *string, groupID *int64, email *string, admin *Admin, reason string, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, ok := s.engine.GetAccount(accountID)
	if !ok {
		return errors.New("account not found")
	}

	changes := make(map[string]interface{})
	oldValues := make(map[string]interface{})

	// Update leverage
	if leverage != nil && *leverage != account.Leverage {
		oldValues["leverage"] = account.Leverage
		changes["leverage"] = *leverage

		if err := s.engine.UpdateAccount(accountID, *leverage, ""); err != nil {
			return fmt.Errorf("failed to update leverage: %w", err)
		}
	}

	// Update margin mode
	if marginMode != nil && *marginMode != account.MarginMode {
		oldValues["marginMode"] = account.MarginMode
		changes["marginMode"] = *marginMode

		if err := s.engine.UpdateAccount(accountID, 0, *marginMode); err != nil {
			return fmt.Errorf("failed to update margin mode: %w", err)
		}
	}

	// Update group
	if groupID != nil && s.userGroups[accountID] != *groupID {
		oldValues["groupID"] = s.userGroups[accountID]
		changes["groupID"] = *groupID
		s.userGroups[accountID] = *groupID
	}

	// Update email
	if email != nil && s.userEmails[accountID] != *email {
		oldValues["email"] = s.userEmails[accountID]
		changes["email"] = *email
		s.userEmails[accountID] = *email
	}

	// Log audit
	s.auditLog.Log(admin.ID, admin.Username, "USER_UPDATE", "USER", accountID, map[string]interface{}{
		"old":    oldValues,
		"new":    changes,
		"reason": reason,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[UserMgmt] Account %s updated by %s: %v", account.AccountNumber, admin.Username, changes)

	return nil
}

// EnableUserAccount enables a disabled account
func (s *UserManagementService) EnableUserAccount(accountID int64, admin *Admin, reason string, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, ok := s.engine.GetAccount(accountID)
	if !ok {
		return errors.New("account not found")
	}

	if account.Status == "ACTIVE" {
		return errors.New("account is already active")
	}

	oldStatus := account.Status
	account.Status = "ACTIVE"

	s.auditLog.Log(admin.ID, admin.Username, "USER_ENABLE", "USER", accountID, map[string]interface{}{
		"oldStatus": oldStatus,
		"newStatus": "ACTIVE",
		"reason":    reason,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[UserMgmt] Account %s enabled by %s: %s", account.AccountNumber, admin.Username, reason)

	return nil
}

// DisableUserAccount disables an account
func (s *UserManagementService) DisableUserAccount(accountID int64, admin *Admin, reason string, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, ok := s.engine.GetAccount(accountID)
	if !ok {
		return errors.New("account not found")
	}

	if account.Status == "DISABLED" {
		return errors.New("account is already disabled")
	}

	// Check if user has open positions
	positions := s.engine.GetPositions(accountID)
	if len(positions) > 0 {
		return fmt.Errorf("cannot disable account with %d open positions", len(positions))
	}

	oldStatus := account.Status
	account.Status = "DISABLED"

	s.auditLog.Log(admin.ID, admin.Username, "USER_DISABLE", "USER", accountID, map[string]interface{}{
		"oldStatus": oldStatus,
		"newStatus": "DISABLED",
		"reason":    reason,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[UserMgmt] Account %s disabled by %s: %s", account.AccountNumber, admin.Username, reason)

	return nil
}

// ResetUserPassword resets a user's password
func (s *UserManagementService) ResetUserPassword(accountID int64, newPassword string, admin *Admin, reason string, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, ok := s.engine.GetAccount(accountID)
	if !ok {
		return errors.New("account not found")
	}

	if err := s.engine.UpdatePassword(accountID, newPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.auditLog.Log(admin.ID, admin.Username, "USER_PASSWORD_RESET", "USER", accountID, map[string]interface{}{
		"reason": reason,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[UserMgmt] Password reset for %s by %s", account.AccountNumber, admin.Username)

	return nil
}

// GetUserTradingHistory returns trading history for a user
func (s *UserManagementService) GetUserTradingHistory(accountID int64) ([]core.Trade, error) {
	trades := s.engine.GetTrades(accountID)
	return trades, nil
}

// GetUserPositions returns open positions for a user
func (s *UserManagementService) GetUserPositions(accountID int64) ([]*core.Position, error) {
	positions := s.engine.GetPositions(accountID)
	return positions, nil
}

// GetUserOrders returns orders for a user
func (s *UserManagementService) GetUserOrders(accountID int64, status string) ([]*core.Order, error) {
	orders := s.engine.GetOrders(accountID, status)
	return orders, nil
}

// SetUserEmail sets email for a user
func (s *UserManagementService) SetUserEmail(accountID int64, email string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userEmails[accountID] = email
}

// RecordLogin records user login time
func (s *UserManagementService) RecordLogin(accountID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastLogins[accountID] = time.Now()
}

// AssignUserToGroup assigns a user to a trading group
func (s *UserManagementService) AssignUserToGroup(accountID, groupID int64, admin *Admin, reason string, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, ok := s.engine.GetAccount(accountID)
	if !ok {
		return errors.New("account not found")
	}

	oldGroupID := s.userGroups[accountID]
	s.userGroups[accountID] = groupID

	s.auditLog.Log(admin.ID, admin.Username, "USER_GROUP_ASSIGN", "USER", accountID, map[string]interface{}{
		"oldGroupID": oldGroupID,
		"newGroupID": groupID,
		"reason":     reason,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[UserMgmt] Account %s assigned to group %d by %s", account.AccountNumber, groupID, admin.Username)

	return nil
}
