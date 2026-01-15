package position

import (
	"time"
)

// Position represents an open trading position with pure business logic
type Position struct {
	ID            int64
	AccountID     int64
	Symbol        string
	Side          string // BUY or SELL
	Volume        float64
	OpenPrice     float64
	CurrentPrice  float64
	OpenTime      time.Time
	SL            float64
	TP            float64
	Swap          float64
	Commission    float64
	UnrealizedPnL float64
	Status        string
	ClosePrice    float64
	CloseTime     time.Time
	CloseReason   string
}

// CalculatePnL computes unrealized P&L at the given current price
func (p *Position) CalculatePnL(currentPrice float64) float64 {
	priceDiff := currentPrice - p.OpenPrice
	if p.Side == "SELL" {
		priceDiff = -priceDiff
	}
	return priceDiff * p.Volume
}

// IsProfitable checks if position has unrealized profit at current price
func (p *Position) IsProfitable(currentPrice float64) bool {
	return p.CalculatePnL(currentPrice) > 0
}

// UpdateCurrentPrice updates the current price and recalculates P&L
func (p *Position) UpdateCurrentPrice(price float64) {
	p.CurrentPrice = price
	p.UnrealizedPnL = p.CalculatePnL(price)
}

// SetStopLoss sets the stop-loss price
func (p *Position) SetStopLoss(sl float64) {
	p.SL = sl
}

// SetTakeProfit sets the take-profit price
func (p *Position) SetTakeProfit(tp float64) {
	p.TP = tp
}

// Close closes the position at the given price and reason
func (p *Position) Close(closePrice float64, reason string) {
	p.Status = "CLOSED"
	p.ClosePrice = closePrice
	p.CloseTime = time.Now()
	p.CloseReason = reason
	p.UpdateCurrentPrice(closePrice)
}

// IsOpen returns true if the position is open
func (p *Position) IsOpen() bool {
	return p.Status == "OPEN"
}

// ShouldTriggerStopLoss checks if stop-loss should trigger
func (p *Position) ShouldTriggerStopLoss(currentPrice float64) bool {
	if p.SL == 0 {
		return false
	}

	if p.Side == "BUY" {
		return currentPrice <= p.SL
	}
	return currentPrice >= p.SL
}

// ShouldTriggerTakeProfit checks if take-profit should trigger
func (p *Position) ShouldTriggerTakeProfit(currentPrice float64) bool {
	if p.TP == 0 {
		return false
	}

	if p.Side == "BUY" {
		return currentPrice >= p.TP
	}
	return currentPrice <= p.TP
}

// AddSwap adds swap/rollover charges
func (p *Position) AddSwap(swap float64) {
	p.Swap += swap
}

// SetCommission sets the commission for this position
func (p *Position) SetCommission(commission float64) {
	p.Commission = commission
}

// GetRealizedPnL calculates realized P&L including swap and commission
func (p *Position) GetRealizedPnL() float64 {
	if !p.IsOpen() {
		return p.UnrealizedPnL - p.Commission - p.Swap
	}
	return 0
}
