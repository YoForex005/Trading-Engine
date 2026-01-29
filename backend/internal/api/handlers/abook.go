package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/epic1st/rtx/backend/abook"
)

// ABookHandler handles A-Book execution API endpoints
type ABookHandler struct {
	engine *abook.ExecutionEngine
}

// NewABookHandler creates a new A-Book API handler
func NewABookHandler(engine *abook.ExecutionEngine) *ABookHandler {
	return &ABookHandler{
		engine: engine,
	}
}

// HandlePlaceOrder handles A-Book order placement
func (h *ABookHandler) HandlePlaceOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		AccountID string  `json:"accountId"`
		Symbol    string  `json:"symbol"`
		Side      string  `json:"side"` // BUY or SELL
		Type      string  `json:"type"` // MARKET or LIMIT
		Volume    float64 `json:"volume"`
		Price     float64 `json:"price,omitempty"`
		SL        float64 `json:"sl,omitempty"`
		TP        float64 `json:"tp,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create order request
	orderReq := &abook.OrderRequest{
		AccountID: req.AccountID,
		Symbol:    req.Symbol,
		Side:      req.Side,
		Type:      req.Type,
		Volume:    req.Volume,
		Price:     req.Price,
		SL:        req.SL,
		TP:        req.TP,
	}

	// Place order
	order, err := h.engine.PlaceOrder(orderReq)
	if err != nil {
		log.Printf("[A-Book API] Order placement failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[A-Book API] Order placed: %s %s %.2f %s @ %.5f (Status: %s)",
		order.Symbol, order.Side, order.Volume, order.Type, order.Price, order.Status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"order":   order,
	})
}

// HandleCancelOrder handles order cancellation
func (h *ABookHandler) HandleCancelOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		OrderID string `json:"orderId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.engine.CancelOrder(req.OrderID); err != nil {
		log.Printf("[A-Book API] Cancel failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"orderId": req.OrderID,
	})
}

// HandleGetOrder handles order status retrieval
func (h *ABookHandler) HandleGetOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	orderID := r.URL.Query().Get("orderId")
	if orderID == "" {
		http.Error(w, "orderId parameter required", http.StatusBadRequest)
		return
	}

	order, err := h.engine.GetOrder(orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// HandleGetPositions handles position listing
func (h *ABookHandler) HandleGetPositions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	accountID := r.URL.Query().Get("accountId")

	positions := h.engine.GetPositions(accountID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(positions)
}

// HandleClosePosition handles position closing
func (h *ABookHandler) HandleClosePosition(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		PositionID string `json:"positionId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.engine.ClosePosition(req.PositionID); err != nil {
		log.Printf("[A-Book API] Close position failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"positionId": req.PositionID,
	})
}

// HandleGetMetrics handles execution metrics retrieval
func (h *ABookHandler) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	metrics := h.engine.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// HandleGetLPHealth handles LP health status retrieval
func (h *ABookHandler) HandleGetLPHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// This would need to be implemented in the SOR
	// For now, return basic status
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"lps":    []string{},
	})
}

// HandleGetQuotes handles aggregated quote retrieval
func (h *ABookHandler) HandleGetQuotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "symbol parameter required", http.StatusBadRequest)
		return
	}

	// This would need to be implemented via SOR
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"symbol": symbol,
		"quotes": []string{},
	})
}
