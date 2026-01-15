package ports

import (
	"context"

	"github.com/epic1st/rtx/backend/internal/domain/account"
	"github.com/epic1st/rtx/backend/internal/domain/order"
	"github.com/epic1st/rtx/backend/internal/domain/position"
	"github.com/epic1st/rtx/backend/internal/domain/trade"
)

// AccountRepository defines account persistence operations
type AccountRepository interface {
	Get(ctx context.Context, id int64) (*account.Account, error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (*account.Account, error)
	GetByUsername(ctx context.Context, username string) (*account.Account, error)
	Create(ctx context.Context, acc *account.Account) error
	Update(ctx context.Context, acc *account.Account) error
	List(ctx context.Context) ([]*account.Account, error)
	Delete(ctx context.Context, id int64) error
}

// PositionRepository defines position persistence operations
type PositionRepository interface {
	Get(ctx context.Context, id int64) (*position.Position, error)
	GetByAccount(ctx context.Context, accountID int64) ([]*position.Position, error)
	GetOpenPositions(ctx context.Context, accountID int64) ([]*position.Position, error)
	GetBySymbol(ctx context.Context, accountID int64, symbol string) ([]*position.Position, error)
	Create(ctx context.Context, pos *position.Position) error
	Update(ctx context.Context, pos *position.Position) error
	Delete(ctx context.Context, id int64) error
}

// OrderRepository defines order persistence operations
type OrderRepository interface {
	Get(ctx context.Context, id int64) (*order.Order, error)
	GetByAccount(ctx context.Context, accountID int64) ([]*order.Order, error)
	GetPending(ctx context.Context, accountID int64) ([]*order.Order, error)
	GetByStatus(ctx context.Context, accountID int64, status string) ([]*order.Order, error)
	GetByPosition(ctx context.Context, positionID int64) ([]*order.Order, error)
	Create(ctx context.Context, ord *order.Order) error
	Update(ctx context.Context, ord *order.Order) error
	Cancel(ctx context.Context, id int64, reason string) error
	Delete(ctx context.Context, id int64) error
}

// TradeRepository defines trade persistence operations (read-only after creation)
type TradeRepository interface {
	Create(ctx context.Context, trd *trade.Trade) error
	Get(ctx context.Context, id int64) (*trade.Trade, error)
	GetByAccount(ctx context.Context, accountID int64, limit, offset int) ([]*trade.Trade, error)
	GetByPosition(ctx context.Context, positionID int64) ([]*trade.Trade, error)
	GetBySymbol(ctx context.Context, accountID int64, symbol string) ([]*trade.Trade, error)
}
