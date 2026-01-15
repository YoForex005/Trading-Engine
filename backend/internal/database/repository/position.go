package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Position matches bbook.Position structure
type Position struct {
	ID            int64
	AccountID     int64
	Symbol        string
	Side          string
	Volume        float64
	OpenPrice     float64
	CurrentPrice  float64
	OpenTime      time.Time
	SL            float64
	TP            float64
	Swap          float64
	Commission    float64
	UnrealizedPnL float64
	Status        string
	ClosePrice    float64
	CloseTime     time.Time
	CloseReason   string
}

type PositionRepository struct {
	pool *pgxpool.Pool
}

func NewPositionRepository(pool *pgxpool.Pool) *PositionRepository {
	return &PositionRepository{pool: pool}
}

// Create inserts a new position
func (r *PositionRepository) Create(ctx context.Context, pos *Position) error {
	query := `
		INSERT INTO positions (
			account_id, symbol, side, volume, open_price, current_price,
			open_time, sl, tp, swap, commission, unrealized_pnl, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, open_time
	`

	err := r.pool.QueryRow(ctx, query,
		pos.AccountID, pos.Symbol, pos.Side, pos.Volume, pos.OpenPrice, pos.CurrentPrice,
		pos.OpenTime, pos.SL, pos.TP, pos.Swap, pos.Commission, pos.UnrealizedPnL, pos.Status,
	).Scan(&pos.ID, &pos.OpenTime)

	if err != nil {
		return fmt.Errorf("failed to create position: %w", err)
	}
	return nil
}

// GetByID retrieves position by ID
func (r *PositionRepository) GetByID(ctx context.Context, id int64) (*Position, error) {
	query := `
		SELECT id, account_id, symbol, side, volume, open_price, current_price,
		       open_time, sl, tp, swap, commission, unrealized_pnl, status,
		       close_price, close_time, close_reason
		FROM positions WHERE id = $1
	`

	var pos Position
	var closeTime *time.Time
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&pos.ID, &pos.AccountID, &pos.Symbol, &pos.Side, &pos.Volume,
		&pos.OpenPrice, &pos.CurrentPrice, &pos.OpenTime, &pos.SL, &pos.TP,
		&pos.Swap, &pos.Commission, &pos.UnrealizedPnL, &pos.Status,
		&pos.ClosePrice, &closeTime, &pos.CloseReason,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("position not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get position: %w", err)
	}

	if closeTime != nil {
		pos.CloseTime = *closeTime
	}

	return &pos, nil
}

// ListByAccount retrieves all positions for an account
func (r *PositionRepository) ListByAccount(ctx context.Context, accountID int64) ([]*Position, error) {
	query := `
		SELECT id, account_id, symbol, side, volume, open_price, current_price,
		       open_time, sl, tp, swap, commission, unrealized_pnl, status,
		       close_price, close_time, close_reason
		FROM positions
		WHERE account_id = $1
		ORDER BY open_time DESC
	`

	rows, err := r.pool.Query(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to list positions: %w", err)
	}
	defer rows.Close()

	var positions []*Position
	for rows.Next() {
		var pos Position
		var closeTime *time.Time
		err := rows.Scan(
			&pos.ID, &pos.AccountID, &pos.Symbol, &pos.Side, &pos.Volume,
			&pos.OpenPrice, &pos.CurrentPrice, &pos.OpenTime, &pos.SL, &pos.TP,
			&pos.Swap, &pos.Commission, &pos.UnrealizedPnL, &pos.Status,
			&pos.ClosePrice, &closeTime, &pos.CloseReason,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}

		if closeTime != nil {
			pos.CloseTime = *closeTime
		}

		positions = append(positions, &pos)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating positions: %w", err)
	}

	return positions, nil
}

// UpdatePrice updates current price and unrealized P/L
func (r *PositionRepository) UpdatePrice(ctx context.Context, id int64, currentPrice, unrealizedPnL float64) error {
	query := `
		UPDATE positions
		SET current_price = $1, unrealized_pnl = $2
		WHERE id = $3 AND status = 'OPEN'
	`

	result, err := r.pool.Exec(ctx, query, currentPrice, unrealizedPnL, id)
	if err != nil {
		return fmt.Errorf("failed to update price: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("position not found or not open: %d", id)
	}

	return nil
}

// Close marks position as closed
func (r *PositionRepository) Close(ctx context.Context, id int64, closePrice float64, closeReason string) error {
	// Use transaction for financial consistency
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE positions
		SET status = 'CLOSED', close_price = $1, close_time = NOW(), close_reason = $2
		WHERE id = $3 AND status = 'OPEN'
	`

	result, err := tx.Exec(ctx, query, closePrice, closeReason, id)
	if err != nil {
		return fmt.Errorf("failed to close position: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("position not found or already closed: %d", id)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
