# WebSocket API Documentation

## Overview

The RTX Trading Engine provides real-time market data and account updates via WebSocket connections.

**WebSocket Endpoint:** `ws://localhost:7999/ws`

**Production:** `wss://api.rtxtrading.com/ws`

## Connection

### Establishing Connection

```javascript
const ws = new WebSocket('ws://localhost:7999/ws');

ws.onopen = () => {
  console.log('Connected to RTX Trading Engine');
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  handleMessage(data);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected from RTX Trading Engine');
};
```

### Python Example

```python
import websocket
import json

def on_message(ws, message):
    data = json.loads(message)
    print(f"Received: {data}")

def on_error(ws, error):
    print(f"Error: {error}")

def on_close(ws, close_status_code, close_msg):
    print("Connection closed")

def on_open(ws):
    print("Connected to RTX Trading Engine")

ws = websocket.WebSocketApp("ws://localhost:7999/ws",
                            on_open=on_open,
                            on_message=on_message,
                            on_error=on_error,
                            on_close=on_close)

ws.run_forever()
```

## Message Types

### 1. Market Tick (Price Updates)

Real-time price updates for all enabled symbols.

**Message Structure:**

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

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Always "tick" for price updates |
| `symbol` | string | Trading symbol (e.g., EURUSD, BTCUSD) |
| `bid` | number | Bid price |
| `ask` | number | Ask price |
| `spread` | number | Ask - Bid |
| `timestamp` | integer | Unix timestamp in milliseconds |
| `lp` | string | Liquidity provider source |

**Frequency:** Real-time as prices change

**Example Handler:**

```javascript
function handleMessage(data) {
  if (data.type === 'tick') {
    updatePriceDisplay(data.symbol, data.bid, data.ask);
    updateChart(data.symbol, data.bid, data.timestamp);
  }
}
```

### 2. Position Updates (Future)

**Planned:** Real-time position P/L updates

```json
{
  "type": "position_update",
  "positionId": 12345,
  "unrealizedPL": 125.50,
  "currentPrice": 1.09545
}
```

### 3. Order Filled (Future)

**Planned:** Order execution notifications

```json
{
  "type": "order_filled",
  "orderId": "ORD-12345",
  "symbol": "EURUSD",
  "side": "BUY",
  "volume": 0.1,
  "fillPrice": 1.09525,
  "timestamp": 1705838400000
}
```

### 4. Account Update (Future)

**Planned:** Balance and equity changes

```json
{
  "type": "account_update",
  "accountId": 1,
  "balance": 5125.50,
  "equity": 5250.75,
  "margin": 250.00,
  "freeMargin": 5000.75
}
```

## Subscription Management (Future)

**Planned feature:** Subscribe to specific symbols

```json
// Subscribe to specific symbols
{
  "action": "subscribe",
  "symbols": ["EURUSD", "BTCUSD", "XAUUSD"]
}

// Unsubscribe
{
  "action": "unsubscribe",
  "symbols": ["XAUUSD"]
}

// Subscribe to all
{
  "action": "subscribe_all"
}
```

## Connection Management

### Heartbeat / Ping-Pong

The server automatically maintains the connection. No explicit ping/pong required.

### Reconnection Strategy

Implement exponential backoff for reconnection:

```javascript
let reconnectDelay = 1000; // Start at 1 second
const maxDelay = 30000; // Max 30 seconds

function connect() {
  const ws = new WebSocket('ws://localhost:7999/ws');

  ws.onclose = () => {
    console.log(`Reconnecting in ${reconnectDelay}ms...`);
    setTimeout(() => {
      reconnectDelay = Math.min(reconnectDelay * 2, maxDelay);
      connect();
    }, reconnectDelay);
  };

  ws.onopen = () => {
    reconnectDelay = 1000; // Reset delay on successful connection
  };
}
```

### Handling Disconnections

```javascript
ws.onclose = (event) => {
  if (event.wasClean) {
    console.log('Clean disconnect');
  } else {
    console.log('Connection died, reconnecting...');
    reconnect();
  }
};
```

## Rate Limits

- **No explicit rate limits** on message reception
- Messages are buffered server-side (2048 message buffer)
- If client cannot keep up, older messages may be dropped

## Error Handling

### Connection Errors

```javascript
ws.onerror = (error) => {
  console.error('WebSocket error:', error);
  // Implement retry logic
};
```

### Message Parsing Errors

```javascript
ws.onmessage = (event) => {
  try {
    const data = JSON.parse(event.data);
    handleMessage(data);
  } catch (error) {
    console.error('Failed to parse message:', error);
  }
};
```

## Complete Example - React Hook

