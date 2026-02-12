package core

import (
	"errors"
	"fmt"
	"log"
	"time"
)

// PositionMode represents the position management mode
type PositionMode string

const (
	// PositionModeHedging allows multiple positions per symbol (MT5-style)
	PositionModeHedging PositionMode = "hedging"
	// PositionModeNetting allows only one aggregated position per symbol
	PositionModeNetting PositionMode = "netting"
)

// NettingEngine handles netting mode position management
type NettingEngine struct {
	engine *Engine // Reference to parent engine
}

// NewNettingEngine creates a new netting engine
func NewNettingEngine(engine *Engine) *NettingEngine {
	return &NettingEngine{
		engine: engine,
	}
}

// GetNetPosition returns the aggregated net position for a symbol
// Returns nil if no position exists
func (ne *NettingEngine) GetNetPosition(accountID int64, symbol string) *Position {
	ne.engine.mu.RLock()
	defer ne.engine.mu.RUnlock()

	return ne.getNetPositionUnlocked(accountID, symbol)
}

// getNetPositionUnlocked is the unlocked version (assumes mutex is already held)
func (ne *NettingEngine) getNetPositionUnlocked(accountID int64, symbol string) *Position {
	for _, pos := range ne.engine.positions {
		if pos.AccountID == accountID && pos.Symbol == symbol && pos.Status == "OPEN" {
			return pos
		}
	}
	return nil
}

// AdjustPosition applies netting logic when opening a new order
// Returns the resulting position (or nil if fully closed) and realized P&L
func (ne *NettingEngine) AdjustPosition(accountID int64, symbol, newSide string, newVolume, newPrice float64, sl, tp, commission float64) (*Position, float64, error) {
	existingPos := ne.GetNetPosition(accountID, symbol)

	// No existing position - create new position
	if existingPos == nil {
		return ne.createNewPosition(accountID, symbol, newSide, newVolume, newPrice, sl, tp, commission)
	}

	// Same direction - increase position volume
	if existingPos.Side == newSide {
		return ne.increasePosition(existingPos, newVolume, newPrice, commission)
	}

	// Opposite direction - apply reduction/reversal logic
	return ne.reduceOrReversePosition(existingPos, newSide, newVolume, newPrice, sl, tp, commission)
}

// AdjustPositionUnlocked is the unlocked version (assumes mutex is already held)
func (ne *NettingEngine) AdjustPositionUnlocked(accountID int64, symbol, newSide string, newVolume, newPrice float64, sl, tp, commission float64) (*Position, float64, error) {
	existingPos := ne.getNetPositionUnlocked(accountID, symbol)

	// No existing position - create new position
	if existingPos == nil {
		return ne.createNewPositionUnlocked(accountID, symbol, newSide, newVolume, newPrice, sl, tp, commission)
	}

	// Same direction - increase position volume
	if existingPos.Side == newSide {
		return ne.increasePositionUnlocked(existingPos, newVolume, newPrice, commission)
	}

	// Opposite direction - apply reduction/reversal logic
	return ne.reduceOrReversePositionUnlocked(existingPos, newSide, newVolume, newPrice, sl, tp, commission)
}

// createNewPosition creates a new position for netting mode
func (ne *NettingEngine) createNewPosition(accountID int64, symbol, side string, volume, price, sl, tp, commission float64) (*Position, float64, error) {
	ne.engine.mu.Lock()
	defer ne.engine.mu.Unlock()

	return ne.createNewPositionUnlocked(accountID, symbol, side, volume, price, sl, tp, commission)
}

