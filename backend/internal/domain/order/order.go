package order

import (
	"errors"
	"time"
)

// Order represents a trading order with pure business logic
type Order struct {
	ID           int64
	AccountID    int64
	Symbol       string
	Type         string // MARKET, BUY_LIMIT, SELL_LIMIT, BUY_STOP, SELL_STOP, STOP_LOSS, TAKE_PROFIT, TRAILING_STOP
	Side         string // BUY or SELL
	Volume       float64
	Price        float64
	TriggerPrice float64
	SL           float64
	TP           float64
	Status       string
	FilledPrice  float64
	FilledAt     *time.Time
	PositionID   int64
	RejectReason string
	CreatedAt    time.Time
	ExpiresAt    *time.Time
	OCOLinkID    int64
	TrailingDelta float64
	ParentPositionID int64
}

// Validate validates the order parameters
func (o *Order) Validate() error {
	if o.Volume <= 0 {
		return errors.New("volume must be positive")
	}
	if o.Symbol == "" {
		return errors.New("symbol is required")
	}
	if o.Side != "BUY" && o.Side != "SELL" {
		return errors.New("side must be BUY or SELL")
	}

	// Validate order type specific fields
	switch o.Type {
	case "MARKET":
		// No additional validation needed
	case "BUY_LIMIT", "SELL_LIMIT":
		if o.TriggerPrice <= 0 {
			return errors.New("limit orders require trigger price")
		}
	case "BUY_STOP", "SELL_STOP":
		if o.TriggerPrice <= 0 {
			return errors.New("stop orders require trigger price")
		}
	case "STOP_LOSS", "TAKE_PROFIT":
		if o.ParentPositionID == 0 {
			return errors.New("SL/TP orders require parent position ID")
		}
		if o.TriggerPrice <= 0 {
			return errors.New("SL/TP orders require trigger price")
		}
	case "TRAILING_STOP":
		if o.ParentPositionID == 0 {
			return errors.New("trailing stop requires parent position ID")
		}
		if o.TrailingDelta <= 0 {
			return errors.New("trailing stop requires positive delta")
		}
	default:
		return errors.New("invalid order type")
	}

	return nil
}

// Fill marks the order as filled at the given price
func (o *Order) Fill(price float64) {
	o.Status = "FILLED"
	o.FilledPrice = price
	now := time.Now()
	o.FilledAt = &now
}

// Reject marks the order as rejected with the given reason
func (o *Order) Reject(reason string) {
	o.Status = "REJECTED"
	o.RejectReason = reason
}

// Cancel marks the order as cancelled
func (o *Order) Cancel(reason string) {
	o.Status = "CANCELLED"
	o.RejectReason = reason
}

// IsPending returns true if the order is still pending
func (o *Order) IsPending() bool {
	return o.Status == "PENDING"
}

// IsFilled returns true if the order is filled
func (o *Order) IsFilled() bool {
	return o.Status == "FILLED"
}

// IsExpired checks if the order has expired
func (o *Order) IsExpired(currentTime time.Time) bool {
	if o.ExpiresAt == nil {
		return false
	}
	return currentTime.After(*o.ExpiresAt)
}

// ShouldTrigger checks if a pending order should trigger at the given price
func (o *Order) ShouldTrigger(bid, ask float64) bool {
	if !o.IsPending() {
		return false
	}

	switch o.Type {
	case "BUY_LIMIT":
		// Trigger when ask price drops to or below trigger price
		return ask <= o.TriggerPrice
	case "SELL_LIMIT":
		// Trigger when bid price rises to or above trigger price
		return bid >= o.TriggerPrice
	case "BUY_STOP":
		// Trigger when ask price rises to or above trigger price
		return ask >= o.TriggerPrice
	case "SELL_STOP":
		// Trigger when bid price drops to or below trigger price
		return bid <= o.TriggerPrice
	case "STOP_LOSS", "TAKE_PROFIT":
		// Same logic as stop orders
		if o.Side == "BUY" {
			return ask >= o.TriggerPrice
		}
		return bid <= o.TriggerPrice
	}

	return false
}

// UpdateTriggerPrice updates the trigger price (for modification)
func (o *Order) UpdateTriggerPrice(newTriggerPrice float64) error {
	if !o.IsPending() {
		return errors.New("cannot modify non-pending order")
	}
	if newTriggerPrice <= 0 {
		return errors.New("trigger price must be positive")
	}
	o.TriggerPrice = newTriggerPrice
	return nil
}

// UpdateTrailingDelta updates the trailing delta
func (o *Order) UpdateTrailingDelta(newDelta float64) error {
	if o.Type != "TRAILING_STOP" {
		return errors.New("can only update trailing delta for trailing stop orders")
	}
	if newDelta <= 0 {
		return errors.New("trailing delta must be positive")
	}
	o.TrailingDelta = newDelta
	return nil
}

// IsMarketOrder returns true if this is a market order
func (o *Order) IsMarketOrder() bool {
	return o.Type == "MARKET"
}

// IsPendingOrder returns true if this is a pending order type
func (o *Order) IsPendingOrder() bool {
	return o.Type == "BUY_LIMIT" || o.Type == "SELL_LIMIT" ||
		o.Type == "BUY_STOP" || o.Type == "SELL_STOP"
}

// IsStopLossOrder returns true if this is a stop-loss order
func (o *Order) IsStopLossOrder() bool {
	return o.Type == "STOP_LOSS"
}

// IsTakeProfitOrder returns true if this is a take-profit order
func (o *Order) IsTakeProfitOrder() bool {
	return o.Type == "TAKE_PROFIT"
}

// IsTrailingStopOrder returns true if this is a trailing stop order
func (o *Order) IsTrailingStopOrder() bool {
	return o.Type == "TRAILING_STOP"
}

// HasOCOLink returns true if this order is linked to another via OCO
func (o *Order) HasOCOLink() bool {
	return o.OCOLinkID != 0
}
