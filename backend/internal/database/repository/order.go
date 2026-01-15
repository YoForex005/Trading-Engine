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

// Order matches bbook.Order structure
type Order struct {
	ID               int64
	AccountID        int64
	Symbol           string
	Type             string
	Side             string
	Volume           float64
	Price            float64
	TriggerPrice     float64
	SL               float64
	TP               float64
	Status           string
	FilledPrice      float64
	FilledAt         *time.Time
	PositionID       int64
	RejectReason     string
	CreatedAt        time.Time
	ParentPositionID *int64     // Links SL/TP to position (NULL for standalone)
	TrailingDelta    *float64   // Distance for trailing stops (NULL for non-trailing)
	ExpiryTime       *time.Time // Auto-cancel after this time (NULL for GTC)
	OCOLinkID        *int64     // References another order for OCO (NULL if no link)
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
			sl, tp, status, created_at, parent_position_id, trailing_delta,
			expiry_time, oco_link_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at
	`

	err := r.pool.QueryRow(ctx, query,
		ord.AccountID, ord.Symbol, ord.Type, ord.Side, ord.Volume,
		ord.Price, ord.TriggerPrice, ord.SL, ord.TP, ord.Status, ord.CreatedAt,
		ord.ParentPositionID, ord.TrailingDelta, ord.ExpiryTime, ord.OCOLinkID,
	).Scan(&ord.ID, &ord.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create order for account %d symbol %s: %w", ord.AccountID, ord.Symbol, err)
	}
	return nil
}

// GetByID retrieves order by ID
func (r *OrderRepository) GetByID(ctx context.Context, id int64) (*Order, error) {
	query := `
		SELECT id, account_id, symbol, type, side, volume, price, trigger_price,
		       sl, tp, status, filled_price, filled_at, position_id, reject_reason, created_at,
		       parent_position_id, trailing_delta, expiry_time, oco_link_id
		FROM orders WHERE id = $1
	`

	var ord Order
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ord.ID, &ord.AccountID, &ord.Symbol, &ord.Type, &ord.Side, &ord.Volume,
		&ord.Price, &ord.TriggerPrice, &ord.SL, &ord.TP, &ord.Status,
		&ord.FilledPrice, &ord.FilledAt, &ord.PositionID, &ord.RejectReason, &ord.CreatedAt,
		&ord.ParentPositionID, &ord.TrailingDelta, &ord.ExpiryTime, &ord.OCOLinkID,
	)

	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NewNotFound("order", fmt.Sprintf("%d", id))
		}
		return nil, fmt.Errorf("failed to get order %d: %w", id, err)
	}
	return &ord, nil
}

// ListByAccount retrieves all orders for an account
func (r *OrderRepository) ListByAccount(ctx context.Context, accountID int64) ([]*Order, error) {
	query := `
		SELECT id, account_id, symbol, type, side, volume, price, trigger_price,
		       sl, tp, status, filled_price, filled_at, position_id, reject_reason, created_at,
		       parent_position_id, trailing_delta, expiry_time, oco_link_id
		FROM orders
		WHERE account_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders for account %d: %w", accountID, err)
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var ord Order
		err := rows.Scan(
			&ord.ID, &ord.AccountID, &ord.Symbol, &ord.Type, &ord.Side, &ord.Volume,
			&ord.Price, &ord.TriggerPrice, &ord.SL, &ord.TP, &ord.Status,
			&ord.FilledPrice, &ord.FilledAt, &ord.PositionID, &ord.RejectReason, &ord.CreatedAt,
			&ord.ParentPositionID, &ord.TrailingDelta, &ord.ExpiryTime, &ord.OCOLinkID,
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
		return fmt.Errorf("failed to update order %d status to %s: %w", id, status, err)
	}

	if result.RowsAffected() == 0 {
		return errors.NewNotFound("order", fmt.Sprintf("%d", id))
	}

	return nil
}

// Delete removes an order
func (r *OrderRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM orders WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete order %d: %w", id, err)
	}

	if result.RowsAffected() == 0 {
		return errors.NewNotFound("order", fmt.Sprintf("%d", id))
	}

	return nil
}

// ListPendingWithTriggers retrieves all pending orders that have trigger prices set
// Used by OrderMonitor to check for price-triggered executions
func (r *OrderRepository) ListPendingWithTriggers(ctx context.Context) ([]*Order, error) {
	query := `
		SELECT id, account_id, symbol, type, side, volume, price, trigger_price,
		       sl, tp, status, filled_price, filled_at, position_id, reject_reason, created_at,
		       parent_position_id, trailing_delta, expiry_time, oco_link_id
		FROM orders
		WHERE status = 'PENDING' AND trigger_price IS NOT NULL
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending orders: %w", err)
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var ord Order
		err := rows.Scan(
			&ord.ID, &ord.AccountID, &ord.Symbol, &ord.Type, &ord.Side, &ord.Volume,
			&ord.Price, &ord.TriggerPrice, &ord.SL, &ord.TP, &ord.Status,
			&ord.FilledPrice, &ord.FilledAt, &ord.PositionID, &ord.RejectReason, &ord.CreatedAt,
			&ord.ParentPositionID, &ord.TrailingDelta, &ord.ExpiryTime, &ord.OCOLinkID,
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

// UpdateTriggerPrice updates the trigger price for a trailing stop order
// Used by OrderMonitor to adjust trailing stops as market moves favorably
func (r *OrderRepository) UpdateTriggerPrice(ctx context.Context, id int64, newTriggerPrice float64) error {
	query := `
		UPDATE orders
		SET trigger_price = $1
		WHERE id = $2
	`

	result, err := r.pool.Exec(ctx, query, newTriggerPrice, id)
	if err != nil {
		return fmt.Errorf("failed to update trigger price for order %d: %w", id, err)
	}

	if result.RowsAffected() == 0 {
		return errors.NewNotFound("order", fmt.Sprintf("%d", id))
	}

	return nil
}

// UpdateOCOLink updates the oco_link_id for an order to create bidirectional OCO relationship
func (r *OrderRepository) UpdateOCOLink(ctx context.Context, orderID, linkedOrderID int64) error {
	query := `
		UPDATE orders
		SET oco_link_id = $1
		WHERE id = $2 AND status = 'PENDING'
	`

	result, err := r.pool.Exec(ctx, query, linkedOrderID, orderID)
	if err != nil {
		return fmt.Errorf("failed to update OCO link for order %d: %w", orderID, err)
	}

	if result.RowsAffected() == 0 {
		return errors.NewNotFound("order or not pending", fmt.Sprintf("%d", orderID))
	}

	return nil
}

// UpdateModifiable updates modifiable fields of a pending order
func (r *OrderRepository) UpdateModifiable(ctx context.Context, ord *Order) error {
	query := `
		UPDATE orders
		SET trigger_price = $1, sl = $2, tp = $3, volume = $4, expiry_time = $5
		WHERE id = $6 AND status = 'PENDING'
	`

	result, err := r.pool.Exec(ctx, query,
		ord.TriggerPrice, ord.SL, ord.TP, ord.Volume, ord.ExpiryTime, ord.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update order %d: %w", ord.ID, err)
	}

	if result.RowsAffected() == 0 {
		return errors.NewNotFound("order or not pending", fmt.Sprintf("%d", ord.ID))
	}

	return nil
}
