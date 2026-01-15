package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/epic1st/rtx/backend/internal/database/repository"
	"github.com/epic1st/rtx/backend/internal/domain/position"
	"github.com/epic1st/rtx/backend/internal/ports"
)

// PositionRepositoryAdapter implements ports.PositionRepository
type PositionRepositoryAdapter struct {
	repo *repository.PositionRepository
}

// Verify interface implementation at compile time
var _ ports.PositionRepository = (*PositionRepositoryAdapter)(nil)

// NewPositionRepositoryAdapter creates a new position repository adapter
func NewPositionRepositoryAdapter(pool *pgxpool.Pool) *PositionRepositoryAdapter {
	return &PositionRepositoryAdapter{
		repo: repository.NewPositionRepository(pool),
	}
}

// toDomain converts repository.Position to domain.Position
func (a *PositionRepositoryAdapter) toDomain(repoPos *repository.Position) *position.Position {
	return &position.Position{
		ID:            repoPos.ID,
		AccountID:     repoPos.AccountID,
		Symbol:        repoPos.Symbol,
		Side:          repoPos.Side,
		Volume:        repoPos.Volume,
		OpenPrice:     repoPos.OpenPrice,
		CurrentPrice:  repoPos.CurrentPrice,
		OpenTime:      repoPos.OpenTime,
		SL:            repoPos.SL,
		TP:            repoPos.TP,
		Swap:          repoPos.Swap,
		Commission:    repoPos.Commission,
		UnrealizedPnL: repoPos.UnrealizedPnL,
		Status:        repoPos.Status,
		ClosePrice:    repoPos.ClosePrice,
		CloseTime:     repoPos.CloseTime,
		CloseReason:   repoPos.CloseReason,
	}
}

// toRepository converts domain.Position to repository.Position
func (a *PositionRepositoryAdapter) toRepository(domainPos *position.Position) *repository.Position {
	return &repository.Position{
		ID:            domainPos.ID,
		AccountID:     domainPos.AccountID,
		Symbol:        domainPos.Symbol,
		Side:          domainPos.Side,
		Volume:        domainPos.Volume,
		OpenPrice:     domainPos.OpenPrice,
		CurrentPrice:  domainPos.CurrentPrice,
		OpenTime:      domainPos.OpenTime,
		SL:            domainPos.SL,
		TP:            domainPos.TP,
		Swap:          domainPos.Swap,
		Commission:    domainPos.Commission,
		UnrealizedPnL: domainPos.UnrealizedPnL,
		Status:        domainPos.Status,
		ClosePrice:    domainPos.ClosePrice,
		CloseTime:     domainPos.CloseTime,
		CloseReason:   domainPos.CloseReason,
	}
}

// Get retrieves a position by ID
func (a *PositionRepositoryAdapter) Get(ctx context.Context, id int64) (*position.Position, error) {
	repoPos, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get position %d: %w", id, err)
	}
	return a.toDomain(repoPos), nil
}

// GetByAccount retrieves all positions for an account
func (a *PositionRepositoryAdapter) GetByAccount(ctx context.Context, accountID int64) ([]*position.Position, error) {
	repoPositions, err := a.repo.ListByAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to list positions for account %d: %w", accountID, err)
	}

	positions := make([]*position.Position, len(repoPositions))
	for i, repoPos := range repoPositions {
		positions[i] = a.toDomain(repoPos)
	}
	return positions, nil
}

// GetOpenPositions retrieves open positions for an account
func (a *PositionRepositoryAdapter) GetOpenPositions(ctx context.Context, accountID int64) ([]*position.Position, error) {
	// Filter from all positions
	allPositions, err := a.GetByAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	var positions []*position.Position
	for _, pos := range allPositions {
		if pos.Status == "OPEN" {
			positions = append(positions, pos)
		}
	}
	return positions, nil
}

// GetBySymbol retrieves positions for an account and symbol
func (a *PositionRepositoryAdapter) GetBySymbol(ctx context.Context, accountID int64, symbol string) ([]*position.Position, error) {
	// This needs to be added to repository.PositionRepository
	// For now, filter open positions
	allPositions, err := a.GetOpenPositions(ctx, accountID)
	if err != nil {
		return nil, err
	}

	var positions []*position.Position
	for _, pos := range allPositions {
		if pos.Symbol == symbol {
			positions = append(positions, pos)
		}
	}
	return positions, nil
}

// Create creates a new position
func (a *PositionRepositoryAdapter) Create(ctx context.Context, pos *position.Position) error {
	repoPos := a.toRepository(pos)
	if err := a.repo.Create(ctx, repoPos); err != nil {
		return fmt.Errorf("failed to create position: %w", err)
	}
	// Update the domain position with generated ID and timestamps
	pos.ID = repoPos.ID
	pos.OpenTime = repoPos.OpenTime
	return nil
}

// Update updates an existing position
func (a *PositionRepositoryAdapter) Update(ctx context.Context, pos *position.Position) error {
	// Use UpdatePrice for price updates
	if err := a.repo.UpdatePrice(ctx, pos.ID, pos.CurrentPrice, pos.UnrealizedPnL); err != nil {
		return fmt.Errorf("failed to update position %d: %w", pos.ID, err)
	}
	return nil
}

// Delete deletes a position (not typically used, positions are closed not deleted)
func (a *PositionRepositoryAdapter) Delete(ctx context.Context, id int64) error {
	// Positions should be closed, not deleted
	// This is a limitation of the current repository
	return fmt.Errorf("delete operation not implemented for positions, use Close instead")
}
