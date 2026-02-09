# Profile Data Architecture & State Management

## Executive Summary

The Trading Engine desktop application uses a **multi-tier state management architecture** combining Zustand stores, localStorage persistence, singleton service managers, and React Context. Profile data is distributed across multiple storage systems with no centralized profile configuration file or dedicated profile switching mechanism.

---

## 1. State Management Overview

### 1.1 Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│            React Components                             │
├─────────────────────────────────────────────────────────┤
│  Zustand Stores (Persistent)                            │
│  ├─ useAppStore (trading-app-storage)                  │
│  ├─ useMarketDataStore (MarketDataStore)               │
│  └─ Custom Hooks (useCurrentTick, useSymbolStats)       │
├─────────────────────────────────────────────────────────┤
│  Singleton Services & Managers                          │
│  ├─ windowManager (localStorage: windowManagerState)    │
│  ├─ indicatorManager (localStorage: chart-indicators-*) │
│  ├─ drawingManager (localStorage: drawings-*)           │
│  └─ WebSocket Services                                  │
├─────────────────────────────────────────────────────────┤
│  Browser localStorage                                   │
│  └─ RTX5-specific keys (rtx5_*, rtx_*)                 │
└─────────────────────────────────────────────────────────┘
```

---

## 2. Zustand Global State Stores

### 2.1 useAppStore (Main Trading State)

**File**: `clients/desktop/src/store/useAppStore.ts`

**Persistence Configuration**:
- Storage Key: `trading-app-storage`
- Middleware: `persist` + `devtools`
- Only persists selected fields (partialize)

**State Structure**:
```typescript
interface AppState {
  // Authentication
  isAuthenticated: boolean;
  accountId: string | null;
  authToken: string | null;

  // Market Data (NOT persisted)
  ticks: Record<string, Tick>;
  selectedSymbol: string;

  // Trading
  positions: Position[];
  orders: Order[];
  trades: Trade[];

  // Account (NOT persisted)
  account: Account | null;

  // UI State
  isChartMaximized: boolean;
  chartType: 'candlestick' | 'line' | 'area';
  timeframe: '1m' | '5m' | '15m' | '1h' | '4h' | '1d';
  orderVolume: number;

  // WebSocket
  wsConnected: boolean;

  // Loading States
  isLoadingAccount: boolean;
  isLoadingPositions: boolean;
  isPlacingOrder: boolean;
}
```

**Persisted Fields** (via partialize):
```typescript
{
  selectedSymbol: state.selectedSymbol,
  chartType: state.chartType,
  timeframe: state.timeframe,
  orderVolume: state.orderVolume,
}
```

**Actions Available**:
- `setAuthenticated(isAuth, accountId, authToken)`
- `setAuthToken(token)`
- `clearAuth()`
- `setTick(symbol, tick)`
- `setSelectedSymbol(symbol)`
- `setChartType(type)`
- `setTimeframe(tf)`
- `setOrderVolume(volume)`
- `setWsConnected(connected)`
- `setLoadingStates(states)`
- `reset()`

---

### 2.2 useMarketDataStore (Real-time Market Data)

**File**: `clients/desktop/src/store/useMarketDataStore.ts`

**Persistence Configuration**:
- Storage Key: `MarketDataStore`
- Middleware: `subscribeWithSelector` + `devtools`
- NOT persisted by default (live data only)

**State Structure**:
```typescript
interface MarketDataState {
  // Symbol data organized by symbol
  symbolData: Record<string, SymbolData>;

  // Subscription tracking
  subscribedSymbols: Set<string>;
}

