package featureflags

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// Experiment represents an A/B testing experiment
type Experiment struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	FlagID          string                 `json:"flag_id"`
	Status          ExperimentStatus       `json:"status"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	StartedAt       *time.Time             `json:"started_at,omitempty"`
	EndedAt         *time.Time             `json:"ended_at,omitempty"`
	CreatedBy       string                 `json:"created_by"`

	// Variants
	Variants        []ExperimentVariant    `json:"variants"`
	ControlVariant  string                 `json:"control_variant"` // ID of control variant

	// Traffic allocation
	TrafficAllocation int                  `json:"traffic_allocation"` // Percentage of users in experiment

	// Goals and metrics
	PrimaryMetric   string                 `json:"primary_metric"`
	SecondaryMetrics []string              `json:"secondary_metrics,omitempty"`
	Goals           []ExperimentGoal       `json:"goals,omitempty"`

	// Targeting
	Rules           []TargetingRule        `json:"rules,omitempty"`

	// Configuration
	MinSampleSize   int                    `json:"min_sample_size"`
	MinDuration     time.Duration          `json:"min_duration"`
	MaxDuration     time.Duration          `json:"max_duration"`
	Confidence      float64                `json:"confidence"` // 0.95 = 95% confidence
	PowerAnalysis   float64                `json:"power_analysis"` // 0.8 = 80% power

	// Early stopping
	EarlyStoppingEnabled bool              `json:"early_stopping_enabled"`
	EarlyStoppingRules   []EarlyStoppingRule `json:"early_stopping_rules,omitempty"`

	// Winner declaration
	WinnerVariantID *string                `json:"winner_variant_id,omitempty"`
	WinnerDeclaredAt *time.Time            `json:"winner_declared_at,omitempty"`

	// Metadata
	Tags            []string               `json:"tags,omitempty"`
	Environment     string                 `json:"environment,omitempty"`
}

// ExperimentStatus represents the status of an experiment
type ExperimentStatus string

const (
	ExperimentStatusDraft     ExperimentStatus = "draft"
	ExperimentStatusRunning   ExperimentStatus = "running"
	ExperimentStatusPaused    ExperimentStatus = "paused"
	ExperimentStatusCompleted ExperimentStatus = "completed"
	ExperimentStatusStopped   ExperimentStatus = "stopped"
)

// ExperimentVariant represents a variant in an experiment
type ExperimentVariant struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	IsControl   bool                   `json:"is_control"`
	Weight      int                    `json:"weight"` // Traffic split percentage
	Config      map[string]interface{} `json:"config,omitempty"`
}

// ExperimentGoal represents a conversion goal
type ExperimentGoal struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        string  `json:"type"` // conversion, revenue, engagement, custom
	Target      float64 `json:"target,omitempty"`
}

// EarlyStoppingRule represents a rule for early stopping
type EarlyStoppingRule struct {
	Type      string  `json:"type"` // statistical_significance, futility, harm
	Threshold float64 `json:"threshold"`
	Metric    string  `json:"metric"`
}

// ExperimentResult represents the results of an experiment
type ExperimentResult struct {
	ExperimentID    string                    `json:"experiment_id"`
	VariantResults  map[string]*VariantResult `json:"variant_results"`
	Winner          *string                   `json:"winner,omitempty"`
	Confidence      float64                   `json:"confidence"`
	CalculatedAt    time.Time                 `json:"calculated_at"`
	SampleSizeReached bool                    `json:"sample_size_reached"`
	Duration        time.Duration             `json:"duration"`
}

// VariantResult represents the results for a single variant
type VariantResult struct {
	VariantID       string                 `json:"variant_id"`
	VariantName     string                 `json:"variant_name"`
	IsControl       bool                   `json:"is_control"`
	Impressions     int64                  `json:"impressions"`
	UniqueUsers     int64                  `json:"unique_users"`
	Conversions     int64                  `json:"conversions"`
	ConversionRate  float64                `json:"conversion_rate"`
	Revenue         float64                `json:"revenue,omitempty"`
	RevenuePerUser  float64                `json:"revenue_per_user,omitempty"`
	CustomMetrics   map[string]float64     `json:"custom_metrics,omitempty"`

	// Statistical analysis
	StandardError   float64                `json:"standard_error"`
	ConfidenceInterval [2]float64          `json:"confidence_interval"`
	PValue          float64                `json:"p_value,omitempty"`
	Uplift          float64                `json:"uplift,omitempty"` // vs control
	IsSignificant   bool                   `json:"is_significant"`
}

// ExperimentManager manages A/B testing experiments
type ExperimentManager struct {
	experiments map[string]*Experiment
	results     map[string]*ExperimentResult
	events      []ExperimentEvent
	mu          sync.RWMutex

	// Analytics tracking
	impressions map[string]map[string]int64 // experiment -> variant -> count
	conversions map[string]map[string]map[string]int64 // experiment -> variant -> goal -> count
	uniqueUsers map[string]map[string]map[string]bool // experiment -> variant -> userID -> bool
	revenue     map[string]map[string]float64 // experiment -> variant -> total revenue

	eventsMu    sync.RWMutex
}

// ExperimentEvent represents an event in an experiment
type ExperimentEvent struct {
	ID           string                 `json:"id"`
	ExperimentID string                 `json:"experiment_id"`
	VariantID    string                 `json:"variant_id"`
	UserID       string                 `json:"user_id"`
	EventType    string                 `json:"event_type"` // impression, conversion, custom
	EventName    string                 `json:"event_name,omitempty"`
	Value        float64                `json:"value,omitempty"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// NewExperimentManager creates a new experiment manager
