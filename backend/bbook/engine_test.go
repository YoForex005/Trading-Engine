package bbook

import (
	"testing"
)

func TestOrderExecution_MarketOrder(t *testing.T) {
	tests := []struct {
		name              string
		side              string
		volume            float64
		accountBalance    float64
		leverage          float64
		bidPrice          float64
		askPrice          float64
		expectedExecution bool
		expectedFillPrice float64
	}{
		{
			name:              "buy market order executes",
			side:              "BUY",
			volume:            1.0,
			accountBalance:    10000.00,
			leverage:          100,
			bidPrice:          1.0850,
			askPrice:          1.0852,
			expectedExecution: true,
			expectedFillPrice: 1.0852, // Buy at ask
		},
		{
			name:              "sell market order executes",
			side:              "SELL",
			volume:            1.0,
			accountBalance:    10000.00,
			leverage:          100,
			bidPrice:          1.0850,
			askPrice:          1.0852,
			expectedExecution: true,
			expectedFillPrice: 1.0850, // Sell at bid
		},
		{
			name:              "insufficient margin rejects order",
			side:              "BUY",
			volume:            1.0,
			accountBalance:    500.00,
			leverage:          100,
			bidPrice:          1.0850,
			askPrice:          1.0852,
			expectedExecution: false,
			expectedFillPrice: 0,
		},
		{
			name:              "mini lot executes",
			side:              "BUY",
			volume:            0.1,
			accountBalance:    1000.00,
			leverage:          100,
			bidPrice:          1.0850,
			askPrice:          1.0852,
			expectedExecution: true,
			expectedFillPrice: 1.0852,
		},
		{
			name:              "high leverage allows larger position",
			side:              "BUY",
			volume:            1.0,
			accountBalance:    500.00,
			leverage:          500,
			bidPrice:          1.0850,
			askPrice:          1.0852,
			expectedExecution: true,
			expectedFillPrice: 1.0852,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			engine := NewEngine()

			// Create account with specified balance
			account := engine.CreateAccount("user1", "testuser", "pass", true)
			account.Balance = tt.accountBalance
			account.Leverage = tt.leverage
			engine.ledger.SetBalance(account.ID, tt.accountBalance)

			// Set price callback
			engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
				return tt.bidPrice, tt.askPrice, true
			})

			// Execute order
			position, err := engine.ExecuteMarketOrder(account.ID, "EURUSD", tt.side, tt.volume, 0, 0)

			if tt.expectedExecution {
				if err != nil {
					t.Fatalf("expected order to execute but got error: %v", err)
				}
				if position == nil {
					t.Fatal("expected position but got nil")
				}
				if position.OpenPrice != tt.expectedFillPrice {
					t.Errorf("fill price: got %.4f, want %.4f", position.OpenPrice, tt.expectedFillPrice)
				}
				if position.Volume != tt.volume {
					t.Errorf("position volume: got %.2f, want %.2f", position.Volume, tt.volume)
				}
				if position.Side != tt.side {
					t.Errorf("position side: got %s, want %s", position.Side, tt.side)
				}
			} else {
				if err == nil {
					t.Error("expected error but order executed")
				}
			}
		})
	}
}

func TestOrderExecution_VolumeValidation(t *testing.T) {
	tests := []struct {
		name        string
		volume      float64
		expectError bool
	}{
		{
			name:        "standard lot accepted",
			volume:      1.0,
			expectError: false,
		},
		{
			name:        "mini lot accepted",
			volume:      0.1,
			expectError: false,
		},
		{
			name:        "micro lot accepted",
			volume:      0.01,
			expectError: false,
		},
		{
			name:        "volume below minimum rejected",
			volume:      0.001,
			expectError: true,
		},
		{
			name:        "volume above maximum rejected",
			volume:      200.0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			engine := NewEngine()
			account := engine.CreateAccount("user1", "testuser", "pass", true)
			account.Balance = 100000.00
			engine.ledger.SetBalance(account.ID, 100000.00)

			engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
				return 1.0850, 1.0852, true
			})

			_, err := engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", tt.volume, 0, 0)

			if tt.expectError {
				if err == nil {
					t.Error("expected error for invalid volume but order executed")
				}
			} else {
				if err != nil {
					t.Errorf("expected order to execute but got error: %v", err)
				}
			}
		})
	}
}

func TestOrderExecution_AccountValidation(t *testing.T) {
	tests := []struct {
		name        string
		accountID   int64
		status      string
		expectError bool
	}{
		{
			name:        "active account accepted",
			accountID:   1,
			status:      "ACTIVE",
			expectError: false,
		},
		{
			name:        "disabled account rejected",
			accountID:   2,
			status:      "DISABLED",
			expectError: true,
		},
		{
			name:        "non-existent account rejected",
			accountID:   999,
			status:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			engine := NewEngine()

			// Create accounts if needed
			if tt.accountID == 1 {
				account := engine.CreateAccount("user1", "testuser", "pass", true)
				account.Balance = 10000.00
				account.Status = tt.status
				engine.ledger.SetBalance(account.ID, 10000.00)
			} else if tt.accountID == 2 {
				account := engine.CreateAccount("user2", "testuser2", "pass", true)
				account.Balance = 10000.00
				account.Status = tt.status
				engine.ledger.SetBalance(account.ID, 10000.00)
				tt.accountID = account.ID
			}

			engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
				return 1.0850, 1.0852, true
			})

			_, err := engine.ExecuteMarketOrder(tt.accountID, "EURUSD", "BUY", 1.0, 0, 0)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but order executed")
				}
			} else {
				if err != nil {
					t.Errorf("expected order to execute but got error: %v", err)
				}
			}
		})
	}
}