```typescript
import { useEffect, useRef, useState } from 'react';

interface MarketTick {
  type: 'tick';
  symbol: string;
  bid: number;
  ask: number;
  spread: number;
  timestamp: number;
  lp: string;
}

export function useMarketData() {
  const [prices, setPrices] = useState<Map<string, MarketTick>>(new Map());
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const connect = () => {
      const ws = new WebSocket('ws://localhost:7999/ws');

      ws.onopen = () => {
        console.log('WebSocket connected');
        setConnected(true);
      };

      ws.onmessage = (event) => {
        try {
          const data: MarketTick = JSON.parse(event.data);
          if (data.type === 'tick') {
            setPrices((prev) => new Map(prev).set(data.symbol, data));
          }
        } catch (error) {
          console.error('Parse error:', error);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

      ws.onclose = () => {
        setConnected(false);
        console.log('Reconnecting in 3s...');
        setTimeout(connect, 3000);
      };

      wsRef.current = ws;
    };

    connect();

    return () => {
      wsRef.current?.close();
    };
  }, []);

  return { prices, connected };
}
```

## Complete Example - Go Client

```go
package main

import (
    "encoding/json"
    "log"
    "time"

    "github.com/gorilla/websocket"
)

type MarketTick struct {
    Type      string  `json:"type"`
    Symbol    string  `json:"symbol"`
    Bid       float64 `json:"bid"`
    Ask       float64 `json:"ask"`
    Spread    float64 `json:"spread"`
    Timestamp int64   `json:"timestamp"`
    LP        string  `json:"lp"`
}

func main() {
    url := "ws://localhost:7999/ws"

    for {
        conn, _, err := websocket.DefaultDialer.Dial(url, nil)
        if err != nil {
            log.Printf("Connection failed: %v, retrying in 5s...", err)
            time.Sleep(5 * time.Second)
            continue
        }

        log.Println("Connected to RTX Trading Engine")

        for {
            _, message, err := conn.ReadMessage()
            if err != nil {
                log.Printf("Read error: %v", err)
                conn.Close()
                break
            }

            var tick MarketTick
            if err := json.Unmarshal(message, &tick); err != nil {
                log.Printf("Parse error: %v", err)
                continue
            }

            if tick.Type == "tick" {
                log.Printf("%s: Bid=%.5f Ask=%.5f LP=%s",
                    tick.Symbol, tick.Bid, tick.Ask, tick.LP)
            }
        }

        log.Println("Reconnecting in 5s...")
        time.Sleep(5 * time.Second)
    }
}
```

## Performance Considerations

### Client-Side Buffering

Implement client-side buffering for high-frequency data:

```javascript
let tickBuffer = [];
const FLUSH_INTERVAL = 100; // ms

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  tickBuffer.push(data);
};

setInterval(() => {
  if (tickBuffer.length > 0) {
    processTicks(tickBuffer);
    tickBuffer = [];
  }
}, FLUSH_INTERVAL);
```

### Throttling Updates

For UI updates, throttle to prevent rendering bottlenecks:

```javascript
import { throttle } from 'lodash';

const updateUI = throttle((tick) => {
  // Update DOM/React state
  updatePriceDisplay(tick);
}, 100); // Max 10 updates per second

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  updateUI(data);
};
```

## Security

### Production Considerations

1. **Use WSS (TLS)** in production: `wss://api.rtxtrading.com/ws`
2. **Authentication:** Pass JWT token in query string or initial message
3. **Origin validation:** Server checks `Origin` header

**Future:** Token-based authentication

```javascript
const token = 'your-jwt-token';
const ws = new WebSocket(`wss://api.rtxtrading.com/ws?token=${token}`);
```

## Troubleshooting

### Connection Refused

**Problem:** `WebSocket connection failed: Connection refused`

**Solution:**
- Check if backend server is running on port 7999
- Verify firewall settings
- Ensure correct WebSocket URL

### No Messages Received

**Problem:** Connected but no price updates

**Solution:**
- Check if LP connections are active (`GET /admin/lp-status`)
- Verify symbols are enabled (`GET /api/config`)
- Check browser console for errors

### High Latency

**Problem:** Delayed price updates

**Solution:**
- Check network connection
- Monitor server load
- Implement client-side buffering
- Reduce number of active symbols

## Testing

### WebSocket Testing Tools

1. **Postman** - WebSocket client support
2. **wscat** - Command-line WebSocket client
3. **Browser DevTools** - Network → WS tab

### wscat Example

```bash
# Install
npm install -g wscat

# Connect
wscat -c ws://localhost:7999/ws

# You'll see real-time tick messages
```

## Summary

| Feature | Status | Description |
|---------|--------|-------------|
| Market Ticks | ✅ Live | Real-time price updates |
| Position Updates | 🔄 Planned | P/L and position changes |
| Order Notifications | 🔄 Planned | Order fills and rejections |
| Account Updates | 🔄 Planned | Balance changes |
| Symbol Subscription | 🔄 Planned | Filter specific symbols |
| Authentication | 🔄 Planned | JWT token auth |

**Current:** All connected clients receive all market ticks for enabled symbols.

**Future:** Symbol-specific subscriptions and authenticated channels.
