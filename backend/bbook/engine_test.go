package bbook

import (
	"sync"
	"testing"
)

// TestNewEngine tests engine initialization
func TestNewEngine(t *testing.T) {
	engine := NewEngine()

	if engine == nil {
		t.Fatal("NewEngine() returned nil")
	}

	if engine.accounts == nil {
		t.Error("accounts map not initialized")
	}

	if engine.positions == nil {
		t.Error("positions map not initialized")
	}

	if engine.orders == nil {
		t.Error("orders map not initialized")
	}

	if engine.symbols == nil {
		t.Error("symbols map not initialized")
	}

	// Verify default symbols are loaded
	symbols := engine.GetSymbols()
	if len(symbols) == 0 {
		t.Error("no default symbols initialized")
	}

	// Check specific symbols
	found := false
	for _, s := range symbols {
		if s.Symbol == "EURUSD" {
			found = true
			if s.ContractSize != 100000 {
				t.Errorf("EURUSD contract size = %f, want 100000", s.ContractSize)
			}
			break
		}
	}

	if !found {
		t.Error("EURUSD symbol not found in default symbols")
	}
}

// TestCreateAccount tests account creation
func TestCreateAccount(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		username string
		password string
		isDemo   bool
	}{
		{
			name:     "Live account with username",
			userID:   "user123",
			username: "trader1",
			password: "pass123",
			isDemo:   false,
		},
		{
			name:     "Demo account",
			userID:   "user456",
			username: "demo_user",
			password: "demo123",
			isDemo:   true,
		},
		{
			name:     "Account without username",
			userID:   "user789",
			username: "",
			password: "pass789",
			isDemo:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			account := engine.CreateAccount(tt.userID, tt.username, tt.password, tt.isDemo)

			if account == nil {
				t.Fatal("CreateAccount() returned nil")
			}

			if account.UserID != tt.userID {
				t.Errorf("UserID = %s, want %s", account.UserID, tt.userID)
			}

			if account.IsDemo != tt.isDemo {
				t.Errorf("IsDemo = %v, want %v", account.IsDemo, tt.isDemo)
			}

			if account.Balance != 0 {
				t.Errorf("Initial balance = %f, want 0", account.Balance)
			}

			if account.Leverage != 100 {
				t.Errorf("Default leverage = %f, want 100", account.Leverage)
			}

			if account.Status != "ACTIVE" {
				t.Errorf("Status = %s, want ACTIVE", account.Status)
			}

			// Test username defaulting
			if tt.username == "" && account.Username == "" {
				t.Error("Username should default to account number when empty")
			}
		})
	}
}

// TestGetAccount tests account retrieval
func TestGetAccount(t *testing.T) {
	engine := NewEngine()
	account := engine.CreateAccount("user1", "trader1", "pass1", false)

	// Test successful retrieval
	retrieved, ok := engine.GetAccount(account.ID)
	if !ok {
		t.Fatal("GetAccount() failed to retrieve existing account")
	}

	if retrieved.ID != account.ID {
		t.Errorf("Retrieved account ID = %d, want %d", retrieved.ID, account.ID)
	}

	// Test non-existent account
	_, ok = engine.GetAccount(999999)
	if ok {
		t.Error("GetAccount() should return false for non-existent account")
	}
}

// TestUpdatePassword tests password updates
func TestUpdatePassword(t *testing.T) {
	engine := NewEngine()
	account := engine.CreateAccount("user1", "trader1", "oldpass", false)

	newPassword := "newpass123"
	err := engine.UpdatePassword(account.ID, newPassword)
	if err != nil {
		t.Fatalf("UpdatePassword() error = %v", err)
	}

	// Verify password was updated
	retrieved, _ := engine.GetAccount(account.ID)
	if retrieved.Password != newPassword {
		t.Errorf("Password = %s, want %s", retrieved.Password, newPassword)
	}

	// Test updating non-existent account
	err = engine.UpdatePassword(999999, "anypass")
	if err == nil {
		t.Error("UpdatePassword() should return error for non-existent account")
	}
}

