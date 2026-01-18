package cbook

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// CBookEngine is the main orchestrator for the C-Book hybrid routing system
type CBookEngine struct {
	mu sync.RWMutex

	// Core components
	profileEngine    *ClientProfileEngine
	routingEngine    *RoutingEngine
	mlPredictor      *MLPredictor
	analytics        *RoutingAnalytics
	compliance       *ComplianceEngine

	// Configuration
	mlEnabled        bool
	autoLearn        bool
	strictCompliance bool

	// Statistics
	totalDecisions   int64
	startTime        time.Time
}

// NewCBookEngine creates a new C-Book hybrid routing engine
func NewCBookEngine() *CBookEngine {
	profileEngine := NewClientProfileEngine()
	routingEngine := NewRoutingEngine(profileEngine)
	mlPredictor := NewMLPredictor()
	analytics := NewRoutingAnalytics(profileEngine, routingEngine)
	compliance := NewComplianceEngine()

	engine := &CBookEngine{
		profileEngine:    profileEngine,
		routingEngine:    routingEngine,
		mlPredictor:      mlPredictor,
		analytics:        analytics,
		compliance:       compliance,
		mlEnabled:        true,
		autoLearn:        true,
		strictCompliance: true,
		startTime:        time.Now(),
	}

	log.Println("[C-Book Engine] Initialized with ML prediction and compliance tracking")
	return engine
}

// RouteOrder is the main entry point for routing decisions
func (cbe *CBookEngine) RouteOrder(
	accountID int64,
	userID, username string,
	symbol, side string,
	volume float64,
	currentVolatility float64,
) (*RoutingDecision, error) {

	cbe.mu.Lock()
	cbe.totalDecisions++
	cbe.mu.Unlock()

	// 1. Get or create client profile
	profile := cbe.profileEngine.GetOrCreateProfile(accountID, userID, username)

	// 2. Get ML prediction if enabled
	var mlPrediction *ProfitabilityPrediction
	if cbe.mlEnabled {
		pred, err := cbe.mlPredictor.Predict(profile)
		if err == nil {
			mlPrediction = pred
			log.Printf("[C-Book] ML Prediction for %s: WinRate=%.1f%%, Risk=%.1f, Confidence=%.2f",
				username, pred.PredictedWinRate, pred.RiskScore, pred.Confidence)
		}
	}

	// 3. Make routing decision
	decision, err := cbe.routingEngine.Route(accountID, symbol, side, volume, currentVolatility)
	if err != nil {
		return nil, err
	}

	// 4. Override with ML recommendation if confidence is high
	if cbe.mlEnabled && mlPrediction != nil && mlPrediction.Confidence > 0.7 {
		// Use ML recommendation if it's more conservative (safer for broker)
		if mlPrediction.RiskScore > 60 && decision.BBookPercent > 50 {
			log.Printf("[C-Book] ML override: Increasing A-Book from %.0f%% to %.0f%% based on risk score",
				decision.ABookPercent, mlPrediction.RecommendedHedge)

			decision.Action = mlPrediction.RecommendedAction
			decision.ABookPercent = mlPrediction.RecommendedHedge
			decision.BBookPercent = 100 - mlPrediction.RecommendedHedge
			decision.Reason += fmt.Sprintf(" [ML override: risk=%.1f]", mlPrediction.RiskScore)
		}
	}

	// 5. Compliance check and audit logging
	if cbe.strictCompliance {
		cbe.compliance.LogRoutingDecision(
			accountID, username, symbol, side, volume,
			decision, profile, mlPrediction,
		)
	}

	log.Printf("[C-Book] ROUTED: Account=%d, %s %s %.2f lots -> %s (A-Book: %.0f%%, B-Book: %.0f%%) | %s",
		accountID, side, symbol, volume, decision.Action,
		decision.ABookPercent, decision.BBookPercent, decision.Reason)

	return decision, nil
}