interface SymbolData {
  currentTick: Tick | null;
  previousTick: Tick | null;
  stats: SymbolStats | null;
  ohlcv1m: OHLCV[];
  ohlcv5m: OHLCV[];
  ohlcv15m: OHLCV[];
  ohlcv1h: OHLCV[];
  tickBuffer: Tick[];
  lastAggregation: number;
}
```

**Performance Optimization**:
- OHLCV aggregation offloaded to Web Worker
- Tick buffer limited to 10,000 ticks per symbol
- OHLCV data limited to: 1000 (1m), 500 (5m), 300 (15m), 200 (1h)

**Selectors**:
```typescript
useCurrentTick(symbol)        // Get latest tick
useSymbolStats(symbol)        // Get symbol statistics
useOHLCV(symbol, timeframe)   // Get OHLCV candles
useSubscribedSymbols()        // Get all subscribed symbols
useRecentOHLCV(symbol)        // Get most recent candle
```

---

## 3. Browser localStorage Keys

### 3.1 Trading Configuration Keys

| Key | Purpose | Format | Managed By |
|-----|---------|--------|-----------|
| `dockHeight` | Bottom panel height | string (number) | App.tsx |
| `trading-app-storage` | Zustand persistence | JSON | useAppStore |
| `MarketDataStore` | Market data cache | JSON | useMarketDataStore |

### 3.2 Window & Layout Keys

| Key | Purpose | Format | Managed By |
|-----|---------|--------|-----------|
| `windowManagerState` | Chart window layout | JSON | windowManager |

**Structure**:
```json
{
  "layoutMode": "single|horizontal|vertical|grid",
  "charts": [
    {
      "id": "chart-1234567890-abc123def",
      "symbol": "EURUSD",
      "position": { "row": 0, "col": 0 }
    }
  ]
}
```

### 3.3 Chart Indicators Keys

| Key | Purpose | Format | Managed By |
|-----|---------|--------|-----------|
| `chart-indicators-{SYMBOL}` | Indicator configs | JSON Array | indicatorManager |

**Structure**:
```json
[
  {
    "id": "indicator-1",
    "name": "SMA 20",
    "type": "sma",
    "parameters": { "period": 20 },
    "visible": true,
    "color": "#FF0000"
  },
  {
    "id": "indicator-2",
    "name": "EMA 12",
    "type": "ema",
    "parameters": { "period": 12 },
    "visible": true
  }
]
```

**Supported Indicators**:
- SMA (Simple Moving Average)
- EMA (Exponential Moving Average)
- RSI (Relative Strength Index)
- MACD (Moving Average Convergence Divergence)

### 3.4 Chart Drawings Keys

| Key | Purpose | Format | Managed By |
|-----|---------|--------|-----------|
| `drawings-{SYMBOL}` | Drawing objects | JSON Array | drawingManager |

**Structure**:
```json
[
  {
    "id": "drawing-1",
    "type": "TREND_LINE",
    "points": [
      { "x": 100, "y": 1.2500 },
      { "x": 200, "y": 1.2550 }
    ],
    "style": {
      "color": "#FF0000",
      "lineWidth": 2,
      "lineStyle": "solid"
    },
    "locked": false,
    "visible": true
  }
]
```

**Supported Drawing Tools**:
- TREND_LINE
- HORIZONTAL_LINE
- VERTICAL_LINE
- FIBONACCI
- SUPPORT_RESISTANCE
- CHANNEL
- RECTANGLE
- TRIANGLE
- TEXT_NOTE

### 3.5 RTX5 Market Watch Keys

| Key | Purpose | Format | Managed By |
|-----|---------|--------|-----------|
| `rtx5_hidden_symbols` | Hidden from display | JSON Array | marketWatchActions |
| `rtx5_symbol_sets` | Custom symbol groups | JSON Object | marketWatchActions |
| `rtx5_marketwatch_cols` | Column visibility | JSON Array | marketWatchActions |
| `rtx5_marketwatch_options` | System settings | JSON Object | marketWatchActions |
| `rtx5_marketwatch_lp_filter` | LP filter toggle | string (boolean) | MarketWatchPanel |
| `rtx5_subscribed_symbols` | Active subscriptions | JSON Array | MarketWatchPanel |
| `rtx_token` | Authentication token | string | websocket-enhanced.ts |

**Details**:

#### rtx5_hidden_symbols
```json
["SYMBOL1", "SYMBOL2"]
```

#### rtx5_symbol_sets
```json
{
  "My Favorites": ["EURUSD", "GBPUSD"],
  "Forex Major": ["EURUSD", "GBPUSD", "USDJPY"],
  "Custom Set": ["XAUUSD", "USOIL"]
}
```

Default sets available:
- `forex.major`: EURUSD, GBPUSD, USDJPY, USDCHF, AUDUSD, USDCAD, NZDUSD
- `forex.crosses`: EURGBP, EURJPY, GBPJPY, EURAUD, EURCHF, AUDJPY, GBPAUD
- `forex.exotic`: USDTRY, USDZAR, USDMXN, USDSEK, USDNOK, USDDKK
- `commodities`: XAUUSD, XAGUSD, USOIL, UKOIL, NATGAS
- `indices`: US30, US500, US100, UK100, GER40, FRA40, JPN225

#### rtx5_marketwatch_cols
```json
["symbol", "bid", "ask", "spread", "dailyChange", "last", "high", "low", "volume", "time"]
```

Available columns:
- symbol
- bid
- ask
- spread
- dailyChange
- last
- high
- low
- volume
- time

#### rtx5_marketwatch_options
```json
{
  "useSystemColors": true,
  "showMilliseconds": false,
  "autoRemoveExpired": true,
  "autoArrange": true,
  "showGrid": true
}
```

---

## 4. Singleton Service Managers

### 4.1 WindowManager

**File**: `clients/desktop/src/services/windowManager.ts`

**Purpose**: Manages multi-chart window layouts

**Layout Modes**:
- `single`: Only one chart visible
- `horizontal`: Charts arranged in horizontal row
- `vertical`: Charts arranged in vertical column
- `grid`: Charts in grid layout (2x2, 3x3, etc.)

**Public API**:
```typescript
setLayoutMode(mode: LayoutMode): void
getLayoutMode(): LayoutMode
addChart(symbol: string): string  // Returns chart ID
removeChart(id: string): void
getCharts(): ChartWindow[]
getChartCount(): number
getGridDimensions(): { rows: number; cols: number }
getGridCSSClass(): string  // Tailwind classes
subscribe(callback: (mode: LayoutMode) => void): () => void  // Listener pattern
reset(): void
```

**Persistence**:
- Automatically loads/saves state to `windowManagerState` on changes
- Uses singleton pattern: `export const windowManager = new WindowManager();`

---

### 4.2 IndicatorManager

**File**: `clients/desktop/src/services/indicatorManager.ts`

**Purpose**: Manages technical indicators on charts

**IndicatorConfig Structure**:
```typescript
interface IndicatorConfig {
  id: string;
  name: string;
  type: string;  // 'sma', 'ema', 'rsi', 'macd'
  parameters: Record<string, any>;
  visible: boolean;
  color?: string;
}
```

**Public API**:
```typescript
addIndicator(config: IndicatorConfig): void
removeIndicator(id: string): void
toggleIndicator(id: string): void
updateIndicator(id: string, parameters: Record<string, any>): void
getIndicators(): IndicatorConfig[]
saveToStorage(symbol: string): void
loadFromStorage(symbol: string): void
setChart(chart: IChartApi | null): void
setOHLCData(data: OHLCData[]): void
```

**Persistence**:
- Symbol-specific: `chart-indicators-{SYMBOL}`
- Manual save/load methods
- Stores to localStorage as JSON array

---

### 4.3 DrawingManager

**File**: `clients/desktop/src/services/drawingManager.ts`

**Purpose**: Manages chart drawings (trendlines, patterns, etc.)

**Drawing Types Supported**:
- TREND_LINE
- HORIZONTAL_LINE
- VERTICAL_LINE
- FIBONACCI
- SUPPORT_RESISTANCE
- CHANNEL
- RECTANGLE
- TRIANGLE
- TEXT_NOTE

**Public API**:
```typescript
addDrawing(drawing: Drawing): void
removeDrawing(id: string): void
getDrawings(): Drawing[]
getDrawing(id: string): Drawing | undefined
toggleVisibility(id: string): void
lockDrawing(id: string): void
unlockDrawing(id: string): void
saveToStorage(symbol: string): void
loadFromStorage(symbol: string): void
clear(): void
```

**Persistence**:
- Symbol-specific: `drawings-{SYMBOL}`
- Stores array of drawing objects with coordinates and styling

---

## 5. React Context & Providers

### 5.1 CommandBusContext

**File**: `clients/desktop/src/contexts/CommandBusContext.tsx`

**Purpose**: Event-driven architecture for cross-component communication

**Custom Events Dispatched**:
```typescript
window.dispatchEvent(new CustomEvent('openChart', {
  detail: { symbol, timeframe, type }
}));

