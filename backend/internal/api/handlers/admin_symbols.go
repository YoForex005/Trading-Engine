package handlers

import (
	"encoding/json"
	"net/http"

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
