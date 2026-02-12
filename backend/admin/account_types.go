package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Account Type & Leverage Configuration API
// ============================================
// Manages trading account types with leverage configurations
// and per-symbol leverage overrides.

// AccountType represents a trading account type configuration
type AccountType struct {
	ID               int64              `json:"id"`
	Name             string             `json:"name"`
	Description      string             `json:"description"`
	MinDeposit       float64            `json:"minDeposit"`
	MaxLeverage      int                `json:"maxLeverage"`          // e.g., 500 for 1:500
	SpreadType       string             `json:"spreadType"`           // fixed, variable, raw
	CommissionPerLot float64            `json:"commissionPerLot"`     // Commission per standard lot
	MarginCallLevel  float64            `json:"marginCallLevel"`      // % e.g., 80.0
	StopOutLevel     float64            `json:"stopOutLevel"`         // % e.g., 50.0
	MaxOpenPositions int                `json:"maxOpenPositions"`     // 0 = unlimited
	HedgingAllowed   bool               `json:"hedgingAllowed"`       // Allow hedging positions
	SwapFree         bool               `json:"swapFree"`             // Islamic/swap-free account
	OneClickDefault  bool               `json:"oneClickDefault"`      // Enable one-click trading by default
	Status           string             `json:"status"`               // active, disabled
	LeverageOverrides []LeverageOverride `json:"leverageOverrides"`   // Per-symbol leverage caps
	CreatedAt        time.Time          `json:"createdAt"`
	UpdatedAt        *time.Time         `json:"updatedAt,omitempty"`
	DeletedAt        *time.Time         `json:"deletedAt,omitempty"`  // Soft delete
}

// LeverageOverride represents a per-symbol leverage cap
type LeverageOverride struct {
	Symbol      string `json:"symbol"`      // e.g., "EURUSD", "BTCUSD"
	MaxLeverage int    `json:"maxLeverage"` // Override max leverage for this symbol
}

// AccountTypeService manages account types and leverage configurations
type AccountTypeService struct {
	mu           sync.RWMutex
	accountTypes map[int64]*AccountType
	nextID       int64
}

