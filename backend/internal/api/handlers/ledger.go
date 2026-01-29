package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/epic1st/rtx/backend/internal/core"
)

// HandleGetTrades returns trade history
func (h *APIHandler) HandleGetTrades(w http.ResponseWriter, r *http.Request) {
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

	trades := h.engine.GetTrades(accountID)
	if trades == nil {
		trades = []core.Trade{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trades)
}

// HandleGetLedger returns ledger history
func (h *APIHandler) HandleGetLedger(w http.ResponseWriter, r *http.Request) {
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

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	entries := h.engine.GetLedger().GetHistory(accountID, limit)
	if entries == nil {
		entries = []core.LedgerEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
