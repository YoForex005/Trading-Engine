# CandleEngine - Quick Start Guide

## Installation

The CandleEngine is part of the services package and is automatically available:

```typescript
import { CandleEngine, createCandleEngine, type OHLCData } from '../services';
```

## Basic Usage (30 seconds)

### Setup
```typescript
// Create engine for USDJPY 1-minute candles
const engine = new CandleEngine('USDJPY', '1m');

// Load last 500 historical candles
await engine.loadHistorical(500);
```

### Process Ticks
```typescript
const tick = {
  symbol: 'USDJPY',
  bid: 110.500,
  ask: 110.505,
  timestamp: Date.now(),
  volume: 100
};

const result = engine.processTick(tick);

// Check if new candle was formed
if (result.isNewCandle) {
  console.log('New candle formed!');
  const lastCandle = engine.getLastClosedCandle();
  console.log('Open:', lastCandle?.open, 'Close:', lastCandle?.close);
}
```

### Get Candles
```typescript
// All candles (including forming)
const allCandles = engine.getAllCandles();

// Just the closed ones
const historical = engine.getHistoricalCandles();

// Last 10 candles
const last10 = engine.getLastCandles(10);

// Candles in specific time range
const range = engine.getCandlesInRange(startTime, endTime);
```

## Common Patterns

### Multi-Timeframe Analysis
```typescript
const engines = {
  m1: await createCandleEngine('EURUSD', '1m'),
  m5: await createCandleEngine('EURUSD', '5m'),
  h1: await createCandleEngine('EURUSD', '1h')
};

// Process same tick through all
Object.values(engines).forEach(e => e.processTick(tick));

// Analyze each timeframe
const trend1m = analyzeTrend(engines.m1.getLastCandles(20));
const trend5m = analyzeTrend(engines.m5.getLastCandles(20));
const trend1h = analyzeTrend(engines.h1.getLastCandles(20));
```

### React Hook
```typescript
import { useState, useEffect } from 'react';

export function useCandles(symbol: string, timeframe: '1m' | '5m' | '15m' | '1h') {
  const [candles, setCandles] = useState<OHLCData[]>([]);
  const [engine, setEngine] = useState<CandleEngine | null>(null);

  useEffect(() => {
    createCandleEngine(symbol, timeframe, 500).then(e => {
      setEngine(e);
      setCandles(e.getAllCandles());
    });
  }, [symbol, timeframe]);

  const handleTick = (tick: TickData) => {
    if (!engine) return;
    const result = engine.processTick(tick);
    setCandles(result.candles);
    return result.isNewCandle;
  };

  return { candles, engine, handleTick };
}

// Usage in component
const { candles, handleTick } = useCandles('USDJPY', '1m');
```

### Timeframe Switching
```typescript
const engine = new CandleEngine('USDJPY', '1m');
await engine.loadHistorical(500);

// User switches timeframe
await engine.changeTimeframe('5m');

// Continue processing ticks normally
const result = engine.processTick(newTick);
```

### WebSocket Integration
```typescript
import { WebSocketService } from '../services/websocket';

const engine = await createCandleEngine('USDJPY', '1m');
const ws = new WebSocketService();

ws.on('tick', (tick: TickData) => {
  const result = engine.processTick(tick);

  // Update chart
  chartRef.current?.update(result.candles);

  // Handle new candle
  if (result.isNewCandle) {
    const closed = engine.getLastClosedCandle();
    updateIndicators(closed);
    checkTradingSignals(engine.getLastCandles(20));
  }
});
```

## API Reference (One Page)

### Constructor
```typescript
// Full form
new CandleEngine(symbol: string, timeframe?: Timeframe, maxHistoricalCandles?: number)

// Factory (auto-initializes)
await createCandleEngine(symbol: string, timeframe?: Timeframe, limit?: number)
```

### Loading Data
```typescript
await engine.loadHistorical(limit?: number)  // Default: 500
```

### Processing Ticks
```typescript
const result = engine.processTick(tick: TickData)
// result.candles: OHLCData[]
// result.isNewCandle: boolean
// result.candleIndex: number
```

