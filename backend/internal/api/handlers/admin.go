package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/epic1st/rtx/backend/internal/core"
)

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

	var accounts []*core.Account
	// Get all accounts from engine
	for i := int64(1); i <= 100; i++ {
		acc, ok := h.engine.GetAccount(i)
		if ok {
			accounts = append(accounts, acc)
		}
	}

	if accounts == nil {
		accounts = []*core.Account{}
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

	var entries []core.LedgerEntry
	if typeFilter != "" {
		entries = h.engine.GetLedger().GetEntriesByType(typeFilter, limit)
	} else {
		entries = h.engine.GetLedger().GetAllEntries(limit)
	}

	if entries == nil {
		entries = []core.LedgerEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
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
