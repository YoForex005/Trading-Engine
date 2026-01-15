package bbook

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/govalues/decimal"

	decutil "github.com/epic1st/rtx/backend/internal/decimal"
	"github.com/epic1st/rtx/backend/internal/database/repository"
)

// CalculatePositionMargin calculates required margin for a single position.
// Standard forex/CFD formula: (volume * contract_size * open_price) / leverage
// This is the industry standard (MT5, all major brokers).
func CalculatePositionMargin(
	volume decimal.Decimal,
	contractSize decimal.Decimal,
	openPrice decimal.Decimal,
	leverage decimal.Decimal,
) decimal.Decimal {
	notionalValue, err := volume.Mul(contractSize)
	if err != nil {
		return decutil.Zero()
	}
	notionalValue, err = notionalValue.Mul(openPrice)
	if err != nil {
		return decutil.Zero()
	}
	margin, err := notionalValue.Quo(leverage)
	if err != nil {
		return decutil.Zero()
	}
	return margin
}

// CalculateUnrealizedPL calculates P&L for an open position.
func CalculateUnrealizedPL(
	side string, // "BUY" or "SELL"
	volume decimal.Decimal,
	openPrice decimal.Decimal,
	currentPrice decimal.Decimal,
	contractSize decimal.Decimal,
) decimal.Decimal {
	var priceDiff decimal.Decimal
	var err error
	if side == "BUY" {
		priceDiff, err = currentPrice.Sub(openPrice)
	} else { // SELL
		priceDiff, err = openPrice.Sub(currentPrice)
	}
	if err != nil {
		return decutil.Zero()
	}
	pl, err := priceDiff.Mul(volume)
	if err != nil {
		return decutil.Zero()
	}
	pl, err = pl.Mul(contractSize)
	if err != nil {
		return decutil.Zero()
	}
	return pl
}

// CalculateEquity computes equity as balance + unrealized P&L.
func CalculateEquity(balance decimal.Decimal, unrealizedPL decimal.Decimal) decimal.Decimal {
	equity, err := balance.Add(unrealizedPL)
	if err != nil {
		return balance // If error, return balance as fallback
	}
	return equity
}

// CalculateMarginLevel returns margin level percentage.
// Formula: (Equity / Used Margin) * 100
// Returns:
//   - 200% = healthy
//   - 100% = margin call threshold
//   - 50% = stop-out threshold
func CalculateMarginLevel(equity decimal.Decimal, usedMargin decimal.Decimal) decimal.Decimal {
	if decutil.IsZero(usedMargin) {
		return decutil.MustParse("99999.00") // Effectively infinite when no margin used
	}
	level, err := equity.Quo(usedMargin)
	if err != nil {
		return decutil.Zero()
	}
	level, err = level.Mul(decutil.NewFromInt64(100))
	if err != nil {
		return decutil.Zero()
	}
	return level
}

// CheckThresholds checks if margin call or stop-out is triggered.
// Returns (marginCall, stopOut) booleans.
func CheckThresholds(
	marginLevel decimal.Decimal,
	marginCallLevel decimal.Decimal,
	stopOutLevel decimal.Decimal,
) (marginCall bool, stopOut bool) {
	if marginLevel.Cmp(stopOutLevel) <= 0 {
		return true, true // Both triggered if at/below stop-out
	}
	if marginLevel.Cmp(marginCallLevel) <= 0 {
		return true, false // Only margin call if at/below call level
	}
	return false, false
}

