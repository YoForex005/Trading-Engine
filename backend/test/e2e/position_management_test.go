package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/epic1st/rtx/backend/internal/database"
	"github.com/epic1st/rtx/backend/internal/database/repository"
)

// TestE2E_PositionModify verifies modifying open positions:
// 1. Open position
// 2. Modify stop-loss and take-profit
// 3. Verify modifications persisted
func TestE2E_PositionModify(t *testing.T) {
	ctx := context.Background()
	err := database.InitPool(ctx, getDatabaseURL(t))
	if err != nil {
		t.Fatalf("database connect failed: %v", err)
	}
	defer database.Close()

	pool := database.GetPool()
	if pool == nil {
		t.Fatal("database pool is nil")
	}

	defer cleanupTestData(t, pool)

	accountRepo := repository.NewAccountRepository(pool)
	positionRepo := repository.NewPositionRepository(pool)

	// Create account
	account := &repository.Account{
		AccountNumber: "E2E_TEST_003",
		UserID:        "test_user_003",
		Username:      "testuser3",
		Password:      "hashed_password",
		Balance:       10000.00,
		Equity:        10000.00,
		Margin:        0,
		FreeMargin:    10000.00,
		MarginLevel:   0,
		Leverage:      100,
		MarginMode:    "retail",
		Currency:      "USD",
		Status:        "active",
		IsDemo:        true,
	}
	accountRepo.Create(ctx, account)

	// Create position
	entryPrice := 1.0850
	position := &repository.Position{
		AccountID:     account.ID,
		Symbol:        "EURUSD",
		Side:          "buy",
		Volume:        1.0,
		OpenPrice:     entryPrice,
		CurrentPrice:  entryPrice,
		OpenTime:      time.Now(),
		SL:            0,
		TP:            0,
		Swap:          0,
		Commission:    0,
		UnrealizedPnL: 0,
		Status:        "open",
	}
	positionRepo.Create(ctx, position)

	// Modify stop-loss and take-profit
	stopLoss := 1.0800
	takeProfit := 1.0950
	err = positionRepo.UpdateSLTP(ctx, position.ID, stopLoss, takeProfit)
	if err != nil {
		t.Fatalf("update SL/TP failed: %v", err)
	}

	// Verify modifications persisted
	updatedPos, err := positionRepo.GetByID(ctx, position.ID)
	if err != nil {
		t.Fatalf("failed to fetch position: %v", err)
	}

	if updatedPos.SL != stopLoss {
		t.Errorf("stop-loss: got %f, want %f", updatedPos.SL, stopLoss)
	}

	if updatedPos.TP != takeProfit {
		t.Errorf("take-profit: got %f, want %f", updatedPos.TP, takeProfit)
	}
}

// TestE2E_MultiplePositions verifies managing multiple concurrent positions:
// 1. Open multiple positions on different symbols
// 2. Verify all positions tracked independently
// 3. Close one position
// 4. Verify other positions unaffected
func TestE2E_MultiplePositions(t *testing.T) {
	ctx := context.Background()
	err := database.InitPool(ctx, getDatabaseURL(t))
	if err != nil {
		t.Fatalf("database connect failed: %v", err)
	}
	defer database.Close()

	pool := database.GetPool()
	if pool == nil {
		t.Fatal("database pool is nil")
	}

	defer cleanupTestData(t, pool)

	accountRepo := repository.NewAccountRepository(pool)
	positionRepo := repository.NewPositionRepository(pool)

	// Create account
	account := &repository.Account{
		AccountNumber: "E2E_TEST_004",
		UserID:        "test_user_004",
		Username:      "testuser4",
		Password:      "hashed_password",
		Balance:       20000.00,
		Equity:        20000.00,
		Margin:        0,
		FreeMargin:    20000.00,
		MarginLevel:   0,
		Leverage:      100,
		MarginMode:    "retail",
		Currency:      "USD",
		Status:        "active",
		IsDemo:        true,
	}
	accountRepo.Create(ctx, account)

	// Open three positions on different symbols
	positions := []*repository.Position{
		{
			AccountID:     account.ID,
			Symbol:        "EURUSD",
			Side:          "buy",
			Volume:        1.0,
			OpenPrice:     1.0850,
			CurrentPrice:  1.0850,
			OpenTime:      time.Now(),
			SL:            0,
			TP:            0,
			Swap:          0,
			Commission:    0,
			UnrealizedPnL: 0,
			Status:        "open",
		},
		{
			AccountID:     account.ID,
			Symbol:        "GBPUSD",
			Side:          "buy",
			Volume:        1.0,
			OpenPrice:     1.2650,
			CurrentPrice:  1.2650,
			OpenTime:      time.Now(),
			SL:            0,
			TP:            0,
			Swap:          0,
			Commission:    0,
			UnrealizedPnL: 0,
			Status:        "open",
		},
		{
			AccountID:     account.ID,
			Symbol:        "USDJPY",
			Side:          "sell",
			Volume:        1.0,
			OpenPrice:     148.50,
			CurrentPrice:  148.50,
			OpenTime:      time.Now(),
			SL:            0,
			TP:            0,
			Swap:          0,
			Commission:    0,
			UnrealizedPnL: 0,
			Status:        "open",
		},
	}

	// Create all positions
	for _, pos := range positions {
		err := positionRepo.Create(ctx, pos)
		if err != nil {
			t.Fatalf("failed to create position %s: %v", pos.Symbol, err)
		}
	}

	// Verify all three positions exist
	allPositions, err := positionRepo.ListByAccount(ctx, account.ID)
	if err != nil {
		t.Fatalf("failed to fetch positions: %v", err)
	}

	if len(allPositions) != 3 {
		t.Fatalf("expected 3 positions, got %d", len(allPositions))
	}

	// Close GBPUSD position (middle one)
	gbpusdPos := positions[1]
	err = positionRepo.Close(ctx, gbpusdPos.ID, 1.2700, "manual")
	if err != nil {
		t.Fatalf("close position failed: %v", err)
	}

	// Verify only EURUSD and USDJPY remain open
	openPositions, err := positionRepo.ListByAccount(ctx, account.ID)
	if err != nil {
		t.Fatalf("failed to fetch positions: %v", err)
	}

	openCount := 0
	closedCount := 0
	for _, pos := range openPositions {
		if pos.Status == "open" {
			openCount++
		} else if pos.Status == "closed" {
			closedCount++
		}
	}

	if openCount != 2 {
		t.Errorf("expected 2 open positions, got %d", openCount)
	}

	if closedCount != 1 {
		t.Errorf("expected 1 closed position, got %d", closedCount)
	}

	// Verify correct positions remain open
	symbols := make(map[string]bool)
	for _, pos := range openPositions {
		if pos.Status == "open" {
			symbols[pos.Symbol] = true
		}
	}

	if !symbols["EURUSD"] || !symbols["USDJPY"] {
		t.Error("wrong positions remain after close - expected EURUSD and USDJPY")
	}

	if symbols["GBPUSD"] {
		// Note: GBPUSD might still be in list but should be marked closed
		for _, pos := range openPositions {
			if pos.Symbol == "GBPUSD" && pos.Status == "open" {
				t.Error("GBPUSD should be closed but is still open")
			}
		}
	}
}

