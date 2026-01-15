package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/epic1st/rtx/backend/internal/database/repository"
	"github.com/epic1st/rtx/backend/internal/domain/order"
	"github.com/epic1st/rtx/backend/internal/ports"
)

// OrderRepositoryAdapter implements ports.OrderRepository
type OrderRepositoryAdapter struct {
	repo *repository.OrderRepository
}

// Verify interface implementation at compile time
var _ ports.OrderRepository = (*OrderRepositoryAdapter)(nil)

// NewOrderRepositoryAdapter creates a new order repository adapter
func NewOrderRepositoryAdapter(pool *pgxpool.Pool) *OrderRepositoryAdapter {
	return &OrderRepositoryAdapter{
		repo: repository.NewOrderRepository(pool),
	}
}

// toDomain converts repository.Order to domain.Order
func (a *OrderRepositoryAdapter) toDomain(repoOrd *repository.Order) *order.Order {
	domainOrd := &order.Order{
		ID:           repoOrd.ID,
		AccountID:    repoOrd.AccountID,
		Symbol:       repoOrd.Symbol,
		Type:         repoOrd.Type,
		Side:         repoOrd.Side,
		Volume:       repoOrd.Volume,
		Price:        repoOrd.Price,
		TriggerPrice: repoOrd.TriggerPrice,
		SL:           repoOrd.SL,
		TP:           repoOrd.TP,
		Status:       repoOrd.Status,
		FilledPrice:  repoOrd.FilledPrice,
		FilledAt:     repoOrd.FilledAt,
		PositionID:   repoOrd.PositionID,
		RejectReason: repoOrd.RejectReason,
		CreatedAt:    repoOrd.CreatedAt,
	}

	// Handle nullable fields
	if repoOrd.ParentPositionID != nil {
		domainOrd.ParentPositionID = *repoOrd.ParentPositionID
	}
	if repoOrd.TrailingDelta != nil {
		domainOrd.TrailingDelta = *repoOrd.TrailingDelta
	}
	if repoOrd.ExpiryTime != nil {
		domainOrd.ExpiresAt = repoOrd.ExpiryTime
	}
	if repoOrd.OCOLinkID != nil {
		domainOrd.OCOLinkID = *repoOrd.OCOLinkID
	}

	return domainOrd
}

// toRepository converts domain.Order to repository.Order
func (a *OrderRepositoryAdapter) toRepository(domainOrd *order.Order) *repository.Order {
	repoOrd := &repository.Order{
		ID:           domainOrd.ID,
		AccountID:    domainOrd.AccountID,
		Symbol:       domainOrd.Symbol,
		Type:         domainOrd.Type,
		Side:         domainOrd.Side,
		Volume:       domainOrd.Volume,
		Price:        domainOrd.Price,
		TriggerPrice: domainOrd.TriggerPrice,
		SL:           domainOrd.SL,
		TP:           domainOrd.TP,
		Status:       domainOrd.Status,
		FilledPrice:  domainOrd.FilledPrice,
		FilledAt:     domainOrd.FilledAt,
		PositionID:   domainOrd.PositionID,
		RejectReason: domainOrd.RejectReason,
		CreatedAt:    domainOrd.CreatedAt,
	}

	// Handle nullable fields
	if domainOrd.ParentPositionID != 0 {
		repoOrd.ParentPositionID = &domainOrd.ParentPositionID
	}
	if domainOrd.TrailingDelta != 0 {
		repoOrd.TrailingDelta = &domainOrd.TrailingDelta
	}
	if domainOrd.ExpiresAt != nil {
		repoOrd.ExpiryTime = domainOrd.ExpiresAt
	}
	if domainOrd.OCOLinkID != 0 {
		repoOrd.OCOLinkID = &domainOrd.OCOLinkID
	}

	return repoOrd
}

// Get retrieves an order by ID
func (a *OrderRepositoryAdapter) Get(ctx context.Context, id int64) (*order.Order, error) {
	repoOrd, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order %d: %w", id, err)
	}
	return a.toDomain(repoOrd), nil
}

// GetByAccount retrieves all orders for an account
func (a *OrderRepositoryAdapter) GetByAccount(ctx context.Context, accountID int64) ([]*order.Order, error) {
	repoOrders, err := a.repo.ListByAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders for account %d: %w", accountID, err)
	}

	orders := make([]*order.Order, len(repoOrders))
	for i, repoOrd := range repoOrders {
		orders[i] = a.toDomain(repoOrd)
	}
	return orders, nil
}

// GetPending retrieves pending orders for an account
func (a *OrderRepositoryAdapter) GetPending(ctx context.Context, accountID int64) ([]*order.Order, error) {
	// Filter by PENDING status
	return a.GetByStatus(ctx, accountID, "PENDING")
}

// GetByStatus retrieves orders by status for an account
func (a *OrderRepositoryAdapter) GetByStatus(ctx context.Context, accountID int64, status string) ([]*order.Order, error) {
	// Filter from all orders
	allOrders, err := a.GetByAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	var orders []*order.Order
	for _, ord := range allOrders {
		if ord.Status == status {
			orders = append(orders, ord)
		}
	}
	return orders, nil
}

// GetByPosition retrieves orders for a position
func (a *OrderRepositoryAdapter) GetByPosition(ctx context.Context, positionID int64) ([]*order.Order, error) {
	// This needs to be added to repository.OrderRepository
	// For now, return empty list
	return []*order.Order{}, nil
}

// Create creates a new order
func (a *OrderRepositoryAdapter) Create(ctx context.Context, ord *order.Order) error {
	repoOrd := a.toRepository(ord)
	if err := a.repo.Create(ctx, repoOrd); err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	// Update the domain order with generated ID and timestamps
	ord.ID = repoOrd.ID
	ord.CreatedAt = repoOrd.CreatedAt
	return nil
}

// Update updates an existing order
func (a *OrderRepositoryAdapter) Update(ctx context.Context, ord *order.Order) error {
	repoOrd := a.toRepository(ord)
	if err := a.repo.UpdateModifiable(ctx, repoOrd); err != nil {
		return fmt.Errorf("failed to update order %d: %w", ord.ID, err)
	}
	return nil
}

// Cancel cancels an order
func (a *OrderRepositoryAdapter) Cancel(ctx context.Context, id int64, reason string) error {
	// Use UpdateStatus to cancel
	if err := a.repo.UpdateStatus(ctx, id, "CANCELLED", nil); err != nil {
		return fmt.Errorf("failed to cancel order %d: %w", id, err)
	}
	return nil
}

// Delete deletes an order
func (a *OrderRepositoryAdapter) Delete(ctx context.Context, id int64) error {
	if err := a.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete order %d: %w", id, err)
	}
	return nil
}
