package cbook

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"sync"
	"time"
)

// MLPredictor provides machine learning-based client profitability predictions
type MLPredictor struct {
	mu sync.RWMutex

	// Feature weights (learned via online learning)
	featureWeights map[string]float64

	// Model configuration
	learningRate   float64
	regularization float64
	minSamples     int

	// Training data
	trainingData   []TrainingSample
	maxSamples     int

	// Model metadata
	lastTrainedAt  time.Time
	modelVersion   int
	trainingCount  int64
}

// TrainingSample represents a labeled training example
type TrainingSample struct {
	Features      map[string]float64 `json:"features"`
	ActualPnL     float64            `json:"actualPnL"`
	ActualWinRate float64            `json:"actualWinRate"`
	Label         int                `json:"label"` // 1 = profitable, 0 = unprofitable
	Weight        float64            `json:"weight"` // Sample importance
	RecordedAt    time.Time          `json:"recordedAt"`
}

// ProfitabilityPrediction contains ML model output
type ProfitabilityPrediction struct {
	AccountID          int64              `json:"accountId"`
	PredictedWinRate   float64            `json:"predictedWinRate"`   // 0-100
	Confidence         float64            `json:"confidence"`         // 0-1
	RiskScore          float64            `json:"riskScore"`          // 0-100 (higher = riskier for broker)
	RecommendedAction  RoutingAction      `json:"recommendedAction"`
	RecommendedHedge   float64            `json:"recommendedHedge"`   // % to A-Book
	Features           map[string]float64 `json:"features"`
	ModelVersion       int                `json:"modelVersion"`
	PredictedAt        time.Time          `json:"predictedAt"`
}

// NewMLPredictor creates a new ML predictor with online learning
func NewMLPredictor() *MLPredictor {
	return &MLPredictor{
		featureWeights: map[string]float64{
			"win_rate":            0.3,
			"sharpe_ratio":        0.2,
			"avg_hold_time":       -0.15,
			"order_fill_ratio":    -0.1,
			"instrument_concentration": -0.05,
			"total_volume":        0.1,
			"toxicity_score":      0.15,
			"max_drawdown":        -0.1,
			"avg_trade_size":      0.05,
			"time_consistency":    0.05,
		},
		learningRate:   0.01,
		regularization: 0.001,
		minSamples:     50,
		maxSamples:     10000,
		trainingData:   make([]TrainingSample, 0, 10000),
		modelVersion:   1,
	}
}

// Predict generates profitability prediction for a client
func (ml *MLPredictor) Predict(profile *ClientProfile) (*ProfitabilityPrediction, error) {
	if profile == nil {
		return nil, errors.New("profile is nil")
	}

	ml.mu.RLock()
	defer ml.mu.RUnlock()

	// Extract features
	features := ml.extractFeatures(profile)

	// Calculate prediction score using learned weights
	score := 0.0
	for featureName, value := range features {
		weight, exists := ml.featureWeights[featureName]
		if exists {
			score += weight * value
		}
	}

	// Apply sigmoid to get probability
	predictedWinRate := sigmoid(score) * 100

	// Calculate confidence based on sample size
	confidence := calculateConfidence(profile.TotalTrades, ml.minSamples)

	// Calculate risk score (inverse of broker advantage)
	riskScore := calculateRiskScore(predictedWinRate, profile.ToxicityScore, confidence)

	// Determine recommended action
	recommendedAction, recommendedHedge := ml.recommendRouting(riskScore, predictedWinRate, profile.ToxicityScore)

	prediction := &ProfitabilityPrediction{
		AccountID:         profile.AccountID,
		PredictedWinRate:  predictedWinRate,
		Confidence:        confidence,
		RiskScore:         riskScore,
		RecommendedAction: recommendedAction,
		RecommendedHedge:  recommendedHedge,
		Features:          features,
		ModelVersion:      ml.modelVersion,
		PredictedAt:       time.Now(),
	}

	return prediction, nil
}

