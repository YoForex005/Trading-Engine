# Trading Engine: Complete Settings & Preferences Persistence Architecture

## Executive Summary

The Trading Engine currently has **fragmented settings persistence** across multiple layers:
- **Frontend localStorage**: Chart state, toolbar configuration, window layouts
- **Backend in-memory stores**: Notification preferences
- **Database tables**: Core trading data (users, accounts, orders, trades)
- **No unified profile system**: Settings are not grouped by user/account profile

This document maps the complete architecture and identifies opportunities to consolidate settings into **profile-based persistence**.

---

## 1. CURRENT PERSISTENCE LAYERS

### 1.1 Frontend Client (Browser localStorage)

**File**: `C:\Users\s s laptop bazar\Trading-Engine2\clients\desktop\src`

#### A. Zustand Store (Main App State)
**File**: `useAppStore.ts`
```typescript
// Persisted via Zustand middleware
partialize: (state) => ({
  selectedSymbol: state.selectedSymbol,      // Currently selected trading pair
  chartType: state.chartType,                // 'candlestick' | 'line' | 'area'
  timeframe: state.timeframe,                // '1m' | '5m' | '15m' | '1h' | '4h' | '1d'
  orderVolume: state.orderVolume,            // Default lot size (0.01, 0.1, etc.)
})
```
**Storage Key**: `trading-app-storage`

#### B. Window Manager (Layout Persistence)
**File**: `services/windowManager.ts`
```typescript
private loadState(): void {
  const saved = localStorage.getItem('windowManagerState');
  if (saved) {
    const state = JSON.parse(saved);
    this.layoutMode = state.layoutMode || 'single';  // 'single' | 'horizontal' | 'vertical' | 'grid'
    this.charts = state.charts || [];                 // Array of { id, symbol, position }
  }
}
```
**Storage Key**: `windowManagerState`
**Persisted Data**:
- Layout mode (single/horizontal/vertical/grid)
- Chart window positions and symbols

#### C. Dock Height (Bottom Panel)
**File**: `App.tsx`
```typescript
const [dockHeight, setDockHeight] = useState(() => {
  const saved = localStorage.getItem('dockHeight');
  return saved ? parseInt(saved, 10) : 250;
});

// Persist on change
useEffect(() => {
  localStorage.setItem('dockHeight', dockHeight.toString());
}, [dockHeight]);
```
**Storage Key**: `dockHeight`
**Persisted Data**: Height in pixels (default 250px)

#### D. Market Watch Customization
**File**: `services/marketWatchActions.ts`
```typescript
// Column configuration
localStorage.setItem('rtx5_marketwatch_cols', JSON.stringify(columns));
// Columns: 'symbol' | 'bid' | 'ask' | 'spread' | 'dailyChange' | 'last' | 'high' | 'low' | 'volume' | 'time'

// Hidden symbols list
localStorage.setItem('rtx5_hidden_symbols', JSON.stringify(hiddenSymbols));

// Symbol sets (favorites/groups)
localStorage.setItem('rtx5_symbol_sets', JSON.stringify({
  'My Favorites': ['EURUSD', 'GBPUSD'],
  'custom-set': ['XAUUSD', 'BTCUSD']
}));

// System options
localStorage.setItem('rtx5_marketwatch_options', JSON.stringify({
  useSystemColors: true,
  showMilliseconds: false,
  autoRemoveExpired: true,
  autoArrange: true,
  showGrid: true
}));
```
**Storage Keys**:
- `rtx5_marketwatch_cols`
- `rtx5_hidden_symbols`
- `rtx5_symbol_sets`
- `rtx5_marketwatch_options`

#### E. Toolbar & Drawing State
**File**: `store/toolbarState.ts`
```typescript
export interface ToolbarState {
  activeTool: 'cursor' | 'trendline' | 'hline' | 'vline' | 'text' | 'channel' | 'fibonacci' | 'shapes' | null;
  chartType: 'candlestick' | 'bar' | 'line' | 'area';
  timeframe: string;
  crosshairEnabled: boolean;
  autoScrollEnabled: boolean;
  chartShiftEnabled: boolean;
  candleWidth: number;  // Zoom level (2-50px)
  indicators: Array<{ name: string; params: any }>;
  drawings: Array<{ id: string; type: string; points: any[] }>;
}
```
**Current State**: In-memory reducer (NOT persisted)

