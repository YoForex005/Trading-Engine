package handlers

import (
	"math"
	"testing"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
)

// TestCalculateSharpeRatio tests the Sharpe Ratio calculation
func TestCalculateSharpeRatio(t *testing.T) {
	tests := []struct {
		name         string
		trades       []core.Trade
		riskFreeRate float64
		expected     float64
		tolerance    float64
	}{
		{
			name: "Positive consistent returns",
			trades: []core.Trade{
				{RealizedPnL: 100, ExecutedAt: time.Now()},
				{RealizedPnL: 120, ExecutedAt: time.Now().Add(time.Hour)},
				{RealizedPnL: 110, ExecutedAt: time.Now().Add(2 * time.Hour)},
				{RealizedPnL: 130, ExecutedAt: time.Now().Add(3 * time.Hour)},
			},
			riskFreeRate: 0.02,
			expected:     100.0, // High Sharpe for consistent positive returns
			tolerance:    50.0,
		},
		{
			name: "Mixed returns",
			trades: []core.Trade{
				{RealizedPnL: 100, ExecutedAt: time.Now()},
				{RealizedPnL: -50, ExecutedAt: time.Now().Add(time.Hour)},
				{RealizedPnL: 80, ExecutedAt: time.Now().Add(2 * time.Hour)},
				{RealizedPnL: -30, ExecutedAt: time.Now().Add(3 * time.Hour)},
			},
			riskFreeRate: 0.02,
			expected:     5.0, // Moderate Sharpe for mixed returns
			tolerance:    10.0,
		},
		{
			name:         "Empty trades",
			trades:       []core.Trade{},
			riskFreeRate: 0.02,
			expected:     0.0,
			tolerance:    0.01,
		},
		{
			name: "Single trade",
			trades: []core.Trade{
				{RealizedPnL: 100, ExecutedAt: time.Now()},
			},
			riskFreeRate: 0.02,
			expected:     0.0, // Not enough data for meaningful Sharpe
			tolerance:    0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateSharpeRatio(tt.trades, tt.riskFreeRate)
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("calculateSharpeRatio() = %v, want %v (±%v)", result, tt.expected, tt.tolerance)
			}
		})
	}
}

