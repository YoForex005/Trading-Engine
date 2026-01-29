package featureflags

import (
	"regexp"
	"strings"
	"time"
)

// EvaluationContext contains the context for evaluating targeting rules
type EvaluationContext struct {
	UserID     string                 `json:"user_id"`
	Attributes map[string]interface{} `json:"attributes"`
	Timestamp  time.Time              `json:"timestamp"`

	// Common attributes
	Country    string `json:"country,omitempty"`
	Language   string `json:"language,omitempty"`
	Tier       string `json:"tier,omitempty"` // free, premium, vip
	DeviceType string `json:"device_type,omitempty"` // mobile, desktop, tablet
	Browser    string `json:"browser,omitempty"`
	OS         string `json:"os,omitempty"`
	AppVersion string `json:"app_version,omitempty"`

	// Trading specific
	AccountType    string  `json:"account_type,omitempty"` // demo, live
	AccountBalance float64 `json:"account_balance,omitempty"`
	TradingVolume  float64 `json:"trading_volume,omitempty"`
	IsNewUser      bool    `json:"is_new_user,omitempty"`
	IsBetaTester   bool    `json:"is_beta_tester,omitempty"`
	IsVIP          bool    `json:"is_vip,omitempty"`
}

// TargetingRule represents a rule for targeting users
type TargetingRule struct {
	ID          string                 `json:"id"`
	Type        RuleType               `json:"type"`
	Attribute   string                 `json:"attribute,omitempty"`
	Operator    RuleOperator           `json:"operator,omitempty"`
	Value       interface{}            `json:"value,omitempty"`
	Values      []interface{}          `json:"values,omitempty"`
	Conditions  []TargetingRule        `json:"conditions,omitempty"` // For AND/OR rules
}

// RuleType represents the type of targeting rule
type RuleType string

const (
	RuleTypeUserID        RuleType = "user_id"
	RuleTypeUserIDHash    RuleType = "user_id_hash"
	RuleTypeCountry       RuleType = "country"
	RuleTypeLanguage      RuleType = "language"
	RuleTypeTier          RuleType = "tier"
	RuleTypeDeviceType    RuleType = "device_type"
	RuleTypeBrowser       RuleType = "browser"
	RuleTypeOS            RuleType = "os"
	RuleTypeAppVersion    RuleType = "app_version"
	RuleTypeCustom        RuleType = "custom"
	RuleTypeBetaTester    RuleType = "beta_tester"
	RuleTypeVIP           RuleType = "vip"
	RuleTypeNewUser       RuleType = "new_user"
	RuleTypePercentage    RuleType = "percentage"
	RuleTypeAccountType   RuleType = "account_type"
	RuleTypeAccountBalance RuleType = "account_balance"
	RuleTypeTradingVolume RuleType = "trading_volume"
	RuleTypeDate          RuleType = "date"
	RuleTypeAnd           RuleType = "and"
	RuleTypeOr            RuleType = "or"
	RuleTypeNot           RuleType = "not"
)

// RuleOperator represents the comparison operator
type RuleOperator string

const (
	OpEquals            RuleOperator = "equals"
	OpNotEquals         RuleOperator = "not_equals"
	OpIn                RuleOperator = "in"
	OpNotIn             RuleOperator = "not_in"
	OpContains          RuleOperator = "contains"
	OpNotContains       RuleOperator = "not_contains"
	OpStartsWith        RuleOperator = "starts_with"
	OpEndsWith          RuleOperator = "ends_with"
	OpMatches           RuleOperator = "matches" // regex
	OpGreaterThan       RuleOperator = "greater_than"
	OpGreaterThanOrEqual RuleOperator = "greater_than_or_equal"
	OpLessThan          RuleOperator = "less_than"
	OpLessThanOrEqual   RuleOperator = "less_than_or_equal"
	OpBefore            RuleOperator = "before" // date
	OpAfter             RuleOperator = "after"  // date
	OpBetween           RuleOperator = "between"
	OpExists            RuleOperator = "exists"
	OpNotExists         RuleOperator = "not_exists"
)

