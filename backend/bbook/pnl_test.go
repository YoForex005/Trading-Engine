package bbook

import (
	"testing"
)

func TestCalculatePositionProfit(t *testing.T) {
	tests := []struct {
		name           string
		symbol         string
		side           string
		volume         float64
		openPrice      float64
		currentPrice   float64
		expectedProfit float64
	}{
		{
			name:           "buy position in profit",
			symbol:         "EURUSD",
			side:           "BUY",
			volume:         1.0,
			openPrice:      1.0850,
			currentPrice:   1.0900,
			expectedProfit: 500.00, // 50 pips * $10/pip for 1 lot
		},
		{
			name:           "buy position in loss",
			symbol:         "EURUSD",
			side:           "BUY",
			volume:         1.0,
			openPrice:      1.0850,
			currentPrice:   1.0800,
			expectedProfit: -500.00, // -50 pips
		},
		{
			name:           "sell position in profit",
			symbol:         "EURUSD",
			side:           "SELL",
			volume:         1.0,
			openPrice:      1.0850,
			currentPrice:   1.0800,
			expectedProfit: 500.00, // Price went down, profit for sell
		},
		{
			name:           "sell position in loss",
			symbol:         "EURUSD",
			side:           "SELL",
			volume:         1.0,
			openPrice:      1.0850,
			currentPrice:   1.0900,
			expectedProfit: -500.00,
		},
		{
			name:           "mini lot position",
			symbol:         "EURUSD",
			side:           "BUY",
			volume:         0.1,
			openPrice:      1.0850,
			currentPrice:   1.0900,
			expectedProfit: 50.00, // 1/10 of standard lot
		},
		{
			name:           "micro lot position",
			symbol:         "EURUSD",
			side:           "BUY",
			volume:         0.01,
			openPrice:      1.0850,
			currentPrice:   1.0900,
			expectedProfit: 5.00,
		},
		{
			name:           "no movement = no profit",
			symbol:         "EURUSD",
			side:           "BUY",
			volume:         1.0,
			openPrice:      1.0850,
			currentPrice:   1.0850,
			expectedProfit: 0.00,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			engine := NewEngine()
			spec := engine.symbols[tt.symbol]

			position := &Position{
				Symbol:    tt.symbol,
				Side:      tt.side,
				Volume:    tt.volume,
				OpenPrice: tt.openPrice,
			}

			profit := engine.calculatePnL(position, tt.currentPrice, tt.volume, spec)

			// Allow small floating point variance
			tolerance := 0.01
			if abs(profit-tt.expectedProfit) > tolerance {
				t.Errorf("position profit: got %.2f, want %.2f", profit, tt.expectedProfit)
			}
		})
	}
}

func TestCalculateMarginRequirement(t *testing.T) {
	tests := []struct {
		name           string
		symbol         string
		volume         float64
		price          float64
		leverage       float64
		expectedMargin float64
	}{
		{
			name:           "standard lot with 100:1 leverage",
			symbol:         "EURUSD",
			volume:         1.0,
			price:          1.0850,
			leverage:       100,
			expectedMargin: 1085.00,
		},
		{
			name:           "mini lot with 100:1 leverage",
			symbol:         "EURUSD",
			volume:         0.1,
			price:          1.0850,
			leverage:       100,
			expectedMargin: 108.50,
		},
		{
			name:           "high leverage reduces margin",
			symbol:         "EURUSD",
			volume:         1.0,
			price:          1.0850,
			leverage:       500,
			expectedMargin: 217.00,
		},
		{
			name:           "low leverage increases margin",
			symbol:         "EURUSD",
			volume:         1.0,
			price:          1.0850,
			leverage:       50,
			expectedMargin: 2170.00,
		},
		{
			name:           "large position size",
			symbol:         "EURUSD",
			volume:         10.0,
			price:          1.0850,
			leverage:       100,
			expectedMargin: 10850.00,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			engine := NewEngine()

			margin := engine.calculateMargin(tt.symbol, tt.volume, tt.price, tt.leverage)

			// Allow small floating point variance
			tolerance := 0.01
			if abs(margin-tt.expectedMargin) > tolerance {
				t.Errorf("margin requirement: got %.2f, want %.2f", margin, tt.expectedMargin)
			}
		})
	}
}

func TestPositionMarginCalculation(t *testing.T) {
	tests := []struct {
		name           string
		symbol         string
		volume         float64
		openPrice      float64
		leverage       float64
		expectedMargin float64
	}{
		{
			name:           "EUR position margin",
			symbol:         "EURUSD",
			volume:         1.0,
			openPrice:      1.1000,
			leverage:       100,
			expectedMargin: 1100.00,
		},
		{
			name:           "GBP position margin",
			symbol:         "GBPUSD",
			volume:         0.5,
			openPrice:      1.2500,
			leverage:       100,
			expectedMargin: 625.00,
		},
		{
			name:           "high leverage position",
			symbol:         "EURUSD",
			volume:         1.0,
			openPrice:      1.1000,
			leverage:       500,
			expectedMargin: 220.00,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			engine := NewEngine()
			spec := engine.symbols[tt.symbol]

			position := &Position{
				Symbol:    tt.symbol,
				Volume:    tt.volume,
				OpenPrice: tt.openPrice,
			}

			margin := engine.calculatePositionMargin(position, spec, tt.leverage)

			// Allow small floating point variance
			tolerance := 0.01
			if abs(margin-tt.expectedMargin) > tolerance {
				t.Errorf("position margin: got %.2f, want %.2f", margin, tt.expectedMargin)
			}
		})
	}
}

func TestPnLWithDifferentSymbols(t *testing.T) {
	tests := []struct {
		name           string
		symbol         string
		side           string
		volume         float64
		openPrice      float64
		currentPrice   float64
		expectedProfit float64
	}{
		{
			name:           "GBPUSD buy profit",
			symbol:         "GBPUSD",
			side:           "BUY",
			volume:         1.0,
			openPrice:      1.2500,
			currentPrice:   1.2550,
			expectedProfit: 500.00, // 50 pips * $10/pip
		},
		{
			name:           "USDJPY buy profit",
			symbol:         "USDJPY",
			side:           "BUY",
			volume:         1.0,
			openPrice:      110.00,
			currentPrice:   110.50,
			expectedProfit: 454.50, // 50 pips * $9.09/pip
		},
		{
			name:           "AUDUSD sell profit",
			symbol:         "AUDUSD",
			side:           "SELL",
			volume:         1.0,
			openPrice:      0.7200,
			currentPrice:   0.7150,
			expectedProfit: 500.00, // 50 pips * $10/pip
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			engine := NewEngine()
			spec := engine.symbols[tt.symbol]

			position := &Position{
				Symbol:    tt.symbol,
				Side:      tt.side,
				Volume:    tt.volume,
				OpenPrice: tt.openPrice,
			}

			profit := engine.calculatePnL(position, tt.currentPrice, tt.volume, spec)

			// Allow larger tolerance for JPY pairs due to pip value differences
			tolerance := 1.0
			if abs(profit-tt.expectedProfit) > tolerance {
				t.Errorf("position profit: got %.2f, want %.2f", profit, tt.expectedProfit)
			}
		})
	}
}

// abs returns the absolute value of x
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
