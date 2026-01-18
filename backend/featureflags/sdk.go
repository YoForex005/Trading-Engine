package featureflags

import (
	"sync"
	"time"
)

// SDK provides a unified interface for feature flags and experiments
type SDK struct {
	flagManager       *FlagManager
	experimentManager *ExperimentManager
	analyticsManager  *AnalyticsManager
	rolloutManager    *RolloutManager

	// Cache
	cache       sync.Map // user evaluations cache
	cacheTTL    time.Duration

	// Callbacks
	onFlagEval  []func(*FlagEvaluation)
	onExperimentAssignment []func(string, string, string) // experimentID, variantID, userID
}

// NewSDK creates a new feature flags SDK
func NewSDK() *SDK {
	return &SDK{
		flagManager:       NewFlagManager(),
		experimentManager: NewExperimentManager(),
		analyticsManager:  NewAnalyticsManager(),
		rolloutManager:    NewRolloutManager(true),
		cacheTTL:          5 * time.Second,
		onFlagEval:        make([]func(*FlagEvaluation), 0),
		onExperimentAssignment: make([]func(string, string, string), 0),
	}
}

// Feature Flags API

// IsEnabled checks if a feature flag is enabled for a user
func (sdk *SDK) IsEnabled(flagID, userID string, ctx *EvaluationContext) bool {
	if ctx == nil {
		ctx = &EvaluationContext{
			UserID:    userID,
			Timestamp: time.Now(),
		}
	}
	ctx.UserID = userID

	eval := sdk.flagManager.Evaluate(flagID, ctx)

	// Notify listeners
	for _, callback := range sdk.onFlagEval {
		go callback(eval)
	}

	return eval.Enabled
}

// GetVariant gets the assigned variant for a multivariate flag
func (sdk *SDK) GetVariant(flagID, userID string, ctx *EvaluationContext) *Variant {
	if ctx == nil {
		ctx = &EvaluationContext{
			UserID:    userID,
			Timestamp: time.Now(),
		}
	}
	ctx.UserID = userID

	return sdk.flagManager.GetVariant(flagID, ctx)
}

// GetVariantPayload gets the payload for an assigned variant
func (sdk *SDK) GetVariantPayload(flagID, userID string, ctx *EvaluationContext) map[string]interface{} {
	variant := sdk.GetVariant(flagID, userID, ctx)
	if variant == nil {
		return nil
	}
	return variant.Payload
}

// Experiments API

// GetExperimentVariant gets the assigned experiment variant for a user
func (sdk *SDK) GetExperimentVariant(experimentID, userID string, ctx *EvaluationContext) *ExperimentVariant {
	exp, exists := sdk.experimentManager.GetExperiment(experimentID)
	if !exists || exp.Status != ExperimentStatusRunning {
		return nil
	}

	// Check if user matches targeting rules
	if len(exp.Rules) > 0 {
		if ctx == nil {
			ctx = &EvaluationContext{UserID: userID, Timestamp: time.Now()}
		}
		matched := true
		for _, rule := range exp.Rules {
			if !rule.Evaluate(ctx) {
				matched = false
				break
			}
		}
		if !matched {
			return nil
		}
	}

	// Check traffic allocation
	if exp.TrafficAllocation < 100 {
		hash := hashString(userID + experimentID)
		value := hash % 100
		if int(value) >= exp.TrafficAllocation {
			return nil
		}
	}

	// Assign variant using consistent hashing
	hash := hashString(userID + experimentID)
	value := hash % 100

	cumulative := 0
	for i := range exp.Variants {
		cumulative += exp.Variants[i].Weight
		if int(value) < cumulative {
			variant := &exp.Variants[i]

			// Track impression
			sdk.experimentManager.TrackImpression(experimentID, variant.ID, userID)

			// Notify listeners
			for _, callback := range sdk.onExperimentAssignment {
				go callback(experimentID, variant.ID, userID)
			}

			return variant
		}
	}

	// Fallback to first variant
	if len(exp.Variants) > 0 {
		return &exp.Variants[0]
	}
	return nil
}

// TrackConversion tracks a conversion event for an experiment
func (sdk *SDK) TrackConversion(experimentID, userID, goalID string, value float64) {
	// Get the user's assigned variant
	variant := sdk.GetExperimentVariant(experimentID, userID, nil)
	if variant == nil {
		return
	}

	sdk.experimentManager.TrackConversion(experimentID, variant.ID, userID, goalID, value)
}