#### F. Drawings & Indicators
**Files**:
- `services/drawingManager.ts`
- `services/indicatorManager.ts`

**Drawing Manager**:
```typescript
export interface Drawing {
  id: string;
  type: DrawingType;  // 'trendline' | 'hline' | 'vline' | 'text' | 'channel' | 'fibonacci' | 'shapes'
  subtype?: string;
  points: DrawingPoint[];
  text?: string;
  color?: string;
  lineWidth?: number;
  lineStyle?: 'solid' | 'dashed' | 'dotted';
  selected?: boolean;
  locked?: boolean;
}
```
**Indicator Manager**:
```typescript
export interface IndicatorConfig {
  id: string;
  name: string;
  type: string;
  parameters: Record<string, any>;
  visible: boolean;
  color?: string;
}
```
**Current State**: In-memory only (NOT persisted to localStorage/backend)
**Backend Support**: `services/drawingManager.ts` has `loadFromBackend()` method suggesting backend support exists

#### G. Types Definition
**File**: `types/trading.ts`

**Chart Settings**:
```typescript
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

**Theme Colors**:
```typescript
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
```

**Panel Layout**:
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
```

---

### 1.2 Backend In-Memory Stores

#### Notification Preferences
**File**: `backend/notifications/preferences.go`

```go
type UserPreferences struct {
  UserID         string
  Preferences    map[NotificationType]ChannelPreference  // Notification type -> channels
  Locale         string    // "en"
  Timezone       string    // "UTC"
  QuietHours     *QuietHours
  UnsubscribeAll bool
  UpdatedAt      time.Time
}

type ChannelPreference struct {
  Enabled         bool
  Channels        []NotificationChannel  // EMAIL, SMS, PUSH, IN_APP
  MinimumPriority Priority               // LOW, NORMAL, HIGH, CRITICAL
}

type QuietHours struct {
  Enabled   bool
  StartTime string  // "22:00"
  EndTime   string  // "08:00"
  Timezone  string  // "America/New_York"
}
```

**Supported Notification Types**:
- `NotifMarginCallWarning`
- `NotifStopOut`
- `NotifSecurityAlert`
- `NotifOrderExecuted`
- `NotifPositionClosed`
- `NotifLoginNewDevice`
- `NotifBalanceChange`
- `NotifPriceMovement`
- `NotifNewsAlert`
- `NotifTradingHoursChange`
- `NotifSystemMaintenance`

**Interface**: `PreferencesStore` with `Get()`, `Save()`, `Delete()` methods
**Implementation**: `InMemoryPreferencesStore` (resets on server restart)

---

### 1.3 Database Schema

**File**: `backend/database/schema.sql`

**Core Tables**:
```sql
-- User identity
CREATE TABLE users (
  user_id UUID PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  first_name VARCHAR(100),
  last_name VARCHAR(100),
  role VARCHAR(50) NOT NULL DEFAULT 'TRADER',
  created_at TIMESTAMP,
  is_active BOOLEAN DEFAULT TRUE
);

-- Trading accounts (one user -> multiple accounts)
CREATE TABLE accounts (
  account_id BIGSERIAL PRIMARY KEY,
  user_id UUID REFERENCES users(user_id),
  currency VARCHAR(10) NOT NULL DEFAULT 'USD',
  balance DECIMAL(18, 8) DEFAULT 0.0,
  leverage INT DEFAULT 100,
  account_type VARCHAR(50) DEFAULT 'RETAIL',
  group_id VARCHAR(50) DEFAULT 'demo-forex',
  created_at TIMESTAMP
);
```

**Current Issues**:
- No `user_profiles` or `account_profiles` table
- Settings scattered across localStorage and in-memory stores
- No centralized user preferences table in database

---

## 2. CURRENTLY PERSISTED SETTINGS BY CATEGORY

### Chart & Visualization
| Setting | Type | Current Storage | Persists On |
|---------|------|-----------------|------------|
| Chart Type | candlestick\|line\|area | localStorage | Manual (via Zustand) |
| Timeframe | 1m-1d | localStorage | Manual (via Zustand) |
| Zoom Level | number (2-50px) | Memory | Session only |
| Grid Visible | boolean | Memory | Session only |
| Volume Visible | boolean | Memory | Session only |
| Crosshair Visible | boolean | Memory | Session only |
| Price Scale | AUTO\|LOG\|PERCENT | Memory | Session only |
| Background Color | hex | Memory | Session only |
| Grid Color | hex | Memory | Session only |
| Text Color | hex | Memory | Session only |

