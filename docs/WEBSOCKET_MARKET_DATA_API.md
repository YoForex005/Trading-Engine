# WebSocket Market Data API Documentation

**Version**: 1.1
**Last Updated**: January 30, 2026
**Status**: Production Ready

---

## Overview

The Trading Engine provides real-time market data via WebSocket connection. All data comes from the YOFX FIX gateway with 100% real market data (no simulation).

---

## Connection

### URL
```
ws://localhost:7999/ws?token=YOUR_JWT_TOKEN
```

### Headers
```
Authorization: Bearer YOUR_JWT_TOKEN
```

### Authentication
- JWT token required in query parameter or header
- Token obtained from `/login` endpoint
- Token expires based on `JWT_EXPIRY` configuration (default: 24 hours)

### Connection Example (JavaScript)
```javascript
const token = localStorage.getItem('authToken');
const ws = new WebSocket(`ws://localhost:7999/ws?token=${token}`);

ws.onopen = () => {
    console.log('Connected to market data');
};

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log('Market tick received:', data);
};

ws.onerror = (error) => {
    console.error('WebSocket error:', error);
};

ws.onclose = () => {
    console.log('Disconnected from market data');
    // Implement reconnection logic
};
```

---

## Message Format

### Tick Message Structure

```json
{
  "type": "tick",
  "symbol": "EURUSD",
  "bid": 1.1045156847,
  "ask": 1.1046156847,
  "spread": 0.0001,
  "timestamp": 1769772266771,
  "lp": "YOFX",
  "dailyChange": 0.0234,
  "high24h": 1.1051234567,
  "low24h": 1.1032456789
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | Always `"tick"` for market data messages |
| `symbol` | string | Yes | Currency pair symbol (e.g., `"EURUSD"`, `"GBPUSD"`) |
| `bid` | float64 | Yes | Current bid price (5-8 decimal places) |
| `ask` | float64 | Yes | Current ask price (5-8 decimal places) |
| `spread` | float64 | Yes | Bid-ask spread = ask - bid |
| `timestamp` | int64 | Yes | Unix milliseconds when quote was generated |
| `lp` | string | Yes | Liquidity provider source. Always `"YOFX"` |
| `dailyChange` | float64 | Yes | Daily percentage change from open (e.g., 0.0234 = 0.0234%) |
| `high24h` | float64 | Yes | 24-hour high price from YOFX feed |
| `low24h` | float64 | Yes | 24-hour low price from YOFX feed |

---

## Data Characteristics

### Update Frequency
- **Standard Mode**: 60-80% reduced broadcasts (throttling enabled)
- **MT5 Mode** (`MT5_MODE=true`): All ticks broadcast (no throttling)
- **Typical Rate**: 5-10 ticks/second per symbol

### Pricing Precision
- **Bid/Ask**: Configured for 5-8 decimal places depending on symbol
- **Spread**: Calculated as `ask - bid`
- **Example**: `1.1045156847` (10 decimal places)

### Timestamp
- **Format**: Unix milliseconds (not seconds)
- **Example**: `1769772266771` represents January 30, 2026
- **Conversion**:
  ```javascript
  const date = new Date(timestamp);
  // 2026-01-30T12:34:26.771Z
  ```

### Daily Change
- **Format**: Percentage (e.g., 0.0234 = 0.0234%)
- **Calculation**: `(high24h - open) / open * 100`
- **Range**: -100 to +100
- **Use Case**: Display daily performance in UI

### High/Low Tracking
- **Source**: YOFX FIX gateway market data
- **Update Frequency**: Updated with each market data refresh
- **Purpose**: Technical analysis, support/resistance levels
- **Persistence**: Stored in tick history database

---

## Usage Examples

### Example 1: Display Real-Time Quotes

```javascript
const ws = new WebSocket(`ws://localhost:7999/ws?token=${token}`);

const prices = {};

ws.onmessage = (event) => {
    const tick = JSON.parse(event.data);

    if (tick.type === 'tick') {
        prices[tick.symbol] = {
            bid: tick.bid,
            ask: tick.ask,
            spread: tick.spread,
            timestamp: new Date(tick.timestamp),
        };

        // Update UI with new price
        updatePriceDisplay(tick.symbol, tick.bid, tick.ask);
    }
};
```

### Example 2: Calculate Daily Change

```javascript
ws.onmessage = (event) => {
    const tick = JSON.parse(event.data);

    if (tick.type === 'tick') {
        // Display daily change with color
        const changePercent = tick.dailyChange.toFixed(2);
        const color = tick.dailyChange >= 0 ? 'green' : 'red';

        console.log(`${tick.symbol}: ${changePercent}% (${color})`);

        // Display high/low range
        const range = tick.high24h - tick.low24h;
        console.log(`Range: ${tick.low24h} - ${tick.high24h} (${range.toFixed(5)})`);
    }
};
```

### Example 3: Technical Analysis with High/Low

```javascript
class TechnicalAnalyzer {
    constructor(symbol) {
        this.symbol = symbol;
        this.prices = [];
        this.highLow = null;
    }

    processTick(tick) {
        if (tick.symbol !== this.symbol) return;

        // Store price for analysis
        this.prices.push({
            bid: tick.bid,
            ask: tick.ask,
            time: new Date(tick.timestamp),
        });

        // Update high/low
        this.highLow = {
            high24h: tick.high24h,
            low24h: tick.low24h,
            range: tick.high24h - tick.low24h,
            midpoint: (tick.high24h + tick.low24h) / 2,
        };

        // Calculate support/resistance
        this.calculateLevels();
    }

    calculateLevels() {
        const mid = this.highLow.midpoint;
        const range = this.highLow.range;

        return {
            resistance1: mid + (range * 0.25),
            support1: mid - (range * 0.25),
            resistance2: this.highLow.high24h,
            support2: this.highLow.low24h,
        };
    }
}
```

### Example 4: Zustand Store Integration

```typescript
import { create } from 'zustand';

interface Tick {
    symbol: string;
    bid: number;
    ask: number;
    spread: number;
    timestamp: number;
    lp: string;
    dailyChange: number;
    high24h: number;
    low24h: number;
}

interface AppStore {
    ticks: Record<string, Tick>;
    setTick: (symbol: string, tick: Tick) => void;
    getTick: (symbol: string) => Tick | undefined;
}

export const useAppStore = create<AppStore>((set, get) => ({
    ticks: {},

    setTick: (symbol: string, tick: Tick) => {
        set((state) => ({
            ticks: {
                ...state.ticks,
                [symbol]: tick,
            },
        }));
    },

    getTick: (symbol: string) => {
        return get().ticks[symbol];
    },
}));

// Usage in component
function MarketWatch() {
    const { ticks } = useAppStore();

    return (
        <div>
            {Object.entries(ticks).map(([symbol, tick]) => (
                <div key={symbol}>
                    <span>{symbol}</span>
                    <span>Bid: {tick.bid.toFixed(5)}</span>
                    <span>Ask: {tick.ask.toFixed(5)}</span>
                    <span>Change: {tick.dailyChange.toFixed(2)}%</span>
                    <span>High: {tick.high24h.toFixed(5)}</span>
                    <span>Low: {tick.low24h.toFixed(5)}</span>
                </div>
            ))}
        </div>
    );
}
```

---

## Error Handling

### Connection Errors

```javascript
ws.onerror = (error) => {
    console.error('WebSocket error:', error);
};

ws.onclose = (event) => {
    if (event.code === 1008 || event.reason === 'Unauthorized') {
        // Authentication failed
        console.error('Auth failed - redirect to login');
        window.location.href = '/login';
        return;
    }

    if (event.code !== 1000) {
        // Unexpected closure - reconnect
        console.log('Reconnecting in 2 seconds...');
        setTimeout(reconnect, 2000);
    }
};
```

### Message Validation

```javascript
ws.onmessage = (event) => {
    let tick;
    try {
        tick = JSON.parse(event.data);
    } catch (e) {
        console.error('Invalid JSON:', e);
        return;
    }

    // Validate required fields
    if (!tick.symbol || typeof tick.bid !== 'number' || typeof tick.ask !== 'number') {
        console.error('Invalid tick structure:', tick);
        return;
    }

    // Sanity check: spread should be positive
    if (tick.ask < tick.bid) {
        console.error('Invalid spread (ask < bid):', tick);
        return;
    }

    // Process valid tick
    processTick(tick);
};
```

---

## Performance Optimization

### Throttling Settings

#### Standard Mode (Default)
```bash
# Environment
ENVIRONMENT=production

# Result: 60-80% reduction in broadcast frequency
# Network: ~50-100 Kbps per client
# CPU: Minimal
# Latency: <20ms
```

#### MT5 Mode (Professional Trading)
```bash
# Environment
MT5_MODE=true

# Result: 100% tick delivery (no throttling)
# Network: ~500-800 Kbps per client
# CPU: Higher
# Latency: <10ms
# Use Case: Professional terminals, algo trading, backtesting
```

### Client-Side Optimization

```javascript
// Debounce updates to prevent excessive re-renders
import { debounce } from 'lodash-es';

const updateUI = debounce((tick) => {
    // Update React state/Zustand
    useAppStore.setState((state) => ({
        ticks: {
            ...state.ticks,
            [tick.symbol]: tick,
        },
    }));
}, 50);  // Max 20 updates per second

ws.onmessage = (event) => {
    const tick = JSON.parse(event.data);
    updateUI(tick);
};
```

---

## Monitoring & Diagnostics

### Real-Time Stats Endpoint

```bash
GET /admin/fix/ticks
```

**Response**:
```json
{
  "totalTickCount": 523641,
  "symbolCount": 29,
  "ticksPerSecond": 145.3,
  "latestTicks": {
    "EURUSD": {
      "bid": 1.1045156847,
      "ask": 1.1046156847,
      "high24h": 1.1051234567,
      "low24h": 1.1032456789,
      "timestamp": 1769772266771,
      "lp": "YOFX"
    }
  }
}
```

### Health Check Endpoint

```bash
GET /health
```

**Response**:
```json
{
  "status": "ok",
  "checks": {
    "fix_yofx2": "connected",
    "websocket": "ok",
    "latency_ms": 8.5
  }
}
```

---

## Troubleshooting

### Issue: "401 Unauthorized"

**Cause**: Invalid or expired JWT token

**Solution**:
```javascript
// Get new token
const response = await fetch('/login', {
    method: 'POST',
    body: JSON.stringify({ username: 'user', password: 'pass' }),
});
const { token } = await response.json();
localStorage.setItem('authToken', token);

// Reconnect with new token
ws.close();
connectWebSocket(token);
```

### Issue: No ticks received

**Cause**:
1. YOFX FIX connection down
2. Symbol not subscribed
3. Network issue

**Solution**:
```bash
# Check FIX status
curl http://localhost:7999/admin/fix/status

# Check subscribed symbols
curl http://localhost:7999/api/symbols/subscribed

# Check market data
curl http://localhost:7999/admin/fix/ticks

# Check network
ping -c 5 $YOFX_PROXY_HOST
```

### Issue: High latency or dropped ticks

**Cause**:
1. Network congestion
2. Client buffer full
3. Server overload

**Solution**:
```javascript
// Monitor latency
let lastTickTime = Date.now();
ws.onmessage = (event) => {
    const tick = JSON.parse(event.data);
    const latency = Date.now() - tick.timestamp;
    console.log(`Latency: ${latency}ms`);
    lastTickTime = Date.now();
};

// Drop old ticks if buffer full
const maxTicksPerSymbol = 1000;
const tickCount = {};

ws.onmessage = (event) => {
    const tick = JSON.parse(event.data);

    tickCount[tick.symbol] = (tickCount[tick.symbol] || 0) + 1;

    if (tickCount[tick.symbol] > maxTicksPerSymbol) {
        console.warn(`Dropping old tick for ${tick.symbol}`);
        return;
    }

    processTick(tick);
};
```

---

## Security

### Token Management

```javascript
// Store token securely
// ✅ DO: Use httpOnly cookies (set by server)
// ❌ DON'T: Store in localStorage (vulnerable to XSS)

// Cookie-based (recommended)
const ws = new WebSocket('ws://localhost:7999/ws');
// Cookie automatically sent by browser

// Header-based (if needed)
const token = getCookie('authToken');
ws = new WebSocket(`ws://localhost:7999/ws?token=${token}`);
```

### Data Validation

```javascript
// Validate all received data
function validateTick(tick) {
    const errors = [];

    if (typeof tick.symbol !== 'string') {
        errors.push('Invalid symbol');
    }

    if (typeof tick.bid !== 'number' || tick.bid <= 0) {
        errors.push('Invalid bid');
    }

    if (typeof tick.ask !== 'number' || tick.ask <= 0) {
        errors.push('Invalid ask');
    }

    if (tick.ask < tick.bid) {
        errors.push('Ask < Bid (invalid spread)');
    }

    if (typeof tick.timestamp !== 'number' || tick.timestamp < 0) {
        errors.push('Invalid timestamp');
    }

    if (tick.lp !== 'YOFX') {
        errors.push(`Invalid LP: ${tick.lp} (expected YOFX)`);
    }

    return {
        valid: errors.length === 0,
        errors,
    };
}

ws.onmessage = (event) => {
    const tick = JSON.parse(event.data);
    const validation = validateTick(tick);

    if (!validation.valid) {
        console.error('Invalid tick:', validation.errors);
        return;
    }

    processTick(tick);
};
```

---

## Backward Compatibility

### Legacy Fields

If you're upgrading from an older version, note the following:

**New Required Fields** (as of Jan 30, 2026):
- `dailyChange` (float64)
- `high24h` (float64)
- `low24h` (float64)

**Deprecated Fields**:
- None (all existing fields preserved)

**Migration**:
```javascript
// Old code (still works)
const spread = tick.ask - tick.bid;

// New code (recommended)
const spread = tick.spread;
const change = tick.dailyChange;
const range = tick.high24h - tick.low24h;
```

---

## API Endpoints for Market Data

### Get Latest Price for Symbol

```bash
GET /api/prices/:symbol
```

**Response**:
```json
{
  "symbol": "EURUSD",
  "bid": 1.1045156847,
  "ask": 1.1046156847,
  "spread": 0.0001,
  "timestamp": 1769772266771,
  "lp": "YOFX",
  "high24h": 1.1051234567,
  "low24h": 1.1032456789
}
```

### Get Subscribed Symbols

```bash
GET /api/symbols/subscribed
```

**Response**:
```json
{
  "symbols": [
    "EURUSD",
    "GBPUSD",
    "USDJPY",
    // ... 29 total
  ],
  "count": 29
}
```

### Get All Available Symbols

```bash
GET /api/symbols/all
```

**Response**:
```json
{
  "symbols": [
    {
      "symbol": "EURUSD",
      "name": "Euro vs US Dollar",
      "enabled": true,
      "category": "Major"
    },
    // ... more symbols
  ],
  "total": 50
}
```

---

## Rate Limits

- **WebSocket Connections**: Unlimited
- **Historical Data Requests**: 1000 requests/hour
- **Administrative Endpoints**: 100 requests/hour
- **Per Symbol**: Up to 100 ticks/second storage

---

## Support

- **Documentation**: https://docs.example.com
- **Issues**: Please report via admin panel
- **Contact**: support@trading-engine.com

---

**Document Version**: 1.1
**Last Updated**: January 30, 2026
**Status**: Production Ready
