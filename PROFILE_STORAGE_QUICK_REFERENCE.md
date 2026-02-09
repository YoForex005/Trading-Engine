# Profile & Storage Data - Quick Reference

## All localStorage Keys (A-Z)

```
chart-indicators-{SYMBOL}      Technical indicator configurations per symbol
dockHeight                      Bottom dock panel height (pixels as string)
drawings-{SYMBOL}              Chart drawing objects per symbol
rtx5_hidden_symbols            Array of hidden symbols
rtx5_marketwatch_cols          Array of visible column IDs
rtx5_marketwatch_lp_filter     Boolean as string (show only real data)
rtx5_marketwatch_options       System options object
rtx5_subscribed_symbols        Array of currently subscribed symbols
rtx5_symbol_sets               Custom symbol groupings (workspace-like)
rtx_token                       JWT authentication token
trading-app-storage            Main Zustand store (Zustand persist)
windowManagerState             Multi-chart layout configuration
MarketDataStore                Market data store (live, not persisted)
```

**Total: 13 different key patterns + 1 per symbol for indicators/drawings**

---

## Where Persisted Data is Used

### On Application Startup

```typescript
// 1. App.tsx
const [dockHeight, setDockHeight] = useState(() => {
  const saved = localStorage.getItem('dockHeight');
  return saved ? parseInt(saved, 10) : 250;
});

// 2. Zustand hydration (automatic via persist middleware)
useAppStore → reads 'trading-app-storage'

// 3. WindowManager constructor
windowManager.loadState() → reads 'windowManagerState'

// 4. MarketWatchPanel useEffect
const saved = localStorage.getItem('rtx5_marketwatch_cols')
const saved = localStorage.getItem('rtx5_hidden_symbols')
const saved = localStorage.getItem('rtx5_marketwatch_options')
const saved = localStorage.getItem('rtx5_subscribed_symbols')

// 5. TradingChart component
indicatorManager.loadFromStorage(symbol)
drawingManager.loadFromStorage(symbol)

// 6. WebSocket connection
const token = localStorage.getItem('rtx_token')
```

### During User Interaction

```typescript
// Changing chart preferences
useAppStore.setState({ chartType, timeframe, orderVolume })
// AUTO-PERSISTED to 'trading-app-storage'

// Adding indicator
indicatorManager.addIndicator(config)
// Manual: indicatorManager.saveToStorage(symbol)

// Adding drawing
drawingManager.addDrawing(drawing)
// Manual: drawingManager.saveToStorage(symbol)

// Changing layout
windowManager.setLayoutMode('grid')
// AUTO-PERSISTED to 'windowManagerState'

// Hiding symbol
hideSymbol('EURUSD')
// Updated 'rtx5_hidden_symbols'

// Changing columns
saveColumnConfig(['symbol', 'bid', 'ask'])
// Updated 'rtx5_marketwatch_cols'
```

---

## File Paths (Absolute)

**Main Storage Files**:
- `C:\Users\s s laptop bazar\Trading-Engine2\clients\desktop\src\store\useAppStore.ts`
- `C:\Users\s s laptop bazar\Trading-Engine2\clients\desktop\src\store\useMarketDataStore.ts`
- `C:\Users\s s laptop bazar\Trading-Engine2\clients\desktop\src\services\windowManager.ts`
- `C:\Users\s s laptop bazar\Trading-Engine2\clients\desktop\src\services\indicatorManager.ts`
- `C:\Users\s s laptop bazar\Trading-Engine2\clients\desktop\src\services\drawingManager.ts`
- `C:\Users\s s laptop bazar\Trading-Engine2\clients\desktop\src\services\marketWatchActions.ts`
- `C:\Users\s s laptop bazar\Trading-Engine2\clients\desktop\src\components\layout\MarketWatchPanel.tsx`
- `C:\Users\s s laptop bazar\Trading-Engine2\clients\desktop\src\App.tsx`

---

## State Management Architecture

