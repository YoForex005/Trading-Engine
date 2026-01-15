package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DailyAccountStats represents daily statistics for an account
// CRITICAL: All DECIMAL columns stored as strings to avoid float precision issues
type DailyAccountStats struct {
	AccountID              int64
	StatDate               time.Time
	StartingBalance        string // DECIMAL(20,8) as string
	EndingBalance          string // DECIMAL(20,8) as string
	RealizedPL             string // DECIMAL(20,8) as string
	UnrealizedPL           string // DECIMAL(20,8) as string
	TotalPL                string // DECIMAL(20,8) as string
	HighWaterMark          string // DECIMAL(20,8) as string
	CurrentDrawdownPct     string // DECIMAL(5,2) as string
	MaxDrawdownPct         string // DECIMAL(5,2) as string
	TradesOpened           int
	TradesClosed           int
	WinningTrades          int
	LosingTrades           int
	DailyLossLimitBreached bool
	DrawdownLimitBreached  bool
	AccountDisabledAt      *time.Time
	UpdatedAt              time.Time
}

type DailyStatsRepository struct {
	pool *pgxpool.Pool
}

func NewDailyStatsRepository(pool *pgxpool.Pool) *DailyStatsRepository {
	return &DailyStatsRepository{pool: pool}
}

// GetByAccountAndDate retrieves daily stats for a specific account and date
func (r *DailyStatsRepository) GetByAccountAndDate(ctx context.Context, accountID int64, date time.Time) (*DailyAccountStats, error) {
	query := `
		SELECT account_id, stat_date, starting_balance, ending_balance,
		       realized_pl, unrealized_pl, total_pl, high_water_mark,
		       current_drawdown_pct, max_drawdown_pct, trades_opened, trades_closed,
		       winning_trades, losing_trades, daily_loss_limit_breached,
		       drawdown_limit_breached, account_disabled_at, updated_at
		FROM daily_account_stats
		WHERE account_id = $1 AND stat_date = $2
	`

	var stats DailyAccountStats
	err := r.pool.QueryRow(ctx, query, accountID, date).Scan(
		&stats.AccountID, &stats.StatDate, &stats.StartingBalance, &stats.EndingBalance,
		&stats.RealizedPL, &stats.UnrealizedPL, &stats.TotalPL, &stats.HighWaterMark,
		&stats.CurrentDrawdownPct, &stats.MaxDrawdownPct, &stats.TradesOpened, &stats.TradesClosed,
		&stats.WinningTrades, &stats.LosingTrades, &stats.DailyLossLimitBreached,
		&stats.DrawdownLimitBreached, &stats.AccountDisabledAt, &stats.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, pgx.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// Upsert inserts or updates daily statistics
func (r *DailyStatsRepository) Upsert(ctx context.Context, stats *DailyAccountStats) error {
	query := `
		INSERT INTO daily_account_stats (
			account_id, stat_date, starting_balance, ending_balance,
			realized_pl, unrealized_pl, total_pl, high_water_mark,
			current_drawdown_pct, max_drawdown_pct, trades_opened, trades_closed,
			winning_trades, losing_trades, daily_loss_limit_breached,
			drawdown_limit_breached, account_disabled_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW())
		ON CONFLICT (account_id, stat_date)
		DO UPDATE SET
			ending_balance = EXCLUDED.ending_balance,
			realized_pl = EXCLUDED.realized_pl,
			unrealized_pl = EXCLUDED.unrealized_pl,
			total_pl = EXCLUDED.total_pl,
			high_water_mark = EXCLUDED.high_water_mark,
			current_drawdown_pct = EXCLUDED.current_drawdown_pct,
			max_drawdown_pct = EXCLUDED.max_drawdown_pct,
			trades_opened = EXCLUDED.trades_opened,
			trades_closed = EXCLUDED.trades_closed,
			winning_trades = EXCLUDED.winning_trades,
			losing_trades = EXCLUDED.losing_trades,
			daily_loss_limit_breached = EXCLUDED.daily_loss_limit_breached,
			drawdown_limit_breached = EXCLUDED.drawdown_limit_breached,
			account_disabled_at = EXCLUDED.account_disabled_at,
			updated_at = NOW()
	`

	_, err := r.pool.Exec(ctx, query,
		stats.AccountID, stats.StatDate, stats.StartingBalance, stats.EndingBalance,
		stats.RealizedPL, stats.UnrealizedPL, stats.TotalPL, stats.HighWaterMark,
		stats.CurrentDrawdownPct, stats.MaxDrawdownPct, stats.TradesOpened, stats.TradesClosed,
		stats.WinningTrades, stats.LosingTrades, stats.DailyLossLimitBreached,
		stats.DrawdownLimitBreached, stats.AccountDisabledAt,
	)

	return err
}

// GetHistoricalHighWaterMark retrieves the highest high-water mark ever recorded for an account
func (r *DailyStatsRepository) GetHistoricalHighWaterMark(ctx context.Context, accountID int64) (string, error) {
	query := `
		SELECT COALESCE(MAX(high_water_mark), '0')
		FROM daily_account_stats
		WHERE account_id = $1
	`

	var highWaterMark string
	err := r.pool.QueryRow(ctx, query, accountID).Scan(&highWaterMark)
	return highWaterMark, err
}
