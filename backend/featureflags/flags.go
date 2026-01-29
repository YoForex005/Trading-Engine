package featureflags

import (
	"crypto/sha256"
	"encoding/binary"
	"sync"
	"time"
)

// FlagType represents the type of feature flag
type FlagType string

const (
	FlagTypeBoolean     FlagType = "boolean"
	FlagTypeMultivariate FlagType = "multivariate"
	FlagTypePercentage  FlagType = "percentage"
	FlagTypeKillSwitch  FlagType = "killswitch"
)

// Flag represents a feature flag configuration
type Flag struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Type         FlagType               `json:"type"`
	Enabled      bool                   `json:"enabled"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	CreatedBy    string                 `json:"created_by"`

	// Boolean flag
	DefaultValue bool                   `json:"default_value,omitempty"`

	// Multivariate flag (A/B/C testing)
	Variants     []Variant              `json:"variants,omitempty"`

	// Percentage rollout
	Percentage   int                    `json:"percentage,omitempty"` // 0-100

	// Targeting rules
	Rules        []TargetingRule        `json:"rules,omitempty"`

	// Dependencies
	Dependencies []FlagDependency       `json:"dependencies,omitempty"`

	// Scheduling
	EnableAt     *time.Time             `json:"enable_at,omitempty"`
	DisableAt    *time.Time             `json:"disable_at,omitempty"`

	// Metadata
	Tags         []string               `json:"tags,omitempty"`
	Environment  string                 `json:"environment,omitempty"` // dev, staging, prod

	// Audit
	Version      int                    `json:"version"`
}

// Variant represents a variant in multivariate testing
type Variant struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Weight      int                    `json:"weight"` // 0-100
	Payload     map[string]interface{} `json:"payload,omitempty"`
}

// FlagDependency represents a dependency between flags
type FlagDependency struct {
	FlagID   string `json:"flag_id"`
	Operator string `json:"operator"` // "enabled", "disabled", "equals"
	Value    string `json:"value,omitempty"`
}

// FlagEvaluation represents the result of evaluating a flag
type FlagEvaluation struct {
	FlagID      string                 `json:"flag_id"`
	Enabled     bool                   `json:"enabled"`
	Variant     *Variant               `json:"variant,omitempty"`
	Reason      string                 `json:"reason"`
	EvaluatedAt time.Time              `json:"evaluated_at"`
	UserID      string                 `json:"user_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// FlagHistory represents a historical change to a flag
type FlagHistory struct {
	ID        string                 `json:"id"`
	FlagID    string                 `json:"flag_id"`
	Action    string                 `json:"action"` // created, updated, deleted, toggled
	ChangedBy string                 `json:"changed_by"`
	ChangedAt time.Time              `json:"changed_at"`
	Before    *Flag                  `json:"before,omitempty"`
	After     *Flag                  `json:"after,omitempty"`
	Reason    string                 `json:"reason,omitempty"`
}

// FlagManager manages feature flags with <1ms evaluation
type FlagManager struct {
	flags       map[string]*Flag
	history     []FlagHistory
	mu          sync.RWMutex

	// Cache for fast lookups
	cache       sync.Map // map[string]*FlagEvaluation
	cacheTTL    time.Duration

	// Callbacks
	onChange    []func(*Flag, *Flag) // before, after
	onEvaluate  []func(*FlagEvaluation)
}

// NewFlagManager creates a new flag manager
func NewFlagManager() *FlagManager {
	return &FlagManager{
		flags:    make(map[string]*Flag),
		history:  make([]FlagHistory, 0),
		cacheTTL: 5 * time.Second,
		onChange: make([]func(*Flag, *Flag), 0),
		onEvaluate: make([]func(*FlagEvaluation), 0),
	}
}

// CreateFlag creates a new feature flag
func (fm *FlagManager) CreateFlag(flag *Flag, createdBy string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if _, exists := fm.flags[flag.ID]; exists {
		return ErrFlagExists
	}

	now := time.Now()
	flag.CreatedAt = now
	flag.UpdatedAt = now
	flag.CreatedBy = createdBy
	flag.Version = 1

	// Validate flag
	if err := fm.validateFlag(flag); err != nil {
		return err
	}

	fm.flags[flag.ID] = flag

	// Record history
	fm.recordHistory(FlagHistory{
		FlagID:    flag.ID,
		Action:    "created",
		ChangedBy: createdBy,
		ChangedAt: now,
		After:     flag,
	})

	// Notify listeners
	for _, callback := range fm.onChange {
		go callback(nil, flag)
	}

	return nil
}

