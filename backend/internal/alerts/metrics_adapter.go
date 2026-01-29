package alerts

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
)

// BBookMetricsAdapter adapts the B-Book engine to provide account metrics
type BBookMetricsAdapter struct {
	engine    *core.Engine
	pnlEngine *core.PnLEngine
}

// NewBBookMetricsAdapter creates a metrics adapter for the B-Book engine
func NewBBookMetricsAdapter(engine *core.Engine, pnlEngine *core.PnLEngine) *BBookMetricsAdapter {
	return &BBookMetricsAdapter{
		engine:    engine,
		pnlEngine: pnlEngine,
	}
}

// GetSnapshot retrieves current account metrics for alert evaluation
func (a *BBookMetricsAdapter) GetSnapshot(accountID string) (*MetricSnapshot, error) {
	// Convert string accountID to int64 (or use accountID directly if it's a userID)
	// For now, try to get account by userID first
	accounts := a.engine.GetAccountByUser(accountID)
	if len(accounts) == 0 {
		// Try parsing as int64 ID
		id, err := strconv.ParseInt(accountID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("account not found: %s", accountID)
		}

		// Try getting by numeric ID
		account, exists := a.engine.GetAccount(id)
		if !exists {
			return nil, fmt.Errorf("account not found: %s", accountID)
		}
		accounts = []*core.Account{account}
	}

	account := accounts[0] // Use first account for this user

	// Use GetAccountSummary which calculates all metrics
	summary, err := a.engine.GetAccountSummary(account.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account summary: %v", err)
	}

	// Get positions count
	positions := a.engine.GetPositions(account.ID)

	// Convert core.AccountSummary to alerts.MetricSnapshot
	snapshot := &MetricSnapshot{
		AccountID:       accountID,
		Timestamp:       time.Now(),
		Balance:         summary.Balance,
		Equity:          summary.Equity,
		Margin:          summary.Margin,
		FreeMargin:      summary.FreeMargin,
		MarginLevel:     summary.MarginLevel,
		ExposurePercent: 0.0, // Will calculate below
		PositionCount:   len(positions),
		PnL:             summary.UnrealizedPnL,
	}

	// Calculate exposure percentage (used margin / equity)
	if snapshot.Equity > 0 {
		snapshot.ExposurePercent = (snapshot.Margin / snapshot.Equity) * 100
	}

	// Log metrics for debugging (can be disabled in production)
	if len(positions) > 0 {
		log.Printf("[MetricsAdapter] Snapshot for %s: Balance=%.2f, Equity=%.2f, Margin=%.2f, MarginLevel=%.2f%%, Exposure=%.2f%%",
			accountID, snapshot.Balance, snapshot.Equity, snapshot.Margin, snapshot.MarginLevel, snapshot.ExposurePercent)
	}

	return snapshot, nil
}