```
┌─────────────────────────┐
│ Zustand Stores          │
├─────────────────────────┤
│ useAppStore             │  ← Trading state (partially persisted)
│ useMarketDataStore      │  ← Market data (live, not persisted)
└─────────────────────────┘
         ↕
┌─────────────────────────┐
│ Singleton Services      │
├─────────────────────────┤
│ windowManager           │  ← Manual persist (windowManagerState)
│ indicatorManager        │  ← Manual persist (chart-indicators-*)
│ drawingManager          │  ← Manual persist (drawings-*)
└─────────────────────────┘
         ↕
┌─────────────────────────┐
│ localStorage Keys       │
├─────────────────────────┤
│ Direct persistence      │  ← App.tsx (dockHeight)
│ Function-based          │  ← marketWatchActions (hide, columns, etc)
└─────────────────────────┘
```

---

## Data Types Reference

### Tick (Market Data)
```typescript
{
  symbol: string;
  bid: number;
  ask: number;
  spread: number;
  timestamp: number;
  lp?: string;
  prevBid?: number;
  prevAsk?: number;
  dailyChange?: number;
  high?: number;
  low?: number;
  high24h?: number;
  low24h?: number;
  volume?: number;
}
```

### Position (Open Trade)
```typescript
{
  id: number;
  symbol: string;
  side: 'BUY' | 'SELL';
  volume: number;
  openPrice: number;
  currentPrice: number;
  openTime: string;
  sl: number;
  tp: number;
  swap: number;
  commission: number;
  unrealizedPnL: number;
}
```

### IndicatorConfig (Chart Indicator)
```typescript
{
  id: string;
  name: string;
  type: 'sma' | 'ema' | 'rsi' | 'macd';
  parameters: Record<string, any>;  // { period: 20 }
  visible: boolean;
  color?: string;
}
```

### Drawing (Chart Annotation)
```typescript
{
  id: string;
  type: DrawingTool;  // 'TREND_LINE', 'HORIZONTAL_LINE', etc
  points: Array<{ x: number; y: number }>;
  style: {
    color: string;
    lineWidth: number;
    lineStyle: 'solid' | 'dashed' | 'dotted';
  };
  locked?: boolean;
  visible?: boolean;
}
```

### ChartWindow (Multi-Chart Layout)
```typescript
{
  id: string;
  symbol: string;
  position?: { row: number; col: number };
}
```

### SystemOptions (Market Watch Settings)
```typescript
{
  useSystemColors: boolean;
  showMilliseconds: boolean;
  autoRemoveExpired: boolean;
  autoArrange: boolean;
  showGrid: boolean;
}
```

---

## Key Operations

### Reading Profile Data

```typescript
// Zustand
const state = useAppStore.getState();
console.log(state.selectedSymbol);  // EURUSD
console.log(state.chartType);       // candlestick
console.log(state.timeframe);       // 1m
console.log(state.orderVolume);     // 0.01

// Services
const layoutMode = windowManager.getLayoutMode();
const charts = windowManager.getCharts();
const indicators = indicatorManager.getIndicators();
const drawings = drawingManager.getDrawings();

// Market Watch
const hidden = getHiddenSymbols();
const sets = getSymbolSets();
const cols = loadColumnConfig();
const opts = loadSystemOptions();
```

### Writing Profile Data

```typescript
// Zustand (auto-persisted)
useAppStore.setState({
  selectedSymbol: 'GBPUSD',
  chartType: 'line',
  timeframe: '5m',
  orderVolume: 0.1
});

// Services (manual persistence)
windowManager.setLayoutMode('grid');
windowManager.addChart('USDJPY');
// Changes saved to localStorage automatically

indicatorManager.addIndicator({
  id: 'ind-1',
  name: 'SMA 20',
  type: 'sma',
  parameters: { period: 20 },
  visible: true
});
indicatorManager.saveToStorage('EURUSD');  // ← Manual save required

// Market Watch
hideSymbol('EURUSD');
saveSymbolSet('My Favorites', ['GBPUSD', 'USDJPY']);
saveColumnConfig(['symbol', 'bid', 'ask', 'spread']);
saveSystemOptions({ useSystemColors: true, showGrid: true });
```

### Clearing Profile Data

```typescript
// Reset all trading state
useAppStore.getState().reset();

// Clear layout
windowManager.reset();
localStorage.removeItem('windowManagerState');

// Clear indicators
indicatorManager.clearAllIndicators();
indicatorManager.saveToStorage('EURUSD');

// Clear drawings
drawingManager.clear();
drawingManager.saveToStorage('EURUSD');

// Clear market watch
localStorage.removeItem('rtx5_hidden_symbols');
localStorage.removeItem('rtx5_symbol_sets');
localStorage.removeItem('rtx5_marketwatch_cols');
localStorage.removeItem('rtx5_marketwatch_options');
```

