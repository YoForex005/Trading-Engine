# TradingDashboard Component Map & Configuration Points

## Component Hierarchy

```
TradingDashboard (Main Container)
├── Sidebar (LAYOUT: w-16, hardcoded icons/labels)
│   ├── Logo Display ("RT" - hardcoded)
│   └── NavButtons (5 items - all hardcoded)
│       ├── Trading (LayoutDashboard icon)
│       ├── Positions (TrendingUp icon)
│       ├── Order Book (BookOpen icon)
│       ├── History (History icon)
│       └── Admin (Settings icon)
│
├── Header (LAYOUT: h-14)
│   ├── Title ("Trading Platform" - hardcoded)
│   ├── Symbol Display (dynamic from store)
│   ├── Bid/Ask Display (dynamic)
│   └── Account Info Display (dynamic)
│
└── Content Area (Dynamic based on viewMode)
    ├── TradingView (active when viewMode === 'trading')
    │   ├── TradingChart (bg-[#131722] - hardcoded)
    │   └── Sidebar (w-80 - hardcoded)
    │       ├── AccountInfoDashboard
    │       └── OrderEntry
    │           ├── Volume Input (default 0.01 - hardcoded)
    │           ├── Price/Trigger Inputs
    │           ├── SL/TP Inputs
    │           ├── Risk Calculator (defaults: 1%, 20 pips - hardcoded)
    │           └── Place Order Button
    │
    ├── PositionsView
    │   └── PositionList
    │       └── Uses hardcoded: 100000 units/lot, pip calculation
    │
    ├── OrderBookView
    │   ├── OrderBook
    │   └── AccountInfoDashboard
    │
    ├── HistoryView
    │   └── TradeHistory
    │
    └── AdminView
        └── AdminPanel
            ├── Broker Config Tab
            ├── Liquidity Providers Tab
            └── Symbols Tab
```

---

## Configuration Points Matrix

### By Component

| Component | Line(s) | Config Type | Current Value | Configurable? |
|-----------|---------|-------------|---------------|---------------|
| TradingDashboard | 29-31 | ViewMode State | 'trading' | No (hardcoded) |
| Sidebar | 37 | Width | 64px (w-16) | No |
| Header | 71 | Height | 56px (h-14) | No |
| Header | 78 | Title | "Trading Platform" | No |
| Logo | 41-45 | Text/Style | "RT" + gradient | No |
| NavButton | 46-60 | Items/Labels | 5 hardcoded | No |
| TradingView | 157 | Panel Width | 320px (w-80) | No |
| OrderBook | 197 | Panel Width | 384px (w-96) | No |
| TradingChart | 161 | Background | #131722 | No |
| Theme | Global | Colors | Multiple | No |
| OrderEntry | 32 | Default Volume | 0.01 | No |
| OrderEntry | 41-42 | Risk Defaults | 1%, 20pips | No |
| API | AdminPanel | Base URL | localhost:8080 | No (hardcoded) |
| Formats | App.tsx | Price Decimals | JPY=3, Other=5 | Partially |

---

## Configuration Requirements by Level

### 1. Theme Level (Colors)

**Currently Hardcoded In:**
- TradingDashboard.tsx (16 class references)
- OrderEntry.tsx (12 class references)
- AdminPanel.tsx (8 class references)
- App.tsx (14 class references)
- Other components (10+ references)

**Should Be In:**
```typescript
// src/theme/colors.ts
export const COLORS = {
  background: '#09090b',
  sidebar: 'zinc-900/30',
  header: 'zinc-900/20',
  chart: '#131722',
  borders: 'zinc-800',
  text: {
    primary: 'zinc-300',
    secondary: 'zinc-500',
  },
  accent: {
    primary: 'emerald-500',
    dark: 'emerald-600',
    light: 'emerald-400',
  },
  status: {
    buy: 'emerald-400',
    sell: 'red-400',
  },
};
```

### 2. Layout Level (Dimensions)

**Currently Hardcoded In:**
- TradingDashboard.tsx (6 dimension values)
- OrderEntry.tsx (inline styles)
- Components (various)

**Should Be In:**
```typescript
// src/config/layout.config.ts
export const LAYOUT = {
  sidebar: { width: 64 },
  header: { height: 56 },
  panels: {
    orderEntry: { width: 320 },
    orderBook: { width: 384 },
  },
  spacing: {
    default: 16, // gap-4
    compact: 8,  // gap-2
  },
};
```

### 3. Branding Level

**Currently Hardcoded In:**
- TradingDashboard.tsx:43 ("RT")
- TradingDashboard.tsx:78 ("Trading Platform")

**Should Be In:**
```typescript
// src/config/branding.config.ts
export const BRANDING = {
  logo: 'RT',
  title: 'Trading Platform',
  faviconUrl: undefined,
  companyName: 'RTX Trading',
};
```