// UpdateFlag updates an existing flag
func (fm *FlagManager) UpdateFlag(flagID string, updates *Flag, updatedBy string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	oldFlag, exists := fm.flags[flagID]
	if !exists {
		return ErrFlagNotFound
	}

	// Create a copy for history
	before := *oldFlag

	// Update fields
	updates.UpdatedAt = time.Now()
	updates.Version = oldFlag.Version + 1

	// Validate updates
	if err := fm.validateFlag(updates); err != nil {
		return err
	}

	fm.flags[flagID] = updates

	// Clear cache for this flag
	fm.cache.Delete(flagID)

	// Record history
	fm.recordHistory(FlagHistory{
		FlagID:    flagID,
		Action:    "updated",
		ChangedBy: updatedBy,
		ChangedAt: time.Now(),
		Before:    &before,
		After:     updates,
	})

	// Notify listeners
	for _, callback := range fm.onChange {
		go callback(&before, updates)
	}

	return nil
}

// DeleteFlag deletes a flag
func (fm *FlagManager) DeleteFlag(flagID string, deletedBy string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	flag, exists := fm.flags[flagID]
	if !exists {
		return ErrFlagNotFound
	}

	// Create a copy for history
	before := *flag

	delete(fm.flags, flagID)
	fm.cache.Delete(flagID)

	// Record history
	fm.recordHistory(FlagHistory{
		FlagID:    flagID,
		Action:    "deleted",
		ChangedBy: deletedBy,
		ChangedAt: time.Now(),
		Before:    &before,
	})

	// Notify listeners
	for _, callback := range fm.onChange {
		go callback(&before, nil)
	}

	return nil
}

// ToggleFlag quickly toggles a flag on/off
func (fm *FlagManager) ToggleFlag(flagID string, enabled bool, toggledBy string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	flag, exists := fm.flags[flagID]
	if !exists {
		return ErrFlagNotFound
	}

	// Create a copy for history
	before := *flag

	flag.Enabled = enabled
	flag.UpdatedAt = time.Now()
	flag.Version++

	fm.cache.Delete(flagID)

	// Record history
	fm.recordHistory(FlagHistory{
		FlagID:    flagID,
		Action:    "toggled",
		ChangedBy: toggledBy,
		ChangedAt: time.Now(),
		Before:    &before,
		After:     flag,
		Reason:    map[bool]string{true: "enabled", false: "disabled"}[enabled],
	})

	// Notify listeners
	for _, callback := range fm.onChange {
		go callback(&before, flag)
	}

	return nil
}

