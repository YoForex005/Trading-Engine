package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

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