### Window & Layout
| Setting | Type | Current Storage | Persists On |
|---------|------|-----------------|------------|
| Layout Mode | single\|horizontal\|vertical\|grid | localStorage | Manual |
| Chart Positions | { row, col } | localStorage | Manual |
| Open Charts | Array<{symbol, timeframe}> | Memory | Session only |
| Dock Height | number | localStorage | Manual |
| Market Watch Width | Implicit | Memory | Session only |
| Navigator Width | Implicit | Memory | Session only |

### Trading & Execution
| Setting | Type | Current Storage | Persists On |
|---------|------|-----------------|------------|
| Default Order Volume | number | localStorage | Manual (via Zustand) |
| Selected Symbol | string | localStorage | Manual (via Zustand) |
| Default Order Type | string | Memory | Session only |
| Slippage Tolerance | Implicit | Memory | Session only |
| Time in Force | Implicit | Memory | Session only |

### Market Watch
| Setting | Type | Current Storage | Persists On |
|---------|------|-----------------|------------|
| Visible Columns | string[] | localStorage | Manual |
| Hidden Symbols | string[] | localStorage | Manual |
| Symbol Sets | Record<name, symbols[]> | localStorage | Manual |
| Sort Field | string | Memory | Session only |
| Sort Direction | asc\|desc | Memory | Session only |

### Indicators & Drawings (NOT PERSISTED)
| Setting | Type | Current Storage | Persists On |
|---------|------|-----------------|------------|
| Indicator Templates | - | Memory | Session only |
| Indicator Parameters | - | Memory | Session only |
| Indicator Visibility | - | Memory | Session only |
| Drawing Templates | - | Memory | Session only |
| Drawing Colors/Style | - | Memory | Session only |
| Drawing Annotations | - | Memory | Session only |

### System & Notifications
| Setting | Type | Current Storage | Persists On |
|---------|------|-----------------|------------|
| Use System Colors | boolean | localStorage | Manual |
| Show Milliseconds | boolean | localStorage | Manual |
| Auto Remove Expired | boolean | localStorage | Manual |
| Auto Arrange | boolean | localStorage | Manual |
| Notification Channels | string[] | In-Memory (backend) | Session only |
| Notification Preferences | Record<type, preference> | In-Memory (backend) | Session only |
| Quiet Hours | { start, end, tz } | In-Memory (backend) | Session only |
| Locale | string | In-Memory (backend) | Session only |
| Timezone | string | In-Memory (backend) | Session only |

---

## 3. MISSING/INCOMPLETE PERSISTENCE

### HIGH PRIORITY - Critical Features Not Persisted
1. **Indicator Templates** - No way to save custom indicator setups
2. **Drawing Templates** - Trendlines, channels, annotations are lost on refresh
3. **Drawing Styling** - Colors, line widths, dash patterns not saved
4. **Chart Overlay Settings** - Volume bars, studies visibility per timeframe
5. **Keyboard Shortcuts** - No customization or persistence
6. **Alert Conditions** - Price alerts, volume alerts not saved
7. **Watchlist Alerts** - Custom alerts per symbol not persisted

### MEDIUM PRIORITY - Partial/Session-Only Persistence
1. **Theme/Color Scheme** - Defined in types but no UI to customize globally
2. **Panel Visibility** - Which panels (orderbook, time sales) are visible
3. **Tab State** - Indicator/drawing active tabs not persisted
4. **Zoom/Pan** - Chart viewport state not saved between sessions
5. **Cursor Tools** - Last selected drawing tool not remembered

### LOW PRIORITY - Advanced Features (Not Yet Implemented)
1. **Multi-Language Support** - Locale stored in backend but no UI
2. **Font Preferences** - Chart font size/family customization
3. **Time Format** - 24h vs 12h preference
4. **Number Format** - Decimal separator, thousands separator
5. **Keyboard Layouts** - Custom hotkey configurations

---

## 4. PROPOSED PROFILE-BASED ARCHITECTURE

### 4.1 Multi-Profile Support Strategy

Instead of storing settings globally per account, implement **trading profiles** where each profile contains a complete settings snapshot:

```typescript
// Profile Types - Organized by purpose
export enum ProfileCategory {
  DAY_TRADER = 'day_trader',        // Fast timeframes (1m-5m), scalping setup
  SWING_TRADER = 'swing_trader',    // Mid timeframes (1h-4h)
  POSITION_TRADER = 'position',     // Long timeframes (1d-1w)
  CUSTOM = 'custom'                 // User-defined
}

export interface TradingProfile {
  // Identity
  profileId: string;                          // UUID
  accountId: string;                          // Which account this profile belongs to
  name: string;                               // "Scalping-EURUSD", "Swing-Setup"
  category: ProfileCategory;                  // Classification

  // Metadata
  createdAt: Date;
  updatedAt: Date;
  isDefault: boolean;                         // Load on app start
  isActive: boolean;                          // Currently in use

  // ========== CHART SETTINGS ==========
  chartPreferences: {
    chartType: 'candlestick' | 'bar' | 'line' | 'area';
    timeframe: '1m' | '5m' | '15m' | '1h' | '4h' | '1d';
    zoomLevel: number;                        // 2-50px

    // Visual settings
    gridSettings: {
      visible: boolean;
      color: string;
      style: 'solid' | 'dashed' | 'dotted';
    };

    volumeSettings: {
      visible: boolean;
      color: string;
      opacity: number;
    };

    crosshairSettings: {
      visible: boolean;
      style: 'solid' | 'dashed' | 'dotted';
      color: string;
    };

    priceScaleSettings: {
      type: 'AUTO' | 'LOGARITHMIC' | 'PERCENTAGE';
      position: 'left' | 'right';
    };

    colors: {
      background: string;
      upCandle: string;
      downCandle: string;
      wickColor: string;
      gridColor: string;
      textColor: string;
    };
  };

  // ========== LAYOUT SETTINGS ==========
  layoutPreferences: {
    layoutMode: 'single' | 'horizontal' | 'vertical' | 'grid';

    // Panel visibility
    panels: {
      marketWatch: { visible: boolean; width: number };
      orderBook: { visible: boolean; width: number };
      timeSales: { visible: boolean; width: number };
      positions: { visible: boolean; height: number };
      orderEntry: { visible: boolean; height: number };
      alerts: { visible: boolean; height: number };
    };

    // Dock configuration
    dockHeight: number;
    dockPosition: 'bottom' | 'top' | 'right';
  };

  // ========== TRADING EXECUTION SETTINGS ==========
  tradingPreferences: {
    defaultOrderSize: number;                 // 0.01, 0.1, 1.0, etc.
    defaultOrderType: 'MARKET' | 'LIMIT' | 'STOP';
    defaultTimeInForce: 'GTC' | 'IOC' | 'FOK' | 'DAY';

    slippageTolerance: {
      enabled: boolean;
      pips: number;                           // 1.0, 2.5, etc.
    };

    // Risk management defaults
    riskManagement: {
      useSL: boolean;
      useTPs: boolean;
      defaultSLPips: number;
      defaultTPs: Array<{ percent: number; pips: number }>;  // Ladder
    };
  };

  // ========== MARKET WATCH SETTINGS ==========
  marketWatchPreferences: {
    visibleColumns: Array<
      'symbol' | 'bid' | 'ask' | 'spread' | 'change' |
      'changePercent' | 'volume' | 'high' | 'low' | 'time'
    >;

    hiddenSymbols: string[];                  // Symbols to hide from view

    symbolSets: Array<{
      name: string;
      symbols: string[];
      isDefault: boolean;
    }>;

    sorting: {
      column: string | null;
      direction: 'asc' | 'desc';
    };

    columnWidths: Record<string, number>;
  };

  // ========== INDICATORS SETTINGS ==========
  indicatorPreferences: {
    templates: Array<{
      id: string;
      name: string;                           // "EMA Cross 12/26"
      indicators: Array<IndicatorTemplate>;
      isDefault: boolean;
    }>;

    savedIndicators: Array<{
      id: string;
      name: string;
      type: string;                           // 'SMA', 'EMA', 'RSI', etc.
      parameters: Record<string, any>;
      visible: boolean;
      style: {
        color: string;
        lineWidth: number;
        lineStyle: 'solid' | 'dashed' | 'dotted';
      };
    }>;
  };

  // ========== DRAWINGS SETTINGS ==========
  drawingPreferences: {
    templates: Array<{
      id: string;
      name: string;                           // "Support/Resistance"
      drawings: Array<DrawingTemplate>;
      isDefault: boolean;
    }>;

    defaultStyles: {
      lineColor: string;
      fillColor: string;
      lineWidth: number;
      lineStyle: 'solid' | 'dashed' | 'dotted';
      fontSize: number;
      textColor: string;
    };

    // Saved drawing annotations
    savedAnnotations: Array<{
      id: string;
      type: string;
      symbol: string;
      timeframe: string;
      points: Array<{ time: number; price: number }>;
      text?: string;
      style: {
        color: string;
        lineWidth: number;
        lineStyle: string;
      };
    }>;
  };

  // ========== ALERTS SETTINGS ==========
  alertPreferences: {
    // Price alerts
    priceAlerts: Array<{
      id: string;
      symbol: string;
      condition: 'ABOVE' | 'BELOW' | 'CROSSES_UP' | 'CROSSES_DOWN';
      targetPrice: number;
      enabled: boolean;
      sound: string;                         // Alert sound filename
      notificationChannels: string[];        // 'EMAIL', 'SMS', 'PUSH', 'IN_APP'
    }>;

    // Volume/Pattern alerts
    technicalAlerts: Array<{
      id: string;
      symbol: string;
      condition: string;                     // "Volume > 1M", "RSI > 70"
      enabled: boolean;
    }>;
  };

  // ========== KEYBOARD & HOTKEYS ==========
  keyboardPreferences: {
    shortcuts: Array<{
      key: string;                            // 'F9', 'Alt+B', 'Ctrl+U'
      action: string;                         // 'NEW_ORDER', 'QUICK_BUY', 'SYMBOLS_DIALOG'
      enabled: boolean;
      description: string;
    }>;
  };

  // ========== NOTIFICATION PREFERENCES ==========
  notificationPreferences: {
    locale: string;                           // 'en', 'es', 'fr'
    timezone: string;                         // 'UTC', 'America/New_York'

    quietHours: {
      enabled: boolean;
      startTime: string;                      // "22:00"
      endTime: string;                        // "08:00"
      timezone: string;
    };

    channels: Record<
      string,  // NotificationType
      {
        enabled: boolean;
        channels: string[];                   // EMAIL, SMS, PUSH, IN_APP
        minimumPriority: string;              // LOW, NORMAL, HIGH, CRITICAL
      }
    >;
  };

  // ========== THEME & APPEARANCE ==========
  appearancePreferences: {
    theme: 'light' | 'dark' | 'auto';

    colors: {
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

    font: {
      family: string;                        // 'monospace', 'sans-serif'
      size: number;                          // 12, 14, 16
    };

    formatting: {
      timeFormat: '24h' | '12h';
      decimalSeparator: '.' | ',';
      thousandsSeparator: ',' | '.';
      dateFormat: 'DD/MM/YYYY' | 'MM/DD/YYYY' | 'YYYY-MM-DD';
    };
  };
}

// Template types
interface IndicatorTemplate {
  id: string;
  type: string;                              // 'SMA', 'EMA', 'RSI'
  parameters: Record<string, any>;
  style: {
    color: string;
    lineWidth: number;
    lineStyle: string;
  };
}

interface DrawingTemplate {
  id: string;
  type: string;                              // 'TRENDLINE', 'CHANNEL'
  style: {
    color: string;
    lineWidth: number;
    lineStyle: string;
    fillColor?: string;
    fillOpacity?: number;
  };
  defaultText?: string;
}
```

