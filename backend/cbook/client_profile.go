package cbook

import (
	"math"
	"sync"
	"time"
)

// ClientClassification defines client risk categories
type ClientClassification string

const (
	ClassificationRetail      ClientClassification = "RETAIL"        // Likely losers, high B-Book
	ClassificationSemiPro     ClientClassification = "SEMI_PRO"      // Mixed results
	ClassificationProfessional ClientClassification = "PROFESSIONAL" // Likely winners, A-Book
	ClassificationToxic       ClientClassification = "TOXIC"         // Arbitrageurs, reject/A-Book
	ClassificationUnknown     ClientClassification = "UNKNOWN"       // New clients
)

// ClientProfile tracks trading metrics for routing decisions
type ClientProfile struct {
	AccountID   int64                `json:"accountId"`
	UserID      string               `json:"userId"`
	Username    string               `json:"username"`

	// Trading Metrics
	TotalTrades     int64   `json:"totalTrades"`
	WinningTrades   int64   `json:"winningTrades"`
	LosingTrades    int64   `json:"losingTrades"`
	WinRate         float64 `json:"winRate"`         // Percentage
	AverageTradeSize float64 `json:"avgTradeSize"`   // In lots
	TotalVolume     float64 `json:"totalVolume"`     // Cumulative lots
	TotalPnL        float64 `json:"totalPnL"`        // Cumulative P&L

	// Risk Metrics
	SharpeRatio       float64 `json:"sharpeRatio"`
	MaxDrawdown       float64 `json:"maxDrawdown"`
	AverageHoldTime   int64   `json:"avgHoldTime"` // Seconds
	ToxicityScore     float64 `json:"toxicityScore"` // 0-100

	// Pattern Detection
	OrderToFillRatio  float64            `json:"orderToFillRatio"`  // Cancelled/Total
	InstrumentConc    map[string]float64 `json:"instrumentConc"`    // Symbol concentration
	TimeOfDayPattern  map[int]int64      `json:"timeOfDayPattern"`  // Hour -> trade count

	// Classification
	Classification ClientClassification `json:"classification"`
	LastUpdated    time.Time            `json:"lastUpdated"`

	// Internal tracking
	tradeHistory   []TradeRecord
	mu             sync.RWMutex
}

// TradeRecord stores individual trade data for analysis
type TradeRecord struct {
	TradeID     int64     `json:"tradeId"`
	Symbol      string    `json:"symbol"`
	Volume      float64   `json:"volume"`
	OpenPrice   float64   `json:"openPrice"`
	ClosePrice  float64   `json:"closePrice"`
	PnL         float64   `json:"pnl"`
	OpenTime    time.Time `json:"openTime"`
	CloseTime   time.Time `json:"closeTime"`
	HoldTime    int64     `json:"holdTime"` // Seconds
	IsWinner    bool      `json:"isWinner"`
}

// ClientProfileEngine manages client profiling
type ClientProfileEngine struct {
	profiles map[int64]*ClientProfile
	mu       sync.RWMutex

	// Configuration thresholds
	toxicWinRateThreshold      float64 // Default: 55%
	toxicSharpeThreshold       float64 // Default: 2.0
	proWinRateThreshold        float64 // Default: 52%
	minTradesForClassification int64   // Default: 50
}

// NewClientProfileEngine creates a new profiling engine
func NewClientProfileEngine() *ClientProfileEngine {
	return &ClientProfileEngine{
		profiles:                   make(map[int64]*ClientProfile),
		toxicWinRateThreshold:      55.0,
		toxicSharpeThreshold:       2.0,
		proWinRateThreshold:        52.0,
		minTradesForClassification: 50,
	}
}

// GetOrCreateProfile retrieves or creates a client profile
func (cpe *ClientProfileEngine) GetOrCreateProfile(accountID int64, userID, username string) *ClientProfile {
	cpe.mu.RLock()
	profile, exists := cpe.profiles[accountID]
	cpe.mu.RUnlock()

	if exists {
		return profile
	}

	// Create new profile
	cpe.mu.Lock()
	defer cpe.mu.Unlock()

	profile = &ClientProfile{
		AccountID:        accountID,
		UserID:           userID,
		Username:         username,
		Classification:   ClassificationUnknown,
		InstrumentConc:   make(map[string]float64),
		TimeOfDayPattern: make(map[int]int64),
		tradeHistory:     make([]TradeRecord, 0, 1000),
		LastUpdated:      time.Now(),
	}

	cpe.profiles[accountID] = profile
	return profile
}

