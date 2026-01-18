package handlers

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
)

// RuleEffectivenessMetrics contains performance metrics for a routing rule
type RuleEffectivenessMetrics struct {
	RuleID            string              `json:"rule_id"`
	SharpeRatio       float64             `json:"sharpe_ratio"`
	ProfitFactor      float64             `json:"profit_factor"`
	MaxDrawdown       float64             `json:"max_drawdown"`
	TotalPnL          float64             `json:"total_pnl"`
	TradeCount        int                 `json:"trade_count"`
	WinRate           float64             `json:"win_rate"`
	AvgTradeDuration  float64             `json:"avg_trade_duration,omitempty"`
	ConsecutiveWins   int                 `json:"consecutive_wins,omitempty"`
	ConsecutiveLosses int                 `json:"consecutive_losses,omitempty"`
	Rank              int                 `json:"rank,omitempty"`
	Timeline          []TimelineDataPoint `json:"timeline,omitempty"`
}

// TimelineDataPoint represents a point in the performance timeline
type TimelineDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	PnL       float64   `json:"pnl"`
	Equity    float64   `json:"equity"`
}

// RuleEffectivenessResponse contains all rules with their metrics
type RuleEffectivenessResponse struct {
	Rules []RuleEffectivenessMetrics `json:"rules"`
}

// CalculateMetricsRequest contains trades for metric calculation
type CalculateMetricsRequest struct {
	Trades       []TradeData `json:"trades"`
	RiskFreeRate float64     `json:"risk_free_rate"`
}

// TradeData represents a trade for metric calculation
type TradeData struct {
	PnL       float64   `json:"pnl"`
	Timestamp time.Time `json:"timestamp"`
	Volume    float64   `json:"volume,omitempty"`
	Duration  float64   `json:"duration,omitempty"` // in seconds
}

// CalculateMetricsResponse contains calculated financial metrics
type CalculateMetricsResponse struct {
	SharpeRatio       float64 `json:"sharpe_ratio"`
	ProfitFactor      float64 `json:"profit_factor"`
	MaxDrawdown       float64 `json:"max_drawdown"`
	WinRate           float64 `json:"win_rate"`
	ConsecutiveWins   int     `json:"consecutive_wins"`
	ConsecutiveLosses int     `json:"consecutive_losses"`
}