window.dispatchEvent(new CustomEvent('openOrderDialog', {
  detail: { symbol, type }
}));

window.dispatchEvent(new CustomEvent('openDepthOfMarket', {
  detail: { symbol }
}));

window.dispatchEvent(new CustomEvent('openSymbolsDialog'));

window.dispatchEvent(new CustomEvent('openPopupPrices', {
  detail: { symbol }
}));
```

---

## 6. Keyboard Shortcuts Storage

**File**: `clients/desktop/src/hooks/useKeyboardShortcuts.ts`

**Shortcuts** (hardcoded, not persisted):
```typescript
F9       - New Order Dialog
F10      - Chart Window
Alt+B    - Quick Buy (executes market order)
Alt+S    - Quick Sell (executes market order)
Ctrl+U   - Symbols Dialog
Ctrl+Shift+S - Save Layout
Ctrl+Shift+L - Load Layout
```

Currently **shortcuts are not configurable** - they are hardcoded in the application.

---

## 7. UI State Persistence

### 7.1 Dock Height

**File**: `clients/desktop/src/App.tsx` (lines 83-168)

```typescript
const [dockHeight, setDockHeight] = useState(() => {
  const saved = localStorage.getItem('dockHeight');
  return saved ? parseInt(saved, 10) : 250;
});

