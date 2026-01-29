package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/epic1st/rtx/backend/internal/core"
)

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
		orders = []*core.Order{}
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