// Evaluate evaluates a flag for a given context (< 1ms target)
func (fm *FlagManager) Evaluate(flagID string, ctx *EvaluationContext) *FlagEvaluation {
	startTime := time.Now()

	// Check cache first (fastest path)
	cacheKey := fm.getCacheKey(flagID, ctx)
	if cached, ok := fm.cache.Load(cacheKey); ok {
		eval := cached.(*FlagEvaluation)
		// Check if cache is still valid
		if time.Since(eval.EvaluatedAt) < fm.cacheTTL {
			return eval
		}
		fm.cache.Delete(cacheKey)
	}

	// Get flag (read lock for performance)
	fm.mu.RLock()
	flag, exists := fm.flags[flagID]
	fm.mu.RUnlock()

	if !exists {
		return &FlagEvaluation{
			FlagID:      flagID,
			Enabled:     false,
			Reason:      "flag_not_found",
			EvaluatedAt: startTime,
			UserID:      ctx.UserID,
		}
	}

	// Quick checks
	if !flag.Enabled {
		eval := &FlagEvaluation{
			FlagID:      flagID,
			Enabled:     false,
			Reason:      "flag_disabled",
			EvaluatedAt: startTime,
			UserID:      ctx.UserID,
		}
		fm.cacheEvaluation(cacheKey, eval)
		return eval
	}

	// Check scheduling
	now := time.Now()
	if flag.EnableAt != nil && now.Before(*flag.EnableAt) {
		eval := &FlagEvaluation{
			FlagID:      flagID,
			Enabled:     false,
			Reason:      "not_yet_enabled",
			EvaluatedAt: startTime,
			UserID:      ctx.UserID,
		}
		fm.cacheEvaluation(cacheKey, eval)
		return eval
	}

	if flag.DisableAt != nil && now.After(*flag.DisableAt) {
		eval := &FlagEvaluation{
			FlagID:      flagID,
			Enabled:     false,
			Reason:      "expired",
			EvaluatedAt: startTime,
			UserID:      ctx.UserID,
		}
		fm.cacheEvaluation(cacheKey, eval)
		return eval
	}

	// Check dependencies
	if !fm.checkDependencies(flag, ctx) {
		eval := &FlagEvaluation{
			FlagID:      flagID,
			Enabled:     false,
			Reason:      "dependency_not_met",
			EvaluatedAt: startTime,
			UserID:      ctx.UserID,
		}
		fm.cacheEvaluation(cacheKey, eval)
		return eval
	}

	// Evaluate targeting rules
	if len(flag.Rules) > 0 {
		if !fm.evaluateRules(flag.Rules, ctx) {
			eval := &FlagEvaluation{
				FlagID:      flagID,
				Enabled:     false,
				Reason:      "targeting_rule_not_met",
				EvaluatedAt: startTime,
				UserID:      ctx.UserID,
			}
			fm.cacheEvaluation(cacheKey, eval)
			return eval
		}
	}

	// Evaluate based on flag type
	var eval *FlagEvaluation
	switch flag.Type {
	case FlagTypeBoolean:
		eval = &FlagEvaluation{
			FlagID:      flagID,
			Enabled:     flag.DefaultValue,
			Reason:      "boolean_flag",
			EvaluatedAt: startTime,
			UserID:      ctx.UserID,
		}

	case FlagTypePercentage:
		enabled := fm.isInPercentage(ctx.UserID, flag.Percentage)
		eval = &FlagEvaluation{
			FlagID:      flagID,
			Enabled:     enabled,
			Reason:      map[bool]string{true: "in_percentage", false: "out_of_percentage"}[enabled],
			EvaluatedAt: startTime,
			UserID:      ctx.UserID,
		}

	case FlagTypeMultivariate:
		variant := fm.selectVariant(flag.Variants, ctx.UserID)
		eval = &FlagEvaluation{
			FlagID:      flagID,
			Enabled:     variant != nil,
			Variant:     variant,
			Reason:      "multivariate_assignment",
			EvaluatedAt: startTime,
			UserID:      ctx.UserID,
		}

	case FlagTypeKillSwitch:
		eval = &FlagEvaluation{
			FlagID:      flagID,
			Enabled:     false,
			Reason:      "kill_switch_active",
			EvaluatedAt: startTime,
			UserID:      ctx.UserID,
		}

	default:
		eval = &FlagEvaluation{
			FlagID:      flagID,
			Enabled:     false,
			Reason:      "unknown_flag_type",
			EvaluatedAt: startTime,
			UserID:      ctx.UserID,
		}
	}

	// Cache the evaluation
	fm.cacheEvaluation(cacheKey, eval)

	// Notify listeners (async to not impact performance)
	for _, callback := range fm.onEvaluate {
		go callback(eval)
	}

	return eval
}

// IsEnabled is a convenience method for boolean flags
func (fm *FlagManager) IsEnabled(flagID string, ctx *EvaluationContext) bool {
	eval := fm.Evaluate(flagID, ctx)
	return eval.Enabled
}

// GetVariant gets the variant for multivariate flags
func (fm *FlagManager) GetVariant(flagID string, ctx *EvaluationContext) *Variant {
	eval := fm.Evaluate(flagID, ctx)
	return eval.Variant
}

// Helper methods

func (fm *FlagManager) validateFlag(flag *Flag) error {
	if flag.ID == "" {
		return ErrInvalidFlag("id is required")
	}
	if flag.Name == "" {
		return ErrInvalidFlag("name is required")
	}

	// Validate variants for multivariate flags
	if flag.Type == FlagTypeMultivariate {
		if len(flag.Variants) == 0 {
			return ErrInvalidFlag("multivariate flag requires at least one variant")
		}
		totalWeight := 0
		for _, v := range flag.Variants {
			totalWeight += v.Weight
		}
		if totalWeight != 100 {
			return ErrInvalidFlag("variant weights must sum to 100")
		}
	}

	// Validate percentage
	if flag.Type == FlagTypePercentage {
		if flag.Percentage < 0 || flag.Percentage > 100 {
			return ErrInvalidFlag("percentage must be between 0 and 100")
		}
	}

	return nil
}

