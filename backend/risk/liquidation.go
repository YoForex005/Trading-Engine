package risk

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"
)

// LiquidationEngine handles stop-out and auto-liquidation
type LiquidationEngine struct {
	engine            *Engine
	monitorInterval   time.Duration
	stopChan          chan struct{}
	liquidationChan   chan *LiquidationEvent
	onLiquidation     func(*LiquidationEvent)
}

// NewLiquidationEngine creates a new liquidation engine
func NewLiquidationEngine(engine *Engine) *LiquidationEngine {
	return &LiquidationEngine{
		engine:           engine,
		monitorInterval:  time.Second, // Check every second
		stopChan:         make(chan struct{}),
		liquidationChan:  make(chan *LiquidationEvent, 100),
	}
}

// Start begins monitoring for liquidation conditions
func (l *LiquidationEngine) Start() {
	go l.monitorLoop()
	log.Println("[Liquidation] Engine started")
}

// Stop stops the liquidation engine
func (l *LiquidationEngine) Stop() {
	close(l.stopChan)
	log.Println("[Liquidation] Engine stopped")
}

// SetLiquidationCallback sets callback for liquidation events
func (l *LiquidationEngine) SetLiquidationCallback(fn func(*LiquidationEvent)) {
	l.onLiquidation = fn
}

// monitorLoop continuously monitors accounts for liquidation conditions
func (l *LiquidationEngine) monitorLoop() {
	ticker := time.NewTicker(l.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-l.stopChan:
			return
		case <-ticker.C:
			l.checkAllAccounts()
		}
	}
}

// checkAllAccounts checks all accounts for liquidation conditions
func (l *LiquidationEngine) checkAllAccounts() {
	accounts := l.engine.GetAllAccounts()

	for _, account := range accounts {
		// Skip accounts with no positions
		positions := l.engine.GetAllPositions(account.ID)
		if len(positions) == 0 {
			continue
		}

		clientProfile := l.engine.GetClientRiskProfile(account.ClientID)
		if clientProfile == nil {
			continue
		}

		// Check if margin level is below stop-out level
		if account.MarginLevel > 0 && account.MarginLevel < clientProfile.StopOutLevel {
			log.Printf("[Liquidation] STOP-OUT triggered for account #%d: margin level %.2f%% < %.2f%%",
				account.ID, account.MarginLevel, clientProfile.StopOutLevel)

			event := l.liquidateAccount(account.ID, "STOP_OUT", clientProfile.StopOutLevel)
			if event != nil {
				l.liquidationChan <- event
				if l.onLiquidation != nil {
					l.onLiquidation(event)
				}
			}
		}

		// Check daily loss limit
		if clientProfile.DailyLossLimit > 0 {
			dailyPnL := l.engine.GetDailyPnL(account.ID)
			if dailyPnL < 0 && math.Abs(dailyPnL) >= clientProfile.DailyLossLimit {
				log.Printf("[Liquidation] DAILY LOSS LIMIT triggered for account #%d: loss %.2f >= %.2f",
					account.ID, math.Abs(dailyPnL), clientProfile.DailyLossLimit)

				event := l.liquidateAccount(account.ID, "DAILY_LOSS_LIMIT", clientProfile.DailyLossLimit)
				if event != nil {
					l.liquidationChan <- event
					if l.onLiquidation != nil {
						l.onLiquidation(event)
					}
				}
			}
		}

		// Check max drawdown
		if clientProfile.MaxDrawdownPercent > 0 {
			peakEquity := l.engine.GetPeakEquity(account.ID)
			if peakEquity > 0 {
				drawdownPercent := ((peakEquity - account.Equity) / peakEquity) * 100
				if drawdownPercent >= clientProfile.MaxDrawdownPercent {
					log.Printf("[Liquidation] MAX DRAWDOWN triggered for account #%d: %.2f%% >= %.2f%%",
						account.ID, drawdownPercent, clientProfile.MaxDrawdownPercent)

					event := l.liquidateAccount(account.ID, "MAX_DRAWDOWN", clientProfile.MaxDrawdownPercent)
					if event != nil {
						l.liquidationChan <- event
						if l.onLiquidation != nil {
							l.onLiquidation(event)
						}
					}
				}
			}
		}
	}
}