### 4. Trading Level (Defaults)

**Currently Hardcoded In:**
- OrderEntry.tsx:32 (0.01 volume)
- OrderEntry.tsx:41 (1% risk)
- OrderEntry.tsx:42 (20 pips SL)
- Multiple components (100000 lot size)
- App.tsx:55 (EURUSD default symbol)

**Should Be In:**
```typescript
// src/config/trading.config.ts
export const TRADING = {
  defaults: {
    volume: 0.01,
    riskPercent: 1,
    stopLossPips: 20,
  },
  symbols: {
    lotSize: 100000,
    defaultSymbol: 'EURUSD',
    priceDecimals: {
      JPY: 3,
      default: 5,
    },
  },
};
```

### 5. API Level

**Currently Hardcoded In:**
- AdminPanel.tsx:72 (config endpoint)
- AdminPanel.tsx:87 (LP endpoint)
- AdminPanel.tsx:99 (symbols endpoint)
- App.tsx:79 (config endpoint)
- App.tsx:93-94 (account/positions endpoints)
- App.tsx:135 (WebSocket URL)

**Should Be In:**
```typescript
// src/config/api.config.ts
export const API_CONFIG = {
  baseUrl: process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080',
  endpoints: {
    config: '/api/config',
    admin: {
      config: '/api/admin/config',
      lp: '/api/admin/liquidity-providers',
      symbols: '/api/admin/symbols',
    },
    account: '/api/account/summary',
    positions: '/api/positions',
    websocket: 'ws://localhost:8080/ws',
  },
};
```

### 6. View Registry Level

**Currently Hardcoded In:**
- TradingDashboard.tsx:26 (type ViewMode)
- TradingDashboard.tsx:28 (useState default)
- TradingDashboard.tsx:46-60 (NavButton items)
- TradingDashboard.tsx:168-174 (Content conditionals)

**Should Be In:**
```typescript
// src/config/views.config.ts
export const VIEW_REGISTRY = [
  {
    id: 'trading',
    label: 'Trading',
    icon: LayoutDashboard,
    component: TradingView,
    enabled: true,
  },
  {
    id: 'positions',
    label: 'Positions',
    icon: TrendingUp,
    component: PositionsView,
    enabled: true,
  },
  // ... more views
];

export const DEFAULT_VIEW = 'trading';
```

---

## Configuration Dependency Graph

```
CONFIG ROOT
├── BRANDING
│   └── Used by: Header, Sidebar
├── THEME (COLORS)
│   └── Used by: All components
├── LAYOUT
│   └── Used by: TradingDashboard, sub-components
├── TRADING
│   ├── Used by: OrderEntry, PositionList
│   └── Includes: Defaults, Symbol Config
├── API
│   └── Used by: AdminPanel, App, components
├── VIEWS
│   ├── Used by: TradingDashboard (nav, routing)
│   └── Includes: Label, Icon, Component, Enabled
└── ENVIRONMENT
    └── Used by: API Config, Feature Flags
```

---

## Implementation Roadmap

### Step 1: Extract Constants (2 hours)
- Create `/src/config/` directory
- Create `dashboard.config.ts` with all hardcoded values
- Update imports in components

### Step 2: Create Config Interface (1 hour)
- Define `TradingDashboardConfig` type
- Export default configuration
- Add type validation

### Step 3: Create Theme Provider (3 hours)
- Create `/src/theme/` directory
- Build ThemeProvider component
- Create useTheme hook
- Replace hardcoded colors with theme values

### Step 4: Make Components Configurable (4 hours)
- Update TradingDashboard props
- Pass config to child components
- Update OrderEntry defaults
- Update PositionList calculations

### Step 5: Environment Configuration (1 hour)
- Add .env support
- Load API URLs from environment
- Support different environments (dev/staging/prod)

**Total Effort: 11 hours**

---

## Files That Need Changes

1. **TradingDashboard.tsx** - Extract layout, branding, views
2. **OrderEntry.tsx** - Extract trading defaults
3. **PositionList.tsx** - Make lot size configurable
4. **AdminPanel.tsx** - Use API config
5. **App.tsx** - Use config, API config, theme
6. **Create new:**
   - `/src/config/dashboard.config.ts`
   - `/src/config/api.config.ts`
   - `/src/theme/colors.ts`
   - `/src/theme/ThemeProvider.tsx`
   - `/src/theme/useTheme.ts`

---

## Success Criteria

After refactoring:
1. All colors in theme config
2. All layout dimensions in layout config
3. All trading defaults in trading config
4. All API endpoints in api config
5. Views dynamically configurable
6. Can change theme without code changes
7. Can modify defaults without code changes
8. API URL configurable per environment

