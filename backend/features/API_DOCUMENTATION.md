

# Advanced Trading Features - API Documentation

Complete API reference for all advanced trading features.

## Base URL

```
http://localhost:7999/api
```

All endpoints return JSON and support CORS.

---

## Advanced Order Types

### Place Bracket Order

Creates an entry order with automatic stop-loss and take-profit.

**Endpoint:** `POST /api/orders/bracket`

**Request Body:**
```json
{
  "symbol": "EURUSD",
  "side": "BUY",
  "volume": 1.0,
  "entryPrice": 1.10000,
  "stopLoss": 1.09500,
  "takeProfit": 1.10500,
  "entryType": "LIMIT",
  "timeInForce": "GTC"
}
```

**Response:**
```json
{
  "id": "bracket-123",
  "symbol": "EURUSD",
  "side": "BUY",
  "volume": 1.0,
  "entryPrice": 1.10000,
  "stopLoss": 1.09500,
  "takeProfit": 1.10500,
  "status": "PENDING",
  "createdAt": "2024-01-15T10:30:00Z"
}
```

### Place TWAP Order

Time-Weighted Average Price execution.

**Endpoint:** `POST /api/orders/twap`

**Request Body:**
```json
{
  "symbol": "EURUSD",
  "side": "BUY",
  "totalVolume": 10.0,
  "durationMinutes": 60,
  "intervalSeconds": 300,
  "minPrice": 1.09900,
  "maxPrice": 1.10100
}
```

**Response:**
```json
{
  "id": "twap-456",
  "symbol": "EURUSD",
  "side": "BUY",
  "totalVolume": 10.0,
  "filledVolume": 0,
  "startTime": "2024-01-15T10:30:00Z",
  "endTime": "2024-01-15T11:30:00Z",
  "sliceCount": 12,
  "status": "PENDING"
}
```

### List Bracket Orders

**Endpoint:** `GET /api/orders/bracket/list`

**Response:**
```json
[
  {
    "id": "bracket-123",
    "symbol": "EURUSD",
    "status": "ACTIVE",
    ...
  }
]
```

---

## Technical Indicators

### Calculate Indicator

**Endpoint:** `GET /api/indicators/calculate`

**Query Parameters:**
- `symbol` (required): Trading symbol (e.g., "EURUSD")
- `indicator` (required): Indicator name
- `period` (optional): Period for calculation (default: 14)

**Supported Indicators:**
- `sma` - Simple Moving Average
- `ema` - Exponential Moving Average
- `wma` - Weighted Moving Average
- `rsi` - Relative Strength Index
- `macd` - Moving Average Convergence Divergence
- `bb` - Bollinger Bands
- `atr` - Average True Range
- `adx` - Average Directional Index
- `stochastic` - Stochastic Oscillator
- `pivot` - Pivot Points

**Example Request:**
```
GET /api/indicators/calculate?symbol=EURUSD&indicator=rsi&period=14
```

**Response (RSI):**
```json
[
  {
    "timestamp": 1705315200,
    "values": {
      "rsi": 65.34
    }
  },
  ...
]
```

**Response (MACD):**
```json
[
  {
    "timestamp": 1705315200,
    "values": {
      "macd": 0.00123,
      "signal": 0.00115,
      "histogram": 0.00008
    }
  },
  ...
]
```

**Response (Bollinger Bands):**
```json
[
  {
    "timestamp": 1705315200,
    "values": {
      "upper": 1.10250,
      "middle": 1.10000,
      "lower": 1.09750
    }
  },
  ...
]
```

**Response (Pivot Points):**
```json
{
  "pivot": 1.10000,
  "r1": 1.10150,
  "r2": 1.10300,
  "r3": 1.10450,
  "s1": 1.09850,
  "s2": 1.09700,
  "s3": 1.09550
}
```

---

## Strategy Automation

### List Strategies

**Endpoint:** `GET /api/strategies`

**Response:**
```json
[
  {
    "id": "strat-789",
    "name": "MA Crossover",
    "type": "INDICATOR",
    "mode": "PAPER",
    "symbols": ["EURUSD", "GBPUSD"],
    "enabled": true,
    "stats": {
      "totalTrades": 45,
      "winRate": 62.5,
      "netProfit": 1250.50
    }
  }
]
```

