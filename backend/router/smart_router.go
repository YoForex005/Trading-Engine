package router

import (
	"errors"
	"sync"
)

// RoutingRule defines how orders are routed
type RoutingRule struct {
	ID            string  `json:"id"`
	GroupPattern  string  `json:"groupPattern"`  // e.g. "VIP-*"
	SymbolPattern string  `json:"symbolPattern"` // e.g. "EURUSD", "*"
	MinVolume     float64 `json:"minVolume"`
	MaxVolume     float64 `json:"maxVolume"`
	Action        string  `json:"action"`   // A_BOOK, B_BOOK, REJECT
	TargetLP      string  `json:"targetLp"` // e.g. "LMAX_PROD"
	Priority      int     `json:"priority"` // Higher = checked first
}

// Decision represents the routing decision for an order
type Decision struct {
	Action   string `json:"action"`
	TargetLP string `json:"targetLp,omitempty"`
	Reason   string `json:"reason"`
}

// SmartRouter handles A-Book / B-Book routing decisions
type SmartRouter struct {
	rules []RoutingRule
	mu    sync.RWMutex
}

func NewSmartRouter() *SmartRouter {
	return &SmartRouter{
		rules: []RoutingRule{
			// Default rules
			{
				ID:            "rule_001",
				GroupPattern:  "VIP-*",
				SymbolPattern: "*",
				MinVolume:     0,
				MaxVolume:     1000,
				Action:        "A_BOOK",
				TargetLP:      "LMAX_PROD",
				Priority:      100,
			},
			{
				ID:            "rule_002",
				GroupPattern:  "*",
				SymbolPattern: "*",
				MinVolume:     10,
				MaxVolume:     1000,
				Action:        "A_BOOK",
				TargetLP:      "LMAX_PROD",
				Priority:      50,
			},
			{
				ID:            "rule_003",
				GroupPattern:  "*",
				SymbolPattern: "*",
				MinVolume:     0,
				MaxVolume:     10,
				Action:        "B_BOOK",
				TargetLP:      "",
				Priority:      10,
			},
		},
	}
}

// Route determines where an order should go
func (r *SmartRouter) Route(group string, symbol string, volume float64) (*Decision, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rule := range r.rules {
		if r.matchesPattern(group, rule.GroupPattern) &&
			r.matchesPattern(symbol, rule.SymbolPattern) &&
			volume >= rule.MinVolume &&
			volume <= rule.MaxVolume {
			return &Decision{
				Action:   rule.Action,
				TargetLP: rule.TargetLP,
				Reason:   "Matched rule: " + rule.ID,
			}, nil
		}
	}

	// Default fallback
	return &Decision{
		Action: "B_BOOK",
		Reason: "No matching rule, defaulting to B-Book",
	}, nil
}

// matchesPattern checks if a value matches a glob pattern
func (r *SmartRouter) matchesPattern(value, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(value) >= len(prefix) && value[:len(prefix)] == prefix
	}
	return value == pattern
}

// AddRule adds a new routing rule
func (r *SmartRouter) AddRule(rule RoutingRule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules = append(r.rules, rule)
	// Sort by priority (would implement proper sorting)
}

// GetRules returns all routing rules
func (r *SmartRouter) GetRules() []RoutingRule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.rules
}

// UpdateRule updates an existing rule
func (r *SmartRouter) UpdateRule(ruleID string, updated RoutingRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, rule := range r.rules {
		if rule.ID == ruleID {
			r.rules[i] = updated
			return nil
		}
	}
	return errors.New("rule not found")
}
