# Real-Time Analytics API Design Patterns

## Overview

This document details the API design for the real-time analytics dashboard, covering WebSocket, REST, and SSE protocols with example implementations.

---

## 1. WebSocket API (Real-Time Prices & Orders)

### Connection Flow

```
Client                          Server
   |                               |
   |-- CONNECT /ws/prices -------> |
   |                               | (Authenticate)
   | <------ 200 OK/101 Upgrade --- |
   |                               |
   |-- SUBSCRIBE (JSON) ---------> | Subscribe to AAPL
   |                               | (Add to group)
   |                               |
   |                               | Kafka: "AAPL tick arrived"
   | <------ Price Update -------- |
   |                               |
   |-- UNSUBSCRIBE (JSON) -------> | Unsubscribe from AAPL
   |                               |
   |-- DISCONNECT ----------------> |
   |                               |
```

### WebSocket Message Protocol

**Subscribe Message:**
```json
{
  "type": "subscribe",
  "symbols": ["AAPL", "GOOGL", "MSFT"],
  "timeframe": "real-time",
  "includeDepth": true
}
```

**Unsubscribe Message:**
```json
{
  "type": "unsubscribe",
  "symbols": ["AAPL"]
}
```

**Initial Snapshot (Full Price Data):**
```json
{
  "type": "snapshot",
  "symbol": "AAPL",
  "price": 150.25,
  "bid": 150.24,
  "ask": 150.26,
  "volume": 1000000,
  "open": 149.50,
  "high": 150.75,
  "low": 149.25,
  "timestamp": 1705690245123,
  "sequence": 1000000,
  "exchange": "NASDAQ"
}
```

**Delta Update (Only Changed Fields):**
```json
{
  "type": "delta",
  "symbol": "AAPL",
  "price": 150.26,
  "bid": 150.25,
  "volume": 1000050,
  "timestamp": 1705690245124,
  "sequence": 1000001
}
```

**Error Message:**
```json
{
  "type": "error",
  "code": "INVALID_SYMBOL",
  "message": "Symbol INVALID not found",
  "timestamp": 1705690245125
}
```

### Server-Side Implementation (Node.js)

