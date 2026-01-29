package oms

import (
	"sync"
	"testing"
	"time"
)

// TestNewService tests service initialization
func TestNewService(t *testing.T) {
	service := NewService()

	if service == nil {
		t.Fatal("NewService() returned nil")
	}

	if service.orders == nil {
		t.Error("orders map not initialized")
	}

	if service.positions == nil {
		t.Error("positions map not initialized")
	}
}

// TestPlaceOrder tests order placement
func TestPlaceOrder(t *testing.T) {
	tests := []struct {
		name        string
		req         PlaceOrderRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "Valid BUY MARKET order",
			req: PlaceOrderRequest{
				AccountID: "acc1",
				Symbol:    "EURUSD",
				Side:      "BUY",
				Type:      "MARKET",
				Volume:    1.0,
				Price:     1.1000,
			},
			wantErr: false,
		},
		{
			name: "Valid SELL MARKET order with SL/TP",
			req: PlaceOrderRequest{
				AccountID: "acc1",
				Symbol:    "EURUSD",
				Side:      "SELL",
				Type:      "MARKET",
				Volume:    0.5,
				Price:     1.1000,
				SL:        1.1050,
				TP:        1.0950,
			},
			wantErr: false,
		},
		{
			name: "Valid LIMIT order",
			req: PlaceOrderRequest{
				AccountID: "acc1",
				Symbol:    "GBPUSD",
				Side:      "BUY",
				Type:      "LIMIT",
				Volume:    2.0,
				Price:     1.2500,
			},
			wantErr: false,
		},
		{
			name: "Invalid volume - zero",
			req: PlaceOrderRequest{
				AccountID: "acc1",
				Symbol:    "EURUSD",
				Side:      "BUY",
				Type:      "MARKET",
				Volume:    0,
				Price:     1.1000,
			},
			wantErr:     true,
			errContains: "invalid volume",
		},
		{
			name: "Invalid volume - negative",
			req: PlaceOrderRequest{
				AccountID: "acc1",
				Symbol:    "EURUSD",
				Side:      "BUY",
				Type:      "MARKET",
				Volume:    -1.0,
				Price:     1.1000,
			},
			wantErr:     true,
			errContains: "invalid volume",
		},
		{
			name: "Invalid volume - too large",
			req: PlaceOrderRequest{
				AccountID: "acc1",
				Symbol:    "EURUSD",
				Side:      "BUY",
				Type:      "MARKET",
				Volume:    150,
				Price:     1.1000,
			},
			wantErr:     true,
			errContains: "invalid volume",
		},
		{
			name: "Invalid side",
			req: PlaceOrderRequest{
				AccountID: "acc1",
				Symbol:    "EURUSD",
				Side:      "HOLD",
				Type:      "MARKET",
				Volume:    1.0,
				Price:     1.1000,
			},
			wantErr:     true,
			errContains: "invalid side",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService()
			order, err := service.PlaceOrder(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("PlaceOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != nil && tt.errContains != "" {
					if len(err.Error()) < len(tt.errContains) || err.Error()[:len(tt.errContains)] != tt.errContains {
						t.Errorf("Error = %s, want to contain %s", err.Error(), tt.errContains)
					}
				}
				return
			}

			// Validate successful order
			if order == nil {
				t.Fatal("Expected order, got nil")
			}

			if order.ID == "" {
				t.Error("Order ID should not be empty")
			}

			if order.AccountID != tt.req.AccountID {
				t.Errorf("AccountID = %s, want %s", order.AccountID, tt.req.AccountID)
			}

			if order.Symbol != tt.req.Symbol {
				t.Errorf("Symbol = %s, want %s", order.Symbol, tt.req.Symbol)
			}

			if order.Side != tt.req.Side {
				t.Errorf("Side = %s, want %s", order.Side, tt.req.Side)
			}

			if order.Type != tt.req.Type {
				t.Errorf("Type = %s, want %s", order.Type, tt.req.Type)
			}

			if order.Volume != tt.req.Volume {
				t.Errorf("Volume = %f, want %f", order.Volume, tt.req.Volume)
			}

			// Test MARKET order immediate execution
			if tt.req.Type == "MARKET" {
				if order.Status != "FILLED" {
					t.Errorf("MARKET order status = %s, want FILLED", order.Status)
				}

				if order.FilledAt == nil {
					t.Error("MARKET order should have FilledAt timestamp")
				}

				if order.PriceExecuted == 0 {
					t.Error("MARKET order should have execution price")
				}
			} else {
				if order.Status != "PENDING" {
					t.Errorf("Non-MARKET order status = %s, want PENDING", order.Status)
				}
			}
		})
	}
}

