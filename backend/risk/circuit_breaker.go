package risk

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

// CircuitBreakerManager manages all circuit breakers
type CircuitBreakerManager struct {
	mu              sync.RWMutex
	breakers        map[string]*CircuitBreaker
	engine          *Engine
	volatilityCache map[string]float64 // symbol -> recent volatility
	priceHistory    map[string][]PricePoint
	monitorInterval time.Duration
	stopChan        chan struct{}
}

// PricePoint stores a price with timestamp
type PricePoint struct {
	Price float64
	Time  time.Time
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(engine *Engine) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers:        make(map[string]*CircuitBreaker),
		engine:          engine,
		volatilityCache: make(map[string]float64),
		priceHistory:    make(map[string][]PricePoint),
		monitorInterval: 5 * time.Second,
		stopChan:        make(chan struct{}),
	}
}

// Start begins monitoring for circuit breaker conditions
func (cbm *CircuitBreakerManager) Start() {
	go cbm.monitorLoop()
	log.Println("[CircuitBreaker] Manager started")
}

// Stop stops the circuit breaker manager
func (cbm *CircuitBreakerManager) Stop() {
	close(cbm.stopChan)
	log.Println("[CircuitBreaker] Manager stopped")
}

// monitorLoop continuously monitors conditions
func (cbm *CircuitBreakerManager) monitorLoop() {
	ticker := time.NewTicker(cbm.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cbm.stopChan:
			return
		case <-ticker.C:
			cbm.checkAllBreakers()
			cbm.autoResetBreakers()
		}
	}
}

// AddVolatilityBreaker adds a volatility-based circuit breaker
func (cbm *CircuitBreakerManager) AddVolatilityBreaker(
	symbol string,
	threshold float64,
	autoReset bool,
	resetAfter time.Duration,
) {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	id := fmt.Sprintf("VOL_%s", symbol)
	breaker := &CircuitBreaker{
		ID:         id,
		Type:       "VOLATILITY",
		Symbol:     symbol,
		Status:     CircuitNormal,
		Threshold:  threshold,
		CurrentValue: 0,
		Message:    fmt.Sprintf("Volatility circuit breaker for %s", symbol),
		AutoReset:  autoReset,
		ResetAfter: resetAfter,
	}

	cbm.breakers[id] = breaker
	log.Printf("[CircuitBreaker] Added volatility breaker for %s: threshold %.2f%%", symbol, threshold)
}

// AddDailyLossBreaker adds a daily loss circuit breaker
func (cbm *CircuitBreakerManager) AddDailyLossBreaker(
	accountID int64,
	lossLimit float64,
	autoReset bool,
) {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	id := fmt.Sprintf("LOSS_%d", accountID)
	breaker := &CircuitBreaker{
		ID:         id,
		Type:       "LOSS_LIMIT",
		Status:     CircuitNormal,
		Threshold:  lossLimit,
		CurrentValue: 0,
		Message:    fmt.Sprintf("Daily loss limit for account %d", accountID),
		AutoReset:  autoReset,
		ResetAfter: 24 * time.Hour,
	}

	cbm.breakers[id] = breaker
	log.Printf("[CircuitBreaker] Added daily loss breaker for account %d: limit %.2f", accountID, lossLimit)
}

// AddNewsEventBreaker adds a news event circuit breaker
func (cbm *CircuitBreakerManager) AddNewsEventBreaker(
	symbol string,
	duration time.Duration,
	message string,
) {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	id := fmt.Sprintf("NEWS_%s_%d", symbol, time.Now().Unix())
	now := time.Now()
	resetTime := now.Add(duration)

	breaker := &CircuitBreaker{
		ID:          id,
		Type:        "NEWS_EVENT",
		Symbol:      symbol,
		Status:      CircuitTripped,
		Threshold:   0,
		CurrentValue: 0,
		TriggeredAt: &now,
		ResetAt:     &resetTime,
		Message:     message,
		AutoReset:   true,
		ResetAfter:  duration,
	}

	cbm.breakers[id] = breaker
	log.Printf("[CircuitBreaker] Added news event breaker for %s: %s (duration: %v)",
		symbol, message, duration)

	// Create alert
	alert := &RiskAlert{
		ID:        fmt.Sprintf("ALERT_NEWS_%d", time.Now().Unix()),
		Symbol:    symbol,
		AlertType: "NEWS_EVENT_HALT",
		Severity:  RiskLevelHigh,
		Message:   fmt.Sprintf("Trading halted for %s: %s", symbol, message),
		Data: map[string]interface{}{
			"duration_minutes": duration.Minutes(),
		},
		CreatedAt: time.Now(),
	}
	cbm.engine.StoreAlert(alert)
}

