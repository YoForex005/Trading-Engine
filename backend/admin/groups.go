package admin

import (
	"errors"
	"log"
	"sync"
	"time"
)

// GroupManagementService handles trading group operations
type GroupManagementService struct {
	mu         sync.RWMutex
	groups     map[int64]*UserGroup
	auditLog   *AuditLog
	nextGroupID int64
}

// NewGroupManagementService creates a new group management service
func NewGroupManagementService(auditLog *AuditLog) *GroupManagementService {
	svc := &GroupManagementService{
		groups:      make(map[int64]*UserGroup),
		auditLog:    auditLog,
		nextGroupID: 1,
	}

	// Create default groups
	svc.createDefaultGroups()

	return svc
}

func (s *GroupManagementService) createDefaultGroups() {
	// Standard Group
	s.groups[1] = &UserGroup{
		ID:             1,
		Name:           "Standard",
		Description:    "Default group for standard clients",
		ExecutionMode:  "BBOOK",
		Markup:         0.5,  // 0.5 pip markup
		Commission:     5.0,  // $5 per lot
		MaxLeverage:    100.0,
		EnabledSymbols: []string{"EURUSD", "GBPUSD", "USDJPY", "BTCUSD", "ETHUSD"},
		SymbolSettings: make(map[string]SymbolGroupSettings),
		DefaultBalance: 5000.0,
		MarginMode:     "HEDGING",
		Status:         "ACTIVE",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      "SYSTEM",
	}

	// Premium Group
	s.groups[2] = &UserGroup{
		ID:             2,
		Name:           "Premium",
		Description:    "Premium group with lower spreads and higher leverage",
		ExecutionMode:  "ABOOK",
		Markup:         0.2,  // 0.2 pip markup
		Commission:     3.0,  // $3 per lot
		MaxLeverage:    500.0,
		EnabledSymbols: []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD", "BTCUSD", "ETHUSD", "BNBUSD"},
		SymbolSettings: make(map[string]SymbolGroupSettings),
		DefaultBalance: 10000.0,
		MarginMode:     "HEDGING",
		Status:         "ACTIVE",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      "SYSTEM",
	}

	// VIP Group
	s.groups[3] = &UserGroup{
		ID:             3,
		Name:           "VIP",
		Description:    "VIP group with raw spreads and commission-only pricing",
		ExecutionMode:  "ABOOK",
		Markup:         0.0,  // Raw spreads
		Commission:     2.0,  // $2 per lot
		MaxLeverage:    1000.0,
		EnabledSymbols: []string{"*"}, // All symbols
		SymbolSettings: make(map[string]SymbolGroupSettings),
		DefaultBalance: 50000.0,
		MarginMode:     "HEDGING",
		Status:         "ACTIVE",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      "SYSTEM",
	}

	s.nextGroupID = 4

	log.Println("[GroupMgmt] Default groups created: Standard, Premium, VIP")
}

// CreateGroup creates a new trading group
func (s *GroupManagementService) CreateGroup(name, description, executionMode string, markup, commission, maxLeverage, defaultBalance float64, enabledSymbols []string, marginMode string, admin *Admin, ipAddress string) (*UserGroup, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate execution mode
	if executionMode != "BBOOK" && executionMode != "ABOOK" && executionMode != "HYBRID" {
		return nil, errors.New("execution mode must be BBOOK, ABOOK, or HYBRID")
	}

	// Validate margin mode
	if marginMode != "HEDGING" && marginMode != "NETTING" {
		return nil, errors.New("margin mode must be HEDGING or NETTING")
	}

	// Check if group name exists
	for _, g := range s.groups {
		if g.Name == name {
			return nil, errors.New("group name already exists")
		}
	}

	group := &UserGroup{
		ID:             s.nextGroupID,
		Name:           name,
		Description:    description,
		ExecutionMode:  executionMode,
		Markup:         markup,
		Commission:     commission,
		MaxLeverage:    maxLeverage,
		EnabledSymbols: enabledSymbols,
		SymbolSettings: make(map[string]SymbolGroupSettings),
		DefaultBalance: defaultBalance,
		MarginMode:     marginMode,
		Status:         "ACTIVE",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      admin.Username,
	}

	s.nextGroupID++
	s.groups[group.ID] = group

	// Log audit
	s.auditLog.Log(admin.ID, admin.Username, "GROUP_CREATE", "GROUP", group.ID, map[string]interface{}{
		"name":           name,
		"executionMode":  executionMode,
		"markup":         markup,
		"commission":     commission,
		"maxLeverage":    maxLeverage,
		"defaultBalance": defaultBalance,
		"marginMode":     marginMode,
	}, "", ipAddress, "", "SUCCESS", "")

	log.Printf("[GroupMgmt] Group created: %s (%s) by %s", name, executionMode, admin.Username)

	return group, nil
}

