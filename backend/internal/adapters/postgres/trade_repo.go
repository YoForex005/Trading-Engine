package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/epic1st/rtx/backend/internal/database/repository"
	"github.com/epic1st/rtx/backend/internal/domain/trade"
	"github.com/epic1st/rtx/backend/internal/ports"
)

// TradeRepositoryAdapter implements ports.TradeRepository
type TradeRepositoryAdapter struct {
	repo *repository.TradeRepository
}

// Verify interface implementation at compile time
var _ ports.TradeRepository = (*TradeRepositoryAdapter)(nil)

// NewTradeRepositoryAdapter creates a new trade repository adapter
func NewTradeRepositoryAdapter(pool *pgxpool.Pool) *TradeRepositoryAdapter {
	return &TradeRepositoryAdapter{
		repo: repository.NewTradeRepository(pool),
	}
}

// toDomain converts repository.Trade to domain.Trade
func (a *TradeRepositoryAdapter) toDomain(repoTrade *repository.Trade) *trade.Trade {
	domainTrade := &trade.Trade{
		ID:          repoTrade.ID,
		PositionID:  repoTrade.PositionID,
		AccountID:   repoTrade.AccountID,
		Symbol:      repoTrade.Symbol,
		Side:        repoTrade.Side,
		Volume:      repoTrade.Volume,
		Price:       repoTrade.Price,
		RealizedPnL: repoTrade.RealizedPnL,
		Commission:  repoTrade.Commission,
		ExecutedAt:  repoTrade.ExecutedAt,
	}
	if repoTrade.OrderID != nil {
		domainTrade.OrderID = *repoTrade.OrderID
	}
	return domainTrade
}

// toRepository converts domain.Trade to repository.Trade
func (a *TradeRepositoryAdapter) toRepository(domainTrade *trade.Trade) *repository.Trade {
	repoTrade := &repository.Trade{
		ID:          domainTrade.ID,
		PositionID:  domainTrade.PositionID,
		AccountID:   domainTrade.AccountID,
		Symbol:      domainTrade.Symbol,
		Side:        domainTrade.Side,
		Volume:      domainTrade.Volume,
		Price:       domainTrade.Price,
		RealizedPnL: domainTrade.RealizedPnL,
		Commission:  domainTrade.Commission,
		ExecutedAt:  domainTrade.ExecutedAt,
	}
	if domainTrade.OrderID != 0 {
		repoTrade.OrderID = &domainTrade.OrderID
	}
	return repoTrade
}

// Create creates a new trade
func (a *TradeRepositoryAdapter) Create(ctx context.Context, trd *trade.Trade) error {
	repoTrade := a.toRepository(trd)
	if err := a.repo.Create(ctx, repoTrade); err != nil {
		return fmt.Errorf("failed to create trade: %w", err)
	}
	// Update the domain trade with generated ID and timestamps
	trd.ID = repoTrade.ID
	trd.ExecutedAt = repoTrade.ExecutedAt
	return nil
}

// Get retrieves a trade by ID
func (a *TradeRepositoryAdapter) Get(ctx context.Context, id int64) (*trade.Trade, error) {
	// GetByID not implemented in repository, need to list and filter
	// This is a limitation of the current repository
	return nil, fmt.Errorf("Get by ID not implemented for trades")
}

// GetByAccount retrieves trades for an account with pagination
func (a *TradeRepositoryAdapter) GetByAccount(ctx context.Context, accountID int64, limit, offset int) ([]*trade.Trade, error) {
	repoTrades, err := a.repo.ListByAccount(ctx, accountID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list trades for account %d: %w", accountID, err)
	}

	trades := make([]*trade.Trade, len(repoTrades))
	for i, repoTrade := range repoTrades {
		trades[i] = a.toDomain(repoTrade)
	}
	return trades, nil
}

// GetByPosition retrieves trades for a position
func (a *TradeRepositoryAdapter) GetByPosition(ctx context.Context, positionID int64) ([]*trade.Trade, error) {
	// ListByPosition not implemented in repository
	// This is a limitation of the current repository
	return []*trade.Trade{}, nil
}

// GetBySymbol retrieves trades for a symbol
func (a *TradeRepositoryAdapter) GetBySymbol(ctx context.Context, accountID int64, symbol string) ([]*trade.Trade, error) {
	// This needs to be added to repository.TradeRepository
	// For now, get all trades and filter
	allTrades, err := a.GetByAccount(ctx, accountID, 1000, 0)
	if err != nil {
		return nil, err
	}

	var trades []*trade.Trade
	for _, trd := range allTrades {
		if trd.Symbol == symbol {
			trades = append(trades, trd)
		}
	}
	return trades, nil
}