// AddSystemBreaker adds a system-wide emergency breaker
func (cbm *CircuitBreakerManager) AddSystemBreaker(message string) {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	id := "SYSTEM_EMERGENCY"
	now := time.Now()

	breaker := &CircuitBreaker{
		ID:          id,
		Type:        "SYSTEM",
		Status:      CircuitTripped,
		Threshold:   0,
		CurrentValue: 0,
		TriggeredAt: &now,
		Message:     message,
		AutoReset:   false,
	}

	cbm.breakers[id] = breaker
	log.Printf("[CircuitBreaker] SYSTEM EMERGENCY BREAKER ACTIVATED: %s", message)

	// Create critical alert
	alert := &RiskAlert{
		ID:        fmt.Sprintf("ALERT_SYSTEM_%d", time.Now().Unix()),
		AlertType: "SYSTEM_EMERGENCY",
		Severity:  RiskLevelCritical,
		Message:   fmt.Sprintf("SYSTEM EMERGENCY: %s", message),
		CreatedAt: time.Now(),
	}
	cbm.engine.StoreAlert(alert)
}

// checkAllBreakers checks all active breakers
func (cbm *CircuitBreakerManager) checkAllBreakers() {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	for id, breaker := range cbm.breakers {
		if breaker.Status == CircuitTripped {
			continue // Already tripped
		}

		switch breaker.Type {
		case "VOLATILITY":
			cbm.checkVolatilityBreaker(id, breaker)
		case "LOSS_LIMIT":
			cbm.checkDailyLossBreaker(id, breaker)
		case "FAT_FINGER":
			// Checked on order entry
		}
	}
}

// checkVolatilityBreaker checks volatility circuit breaker
func (cbm *CircuitBreakerManager) checkVolatilityBreaker(id string, breaker *CircuitBreaker) {
	// Calculate recent volatility
	volatility := cbm.calculateRecentVolatility(breaker.Symbol)
	breaker.CurrentValue = volatility

	if volatility > breaker.Threshold {
		// Trip the breaker
		now := time.Now()
		breaker.Status = CircuitTripped
		breaker.TriggeredAt = &now

		if breaker.AutoReset {
			resetTime := now.Add(breaker.ResetAfter)
			breaker.ResetAt = &resetTime
		}

		log.Printf("[CircuitBreaker] TRIPPED: %s volatility %.2f%% > %.2f%%",
			breaker.Symbol, volatility, breaker.Threshold)

		// Create alert
		alert := &RiskAlert{
			ID:        fmt.Sprintf("ALERT_VOL_%s_%d", breaker.Symbol, now.Unix()),
			Symbol:    breaker.Symbol,
			AlertType: "VOLATILITY_HALT",
			Severity:  RiskLevelHigh,
			Message:   fmt.Sprintf("Trading halted for %s due to high volatility: %.2f%% > %.2f%%",
				breaker.Symbol, volatility, breaker.Threshold),
			Data: map[string]interface{}{
				"volatility": volatility,
				"threshold":  breaker.Threshold,
			},
			CreatedAt: now,
		}
		cbm.engine.StoreAlert(alert)
	}
}

