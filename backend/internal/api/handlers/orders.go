package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

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

// HandlePlacePendingOrder creates a pending order (BUY_LIMIT, SELL_LIMIT, BUY_STOP, SELL_STOP)
func (h *APIHandler) HandlePlacePendingOrder(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		AccountID    int64   `json:"accountId"`
		Type         string  `json:"type"` // BUY_LIMIT, SELL_LIMIT, BUY_STOP, SELL_STOP
		Symbol       string  `json:"symbol"`
		Volume       float64 `json:"volume"`
		TriggerPrice float64 `json:"triggerPrice"`
		SL           float64 `json:"sl,omitempty"`
		TP           float64 `json:"tp,omitempty"`
		OCOLinkID    *int64  `json:"ocoLinkId,omitempty"` // ID of order to link for OCO
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AccountID == 0 {
		req.AccountID = 1 // Default account
	}

	// Validate required fields
	if req.Type == "" {
		http.Error(w, "Order type is required", http.StatusBadRequest)
		return
	}
	if req.Symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}
	if req.Volume <= 0 {
		http.Error(w, "Volume must be greater than 0", http.StatusBadRequest)
		return
	}
	if req.TriggerPrice <= 0 {
		http.Error(w, "Trigger price must be greater than 0", http.StatusBadRequest)
		return
	}

	order, err := h.engine.CreatePendingOrderWithOCO(req.AccountID, req.Type, req.Symbol, req.Volume, req.TriggerPrice, req.SL, req.TP, req.OCOLinkID)
	if err != nil {
		log.Printf("[API] Pending order rejected: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"order":   order,
	})
}

// HandleModifyOrder modifies a pending order's parameters
func (h *APIHandler) HandleModifyOrder(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract order ID from URL path or query
	orderIDStr := r.URL.Query().Get("id")
	if orderIDStr == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var req struct {
		TriggerPrice *float64   `json:"triggerPrice,omitempty"`
		SL           *float64   `json:"sl,omitempty"`
		TP           *float64   `json:"tp,omitempty"`
		Volume       *float64   `json:"volume,omitempty"`
		ExpiryTime   *time.Time `json:"expiryTime,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Modify the order
	order, err := h.engine.ModifyOrder(orderID, req.TriggerPrice, req.SL, req.TP, req.Volume, req.ExpiryTime)
	if err != nil {
		log.Printf("[API] Order modification rejected: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"order":   order,
	})
}

// HandleCancelOrder cancels a pending order
func (h *APIHandler) HandleCancelOrder(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract order ID from URL path or query
	orderIDStr := r.URL.Query().Get("id")
	if orderIDStr == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// Cancel the order
	if err := h.engine.CancelOrder(orderID); err != nil {
		log.Printf("[API] Order cancellation rejected: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Order cancelled",
	})
}