### Querying Data
```typescript
engine.getAllCandles()           // All (historical + forming)
engine.getHistoricalCandles()    // Only closed
engine.getFormingCandle()        // Currently forming or null
engine.getLastClosedCandle()     // Most recent closed
engine.getCandleAt(index)        // By index
engine.getLastCandles(count)     // Last N
engine.getCandlesInRange(from, to) // Time range
```

### Management
```typescript
await engine.changeTimeframe(newTimeframe: Timeframe)
engine.reset()  // Clear all state
```

### Diagnostics
```typescript
engine.getStats()           // Statistics object
engine.getChartData()       // Chart-ready format
engine.isReady()            // Initialization status
engine.getSymbol()          // Current symbol
engine.getTimeframe()       // Current timeframe
```

## Data Types

### OHLCData
```typescript
{
  time: number;     // Candle time (milliseconds)
  open: number;     // Opening price
  high: number;     // Highest price
  low: number;      // Lowest price
  close: number;    // Closing price
  volume: number;   // Total volume
}
```

### TickData
```typescript
{
  symbol: string;
  bid: number;
  ask: number;
  timestamp: number;
  volume?: number;
}
```

### CandleProcessingResult
```typescript
{
  candles: OHLCData[];     // All current candles
  isNewCandle: boolean;    // True if new candle created
  candleIndex: number;     // Index of current/new candle
}
```

## Supported Timeframes

- `'1m'` - 1 minute
- `'5m'` - 5 minutes
- `'15m'` - 15 minutes
- `'30m'` - 30 minutes
- `'1h'` - 1 hour
- `'4h'` - 4 hours
- `'1d'` - 1 day

## Common Gotchas

### Must Initialize Before Processing
```typescript
const engine = new CandleEngine('USDJPY', '1m');

// WRONG: This will throw
engine.processTick(tick);

// RIGHT: Initialize first
await engine.loadHistorical();
engine.processTick(tick);  // Now works
```

### Immutable Returns
```typescript
const candle = engine.getLastClosedCandle();
candle.close = 99999;  // This is a copy, won't affect engine
```

### Time Bucket Precision
```typescript
// Ticks are bucketed by timeframe
// 1m: ticks between :00-:59 → same candle
// 5m: ticks between :00-:04, :05-:09, etc. → same candle
// A new candle forms when crossing bucket boundary
```

### Memory Management
```typescript
// Default: keeps last 1000 candles
// Customize:
const engine = new CandleEngine('USDJPY', '1m', 2000);  // Keep 2000

// Memory usage: ~40KB for 1000 candles (minimal)
```

## Performance

- **processTick()**: O(1) - Under 1ms
- **loadHistorical()**: O(n log n) - One-time load
- **changeTimeframe()**: O(m log m) - Manual operation
- **getAllCandles()**: O(m) - Fast

Can handle 1000+ ticks per second without issues.

## Troubleshooting

### No historical data
```typescript
// Check if API is working
const allSymbols = await historyDataManager.getAvailableSymbols();
console.log(allSymbols);  // Should have your symbol

// Try loading manually
try {
  await engine.loadHistorical();
} catch (error) {
  console.error('Load failed:', error);
}

// Engine still initialized but with empty history
// Can still process live ticks
```

### Candles seem wrong
```typescript
// Check stats
const stats = engine.getStats();
console.log(stats);

// Verify last few candles
const last5 = engine.getLastCandles(5);
last5.forEach(c => {
  console.log(`Time: ${c.time}, O: ${c.open}, C: ${c.close}`);
});
```

### Out of memory
```typescript
// Reduce maxHistoricalCandles
const engine = new CandleEngine('USDJPY', '1m', 100);  // Keep only 100

// Or clear periodically
engine.reset();
await engine.loadHistorical(100);
```

## Next Steps

1. **Import** the service in your component
2. **Create** an engine instance for your symbol
3. **Load** historical data
4. **Connect** WebSocket tick events
5. **Process** ticks with `processTick()`
6. **Display** with `getChartData()`

That's it! The engine handles all the candle logic.

---

**Need more help?** Check the full documentation or examples in the code comments.
