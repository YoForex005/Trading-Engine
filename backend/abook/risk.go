package abook

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// RiskManager handles A-Book risk management and limits
type RiskManager struct {
	limits       map[string]*AccountLimits
	exposures    map[string]*AccountExposure
	symbolLimits map[string]*SymbolLimits
	mu           sync.RWMutex
}

// AccountLimits defines risk limits per account
type AccountLimits struct {
	AccountID              string
	MaxPositionSize        float64 // Max lots per position
	MaxTotalExposure       float64 // Max total exposure across all positions
	MaxPositionsPerSymbol  int
	MaxTotalPositions      int
	MaxDailyLoss           float64
	MaxDailyTrades         int
	AllowedSymbols         map[string]bool
	MarginLevel            float64 // Minimum margin level (%)
	EnableKillSwitch       bool
	KillSwitchActivated    bool
	LastUpdate             time.Time
}

// AccountExposure tracks current exposure for an account
type AccountExposure struct {
	AccountID         string
	TotalExposure     float64
	PositionCount     int
	DailyTrades       int
	DailyPnL          float64
	ExposureBySymbol  map[string]float64
	PositionsBySymbol map[string]int
	LastReset         time.Time
}

// SymbolLimits defines global symbol limits
type SymbolLimits struct {
	Symbol             string
	MaxPositionSize    float64
	MaxTotalExposure   float64
	MinSpread          float64
	MaxSpread          float64
	TradingHours       *TradingHours
	CircuitBreaker     *CircuitBreaker
}

// TradingHours defines allowed trading hours
type TradingHours struct {
	Enabled   bool
	StartTime string // HH:MM format
	EndTime   string
	Weekdays  []time.Weekday
}

// CircuitBreaker implements volatility-based trading halts
type CircuitBreaker struct {
	Enabled            bool
	PriceChangePercent float64 // % change to trigger
	TimeWindow         time.Duration
	CooldownPeriod     time.Duration
	LastTrigger        time.Time
	IsTriggered        bool
}

// NewRiskManager creates a new risk manager
func NewRiskManager() *RiskManager {
	return &RiskManager{
		limits:       make(map[string]*AccountLimits),
		exposures:    make(map[string]*AccountExposure),
		symbolLimits: make(map[string]*SymbolLimits),
	}
}

// SetAccountLimits sets risk limits for an account
func (rm *RiskManager) SetAccountLimits(accountID string, limits *AccountLimits) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	limits.AccountID = accountID
	limits.LastUpdate = time.Now()
	rm.limits[accountID] = limits

	// Initialize exposure if not exists
	if _, exists := rm.exposures[accountID]; !exists {
		rm.exposures[accountID] = &AccountExposure{
			AccountID:         accountID,
			ExposureBySymbol:  make(map[string]float64),
			PositionsBySymbol: make(map[string]int),
			LastReset:         time.Now(),
		}
	}
}

// SetSymbolLimits sets global limits for a symbol
func (rm *RiskManager) SetSymbolLimits(symbol string, limits *SymbolLimits) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	limits.Symbol = symbol
	rm.symbolLimits[symbol] = limits
}

