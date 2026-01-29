package cbook

import (
	"math"
	"sync"
	"time"
)

// RoutingAnalytics provides comprehensive analytics and reporting
type RoutingAnalytics struct {
	mu sync.RWMutex

	profileEngine *ClientProfileEngine
	routingEngine *RoutingEngine

	// Performance tracking
	performanceMetrics map[int64]*ClientPerformance // accountID -> performance
}

// ClientPerformance tracks P&L by routing decision
type ClientPerformance struct {
	AccountID int64  `json:"accountId"`
	Username  string `json:"username"`

	// P&L breakdown
	TotalPnL        float64 `json:"totalPnL"`
	ABookPnL        float64 `json:"aBookPnL"`  // Loss (broker hedged)
	BBookPnL        float64 `json:"bBookPnL"`  // Profit (broker internalized)
	BrokerNetPnL    float64 `json:"brokerNetPnL"` // B-Book wins - A-Book costs

	// Trade counts
	TotalTrades     int64 `json:"totalTrades"`
	ABookTrades     int64 `json:"aBookTrades"`
	BBookTrades     int64 `json:"bBookTrades"`
	PartialHedges   int64 `json:"partialHedges"`
	RejectedTrades  int64 `json:"rejectedTrades"`

	// Profitability metrics
	WinRate         float64 `json:"winRate"`
	AvgWin          float64 `json:"avgWin"`
	AvgLoss         float64 `json:"avgLoss"`
	ProfitFactor    float64 `json:"profitFactor"`

	// Risk metrics
	MaxDrawdown     float64 `json:"maxDrawdown"`
	SharpeRatio     float64 `json:"sharpeRatio"`
	VaR95           float64 `json:"var95"` // 95% Value at Risk

	LastUpdated     time.Time `json:"lastUpdated"`
}

// SymbolPerformance tracks performance by symbol
type SymbolPerformance struct {
	Symbol          string  `json:"symbol"`
	BrokerNetPnL    float64 `json:"brokerNetPnL"`
	TotalVolume     float64 `json:"totalVolume"`
	ABookPercent    float64 `json:"aBookPercent"`
	BBookPercent    float64 `json:"bBookPercent"`
	ClientCount     int     `json:"clientCount"`
}

// RoutingEffectiveness measures routing accuracy
type RoutingEffectiveness struct {
	Period          string    `json:"period"` // "1D", "1W", "1M"

	// Classification accuracy
	RetailAccuracy  float64   `json:"retailAccuracy"`  // % of retail clients that lost
	ProAccuracy     float64   `json:"proAccuracy"`     // % of pro clients that won
	ToxicAccuracy   float64   `json:"toxicAccuracy"`   // % of toxic correctly identified

	// Routing profitability
	OptimalDecisions      int64   `json:"optimalDecisions"`      // Correct routing choices
	SuboptimalDecisions   int64   `json:"suboptimalDecisions"`   // Wrong routing choices
	OptimalityRate        float64 `json:"optimalityRate"`        // %

	// Cost analysis
	HedgingCosts          float64 `json:"hedgingCosts"`
	BBookProfits          float64 `json:"bBookProfits"`
	NetBrokerProfit       float64 `json:"netBrokerProfit"`
	ROI                   float64 `json:"roi"` // Return on hedging costs

	CalculatedAt          time.Time `json:"calculatedAt"`
}

// NewRoutingAnalytics creates an analytics engine
func NewRoutingAnalytics(profileEngine *ClientProfileEngine, routingEngine *RoutingEngine) *RoutingAnalytics {
	return &RoutingAnalytics{
		profileEngine:      profileEngine,
		routingEngine:      routingEngine,
		performanceMetrics: make(map[int64]*ClientPerformance),
	}
}

