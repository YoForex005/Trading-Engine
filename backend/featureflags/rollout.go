package featureflags

import (
	"sync"
	"time"
)

// RolloutManager manages gradual rollouts of features
type RolloutManager struct {
	rollouts map[string]*Rollout
	mu       sync.RWMutex

	// Monitoring
	metrics  map[string]*RolloutMetrics
	metricsMu sync.RWMutex

	// Auto-rollback
	autoRollback bool
	rollbackRules []RollbackRule
}

// Rollout represents a gradual rollout configuration
type Rollout struct {
	ID          string                 `json:"id"`
	FlagID      string                 `json:"flag_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      RolloutStatus          `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`

	// Rollout strategy
	Strategy    RolloutStrategy        `json:"strategy"`
	Stages      []RolloutStage         `json:"stages"`
	CurrentStage int                   `json:"current_stage"`

	// Monitoring
	HealthChecks []HealthCheck         `json:"health_checks"`
	Metrics      []string              `json:"metrics"`

	// Auto-progression
	AutoProgress bool                  `json:"auto_progress"`
	MinStageDuration time.Duration     `json:"min_stage_duration"`
	ProgressConditions []ProgressCondition `json:"progress_conditions,omitempty"`

	// Rollback
	RollbackOnError bool               `json:"rollback_on_error"`
	RollbackThresholds []RollbackThreshold `json:"rollback_thresholds,omitempty"`
}

// RolloutStatus represents the status of a rollout
type RolloutStatus string

const (
	RolloutStatusPending    RolloutStatus = "pending"
	RolloutStatusInProgress RolloutStatus = "in_progress"
	RolloutStatusPaused     RolloutStatus = "paused"
	RolloutStatusCompleted  RolloutStatus = "completed"
	RolloutStatusRolledBack RolloutStatus = "rolled_back"
	RolloutStatusFailed     RolloutStatus = "failed"
)

// RolloutStrategy represents the rollout strategy
type RolloutStrategy string

const (
	StrategyLinear      RolloutStrategy = "linear"      // Increase by fixed percentage
	StrategyExponential RolloutStrategy = "exponential" // Double each stage
	StrategyCustom      RolloutStrategy = "custom"      // Custom stages
	StrategyCanary      RolloutStrategy = "canary"      // Small group first, then all
	StrategyBlueGreen   RolloutStrategy = "blue_green"  // 0% -> 100%
)

// RolloutStage represents a stage in the rollout
type RolloutStage struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Percentage  int           `json:"percentage"` // Percentage of users
	Duration    time.Duration `json:"duration"`   // Minimum duration
	StartedAt   *time.Time    `json:"started_at,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	Status      string        `json:"status"` // pending, active, completed, failed
}

// HealthCheck represents a health check for monitoring
type HealthCheck struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Type        string        `json:"type"` // error_rate, latency, custom
	Threshold   float64       `json:"threshold"`
	Interval    time.Duration `json:"interval"`
	Enabled     bool          `json:"enabled"`
}

// ProgressCondition represents a condition for auto-progression
type ProgressCondition struct {
	Metric    string  `json:"metric"`
	Operator  string  `json:"operator"` // gt, lt, gte, lte, eq
	Threshold float64 `json:"threshold"`
}

// RollbackThreshold represents a threshold for automatic rollback
type RollbackThreshold struct {
	Metric    string  `json:"metric"`
	Threshold float64 `json:"threshold"`
	Duration  time.Duration `json:"duration"` // Threshold must be exceeded for this duration
}

// RollbackRule represents a rule for automatic rollback
type RollbackRule struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Metric      string  `json:"metric"`
	Operator    string  `json:"operator"`
	Threshold   float64 `json:"threshold"`
	Enabled     bool    `json:"enabled"`
}

// RolloutMetrics represents metrics for a rollout
type RolloutMetrics struct {
	RolloutID      string              `json:"rollout_id"`
	CurrentStage   int                 `json:"current_stage"`
	UsersInRollout int64               `json:"users_in_rollout"`
	ErrorRate      float64             `json:"error_rate"`
	Latency        float64             `json:"latency"` // ms
	CustomMetrics  map[string]float64  `json:"custom_metrics"`
	HealthStatus   string              `json:"health_status"` // healthy, degraded, unhealthy
	LastUpdated    time.Time           `json:"last_updated"`
}

