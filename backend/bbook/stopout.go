package bbook

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/govalues/decimal"

	decutil "github.com/epic1st/rtx/backend/internal/decimal"
)

// SelectPositionsForLiquidation returns positions sorted by unrealized P&L (most losing first).
// Liquidating worst performers first maximizes account recovery chance.
func SelectPositionsForLiquidation(positions []*Position, currentPrices map[string]decimal.Decimal) []*Position {
	type positionWithPL struct {
		position *Position
		pl       decimal.Decimal
	}

	// Calculate P&L for each open position
	positionsWithPL := []positionWithPL{}
	for _, pos := range positions {
		if pos.Status != "OPEN" {
			continue
		}

		currentPrice, ok := currentPrices[pos.Symbol]
		if !ok {
			log.Printf("[StopOut] Warning: no current price for %s, skipping", pos.Symbol)
			continue
		}

		pl := CalculateUnrealizedPL(
			pos.Side,
			decutil.MustParse(fmt.Sprintf("%.8f", pos.Volume)),
			decutil.MustParse(fmt.Sprintf("%.8f", pos.OpenPrice)),
			currentPrice,
			decutil.MustParse("100000"), // TODO: get from symbol config
		)

		positionsWithPL = append(positionsWithPL, positionWithPL{
			position: pos,
			pl:       pl,
		})
	}

	// Sort by P&L ascending (most negative first)
	sort.Slice(positionsWithPL, func(i, j int) bool {
		return positionsWithPL[i].pl.Cmp(positionsWithPL[j].pl) < 0
	})

	// Extract sorted positions
	result := make([]*Position, len(positionsWithPL))
	for i, pwpl := range positionsWithPL {
		result[i] = pwpl.position
	}

	return result
}

// ExecuteStopOut closes positions when margin level drops below stop-out threshold.
// Closes most losing positions first until margin level recovers above threshold.
// Returns number of positions closed and any error.
func ExecuteStopOut(
	ctx context.Context,
	engine *Engine,
	accountID int64,
	currentMarginLevel decimal.Decimal,
	stopOutLevel decimal.Decimal,
) (int, error) {
	account, ok := engine.accounts[accountID]
	if !ok {
		return 0, fmt.Errorf("account %d not found", accountID)
	}

	// Collect all open positions for this account
	var accountPositions []*Position
	for _, pos := range engine.positions {
		if pos.AccountID == accountID && pos.Status == "OPEN" {
			accountPositions = append(accountPositions, pos)
		}
	}

	// Get current prices for all symbols
	currentPrices := make(map[string]decimal.Decimal)
	for symbol := range engine.symbols {
		// Get current price from price callback
		if engine.priceCallback != nil {
			bid, _, ok := engine.priceCallback(symbol)
			if ok {
				// Use bid for BUY positions (conservative estimate)
				currentPrices[symbol] = decutil.NewFromFloat64(bid)
			}
		}
	}

	// Select positions for liquidation (worst performers first)
	positionsToClose := SelectPositionsForLiquidation(accountPositions, currentPrices)

	log.Printf("[STOP OUT] Account %d: Margin level %.2f%% below threshold %.2f%%. Liquidating %d positions...",
		accountID, currentMarginLevel, stopOutLevel, len(positionsToClose))

	closedCount := 0

	// Close positions one by one until margin level recovers
	for _, pos := range positionsToClose {
		// Close position (ClosePosition gets current price from priceCallback)
		// Pass 0 as volume to close entire position
		if _, err := engine.ClosePosition(pos.ID, 0); err != nil {
			log.Printf("[StopOut] Failed to close position %d: %v", pos.ID, err)
			continue
		}

		closedCount++

		// Recalculate margin after each closure
		if err := engine.UpdateMarginState(ctx, accountID); err != nil {
			log.Printf("[StopOut] Failed to recalculate margin: %v", err)
			// Continue closing positions despite error
		}

		// Check if margin level recovered
		marginState, err := engine.marginStateRepo.GetByAccountID(ctx, accountID)
		if err != nil {
			log.Printf("[StopOut] Failed to get margin state: %v", err)
			continue
		}

		newMarginLevel := decutil.MustParse(marginState.MarginLevel)
		if newMarginLevel.Cmp(stopOutLevel) > 0 {
			log.Printf("[STOP OUT COMPLETE] Account %d: Margin level recovered to %.2f%% after closing %d positions",
				accountID, newMarginLevel, closedCount)
			break
		}
	}

	if closedCount == 0 {
		return 0, fmt.Errorf("failed to close any positions during stop-out")
	}

	// Final margin check
	marginState, err := engine.marginStateRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return closedCount, fmt.Errorf("stop-out completed but margin state check failed: %w", err)
	}

	finalMarginLevel := decutil.MustParse(marginState.MarginLevel)
	if finalMarginLevel.Cmp(stopOutLevel) <= 0 {
		log.Printf("[STOP OUT WARNING] Account %d: Closed all %d positions but margin level still %.2f%% (threshold: %.2f%%)",
			accountID, closedCount, finalMarginLevel, stopOutLevel)
	}

	return closedCount, nil
}