func NewExperimentManager() *ExperimentManager {
	return &ExperimentManager{
		experiments: make(map[string]*Experiment),
		results:     make(map[string]*ExperimentResult),
		events:      make([]ExperimentEvent, 0),
		impressions: make(map[string]map[string]int64),
		conversions: make(map[string]map[string]map[string]int64),
		uniqueUsers: make(map[string]map[string]map[string]bool),
		revenue:     make(map[string]map[string]float64),
	}
}

// CreateExperiment creates a new experiment
func (em *ExperimentManager) CreateExperiment(exp *Experiment, createdBy string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if _, exists := em.experiments[exp.ID]; exists {
		return ErrExperimentExists
	}

	now := time.Now()
	exp.CreatedAt = now
	exp.UpdatedAt = now
	exp.CreatedBy = createdBy
	exp.Status = ExperimentStatusDraft

	// Validate experiment
	if err := em.validateExperiment(exp); err != nil {
		return err
	}

	em.experiments[exp.ID] = exp
	em.initializeTracking(exp.ID)

	return nil
}

// StartExperiment starts an experiment
func (em *ExperimentManager) StartExperiment(experimentID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	exp, exists := em.experiments[experimentID]
	if !exists {
		return ErrExperimentNotFound
	}

	if exp.Status != ExperimentStatusDraft && exp.Status != ExperimentStatusPaused {
		return fmt.Errorf("experiment must be in draft or paused status to start")
	}

	now := time.Now()
	exp.Status = ExperimentStatusRunning
	exp.StartedAt = &now
	exp.UpdatedAt = now

	return nil
}

// PauseExperiment pauses a running experiment
func (em *ExperimentManager) PauseExperiment(experimentID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	exp, exists := em.experiments[experimentID]
	if !exists {
		return ErrExperimentNotFound
	}

	if exp.Status != ExperimentStatusRunning {
		return fmt.Errorf("only running experiments can be paused")
	}

	exp.Status = ExperimentStatusPaused
	exp.UpdatedAt = time.Now()

	return nil
}

// StopExperiment stops an experiment
func (em *ExperimentManager) StopExperiment(experimentID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	exp, exists := em.experiments[experimentID]
	if !exists {
		return ErrExperimentNotFound
	}

	now := time.Now()
	exp.Status = ExperimentStatusStopped
	exp.EndedAt = &now
	exp.UpdatedAt = now

	return nil
}