// RecordTrade records a completed trade for learning and analytics
func (cbe *CBookEngine) RecordTrade(accountID int64, trade TradeRecord) {
	// 1. Update client profile
	cbe.profileEngine.RecordTrade(accountID, trade)

	// 2. Get updated profile
	profile, exists := cbe.profileEngine.GetProfile(accountID)
	if !exists {
		return
	}

	// 3. Train ML model if auto-learning enabled
	if cbe.autoLearn && cbe.mlEnabled {
		actualWinRate := profile.WinRate
		actualPnL := trade.PnL
		cbe.mlPredictor.Train(profile, actualWinRate, actualPnL)
	}

	// 4. Record for analytics (need to find corresponding decision)
	// Simplified: assuming we can match trades to decisions
	// In production, would need to store decision IDs with trades

	log.Printf("[C-Book] Recorded trade for account %d: PnL=%.2f, WinRate=%.1f%%",
		accountID, trade.PnL, profile.WinRate)
}

// RecordTradeOutcome records trade outcome for compliance and analytics
func (cbe *CBookEngine) RecordTradeOutcome(accountID int64, tradeID int64, decision *RoutingDecision, outcome *TradeOutcome) {
	// 1. Record in compliance engine
	if cbe.strictCompliance {
		cbe.compliance.RecordTradeOutcome(accountID, tradeID, outcome)
	}

	// 2. Record in analytics engine
	profile, _ := cbe.profileEngine.GetProfile(accountID)
	username := ""
	if profile != nil {
		username = profile.Username
	}

	cbe.analytics.RecordTradeOutcome(accountID, username, decision, outcome.RealizedPnL)

	log.Printf("[C-Book] Recorded outcome for account %d: PnL=%.2f, Optimal=%v",
		accountID, outcome.RealizedPnL, outcome.WasOptimal)
}

// UpdateExposure updates symbol exposure tracking
func (cbe *CBookEngine) UpdateExposure(symbol, side string, volume float64, action RoutingAction, bBookPercent float64) {
	// Only update exposure for B-Book portion
	if action == ActionBBook || action == ActionPartialHedge {
		bBookVolume := volume
		if action == ActionPartialHedge {
			bBookVolume = volume * (bBookPercent / 100)
		}

		cbe.routingEngine.UpdateExposure(symbol, side, bBookVolume)
	}
}

// GetClientProfile returns a client's profile and classification
func (cbe *CBookEngine) GetClientProfile(accountID int64) (*ClientProfile, bool) {
	return cbe.profileEngine.GetProfile(accountID)
}

// GetAllProfiles returns all client profiles
func (cbe *CBookEngine) GetAllProfiles() []*ClientProfile {
	return cbe.profileEngine.GetAllProfiles()
}

// GetProfilesByClassification returns profiles by classification
func (cbe *CBookEngine) GetProfilesByClassification(classification ClientClassification) []*ClientProfile {
	return cbe.profileEngine.GetProfilesByClassification(classification)
}

// GetExposure returns current exposure for a symbol
func (cbe *CBookEngine) GetExposure(symbol string) *SymbolExposure {
	return cbe.routingEngine.GetExposure(symbol)
}

// GetAllExposures returns all symbol exposures
func (cbe *CBookEngine) GetAllExposures() map[string]*SymbolExposure {
	cbe.routingEngine.mu.RLock()
	defer cbe.routingEngine.mu.RUnlock()

	exposures := make(map[string]*SymbolExposure)
	for symbol, exp := range cbe.routingEngine.symbolExposure {
		exposures[symbol] = exp
	}
	return exposures
}

// AddRoutingRule adds a manual routing rule
func (cbe *CBookEngine) AddRoutingRule(rule *RoutingRule) {
	cbe.routingEngine.AddRule(rule)
}

// UpdateRoutingRule updates an existing rule
func (cbe *CBookEngine) UpdateRoutingRule(ruleID string, rule *RoutingRule) error {
	return cbe.routingEngine.UpdateRule(ruleID, rule)
}