// extractFeatures converts profile to normalized feature vector
func (ml *MLPredictor) extractFeatures(profile *ClientProfile) map[string]float64 {
	features := make(map[string]float64)

	// Normalize win rate (0-100 -> 0-1)
	features["win_rate"] = normalize(profile.WinRate, 0, 100)

	// Sharpe ratio (clip to -3 to 3, then normalize)
	sharpe := clip(profile.SharpeRatio, -3, 3)
	features["sharpe_ratio"] = normalize(sharpe, -3, 3)

	// Average hold time (log scale, seconds)
	if profile.AverageHoldTime > 0 {
		// Normalize log of hold time (1 sec to 1 week)
		logHoldTime := math.Log(float64(profile.AverageHoldTime))
		features["avg_hold_time"] = normalize(logHoldTime, 0, math.Log(604800))
	}

	// Order to fill ratio
	features["order_fill_ratio"] = clip(profile.OrderToFillRatio, 0, 1)

	// Instrument concentration (max concentration)
	maxConc := 0.0
	totalVol := 0.0
	for _, vol := range profile.InstrumentConc {
		totalVol += vol
		if vol > maxConc {
			maxConc = vol
		}
	}
	if totalVol > 0 {
		features["instrument_concentration"] = maxConc / totalVol
	}

	// Total volume (log scale)
	if profile.TotalVolume > 0 {
		logVolume := math.Log(profile.TotalVolume + 1)
		features["total_volume"] = normalize(logVolume, 0, math.Log(10000))
	}

	// Toxicity score
	features["toxicity_score"] = normalize(profile.ToxicityScore, 0, 100)

	// Max drawdown (normalize to 0-1, assuming max 10000)
	features["max_drawdown"] = normalize(profile.MaxDrawdown, 0, 10000)

	// Average trade size (log scale)
	if profile.AverageTradeSize > 0 {
		logSize := math.Log(profile.AverageTradeSize + 1)
		features["avg_trade_size"] = normalize(logSize, 0, math.Log(100))
	}

	// Time consistency (how evenly distributed trades are across hours)
	features["time_consistency"] = calculateTimeConsistency(profile.TimeOfDayPattern)

	return features
}

// Train performs online learning update with new sample
func (ml *MLPredictor) Train(profile *ClientProfile, actualWinRate, actualPnL float64) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	// Create training sample
	features := ml.extractFeatures(profile)

	label := 0
	if actualWinRate > 50 {
		label = 1 // Client is profitable
	}

	sample := TrainingSample{
		Features:      features,
		ActualPnL:     actualPnL,
		ActualWinRate: actualWinRate,
		Label:         label,
		Weight:        1.0,
		RecordedAt:    time.Now(),
	}

	// Add to training data
	ml.trainingData = append(ml.trainingData, sample)
	if len(ml.trainingData) > ml.maxSamples {
		ml.trainingData = ml.trainingData[1:]
	}

	// Perform gradient descent update
	ml.gradientDescentUpdate(sample)

	ml.trainingCount++

	// Retrain full model every 100 samples
	if ml.trainingCount%100 == 0 && len(ml.trainingData) >= ml.minSamples {
		ml.retrainModel()
	}
}

// gradientDescentUpdate performs single SGD step
func (ml *MLPredictor) gradientDescentUpdate(sample TrainingSample) {
	// Calculate prediction with current weights
	score := 0.0
	for featureName, value := range sample.Features {
		weight, exists := ml.featureWeights[featureName]
		if exists {
			score += weight * value
		}
	}

	prediction := sigmoid(score)
	target := float64(sample.Label)

	// Calculate error
	error := prediction - target

	// Update weights (gradient descent with L2 regularization)
	for featureName, value := range sample.Features {
		weight := ml.featureWeights[featureName]

		// Gradient: error * feature_value + regularization
		gradient := error*value + ml.regularization*weight

		// Weight update
		ml.featureWeights[featureName] = weight - ml.learningRate*gradient
	}
}

// retrainModel performs full batch training
func (ml *MLPredictor) retrainModel() {
	if len(ml.trainingData) < ml.minSamples {
		return
	}

	log.Printf("[ML Predictor] Retraining model with %d samples", len(ml.trainingData))

	// Perform multiple epochs of batch gradient descent
	epochs := 10
	batchSize := 32

	for epoch := 0; epoch < epochs; epoch++ {
		// Shuffle training data (simple shuffle)
		for i := len(ml.trainingData) - 1; i > 0; i-- {
			j := i % (i + 1) // Simplified random
			ml.trainingData[i], ml.trainingData[j] = ml.trainingData[j], ml.trainingData[i]
		}

		// Mini-batch gradient descent
		for i := 0; i < len(ml.trainingData); i += batchSize {
			end := i + batchSize
			if end > len(ml.trainingData) {
				end = len(ml.trainingData)
			}

			batch := ml.trainingData[i:end]
			ml.batchUpdate(batch)
		}
	}

	ml.lastTrainedAt = time.Now()
	ml.modelVersion++

	log.Printf("[ML Predictor] Model retrained. Version: %d", ml.modelVersion)
}

// batchUpdate performs gradient descent on a batch
func (ml *MLPredictor) batchUpdate(batch []TrainingSample) {
	gradients := make(map[string]float64)

	// Accumulate gradients
	for _, sample := range batch {
		score := 0.0
		for featureName, value := range sample.Features {
			weight, exists := ml.featureWeights[featureName]
			if exists {
				score += weight * value
			}
		}

		prediction := sigmoid(score)
		target := float64(sample.Label)
		error := prediction - target

		for featureName, value := range sample.Features {
			gradients[featureName] += error * value
		}
	}

	// Average gradients and update weights
	batchSize := float64(len(batch))
	for featureName := range ml.featureWeights {
		gradient := gradients[featureName] / batchSize
		regularizationTerm := ml.regularization * ml.featureWeights[featureName]

		ml.featureWeights[featureName] -= ml.learningRate * (gradient + regularizationTerm)
	}
}

