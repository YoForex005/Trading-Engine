# Professional Trading Terminal - UI Architecture

**Version:** 1.0.0
**Date:** 2026-01-18
**Status:** Approved
**Architect:** System Architecture Designer

## Executive Summary

This document outlines the comprehensive architecture for transforming the web trading frontend into a professional-grade terminal with multi-panel layout, real-time data streaming, advanced charting, and institutional-quality features.

## 1. Architecture Overview

### 1.1 Design Principles

| Principle | Description | Implementation |
|-----------|-------------|----------------|
| **Modularity** | Component-based with clear boundaries | Panel-based architecture with registry pattern |
| **Performance** | Optimized for real-time data | Throttling, batching, Web Workers |
| **Scalability** | Handle multiple feeds and layouts | Event-driven architecture |
| **Extensibility** | Easy to add new features | Plugin system for panels |
| **Responsiveness** | Adaptive layouts | Breakpoint-based grid system |

### 1.2 Current vs Target Architecture

| Aspect | Current State | Target State |
|--------|---------------|--------------|
| Layout | Fixed 3-column | Drag-drop multi-panel grid |
| Panels | 5 basic panels | 12+ professional panels |
| Data Flow | Direct WebSocket → Component | Buffered → Store → UI (optimized) |
| State Management | Zustand (monolithic) | Zustand (sliced by domain) |
| Real-time Updates | Unthrottled (performance issues) | Throttled at 10-20 FPS |
| Customization | Fixed layout | Workspace presets + custom layouts |

## 2. Component Architecture

### 2.1 Layout System

**Technology Choice:** React Grid Layout

**Rationale:**
- Mature library with 10k+ stars
- Native drag-drop and resize support
- Breakpoint-responsive layouts
- Persistent state management
- No external dependencies beyond React

**Panel Structure:**

```typescript
interface Panel {
  id: string;
  type: PanelType;
  title: string;
  component: React.ComponentType<PanelProps>;
  defaultSize: { w: number; h: number };
  minSize: { w: number; h: number };
  maxSize?: { w: number; h: number };
  closable: boolean;
  settings?: PanelSettings;
}

type PanelType =
  | 'market-watch'
  | 'chart'
  | 'orderbook'
  | 'time-sales'
  | 'order-entry'
  | 'positions'
  | 'orders'
  | 'account-summary'
  | 'news'
  | 'watchlist'
  | 'options-chain'
  | 'strategy-builder';
```

### 2.2 Panel Registry

**Pattern:** Factory + Registry

```typescript
class PanelRegistry {
  private panels: Map<PanelType, PanelDefinition>;

  register(type: PanelType, definition: PanelDefinition): void;
  create(type: PanelType, props?: PanelProps): Panel;
  getAvailable(): PanelType[];
}
```

**Benefits:**
- Dynamic panel loading
- Easy to add new panel types
- Centralized panel metadata
- Type-safe panel creation

### 2.3 Core Panels

#### 2.3.1 Market Watch Panel

**Features:**
- Real-time quotes for multiple symbols
- Customizable columns (Bid, Ask, Spread, Change%, Volume)
- Symbol search with filtering
- Favorites/watchlist support
- Color-coded price direction indicators
- Right-click context menu (Chart, Trade, Add to Watchlist)

**Data Source:** `useMarketStore` → ticks stream (throttled 100ms)

#### 2.3.2 Advanced Chart Panel

**Current:** Lightweight Charts
**Enhancements:**

- **Technical Indicators:**
  - Moving Averages (SMA, EMA, WMA)
  - Bollinger Bands
  - RSI, MACD, Stochastic
  - Volume Profile
  - Fibonacci Retracements

- **Drawing Tools:**
  - Trend lines, channels
  - Horizontal/vertical lines
  - Rectangles, circles
  - Text annotations

- **Features:**
  - Multi-timeframe (1m to 1M)
  - Chart templates (save/load)
  - Position markers on chart
  - Order entry from chart
  - Screenshot export

#### 2.3.3 Order Book Panel

**Visualization:**
- Depth ladder with bid/ask levels
- Volume bars (horizontal)
- Spread indicator
- Cumulative volume
- Last trade marker

**Optimizations:**
- Virtual scrolling for deep orderbook
- Throttled updates (200ms)
- Level aggregation (group by price)

#### 2.3.4 Time & Sales Panel

**Features:**
- Real-time trade ticker
- Color-coded (green=buy, red=sell)
- Volume weighting
- Filtering by size/side
- Export to CSV

**Performance:**
- Window virtualization (react-window)
- Max 1000 rows in memory
- Scrollback buffer

