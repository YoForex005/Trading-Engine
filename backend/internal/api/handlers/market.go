package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/epic1st/rtx/backend/internal/core"
)

// HandleGetSymbols returns all symbols
func (h *APIHandler) HandleGetSymbols(w http.ResponseWriter, r *http.Request) {
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
