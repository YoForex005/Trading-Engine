package bbook

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"

	decutil "github.com/epic1st/rtx/backend/internal/decimal"
	"github.com/epic1st/rtx/backend/internal/database/repository"
)

// CheckDailyLossLimit validates if opening this order would breach daily loss limit.
// Returns error if limit would be breached, nil otherwise.
func CheckDailyLossLimit(
	ctx context.Context,
	accountID int64,
	currentBalance decimal.Decimal,
	dailyStatsRepo *repository.DailyStatsRepository,
	riskLimitRepo *repository.RiskLimitRepository,
) error {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Get today's stats
	stats, err := dailyStatsRepo.GetByAccountAndDate(ctx, accountID, today)
	if err == pgx.ErrNoRows {
		// No stats for today yet - first trade of the day
		// Starting balance = current balance, no loss yet
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get daily stats: %w", err)
	}

	// Check if already breached
	if stats.DailyLossLimitBreached {
		return fmt.Errorf("daily loss limit already breached - account trading disabled for today")
	}

	// Calculate current daily P&L
	startingBalance := decutil.MustParse(stats.StartingBalance)
	dailyPL := decutil.Sub(currentBalance, startingBalance)

	// Get account limits
	limits, err := riskLimitRepo.GetByAccountID(ctx, accountID)
	if err != nil || limits.DailyLossLimit == nil {
		// No daily loss limit configured
		return nil
	}

	dailyLossLimit := decutil.MustParse(*limits.DailyLossLimit)

	// Check if loss exceeds limit (dailyPL is negative for losses)
	if dailyPL.Cmp(decutil.Zero()) < 0 {
		loss := dailyPL.Neg() // Convert to positive value
		if loss.Cmp(dailyLossLimit) >= 0 {
			return fmt.Errorf(
				"daily loss limit reached: loss $%s exceeds limit $%s",
				decutil.ToStringFixed(loss, 2),
				decutil.ToStringFixed(dailyLossLimit, 2),
			)
		}
	}

	return nil
}

// CheckDrawdownLimit validates if current drawdown from high-water mark exceeds limit.
func CheckDrawdownLimit(
	ctx context.Context,
	accountID int64,
	currentBalance decimal.Decimal,
	dailyStatsRepo *repository.DailyStatsRepository,
	riskLimitRepo *repository.RiskLimitRepository,
) error {
	// Get historical high-water mark
	highWaterMarkStr, err := dailyStatsRepo.GetHistoricalHighWaterMark(ctx, accountID)
	if err != nil {
		log.Printf("[Drawdown] Failed to get high-water mark: %v", err)
		// Use current balance as high-water mark if no history
		highWaterMarkStr = decutil.ToString(currentBalance)
	}

	highWaterMark := decutil.MustParse(highWaterMarkStr)

	// Update high-water mark if current balance higher
	if currentBalance.Cmp(highWaterMark) > 0 {
		highWaterMark = currentBalance
	}

	// Calculate drawdown percentage
	drawdown := decutil.Sub(highWaterMark, currentBalance)
	drawdownPct := decutil.Zero()
	if !decutil.IsZero(highWaterMark) {
		drawdownPct = decutil.Mul(
			decutil.Div(drawdown, highWaterMark),
			decutil.NewFromInt64(100),
		)
	}

	// Get account limits
	limits, err := riskLimitRepo.GetByAccountID(ctx, accountID)
	if err != nil || limits.MaxDrawdownPct == nil {
		// No drawdown limit configured
		return nil
	}

	maxDrawdownPct := decutil.MustParse(*limits.MaxDrawdownPct)

	// Check if drawdown exceeds limit
	if drawdownPct.Cmp(maxDrawdownPct) >= 0 {
		return fmt.Errorf(
			"maximum drawdown exceeded: %s%% from peak $%s (limit: %s%%)",
			decutil.ToStringFixed(drawdownPct, 2),
			decutil.ToStringFixed(highWaterMark, 2),
			decutil.ToStringFixed(maxDrawdownPct, 2),
		)
	}

	return nil
}