#### 2.3.5 Professional Order Entry Panel

**Order Types:**
- Market
- Limit
- Stop
- Stop-Limit
- OCO (One-Cancels-Other)
- Bracket (Entry + SL + TP)

**Features:**
- Quick order buttons (1-click trading)
- Volume presets (0.01, 0.1, 1.0)
- Risk calculator (% risk → lot size)
- Margin preview
- SL/TP in pips or price
- Time in force (GTC, IOC, FOK)

#### 2.3.6 Positions Panel

**Current State:** Basic position list
**Enhancements:**

- Group by symbol
- Net position calculation
- P&L charts (sparklines)
- Quick modify (drag SL/TP)
- Partial close slider
- Bulk actions (Close All, Close Winners, Close Losers)
- Export to CSV

#### 2.3.7 Account Summary Panel

**Metrics:**
- Balance, Equity, Margin
- Free Margin, Margin Level
- P&L (Daily, Weekly, Total)
- Win Rate, Profit Factor
- Largest Win/Loss
- Average Trade

**Visualizations:**
- Equity curve (mini chart)
- Margin usage gauge
- Daily P&L bar chart

#### 2.3.8 News & Alerts Panel

**Sources:**
- Economic calendar integration
- Custom alerts (price levels)
- System notifications
- Trade confirmations

**Features:**
- Filter by severity/category
- Audio alerts
- Desktop notifications
- Snooze/dismiss

### 2.4 Layout Presets

**Predefined Workspaces:**

1. **Trading Workspace**
   - Large chart (60%)
   - Market watch (15%)
   - Order entry (10%)
   - Positions (15%)

2. **Analysis Workspace**
   - Multi-chart layout (2x2)
   - Technical indicators
   - Market correlation heatmap

3. **Scalping Workspace**
   - Order book (30%)
   - Time & Sales (20%)
   - Chart with Level II (40%)
   - Quick order panel (10%)

4. **News Trader Workspace**
   - Economic calendar (30%)
   - News feed (30%)
   - Chart (30%)
   - Positions (10%)

## 3. State Management Architecture

### 3.1 Store Slicing Pattern

**Current:** Monolithic Zustand store
**Target:** Domain-sliced stores with selective subscriptions

```typescript
// Market data store
const useMarketStore = create<MarketState>()(
  persist(
    subscribeWithSelector(
      immer((set, get) => ({
        ticks: {},
        symbols: [],
        orderbook: {},
        updateTick: (tick) => set((state) => {
          state.ticks[tick.symbol] = tick;
        }),
        // ... other actions
      }))
    ),
    { name: 'market-store' }
  )
);

// Account store
const useAccountStore = create<AccountState>()(
  persist(
    subscribeWithSelector(
      immer((set) => ({
        balance: 0,
        equity: 0,
        positions: [],
        // ... other state
      }))
    ),
    { name: 'account-store' }
  )
);

// Layout store
const useLayoutStore = create<LayoutState>()(
  persist((set) => ({
    panels: [],
    layout: [],
    activeWorkspace: 'trading',
    // ... layout state
  }), { name: 'layout-store' })
);
```

### 3.2 Middleware Stack

| Middleware | Purpose | Usage |
|------------|---------|-------|
| `persist` | LocalStorage persistence | All stores |
| `subscribeWithSelector` | Granular subscriptions | Market, Account stores |
| `immer` | Immutable updates | Complex state updates |
| `devtools` | Redux DevTools integration | Development only |

### 3.3 Subscription Optimization

**Problem:** Re-rendering entire component tree on every tick
**Solution:** Selective subscriptions

```typescript
// ❌ Bad: Re-renders on any tick update
const ticks = useMarketStore(state => state.ticks);

// ✅ Good: Only re-renders on specific symbol
const eurUsdTick = useMarketStore(
  state => state.ticks['EURUSD'],
  shallow
);

// ✅ Better: Memoized selector
const useSymbolTick = (symbol: string) =>
  useMarketStore(
    useCallback(state => state.ticks[symbol], [symbol]),
    shallow
  );
```

## 4. Real-Time Data Flow

### 4.1 WebSocket Manager Architecture

**Current Issues:**
- No reconnection logic
- Unthrottled updates (causes UI lag)
- No message queuing during disconnection

**Target Architecture:**