// createNewPositionUnlocked is the unlocked version
func (ne *NettingEngine) createNewPositionUnlocked(accountID int64, symbol, side string, volume, price, sl, tp, commission float64) (*Position, float64, error) {
	positionID := ne.engine.nextPositionID
	ne.engine.nextPositionID++

	position := &Position{
		ID:           positionID,
		AccountID:    accountID,
		Symbol:       symbol,
		Side:         side,
		Volume:       volume,
		OpenPrice:    price,
		CurrentPrice: price,
		OpenTime:     ne.engine.getCurrentTime(),
		SL:           sl,
		TP:           tp,
		Commission:   commission,
		Status:       "OPEN",
	}
	ne.engine.positions[positionID] = position

	log.Printf("[NettingEngine] Created new position: ID=%d, Symbol=%s, Side=%s, Volume=%.2f, Price=%.5f",
		positionID, symbol, side, volume, price)

	return position, 0.0, nil
}

// increasePosition increases an existing position's volume and recalculates average price
func (ne *NettingEngine) increasePosition(pos *Position, addVolume, newPrice, commission float64) (*Position, float64, error) {
	ne.engine.mu.Lock()
	defer ne.engine.mu.Unlock()

	return ne.increasePositionUnlocked(pos, addVolume, newPrice, commission)
}

// increasePositionUnlocked is the unlocked version
func (ne *NettingEngine) increasePositionUnlocked(pos *Position, addVolume, newPrice, commission float64) (*Position, float64, error) {
	oldVolume := pos.Volume
	oldPrice := pos.OpenPrice

	// Calculate new average price
	newAvgPrice := ne.CalculateAveragePrice(oldPrice, oldVolume, newPrice, addVolume)

	// Update position
	pos.Volume += addVolume
	pos.OpenPrice = newAvgPrice
	pos.Commission += commission

	log.Printf("[NettingEngine] Increased position %d: Volume %.2f→%.2f, AvgPrice %.5f→%.5f",
		pos.ID, oldVolume, pos.Volume, oldPrice, newAvgPrice)

	return pos, 0.0, nil
}

// reduceOrReversePosition handles opposite direction orders
func (ne *NettingEngine) reduceOrReversePosition(existingPos *Position, newSide string, newVolume, newPrice, sl, tp, commission float64) (*Position, float64, error) {
	ne.engine.mu.Lock()
	defer ne.engine.mu.Unlock()

	return ne.reduceOrReversePositionUnlocked(existingPos, newSide, newVolume, newPrice, sl, tp, commission)
}