// UpdateDailyStats updates daily statistics after position close or account change.
func UpdateDailyStats(
	ctx context.Context,
	accountID int64,
	currentBalance decimal.Decimal,
	realizedPL decimal.Decimal,
	unrealizedPL decimal.Decimal,
	tradeClosed bool,
	tradeWon bool,
	dailyStatsRepo *repository.DailyStatsRepository,
	riskLimitRepo *repository.RiskLimitRepository,
) error {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Get or create today's stats
	stats, err := dailyStatsRepo.GetByAccountAndDate(ctx, accountID, today)
	if err == pgx.ErrNoRows {
		// First trade of the day - initialize
		stats = &repository.DailyAccountStats{
			AccountID:       accountID,
			StatDate:        today,
			StartingBalance: decutil.ToString(decutil.Sub(currentBalance, realizedPL)),
			EndingBalance:   decutil.ToString(currentBalance),
			RealizedPL:      "0",
			UnrealizedPL:    "0",
			TotalPL:         "0",
			HighWaterMark:   decutil.ToString(currentBalance),
			TradesOpened:    0,
			TradesClosed:    0,
			WinningTrades:   0,
			LosingTrades:    0,
		}
	} else if err != nil {
		return fmt.Errorf("failed to get daily stats: %w", err)
	}

	// Update ending balance
	stats.EndingBalance = decutil.ToString(currentBalance)

	// Update P&L
	currentRealizedPL := decutil.MustParse(stats.RealizedPL)
	stats.RealizedPL = decutil.ToString(decutil.Add(currentRealizedPL, realizedPL))
	stats.UnrealizedPL = decutil.ToString(unrealizedPL)
	stats.TotalPL = decutil.ToString(
		decutil.Add(
			decutil.Add(currentRealizedPL, realizedPL),
			unrealizedPL,
		),
	)

	// Update trade counts
	if tradeClosed {
		stats.TradesClosed++
		if tradeWon {
			stats.WinningTrades++
		} else {
			stats.LosingTrades++
		}
	}

	// Update high-water mark
	highWaterMark := decutil.MustParse(stats.HighWaterMark)
	if currentBalance.Cmp(highWaterMark) > 0 {
		stats.HighWaterMark = decutil.ToString(currentBalance)
		highWaterMark = currentBalance
	}

	// Calculate drawdown
	drawdown := decutil.Sub(highWaterMark, currentBalance)
	drawdownPct := decutil.Zero()
	if !decutil.IsZero(highWaterMark) {
		drawdownPct = decutil.Mul(
			decutil.Div(drawdown, highWaterMark),
			decutil.NewFromInt64(100),
		)
	}
	stats.CurrentDrawdownPct = decutil.ToStringFixed(drawdownPct, 2)

	// Update max drawdown
	maxDrawdown := decutil.MustParse(stats.MaxDrawdownPct)
	if drawdownPct.Cmp(maxDrawdown) > 0 {
		stats.MaxDrawdownPct = decutil.ToStringFixed(drawdownPct, 2)
	}

	// Check limits
	if err := CheckDailyLossLimit(ctx, accountID, currentBalance, dailyStatsRepo, riskLimitRepo); err != nil {
		stats.DailyLossLimitBreached = true
		now := time.Now()
		stats.AccountDisabledAt = &now
		log.Printf("[DAILY LOSS LIMIT BREACH] Account %d: %v", accountID, err)
	}

	if err := CheckDrawdownLimit(ctx, accountID, currentBalance, dailyStatsRepo, riskLimitRepo); err != nil {
		stats.DrawdownLimitBreached = true
		now := time.Now()
		stats.AccountDisabledAt = &now
		log.Printf("[DRAWDOWN LIMIT BREACH] Account %d: %v", accountID, err)
	}

	// Persist stats
	return dailyStatsRepo.Upsert(ctx, stats)
}
