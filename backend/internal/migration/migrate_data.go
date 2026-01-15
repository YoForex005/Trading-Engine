package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
	"github.com/epic1st/rtx/backend/internal/database/repository"
)

// PersistenceData represents the old JSON file structure
type PersistenceData struct {
	Accounts       map[int64]*core.Account  `json:"accounts"`
	Positions      map[int64]*core.Position `json:"positions"`
	Orders         map[int64]*core.Order    `json:"orders"`
	Trades         []core.Trade             `json:"trades"`
	NextPositionID int64                    `json:"nextPositionId"`
	NextOrderID    int64                    `json:"nextOrderId"`
	NextTradeID    int64                    `json:"nextTradeId"`
	SavedAt        time.Time                `json:"savedAt"`
}

// MigrateFromJSON reads existing JSON persistence file and migrates to PostgreSQL
// Safe to run multiple times (idempotent - checks if data already exists)
func MigrateFromJSON(
	ctx context.Context,
	jsonPath string,
	accountRepo *repository.AccountRepository,
	positionRepo *repository.PositionRepository,
	orderRepo *repository.OrderRepository,
	tradeRepo *repository.TradeRepository,
) error {
	// Check if database already has data (avoid duplicate migration)
	existingAccounts, err := accountRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to check existing accounts: %w", err)
	}
	if len(existingAccounts) > 0 {
		fmt.Printf("Database already contains %d accounts, skipping migration\n", len(existingAccounts))
		return nil
	}

	// Check if JSON file exists
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		fmt.Println("No existing JSON file found, starting with empty database")
		return nil
	}

	// Read JSON file
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	// Parse persistence data
	var persistenceData PersistenceData
	if err := json.Unmarshal(data, &persistenceData); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	fmt.Printf("Migrating %d accounts, %d positions, %d orders, %d trades\n",
		len(persistenceData.Accounts), len(persistenceData.Positions),
		len(persistenceData.Orders), len(persistenceData.Trades))

	// Migrate accounts
	for _, acc := range persistenceData.Accounts {
		repoAcc := &repository.Account{
			AccountNumber: acc.AccountNumber,
			UserID:        acc.UserID,
			Username:      acc.Username,
			Password:      acc.Password,
			Balance:       acc.Balance,
			Equity:        acc.Equity,
			Margin:        acc.Margin,
			FreeMargin:    acc.FreeMargin,
			MarginLevel:   acc.MarginLevel,
			Leverage:      acc.Leverage,
			MarginMode:    acc.MarginMode,
			Currency:      acc.Currency,
			Status:        acc.Status,
			IsDemo:        acc.IsDemo,
		}
		if err := accountRepo.Create(ctx, repoAcc); err != nil {
			return fmt.Errorf("failed to migrate account %s: %w", acc.AccountNumber, err)
		}
	}

	// Migrate positions
	for _, pos := range persistenceData.Positions {
		repoPos := convertPosition(pos)
		if err := positionRepo.Create(ctx, repoPos); err != nil {
			return fmt.Errorf("failed to migrate position %d: %w", pos.ID, err)
		}
	}

	// Migrate orders
	for _, ord := range persistenceData.Orders {
		repoOrd := convertOrder(ord)
		if err := orderRepo.Create(ctx, repoOrd); err != nil {
			return fmt.Errorf("failed to migrate order %d: %w", ord.ID, err)
		}
	}

	// Migrate trades
	for _, trade := range persistenceData.Trades {
		repoTrade := convertTrade(&trade)
		if err := tradeRepo.Create(ctx, repoTrade); err != nil {
			return fmt.Errorf("failed to migrate trade %d: %w", trade.ID, err)
		}
	}

	fmt.Println("Migration completed successfully")
	return nil
}

// convertPosition converts core.Position to repository.Position
func convertPosition(pos *core.Position) *repository.Position {
	return &repository.Position{
		ID:            pos.ID,
		AccountID:     pos.AccountID,
		Symbol:        pos.Symbol,
		Side:          pos.Side,
		Volume:        pos.Volume,
		OpenPrice:     pos.OpenPrice,
		CurrentPrice:  pos.CurrentPrice,
		OpenTime:      pos.OpenTime,
		SL:            pos.SL,
		TP:            pos.TP,
		Swap:          pos.Swap,
		Commission:    pos.Commission,
		UnrealizedPnL: pos.UnrealizedPnL,
		Status:        pos.Status,
		ClosePrice:    pos.ClosePrice,
		CloseTime:     pos.CloseTime,
		CloseReason:   pos.CloseReason,
	}
}

// convertOrder converts core.Order to repository.Order
func convertOrder(ord *core.Order) *repository.Order {
	return &repository.Order{
		ID:           ord.ID,
		AccountID:    ord.AccountID,
		Symbol:       ord.Symbol,
		Type:         ord.Type,
		Side:         ord.Side,
		Volume:       ord.Volume,
		Price:        ord.Price,
		TriggerPrice: ord.TriggerPrice,
		SL:           ord.SL,
		TP:           ord.TP,
		Status:       ord.Status,
		FilledPrice:  ord.FilledPrice,
		FilledAt:     ord.FilledAt,
		PositionID:   ord.PositionID,
		RejectReason: ord.RejectReason,
		CreatedAt:    ord.CreatedAt,
	}
}

// convertTrade converts core.Trade to repository.Trade
func convertTrade(trade *core.Trade) *repository.Trade {
	var orderID *int64
	if trade.OrderID != 0 {
		orderID = &trade.OrderID
	}

	return &repository.Trade{
		ID:          trade.ID,
		OrderID:     orderID,
		PositionID:  trade.PositionID,
		AccountID:   trade.AccountID,
		Symbol:      trade.Symbol,
		Side:        trade.Side,
		Volume:      trade.Volume,
		Price:       trade.Price,
		RealizedPnL: trade.RealizedPnL,
		Commission:  trade.Commission,
		ExecutedAt:  trade.ExecutedAt,
	}
}
