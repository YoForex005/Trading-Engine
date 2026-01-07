package bbook

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// APIHandler handles B-Book API requests
type APIHandler struct {
	engine    *Engine
	pnlEngine *PnLEngine
}

// NewAPIHandler creates API handlers for B-Book
func NewAPIHandler(engine *Engine, pnlEngine *PnLEngine) *APIHandler {
	return &APIHandler{
		engine:    engine,
		pnlEngine: pnlEngine,
	}
}

// cors adds CORS headers
func cors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
}

// HandleGetAccountSummary returns account balance/equity/margin
func (h *APIHandler) HandleGetAccountSummary(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get account ID from query or default to 1
	accountID := int64(1)
	if id := r.URL.Query().Get("accountId"); id != "" {
		if parsed, err := strconv.ParseInt(id, 10, 64); err == nil {
			accountID = parsed
		}
	}

	summary, err := h.engine.GetAccountSummary(accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// HandleGetPositions returns open positions
func (h *APIHandler) HandleGetPositions(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	accountID := int64(1)
	if id := r.URL.Query().Get("accountId"); id != "" {
		if parsed, err := strconv.ParseInt(id, 10, 64); err == nil {
			accountID = parsed
		}
	}

	positions := h.engine.GetPositions(accountID)
	if positions == nil {
		positions = []*Position{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(positions)
}

// HandleGetOrders returns orders
func (h *APIHandler) HandleGetOrders(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	accountID := int64(1)
	if id := r.URL.Query().Get("accountId"); id != "" {
		if parsed, err := strconv.ParseInt(id, 10, 64); err == nil {
			accountID = parsed
		}
	}

	status := r.URL.Query().Get("status")
	orders := h.engine.GetOrders(accountID, status)
	if orders == nil {
		orders = []*Order{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// HandlePlaceMarketOrder executes a market order
func (h *APIHandler) HandlePlaceMarketOrder(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		AccountID int64   `json:"accountId"`
		Symbol    string  `json:"symbol"`
		Side      string  `json:"side"`
		Volume    float64 `json:"volume"`
		SL        float64 `json:"sl,omitempty"`
		TP        float64 `json:"tp,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AccountID == 0 {
		req.AccountID = 1 // Default account
	}

	position, err := h.engine.ExecuteMarketOrder(req.AccountID, req.Symbol, req.Side, req.Volume, req.SL, req.TP)
	if err != nil {
		log.Printf("[API] Order rejected: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Force P/L update
	if h.pnlEngine != nil {
		h.pnlEngine.ForceUpdate()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"position": position,
	})
}

// HandleClosePosition closes a position
func (h *APIHandler) HandleClosePosition(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		PositionID int64   `json:"positionId"`
		Volume     float64 `json:"volume,omitempty"` // 0 = close all
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Also check URL path for position ID
	if req.PositionID == 0 {
		parts := strings.Split(r.URL.Path, "/")
		for i, p := range parts {
			if p == "positions" && i+1 < len(parts) {
				if id, err := strconv.ParseInt(parts[i+1], 10, 64); err == nil {
					req.PositionID = id
				}
			}
		}
	}

	trade, err := h.engine.ClosePosition(req.PositionID, req.Volume)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Force P/L update
	if h.pnlEngine != nil {
		h.pnlEngine.ForceUpdate()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"trade":   trade,
	})
}

// HandleCloseBulk closes multiple positions based on filter
func (h *APIHandler) HandleCloseBulk(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		AccountID int64  `json:"accountId"`
		Type      string `json:"type"`             // ALL, WINNERS, LOSERS
		Symbol    string `json:"symbol,omitempty"` // Optional: limit to one symbol
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AccountID == 0 {
		req.AccountID = 1 // Default account
	}

	positions := h.engine.GetPositions(req.AccountID)

	// Update prices to ensure P/L is fresh
	h.engine.UpdatePositionPrices()

	closedCount := 0
	var errors []string

	for _, pos := range positions {
		// Filter by symbol if provided
		if req.Symbol != "" && pos.Symbol != req.Symbol {
			continue
		}

		shouldClose := false
		switch req.Type {
		case "ALL":
			shouldClose = true
		case "WINNERS":
			shouldClose = pos.UnrealizedPnL > 0
		case "LOSERS":
			shouldClose = pos.UnrealizedPnL < 0
		default:
			// Invalid type
			http.Error(w, "Invalid Type: must be ALL, WINNERS, or LOSERS", http.StatusBadRequest)
			return
		}

		if shouldClose {
			_, err := h.engine.ClosePosition(pos.ID, 0) // 0 volume = full close
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to close position %d: %v", pos.ID, err))
			} else {
				closedCount++
			}
		}
	}

	// Force P/L update
	if h.pnlEngine != nil {
		h.pnlEngine.ForceUpdate()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"closedCount": closedCount,
		"errors":      errors,
	})
}

// HandleModifyPosition modifes SL/TP
func (h *APIHandler) HandleModifyPosition(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		PositionID int64   `json:"positionId"`
		SL         float64 `json:"sl"`
		TP         float64 `json:"tp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Also check URL if PositionID is 0
	if req.PositionID == 0 {
		parts := strings.Split(r.URL.Path, "/")
		for i, p := range parts {
			if p == "positions" && i+1 < len(parts) {
				if id, err := strconv.ParseInt(parts[i+1], 10, 64); err == nil {
					req.PositionID = id
				}
			}
		}
	}

	position, err := h.engine.ModifyPosition(req.PositionID, req.SL, req.TP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"position": position,
	})
}

// HandleGetTrades returns trade history
func (h *APIHandler) HandleGetTrades(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	accountID := int64(1)
	if id := r.URL.Query().Get("accountId"); id != "" {
		if parsed, err := strconv.ParseInt(id, 10, 64); err == nil {
			accountID = parsed
		}
	}

	trades := h.engine.GetTrades(accountID)
	if trades == nil {
		trades = []Trade{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trades)
}

// HandleGetLedger returns ledger history
func (h *APIHandler) HandleGetLedger(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	accountID := int64(1)
	if id := r.URL.Query().Get("accountId"); id != "" {
		if parsed, err := strconv.ParseInt(id, 10, 64); err == nil {
			accountID = parsed
		}
	}

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	entries := h.engine.GetLedger().GetHistory(accountID, limit)
	if entries == nil {
		entries = []LedgerEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// ===== ADMIN ENDPOINTS =====

// HandleAdminDeposit adds funds to an account
func (h *APIHandler) HandleAdminDeposit(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		AccountID   int64   `json:"accountId"`
		Amount      float64 `json:"amount"`
		Method      string  `json:"method"` // BANK/CRYPTO/CARD/MANUAL
		Reference   string  `json:"reference"`
		Description string  `json:"description"`
		AdminID     string  `json:"adminId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Method == "" {
		req.Method = "MANUAL"
	}
	if req.Description == "" {
		req.Description = "Deposit via " + req.Method
	}

	// Get account to update balance
	account, ok := h.engine.GetAccount(req.AccountID)
	if !ok {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	entry, err := h.engine.GetLedger().Deposit(req.AccountID, req.Amount, req.Method, req.Reference, req.Description, req.AdminID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update account balance
	account.Balance = entry.BalanceAfter

	log.Printf("[ADMIN] Deposit: Account #%d +%.2f %s by %s", req.AccountID, req.Amount, req.Method, req.AdminID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"entry":      entry,
		"newBalance": entry.BalanceAfter,
	})
}

// HandleAdminWithdraw removes funds from an account
func (h *APIHandler) HandleAdminWithdraw(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		AccountID   int64   `json:"accountId"`
		Amount      float64 `json:"amount"`
		Method      string  `json:"method"` // BANK/CRYPTO
		Reference   string  `json:"reference"`
		Description string  `json:"description"`
		AdminID     string  `json:"adminId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Method == "" {
		req.Method = "MANUAL"
	}
	if req.Description == "" {
		req.Description = "Withdrawal via " + req.Method
	}

	// Get account to update balance
	account, ok := h.engine.GetAccount(req.AccountID)
	if !ok {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	entry, err := h.engine.GetLedger().Withdraw(req.AccountID, req.Amount, req.Method, req.Reference, req.Description, req.AdminID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update account balance
	account.Balance = entry.BalanceAfter

	log.Printf("[ADMIN] Withdraw: Account #%d -%.2f %s by %s", req.AccountID, req.Amount, req.Method, req.AdminID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"entry":      entry,
		"newBalance": entry.BalanceAfter,
	})
}

// HandleAdminAdjust makes a balance adjustment
func (h *APIHandler) HandleAdminAdjust(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		AccountID   int64   `json:"accountId"`
		Amount      float64 `json:"amount"` // Can be negative
		Description string  `json:"description"`
		AdminID     string  `json:"adminId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get account to update balance
	account, ok := h.engine.GetAccount(req.AccountID)
	if !ok {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	entry, err := h.engine.GetLedger().Adjust(req.AccountID, req.Amount, req.Description, req.AdminID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update account balance
	account.Balance = entry.BalanceAfter

	log.Printf("[ADMIN] Adjustment: Account #%d %+.2f by %s: %s", req.AccountID, req.Amount, req.AdminID, req.Description)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"entry":      entry,
		"newBalance": entry.BalanceAfter,
	})
}

// HandleAdminBonus adds a bonus
func (h *APIHandler) HandleAdminBonus(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		AccountID   int64   `json:"accountId"`
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
		AdminID     string  `json:"adminId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Description == "" {
		req.Description = "Bonus"
	}

	// Get account to update balance
	account, ok := h.engine.GetAccount(req.AccountID)
	if !ok {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	entry, err := h.engine.GetLedger().AddBonus(req.AccountID, req.Amount, req.Description, req.AdminID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update account balance
	account.Balance = entry.BalanceAfter

	log.Printf("[ADMIN] Bonus: Account #%d +%.2f by %s", req.AccountID, req.Amount, req.AdminID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"entry":      entry,
		"newBalance": entry.BalanceAfter,
	})
}

// HandleAdminGetAccounts returns all accounts
func (h *APIHandler) HandleAdminGetAccounts(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var accounts []*Account
	// Get all accounts from engine
	for i := int64(1); i <= 100; i++ {
		acc, ok := h.engine.GetAccount(i)
		if ok {
			accounts = append(accounts, acc)
		}
	}

	if accounts == nil {
		accounts = []*Account{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

// HandleAdminGetLedgerAll returns all ledger entries
func (h *APIHandler) HandleAdminGetLedgerAll(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	limit := 500
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	typeFilter := r.URL.Query().Get("type")

	var entries []LedgerEntry
	if typeFilter != "" {
		entries = h.engine.GetLedger().GetEntriesByType(typeFilter, limit)
	} else {
		entries = h.engine.GetLedger().GetAllEntries(limit)
	}

	if entries == nil {
		entries = []LedgerEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// HandleCreateAccount creates a new account
func (h *APIHandler) HandleCreateAccount(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		UserID   string `json:"userId"`             // Optional: Client User ID
		Username string `json:"username,omitempty"` // Admin-assigned username
		Password string `json:"password,omitempty"` // Admin-assigned password
		IsDemo   bool   `json:"isDemo"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		req.UserID = "default"
	}

	account := h.engine.CreateAccount(req.UserID, req.Username, req.Password, req.IsDemo)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

// HandleAdminResetPassword resets an account password
func (h *APIHandler) HandleAdminResetPassword(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		AccountID   int64  `json:"accountId"`
		NewPassword string `json:"newPassword"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.NewPassword == "" {
		http.Error(w, "New Password cannot be empty", http.StatusBadRequest)
		return
	}

	err := h.engine.UpdatePassword(req.AccountID, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// HandleAdminUpdateAccount updates account configuration
func (h *APIHandler) HandleAdminUpdateAccount(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		AccountID  int64   `json:"accountId"`
		Leverage   float64 `json:"leverage"`
		MarginMode string  `json:"marginMode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.engine.UpdateAccount(req.AccountID, req.Leverage, req.MarginMode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// HandleGetSymbols returns all symbols
func (h *APIHandler) HandleGetSymbols(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	symbols := h.engine.GetSymbols()
	if symbols == nil {
		symbols = []*SymbolSpec{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(symbols)
}