// NewRolloutManager creates a new rollout manager
func NewRolloutManager(autoRollback bool) *RolloutManager {
	return &RolloutManager{
		rollouts:     make(map[string]*Rollout),
		metrics:      make(map[string]*RolloutMetrics),
		autoRollback: autoRollback,
		rollbackRules: []RollbackRule{
			{
				ID:        "error-rate",
				Name:      "High Error Rate",
				Metric:    "error_rate",
				Operator:  "gt",
				Threshold: 0.05, // 5% error rate
				Enabled:   true,
			},
			{
				ID:        "latency",
				Name:      "High Latency",
				Metric:    "latency",
				Operator:  "gt",
				Threshold: 1000, // 1000ms
				Enabled:   true,
			},
		},
	}
}

// CreateRollout creates a new rollout
func (rm *RolloutManager) CreateRollout(rollout *Rollout) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.rollouts[rollout.ID]; exists {
		return ErrRolloutExists
	}

	rollout.CreatedAt = time.Now()
	rollout.UpdatedAt = time.Now()
	rollout.Status = RolloutStatusPending
	rollout.CurrentStage = 0

	// Generate stages based on strategy if not custom
	if rollout.Strategy != StrategyCustom && len(rollout.Stages) == 0 {
		rollout.Stages = rm.generateStages(rollout.Strategy)
	}

	// Validate stages
	if err := rm.validateStages(rollout.Stages); err != nil {
		return err
	}

	rm.rollouts[rollout.ID] = rollout

	// Initialize metrics
	rm.initializeMetrics(rollout.ID)

	return nil
}

// StartRollout starts a rollout
func (rm *RolloutManager) StartRollout(rolloutID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rollout, exists := rm.rollouts[rolloutID]
	if !exists {
		return ErrRolloutNotFound
	}

	if rollout.Status != RolloutStatusPending && rollout.Status != RolloutStatusPaused {
		return ErrInvalidRolloutStatus
	}

	now := time.Now()
	rollout.Status = RolloutStatusInProgress
	rollout.StartedAt = &now
	rollout.UpdatedAt = now

	// Start first stage
	if len(rollout.Stages) > 0 {
		rollout.Stages[0].Status = "active"
		rollout.Stages[0].StartedAt = &now
	}

	return nil
}

// ProgressToNextStage progresses rollout to next stage
func (rm *RolloutManager) ProgressToNextStage(rolloutID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rollout, exists := rm.rollouts[rolloutID]
	if !exists {
		return ErrRolloutNotFound
	}

	if rollout.Status != RolloutStatusInProgress {
		return ErrInvalidRolloutStatus
	}

	// Check if there's a next stage
	if rollout.CurrentStage >= len(rollout.Stages)-1 {
		// Complete the rollout
		return rm.completeRolloutInternal(rollout)
	}

	// Complete current stage
	now := time.Now()
	currentStage := &rollout.Stages[rollout.CurrentStage]
	currentStage.Status = "completed"
	currentStage.CompletedAt = &now

	// Start next stage
	rollout.CurrentStage++
	nextStage := &rollout.Stages[rollout.CurrentStage]
	nextStage.Status = "active"
	nextStage.StartedAt = &now
	rollout.UpdatedAt = now

	return nil
}

// PauseRollout pauses a rollout
func (rm *RolloutManager) PauseRollout(rolloutID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rollout, exists := rm.rollouts[rolloutID]
	if !exists {
		return ErrRolloutNotFound
	}

	if rollout.Status != RolloutStatusInProgress {
		return ErrInvalidRolloutStatus
	}

	rollout.Status = RolloutStatusPaused
	rollout.UpdatedAt = time.Now()

	return nil
}

// RollbackRollout rolls back a rollout
func (rm *RolloutManager) RollbackRollout(rolloutID string, reason string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rollout, exists := rm.rollouts[rolloutID]
	if !exists {
		return ErrRolloutNotFound
	}

	now := time.Now()
	rollout.Status = RolloutStatusRolledBack
	rollout.CompletedAt = &now
	rollout.UpdatedAt = now

	// Mark current stage as failed
	if rollout.CurrentStage < len(rollout.Stages) {
		currentStage := &rollout.Stages[rollout.CurrentStage]
		currentStage.Status = "failed"
		currentStage.CompletedAt = &now
	}

	return nil
}

