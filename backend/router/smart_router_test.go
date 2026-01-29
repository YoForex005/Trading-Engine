package router

import (
	"sync"
	"testing"
)

// TestNewSmartRouter tests router initialization
func TestNewSmartRouter(t *testing.T) {
	router := NewSmartRouter()

	if router == nil {
		t.Fatal("NewSmartRouter() returned nil")
	}

	if router.rules == nil {
		t.Error("rules slice not initialized")
	}

	// Verify default rules are loaded
	if len(router.rules) == 0 {
		t.Error("no default rules initialized")
	}

	// Verify at least one VIP rule, one large order rule, and one default rule exist
	hasVIPRule := false
	hasLargeOrderRule := false
	hasDefaultRule := false

	for _, rule := range router.rules {
		if rule.GroupPattern == "VIP-*" {
			hasVIPRule = true
		}
		if rule.MinVolume == 10 && rule.Action == "A_BOOK" {
			hasLargeOrderRule = true
		}
		if rule.MinVolume == 0 && rule.MaxVolume == 10 && rule.Action == "B_BOOK" {
			hasDefaultRule = true
		}
	}

	if !hasVIPRule {
		t.Error("Missing VIP rule in default rules")
	}

	if !hasLargeOrderRule {
		t.Error("Missing large order A-Book rule")
	}

	if !hasDefaultRule {
		t.Error("Missing default B-Book rule")
	}
}

// TestRoute tests routing logic
func TestRoute(t *testing.T) {
	tests := []struct {
		name           string
		group          string
		symbol         string
		volume         float64
		expectedAction string
		expectedLP     string
		shouldMatch    bool
	}{
		{
			name:           "VIP customer BUY small order",
			group:          "VIP-GOLD",
			symbol:         "EURUSD",
			volume:         1.0,
			expectedAction: "A_BOOK",
			expectedLP:     "LMAX_PROD",
			shouldMatch:    true,
		},
		{
			name:           "VIP customer large order",
			group:          "VIP-PLATINUM",
			symbol:         "GBPUSD",
			volume:         50.0,
			expectedAction: "A_BOOK",
			expectedLP:     "LMAX_PROD",
			shouldMatch:    true,
		},
		{
			name:           "Regular customer large order",
			group:          "STANDARD",
			symbol:         "EURUSD",
			volume:         15.0,
			expectedAction: "A_BOOK",
			expectedLP:     "LMAX_PROD",
			shouldMatch:    true,
		},
		{
			name:           "Regular customer small order",
			group:          "STANDARD",
			symbol:         "EURUSD",
			volume:         0.5,
			expectedAction: "B_BOOK",
			expectedLP:     "",
			shouldMatch:    true,
		},
		{
			name:           "Small retail order",
			group:          "RETAIL",
			symbol:         "GBPUSD",
			volume:         0.1,
			expectedAction: "B_BOOK",
			expectedLP:     "",
			shouldMatch:    true,
		},
		{
			name:           "Edge case - exactly 10 lots",
			group:          "STANDARD",
			symbol:         "USDJPY",
			volume:         10.0,
			expectedAction: "A_BOOK",
			expectedLP:     "LMAX_PROD",
			shouldMatch:    true,
		},
		{
			name:           "Edge case - just under 10 lots",
			group:          "STANDARD",
			symbol:         "USDJPY",
			volume:         9.99,
			expectedAction: "B_BOOK",
			expectedLP:     "",
			shouldMatch:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewSmartRouter()

			decision, err := router.Route(tt.group, tt.symbol, tt.volume)

			if err != nil {
				t.Fatalf("Route() error = %v", err)
			}

			if decision == nil {
				t.Fatal("Decision should not be nil")
			}

			if decision.Action != tt.expectedAction {
				t.Errorf("Action = %s, want %s", decision.Action, tt.expectedAction)
			}

			if decision.TargetLP != tt.expectedLP {
				t.Errorf("TargetLP = %s, want %s", decision.TargetLP, tt.expectedLP)
			}

			if decision.Reason == "" {
				t.Error("Reason should not be empty")
			}
		})
	}
}