// TestE2E_PositionPriceUpdate verifies position price updates:
// 1. Create position
// 2. Update current price and unrealized PnL
// 3. Verify updates persisted
func TestE2E_PositionPriceUpdate(t *testing.T) {
	ctx := context.Background()
	err := database.InitPool(ctx, getDatabaseURL(t))
	if err != nil {
		t.Fatalf("database connect failed: %v", err)
	}
	defer database.Close()

	pool := database.GetPool()
	if pool == nil {
		t.Fatal("database pool is nil")
	}

	defer cleanupTestData(t, pool)

	accountRepo := repository.NewAccountRepository(pool)
	positionRepo := repository.NewPositionRepository(pool)

	// Create account
	account := &repository.Account{
		AccountNumber: "E2E_TEST_005",
		UserID:        "test_user_005",
		Username:      "testuser5",
		Password:      "hashed_password",
		Balance:       10000.00,
		Equity:        10000.00,
		Margin:        0,
		FreeMargin:    10000.00,
		MarginLevel:   0,
		Leverage:      100,
		MarginMode:    "retail",
		Currency:      "USD",
		Status:        "active",
		IsDemo:        true,
	}
	accountRepo.Create(ctx, account)

	// Create position
	entryPrice := 1.0850
	position := &repository.Position{
		AccountID:     account.ID,
		Symbol:        "EURUSD",
		Side:          "buy",
		Volume:        1.0,
		OpenPrice:     entryPrice,
		CurrentPrice:  entryPrice,
		OpenTime:      time.Now(),
		SL:            0,
		TP:            0,
		Swap:          0,
		Commission:    0,
		UnrealizedPnL: 0,
		Status:        "open",
	}
	positionRepo.Create(ctx, position)

	// Update price (market moved up 50 pips)
	newPrice := 1.0900
	// Calculate unrealized PnL: (newPrice - openPrice) * volume * contract size
	// For EURUSD: (1.0900 - 1.0850) * 1.0 * 100000 = 500 USD
	unrealizedPnL := 500.0

	err = positionRepo.UpdatePrice(ctx, position.ID, newPrice, unrealizedPnL)
	if err != nil {
		t.Fatalf("update price failed: %v", err)
	}

	// Verify updates persisted
	updatedPos, err := positionRepo.GetByID(ctx, position.ID)
	if err != nil {
		t.Fatalf("failed to fetch position: %v", err)
	}

	if updatedPos.CurrentPrice != newPrice {
		t.Errorf("current price: got %f, want %f", updatedPos.CurrentPrice, newPrice)
	}

	if updatedPos.UnrealizedPnL != unrealizedPnL {
		t.Errorf("unrealized PnL: got %f, want %f", updatedPos.UnrealizedPnL, unrealizedPnL)
	}

	// Verify position is still open
	if updatedPos.Status != "open" {
		t.Errorf("position status: got %s, want open", updatedPos.Status)
	}
}