### 4.2 Database Schema Addition

```sql
-- User Trading Profiles (per account)
CREATE TABLE user_profiles (
    profile_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id BIGINT NOT NULL REFERENCES accounts(account_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,

    -- Identity
    name VARCHAR(100) NOT NULL,               -- "Scalping-EURUSD", "Swing-Setup"
    category VARCHAR(50) NOT NULL,            -- 'day_trader', 'swing_trader', 'position', 'custom'
    description TEXT,

    -- Status
    is_active BOOLEAN DEFAULT FALSE,
    is_default BOOLEAN DEFAULT FALSE,         -- Load on app start

    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE,

    -- JSONB columns for flexible settings
    chart_settings JSONB NOT NULL,            -- ChartPreferences
    layout_settings JSONB NOT NULL,           -- LayoutPreferences
    trading_settings JSONB NOT NULL,          -- TradingPreferences
    market_watch_settings JSONB NOT NULL,     -- MarketWatchPreferences
    indicator_settings JSONB NOT NULL,        -- IndicatorPreferences
    drawing_settings JSONB NOT NULL,          -- DrawingPreferences
    alert_settings JSONB NOT NULL,            -- AlertPreferences
    keyboard_settings JSONB NOT NULL,         -- KeyboardPreferences
    notification_settings JSONB NOT NULL,     -- NotificationPreferences
    appearance_settings JSONB NOT NULL,       -- AppearancePreferences

    -- Full profile backup
    full_settings JSONB NOT NULL,             -- Complete TradingProfile object

    UNIQUE (account_id, name),
    CONSTRAINT fk_profile_account FOREIGN KEY (account_id) REFERENCES accounts(account_id)
);

-- Index for quick lookups
CREATE INDEX idx_user_profiles_active ON user_profiles(account_id, is_active);
CREATE INDEX idx_user_profiles_default ON user_profiles(account_id, is_default);

-- Audit/History of profile changes
CREATE TABLE profile_audit_log (
    log_id BIGSERIAL PRIMARY KEY,
    profile_id UUID NOT NULL REFERENCES user_profiles(profile_id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,              -- 'CREATED', 'UPDATED', 'ACTIVATED', 'DELETED'
    changed_fields JSONB,                     -- Which settings changed
    old_values JSONB,                         -- Previous values
    new_values JSONB,                         -- New values
    changed_by UUID REFERENCES users(user_id),
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indicator Templates (reusable across profiles)
CREATE TABLE indicator_templates (
    template_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id BIGINT NOT NULL REFERENCES accounts(account_id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    indicators JSONB NOT NULL,               -- Array of indicator configs
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Drawing Templates (reusable across profiles)
CREATE TABLE drawing_templates (
    template_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id BIGINT NOT NULL REFERENCES accounts(account_id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    drawing_configs JSONB NOT NULL,          -- Array of drawing configs
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Saved Price Alerts
CREATE TABLE price_alerts (
    alert_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id BIGINT NOT NULL REFERENCES accounts(account_id),
    symbol VARCHAR(20) NOT NULL,
    condition VARCHAR(50) NOT NULL,          -- 'ABOVE', 'BELOW', 'CROSSES_UP', 'CROSSES_DOWN'
    target_price DECIMAL(18, 8) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    notification_channels JSONB NOT NULL,    -- Array of channels
    sound VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    triggered_count INT DEFAULT 0,
    last_triggered_at TIMESTAMP WITH TIME ZONE
);

-- Keyboard Shortcut Customizations
CREATE TABLE keyboard_shortcuts (
    shortcut_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id BIGINT NOT NULL REFERENCES accounts(account_id),
    key_sequence VARCHAR(50) NOT NULL,       -- 'F9', 'Alt+B', 'Ctrl+Shift+U'
    action VARCHAR(100) NOT NULL,            -- 'NEW_ORDER', 'QUICK_BUY'
    enabled BOOLEAN DEFAULT TRUE,
    UNIQUE (account_id, key_sequence)
);
```