// checkDailyLossBreaker checks daily loss circuit breaker
func (cbm *CircuitBreakerManager) checkDailyLossBreaker(id string, breaker *CircuitBreaker) {
	// Extract account ID from breaker ID (format: LOSS_{accountID})
	var accountID int64
	fmt.Sscanf(id, "LOSS_%d", &accountID)

	dailyPnL := cbm.engine.GetDailyPnL(accountID)
	loss := math.Abs(dailyPnL)
	breaker.CurrentValue = loss

	if dailyPnL < 0 && loss >= breaker.Threshold {
		// Trip the breaker
		now := time.Now()
		breaker.Status = CircuitTripped
		breaker.TriggeredAt = &now

		if breaker.AutoReset {
			// Reset at next day
			tomorrow := now.Add(24 * time.Hour)
			resetTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
				0, 0, 0, 0, tomorrow.Location())
			breaker.ResetAt = &resetTime
		}

		log.Printf("[CircuitBreaker] TRIPPED: Account %d daily loss %.2f >= %.2f",
			accountID, loss, breaker.Threshold)

		// Create alert
		alert := &RiskAlert{
			ID:        fmt.Sprintf("ALERT_LOSS_%d_%d", accountID, now.Unix()),
			AccountID: accountID,
			AlertType: "DAILY_LOSS_LIMIT",
			Severity:  RiskLevelCritical,
			Message:   fmt.Sprintf("Account %d reached daily loss limit: %.2f", accountID, loss),
			Data: map[string]interface{}{
				"daily_pnl": dailyPnL,
				"limit":     breaker.Threshold,
			},
			CreatedAt: now,
		}
		cbm.engine.StoreAlert(alert)
	}
}

// calculateRecentVolatility calculates volatility from recent price history
func (cbm *CircuitBreakerManager) calculateRecentVolatility(symbol string) float64 {
	// Get price history
	history := cbm.priceHistory[symbol]
	if len(history) < 10 {
		// Need at least 10 data points
		return 0
	}

	// Calculate returns
	returns := make([]float64, 0)
	for i := 1; i < len(history); i++ {
		ret := (history[i].Price - history[i-1].Price) / history[i-1].Price
		returns = append(returns, ret)
	}

	// Calculate standard deviation
	mean := 0.0
	for _, ret := range returns {
		mean += ret
	}
	mean /= float64(len(returns))

	variance := 0.0
	for _, ret := range returns {
		diff := ret - mean
		variance += diff * diff
	}
	variance /= float64(len(returns))

	stdDev := math.Sqrt(variance)

	// Annualize (assuming 5-minute bars, 288 per day, 252 trading days)
	annualizedVol := stdDev * math.Sqrt(288*252) * 100 // Convert to percentage

	return annualizedVol
}

// UpdatePrice updates price history for volatility calculation
func (cbm *CircuitBreakerManager) UpdatePrice(symbol string, price float64) {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	if cbm.priceHistory[symbol] == nil {
		cbm.priceHistory[symbol] = make([]PricePoint, 0)
	}

	point := PricePoint{
		Price: price,
		Time:  time.Now(),
	}
	cbm.priceHistory[symbol] = append(cbm.priceHistory[symbol], point)

	// Keep only recent history (e.g., last 100 points)
	if len(cbm.priceHistory[symbol]) > 100 {
		cbm.priceHistory[symbol] = cbm.priceHistory[symbol][1:]
	}
}

// autoResetBreakers automatically resets breakers that are due
func (cbm *CircuitBreakerManager) autoResetBreakers() {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	now := time.Now()

	for id, breaker := range cbm.breakers {
		if breaker.Status == CircuitTripped && breaker.AutoReset && breaker.ResetAt != nil {
			if now.After(*breaker.ResetAt) {
				breaker.Status = CircuitNormal
				breaker.TriggeredAt = nil
				breaker.ResetAt = nil

				log.Printf("[CircuitBreaker] AUTO-RESET: %s", id)

				// Create alert
				alert := &RiskAlert{
					ID:        fmt.Sprintf("ALERT_RESET_%s_%d", id, now.Unix()),
					Symbol:    breaker.Symbol,
					AlertType: "CIRCUIT_BREAKER_RESET",
					Severity:  RiskLevelLow,
					Message:   fmt.Sprintf("Circuit breaker reset: %s", breaker.Message),
					CreatedAt: now,
				}
				cbm.engine.StoreAlert(alert)
			}
		}
	}
}