// recommendRouting determines optimal routing based on predictions
func (ml *MLPredictor) recommendRouting(riskScore, predictedWinRate, toxicityScore float64) (RoutingAction, float64) {
	// Very high risk - reject or full A-Book
	if riskScore > 80 || toxicityScore > 80 {
		if toxicityScore > 90 {
			return ActionReject, 0
		}
		return ActionABook, 100
	}

	// High risk - mostly A-Book
	if riskScore > 60 || predictedWinRate > 55 {
		return ActionPartialHedge, 80
	}

	// Medium risk - balanced
	if riskScore > 40 || predictedWinRate > 52 {
		return ActionPartialHedge, 60
	}

	// Low risk (profitable for broker) - mostly B-Book
	if riskScore < 30 && predictedWinRate < 48 {
		return ActionBBook, 10 // 10% hedge for safety
	}

	// Default - moderate B-Book
	return ActionPartialHedge, 40
}

// GetModelStats returns model performance statistics
func (ml *MLPredictor) GetModelStats() map[string]interface{} {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	stats := make(map[string]interface{})

	stats["model_version"] = ml.modelVersion
	stats["last_trained"] = ml.lastTrainedAt
	stats["training_count"] = ml.trainingCount
	stats["training_samples"] = len(ml.trainingData)
	stats["feature_weights"] = ml.featureWeights
	stats["learning_rate"] = ml.learningRate

	// Calculate training accuracy if enough samples
	if len(ml.trainingData) >= ml.minSamples {
		correct := 0
		for _, sample := range ml.trainingData {
			score := 0.0
			for featureName, value := range sample.Features {
				score += ml.featureWeights[featureName] * value
			}
			prediction := 0
			if sigmoid(score) > 0.5 {
				prediction = 1
			}
			if prediction == sample.Label {
				correct++
			}
		}
		stats["training_accuracy"] = float64(correct) / float64(len(ml.trainingData)) * 100
	}

	return stats
}

// ExportModel exports model weights for backup/transfer
func (ml *MLPredictor) ExportModel() ([]byte, error) {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	modelData := map[string]interface{}{
		"version":        ml.modelVersion,
		"weights":        ml.featureWeights,
		"learning_rate":  ml.learningRate,
		"regularization": ml.regularization,
		"trained_at":     ml.lastTrainedAt,
	}

	return json.Marshal(modelData)
}

// ImportModel imports model weights
func (ml *MLPredictor) ImportModel(data []byte) error {
	var modelData map[string]interface{}
	if err := json.Unmarshal(data, &modelData); err != nil {
		return err
	}

	ml.mu.Lock()
	defer ml.mu.Unlock()

	if weights, ok := modelData["weights"].(map[string]interface{}); ok {
		for key, value := range weights {
			if floatVal, ok := value.(float64); ok {
				ml.featureWeights[key] = floatVal
			}
		}
	}

	if version, ok := modelData["version"].(float64); ok {
		ml.modelVersion = int(version)
	}

	log.Printf("[ML Predictor] Model imported. Version: %d", ml.modelVersion)
	return nil
}

// Helper functions

func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func normalize(value, min, max float64) float64 {
	if max == min {
		return 0
	}
	normalized := (value - min) / (max - min)
	return clip(normalized, 0, 1)
}

func clip(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func calculateConfidence(samples int64, minSamples int) float64 {
	if samples < int64(minSamples) {
		return float64(samples) / float64(minSamples)
	}
	// Logarithmic confidence growth
	return math.Min(1.0, 0.7+0.3*math.Log(float64(samples))/math.Log(1000))
}

func calculateRiskScore(predictedWinRate, toxicityScore, confidence float64) float64 {
	// Risk = how likely client is to be profitable (bad for broker)
	baseRisk := predictedWinRate // 0-100

	// Adjust for toxicity
	toxicityWeight := toxicityScore * 0.3

	// Adjust for confidence (low confidence = higher uncertainty risk)
	confidenceAdjustment := (1 - confidence) * 20

	riskScore := baseRisk + toxicityWeight + confidenceAdjustment

	return clip(riskScore, 0, 100)
}

func calculateTimeConsistency(pattern map[int]int64) float64 {
	if len(pattern) == 0 {
		return 0
	}

	// Calculate variance in hourly distribution
	var total int64
	for _, count := range pattern {
		total += count
	}

	if total == 0 {
		return 0
	}

	expectedPerHour := float64(total) / 24.0
	var variance float64

	for hour := 0; hour < 24; hour++ {
		count := float64(pattern[hour])
		diff := count - expectedPerHour
		variance += diff * diff
	}

	variance /= 24.0

	// Lower variance = higher consistency
	// Normalize to 0-1 (assuming max variance of total^2/24)
	maxVariance := float64(total*total) / 24.0
	consistency := 1.0 - math.Min(variance/maxVariance, 1.0)

	return consistency
}
