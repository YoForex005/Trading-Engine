# Advanced Trading Features

This package provides MT5-competitive advanced trading features for the RTX Trading Engine.

## Features

### 1. Advanced Order Types (`order_types.go`)

#### Supported Order Types:

- **Bracket Orders**: Entry + automatic SL + TP in one order
- **Iceberg Orders**: Hidden volume execution with visible portions
- **TWAP (Time-Weighted Average Price)**: Execute volume evenly over time
- **VWAP (Volume-Weighted Average Price)**: Target VWAP price execution
- **OCO (One-Cancels-Other)**: Already implemented in base system
- **Trailing Stop**: Already implemented in base system

#### Time-in-Force Options:

- **GTC (Good-Till-Cancelled)**: Order stays until filled or cancelled
- **GTD (Good-Till-Date)**: Order expires at specified date
- **FOK (Fill-or-Kill)**: Execute entire order immediately or cancel
- **IOC (Immediate-or-Cancel)**: Execute available portion immediately
- **DAY**: Order expires at end of trading day

#### Example Usage:

```go
// Create bracket order (entry + SL + TP)
bracket, err := orderService.PlaceBracketOrder(
    "EURUSD",      // symbol
    "BUY",         // side
    1.0,           // volume
    1.10000,       // entry price
    1.09500,       // stop loss
    1.10500,       // take profit
    "LIMIT",       // entry type (MARKET, LIMIT, STOP)
    TIF_GTC,       // time in force
    nil,           // expiry time (optional)
)

// Create TWAP order for 10 lots over 60 minutes
twap, err := orderService.PlaceTWAPOrder(
    "EURUSD",                          // symbol
    "BUY",                             // side
    10.0,                              // total volume
    time.Now(),                        // start time
    time.Now().Add(60*time.Minute),    // end time
    60,                                // interval seconds
    1.09900,                           // min price (optional)
    1.10100,                           // max price (optional)
)
```

### 2. Technical Indicators (`indicators.go`)

#### Supported Indicators:

**Moving Averages:**
- SMA (Simple Moving Average)
- EMA (Exponential Moving Average)
- WMA (Weighted Moving Average)

**Oscillators:**
- RSI (Relative Strength Index)
- MACD (Moving Average Convergence Divergence)
- Stochastic Oscillator

**Volatility:**
- Bollinger Bands
- ATR (Average True Range)
- ADX (Average Directional Index)

**Support/Resistance:**
- Pivot Points (Standard, Fibonacci, Camarilla)
- Fibonacci Retracements

#### Example Usage:

```go
indicatorService := NewIndicatorService(200) // Keep 200 bars in cache

// Calculate RSI(14)
rsi, err := indicatorService.RSI("EURUSD", 14)

// Calculate MACD(12,26,9)
macd, err := indicatorService.MACD("EURUSD", 12, 26, 9)

// Calculate Bollinger Bands(20, 2.0)
bb, err := indicatorService.BollingerBands("EURUSD", 20, 2.0)

// Get Pivot Points
pivots, err := indicatorService.PivotPoints("EURUSD")
// Returns: pivot, r1, r2, r3, s1, s2, s3

// Calculate Fibonacci levels
fib := indicatorService.FibonacciRetracement(1.10500, 1.09500)
// Returns: 0%, 23.6%, 38.2%, 50%, 61.8%, 78.6%, 100%
```

### 3. Strategy Automation (`strategies.go`)

#### Features:

- **Strategy Types**: Indicator-based, Pattern-based, ML-based, Arbitrage, Custom
- **Execution Modes**: Live, Paper Trading, Backtesting
- **Backtesting Framework**: Full historical simulation
- **Performance Analytics**: Win rate, Sharpe ratio, profit factor, etc.
- **Risk Management**: Per-trade risk limits, drawdown limits, position limits

#### Example Usage:

```go
strategyService := NewStrategyService(indicatorService)

// Create a strategy
strategy := &Strategy{
    Name:        "MA Crossover",
    Type:        StrategyTypeIndicatorBased,
    Mode:        ModePaper,  // Paper trading mode
    Symbols:     []string{"EURUSD", "GBPUSD"},
    Timeframe:   "M15",
    Enabled:     true,
    Config: map[string]interface{}{
        "fastPeriod": 10,
        "slowPeriod": 20,
    },
    MaxPositions:    3,
    MaxRiskPerTrade: 2.0,  // 2% per trade
    MaxDrawdown:     10.0, // 10% max drawdown
}

created, err := strategyService.CreateStrategy(strategy)

// Run backtest
result, err := strategyService.RunBacktest(
    created.ID,
    time.Now().AddDate(0, -6, 0), // 6 months ago
    time.Now(),
    10000.0, // initial balance
)

// View results
fmt.Printf("Win Rate: %.2f%%\n", result.Stats.WinRate)
fmt.Printf("Net Profit: $%.2f\n", result.Stats.NetProfit)
fmt.Printf("Sharpe Ratio: %.2f\n", result.Stats.SharpeRatio)
```

### 4. Alerts System (`alerts.go`)

#### Alert Types:

- **Price Alerts**: Above, Below, Cross Above, Cross Below
- **Indicator Alerts**: RSI overbought/oversold, MACD crossover
- **Price Change Alerts**: Percentage change threshold
- **News Alerts**: Keyword-based news monitoring
- **Account Alerts**: Balance changes, margin calls

#### Notification Channels:

- Email
- SMS
- Push Notifications
- Webhooks (custom integrations)
- In-App notifications

#### Example Usage:

