# Quick Start Guide - Advanced Trading Features

Get up and running with advanced trading features in 5 minutes.

## Step 1: Add to Your main.go

Add this import at the top:

```go
import "github.com/epic1st/rtx/backend/features"
```

Add this in your `main()` function after initializing the B-Book engine and hub:

```go
// Initialize advanced features
featureHandlers := features.InitializeAdvancedFeatures(bbookEngine, hub, lpMgr)

// Optional: Create demo strategy and alerts
features.CreateDemoStrategy(featureHandlers.strategyService)
features.CreateDemoAlerts(featureHandlers.alertService)
```

That's it! All features are now available.

## Step 2: Test the API

### Place a Bracket Order

```bash
curl -X POST http://localhost:7999/api/orders/bracket \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "EURUSD",
    "side": "BUY",
    "volume": 1.0,
    "entryPrice": 1.10000,
    "stopLoss": 1.09500,
    "takeProfit": 1.10500,
    "entryType": "LIMIT",
    "timeInForce": "GTC"
  }'
```

### Calculate RSI

```bash
curl "http://localhost:7999/api/indicators/calculate?symbol=EURUSD&indicator=rsi&period=14"
```

### Create a Price Alert

```bash
curl -X POST http://localhost:7999/api/alerts/create \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "demo-user",
    "name": "EURUSD Above 1.10",
    "type": "PRICE",
    "symbol": "EURUSD",
    "condition": "ABOVE",
    "value": 1.10000,
    "message": "EURUSD crossed above 1.10",
    "channels": ["IN_APP"]
  }'
```

### Generate Performance Report

```bash
curl "http://localhost:7999/api/reports/performance?accountId=demo-user&startDate=2024-01-01&endDate=2024-12-31"
```

## Step 3: View Available Features

Visit these endpoints to see what's available:

- **List Strategies**: `GET http://localhost:7999/api/strategies`
- **List Alerts**: `GET http://localhost:7999/api/alerts/list?userId=demo-user`
- **List Bracket Orders**: `GET http://localhost:7999/api/orders/bracket/list`

## Common Use Cases

### 1. Auto-Trading Strategy

```bash
# Create strategy
curl -X POST http://localhost:7999/api/strategies/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MA Crossover",
    "type": "INDICATOR",
    "mode": "PAPER",
    "symbols": ["EURUSD"],
    "timeframe": "M15",
    "config": {"fastPeriod": 10, "slowPeriod": 20},
    "maxPositions": 3,
    "maxRiskPerTrade": 2.0
  }'

# Get the strategy ID from response, then enable it:
curl -X PUT http://localhost:7999/api/strategies/{strategyId} \
  -d '{"enabled": true}'
```

### 2. Large Order Execution with TWAP

```bash
curl -X POST http://localhost:7999/api/orders/twap \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "EURUSD",
    "side": "BUY",
    "totalVolume": 10.0,
    "durationMinutes": 60,
    "intervalSeconds": 300,
    "minPrice": 1.09900,
    "maxPrice": 1.10100
  }'
```

### 3. Risk Management with Alerts

```bash
# Alert when RSI is overbought
curl -X POST http://localhost:7999/api/alerts/create \
  -d '{
    "userId": "demo-user",
    "name": "EURUSD RSI Overbought",
    "type": "INDICATOR",
    "symbol": "EURUSD",
    "indicator": "RSI",
    "period": 14,
    "condition": "OVERBOUGHT",
    "value": 70.0,
    "message": "EURUSD RSI > 70",
    "channels": ["EMAIL", "PUSH"]
  }'
```

### 4. Strategy Backtesting

```bash
curl -X POST http://localhost:7999/api/strategies/backtest \
  -H "Content-Type: application/json" \
  -d '{
    "strategyId": "{your-strategy-id}",
    "startDate": "2023-01-01T00:00:00Z",
    "endDate": "2024-01-01T00:00:00Z",
    "initialBalance": 10000.0
  }'
```