// NewAccountTypeService creates a new account type service with default account types
func NewAccountTypeService() *AccountTypeService {
	service := &AccountTypeService{
		accountTypes: make(map[int64]*AccountType),
		nextID:       1,
	}

	// Create 6 default account types
	defaultTypes := []struct {
		name             string
		description      string
		minDeposit       float64
		maxLeverage      int
		spreadType       string
		commissionPerLot float64
		marginCallLevel  float64
		stopOutLevel     float64
		maxOpenPositions int
		hedgingAllowed   bool
		swapFree         bool
		oneClickDefault  bool
	}{
		{
			name:             "Demo",
			description:      "Practice account with virtual funds - perfect for testing strategies",
			minDeposit:       0,
			maxLeverage:      500,
			spreadType:       "variable",
			commissionPerLot: 0,
			marginCallLevel:  80.0,
			stopOutLevel:     50.0,
			maxOpenPositions: 0, // unlimited
			hedgingAllowed:   true,
			swapFree:         false,
			oneClickDefault:  true,
		},
		{
			name:             "Micro",
			description:      "Entry-level account with low minimum deposit and high leverage",
			minDeposit:       10,
			maxLeverage:      1000,
			spreadType:       "variable",
			commissionPerLot: 0,
			marginCallLevel:  100.0,
			stopOutLevel:     50.0,
			maxOpenPositions: 50,
			hedgingAllowed:   true,
			swapFree:         false,
			oneClickDefault:  true,
		},
		{
			name:             "Standard",
			description:      "Most popular account type with balanced leverage and spreads",
			minDeposit:       100,
			maxLeverage:      500,
			spreadType:       "variable",
			commissionPerLot: 0,
			marginCallLevel:  100.0,
			stopOutLevel:     50.0,
			maxOpenPositions: 100,
			hedgingAllowed:   true,
			swapFree:         false,
			oneClickDefault:  true,
		},
		{
			name:             "ECN",
			description:      "Professional account with raw spreads and low commission",
			minDeposit:       1000,
			maxLeverage:      200,
			spreadType:       "raw",
			commissionPerLot: 3.50,
			marginCallLevel:  100.0,
			stopOutLevel:     50.0,
			maxOpenPositions: 200,
			hedgingAllowed:   true,
			swapFree:         false,
			oneClickDefault:  false,
		},
		{
			name:             "VIP",
			description:      "Premium account with tighter spreads and dedicated support",
			minDeposit:       10000,
			maxLeverage:      300,
			spreadType:       "variable",
			commissionPerLot: 0,
			marginCallLevel:  120.0,
			stopOutLevel:     60.0,
			maxOpenPositions: 0, // unlimited
			hedgingAllowed:   true,
			swapFree:         false,
			oneClickDefault:  false,
		},
		{
			name:             "Islamic",
			description:      "Sharia-compliant swap-free account for Islamic traders",
			minDeposit:       500,
			maxLeverage:      200,
			spreadType:       "variable",
			commissionPerLot: 0,
			marginCallLevel:  100.0,
			stopOutLevel:     50.0,
			maxOpenPositions: 100,
			hedgingAllowed:   true,
			swapFree:         true,
			oneClickDefault:  true,
		},
	}

	for _, dt := range defaultTypes {
		accountType := &AccountType{
			ID:               service.nextID,
			Name:             dt.name,
			Description:      dt.description,
			MinDeposit:       dt.minDeposit,
			MaxLeverage:      dt.maxLeverage,
			SpreadType:       dt.spreadType,
			CommissionPerLot: dt.commissionPerLot,
			MarginCallLevel:  dt.marginCallLevel,
			StopOutLevel:     dt.stopOutLevel,
			MaxOpenPositions: dt.maxOpenPositions,
			HedgingAllowed:   dt.hedgingAllowed,
			SwapFree:         dt.swapFree,
			OneClickDefault:  dt.oneClickDefault,
			Status:           "active",
			LeverageOverrides: []LeverageOverride{},
			CreatedAt:        time.Now(),
		}

		// Add some default leverage overrides for high-risk symbols
		if dt.name == "ECN" || dt.name == "VIP" {
			accountType.LeverageOverrides = []LeverageOverride{
				{Symbol: "BTCUSD", MaxLeverage: 50},
				{Symbol: "ETHUSD", MaxLeverage: 50},
				{Symbol: "XAUUSD", MaxLeverage: 100},
			}
		}

		service.accountTypes[service.nextID] = accountType
		service.nextID++
	}

	log.Printf("[AccountTypes] Initialized with %d default account types (Demo, Micro, Standard, ECN, VIP, Islamic)", len(defaultTypes))
	return service
}

// ListAccountTypes returns all account types (excluding soft-deleted)
func (ats *AccountTypeService) ListAccountTypes(includeDeleted bool) []*AccountType {
	ats.mu.RLock()
	defer ats.mu.RUnlock()

	types := make([]*AccountType, 0)
	for _, accountType := range ats.accountTypes {
		if !includeDeleted && accountType.DeletedAt != nil {
			continue
		}
		types = append(types, accountType)
	}
	return types
}

// GetAccountType retrieves a specific account type by ID
func (ats *AccountTypeService) GetAccountType(id int64) (*AccountType, error) {
	ats.mu.RLock()
	defer ats.mu.RUnlock()

	accountType, exists := ats.accountTypes[id]
	if !exists {
		return nil, fmt.Errorf("account type %d not found", id)
	}
	if accountType.DeletedAt != nil {
		return nil, fmt.Errorf("account type %d has been deleted", id)
	}
	return accountType, nil
}