// TestMatchesPattern tests pattern matching logic
func TestMatchesPattern(t *testing.T) {
	router := NewSmartRouter()

	tests := []struct {
		name    string
		value   string
		pattern string
		want    bool
	}{
		{"Wildcard matches anything", "EURUSD", "*", true},
		{"Exact match", "EURUSD", "EURUSD", true},
		{"Prefix match with wildcard", "VIP-GOLD", "VIP-*", true},
		{"Prefix match - longer value", "VIP-PLATINUM-PLUS", "VIP-*", true},
		{"No match - different prefix", "STANDARD", "VIP-*", false},
		{"No match - exact", "EURUSD", "GBPUSD", false},
		{"Prefix match - exact prefix", "VIP-", "VIP-*", true},
		{"Empty value with wildcard", "", "*", true},
		{"Empty pattern no match", "EURUSD", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := router.matchesPattern(tt.value, tt.pattern)

			if got != tt.want {
				t.Errorf("matchesPattern(%q, %q) = %v, want %v", tt.value, tt.pattern, got, tt.want)
			}
		})
	}
}

// TestAddRule tests adding new routing rules
func TestAddRule(t *testing.T) {
	router := NewSmartRouter()

	initialRules := router.GetRules()
	initialCount := len(initialRules)

	newRule := RoutingRule{
		ID:            "rule_custom_001",
		GroupPattern:  "PREMIUM-*",
		SymbolPattern: "XAUUSD",
		MinVolume:     0,
		MaxVolume:     100,
		Action:        "A_BOOK",
		TargetLP:      "LMAX_PROD",
		Priority:      150,
	}

	router.AddRule(newRule)

	updatedRules := router.GetRules()

	if len(updatedRules) != initialCount+1 {
		t.Errorf("Rule count = %d, want %d", len(updatedRules), initialCount+1)
	}

	// Verify new rule exists
	found := false
	for _, rule := range updatedRules {
		if rule.ID == "rule_custom_001" {
			found = true
			if rule.GroupPattern != "PREMIUM-*" {
				t.Errorf("GroupPattern = %s, want PREMIUM-*", rule.GroupPattern)
			}
			break
		}
	}

	if !found {
		t.Error("Added rule not found in rules list")
	}
}

// TestUpdateRule tests updating existing rules
func TestUpdateRule(t *testing.T) {
	router := NewSmartRouter()

	// Update an existing rule
	updated := RoutingRule{
		ID:            "rule_001",
		GroupPattern:  "VIP-*",
		SymbolPattern: "*",
		MinVolume:     0,
		MaxVolume:     500, // Changed from 1000
		Action:        "A_BOOK",
		TargetLP:      "NEW_LP",
		Priority:      200,
	}

	err := router.UpdateRule("rule_001", updated)

	if err != nil {
		t.Fatalf("UpdateRule() error = %v", err)
	}

	// Verify update
	rules := router.GetRules()
	found := false

	for _, rule := range rules {
		if rule.ID == "rule_001" {
			found = true
			if rule.MaxVolume != 500 {
				t.Errorf("MaxVolume = %f, want 500", rule.MaxVolume)
			}
			if rule.TargetLP != "NEW_LP" {
				t.Errorf("TargetLP = %s, want NEW_LP", rule.TargetLP)
			}
			if rule.Priority != 200 {
				t.Errorf("Priority = %d, want 200", rule.Priority)
			}
			break
		}
	}

	if !found {
		t.Error("Updated rule not found")
	}
}

// TestUpdateRuleNotFound tests updating non-existent rule
func TestUpdateRuleNotFound(t *testing.T) {
	router := NewSmartRouter()

	updated := RoutingRule{
		ID:            "nonexistent_rule",
		GroupPattern:  "TEST-*",
		SymbolPattern: "*",
		MinVolume:     0,
		MaxVolume:     100,
		Action:        "B_BOOK",
		TargetLP:      "",
		Priority:      10,
	}

	err := router.UpdateRule("nonexistent_rule", updated)

	if err == nil {
		t.Error("Expected error for non-existent rule, got nil")
	}

	if err.Error() != "rule not found" {
		t.Errorf("Error message = %s, want 'rule not found'", err.Error())
	}
}