```typescript
class WebSocketManager {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private messageQueue: Message[] = [];
  private subscriptions: Map<string, Set<Callback>>;
  private updateBuffer: Map<string, any> = new Map();
  private flushInterval: NodeJS.Timer | null = null;

  connect(url: string, token: string): void {
    // Implement exponential backoff
    // Start flush interval (100ms)
  }

  subscribe(channel: string, callback: Callback): Unsubscribe {
    // Add to subscriptions map
  }

  send(message: Message): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      this.messageQueue.push(message);
    }
  }

  private onMessage(event: MessageEvent): void {
    const data = JSON.parse(event.data);

    // Buffer updates instead of immediate callback
    this.updateBuffer.set(data.channel, data);
  }

  private flushUpdates(): void {
    // Called every 100ms
    this.updateBuffer.forEach((data, channel) => {
      const callbacks = this.subscriptions.get(channel);
      callbacks?.forEach(cb => cb(data));
    });
    this.updateBuffer.clear();
  }
}
```

**Benefits:**
- Automatic reconnection with backoff
- Throttled updates (10 FPS = 100ms)
- Message queue during disconnection
- Subscription-based architecture
- Shared WebSocket for all components

### 4.2 Data Streams

| Stream | Channel | Update Rate | UI Throttle | Priority |
|--------|---------|-------------|-------------|----------|
| Ticks | `market.ticks` | Real-time | 100ms | High |
| Order Book | `market.depth.{symbol}` | Real-time | 200ms | Medium |
| Trades | `market.trades.{symbol}` | Real-time | Immediate | High |
| Account | `account.{accountId}` | On change | 500ms | Medium |
| News | `news.feed` | On publish | Immediate | Low |

### 4.3 Tick Buffer Implementation

```typescript
class TickBuffer {
  private buffer: Map<string, Tick> = new Map();
  private flushCallback: (ticks: Map<string, Tick>) => void;
  private interval: NodeJS.Timer;

  constructor(flushCallback: typeof this.flushCallback) {
    this.flushCallback = flushCallback;
    this.interval = setInterval(() => this.flush(), 100);
  }

  add(tick: Tick): void {
    this.buffer.set(tick.symbol, tick);
  }

  private flush(): void {
    if (this.buffer.size > 0) {
      this.flushCallback(new Map(this.buffer));
      this.buffer.clear();
    }
  }

  destroy(): void {
    clearInterval(this.interval);
  }
}
```

## 5. Performance Optimizations

### 5.1 Rendering Optimizations

| Technique | Component | Impact |
|-----------|-----------|--------|
| `React.memo` | All panels, rows | 50% reduction in re-renders |
| `useMemo` | Derived calculations | Avoid recalculation |
| `useCallback` | Event handlers | Prevent child re-renders |
| Virtual scrolling | Lists (Positions, Trades) | Handle 10k+ rows |
| Lazy loading | Non-visible panels | Faster initial load |
| Code splitting | Panel components | Smaller bundle size |

### 5.2 Data Optimizations

**Tick Buffer with Priority Queue:**

```typescript
interface TickUpdate {
  symbol: string;
  tick: Tick;
  priority: number; // Higher = more important
}

class PriorityTickBuffer {
  private buffer: TickUpdate[] = [];
  private visibleSymbols: Set<string> = new Set();

  add(symbol: string, tick: Tick): void {
    const priority = this.visibleSymbols.has(symbol) ? 10 : 1;
    this.buffer.push({ symbol, tick, priority });
  }

  flush(): TickUpdate[] {
    // Sort by priority, return top N
    return this.buffer
      .sort((a, b) => b.priority - a.priority)
      .slice(0, 50); // Only update top 50 symbols
  }
}
```

### 5.3 Network Optimizations

- **WebSocket Compression:** Enable `permessage-deflate`
- **Binary Protocol:** Use MessagePack instead of JSON (30% smaller)
- **Differential Updates:** Only send changed orderbook levels
- **HTTP/2:** For API requests
- **CDN:** Static assets

### 5.4 Memory Optimizations

- **Tick History Limit:** Max 500 ticks per symbol (circular buffer)
- **Virtual Scrolling:** Render only visible rows
- **WeakMap for Memoization:** Auto garbage collection
- **IndexedDB for Historical Data:** Offload chart data
- **Web Worker for Parsing:** Offload JSON parsing

## 6. Responsive Layout System

### 6.1 Breakpoints

```typescript
const breakpoints = {
  xs: '0px',      // Mobile portrait
  sm: '640px',    // Mobile landscape
  md: '768px',    // Tablet
  lg: '1024px',   // Desktop
  xl: '1440px',   // Large desktop
  '2xl': '1920px' // Ultra-wide
};
```

### 6.2 Adaptive Behavior