### Create Strategy

**Endpoint:** `POST /api/strategies/create`

**Request Body:**
```json
{
  "name": "RSI Mean Reversion",
  "description": "Buy oversold, sell overbought",
  "type": "INDICATOR",
  "mode": "PAPER",
  "symbols": ["EURUSD"],
  "timeframe": "M15",
  "config": {
    "indicator": "RSI",
    "period": 14,
    "oversold": 30,
    "overbought": 70
  },
  "maxPositions": 3,
  "maxRiskPerTrade": 2.0,
  "defaultStopLoss": 50,
  "defaultTakeProfit": 100
}
```

**Response:**
```json
{
  "id": "strat-890",
  "name": "RSI Mean Reversion",
  "enabled": false,
  "createdAt": "2024-01-15T10:30:00Z",
  ...
}
```

### Run Backtest

**Endpoint:** `POST /api/strategies/backtest`

**Request Body:**
```json
{
  "strategyId": "strat-789",
  "startDate": "2023-07-01T00:00:00Z",
  "endDate": "2024-01-01T00:00:00Z",
  "initialBalance": 10000.0
}
```

**Response:**
```json
{
  "strategyId": "strat-789",
  "startDate": "2023-07-01T00:00:00Z",
  "endDate": "2024-01-01T00:00:00Z",
  "initialBalance": 10000.0,
  "finalBalance": 12500.75,
  "stats": {
    "totalTrades": 125,
    "winningTrades": 78,
    "losingTrades": 47,
    "winRate": 62.4,
    "netProfit": 2500.75,
    "profitFactor": 1.85,
    "sharpeRatio": 1.42,
    "maxDrawdown": 850.25,
    "largestWin": 250.50,
    "largestLoss": -180.30
  },
  "equityCurve": [...],
  "completedAt": "2024-01-15T10:35:00Z"
}
```

---

## Alerts System

### Create Alert

**Endpoint:** `POST /api/alerts/create`

**Request Body (Price Alert):**
```json
{
  "userId": "user123",
  "name": "EURUSD Above 1.10",
  "type": "PRICE",
  "symbol": "EURUSD",
  "condition": "ABOVE",
  "value": 1.10000,
  "message": "EURUSD crossed above 1.10",
  "channels": ["EMAIL", "PUSH"],
  "triggerOnce": true
}
```

**Request Body (Indicator Alert):**
```json
{
  "userId": "user123",
  "name": "EURUSD RSI Overbought",
  "type": "INDICATOR",
  "symbol": "EURUSD",
  "indicator": "RSI",
  "period": 14,
  "condition": "OVERBOUGHT",
  "value": 70.0,
  "message": "EURUSD RSI is overbought",
  "channels": ["WEBHOOK"],
  "webhook": "https://your-server.com/webhook"
}
```

**Response:**
```json
{
  "id": "alert-123",
  "userId": "user123",
  "name": "EURUSD Above 1.10",
  "enabled": true,
  "triggered": false,
  "createdAt": "2024-01-15T10:30:00Z"
}
```

### List User Alerts

**Endpoint:** `GET /api/alerts/list?userId=user123`

**Response:**
```json
[
  {
    "id": "alert-123",
    "name": "EURUSD Above 1.10",
    "type": "PRICE",
    "enabled": true,
    "triggered": false
  },
  ...
]
```

### Get Alert Triggers

**Endpoint:** `GET /api/alerts/triggers?userId=user123`

**Response:**
```json
[
  {
    "id": "trigger-456",
    "alertId": "alert-123",
    "alertName": "EURUSD Above 1.10",
    "message": "EURUSD crossed above 1.10",
    "currentValue": 1.10050,
    "triggeredAt": "2024-01-15T10:45:00Z",
    "sent": true
  },
  ...
]
```

---

## Advanced Reports

### Tax Report

**Endpoint:** `GET /api/reports/tax?accountId=acc123&year=2024`

