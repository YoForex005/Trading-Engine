package handlers

import (
	"math"
	"testing"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
)

// Standalone tests for calculation functions

func TestSharpeRatioCalculation(t *testing.T) {
	trades := []core.Trade{
		{RealizedPnL: 100, ExecutedAt: time.Now()},
		{RealizedPnL: 120, ExecutedAt: time.Now().Add(time.Hour)},
		{RealizedPnL: 110, ExecutedAt: time.Now().Add(2 * time.Hour)},
		{RealizedPnL: 130, ExecutedAt: time.Now().Add(3 * time.Hour)},
	}

	sharpe := calculateSharpeRatio(trades, 0.02)

	// With consistent positive returns, Sharpe should be high
	if sharpe < 50 {
		t.Errorf("Sharpe ratio too low for consistent positive returns: %v", sharpe)
	}

	t.Logf("Sharpe Ratio: %.2f", sharpe)
}

func TestProfitFactorCalculation(t *testing.T) {
	trades := []core.Trade{
		{RealizedPnL: 200},
		{RealizedPnL: -100},
	}

	pf := calculateProfitFactor(trades)

	if pf != 2.0 {
		t.Errorf("Profit Factor = %v, expected 2.0", pf)
	}

	t.Logf("Profit Factor: %.2f", pf)
}

func TestMaxDrawdownCalculation(t *testing.T) {
	trades := []core.Trade{
		{RealizedPnL: 100, ExecutedAt: time.Unix(1000, 0)},
		{RealizedPnL: -150, ExecutedAt: time.Unix(2000, 0)}, // Peak 100, trough -50
		{RealizedPnL: 200, ExecutedAt: time.Unix(3000, 0)},  // Recover to 150
	}

	dd := calculateMaxDrawdown(trades)

	// Drawdown from peak of 100 to trough of -50 is 150
	// As percentage of peak: 150/100 = 1.5 (150%)
	if math.Abs(dd-1.5) > 0.01 {
		t.Errorf("Max Drawdown = %v, expected 1.5", dd)
	}

	t.Logf("Max Drawdown: %.2f%%", dd*100)
}

func TestWinRateCalculation(t *testing.T) {
	trades := []core.Trade{
		{RealizedPnL: 100},
		{RealizedPnL: -50},
		{RealizedPnL: 75},
		{RealizedPnL: -25},
	}

	wr := calculateWinRate(trades)

	if wr != 50.0 {
		t.Errorf("Win Rate = %v%%, expected 50.0%%", wr)
	}

	t.Logf("Win Rate: %.2f%%", wr)
}

func TestConsecutiveWinsLosses(t *testing.T) {
	trades := []core.Trade{
		{RealizedPnL: 100, ExecutedAt: time.Unix(1000, 0)},
		{RealizedPnL: 50, ExecutedAt: time.Unix(2000, 0)},
		{RealizedPnL: 75, ExecutedAt: time.Unix(3000, 0)},
		{RealizedPnL: -25, ExecutedAt: time.Unix(4000, 0)},
		{RealizedPnL: -30, ExecutedAt: time.Unix(5000, 0)},
	}

	wins, losses := calculateConsecutiveWinsLosses(trades)

	if wins != 3 {
		t.Errorf("Consecutive Wins = %v, expected 3", wins)
	}

	if losses != 2 {
		t.Errorf("Consecutive Losses = %v, expected 2", losses)
	}

	t.Logf("Max Consecutive Wins: %d, Losses: %d", wins, losses)
}

func TestTimelineGeneration(t *testing.T) {
	now := time.Now()
	trades := []core.Trade{
		{RealizedPnL: 100, ExecutedAt: now},
		{RealizedPnL: -50, ExecutedAt: now.Add(time.Hour)},
		{RealizedPnL: 75, ExecutedAt: now.Add(2 * time.Hour)},
	}

	timeline := calculateTimeline(trades)

	if len(timeline) != 3 {
		t.Errorf("Timeline length = %d, expected 3", len(timeline))
	}

	// Check equity curve
	expectedEquity := []float64{100, 50, 125}
	for i, point := range timeline {
		if point.Equity != expectedEquity[i] {
			t.Errorf("Timeline[%d].Equity = %v, expected %v", i, point.Equity, expectedEquity[i])
		}
	}

	t.Logf("Timeline: %d points generated", len(timeline))
}