// Auto-save on change
useEffect(() => {
  localStorage.setItem('dockHeight', dockHeight.toString());
}, [dockHeight]);
```

**Default**: 250px

### 7.2 Chart Tabs

**Not persisted** - Open charts reset on page reload

**Structure**:
```typescript
interface ChartTab {
  id: string;
  symbol: string;
  timeframe: string;
}
```

---

## 8. Authentication & Session Management

### 8.1 Auth State (Zustand)

```typescript
interface AppState {
  isAuthenticated: boolean;
  accountId: string | null;
  authToken: string | null;
}
```

**Actions**:
- `setAuthenticated(isAuth, accountId, authToken)`
- `setAuthToken(token)`
- `clearAuth()`

**Storage**:
- Persisted in `trading-app-storage` (Zustand default)
- Auth token also stored in `rtx_token` for WebSocket

### 8.2 WebSocket Authentication

**Files**:
- `clients/desktop/src/services/websocket.ts`
- `clients/desktop/src/services/websocket-enhanced.ts`

```typescript
const token = localStorage.getItem('rtx_token');
const urlWithAuth = this.addTokenToUrl(this.url, token);
```

---

## 9. Market Watch Configuration

### 9.1 Column Configuration

**File**: `clients/desktop/src/components/layout/MarketWatchPanel.tsx`

```typescript
const [visibleColumns, setVisibleColumns] = useState<ColumnId[]>(() => {
  const saved = localStorage.getItem('rtx5_marketwatch_cols');
  return saved ? JSON.parse(saved) : DEFAULT_VISIBLE_COLUMNS;
});

// Persist on change
useEffect(() => {
  localStorage.setItem('rtx5_marketwatch_cols', JSON.stringify(visibleColumns));
}, [visibleColumns]);
```

### 9.2 Hidden Symbols

**File**: `clients/desktop/src/services/marketWatchActions.ts`

```typescript
export function hideSymbol(symbol: string): string[] {
  const stored = localStorage.getItem('rtx5_hidden_symbols');
  const hidden = stored ? JSON.parse(stored) : [];

  if (!hidden.includes(symbol)) {
    const updated = [...hidden, symbol];
    localStorage.setItem('rtx5_hidden_symbols', JSON.stringify(updated));
    return updated;
  }
  return hidden;
}

export function getHiddenSymbols(): string[] {
  const stored = localStorage.getItem('rtx5_hidden_symbols');
  return stored ? JSON.parse(stored) : [];
}
```

### 9.3 Symbol Sets (Workspaces)

**File**: `clients/desktop/src/services/marketWatchActions.ts`

```typescript
export async function loadSymbolSet(setName: string): Promise<string[]> {
  const customSets = localStorage.getItem('rtx5_symbol_sets');
  const sets: Record<string, string[]> = customSets ? JSON.parse(customSets) : {};
  const symbols = sets[setName] || DEFAULT_SETS[setName] || [];

  for (const symbol of symbols) {
    await subscribeToSymbol(symbol);
  }
  return symbols;
}

