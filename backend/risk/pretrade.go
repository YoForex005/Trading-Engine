package risk

import (
	"errors"
	"fmt"
	"log"
	"math"
)

// PreTradeValidator handles pre-trade risk checks
type PreTradeValidator struct {
	engine *Engine
}

// NewPreTradeValidator creates a new pre-trade validator
func NewPreTradeValidator(engine *Engine) *PreTradeValidator {
	return &PreTradeValidator{
		engine: engine,
	}
}

// ValidateOrder performs comprehensive pre-trade checks
func (v *PreTradeValidator) ValidateOrder(
	accountID int64,
	symbol string,
	side string,
	volume float64,
	price float64,
	orderType string,
) (*PreTradeCheckResult, error) {

	result := &PreTradeCheckResult{
		Allowed:  true,
		Checks:   make(map[string]bool),
		Warnings: make([]string, 0),
		Metadata: make(map[string]interface{}),
	}

	// Get account
	account, err := v.engine.GetAccountByID(accountID)
	if err != nil {
		result.Allowed = false
		result.Reason = "Account not found"
		return result, err
	}

	// Get instrument risk params
	instrParams := v.engine.GetInstrumentRiskParams(symbol)
	if instrParams == nil {
		result.Allowed = false
		result.Reason = "Instrument not configured"
		return result, errors.New("instrument not configured")
	}

	// Get client risk profile
	clientProfile := v.engine.GetClientRiskProfile(account.ClientID)
	if clientProfile == nil {
		result.Allowed = false
		result.Reason = "Client risk profile not found"
		return result, errors.New("client risk profile not found")
	}

	// Check 1: Instrument trading allowed
	if !instrParams.AllowNewPositions {
		result.Checks["instrument_allowed"] = false
		result.Allowed = false
		result.Reason = "New positions not allowed for this instrument"
		return result, nil
	}
	result.Checks["instrument_allowed"] = true

	// Check 2: Calculate required margin
	contractSize := 100000.0 // Standard lot
	if symbol == "XAUUSD" {
		contractSize = 100.0 // Gold is 100 oz
	} else if symbol == "XAGUSD" {
		contractSize = 5000.0 // Silver is 5000 oz
	}

	notionalValue := volume * contractSize * price

	// Use instrument-specific margin requirement or client leverage
	marginPercent := instrParams.MarginRequirement
	if marginPercent == 0 {
		marginPercent = 100.0 / float64(clientProfile.MaxLeverage)
	}

	requiredMargin := notionalValue * (marginPercent / 100.0)
	result.RequiredMargin = requiredMargin
	result.Metadata["notional_value"] = notionalValue
	result.Metadata["margin_percent"] = marginPercent

	// Check 3: Margin availability
	result.FreeMargin = account.FreeMargin
	if requiredMargin > account.FreeMargin {
		result.Checks["margin_available"] = false
		result.Allowed = false
		result.Reason = fmt.Sprintf("Insufficient margin: required %.2f, available %.2f",
			requiredMargin, account.FreeMargin)
		return result, nil
	}
	result.Checks["margin_available"] = true

	// Calculate margin level after trade
	newMargin := account.Margin + requiredMargin
	newMarginLevel := 0.0
	if newMargin > 0 {
		newMarginLevel = (account.Equity / newMargin) * 100
	}
	result.MarginLevel = newMarginLevel
	result.Metadata["margin_level_after"] = newMarginLevel

	// Check 4: Margin level after trade
	if newMarginLevel < clientProfile.StopOutLevel {
		result.Checks["margin_level_ok"] = false
		result.Allowed = false
		result.Reason = fmt.Sprintf("Trade would trigger stop-out (margin level %.2f%% < %.2f%%)",
			newMarginLevel, clientProfile.StopOutLevel)
		return result, nil
	}
	result.Checks["margin_level_ok"] = true

	// Warning if close to margin call
	if newMarginLevel < clientProfile.MarginCallLevel {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Trade will bring margin level to %.2f%%, below margin call level %.2f%%",
				newMarginLevel, clientProfile.MarginCallLevel))
	}

	// Check 5: Position size limits
	if volume > instrParams.MaxPositionSize {
		result.Checks["position_size_ok"] = false
		result.Allowed = false
		result.Reason = fmt.Sprintf("Position size %.2f lots exceeds limit %.2f lots",
			volume, instrParams.MaxPositionSize)
		return result, nil
	}
	result.Checks["position_size_ok"] = true

	// Check per-instrument limit from client profile
	if limit, ok := clientProfile.InstrumentLimits[symbol]; ok {
		currentExposure := v.engine.GetSymbolExposure(symbol)
		newExposure := currentExposure + (notionalValue * v.getDirectionMultiplier(side))
		if math.Abs(newExposure) > limit {
			result.Checks["instrument_limit_ok"] = false
			result.Allowed = false
			result.Reason = fmt.Sprintf("Instrument exposure limit exceeded: %.2f > %.2f",
				math.Abs(newExposure), limit)
			return result, nil
		}
	}
	result.Checks["instrument_limit_ok"] = true

	// Check 6: Fat finger protection
	if clientProfile.FatFingerThreshold > 0 {
		avgOrderSize := v.engine.GetAverageOrderSize(accountID, symbol)
		if avgOrderSize > 0 && volume > avgOrderSize*clientProfile.FatFingerThreshold {
			result.Checks["fat_finger_ok"] = false
			result.Allowed = false
			result.Reason = fmt.Sprintf("Order size %.2f lots is %.0fx larger than typical (%.2f lots)",
				volume, volume/avgOrderSize, avgOrderSize)
			return result, nil
		}
	}
	result.Checks["fat_finger_ok"] = true

	// Check 7: Daily loss limit
	dailyPnL := v.engine.GetDailyPnL(accountID)
	if dailyPnL < 0 && math.Abs(dailyPnL) >= clientProfile.DailyLossLimit {
		result.Checks["daily_loss_ok"] = false
		result.Allowed = false
		result.Reason = fmt.Sprintf("Daily loss limit reached: %.2f >= %.2f",
			math.Abs(dailyPnL), clientProfile.DailyLossLimit)
		return result, nil
	}
	result.Checks["daily_loss_ok"] = true

	// Check 8: Max drawdown
	peakEquity := v.engine.GetPeakEquity(accountID)
	currentDrawdown := ((peakEquity - account.Equity) / peakEquity) * 100
	if currentDrawdown >= clientProfile.MaxDrawdownPercent {
		result.Checks["drawdown_ok"] = false
		result.Allowed = false
		result.Reason = fmt.Sprintf("Max drawdown reached: %.2f%% >= %.2f%%",
			currentDrawdown, clientProfile.MaxDrawdownPercent)
		return result, nil
	}
	result.Checks["drawdown_ok"] = true

	// Check 9: Max positions
	currentPositions := v.engine.GetPositionCount(accountID)
	if currentPositions >= clientProfile.MaxPositions {
		result.Checks["max_positions_ok"] = false
		result.Allowed = false
		result.Reason = fmt.Sprintf("Max positions limit reached: %d >= %d",
			currentPositions, clientProfile.MaxPositions)
		return result, nil
	}
	result.Checks["max_positions_ok"] = true

	// Check 10: Total exposure limit
	currentExposure := v.engine.GetTotalExposure(accountID)
	exposurePercent := (currentExposure / account.Equity) * 100
	if exposurePercent >= clientProfile.MaxExposurePercent {
		result.Checks["exposure_ok"] = false
		result.Allowed = false
		result.Reason = fmt.Sprintf("Total exposure limit reached: %.2f%% >= %.2f%%",
			exposurePercent, clientProfile.MaxExposurePercent)
		return result, nil
	}
	result.Checks["exposure_ok"] = true

	// Check 11: Stop loss requirement
	if clientProfile.RequireStopLoss || instrParams.RequireStopLoss {
		// This check would require SL parameter - add warning for now
		result.Warnings = append(result.Warnings, "Stop loss is required for this account/instrument")
	}

	// Check 12: Circuit breakers
	breaker := v.engine.GetCircuitBreaker(symbol)
	if breaker != nil && breaker.Status == CircuitTripped {
		result.Checks["circuit_breaker_ok"] = false
		result.Allowed = false
		result.Reason = fmt.Sprintf("Circuit breaker tripped for %s: %s", symbol, breaker.Message)
		return result, nil
	}
	result.Checks["circuit_breaker_ok"] = true

	// Check 13: Trading session (if required)
	if instrParams.TradingSessionOnly {
		if !v.engine.IsMarketOpen(symbol) {
			result.Checks["market_hours_ok"] = false
			result.Allowed = false
			result.Reason = fmt.Sprintf("Market closed for %s", symbol)
			return result, nil
		}
	}
	result.Checks["market_hours_ok"] = true

	// Check 14: Regulatory limits (ESMA leverage limits for retail)
	if clientProfile.RiskTier == "RETAIL" {
		maxLeverage := v.getESMAMaxLeverage(symbol)
		if maxLeverage > 0 {
			effectiveLeverage := notionalValue / account.Equity
			if effectiveLeverage > float64(maxLeverage) {
				result.Checks["regulatory_leverage_ok"] = false
				result.Allowed = false
				result.Reason = fmt.Sprintf("ESMA leverage limit exceeded: %.0f:1 > %d:1",
					effectiveLeverage, maxLeverage)
				return result, nil
			}
		}
	}
	result.Checks["regulatory_leverage_ok"] = true

	// Check 15: Credit limit
	usedCredit := v.engine.GetUsedCredit(account.ClientID)
	if usedCredit+notionalValue > clientProfile.CreditLimit {
		result.Checks["credit_limit_ok"] = false
		result.Allowed = false
		result.Reason = fmt.Sprintf("Credit limit exceeded: %.2f > %.2f",
			usedCredit+notionalValue, clientProfile.CreditLimit)
		return result, nil
	}
	result.Checks["credit_limit_ok"] = true

	log.Printf("[PreTrade] Account #%d: %s %.2f lots %s - APPROVED (margin: %.2f, free: %.2f)",
		accountID, symbol, volume, side, requiredMargin, account.FreeMargin)

	return result, nil
}