// RecordTrade adds a trade to client's history and updates metrics
func (cpe *ClientProfileEngine) RecordTrade(accountID int64, trade TradeRecord) {
	profile := cpe.GetOrCreateProfile(accountID, "", "")

	profile.mu.Lock()
	defer profile.mu.Unlock()

	// Add to history (keep last 1000 trades)
	profile.tradeHistory = append(profile.tradeHistory, trade)
	if len(profile.tradeHistory) > 1000 {
		profile.tradeHistory = profile.tradeHistory[1:]
	}

	// Update counters
	profile.TotalTrades++
	profile.TotalVolume += trade.Volume
	profile.TotalPnL += trade.PnL

	if trade.IsWinner {
		profile.WinningTrades++
	} else {
		profile.LosingTrades++
	}

	// Update instrument concentration
	if profile.InstrumentConc == nil {
		profile.InstrumentConc = make(map[string]float64)
	}
	profile.InstrumentConc[trade.Symbol] += trade.Volume

	// Update time-of-day pattern
	hour := trade.OpenTime.Hour()
	if profile.TimeOfDayPattern == nil {
		profile.TimeOfDayPattern = make(map[int]int64)
	}
	profile.TimeOfDayPattern[hour]++

	// Recalculate metrics
	profile.recalculateMetrics()

	// Reclassify if enough data
	if profile.TotalTrades >= cpe.minTradesForClassification {
		profile.Classification = cpe.classifyClient(profile)
	}

	profile.LastUpdated = time.Now()
}

// RecordOrderCancellation tracks cancelled orders for toxicity detection
func (cpe *ClientProfileEngine) RecordOrderCancellation(accountID int64) {
	profile := cpe.GetOrCreateProfile(accountID, "", "")

	profile.mu.Lock()
	defer profile.mu.Unlock()

	// Update order-to-fill ratio
	// This is a simplified version; full implementation would track all orders
	totalOrders := float64(profile.TotalTrades) * (1.0 + profile.OrderToFillRatio)
	profile.OrderToFillRatio = (totalOrders - float64(profile.TotalTrades)) / (totalOrders + 1)

	profile.LastUpdated = time.Now()
}

// recalculateMetrics updates derived metrics (must be called with lock held)
func (profile *ClientProfile) recalculateMetrics() {
	if profile.TotalTrades == 0 {
		return
	}

	// Win rate
	profile.WinRate = (float64(profile.WinningTrades) / float64(profile.TotalTrades)) * 100

	// Average trade size
	profile.AverageTradeSize = profile.TotalVolume / float64(profile.TotalTrades)

	// Average hold time
	var totalHoldTime int64
	for _, trade := range profile.tradeHistory {
		totalHoldTime += trade.HoldTime
	}
	if len(profile.tradeHistory) > 0 {
		profile.AverageHoldTime = totalHoldTime / int64(len(profile.tradeHistory))
	}

	// Sharpe ratio (simplified)
	if len(profile.tradeHistory) > 10 {
		profile.SharpeRatio = calculateSharpeRatio(profile.tradeHistory)
	}

	// Max drawdown
	profile.MaxDrawdown = calculateMaxDrawdown(profile.tradeHistory)

	// Toxicity score (composite metric)
	profile.ToxicityScore = calculateToxicityScore(profile)
}

// classifyClient determines client classification based on metrics
func (cpe *ClientProfileEngine) classifyClient(profile *ClientProfile) ClientClassification {
	// Toxic client detection (highest priority)
	if profile.WinRate > cpe.toxicWinRateThreshold {
		if profile.SharpeRatio > cpe.toxicSharpeThreshold {
			return ClassificationToxic
		}
	}

	// Check for latency arbitrage patterns
	if profile.AverageHoldTime < 60 && profile.WinRate > 60 { // < 1 minute hold, high win rate
		return ClassificationToxic
	}

	// High order cancellation rate
	if profile.OrderToFillRatio > 0.5 { // More than 50% orders cancelled
		return ClassificationToxic
	}

	// Professional trader
	if profile.WinRate > cpe.proWinRateThreshold && profile.TotalTrades > 100 {
		if profile.SharpeRatio > 1.0 && profile.ToxicityScore < 50 {
			return ClassificationProfessional
		}
	}

	// Semi-professional (borderline profitable)
	if profile.WinRate >= 48 && profile.WinRate <= cpe.proWinRateThreshold {
		return ClassificationSemiPro
	}

	// Default to retail (likely losers)
	if profile.WinRate < 48 {
		return ClassificationRetail
	}

	return ClassificationUnknown
}