// TestUpdateAccount tests account configuration updates
func TestUpdateAccount(t *testing.T) {
	tests := []struct {
		name       string
		leverage   float64
		marginMode string
		wantErr    bool
	}{
		{"Update leverage", 200, "", false},
		{"Update margin mode", 0, "NETTING", false},
		{"Update both", 500, "HEDGING", false},
		{"Invalid account", 100, "NETTING", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			var accountID int64

			if !tt.wantErr {
				account := engine.CreateAccount("user1", "trader1", "pass1", false)
				accountID = account.ID
			} else {
				accountID = 999999 // Non-existent
			}

			err := engine.UpdateAccount(accountID, tt.leverage, tt.marginMode)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				retrieved, _ := engine.GetAccount(accountID)
				if tt.leverage > 0 && retrieved.Leverage != tt.leverage {
					t.Errorf("Leverage = %f, want %f", retrieved.Leverage, tt.leverage)
				}
				if tt.marginMode != "" && retrieved.MarginMode != tt.marginMode {
					t.Errorf("MarginMode = %s, want %s", retrieved.MarginMode, tt.marginMode)
				}
			}
		})
	}
}

// TestExecuteMarketOrder tests market order execution
func TestExecuteMarketOrder(t *testing.T) {
	tests := []struct {
		name      string
		symbol    string
		side      string
		volume    float64
		sl        float64
		tp        float64
		wantErr   bool
		errString string
	}{
		{"Valid BUY order", "EURUSD", "BUY", 0.1, 0, 0, false, ""},
		{"Valid SELL order", "EURUSD", "SELL", 0.5, 1.0500, 1.1000, false, ""},
		{"Invalid volume - too small", "EURUSD", "BUY", 0.001, 0, 0, true, "volume must be between"},
		{"Invalid volume - too large", "EURUSD", "BUY", 200, 0, 0, true, "volume must be between"},
		{"Invalid side", "EURUSD", "HOLD", 0.1, 0, 0, true, "invalid side"},
		{"Unknown symbol", "UNKNOWN", "BUY", 0.1, 0, 0, true, "symbol UNKNOWN not found"},
		{"No price feed", "BTCUSD", "BUY", 0.01, 0, 0, true, "price feed not available"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			account := engine.CreateAccount("user1", "trader1", "pass1", false)
			account.Balance = 10000 // Fund account

			// Set price callback for valid symbols
			if tt.symbol != "UNKNOWN" && tt.name != "No price feed" {
				engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
					if symbol == "EURUSD" {
						return 1.1000, 1.1002, true
					}
					return 0, 0, false
				})
			}

			position, err := engine.ExecuteMarketOrder(account.ID, tt.symbol, tt.side, tt.volume, tt.sl, tt.tp)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteMarketOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errString != "" && err != nil {
				if len(err.Error()) < len(tt.errString) || err.Error()[:len(tt.errString)] != tt.errString {
					t.Errorf("Error message = %s, want to contain %s", err.Error(), tt.errString)
				}
			}

			if !tt.wantErr {
				if position == nil {
					t.Fatal("Expected position, got nil")
				}

				if position.Symbol != tt.symbol {
					t.Errorf("Position symbol = %s, want %s", position.Symbol, tt.symbol)
				}

				if position.Side != tt.side {
					t.Errorf("Position side = %s, want %s", position.Side, tt.side)
				}

				if position.Volume != tt.volume {
					t.Errorf("Position volume = %f, want %f", position.Volume, tt.volume)
				}

				if position.Status != "OPEN" {
					t.Errorf("Position status = %s, want OPEN", position.Status)
				}
			}
		})
	}
}

// TestExecuteMarketOrderInsufficientMargin tests margin validation
func TestExecuteMarketOrderInsufficientMargin(t *testing.T) {
	engine := NewEngine()
	account := engine.CreateAccount("user1", "trader1", "pass1", false)
	account.Balance = 100 // Insufficient for large order

	engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		return 1.1000, 1.1002, true
	})

	_, err := engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", 10.0, 0, 0)
	if err == nil {
		t.Error("Expected insufficient margin error, got nil")
	}

	if err != nil && len(err.Error()) < 19 || err.Error()[:19] != "insufficient margin" {
		t.Errorf("Error = %s, want insufficient margin error", err.Error())
	}
}

