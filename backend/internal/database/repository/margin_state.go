package repository

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/epic1st/rtx/backend/internal/shared/errors"
)

// MarginState represents real-time margin state per account
// CRITICAL: All DECIMAL columns stored as strings to avoid float precision issues
type MarginState struct {
	AccountID           int64
	Equity              string // DECIMAL(20,8) as string
	UsedMargin          string // DECIMAL(20,8) as string
	FreeMargin          string // DECIMAL(20,8) as string (computed column)
	MarginLevel         string // DECIMAL(10,2) as string
	MarginCallTriggered bool
	StopOutTriggered    bool
	LastUpdated         time.Time
}

type MarginStateRepository struct {
	pool *pgxpool.Pool
}

func NewMarginStateRepository(pool *pgxpool.Pool) *MarginStateRepository {
	return &MarginStateRepository{pool: pool}
}

// GetByAccountID retrieves margin state for an account
func (r *MarginStateRepository) GetByAccountID(ctx context.Context, accountID int64) (*MarginState, error) {
	query := `
		SELECT account_id, equity, used_margin, free_margin, margin_level,
		       margin_call_triggered, stop_out_triggered, last_updated
		FROM margin_state
		WHERE account_id = $1
	`

	var state MarginState
	err := r.pool.QueryRow(ctx, query, accountID).Scan(
		&state.AccountID,
		&state.Equity,
		&state.UsedMargin,
		&state.FreeMargin,
		&state.MarginLevel,
		&state.MarginCallTriggered,
		&state.StopOutTriggered,
		&state.LastUpdated,
	)

	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NewNotFound("margin_state", fmt.Sprintf("account_%d", accountID))
		}
		return nil, fmt.Errorf("failed to get margin state for account %d: %w", accountID, err)
	}

	return &state, nil
}

// Upsert inserts or updates margin state (for real-time updates)
func (r *MarginStateRepository) Upsert(ctx context.Context, state *MarginState) error {
	query := `
		INSERT INTO margin_state (
			account_id, equity, used_margin, margin_level,
			margin_call_triggered, stop_out_triggered, last_updated
		) VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (account_id) DO UPDATE SET
			equity = EXCLUDED.equity,
			used_margin = EXCLUDED.used_margin,
			margin_level = EXCLUDED.margin_level,
			margin_call_triggered = EXCLUDED.margin_call_triggered,
			stop_out_triggered = EXCLUDED.stop_out_triggered,
			last_updated = NOW()
		RETURNING account_id, equity, used_margin, free_margin, margin_level,
		          margin_call_triggered, stop_out_triggered, last_updated
	`

	err := r.pool.QueryRow(ctx, query,
		state.AccountID,
		state.Equity,
		state.UsedMargin,
		state.MarginLevel,
		state.MarginCallTriggered,
		state.StopOutTriggered,
	).Scan(
		&state.AccountID,
		&state.Equity,
		&state.UsedMargin,
		&state.FreeMargin,
		&state.MarginLevel,
		&state.MarginCallTriggered,
		&state.StopOutTriggered,
		&state.LastUpdated,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert margin state: %w", err)
	}

	return nil
}

// GetByAccountIDWithLock retrieves margin state with row lock for concurrent safety
// Used during margin calculation with transaction isolation
func (r *MarginStateRepository) GetByAccountIDWithLock(ctx context.Context, tx pgx.Tx, accountID int64) (*MarginState, error) {
	query := `
		SELECT account_id, equity, used_margin, free_margin, margin_level,
		       margin_call_triggered, stop_out_triggered, last_updated
		FROM margin_state
		WHERE account_id = $1
		FOR UPDATE
	`

	var state MarginState
	err := tx.QueryRow(ctx, query, accountID).Scan(
		&state.AccountID,
		&state.Equity,
		&state.UsedMargin,
		&state.FreeMargin,
		&state.MarginLevel,
		&state.MarginCallTriggered,
		&state.StopOutTriggered,
		&state.LastUpdated,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("margin state not found for account: %d", accountID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get margin state with lock: %w", err)
	}

	return &state, nil
}

// UpsertWithTx inserts or updates margin state within a transaction
func (r *MarginStateRepository) UpsertWithTx(ctx context.Context, tx pgx.Tx, state *MarginState) error {
	query := `
		INSERT INTO margin_state (
			account_id, equity, used_margin, margin_level,
			margin_call_triggered, stop_out_triggered, last_updated
		) VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (account_id) DO UPDATE SET
			equity = EXCLUDED.equity,
			used_margin = EXCLUDED.used_margin,
			margin_level = EXCLUDED.margin_level,
			margin_call_triggered = EXCLUDED.margin_call_triggered,
			stop_out_triggered = EXCLUDED.stop_out_triggered,
			last_updated = NOW()
		RETURNING account_id, equity, used_margin, free_margin, margin_level,
		          margin_call_triggered, stop_out_triggered, last_updated
	`

	err := tx.QueryRow(ctx, query,
		state.AccountID,
		state.Equity,
		state.UsedMargin,
		state.MarginLevel,
		state.MarginCallTriggered,
		state.StopOutTriggered,
	).Scan(
		&state.AccountID,
		&state.Equity,
		&state.UsedMargin,
		&state.FreeMargin,
		&state.MarginLevel,
		&state.MarginCallTriggered,
		&state.StopOutTriggered,
		&state.LastUpdated,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert margin state in transaction: %w", err)
	}

	return nil
}