// CalculateToxicityScore computes a composite toxicity score (0-100)
func calculateToxicityScore(profile *ClientProfile) float64 {
	score := 0.0

	// High win rate (30 points max)
	if profile.WinRate > 55 {
		score += math.Min((profile.WinRate-55)*3, 30)
	}

	// High Sharpe ratio (20 points max)
	if profile.SharpeRatio > 1.5 {
		score += math.Min((profile.SharpeRatio-1.5)*10, 20)
	}

	// High order cancellation rate (20 points max)
	if profile.OrderToFillRatio > 0.2 {
		score += math.Min(profile.OrderToFillRatio*100, 20)
	}

	// Short hold times (15 points max)
	if profile.AverageHoldTime < 300 { // Less than 5 minutes
		score += math.Min((300-float64(profile.AverageHoldTime))/20, 15)
	}

	// Instrument concentration (15 points max)
	maxConcentration := 0.0
	for _, concentration := range profile.InstrumentConc {
		if concentration > maxConcentration {
			maxConcentration = concentration
		}
	}
	totalVolume := 0.0
	for _, vol := range profile.InstrumentConc {
		totalVolume += vol
	}
	if totalVolume > 0 {
		concentrationPct := (maxConcentration / totalVolume) * 100
		if concentrationPct > 60 {
			score += math.Min((concentrationPct-60)*0.5, 15)
		}
	}

	return math.Min(score, 100)
}

// calculateSharpeRatio computes Sharpe ratio from trade history
func calculateSharpeRatio(trades []TradeRecord) float64 {
	if len(trades) < 2 {
		return 0
	}

	// Calculate mean return
	var totalReturn float64
	for _, trade := range trades {
		totalReturn += trade.PnL
	}
	meanReturn := totalReturn / float64(len(trades))

	// Calculate standard deviation
	var variance float64
	for _, trade := range trades {
		diff := trade.PnL - meanReturn
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(trades)))

	if stdDev == 0 {
		return 0
	}

	// Sharpe ratio (assuming risk-free rate = 0 for simplicity)
	return meanReturn / stdDev
}

// calculateMaxDrawdown computes maximum drawdown from trade history
func calculateMaxDrawdown(trades []TradeRecord) float64 {
	if len(trades) == 0 {
		return 0
	}

	peak := 0.0
	maxDrawdown := 0.0
	cumulative := 0.0

	for _, trade := range trades {
		cumulative += trade.PnL
		if cumulative > peak {
			peak = cumulative
		}
		drawdown := peak - cumulative
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}

// GetProfile returns a client's profile
func (cpe *ClientProfileEngine) GetProfile(accountID int64) (*ClientProfile, bool) {
	cpe.mu.RLock()
	defer cpe.mu.RUnlock()

	profile, exists := cpe.profiles[accountID]
	return profile, exists
}

// GetAllProfiles returns all client profiles
func (cpe *ClientProfileEngine) GetAllProfiles() []*ClientProfile {
	cpe.mu.RLock()
	defer cpe.mu.RUnlock()

	profiles := make([]*ClientProfile, 0, len(cpe.profiles))
	for _, profile := range cpe.profiles {
		profiles = append(profiles, profile)
	}
	return profiles
}

// GetProfilesByClassification returns profiles matching a classification
func (cpe *ClientProfileEngine) GetProfilesByClassification(classification ClientClassification) []*ClientProfile {
	cpe.mu.RLock()
	defer cpe.mu.RUnlock()

	var profiles []*ClientProfile
	for _, profile := range cpe.profiles {
		if profile.Classification == classification {
			profiles = append(profiles, profile)
		}
	}
	return profiles
}

// UpdateThresholds allows dynamic configuration of classification thresholds
func (cpe *ClientProfileEngine) UpdateThresholds(toxicWinRate, toxicSharpe, proWinRate float64, minTrades int64) {
	cpe.mu.Lock()
	defer cpe.mu.Unlock()

	if toxicWinRate > 0 {
		cpe.toxicWinRateThreshold = toxicWinRate
	}
	if toxicSharpe > 0 {
		cpe.toxicSharpeThreshold = toxicSharpe
	}
	if proWinRate > 0 {
		cpe.proWinRateThreshold = proWinRate
	}
	if minTrades > 0 {
		cpe.minTradesForClassification = minTrades
	}
}
