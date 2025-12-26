package orders

import (
	"errors"
	"log"
	"sync"
)

// Position represents an open trading position
type Position struct {
	ID           string  `json:"id"`
	Symbol       string  `json:"symbol"`
	Side         string  `json:"side"` // BUY or SELL
	Volume       float64 `json:"volume"`
	OpenPrice    float64 `json:"openPrice"`
	CurrentPrice float64 `json:"currentPrice"`
	SL           float64 `json:"sl,omitempty"`
	TP           float64 `json:"tp,omitempty"`
	UnrealizedPL float64 `json:"unrealizedPL"`
	LP           string  `json:"lp"`
}

// PositionManager handles position operations
type PositionManager struct {
	mu              sync.RWMutex
	positions       map[string]*Position
	hedgingMode     bool // true = hedging (multiple positions per symbol), false = netting
	closeCallback   func(tradeID string, units int) error
	partialCallback func(tradeID string, units int) error
	reverseCallback func(symbol string, currentUnits int) error
}

// NewPositionManager creates a new position manager
func NewPositionManager(hedgingMode bool) *PositionManager {
	return &PositionManager{
		positions:   make(map[string]*Position),
		hedgingMode: hedgingMode,
	}
}

// SetCloseCallback sets the function to close trades via LP
func (pm *PositionManager) SetCloseCallback(fn func(tradeID string, units int) error) {
	pm.closeCallback = fn
}

// SetPartialCallback sets the function for partial closes
func (pm *PositionManager) SetPartialCallback(fn func(tradeID string, units int) error) {
	pm.partialCallback = fn
}

// SetReverseCallback sets the function for reversing positions
func (pm *PositionManager) SetReverseCallback(fn func(symbol string, currentUnits int) error) {
	pm.reverseCallback = fn
}

// IsHedgingMode returns current mode
func (pm *PositionManager) IsHedgingMode() bool {
	return pm.hedgingMode
}

// SetHedgingMode changes the position mode
func (pm *PositionManager) SetHedgingMode(hedging bool) {
	pm.hedgingMode = hedging
	log.Printf("[PositionManager] Mode set to: %s", map[bool]string{true: "HEDGING", false: "NETTING"}[hedging])
}

// UpdatePosition updates or adds a position
func (pm *PositionManager) UpdatePosition(pos *Position) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.positions[pos.ID] = pos
}

// GetPosition returns a position by ID
func (pm *PositionManager) GetPosition(id string) (*Position, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	pos, ok := pm.positions[id]
	return pos, ok
}

// GetAllPositions returns all open positions
func (pm *PositionManager) GetAllPositions() []*Position {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	positions := make([]*Position, 0, len(pm.positions))
	for _, p := range pm.positions {
		positions = append(positions, p)
	}
	return positions
}

// GetPositionsBySymbol returns positions for a specific symbol
func (pm *PositionManager) GetPositionsBySymbol(symbol string) []*Position {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var positions []*Position
	for _, p := range pm.positions {
		if p.Symbol == symbol {
			positions = append(positions, p)
		}
	}
	return positions
}

// PartialClose closes a percentage of a position
func (pm *PositionManager) PartialClose(tradeID string, percent float64) error {
	if percent <= 0 || percent > 100 {
		return errors.New("invalid percentage (must be 1-100)")
	}

	pm.mu.RLock()
	pos, exists := pm.positions[tradeID]
	pm.mu.RUnlock()

	if !exists {
		return errors.New("position not found")
	}

	// Calculate units to close
	unitsToClose := int(pos.Volume * 100000 * (percent / 100))
	if unitsToClose == 0 {
		return errors.New("partial close too small")
	}

	log.Printf("[PositionManager] Partial close: %s %.0f%% (%d units)", tradeID, percent, unitsToClose)

	if pm.partialCallback != nil {
		return pm.partialCallback(tradeID, unitsToClose)
	}

	// Update local position
	pm.mu.Lock()
	pos.Volume = pos.Volume * (1 - percent/100)
	if pos.Volume <= 0.0001 {
		delete(pm.positions, tradeID)
	}
	pm.mu.Unlock()

	return nil
}