| Breakpoint | Layout Strategy | Panel Behavior |
|------------|----------------|----------------|
| xs, sm | Single column | Bottom sheets, tabs |
| md | 2-column grid | Collapsible panels |
| lg, xl | 3-4 column grid | Full multi-panel |
| 2xl | Multi-monitor sync | Extended workspace |

### 6.3 Mobile-First Panels

**Bottom Sheets for Mobile:**
- Order Entry → Bottom sheet
- Positions → Swipeable list
- Chart → Full screen with controls
- Market Watch → Searchable list

## 7. Theme System

### 7.1 Theme Structure

```typescript
interface Theme {
  name: string;
  colors: {
    background: string;
    surface: string;
    primary: string;
    secondary: string;
    success: string;
    danger: string;
    warning: string;
    text: {
      primary: string;
      secondary: string;
      muted: string;
    };
    chart: {
      background: string;
      grid: string;
      up: string;
      down: string;
    };
  };
  fonts: {
    ui: string;
    mono: string;
  };
  spacing: 'compact' | 'normal' | 'comfortable';
}
```

### 7.2 Predefined Themes

1. **Dark (Default):** Current zinc/slate palette
2. **Light:** White background, blue accents
3. **Midnight:** Deep blacks, OLED-optimized
4. **TradingView:** TradingView-inspired colors
5. **Bloomberg:** Bloomberg Terminal aesthetic

### 7.3 CSS Variables Implementation

```css
:root {
  --color-bg-primary: #09090b;
  --color-bg-surface: #18181b;
  --color-text-primary: #fafafa;
  --color-success: #10b981;
  --color-danger: #ef4444;
  /* ... more variables */
}

[data-theme="light"] {
  --color-bg-primary: #ffffff;
  --color-bg-surface: #f4f4f5;
  --color-text-primary: #18181b;
  /* ... overrides */
}
```

## 8. Accessibility

### 8.1 Keyboard Shortcuts

**Global Shortcuts:**
- `F2`: Quick order entry
- `F9`: Close all positions
- `F12`: Developer tools
- `Ctrl+L`: Focus symbol search
- `Ctrl+T`: New workspace tab
- `Ctrl+W`: Close active panel
- `Ctrl+1-9`: Switch to workspace 1-9

**Trading Shortcuts:**
- `B`: Market buy current symbol
- `S`: Market sell current symbol
- `Q`: Quick close selected position
- `M`: Modify selected position (SL/TP)
- `Esc`: Cancel pending order entry

**Navigation Shortcuts:**
- `Tab`: Next panel
- `Shift+Tab`: Previous panel
- `Arrow Keys`: Navigate lists/tables
- `Enter`: Confirm action
- `Esc`: Cancel/close

### 8.2 Screen Reader Support

- ARIA labels on all interactive elements
- Live regions for account updates
- Descriptive button text (not just icons)
- Keyboard focus indicators

### 8.3 Color Contrast

- WCAG AA compliant (4.5:1 for text)
- High contrast mode option
- Color-blind friendly palettes

## 9. Security Considerations

### 9.1 Authentication & Authorization

- **JWT Tokens:** Current implementation ✅
- **Token Refresh:** Add refresh token mechanism
- **Secure Storage:** httpOnly cookies for tokens
- **Session Timeout:** Auto-logout after inactivity

### 9.2 Data Protection

- **TLS for WebSocket:** wss:// protocol
- **XSS Protection:** Sanitize all user inputs
- **CSRF Protection:** CSRF tokens for mutations
- **Content Security Policy:** Strict CSP headers

### 9.3 Secure Coding Practices

- Input validation on client and server
- No sensitive data in localStorage (only non-sensitive prefs)
- Obfuscate API keys in production builds
- Regular dependency audits

## 10. Testing Strategy

### 10.1 Unit Testing

**Tools:** Vitest + React Testing Library

**Coverage Targets:**
- Components: 80%
- Stores: 90%
- Utils/helpers: 95%

**Test Categories:**
- Component rendering
- User interactions
- Store mutations
- WebSocket message handling

### 10.2 Integration Testing

**Scenarios:**
- Full order flow (Entry → Execution → Position → Close)
- WebSocket reconnection
- Layout persistence
- Multi-panel synchronization

### 10.3 E2E Testing

**Tools:** Playwright or Cypress

**Critical Paths:**
- Login → Trade → View Position → Close
- Symbol search → Chart → Order from chart
- Layout customization → Save → Restore

### 10.4 Performance Testing

**Metrics:**
- Time to Interactive (TTI) < 3s
- First Contentful Paint (FCP) < 1.5s
- WebSocket latency < 50ms
- UI update rate: 10-20 FPS
- Memory usage < 200MB