// ManualReset manually resets a circuit breaker
func (cbm *CircuitBreakerManager) ManualReset(id string, adminID string) error {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	breaker, exists := cbm.breakers[id]
	if !exists {
		return fmt.Errorf("circuit breaker %s not found", id)
	}

	if breaker.Status != CircuitTripped {
		return fmt.Errorf("circuit breaker %s is not tripped", id)
	}

	breaker.Status = CircuitNormal
	breaker.TriggeredAt = nil
	breaker.ResetAt = nil

	log.Printf("[CircuitBreaker] MANUAL RESET by %s: %s", adminID, id)

	// Create alert
	alert := &RiskAlert{
		ID:        fmt.Sprintf("ALERT_MANUAL_RESET_%s_%d", id, time.Now().Unix()),
		Symbol:    breaker.Symbol,
		AlertType: "MANUAL_CIRCUIT_RESET",
		Severity:  RiskLevelMedium,
		Message:   fmt.Sprintf("Circuit breaker manually reset by %s: %s", adminID, breaker.Message),
		Data: map[string]interface{}{
			"admin_id": adminID,
		},
		CreatedAt: time.Now(),
	}
	cbm.engine.StoreAlert(alert)

	return nil
}

// IsBlocked checks if trading is blocked for a symbol
func (cbm *CircuitBreakerManager) IsBlocked(symbol string) (bool, string) {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	// Check system-wide breaker first
	if breaker, exists := cbm.breakers["SYSTEM_EMERGENCY"]; exists {
		if breaker.Status == CircuitTripped {
			return true, breaker.Message
		}
	}

	// Check symbol-specific breakers
	for _, breaker := range cbm.breakers {
		if breaker.Symbol == symbol && breaker.Status == CircuitTripped {
			return true, breaker.Message
		}
	}

	return false, ""
}

// GetAllBreakers returns all circuit breakers
func (cbm *CircuitBreakerManager) GetAllBreakers() []CircuitBreaker {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	breakers := make([]CircuitBreaker, 0, len(cbm.breakers))
	for _, breaker := range cbm.breakers {
		breakers = append(breakers, *breaker)
	}
	return breakers
}

// GetActiveBreakers returns all tripped circuit breakers
func (cbm *CircuitBreakerManager) GetActiveBreakers() []CircuitBreaker {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	breakers := make([]CircuitBreaker, 0)
	for _, breaker := range cbm.breakers {
		if breaker.Status == CircuitTripped {
			breakers = append(breakers, *breaker)
		}
	}
	return breakers
}

// RemoveBreaker removes a circuit breaker
func (cbm *CircuitBreakerManager) RemoveBreaker(id string) error {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	if _, exists := cbm.breakers[id]; !exists {
		return fmt.Errorf("circuit breaker %s not found", id)
	}

	delete(cbm.breakers, id)
	log.Printf("[CircuitBreaker] Removed breaker: %s", id)
	return nil
}

// GetBreaker retrieves a circuit breaker by symbol or ID
func (cbm *CircuitBreakerManager) GetBreaker(symbolOrID string) *CircuitBreaker {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	// Direct lookup by ID
	if breaker, exists := cbm.breakers[symbolOrID]; exists {
		return breaker
	}

	// Search by symbol in breaker metadata
	for _, breaker := range cbm.breakers {
		// Check if this is a volatility breaker for the symbol
		if breaker.Type == "VOLATILITY" {
			// Extract symbol from breaker ID (format: VOLATILITY_SYMBOL_...)
			if len(breaker.ID) > 11 && breaker.ID[11:11+len(symbolOrID)] == symbolOrID {
				return breaker
			}
		}
	}

	return nil
}