// reduceOrReversePositionUnlocked is the unlocked version
func (ne *NettingEngine) reduceOrReversePositionUnlocked(existingPos *Position, newSide string, newVolume, newPrice, sl, tp, commission float64) (*Position, float64, error) {
	oldVolume := existingPos.Volume
	volumeDiff := newVolume - oldVolume

	// Get symbol spec for P&L calculation
	spec, ok := ne.engine.symbols[existingPos.Symbol]
	if !ok {
		return nil, 0.0, fmt.Errorf("symbol %s not found", existingPos.Symbol)
	}

	// Case 1: Opposite direction with smaller volume - reduce position
	if newVolume < oldVolume {
		realizedPnL := ne.calculatePnL(existingPos, newPrice, newVolume, spec)
		existingPos.Volume -= newVolume
		existingPos.Commission += commission

		log.Printf("[NettingEngine] Reduced position %d: Volume %.2f→%.2f, Realized P&L: %.2f",
			existingPos.ID, oldVolume, existingPos.Volume, realizedPnL)

		// Record realized P&L
		if ne.engine.ledger != nil {
			ne.engine.ledger.RecordRealizedPnL(existingPos.AccountID, realizedPnL, existingPos.ID)
		}

		return existingPos, realizedPnL, nil
	}

	// Case 2: Opposite direction with equal volume - close position entirely
	if newVolume == oldVolume {
		realizedPnL := ne.calculatePnL(existingPos, newPrice, oldVolume, spec)

		// Close the position
		existingPos.Status = "CLOSED"
		existingPos.ClosePrice = newPrice
		existingPos.CloseTime = ne.engine.getCurrentTime()
		existingPos.CloseReason = "NETTING_FULL_CLOSE"
		existingPos.Commission += commission

		log.Printf("[NettingEngine] Closed position %d: Volume %.2f fully closed, Realized P&L: %.2f",
			existingPos.ID, oldVolume, realizedPnL)

		// Record realized P&L
		if ne.engine.ledger != nil {
			ne.engine.ledger.RecordRealizedPnL(existingPos.AccountID, realizedPnL, existingPos.ID)
		}

		return nil, realizedPnL, nil
	}

	// Case 3: Opposite direction with larger volume - close + reverse
	realizedPnL := ne.calculatePnL(existingPos, newPrice, oldVolume, spec)

	// Close existing position
	existingPos.Status = "CLOSED"
	existingPos.ClosePrice = newPrice
	existingPos.CloseTime = ne.engine.getCurrentTime()
	existingPos.CloseReason = "NETTING_REVERSE"

	log.Printf("[NettingEngine] Closed position %d for reversal: Volume %.2f, Realized P&L: %.2f",
		existingPos.ID, oldVolume, realizedPnL)

	// Record realized P&L
	if ne.engine.ledger != nil {
		ne.engine.ledger.RecordRealizedPnL(existingPos.AccountID, realizedPnL, existingPos.ID)
	}

	// Create reversed position with remaining volume
	remainingVolume := newVolume - oldVolume
	newPosID := ne.engine.nextPositionID
	ne.engine.nextPositionID++

	reversedPos := &Position{
		ID:           newPosID,
		AccountID:    existingPos.AccountID,
		Symbol:       existingPos.Symbol,
		Side:         newSide,
		Volume:       remainingVolume,
		OpenPrice:    newPrice,
		CurrentPrice: newPrice,
		OpenTime:     ne.engine.getCurrentTime(),
		SL:           sl,
		TP:           tp,
		Commission:   commission,
		Status:       "OPEN",
	}
	ne.engine.positions[newPosID] = reversedPos

	log.Printf("[NettingEngine] Created reversed position %d: Side=%s, Volume=%.2f (from %.2f close + %.2f new)",
		newPosID, newSide, remainingVolume, oldVolume, volumeDiff)

	return reversedPos, realizedPnL, nil
}

// CalculateAveragePrice calculates the weighted average price when adding to a position
func (ne *NettingEngine) CalculateAveragePrice(existingPrice, existingVolume, newPrice, newVolume float64) float64 {
	totalValue := (existingPrice * existingVolume) + (newPrice * newVolume)
	totalVolume := existingVolume + newVolume
	if totalVolume == 0 {
		return 0
	}
	return totalValue / totalVolume
}

// calculatePnL calculates P&L for a position close or partial close
func (ne *NettingEngine) calculatePnL(pos *Position, closePrice, volume float64, spec *SymbolSpec) float64 {
	priceDiff := closePrice - pos.OpenPrice
	if pos.Side == "SELL" {
		priceDiff = -priceDiff
	}

	pipValue := spec.PipValue
	if spec.PipSize > 0 {
		pips := priceDiff / spec.PipSize
		pnl := pips * pipValue * volume
		return pnl
	}

	// Fallback for symbols without pip size
	return priceDiff * volume * spec.ContractSize
}

// CanChangePositionMode checks if position mode can be changed
// Returns error if account has open positions
func (ne *NettingEngine) CanChangePositionMode(accountID int64) error {
	ne.engine.mu.RLock()
	defer ne.engine.mu.RUnlock()

	for _, pos := range ne.engine.positions {
		if pos.AccountID == accountID && pos.Status == "OPEN" {
			return errors.New("cannot change position mode while positions are open")
		}
	}
	return nil
}

// ValidatePositionMode validates if a mode string is valid
func ValidatePositionMode(mode string) error {
	if mode != string(PositionModeHedging) && mode != string(PositionModeNetting) {
		return fmt.Errorf("invalid position mode: must be 'hedging' or 'netting'")
	}
	return nil
}

// getCurrentTime helper
func (e *Engine) getCurrentTime() time.Time {
	return time.Now()
}