// CompleteExperiment marks an experiment as completed
func (em *ExperimentManager) CompleteExperiment(experimentID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	exp, exists := em.experiments[experimentID]
	if !exists {
		return ErrExperimentNotFound
	}

	now := time.Now()
	exp.Status = ExperimentStatusCompleted
	exp.EndedAt = &now
	exp.UpdatedAt = now

	// Calculate final results
	em.mu.Unlock()
	result := em.CalculateResults(experimentID)
	em.mu.Lock()

	em.results[experimentID] = result

	return nil
}

// TrackImpression tracks when a user sees a variant
func (em *ExperimentManager) TrackImpression(experimentID, variantID, userID string) {
	em.eventsMu.Lock()
	defer em.eventsMu.Unlock()

	// Track impression
	if em.impressions[experimentID] == nil {
		em.impressions[experimentID] = make(map[string]int64)
	}
	em.impressions[experimentID][variantID]++

	// Track unique user
	if em.uniqueUsers[experimentID] == nil {
		em.uniqueUsers[experimentID] = make(map[string]map[string]bool)
	}
	if em.uniqueUsers[experimentID][variantID] == nil {
		em.uniqueUsers[experimentID][variantID] = make(map[string]bool)
	}
	em.uniqueUsers[experimentID][variantID][userID] = true

	// Record event
	event := ExperimentEvent{
		ID:           generateID(),
		ExperimentID: experimentID,
		VariantID:    variantID,
		UserID:       userID,
		EventType:    "impression",
		Timestamp:    time.Now(),
	}
	em.events = append(em.events, event)
}

// TrackConversion tracks a conversion event
func (em *ExperimentManager) TrackConversion(experimentID, variantID, userID, goalID string, value float64) {
	em.eventsMu.Lock()
	defer em.eventsMu.Unlock()

	// Track conversion
	if em.conversions[experimentID] == nil {
		em.conversions[experimentID] = make(map[string]map[string]int64)
	}
	if em.conversions[experimentID][variantID] == nil {
		em.conversions[experimentID][variantID] = make(map[string]int64)
	}
	em.conversions[experimentID][variantID][goalID]++

	// Track revenue
	if value > 0 {
		if em.revenue[experimentID] == nil {
			em.revenue[experimentID] = make(map[string]float64)
		}
		em.revenue[experimentID][variantID] += value
	}

	// Record event
	event := ExperimentEvent{
		ID:           generateID(),
		ExperimentID: experimentID,
		VariantID:    variantID,
		UserID:       userID,
		EventType:    "conversion",
		EventName:    goalID,
		Value:        value,
		Timestamp:    time.Now(),
	}
	em.events = append(em.events, event)
}

// TrackCustomEvent tracks a custom event
func (em *ExperimentManager) TrackCustomEvent(experimentID, variantID, userID, eventName string, value float64, properties map[string]interface{}) {
	em.eventsMu.Lock()
	defer em.eventsMu.Unlock()

	event := ExperimentEvent{
		ID:           generateID(),
		ExperimentID: experimentID,
		VariantID:    variantID,
		UserID:       userID,
		EventType:    "custom",
		EventName:    eventName,
		Value:        value,
		Properties:   properties,
		Timestamp:    time.Now(),
	}
	em.events = append(em.events, event)
}