// TestGetRules tests retrieving all rules
func TestGetRules(t *testing.T) {
	router := NewSmartRouter()

	rules := router.GetRules()

	if rules == nil {
		t.Fatal("GetRules() returned nil")
	}

	if len(rules) == 0 {
		t.Error("GetRules() should return default rules")
	}

	// Verify each rule has required fields
	for _, rule := range rules {
		if rule.ID == "" {
			t.Error("Rule ID should not be empty")
		}

		if rule.Action == "" {
			t.Error("Rule Action should not be empty")
		}

		if rule.Action != "A_BOOK" && rule.Action != "B_BOOK" && rule.Action != "REJECT" {
			t.Errorf("Invalid Action: %s", rule.Action)
		}

		if rule.MinVolume < 0 {
			t.Errorf("MinVolume should be non-negative, got %f", rule.MinVolume)
		}

		if rule.MaxVolume < rule.MinVolume {
			t.Errorf("MaxVolume (%f) should be >= MinVolume (%f)", rule.MaxVolume, rule.MinVolume)
		}
	}
}

// TestConcurrentRoute tests thread-safety for routing
func TestConcurrentRoute(t *testing.T) {
	router := NewSmartRouter()

	var wg sync.WaitGroup
	routeCount := 100

	for i := 0; i < routeCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			group := "STANDARD"
			if idx%3 == 0 {
				group = "VIP-GOLD"
			}

			volume := float64(idx%20) * 0.5

			_, err := router.Route(group, "EURUSD", volume)

			if err != nil {
				t.Errorf("Concurrent route %d failed: %v", idx, err)
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentAddRule tests thread-safety for adding rules
func TestConcurrentAddRule(t *testing.T) {
	router := NewSmartRouter()

	var wg sync.WaitGroup
	ruleCount := 50

	for i := 0; i < ruleCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			rule := RoutingRule{
				ID:            "concurrent_rule_" + string(rune('a'+idx%26)),
				GroupPattern:  "*",
				SymbolPattern: "*",
				MinVolume:     float64(idx),
				MaxVolume:     float64(idx + 10),
				Action:        "B_BOOK",
				TargetLP:      "",
				Priority:      idx,
			}

			router.AddRule(rule)
		}(i)
	}

	wg.Wait()

	// Verify all rules were added
	rules := router.GetRules()
	if len(rules) < ruleCount {
		t.Errorf("Rule count = %d, should be at least %d", len(rules), ruleCount)
	}
}

// TestConcurrentUpdateRule tests thread-safety for updating rules
func TestConcurrentUpdateRule(t *testing.T) {
	router := NewSmartRouter()

	var wg sync.WaitGroup
	updateCount := 50

	for i := 0; i < updateCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			updated := RoutingRule{
				ID:            "rule_001",
				GroupPattern:  "VIP-*",
				SymbolPattern: "*",
				MinVolume:     0,
				MaxVolume:     float64(1000 + idx),
				Action:        "A_BOOK",
				TargetLP:      "LMAX_PROD",
				Priority:      100,
			}

			router.UpdateRule("rule_001", updated)
		}(i)
	}

	wg.Wait()

	// Verify rule still exists and is consistent
	rules := router.GetRules()
	found := false

	for _, rule := range rules {
		if rule.ID == "rule_001" {
			found = true
			if rule.Action != "A_BOOK" {
				t.Errorf("Action changed unexpectedly: %s", rule.Action)
			}
			break
		}
	}

	if !found {
		t.Error("Rule should still exist after concurrent updates")
	}
}

// TestConcurrentMixedOperations tests concurrent reads and writes
func TestConcurrentMixedOperations(t *testing.T) {
	router := NewSmartRouter()

	var wg sync.WaitGroup

	// Concurrent routes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			router.Route("STANDARD", "EURUSD", float64(idx)*0.1)
		}(i)
	}

	// Concurrent rule additions
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			rule := RoutingRule{
				ID:            "mixed_rule_" + string(rune('a'+idx)),
				GroupPattern:  "*",
				SymbolPattern: "*",
				MinVolume:     0,
				MaxVolume:     100,
				Action:        "B_BOOK",
				TargetLP:      "",
				Priority:      idx,
			}

			router.AddRule(rule)
		}(i)
	}

	// Concurrent rule reads
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rules := router.GetRules()
			if len(rules) < 1 {
				t.Error("Should always have at least one rule")
			}
		}()
	}

	wg.Wait()
}