---

## 5. API ENDPOINTS FOR PROFILE MANAGEMENT

### 5.1 Profile CRUD Operations

```typescript
// Create new profile
POST /api/user/profiles
{
  name: "Scalping Setup",
  category: "day_trader",
  description: "Fast scalping on EURUSD 1m",
  settings: { /* full TradingProfile settings */ }
}
Response: { profileId, createdAt, ... }

// Get user's profiles
GET /api/user/profiles
Response: [ { profileId, name, category, isActive, ... } ]

// Get specific profile
GET /api/user/profiles/:profileId
Response: { /* complete TradingProfile */ }

// Update profile
PUT /api/user/profiles/:profileId
{
  name?: "Updated Name",
  settings?: { /* partial or full settings update */ }
}

// Delete profile
DELETE /api/user/profiles/:profileId

// Set active profile
POST /api/user/profiles/:profileId/activate
Response: { activeProfileId, timestamp }

// Clone profile
POST /api/user/profiles/:profileId/clone
{
  newName: "Scalping Setup v2"
}
Response: { newProfileId, ... }

// Export profile
GET /api/user/profiles/:profileId/export
Response: Downloadable JSON file with profile settings

// Import profile
POST /api/user/profiles/import
File: multipart/form-data with JSON file
Response: { profileId, importedAt }
```