---

## Keyboard Shortcuts (Hardcoded)

| Key | Action |
|-----|--------|
| F9 | New Order Dialog |
| F10 | Chart Window |
| Alt+B | Quick Buy |
| Alt+S | Quick Sell |
| Ctrl+U | Symbols Dialog |
| Ctrl+Shift+S | Save Workspace Layout (not implemented) |
| Ctrl+Shift+L | Load Workspace Layout (not implemented) |

---

## Custom Events (CommandBusContext)

```typescript
// Open new chart
window.dispatchEvent(new CustomEvent('openChart', {
  detail: { symbol: 'GBPUSD', timeframe: '1H', type: 'candlestick' }
}));

// Open order dialog
window.dispatchEvent(new CustomEvent('openOrderDialog', {
  detail: { symbol: 'EURUSD', type: 'MARKET' }
}));

// Open depth of market
window.dispatchEvent(new CustomEvent('openDepthOfMarket', {
  detail: { symbol: 'EURUSD' }
}));

// Open symbols management
window.dispatchEvent(new CustomEvent('openSymbolsDialog'));

// Open popup prices
window.dispatchEvent(new CustomEvent('openPopupPrices', {
  detail: { symbol: 'EURUSD' }
}));
```

---

## Common Issues & Solutions

### Indicator not persisting
```typescript
// Problem: Indicator added but not saved
indicatorManager.addIndicator(config);
// Missing: indicatorManager.saveToStorage(symbol);

// Solution:
indicatorManager.addIndicator(config);
indicatorManager.saveToStorage(symbol);  // ← Add this
```

### Drawing not showing on reload
```typescript
// Problem: Drawing added but not saved
drawingManager.addDrawing(drawing);
// Missing: drawingManager.saveToStorage(symbol);

// Solution:
drawingManager.addDrawing(drawing);
drawingManager.saveToStorage(symbol);  // ← Add this
```

### Layout changes not persisting
```typescript
// windowManager auto-persists, so this should work:
windowManager.setLayoutMode('grid');  // ← Auto-saved

// If not working, check:
localStorage.getItem('windowManagerState');
```

### Chart preferences reset on reload
```typescript
// Problem: selectedSymbol, chartType, timeframe, orderVolume not saving
// Solution: They ARE saved by Zustand (trading-app-storage)
// Check: localStorage.getItem('trading-app-storage');
```

---

## Browser DevTools Tips

### View all stored data
```javascript
// In console:
Object.keys(localStorage).forEach(key => {
  console.log(key, JSON.parse(localStorage.getItem(key)));
});
```

### Monitor specific key changes
```javascript
const original = localStorage.setItem;
localStorage.setItem = function(key, value) {
  if (key === 'trading-app-storage') {
    console.log('AppStore changed:', JSON.parse(value));
  }
  original.apply(this, arguments);
};
```

### Clear all trading data
```javascript
localStorage.clear();  // ⚠️ Clears everything
location.reload();
```

### Export current profile as JSON
```javascript
const profile = {
  appStore: JSON.parse(localStorage.getItem('trading-app-storage')),
  windowManager: JSON.parse(localStorage.getItem('windowManagerState')),
  marketWatch: {
    cols: JSON.parse(localStorage.getItem('rtx5_marketwatch_cols')),
    hidden: JSON.parse(localStorage.getItem('rtx5_hidden_symbols')),
    sets: JSON.parse(localStorage.getItem('rtx5_symbol_sets')),
    options: JSON.parse(localStorage.getItem('rtx5_marketwatch_options'))
  }
};
console.log(JSON.stringify(profile, null, 2));
```

---

## Migration Path (If Implementing Profiles)

```typescript
// Current structure (single profile):
localStorage: {
  'trading-app-storage': {...},
  'windowManagerState': {...},
  'chart-indicators-EURUSD': [...],
  ...
}

// Proposed structure (multi-profile):
localStorage: {
  'profile:default': {
    appStore: {...},
    windowManager: {...},
    indicators: { EURUSD: [...], GBPUSD: [...] },
    drawings: { EURUSD: [...] },
    marketWatch: {...}
  },
  'profile:scalping': {...},
  'profile:swing': {...},
  'profile:active': 'default'
}
```

