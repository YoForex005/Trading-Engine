package cbook

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// AuditLog records all routing decisions for compliance
type AuditLog struct {
	ID              int64             `json:"id"`
	Timestamp       time.Time         `json:"timestamp"`
	AccountID       int64             `json:"accountId"`
	Username        string            `json:"username"`
	Symbol          string            `json:"symbol"`
	Side            string            `json:"side"`
	Volume          float64           `json:"volume"`
	Decision        *RoutingDecision  `json:"decision"`
	ClientProfile   *ClientProfile    `json:"clientProfile,omitempty"`
	MLPrediction    *ProfitabilityPrediction `json:"mlPrediction,omitempty"`
	ActualOutcome   *TradeOutcome     `json:"actualOutcome,omitempty"`
	ComplianceFlags []string          `json:"complianceFlags,omitempty"`
}

// TradeOutcome records actual trade result for audit
type TradeOutcome struct {
	TradeID       int64     `json:"tradeId"`
	ClosedAt      time.Time `json:"closedAt"`
	ClosePrice    float64   `json:"closePrice"`
	RealizedPnL   float64   `json:"realizedPnL"`
	HoldTime      int64     `json:"holdTime"` // Seconds
	ExecutedRoute string    `json:"executedRoute"` // A_BOOK, B_BOOK, PARTIAL
	WasOptimal    bool      `json:"wasOptimal"` // Did we make the right decision?
}

// ComplianceEngine handles audit trails and regulatory compliance
type ComplianceEngine struct {
	mu sync.RWMutex

	// Audit trail
	auditLogs      []*AuditLog
	maxLogs        int
	nextLogID      int64

	// Compliance rules
	requireBestExecution bool
	maxClientExposure    float64 // Max B-Book exposure per client
	fairnessThreshold    float64 // Max allowed discrimination

	// Alerts
	alerts         []ComplianceAlert
	maxAlerts      int
}

// ComplianceAlert represents a compliance concern
type ComplianceAlert struct {
	ID          int64     `json:"id"`
	Severity    string    `json:"severity"` // INFO, WARNING, CRITICAL
	Type        string    `json:"type"`
	AccountID   int64     `json:"accountId"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"createdAt"`
	Resolved    bool      `json:"resolved"`
	ResolvedAt  *time.Time `json:"resolvedAt,omitempty"`
}

// NewComplianceEngine creates a compliance engine
func NewComplianceEngine() *ComplianceEngine {
	return &ComplianceEngine{
		auditLogs:            make([]*AuditLog, 0, 100000),
		maxLogs:              100000,
		nextLogID:            1,
		requireBestExecution: true,
		maxClientExposure:    10000, // $10k max B-Book exposure per client
		fairnessThreshold:    0.15,  // 15% max discrimination
		alerts:               make([]ComplianceAlert, 0, 10000),
		maxAlerts:            10000,
	}
}

// LogRoutingDecision records a routing decision in audit trail
func (ce *ComplianceEngine) LogRoutingDecision(
	accountID int64,
	username, symbol, side string,
	volume float64,
	decision *RoutingDecision,
	profile *ClientProfile,
	prediction *ProfitabilityPrediction,
) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	log := &AuditLog{
		ID:            ce.nextLogID,
		Timestamp:     time.Now(),
		AccountID:     accountID,
		Username:      username,
		Symbol:        symbol,
		Side:          side,
		Volume:        volume,
		Decision:      decision,
		ClientProfile: profile,
		MLPrediction:  prediction,
	}

	ce.nextLogID++

	// Check compliance
	flags := ce.checkCompliance(log, profile)
	log.ComplianceFlags = flags

	// Generate alerts if needed
	if len(flags) > 0 {
		ce.generateAlerts(log, flags)
	}

	// Store log
	ce.auditLogs = append(ce.auditLogs, log)
	if len(ce.auditLogs) > ce.maxLogs {
		ce.auditLogs = ce.auditLogs[1:]
	}
}

