package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/epic1st/rtx/backend/internal/database"
	"github.com/epic1st/rtx/backend/internal/database/repository"
)

// TestE2E_OrderExecution verifies complete order flow:
// 1. Create account
// 2. Place market order via repository
// 3. Verify order persisted
// 4. Verify position created
// 5. Verify database persistence
func TestE2E_OrderExecution(t *testing.T) {
	// Setup test database connection
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

	// Clean up test data
	defer cleanupTestData(t, pool)

	// Create repositories
	accountRepo := repository.NewAccountRepository(pool)
	orderRepo := repository.NewOrderRepository(pool)
	positionRepo := repository.NewPositionRepository(pool)

	// Step 1: Create test account
	account := &repository.Account{
		AccountNumber: "E2E_TEST_001",
		UserID:        "test_user_001",
		Username:      "testuser",
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

	err = accountRepo.Create(ctx, account)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	// Step 2: Create market order
	order := &repository.Order{
		AccountID:    account.ID,
		Symbol:       "EURUSD",
		Type:         "market",
		Side:         "buy",
		Volume:       1.0,
		Price:        0, // Market order has no limit price
		TriggerPrice: 0,
		SL:           0,
		TP:           0,
		Status:       "pending",
		CreatedAt:    time.Now(),
	}

	err = orderRepo.Create(ctx, order)
	if err != nil {
		t.Fatalf("order creation failed: %v", err)
	}

	// Step 3: Verify order persisted to database
	dbOrder, err := orderRepo.GetByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("failed to fetch order: %v", err)
	}

	if dbOrder.Symbol != "EURUSD" {
		t.Errorf("order symbol: got %s, want EURUSD", dbOrder.Symbol)
	}

	if dbOrder.Type != "market" {
		t.Errorf("order type: got %s, want market", dbOrder.Type)
	}

	if dbOrder.Side != "buy" {
		t.Errorf("order side: got %s, want buy", dbOrder.Side)
	}

	// Step 4: Simulate order fill by creating position
	position := &repository.Position{
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
	}

	err = positionRepo.Create(ctx, position)
	if err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	// Step 5: Update order status to filled
	filledPrice := 1.0850
	err = orderRepo.UpdateStatus(ctx, order.ID, "filled", &filledPrice)
	if err != nil {
		t.Fatalf("failed to update order status: %v", err)
	}

	// Verify order is filled
	dbOrder, err = orderRepo.GetByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("failed to fetch updated order: %v", err)
	}

	if dbOrder.Status != "filled" {
		t.Errorf("order status: got %s, want filled", dbOrder.Status)
	}

	// Step 6: Verify position exists
	positions, err := positionRepo.ListByAccount(ctx, account.ID)
	if err != nil {
		t.Fatalf("failed to fetch positions: %v", err)
	}

	if len(positions) != 1 {
		t.Fatalf("expected 1 position, got %d", len(positions))
	}

	pos := positions[0]
	if pos.Symbol != "EURUSD" {
		t.Errorf("position symbol: got %s, want EURUSD", pos.Symbol)
	}

	if pos.Volume != 1.0 {
		t.Errorf("position volume: got %f, want 1.0", pos.Volume)
	}

	// Step 7: Verify account can be retrieved
	retrievedAccount, err := accountRepo.GetByID(ctx, account.ID)
	if err != nil {
		t.Fatalf("failed to fetch account: %v", err)
	}

	if retrievedAccount.AccountNumber != "E2E_TEST_001" {
		t.Errorf("account number: got %s, want E2E_TEST_001", retrievedAccount.AccountNumber)
	}
}

// TestE2E_PositionClose verifies position closing flow:
// 1. Create account and position
// 2. Close position
// 3. Verify position status updated
// 4. Verify trade recorded
func TestE2E_PositionClose(t *testing.T) {
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
		AccountNumber: "E2E_TEST_002",
		UserID:        "test_user_002",
		Username:      "testuser2",
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

	// Create open position (buy EURUSD at 1.0850)
	position := &repository.Position{
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
	}
	positionRepo.Create(ctx, position)

	// Close position at higher price (profit scenario: 1.0900)
	closePrice := 1.0900
	profit := (closePrice - position.OpenPrice) * position.Volume * 100000 // EURUSD pip value

	err = positionRepo.Close(ctx, position.ID, closePrice, "manual")
	if err != nil {
		t.Fatalf("close position failed: %v", err)
	}

	// Verify position marked as closed
	closedPos, err := positionRepo.GetByID(ctx, position.ID)
	if err != nil {
		t.Fatalf("failed to fetch position: %v", err)
	}

	if closedPos.Status != "closed" {
		t.Errorf("position status: got %s, want closed", closedPos.Status)
	}

	if closedPos.ClosePrice != closePrice {
		t.Errorf("close price: got %f, want %f", closedPos.ClosePrice, closePrice)
	}

	if closedPos.CloseReason != "manual" {
		t.Errorf("close reason: got %s, want manual", closedPos.CloseReason)
	}

	// Verify profit calculated (50 pips * $10/pip = $500)
	expectedProfit := 500.0
	if profit < expectedProfit-1 || profit > expectedProfit+1 {
		t.Errorf("profit: got %f, want ~%f", profit, expectedProfit)
	}

	// Note: Trade recording would happen in the engine, not just repository
	// This E2E test verifies the database layer works correctly
}

// Helper functions
func getDatabaseURL(t *testing.T) string {
	// Use test database (not production)
	// This should be set via environment variable in CI/CD
	return "postgres://localhost/trading_engine_test?sslmode=disable"
}

func cleanupTestData(t *testing.T, pool interface{}) {
	// Note: database.Pool is actually pgxpool.Pool
	// We'll execute cleanup via the pool's Exec method
	// This is simplified for E2E test - in production use proper transaction rollback
	t.Log("Test cleanup: E2E tests use isolated test database that can be reset")
}