```go
alertService := NewAlertService()

// Create price alert
alert := &Alert{
    UserID:      "user123",
    Name:        "EURUSD Above 1.10",
    Type:        AlertTypePrice,
    Symbol:      "EURUSD",
    Condition:   ConditionAbove,
    Value:       1.10000,
    Message:     "EURUSD has crossed above 1.10000",
    Channels:    []NotificationChannel{ChannelEmail, ChannelPush},
    TriggerOnce: true,
}

created, err := alertService.CreateAlert(alert)

// Create indicator alert
rsiAlert := &Alert{
    UserID:    "user123",
    Name:      "EURUSD RSI Overbought",
    Type:      AlertTypeIndicator,
    Symbol:    "EURUSD",
    Indicator: "RSI",
    Period:    14,
    Condition: ConditionOverbought,
    Value:     70.0,
    Message:   "EURUSD RSI is overbought",
    Channels:  []NotificationChannel{ChannelWebhook},
    Webhook:   "https://your-server.com/webhook",
}

// Create news alert
newsAlert := &NewsAlert{
    UserID:      "user123",
    Keywords:    []string{"Fed", "interest rate", "NFP"},
    Symbols:     []string{"EURUSD", "GBPUSD"},
    MinSeverity: "HIGH",
}

alertService.CreateNewsAlert(newsAlert)
```

### 5. Advanced Reporting (`reports.go`)

#### Report Types:

**Tax Reports:**
- Yearly P&L summary
- Short-term vs long-term gains
- Commission and swap totals
- Symbol-by-symbol breakdown

**Performance Reports:**
- Win rate, profit factor
- Sharpe ratio, Sortino ratio
- MAE/MFE analysis
- R-multiple analysis
- Time-based attribution
- Symbol-based attribution

**Drawdown Analysis:**
- Maximum drawdown
- Drawdown periods
- Recovery time analysis
- Current drawdown status

#### Example Usage:

```go
reportService := NewReportService()

// Generate tax report for 2024
taxReport, err := reportService.GenerateTaxReport("account123", 2024)

fmt.Printf("Taxable Profit: $%.2f\n", taxReport.TaxableProfitUSD)
fmt.Printf("Short-term Gains: $%.2f\n", taxReport.ShortTermGains)
fmt.Printf("Long-term Gains: $%.2f\n", taxReport.LongTermGains)

// Generate performance report
perfReport, err := reportService.GeneratePerformanceReport(
    "account123",
    time.Now().AddDate(0, -3, 0), // Last 3 months
    time.Now(),
)

fmt.Printf("Win Rate: %.2f%%\n", perfReport.WinRate)
fmt.Printf("Sharpe Ratio: %.2f\n", perfReport.SharpeRatio)
fmt.Printf("Profit Factor: %.2f\n", perfReport.ProfitFactor)
fmt.Printf("Average MAE: $%.2f\n", perfReport.AverageMAE)
fmt.Printf("Average MFE: $%.2f\n", perfReport.AverageMFE)

// Generate drawdown analysis
ddAnalysis, err := reportService.GenerateDrawdownAnalysis(
    "account123",
    time.Now().AddDate(0, -6, 0),
    time.Now(),
    10000.0, // initial balance
)

fmt.Printf("Max Drawdown: $%.2f (%.2f%%)\n",
    ddAnalysis.MaxDrawdown,
    ddAnalysis.MaxDrawdownPct)
fmt.Printf("Recovery Factor: %.2f\n", ddAnalysis.RecoveryFactor)
```

## REST API Endpoints

All endpoints are CORS-enabled and return JSON responses.

### Advanced Orders

```
POST   /api/orders/bracket         - Place bracket order
POST   /api/orders/twap            - Place TWAP order
POST   /api/orders/vwap            - Place VWAP order
POST   /api/orders/iceberg         - Place iceberg order
GET    /api/orders/bracket/list    - List bracket orders
DELETE /api/orders/bracket/:id     - Cancel bracket order
```

### Indicators

```
GET /api/indicators/calculate?symbol=EURUSD&indicator=rsi&period=14
```

Supported indicators: `sma`, `ema`, `wma`, `rsi`, `macd`, `bb`, `atr`, `adx`, `stochastic`, `pivot`

### Strategies

```
GET    /api/strategies             - List all strategies
POST   /api/strategies/create      - Create strategy
PUT    /api/strategies/:id         - Update strategy
DELETE /api/strategies/:id         - Delete strategy
POST   /api/strategies/backtest    - Run backtest
GET    /api/strategies/:id/signals - Get strategy signals
```

### Alerts

```
GET    /api/alerts/list?userId=xxx         - List user alerts
POST   /api/alerts/create                  - Create alert
PUT    /api/alerts/:id                     - Update alert
DELETE /api/alerts/:id                     - Delete alert
GET    /api/alerts/triggers?userId=xxx     - Get triggered alerts
```

### Reports

```
GET /api/reports/tax?accountId=xxx&year=2024
GET /api/reports/performance?accountId=xxx&startDate=2024-01-01&endDate=2024-12-31
GET /api/reports/drawdown?accountId=xxx&initialBalance=10000
```

## Integration Example

See `integration_example.go` for a complete example of how to wire all features together with the main trading engine.

## Performance Considerations

- **Indicator Caching**: Indicators cache up to 200 bars per symbol by default
- **TWAP/VWAP**: Process in background with minimal latency
- **Backtesting**: Can process years of data efficiently
- **Alerts**: Check every 1 second for price/indicator alerts
- **Reports**: Generate on-demand, can be cached

## Thread Safety

All services are thread-safe and use `sync.RWMutex` for concurrent access.

## Future Enhancements

- [ ] Custom indicator plugins
- [ ] Machine learning strategy framework
- [ ] Real-time news feed integration
- [ ] Advanced chart pattern recognition
- [ ] Multi-leg options strategies
- [ ] Portfolio optimization algorithms