// RecordTradeOutcome records trade result for analytics
func (ra *RoutingAnalytics) RecordTradeOutcome(accountID int64, username string, decision *RoutingDecision, pnl float64) {
	ra.mu.Lock()
	defer ra.mu.Unlock()

	perf, exists := ra.performanceMetrics[accountID]
	if !exists {
		perf = &ClientPerformance{
			AccountID: accountID,
			Username:  username,
		}
		ra.performanceMetrics[accountID] = perf
	}

	perf.TotalPnL += pnl
	perf.TotalTrades++

	// Track by routing action
	switch decision.Action {
	case ActionABook:
		perf.ABookTrades++
		// A-Book: broker loses client's profit (hedging cost)
		perf.ABookPnL -= pnl
		perf.BrokerNetPnL -= pnl

	case ActionBBook:
		perf.BBookTrades++
		// B-Book: broker gains client's loss (or loses client's profit)
		perf.BBookPnL -= pnl  // Client profit = broker loss
		perf.BrokerNetPnL -= pnl

	case ActionPartialHedge:
		perf.PartialHedges++
		// Split based on percentages
		aBookPortion := pnl * (decision.ABookPercent / 100)
		bBookPortion := pnl * (decision.BBookPercent / 100)

		perf.ABookPnL -= aBookPortion
		perf.BBookPnL -= bBookPortion
		perf.BrokerNetPnL -= (aBookPortion + bBookPortion)

	case ActionReject:
		perf.RejectedTrades++
		// No P&L impact
	}

	perf.LastUpdated = time.Now()

	// Recalculate metrics
	ra.recalculatePerformanceMetrics(perf)
}

// recalculatePerformanceMetrics updates derived metrics
func (ra *RoutingAnalytics) recalculatePerformanceMetrics(perf *ClientPerformance) {
	// Get full trade history from profile engine
	profile, exists := ra.profileEngine.GetProfile(perf.AccountID)
	if !exists || len(profile.tradeHistory) == 0 {
		return
	}

	// Win rate
	wins := 0
	losses := 0
	var totalWins, totalLosses float64

	for _, trade := range profile.tradeHistory {
		if trade.IsWinner {
			wins++
			totalWins += trade.PnL
		} else {
			losses++
			totalLosses += math.Abs(trade.PnL)
		}
	}

	if wins+losses > 0 {
		perf.WinRate = (float64(wins) / float64(wins+losses)) * 100
	}

	if wins > 0 {
		perf.AvgWin = totalWins / float64(wins)
	}
	if losses > 0 {
		perf.AvgLoss = totalLosses / float64(losses)
	}

	// Profit factor
	if totalLosses > 0 {
		perf.ProfitFactor = totalWins / totalLosses
	}

	// Sharpe ratio
	perf.SharpeRatio = calculateSharpeRatio(profile.tradeHistory)

	// Max drawdown
	perf.MaxDrawdown = calculateMaxDrawdown(profile.tradeHistory)

	// VaR 95%
	perf.VaR95 = calculateVaR(profile.tradeHistory, 0.95)
}

// calculateVaR calculates Value at Risk at given confidence level
func calculateVaR(trades []TradeRecord, confidence float64) float64 {
	if len(trades) == 0 {
		return 0
	}

	// Extract P&L values
	pnls := make([]float64, len(trades))
	for i, trade := range trades {
		pnls[i] = trade.PnL
	}

	// Sort P&L (ascending)
	for i := 0; i < len(pnls); i++ {
		for j := i + 1; j < len(pnls); j++ {
			if pnls[j] < pnls[i] {
				pnls[i], pnls[j] = pnls[j], pnls[i]
			}
		}
	}

	// Find VaR at confidence level
	index := int(float64(len(pnls)) * (1 - confidence))
	if index >= len(pnls) {
		index = len(pnls) - 1
	}

	return math.Abs(pnls[index])
}

// GetClientPerformance returns performance metrics for a client
func (ra *RoutingAnalytics) GetClientPerformance(accountID int64) (*ClientPerformance, bool) {
	ra.mu.RLock()
	defer ra.mu.RUnlock()

	perf, exists := ra.performanceMetrics[accountID]
	return perf, exists
}

// GetTopPerformers returns top clients by broker profit
func (ra *RoutingAnalytics) GetTopPerformers(limit int) []*ClientPerformance {
	ra.mu.RLock()
	defer ra.mu.RUnlock()

	// Convert to slice
	perfs := make([]*ClientPerformance, 0, len(ra.performanceMetrics))
	for _, perf := range ra.performanceMetrics {
		perfs = append(perfs, perf)
	}

	// Sort by broker net P&L (descending)
	for i := 0; i < len(perfs); i++ {
		for j := i + 1; j < len(perfs); j++ {
			if perfs[j].BrokerNetPnL > perfs[i].BrokerNetPnL {
				perfs[i], perfs[j] = perfs[j], perfs[i]
			}
		}
	}

	if limit > 0 && limit < len(perfs) {
		perfs = perfs[:limit]
	}

	return perfs
}