## Frontend Integration

### JavaScript Example

```javascript
// Place bracket order
async function placeBracketOrder() {
  const response = await fetch('http://localhost:7999/api/orders/bracket', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      symbol: 'EURUSD',
      side: 'BUY',
      volume: 1.0,
      entryPrice: 1.10000,
      stopLoss: 1.09500,
      takeProfit: 1.10500,
      entryType: 'LIMIT',
      timeInForce: 'GTC'
    })
  });
  const order = await response.json();
  console.log('Bracket order placed:', order);
}

// Calculate indicator
async function calculateRSI(symbol) {
  const response = await fetch(
    `http://localhost:7999/api/indicators/calculate?symbol=${symbol}&indicator=rsi&period=14`
  );
  const rsi = await response.json();
  console.log('RSI values:', rsi);
}

// Create alert
async function createAlert() {
  const response = await fetch('http://localhost:7999/api/alerts/create', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      userId: 'user123',
      name: 'Price Alert',
      type: 'PRICE',
      symbol: 'EURUSD',
      condition: 'ABOVE',
      value: 1.10000,
      message: 'EURUSD above 1.10',
      channels: ['IN_APP', 'EMAIL']
    })
  });
  const alert = await response.json();
  console.log('Alert created:', alert);
}
```

### React Example

```jsx
import { useState, useEffect } from 'react';

function IndicatorChart({ symbol }) {
  const [rsi, setRsi] = useState([]);

  useEffect(() => {
    fetch(`http://localhost:7999/api/indicators/calculate?symbol=${symbol}&indicator=rsi&period=14`)
      .then(res => res.json())
      .then(data => setRsi(data));
  }, [symbol]);

  return (
    <div>
      <h3>RSI(14) - {symbol}</h3>
      {rsi.length > 0 && (
        <div>Current: {rsi[rsi.length - 1].values.rsi.toFixed(2)}</div>
      )}
    </div>
  );
}
```

## Admin Panel Integration

Add these sections to your admin panel:

1. **Strategy Management**
   - List all strategies
   - Enable/disable strategies
   - View performance stats
   - Run backtests

2. **Alert Management**
   - View active alerts
   - View triggered alerts
   - Create/edit alerts

3. **Advanced Orders**
   - Monitor TWAP/VWAP orders
   - View bracket orders
   - Track execution progress

4. **Reports**
   - Generate tax reports
   - View performance analytics
   - Analyze drawdowns

## Troubleshooting

### Features not working?

1. Check logs for initialization messages:
```
[✓] Indicator Service initialized
[✓] Advanced Order Service initialized
[✓] Strategy Service initialized
[✓] Alert Service initialized
[✓] Report Service initialized
```

2. Verify API routes are registered:
```bash
curl http://localhost:7999/health
```

3. Check CORS headers if calling from browser

### Need historical data for indicators/backtesting?

Implement the data callback in `InitializeAdvancedFeatures`:

```go
indicatorService.SetDataCallback(func(symbol, timeframe string, count int) ([]OHLCBar, error) {
  // Fetch from your OHLC storage
  bars := tickStore.GetOHLCBars(symbol, timeframe, count)
  return convertToOHLCBars(bars), nil
})
```

## Performance Tips

1. **Indicators**: Cache is limited to 200 bars per symbol. Increase if needed:
```go
indicatorService := NewIndicatorService(500) // 500 bars
```

2. **Alerts**: Check frequency is 1 second. Adjust if needed in `alerts.go`

3. **TWAP/VWAP**: Process every 100ms. Suitable for most use cases.

## Next Steps

- Read [README.md](README.md) for detailed feature documentation
- Check [API_DOCUMENTATION.md](API_DOCUMENTATION.md) for complete API reference
- Review [integration_example.go](integration_example.go) for advanced integration
- See [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md) for architecture details

## Support

Questions? Check the documentation or reach out to the development team.

Happy Trading! 🚀
