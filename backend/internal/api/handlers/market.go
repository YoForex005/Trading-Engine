package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/epic1st/rtx/backend/internal/core"
)

// HandleGetSymbols returns all enabled symbols
func (h *APIHandler) HandleGetSymbols(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	allSymbols := h.engine.GetSymbols()
	if allSymbols == nil {
		allSymbols = []*core.SymbolSpec{}
	}

	// Filter out disabled symbols
	enabledSymbols := make([]*core.SymbolSpec, 0)
	for _, symbol := range allSymbols {
		if !symbol.Disabled {
			enabledSymbols = append(enabledSymbols, symbol)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enabledSymbols)
}