// UpdateGroup updates an existing group
func (s *GroupManagementService) UpdateGroup(groupID int64, name, description, executionMode *string, markup, commission, maxLeverage, defaultBalance *float64, enabledSymbols []string, marginMode *string, admin *Admin, reason string, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	group, exists := s.groups[groupID]
	if !exists {
		return errors.New("group not found")
	}

	changes := make(map[string]interface{})
	oldValues := make(map[string]interface{})

	if name != nil && *name != group.Name {
		oldValues["name"] = group.Name
		changes["name"] = *name
		group.Name = *name
	}

	if description != nil && *description != group.Description {
		oldValues["description"] = group.Description
		changes["description"] = *description
		group.Description = *description
	}

	if executionMode != nil && *executionMode != group.ExecutionMode {
		if *executionMode != "BBOOK" && *executionMode != "ABOOK" && *executionMode != "HYBRID" {
			return errors.New("invalid execution mode")
		}
		oldValues["executionMode"] = group.ExecutionMode
		changes["executionMode"] = *executionMode
		group.ExecutionMode = *executionMode
	}

	if markup != nil && *markup != group.Markup {
		oldValues["markup"] = group.Markup
		changes["markup"] = *markup
		group.Markup = *markup
	}

	if commission != nil && *commission != group.Commission {
		oldValues["commission"] = group.Commission
		changes["commission"] = *commission
		group.Commission = *commission
	}

	if maxLeverage != nil && *maxLeverage != group.MaxLeverage {
		oldValues["maxLeverage"] = group.MaxLeverage
		changes["maxLeverage"] = *maxLeverage
		group.MaxLeverage = *maxLeverage
	}

	if defaultBalance != nil && *defaultBalance != group.DefaultBalance {
		oldValues["defaultBalance"] = group.DefaultBalance
		changes["defaultBalance"] = *defaultBalance
		group.DefaultBalance = *defaultBalance
	}

	if marginMode != nil && *marginMode != group.MarginMode {
		if *marginMode != "HEDGING" && *marginMode != "NETTING" {
			return errors.New("invalid margin mode")
		}
		oldValues["marginMode"] = group.MarginMode
		changes["marginMode"] = *marginMode
		group.MarginMode = *marginMode
	}

	if enabledSymbols != nil {
		oldValues["enabledSymbols"] = group.EnabledSymbols
		changes["enabledSymbols"] = enabledSymbols
		group.EnabledSymbols = enabledSymbols
	}

	group.UpdatedAt = time.Now()

	// Log audit
	s.auditLog.Log(admin.ID, admin.Username, "GROUP_UPDATE", "GROUP", groupID, map[string]interface{}{
		"old":    oldValues,
		"new":    changes,
		"reason": reason,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[GroupMgmt] Group %s updated by %s: %v", group.Name, admin.Username, changes)

	return nil
}

// DeleteGroup deletes a group
func (s *GroupManagementService) DeleteGroup(groupID int64, admin *Admin, reason string, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	group, exists := s.groups[groupID]
	if !exists {
		return errors.New("group not found")
	}

	// Cannot delete default groups (1, 2, 3)
	if groupID <= 3 {
		return errors.New("cannot delete default groups")
	}

	delete(s.groups, groupID)

	// Log audit
	s.auditLog.Log(admin.ID, admin.Username, "GROUP_DELETE", "GROUP", groupID, map[string]interface{}{
		"name":   group.Name,
		"reason": reason,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[GroupMgmt] Group %s deleted by %s: %s", group.Name, admin.Username, reason)

	return nil
}

// GetGroup returns a specific group
func (s *GroupManagementService) GetGroup(groupID int64) (*UserGroup, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	group, exists := s.groups[groupID]
	if !exists {
		return nil, errors.New("group not found")
	}

	return group, nil
}

// ListGroups returns all groups
func (s *GroupManagementService) ListGroups() []*UserGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()

	groups := make([]*UserGroup, 0, len(s.groups))
	for _, group := range s.groups {
		groups = append(groups, group)
	}

	return groups
}

// SetSymbolSettings sets per-symbol settings for a group
func (s *GroupManagementService) SetSymbolSettings(groupID int64, symbol string, settings SymbolGroupSettings, admin *Admin, reason string, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	group, exists := s.groups[groupID]
	if !exists {
		return errors.New("group not found")
	}

	oldSettings := group.SymbolSettings[symbol]
	group.SymbolSettings[symbol] = settings
	group.UpdatedAt = time.Now()

	// Log audit
	s.auditLog.Log(admin.ID, admin.Username, "GROUP_SYMBOL_UPDATE", "GROUP", groupID, map[string]interface{}{
		"symbol": symbol,
		"old":    oldSettings,
		"new":    settings,
		"reason": reason,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[GroupMgmt] Symbol settings for %s in group %s updated by %s", symbol, group.Name, admin.Username)

	return nil
}

// EnableGroup enables a disabled group
func (s *GroupManagementService) EnableGroup(groupID int64, admin *Admin, reason string, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	group, exists := s.groups[groupID]
	if !exists {
		return errors.New("group not found")
	}

	if group.Status == "ACTIVE" {
		return errors.New("group is already active")
	}

	oldStatus := group.Status
	group.Status = "ACTIVE"
	group.UpdatedAt = time.Now()

	s.auditLog.Log(admin.ID, admin.Username, "GROUP_ENABLE", "GROUP", groupID, map[string]interface{}{
		"oldStatus": oldStatus,
		"newStatus": "ACTIVE",
		"reason":    reason,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[GroupMgmt] Group %s enabled by %s", group.Name, admin.Username)

	return nil
}

// DisableGroup disables a group
func (s *GroupManagementService) DisableGroup(groupID int64, admin *Admin, reason string, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	group, exists := s.groups[groupID]
	if !exists {
		return errors.New("group not found")
	}

	// Cannot disable default groups
	if groupID <= 3 {
		return errors.New("cannot disable default groups")
	}

	if group.Status == "DISABLED" {
		return errors.New("group is already disabled")
	}

	oldStatus := group.Status
	group.Status = "DISABLED"
	group.UpdatedAt = time.Now()

	s.auditLog.Log(admin.ID, admin.Username, "GROUP_DISABLE", "GROUP", groupID, map[string]interface{}{
		"oldStatus": oldStatus,
		"newStatus": "DISABLED",
		"reason":    reason,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[GroupMgmt] Group %s disabled by %s: %s", group.Name, admin.Username, reason)

	return nil
}

// GetGroupStats returns statistics for a group
func (s *GroupManagementService) GetGroupStats(groupID int64, userMgmt *UserManagementService) (map[string]interface{}, error) {
	s.mu.RLock()
	group, exists := s.groups[groupID]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("group not found")
	}

	// Get all users in this group
	allUsers, _ := userMgmt.GetAllUsers()

	userCount := 0
	totalBalance := 0.0
	totalEquity := 0.0
	totalVolume := 0.0
	totalPnL := 0.0

	for _, user := range allUsers {
		if user.GroupID == groupID {
			userCount++
			totalBalance += user.Balance
			totalEquity += user.Equity
			totalVolume += user.TotalVolume
			totalPnL += user.TotalPnL
		}
	}

	return map[string]interface{}{
		"groupID":      group.ID,
		"groupName":    group.Name,
		"userCount":    userCount,
		"totalBalance": totalBalance,
		"totalEquity":  totalEquity,
		"totalVolume":  totalVolume,
		"totalPnL":     totalPnL,
	}, nil
}