// TrackEvent tracks a custom event for an experiment
func (sdk *SDK) TrackEvent(experimentID, userID, eventName string, value float64, properties map[string]interface{}) {
	variant := sdk.GetExperimentVariant(experimentID, userID, nil)
	if variant == nil {
		return
	}

	sdk.experimentManager.TrackCustomEvent(experimentID, variant.ID, userID, eventName, value, properties)
}

// Rollout API

// IsInRollout checks if a user is included in a gradual rollout
func (sdk *SDK) IsInRollout(rolloutID, userID string) bool {
	rollout, exists := sdk.rolloutManager.GetRollout(rolloutID)
	if !exists || rollout.Status != RolloutStatusInProgress {
		return false
	}

	return sdk.rolloutManager.IsUserInRollout(rolloutID, userID)
}

// GetRolloutPercentage gets the current rollout percentage
func (sdk *SDK) GetRolloutPercentage(rolloutID string) int {
	return sdk.rolloutManager.GetCurrentPercentage(rolloutID)
}

// Management API (for admin interfaces)

// CreateFlag creates a new feature flag
func (sdk *SDK) CreateFlag(flag *Flag, createdBy string) error {
	return sdk.flagManager.CreateFlag(flag, createdBy)
}

// UpdateFlag updates a feature flag
func (sdk *SDK) UpdateFlag(flagID string, updates *Flag, updatedBy string) error {
	return sdk.flagManager.UpdateFlag(flagID, updates, updatedBy)
}

// ToggleFlag toggles a flag on/off
func (sdk *SDK) ToggleFlag(flagID string, enabled bool, toggledBy string) error {
	return sdk.flagManager.ToggleFlag(flagID, enabled, toggledBy)
}

// CreateExperiment creates a new A/B test experiment
func (sdk *SDK) CreateExperiment(exp *Experiment, createdBy string) error {
	return sdk.experimentManager.CreateExperiment(exp, createdBy)
}

// StartExperiment starts an experiment
func (sdk *SDK) StartExperiment(experimentID string) error {
	return sdk.experimentManager.StartExperiment(experimentID)
}

// GetExperimentResults gets the results of an experiment
func (sdk *SDK) GetExperimentResults(experimentID string) (*ExperimentResult, bool) {
	return sdk.experimentManager.GetResults(experimentID)
}

// CreateRollout creates a gradual rollout
func (sdk *SDK) CreateRollout(rollout *Rollout) error {
	return sdk.rolloutManager.CreateRollout(rollout)
}

// StartRollout starts a gradual rollout
func (sdk *SDK) StartRollout(rolloutID string) error {
	return sdk.rolloutManager.StartRollout(rolloutID)
}

// Callbacks

// OnFlagEvaluation registers a callback for flag evaluations
func (sdk *SDK) OnFlagEvaluation(callback func(*FlagEvaluation)) {
	sdk.onFlagEval = append(sdk.onFlagEval, callback)
}

// OnExperimentAssignment registers a callback for experiment assignments
func (sdk *SDK) OnExperimentAssignment(callback func(experimentID, variantID, userID string)) {
	sdk.onExperimentAssignment = append(sdk.onExperimentAssignment, callback)
}

// Managers (for advanced usage)

// Flags returns the flag manager
func (sdk *SDK) Flags() *FlagManager {
	return sdk.flagManager
}

// Experiments returns the experiment manager
func (sdk *SDK) Experiments() *ExperimentManager {
	return sdk.experimentManager
}

// Analytics returns the analytics manager
func (sdk *SDK) Analytics() *AnalyticsManager {
	return sdk.analyticsManager
}

// Rollouts returns the rollout manager
func (sdk *SDK) Rollouts() *RolloutManager {
	return sdk.rolloutManager
}

// Helper builders for common use cases

// FlagBuilder provides a fluent API for creating flags
type FlagBuilder struct {
	flag *Flag
}

// NewFlag creates a new flag builder
func NewFlag(id, name string) *FlagBuilder {
	return &FlagBuilder{
		flag: &Flag{
			ID:           id,
			Name:         name,
			Type:         FlagTypeBoolean,
			Enabled:      true,
			DefaultValue: false,
			Rules:        make([]TargetingRule, 0),
			Dependencies: make([]FlagDependency, 0),
			Tags:         make([]string, 0),
		},
	}
}

// Description sets the flag description
func (fb *FlagBuilder) Description(desc string) *FlagBuilder {
	fb.flag.Description = desc
	return fb
}

// Boolean sets the flag as a boolean flag
func (fb *FlagBuilder) Boolean(defaultValue bool) *FlagBuilder {
	fb.flag.Type = FlagTypeBoolean
	fb.flag.DefaultValue = defaultValue
	return fb
}

