package risk

import (
	"fmt"

	"github.com/epic1st/rtx/backend/internal/core"
)

// CoreEngineAdapter adapts core.Engine to TradingEngine interface
type CoreEngineAdapter struct {
	engine *core.Engine
}

// NewCoreEngineAdapter creates an adapter for core.Engine
func NewCoreEngineAdapter(engine *core.Engine) *CoreEngineAdapter {
	return &CoreEngineAdapter{engine: engine}
}

// GetAllAccounts returns all accounts
func (a *CoreEngineAdapter) GetAllAccounts() []*AccountInfo {
	accounts := a.engine.GetAllAccounts()
	result := make([]*AccountInfo, len(accounts))
	for i, acc := range accounts {
		result[i] = &AccountInfo{
			ID:          acc.ID,
			UserID:      acc.UserID,
			Balance:     acc.Balance,
			Equity:      acc.Equity,
			Margin:      acc.Margin,
			FreeMargin:  acc.FreeMargin,
			MarginLevel: acc.MarginLevel,
		}
	}
	return result
}

// GetAccount returns a specific account
func (a *CoreEngineAdapter) GetAccount(accountID int64) (*AccountInfo, error) {
	acc, ok := a.engine.GetAccount(accountID)
	if !ok {
		return nil, fmt.Errorf("account not found")
	}
	return &AccountInfo{
		ID:          acc.ID,
		UserID:      acc.UserID,
		Balance:     acc.Balance,
		Equity:      acc.Equity,
		Margin:      acc.Margin,
		FreeMargin:  acc.FreeMargin,
		MarginLevel: acc.MarginLevel,
	}, nil
}

// GetPositions returns positions for an account
func (a *CoreEngineAdapter) GetPositions(accountID int64) []*PositionInfo {
	positions := a.engine.GetPositions(accountID)
	result := make([]*PositionInfo, len(positions))
	for i, pos := range positions {
		result[i] = &PositionInfo{
			ID:            pos.ID,
			AccountID:     pos.AccountID,
			Symbol:        pos.Symbol,
			Side:          pos.Side,
			Volume:        pos.Volume,
			OpenPrice:     pos.OpenPrice,
			CurrentPrice:  pos.CurrentPrice,
			UnrealizedPnL: pos.UnrealizedPnL,
		}
	}
	return result
}

// ClosePosition closes a position
func (a *CoreEngineAdapter) ClosePosition(positionID int64, closePrice float64) error {
	// core.Engine ClosePosition takes (positionID, closeVolume) and returns (*Trade, error)
	// We need to get the position first to know the full volume
	positions := a.engine.GetAllPositions()
	var targetPos *core.Position
	for _, pos := range positions {
		if pos.ID == positionID {
			targetPos = pos
			break
		}
	}

	if targetPos == nil {
		return fmt.Errorf("position not found")
	}

	// Close the full volume
	_, err := a.engine.ClosePosition(positionID, targetPos.Volume)
	return err
}

// GetCurrentPrice returns current bid/ask price for a symbol
func (a *CoreEngineAdapter) GetCurrentPrice(symbol string) (bid, ask float64) {
	// The core.Engine doesn't have GetCurrentPrice, but it has a priceCallback
	// We need to call the price callback directly
	// For now, return zeros - this will need to be implemented properly
	return 0, 0
}

// StoreAlert stores a risk alert
func (a *CoreEngineAdapter) StoreAlert(alert *RiskAlert) error {
	// core.Engine doesn't have StoreAlert - just log it for now
	// In production, this would integrate with an alerting system
	return nil
}

// StoreLiquidationEvent stores a liquidation event
func (a *CoreEngineAdapter) StoreLiquidationEvent(event LiquidationEvent) {
	// core.Engine doesn't have StoreLiquidationEvent - just log it for now
	// In production, this would be stored in the database
}