export function saveSymbolSet(name: string, symbols: string[]): void {
  const customSets = localStorage.getItem('rtx5_symbol_sets');
  const sets: Record<string, string[]> = customSets ? JSON.parse(customSets) : {};
  sets[name] = symbols;
  localStorage.setItem('rtx5_symbol_sets', JSON.stringify(sets));
}
```

---

## 10. Complete Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    App Component (App.tsx)                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────┐    ┌──────────────────────────────┐  │
│  │  Local State (useState)  │    │ Zustand Stores              │  │
│  │  ─ selectedSymbol    │    │ ─ useAppStore               │  │
│  │  ─ positions         │    │   (selectedSymbol, chartType,│  │
│  │  ─ account           │    │    timeframe, orderVolume)   │  │
│  │  ─ openCharts        │    │ ─ useMarketDataStore        │  │
│  │  ─ orderPanel*       │    │   (ticks, OHLCV, stats)     │  │
│  └──────────────────────┘    └──────────────────────────────┘  │
│           ↓                                ↓                     │
│  ┌──────────────────────────────────────────────────────┐       │
│  │ Child Components                                      │       │
│  │ ├─ ChartTabs (tabs management)                       │       │
│  │ ├─ TradingChart (main chart display)                │       │
│  │ ├─ MarketWatchPanel (symbol list)                   │       │
│  │ ├─ BottomDock (positions/orders)                    │       │
│  │ ├─ TopToolbar (controls)                            │       │
│  │ └─ OrderPanelDialog (order entry)                   │       │
│  └──────────────────────────────────────────────────────┘       │
│           ↓                                                      │
│  ┌──────────────────────────────────────────────────────┐       │
│  │ Singleton Managers (Services)                        │       │
│  │ ├─ windowManager (layout state)                     │       │
│  │ ├─ indicatorManager (chart indicators)              │       │
│  │ ├─ drawingManager (chart drawings)                  │       │
│  │ └─ WebSocket services (data stream)                 │       │
│  └──────────────────────────────────────────────────────┘       │
│           ↓                                                      │
│  ┌──────────────────────────────────────────────────────┐       │
│  │ Browser Storage                                      │       │
│  │                                                      │       │
│  │ localStorage:                                        │       │
│  │ ├─ trading-app-storage (Zustand persist)            │       │
│  │ ├─ windowManagerState (layout)                      │       │
│  │ ├─ dockHeight (UI state)                            │       │
│  │ ├─ chart-indicators-{SYMBOL} (indicators)          │       │
│  │ ├─ drawings-{SYMBOL} (drawings)                    │       │
│  │ ├─ rtx5_* (market watch config)                    │       │
│  │ └─ rtx_token (auth token)                          │       │
│  └──────────────────────────────────────────────────────┘       │
│           ↓                                                      │
│  ┌──────────────────────────────────────────────────────┐       │
│  │ Backend API & WebSocket                             │       │
│  │ └─ http://localhost:7999                            │       │
│  └──────────────────────────────────────────────────────┘       │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 11. Data Persistence Patterns

### 11.1 Automatic Persistence (Zustand)

**Stores**: useAppStore, useMarketDataStore

**Mechanism**:
- Redux devtools middleware for debugging
- Persist middleware saves to localStorage on state change
- Storage key configured in persist options
- On reload: localStorage → Zustand store → Components

**Pros**:
- Automatic, transparent
- Selective field persistence (partialize)
- Integrated debugging

**Cons**:
- No profile switching
- All data in single key
- No versioning

### 11.2 Manual Persistence (Services)

**Services**: indicatorManager, drawingManager, windowManager

**Mechanism**:
```typescript
// Save
saveToStorage(symbol): void  // Called explicitly by component

// Load
loadFromStorage(symbol): void  // Called on mount

// Direct localStorage API
localStorage.setItem(key, JSON.stringify(data))
localStorage.getItem(key)
```

**Pros**:
- Fine-grained control
- Symbol-specific data
- Easy to debug

**Cons**:
- Manual management required
- Prone to inconsistency
- No transactions

### 11.3 Event-Based Updates (Market Watch)

**Pattern**: localStorage changes trigger component updates

```typescript
const [hiddenSymbols, setHiddenSymbols] = useState(() => {
  const saved = localStorage.getItem('rtx5_hidden_symbols');
  return saved ? JSON.parse(saved) : [];
});

