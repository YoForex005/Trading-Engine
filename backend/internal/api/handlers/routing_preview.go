package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/epic1st/rtx/backend/cbook"
)

// RoutingPreviewRequest represents a request to preview routing decision
type RoutingPreviewRequest struct {
	Symbol       string  `json:"symbol" form:"symbol"`
	Volume       float64 `json:"volume" form:"volume"`
	AccountID    int64   `json:"accountId" form:"accountId"`
	Side         string  `json:"side" form:"side"` // BUY or SELL
	Volatility   float64 `json:"volatility" form:"volatility"`
	UserID       string  `json:"userId" form:"userId"`
	Username     string  `json:"username" form:"username"`
}

// RoutingPreviewResponse represents the API response for routing decision preview
type RoutingPreviewResponse struct {
	Action          string  `json:"action"`          // A_BOOK, B_BOOK, PARTIAL_HEDGE, REJECT
	TargetLP        string  `json:"targetLp"`        // Target LP name
	HedgePercent    float64 `json:"hedgePercent"`    // Hedge percentage (0-100)
	ABookPercent    float64 `json:"aBookPercent"`    // A-Book percentage (0-100)
	BBookPercent    float64 `json:"bBookPercent"`    // B-Book percentage (0-100)
	Reason          string  `json:"reason"`          // Explanation of routing decision
	ToxicityScore   float64 `json:"toxicityScore"`   // Client toxicity score (0-100)
	ExposureRisk    float64 `json:"exposureRisk"`    // Exposure risk level (0-100)
	DecisionTime    string  `json:"decisionTime"`    // ISO 8601 timestamp
	ExposureImpact  string  `json:"exposureImpact"`  // Impact on portfolio exposure
}

// HandleRoutingPreview handles GET /api/routing/preview
// Returns a preview of the routing decision WITHOUT executing the order
func (h *APIHandler) HandleRoutingPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse query parameters
	req := RoutingPreviewRequest{}

	if r.Method == http.MethodPost {
		// Parse JSON body for POST
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	} else {
		// Parse query parameters for GET
		req.Symbol = r.URL.Query().Get("symbol")
		req.Side = r.URL.Query().Get("side")
		req.UserID = r.URL.Query().Get("userId")
		req.Username = r.URL.Query().Get("username")

		volumeStr := r.URL.Query().Get("volume")
		if volumeStr != "" {
			v, err := strconv.ParseFloat(volumeStr, 64)
			if err != nil {
				http.Error(w, "Invalid volume parameter", http.StatusBadRequest)
				return
			}
			req.Volume = v
		}

		accountIDStr := r.URL.Query().Get("accountId")
		if accountIDStr != "" {
			accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
			if err != nil {
				http.Error(w, "Invalid accountId parameter", http.StatusBadRequest)
				return
			}
			req.AccountID = accountID
		}

		volatilityStr := r.URL.Query().Get("volatility")
		if volatilityStr != "" {
			v, err := strconv.ParseFloat(volatilityStr, 64)
			if err != nil {
				http.Error(w, "Invalid volatility parameter", http.StatusBadRequest)
				return
			}
			req.Volatility = v
		}
	}

	// Validate request
	if req.Symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}
	if req.Volume <= 0 {
		http.Error(w, "Volume must be greater than 0", http.StatusBadRequest)
		return
	}
	if req.AccountID <= 0 {
		http.Error(w, "Valid accountId is required", http.StatusBadRequest)
		return
	}

	// Default values
	if req.Side == "" {
		req.Side = "BUY"
	}
	if req.Volatility < 0 {
		req.Volatility = 0.015 // Default 1.5% volatility
	}

	// Check if C-Book engine is available
	if h.cbookEngine == nil {
		http.Error(w, "C-Book engine not available", http.StatusServiceUnavailable)
		return
	}

	// Get routing decision from C-Book engine
	decision, err := h.cbookEngine.RouteOrder(
		req.AccountID,
		req.UserID,
		req.Username,
		req.Symbol,
		req.Side,
		req.Volume,
		req.Volatility,
	)

	if err != nil {
		http.Error(w, "Failed to determine routing decision: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert RoutingDecision to API response
	response := RoutingPreviewResponse{
		Action:        string(decision.Action),
		TargetLP:      decision.TargetLP,
		ABookPercent:  decision.ABookPercent,
		BBookPercent:  decision.BBookPercent,
		Reason:        decision.Reason,
		ToxicityScore: decision.ToxicityScore,
		ExposureRisk:  decision.ExposureRisk,
		DecisionTime:  decision.DecisionTime.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Calculate hedge percent for partial hedge
	if decision.Action == cbook.ActionPartialHedge {
		response.HedgePercent = decision.ABookPercent
	}

	// Determine exposure impact
	if decision.ExposureRisk > 75 {
		response.ExposureImpact = "HIGH - Consider hedging"
	} else if decision.ExposureRisk > 50 {
		response.ExposureImpact = "MEDIUM - Monitor closely"
	} else {
		response.ExposureImpact = "LOW - Within acceptable range"
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