### 5.2 Indicator & Drawing Templates

```typescript
// Save indicator template
POST /api/user/indicator-templates
{
  name: "EMA Cross 12/26",
  indicators: [
    { type: 'EMA', parameters: { period: 12 }, color: '#FF00FF' },
    { type: 'EMA', parameters: { period: 26 }, color: '#00FF00' }
  ]
}

// Get templates
GET /api/user/indicator-templates
Response: Array of IndicatorTemplate

// Apply template to current chart
POST /api/user/indicator-templates/:templateId/apply

// Similar endpoints for drawing-templates
POST /api/user/drawing-templates
GET /api/user/drawing-templates
POST /api/user/drawing-templates/:templateId/apply
```

### 5.3 Alert Management

```typescript
// Create price alert
POST /api/user/alerts
{
  symbol: "EURUSD",
  condition: "ABOVE",
  targetPrice: 1.1050,
  notificationChannels: ["EMAIL", "PUSH"]
}

// Get alerts
GET /api/user/alerts
GET /api/user/alerts/:symbol

// Update alert
PUT /api/user/alerts/:alertId

// Delete alert
DELETE /api/user/alerts/:alertId

// Test alert notification
POST /api/user/alerts/:alertId/test
```

---

## 6. IMPLEMENTATION ROADMAP

### Phase 1: Database & API (Week 1-2)
- [ ] Create profile tables in PostgreSQL
- [ ] Add JSONB columns for settings
- [ ] Create profile API endpoints (CRUD)
- [ ] Add basic validation and error handling
- [ ] Create database migrations

### Phase 2: Backend Integration (Week 2-3)
- [ ] Implement ProfileService (Go)
- [ ] Wire profiles to existing handlers
- [ ] Add profile activation logic
- [ ] Implement profile export/import
- [ ] Add audit logging

### Phase 3: Frontend Persistence Layer (Week 3-4)
- [ ] Create ProfileManager service (TypeScript)
- [ ] Migrate localStorage keys to profile structure
- [ ] Add Profile selection UI
- [ ] Implement profile switching
- [ ] Add profile creation/deletion forms

### Phase 4: UI Components (Week 4-5)
- [ ] Profile selector dropdown
- [ ] Profile management modal
- [ ] Settings export/import UI
- [ ] Profile templates dialog
- [ ] Indicator template builder

### Phase 5: Testing & Polish (Week 5-6)
- [ ] E2E testing for profile workflows
- [ ] Migration script for existing settings
- [ ] Performance optimization
- [ ] Documentation
- [ ] User guide

---

## 7. MIGRATION STRATEGY

### From Current Scattered Settings to Profiles

```typescript
// Migration function
async function migrateExistingSettingsToProfile(accountId: string) {
  const profile: TradingProfile = {
    profileId: generateUUID(),
    accountId,
    name: "Default - Migrated Settings",
    category: 'custom',

    chartPreferences: {
      chartType: JSON.parse(localStorage.getItem('trading-app-storage')).chartType,
      timeframe: JSON.parse(localStorage.getItem('trading-app-storage')).timeframe,
      // ... migrate other settings
    },

    layoutPreferences: {
      layoutMode: JSON.parse(localStorage.getItem('windowManagerState')).layoutMode,
      // ...
    },

    // ... continue for all settings categories
  };

  await saveProfileToDatabase(profile);
  localStorage.clear();  // or archive for rollback
}
```

### Backward Compatibility
1. Keep localStorage keys functional during transition
2. Sync database profiles back to localStorage if needed
3. Gradual migration: offer users opt-in to profiles initially
4. Full deprecation after 2-3 releases

---

## 8. BENEFITS OF PROFILE-BASED ARCHITECTURE

### For Users
- **Quick Setup Switching**: Switch between day trading and swing trading setups instantly
- **Saved Workflows**: Complete chart setup (indicators, drawings) saved per scenario
- **Cross-Device Sync**: Profiles stored server-side, accessible from any device
- **Shareable Setups**: Export/import profiles with other traders
- **Organized Settings**: All preferences grouped logically instead of scattered