// Component mutates localStorage directly
export function hideSymbol(symbol: string): void {
  const stored = localStorage.getItem('rtx5_hidden_symbols');
  const hidden = stored ? JSON.parse(stored) : [];
  const updated = [...hidden, symbol];
  localStorage.setItem('rtx5_hidden_symbols', JSON.stringify(updated));
  // Parent component useEffect detects change
}
```

**Pros**:
- Decoupled action handlers
- Can work across tabs

**Cons**:
- No built-in change detection
- Stale closure issues
- Must use useEffect to detect

---

## 12. Type Definitions

### 12.1 Key Types (from `clients/desktop/src/types/trading.ts`)

```typescript
export type PanelLayout = {
  marketWatch: { visible: boolean; width: number };
  orderBook: { visible: boolean; width: number };
  timeSales: { visible: boolean; width: number };
  orderEntry: { visible: boolean; height: number };
  chart: { visible: boolean; maximized: boolean };
  positions: { visible: boolean; height: number };
  alerts: { visible: boolean; height: number };
};

export type ThemeColors = {
  background: string;
  surface: string;
  border: string;
  textPrimary: string;
  textSecondary: string;
  textMuted: string;
  success: string;
  danger: string;
  warning: string;
  info: string;
  accent: string;
  buy: string;
  sell: string;
  bidColor: string;
  askColor: string;
};

export type ChartSettings = {
  showGrid: boolean;
  showVolume: boolean;
  showCrosshair: boolean;
  priceScale: 'AUTO' | 'LOGARITHMIC' | 'PERCENTAGE';
  backgroundColor: string;
  gridColor: string;
  textColor: string;
};
```

**Note**: These types are defined but **NOT currently persisted** - they're only in type definitions.

---

## 13. Current Limitations & Missing Features

### 13.1 No Profile Switching

**Current State**:
- ❌ No profile configuration files
- ❌ No profile selection UI
- ❌ No profile export/import
- ❌ No named workspaces (only one workspace)

### 13.2 No Workspace Management

**Current State**:
- ❌ Cannot save multiple workspace layouts
- ❌ Cannot switch between workspaces
- ❌ Chart tabs not persisted

### 13.3 No Theme Settings Persistence

**Current State**:
- ❌ ThemeColors defined but not stored
- ❌ No theme switching UI
- ❌ Dark theme hardcoded

### 13.4 No Keyboard Shortcut Customization

**Current State**:
- ❌ Shortcuts hardcoded in useKeyboardShortcuts
- ❌ No shortcut editor
- ❌ Cannot remap keys

### 13.5 No Chart Template Storage

**Current State**:
- ❌ Cannot save indicator/drawing combinations as templates
- ❌ Manual re-setup on new charts

### 13.6 No Multi-Account Support

**Current State**:
- Single accountId in state (accountId = "1")
- Account-specific settings not supported

---

## 14. Storage Statistics

### 14.1 Typical localStorage Usage

**Approximate sizes**:
- `trading-app-storage`: ~200 bytes
- `windowManagerState`: ~300 bytes
- `chart-indicators-{SYMBOL}`: ~500-1000 bytes per symbol (5-10 indicators)
- `drawings-{SYMBOL}`: ~2000-5000 bytes per symbol (if many drawings)
- `rtx5_symbol_sets`: ~500-2000 bytes
- `rtx5_marketwatch_*`: ~500 bytes total
- **Total**: ~5-15 KB typical

### 14.2 Scalability Limits

**Practical limits** (browser quota ~5-50 MB):
- Symbols with indicators: ~5000 before issues
- Drawings per symbol: ~1000 before performance degradation
- Custom symbol sets: ~100 before UI slowdown

---

## 15. Integration Points

### 15.1 Components Using AppStore

```
App.tsx
├─ TradingChart (chartType, timeframe, selectedSymbol)
├─ ChartTabs (selectedSymbol)
├─ OrderPanelDialog (orderVolume)
├─ MarketWatchPanel (selectedSymbol)
├─ BottomDock (positions, orders, trades, account)
└─ MenuBar (authentication, account)
```

### 15.2 Components Using localStorage

```
App.tsx (dockHeight)
MarketWatchPanel (columns, hidden, options, LP filter)
MarketWatch.tsx (LP filter)
TradingChart.tsx (indicators, drawings)
windowManager (layout)
WebSocket services (auth token)
```

---

## 16. Recommendations for Enhancement

### 16.1 Profile Support

```typescript
interface Profile {
  id: string;
  name: string;
  created: number;
  modified: number;
  config: {
    layout: LayoutMode;
    chartType: ChartType;
    timeframe: Timeframe;
    orderVolume: number;
    dockHeight: number;
    indicators: Record<string, IndicatorConfig[]>;
    drawings: Record<string, Drawing[]>;
    marketWatch: {
      columns: ColumnId[];
      hiddenSymbols: string[];
      symbolSets: Record<string, string[]>;
      options: SystemOptions;
    };
  };
}

