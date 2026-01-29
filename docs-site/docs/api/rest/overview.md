---
id: overview
title: REST API Overview
sidebar_label: Overview
sidebar_position: 1
description: Complete REST API reference for Trading Platform
---

# REST API Overview

The Trading Platform REST API provides programmatic access to all platform features using standard HTTP methods and JSON responses.

## Base URL

```
Production: https://api.yourtradingplatform.com/v1
Sandbox:    https://api-sandbox.yourtradingplatform.com/v1
```

## Authentication

All requests require an API key in the Authorization header:

```bash
curl -X GET https://api.yourtradingplatform.com/v1/account/balance \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json"
```

## Quick Reference

### Account Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/account/balance` | Get account balance |
| GET | `/account/info` | Get account information |
| GET | `/account/positions` | List all open positions |
| GET | `/account/orders` | List all orders |
| GET | `/account/history` | Get trade history |

### Trading Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/orders` | Place new order |
| GET | `/orders/:id` | Get order details |
| DELETE | `/orders/:id` | Cancel order |
| PUT | `/orders/:id` | Modify order |
| POST | `/positions/close` | Close position |

### Market Data Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/market/symbols` | List all symbols |
| GET | `/market/ticker/:symbol` | Get latest price |
| GET | `/market/depth/:symbol` | Get order book depth |
| GET | `/market/ohlc/:symbol` | Get OHLC candles |
| GET | `/market/ticks/:symbol` | Get tick data |

## Example Requests

### Get Account Balance

```bash
GET /v1/account/balance
```

Response:
```json
{
  "balance": 10000.00,
  "currency": "USD",
  "equity": 10250.50,
  "margin_used": 500.00,
  "margin_available": 9750.50,
  "margin_level": 2050.10,
  "profit_loss": 250.50
}
```

### Place Market Order

```bash
POST /v1/orders
Content-Type: application/json

{
  "symbol": "EURUSD",
  "type": "market",
  "side": "buy",
  "volume": 0.1,
  "stop_loss": 1.0800,
  "take_profit": 1.0900
}
```

Response:
```json
{
  "order_id": "ORD-123456789",
  "symbol": "EURUSD",
  "type": "market",
  "side": "buy",
  "volume": 0.1,
  "price": 1.0850,
  "stop_loss": 1.0800,
  "take_profit": 1.0900,
  "status": "filled",
  "filled_volume": 0.1,
  "timestamp": "2026-01-18T14:30:00Z"
}
```

### Get Market Data

```bash
GET /v1/market/ticker/EURUSD
```

Response:
```json
{
  "symbol": "EURUSD",
  "bid": 1.0849,
  "ask": 1.0851,
  "spread": 0.0002,
  "timestamp": "2026-01-18T14:30:00.123Z",
  "volume_24h": 125000000,
  "high_24h": 1.0875,
  "low_24h": 1.0825
}
```

For complete API reference, see individual endpoint documentation.