// CreateAccountType creates a new account type
func (ats *AccountTypeService) CreateAccountType(accountType *AccountType) (*AccountType, error) {
	ats.mu.Lock()
	defer ats.mu.Unlock()

	// Validate required fields
	if accountType.Name == "" {
		return nil, fmt.Errorf("account type name is required")
	}
	if accountType.MaxLeverage <= 0 {
		return nil, fmt.Errorf("max leverage must be greater than 0")
	}
	if accountType.SpreadType != "fixed" && accountType.SpreadType != "variable" && accountType.SpreadType != "raw" {
		return nil, fmt.Errorf("spread type must be one of: fixed, variable, raw")
	}

	// Assign ID and timestamps
	accountType.ID = ats.nextID
	ats.nextID++
	accountType.CreatedAt = time.Now()
	accountType.Status = "active"

	if accountType.LeverageOverrides == nil {
		accountType.LeverageOverrides = []LeverageOverride{}
	}

	ats.accountTypes[accountType.ID] = accountType
	log.Printf("[AccountTypes] Created new account type: ID=%d, Name=%s, Leverage=1:%d", accountType.ID, accountType.Name, accountType.MaxLeverage)
	return accountType, nil
}

// UpdateAccountType updates an existing account type
func (ats *AccountTypeService) UpdateAccountType(id int64, updates *AccountType) (*AccountType, error) {
	ats.mu.Lock()
	defer ats.mu.Unlock()

	accountType, exists := ats.accountTypes[id]
	if !exists {
		return nil, fmt.Errorf("account type %d not found", id)
	}
	if accountType.DeletedAt != nil {
		return nil, fmt.Errorf("cannot update deleted account type %d", id)
	}

	// Update fields (preserve ID and CreatedAt)
	if updates.Name != "" {
		accountType.Name = updates.Name
	}
	if updates.Description != "" {
		accountType.Description = updates.Description
	}
	if updates.MinDeposit >= 0 {
		accountType.MinDeposit = updates.MinDeposit
	}
	if updates.MaxLeverage > 0 {
		accountType.MaxLeverage = updates.MaxLeverage
	}
	if updates.SpreadType != "" {
		if updates.SpreadType != "fixed" && updates.SpreadType != "variable" && updates.SpreadType != "raw" {
			return nil, fmt.Errorf("spread type must be one of: fixed, variable, raw")
		}
		accountType.SpreadType = updates.SpreadType
	}
	if updates.CommissionPerLot >= 0 {
		accountType.CommissionPerLot = updates.CommissionPerLot
	}
	if updates.MarginCallLevel > 0 {
		accountType.MarginCallLevel = updates.MarginCallLevel
	}
	if updates.StopOutLevel > 0 {
		accountType.StopOutLevel = updates.StopOutLevel
	}
	if updates.MaxOpenPositions >= 0 {
		accountType.MaxOpenPositions = updates.MaxOpenPositions
	}
	if updates.Status != "" {
		accountType.Status = updates.Status
	}

	accountType.HedgingAllowed = updates.HedgingAllowed
	accountType.SwapFree = updates.SwapFree
	accountType.OneClickDefault = updates.OneClickDefault

	now := time.Now()
	accountType.UpdatedAt = &now

	log.Printf("[AccountTypes] Updated account type: ID=%d, Name=%s", id, accountType.Name)
	return accountType, nil
}

// DeleteAccountType soft-deletes an account type
func (ats *AccountTypeService) DeleteAccountType(id int64) error {
	ats.mu.Lock()
	defer ats.mu.Unlock()

	accountType, exists := ats.accountTypes[id]
	if !exists {
		return fmt.Errorf("account type %d not found", id)
	}
	if accountType.DeletedAt != nil {
		return fmt.Errorf("account type %d already deleted", id)
	}

	now := time.Now()
	accountType.DeletedAt = &now
	accountType.Status = "disabled"

	log.Printf("[AccountTypes] Soft-deleted account type: ID=%d, Name=%s", id, accountType.Name)
	return nil
}

// GetLeverageOverrides returns leverage overrides for an account type
func (ats *AccountTypeService) GetLeverageOverrides(id int64) ([]LeverageOverride, error) {
	accountType, err := ats.GetAccountType(id)
	if err != nil {
		return nil, err
	}
	return accountType.LeverageOverrides, nil
}