// getDirectionMultiplier returns 1 for buy, -1 for sell
func (v *PreTradeValidator) getDirectionMultiplier(side string) float64 {
	if side == "BUY" || side == "LONG" {
		return 1.0
	}
	return -1.0
}

// getESMAMaxLeverage returns ESMA leverage limits for retail clients
func (v *PreTradeValidator) getESMAMaxLeverage(symbol string) int {
	// ESMA leverage limits (EU regulation)
	// Major FX: 30:1
	majorPairs := map[string]bool{
		"EURUSD": true, "GBPUSD": true, "USDJPY": true,
		"USDCHF": true, "AUDUSD": true, "USDCAD": true, "NZDUSD": true,
	}
	if majorPairs[symbol] {
		return 30
	}

	// Minor FX, Gold, Major indices: 20:1
	if len(symbol) == 6 || symbol == "XAUUSD" ||
	   symbol == "SPX500USD" || symbol == "US30USD" || symbol == "NAS100USD" {
		return 20
	}

	// Commodities (non-gold): 10:1
	commodities := map[string]bool{
		"XAGUSD": true, "WTICOUSD": true, "BCOUSD": true,
	}
	if commodities[symbol] {
		return 10
	}

	// Individual stocks: 5:1
	// Crypto: 2:1
	cryptos := map[string]bool{
		"BTCUSD": true, "ETHUSD": true, "XRPUSD": true,
	}
	if cryptos[symbol] {
		return 2
	}

	return 0 // No limit
}