// CalculateResults calculates experiment results with statistical analysis
func (em *ExperimentManager) CalculateResults(experimentID string) *ExperimentResult {
	em.mu.RLock()
	exp, exists := em.experiments[experimentID]
	em.mu.RUnlock()

	if !exists {
		return nil
	}

	em.eventsMu.RLock()
	defer em.eventsMu.RUnlock()

	result := &ExperimentResult{
		ExperimentID:   experimentID,
		VariantResults: make(map[string]*VariantResult),
		Confidence:     exp.Confidence,
		CalculatedAt:   time.Now(),
	}

	// Calculate duration
	if exp.StartedAt != nil {
		endTime := time.Now()
		if exp.EndedAt != nil {
			endTime = *exp.EndedAt
		}
		result.Duration = endTime.Sub(*exp.StartedAt)
	}

	// Calculate results for each variant
	var controlResult *VariantResult
	totalSampleSize := int64(0)

	for _, variant := range exp.Variants {
		vr := em.calculateVariantResult(experimentID, variant)
		result.VariantResults[variant.ID] = vr
		totalSampleSize += vr.UniqueUsers

		if variant.IsControl {
			controlResult = vr
		}
	}

	// Check if minimum sample size reached
	result.SampleSizeReached = totalSampleSize >= int64(exp.MinSampleSize)

	// Calculate uplift and significance vs control
	if controlResult != nil {
		for variantID, vr := range result.VariantResults {
			if variantID != controlResult.VariantID {
				vr.Uplift = em.calculateUplift(controlResult.ConversionRate, vr.ConversionRate)
				vr.PValue = em.calculatePValue(controlResult, vr)
				vr.IsSignificant = vr.PValue < (1 - exp.Confidence)
			}
		}
	}

	// Determine winner
	result.Winner = em.determineWinner(result, exp)

	return result
}

// Helper methods

func (em *ExperimentManager) validateExperiment(exp *Experiment) error {
	if exp.ID == "" {
		return ErrInvalidExperiment("id is required")
	}
	if exp.Name == "" {
		return ErrInvalidExperiment("name is required")
	}
	if len(exp.Variants) < 2 {
		return ErrInvalidExperiment("at least 2 variants required")
	}

	// Validate variant weights
	totalWeight := 0
	hasControl := false
	for _, v := range exp.Variants {
		totalWeight += v.Weight
		if v.IsControl {
			hasControl = true
			exp.ControlVariant = v.ID
		}
	}

	if totalWeight != 100 {
		return ErrInvalidExperiment("variant weights must sum to 100")
	}

	if !hasControl {
		return ErrInvalidExperiment("one variant must be marked as control")
	}

	// Validate confidence
	if exp.Confidence <= 0 || exp.Confidence >= 1 {
		exp.Confidence = 0.95 // Default to 95%
	}

	// Validate power analysis
	if exp.PowerAnalysis <= 0 || exp.PowerAnalysis >= 1 {
		exp.PowerAnalysis = 0.8 // Default to 80%
	}

	return nil
}

func (em *ExperimentManager) initializeTracking(experimentID string) {
	em.impressions[experimentID] = make(map[string]int64)
	em.conversions[experimentID] = make(map[string]map[string]int64)
	em.uniqueUsers[experimentID] = make(map[string]map[string]bool)
	em.revenue[experimentID] = make(map[string]float64)
}

func (em *ExperimentManager) calculateVariantResult(experimentID string, variant ExperimentVariant) *VariantResult {
	vr := &VariantResult{
		VariantID:     variant.ID,
		VariantName:   variant.Name,
		IsControl:     variant.IsControl,
		CustomMetrics: make(map[string]float64),
	}

	// Get impressions
	if impressions, ok := em.impressions[experimentID]; ok {
		vr.Impressions = impressions[variant.ID]
	}

	// Get unique users
	if users, ok := em.uniqueUsers[experimentID]; ok {
		if variantUsers, ok := users[variant.ID]; ok {
			vr.UniqueUsers = int64(len(variantUsers))
		}
	}

	// Get conversions
	totalConversions := int64(0)
	if conversions, ok := em.conversions[experimentID]; ok {
		if variantConversions, ok := conversions[variant.ID]; ok {
			for _, count := range variantConversions {
				totalConversions += count
			}
		}
	}
	vr.Conversions = totalConversions

	// Calculate conversion rate
	if vr.UniqueUsers > 0 {
		vr.ConversionRate = float64(vr.Conversions) / float64(vr.UniqueUsers)
	}

	// Get revenue
	if revenue, ok := em.revenue[experimentID]; ok {
		vr.Revenue = revenue[variant.ID]
		if vr.UniqueUsers > 0 {
			vr.RevenuePerUser = vr.Revenue / float64(vr.UniqueUsers)
		}
	}

	// Calculate standard error
	vr.StandardError = em.calculateStandardError(vr.ConversionRate, vr.UniqueUsers)

	// Calculate confidence interval
	vr.ConfidenceInterval = em.calculateConfidenceInterval(vr.ConversionRate, vr.StandardError, 0.95)

	return vr
}