---

## Zustand Store Methods

### useAppStore Actions

```typescript
setAuthenticated(isAuth, accountId?, authToken?)
setAuthToken(token)
clearAuth()
setTick(symbol, tick)
setTicks(ticks)
setSelectedSymbol(symbol)
setPositions(positions)
setOrders(orders)
setTrades(trades)
setAccount(account)
setChartMaximized(isMaximized)
setChartType(type)
setTimeframe(tf)
setOrderVolume(volume)
setWsConnected(connected)
setLoadingStates(states)
reset()
```

### useAppStore Selectors

```typescript
const isAuthenticated = useAppStore(s => s.isAuthenticated);
const ticks = useAppStore(s => s.ticks);
const selectedSymbol = useAppStore(s => s.selectedSymbol);
const tick = useAppStore(s => s.ticks[symbol]);
```

### useMarketDataStore Actions

```typescript
updateTick(symbol, tick)
updateCandle(update)
updateBulkTicks(ticks)
subscribeSymbol(symbol)
unsubscribeSymbol(symbol)
aggregateOHLCV(symbol)
clearSymbolData(symbol)
clearAllData()
```

### useMarketDataStore Selectors

```typescript
useCurrentTick(symbol)              // Hook
useSymbolStats(symbol)              // Hook
useOHLCV(symbol, timeframe)         // Hook
useSubscribedSymbols()              // Hook
useRecentOHLCV(symbol)              // Hook
```

---

## Performance Notes

- **Zustand stores update instantly** (no debouncing)
- **localStorage.setItem() is synchronous** (~1-10ms per write)
- **OHLCV aggregation offloaded to Web Worker** (non-blocking)
- **Market data store limited**: 1000 (1m), 500 (5m), 300 (15m), 200 (1h) candles
- **Tick buffer limited**: 10,000 ticks per symbol maximum
- **Large drawing count degrades performance**: >1000 drawings per symbol

---

## Common Code Patterns

### Auto-persist on state change
```typescript
const [value, setValue] = useState(() => {
  const saved = localStorage.getItem('key');
  return saved ? JSON.parse(saved) : defaultValue;
});

useEffect(() => {
  localStorage.setItem('key', JSON.stringify(value));
}, [value]);
```

### Symbol-specific data
```typescript
const key = `chart-indicators-${symbol}`;
localStorage.setItem(key, JSON.stringify(data));
const data = localStorage.getItem(key);
```

### Service singleton pattern
```typescript
class Service { ... }
export const service = new Service();  // Single instance
```

### Event-driven updates
```typescript
function hideSymbol(symbol: string) {
  const stored = localStorage.getItem('rtx5_hidden_symbols');
  const hidden = stored ? JSON.parse(stored) : [];
  const updated = [...hidden, symbol];
  localStorage.setItem('rtx5_hidden_symbols', JSON.stringify(updated));
  return updated;
}

// Component detects change via useEffect
const [hidden, setHidden] = useState(() => getHiddenSymbols());
useEffect(() => {
  const updated = getHiddenSymbols();
  setHidden(updated);
}, [/* needs proper dependency tracking */]);
```

---

## Testing Tips

### Test profile persistence
```typescript
it('should persist appStore on state change', () => {
  useAppStore.setState({ selectedSymbol: 'GBPUSD' });
  const stored = JSON.parse(localStorage.getItem('trading-app-storage'));
  expect(stored.selectedSymbol).toBe('GBPUSD');
});

it('should restore appStore on mount', () => {
  localStorage.setItem('trading-app-storage',
    JSON.stringify({ selectedSymbol: 'GBPUSD' }));
  // Remount component
  expect(useAppStore.getState().selectedSymbol).toBe('GBPUSD');
});
```

### Test service persistence
```typescript
it('should save indicators to storage', () => {
  indicatorManager.addIndicator(config);
  indicatorManager.saveToStorage('EURUSD');
  const stored = JSON.parse(localStorage.getItem('chart-indicators-EURUSD'));
  expect(stored).toContainEqual(config);
});
```