// ValidateModification validates SL/TP modification
func (v *PreTradeValidator) ValidateModification(
	accountID int64,
	positionID int64,
	newSL float64,
	newTP float64,
) error {
	// Get position
	position := v.engine.GetPosition(positionID)
	if position == nil {
		return errors.New("position not found")
	}

	if position.AccountID != accountID {
		return errors.New("position does not belong to account")
	}

	// Validate SL is on correct side of entry
	if newSL > 0 {
		if position.Side == "BUY" && newSL >= position.OpenPrice {
			return errors.New("stop loss for long position must be below entry price")
		}
		if position.Side == "SELL" && newSL <= position.OpenPrice {
			return errors.New("stop loss for short position must be above entry price")
		}
	}

	// Validate TP is on correct side of entry
	if newTP > 0 {
		if position.Side == "BUY" && newTP <= position.OpenPrice {
			return errors.New("take profit for long position must be above entry price")
		}
		if position.Side == "SELL" && newTP >= position.OpenPrice {
			return errors.New("take profit for short position must be below entry price")
		}
	}

	return nil
}

// ValidateClose validates position close request
func (v *PreTradeValidator) ValidateClose(accountID int64, positionID int64) error {
	position := v.engine.GetPosition(positionID)
	if position == nil {
		return errors.New("position not found")
	}

	if position.AccountID != accountID {
		return errors.New("position does not belong to account")
	}

	// Check circuit breakers
	breaker := v.engine.GetCircuitBreaker(position.Symbol)
	if breaker != nil && breaker.Status == CircuitTripped && breaker.Type == "SYSTEM" {
		return errors.New("system circuit breaker active - cannot close positions")
	}

	return nil
}