func TestOrderExecution_CommissionDeduction(t *testing.T) {
	engine := NewEngine()

	// Set commission for EURUSD
	spec := engine.symbols["EURUSD"]
	spec.CommissionPerLot = 7.0

	account := engine.CreateAccount("user1", "testuser", "pass", true)
	account.Balance = 10000.00
	engine.ledger.SetBalance(account.ID, 10000.00)

	engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		return 1.0850, 1.0852, true
	})

	// Execute order
	position, err := engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", 1.0, 0, 0)
	if err != nil {
		t.Fatalf("order execution failed: %v", err)
	}

	// Verify commission was calculated
	expectedCommission := 7.0 * 1.0 // $7 per lot * 1 lot
	if position.Commission != expectedCommission {
		t.Errorf("commission: got %.2f, want %.2f", position.Commission, expectedCommission)
	}

	// Verify balance was reduced by commission
	expectedBalance := 10000.00 - position.Commission
	actualBalance := engine.ledger.GetBalance(account.ID)
	tolerance := 0.01
	if abs(actualBalance-expectedBalance) > tolerance {
		t.Errorf("balance after commission: got %.2f, want %.2f", actualBalance, expectedBalance)
	}
}

func TestClosePosition(t *testing.T) {
	tests := []struct {
		name           string
		openSide       string
		openPrice      float64
		closePrice     float64
		volume         float64
		expectedProfit float64
	}{
		{
			name:           "close profitable buy position",
			openSide:       "BUY",
			openPrice:      1.0850,
			closePrice:     1.0900,
			volume:         1.0,
			expectedProfit: 500.00, // 50 pips profit
		},
		{
			name:           "close losing buy position",
			openSide:       "BUY",
			openPrice:      1.0850,
			closePrice:     1.0800,
			volume:         1.0,
			expectedProfit: -500.00, // 50 pips loss
		},
		{
			name:           "close profitable sell position",
			openSide:       "SELL",
			openPrice:      1.0850,
			closePrice:     1.0800,
			volume:         1.0,
			expectedProfit: 500.00, // 50 pips profit
		},
		{
			name:           "close losing sell position",
			openSide:       "SELL",
			openPrice:      1.0850,
			closePrice:     1.0900,
			volume:         1.0,
			expectedProfit: -500.00, // 50 pips loss
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			engine := NewEngine()
			account := engine.CreateAccount("user1", "testuser", "pass", true)
			account.Balance = 10000.00
			engine.ledger.SetBalance(account.ID, 10000.00)

			// Open position
			engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
				return tt.openPrice, tt.openPrice, true
			})
			position, err := engine.ExecuteMarketOrder(account.ID, "EURUSD", tt.openSide, tt.volume, 0, 0)
			if err != nil {
				t.Fatalf("failed to open position: %v", err)
			}

			// Close position
			engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
				return tt.closePrice, tt.closePrice, true
			})
			trade, err := engine.ClosePosition(position.ID, tt.volume)
			if err != nil {
				t.Fatalf("failed to close position: %v", err)
			}

			// Verify realized P&L
			tolerance := 1.0
			if abs(trade.RealizedPnL-tt.expectedProfit) > tolerance {
				t.Errorf("realized P&L: got %.2f, want %.2f", trade.RealizedPnL, tt.expectedProfit)
			}
		})
	}
}

func TestGetAccountSummary(t *testing.T) {
	engine := NewEngine()
	account := engine.CreateAccount("user1", "testuser", "pass", true)
	account.Balance = 10000.00
	engine.ledger.SetBalance(account.ID, 10000.00)

	engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		return 1.0850, 1.0852, true
	})

	// Get summary with no positions
	summary, err := engine.GetAccountSummary(account.ID)
	if err != nil {
		t.Fatalf("failed to get account summary: %v", err)
	}

	if summary.Balance != 10000.00 {
		t.Errorf("balance: got %.2f, want 10000.00", summary.Balance)
	}
	if summary.OpenPositions != 0 {
		t.Errorf("open positions: got %d, want 0", summary.OpenPositions)
	}

	// Open position and get summary
	position, err := engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", 1.0, 0, 0)
	if err != nil {
		t.Fatalf("failed to open position: %v", err)
	}

	summary, err = engine.GetAccountSummary(account.ID)
	if err != nil {
		t.Fatalf("failed to get account summary: %v", err)
	}

	if summary.OpenPositions != 1 {
		t.Errorf("open positions: got %d, want 1", summary.OpenPositions)
	}
	if summary.Margin <= 0 {
		t.Error("margin should be greater than 0 with open position")
	}
	if summary.FreeMargin >= 10000.00-position.Commission {
		t.Error("free margin should be reduced by used margin and commission")
	}
}