// ClosePosition fully closes a position
func (pm *PositionManager) ClosePosition(tradeID string) error {
	pm.mu.RLock()
	pos, exists := pm.positions[tradeID]
	pm.mu.RUnlock()

	if !exists {
		return errors.New("position not found")
	}

	units := int(pos.Volume * 100000)
	log.Printf("[PositionManager] Closing position: %s (%d units)", tradeID, units)

	if pm.closeCallback != nil {
		if err := pm.closeCallback(tradeID, units); err != nil {
			return err
		}
	}

	pm.mu.Lock()
	delete(pm.positions, tradeID)
	pm.mu.Unlock()

	return nil
}

// CloseAllBySymbol closes all positions for a symbol
func (pm *PositionManager) CloseAllBySymbol(symbol string) (int, error) {
	positions := pm.GetPositionsBySymbol(symbol)
	
	closed := 0
	for _, pos := range positions {
		if err := pm.ClosePosition(pos.ID); err != nil {
			log.Printf("[PositionManager] Failed to close %s: %v", pos.ID, err)
			continue
		}
		closed++
	}

	log.Printf("[PositionManager] Closed %d/%d positions for %s", closed, len(positions), symbol)
	return closed, nil
}

// CloseAll closes all open positions
func (pm *PositionManager) CloseAll() (int, error) {
	positions := pm.GetAllPositions()

	closed := 0
	for _, pos := range positions {
		if err := pm.ClosePosition(pos.ID); err != nil {
			log.Printf("[PositionManager] Failed to close %s: %v", pos.ID, err)
			continue
		}
		closed++
	}

	log.Printf("[PositionManager] Closed %d/%d total positions", closed, len(positions))
	return closed, nil
}

// ReversePosition closes current position and opens opposite
func (pm *PositionManager) ReversePosition(tradeID string) error {
	pm.mu.RLock()
	pos, exists := pm.positions[tradeID]
	pm.mu.RUnlock()

	if !exists {
		return errors.New("position not found")
	}

	units := int(pos.Volume * 100000)
	if pos.Side == "SELL" {
		units = -units // Make positive for calculation
	}

	log.Printf("[PositionManager] Reversing position: %s %s %d units", pos.Symbol, pos.Side, units)

	if pm.reverseCallback != nil {
		return pm.reverseCallback(pos.Symbol, units)
	}

	return nil
}

// CloseBy hedges/closes opposite positions
func (pm *PositionManager) CloseBy(tradeID1, tradeID2 string) error {
	pm.mu.RLock()
	pos1, exists1 := pm.positions[tradeID1]
	pos2, exists2 := pm.positions[tradeID2]
	pm.mu.RUnlock()

	if !exists1 || !exists2 {
		return errors.New("one or both positions not found")
	}

	if pos1.Symbol != pos2.Symbol {
		return errors.New("positions must be same symbol")
	}

	if pos1.Side == pos2.Side {
		return errors.New("positions must be opposite sides for close-by")
	}

	// Calculate net result
	vol1 := pos1.Volume
	vol2 := pos2.Volume

	log.Printf("[PositionManager] Close-by: %s (%.2f lots) vs %s (%.2f lots)",
		tradeID1, vol1, tradeID2, vol2)

	if vol1 == vol2 {
		// Equal volumes - close both
		pm.ClosePosition(tradeID1)
		pm.ClosePosition(tradeID2)
	} else if vol1 > vol2 {
		// Reduce pos1 by vol2
		pm.PartialClose(tradeID1, (vol2/vol1)*100)
		pm.ClosePosition(tradeID2)
	} else {
		// Reduce pos2 by vol1
		pm.PartialClose(tradeID2, (vol1/vol2)*100)
		pm.ClosePosition(tradeID1)
	}

	return nil
}

// SetBreakeven sets SL to entry price for a position
func (pm *PositionManager) SetBreakeven(tradeID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pos, exists := pm.positions[tradeID]
	if !exists {
		return errors.New("position not found")
	}

	pos.SL = pos.OpenPrice
	log.Printf("[PositionManager] Break-even set for %s at %.5f", tradeID, pos.OpenPrice)
	return nil
}

// ModifySLTP updates stop loss and take profit
func (pm *PositionManager) ModifySLTP(tradeID string, sl, tp float64) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pos, exists := pm.positions[tradeID]
	if !exists {
		return errors.New("position not found")
	}

	pos.SL = sl
	pos.TP = tp
	log.Printf("[PositionManager] Modified %s: SL=%.5f TP=%.5f", tradeID, sl, tp)
	return nil
}