func (em *ExperimentManager) calculateStandardError(rate float64, sampleSize int64) float64 {
	if sampleSize == 0 {
		return 0
	}
	// Standard error = sqrt(p * (1 - p) / n)
	return math.Sqrt(rate * (1 - rate) / float64(sampleSize))
}

func (em *ExperimentManager) calculateConfidenceInterval(rate, stdError, confidence float64) [2]float64 {
	// Z-score for 95% confidence = 1.96
	zScore := 1.96
	if confidence == 0.99 {
		zScore = 2.576
	}

	margin := zScore * stdError
	return [2]float64{
		math.Max(0, rate-margin),
		math.Min(1, rate+margin),
	}
}

func (em *ExperimentManager) calculateUplift(controlRate, variantRate float64) float64 {
	if controlRate == 0 {
		return 0
	}
	return ((variantRate - controlRate) / controlRate) * 100
}

func (em *ExperimentManager) calculatePValue(control, variant *VariantResult) float64 {
	// Two-proportion z-test
	p1 := control.ConversionRate
	p2 := variant.ConversionRate
	n1 := float64(control.UniqueUsers)
	n2 := float64(variant.UniqueUsers)

	if n1 == 0 || n2 == 0 {
		return 1.0
	}

	// Pooled proportion
	pPool := float64(control.Conversions + variant.Conversions) / (n1 + n2)

	// Standard error
	se := math.Sqrt(pPool * (1 - pPool) * (1/n1 + 1/n2))

	if se == 0 {
		return 1.0
	}

	// Z-score
	z := (p2 - p1) / se

	// Two-tailed p-value approximation
	pValue := 2 * (1 - em.normalCDF(math.Abs(z)))

	return pValue
}

func (em *ExperimentManager) normalCDF(z float64) float64 {
	// Approximation of standard normal CDF
	return 0.5 * (1 + math.Erf(z/math.Sqrt2))
}

func (em *ExperimentManager) determineWinner(result *ExperimentResult, exp *Experiment) *string {
	var bestVariant string
	bestRate := 0.0

	for variantID, vr := range result.VariantResults {
		if !vr.IsControl && vr.IsSignificant && vr.ConversionRate > bestRate {
			bestRate = vr.ConversionRate
			bestVariant = variantID
		}
	}

	if bestVariant != "" {
		return &bestVariant
	}
	return nil
}

// DeclareWinner manually declares a winner
func (em *ExperimentManager) DeclareWinner(experimentID, variantID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	exp, exists := em.experiments[experimentID]
	if !exists {
		return ErrExperimentNotFound
	}

	now := time.Now()
	exp.WinnerVariantID = &variantID
	exp.WinnerDeclaredAt = &now
	exp.UpdatedAt = now

	return nil
}

// GetExperiment retrieves an experiment
func (em *ExperimentManager) GetExperiment(experimentID string) (*Experiment, bool) {
	em.mu.RLock()
	defer em.mu.RUnlock()
	exp, exists := em.experiments[experimentID]
	return exp, exists
}

// ListExperiments lists all experiments
func (em *ExperimentManager) ListExperiments() []*Experiment {
	em.mu.RLock()
	defer em.mu.RUnlock()

	experiments := make([]*Experiment, 0, len(em.experiments))
	for _, exp := range em.experiments {
		experiments = append(experiments, exp)
	}
	return experiments
}

// GetResults retrieves experiment results
func (em *ExperimentManager) GetResults(experimentID string) (*ExperimentResult, bool) {
	// Calculate fresh results
	result := em.CalculateResults(experimentID)
	return result, result != nil
}
