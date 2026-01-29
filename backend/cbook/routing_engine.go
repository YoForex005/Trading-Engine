package cbook

import (
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

// RoutingAction defines where to route an order
type RoutingAction string

const (
	ActionABook         RoutingAction = "A_BOOK"         // Route to LP
	ActionBBook         RoutingAction = "B_BOOK"         // Internalize
	ActionPartialHedge  RoutingAction = "PARTIAL_HEDGE"  // Split between A and B
	ActionReject        RoutingAction = "REJECT"         // Reject order
)

// RoutingDecision represents the routing decision for an order
type RoutingDecision struct {
	Action         RoutingAction `json:"action"`
	TargetLP       string        `json:"targetLp,omitempty"`
	ABookPercent   float64       `json:"aBookPercent"`   // 0-100
	BBookPercent   float64       `json:"bBookPercent"`   // 0-100
	Reason         string        `json:"reason"`
	ToxicityScore  float64       `json:"toxicityScore"`
	ExposureRisk   float64       `json:"exposureRisk"`
	DecisionTime   time.Time     `json:"decisionTime"`
}

// ExposureLimit defines risk limits per instrument
type ExposureLimit struct {
	Symbol           string  `json:"symbol"`
	MaxNetExposure   float64 `json:"maxNetExposure"`   // Max net lots
	MaxGrossExposure float64 `json:"maxGrossExposure"` // Max total lots
	AutoHedgeLevel   float64 `json:"autoHedgeLevel"`   // Trigger A-Book when exceeded
}

// RoutingRule defines manual routing rules (legacy compatibility)
type RoutingRule struct {
	ID            string        `json:"id"`
	Priority      int           `json:"priority"` // Higher = evaluated first

	// Filters
	AccountIDs    []int64       `json:"accountIds,omitempty"`
	UserGroups    []string      `json:"userGroups,omitempty"`
	Symbols       []string      `json:"symbols,omitempty"`
	MinVolume     float64       `json:"minVolume"`
	MaxVolume     float64       `json:"maxVolume"`

	// Classification filters
	Classifications []ClientClassification `json:"classifications,omitempty"`
	MinToxicity     float64                `json:"minToxicity"`
	MaxToxicity     float64                `json:"maxToxicity"`

	// Action
	Action       RoutingAction `json:"action"`
	TargetLP     string        `json:"targetLp,omitempty"`
	HedgePercent float64       `json:"hedgePercent"` // For partial hedge

	// Metadata
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

// RoutingEngine handles intelligent A-Book/B-Book routing
type RoutingEngine struct {
	mu sync.RWMutex

	// Components
	profileEngine *ClientProfileEngine

	// Rules and limits
	rules          []*RoutingRule
	exposureLimits map[string]*ExposureLimit

	// Exposure tracking
	symbolExposure map[string]*SymbolExposure

	// Configuration
	defaultLP            string
	defaultHedgePercent  float64 // Default partial hedge ratio
	maxBBookExposure     float64 // Global B-Book limit
	volatilityThreshold  float64 // Route to A-Book when volatility > threshold

	// Analytics
	decisions      []RoutingDecision
	maxDecisions   int // Keep last N decisions
}

// SymbolExposure tracks current exposure for a symbol
type SymbolExposure struct {
	Symbol         string    `json:"symbol"`
	NetExposure    float64   `json:"netExposure"`    // Net lots (long - short)
	GrossExposure  float64   `json:"grossExposure"`  // Total lots (long + short)
	LongExposure   float64   `json:"longExposure"`
	ShortExposure  float64   `json:"shortExposure"`
	LastUpdated    time.Time `json:"lastUpdated"`
}

// NewRoutingEngine creates a new routing engine
func NewRoutingEngine(profileEngine *ClientProfileEngine) *RoutingEngine {
	return &RoutingEngine{
		profileEngine:       profileEngine,
		rules:               make([]*RoutingRule, 0),
		exposureLimits:      make(map[string]*ExposureLimit),
		symbolExposure:      make(map[string]*SymbolExposure),
		decisions:           make([]RoutingDecision, 0, 10000),
		maxDecisions:        10000,
		defaultLP:           "LMAX_PROD",
		defaultHedgePercent: 70, // Default: 70% A-Book, 30% B-Book
		maxBBookExposure:    1000, // 1000 lots
		volatilityThreshold: 0.02, // 2% volatility
	}
}

// Route makes a routing decision for an order
func (re *RoutingEngine) Route(accountID int64, symbol string, side string, volume float64, currentVolatility float64) (*RoutingDecision, error) {
	re.mu.Lock()
	defer re.mu.Unlock()

	decision := &RoutingDecision{
		DecisionTime: time.Now(),
	}

	// Get client profile
	profile, exists := re.profileEngine.GetProfile(accountID)
	if !exists {
		// New client - default to B-Book with caution
		decision.Action = ActionBBook
		decision.BBookPercent = 50
		decision.ABookPercent = 50
		decision.Reason = "New client - conservative routing"
		re.recordDecision(decision)
		return decision, nil
	}

	decision.ToxicityScore = profile.ToxicityScore

	// 1. Check manual rules first (highest priority)
	if ruleDecision := re.checkRules(accountID, symbol, volume, profile); ruleDecision != nil {
		re.recordDecision(ruleDecision)
		return ruleDecision, nil
	}

	// 2. Classification-based routing
	switch profile.Classification {
	case ClassificationToxic:
		// Toxic clients - A-Book or reject
		if profile.ToxicityScore > 80 {
			decision.Action = ActionReject
			decision.Reason = fmt.Sprintf("Toxic client (score: %.1f) - rejected", profile.ToxicityScore)
		} else {
			decision.Action = ActionABook
			decision.ABookPercent = 100
			decision.TargetLP = re.defaultLP
			decision.Reason = fmt.Sprintf("Toxic client (score: %.1f) - full A-Book", profile.ToxicityScore)
		}
		re.recordDecision(decision)
		return decision, nil

	case ClassificationProfessional:
		// Professional - mostly A-Book
		decision.Action = ActionPartialHedge
		decision.ABookPercent = 80
		decision.BBookPercent = 20
		decision.TargetLP = re.defaultLP
		decision.Reason = "Professional trader - 80% A-Book hedge"

	case ClassificationSemiPro:
		// Semi-pro - balanced
		decision.Action = ActionPartialHedge
		decision.ABookPercent = 50
		decision.BBookPercent = 50
		decision.TargetLP = re.defaultLP
		decision.Reason = "Semi-professional - 50/50 split"

	case ClassificationRetail:
		// Retail - mostly B-Book (broker profit opportunity)
		decision.Action = ActionBBook
		decision.BBookPercent = 90
		decision.ABookPercent = 10
		decision.Reason = "Retail trader - 90% B-Book"

	default:
		// Unknown - conservative
		decision.Action = ActionBBook
		decision.BBookPercent = 60
		decision.ABookPercent = 40
		decision.Reason = "Unknown classification - conservative B-Book"
	}

	// 3. Volume-based override
	if volume >= 10 {
		// Large orders always A-Book
		decision.Action = ActionABook
		decision.ABookPercent = 100
		decision.BBookPercent = 0
		decision.TargetLP = re.defaultLP
		decision.Reason = fmt.Sprintf("Large volume (%.2f lots) - full A-Book", volume)
		re.recordDecision(decision)
		return decision, nil
	}

	// 4. Exposure-based adjustment
	exposure := re.getOrCreateExposure(symbol)
	limit := re.getExposureLimit(symbol)

	// Calculate what new exposure would be
	var projectedNet float64
	if side == "BUY" {
		projectedNet = exposure.NetExposure + volume
	} else {
		projectedNet = exposure.NetExposure - volume
	}

	decision.ExposureRisk = (projectedNet / limit.MaxNetExposure) * 100

	// If exposure too high, route to A-Book
	if abs(projectedNet) > limit.AutoHedgeLevel {
		decision.Action = ActionABook
		decision.ABookPercent = 100
		decision.BBookPercent = 0
		decision.TargetLP = re.defaultLP
		decision.Reason = fmt.Sprintf("Exposure limit reached (%.2f/%.2f lots) - A-Book hedge",
			abs(projectedNet), limit.MaxNetExposure)
		re.recordDecision(decision)
		return decision, nil
	}

	// If approaching limit, increase A-Book percentage
	if abs(projectedNet) > limit.AutoHedgeLevel * 0.7 {
		exposureFactor := abs(projectedNet) / limit.AutoHedgeLevel
		decision.ABookPercent = math.Min(decision.ABookPercent + (exposureFactor * 30), 100)
		decision.BBookPercent = 100 - decision.ABookPercent
		decision.Reason += fmt.Sprintf(" + exposure adjustment (%.1f%% A-Book)", decision.ABookPercent)
	}

	// 5. Volatility-based adjustment
	if currentVolatility > re.volatilityThreshold {
		// High volatility - increase A-Book
		decision.ABookPercent = math.Min(decision.ABookPercent + 30, 100)
		decision.BBookPercent = 100 - decision.ABookPercent
		decision.Reason += fmt.Sprintf(" + high volatility (%.2f%%)", currentVolatility*100)
	}

	// Final validation
	if decision.ABookPercent > 0 {
		decision.Action = ActionPartialHedge
		decision.TargetLP = re.defaultLP
	}
	if decision.ABookPercent >= 100 {
		decision.Action = ActionABook
	}
	if decision.BBookPercent >= 100 {
		decision.Action = ActionBBook
	}

	re.recordDecision(decision)
	return decision, nil
}

// checkRules evaluates manual routing rules
func (re *RoutingEngine) checkRules(accountID int64, symbol string, volume float64, profile *ClientProfile) *RoutingDecision {
	// Sort rules by priority
	sortedRules := make([]*RoutingRule, len(re.rules))
	copy(sortedRules, re.rules)

	// Simple bubble sort by priority (descending)
	for i := 0; i < len(sortedRules); i++ {
		for j := i + 1; j < len(sortedRules); j++ {
			if sortedRules[j].Priority > sortedRules[i].Priority {
				sortedRules[i], sortedRules[j] = sortedRules[j], sortedRules[i]
			}
		}
	}

	for _, rule := range sortedRules {
		if !rule.Enabled {
			continue
		}

		// Check filters
		if !re.ruleMatches(rule, accountID, symbol, volume, profile) {
			continue
		}

		// Rule matched - create decision
		decision := &RoutingDecision{
			Action:       rule.Action,
			TargetLP:     rule.TargetLP,
			DecisionTime: time.Now(),
			Reason:       fmt.Sprintf("Matched rule: %s (%s)", rule.ID, rule.Description),
		}

		if rule.Action == ActionPartialHedge {
			decision.ABookPercent = rule.HedgePercent
			decision.BBookPercent = 100 - rule.HedgePercent
		} else if rule.Action == ActionABook {
			decision.ABookPercent = 100
		} else if rule.Action == ActionBBook {
			decision.BBookPercent = 100
		}

		if profile != nil {
			decision.ToxicityScore = profile.ToxicityScore
		}

		return decision
	}

	return nil // No rule matched
}

// ruleMatches checks if a rule applies to the order
func (re *RoutingEngine) ruleMatches(rule *RoutingRule, accountID int64, symbol string, volume float64, profile *ClientProfile) bool {
	// Check account filter
	if len(rule.AccountIDs) > 0 {
		found := false
		for _, id := range rule.AccountIDs {
			if id == accountID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check symbol filter
	if len(rule.Symbols) > 0 {
		found := false
		for _, sym := range rule.Symbols {
			if sym == symbol || sym == "*" {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check volume
	if volume < rule.MinVolume || (rule.MaxVolume > 0 && volume > rule.MaxVolume) {
		return false
	}

	// Check classification
	if profile != nil && len(rule.Classifications) > 0 {
		found := false
		for _, classification := range rule.Classifications {
			if classification == profile.Classification {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check toxicity
	if profile != nil {
		if rule.MinToxicity > 0 && profile.ToxicityScore < rule.MinToxicity {
			return false
		}
		if rule.MaxToxicity > 0 && profile.ToxicityScore > rule.MaxToxicity {
			return false
		}
	}

	return true
}

// UpdateExposure updates symbol exposure after trade execution
func (re *RoutingEngine) UpdateExposure(symbol, side string, volume float64) {
	re.mu.Lock()
	defer re.mu.Unlock()

	exposure := re.getOrCreateExposure(symbol)

	if side == "BUY" {
		exposure.LongExposure += volume
		exposure.NetExposure += volume
	} else {
		exposure.ShortExposure += volume
		exposure.NetExposure -= volume
	}

	exposure.GrossExposure = exposure.LongExposure + exposure.ShortExposure
	exposure.LastUpdated = time.Now()

	log.Printf("[RoutingEngine] Updated %s exposure: Net=%.2f, Gross=%.2f",
		symbol, exposure.NetExposure, exposure.GrossExposure)
}

// GetExposure returns current exposure for a symbol
func (re *RoutingEngine) GetExposure(symbol string) *SymbolExposure {
	re.mu.RLock()
	defer re.mu.RUnlock()

	return re.getOrCreateExposure(symbol)
}

// getOrCreateExposure gets or creates exposure tracking (must be called with lock)
func (re *RoutingEngine) getOrCreateExposure(symbol string) *SymbolExposure {
	exposure, exists := re.symbolExposure[symbol]
	if !exists {
		exposure = &SymbolExposure{
			Symbol:      symbol,
			LastUpdated: time.Now(),
		}
		re.symbolExposure[symbol] = exposure
	}
	return exposure
}

// getExposureLimit gets exposure limit for symbol (must be called with lock)
func (re *RoutingEngine) getExposureLimit(symbol string) *ExposureLimit {
	limit, exists := re.exposureLimits[symbol]
	if !exists {
		// Default limits
		limit = &ExposureLimit{
			Symbol:           symbol,
			MaxNetExposure:   500,  // 500 lots
			MaxGrossExposure: 1000, // 1000 lots
			AutoHedgeLevel:   300,  // Hedge when > 300 lots
		}
		re.exposureLimits[symbol] = limit
	}
	return limit
}

// AddRule adds a routing rule
func (re *RoutingEngine) AddRule(rule *RoutingRule) {
	re.mu.Lock()
	defer re.mu.Unlock()

	re.rules = append(re.rules, rule)
	log.Printf("[RoutingEngine] Added rule: %s (priority: %d)", rule.ID, rule.Priority)
}

// UpdateRule updates an existing rule
func (re *RoutingEngine) UpdateRule(ruleID string, updated *RoutingRule) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	for i, rule := range re.rules {
		if rule.ID == ruleID {
			re.rules[i] = updated
			log.Printf("[RoutingEngine] Updated rule: %s", ruleID)
			return nil
		}
	}

	return errors.New("rule not found")
}

// DeleteRule removes a rule
func (re *RoutingEngine) DeleteRule(ruleID string) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	for i, rule := range re.rules {
		if rule.ID == ruleID {
			re.rules = append(re.rules[:i], re.rules[i+1:]...)
			log.Printf("[RoutingEngine] Deleted rule: %s", ruleID)
			return nil
		}
	}

	return errors.New("rule not found")
}

// GetRules returns all routing rules
func (re *RoutingEngine) GetRules() []*RoutingRule {
	re.mu.RLock()
	defer re.mu.RUnlock()

	rules := make([]*RoutingRule, len(re.rules))
	copy(rules, re.rules)
	return rules
}

// SetExposureLimit sets exposure limits for a symbol
func (re *RoutingEngine) SetExposureLimit(symbol string, limit *ExposureLimit) {
	re.mu.Lock()
	defer re.mu.Unlock()

	limit.Symbol = symbol
	re.exposureLimits[symbol] = limit
	log.Printf("[RoutingEngine] Set exposure limit for %s: MaxNet=%.2f, AutoHedge=%.2f",
		symbol, limit.MaxNetExposure, limit.AutoHedgeLevel)
}

// recordDecision stores decision for analytics
func (re *RoutingEngine) recordDecision(decision *RoutingDecision) {
	re.decisions = append(re.decisions, *decision)
	if len(re.decisions) > re.maxDecisions {
		re.decisions = re.decisions[1:]
	}
}

// GetDecisionHistory returns recent routing decisions
func (re *RoutingEngine) GetDecisionHistory(limit int) []RoutingDecision {
	re.mu.RLock()
	defer re.mu.RUnlock()

	if limit <= 0 || limit > len(re.decisions) {
		limit = len(re.decisions)
	}

	start := len(re.decisions) - limit
	decisions := make([]RoutingDecision, limit)
	copy(decisions, re.decisions[start:])
	return decisions
}

// GetRoutingStats returns routing statistics
func (re *RoutingEngine) GetRoutingStats() map[string]interface{} {
	re.mu.RLock()
	defer re.mu.RUnlock()

	stats := make(map[string]interface{})

	// Count decisions by action
	actionCounts := make(map[RoutingAction]int)
	for _, decision := range re.decisions {
		actionCounts[decision.Action]++
	}

	stats["total_decisions"] = len(re.decisions)
	stats["action_counts"] = actionCounts
	stats["active_rules"] = len(re.rules)
	stats["tracked_symbols"] = len(re.symbolExposure)

	// Exposure summary
	exposures := make(map[string]map[string]float64)
	for symbol, exp := range re.symbolExposure {
		exposures[symbol] = map[string]float64{
			"net":   exp.NetExposure,
			"gross": exp.GrossExposure,
			"long":  exp.LongExposure,
			"short": exp.ShortExposure,
		}
	}
	stats["exposures"] = exposures

	return stats
}

// Helper function
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
