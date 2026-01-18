package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/epic1st/rtx/backend/internal/core"
)

// HandleAdminGetSymbols returns all symbols including disabled ones
func (h *APIHandler) HandleAdminGetSymbols(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	symbols := h.engine.GetSymbols()
	if symbols == nil {
		symbols = []*core.SymbolSpec{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(symbols)
}

// HandleAdminToggleSymbol toggles a symbol's enabled/disabled status
func (h *APIHandler) HandleAdminToggleSymbol(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Symbol   string `json:"symbol"`
		Disabled bool   `json:"disabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Update Engine (Trading Logic)
	err := h.engine.ToggleSymbol(req.Symbol, req.Disabled)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// 2. Update Hub (Streaming Logic)
	if h.hub != nil {
		h.hub.ToggleSymbol(req.Symbol, req.Disabled)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true, "disabled": req.Disabled})
}

// UpdateSymbolRequest represents the request body for updating a symbol
type UpdateSymbolRequest struct {
	ContractSize     *float64 `json:"contract_size,omitempty"`
	PipSize          *float64 `json:"pip_size,omitempty"`
	PipValue         *float64 `json:"pip_value,omitempty"`
	MarginPercent    *float64 `json:"margin_percent,omitempty"`
	CommissionPerLot *float64 `json:"commission_per_lot,omitempty"`
	SpreadMarkup     *float64 `json:"spread_markup,omitempty"`
}

// HandleAdminUpdateSymbol updates symbol parameters via PATCH request
// URL: PATCH /api/admin/symbols/:symbol
// Validates input values, updates symbol in engine, and persists to database
func (h *APIHandler) HandleAdminUpdateSymbol(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PATCH" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract symbol from URL path: /api/admin/symbols/:symbol
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	symbol := parts[4]

	if symbol == "" {
		http.Error(w, "Symbol parameter required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req UpdateSymbolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get current symbol spec from engine
	symbols := h.engine.GetSymbols()
	var current *core.SymbolSpec
	for _, s := range symbols {
		if s.Symbol == symbol {
			current = s
			break
		}
	}

	if current == nil {
		http.Error(w, "Symbol not found", http.StatusNotFound)
		return
	}

	// Validate and apply updates
	if req.ContractSize != nil {
		if *req.ContractSize <= 0 {
			http.Error(w, "contract_size must be greater than 0", http.StatusBadRequest)
			return
		}
		current.ContractSize = *req.ContractSize
	}

	if req.PipSize != nil {
		if *req.PipSize <= 0 {
			http.Error(w, "pip_size must be greater than 0", http.StatusBadRequest)
			return
		}
		current.PipSize = *req.PipSize
	}

	if req.PipValue != nil {
		if *req.PipValue <= 0 {
			http.Error(w, "pip_value must be greater than 0", http.StatusBadRequest)
			return
		}
		current.PipValue = *req.PipValue
	}

	if req.MarginPercent != nil {
		if *req.MarginPercent < 0 {
			http.Error(w, "margin_percent must be non-negative", http.StatusBadRequest)
			return
		}
		current.MarginPercent = *req.MarginPercent
	}

	if req.CommissionPerLot != nil {
		if *req.CommissionPerLot < 0 {
			http.Error(w, "commission_per_lot must be non-negative", http.StatusBadRequest)
			return
		}
		current.CommissionPerLot = *req.CommissionPerLot
	}

	// Note: spread_markup is currently not stored in SymbolSpec, but is accepted for future compatibility
	// It can be implemented in a future update if needed

	// Update symbol in engine
	h.engine.UpdateSymbol(current)

	// Update hub if available (for streaming logic consistency)
	if h.hub != nil {
		h.hub.UpdateSymbol(current)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"symbol":  current,
	})
}