```typescript
import { WebSocketServer } from 'ws';
import Redis from 'redis';
import { Kafka } from 'kafkajs';

interface SubscriptionState {
  symbols: Set<string>;
  lastPrices: Map<string, PriceData>;
  subscriptionTime: number;
}

class WebSocketPriceServer {
  private wss: WebSocketServer;
  private redis: Redis.RedisClient;
  private kafka: Kafka;
  private subscriptions = new Map<string, SubscriptionState>();

  async handleConnection(ws: WebSocket, req: Request) {
    const clientId = req.headers['sec-websocket-key'];
    const state: SubscriptionState = {
      symbols: new Set(),
      lastPrices: new Map(),
      subscriptionTime: Date.now(),
    };

    this.subscriptions.set(clientId, state);

    ws.on('message', async (data: Buffer) => {
      try {
        const message = JSON.parse(data.toString());
        await this.handleMessage(ws, clientId, message, state);
      } catch (error) {
        this.sendError(ws, 'PARSE_ERROR', error.message);
      }
    });

    ws.on('close', () => {
      // Cleanup subscriptions
      this.subscriptions.delete(clientId);
      console.log(`Client ${clientId} disconnected`);
    });

    ws.on('error', (error) => {
      console.error(`WebSocket error for ${clientId}:`, error);
    });
  }

  private async handleMessage(
    ws: WebSocket,
    clientId: string,
    message: any,
    state: SubscriptionState
  ) {
    if (message.type === 'subscribe') {
      const symbols = Array.isArray(message.symbols)
        ? message.symbols
        : [message.symbols];

      for (const symbol of symbols) {
        // Validate symbol
        const exists = await this.redis.exists(`symbol:${symbol}`);
        if (!exists) {
          this.sendError(ws, 'INVALID_SYMBOL', `Symbol ${symbol} not found`);
          continue;
        }

        state.symbols.add(symbol);

        // Send current snapshot from cache
        const cached = await this.redis.get(`price:${symbol}`);
        if (cached) {
          const price = JSON.parse(cached);
          state.lastPrices.set(symbol, price);
          this.send(ws, {
            type: 'snapshot',
            ...price,
          });
        }

        // Subscribe to Kafka topic for updates
        this.subscribeToKafka(symbol, (price) => {
          if (state.symbols.has(symbol)) {
            this.sendUpdate(ws, symbol, price, state);
          }
        });
      }
    } else if (message.type === 'unsubscribe') {
      const symbols = Array.isArray(message.symbols)
        ? message.symbols
        : [message.symbols];

      for (const symbol of symbols) {
        state.symbols.delete(symbol);
        state.lastPrices.delete(symbol);
      }
    }
  }

  private sendUpdate(
    ws: WebSocket,
    symbol: string,
    price: PriceData,
    state: SubscriptionState
  ) {
    const lastPrice = state.lastPrices.get(symbol);

    if (!lastPrice) {
      // First update, send snapshot
      state.lastPrices.set(symbol, price);
      this.send(ws, {
        type: 'snapshot',
        symbol,
        ...price,
      });
    } else {
      // Send delta with only changed fields
      const delta: any = {
        type: 'delta',
        symbol,
        timestamp: price.timestamp,
        sequence: price.sequence,
      };

      if (price.price !== lastPrice.price) {
        delta.price = price.price;
      }
      if (price.bid !== lastPrice.bid) {
        delta.bid = price.bid;
      }
      if (price.ask !== lastPrice.ask) {
        delta.ask = price.ask;
      }
      if (price.volume !== lastPrice.volume) {
        delta.volume = price.volume;
      }

      state.lastPrices.set(symbol, price);
      this.send(ws, delta);
    }
  }

  private send(ws: WebSocket, data: any) {
    try {
      // Use MessagePack for compression
      const packed = msgpack.encode(data);
      ws.send(packed);
    } catch (error) {
      console.error('Failed to send message:', error);
    }
  }

  private sendError(ws: WebSocket, code: string, message: string) {
    this.send(ws, {
      type: 'error',
      code,
      message,
      timestamp: Date.now(),
    });
  }

  private subscribeToKafka(symbol: string, onMessage: (price: any) => void) {
    // Create Kafka consumer for this symbol
    const consumer = this.kafka.consumer({
      groupId: `ws-${symbol}`,
    });

    consumer.subscribe({ topic: 'trades', fromBeginning: false });

    consumer.run({
      eachMessage: async ({ topic, partition, message }) => {
        const trade = JSON.parse(message.value.toString());
        if (trade.symbol === symbol) {
          onMessage(trade);
        }
      },
    });
  }
}
```

---

## 2. REST API for Historical Data

### Endpoints

#### Get Last Price
```
GET /api/prices/:symbol

Response 200:
{
  "symbol": "AAPL",
  "price": 150.25,
  "bid": 150.24,
  "ask": 150.26,
  "volume": 1000000,
  "timestamp": 1705690245123,
  "exchange": "NASDAQ"
}

Response 404:
{
  "error": "SYMBOL_NOT_FOUND",
  "message": "Symbol AAPL not found"
}
```

#### Get OHLC Bars
```
GET /api/ohlc/:symbol?timeframe=1m&limit=100&startTime=1705600000000

Query Parameters:
  timeframe: 1m, 5m, 15m, 1h, 1d
  limit: 1-1000 (default 100)
  startTime: Unix timestamp (ms)
  endTime: Unix timestamp (ms)

Response 200:
[
  {
    "symbol": "AAPL",
    "time": 1705690200000,
    "open": 150.10,
    "high": 150.75,
    "low": 149.90,
    "close": 150.25,
    "volume": 100000
  },
  ...
]
```

#### Get Trade History
```
GET /api/trades/:symbol?limit=100&offset=0&startTime=...&endTime=...

Response 200:
[
  {
    "symbol": "AAPL",
    "price": 150.25,
    "volume": 1000,
    "side": "BUY",
    "timestamp": 1705690245123,
    "sequence": 1000000
  },
  ...
]
```

#### Get Statistics
```
GET /api/stats/:symbol?period=1d

Query Parameters:
  period: 1h, 1d, 1w, 1m, ytd, all

Response 200:
{
  "symbol": "AAPL",
  "period": "1d",
  "open": 149.50,
  "high": 150.75,
  "low": 149.25,
  "close": 150.25,
  "volume": 50000000,
  "avgPrice": 150.10,
  "tickCount": 100000,
  "trades": 50000,
  "volatility": 0.025,
  "startTime": 1705603200000,
  "endTime": 1705689600000
}
```