// GetSymbolPerformance returns performance breakdown by symbol
func (ra *RoutingAnalytics) GetSymbolPerformance() map[string]*SymbolPerformance {
	ra.mu.RLock()
	defer ra.mu.RUnlock()

	symbolPerf := make(map[string]*SymbolPerformance)

	// Aggregate from all client profiles
	profiles := ra.profileEngine.GetAllProfiles()

	for _, profile := range profiles {
		_, exists := ra.performanceMetrics[profile.AccountID]
		if !exists {
			continue
		}

		for _, trade := range profile.tradeHistory {
			perf, exists := symbolPerf[trade.Symbol]
			if !exists {
				perf = &SymbolPerformance{
					Symbol: trade.Symbol,
				}
				symbolPerf[trade.Symbol] = perf
			}

			// Add to symbol totals
			perf.BrokerNetPnL -= trade.PnL // Broker gains from client losses
			perf.TotalVolume += trade.Volume

			// Count unique clients
			// (simplified - would need set tracking in production)
			perf.ClientCount++
		}
	}

	// Calculate routing percentages from decisions
	// TODO: Symbol not in decision struct, would need to add it
	// Commented out incomplete implementation
	// decisions := ra.routingEngine.GetDecisionHistory(10000)
	// symbolRoutings := make(map[string]struct{ aBook, bBook, total int })
	// for _, decision := range decisions {
	//     // Calculate routing stats
	// }

	return symbolPerf
}

// CalculateRoutingEffectiveness evaluates routing quality
func (ra *RoutingAnalytics) CalculateRoutingEffectiveness(period string) *RoutingEffectiveness {
	ra.mu.RLock()
	defer ra.mu.RUnlock()

	effectiveness := &RoutingEffectiveness{
		Period:       period,
		CalculatedAt: time.Now(),
	}

	profiles := ra.profileEngine.GetAllProfiles()

	var retailWins, retailTotal int64
	var proWins, proTotal int64
	var toxicDetected, toxicTotal int64

	var totalHedgingCosts float64
	var totalBBookProfits float64

	for _, profile := range profiles {
		perf, exists := ra.performanceMetrics[profile.AccountID]
		if !exists {
			continue
		}

		// Classification accuracy
		switch profile.Classification {
		case ClassificationRetail:
			retailTotal++
			// Retail should lose money (good for broker)
			if perf.TotalPnL < 0 {
				retailWins++ // "Win" = correct classification
			}

		case ClassificationProfessional:
			proTotal++
			// Pro should win money
			if perf.TotalPnL > 0 {
				proWins++ // Correct classification
			}

		case ClassificationToxic:
			toxicTotal++
			// Toxic should be highly profitable
			if perf.TotalPnL > 0 && perf.WinRate > 55 {
				toxicDetected++ // Correctly identified toxic
			}
		}

		// P&L breakdown
		totalHedgingCosts += math.Abs(perf.ABookPnL) // Cost of hedging
		totalBBookProfits += perf.BBookPnL           // Profit from B-Book
	}

	// Calculate accuracy rates
	if retailTotal > 0 {
		effectiveness.RetailAccuracy = (float64(retailWins) / float64(retailTotal)) * 100
	}
	if proTotal > 0 {
		effectiveness.ProAccuracy = (float64(proWins) / float64(proTotal)) * 100
	}
	if toxicTotal > 0 {
		effectiveness.ToxicAccuracy = (float64(toxicDetected) / float64(toxicTotal)) * 100
	}

	// Routing optimality
	decisions := ra.routingEngine.GetDecisionHistory(10000)
	for _, decision := range decisions {
		// Simplified optimality check
		// In production, would compare actual outcome vs decision
		switch decision.Action {
		case ActionABook, ActionPartialHedge:
			// Conservative, usually optimal for high toxicity
			if decision.ToxicityScore > 50 {
				effectiveness.OptimalDecisions++
			} else {
				effectiveness.SuboptimalDecisions++
			}
		case ActionBBook:
			// Optimal for low toxicity retail
			if decision.ToxicityScore < 30 {
				effectiveness.OptimalDecisions++
			} else {
				effectiveness.SuboptimalDecisions++
			}
		}
	}

	total := effectiveness.OptimalDecisions + effectiveness.SuboptimalDecisions
	if total > 0 {
		effectiveness.OptimalityRate = (float64(effectiveness.OptimalDecisions) / float64(total)) * 100
	}

	// Cost analysis
	effectiveness.HedgingCosts = totalHedgingCosts
	effectiveness.BBookProfits = totalBBookProfits
	effectiveness.NetBrokerProfit = totalBBookProfits - totalHedgingCosts

	if totalHedgingCosts > 0 {
		effectiveness.ROI = (effectiveness.NetBrokerProfit / totalHedgingCosts) * 100
	}

	return effectiveness
}