// UpdateMetrics updates rollout metrics
func (rm *RolloutManager) UpdateMetrics(rolloutID string, metrics *RolloutMetrics) {
	rm.metricsMu.Lock()
	defer rm.metricsMu.Unlock()

	metrics.RolloutID = rolloutID
	metrics.LastUpdated = time.Now()
	rm.metrics[rolloutID] = metrics

	// Check for auto-rollback conditions
	if rm.autoRollback {
		go rm.checkRollbackConditions(rolloutID, metrics)
	}

	// Check for auto-progression
	rm.mu.RLock()
	rollout, exists := rm.rollouts[rolloutID]
	rm.mu.RUnlock()

	if exists && rollout.AutoProgress {
		go rm.checkProgressConditions(rolloutID, rollout, metrics)
	}
}

// checkRollbackConditions checks if rollback conditions are met
func (rm *RolloutManager) checkRollbackConditions(rolloutID string, metrics *RolloutMetrics) {
	rm.mu.RLock()
	rollout, exists := rm.rollouts[rolloutID]
	rm.mu.RUnlock()

	if !exists || rollout.Status != RolloutStatusInProgress {
		return
	}

	// Check rollout-specific thresholds
	for _, threshold := range rollout.RollbackThresholds {
		if rm.checkThreshold(metrics, threshold.Metric, threshold.Threshold) {
			rm.RollbackRollout(rolloutID, "threshold exceeded: "+threshold.Metric)
			return
		}
	}

	// Check global rollback rules
	for _, rule := range rm.rollbackRules {
		if !rule.Enabled {
			continue
		}

		if rm.checkRule(metrics, rule) {
			rm.RollbackRollout(rolloutID, "rule triggered: "+rule.Name)
			return
		}
	}
}

// checkProgressConditions checks if progression conditions are met
func (rm *RolloutManager) checkProgressConditions(rolloutID string, rollout *Rollout, metrics *RolloutMetrics) {
	if rollout.Status != RolloutStatusInProgress {
		return
	}

	// Check minimum stage duration
	if rollout.CurrentStage < len(rollout.Stages) {
		currentStage := rollout.Stages[rollout.CurrentStage]
		if currentStage.StartedAt != nil {
			elapsed := time.Since(*currentStage.StartedAt)
			if elapsed < rollout.MinStageDuration {
				return
			}
		}
	}

	// Check progress conditions
	allConditionsMet := true
	for _, condition := range rollout.ProgressConditions {
		if !rm.checkProgressCondition(metrics, condition) {
			allConditionsMet = false
			break
		}
	}

	if allConditionsMet {
		rm.ProgressToNextStage(rolloutID)
	}
}

// Helper methods

func (rm *RolloutManager) generateStages(strategy RolloutStrategy) []RolloutStage {
	stages := []RolloutStage{}

	switch strategy {
	case StrategyLinear:
		// 10%, 25%, 50%, 75%, 100%
		percentages := []int{10, 25, 50, 75, 100}
		for i, pct := range percentages {
			stages = append(stages, RolloutStage{
				ID:         generateID(),
				Name:       "Stage " + string(rune('A'+i)),
				Percentage: pct,
				Duration:   1 * time.Hour,
				Status:     "pending",
			})
		}

	case StrategyExponential:
		// 1%, 2%, 4%, 8%, 16%, 32%, 64%, 100%
		percentages := []int{1, 2, 4, 8, 16, 32, 64, 100}
		for i, pct := range percentages {
			stages = append(stages, RolloutStage{
				ID:         generateID(),
				Name:       "Stage " + string(rune('A'+i)),
				Percentage: pct,
				Duration:   30 * time.Minute,
				Status:     "pending",
			})
		}

	case StrategyCanary:
		// 1%, 100%
		stages = []RolloutStage{
			{
				ID:         generateID(),
				Name:       "Canary",
				Percentage: 1,
				Duration:   2 * time.Hour,
				Status:     "pending",
			},
			{
				ID:         generateID(),
				Name:       "Full Rollout",
				Percentage: 100,
				Duration:   1 * time.Hour,
				Status:     "pending",
			},
		}

	case StrategyBlueGreen:
		// 0%, 100%
		stages = []RolloutStage{
			{
				ID:         generateID(),
				Name:       "Preparation",
				Percentage: 0,
				Duration:   30 * time.Minute,
				Status:     "pending",
			},
			{
				ID:         generateID(),
				Name:       "Switch",
				Percentage: 100,
				Duration:   0,
				Status:     "pending",
			},
		}
	}

	return stages
}

func (rm *RolloutManager) validateStages(stages []RolloutStage) error {
	if len(stages) == 0 {
		return ErrInvalidRollout("at least one stage required")
	}

	// Last stage must be 100%
	if stages[len(stages)-1].Percentage != 100 {
		return ErrInvalidRollout("final stage must be 100%")
	}

	// Stages must be in increasing order
	for i := 1; i < len(stages); i++ {
		if stages[i].Percentage <= stages[i-1].Percentage {
			return ErrInvalidRollout("stages must be in increasing percentage order")
		}
	}

	return nil
}