### Implementation

```typescript
import Express from 'express';
import { ClickHouse } from '@clickhouse/client';
import Redis from 'redis';

const app = Express();
const ch = new ClickHouse({
  url: 'http://localhost:8123',
});
const redis = Redis.createClient();

// Get last price
app.get('/api/prices/:symbol', async (req, res) => {
  const { symbol } = req.params;

  try {
    // Try cache first
    const cached = await redis.get(`price:${symbol}`);
    if (cached) {
      return res.json(JSON.parse(cached));
    }

    // Query ClickHouse
    const result = await ch.query({
      query: `
        SELECT symbol, price, bid, ask, volume, timestamp, exchange
        FROM trades
        WHERE symbol = ?
        ORDER BY timestamp DESC
        LIMIT 1
      `,
      query_params: [symbol],
    });

    if (result.length === 0) {
      return res.status(404).json({
        error: 'SYMBOL_NOT_FOUND',
        message: `Symbol ${symbol} not found`,
      });
    }

    const price = result[0];

    // Cache for 5 seconds
    await redis.setex(`price:${symbol}`, 5, JSON.stringify(price));

    res.json(price);
  } catch (error) {
    res.status(500).json({
      error: 'INTERNAL_ERROR',
      message: error.message,
    });
  }
});

// Get OHLC bars
app.get('/api/ohlc/:symbol', async (req, res) => {
  const { symbol } = req.params;
  const { timeframe = '1m', limit = 100, startTime, endTime } = req.query;

  try {
    // Validate inputs
    if (!['1m', '5m', '15m', '1h', '1d'].includes(timeframe)) {
      return res.status(400).json({ error: 'Invalid timeframe' });
    }

    const bucketSize = this.getTimeBucketSize(timeframe);
    const cacheKey = `ohlc:${symbol}:${timeframe}`;

    // Try cache
    const cached = await redis.get(cacheKey);
    if (cached) {
      return res.json(JSON.parse(cached));
    }

    // Query ClickHouse
    const query = `
      SELECT
        symbol,
        toStartOfInterval(timestamp, INTERVAL ${bucketSize}) as time,
        first(price) as open,
        max(price) as high,
        min(price) as low,
        last(price) as close,
        sum(volume) as volume
      FROM trades
      WHERE symbol = ?
        ${startTime ? `AND timestamp >= fromUnixTimestamp(${startTime / 1000})` : ''}
        ${endTime ? `AND timestamp <= fromUnixTimestamp(${endTime / 1000})` : ''}
      GROUP BY symbol, time
      ORDER BY time DESC
      LIMIT ?
    `;

    const result = await ch.query({
      query,
      query_params: [symbol, parseInt(limit)],
    });

    // Cache for 1 minute
    await redis.setex(cacheKey, 60, JSON.stringify(result));

    res.json(result);
  } catch (error) {
    res.status(500).json({
      error: 'INTERNAL_ERROR',
      message: error.message,
    });
  }
});

// Get statistics
app.get('/api/stats/:symbol', async (req, res) => {
  const { symbol } = req.params;
  const { period = '1d' } = req.query;

  try {
    const timeFilter = this.getTimeFilter(period);

    const result = await ch.query({
      query: `
        SELECT
          symbol,
          first(price) as open,
          max(price) as high,
          min(price) as low,
          last(price) as close,
          sum(volume) as volume,
          avg(price) as avgPrice,
          count() as tickCount,
          count(distinct timestamp) as trades,
          stddevPop(price) as volatility,
          min(timestamp) as startTime,
          max(timestamp) as endTime
        FROM trades
        WHERE symbol = ? ${timeFilter}
        GROUP BY symbol
      `,
      query_params: [symbol],
    });

    if (result.length === 0) {
      return res.status(404).json({
        error: 'SYMBOL_NOT_FOUND',
        message: `Symbol ${symbol} not found`,
      });
    }

    res.json(result[0]);
  } catch (error) {
    res.status(500).json({
      error: 'INTERNAL_ERROR',
      message: error.message,
    });
  }
});

private getTimeFilter(period: string): string {
  const now = Math.floor(Date.now() / 1000);
  const periods: { [key: string]: number } = {
    '1h': 3600,
    '1d': 86400,
    '1w': 604800,
    '1m': 2592000,
    'ytd': this.getYTDSeconds(),
  };

  if (period === 'all') {
    return '';
  }

  const seconds = periods[period] || 86400;
  return `AND timestamp > fromUnixTimestamp(${now - seconds})`;
}

private getTimeBucketSize(timeframe: string): string {
  const buckets: { [key: string]: string } = {
    '1m': '1 minute',
    '5m': '5 minutes',
    '15m': '15 minutes',
    '1h': '1 hour',
    '1d': '1 day',
  };
  return buckets[timeframe];
}
```

