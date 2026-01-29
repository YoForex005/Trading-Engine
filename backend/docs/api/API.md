# RTX Trading Engine - API Documentation

## Table of Contents

1. [Introduction](#introduction)
2. [Quick Start](#quick-start)
3. [Authentication](#authentication)
4. [Base URL](#base-url)
5. [API Endpoints](#api-endpoints)
   - [Authentication](#authentication-endpoints)
   - [B-Book Orders](#b-book-orders)
   - [A-Book Orders](#a-book-orders)
   - [Position Management](#position-management)
   - [Account Information](#account-information)
   - [Market Data](#market-data)
   - [Risk Management](#risk-management)
   - [Admin Endpoints](#admin-endpoints)
6. [WebSocket API](#websocket-api)
7. [Error Handling](#error-handling)
8. [Rate Limiting](#rate-limiting)
9. [Code Examples](#code-examples)

## Introduction

The RTX Trading Engine is a hybrid B-Book/A-Book trading platform that provides:

- **B-Book Execution:** Internal order matching using internal balance/equity
- **A-Book Execution:** LP passthrough to OANDA, Binance, YoFX (via FIX 4.4)
- **Real-time Market Data:** WebSocket streaming for live prices
- **Multi-LP Aggregation:** Best price aggregation from multiple liquidity providers
- **Complete Order Management:** Market, limit, stop, and stop-limit orders
- **Risk Management:** Margin calculations, position sizing, trailing stops
- **Admin Controls:** LP routing, symbol management, account administration

**Architecture:**
```
┌─────────────┐
│   Clients   │
└──────┬──────┘
       │ REST API / WebSocket
┌──────▼───────────────────┐
│   RTX Trading Engine     │
├──────────────────────────┤
│  B-Book      │  A-Book   │
│  (Internal)  │  (LP)     │
└──────────────┴───────────┘
       │
┌──────▼──────────────────────────┐
│  Liquidity Providers (LPs)      │
├───────────┬──────────┬───────────┤
│  OANDA    │ Binance  │  YoFX     │
│  (REST)   │ (WS)     │  (FIX)    │
└───────────┴──────────┴───────────┘
```

## Quick Start

### 1. Login

```bash
curl -X POST http://localhost:7999/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "0",
    "username": "admin",
    "role": "ADMIN"
  }
}
```

### 2. Get Account Summary

```bash
curl -X GET http://localhost:7999/api/account/summary \
  -H "Authorization: Bearer <your-token>"
```

### 3. Place Market Order (B-Book)

```bash
curl -X POST http://localhost:7999/api/orders/market \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "EURUSD",
    "side": "BUY",
    "volume": 0.1,
    "sl": 1.0850,
    "tp": 1.1050
  }'
```

### 4. Connect to WebSocket

```javascript
const ws = new WebSocket('ws://localhost:7999/ws');

ws.onmessage = (event) => {
  const tick = JSON.parse(event.data);
  console.log(`${tick.symbol}: ${tick.bid} / ${tick.ask}`);
};
```

## Authentication

All endpoints except `/health`, `/login`, and `/docs` require JWT authentication.

**Include token in Authorization header:**
```
Authorization: Bearer <your-jwt-token>
```

**Token lifetime:** 24 hours

See [AUTHENTICATION.md](./AUTHENTICATION.md) for complete details.

## Base URL

- **Development:** `http://localhost:7999`
- **Production:** `https://api.rtxtrading.com`

## API Endpoints

### Authentication Endpoints

#### POST /login

Authenticate and receive JWT token.

**Request:**
```json
{
  "username": "admin",
  "password": "password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "0",
    "username": "admin",
    "role": "ADMIN"
  }
}
```

---

### B-Book Orders

Internal execution using RTX balance/equity.

#### POST /api/orders/market

Execute market order internally (B-Book).

**Request:**
```json
{
  "accountId": 1,
  "symbol": "EURUSD",
  "side": "BUY",
  "volume": 0.1,
  "sl": 1.0850,
  "tp": 1.1050
}
```

**Response:**
```json
{
  "success": true,
  "position": {
    "id": 12345,
    "accountId": 1,
    "symbol": "EURUSD",
    "side": "BUY",
    "volume": 0.1,
    "entryPrice": 1.0950,
    "sl": 1.0850,
    "tp": 1.1050,
    "openTime": "2026-01-18T12:00:00Z"
  }
}
```

**Parameters:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| accountId | integer | No | Account ID (default: 1) |
| symbol | string | Yes | Trading symbol (e.g., EURUSD) |
| side | string | Yes | BUY or SELL |
| volume | number | Yes | Lot size |
| sl | number | No | Stop loss price |
| tp | number | No | Take profit price |

---

### A-Book Orders

LP passthrough execution (OANDA/YoFX).

#### POST /order

Route order to liquidity provider.

**Request:**
```json
{
  "accountId": "demo_001",
  "symbol": "EURUSD",
  "side": "BUY",
  "volume": 0.1,
  "type": "MARKET",
  "sl": 1.0850,
  "tp": 1.1050
}
```

**Response:**
```json
{
  "success": true,
  "order": {
    "id": "LP-12345",
    "symbol": "EURUSD",
    "side": "BUY",
    "volume": 0.1,
    "status": "FILLED"
  },
  "message": "Order sent to LP"
}
```

#### POST /order/limit

Place limit order via LP.

**Request:**
```json
{
  "symbol": "GBPUSD",
  "side": "BUY",
  "volume": 0.5,
  "price": 1.2650,
  "sl": 1.2600,
  "tp": 1.2750
}
```

#### POST /order/stop

Place stop order.

**Request:**
```json
{
  "symbol": "BTCUSD",
  "side": "SELL",
  "volume": 0.01,
  "triggerPrice": 95000
}
```

#### GET /orders/pending

Get all pending orders.

**Response:**
```json
[
  {
    "id": "ORD-12345",
    "symbol": "GBPUSD",
    "side": "BUY",
    "type": "LIMIT",
    "volume": 0.5,
    "price": 1.2650,
    "status": "PENDING"
  }
]
```

#### POST /order/cancel

Cancel pending order.

**Request:**
```json
{
  "orderId": "ORD-12345"
}
```

---

### Position Management

#### GET /api/positions

Get open positions.

**Query Parameters:**
- `accountId` (optional): Filter by account

**Response:**
```json
[
  {
    "id": 12345,
    "accountId": 1,
    "symbol": "EURUSD",
    "side": "BUY",
    "volume": 0.1,
    "entryPrice": 1.0950,
    "currentPrice": 1.0975,
    "unrealizedPL": 25.00,
    "sl": 1.0850,
    "tp": 1.1050,
    "openTime": "2026-01-18T12:00:00Z"
  }
]
```

#### POST /api/positions/close

Close position (full or partial).

**Request:**
```json
{
  "positionId": 12345,
  "volume": 0.05
}
```

**Response:**
```json
{
  "success": true,
  "position": {
    "id": 12345,
    "volume": 0.05,
    "realizedPL": 12.50
  }
}
```

#### POST /api/positions/modify

Modify stop loss and take profit.

**Request:**
```json
{
  "positionId": 12345,
  "sl": 1.0900,
  "tp": 1.1100
}
```

#### POST /position/trailing-stop

Set trailing stop.

**Request:**
```json
{
  "tradeId": "12345",
  "symbol": "EURUSD",
  "side": "BUY",
  "type": "FIXED",
  "distance": 50
}
```

**Trailing Stop Types:**
- `FIXED`: Fixed distance
- `STEP`: Step-based trailing
- `ATR`: ATR-based trailing

---

### Account Information

#### GET /api/account/summary

Get account balance, equity, margin.

**Response:**
```json
{
  "accountId": 1,
  "balance": 5000.00,
  "equity": 5125.50,
  "margin": 250.00,
  "freeMargin": 4875.50,
  "marginLevel": 2050.20,
  "unrealizedPL": 125.50,
  "realizedPL": 0.00
}
```

**Fields:**

| Field | Description |
|-------|-------------|
| balance | Account balance |
| equity | Balance + unrealized P/L |
| margin | Used margin for open positions |
| freeMargin | Available margin for new orders |
| marginLevel | (Equity / Margin) * 100 |
| unrealizedPL | Floating profit/loss |
| realizedPL | Closed position P/L |

#### POST /api/account/create

Create new trading account (admin only).

**Request:**
```json
{
  "username": "trader001",
  "fullName": "John Doe",
  "password": "securePassword123",
  "isDemo": false
}
```

---

### Market Data

#### GET /ticks

Get tick history.

**Query Parameters:**
- `symbol` (required): Trading symbol
- `limit` (optional): Max ticks (default: 500)

**Response:**
```json
[
  {
    "symbol": "EURUSD",
    "bid": 1.09523,
    "ask": 1.09525,
    "spread": 0.00002,
    "timestamp": 1705838400000,
    "lp": "OANDA"
  }
]
```

#### GET /ohlc

Get OHLC (candlestick) data.

**Query Parameters:**
- `symbol` (required)
- `timeframe` (optional): 1m, 5m, 15m, 1h, 4h, 1d (default: 1h)
- `limit` (optional): Max candles (default: 500)

**Response:**
```json
[
  {
    "timestamp": 1705838400000,
    "open": 1.0950,
    "high": 1.0975,
    "low": 1.0945,
    "close": 1.0970,
    "volume": 1250
  }
]
```

---

### Risk Management

#### GET /risk/calculate-lot

Calculate lot size from risk percentage.

**Query Parameters:**
- `symbol` (required)
- `riskPercent` (required): Risk % (e.g., 2.0 for 2%)
- `slPips` (required): Stop loss distance in pips

**Response:**
```json
{
  "symbol": "EURUSD",
  "riskPercent": 2.0,
  "slPips": 50,
  "recommendedLot": 0.4,
  "riskAmount": 100.00
}
```

#### GET /risk/margin-preview

Preview margin requirements.

**Query Parameters:**
- `symbol` (required)
- `volume` (required)
- `side` (optional)

**Response:**
```json
{
  "requiredMargin": 250.00,
  "freeMarginAfter": 4750.00,
  "marginLevel": 1900.00
}
```

---

### Admin Endpoints

#### GET /admin/accounts

List all trading accounts.

**Response:**
```json
[
  {
    "id": 1,
    "accountNumber": "ACC-001",
    "username": "trader001",
    "fullName": "John Doe",
    "balance": 5000.00,
    "equity": 5125.50,
    "isDemo": false,
    "createdAt": "2026-01-01T00:00:00Z"
  }
]
```

#### POST /admin/deposit

Deposit funds to account.

**Request:**
```json
{
  "accountId": 1,
  "amount": 1000.00,
  "method": "BANK_TRANSFER",
  "reference": "TXN-123456"
}
```

**Methods:** BANK_TRANSFER, CRYPTO, CARD

#### POST /admin/withdraw

Withdraw funds.

**Request:**
```json
{
  "accountId": 1,
  "amount": 500.00,
  "method": "BANK_TRANSFER"
}
```

#### GET /admin/lps

List liquidity providers.

**Response:**
```json
[
  {
    "id": "oanda",
    "name": "OANDA",
    "type": "REST",
    "enabled": true,
    "priority": 1,
    "symbols": ["EURUSD", "GBPUSD", "USDJPY"]
  }
]
```

#### POST /admin/lps

Add new liquidity provider.

**Request:**
```json
{
  "id": "binance",
  "name": "Binance",
  "type": "WEBSOCKET",
  "enabled": true,
  "priority": 2,
  "symbols": ["BTCUSD", "ETHUSD"]
}
```

#### POST /admin/lps/{lpId}/toggle

Enable/disable LP.

#### GET /admin/lp-status

Get LP connection status.

**Response:**
```json
{
  "lps": [
    {
      "id": "oanda",
      "name": "OANDA",
      "enabled": true,
      "connected": true,
      "lastQuoteTime": 1705838400000
    }
  ]
}
```

#### GET /admin/fix/status

Get FIX session status.

**Response:**
```json
{
  "sessions": {
    "YOFX1": "CONNECTED",
    "YOFX2": "DISCONNECTED"
  }
}
```

#### POST /admin/fix/connect

Connect FIX session.

**Request:**
```json
{
  "sessionId": "YOFX1"
}
```

#### POST /admin/fix/disconnect

Disconnect FIX session.

#### GET /api/config

Get broker configuration.

**Response:**
```json
{
  "brokerName": "RTX Trading",
  "priceFeedLP": "OANDA",
  "executionMode": "BBOOK",
  "defaultLeverage": 100,
  "defaultBalance": 5000.00,
  "marginMode": "HEDGING",
  "maxTicksPerSymbol": 50000
}
```

#### POST /api/config

Update broker configuration (admin only).

---

## WebSocket API

**Endpoint:** `ws://localhost:7999/ws`

See [WEBHOOKS.md](./WEBHOOKS.md) for complete WebSocket documentation.

### Market Ticks

**Message Format:**
```json
{
  "type": "tick",
  "symbol": "EURUSD",
  "bid": 1.09523,
  "ask": 1.09525,
  "spread": 0.00002,
  "timestamp": 1705838400000,
  "lp": "OANDA"
}
```

**Connection Example:**
```javascript
const ws = new WebSocket('ws://localhost:7999/ws');

ws.onopen = () => console.log('Connected');

ws.onmessage = (event) => {
  const tick = JSON.parse(event.data);
  if (tick.type === 'tick') {
    updateChart(tick.symbol, tick.bid);
  }
};
```

---

## Error Handling

See [ERRORS.md](./ERRORS.md) for complete error documentation.

**Error Response Format:**
```json
{
  "error": "Human-readable message",
  "code": "ERROR_CODE",
  "details": {
    "field": "additional context"
  }
}
```

**Common Error Codes:**

| Code | HTTP | Description |
|------|------|-------------|
| AUTH_REQUIRED | 401 | Missing authentication |
| AUTH_EXPIRED | 401 | Token expired |
| MARGIN_INSUFFICIENT | 400 | Not enough margin |
| INVALID_SYMBOL | 400 | Unknown symbol |
| POSITION_NOT_FOUND | 404 | Position not found |
| LP_NOT_CONNECTED | 503 | LP unavailable |

---

## Rate Limiting

See [RATE_LIMITS.md](./RATE_LIMITS.md) for details.

**Headers:**
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1705838445
```

**Current Limits (Planned):**
- Orders: 100 requests/minute
- Market data: 120 requests/minute
- Admin: 200 requests/minute

---

## Code Examples

### JavaScript/TypeScript

```typescript
class RTXClient {
  constructor(private baseURL: string, private token: string) {}

  async placeOrder(symbol: string, side: 'BUY' | 'SELL', volume: number) {
    const response = await fetch(`${this.baseURL}/api/orders/market`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ symbol, side, volume })
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error);
    }

    return response.json();
  }

  async getPositions() {
    const response = await fetch(`${this.baseURL}/api/positions`, {
      headers: { 'Authorization': `Bearer ${this.token}` }
    });
    return response.json();
  }

  async closePosition(positionId: number, volume?: number) {
    const response = await fetch(`${this.baseURL}/api/positions/close`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ positionId, volume })
    });
    return response.json();
  }
}

// Usage
const client = new RTXClient('http://localhost:7999', token);
const order = await client.placeOrder('EURUSD', 'BUY', 0.1);
const positions = await client.getPositions();
```

### Python

```python
import requests

class RTXClient:
    def __init__(self, base_url: str, token: str):
        self.base_url = base_url
        self.token = token
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }

    def place_order(self, symbol: str, side: str, volume: float):
        response = requests.post(
            f'{self.base_url}/api/orders/market',
            headers=self.headers,
            json={'symbol': symbol, 'side': side, 'volume': volume}
        )
        response.raise_for_status()
        return response.json()

    def get_positions(self):
        response = requests.get(
            f'{self.base_url}/api/positions',
            headers=self.headers
        )
        return response.json()

    def close_position(self, position_id: int, volume: float = None):
        response = requests.post(
            f'{self.base_url}/api/positions/close',
            headers=self.headers,
            json={'positionId': position_id, 'volume': volume}
        )
        return response.json()

# Usage
client = RTXClient('http://localhost:7999', token)
order = client.place_order('EURUSD', 'BUY', 0.1)
positions = client.get_positions()
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type RTXClient struct {
    BaseURL string
    Token   string
    Client  *http.Client
}

type OrderRequest struct {
    Symbol string  `json:"symbol"`
    Side   string  `json:"side"`
    Volume float64 `json:"volume"`
}

func (c *RTXClient) PlaceOrder(symbol, side string, volume float64) error {
    req := OrderRequest{Symbol: symbol, Side: side, Volume: volume}
    body, _ := json.Marshal(req)

    httpReq, _ := http.NewRequest("POST", c.BaseURL+"/api/orders/market", bytes.NewBuffer(body))
    httpReq.Header.Set("Authorization", "Bearer "+c.Token)
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := c.Client.Do(httpReq)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return fmt.Errorf("order failed: %d", resp.StatusCode)
    }

    return nil
}

func main() {
    client := &RTXClient{
        BaseURL: "http://localhost:7999",
        Token:   "your-token",
        Client:  &http.Client{},
    }

    client.PlaceOrder("EURUSD", "BUY", 0.1)
}
```

---

## Additional Resources

- [OpenAPI Specification](./openapi.yaml) - Complete OpenAPI 3.0 spec
- [Postman Collection](./postman-collection.json) - Ready-to-import collection
- [Authentication Guide](./AUTHENTICATION.md) - JWT auth details
- [WebSocket Guide](./WEBHOOKS.md) - Real-time data streaming
- [Error Reference](./ERRORS.md) - Complete error codes
- [Rate Limits](./RATE_LIMITS.md) - API rate limiting

---

## Support

For issues, questions, or feature requests:

- **Email:** support@rtxtrading.com
- **GitHub:** https://github.com/epic1st/rtx
- **Documentation:** http://localhost:7999/docs

---

**Version:** 3.0.0
**Last Updated:** 2026-01-18