func (rm *RolloutManager) completeRolloutInternal(rollout *Rollout) error {
	now := time.Now()
	rollout.Status = RolloutStatusCompleted
	rollout.CompletedAt = &now
	rollout.UpdatedAt = now

	// Complete final stage
	if rollout.CurrentStage < len(rollout.Stages) {
		finalStage := &rollout.Stages[rollout.CurrentStage]
		finalStage.Status = "completed"
		finalStage.CompletedAt = &now
	}

	return nil
}

func (rm *RolloutManager) initializeMetrics(rolloutID string) {
	rm.metricsMu.Lock()
	defer rm.metricsMu.Unlock()

	rm.metrics[rolloutID] = &RolloutMetrics{
		RolloutID:     rolloutID,
		CurrentStage:  0,
		CustomMetrics: make(map[string]float64),
		HealthStatus:  "healthy",
		LastUpdated:   time.Now(),
	}
}

func (rm *RolloutManager) checkThreshold(metrics *RolloutMetrics, metricName string, threshold float64) bool {
	var value float64

	switch metricName {
	case "error_rate":
		value = metrics.ErrorRate
	case "latency":
		value = metrics.Latency
	default:
		if v, ok := metrics.CustomMetrics[metricName]; ok {
			value = v
		} else {
			return false
		}
	}

	return value > threshold
}

func (rm *RolloutManager) checkRule(metrics *RolloutMetrics, rule RollbackRule) bool {
	var value float64

	switch rule.Metric {
	case "error_rate":
		value = metrics.ErrorRate
	case "latency":
		value = metrics.Latency
	default:
		if v, ok := metrics.CustomMetrics[rule.Metric]; ok {
			value = v
		} else {
			return false
		}
	}

	switch rule.Operator {
	case "gt":
		return value > rule.Threshold
	case "lt":
		return value < rule.Threshold
	case "gte":
		return value >= rule.Threshold
	case "lte":
		return value <= rule.Threshold
	case "eq":
		return value == rule.Threshold
	default:
		return false
	}
}

func (rm *RolloutManager) checkProgressCondition(metrics *RolloutMetrics, condition ProgressCondition) bool {
	var value float64

	switch condition.Metric {
	case "error_rate":
		value = metrics.ErrorRate
	case "latency":
		value = metrics.Latency
	default:
		if v, ok := metrics.CustomMetrics[condition.Metric]; ok {
			value = v
		} else {
			return false
		}
	}

	switch condition.Operator {
	case "gt":
		return value > condition.Threshold
	case "lt":
		return value < condition.Threshold
	case "gte":
		return value >= condition.Threshold
	case "lte":
		return value <= condition.Threshold
	case "eq":
		return value == condition.Threshold
	default:
		return false
	}
}

// GetRollout retrieves a rollout
func (rm *RolloutManager) GetRollout(rolloutID string) (*Rollout, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	rollout, exists := rm.rollouts[rolloutID]
	return rollout, exists
}

// GetMetrics retrieves rollout metrics
func (rm *RolloutManager) GetMetrics(rolloutID string) (*RolloutMetrics, bool) {
	rm.metricsMu.RLock()
	defer rm.metricsMu.RUnlock()
	metrics, exists := rm.metrics[rolloutID]
	return metrics, exists
}

// ListRollouts lists all rollouts
func (rm *RolloutManager) ListRollouts() []*Rollout {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rollouts := make([]*Rollout, 0, len(rm.rollouts))
	for _, rollout := range rm.rollouts {
		rollouts = append(rollouts, rollout)
	}
	return rollouts
}

// GetCurrentPercentage gets the current rollout percentage
func (rm *RolloutManager) GetCurrentPercentage(rolloutID string) int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rollout, exists := rm.rollouts[rolloutID]
	if !exists || rollout.CurrentStage >= len(rollout.Stages) {
		return 0
	}

	return rollout.Stages[rollout.CurrentStage].Percentage
}

// IsUserInRollout checks if a user is included in the current rollout stage
func (rm *RolloutManager) IsUserInRollout(rolloutID, userID string) bool {
	percentage := rm.GetCurrentPercentage(rolloutID)
	if percentage == 0 {
		return false
	}
	if percentage >= 100 {
		return true
	}

	// Use consistent hashing
	hash := hashString(userID)
	value := hash % 100
	return int(value) < percentage
}
