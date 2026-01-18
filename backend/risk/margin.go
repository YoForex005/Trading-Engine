package risk

import (
	"fmt"
	"log"
	"math"
	"time"
)

// MarginCalculator handles margin calculations with multiple methods
type MarginCalculator struct {
	engine *Engine
}

// NewMarginCalculator creates a new margin calculator
func NewMarginCalculator(engine *Engine) *MarginCalculator {
	return &MarginCalculator{
		engine: engine,
	}
}

// CalculateMargin calculates required margin based on method
func (m *MarginCalculator) CalculateMargin(
	accountID int64,
	symbol string,
	volume float64,
	price float64,
	method MarginCalculationType,
) (float64, error) {

	switch method {
	case MarginRetail:
		return m.calculateRetailMargin(accountID, symbol, volume, price)
	case MarginPortfolio:
		return m.calculatePortfolioMargin(accountID, symbol, volume, price)
	case MarginSPAN:
		return m.calculateSPANMargin(accountID, symbol, volume, price)
	default:
		return m.calculateRetailMargin(accountID, symbol, volume, price)
	}
}

// calculateRetailMargin calculates margin using fixed percentage per instrument
func (m *MarginCalculator) calculateRetailMargin(
	accountID int64,
	symbol string,
	volume float64,
	price float64,
) (float64, error) {

	// Get instrument parameters
	instrParams := m.engine.GetInstrumentRiskParams(symbol)
	if instrParams == nil {
		return 0, fmt.Errorf("instrument %s not configured", symbol)
	}

	// Get account for leverage
	account, err := m.engine.GetAccountByID(accountID)
	if err != nil {
		return 0, err
	}

	// Calculate contract size
	contractSize := m.getContractSize(symbol)

	// Notional value
	notionalValue := volume * contractSize * price

	// Margin percentage (use instrument specific or account leverage)
	marginPercent := instrParams.MarginRequirement
	if marginPercent == 0 {
		marginPercent = 100.0 / float64(account.Leverage)
	}

	requiredMargin := notionalValue * (marginPercent / 100.0)

	log.Printf("[Margin-Retail] %s: %.2f lots @ %.5f = notional %.2f, margin %.2f (%.2f%%)",
		symbol, volume, price, notionalValue, requiredMargin, marginPercent)

	return requiredMargin, nil
}

// calculatePortfolioMargin implements cross-margining across positions
// This allows offsetting positions to reduce total margin requirement
func (m *MarginCalculator) calculatePortfolioMargin(
	accountID int64,
	symbol string,
	volume float64,
	price float64,
) (float64, error) {

	// Get all positions for the account
	positions := m.engine.GetAllPositions(accountID)

	// Calculate base margin for new position
	baseMargin, err := m.calculateRetailMargin(accountID, symbol, volume, price)
	if err != nil {
		return 0, err
	}

	// Calculate offsets from existing positions
	offset := 0.0

	// Check for offsetting positions in same instrument
	for _, pos := range positions {
		if pos.Symbol == symbol {
			// Calculate correlation-based offset
			correlationFactor := 0.8 // 80% offset for same instrument opposite side

			posNotional := pos.Volume * m.getContractSize(symbol) * pos.CurrentPrice
			newNotional := volume * m.getContractSize(symbol) * price

			// Opposite sides offset each other
			if (pos.Side == "BUY" && volume < 0) || (pos.Side == "SELL" && volume > 0) {
				overlapNotional := math.Min(math.Abs(posNotional), math.Abs(newNotional))
				offset += overlapNotional * correlationFactor * 0.01 // 1% margin on offset
			}
		}
	}

	// Check for correlated instruments
	correlation := m.engine.GetCorrelation(symbol, "")
	if correlation > 0.7 {
		// Highly correlated - apply discount
		offset += baseMargin * (correlation - 0.7) * 0.5 // Up to 15% discount
	}

	portfolioMargin := baseMargin - offset
	if portfolioMargin < baseMargin*0.5 {
		portfolioMargin = baseMargin * 0.5 // Minimum 50% of base margin
	}

	log.Printf("[Margin-Portfolio] %s: base %.2f, offset %.2f, portfolio %.2f",
		symbol, baseMargin, offset, portfolioMargin)

	return portfolioMargin, nil
}