// Storage key: profiles:list (array of profile IDs)
// Storage key: profile:{profileId} (full profile config)
// Storage key: profile:active (current profile ID)
```

### 16.2 Workspace Persistence

```typescript
interface Workspace {
  id: string;
  name: string;
  chartTabs: ChartTab[];
  activeChartId: string | null;
  modified: number;
}

// Storage key: workspaces:list
// Storage key: workspace:{workspaceId}
// Storage key: workspace:active
```

### 16.3 Theme Management

```typescript
interface Theme {
  id: string;
  name: string;
  colors: ThemeColors;
  chartSettings: ChartSettings;
}

// Storage key: themes:list
// Storage key: theme:{themeId}
// Storage key: theme:active
```

---

## 17. Summary Table

| Feature | Status | Storage | Component |
|---------|--------|---------|-----------|
| Trading State | ✅ Active | Zustand | useAppStore |
| Market Data | ✅ Active | Zustand | useMarketDataStore |
| Window Layout | ✅ Active | localStorage | windowManager |
| Indicators | ✅ Active | localStorage | indicatorManager |
| Drawings | ✅ Active | localStorage | drawingManager |
| Chart Tabs | ❌ Not Persisted | Memory | ChartTabs |
| Keyboard Shortcuts | ✅ Hardcoded | Code | useKeyboardShortcuts |
| Theme Settings | ❌ Not Persisted | Theme CSS | Tailwind |
| Profiles | ❌ Not Implemented | - | - |
| Workspaces | ✅ Single | localStorage | windowManager |
| Market Watch Config | ✅ Active | localStorage | marketWatchActions |
| Authentication | ✅ Active | Zustand + localStorage | useAppStore |

---

## 18. Code Examples

### 18.1 Reading Profile Data

```typescript
// Get current trading preferences
const appState = useAppStore.getState();
const { chartType, timeframe, orderVolume, selectedSymbol } = appState;

// Get market watch config
const hiddenSymbols = getHiddenSymbols();
const symbolSets = getSymbolSets();
const columns = loadColumnConfig();
const options = loadSystemOptions();

// Get layout
const layoutMode = windowManager.getLayoutMode();
const charts = windowManager.getCharts();

// Get chart indicators for active symbol
indicatorManager.loadFromStorage(selectedSymbol);
const indicators = indicatorManager.getIndicators();
```

### 18.2 Writing Profile Data

```typescript
// Update trading preferences
useAppStore.setState({
  chartType: 'candlestick',
  timeframe: '5m',
  orderVolume: 0.1
});

// Update market watch
hideSymbol('EURUSD');
saveSymbolSet('My Favorites', ['GBPUSD', 'USDJPY']);
saveColumnConfig(['symbol', 'bid', 'ask', 'spread']);
saveSystemOptions({ useSystemColors: true, showGrid: true });

// Update layout
windowManager.setLayoutMode('grid');
windowManager.addChart('GBPUSD');

// Update indicators
indicatorManager.addIndicator({
  id: 'ind-1',
  name: 'SMA 20',
  type: 'sma',
  parameters: { period: 20 },
  visible: true
});
indicatorManager.saveToStorage('EURUSD');
```

### 18.3 Clearing Profile Data

```typescript
// Clear all trading state
useAppStore.getState().reset();

// Clear layout
windowManager.reset();

// Clear indicators for symbol
indicatorManager.clearAllIndicators();
indicatorManager.saveToStorage('EURUSD');

// Clear drawings for symbol
drawingManager.clear();
drawingManager.saveToStorage('EURUSD');

// Clear market watch settings
localStorage.removeItem('rtx5_hidden_symbols');
localStorage.removeItem('rtx5_symbol_sets');
localStorage.removeItem('rtx5_marketwatch_cols');
localStorage.removeItem('rtx5_marketwatch_options');
```

---

## Conclusion

The Trading Engine application uses a **distributed profile architecture** with:
- **Zustand stores** for reactive state management
- **localStorage** for browser-level persistence
- **Singleton services** for chart configuration
- **No centralized profile system**

For full profile support, the application would need a dedicated profile management system with export/import capabilities, allowing users to save and restore complete trading configurations including layouts, indicators, drawings, and preferences.