// liquidateAccount performs auto-liquidation of an account
func (l *LiquidationEngine) liquidateAccount(
	accountID int64,
	reason string,
	threshold float64,
) *LiquidationEvent {

	account, err := l.engine.GetAccountByID(accountID)
	if err != nil {
		log.Printf("[Liquidation] Error getting account #%d: %v", accountID, err)
		return nil
	}

	marginLevelBefore := account.MarginLevel

	// Get all positions sorted by priority
	positions := l.getSortedPositionsForLiquidation(accountID, LiquidationLargestLoss)

	liquidatedPositions := make([]LiquidatedPosition, 0)
	totalPnL := 0.0
	totalSlippage := 0.0

	// Close positions until margin level is restored or all positions are closed
	for _, pos := range positions {
		// Attempt to close position
		closePrice := l.getClosePrice(*pos)
		slippage := l.calculateSlippage(*pos, closePrice)

		pnl := l.calculatePositionPnL(*pos, closePrice)

		liquidatedPos := LiquidatedPosition{
			PositionID: pos.ID,
			Symbol:     pos.Symbol,
			Volume:     pos.Volume,
			OpenPrice:  pos.OpenPrice,
			ClosePrice: closePrice,
			PnL:        pnl,
			Slippage:   slippage,
		}
		liquidatedPositions = append(liquidatedPositions, liquidatedPos)

		totalPnL += pnl
		totalSlippage += slippage

		// Close the position in the engine
		err := l.engine.ClosePosition(pos.ID, closePrice, "LIQUIDATION")
		if err != nil {
			log.Printf("[Liquidation] Error closing position %d: %v", pos.ID, err)
			continue
		}

		log.Printf("[Liquidation] Closed position %d: %s %.2f lots @ %.5f, PnL: %.2f, Slippage: %.2f",
			pos.ID, pos.Symbol, pos.Volume, closePrice, pnl, slippage)

		// Check if margin is restored
		account, _ = l.engine.GetAccountByID(accountID)
		clientProfile := l.engine.GetClientRiskProfile(account.ClientID)
		if clientProfile != nil && account.MarginLevel >= clientProfile.MarginCallLevel*1.5 {
			// Margin restored to 150% of margin call level - stop liquidating
			log.Printf("[Liquidation] Margin restored for account #%d: %.2f%%",
				accountID, account.MarginLevel)
			break
		}
	}

	event := &LiquidationEvent{
		ID:                fmt.Sprintf("LIQ_%d_%d", accountID, time.Now().Unix()),
		AccountID:         accountID,
		TriggerReason:     reason,
		MarginLevelBefore: marginLevelBefore,
		PositionsClosed:   liquidatedPositions,
		TotalPnL:          totalPnL,
		Slippage:          totalSlippage,
		ExecutedAt:        time.Now(),
	}

	// Store liquidation event
	l.engine.StoreLiquidationEvent(*event)

	// Create risk alert
	alert := &RiskAlert{
		ID:        fmt.Sprintf("ALERT_LIQ_%d", time.Now().Unix()),
		AccountID: accountID,
		AlertType: "LIQUIDATION",
		Severity:  RiskLevelCritical,
		Message: fmt.Sprintf("Account liquidated: %s (threshold: %.2f). Closed %d positions, total PnL: %.2f",
			reason, threshold, len(liquidatedPositions), totalPnL),
		Data: map[string]interface{}{
			"reason":             reason,
			"positions_closed":   len(liquidatedPositions),
			"total_pnl":          totalPnL,
			"margin_level_before": marginLevelBefore,
		},
		CreatedAt: time.Now(),
	}
	l.engine.StoreAlert(alert)

	log.Printf("[Liquidation] Account #%d liquidated: %s, closed %d positions, PnL: %.2f",
		accountID, reason, len(liquidatedPositions), totalPnL)

	return event
}

// getSortedPositionsForLiquidation returns positions sorted by liquidation priority
func (l *LiquidationEngine) getSortedPositionsForLiquidation(
	accountID int64,
	priority LiquidationPriority,
) []*Position {

	positions := l.engine.GetAllPositions(accountID)

	switch priority {
	case LiquidationLargestLoss:
		// Sort by unrealized P/L (most negative first)
		sort.Slice(positions, func(i, j int) bool {
			return positions[i].UnrealizedPnL < positions[j].UnrealizedPnL
		})

	case LiquidationHighestMargin:
		// Sort by margin requirement (highest first)
		sort.Slice(positions, func(i, j int) bool {
			marginI := l.getPositionMargin(positions[i])
			marginJ := l.getPositionMargin(positions[j])
			return marginI > marginJ
		})

	case LiquidationOldestPosition:
		// Sort by open time (oldest first)
		sort.Slice(positions, func(i, j int) bool {
			return positions[i].OpenTime.Before(positions[j].OpenTime)
		})

	case LiquidationLowestProfit:
		// Sort by unrealized P/L (lowest profit/highest loss first)
		sort.Slice(positions, func(i, j int) bool {
			return positions[i].UnrealizedPnL < positions[j].UnrealizedPnL
		})
	}

	return positions
}