---

## 3. Server-Sent Events (SSE) for Broadcasts

### Connection Flow

```
Client                        Server
   |                             |
   |-- GET /api/events --------> |
   |                             | (Authenticate)
   | <---- 200 OK, text/event-- |
   |       stream                |
   |                             |
   | <---- data: {...} -------- | News event
   |       retry: 5000           |
   |                             |
   | <---- data: {...} -------- | Alert event
   |                             |
   | [Network failure]           |
   |                             |
   |-- Auto-reconnect ---------> | (Browser native)
   |                             |
   | <---- data: {...} -------- | Resume from last ID
   |       id: 1000              |
   |                             |
```

### SSE Message Protocol

**News Event:**
```
event: news
id: 1000
data: {"title": "Fed announces rate decision", "source": "Reuters", "timestamp": 1705690245123}

retry: 5000
```

**Price Alert:**
```
event: price_alert
id: 1001
data: {"symbol": "AAPL", "price": 150.25, "alert_type": "threshold_exceeded", "threshold": 150.00}
```

**System Status:**
```
event: system_status
id: 1002
data: {"status": "maintenance_window", "duration": 60, "affected_symbols": ["AAPL", "GOOGL"]}
```

### Server-Side Implementation

```typescript
app.get('/api/events', authMiddleware, (req, res) => {
  const userId = req.user.id;
  const lastEventId = parseInt(req.headers['last-event-id'] || '0');

  // Set SSE headers
  res.setHeader('Content-Type', 'text/event-stream');
  res.setHeader('Cache-Control', 'no-cache');
  res.setHeader('Connection', 'keep-alive');
  res.setHeader('Access-Control-Allow-Origin', '*');

  // Send event queue missed while offline
  this.sendMissedEvents(res, userId, lastEventId);

  // Subscribe to event broadcaster
  const unsubscribe = eventBroadcaster.subscribe((event) => {
    if (event.shouldSendTo(userId)) {
      res.write(`event: ${event.type}\n`);
      res.write(`id: ${event.id}\n`);
      res.write(`data: ${JSON.stringify(event.data)}\n`);
      res.write(`retry: 5000\n\n`);
    }
  });

  // Keep-alive ping every 30 seconds
  const keepAlive = setInterval(() => {
    res.write(': keep-alive\n\n');
  }, 30000);

  // Cleanup on disconnect
  req.on('close', () => {
    clearInterval(keepAlive);
    unsubscribe();
  });
});
```

### Client-Side

```typescript
class EventClient {
  private eventSource: EventSource;
  private lastEventId = 0;

  connect() {
    const url = `/api/events?token=${this.token}`;

    this.eventSource = new EventSource(url, {
      withCredentials: true,
    });

    this.eventSource.addEventListener('news', (event) => {
      const data = JSON.parse(event.data);
      this.onNews(data);
    });

    this.eventSource.addEventListener('price_alert', (event) => {
      const data = JSON.parse(event.data);
      this.onPriceAlert(data);
    });

    this.eventSource.addEventListener('system_status', (event) => {
      const data = JSON.parse(event.data);
      this.onSystemStatus(data);
    });

    this.eventSource.onerror = () => {
      console.error('SSE connection error, reconnecting...');
      // Browser auto-reconnects with last-event-id header
    };
  }

  disconnect() {
    this.eventSource.close();
  }

  private onNews(data: any) {
    console.log('News:', data);
    // Update UI
  }

  private onPriceAlert(data: any) {
    console.log('Price Alert:', data);
    // Show notification
  }

  private onSystemStatus(data: any) {
    console.log('System Status:', data);
    // Update status indicator
  }
}

// Usage
const client = new EventClient('token123');
client.connect();
```