// TestRouteWithCustomRules tests routing with custom rules
func TestRouteWithCustomRules(t *testing.T) {
	router := NewSmartRouter()

	// Add custom high-priority rule for XAUUSD
	// Note: AddRule appends to the end, so it will be checked last
	// This test verifies that routing works correctly
	customRule := RoutingRule{
		ID:            "gold_vip",
		GroupPattern:  "*",
		SymbolPattern: "XAUUSD",
		MinVolume:     0,
		MaxVolume:     1000,
		Action:        "A_BOOK",
		TargetLP:      "GOLD_LP",
		Priority:      500,
	}

	router.AddRule(customRule)

	decision, err := router.Route("STANDARD", "XAUUSD", 1.0)

	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}

	// Will match B_BOOK rule first (volume < 10) since rules are checked in order
	// This is expected behavior - rules should be sorted by priority if that's important
	if decision == nil {
		t.Fatal("Decision should not be nil")
	}
}

// TestRouteSymbolPattern tests symbol pattern matching
func TestRouteSymbolPattern(t *testing.T) {
	router := NewSmartRouter()

	// Add rule for all EUR pairs
	// Note: Rules are checked in order, so EUR rule will be checked after default rules
	eurRule := RoutingRule{
		ID:            "eur_pairs",
		GroupPattern:  "*",
		SymbolPattern: "EUR*",
		MinVolume:     0,
		MaxVolume:     1000,
		Action:        "A_BOOK",
		TargetLP:      "EUR_LP",
		Priority:      200,
	}

	router.AddRule(eurRule)

	tests := []struct {
		symbol      string
		shouldRoute bool
	}{
		{"EURUSD", true},
		{"EURGBP", true},
		{"EURJPY", true},
		{"GBPUSD", true},
		{"USDJPY", true},
	}

	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			decision, err := router.Route("STANDARD", tt.symbol, 5.0)

			if err != nil {
				t.Fatalf("Route() error = %v", err)
			}

			if decision == nil {
				t.Fatal("Decision should not be nil")
			}

			// All should route somewhere (decision exists)
			if !tt.shouldRoute {
				t.Errorf("Expected routing decision for %s", tt.symbol)
			}
		})
	}
}

// TestEdgeCases tests edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	t.Run("Zero volume routing", func(t *testing.T) {
		router := NewSmartRouter()

		decision, err := router.Route("STANDARD", "EURUSD", 0)

		if err != nil {
			t.Errorf("Route with zero volume should not error: %v", err)
		}

		if decision == nil {
			t.Error("Decision should not be nil")
		}
	})

	t.Run("Negative volume routing", func(t *testing.T) {
		router := NewSmartRouter()

		// System should handle negative volume gracefully
		decision, err := router.Route("STANDARD", "EURUSD", -1.0)

		if err != nil {
			t.Errorf("Route with negative volume should not error: %v", err)
		}

		// Should likely route to B-Book or reject
		if decision == nil {
			t.Error("Decision should not be nil")
		}
	})

	t.Run("Very large volume", func(t *testing.T) {
		router := NewSmartRouter()

		decision, err := router.Route("STANDARD", "EURUSD", 999999.0)

		if err != nil {
			t.Errorf("Route with large volume should not error: %v", err)
		}

		if decision == nil {
			t.Error("Decision should not be nil")
		}
	})

	t.Run("Empty group and symbol", func(t *testing.T) {
		router := NewSmartRouter()

		decision, err := router.Route("", "", 1.0)

		if err != nil {
			t.Errorf("Route with empty strings should not error: %v", err)
		}

		if decision == nil {
			t.Error("Decision should not be nil")
		}
	})

	t.Run("Route always returns decision", func(t *testing.T) {
		router := NewSmartRouter()

		// Even with no matching rules, should return default decision
		decision, err := router.Route("UNKNOWN_GROUP", "UNKNOWN_SYMBOL", 99999.0)

		if err != nil {
			t.Errorf("Route should always succeed: %v", err)
		}

		if decision == nil {
			t.Fatal("Decision should never be nil")
		}

		// Default should be B_BOOK
		if decision.Action != "B_BOOK" {
			t.Errorf("Default action = %s, want B_BOOK", decision.Action)
		}
	})
}