// calculateSPANMargin implements SPAN (Standard Portfolio Analysis of Risk)
// Used primarily for futures and options
func (m *MarginCalculator) calculateSPANMargin(
	accountID int64,
	symbol string,
	volume float64,
	price float64,
) (float64, error) {

	// SPAN margin calculation involves:
	// 1. Price scan range (typically ±3 standard deviations)
	// 2. Volatility scan range (typically ±10%)
	// 3. Worst-case scenario across all scans

	volatility := m.engine.GetVolatility(symbol)
	if volatility == 0 {
		volatility = 0.15 // Default 15% annualized
	}

	contractSize := m.getContractSize(symbol)
	notionalValue := volume * contractSize * price

	// Price scan range: ±3 standard deviations (daily)
	dailyVol := volatility / math.Sqrt(252) // Annualized to daily
	priceScanRange := price * 3 * dailyVol

	// Calculate worst-case loss across 16 scenarios
	scenarios := []struct {
		priceMove float64
		volMove   float64
	}{
		{0, 0},           // No change
		{1, 0}, {-1, 0},  // Price up/down
		{0, 0.1}, {0, -0.1}, // Vol up/down
		{1, 0.1}, {1, -0.1},   // Price up, vol up/down
		{-1, 0.1}, {-1, -0.1}, // Price down, vol up/down
		{0.5, 0.05}, {-0.5, 0.05},   // Mid price moves
		{0.5, -0.05}, {-0.5, -0.05},
		{0.33, 0}, {-0.33, 0}, // Third moves
		{0.67, 0}, {-0.67, 0}, // Two-thirds moves
	}

	maxLoss := 0.0
	for _, scenario := range scenarios {
		newPrice := price + (priceScanRange * scenario.priceMove)
		loss := math.Abs(newPrice-price) * volume * contractSize
		if loss > maxLoss {
			maxLoss = loss
		}
	}

	// SPAN margin is worst-case loss plus a volatility adjustment
	spanMargin := maxLoss + (notionalValue * volatility * 0.1)

	// Apply inter-commodity spreading credits (simplified)
	positions := m.engine.GetAllPositions(accountID)
	spreadCredit := 0.0
	for _, pos := range positions {
		if m.isRelatedInstrument(symbol, pos.Symbol) {
			spreadCredit += spanMargin * 0.15 // 15% credit for related instruments
		}
	}

	spanMargin -= spreadCredit
	if spanMargin < maxLoss*0.6 {
		spanMargin = maxLoss * 0.6 // Minimum 60% of max loss
	}

	log.Printf("[Margin-SPAN] %s: max loss %.2f, vol adj %.2f, spread credit %.2f, SPAN %.2f",
		symbol, maxLoss, notionalValue*volatility*0.1, spreadCredit, spanMargin)

	return spanMargin, nil
}

// UpdateAccountMargin recalculates total margin for an account
func (m *MarginCalculator) UpdateAccountMargin(accountID int64) error {
	account, err := m.engine.GetAccountByID(accountID)
	if err != nil {
		return err
	}

	clientProfile := m.engine.GetClientRiskProfile(account.ClientID)
	if clientProfile == nil {
		return fmt.Errorf("client profile not found")
	}

	positions := m.engine.GetAllPositions(accountID)
	totalMargin := 0.0

	// Calculate margin for each position
	for _, pos := range positions {
		margin, err := m.CalculateMargin(
			accountID,
			pos.Symbol,
			pos.Volume,
			pos.CurrentPrice,
			clientProfile.MarginMethod,
		)
		if err != nil {
			log.Printf("[Margin] Error calculating margin for position %d: %v", pos.ID, err)
			continue
		}
		totalMargin += margin
	}

	// Update account margin
	account.Margin = totalMargin
	account.FreeMargin = account.Equity - account.Margin

	// Calculate margin level
	if account.Margin > 0 {
		account.MarginLevel = (account.Equity / account.Margin) * 100
	} else {
		account.MarginLevel = 0
	}

	log.Printf("[Margin] Account #%d: total margin %.2f, free %.2f, level %.2f%%",
		accountID, account.Margin, account.FreeMargin, account.MarginLevel)

	return nil
}