// UpdateLeverageOverrides updates the leverage overrides for an account type
func (ats *AccountTypeService) UpdateLeverageOverrides(id int64, overrides []LeverageOverride) error {
	ats.mu.Lock()
	defer ats.mu.Unlock()

	accountType, exists := ats.accountTypes[id]
	if !exists {
		return fmt.Errorf("account type %d not found", id)
	}
	if accountType.DeletedAt != nil {
		return fmt.Errorf("cannot update leverage for deleted account type %d", id)
	}

	// Validate overrides
	for _, override := range overrides {
		if override.Symbol == "" {
			return fmt.Errorf("symbol is required for leverage override")
		}
		if override.MaxLeverage <= 0 {
			return fmt.Errorf("max leverage must be greater than 0 for symbol %s", override.Symbol)
		}
	}

	accountType.LeverageOverrides = overrides
	now := time.Now()
	accountType.UpdatedAt = &now

	log.Printf("[AccountTypes] Updated leverage overrides for account type %d: %d symbols configured", id, len(overrides))
	return nil
}

// ============================================
// HTTP Handlers
// ============================================

// AccountTypeHandler handles HTTP requests for account types
type AccountTypeHandler struct {
	service     *AccountTypeService
	authService *auth.Service
}

// NewAccountTypeHandler creates a new account type handler
func NewAccountTypeHandler(service *AccountTypeService, authService *auth.Service) *AccountTypeHandler {
	return &AccountTypeHandler{
		service:     service,
		authService: authService,
	}
}

// ListAccountTypes handles GET /admin/account-types
func (ath *AccountTypeHandler) ListAccountTypes(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accountTypes := ath.service.ListAccountTypes(false)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"accountTypes": accountTypes,
		"count":        len(accountTypes),
	})
}

// GetAccountType handles GET /admin/account-types/:id
func (ath *AccountTypeHandler) GetAccountType(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	idStr := parts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid account type ID", http.StatusBadRequest)
		return
	}

	accountType, err := ath.service.GetAccountType(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accountType)
}

// CreateAccountType handles POST /admin/account-types
func (ath *AccountTypeHandler) CreateAccountType(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var accountType AccountType
	if err := json.NewDecoder(r.Body).Decode(&accountType); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	created, err := ath.service.CreateAccountType(&accountType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// UpdateAccountType handles PUT /admin/account-types/:id
func (ath *AccountTypeHandler) UpdateAccountType(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	idStr := parts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid account type ID", http.StatusBadRequest)
		return
	}

	var updates AccountType
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := ath.service.UpdateAccountType(id, &updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// DeleteAccountType handles DELETE /admin/account-types/:id
func (ath *AccountTypeHandler) DeleteAccountType(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	idStr := parts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid account type ID", http.StatusBadRequest)
		return
	}

	if err := ath.service.DeleteAccountType(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Account type deleted successfully",
		"id":      id,
	})
}

// GetLeverageOverrides handles GET /admin/account-types/:id/leverage
func (ath *AccountTypeHandler) GetLeverageOverrides(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	idStr := parts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid account type ID", http.StatusBadRequest)
		return
	}

	overrides, err := ath.service.GetLeverageOverrides(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"accountTypeId":     id,
		"leverageOverrides": overrides,
		"count":             len(overrides),
	})
}

// UpdateLeverageOverrides handles PUT /admin/account-types/:id/leverage
func (ath *AccountTypeHandler) UpdateLeverageOverrides(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	idStr := parts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid account type ID", http.StatusBadRequest)
		return
	}

	var request struct {
		LeverageOverrides []LeverageOverride `json:"leverageOverrides"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ath.service.UpdateLeverageOverrides(id, request.LeverageOverrides); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":           "Leverage overrides updated successfully",
		"accountTypeId":     id,
		"leverageOverrides": request.LeverageOverrides,
	})
}
