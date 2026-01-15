package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Trade struct {
	ID          int64
	OrderID     *int64 // Nullable
	PositionID  int64
	AccountID   int64
	Symbol      string
	Side        string
	Volume      float64
	Price       float64
	RealizedPnL float64
	Commission  float64
	ExecutedAt  time.Time
}

type TradeRepository struct {
	pool *pgxpool.Pool
}

func NewTradeRepository(pool *pgxpool.Pool) *TradeRepository {
	return &TradeRepository{pool: pool}
}

// Create inserts a new trade record
func (r *TradeRepository) Create(ctx context.Context, trade *Trade) error {
	query := `
		INSERT INTO trades (
			order_id, position_id, account_id, symbol, side, volume,
			price, realized_pnl, commission, executed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, executed_at
	`

	err := r.pool.QueryRow(ctx, query,
		trade.OrderID, trade.PositionID, trade.AccountID, trade.Symbol,
		trade.Side, trade.Volume, trade.Price, trade.RealizedPnL,
		trade.Commission, trade.ExecutedAt,
	).Scan(&trade.ID, &trade.ExecutedAt)

	if err != nil {
		return fmt.Errorf("failed to create trade: %w", err)
	}
	return nil
}

// ListByAccount retrieves trades for an account with pagination
func (r *TradeRepository) ListByAccount(ctx context.Context, accountID int64, limit, offset int) ([]*Trade, error) {
	query := `
		SELECT id, order_id, position_id, account_id, symbol, side, volume,
		       price, realized_pnl, commission, executed_at
		FROM trades
		WHERE account_id = $1
		ORDER BY executed_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, accountID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list trades: %w", err)
	}
	defer rows.Close()

	var trades []*Trade
	for rows.Next() {
		var t Trade
		err := rows.Scan(
			&t.ID, &t.OrderID, &t.PositionID, &t.AccountID, &t.Symbol,
			&t.Side, &t.Volume, &t.Price, &t.RealizedPnL, &t.Commission,
			&t.ExecutedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}
		trades = append(trades, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trades: %w", err)
	}

	return trades, nil
}

// ListBySymbol retrieves trades for a symbol with time range
func (r *TradeRepository) ListBySymbol(ctx context.Context, symbol string, startTime, endTime time.Time, limit int) ([]*Trade, error) {
	query := `
		SELECT id, order_id, position_id, account_id, symbol, side, volume,
		       price, realized_pnl, commission, executed_at
		FROM trades
		WHERE symbol = $1 AND executed_at BETWEEN $2 AND $3
		ORDER BY executed_at DESC
		LIMIT $4
	`

	rows, err := r.pool.Query(ctx, query, symbol, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list trades by symbol: %w", err)
	}
	defer rows.Close()

	var trades []*Trade
	for rows.Next() {
		var t Trade
		err := rows.Scan(
			&t.ID, &t.OrderID, &t.PositionID, &t.AccountID, &t.Symbol,
			&t.Side, &t.Volume, &t.Price, &t.RealizedPnL, &t.Commission,
			&t.ExecutedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}
		trades = append(trades, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trades: %w", err)
	}

	return trades, nil
}