// PreTradeCheck performs comprehensive pre-trade validation
func (rm *RiskManager) PreTradeCheck(accountID, symbol, side string, volume, price float64) error {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// 1. Get account limits
	limits, exists := rm.limits[accountID]
	if !exists {
		// Use default limits
		limits = rm.getDefaultLimits(accountID)
	}

	// 2. Check kill switch
	if limits.KillSwitchActivated {
		return errors.New("kill switch activated - trading disabled")
	}

	// 3. Check allowed symbols
	if len(limits.AllowedSymbols) > 0 {
		if !limits.AllowedSymbols[symbol] {
			return fmt.Errorf("symbol %s not allowed for this account", symbol)
		}
	}

	// 4. Check position size
	if volume > limits.MaxPositionSize {
		return fmt.Errorf("position size %.2f exceeds maximum %.2f",
			volume, limits.MaxPositionSize)
	}

	// 5. Get current exposure
	exposure, exists := rm.exposures[accountID]
	if !exists {
		exposure = &AccountExposure{
			AccountID:         accountID,
			ExposureBySymbol:  make(map[string]float64),
			PositionsBySymbol: make(map[string]int),
			LastReset:         time.Now(),
		}
		rm.exposures[accountID] = exposure
	}

	// Reset daily counters if needed
	if time.Since(exposure.LastReset) > 24*time.Hour {
		exposure.DailyTrades = 0
		exposure.DailyPnL = 0
		exposure.LastReset = time.Now()
	}

	// 6. Check daily trade limit
	if exposure.DailyTrades >= limits.MaxDailyTrades {
		return fmt.Errorf("daily trade limit (%d) exceeded", limits.MaxDailyTrades)
	}

	// 7. Check daily loss limit
	if limits.MaxDailyLoss > 0 && exposure.DailyPnL < -limits.MaxDailyLoss {
		return fmt.Errorf("daily loss limit exceeded: %.2f", exposure.DailyPnL)
	}

	// 8. Check total exposure
	newExposure := volume * price
	if exposure.TotalExposure+newExposure > limits.MaxTotalExposure {
		return fmt.Errorf("total exposure would exceed limit: %.2f > %.2f",
			exposure.TotalExposure+newExposure, limits.MaxTotalExposure)
	}

	// 9. Check positions per symbol
	if exposure.PositionsBySymbol[symbol] >= limits.MaxPositionsPerSymbol {
		return fmt.Errorf("max positions per symbol exceeded: %d",
			limits.MaxPositionsPerSymbol)
	}

	// 10. Check total positions
	if exposure.PositionCount >= limits.MaxTotalPositions {
		return fmt.Errorf("max total positions exceeded: %d", limits.MaxTotalPositions)
	}

	// 11. Check symbol-specific limits
	if symbolLimits, exists := rm.symbolLimits[symbol]; exists {
		if err := rm.checkSymbolLimits(symbolLimits, volume); err != nil {
			return err
		}
	}

	return nil
}

// RecordTrade records a trade for exposure tracking
func (rm *RiskManager) RecordTrade(accountID, symbol string, volume, price float64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	exposure, exists := rm.exposures[accountID]
	if !exists {
		exposure = &AccountExposure{
			AccountID:         accountID,
			ExposureBySymbol:  make(map[string]float64),
			PositionsBySymbol: make(map[string]int),
			LastReset:         time.Now(),
		}
		rm.exposures[accountID] = exposure
	}

	// Update exposure
	tradeExposure := volume * price
	exposure.TotalExposure += tradeExposure
	exposure.ExposureBySymbol[symbol] += tradeExposure
	exposure.PositionsBySymbol[symbol]++
	exposure.PositionCount++
	exposure.DailyTrades++
}

// RecordClosedTrade updates exposure when a position is closed
func (rm *RiskManager) RecordClosedTrade(accountID, symbol string, volume, openPrice, closePrice float64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	exposure, exists := rm.exposures[accountID]
	if !exists {
		return
	}

	// Update exposure
	tradeExposure := volume * openPrice
	exposure.TotalExposure -= tradeExposure
	if exposure.TotalExposure < 0 {
		exposure.TotalExposure = 0
	}

	exposure.ExposureBySymbol[symbol] -= tradeExposure
	if exposure.ExposureBySymbol[symbol] < 0 {
		exposure.ExposureBySymbol[symbol] = 0
	}

	exposure.PositionsBySymbol[symbol]--
	if exposure.PositionsBySymbol[symbol] < 0 {
		exposure.PositionsBySymbol[symbol] = 0
	}

	exposure.PositionCount--
	if exposure.PositionCount < 0 {
		exposure.PositionCount = 0
	}

	// Calculate PnL
	pnl := (closePrice - openPrice) * volume
	exposure.DailyPnL += pnl

	// Check if daily loss limit triggered kill switch
	limits, exists := rm.limits[accountID]
	if exists && limits.EnableKillSwitch {
		if limits.MaxDailyLoss > 0 && exposure.DailyPnL < -limits.MaxDailyLoss {
			limits.KillSwitchActivated = true
		}
	}
}

// ActivateKillSwitch manually activates the kill switch for an account
func (rm *RiskManager) ActivateKillSwitch(accountID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	limits, exists := rm.limits[accountID]
	if !exists {
		return errors.New("account limits not found")
	}

	limits.KillSwitchActivated = true
	return nil
}

