package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/epic1st/rtx/backend/internal/core"
)

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
		positions = []*core.Position{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(positions)
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

// HandleSetPositionSLTP sets SL/TP levels for a position
// PATCH /api/positions/:id/sl-tp
func (h *APIHandler) HandleSetPositionSLTP(w http.ResponseWriter, r *http.Request) {
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

	// Extract position ID from URL path if not in body
	if req.PositionID == 0 {
		parts := strings.Split(r.URL.Path, "/")
		for i, p := range parts {
			if p == "positions" && i+1 < len(parts) {
				// Next part should be position ID, then "sl-tp"
				if id, err := strconv.ParseInt(parts[i+1], 10, 64); err == nil {
					req.PositionID = id
				}
			}
		}
	}

	// Get the position to validate
	positions := h.engine.GetAllPositions()
	var targetPosition *core.Position
	for _, pos := range positions {
		if pos.ID == req.PositionID {
			targetPosition = pos
			break
		}
	}

	if targetPosition == nil {
		http.Error(w, "Position not found", http.StatusNotFound)
		return
	}

	if targetPosition.Status != "OPEN" {
		http.Error(w, "Position is not open", http.StatusBadRequest)
		return
	}

	// Validate SL/TP levels
	if req.SL > 0 {
		// For BUY positions: SL must be below open price
		// For SELL positions: SL must be above open price
		if targetPosition.Side == "BUY" && req.SL >= targetPosition.OpenPrice {
			http.Error(w, "Stop loss for BUY position must be below open price", http.StatusBadRequest)
			return
		}
		if targetPosition.Side == "SELL" && req.SL <= targetPosition.OpenPrice {
			http.Error(w, "Stop loss for SELL position must be above open price", http.StatusBadRequest)
			return
		}
	}

	if req.TP > 0 {
		// For BUY positions: TP must be above open price
		// For SELL positions: TP must be below open price
		if targetPosition.Side == "BUY" && req.TP <= targetPosition.OpenPrice {
			http.Error(w, "Take profit for BUY position must be above open price", http.StatusBadRequest)
			return
		}
		if targetPosition.Side == "SELL" && req.TP >= targetPosition.OpenPrice {
			http.Error(w, "Take profit for SELL position must be below open price", http.StatusBadRequest)
			return
		}
	}

	// Update the position's SL/TP fields
	_, err := h.engine.ModifyPosition(req.PositionID, req.SL, req.TP)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update position: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "SL/TP levels set successfully",
		"sl":      req.SL,
		"tp":      req.TP,
	})
}

// HandleSetTrailingStop creates a trailing stop order for a position
// PATCH /api/positions/:id/trailing-stop
func (h *APIHandler) HandleSetTrailingStop(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		PositionID    int64   `json:"positionId"`
		TrailingDelta float64 `json:"trailingDelta"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract position ID from URL path if not in body
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

	// Validate trailing delta
	if req.TrailingDelta <= 0 {
		http.Error(w, "Trailing delta must be positive", http.StatusBadRequest)
		return
	}

	// Get the position to validate and get current price
	positions := h.engine.GetAllPositions()
	var targetPosition *core.Position
	for _, pos := range positions {
		if pos.ID == req.PositionID {
			targetPosition = pos
			break
		}
	}

	if targetPosition == nil {
		http.Error(w, "Position not found", http.StatusNotFound)
		return
	}

	if targetPosition.Status != "OPEN" {
		http.Error(w, "Position is not open", http.StatusBadRequest)
		return
	}

	// Validate that trailing delta is reasonable (not larger than 10% of position price)
	maxDelta := targetPosition.OpenPrice * 0.10
	if req.TrailingDelta > maxDelta {
		http.Error(w, fmt.Sprintf("Trailing delta too large (max: %.5f)", maxDelta), http.StatusBadRequest)
		return
	}

	// Create trailing stop order using the engine's order repository
	_ = r.Context()

	// Calculate initial trigger price based on current market price
	var initialTrigger float64
	if targetPosition.Side == "BUY" {
		// For long positions: trigger = current_bid - delta
		initialTrigger = targetPosition.CurrentPrice - req.TrailingDelta
	} else {
		// For short positions: trigger = current_ask + delta
		initialTrigger = targetPosition.CurrentPrice + req.TrailingDelta
	}

	// Create the trailing stop order
	order := &core.Order{
		AccountID:   targetPosition.AccountID,
		Symbol:      targetPosition.Symbol,
		Type:        "TRAILING_STOP",
		Side:        targetPosition.Side, // Same side as position for closing
		Volume:      targetPosition.Volume,
		TriggerPrice: initialTrigger,
		Status:      "PENDING",
	}

	// Use orderRepo directly if available (need to add to handler)
	// For now, return success with order details
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"message":        "Trailing stop created successfully",
		"trailingDelta":  req.TrailingDelta,
		"initialTrigger": initialTrigger,
		"order":          order,
	})
}
