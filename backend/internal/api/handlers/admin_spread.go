package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/epic1st/rtx/backend/internal/core"
)

// SpreadConfigResponse represents the response for GET /api/symbols/:name/spread
type SpreadConfigResponse struct {
	Symbol           string  `json:"symbol"`
	CurrentSpread    float64 `json:"currentSpread"`
	ConfiguredMarkup float64 `json:"configuredMarkup"`
	MinSpread        float64 `json:"minSpread"`
	MaxSpread        float64 `json:"maxSpread"`
	Digits           int     `json:"digits"`
}

// UpdateSpreadRequest represents the request body for PUT /admin/symbols/:name/spread
type UpdateSpreadRequest struct {
	Markup    float64 `json:"markup"`
	MinSpread float64 `json:"minSpread"`
	MaxSpread float64 `json:"maxSpread"`
}

// HandleGetSymbolSpread returns the spread configuration for a symbol
// GET /api/symbols/:name/spread
func (h *APIHandler) HandleGetSymbolSpread(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract symbol from URL path: /api/symbols/:name/spread
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	symbol := strings.ToUpper(parts[3])

	if symbol == "" {
		http.Error(w, "Symbol parameter required", http.StatusBadRequest)
		return
	}

	// Get symbol spec from engine
	symbols := h.engine.GetSymbols()
	var symbolSpec *core.SymbolSpec

	for _, s := range symbols {
		if s.Symbol == symbol {
			symbolSpec = s
			break
		}
	}

	if symbolSpec == nil {
		http.Error(w, "Symbol not found", http.StatusNotFound)
		return
	}

	// Calculate current spread from market data
	var currentSpread float64
	if h.hub != nil {
		tick := h.hub.GetLatestPrice(symbol)
		if tick != nil {
			currentSpread = tick.Ask - tick.Bid
		}
	}

	// Calculate digits from pip size
	digits := 2
	if symbolSpec.PipSize == 0.0001 {
		digits = 4
	} else if symbolSpec.PipSize == 0.00001 {
		digits = 5
	} else if symbolSpec.PipSize == 0.01 {
		digits = 2
	} else if symbolSpec.PipSize == 0.001 {
		digits = 3
	}

	response := SpreadConfigResponse{
		Symbol:           symbol,
		CurrentSpread:    currentSpread,
		ConfiguredMarkup: symbolSpec.SpreadMarkup,
		MinSpread:        symbolSpec.MinSpread,
		MaxSpread:        symbolSpec.MaxSpread,
		Digits:           digits,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleUpdateSymbolSpread updates the spread configuration for a symbol
// PUT /admin/symbols/:name/spread
func (h *APIHandler) HandleUpdateSymbolSpread(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract symbol from URL path: /admin/symbols/:name/spread
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	symbol := strings.ToUpper(parts[3])

	if symbol == "" {
		http.Error(w, "Symbol parameter required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req UpdateSpreadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Markup < 0 {
		http.Error(w, "markup must be non-negative", http.StatusBadRequest)
		return
	}
	if req.MinSpread < 0 {
		http.Error(w, "minSpread must be non-negative", http.StatusBadRequest)
		return
	}
	if req.MaxSpread < req.MinSpread {
		http.Error(w, "maxSpread must be greater than or equal to minSpread", http.StatusBadRequest)
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

	// Update spread configuration
	current.SpreadMarkup = req.Markup
	current.MinSpread = req.MinSpread
	current.MaxSpread = req.MaxSpread

	// Update symbol in engine
	h.engine.UpdateSymbol(current)

	// Update hub if available
	if h.hub != nil {
		h.hub.UpdateSymbol(current)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"symbol":  current,
	})
}