// RecordTradeOutcome updates audit log with actual outcome
func (ce *ComplianceEngine) RecordTradeOutcome(accountID int64, tradeID int64, outcome *TradeOutcome) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	// Find the corresponding audit log
	// In production, would use indexed lookup
	for i := len(ce.auditLogs) - 1; i >= 0; i-- {
		log := ce.auditLogs[i]
		if log.AccountID == accountID {
			// Simple heuristic: most recent trade for this account
			log.ActualOutcome = outcome

			// Evaluate if decision was optimal
			outcome.WasOptimal = ce.evaluateDecisionQuality(log)

			// If decision was suboptimal, create alert
			if !outcome.WasOptimal {
				ce.createSuboptimalDecisionAlert(log)
			}

			return nil
		}
	}

	return fmt.Errorf("audit log not found for account %d", accountID)
}

// checkCompliance validates routing decision against compliance rules
func (ce *ComplianceEngine) checkCompliance(log *AuditLog, profile *ClientProfile) []string {
	flags := make([]string, 0)

	// 1. Best execution check
	if ce.requireBestExecution {
		if log.Decision.Action == ActionBBook && log.Volume >= 5 {
			flags = append(flags, "LARGE_BBOOK_ORDER: Large order routed to B-Book may violate best execution")
		}
	}

	// 2. Discrimination check
	if profile != nil {
		// Check if similar clients are treated differently
		// This is a simplified version; full implementation would compare across clients
		if log.Decision.Action == ActionReject && profile.ToxicityScore < 70 {
			flags = append(flags, "QUESTIONABLE_REJECT: Client rejected with moderate toxicity score")
		}
	}

	// 3. Excessive B-Book exposure
	// Would need to check total B-Book exposure for this client
	// Simplified check here
	if log.Decision.Action == ActionBBook && log.Volume > 10 {
		flags = append(flags, "EXCESSIVE_BBOOK: Large B-Book exposure for single client")
	}

	// 4. Toxic client in B-Book
	if profile != nil && profile.ToxicityScore > 60 && log.Decision.BBookPercent > 50 {
		flags = append(flags, "TOXIC_BBOOK: High-toxicity client routed to B-Book")
	}

	// 5. Consistency check (same client shouldn't have wildly different routing)
	// Would compare with historical routing for this client
	// Simplified check here

	return flags
}

// evaluateDecisionQuality determines if routing decision was optimal
func (ce *ComplianceEngine) evaluateDecisionQuality(log *AuditLog) bool {
	if log.ActualOutcome == nil {
		return true // Can't evaluate yet
	}

	outcome := log.ActualOutcome
	decision := log.Decision

	// Client made profit - should have been A-Book
	if outcome.RealizedPnL > 0 {
		// If we B-Booked a profitable trade, that's a loss for broker
		if decision.Action == ActionBBook || decision.BBookPercent > 50 {
			return false // Suboptimal - should have hedged
		}
		return true
	}

	// Client made loss - B-Book was good
	if outcome.RealizedPnL < 0 {
		// If we A-Booked a losing trade, we missed profit opportunity
		if decision.Action == ActionABook || decision.ABookPercent > 70 {
			// Only suboptimal if it was a clearly retail client
			if log.ClientProfile != nil && log.ClientProfile.Classification == ClassificationRetail {
				return false // Suboptimal - should have B-Booked
			}
		}
		return true
	}

	return true // Break-even trades are fine
}

// generateAlerts creates compliance alerts from flags
func (ce *ComplianceEngine) generateAlerts(log *AuditLog, flags []string) {
	for _, flag := range flags {
		severity := "WARNING"

		// Determine severity
		if flag == "LARGE_BBOOK_ORDER" || flag == "EXCESSIVE_BBOOK" {
			severity = "CRITICAL"
		} else if flag == "TOXIC_BBOOK" {
			severity = "WARNING"
		} else {
			severity = "INFO"
		}

		alert := ComplianceAlert{
			ID:        int64(len(ce.alerts) + 1),
			Severity:  severity,
			Type:      flag,
			AccountID: log.AccountID,
			Message:   fmt.Sprintf("Compliance flag: %s for account %d", flag, log.AccountID),
			CreatedAt: time.Now(),
			Resolved:  false,
		}

		ce.alerts = append(ce.alerts, alert)
		if len(ce.alerts) > ce.maxAlerts {
			ce.alerts = ce.alerts[1:]
		}
	}
}

