package trade

import (
	"time"
)

// Trade represents an immutable execution record
type Trade struct {
	ID          int64
	OrderID     int64
	PositionID  int64
	AccountID   int64
	Symbol      string
	Side        string
	Volume      float64
	Price       float64
	RealizedPnL float64
	Commission  float64
	ExecutedAt  time.Time
}

// NewTrade creates a new trade record (immutable after creation)
func NewTrade(
	id, orderID, positionID, accountID int64,
	symbol, side string,
	volume, price, realizedPnL, commission float64,
) *Trade {
	return &Trade{
		ID:          id,
		OrderID:     orderID,
		PositionID:  positionID,
		AccountID:   accountID,
		Symbol:      symbol,
		Side:        side,
		Volume:      volume,
		Price:       price,
		RealizedPnL: realizedPnL,
		Commission:  commission,
		ExecutedAt:  time.Now(),
	}
}

// GetNetPnL returns the net P&L after commission
func (t *Trade) GetNetPnL() float64 {
	return t.RealizedPnL - t.Commission
}

// IsBuy returns true if this is a buy trade
func (t *Trade) IsBuy() bool {
	return t.Side == "BUY"
}

// IsSell returns true if this is a sell trade
func (t *Trade) IsSell() bool {
	return t.Side == "SELL"
}

// IsProfitable returns true if the net P&L is positive
func (t *Trade) IsProfitable() bool {
	return t.GetNetPnL() > 0
}