// Evaluate evaluates the targeting rule against the context
func (r *TargetingRule) Evaluate(ctx *EvaluationContext) bool {
	switch r.Type {
	case RuleTypeAnd:
		return r.evaluateAnd(ctx)
	case RuleTypeOr:
		return r.evaluateOr(ctx)
	case RuleTypeNot:
		return r.evaluateNot(ctx)
	case RuleTypeUserID:
		return r.evaluateUserID(ctx)
	case RuleTypeUserIDHash:
		return r.evaluateUserIDHash(ctx)
	case RuleTypeCountry:
		return r.evaluateStringMatch(ctx.Country)
	case RuleTypeLanguage:
		return r.evaluateStringMatch(ctx.Language)
	case RuleTypeTier:
		return r.evaluateStringMatch(ctx.Tier)
	case RuleTypeDeviceType:
		return r.evaluateStringMatch(ctx.DeviceType)
	case RuleTypeBrowser:
		return r.evaluateStringMatch(ctx.Browser)
	case RuleTypeOS:
		return r.evaluateStringMatch(ctx.OS)
	case RuleTypeAppVersion:
		return r.evaluateVersion(ctx.AppVersion)
	case RuleTypeBetaTester:
		return ctx.IsBetaTester
	case RuleTypeVIP:
		return ctx.IsVIP
	case RuleTypeNewUser:
		return ctx.IsNewUser
	case RuleTypeAccountType:
		return r.evaluateStringMatch(ctx.AccountType)
	case RuleTypeAccountBalance:
		return r.evaluateNumeric(ctx.AccountBalance)
	case RuleTypeTradingVolume:
		return r.evaluateNumeric(ctx.TradingVolume)
	case RuleTypeDate:
		return r.evaluateDate(ctx.Timestamp)
	case RuleTypePercentage:
		return r.evaluatePercentage(ctx.UserID)
	case RuleTypeCustom:
		return r.evaluateCustomAttribute(ctx)
	default:
		return false
	}
}

// Helper evaluation methods

func (r *TargetingRule) evaluateAnd(ctx *EvaluationContext) bool {
	for _, condition := range r.Conditions {
		if !condition.Evaluate(ctx) {
			return false
		}
	}
	return len(r.Conditions) > 0
}

func (r *TargetingRule) evaluateOr(ctx *EvaluationContext) bool {
	for _, condition := range r.Conditions {
		if condition.Evaluate(ctx) {
			return true
		}
	}
	return false
}

func (r *TargetingRule) evaluateNot(ctx *EvaluationContext) bool {
	if len(r.Conditions) == 0 {
		return false
	}
	return !r.Conditions[0].Evaluate(ctx)
}