// DeleteRoutingRule removes a rule
func (cbe *CBookEngine) DeleteRoutingRule(ruleID string) error {
	return cbe.routingEngine.DeleteRule(ruleID)
}

// GetRoutingRules returns all routing rules
func (cbe *CBookEngine) GetRoutingRules() []*RoutingRule {
	return cbe.routingEngine.GetRules()
}

// SetExposureLimit sets exposure limits for a symbol
func (cbe *CBookEngine) SetExposureLimit(symbol string, limit *ExposureLimit) {
	cbe.routingEngine.SetExposureLimit(symbol, limit)
}

// GetRoutingStats returns comprehensive routing statistics
func (cbe *CBookEngine) GetRoutingStats() map[string]interface{} {
	stats := cbe.routingEngine.GetRoutingStats()

	cbe.mu.RLock()
	stats["total_decisions"] = cbe.totalDecisions
	stats["uptime_hours"] = time.Since(cbe.startTime).Hours()
	cbe.mu.RUnlock()

	return stats
}

// GetMLStats returns ML model statistics
func (cbe *CBookEngine) GetMLStats() map[string]interface{} {
	return cbe.mlPredictor.GetModelStats()
}

// GetAnalytics returns analytics data
func (cbe *CBookEngine) GetClientPerformance(accountID int64) (*ClientPerformance, bool) {
	return cbe.analytics.GetClientPerformance(accountID)
}

// GetTopPerformers returns top performing clients (by broker profit)
func (cbe *CBookEngine) GetTopPerformers(limit int) []*ClientPerformance {
	return cbe.analytics.GetTopPerformers(limit)
}

// GetSymbolPerformance returns performance by symbol
func (cbe *CBookEngine) GetSymbolPerformance() map[string]*SymbolPerformance {
	return cbe.analytics.GetSymbolPerformance()
}

// GetRoutingEffectiveness calculates routing quality metrics
func (cbe *CBookEngine) GetRoutingEffectiveness(period string) *RoutingEffectiveness {
	return cbe.analytics.CalculateRoutingEffectiveness(period)
}

// GeneratePnLReport creates P&L report
func (cbe *CBookEngine) GeneratePnLReport() map[string]interface{} {
	return cbe.analytics.GeneratePnLReport()
}

// GetRecommendedAdjustments returns optimization suggestions
func (cbe *CBookEngine) GetRecommendedAdjustments() []string {
	return cbe.analytics.GetRecommendedAdjustments()
}

// GetAuditLogs retrieves compliance audit logs
func (cbe *CBookEngine) GetAuditLogs(accountID int64, startTime, endTime time.Time, limit int) []*AuditLog {
	return cbe.compliance.GetAuditLogs(accountID, startTime, endTime, limit)
}

// GetComplianceAlerts retrieves compliance alerts
func (cbe *CBookEngine) GetComplianceAlerts(severity string, resolved bool, limit int) []ComplianceAlert {
	return cbe.compliance.GetAlerts(severity, resolved, limit)
}

// ResolveAlert marks a compliance alert as resolved
func (cbe *CBookEngine) ResolveAlert(alertID int64) error {
	return cbe.compliance.ResolveAlert(alertID)
}

// GenerateBestExecutionReport creates regulatory report
func (cbe *CBookEngine) GenerateBestExecutionReport(startTime, endTime time.Time) map[string]interface{} {
	return cbe.compliance.GenerateBestExecutionReport(startTime, endTime)
}

// GenerateRoutingAuditReport creates audit report
func (cbe *CBookEngine) GenerateRoutingAuditReport(accountID int64) map[string]interface{} {
	return cbe.compliance.GenerateRoutingAuditReport(accountID)
}

// ExportAuditTrail exports audit trail for regulators
func (cbe *CBookEngine) ExportAuditTrail(startTime, endTime time.Time) ([]byte, error) {
	return cbe.compliance.ExportAuditTrail(startTime, endTime)
}

