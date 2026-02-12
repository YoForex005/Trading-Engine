package payments

import (
	"sync"
	"time"
)

// FraudChecker performs fraud detection checks
type FraudChecker struct {
	mu               sync.RWMutex
	userTransactions map[string][]time.Time // userID -> timestamps
	userAmounts      map[string][]float64   // userID -> amounts
}

// NewFraudChecker creates a new fraud checker
func NewFraudChecker() *FraudChecker {
	return &FraudChecker{
		userTransactions: make(map[string][]time.Time),
		userAmounts:      make(map[string][]float64),
	}
}

// CalculateRiskScore calculates overall risk score (0-100)
func (f *FraudChecker) CalculateRiskScore(userID string, amount float64) int {
	score := 0

	// Velocity check: 3/hour = +20, 10/day = +30
	velocityScore := f.CheckVelocity(userID)
	score += velocityScore

	// Amount anomaly: >5x average = +30
	amountScore := f.CheckAmount(userID, amount)
	score += amountScore

	// Additional checks could be added:
	// - Geolocation score (+20 if suspicious country)
	// - Device fingerprint (+15 if new device)
	// - Time of day (+10 if unusual hours)
	// - IP reputation (+25 if known VPN/proxy)

	// Record this transaction
	f.recordTransaction(userID, amount)

	// Clean up old records
	f.cleanup(userID)

	if score > 100 {
		score = 100
	}

	return score
}

// CheckVelocity checks transaction velocity limits
// Returns score: 0-50
func (f *FraudChecker) CheckVelocity(userID string) int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	timestamps, exists := f.userTransactions[userID]
	if !exists {
		return 0 // No history = no risk
	}

	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	oneDayAgo := now.Add(-24 * time.Hour)

	countHour := 0
	countDay := 0

	for _, ts := range timestamps {
		if ts.After(oneHourAgo) {
			countHour++
		}
		if ts.After(oneDayAgo) {
			countDay++
		}
	}

	score := 0

	// 3+ transactions in last hour = +20
	if countHour >= 3 {
		score += 20
	}

	// 10+ transactions in last day = +30
	if countDay >= 10 {
		score += 30
	}

	return score
}

// CheckAmount checks if amount is anomalous
// Returns score: 0-30
func (f *FraudChecker) CheckAmount(userID string, amount float64) int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	amounts, exists := f.userAmounts[userID]
	if !exists || len(amounts) == 0 {
		// No history - first transaction
		// Large first deposit is suspicious
		if amount > 10000 {
			return 20
		}
		return 0
	}

	// Calculate average
	var sum float64
	for _, a := range amounts {
		sum += a
	}
	average := sum / float64(len(amounts))

	// >5x average = high risk
	if amount > average*5 {
		return 30
	}

	// >3x average = medium risk
	if amount > average*3 {
		return 15
	}

	return 0
}

// recordTransaction records a transaction for velocity/amount tracking
func (f *FraudChecker) recordTransaction(userID string, amount float64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()

	// Record timestamp
	if _, exists := f.userTransactions[userID]; !exists {
		f.userTransactions[userID] = []time.Time{}
	}
	f.userTransactions[userID] = append(f.userTransactions[userID], now)

	// Record amount
	if _, exists := f.userAmounts[userID]; !exists {
		f.userAmounts[userID] = []float64{}
	}
	f.userAmounts[userID] = append(f.userAmounts[userID], amount)
}

// cleanup removes old transaction records (>7 days)
func (f *FraudChecker) cleanup(userID string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)

	// Clean timestamps
	if timestamps, exists := f.userTransactions[userID]; exists {
		filtered := []time.Time{}
		for _, ts := range timestamps {
			if ts.After(sevenDaysAgo) {
				filtered = append(filtered, ts)
			}
		}
		f.userTransactions[userID] = filtered
	}

	// Note: We keep all amounts for average calculation
	// In production, you might want to limit this or use a rolling window
}

// GetUserStats returns fraud statistics for a user
func (f *FraudChecker) GetUserStats(userID string) map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	stats := map[string]interface{}{
		"total_transactions": 0,
		"transactions_24h":   0,
		"transactions_1h":    0,
		"average_amount":     0.0,
		"total_amount":       0.0,
	}

	// Transaction count
	if timestamps, exists := f.userTransactions[userID]; exists {
		stats["total_transactions"] = len(timestamps)

		now := time.Now()
		oneHourAgo := now.Add(-1 * time.Hour)
		oneDayAgo := now.Add(-24 * time.Hour)

		count1h := 0
		count24h := 0
		for _, ts := range timestamps {
			if ts.After(oneHourAgo) {
				count1h++
			}
			if ts.After(oneDayAgo) {
				count24h++
			}
		}

		stats["transactions_1h"] = count1h
		stats["transactions_24h"] = count24h
	}

	// Amount stats
	if amounts, exists := f.userAmounts[userID]; exists {
		var sum float64
		for _, a := range amounts {
			sum += a
		}
		stats["total_amount"] = sum
		if len(amounts) > 0 {
			stats["average_amount"] = sum / float64(len(amounts))
		}
	}

	return stats
}