// TestCalculateProfitFactor tests the Profit Factor calculation
func TestCalculateProfitFactor(t *testing.T) {
	tests := []struct {
		name     string
		trades   []core.Trade
		expected float64
	}{
		{
			name: "2:1 profit factor",
			trades: []core.Trade{
				{RealizedPnL: 200},
				{RealizedPnL: -100},
			},
			expected: 2.0,
		},
		{
			name: "1:1 profit factor",
			trades: []core.Trade{
				{RealizedPnL: 100},
				{RealizedPnL: -100},
			},
			expected: 1.0,
		},
		{
			name: "All winning trades",
			trades: []core.Trade{
				{RealizedPnL: 100},
				{RealizedPnL: 200},
				{RealizedPnL: 150},
			},
			expected: math.Inf(1), // Infinite profit factor
		},
		{
			name: "All losing trades",
			trades: []core.Trade{
				{RealizedPnL: -100},
				{RealizedPnL: -200},
			},
			expected: 0.0,
		},
		{
			name:     "Empty trades",
			trades:   []core.Trade{},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateProfitFactor(tt.trades)
			if math.IsInf(tt.expected, 1) {
				if !math.IsInf(result, 1) {
					t.Errorf("calculateProfitFactor() = %v, want +Inf", result)
				}
			} else if result != tt.expected {
				t.Errorf("calculateProfitFactor() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCalculateMaxDrawdown tests the Maximum Drawdown calculation
func TestCalculateMaxDrawdown(t *testing.T) {
	tests := []struct {
		name      string
		trades    []core.Trade
		expected  float64
		tolerance float64
	}{
		{
			name: "Simple drawdown",
			trades: []core.Trade{
				{RealizedPnL: 100, ExecutedAt: time.Unix(1000, 0)},
				{RealizedPnL: -150, ExecutedAt: time.Unix(2000, 0)}, // Peak at 100, trough at -50
				{RealizedPnL: 200, ExecutedAt: time.Unix(3000, 0)},
			},
			expected:  1.5, // Drawdown of 150 from peak of 100 = 150%
			tolerance: 0.01,
		},
		{
			name: "No drawdown - always increasing",
			trades: []core.Trade{
				{RealizedPnL: 100, ExecutedAt: time.Unix(1000, 0)},
				{RealizedPnL: 50, ExecutedAt: time.Unix(2000, 0)},
				{RealizedPnL: 75, ExecutedAt: time.Unix(3000, 0)},
			},
			expected:  0.0,
			tolerance: 0.01,
		},
		{
			name:      "Empty trades",
			trades:    []core.Trade{},
			expected:  0.0,
			tolerance: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMaxDrawdown(tt.trades)
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("calculateMaxDrawdown() = %v, want %v (±%v)", result, tt.expected, tt.tolerance)
			}
		})
	}
}

// TestCalculateWinRate tests the Win Rate calculation
func TestCalculateWinRate(t *testing.T) {
	tests := []struct {
		name     string
		trades   []core.Trade
		expected float64
	}{
		{
			name: "50% win rate",
			trades: []core.Trade{
				{RealizedPnL: 100},
				{RealizedPnL: -50},
				{RealizedPnL: 75},
				{RealizedPnL: -25},
			},
			expected: 50.0,
		},
		{
			name: "75% win rate",
			trades: []core.Trade{
				{RealizedPnL: 100},
				{RealizedPnL: 50},
				{RealizedPnL: 75},
				{RealizedPnL: -25},
			},
			expected: 75.0,
		},
		{
			name: "100% win rate",
			trades: []core.Trade{
				{RealizedPnL: 100},
				{RealizedPnL: 50},
			},
			expected: 100.0,
		},
		{
			name: "0% win rate",
			trades: []core.Trade{
				{RealizedPnL: -100},
				{RealizedPnL: -50},
			},
			expected: 0.0,
		},
		{
			name:     "Empty trades",
			trades:   []core.Trade{},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateWinRate(tt.trades)
			if result != tt.expected {
				t.Errorf("calculateWinRate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCalculateConsecutiveWinsLosses tests the consecutive wins/losses calculation
func TestCalculateConsecutiveWinsLosses(t *testing.T) {
	tests := []struct {
		name          string
		trades        []core.Trade
		expectedWins  int
		expectedLoses int
	}{
		{
			name: "Streak of 3 wins, 2 losses",
			trades: []core.Trade{
				{RealizedPnL: 100, ExecutedAt: time.Unix(1000, 0)},
				{RealizedPnL: 50, ExecutedAt: time.Unix(2000, 0)},
				{RealizedPnL: 75, ExecutedAt: time.Unix(3000, 0)},
				{RealizedPnL: -25, ExecutedAt: time.Unix(4000, 0)},
				{RealizedPnL: -30, ExecutedAt: time.Unix(5000, 0)},
				{RealizedPnL: 100, ExecutedAt: time.Unix(6000, 0)},
			},
			expectedWins:  3,
			expectedLoses: 2,
		},
		{
			name: "All wins",
			trades: []core.Trade{
				{RealizedPnL: 100, ExecutedAt: time.Unix(1000, 0)},
				{RealizedPnL: 50, ExecutedAt: time.Unix(2000, 0)},
				{RealizedPnL: 75, ExecutedAt: time.Unix(3000, 0)},
			},
			expectedWins:  3,
			expectedLoses: 0,
		},
		{
			name: "All losses",
			trades: []core.Trade{
				{RealizedPnL: -100, ExecutedAt: time.Unix(1000, 0)},
				{RealizedPnL: -50, ExecutedAt: time.Unix(2000, 0)},
			},
			expectedWins:  0,
			expectedLoses: 2,
		},
		{
			name:          "Empty trades",
			trades:        []core.Trade{},
			expectedWins:  0,
			expectedLoses: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wins, losses := calculateConsecutiveWinsLosses(tt.trades)
			if wins != tt.expectedWins {
				t.Errorf("calculateConsecutiveWinsLosses() wins = %v, want %v", wins, tt.expectedWins)
			}
			if losses != tt.expectedLoses {
				t.Errorf("calculateConsecutiveWinsLosses() losses = %v, want %v", losses, tt.expectedLoses)
			}
		})
	}
}

// TestCalculateTimeline tests the timeline generation
func TestCalculateTimeline(t *testing.T) {
	now := time.Now()
	trades := []core.Trade{
		{RealizedPnL: 100, ExecutedAt: now},
		{RealizedPnL: -50, ExecutedAt: now.Add(time.Hour)},
		{RealizedPnL: 75, ExecutedAt: now.Add(2 * time.Hour)},
	}

	timeline := calculateTimeline(trades)

	if len(timeline) != 3 {
		t.Errorf("calculateTimeline() returned %d points, want 3", len(timeline))
	}

	// Check equity progression
	if timeline[0].Equity != 100 {
		t.Errorf("timeline[0].Equity = %v, want 100", timeline[0].Equity)
	}
	if timeline[1].Equity != 50 {
		t.Errorf("timeline[1].Equity = %v, want 50", timeline[1].Equity)
	}
	if timeline[2].Equity != 125 {
		t.Errorf("timeline[2].Equity = %v, want 125", timeline[2].Equity)
	}

	// Check PnL values
	if timeline[0].PnL != 100 {
		t.Errorf("timeline[0].PnL = %v, want 100", timeline[0].PnL)
	}
	if timeline[1].PnL != -50 {
		t.Errorf("timeline[1].PnL = %v, want -50", timeline[1].PnL)
	}
	if timeline[2].PnL != 75 {
		t.Errorf("timeline[2].PnL = %v, want 75", timeline[2].PnL)
	}
}

// TestCalculateRuleMetrics tests the comprehensive metrics calculation
func TestCalculateRuleMetrics(t *testing.T) {
	trades := []core.Trade{
		{RealizedPnL: 100, ExecutedAt: time.Unix(1000, 0)},
		{RealizedPnL: -50, ExecutedAt: time.Unix(2000, 0)},
		{RealizedPnL: 75, ExecutedAt: time.Unix(3000, 0)},
		{RealizedPnL: 125, ExecutedAt: time.Unix(4000, 0)},
	}

	metrics := calculateRuleMetrics("test-rule-1", trades, 0.02)

	if metrics.RuleID != "test-rule-1" {
		t.Errorf("RuleID = %v, want test-rule-1", metrics.RuleID)
	}

	if metrics.TradeCount != 4 {
		t.Errorf("TradeCount = %v, want 4", metrics.TradeCount)
	}

	if metrics.TotalPnL != 250 {
		t.Errorf("TotalPnL = %v, want 250", metrics.TotalPnL)
	}

	if metrics.WinRate != 75.0 {
		t.Errorf("WinRate = %v, want 75.0", metrics.WinRate)
	}

	// Profit factor should be (100 + 75 + 125) / 50 = 6.0
	if metrics.ProfitFactor != 6.0 {
		t.Errorf("ProfitFactor = %v, want 6.0", metrics.ProfitFactor)
	}
}
