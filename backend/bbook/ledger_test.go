package bbook

import (
	"testing"
)

func TestLedgerDeposit(t *testing.T) {
	tests := []struct {
		name            string
		initialBalance  float64
		depositAmount   float64
		expectedBalance float64
		expectError     bool
	}{
		{
			name:            "standard deposit",
			initialBalance:  1000.00,
			depositAmount:   500.00,
			expectedBalance: 1500.00,
			expectError:     false,
		},
		{
			name:            "zero deposit rejected",
			initialBalance:  1000.00,
			depositAmount:   0.00,
			expectedBalance: 1000.00,
			expectError:     true,
		},
		{
			name:            "negative deposit rejected",
			initialBalance:  1000.00,
			depositAmount:   -100.00,
			expectedBalance: 1000.00,
			expectError:     true,
		},
		{
			name:            "large deposit",
			initialBalance:  1000.00,
			depositAmount:   1000000.00,
			expectedBalance: 1001000.00,
			expectError:     false,
		},
		{
			name:            "fractional cents",
			initialBalance:  1000.00,
			depositAmount:   100.125,
			expectedBalance: 1100.125,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			ledger := NewLedger()
			ledger.SetBalance(1, tt.initialBalance)

			// Execute
			_, err := ledger.Deposit(1, tt.depositAmount, "BANK", "REF123", "Test deposit", "admin1")

			// Verify
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				balance := ledger.GetBalance(1)
				if balance != tt.expectedBalance {
					t.Errorf("balance after deposit: got %.2f, want %.2f", balance, tt.expectedBalance)
				}
			}
		})
	}
}

func TestLedgerWithdraw(t *testing.T) {
	tests := []struct {
		name            string
		initialBalance  float64
		withdrawAmount  float64
		expectedBalance float64
		expectError     bool
	}{
		{
			name:            "standard withdrawal",
			initialBalance:  1000.00,
			withdrawAmount:  500.00,
			expectedBalance: 500.00,
			expectError:     false,
		},
		{
			name:            "withdrawal exceeds balance",
			initialBalance:  1000.00,
			withdrawAmount:  1500.00,
			expectedBalance: 1000.00,
			expectError:     true,
		},
		{
			name:            "zero withdrawal rejected",
			initialBalance:  1000.00,
			withdrawAmount:  0.00,
			expectedBalance: 1000.00,
			expectError:     true,
		},
		{
			name:            "negative withdrawal rejected",
			initialBalance:  1000.00,
			withdrawAmount:  -100.00,
			expectedBalance: 1000.00,
			expectError:     true,
		},
		{
			name:            "withdraw entire balance",
			initialBalance:  1000.00,
			withdrawAmount:  1000.00,
			expectedBalance: 0.00,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ledger := NewLedger()
			ledger.SetBalance(1, tt.initialBalance)

			_, err := ledger.Withdraw(1, tt.withdrawAmount, "BANK", "REF123", "Test withdrawal", "admin1")

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				balance := ledger.GetBalance(1)
				if balance != tt.expectedBalance {
					t.Errorf("balance after withdrawal: got %.2f, want %.2f", balance, tt.expectedBalance)
				}
			}
		})
	}
}

func TestLedgerAdjust(t *testing.T) {
	tests := []struct {
		name            string
		initialBalance  float64
		adjustAmount    float64
		expectedBalance float64
		expectError     bool
	}{
		{
			name:            "positive adjustment",
			initialBalance:  1000.00,
			adjustAmount:    250.00,
			expectedBalance: 1250.00,
			expectError:     false,
		},
		{
			name:            "negative adjustment",
			initialBalance:  1000.00,
			adjustAmount:    -300.00,
			expectedBalance: 700.00,
			expectError:     false,
		},
		{
			name:            "adjustment causing negative balance rejected",
			initialBalance:  1000.00,
			adjustAmount:    -1500.00,
			expectedBalance: 1000.00,
			expectError:     true,
		},
		{
			name:            "adjustment to exactly zero",
			initialBalance:  1000.00,
			adjustAmount:    -1000.00,
			expectedBalance: 0.00,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ledger := NewLedger()
			ledger.SetBalance(1, tt.initialBalance)

			_, err := ledger.Adjust(1, tt.adjustAmount, "Manual adjustment", "admin1")

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				balance := ledger.GetBalance(1)
				if balance != tt.expectedBalance {
					t.Errorf("balance after adjustment: got %.2f, want %.2f", balance, tt.expectedBalance)
				}
			}
		})
	}
}

func TestLedgerBonus(t *testing.T) {
	tests := []struct {
		name            string
		initialBalance  float64
		bonusAmount     float64
		expectedBalance float64
		expectError     bool
	}{
		{
			name:            "standard bonus",
			initialBalance:  1000.00,
			bonusAmount:     100.00,
			expectedBalance: 1100.00,
			expectError:     false,
		},
		{
			name:            "zero bonus rejected",
			initialBalance:  1000.00,
			bonusAmount:     0.00,
			expectedBalance: 1000.00,
			expectError:     true,
		},
		{
			name:            "negative bonus rejected",
			initialBalance:  1000.00,
			bonusAmount:     -50.00,
			expectedBalance: 1000.00,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ledger := NewLedger()
			ledger.SetBalance(1, tt.initialBalance)

			_, err := ledger.AddBonus(1, tt.bonusAmount, "Welcome bonus", "admin1")

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				balance := ledger.GetBalance(1)
				if balance != tt.expectedBalance {
					t.Errorf("balance after bonus: got %.2f, want %.2f", balance, tt.expectedBalance)
				}
			}
		})
	}
}

func TestLedgerRealizedPnL(t *testing.T) {
	tests := []struct {
		name            string
		initialBalance  float64
		pnlAmount       float64
		expectedBalance float64
	}{
		{
			name:            "profitable trade",
			initialBalance:  1000.00,
			pnlAmount:       150.00,
			expectedBalance: 1150.00,
		},
		{
			name:            "losing trade",
			initialBalance:  1000.00,
			pnlAmount:       -75.00,
			expectedBalance: 925.00,
		},
		{
			name:            "breakeven trade",
			initialBalance:  1000.00,
			pnlAmount:       0.00,
			expectedBalance: 1000.00,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ledger := NewLedger()
			ledger.SetBalance(1, tt.initialBalance)

			entry := ledger.RecordRealizedPnL(1, tt.pnlAmount, 12345)

			if entry == nil {
				t.Fatal("expected ledger entry but got nil")
			}

			balance := ledger.GetBalance(1)
			if balance != tt.expectedBalance {
				t.Errorf("balance after P&L: got %.2f, want %.2f", balance, tt.expectedBalance)
			}

			if entry.Type != "REALIZED_PNL" {
				t.Errorf("entry type: got %s, want REALIZED_PNL", entry.Type)
			}
		})
	}
}

func TestLedgerGetHistory(t *testing.T) {
	ledger := NewLedger()
	ledger.SetBalance(1, 1000.00)

	// Create multiple transactions
	ledger.Deposit(1, 500.00, "BANK", "REF1", "Deposit 1", "admin1")
	ledger.Withdraw(1, 200.00, "BANK", "REF2", "Withdrawal 1", "admin1")
	ledger.AddBonus(1, 100.00, "Bonus 1", "admin1")

	// Get history
	history := ledger.GetHistory(1, 10)

	if len(history) != 3 {
		t.Errorf("history length: got %d, want 3", len(history))
	}

	// Verify most recent first
	if history[0].Type != "BONUS" {
		t.Errorf("most recent entry type: got %s, want BONUS", history[0].Type)
	}
}
