package ports

import (
	"context"

	"github.com/epic1st/rtx/backend/internal/domain/account"
	"github.com/epic1st/rtx/backend/internal/domain/order"
	"github.com/epic1st/rtx/backend/internal/domain/position"
	"github.com/epic1st/rtx/backend/internal/domain/trade"
)

// TradingService defines trading operations
type TradingService interface {
	// Order execution
	ExecuteMarketOrder(ctx context.Context, ord *order.Order) (*position.Position, error)
	ExecutePendingOrder(ctx context.Context, ord *order.Order) (*position.Position, error)

	// Position management
	ClosePosition(ctx context.Context, positionID int64, reason string) (*trade.Trade, error)
	ModifyPosition(ctx context.Context, positionID int64, sl, tp float64) error

	// Order management
	PlaceOrder(ctx context.Context, ord *order.Order) error
	ModifyOrder(ctx context.Context, orderID int64, triggerPrice, sl, tp float64) error
	CancelOrder(ctx context.Context, orderID int64) error

	// Account operations
	UpdateAccountMargins(ctx context.Context, accountID int64) error
	CalculateMargin(ctx context.Context, accountID int64) error
	CheckStopOut(ctx context.Context, accountID int64) error

	// Account summary
	GetAccountSummary(ctx context.Context, accountID int64) (*account.Summary, error)
}

// MarketDataService defines market data operations
type MarketDataService interface {
	GetCurrentPrice(ctx context.Context, symbol string) (bid, ask float64, err error)
	GetSymbolSpec(ctx context.Context, symbol string) (*SymbolSpec, error)
	SubscribeToSymbol(ctx context.Context, symbol string) error
	UnsubscribeFromSymbol(ctx context.Context, symbol string) error
}

// SymbolSpec contains symbol specifications
type SymbolSpec struct {
	Symbol           string
	ContractSize     float64
	PipSize          float64
	PipValue         float64
	MinVolume        float64
	MaxVolume        float64
	VolumeStep       float64
	MarginPercent    float64
	CommissionPerLot float64
}

// RiskManagementService defines risk management operations
type RiskManagementService interface {
	ValidateOrder(ctx context.Context, accountID int64, ord *order.Order) error
	CheckPositionLimits(ctx context.Context, accountID int64) error
	CheckExposureLimits(ctx context.Context, accountID int64, symbol string, volume float64) error
	CheckDailyLimits(ctx context.Context, accountID int64) error
	UpdateDailyStats(ctx context.Context, accountID int64) error
}

// NotificationService defines notification operations
type NotificationService interface {
	NotifyOrderFilled(ctx context.Context, accountID int64, ord *order.Order) error
	NotifyPositionClosed(ctx context.Context, accountID int64, pos *position.Position) error
	NotifyMarginCall(ctx context.Context, accountID int64, marginLevel float64) error
	NotifyStopOut(ctx context.Context, accountID int64) error
}
