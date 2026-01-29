package fix

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// RuleType defines the type of access rule
type RuleType string

const (
	RuleTypeMinBalance      RuleType = "min_balance"
	RuleTypeMinVolume       RuleType = "min_volume"
	RuleTypeAccountAge      RuleType = "account_age"
	RuleTypeKYCLevel        RuleType = "kyc_level"
	RuleTypeGroupMembership RuleType = "group_membership"
	RuleTypeCustom          RuleType = "custom"
)

// AccessRule defines a rule for FIX API access
type AccessRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        RuleType               `json:"type"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"` // Higher priority rules evaluated first
	Operator    string                 `json:"operator"` // ">", ">=", "==", "in", "contains"
	Value       interface{}            `json:"value"`
	ErrorMsg    string                 `json:"error_msg"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UserContext contains user information for rule evaluation
type UserContext struct {
	UserID          string
	AccountBalance  float64
	TradingVolume   float64 // Last 30 days
	AccountAge      time.Duration
	KYCLevel        int    // 0=none, 1=basic, 2=intermediate, 3=full
	Groups          []string
	CustomFields    map[string]interface{}
	IsAdmin         bool
}

// RuleEvaluationResult contains the result of rule evaluation
type RuleEvaluationResult struct {
	Allowed       bool
	FailedRules   []string
	WarningRules  []string
	EvaluatedAt   time.Time
}

// RulesEngine evaluates access rules for FIX API provisioning
type RulesEngine struct {
	rules       map[string]*AccessRule
	userRules   map[string][]string // userID -> ruleIDs (user-specific overrides)
	mu          sync.RWMutex
	auditLogger AuditLogger
}

// NewRulesEngine creates a new rules engine
func NewRulesEngine(auditLogger AuditLogger) *RulesEngine {
	engine := &RulesEngine{
		rules:       make(map[string]*AccessRule),
		userRules:   make(map[string][]string),
		auditLogger: auditLogger,
	}

	// Add default rules
	engine.addDefaultRules()

	return engine
}

// addDefaultRules adds standard access rules
func (re *RulesEngine) addDefaultRules() {
	defaultRules := []*AccessRule{
		{
			ID:          "rule_min_balance",
			Name:        "Minimum Account Balance",
			Description: "User must have minimum account balance",
			Type:        RuleTypeMinBalance,
			Enabled:     true,
			Priority:    100,
			Operator:    ">=",
			Value:       1000.0, // $1000 minimum
			ErrorMsg:    "Account balance must be at least $1000 to access FIX API",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rule_min_volume",
			Name:        "Minimum Trading Volume",
			Description: "User must have minimum trading volume (30 days)",
			Type:        RuleTypeMinVolume,
			Enabled:     true,
			Priority:    90,
			Operator:    ">=",
			Value:       10000.0, // $10k trading volume
			ErrorMsg:    "Trading volume must be at least $10,000 in the last 30 days",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rule_account_age",
			Name:        "Minimum Account Age",
			Description: "Account must be at least 30 days old",
			Type:        RuleTypeAccountAge,
			Enabled:     true,
			Priority:    80,
			Operator:    ">=",
			Value:       30 * 24 * time.Hour, // 30 days
			ErrorMsg:    "Account must be at least 30 days old to access FIX API",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rule_kyc_verified",
			Name:        "KYC Verification Required",
			Description: "User must have completed KYC verification",
			Type:        RuleTypeKYCLevel,
			Enabled:     true,
			Priority:    100,
			Operator:    ">=",
			Value:       2, // Intermediate KYC level
			ErrorMsg:    "KYC verification level 2 or higher required for FIX API access",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, rule := range defaultRules {
		re.rules[rule.ID] = rule
	}
}

// AddRule adds or updates a rule
func (re *RulesEngine) AddRule(rule *AccessRule) error {
	if rule.ID == "" {
		return errors.New("rule ID cannot be empty")
	}

	re.mu.Lock()
	defer re.mu.Unlock()

	now := time.Now()
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = now
	}
	rule.UpdatedAt = now

	re.rules[rule.ID] = rule
	re.auditLogger.LogCredentialOperation("rule_added", "system", fmt.Sprintf("rule=%s", rule.ID), true)

	return nil
}

// RemoveRule removes a rule
func (re *RulesEngine) RemoveRule(ruleID string) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	if _, exists := re.rules[ruleID]; !exists {
		return errors.New("rule not found")
	}

	delete(re.rules, ruleID)
	re.auditLogger.LogCredentialOperation("rule_removed", "system", fmt.Sprintf("rule=%s", ruleID), true)

	return nil
}

// EnableRule enables a rule
func (re *RulesEngine) EnableRule(ruleID string) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	rule, exists := re.rules[ruleID]
	if !exists {
		return errors.New("rule not found")
	}

	rule.Enabled = true
	rule.UpdatedAt = time.Now()

	return nil
}

// DisableRule disables a rule
func (re *RulesEngine) DisableRule(ruleID string) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	rule, exists := re.rules[ruleID]
	if !exists {
		return errors.New("rule not found")
	}

	rule.Enabled = false
	rule.UpdatedAt = time.Now()

	return nil
}

// SetUserRules sets custom rules for a specific user (overrides global rules)
func (re *RulesEngine) SetUserRules(userID string, ruleIDs []string) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	// Validate that all rules exist
	for _, ruleID := range ruleIDs {
		if _, exists := re.rules[ruleID]; !exists {
			return fmt.Errorf("rule not found: %s", ruleID)
		}
	}

	re.userRules[userID] = ruleIDs
	re.auditLogger.LogCredentialOperation("user_rules_set", userID, fmt.Sprintf("rules=%v", ruleIDs), true)

	return nil
}

// EvaluateAccess evaluates if a user should be granted FIX API access
func (re *RulesEngine) EvaluateAccess(ctx *UserContext) *RuleEvaluationResult {
	re.mu.RLock()
	defer re.mu.RUnlock()

	result := &RuleEvaluationResult{
		Allowed:      true,
		FailedRules:  make([]string, 0),
		WarningRules: make([]string, 0),
		EvaluatedAt:  time.Now(),
	}

	// Admins bypass all rules
	if ctx.IsAdmin {
		re.auditLogger.LogCredentialOperation("access_granted", ctx.UserID, "admin bypass", true)
		return result
	}

	// Get rules to evaluate (user-specific or global)
	rulesToEvaluate := re.getRulesToEvaluate(ctx.UserID)

	// Evaluate rules in priority order
	for _, rule := range rulesToEvaluate {
		if !rule.Enabled {
			continue
		}

		passed, err := re.evaluateRule(rule, ctx)
		if err != nil {
			re.auditLogger.LogCredentialOperation("rule_evaluation_error", ctx.UserID,
				fmt.Sprintf("rule=%s error=%v", rule.ID, err), false)
			continue
		}

		if !passed {
			result.Allowed = false
			result.FailedRules = append(result.FailedRules, rule.ErrorMsg)
		}
	}

	if result.Allowed {
		re.auditLogger.LogCredentialOperation("access_granted", ctx.UserID, "all rules passed", true)
	} else {
		re.auditLogger.LogCredentialOperation("access_denied", ctx.UserID,
			fmt.Sprintf("failed_rules=%v", result.FailedRules), false)
	}

	return result
}

// getRulesToEvaluate returns the rules to evaluate for a user
func (re *RulesEngine) getRulesToEvaluate(userID string) []*AccessRule {
	var rules []*AccessRule

	// Check for user-specific rules
	if userRuleIDs, hasCustom := re.userRules[userID]; hasCustom {
		for _, ruleID := range userRuleIDs {
			if rule, exists := re.rules[ruleID]; exists {
				rules = append(rules, rule)
			}
		}
	} else {
		// Use global rules
		for _, rule := range re.rules {
			rules = append(rules, rule)
		}
	}

	// Sort by priority (higher first)
	for i := 0; i < len(rules)-1; i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[i].Priority < rules[j].Priority {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}

	return rules
}

// evaluateRule evaluates a single rule against user context
func (re *RulesEngine) evaluateRule(rule *AccessRule, ctx *UserContext) (bool, error) {
	switch rule.Type {
	case RuleTypeMinBalance:
		threshold, ok := rule.Value.(float64)
		if !ok {
			return false, errors.New("invalid threshold value")
		}
		return re.compareNumeric(ctx.AccountBalance, rule.Operator, threshold)

	case RuleTypeMinVolume:
		threshold, ok := rule.Value.(float64)
		if !ok {
			return false, errors.New("invalid threshold value")
		}
		return re.compareNumeric(ctx.TradingVolume, rule.Operator, threshold)

	case RuleTypeAccountAge:
		threshold, ok := rule.Value.(time.Duration)
		if !ok {
			return false, errors.New("invalid threshold value")
		}
		return ctx.AccountAge >= threshold, nil

	case RuleTypeKYCLevel:
		threshold, ok := rule.Value.(int)
		if !ok {
			// Handle float64 from JSON unmarshaling
			if f, ok := rule.Value.(float64); ok {
				threshold = int(f)
			} else {
				return false, errors.New("invalid threshold value")
			}
		}
		return ctx.KYCLevel >= threshold, nil

	case RuleTypeGroupMembership:
		requiredGroups, ok := rule.Value.([]string)
		if !ok {
			return false, errors.New("invalid group list")
		}
		return re.hasAnyGroup(ctx.Groups, requiredGroups), nil

	case RuleTypeCustom:
		// Custom rules use metadata for evaluation
		return re.evaluateCustomRule(rule, ctx)

	default:
		return false, fmt.Errorf("unknown rule type: %s", rule.Type)
	}
}

// compareNumeric compares two numeric values based on operator
func (re *RulesEngine) compareNumeric(actual float64, operator string, threshold float64) (bool, error) {
	switch operator {
	case ">":
		return actual > threshold, nil
	case ">=":
		return actual >= threshold, nil
	case "==":
		return actual == threshold, nil
	case "<":
		return actual < threshold, nil
	case "<=":
		return actual <= threshold, nil
	default:
		return false, fmt.Errorf("unknown operator: %s", operator)
	}
}

// hasAnyGroup checks if user is in any of the required groups
func (re *RulesEngine) hasAnyGroup(userGroups, requiredGroups []string) bool {
	for _, required := range requiredGroups {
		for _, userGroup := range userGroups {
			if userGroup == required {
				return true
			}
		}
	}
	return false
}

// evaluateCustomRule evaluates a custom rule using metadata
func (re *RulesEngine) evaluateCustomRule(rule *AccessRule, ctx *UserContext) (bool, error) {
	// Custom rules can access both rule.Metadata and ctx.CustomFields
	// This is where you'd implement custom business logic

	// Example: Check a custom field exists
	if fieldName, ok := rule.Metadata["required_field"].(string); ok {
		if _, exists := ctx.CustomFields[fieldName]; !exists {
			return false, nil
		}
	}

	return true, nil
}

// GetRule retrieves a rule by ID
func (re *RulesEngine) GetRule(ruleID string) (*AccessRule, error) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	rule, exists := re.rules[ruleID]
	if !exists {
		return nil, errors.New("rule not found")
	}

	return rule, nil
}

// ListRules returns all rules
func (re *RulesEngine) ListRules() []*AccessRule {
	re.mu.RLock()
	defer re.mu.RUnlock()

	rules := make([]*AccessRule, 0, len(re.rules))
	for _, rule := range re.rules {
		rules = append(rules, rule)
	}

	return rules
}

// GetUserRules returns custom rules for a user
func (re *RulesEngine) GetUserRules(userID string) []string {
	re.mu.RLock()
	defer re.mu.RUnlock()

	return re.userRules[userID]
}