// getClosePrice determines the close price with slippage
func (l *LiquidationEngine) getClosePrice(pos Position) float64 {
	// Get current bid/ask
	bid, ask := l.engine.GetCurrentPrice(pos.Symbol)

	// Liquidation uses worse price due to urgency
	var basePrice float64
	if pos.Side == "BUY" {
		basePrice = bid // Close long at bid
	} else {
		basePrice = ask // Close short at ask
	}

	// Apply slippage (0.1% - 0.5% depending on volatility and size)
	slippagePercent := l.calculateSlippagePercent(pos)
	slippage := basePrice * slippagePercent / 100.0

	if pos.Side == "BUY" {
		return basePrice - slippage // Worse price for long
	}
	return basePrice + slippage // Worse price for short
}

// calculateSlippagePercent estimates slippage percentage
func (l *LiquidationEngine) calculateSlippagePercent(pos Position) float64 {
	// Base slippage: 0.1%
	slippage := 0.1

	// Add volatility component
	volatility := l.engine.GetVolatility(pos.Symbol)
	slippage += volatility * 0.5 // Up to 0.075% more for 15% volatility

	// Add size component (larger positions = more slippage)
	if pos.Volume > 10 {
		slippage += math.Min((pos.Volume-10)*0.01, 0.3) // Up to 0.3% more
	}

	// Market hours vs off-hours
	if !l.engine.IsMarketOpen(pos.Symbol) {
		slippage += 0.2 // 0.2% more slippage off-hours
	}

	return math.Min(slippage, 0.5) // Cap at 0.5%
}

// calculateSlippage calculates actual slippage amount
func (l *LiquidationEngine) calculateSlippage(pos Position, closePrice float64) float64 {
	expectedPrice := pos.CurrentPrice
	if pos.Side == "BUY" {
		bid, _ := l.engine.GetCurrentPrice(pos.Symbol)
		expectedPrice = bid
	} else {
		_, ask := l.engine.GetCurrentPrice(pos.Symbol)
		expectedPrice = ask
	}

	priceDiff := math.Abs(closePrice - expectedPrice)
	contractSize := l.getContractSize(pos.Symbol)
	slippage := priceDiff * pos.Volume * contractSize

	return slippage
}

// calculatePositionPnL calculates P/L for a position
func (l *LiquidationEngine) calculatePositionPnL(pos Position, closePrice float64) float64 {
	priceDiff := closePrice - pos.OpenPrice
	if pos.Side == "SELL" {
		priceDiff = -priceDiff
	}

	contractSize := l.getContractSize(pos.Symbol)
	pnl := priceDiff * pos.Volume * contractSize

	return pnl
}

// getPositionMargin gets margin requirement for a position
func (l *LiquidationEngine) getPositionMargin(pos *Position) float64 {
	_, err := l.engine.GetAccountByID(pos.AccountID)
	if err != nil {
		return 0
	}

	calculator := NewMarginCalculator(l.engine)
	margin, err := calculator.CalculateMargin(
		pos.AccountID,
		pos.Symbol,
		pos.Volume,
		pos.CurrentPrice,
		MarginRetail,
	)
	if err != nil {
		return 0
	}

	return margin
}

// getContractSize returns contract size for an instrument
func (l *LiquidationEngine) getContractSize(symbol string) float64 {
	switch symbol {
	case "XAUUSD":
		return 100.0
	case "XAGUSD":
		return 5000.0
	case "BTCUSD", "ETHUSD":
		return 1.0
	default:
		if len(symbol) == 6 {
			return 100000.0
		}
		return 1.0
	}
}

// ManualLiquidation allows admin to manually trigger liquidation
func (l *LiquidationEngine) ManualLiquidation(
	accountID int64,
	reason string,
	adminID string,
) (*LiquidationEvent, error) {

	log.Printf("[Liquidation] Manual liquidation triggered by %s for account #%d: %s",
		adminID, accountID, reason)

	event := l.liquidateAccount(accountID, fmt.Sprintf("MANUAL: %s", reason), 0)
	if event == nil {
		return nil, fmt.Errorf("liquidation failed")
	}

	return event, nil
}

// GetLiquidationHistory returns liquidation events for an account
func (l *LiquidationEngine) GetLiquidationHistory(accountID int64, limit int) []LiquidationEvent {
	return l.engine.GetLiquidationEvents(accountID)
}