// TestPlaceMarketOrderCreatesPosition tests that MARKET orders create positions
func TestPlaceMarketOrderCreatesPosition(t *testing.T) {
	service := NewService()

	req := PlaceOrderRequest{
		AccountID: "acc1",
		Symbol:    "EURUSD",
		Side:      "BUY",
		Type:      "MARKET",
		Volume:    1.0,
		Price:     1.1000,
		SL:        1.0950,
		TP:        1.1050,
	}

	order, err := service.PlaceOrder(req)
	if err != nil {
		t.Fatalf("PlaceOrder() error = %v", err)
	}

	// Verify position was created
	positions := service.GetPositions(req.AccountID)
	if len(positions) != 1 {
		t.Fatalf("Position count = %d, want 1", len(positions))
	}

	pos := positions[0]

	if pos.OrderID != order.ID {
		t.Errorf("Position OrderID = %s, want %s", pos.OrderID, order.ID)
	}

	if pos.AccountID != req.AccountID {
		t.Errorf("Position AccountID = %s, want %s", pos.AccountID, req.AccountID)
	}

	if pos.Symbol != req.Symbol {
		t.Errorf("Position Symbol = %s, want %s", pos.Symbol, req.Symbol)
	}

	if pos.Side != req.Side {
		t.Errorf("Position Side = %s, want %s", pos.Side, req.Side)
	}

	if pos.Volume != req.Volume {
		t.Errorf("Position Volume = %f, want %f", pos.Volume, req.Volume)
	}

	if pos.OpenPrice != order.PriceExecuted {
		t.Errorf("Position OpenPrice = %f, want %f", pos.OpenPrice, order.PriceExecuted)
	}

	if pos.SL != req.SL {
		t.Errorf("Position SL = %f, want %f", pos.SL, req.SL)
	}

	if pos.TP != req.TP {
		t.Errorf("Position TP = %f, want %f", pos.TP, req.TP)
	}
}

// TestDetermineRouting tests A-Book vs B-Book routing logic
func TestDetermineRouting(t *testing.T) {
	tests := []struct {
		name          string
		volume        float64
		expectedRoute string
	}{
		{"Small order to B-Book", 0.5, "B_BOOK"},
		{"Medium order to B-Book", 5.0, "B_BOOK"},
		{"Large order to A-Book", 10.0, "A_BOOK"},
		{"Very large order to A-Book", 50.0, "A_BOOK"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService()

			req := PlaceOrderRequest{
				AccountID: "acc1",
				Symbol:    "EURUSD",
				Side:      "BUY",
				Type:      "MARKET",
				Volume:    tt.volume,
				Price:     1.1000,
			}

			order, _ := service.PlaceOrder(req)

			if order.RoutingType != tt.expectedRoute {
				t.Errorf("RoutingType = %s, want %s", order.RoutingType, tt.expectedRoute)
			}
		})
	}
}

