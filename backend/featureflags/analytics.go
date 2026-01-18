package featureflags

import (
	"math"
	"sync"
	"time"
)

// AnalyticsManager handles experiment analytics and funnel analysis
type AnalyticsManager struct {
	funnels    map[string]*Funnel
	segments   map[string]*Segment
	mu         sync.RWMutex

	// Time series data
	timeSeriesData map[string]map[string][]TimeSeriesPoint // experiment -> variant -> data points
	tsmu           sync.RWMutex
}

// Funnel represents a conversion funnel
type Funnel struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Steps       []FunnelStep  `json:"steps"`
	CreatedAt   time.Time     `json:"created_at"`
}

// FunnelStep represents a step in a funnel
type FunnelStep struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	EventType   string `json:"event_type"`
	Order       int    `json:"order"`
}

// FunnelAnalysis represents funnel analysis results
type FunnelAnalysis struct {
	FunnelID       string                      `json:"funnel_id"`
	ExperimentID   string                      `json:"experiment_id"`
	VariantResults map[string]*VariantFunnelResult `json:"variant_results"`
	CalculatedAt   time.Time                   `json:"calculated_at"`
}

// VariantFunnelResult represents funnel results for a variant
type VariantFunnelResult struct {
	VariantID   string              `json:"variant_id"`
	VariantName string              `json:"variant_name"`
	StepResults []FunnelStepResult  `json:"step_results"`
	OverallConversionRate float64   `json:"overall_conversion_rate"`
}

// FunnelStepResult represents results for a single funnel step
type FunnelStepResult struct {
	StepID         string  `json:"step_id"`
	StepName       string  `json:"step_name"`
	StepOrder      int     `json:"step_order"`
	Users          int64   `json:"users"`
	ConversionRate float64 `json:"conversion_rate"` // % of previous step
	DropoffRate    float64 `json:"dropoff_rate"`
}

// Segment represents a user segment
type Segment struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Rules       []TargetingRule `json:"rules"`
	CreatedAt   time.Time       `json:"created_at"`
}

// SegmentPerformance represents performance metrics for a segment
type SegmentPerformance struct {
	SegmentID      string                    `json:"segment_id"`
	SegmentName    string                    `json:"segment_name"`
	ExperimentID   string                    `json:"experiment_id"`
	VariantResults map[string]*VariantResult `json:"variant_results"`
	CalculatedAt   time.Time                 `json:"calculated_at"`
}

// TimeSeriesPoint represents a data point in a time series
type TimeSeriesPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	Impressions    int64     `json:"impressions"`
	Conversions    int64     `json:"conversions"`
	ConversionRate float64   `json:"conversion_rate"`
	Revenue        float64   `json:"revenue,omitempty"`
}

// SampleSizeCalculator calculates required sample size
type SampleSizeCalculator struct {
	BaselineRate      float64 // Control conversion rate
	MinDetectableEffect float64 // Minimum detectable effect (e.g., 0.05 = 5% relative change)
	Confidence        float64 // Confidence level (e.g., 0.95 = 95%)
	Power             float64 // Statistical power (e.g., 0.8 = 80%)
	NumVariants       int     // Number of variants (including control)
}

// NewAnalyticsManager creates a new analytics manager
func NewAnalyticsManager() *AnalyticsManager {
	return &AnalyticsManager{
		funnels:        make(map[string]*Funnel),
		segments:       make(map[string]*Segment),
		timeSeriesData: make(map[string]map[string][]TimeSeriesPoint),
	}
}

// CreateFunnel creates a new conversion funnel
func (am *AnalyticsManager) CreateFunnel(funnel *Funnel) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.funnels[funnel.ID]; exists {
		return ErrFunnelExists
	}

	funnel.CreatedAt = time.Now()
	am.funnels[funnel.ID] = funnel

	return nil
}

