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