// createSuboptimalDecisionAlert creates alert for bad routing decision
func (ce *ComplianceEngine) createSuboptimalDecisionAlert(log *AuditLog) {
	alert := ComplianceAlert{
		ID:        int64(len(ce.alerts) + 1),
		Severity:  "WARNING",
		Type:      "SUBOPTIMAL_ROUTING",
		AccountID: log.AccountID,
		Message: fmt.Sprintf("Suboptimal routing decision for account %d: Client PnL=%.2f, Routed to %s",
			log.AccountID, log.ActualOutcome.RealizedPnL, log.Decision.Action),
		CreatedAt: time.Now(),
		Resolved:  false,
	}

	ce.alerts = append(ce.alerts, alert)
	if len(ce.alerts) > ce.maxAlerts {
		ce.alerts = ce.alerts[1:]
	}
}

// GetAuditLogs retrieves audit logs with filters
func (ce *ComplianceEngine) GetAuditLogs(accountID int64, startTime, endTime time.Time, limit int) []*AuditLog {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	logs := make([]*AuditLog, 0)

	for i := len(ce.auditLogs) - 1; i >= 0 && len(logs) < limit; i-- {
		log := ce.auditLogs[i]

		// Apply filters
		if accountID > 0 && log.AccountID != accountID {
			continue
		}

		if !startTime.IsZero() && log.Timestamp.Before(startTime) {
			continue
		}

		if !endTime.IsZero() && log.Timestamp.After(endTime) {
			continue
		}

		logs = append(logs, log)
	}

	return logs
}

// GetAlerts retrieves compliance alerts
func (ce *ComplianceEngine) GetAlerts(severity string, resolved bool, limit int) []ComplianceAlert {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	alerts := make([]ComplianceAlert, 0)

	for i := len(ce.alerts) - 1; i >= 0 && len(alerts) < limit; i-- {
		alert := ce.alerts[i]

		if severity != "" && alert.Severity != severity {
			continue
		}

		if alert.Resolved != resolved {
			continue
		}

		alerts = append(alerts, alert)
	}

	return alerts
}

// ResolveAlert marks an alert as resolved
func (ce *ComplianceEngine) ResolveAlert(alertID int64) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	for i := range ce.alerts {
		if ce.alerts[i].ID == alertID {
			now := time.Now()
			ce.alerts[i].Resolved = true
			ce.alerts[i].ResolvedAt = &now
			return nil
		}
	}

	return fmt.Errorf("alert %d not found", alertID)
}

// GenerateBestExecutionReport creates regulatory compliance report
func (ce *ComplianceEngine) GenerateBestExecutionReport(startTime, endTime time.Time) map[string]interface{} {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	report := make(map[string]interface{})

	var totalTrades int64
	var aBookTrades int64
	var bBookTrades int64
	var partialHedges int64
	var rejectedTrades int64

	var totalVolume float64
	var aBookVolume float64
	var bBookVolume float64

	var complianceViolations int64

	for _, log := range ce.auditLogs {
		if !log.Timestamp.After(startTime) || log.Timestamp.After(endTime) {
			continue
		}

		totalTrades++
		totalVolume += log.Volume

		switch log.Decision.Action {
		case ActionABook:
			aBookTrades++
			aBookVolume += log.Volume
		case ActionBBook:
			bBookTrades++
			bBookVolume += log.Volume
		case ActionPartialHedge:
			partialHedges++
			aBookVolume += log.Volume * (log.Decision.ABookPercent / 100)
			bBookVolume += log.Volume * (log.Decision.BBookPercent / 100)
		case ActionReject:
			rejectedTrades++
		}

		if len(log.ComplianceFlags) > 0 {
			complianceViolations++
		}
	}

	report["period_start"] = startTime
	report["period_end"] = endTime
	report["total_trades"] = totalTrades
	report["abook_trades"] = aBookTrades
	report["bbook_trades"] = bBookTrades
	report["partial_hedges"] = partialHedges
	report["rejected_trades"] = rejectedTrades
	report["total_volume"] = totalVolume
	report["abook_volume"] = aBookVolume
	report["bbook_volume"] = bBookVolume

	if totalTrades > 0 {
		report["abook_percentage"] = (float64(aBookTrades) / float64(totalTrades)) * 100
		report["bbook_percentage"] = (float64(bBookTrades) / float64(totalTrades)) * 100
	}

	report["compliance_violations"] = complianceViolations
	report["generated_at"] = time.Now()

	return report
}