**Response:**
```json
{
  "accountId": "acc123",
  "year": 2024,
  "totalTrades": 250,
  "totalProfit": 15000.50,
  "totalLoss": -5000.25,
  "netProfit": 9999.25,
  "totalCommission": 500.00,
  "totalSwap": 150.75,
  "taxableProfitUsd": 9999.25,
  "shortTermGains": 8000.00,
  "longTermGains": 1999.25,
  "tradesBySymbol": {
    "EURUSD": {
      "symbol": "EURUSD",
      "trades": 120,
      "netProfit": 5000.00,
      "commission": 240.00
    },
    ...
  },
  "generatedAt": "2024-01-15T10:30:00Z"
}
```

### Performance Report

**Endpoint:** `GET /api/reports/performance?accountId=acc123&startDate=2023-10-01&endDate=2024-01-01`

**Response:**
```json
{
  "accountId": "acc123",
  "startDate": "2023-10-01T00:00:00Z",
  "endDate": "2024-01-01T00:00:00Z",
  "totalTrades": 125,
  "winningTrades": 78,
  "losingTrades": 47,
  "winRate": 62.4,
  "totalProfit": 8500.50,
  "totalLoss": -3250.25,
  "netProfit": 5250.25,
  "averageWin": 109.00,
  "averageLoss": -69.15,
  "largestWin": 450.00,
  "largestLoss": -280.50,
  "profitFactor": 2.62,
  "sharpeRatio": 1.85,
  "sortinoRatio": 2.45,
  "maxDrawdown": 950.00,
  "maxDrawdownPct": 9.5,
  "averageMAE": 45.20,
  "averageMFE": 125.80,
  "maeMfeRatio": 2.78,
  "averageRMultiple": 1.58,
  "longestWinStreak": 8,
  "longestLossStreak": 4,
  "profitBySymbol": {
    "EURUSD": 3250.50,
    "GBPUSD": 1500.25,
    "USDJPY": 499.50
  },
  "generatedAt": "2024-01-15T10:30:00Z"
}
```

### Drawdown Analysis

**Endpoint:** `GET /api/reports/drawdown?accountId=acc123&initialBalance=10000`

**Response:**
```json
{
  "accountId": "acc123",
  "currentDrawdown": 250.50,
  "currentDrawdownPct": 2.1,
  "maxDrawdown": 950.00,
  "maxDrawdownPct": 9.5,
  "maxDrawdownStart": "2023-11-15T00:00:00Z",
  "maxDrawdownEnd": "2023-11-28T00:00:00Z",
  "maxDrawdownDuration": 13,
  "recoveryTime": 8,
  "recoveryFactor": 5.53,
  "drawdownPeriods": [
    {
      "startDate": "2023-11-15T00:00:00Z",
      "endDate": "2023-11-28T00:00:00Z",
      "recoveryDate": "2023-12-06T00:00:00Z",
      "drawdown": 950.00,
      "drawdownPct": 9.5,
      "duration": 13,
      "recoveryTime": 8
    },
    ...
  ],
  "averageDrawdown": 450.00,
  "averageRecoveryTime": 5,
  "generatedAt": "2024-01-15T10:30:00Z"
}
```

---

## WebSocket Integration

All real-time updates (price alerts, strategy signals, etc.) can be streamed via WebSocket:

```javascript
const ws = new WebSocket('ws://localhost:7999/ws');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);

  if (data.type === 'alert_triggered') {
    console.log('Alert:', data.alertName, data.message);
  }

  if (data.type === 'strategy_signal') {
    console.log('Signal:', data.symbol, data.side, data.price);
  }
};
```

---

## Error Responses

All endpoints return standard error responses:

**400 Bad Request:**
```json
{
  "error": "Invalid request parameters"
}
```

**404 Not Found:**
```json
{
  "error": "Resource not found"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Internal server error"
}
```

---

## Rate Limiting

- No rate limiting currently implemented
- Recommended: 100 requests/minute per IP

---

## Authentication

Currently using the existing authentication system. All endpoints require valid session tokens.

**Header:**
```
Authorization: Bearer <token>
```

---

## Best Practices

1. **Backtesting**: Always backtest strategies before live trading
2. **Paper Trading**: Use paper trading mode to validate strategies
3. **Alerts**: Set reasonable thresholds to avoid alert fatigue
4. **TWAP/VWAP**: Use for large orders to minimize market impact
5. **Reports**: Generate regularly for compliance and analysis

---

## Support

For issues or questions:
- GitHub Issues: github.com/epic1st/rtx/issues
- Email: support@rtxtrading.com
