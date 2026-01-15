package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RiskLimit represents risk limits per account or account group
// CRITICAL: All DECIMAL columns stored as strings to avoid float precision issues
type RiskLimit struct {
	ID                   int64
	AccountID            *int64  // Pointer for NULL (group limits)
	AccountGroup         *string // Pointer for NULL (account-specific)
	MaxLeverage          string  // DECIMAL(5,2) as string
	MaxOpenPositions     int
	MaxPositionSizeLots  *string // DECIMAL(10,2) as string, optional
	DailyLossLimit       *string // DECIMAL(20,8) as string, optional
	MaxDrawdownPct       *string // DECIMAL(5,2) as string, optional
	MarginCallLevel      string  // DECIMAL(5,2) as string, default 100.00
	StopOutLevel         string  // DECIMAL(5,2) as string, default 50.00
	MaxSymbolExposurePct *string // DECIMAL(5,2) as string, max exposure to single symbol as % of equity
	MaxTotalExposurePct  *string // DECIMAL(6,2) as string, max total notional exposure as % of equity
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type RiskLimitRepository struct {
	pool *pgxpool.Pool
}

func NewRiskLimitRepository(pool *pgxpool.Pool) *RiskLimitRepository {
	return &RiskLimitRepository{pool: pool}
}

// GetByAccountID retrieves account-specific risk limits
func (r *RiskLimitRepository) GetByAccountID(ctx context.Context, accountID int64) (*RiskLimit, error) {
	query := `
		SELECT id, account_id, account_group, max_leverage, max_open_positions,
		       max_position_size_lots, daily_loss_limit, max_drawdown_pct,
		       margin_call_level, stop_out_level, max_symbol_exposure_pct,
		       max_total_exposure_pct, created_at, updated_at
		FROM risk_limits
		WHERE account_id = $1
	`

	var limit RiskLimit
	err := r.pool.QueryRow(ctx, query, accountID).Scan(
		&limit.ID,
		&limit.AccountID,
		&limit.AccountGroup,
		&limit.MaxLeverage,
		&limit.MaxOpenPositions,
		&limit.MaxPositionSizeLots,
		&limit.DailyLossLimit,
		&limit.MaxDrawdownPct,
		&limit.MarginCallLevel,
		&limit.StopOutLevel,
		&limit.MaxSymbolExposurePct,
		&limit.MaxTotalExposurePct,
		&limit.CreatedAt,
		&limit.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("risk limits not found for account: %d", accountID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get risk limits: %w", err)
	}

	return &limit, nil
}

// GetByAccountGroup retrieves group-level risk limits
func (r *RiskLimitRepository) GetByAccountGroup(ctx context.Context, accountGroup string) (*RiskLimit, error) {
	query := `
		SELECT id, account_id, account_group, max_leverage, max_open_positions,
		       max_position_size_lots, daily_loss_limit, max_drawdown_pct,
		       margin_call_level, stop_out_level, max_symbol_exposure_pct,
		       max_total_exposure_pct, created_at, updated_at
		FROM risk_limits
		WHERE account_group = $1 AND account_id IS NULL
	`

	var limit RiskLimit
	err := r.pool.QueryRow(ctx, query, accountGroup).Scan(
		&limit.ID,
		&limit.AccountID,
		&limit.AccountGroup,
		&limit.MaxLeverage,
		&limit.MaxOpenPositions,
		&limit.MaxPositionSizeLots,
		&limit.DailyLossLimit,
		&limit.MaxDrawdownPct,
		&limit.MarginCallLevel,
		&limit.StopOutLevel,
		&limit.MaxSymbolExposurePct,
		&limit.MaxTotalExposurePct,
		&limit.CreatedAt,
		&limit.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("risk limits not found for account group: %s", accountGroup)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get group risk limits: %w", err)
	}

	return &limit, nil
}

// Create inserts a new risk limit
func (r *RiskLimitRepository) Create(ctx context.Context, limit *RiskLimit) error {
	query := `
		INSERT INTO risk_limits (
			account_id, account_group, max_leverage, max_open_positions,
			max_position_size_lots, daily_loss_limit, max_drawdown_pct,
			margin_call_level, stop_out_level, max_symbol_exposure_pct,
			max_total_exposure_pct, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		limit.AccountID,
		limit.AccountGroup,
		limit.MaxLeverage,
		limit.MaxOpenPositions,
		limit.MaxPositionSizeLots,
		limit.DailyLossLimit,
		limit.MaxDrawdownPct,
		limit.MarginCallLevel,
		limit.StopOutLevel,
		limit.MaxSymbolExposurePct,
		limit.MaxTotalExposurePct,
	).Scan(&limit.ID, &limit.CreatedAt, &limit.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create risk limit: %w", err)
	}

	return nil
}

// Update modifies an existing risk limit
func (r *RiskLimitRepository) Update(ctx context.Context, limit *RiskLimit) error {
	query := `
		UPDATE risk_limits SET
			max_leverage = $1,
			max_open_positions = $2,
			max_position_size_lots = $3,
			daily_loss_limit = $4,
			max_drawdown_pct = $5,
			margin_call_level = $6,
			stop_out_level = $7,
			max_symbol_exposure_pct = $8,
			max_total_exposure_pct = $9,
			updated_at = NOW()
		WHERE id = $10
		RETURNING updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		limit.MaxLeverage,
		limit.MaxOpenPositions,
		limit.MaxPositionSizeLots,
		limit.DailyLossLimit,
		limit.MaxDrawdownPct,
		limit.MarginCallLevel,
		limit.StopOutLevel,
		limit.MaxSymbolExposurePct,
		limit.MaxTotalExposurePct,
		limit.ID,
	).Scan(&limit.UpdatedAt)

	if err == pgx.ErrNoRows {
		return fmt.Errorf("risk limit not found: %d", limit.ID)
	}
	if err != nil {
		return fmt.Errorf("failed to update risk limit: %w", err)
	}

	return nil
}

// GetAll retrieves all risk limits (for admin/reporting)
func (r *RiskLimitRepository) GetAll(ctx context.Context) ([]*RiskLimit, error) {
	query := `
		SELECT id, account_id, account_group, max_leverage, max_open_positions,
		       max_position_size_lots, daily_loss_limit, max_drawdown_pct,
		       margin_call_level, stop_out_level, max_symbol_exposure_pct,
		       max_total_exposure_pct, created_at, updated_at
		FROM risk_limits
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all risk limits: %w", err)
	}
	defer rows.Close()

	var limits []*RiskLimit
	for rows.Next() {
		var limit RiskLimit
		err := rows.Scan(
			&limit.ID,
			&limit.AccountID,
			&limit.AccountGroup,
			&limit.MaxLeverage,
			&limit.MaxOpenPositions,
			&limit.MaxPositionSizeLots,
			&limit.DailyLossLimit,
			&limit.MaxDrawdownPct,
			&limit.MarginCallLevel,
			&limit.StopOutLevel,
			&limit.MaxSymbolExposurePct,
			&limit.MaxTotalExposurePct,
			&limit.CreatedAt,
			&limit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan risk limit: %w", err)
		}
		limits = append(limits, &limit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating risk limits: %w", err)
	}

	return limits, nil
}