// GenerateRoutingAuditReport creates detailed audit report
func (ce *ComplianceEngine) GenerateRoutingAuditReport(accountID int64) map[string]interface{} {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	report := make(map[string]interface{})

	decisions := make([]map[string]interface{}, 0)
	var totalCorrect int64
	var totalDecisions int64

	for _, log := range ce.auditLogs {
		if accountID > 0 && log.AccountID != accountID {
			continue
		}

		if log.ActualOutcome == nil {
			continue
		}

		totalDecisions++
		if log.ActualOutcome.WasOptimal {
			totalCorrect++
		}

		decisionRecord := map[string]interface{}{
			"timestamp":      log.Timestamp,
			"account_id":     log.AccountID,
			"symbol":         log.Symbol,
			"action":         log.Decision.Action,
			"pnl":            log.ActualOutcome.RealizedPnL,
			"was_optimal":    log.ActualOutcome.WasOptimal,
			"toxicity_score": log.Decision.ToxicityScore,
		}

		decisions = append(decisions, decisionRecord)
	}

	report["account_id"] = accountID
	report["total_decisions"] = totalDecisions
	report["optimal_decisions"] = totalCorrect

	if totalDecisions > 0 {
		report["accuracy"] = (float64(totalCorrect) / float64(totalDecisions)) * 100
	}

	report["decisions"] = decisions
	report["generated_at"] = time.Now()

	return report
}

// ExportAuditTrail exports audit trail to JSON for regulatory submission
func (ce *ComplianceEngine) ExportAuditTrail(startTime, endTime time.Time) ([]byte, error) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	logs := make([]*AuditLog, 0)

	for _, log := range ce.auditLogs {
		if !log.Timestamp.After(startTime) || log.Timestamp.After(endTime) {
			continue
		}
		logs = append(logs, log)
	}

	export := map[string]interface{}{
		"export_date":  time.Now(),
		"period_start": startTime,
		"period_end":   endTime,
		"total_logs":   len(logs),
		"logs":         logs,
	}

	return json.MarshalIndent(export, "", "  ")
}

// GetComplianceStats returns compliance statistics
func (ce *ComplianceEngine) GetComplianceStats() map[string]interface{} {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	stats := make(map[string]interface{})

	stats["total_audit_logs"] = len(ce.auditLogs)
	stats["total_alerts"] = len(ce.alerts)

	// Count alerts by severity
	criticalAlerts := 0
	warningAlerts := 0
	infoAlerts := 0
	unresolvedAlerts := 0

	for _, alert := range ce.alerts {
		switch alert.Severity {
		case "CRITICAL":
			criticalAlerts++
		case "WARNING":
			warningAlerts++
		case "INFO":
			infoAlerts++
		}
		if !alert.Resolved {
			unresolvedAlerts++
		}
	}

	stats["critical_alerts"] = criticalAlerts
	stats["warning_alerts"] = warningAlerts
	stats["info_alerts"] = infoAlerts
	stats["unresolved_alerts"] = unresolvedAlerts

	return stats
}