// TestClosePosition tests position closing
func TestClosePosition(t *testing.T) {
	engine := NewEngine()
	account := engine.CreateAccount("user1", "trader1", "pass1", false)
	account.Balance = 10000

	engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		return 1.1000, 1.1002, true
	})

	// Open position
	position, err := engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", 0.1, 0, 0)
	if err != nil {
		t.Fatalf("Failed to open position: %v", err)
	}

	// Close position
	trade, err := engine.ClosePosition(position.ID, 0)
	if err != nil {
		t.Fatalf("ClosePosition() error = %v", err)
	}

	if trade == nil {
		t.Fatal("Expected trade, got nil")
	}

	// Verify position is closed
	closedPos, _ := engine.positions[position.ID]
	if closedPos.Status != "CLOSED" {
		t.Errorf("Position status = %s, want CLOSED", closedPos.Status)
	}

	// Test closing non-existent position
	_, err = engine.ClosePosition(999999, 0)
	if err == nil {
		t.Error("Expected error for non-existent position, got nil")
	}

	// Test closing already closed position
	_, err = engine.ClosePosition(position.ID, 0)
	if err == nil {
		t.Error("Expected error for already closed position, got nil")
	}
}

// TestPartialClose tests partial position closing
func TestPartialClose(t *testing.T) {
	engine := NewEngine()
	account := engine.CreateAccount("user1", "trader1", "pass1", false)
	account.Balance = 10000

	engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		return 1.1000, 1.1002, true
	})

	// Open position with 1.0 lot
	position, _ := engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", 1.0, 0, 0)

	// Close 0.5 lot
	_, err := engine.ClosePosition(position.ID, 0.5)
	if err != nil {
		t.Fatalf("Partial close error = %v", err)
	}

	// Verify remaining volume
	pos, _ := engine.positions[position.ID]
	if pos.Volume != 0.5 {
		t.Errorf("Remaining volume = %f, want 0.5", pos.Volume)
	}

	if pos.Status != "OPEN" {
		t.Errorf("Position status = %s, want OPEN", pos.Status)
	}
}

// TestModifyPosition tests SL/TP modification
func TestModifyPosition(t *testing.T) {
	engine := NewEngine()
	account := engine.CreateAccount("user1", "trader1", "pass1", false)
	account.Balance = 10000

	engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		return 1.1000, 1.1002, true
	})

	position, _ := engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", 0.1, 0, 0)

	newSL := 1.0950
	newTP := 1.1050

	modified, err := engine.ModifyPosition(position.ID, newSL, newTP)
	if err != nil {
		t.Fatalf("ModifyPosition() error = %v", err)
	}

	if modified.SL != newSL {
		t.Errorf("SL = %f, want %f", modified.SL, newSL)
	}

	if modified.TP != newTP {
		t.Errorf("TP = %f, want %f", modified.TP, newTP)
	}

	// Test modifying non-existent position
	_, err = engine.ModifyPosition(999999, 1.0, 1.1)
	if err == nil {
		t.Error("Expected error for non-existent position, got nil")
	}
}

// TestGetPositions tests position retrieval
func TestGetPositions(t *testing.T) {
	engine := NewEngine()
	account1 := engine.CreateAccount("user1", "trader1", "pass1", false)
	account2 := engine.CreateAccount("user2", "trader2", "pass2", false)
	account1.Balance = 10000
	account2.Balance = 10000

	engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		return 1.1000, 1.1002, true
	})

	// Open positions for account1
	engine.ExecuteMarketOrder(account1.ID, "EURUSD", "BUY", 0.1, 0, 0)
	engine.ExecuteMarketOrder(account1.ID, "EURUSD", "SELL", 0.2, 0, 0)

	// Open position for account2
	engine.ExecuteMarketOrder(account2.ID, "EURUSD", "BUY", 0.3, 0, 0)

	// Get positions for account1
	positions := engine.GetPositions(account1.ID)
	if len(positions) != 2 {
		t.Errorf("Account1 positions count = %d, want 2", len(positions))
	}

	// Get positions for account2
	positions = engine.GetPositions(account2.ID)
	if len(positions) != 1 {
		t.Errorf("Account2 positions count = %d, want 1", len(positions))
	}

	// Test GetAllPositions
	allPositions := engine.GetAllPositions()
	if len(allPositions) != 3 {
		t.Errorf("All positions count = %d, want 3", len(allPositions))
	}
}