## 11. Implementation Roadmap

### Phase 1: Foundation (2 weeks)

**Week 1:**
- [ ] Implement Zustand store slicing (market, account, layout)
- [ ] Create WebSocketManager with reconnection
- [ ] Add tick buffering and throttling
- [ ] Set up panel registry pattern

**Week 2:**
- [ ] Integrate React Grid Layout
- [ ] Create base Panel component
- [ ] Implement layout persistence
- [ ] Add workspace presets (Trading, Analysis, Scalping)

### Phase 2: Core Panels (3 weeks)

**Week 3:**
- [ ] Enhanced Market Watch panel
- [ ] Order Book visualization
- [ ] Time & Sales ticker

**Week 4:**
- [ ] Advanced Chart with indicators
- [ ] Drawing tools on chart
- [ ] Position markers on chart

**Week 5:**
- [ ] Professional Order Entry panel
- [ ] Risk calculator
- [ ] Margin preview

### Phase 3: Features (2 weeks)

**Week 6:**
- [ ] Drag-drop panel customization
- [ ] Panel settings/preferences
- [ ] Theme system implementation
- [ ] Keyboard shortcuts

**Week 7:**
- [ ] News & Alerts panel
- [ ] Account summary enhancements
- [ ] Export/import layouts
- [ ] Multi-monitor support

### Phase 4: Polish (1 week)

**Week 8:**
- [ ] Performance optimization pass
- [ ] Accessibility improvements
- [ ] Mobile responsiveness
- [ ] Documentation
- [ ] User acceptance testing

## 12. Success Metrics

### 12.1 Performance Metrics

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| Time to Interactive | < 3s | ~4s | ⚠️ Needs improvement |
| WebSocket Latency | < 50ms | ~30ms | ✅ Good |
| UI Update Rate | 10-20 FPS | Uncapped | ❌ Needs throttling |
| Memory Usage | < 200MB | ~150MB | ✅ Good |
| Bundle Size | < 1MB | ~800KB | ✅ Good |

### 12.2 User Experience Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Layout Customization | 90% users customize | Track localStorage |
| Workspace Presets Usage | 70% use presets | Analytics event |
| Keyboard Shortcut Usage | 40% power users | Track shortcut events |
| Mobile Usage | < 10% on mobile | User agent detection |

### 12.3 Business Metrics

| Metric | Target | Impact |
|--------|--------|--------|
| Trade Execution Speed | < 100ms | Revenue |
| Platform Uptime | 99.9% | Retention |
| Support Tickets | -30% | Cost savings |
| User Retention | +20% | Revenue |

## 13. Risk Analysis

### 13.1 Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Performance degradation with many panels | Medium | High | Virtual scrolling, lazy loading |
| WebSocket reliability issues | Low | High | Auto-reconnect, message queue |
| Layout state corruption | Low | Medium | Schema validation, migration |
| Browser compatibility | Low | Medium | Progressive enhancement |

### 13.2 Business Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| User resistance to new UI | Medium | Medium | Phased rollout, A/B testing |
| Training overhead | Low | Low | Interactive tutorials |
| Mobile experience gaps | Medium | Low | Focus on desktop first |

## 14. Future Enhancements

### 14.1 Short-term (3-6 months)

- [ ] Options trading panel
- [ ] Strategy builder (visual algo trading)
- [ ] Advanced charting (multi-chart sync)
- [ ] Social trading features
- [ ] Mobile app (React Native)

### 14.2 Long-term (6-12 months)

- [ ] AI-powered trading signals
- [ ] Backtesting integration
- [ ] Portfolio analytics
- [ ] API for third-party integrations
- [ ] Desktop app (Electron)

## 15. Conclusion

This architecture provides a solid foundation for building a professional-grade trading terminal that rivals industry leaders like TradingView, MetaTrader, and Bloomberg Terminal. The modular design, performance optimizations, and extensibility ensure the platform can scale with user demands and evolving market requirements.

**Key Takeaways:**
1. **Modularity:** Panel-based architecture for flexibility
2. **Performance:** Throttled real-time updates for smooth UI
3. **Customization:** Drag-drop layouts with workspace presets
4. **Scalability:** Store slicing and WebSocket optimization
5. **Professional UX:** Keyboard shortcuts, themes, accessibility

---

**Next Steps:**
1. Review this architecture with stakeholders
2. Prioritize Phase 1 tasks
3. Set up project tracking (GitHub Projects)
4. Begin implementation with foundation work