---

## 4. Error Handling & Resilience

### Error Response Format

```json
{
  "error": "ERROR_CODE",
  "message": "Human-readable message",
  "details": {
    "symbol": "INVALID",
    "field": "symbol",
    "validation": "not_found"
  },
  "timestamp": 1705690245123,
  "requestId": "req-12345-abcde"
}
```

### Common Error Codes

```
400 - BAD_REQUEST              Invalid input parameters
401 - UNAUTHORIZED             Missing/invalid authentication
403 - FORBIDDEN                Insufficient permissions
404 - NOT_FOUND               Resource not found
429 - RATE_LIMITED            Too many requests
500 - INTERNAL_ERROR          Server error
503 - SERVICE_UNAVAILABLE     Temporary unavailability
```

### Retry Strategy

```typescript
async function withRetry<T>(
  fn: () => Promise<T>,
  maxRetries = 3,
  backoff = 100
): Promise<T> {
  let lastError;

  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error;

      // Don't retry on client errors
      if (error.status >= 400 && error.status < 500) {
        throw error;
      }

      // Wait before retry (exponential backoff)
      if (attempt < maxRetries) {
        await sleep(backoff * Math.pow(2, attempt - 1));
      }
    }
  }

  throw lastError;
}

// Usage
const price = await withRetry(() =>
  fetch(`/api/prices/${symbol}`).then((r) => r.json())
);
```

---

## 5. Rate Limiting

### Per-User Limits

```
WebSocket subscriptions:   1000 symbols per connection
REST requests:             100 requests/minute
Historical data queries:   10 concurrent queries
Large batch requests:      1 request/second
```

### Implementation

```typescript
import rateLimit from 'express-rate-limit';

const generalLimiter = rateLimit({
  windowMs: 60 * 1000, // 1 minute
  max: 100, // 100 requests
  message: { error: 'RATE_LIMITED' },
  standardHeaders: true,
  legacyHeaders: false,
});

const historicalLimiter = rateLimit({
  windowMs: 60 * 1000,
  max: 10, // 10 concurrent
  skipSuccessfulRequests: true,
});

app.use('/api/', generalLimiter);
app.get('/api/historical/:symbol', historicalLimiter, (req, res) => {
  // ...
});
```

---

## 6. Authentication & Authorization

### JWT Token

```typescript
interface JWTPayload {
  userId: string;
  username: string;
  permissions: string[];
  symbolicAccess: string[]; // Which symbols user can access
  issuedAt: number;
  expiresAt: number;
}

// Verify WebSocket connection
wss.on('connection', (ws, req) => {
  const token = req.url.split('token=')[1];

  try {
    const payload = jwt.verify(token, process.env.JWT_SECRET);

    ws.userId = payload.userId;
    ws.permissions = payload.permissions;
    ws.symbolAccess = payload.symbolicAccess;
  } catch (error) {
    ws.close(4001, 'Invalid token');
  }
});

// Verify REST requests
app.use(authMiddleware);

function authMiddleware(req, res, next) {
  const token = req.headers.authorization?.split(' ')[1];

  if (!token) {
    return res.status(401).json({ error: 'UNAUTHORIZED' });
  }

  try {
    req.user = jwt.verify(token, process.env.JWT_SECRET);
    next();
  } catch (error) {
    res.status(401).json({ error: 'UNAUTHORIZED' });
  }
}
```

---

## Summary: Choosing the Right Protocol

| Use Case | Protocol | Reason |
|----------|----------|--------|
| **Live prices** | WebSocket | Bidirectional, low latency, high frequency |
| **Order placement** | WebSocket | Requires real-time response |
| **Market news** | SSE | One-way broadcast, auto-reconnect |
| **System alerts** | SSE | One-way, simple infrastructure |
| **Historical data** | REST | One-off queries, easier caching |
| **Complex reports** | REST | Async job, large responses |
| **User watchlists** | REST + WS | REST for updates, WS for subscriptions |

---

End of API Design Patterns Document
