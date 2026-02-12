package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
)

// OverviewMetrics represents dashboard summary metrics
type OverviewMetrics struct {
	TotalAccounts     int     `json:"totalAccounts"`
	ActiveAccounts    int     `json:"activeAccounts"`
	TotalPositions    int     `json:"totalPositions"`
	OpenPositions     int     `json:"openPositions"`
	TotalVolume24h    float64 `json:"totalVolume24h"`
	TotalPnL24h       float64 `json:"totalPnL24h"`
	TotalDeposits     float64 `json:"totalDeposits"`
	TotalWithdrawals  float64 `json:"totalWithdrawals"`
}

// TradingMetrics represents trading activity metrics
type TradingMetrics struct {
	TradesPerHour   []int              `json:"tradesPerHour"`   // 24 hours
	VolumeBySymbol  map[string]float64 `json:"volumeBySymbol"`
	AvgTradeSize    float64            `json:"avgTradeSize"`
	WinRate         float64            `json:"winRate"`
	TotalTrades     int                `json:"totalTrades"`
}

// RevenueMetrics represents revenue breakdown
type RevenueMetrics struct {
	SpreadRevenue     float64 `json:"spreadRevenue"`
	CommissionRevenue float64 `json:"commissionRevenue"`
	SwapRevenue       float64 `json:"swapRevenue"`
	TotalRevenue      float64 `json:"totalRevenue"`
}