func (fm *FlagManager) recordHistory(entry FlagHistory) {
	entry.ID = generateID()
	fm.history = append(fm.history, entry)

	// Keep only last 1000 entries per flag (prevent memory bloat)
	if len(fm.history) > 10000 {
		fm.history = fm.history[len(fm.history)-1000:]
	}
}

func (fm *FlagManager) checkDependencies(flag *Flag, ctx *EvaluationContext) bool {
	for _, dep := range flag.Dependencies {
		depFlag, exists := fm.flags[dep.FlagID]
		if !exists {
			return false
		}

		switch dep.Operator {
		case "enabled":
			if !depFlag.Enabled {
				return false
			}
		case "disabled":
			if depFlag.Enabled {
				return false
			}
		case "equals":
			// Would need to evaluate the dependent flag and check variant
			depEval := fm.Evaluate(dep.FlagID, ctx)
			if depEval.Variant == nil || depEval.Variant.ID != dep.Value {
				return false
			}
		}
	}
	return true
}

func (fm *FlagManager) evaluateRules(rules []TargetingRule, ctx *EvaluationContext) bool {
	// All rules must pass (AND logic)
	// For OR logic, use multiple flags
	for _, rule := range rules {
		if !rule.Evaluate(ctx) {
			return false
		}
	}
	return true
}

// isInPercentage uses consistent hashing for percentage rollouts
func (fm *FlagManager) isInPercentage(userID string, percentage int) bool {
	if percentage == 0 {
		return false
	}
	if percentage >= 100 {
		return true
	}

	// Hash user ID for consistent assignment
	hash := sha256.Sum256([]byte(userID))
	value := binary.BigEndian.Uint32(hash[:4]) % 100

	return int(value) < percentage
}

// selectVariant selects a variant based on consistent hashing
func (fm *FlagManager) selectVariant(variants []Variant, userID string) *Variant {
	if len(variants) == 0 {
		return nil
	}

	// Hash user ID for consistent assignment
	hash := sha256.Sum256([]byte(userID))
	value := binary.BigEndian.Uint32(hash[:4]) % 100

	// Select variant based on cumulative weights
	cumulative := 0
	for i := range variants {
		cumulative += variants[i].Weight
		if int(value) < cumulative {
			return &variants[i]
		}
	}

	// Fallback to first variant (should not happen if weights sum to 100)
	return &variants[0]
}

func (fm *FlagManager) getCacheKey(flagID string, ctx *EvaluationContext) string {
	// Simple cache key combining flag ID and user ID
	// For more complex caching, include other context attributes
	return flagID + ":" + ctx.UserID
}

func (fm *FlagManager) cacheEvaluation(key string, eval *FlagEvaluation) {
	fm.cache.Store(key, eval)
}

// GetFlag retrieves a flag by ID
func (fm *FlagManager) GetFlag(flagID string) (*Flag, bool) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	flag, exists := fm.flags[flagID]
	return flag, exists
}

// ListFlags lists all flags
func (fm *FlagManager) ListFlags() []*Flag {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	flags := make([]*Flag, 0, len(fm.flags))
	for _, flag := range fm.flags {
		flags = append(flags, flag)
	}
	return flags
}

// GetHistory retrieves flag history
func (fm *FlagManager) GetHistory(flagID string, limit int) []FlagHistory {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	history := make([]FlagHistory, 0)
	count := 0

	// Iterate in reverse (most recent first)
	for i := len(fm.history) - 1; i >= 0 && count < limit; i-- {
		if fm.history[i].FlagID == flagID {
			history = append(history, fm.history[i])
			count++
		}
	}

	return history
}

// OnChange registers a callback for flag changes
func (fm *FlagManager) OnChange(callback func(*Flag, *Flag)) {
	fm.onChange = append(fm.onChange, callback)
}

// OnEvaluate registers a callback for flag evaluations
func (fm *FlagManager) OnEvaluate(callback func(*FlagEvaluation)) {
	fm.onEvaluate = append(fm.onEvaluate, callback)
}

// ClearCache clears the evaluation cache
func (fm *FlagManager) ClearCache() {
	fm.cache = sync.Map{}
}