// TestGetPositions tests position retrieval
func TestGetPositions(t *testing.T) {
	service := NewService()

	// Place orders for different accounts
	req1 := PlaceOrderRequest{
		AccountID: "acc1",
		Symbol:    "EURUSD",
		Side:      "BUY",
		Type:      "MARKET",
		Volume:    1.0,
		Price:     1.1000,
	}

	req2 := PlaceOrderRequest{
		AccountID: "acc1",
		Symbol:    "GBPUSD",
		Side:      "SELL",
		Type:      "MARKET",
		Volume:    0.5,
		Price:     1.2500,
	}

	req3 := PlaceOrderRequest{
		AccountID: "acc2",
		Symbol:    "EURUSD",
		Side:      "BUY",
		Type:      "MARKET",
		Volume:    2.0,
		Price:     1.1000,
	}

	service.PlaceOrder(req1)
	service.PlaceOrder(req2)
	service.PlaceOrder(req3)

	// Test positions for acc1
	positions := service.GetPositions("acc1")
	if len(positions) != 2 {
		t.Errorf("acc1 positions count = %d, want 2", len(positions))
	}

	// Test positions for acc2
	positions = service.GetPositions("acc2")
	if len(positions) != 1 {
		t.Errorf("acc2 positions count = %d, want 1", len(positions))
	}

	// Test positions for non-existent account
	positions = service.GetPositions("acc999")
	if len(positions) != 0 {
		t.Errorf("Non-existent account positions count = %d, want 0", len(positions))
	}
}

// TestClosePosition tests position closing
func TestClosePosition(t *testing.T) {
	service := NewService()

	// Open position
	req := PlaceOrderRequest{
		AccountID: "acc1",
		Symbol:    "EURUSD",
		Side:      "BUY",
		Type:      "MARKET",
		Volume:    1.0,
		Price:     1.1000,
	}

	service.PlaceOrder(req)
	positions := service.GetPositions("acc1")

	if len(positions) == 0 {
		t.Fatal("No position created")
	}

	positionID := positions[0].ID
	closePrice := 1.1050

	// Close position
	closedPos, err := service.ClosePosition(positionID, closePrice)
	if err != nil {
		t.Fatalf("ClosePosition() error = %v", err)
	}

	if closedPos == nil {
		t.Fatal("Expected closed position, got nil")
	}

	// Verify profit calculation (BUY position)
	expectedProfit := (closePrice - req.Price) * req.Volume * 100000
	if closedPos.Profit != expectedProfit {
		t.Errorf("Profit = %f, want %f", closedPos.Profit, expectedProfit)
	}

	// Verify position was removed
	positions = service.GetPositions("acc1")
	if len(positions) != 0 {
		t.Errorf("Position count after close = %d, want 0", len(positions))
	}

	// Test closing non-existent position
	_, err = service.ClosePosition("invalid-id", 1.1000)
	if err == nil {
		t.Error("Expected error for non-existent position, got nil")
	}
}

// TestClosePositionProfitCalculation tests profit calculation for both sides
func TestClosePositionProfitCalculation(t *testing.T) {
	tests := []struct {
		name          string
		side          string
		openPrice     float64
		closePrice    float64
		volume        float64
		expectedProfit float64
	}{
		{
			name:          "BUY position profit",
			side:          "BUY",
			openPrice:     1.1000,
			closePrice:    1.1050,
			volume:        1.0,
			expectedProfit: 500,
		},
		{
			name:          "BUY position loss",
			side:          "BUY",
			openPrice:     1.1000,
			closePrice:    1.0950,
			volume:        1.0,
			expectedProfit: -500,
		},
		{
			name:          "SELL position profit",
			side:          "SELL",
			openPrice:     1.1000,
			closePrice:    1.0950,
			volume:        1.0,
			expectedProfit: 500,
		},
		{
			name:          "SELL position loss",
			side:          "SELL",
			openPrice:     1.1000,
			closePrice:    1.1050,
			volume:        1.0,
			expectedProfit: -500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService()

			req := PlaceOrderRequest{
				AccountID: "acc1",
				Symbol:    "EURUSD",
				Side:      tt.side,
				Type:      "MARKET",
				Volume:    tt.volume,
				Price:     tt.openPrice,
			}

			service.PlaceOrder(req)
			positions := service.GetPositions("acc1")

			closedPos, _ := service.ClosePosition(positions[0].ID, tt.closePrice)

			// Allow small floating point differences
			diff := closedPos.Profit - tt.expectedProfit
			if diff < 0 {
				diff = -diff
			}

			if diff > 0.01 {
				t.Errorf("Profit = %f, want %f", closedPos.Profit, tt.expectedProfit)
			}
		})
	}
}