// GeneratePnLReport creates a comprehensive P&L report
func (ra *RoutingAnalytics) GeneratePnLReport() map[string]interface{} {
	ra.mu.RLock()
	defer ra.mu.RUnlock()

	report := make(map[string]interface{})

	var totalBrokerPnL float64
	var totalABookPnL float64
	var totalBBookPnL float64
	var totalTrades int64

	clientBreakdown := make([]map[string]interface{}, 0)

	for _, perf := range ra.performanceMetrics {
		totalBrokerPnL += perf.BrokerNetPnL
		totalABookPnL += perf.ABookPnL
		totalBBookPnL += perf.BBookPnL
		totalTrades += perf.TotalTrades

		clientBreakdown = append(clientBreakdown, map[string]interface{}{
			"account_id":     perf.AccountID,
			"username":       perf.Username,
			"broker_net_pnl": perf.BrokerNetPnL,
			"total_trades":   perf.TotalTrades,
			"win_rate":       perf.WinRate,
		})
	}

	// Sort by broker P&L
	for i := 0; i < len(clientBreakdown); i++ {
		for j := i + 1; j < len(clientBreakdown); j++ {
			if clientBreakdown[j]["broker_net_pnl"].(float64) > clientBreakdown[i]["broker_net_pnl"].(float64) {
				clientBreakdown[i], clientBreakdown[j] = clientBreakdown[j], clientBreakdown[i]
			}
		}
	}

	report["total_broker_pnl"] = totalBrokerPnL
	report["total_abook_pnl"] = totalABookPnL
	report["total_bbook_pnl"] = totalBBookPnL
	report["total_trades"] = totalTrades
	report["client_breakdown"] = clientBreakdown
	report["generated_at"] = time.Now()

	return report
}

// GetRecommendedAdjustments suggests routing rule changes
func (ra *RoutingAnalytics) GetRecommendedAdjustments() []string {
	recommendations := make([]string, 0)

	effectiveness := ra.CalculateRoutingEffectiveness("1W")

	// Low retail accuracy - tighten classification
	if effectiveness.RetailAccuracy < 70 {
		recommendations = append(recommendations,
			"Retail classification accuracy is low. Consider tightening win rate thresholds.")
	}

	// Low toxic detection - increase vigilance
	if effectiveness.ToxicAccuracy < 80 {
		recommendations = append(recommendations,
			"Toxic client detection is suboptimal. Review toxicity score thresholds.")
	}

	// Negative ROI - too much hedging
	if effectiveness.ROI < 0 {
		recommendations = append(recommendations,
			"Hedging costs exceed B-Book profits. Reduce A-Book percentage for low-risk clients.")
	}

	// Low ROI - increase hedging
	if effectiveness.ROI > 200 {
		recommendations = append(recommendations,
			"Very high ROI suggests insufficient hedging. Increase A-Book exposure for pro clients.")
	}

	// Optimality rate issues
	if effectiveness.OptimalityRate < 75 {
		recommendations = append(recommendations,
			"Routing optimality is below target. Review and adjust routing rules.")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Routing performance is optimal. No changes recommended.")
	}

	return recommendations
}