// Percentage sets the flag as a percentage rollout
func (fb *FlagBuilder) Percentage(pct int) *FlagBuilder {
	fb.flag.Type = FlagTypePercentage
	fb.flag.Percentage = pct
	return fb
}

// Multivariate sets the flag as multivariate with variants
func (fb *FlagBuilder) Multivariate(variants ...Variant) *FlagBuilder {
	fb.flag.Type = FlagTypeMultivariate
	fb.flag.Variants = variants
	return fb
}

// Rules adds targeting rules
func (fb *FlagBuilder) Rules(rules ...TargetingRule) *FlagBuilder {
	fb.flag.Rules = append(fb.flag.Rules, rules...)
	return fb
}

// Tags adds tags
func (fb *FlagBuilder) Tags(tags ...string) *FlagBuilder {
	fb.flag.Tags = append(fb.flag.Tags, tags...)
	return fb
}

// Environment sets the environment
func (fb *FlagBuilder) Environment(env string) *FlagBuilder {
	fb.flag.Environment = env
	return fb
}

// EnableAt schedules the flag to enable at a specific time
func (fb *FlagBuilder) EnableAt(t time.Time) *FlagBuilder {
	fb.flag.EnableAt = &t
	return fb
}

// DisableAt schedules the flag to disable at a specific time
func (fb *FlagBuilder) DisableAt(t time.Time) *FlagBuilder {
	fb.flag.DisableAt = &t
	return fb
}

// Build returns the flag
func (fb *FlagBuilder) Build() *Flag {
	return fb.flag
}

// ExperimentBuilder provides a fluent API for creating experiments
type ExperimentBuilder struct {
	experiment *Experiment
}

// NewExperiment creates a new experiment builder
func NewExperiment(id, name string) *ExperimentBuilder {
	return &ExperimentBuilder{
		experiment: &Experiment{
			ID:               id,
			Name:             name,
			TrafficAllocation: 100,
			MinSampleSize:    1000,
			MinDuration:      24 * time.Hour,
			MaxDuration:      30 * 24 * time.Hour,
			Confidence:       0.95,
			PowerAnalysis:    0.8,
			Variants:         make([]ExperimentVariant, 0),
			Rules:            make([]TargetingRule, 0),
			Tags:             make([]string, 0),
		},
	}
}

// Description sets the experiment description
func (eb *ExperimentBuilder) Description(desc string) *ExperimentBuilder {
	eb.experiment.Description = desc
	return eb
}

// FlagID sets the associated flag ID
func (eb *ExperimentBuilder) FlagID(flagID string) *ExperimentBuilder {
	eb.experiment.FlagID = flagID
	return eb
}

// Control adds the control variant
func (eb *ExperimentBuilder) Control(id, name string, weight int) *ExperimentBuilder {
	eb.experiment.Variants = append(eb.experiment.Variants, ExperimentVariant{
		ID:        id,
		Name:      name,
		IsControl: true,
		Weight:    weight,
	})
	eb.experiment.ControlVariant = id
	return eb
}

// Variant adds a test variant
func (eb *ExperimentBuilder) Variant(id, name string, weight int, config map[string]interface{}) *ExperimentBuilder {
	eb.experiment.Variants = append(eb.experiment.Variants, ExperimentVariant{
		ID:        id,
		Name:      name,
		IsControl: false,
		Weight:    weight,
		Config:    config,
	})
	return eb
}

// Traffic sets the traffic allocation percentage
func (eb *ExperimentBuilder) Traffic(pct int) *ExperimentBuilder {
	eb.experiment.TrafficAllocation = pct
	return eb
}

// SampleSize sets the minimum sample size
func (eb *ExperimentBuilder) SampleSize(size int) *ExperimentBuilder {
	eb.experiment.MinSampleSize = size
	return eb
}

// Duration sets the experiment duration
func (eb *ExperimentBuilder) Duration(min, max time.Duration) *ExperimentBuilder {
	eb.experiment.MinDuration = min
	eb.experiment.MaxDuration = max
	return eb
}

// Confidence sets the confidence level
func (eb *ExperimentBuilder) Confidence(level float64) *ExperimentBuilder {
	eb.experiment.Confidence = level
	return eb
}

// Rules adds targeting rules
func (eb *ExperimentBuilder) Rules(rules ...TargetingRule) *ExperimentBuilder {
	eb.experiment.Rules = append(eb.experiment.Rules, rules...)
	return eb
}

// Tags adds tags
func (eb *ExperimentBuilder) Tags(tags ...string) *ExperimentBuilder {
	eb.experiment.Tags = append(eb.experiment.Tags, tags...)
	return eb
}

// Build returns the experiment
func (eb *ExperimentBuilder) Build() *Experiment {
	return eb.experiment
}