// HandleAdminAnalyticsOverview returns dashboard summary metrics
func (h *APIHandler) HandleAdminAnalyticsOverview(w http.ResponseWriter, r *http.Request) {
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

	allAccounts := h.engine.GetAllAccounts()
	allPositions := h.engine.GetAllPositions()
	ledger := h.engine.GetLedger()

	// Total accounts
	totalAccounts := len(allAccounts)

	// Active accounts (accounts with open positions)
	activeAccountsMap := make(map[int64]bool)
	openPositions := 0
	for _, pos := range allPositions {
		if pos.Status == "OPEN" {
			activeAccountsMap[pos.AccountID] = true
			openPositions++
		}
	}
	activeAccounts := len(activeAccountsMap)

	// Total positions (including closed)
	totalPositions := len(allPositions)

	// Calculate 24h metrics
	now := time.Now()
	cutoff24h := now.Add(-24 * time.Hour)

	var totalVolume24h float64
	var totalPnL24h float64

	// Get all trades from all accounts
	var allTrades []core.Trade
	for _, acc := range allAccounts {
		trades := h.engine.GetTrades(acc.ID)
		allTrades = append(allTrades, trades...)
	}

	// Sum volume and PnL from last 24h
	for _, trade := range allTrades {
		if trade.ExecutedAt.After(cutoff24h) {
			totalVolume24h += trade.Volume
			totalPnL24h += trade.RealizedPnL
		}
	}

	// Get deposits and withdrawals from ledger
	var totalDeposits float64
	var totalWithdrawals float64

	for _, acc := range allAccounts {
		entries := ledger.GetHistory(acc.ID, 0)
		for _, entry := range entries {
			if entry.Type == "DEPOSIT" && entry.Status == "COMPLETED" {
				totalDeposits += entry.Amount
			} else if entry.Type == "WITHDRAW" && entry.Status == "COMPLETED" {
				totalWithdrawals += entry.Amount
			}
		}
	}

	metrics := OverviewMetrics{
		TotalAccounts:    totalAccounts,
		ActiveAccounts:   activeAccounts,
		TotalPositions:   totalPositions,
		OpenPositions:    openPositions,
		TotalVolume24h:   totalVolume24h,
		TotalPnL24h:      totalPnL24h,
		TotalDeposits:    totalDeposits,
		TotalWithdrawals: totalWithdrawals,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// HandleAdminAnalyticsTrading returns trading activity metrics
func (h *APIHandler) HandleAdminAnalyticsTrading(w http.ResponseWriter, r *http.Request) {
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

	allAccounts := h.engine.GetAllAccounts()

	// Get all trades
	var allTrades []core.Trade
	for _, acc := range allAccounts {
		trades := h.engine.GetTrades(acc.ID)
		allTrades = append(allTrades, trades...)
	}

	// Initialize 24-hour array
	tradesPerHour := make([]int, 24)
	now := time.Now()

	// Volume by symbol
	volumeBySymbol := make(map[string]float64)

	// Calculate metrics
	var totalVolume float64
	var winningTrades int
	var totalPnL float64

	for _, trade := range allTrades {
		// Trades per hour (last 24 hours)
		hoursAgo := int(now.Sub(trade.ExecutedAt).Hours())
		if hoursAgo >= 0 && hoursAgo < 24 {
			hourIndex := 23 - hoursAgo // Reverse order (most recent = index 23)
			tradesPerHour[hourIndex]++
		}

		// Volume by symbol
		volumeBySymbol[trade.Symbol] += trade.Volume
		totalVolume += trade.Volume

		// Win rate
		if trade.RealizedPnL > 0 {
			winningTrades++
		}
		totalPnL += trade.RealizedPnL
	}

	// Calculate averages
	avgTradeSize := 0.0
	if len(allTrades) > 0 {
		avgTradeSize = totalVolume / float64(len(allTrades))
	}

	winRate := 0.0
	if len(allTrades) > 0 {
		winRate = (float64(winningTrades) / float64(len(allTrades))) * 100
	}

	metrics := TradingMetrics{
		TradesPerHour:  tradesPerHour,
		VolumeBySymbol: volumeBySymbol,
		AvgTradeSize:   avgTradeSize,
		WinRate:        winRate,
		TotalTrades:    len(allTrades),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// HandleAdminAnalyticsRevenue returns revenue breakdown
func (h *APIHandler) HandleAdminAnalyticsRevenue(w http.ResponseWriter, r *http.Request) {
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

	allAccounts := h.engine.GetAllAccounts()
	ledger := h.engine.GetLedger()

	var commissionRevenue float64
	var swapRevenue float64
	var spreadRevenue float64

	// Calculate from ledger entries
	for _, acc := range allAccounts {
		entries := ledger.GetHistory(acc.ID, 0)
		for _, entry := range entries {
			// Commission revenue (negative amounts in ledger = revenue for broker)
			if entry.Type == "COMMISSION" {
				// Commissions are deducted from client balance, so they're broker revenue
				commissionRevenue += -entry.Amount // Negative in ledger = positive revenue
			}
			// Swap revenue (can be positive or negative for clients)
			if entry.Type == "SWAP" {
				swapRevenue += -entry.Amount // Negative in ledger = positive revenue
			}
		}
	}

	// Get all trades to calculate spread revenue
	var allTrades []core.Trade
	for _, acc := range allAccounts {
		trades := h.engine.GetTrades(acc.ID)
		allTrades = append(allTrades, trades...)
	}

	// Calculate spread revenue from B-Book trades
	// Spread = difference between bid/ask, captured on each trade
	// For simplification, estimate spread revenue from commission
	// In a real implementation, this would be calculated from actual bid/ask spreads
	for _, trade := range allTrades {
		// Commission already captured above
		// Spread revenue would require bid/ask price tracking
		// For now, we'll use a simplified calculation
		// Typical forex spread is 1-3 pips, worth ~$10-30 per lot

		// Get symbol spec to estimate spread
		symbols := h.engine.GetSymbols()
		for _, spec := range symbols {
			if spec.Symbol == trade.Symbol {
				// Estimate spread revenue: volume * pip value * spread in pips
				// This is a simplified calculation
				estimatedSpread := trade.Volume * 10.0 // $10 per lot average spread
				spreadRevenue += estimatedSpread
				break
			}
		}
	}

	totalRevenue := spreadRevenue + commissionRevenue + swapRevenue

	metrics := RevenueMetrics{
		SpreadRevenue:     spreadRevenue,
		CommissionRevenue: commissionRevenue,
		SwapRevenue:       swapRevenue,
		TotalRevenue:      totalRevenue,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
