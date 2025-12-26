package risk

import (
	"errors"
	"log"
	"math"
	"strings"
)

// RiskCalculator provides risk-based calculations
type RiskCalculator struct {
	getBalance func() float64
	getPrice   func(symbol string) (bid, ask float64, ok bool)
}

// NewRiskCalculator creates a new risk calculator
func NewRiskCalculator() *RiskCalculator {
	return &RiskCalculator{}
}

// SetBalanceCallback sets the function to get account balance
func (rc *RiskCalculator) SetBalanceCallback(fn func() float64) {
	rc.getBalance = fn
}

// SetPriceCallback sets the function to get current prices
func (rc *RiskCalculator) SetPriceCallback(fn func(symbol string) (bid, ask float64, ok bool)) {
	rc.getPrice = fn
}

// LotCalcResult contains lot calculation results
type LotCalcResult struct {
	RiskPercent float64 `json:"riskPercent"`
	RiskAmount  float64 `json:"riskAmount"`
	SLPips      float64 `json:"slPips"`
	PipValue    float64 `json:"pipValue"`
	LotSize     float64 `json:"lotSize"`
	Units       int     `json:"units"`
}

// CalculateLotFromRisk calculates lot size based on risk percentage
func (rc *RiskCalculator) CalculateLotFromRisk(riskPercent, slPips float64, symbol string) (*LotCalcResult, error) {
	if riskPercent <= 0 || riskPercent > 100 {
		return nil, errors.New("risk percent must be 1-100")
	}
	if slPips <= 0 {
		return nil, errors.New("stop loss pips must be positive")
	}

	balance := 100000.0 // Default
	if rc.getBalance != nil {
		balance = rc.getBalance()
	}

	// Calculate risk amount
	riskAmount := balance * (riskPercent / 100)

	// Get pip value for symbol
	pipValue := rc.GetPipValuePerLot(symbol)

	// Calculate lot size: risk / (sl_pips * pip_value_per_lot)
	lotSize := riskAmount / (slPips * pipValue)

	// Round to 2 decimal places (0.01 min lot)
	lotSize = math.Floor(lotSize*100) / 100
	if lotSize < 0.01 {
		lotSize = 0.01
	}

	units := int(lotSize * 100000)

	result := &LotCalcResult{
		RiskPercent: riskPercent,
		RiskAmount:  riskAmount,
		SLPips:      slPips,
		PipValue:    pipValue,
		LotSize:     lotSize,
		Units:       units,
	}

	log.Printf("[RiskCalc] %s: %.1f%% risk (%.2f) / %.1f pips = %.2f lots",
		symbol, riskPercent, riskAmount, slPips, lotSize)

	return result, nil
}

// MarginPreview contains margin calculation results
type MarginPreview struct {
	Symbol          string  `json:"symbol"`
	Volume          float64 `json:"volume"`
	Side            string  `json:"side"`
	RequiredMargin  float64 `json:"requiredMargin"`
	CurrentMargin   float64 `json:"currentMargin"`
	FreeMargin      float64 `json:"freeMargin"`
	FreeMarginAfter float64 `json:"freeMarginAfter"`
	MarginLevel     float64 `json:"marginLevel"`     // Percentage
	CanTrade        bool    `json:"canTrade"`
}

// PreviewMargin calculates margin requirements for a trade
func (rc *RiskCalculator) PreviewMargin(symbol string, volume float64, side string, leverage int, currentMargin, freeMargin float64) (*MarginPreview, error) {
	if volume <= 0 {
		return nil, errors.New("volume must be positive")
	}
	if leverage <= 0 {
		leverage = 100 // Default leverage
	}

	// Get current price
	var price float64 = 1.0
	if rc.getPrice != nil {
		bid, ask, ok := rc.getPrice(symbol)
		if ok {
			if side == "BUY" {
				price = ask
			} else {
				price = bid
			}
		}
	}

	// Calculate notional value
	units := volume * 100000
	notional := units * price

	// Required margin = notional / leverage
	requiredMargin := notional / float64(leverage)

	// Calculate free margin after trade
	freeMarginAfter := freeMargin - requiredMargin

	// Calculate margin level
	balance := 100000.0
	if rc.getBalance != nil {
		balance = rc.getBalance()
	}
	
	totalMargin := currentMargin + requiredMargin
	marginLevel := 0.0
	if totalMargin > 0 {
		marginLevel = (balance / totalMargin) * 100
	}

	canTrade := freeMarginAfter > 0

	preview := &MarginPreview{
		Symbol:          symbol,
		Volume:          volume,
		Side:            side,
		RequiredMargin:  requiredMargin,
		CurrentMargin:   currentMargin,
		FreeMargin:      freeMargin,
		FreeMarginAfter: freeMarginAfter,
		MarginLevel:     marginLevel,
		CanTrade:        canTrade,
	}

	log.Printf("[RiskCalc] Margin preview %s %.2f lots: Required=%.2f, After=%.2f, CanTrade=%v",
		symbol, volume, requiredMargin, freeMarginAfter, canTrade)

	return preview, nil
}