// UpdateMarginState recalculates margin for account and persists to database.
// Called after EVERY position change (order fill, position close, tick update affecting P&L).
func (e *Engine) UpdateMarginState(ctx context.Context, accountID int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	account, ok := e.accounts[accountID]
	if !ok {
		return fmt.Errorf("account %d not found", accountID)
	}

	// 1. Calculate used margin across all open positions
	usedMargin := decutil.Zero()
	for _, pos := range e.positions {
		if pos.AccountID != accountID || pos.Status != "OPEN" {
			continue
		}

		// Get symbol config for leverage and contract size
		symbolConfig, err := e.symbolMarginConfigRepo.GetBySymbol(ctx, pos.Symbol)
		if err != nil {
			log.Printf("[Margin] Warning: symbol config not found for %s, using defaults", pos.Symbol)
			// Use defaults: 30:1 leverage, 100000 contract size (standard lot)
			symbolConfig = &repository.SymbolMarginConfig{
				MaxLeverage:  "30.00",
				ContractSize: "100000",
			}
		}

		posMargin := CalculatePositionMargin(
			decutil.MustParse(fmt.Sprintf("%.8f", pos.Volume)),
			decutil.MustParse(symbolConfig.ContractSize),
			decutil.MustParse(fmt.Sprintf("%.8f", pos.OpenPrice)),
			decutil.MustParse(symbolConfig.MaxLeverage),
		)
		newUsedMargin, err := usedMargin.Add(posMargin)
		if err != nil {
			log.Printf("[Margin] Warning: failed to add position margin: %v", err)
			continue
		}
		usedMargin = newUsedMargin
	}

	// 2. Calculate unrealized P&L across all open positions
	totalUnrealizedPL := decutil.Zero()
	for _, pos := range e.positions {
		if pos.AccountID != accountID || pos.Status != "OPEN" {
			continue
		}

		// Get symbol config for contract size
		symbolConfig, err := e.symbolMarginConfigRepo.GetBySymbol(ctx, pos.Symbol)
		contractSize := "100000" // Default standard lot
		if err == nil {
			contractSize = symbolConfig.ContractSize
		}

		pl := CalculateUnrealizedPL(
			pos.Side,
			decutil.MustParse(fmt.Sprintf("%.8f", pos.Volume)),
			decutil.MustParse(fmt.Sprintf("%.8f", pos.OpenPrice)),
			decutil.MustParse(fmt.Sprintf("%.8f", pos.CurrentPrice)),
			decutil.MustParse(contractSize),
		)
		newTotal, err := totalUnrealizedPL.Add(pl)
		if err != nil {
			log.Printf("[Margin] Warning: failed to add unrealized P&L: %v", err)
			continue
		}
		totalUnrealizedPL = newTotal
	}

	// 3. Calculate equity
	balance := decutil.MustParse(fmt.Sprintf("%.8f", account.Balance))
	equity := CalculateEquity(balance, totalUnrealizedPL)

	// 4. Calculate margin level
	marginLevel := CalculateMarginLevel(equity, usedMargin)

	// 5. Get risk limits and check thresholds
	limits, err := e.riskLimitRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		// Use regulatory defaults if not configured (ESMA standards)
		limits = &repository.RiskLimit{
			MarginCallLevel: "100.00", // 100% = equity equals used margin
			StopOutLevel:    "50.00",  // 50% = equity at half of used margin
		}
	}

	marginCall, stopOut := CheckThresholds(
		marginLevel,
		decutil.MustParse(limits.MarginCallLevel),
		decutil.MustParse(limits.StopOutLevel),
	)

	// 6. Prepare margin state for database
	freeMargin, err := equity.Sub(usedMargin)
	if err != nil {
		return fmt.Errorf("failed to calculate free margin: %w", err)
	}
	state := &repository.MarginState{
		AccountID:           accountID,
		Equity:              decutil.ToString(equity),
		UsedMargin:          decutil.ToString(usedMargin),
		FreeMargin:          decutil.ToString(freeMargin),
		MarginLevel:         decutil.ToStringFixed(marginLevel, 2),
		MarginCallTriggered: marginCall,
		StopOutTriggered:    stopOut,
		LastUpdated:         time.Now(),
	}

	// 7. Persist margin state to database
	if err := e.marginStateRepo.Upsert(ctx, state); err != nil {
		return fmt.Errorf("failed to update margin state: %w", err)
	}

	// 8. Log margin events
	if marginCall && !stopOut {
		log.Printf("[MARGIN CALL] Account %d: Margin level %.2f%% (threshold: %s%%)",
			accountID, marginLevel, limits.MarginCallLevel)
	}

	// 9. Execute stop-out if triggered
	if stopOut {
		log.Printf("[STOP OUT TRIGGERED] Account %d: Margin level %.2f%% (threshold: %s%%)",
			accountID, marginLevel, limits.StopOutLevel)

		// Unlock mutex before ExecuteStopOut (it needs to acquire lock)
		e.mu.Unlock()
		closedCount, err := ExecuteStopOut(
			ctx,
			e,
			accountID,
			marginLevel,
			decutil.MustParse(limits.StopOutLevel),
		)
		e.mu.Lock() // Re-acquire lock

		if err != nil {
			log.Printf("[STOP OUT ERROR] Account %d: %v (closed %d positions)", accountID, err, closedCount)
			// Don't return error - margin state already updated, stop-out attempted
		} else {
			log.Printf("[STOP OUT SUCCESS] Account %d: Closed %d positions", accountID, closedCount)
		}
	}

	return nil
}