// HandleGetRuleEffectiveness returns effectiveness metrics for all routing rules
// GET /api/analytics/rules/effectiveness?start_time=<timestamp>&end_time=<timestamp>&min_trades=<n>
func (h *APIHandler) HandleGetRuleEffectiveness(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Verify admin authentication
	if !h.isAdminUser(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	minTradesStr := r.URL.Query().Get("min_trades")

	var startTime, endTime time.Time

	if startTimeStr != "" {
		if startTimeUnix, parseErr := strconv.ParseInt(startTimeStr, 10, 64); parseErr == nil {
			startTime = time.Unix(startTimeUnix, 0)
		}
	}

	if endTimeStr != "" {
		if endTimeUnix, parseErr := strconv.ParseInt(endTimeStr, 10, 64); parseErr == nil {
			endTime = time.Unix(endTimeUnix, 0)
		}
	}

	minTrades := 1
	if minTradesStr != "" {
		if parsed, err := strconv.Atoi(minTradesStr); err == nil && parsed > 0 {
			minTrades = parsed
		}
	}

	// Get routing engine
	routingEngine := h.getRoutingEngine()
	if routingEngine == nil {
		http.Error(w, "Routing engine not available", http.StatusInternalServerError)
		return
	}

	// Get all rules
	rules := routingEngine.GetRules()

	// Calculate metrics for each rule
	ruleMetrics := make([]RuleEffectivenessMetrics, 0)

	for _, rule := range rules {
		// Get trades associated with this rule
		trades := h.getTradesForRule(rule.ID, startTime, endTime)

		// Skip if below minimum trade count
		if len(trades) < minTrades {
			continue
		}

		// Calculate metrics
		metrics := calculateRuleMetrics(rule.ID, trades, 0.02) // 2% risk-free rate

		ruleMetrics = append(ruleMetrics, metrics)
	}

	// Rank rules by Sharpe Ratio
	sort.Slice(ruleMetrics, func(i, j int) bool {
		return ruleMetrics[i].SharpeRatio > ruleMetrics[j].SharpeRatio
	})

	// Assign ranks
	for i := range ruleMetrics {
		ruleMetrics[i].Rank = i + 1
	}

	response := RuleEffectivenessResponse{
		Rules: ruleMetrics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("[AnalyticsAPI] Rule effectiveness query: %d rules analyzed", len(ruleMetrics))
}

// HandleGetRuleMetrics returns detailed metrics for a single rule
// GET /api/analytics/rules/{rule_id}/metrics?start_time=<timestamp>&end_time=<timestamp>
func (h *APIHandler) HandleGetRuleMetrics(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Verify admin authentication
	if !h.isAdminUser(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract rule ID from path: /api/analytics/rules/{id}/metrics
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid path: rule_id not found", http.StatusBadRequest)
		return
	}
	ruleID := pathParts[4]

	// Parse query parameters
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	var startTime, endTime time.Time

	if startTimeStr != "" {
		if startTimeUnix, err := strconv.ParseInt(startTimeStr, 10, 64); err == nil {
			startTime = time.Unix(startTimeUnix, 0)
		}
	}

	if endTimeStr != "" {
		if endTimeUnix, err := strconv.ParseInt(endTimeStr, 10, 64); err == nil {
			endTime = time.Unix(endTimeUnix, 0)
		}
	}

	// Get trades for this rule
	trades := h.getTradesForRule(ruleID, startTime, endTime)

	if len(trades) == 0 {
		http.Error(w, "No trades found for this rule in the specified time range", http.StatusNotFound)
		return
	}

	// Calculate full metrics including timeline
	metrics := calculateRuleMetrics(ruleID, trades, 0.02)
	metrics.Timeline = calculateTimeline(trades)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)

	log.Printf("[AnalyticsAPI] Rule metrics for %s: %d trades, Sharpe: %.2f, PnL: %.2f",
		ruleID, metrics.TradeCount, metrics.SharpeRatio, metrics.TotalPnL)
}

// HandleCalculateMetrics calculates metrics for a given set of trades
// POST /api/analytics/rules/calculate
func (h *APIHandler) HandleCalculateMetrics(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Verify admin authentication
	if !h.isAdminUser(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CalculateMetricsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Trades) == 0 {
		http.Error(w, "No trades provided", http.StatusBadRequest)
		return
	}

	// Convert to internal trade format
	trades := make([]core.Trade, len(req.Trades))
	for i, td := range req.Trades {
		trades[i] = core.Trade{
			RealizedPnL: td.PnL,
			ExecutedAt:  td.Timestamp,
			Volume:      td.Volume,
		}
	}

	// Calculate metrics
	sharpe := calculateSharpeRatio(trades, req.RiskFreeRate)
	profitFactor := calculateProfitFactor(trades)
	maxDrawdown := calculateMaxDrawdown(trades)
	winRate := calculateWinRate(trades)
	consecutiveWins, consecutiveLosses := calculateConsecutiveWinsLosses(trades)

	response := CalculateMetricsResponse{
		SharpeRatio:       sharpe,
		ProfitFactor:      profitFactor,
		MaxDrawdown:       maxDrawdown,
		WinRate:           winRate,
		ConsecutiveWins:   consecutiveWins,
		ConsecutiveLosses: consecutiveLosses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("[AnalyticsAPI] Calculated metrics: Sharpe=%.2f, PF=%.2f, DD=%.2f%%",
		sharpe, profitFactor, maxDrawdown*100)
}

// Helper functions

// getTradesForRule retrieves trades associated with a routing rule
func (h *APIHandler) getTradesForRule(_ string, startTime, endTime time.Time) []core.Trade {
	// In a real implementation, this would query a database for trades
	// that were routed using this specific rule.
	// For now, we'll use all trades from the engine and filter by time.

	// Note: This is a simplified implementation. In production, you would:
	// 1. Store rule_id with each trade in the database
	// 2. Query trades WHERE rule_id = ? AND timestamp BETWEEN ? AND ?

	allTrades := h.getAllTrades()

	filteredTrades := make([]core.Trade, 0)
	for _, trade := range allTrades {
		// Apply time filters
		if !startTime.IsZero() && trade.ExecutedAt.Before(startTime) {
			continue
		}
		if !endTime.IsZero() && trade.ExecutedAt.After(endTime) {
			continue
		}

		filteredTrades = append(filteredTrades, trade)
	}

	return filteredTrades
}

// getAllTrades retrieves all trades from all accounts
func (h *APIHandler) getAllTrades() []core.Trade {
	// This is a simplified implementation
	// In production, this would query from a database
	allTrades := make([]core.Trade, 0)

	// For demonstration, we'll use a placeholder
	// In real implementation, you would iterate through accounts or query DB

	return allTrades
}

// calculateRuleMetrics calculates all metrics for a rule
func calculateRuleMetrics(ruleID string, trades []core.Trade, riskFreeRate float64) RuleEffectivenessMetrics {
	if len(trades) == 0 {
		return RuleEffectivenessMetrics{
			RuleID:     ruleID,
			TradeCount: 0,
		}
	}

	totalPnL := 0.0
	for _, trade := range trades {
		totalPnL += trade.RealizedPnL
	}

	avgDuration := calculateAvgTradeDuration(trades)
	consecutiveWins, consecutiveLosses := calculateConsecutiveWinsLosses(trades)

	return RuleEffectivenessMetrics{
		RuleID:            ruleID,
		SharpeRatio:       calculateSharpeRatio(trades, riskFreeRate),
		ProfitFactor:      calculateProfitFactor(trades),
		MaxDrawdown:       calculateMaxDrawdown(trades),
		TotalPnL:          totalPnL,
		TradeCount:        len(trades),
		WinRate:           calculateWinRate(trades),
		AvgTradeDuration:  avgDuration,
		ConsecutiveWins:   consecutiveWins,
		ConsecutiveLosses: consecutiveLosses,
	}
}

// calculateSharpeRatio calculates the Sharpe Ratio
// Sharpe Ratio = (Average Return - Risk Free Rate) / Standard Deviation of Returns
func calculateSharpeRatio(trades []core.Trade, riskFreeRate float64) float64 {
	if len(trades) < 2 {
		return 0.0
	}

	// Calculate returns for each trade
	returns := make([]float64, len(trades))
	totalReturn := 0.0

	for i, trade := range trades {
		returns[i] = trade.RealizedPnL
		totalReturn += trade.RealizedPnL
	}

	// Calculate average return
	avgReturn := totalReturn / float64(len(trades))

	// Calculate standard deviation
	variance := 0.0
	for _, ret := range returns {
		diff := ret - avgReturn
		variance += diff * diff
	}
	variance /= float64(len(trades))
	stdDev := math.Sqrt(variance)

	// Avoid division by zero
	if stdDev == 0 {
		return 0.0
	}

	// Sharpe Ratio
	// Assuming riskFreeRate is annual, convert to per-trade rate
	periodicRiskFreeRate := riskFreeRate / 252.0 // Assuming 252 trading days
	sharpe := (avgReturn - periodicRiskFreeRate) / stdDev

	// Annualize the Sharpe Ratio (multiply by sqrt of number of periods per year)
	// Assuming daily trades
	return sharpe * math.Sqrt(252)
}

// calculateProfitFactor calculates the Profit Factor
// Profit Factor = Gross Profit / Gross Loss
func calculateProfitFactor(trades []core.Trade) float64 {
	grossProfit := 0.0
	grossLoss := 0.0

	for _, trade := range trades {
		if trade.RealizedPnL > 0 {
			grossProfit += trade.RealizedPnL
		} else if trade.RealizedPnL < 0 {
			grossLoss += math.Abs(trade.RealizedPnL)
		}
	}

	// Avoid division by zero
	if grossLoss == 0 {
		if grossProfit > 0 {
			return math.Inf(1) // Infinite profit factor (no losses)
		}
		return 0.0 // No trades or all breakeven
	}

	return grossProfit / grossLoss
}

// calculateMaxDrawdown calculates the maximum drawdown
// Max Drawdown = Maximum peak-to-trough decline in equity
func calculateMaxDrawdown(trades []core.Trade) float64 {
	if len(trades) == 0 {
		return 0.0
	}

	// Sort trades by execution time
	sortedTrades := make([]core.Trade, len(trades))
	copy(sortedTrades, trades)
	sort.Slice(sortedTrades, func(i, j int) bool {
		return sortedTrades[i].ExecutedAt.Before(sortedTrades[j].ExecutedAt)
	})

	// Calculate running equity
	equity := 0.0
	peak := 0.0
	maxDrawdown := 0.0

	for _, trade := range sortedTrades {
		equity += trade.RealizedPnL

		// Update peak
		if equity > peak {
			peak = equity
		}

		// Calculate drawdown from peak
		drawdown := peak - equity

		// Update max drawdown
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	// Return as percentage of peak (or absolute if peak is 0)
	if peak > 0 {
		return maxDrawdown / peak
	}

	return maxDrawdown
}

// calculateWinRate calculates the win rate percentage
func calculateWinRate(trades []core.Trade) float64 {
	if len(trades) == 0 {
		return 0.0
	}

	winningTrades := 0
	for _, trade := range trades {
		if trade.RealizedPnL > 0 {
			winningTrades++
		}
	}

	return (float64(winningTrades) / float64(len(trades))) * 100.0
}

// calculateAvgTradeDuration calculates average trade duration in seconds
func calculateAvgTradeDuration(_ []core.Trade) float64 {
	// This is a placeholder - actual implementation would need position open/close times
	// For now, return 0 since we don't have duration data in Trade struct
	return 0.0
}

// calculateConsecutiveWinsLosses calculates max consecutive wins and losses
func calculateConsecutiveWinsLosses(trades []core.Trade) (int, int) {
	if len(trades) == 0 {
		return 0, 0
	}

	// Sort trades by execution time
	sortedTrades := make([]core.Trade, len(trades))
	copy(sortedTrades, trades)
	sort.Slice(sortedTrades, func(i, j int) bool {
		return sortedTrades[i].ExecutedAt.Before(sortedTrades[j].ExecutedAt)
	})

	maxConsecutiveWins := 0
	maxConsecutiveLosses := 0
	currentWinStreak := 0
	currentLossStreak := 0

	for _, trade := range sortedTrades {
		if trade.RealizedPnL > 0 {
			// Winning trade
			currentWinStreak++
			currentLossStreak = 0
			if currentWinStreak > maxConsecutiveWins {
				maxConsecutiveWins = currentWinStreak
			}
		} else if trade.RealizedPnL < 0 {
			// Losing trade
			currentLossStreak++
			currentWinStreak = 0
			if currentLossStreak > maxConsecutiveLosses {
				maxConsecutiveLosses = currentLossStreak
			}
		}
		// Skip breakeven trades (PnL == 0)
	}

	return maxConsecutiveWins, maxConsecutiveLosses
}

// calculateTimeline generates equity curve timeline
func calculateTimeline(trades []core.Trade) []TimelineDataPoint {
	if len(trades) == 0 {
		return []TimelineDataPoint{}
	}

	// Sort trades by execution time
	sortedTrades := make([]core.Trade, len(trades))
	copy(sortedTrades, trades)
	sort.Slice(sortedTrades, func(i, j int) bool {
		return sortedTrades[i].ExecutedAt.Before(sortedTrades[j].ExecutedAt)
	})

	timeline := make([]TimelineDataPoint, len(sortedTrades))
	equity := 0.0

	for i, trade := range sortedTrades {
		equity += trade.RealizedPnL
		timeline[i] = TimelineDataPoint{
			Timestamp: trade.ExecutedAt,
			PnL:       trade.RealizedPnL,
			Equity:    equity,
		}
	}

	return timeline
}