// AnalyzeFunnel analyzes a funnel for an experiment
func (am *AnalyticsManager) AnalyzeFunnel(funnelID, experimentID string, events []ExperimentEvent) *FunnelAnalysis {
	am.mu.RLock()
	funnel, exists := am.funnels[funnelID]
	am.mu.RUnlock()

	if !exists {
		return nil
	}

	analysis := &FunnelAnalysis{
		FunnelID:       funnelID,
		ExperimentID:   experimentID,
		VariantResults: make(map[string]*VariantFunnelResult),
		CalculatedAt:   time.Now(),
	}

	// Group events by variant
	variantEvents := make(map[string][]ExperimentEvent)
	for _, event := range events {
		if event.ExperimentID == experimentID {
			variantEvents[event.VariantID] = append(variantEvents[event.VariantID], event)
		}
	}

	// Analyze each variant
	for variantID, events := range variantEvents {
		result := am.analyzeFunnelForVariant(funnel, events)
		result.VariantID = variantID
		analysis.VariantResults[variantID] = result
	}

	return analysis
}

func (am *AnalyticsManager) analyzeFunnelForVariant(funnel *Funnel, events []ExperimentEvent) *VariantFunnelResult {
	result := &VariantFunnelResult{
		StepResults: make([]FunnelStepResult, len(funnel.Steps)),
	}

	// Track users through the funnel
	userProgress := make(map[string]int) // userID -> furthest step reached

	for _, event := range events {
		for i, step := range funnel.Steps {
			if event.EventType == step.EventType || event.EventName == step.ID {
				if userProgress[event.UserID] < i+1 {
					userProgress[event.UserID] = i + 1
				}
			}
		}
	}

	// Calculate step results
	previousStepUsers := int64(len(userProgress))

	for i, step := range funnel.Steps {
		usersReached := int64(0)
		for _, stepReached := range userProgress {
			if stepReached >= i+1 {
				usersReached++
			}
		}

		conversionRate := 0.0
		if previousStepUsers > 0 {
			conversionRate = float64(usersReached) / float64(previousStepUsers)
		}

		dropoffRate := 1 - conversionRate

		result.StepResults[i] = FunnelStepResult{
			StepID:         step.ID,
			StepName:       step.Name,
			StepOrder:      i + 1,
			Users:          usersReached,
			ConversionRate: conversionRate,
			DropoffRate:    dropoffRate,
		}

		previousStepUsers = usersReached
	}

	// Calculate overall conversion rate
	if len(userProgress) > 0 && len(result.StepResults) > 0 {
		lastStep := result.StepResults[len(result.StepResults)-1]
		result.OverallConversionRate = float64(lastStep.Users) / float64(len(userProgress))
	}

	return result
}

// CreateSegment creates a new user segment
func (am *AnalyticsManager) CreateSegment(segment *Segment) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.segments[segment.ID]; exists {
		return ErrSegmentExists
	}

	segment.CreatedAt = time.Now()
	am.segments[segment.ID] = segment

	return nil
}

// AnalyzeSegmentPerformance analyzes performance for a specific segment
func (am *AnalyticsManager) AnalyzeSegmentPerformance(segmentID, experimentID string, events []ExperimentEvent) *SegmentPerformance {
	am.mu.RLock()
	segment, exists := am.segments[segmentID]
	am.mu.RUnlock()

	if !exists {
		return nil
	}

	performance := &SegmentPerformance{
		SegmentID:      segmentID,
		SegmentName:    segment.Name,
		ExperimentID:   experimentID,
		VariantResults: make(map[string]*VariantResult),
		CalculatedAt:   time.Now(),
	}

	// Filter events for users in this segment
	segmentEvents := make([]ExperimentEvent, 0)
	for _, event := range events {
		// Check if user matches segment rules
		// This would need actual user context - simplified here
		segmentEvents = append(segmentEvents, event)
	}

	// Calculate metrics per variant (simplified)
	// In production, this would use the full experiment manager logic
	// but filtered to only segment users

	return performance
}

