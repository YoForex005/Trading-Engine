package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/epic1st/rtx/backend/auth"
	"github.com/epic1st/rtx/backend/internal/core"
)

// ExposureData represents current exposure for a symbol (for Risk Dashboard)
type ExposureData struct {
	Symbol       string  `json:"symbol"`
	LongVolume   float64 `json:"longVolume"`
	ShortVolume  float64 `json:"shortVolume"`
	NetExposure  float64 `json:"netExposure"`
	PnL          float64 `json:"pnl"`
}

// ExposureSummary represents aggregate risk metrics
type ExposureSummary struct {
	TotalExposure      float64                    `json:"totalExposure"`
	NetPnL             float64                    `json:"netPnL"`
	MarginUtilization  float64                    `json:"marginUtilization"`
	ActivePositions    int                        `json:"activePositions"`
	RoutingBreakdown   RoutingBreakdown           `json:"routingBreakdown"`
}

// RoutingBreakdown represents A-Book vs B-Book vs C-Book split
type RoutingBreakdown struct {
	ABook float64 `json:"aBook"` // Percentage
	BBook float64 `json:"bBook"` // Percentage
	CBook float64 `json:"cBook"` // Percentage
}

// TopClient represents a client with high exposure
type TopClient struct {
	AccountID     int64   `json:"accountId"`
	AccountNumber string  `json:"accountNumber"`
	Symbol        string  `json:"symbol"`
	Volume        float64 `json:"volume"`
	PnL           float64 `json:"pnl"`
	MarginLevel   float64 `json:"marginLevel"`
}

// Helper function to validate JWT token from Authorization header
func validateAuthToken(r *http.Request) (*auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, http.ErrNoCookie // Using standard error for "missing auth"
	}

	// Extract Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, http.ErrNoCookie
	}

	tokenString := parts[1]
	return auth.ValidateTokenWithDefault(tokenString)
}