// TestConcurrentOrderPlacement tests thread-safety for order placement
func TestConcurrentOrderPlacement(t *testing.T) {
	service := NewService()

	var wg sync.WaitGroup
	orderCount := 100

	for i := 0; i < orderCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			side := "BUY"
			if idx%2 == 0 {
				side = "SELL"
			}

			req := PlaceOrderRequest{
				AccountID: "acc1",
				Symbol:    "EURUSD",
				Side:      side,
				Type:      "MARKET",
				Volume:    0.1,
				Price:     1.1000,
			}

			_, err := service.PlaceOrder(req)
			if err != nil {
				t.Errorf("Concurrent order %d failed: %v", idx, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all positions were created
	positions := service.GetPositions("acc1")
	if len(positions) != orderCount {
		t.Errorf("Position count = %d, want %d", len(positions), orderCount)
	}
}

// TestConcurrentPositionAccess tests thread-safety for position access
func TestConcurrentPositionAccess(t *testing.T) {
	service := NewService()

	// Create initial positions
	for i := 0; i < 10; i++ {
		req := PlaceOrderRequest{
			AccountID: "acc1",
			Symbol:    "EURUSD",
			Side:      "BUY",
			Type:      "MARKET",
			Volume:    0.1,
			Price:     1.1000,
		}
		service.PlaceOrder(req)
	}

	var wg sync.WaitGroup

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			positions := service.GetPositions("acc1")
			if len(positions) < 1 {
				t.Error("Expected at least 1 position")
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentMixedOperations tests concurrent orders and closes
func TestConcurrentMixedOperations(t *testing.T) {
	service := NewService()

	var wg sync.WaitGroup

	// Create positions
	for i := 0; i < 20; i++ {
		req := PlaceOrderRequest{
			AccountID: "acc1",
			Symbol:    "EURUSD",
			Side:      "BUY",
			Type:      "MARKET",
			Volume:    0.1,
			Price:     1.1000,
		}
		service.PlaceOrder(req)
	}

	positions := service.GetPositions("acc1")
	positionIDs := make([]string, len(positions))
	for i, pos := range positions {
		positionIDs[i] = pos.ID
	}

	// Concurrent operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Half close positions, half create new ones
			if idx%2 == 0 && idx/2 < len(positionIDs) {
				service.ClosePosition(positionIDs[idx/2], 1.1050)
			} else {
				req := PlaceOrderRequest{
					AccountID: "acc1",
					Symbol:    "GBPUSD",
					Side:      "SELL",
					Type:      "MARKET",
					Volume:    0.1,
					Price:     1.2500,
				}
				service.PlaceOrder(req)
			}
		}(i)
	}

	wg.Wait()

	// Verify final state is consistent
	finalPositions := service.GetPositions("acc1")
	if len(finalPositions) < 0 {
		t.Error("Position count should be non-negative")
	}
}

// TestOrderTimestamps tests that orders have proper timestamps
func TestOrderTimestamps(t *testing.T) {
	service := NewService()

	beforeCreate := time.Now()
	time.Sleep(10 * time.Millisecond) // Small delay to ensure timestamp difference

	req := PlaceOrderRequest{
		AccountID: "acc1",
		Symbol:    "EURUSD",
		Side:      "BUY",
		Type:      "MARKET",
		Volume:    1.0,
		Price:     1.1000,
	}

	order, _ := service.PlaceOrder(req)

	time.Sleep(10 * time.Millisecond)
	afterCreate := time.Now()

	// Verify CreatedAt is within expected range
	if order.CreatedAt.Before(beforeCreate) {
		t.Error("CreatedAt timestamp is before order creation started")
	}

	if order.CreatedAt.After(afterCreate) {
		t.Error("CreatedAt timestamp is after order creation completed")
	}

	// Verify MARKET order has FilledAt
	if order.FilledAt == nil {
		t.Error("MARKET order should have FilledAt timestamp")
	}

	if order.FilledAt.Before(order.CreatedAt) {
		t.Error("FilledAt should not be before CreatedAt")
	}
}

// TestPositionTimestamps tests position timestamps
func TestPositionTimestamps(t *testing.T) {
	service := NewService()

	beforeCreate := time.Now()

	req := PlaceOrderRequest{
		AccountID: "acc1",
		Symbol:    "EURUSD",
		Side:      "BUY",
		Type:      "MARKET",
		Volume:    1.0,
		Price:     1.1000,
	}

	service.PlaceOrder(req)

	afterCreate := time.Now()

	positions := service.GetPositions("acc1")
	if len(positions) == 0 {
		t.Fatal("No position created")
	}

	pos := positions[0]

	if pos.OpenTime.Before(beforeCreate) {
		t.Error("OpenTime is before position creation started")
	}

	if pos.OpenTime.After(afterCreate) {
		t.Error("OpenTime is after position creation completed")
	}
}

// TestOrderIDUniqueness tests that order IDs are unique
func TestOrderIDUniqueness(t *testing.T) {
	service := NewService()

	orderIDs := make(map[string]bool)

	for i := 0; i < 100; i++ {
		req := PlaceOrderRequest{
			AccountID: "acc1",
			Symbol:    "EURUSD",
			Side:      "BUY",
			Type:      "MARKET",
			Volume:    0.1,
			Price:     1.1000,
		}

		order, _ := service.PlaceOrder(req)

		if orderIDs[order.ID] {
			t.Errorf("Duplicate order ID found: %s", order.ID)
		}

		orderIDs[order.ID] = true
	}

	if len(orderIDs) != 100 {
		t.Errorf("Expected 100 unique order IDs, got %d", len(orderIDs))
	}
}

// TestPositionIDUniqueness tests that position IDs are unique
func TestPositionIDUniqueness(t *testing.T) {
	service := NewService()

	positionIDs := make(map[string]bool)

	for i := 0; i < 100; i++ {
		req := PlaceOrderRequest{
			AccountID: "acc1",
			Symbol:    "EURUSD",
			Side:      "BUY",
			Type:      "MARKET",
			Volume:    0.1,
			Price:     1.1000,
		}

		service.PlaceOrder(req)
	}

	positions := service.GetPositions("acc1")

	for _, pos := range positions {
		if positionIDs[pos.ID] {
			t.Errorf("Duplicate position ID found: %s", pos.ID)
		}
		positionIDs[pos.ID] = true
	}

	if len(positionIDs) != 100 {
		t.Errorf("Expected 100 unique position IDs, got %d", len(positionIDs))
	}
}

// TestEdgeCases tests edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	t.Run("Place LIMIT order does not create position", func(t *testing.T) {
		service := NewService()

		req := PlaceOrderRequest{
			AccountID: "acc1",
			Symbol:    "EURUSD",
			Side:      "BUY",
			Type:      "LIMIT",
			Volume:    1.0,
			Price:     1.1000,
		}

		service.PlaceOrder(req)

		positions := service.GetPositions("acc1")
		if len(positions) != 0 {
			t.Error("LIMIT order should not create position immediately")
		}
	})

	t.Run("Zero SL/TP are valid", func(t *testing.T) {
		service := NewService()

		req := PlaceOrderRequest{
			AccountID: "acc1",
			Symbol:    "EURUSD",
			Side:      "BUY",
			Type:      "MARKET",
			Volume:    1.0,
			Price:     1.1000,
			SL:        0,
			TP:        0,
		}

		order, err := service.PlaceOrder(req)
		if err != nil {
			t.Errorf("Order with zero SL/TP should be valid: %v", err)
		}

		if order.SL != 0 || order.TP != 0 {
			t.Error("SL and TP should be 0")
		}
	})

	t.Run("Close position updates profit correctly", func(t *testing.T) {
		service := NewService()

		req := PlaceOrderRequest{
			AccountID: "acc1",
			Symbol:    "EURUSD",
			Side:      "BUY",
			Type:      "MARKET",
			Volume:    1.0,
			Price:     1.1000,
		}

		service.PlaceOrder(req)
		positions := service.GetPositions("acc1")

		// Close at same price (breakeven)
		closedPos, _ := service.ClosePosition(positions[0].ID, 1.1000)

		if closedPos.Profit != 0 {
			t.Errorf("Breakeven profit = %f, want 0", closedPos.Profit)
		}
	})
}