// TestGetAccountSummary tests account summary calculation
func TestGetAccountSummary(t *testing.T) {
	engine := NewEngine()
	account := engine.CreateAccount("user1", "trader1", "pass1", false)
	account.Balance = 10000

	engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		return 1.1000, 1.1002, true
	})

	// Get summary with no positions
	summary, err := engine.GetAccountSummary(account.ID)
	if err != nil {
		t.Fatalf("GetAccountSummary() error = %v", err)
	}

	if summary.Balance != 10000 {
		t.Errorf("Balance = %f, want 10000", summary.Balance)
	}

	if summary.OpenPositions != 0 {
		t.Errorf("OpenPositions = %d, want 0", summary.OpenPositions)
	}

	// Open a position
	engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", 0.1, 0, 0)

	// Get summary with position
	summary, _ = engine.GetAccountSummary(account.ID)
	if summary.OpenPositions != 1 {
		t.Errorf("OpenPositions = %d, want 1", summary.OpenPositions)
	}

	if summary.Margin <= 0 {
		t.Error("Margin should be greater than 0 with open position")
	}

	// Test non-existent account
	_, err = engine.GetAccountSummary(999999)
	if err == nil {
		t.Error("Expected error for non-existent account, got nil")
	}
}

// TestConcurrentOrderExecution tests thread-safety
func TestConcurrentOrderExecution(t *testing.T) {
	engine := NewEngine()
	account := engine.CreateAccount("user1", "trader1", "pass1", false)
	account.Balance = 100000

	engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		return 1.1000, 1.1002, true
	})

	var wg sync.WaitGroup
	orderCount := 50

	for i := 0; i < orderCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			side := "BUY"
			if idx%2 == 0 {
				side = "SELL"
			}
			_, err := engine.ExecuteMarketOrder(account.ID, "EURUSD", side, 0.01, 0, 0)
			if err != nil {
				t.Errorf("Concurrent order %d failed: %v", idx, err)
			}
		}(i)
	}

	wg.Wait()

	positions := engine.GetPositions(account.ID)
	if len(positions) != orderCount {
		t.Errorf("Position count = %d, want %d", len(positions), orderCount)
	}
}

// TestConcurrentPositionClose tests concurrent position closing
func TestConcurrentPositionClose(t *testing.T) {
	engine := NewEngine()
	account := engine.CreateAccount("user1", "trader1", "pass1", false)
	account.Balance = 100000

	engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		return 1.1000, 1.1002, true
	})

	// Open multiple positions
	var positionIDs []int64
	for i := 0; i < 20; i++ {
		pos, _ := engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", 0.01, 0, 0)
		positionIDs = append(positionIDs, pos.ID)
	}

	// Close positions concurrently
	var wg sync.WaitGroup
	for _, id := range positionIDs {
		wg.Add(1)
		go func(posID int64) {
			defer wg.Done()
			_, err := engine.ClosePosition(posID, 0)
			if err != nil {
				t.Errorf("Failed to close position %d: %v", posID, err)
			}
		}(id)
	}

	wg.Wait()

	positions := engine.GetPositions(account.ID)
	if len(positions) != 0 {
		t.Errorf("Open positions = %d, want 0", len(positions))
	}
}

// TestUpdatePositionPrices tests price updates
func TestUpdatePositionPrices(t *testing.T) {
	engine := NewEngine()
	account := engine.CreateAccount("user1", "trader1", "pass1", false)
	account.Balance = 10000

	initialBid := 1.1000
	initialAsk := 1.1002

	engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		return initialBid, initialAsk, true
	})

	position, _ := engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", 0.1, 0, 0)

	// Update callback to new price
	newBid := 1.1050
	newAsk := 1.1052
	engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		return newBid, newAsk, true
	})

	// Update prices
	engine.UpdatePositionPrices()

	// Verify price update
	updated, _ := engine.positions[position.ID]
	if updated.CurrentPrice != newBid {
		t.Errorf("Current price = %f, want %f", updated.CurrentPrice, newBid)
	}

	// Verify P/L was calculated
	if updated.UnrealizedPnL == 0 {
		t.Error("UnrealizedPnL should be non-zero after price update")
	}
}

// TestSymbolOperations tests symbol management
func TestSymbolOperations(t *testing.T) {
	engine := NewEngine()

	// Test GetSymbols
	symbols := engine.GetSymbols()
	if len(symbols) == 0 {
		t.Error("GetSymbols() returned empty list")
	}

	// Test UpdateSymbol
	newSymbol := &SymbolSpec{
		Symbol:           "TEST",
		ContractSize:     100,
		PipSize:          0.01,
		PipValue:         1,
		MinVolume:        0.1,
		MaxVolume:        10,
		VolumeStep:       0.1,
		MarginPercent:    5,
		CommissionPerLot: 0.5,
	}

	engine.UpdateSymbol(newSymbol)

	symbols = engine.GetSymbols()
	found := false
	for _, s := range symbols {
		if s.Symbol == "TEST" {
			found = true
			if s.ContractSize != 100 {
				t.Errorf("TEST contract size = %f, want 100", s.ContractSize)
			}
			break
		}
	}

	if !found {
		t.Error("Updated symbol not found")
	}
}