// RecordTimeSeriesData records time series data for an experiment
func (am *AnalyticsManager) RecordTimeSeriesData(experimentID, variantID string, point TimeSeriesPoint) {
	am.tsmu.Lock()
	defer am.tsmu.Unlock()

	if am.timeSeriesData[experimentID] == nil {
		am.timeSeriesData[experimentID] = make(map[string][]TimeSeriesPoint)
	}

	am.timeSeriesData[experimentID][variantID] = append(
		am.timeSeriesData[experimentID][variantID],
		point,
	)
}

// GetTimeSeriesData retrieves time series data
func (am *AnalyticsManager) GetTimeSeriesData(experimentID, variantID string, from, to time.Time) []TimeSeriesPoint {
	am.tsmu.RLock()
	defer am.tsmu.RUnlock()

	if am.timeSeriesData[experimentID] == nil {
		return []TimeSeriesPoint{}
	}

	allPoints := am.timeSeriesData[experimentID][variantID]
	filtered := make([]TimeSeriesPoint, 0)

	for _, point := range allPoints {
		if (point.Timestamp.Equal(from) || point.Timestamp.After(from)) &&
			(point.Timestamp.Equal(to) || point.Timestamp.Before(to)) {
			filtered = append(filtered, point)
		}
	}

	return filtered
}

// CalculateSampleSize calculates required sample size for an experiment
func (calc *SampleSizeCalculator) Calculate() int {
	// Convert confidence to z-score
	alpha := 1 - calc.Confidence
	zAlpha := calc.normalInverse(1 - alpha/2)

	// Convert power to z-score
	zBeta := calc.normalInverse(calc.Power)

	// Calculate effect size
	p1 := calc.BaselineRate
	p2 := p1 * (1 + calc.MinDetectableEffect)

	// Pooled proportion
	pPool := (p1 + p2) / 2

	// Sample size per variant
	numerator := (zAlpha*math.Sqrt(2*pPool*(1-pPool)) + zBeta*math.Sqrt(p1*(1-p1)+p2*(1-p2)))
	numerator = numerator * numerator
	denominator := (p2 - p1) * (p2 - p1)

	sampleSizePerVariant := int(math.Ceil(numerator / denominator))

	// Total sample size
	totalSampleSize := sampleSizePerVariant * calc.NumVariants

	return totalSampleSize
}

func (calc *SampleSizeCalculator) normalInverse(p float64) float64 {
	// Approximation of inverse normal CDF
	// For production, use a proper statistical library
	if p <= 0 {
		return math.Inf(-1)
	}
	if p >= 1 {
		return math.Inf(1)
	}

	// Beasley-Springer-Moro algorithm approximation
	a := [4]float64{2.50662823884, -18.61500062529, 41.39119773534, -25.44106049637}
	b := [4]float64{-8.47351093090, 23.08336743743, -21.06224101826, 3.13082909833}
	c := [9]float64{0.3374754822726147, 0.9761690190917186, 0.1607979714918209,
		0.0276438810333863, 0.0038405729373609, 0.0003951896511919,
		0.0000321767881768, 0.0000002888167364, 0.0000003960315187}

	x := p - 0.5
	var r float64

	if math.Abs(x) < 0.42 {
		r = x * x
		r = x * (((a[3]*r+a[2])*r+a[1])*r + a[0]) /
			((((b[3]*r+b[2])*r+b[1])*r+b[0])*r + 1.0)
		return r
	}

	r = p
	if x > 0.0 {
		r = 1.0 - p
	}

	r = math.Log(-math.Log(r))
	r = c[0] + r*(c[1]+r*(c[2]+r*(c[3]+r*(c[4]+r*(c[5]+r*(c[6]+r*(c[7]+r*c[8])))))))

	if x < 0.0 {
		r = -r
	}

	return r
}