func (r *TargetingRule) evaluateUserID(ctx *EvaluationContext) bool {
	switch r.Operator {
	case OpEquals:
		return ctx.UserID == r.Value.(string)
	case OpNotEquals:
		return ctx.UserID != r.Value.(string)
	case OpIn:
		for _, v := range r.Values {
			if ctx.UserID == v.(string) {
				return true
			}
		}
		return false
	case OpNotIn:
		for _, v := range r.Values {
			if ctx.UserID == v.(string) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (r *TargetingRule) evaluateUserIDHash(ctx *EvaluationContext) bool {
	// Percentage-based selection using consistent hashing
	// Type conversion for validation only - percentage value not directly used
	_, ok := r.Value.(int)
	if !ok {
		if _, ok := r.Value.(float64); !ok {
			return false
		}
	}
	return r.evaluatePercentage(ctx.UserID)
}

func (r *TargetingRule) evaluateStringMatch(value string) bool {
	switch r.Operator {
	case OpEquals:
		return value == r.Value.(string)
	case OpNotEquals:
		return value != r.Value.(string)
	case OpIn:
		for _, v := range r.Values {
			if value == v.(string) {
				return true
			}
		}
		return false
	case OpNotIn:
		for _, v := range r.Values {
			if value == v.(string) {
				return false
			}
		}
		return true
	case OpContains:
		return strings.Contains(strings.ToLower(value), strings.ToLower(r.Value.(string)))
	case OpNotContains:
		return !strings.Contains(strings.ToLower(value), strings.ToLower(r.Value.(string)))
	case OpStartsWith:
		return strings.HasPrefix(strings.ToLower(value), strings.ToLower(r.Value.(string)))
	case OpEndsWith:
		return strings.HasSuffix(strings.ToLower(value), strings.ToLower(r.Value.(string)))
	case OpMatches:
		matched, _ := regexp.MatchString(r.Value.(string), value)
		return matched
	case OpExists:
		return value != ""
	case OpNotExists:
		return value == ""
	default:
		return false
	}
}

func (r *TargetingRule) evaluateNumeric(value float64) bool {
	var target float64
	switch v := r.Value.(type) {
	case float64:
		target = v
	case int:
		target = float64(v)
	case int64:
		target = float64(v)
	default:
		return false
	}

	switch r.Operator {
	case OpEquals:
		return value == target
	case OpNotEquals:
		return value != target
	case OpGreaterThan:
		return value > target
	case OpGreaterThanOrEqual:
		return value >= target
	case OpLessThan:
		return value < target
	case OpLessThanOrEqual:
		return value <= target
	case OpBetween:
		if len(r.Values) != 2 {
			return false
		}
		min := r.Values[0].(float64)
		max := r.Values[1].(float64)
		return value >= min && value <= max
	default:
		return false
	}
}

func (r *TargetingRule) evaluateVersion(version string) bool {
	target, ok := r.Value.(string)
	if !ok {
		return false
	}

	switch r.Operator {
	case OpEquals:
		return version == target
	case OpNotEquals:
		return version != target
	case OpGreaterThan:
		return compareVersions(version, target) > 0
	case OpGreaterThanOrEqual:
		return compareVersions(version, target) >= 0
	case OpLessThan:
		return compareVersions(version, target) < 0
	case OpLessThanOrEqual:
		return compareVersions(version, target) <= 0
	default:
		return r.evaluateStringMatch(version)
	}
}

func (r *TargetingRule) evaluateDate(timestamp time.Time) bool {
	var target time.Time
	switch v := r.Value.(type) {
	case time.Time:
		target = v
	case string:
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return false
		}
		target = parsed
	default:
		return false
	}

	switch r.Operator {
	case OpBefore:
		return timestamp.Before(target)
	case OpAfter:
		return timestamp.After(target)
	case OpEquals:
		return timestamp.Equal(target)
	default:
		return false
	}
}

func (r *TargetingRule) evaluatePercentage(userID string) bool {
	percentage, ok := r.Value.(int)
	if !ok {
		if f, ok := r.Value.(float64); ok {
			percentage = int(f)
		} else {
			return false
		}
	}

	if percentage <= 0 {
		return false
	}
	if percentage >= 100 {
		return true
	}

	// Use consistent hashing (same as FlagManager)
	hash := hashString(userID)
	value := hash % 100
	return int(value) < percentage
}

func (r *TargetingRule) evaluateCustomAttribute(ctx *EvaluationContext) bool {
	if r.Attribute == "" {
		return false
	}

	value, exists := ctx.Attributes[r.Attribute]
	if !exists {
		return r.Operator == OpNotExists
	}

	if r.Operator == OpExists {
		return true
	}

	// Type-specific evaluation
	switch v := value.(type) {
	case string:
		return r.evaluateStringMatch(v)
	case float64:
		return r.evaluateNumeric(v)
	case int:
		return r.evaluateNumeric(float64(v))
	case int64:
		return r.evaluateNumeric(float64(v))
	case bool:
		target, ok := r.Value.(bool)
		if !ok {
			return false
		}
		return v == target
	default:
		return false
	}
}

// Helper functions

func compareVersions(v1, v2 string) int {
	// Simple semantic version comparison
	// Split by "." and compare numerically
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int

		if i < len(parts1) {
			// Parse numeric part (ignore non-numeric suffixes like "1.0.0-beta")
			numPart := strings.Split(parts1[i], "-")[0]
			n1 = parseVersionPart(numPart)
		}

		if i < len(parts2) {
			numPart := strings.Split(parts2[i], "-")[0]
			n2 = parseVersionPart(numPart)
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
}

func parseVersionPart(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			break
		}
	}
	return result
}

func hashString(s string) uint32 {
	// Simple FNV-1a hash
	const (
		offset32 = 2166136261
		prime32  = 16777619
	)

	hash := uint32(offset32)
	for i := 0; i < len(s); i++ {
		hash ^= uint32(s[i])
		hash *= prime32
	}
	return hash
}

// SegmentBuilder provides a fluent API for building targeting rules
type SegmentBuilder struct {
	rules []TargetingRule
}

// NewSegment creates a new segment builder
func NewSegment() *SegmentBuilder {
	return &SegmentBuilder{
		rules: make([]TargetingRule, 0),
	}
}

// UserID targets specific user IDs
func (sb *SegmentBuilder) UserID(ids ...string) *SegmentBuilder {
	values := make([]interface{}, len(ids))
	for i, id := range ids {
		values[i] = id
	}
	sb.rules = append(sb.rules, TargetingRule{
		Type:     RuleTypeUserID,
		Operator: OpIn,
		Values:   values,
	})
	return sb
}

// Country targets specific countries
func (sb *SegmentBuilder) Country(countries ...string) *SegmentBuilder {
	values := make([]interface{}, len(countries))
	for i, c := range countries {
		values[i] = c
	}
	sb.rules = append(sb.rules, TargetingRule{
		Type:     RuleTypeCountry,
		Operator: OpIn,
		Values:   values,
	})
	return sb
}

// Tier targets specific user tiers
func (sb *SegmentBuilder) Tier(tiers ...string) *SegmentBuilder {
	values := make([]interface{}, len(tiers))
	for i, t := range tiers {
		values[i] = t
	}
	sb.rules = append(sb.rules, TargetingRule{
		Type:     RuleTypeTier,
		Operator: OpIn,
		Values:   values,
	})
	return sb
}

// DeviceType targets specific device types
func (sb *SegmentBuilder) DeviceType(types ...string) *SegmentBuilder {
	values := make([]interface{}, len(types))
	for i, t := range types {
		values[i] = t
	}
	sb.rules = append(sb.rules, TargetingRule{
		Type:     RuleTypeDeviceType,
		Operator: OpIn,
		Values:   values,
	})
	return sb
}

// BetaTesters targets beta testers only
func (sb *SegmentBuilder) BetaTesters() *SegmentBuilder {
	sb.rules = append(sb.rules, TargetingRule{
		Type: RuleTypeBetaTester,
	})
	return sb
}

// VIP targets VIP users only
func (sb *SegmentBuilder) VIP() *SegmentBuilder {
	sb.rules = append(sb.rules, TargetingRule{
		Type: RuleTypeVIP,
	})
	return sb
}

// NewUsers targets new users only
func (sb *SegmentBuilder) NewUsers() *SegmentBuilder {
	sb.rules = append(sb.rules, TargetingRule{
		Type: RuleTypeNewUser,
	})
	return sb
}

// Percentage targets a percentage of users
func (sb *SegmentBuilder) Percentage(pct int) *SegmentBuilder {
	sb.rules = append(sb.rules, TargetingRule{
		Type:  RuleTypePercentage,
		Value: pct,
	})
	return sb
}

// Build returns the targeting rules
func (sb *SegmentBuilder) Build() []TargetingRule {
	return sb.rules
}