// HandleAnalyticsExposureCurrent returns current exposure by symbol
func (h *APIHandler) HandleAnalyticsExposureCurrent(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Validate JWT token
	_, err := validateAuthToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all open positions
	allPositions := h.engine.GetAllPositions()

	// Group by symbol and calculate exposure
	exposureMap := make(map[string]*ExposureData)

	for _, pos := range allPositions {
		if pos.Status != "OPEN" {
			continue
		}

		if _, exists := exposureMap[pos.Symbol]; !exists {
			exposureMap[pos.Symbol] = &ExposureData{
				Symbol: pos.Symbol,
			}
		}

		exp := exposureMap[pos.Symbol]

		// Accumulate volumes
		if pos.Side == "BUY" {
			exp.LongVolume += pos.Volume
			exp.NetExposure += pos.Volume
		} else if pos.Side == "SELL" {
			exp.ShortVolume += pos.Volume
			exp.NetExposure -= pos.Volume
		}

		// Accumulate PnL
		exp.PnL += pos.UnrealizedPnL
	}

	// Convert map to slice
	result := make([]ExposureData, 0, len(exposureMap))
	for _, exp := range exposureMap {
		result = append(result, *exp)
	}

	// Sort by absolute net exposure (largest first)
	sort.Slice(result, func(i, j int) bool {
		absI := result[i].NetExposure
		if absI < 0 {
			absI = -absI
		}
		absJ := result[j].NetExposure
		if absJ < 0 {
			absJ = -absJ
		}
		return absI > absJ
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleAnalyticsExposureSummary returns aggregate risk metrics
func (h *APIHandler) HandleAnalyticsExposureSummary(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Validate JWT token
	_, err := validateAuthToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	allPositions := h.engine.GetAllPositions()
	allAccounts := h.engine.GetAllAccounts()

	var totalExposure float64
	var netPnL float64
	var totalUsedMargin float64
	var totalEquity float64
	activePositions := 0

	// Routing breakdown counters
	var aBookVolume, bBookVolume, cBookVolume float64

	// Calculate metrics from positions
	for _, pos := range allPositions {
		if pos.Status != "OPEN" {
			continue
		}

		activePositions++

		// Calculate notional value for exposure
		notional := h.calculateNotionalValue(pos)
		totalExposure += notional

		// Sum PnL
		netPnL += pos.UnrealizedPnL

		// Routing breakdown (based on position metadata or default to B-Book)
		// For now, assume all positions are B-Book unless we have routing info
		// This can be enhanced later with actual routing data
		bBookVolume += pos.Volume
	}

	// Calculate margin metrics from accounts
	for _, acc := range allAccounts {
		summary, err := h.engine.GetAccountSummary(acc.ID)
		if err != nil {
			continue
		}
		totalUsedMargin += summary.Margin
		totalEquity += summary.Equity
	}

	// Calculate margin utilization percentage
	marginUtilization := 0.0
	if totalEquity > 0 {
		marginUtilization = (totalUsedMargin / totalEquity) * 100
	}

	// Calculate routing breakdown percentages
	totalVolume := aBookVolume + bBookVolume + cBookVolume
	routingBreakdown := RoutingBreakdown{
		ABook: 0,
		BBook: 100, // Default: all B-Book for now
		CBook: 0,
	}
	if totalVolume > 0 {
		routingBreakdown.ABook = (aBookVolume / totalVolume) * 100
		routingBreakdown.BBook = (bBookVolume / totalVolume) * 100
		routingBreakdown.CBook = (cBookVolume / totalVolume) * 100
	}

	summary := ExposureSummary{
		TotalExposure:     totalExposure,
		NetPnL:            netPnL,
		MarginUtilization: marginUtilization,
		ActivePositions:   activePositions,
		RoutingBreakdown:  routingBreakdown,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// HandleAnalyticsTopClients returns top 10 clients by exposure
func (h *APIHandler) HandleAnalyticsTopClients(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Validate JWT token
	_, err := validateAuthToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	allPositions := h.engine.GetAllPositions()
	allAccounts := h.engine.GetAllAccounts()

	// Build account lookup map
	accountMap := make(map[int64]*core.Account)
	for _, acc := range allAccounts {
		accountMap[acc.ID] = acc
	}

	// Group positions by account+symbol
	type ClientSymbolKey struct {
		AccountID int64
		Symbol    string
	}

	clientExposureMap := make(map[ClientSymbolKey]*TopClient)

	for _, pos := range allPositions {
		if pos.Status != "OPEN" {
			continue
		}

		key := ClientSymbolKey{
			AccountID: pos.AccountID,
			Symbol:    pos.Symbol,
		}

		if _, exists := clientExposureMap[key]; !exists {
			account, ok := accountMap[pos.AccountID]
			accountNumber := ""
			if ok {
				accountNumber = account.AccountNumber
			}

			clientExposureMap[key] = &TopClient{
				AccountID:     pos.AccountID,
				AccountNumber: accountNumber,
				Symbol:        pos.Symbol,
				Volume:        0,
				PnL:           0,
				MarginLevel:   0,
			}
		}

		client := clientExposureMap[key]
		client.Volume += pos.Volume
		client.PnL += pos.UnrealizedPnL
	}

	// Calculate margin level for each account
	accountMarginLevels := make(map[int64]float64)
	for _, acc := range allAccounts {
		summary, err := h.engine.GetAccountSummary(acc.ID)
		if err == nil {
			accountMarginLevels[acc.ID] = summary.MarginLevel
		}
	}

	// Update margin levels
	for key, client := range clientExposureMap {
		if marginLevel, ok := accountMarginLevels[key.AccountID]; ok {
			client.MarginLevel = marginLevel
		}
	}

	// Convert to slice
	result := make([]TopClient, 0, len(clientExposureMap))
	for _, client := range clientExposureMap {
		result = append(result, *client)
	}

	// Sort by volume descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].Volume > result[j].Volume
	})

	// Take top 10
	if len(result) > 10 {
		result = result[:10]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