// CalculatePowerAnalysis calculates statistical power for given parameters
func CalculatePowerAnalysis(baselineRate, targetRate float64, sampleSize int, confidence float64) float64 {
	// Effect size
	effectSize := targetRate - baselineRate

	// Standard error
	se := math.Sqrt(baselineRate*(1-baselineRate)/float64(sampleSize) +
		targetRate*(1-targetRate)/float64(sampleSize))

	if se == 0 {
		return 0
	}

	// Z-score for confidence level
	_ = 1 - confidence // alpha unused, kept for documentation
	zAlpha := 1.96     // For 95% confidence
	if confidence == 0.99 {
		zAlpha = 2.576
	}

	// Critical value
	criticalValue := zAlpha * se

	// Z-score for power
	zBeta := (effectSize - criticalValue) / se

	// Power = P(Z > -zBeta)
	power := 0.5 * (1 + math.Erf(zBeta/math.Sqrt2))

	return power
}

// CalculateConfidenceInterval calculates confidence interval for a proportion
func CalculateConfidenceInterval(successes, total int64, confidence float64) (float64, float64) {
	if total == 0 {
		return 0, 0
	}

	p := float64(successes) / float64(total)
	n := float64(total)

	// Wilson score interval (more accurate than normal approximation)
	z := 1.96 // 95% confidence
	if confidence == 0.99 {
		z = 2.576
	}

	denominator := 1 + z*z/n
	center := (p + z*z/(2*n)) / denominator
	margin := (z * math.Sqrt(p*(1-p)/n+z*z/(4*n*n))) / denominator

	lower := math.Max(0, center-margin)
	upper := math.Min(1, center+margin)

	return lower, upper
}

// DetectAnomalies detects anomalies in time series data
func DetectAnomalies(data []TimeSeriesPoint, threshold float64) []int {
	if len(data) < 3 {
		return []int{}
	}

	anomalies := make([]int, 0)

	// Calculate mean and standard deviation
	sum := 0.0
	for _, point := range data {
		sum += point.ConversionRate
	}
	mean := sum / float64(len(data))

	variance := 0.0
	for _, point := range data {
		diff := point.ConversionRate - mean
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(data)))

	// Detect anomalies (points outside threshold * stdDev)
	for i, point := range data {
		diff := math.Abs(point.ConversionRate - mean)
		if diff > threshold*stdDev {
			anomalies = append(anomalies, i)
		}
	}

	return anomalies
}

// CalculateLifetimeValue calculates customer lifetime value
func CalculateLifetimeValue(averageRevenue float64, retentionRate float64, discountRate float64) float64 {
	// CLV = (Average Revenue per Customer * Retention Rate) / (1 + Discount Rate - Retention Rate)
	if 1+discountRate-retentionRate == 0 {
		return 0
	}
	return (averageRevenue * retentionRate) / (1 + discountRate - retentionRate)
}

// GetFunnel retrieves a funnel by ID
func (am *AnalyticsManager) GetFunnel(funnelID string) (*Funnel, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()
	funnel, exists := am.funnels[funnelID]
	return funnel, exists
}

// GetSegment retrieves a segment by ID
func (am *AnalyticsManager) GetSegment(segmentID string) (*Segment, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()
	segment, exists := am.segments[segmentID]
	return segment, exists
}

// ListFunnels lists all funnels
func (am *AnalyticsManager) ListFunnels() []*Funnel {
	am.mu.RLock()
	defer am.mu.RUnlock()

	funnels := make([]*Funnel, 0, len(am.funnels))
	for _, funnel := range am.funnels {
		funnels = append(funnels, funnel)
	}
	return funnels
}

// ListSegments lists all segments
func (am *AnalyticsManager) ListSegments() []*Segment {
	am.mu.RLock()
	defer am.mu.RUnlock()

	segments := make([]*Segment, 0, len(am.segments))
	for _, segment := range am.segments {
		segments = append(segments, segment)
	}
	return segments
}