// GetComplianceStats returns compliance statistics
func (cbe *CBookEngine) GetComplianceStats() map[string]interface{} {
	return cbe.compliance.GetComplianceStats()
}

// UpdateClassificationThresholds updates client classification thresholds
func (cbe *CBookEngine) UpdateClassificationThresholds(toxicWinRate, toxicSharpe, proWinRate float64, minTrades int64) {
	cbe.profileEngine.UpdateThresholds(toxicWinRate, toxicSharpe, proWinRate, minTrades)
}

// EnableML enables or disables ML predictions
func (cbe *CBookEngine) EnableML(enabled bool) {
	cbe.mu.Lock()
	defer cbe.mu.Unlock()
	cbe.mlEnabled = enabled
	log.Printf("[C-Book] ML predictions %s", map[bool]string{true: "enabled", false: "disabled"}[enabled])
}

// EnableAutoLearning enables or disables automatic ML training
func (cbe *CBookEngine) EnableAutoLearning(enabled bool) {
	cbe.mu.Lock()
	defer cbe.mu.Unlock()
	cbe.autoLearn = enabled
	log.Printf("[C-Book] Auto-learning %s", map[bool]string{true: "enabled", false: "disabled"}[enabled])
}

// EnableStrictCompliance enables or disables strict compliance mode
func (cbe *CBookEngine) EnableStrictCompliance(enabled bool) {
	cbe.mu.Lock()
	defer cbe.mu.Unlock()
	cbe.strictCompliance = enabled
	log.Printf("[C-Book] Strict compliance %s", map[bool]string{true: "enabled", false: "disabled"}[enabled])
}

// ExportMLModel exports ML model for backup
func (cbe *CBookEngine) ExportMLModel() ([]byte, error) {
	return cbe.mlPredictor.ExportModel()
}

// ImportMLModel imports ML model from backup
func (cbe *CBookEngine) ImportMLModel(data []byte) error {
	return cbe.mlPredictor.ImportModel(data)
}

// GetDashboardData returns comprehensive dashboard data
func (cbe *CBookEngine) GetDashboardData() map[string]interface{} {
	dashboard := make(map[string]interface{})

	// Overview stats
	cbe.mu.RLock()
	dashboard["total_decisions"] = cbe.totalDecisions
	dashboard["uptime_hours"] = time.Since(cbe.startTime).Hours()
	dashboard["ml_enabled"] = cbe.mlEnabled
	dashboard["auto_learn"] = cbe.autoLearn
	cbe.mu.RUnlock()

	// Client breakdown
	profiles := cbe.profileEngine.GetAllProfiles()
	classificationCounts := make(map[ClientClassification]int)
	for _, profile := range profiles {
		classificationCounts[profile.Classification]++
	}
	dashboard["client_classifications"] = classificationCounts
	dashboard["total_clients"] = len(profiles)

	// Routing stats
	dashboard["routing_stats"] = cbe.GetRoutingStats()

	// ML stats
	if cbe.mlEnabled {
		dashboard["ml_stats"] = cbe.GetMLStats()
	}

	// Compliance stats
	dashboard["compliance_stats"] = cbe.GetComplianceStats()

	// Top performers
	dashboard["top_performers"] = cbe.GetTopPerformers(10)

	// Effectiveness metrics
	dashboard["effectiveness_1w"] = cbe.GetRoutingEffectiveness("1W")

	// Recommendations
	dashboard["recommendations"] = cbe.GetRecommendedAdjustments()

	return dashboard
}

// GetRoutingEngine returns the routing engine for external access
func (cbe *CBookEngine) GetRoutingEngine() *RoutingEngine {
	return cbe.routingEngine
}

// GetDecisionHistory returns recent routing decisions for analytics
func (cbe *CBookEngine) GetDecisionHistory(limit int) []RoutingDecision {
	return cbe.routingEngine.GetDecisionHistory(limit)
}
