package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Order matches bbook.Order structure
type Order struct {
	ID           int64
	AccountID    int64
	Symbol       string
	Type         string
	Side         string
	Volume       float64
	Price        float64
	TriggerPrice float64
	SL           float64
	TP           float64
	Status       string
	FilledPrice  float64
	FilledAt     *time.Time
	PositionID   int64
	RejectReason string
	CreatedAt    time.Time
}

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

// Create inserts a new order
func (r *OrderRepository) Create(ctx context.Context, ord *Order) error {
	query := `
		INSERT INTO orders (
			account_id, symbol, type, side, volume, price, trigger_price,
			sl, tp, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at
	`

	err := r.pool.QueryRow(ctx, query,
		ord.AccountID, ord.Symbol, ord.Type, ord.Side, ord.Volume,
		ord.Price, ord.TriggerPrice, ord.SL, ord.TP, ord.Status, ord.CreatedAt,
	).Scan(&ord.ID, &ord.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

// GetByID retrieves order by ID
func (r *OrderRepository) GetByID(ctx context.Context, id int64) (*Order, error) {
	query := `
		SELECT id, account_id, symbol, type, side, volume, price, trigger_price,
		       sl, tp, status, filled_price, filled_at, position_id, reject_reason, created_at
		FROM orders WHERE id = $1
	`

	var ord Order
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ord.ID, &ord.AccountID, &ord.Symbol, &ord.Type, &ord.Side, &ord.Volume,
		&ord.Price, &ord.TriggerPrice, &ord.SL, &ord.TP, &ord.Status,
		&ord.FilledPrice, &ord.FilledAt, &ord.PositionID, &ord.RejectReason, &ord.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("order not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return &ord, nil
}

// ListByAccount retrieves all orders for an account
func (r *OrderRepository) ListByAccount(ctx context.Context, accountID int64) ([]*Order, error) {
	query := `
		SELECT id, account_id, symbol, type, side, volume, price, trigger_price,
		       sl, tp, status, filled_price, filled_at, position_id, reject_reason, created_at
		FROM orders
		WHERE account_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var ord Order
		err := rows.Scan(
			&ord.ID, &ord.AccountID, &ord.Symbol, &ord.Type, &ord.Side, &ord.Volume,
			&ord.Price, &ord.TriggerPrice, &ord.SL, &ord.TP, &ord.Status,
			&ord.FilledPrice, &ord.FilledAt, &ord.PositionID, &ord.RejectReason, &ord.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, &ord)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return orders, nil
}

// UpdateStatus updates order status and optionally filled price
func (r *OrderRepository) UpdateStatus(ctx context.Context, id int64, status string, filledPrice *float64) error {
	var query string
	var args []interface{}

	if filledPrice != nil {
		query = `
			UPDATE orders
			SET status = $1, filled_price = $2, filled_at = NOW()
			WHERE id = $3
		`
		args = []interface{}{status, *filledPrice, id}
	} else {
		query = `
			UPDATE orders
			SET status = $1
			WHERE id = $2
		`
		args = []interface{}{status, id}
	}

	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("order not found: %d", id)
	}

	return nil
}

// Delete removes an order
func (r *OrderRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM orders WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("order not found: %d", id)
	}

	return nil
}