// CalculateMaintenanceMargin calculates maintenance margin requirement
// Typically lower than initial margin
func (m *MarginCalculator) CalculateMaintenanceMargin(
	accountID int64,
	symbol string,
	volume float64,
	price float64,
) (float64, error) {

	// Maintenance margin is typically 50-75% of initial margin
	initialMargin, err := m.calculateRetailMargin(accountID, symbol, volume, price)
	if err != nil {
		return 0, err
	}

	maintenanceMargin := initialMargin * 0.5 // 50% of initial

	return maintenanceMargin, nil
}

// getContractSize returns the contract size for an instrument
func (m *MarginCalculator) getContractSize(symbol string) float64 {
	switch symbol {
	case "XAUUSD":
		return 100.0 // 100 troy ounces
	case "XAGUSD":
		return 5000.0 // 5000 troy ounces
	case "BTCUSD", "ETHUSD":
		return 1.0 // 1 coin
	default:
		if len(symbol) == 6 {
			return 100000.0 // Standard forex lot
		}
		return 1.0 // Indices, stocks
	}
}

// isRelatedInstrument checks if two instruments are related for SPAN credits
func (m *MarginCalculator) isRelatedInstrument(symbol1, symbol2 string) bool {
	// Simplified - check if they share a currency
	if len(symbol1) == 6 && len(symbol2) == 6 {
		base1 := symbol1[:3]
		quote1 := symbol1[3:]
		base2 := symbol2[:3]
		quote2 := symbol2[3:]

		return base1 == base2 || base1 == quote2 || quote1 == base2 || quote1 == quote2
	}

	// Same asset class
	metals := map[string]bool{"XAUUSD": true, "XAGUSD": true, "XPTUSD": true, "XPDUSD": true}
	if metals[symbol1] && metals[symbol2] {
		return true
	}

	cryptos := map[string]bool{"BTCUSD": true, "ETHUSD": true, "XRPUSD": true}
	if cryptos[symbol1] && cryptos[symbol2] {
		return true
	}

	return false
}

// CalculateMarginCall determines if margin call should be triggered
func (m *MarginCalculator) CalculateMarginCall(accountID int64) (*MarginCall, error) {
	account, err := m.engine.GetAccountByID(accountID)
	if err != nil {
		return nil, err
	}

	clientProfile := m.engine.GetClientRiskProfile(account.ClientID)
	if clientProfile == nil {
		return nil, fmt.Errorf("client profile not found")
	}

	// Check if margin level is below margin call threshold
	if account.MarginLevel > 0 && account.MarginLevel < clientProfile.MarginCallLevel {

		// Calculate how much deposit is needed to restore margin
		targetEquity := account.Margin * (clientProfile.MarginCallLevel / 100.0)
		requiredDeposit := targetEquity - account.Equity

		severity := RiskLevelMedium
		if account.MarginLevel < clientProfile.StopOutLevel*1.2 {
			severity = RiskLevelHigh
		}
		if account.MarginLevel < clientProfile.StopOutLevel {
			severity = RiskLevelCritical
		}

		marginCall := &MarginCall{
			ID:              fmt.Sprintf("MC_%d_%d", accountID, time.Now().Unix()),
			AccountID:       accountID,
			MarginLevel:     account.MarginLevel,
			RequiredDeposit: requiredDeposit,
			Severity:        severity,
			TriggeredAt:     m.engine.now(),
			Status:          "ACTIVE",
			Actions:         []string{"Email sent", "Popup notification"},
		}

		log.Printf("[MarginCall] Account #%d: level %.2f%% < %.2f%%, deposit needed: %.2f",
			accountID, account.MarginLevel, clientProfile.MarginCallLevel, requiredDeposit)

		return marginCall, nil
	}

	return nil, nil // No margin call
}