### For Developers
- **Unified Settings API**: Single endpoint to read/write all settings
- **Type Safety**: Structured TradingProfile interface (TypeScript)
- **Auditability**: Track setting changes with ProfileAuditLog
- **Performance**: Single JSONB column instead of N localStorage keys
- **Extensibility**: Easy to add new setting categories
- **Testing**: Standardized profile fixtures for testing

### For the Business
- **Analytics**: Track which settings users prefer
- **Premium Features**: Profile templates, advanced saving options
- **Support**: Easier to diagnose user issues with profile snapshots
- **Recommendations**: Suggest profiles based on trading behavior

---

## 9. TECHNICAL CONSIDERATIONS

### JSONB vs Relational Tables

**Why JSONB for settings**:
- Settings schema evolves frequently
- Nested structures (chartSettings → gridSettings → color)
- User customization depth varies
- No complex queries needed (mostly read/write entire object)
- PostgreSQL JSONB indexing supports fast retrieval

**Exception - Separate tables for**:
- Price alerts (need real-time querying)
- Keyboard shortcuts (validation and uniqueness constraints)
- Indicator/drawing templates (reusable, frequently queried)

### Storage Estimates

Per user account:
- Profile object: ~5-10KB (JSON)
- 10 profiles per account: 50-100KB
- 1M users: 50-100GB (acceptable)

### Performance Considerations

- Profile switching should be <100ms (cached in memory)
- Setting changes should batch/debounce (avoid excessive API calls)
- JSONB indexes on frequently filtered fields (category, is_active)
- Consider caching profiles in Redis for high-frequency reads

---

## 10. CURRENT STATE MAPPING

### What Gets Grouped Into Each Profile

| Category | Current Storage | New Home |
|----------|-----------------|----------|
| Chart visual settings | Memory | profile.chartPreferences.colors |
| Timeframe/chart type | localStorage | profile.chartPreferences |
| Window layout | localStorage | profile.layoutPreferences |
| Dock height | localStorage | profile.layoutPreferences.dockHeight |
| Market watch columns | localStorage | profile.marketWatchPreferences.visibleColumns |
| Hidden symbols | localStorage | profile.marketWatchPreferences.hiddenSymbols |
| Symbol sets/favorites | localStorage | profile.marketWatchPreferences.symbolSets |
| Default order size | localStorage | profile.tradingPreferences.defaultOrderSize |
| Drawing tool state | Memory | profile.drawingPreferences |
| Indicator settings | Memory | profile.indicatorPreferences |
| Notification prefs | Backend in-memory | profile.notificationPreferences |
| Timezone/Locale | Backend in-memory | profile.notificationPreferences |
| Keyboard shortcuts | Memory | profile.keyboardPreferences |

---

## 11. EXISTING BACKEND SUPPORT

**Files with backend support infrastructure already in place**:

1. `backend/notifications/preferences.go` - PreferencesStore interface
   - Already has Get/Save/Delete methods
   - Can be adapted to use ProfileService

2. `backend/cbook/client_profile.go` - ClientProfile structure
   - Tracks client metrics (win rate, Sharpe ratio, etc.)
   - Can be extended to include UI preferences

3. Database migrations system
   - `backend/migrations/` directory with SQL files
   - Can add new profile tables here

---

## 12. SUMMARY & RECOMMENDATIONS

### Current State
- Settings scattered across 3+ layers (localStorage, in-memory, Zustand)
- Many critical features (indicators, drawings) not persisted at all
- No unified user profile concept
- Difficult to implement multi-account or multi-setup scenarios

### Recommended Approach
1. **Implement TradingProfile interface** in TypeScript (types are already modeled)
2. **Create database tables** for profiles and related data
3. **Build ProfileService** API endpoints
4. **Migrate existing localStorage** to database-backed profiles
5. **Add UI** for profile management, switching, export/import

### Quick Wins (Before Full Implementation)
- [ ] Persist indicator templates to localStorage (immediate 80% solution)
- [ ] Persist drawing state to localStorage
- [ ] Add keyboard shortcut customization UI
- [ ] Implement basic theme/color customization

### Long-Term Value
- Cross-device settings sync
- Trading methodology templates for teams
- Analytics on user preferences
- Premium profile features (templates, sharing)
- Foundation for collaborative trading rooms