// TestPnLCalculation tests P/L calculation logic
func TestPnLCalculation(t *testing.T) {
	engine := NewEngine()

	spec := &SymbolSpec{
		PipSize:  0.0001,
		PipValue: 10,
	}

	tests := []struct {
		name         string
		side         string
		openPrice    float64
		currentPrice float64
		volume       float64
		expectedPnL  float64
	}{
		{"BUY profit", "BUY", 1.1000, 1.1050, 1.0, 500},
		{"BUY loss", "BUY", 1.1000, 1.0950, 1.0, -500},
		{"SELL profit", "SELL", 1.1000, 1.0950, 1.0, 500},
		{"SELL loss", "SELL", 1.1000, 1.1050, 1.0, -500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := &Position{
				Side:      tt.side,
				OpenPrice: tt.openPrice,
				Volume:    tt.volume,
			}

			pnl := engine.calculatePnL(pos, tt.currentPrice, tt.volume, spec)

			// Allow small floating point difference
			diff := pnl - tt.expectedPnL
			if diff < 0 {
				diff = -diff
			}

			if diff > 0.01 {
				t.Errorf("P/L = %f, want %f (diff: %f)", pnl, tt.expectedPnL, diff)
			}
		})
	}
}

// TestMarginCalculation tests margin calculation
func TestMarginCalculation(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name           string
		symbol         string
		volume         float64
		price          float64
		leverage       float64
		expectedMargin float64
	}{
		{"EURUSD standard", "EURUSD", 1.0, 1.1000, 100, 1100},
		{"EURUSD high leverage", "EURUSD", 1.0, 1.1000, 500, 220},
		{"ETHUSD crypto", "ETHUSD", 1.0, 2000, 100, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			margin := engine.calculateMargin(tt.symbol, tt.volume, tt.price, tt.leverage)

			// Allow small floating point difference
			diff := margin - tt.expectedMargin
			if diff < 0 {
				diff = -diff
			}

			if diff > 0.01 {
				t.Errorf("Margin = %f, want %f", margin, tt.expectedMargin)
			}
		})
	}
}

// TestLedgerIntegration tests ledger recording
func TestLedgerIntegration(t *testing.T) {
	engine := NewEngine()
	ledger := engine.GetLedger()

	if ledger == nil {
		t.Fatal("Ledger not initialized")
	}

	account := engine.CreateAccount("user1", "trader1", "pass1", false)
	account.Balance = 10000

	engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		return 1.1000, 1.1002, true
	})

	// Execute and close a trade
	position, _ := engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", 0.1, 0, 0)
	engine.ClosePosition(position.ID, 0)

	// Verify ledger has entries
	entries := ledger.GetHistory(account.ID, 100)
	if len(entries) == 0 {
		t.Error("Ledger should have entries after trade")
	}
}

// TestEdgeCases tests edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	t.Run("Execute order on inactive account", func(t *testing.T) {
		engine := NewEngine()
		account := engine.CreateAccount("user1", "trader1", "pass1", false)
		account.Status = "DISABLED"
		account.Balance = 10000

		engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
			return 1.1000, 1.1002, true
		})

		_, err := engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", 0.1, 0, 0)
		if err == nil {
			t.Error("Expected error for inactive account, got nil")
		}
	})

	t.Run("Execute order without price callback", func(t *testing.T) {
		engine := NewEngine()
		account := engine.CreateAccount("user1", "trader1", "pass1", false)
		account.Balance = 10000

		_, err := engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", 0.1, 0, 0)
		if err == nil {
			t.Error("Expected price feed error, got nil")
		}
	})

	t.Run("Close position without price callback", func(t *testing.T) {
		engine := NewEngine()
		account := engine.CreateAccount("user1", "trader1", "pass1", false)
		account.Balance = 10000

		engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
			return 1.1000, 1.1002, true
		})

		position, _ := engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", 0.1, 0, 0)

		// Remove callback
		engine.SetPriceCallback(nil)

		_, err := engine.ClosePosition(position.ID, 0)
		if err == nil {
			t.Error("Expected price feed error, got nil")
		}
	})
}