// GetPipValuePerLot returns pip value for 1 standard lot
func (rc *RiskCalculator) GetPipValuePerLot(symbol string) float64 {
	// Standard lot = 100,000 units

	// For XXX/USD pairs, pip value is $10 per lot
	if strings.HasSuffix(symbol, "USD") && len(symbol) == 6 {
		return 10.0
	}

	// For USD/XXX pairs, need to calculate
	if strings.HasPrefix(symbol, "USD") && len(symbol) == 6 {
		if rc.getPrice != nil {
			_, ask, ok := rc.getPrice(symbol)
			if ok && ask > 0 {
				return 10.0 / ask
			}
		}
		return 10.0 // Approximate
	}

	// For JPY pairs
	if strings.Contains(symbol, "JPY") {
		// JPY pairs: 1 pip = 0.01, standard lot pip value varies
		if strings.HasSuffix(symbol, "JPY") {
			if rc.getPrice != nil {
				_, ask, ok := rc.getPrice("USDJPY")
				if ok && ask > 0 {
					return 1000.0 / ask
				}
			}
			return 9.0 // Approximate at USDJPY ~110
		}
	}

	// Gold (XAU/USD)
	if symbol == "XAUUSD" {
		return 10.0 // Per 0.1 move (1 pip)
	}

	// Silver (XAG/USD)
	if symbol == "XAGUSD" {
		return 50.0 // Per 0.01 move
	}

	// Indices (approximate)
	if symbol == "US30USD" {
		return 1.0
	}
	if symbol == "SPX500USD" {
		return 1.0
	}
	if symbol == "NAS100USD" {
		return 1.0
	}

	// Default
	return 10.0
}

// ConvertPipsToPrice converts pips to price distance
func (rc *RiskCalculator) ConvertPipsToPrice(symbol string, pips float64) float64 {
	// JPY pairs: 1 pip = 0.01
	if strings.Contains(symbol, "JPY") {
		return pips * 0.01
	}
	
	// Gold
	if symbol == "XAUUSD" {
		return pips * 0.1
	}
	
	// Indices
	if symbol == "US30USD" || symbol == "SPX500USD" || symbol == "NAS100USD" {
		return pips * 1.0
	}
	
	// Standard forex: 1 pip = 0.0001
	return pips * 0.0001
}

// ConvertPriceToMoney converts price difference to money (P/L)
func (rc *RiskCalculator) ConvertPriceToMoney(symbol string, priceDiff float64, volume float64) float64 {
	// Convert price diff to pips
	var pips float64
	if strings.Contains(symbol, "JPY") {
		pips = priceDiff / 0.01
	} else if symbol == "XAUUSD" {
		pips = priceDiff / 0.1
	} else {
		pips = priceDiff / 0.0001
	}

	// Calculate P/L
	pipValue := rc.GetPipValuePerLot(symbol)
	return pips * pipValue * volume
}

// CalculateSLFromMoney calculates SL distance from risk amount in money
func (rc *RiskCalculator) CalculateSLFromMoney(symbol string, riskMoney, volume float64) (float64, float64) {
	pipValue := rc.GetPipValuePerLot(symbol)
	
	// pips = money / (pip_value * lots)
	slPips := riskMoney / (pipValue * volume)
	slPrice := rc.ConvertPipsToPrice(symbol, slPips)
	
	return slPips, slPrice
}