// DeactivateKillSwitch deactivates the kill switch
func (rm *RiskManager) DeactivateKillSwitch(accountID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	limits, exists := rm.limits[accountID]
	if !exists {
		return errors.New("account limits not found")
	}

	limits.KillSwitchActivated = false
	return nil
}

// GetAccountExposure returns current exposure for an account
func (rm *RiskManager) GetAccountExposure(accountID string) (*AccountExposure, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	exposure, exists := rm.exposures[accountID]
	if !exists {
		return nil, errors.New("account exposure not found")
	}

	// Create a copy
	copy := &AccountExposure{
		AccountID:         exposure.AccountID,
		TotalExposure:     exposure.TotalExposure,
		PositionCount:     exposure.PositionCount,
		DailyTrades:       exposure.DailyTrades,
		DailyPnL:          exposure.DailyPnL,
		ExposureBySymbol:  make(map[string]float64),
		PositionsBySymbol: make(map[string]int),
		LastReset:         exposure.LastReset,
	}

	for k, v := range exposure.ExposureBySymbol {
		copy.ExposureBySymbol[k] = v
	}
	for k, v := range exposure.PositionsBySymbol {
		copy.PositionsBySymbol[k] = v
	}

	return copy, nil
}

// CheckVolatilityCircuitBreaker checks if circuit breaker should be triggered
func (rm *RiskManager) CheckVolatilityCircuitBreaker(symbol string, priceChangePercent float64) (bool, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	limits, exists := rm.symbolLimits[symbol]
	if !exists || limits.CircuitBreaker == nil || !limits.CircuitBreaker.Enabled {
		return false, nil
	}

	breaker := limits.CircuitBreaker

	// Check if already triggered and in cooldown
	if breaker.IsTriggered {
		if time.Since(breaker.LastTrigger) < breaker.CooldownPeriod {
			return true, errors.New("circuit breaker active - trading halted")
		}
		// Cooldown expired, reset
		breaker.IsTriggered = false
	}

	// Check if threshold exceeded
	if abs(priceChangePercent) >= breaker.PriceChangePercent {
		breaker.IsTriggered = true
		breaker.LastTrigger = time.Now()
		return true, fmt.Errorf("circuit breaker triggered: %.2f%% price change",
			priceChangePercent)
	}

	return false, nil
}

// checkSymbolLimits validates symbol-specific limits
func (rm *RiskManager) checkSymbolLimits(limits *SymbolLimits, volume float64) error {
	// Check position size
	if volume > limits.MaxPositionSize {
		return fmt.Errorf("position size %.2f exceeds symbol limit %.2f",
			volume, limits.MaxPositionSize)
	}

	// Check trading hours
	if limits.TradingHours != nil && limits.TradingHours.Enabled {
		if err := rm.checkTradingHours(limits.TradingHours); err != nil {
			return err
		}
	}

	return nil
}

// checkTradingHours validates current time against allowed trading hours
func (rm *RiskManager) checkTradingHours(hours *TradingHours) error {
	now := time.Now()

	// Check weekday
	dayAllowed := false
	for _, day := range hours.Weekdays {
		if now.Weekday() == day {
			dayAllowed = true
			break
		}
	}

	if !dayAllowed {
		return fmt.Errorf("trading not allowed on %s", now.Weekday())
	}

	// Check time
	currentTime := now.Format("15:04")
	if currentTime < hours.StartTime || currentTime > hours.EndTime {
		return fmt.Errorf("trading hours: %s - %s (current: %s)",
			hours.StartTime, hours.EndTime, currentTime)
	}

	return nil
}

// getDefaultLimits returns default risk limits
func (rm *RiskManager) getDefaultLimits(accountID string) *AccountLimits {
	return &AccountLimits{
		AccountID:             accountID,
		MaxPositionSize:       100.0,
		MaxTotalExposure:      1000000.0,
		MaxPositionsPerSymbol: 10,
		MaxTotalPositions:     50,
		MaxDailyLoss:          10000.0,
		MaxDailyTrades:        100,
		AllowedSymbols:        make(map[string]bool),
		MarginLevel:           100.0,
		EnableKillSwitch:      true,
		KillSwitchActivated:   false,
		LastUpdate:            time.Now(),
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
