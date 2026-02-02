# Trading Profile Architecture - Quick Reference

## Current Settings Locations

```
┌─────────────────────────────────────────────────────────────┐
│                    Trading Application                       │
└─────────────────────────────────────────────────────────────┘
    ↓
┌───────────────────┬────────────────────┬────────────────────┐
│  BROWSER (Client) │   MEMORY (Server)  │  DATABASE (Persist)│
├───────────────────┼────────────────────┼────────────────────┤
│ localStorage:     │ In-Memory Stores:  │ PostgreSQL:        │
│                   │                    │                    │
│ • trading-app-    │ • Toolbar state    │ • users            │
│   storage         │ • Indicators       │ • accounts         │
│                   │ • Drawings         │ • trades           │
│ • windowManager   │ • Layout modes     │ • orders           │
│   State           │ • Profiles (NO!)   │ (NO USER PREFS)    │
│                   │                    │                    │
│ • dockHeight      │ • Notification     │ • notification     │
│                   │   Preferences      │   prefs (temp)     │
│ • rtx5_market-    │   (PreferencesStore│                    │
│   watch_cols      │    interface)      │                    │
│                   │                    │                    │
│ • rtx5_hidden-    │ • Temp session     │                    │
│   symbols         │   state (resets    │                    │
│                   │   on reload)       │                    │
│ • rtx5_symbol-    │                    │                    │
│   sets            │                    │                    │
│                   │                    │                    │
│ • rtx5_market-    │                    │                    │
│   watch_options   │                    │                    │
│                   │                    │                    │
│ [~30KB total]     │ [Session only]     │ [Trading data only]│
│                   │                    │ [NO UI SETTINGS]   │
└───────────────────┴────────────────────┴────────────────────┘
        ↓ Lost on refresh or browser clear
      [SESSION LOST]
```

## Proposed: Profile-Based Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                    Trading Application                                │
│                                                                       │
│     Profile 1: "Scalping"      Profile 2: "Swing Trading"           │
│     ├─ Chart: 1m candlestick   ├─ Chart: 1h candlestick            │
│     ├─ Indicators: [RSI, MACD] ├─ Indicators: [EMA, Bollinger]      │
│     ├─ Drawings: [Grid]        ├─ Drawings: [Support/Resistance]    │
│     ├─ Order Size: 0.01        ├─ Order Size: 0.5                   │
│     └─ Layout: 4x4 grid        └─ Layout: single focused            │
│                                                                       │
│     Profile 3: "Analysis"                                            │
│     ├─ Chart: 4h candlestick                                        │
│     ├─ Indicators: [MA, Volume Profile]                             │
│     └─ ...                                                           │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
    ↓ Save/Load instantly
    ↓ Switch profiles in <100ms
    ↓ Cross-device sync (cloud)
┌──────────────────────────────────────────────────────────────────────┐
│ PostgreSQL Database: user_profiles table                              │
│                                                                       │
│ profile_id | account_id | name          | chart_settings (JSONB)   │
│ ───────────┼────────────┼───────────────┼──────────────────────────│
│ uuid-001   │ 12345      │ "Scalping"    │ { chartType, timeframe..}│
│ uuid-002   │ 12345      │ "Swing"       │ { chartType, timeframe..}│
│ uuid-003   │ 12345      │ "Analysis"    │ { chartType, timeframe..}│
│                                                                       │
│ [Plus columns for:] layout_settings, trading_settings,              │
│ indicator_settings, drawing_settings, alert_settings, etc.          │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

## What Gets Grouped Into a Profile

### Chart Preferences
- Chart type (candlestick, bar, line, area)
- Timeframe (1m to 1d)
- Zoom level
- Grid settings (visible, color, style)
- Volume visibility
- Crosshair settings
- Price scale (Auto/Log/Percent)
- Candle colors (up/down)
- Wicks, shadows styling

### Layout Preferences
- Layout mode (single, horizontal, vertical, grid)
- Panel visibility (MarketWatch, OrderBook, TimeSales, Orders, Positions, Alerts)
- Panel widths/heights
- Dock position and height
- Navigator visibility/size

### Trading Preferences
- Default order size (0.01, 0.1, 1.0 lots)
- Default order type (Market, Limit, Stop)
- Default time in force (GTC, IOC, FOK, Day)
- Slippage tolerance (pips)
- Default stop-loss/take-profit settings
- Risk management presets

### Market Watch Preferences
- Visible columns (Symbol, Bid, Ask, Spread, Change, Volume, etc.)
- Hidden symbols list
- Symbol sets/favorites (Forex Majors, Commodities, Cryptos, Custom)
- Sort preferences (column, direction)
- Column widths

### Indicator Preferences
- **Saved indicator templates** - "EMA Cross 12/26", "RSI Divergence"
- Individual indicator settings (parameters, colors, visibility)
- Indicator organization/grouping
- Overlay vs separate window settings

### Drawing Preferences
- **Saved drawing templates** - "Support/Resistance Grid", "Trendline Channel"
- Default drawing styles (line color, width, dash pattern)
- Text annotation settings (font, size, color)
- Drawing layer visibility/locking
- Saved annotations (linked to chart and timeframe)

### Alert Preferences
- **Price alerts** (Above/Below/Crosses conditions)
- Volume alerts
- Technical alerts (RSI > 70, MACD crossover)
- Alert sounds and notification channels (Email, SMS, Push, In-App)
- Quiet hours

### Keyboard Preferences
- Custom hotkey mappings (F9 → New Order, Alt+B → Quick Buy)
- Enabled/disabled shortcuts
- Keyboard layout support

### Notification Preferences
- Timezone and locale
- Quiet hours (22:00 - 08:00)
- Notification channels per event type
- Priority filtering (only High/Critical during hours)

### Appearance Preferences
- Theme (Light, Dark, Auto)
- Color scheme customization
- Font family and size
- Number formatting (., , separators)
- Date/Time format (24h vs 12h, DD/MM vs MM/DD)

---

## Key Attributes of Each Profile

```typescript
TradingProfile {
  profileId: "uuid-12345"              // Unique identifier
  accountId: "demo-001"                // Which account
  name: "Scalping EURUSD"              // User-friendly name
  category: "day_trader"               // day_trader | swing_trader | position | custom
  description: "Fast scalping setup..."// Optional details

  // Status
  isActive: true                       // Currently in use
  isDefault: false                     // Load on app start

  // Timestamps
  createdAt: "2024-02-02T10:00:00Z"   // When created
  updatedAt: "2024-02-02T14:30:00Z"   // Last modified
  lastUsedAt: "2024-02-02T15:45:00Z"  // Last switched to

  // Complete settings (10+ categories)
  chartPreferences: { ... }
  layoutPreferences: { ... }
  tradingPreferences: { ... }
  // ... more categories
}
```

---

## Database Schema Addition

```sql
-- Main profile table
CREATE TABLE user_profiles (
    profile_id UUID PRIMARY KEY,
    account_id BIGINT REFERENCES accounts(account_id),
    user_id UUID REFERENCES users(user_id),

    -- Identity
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,

    -- Status
    is_active BOOLEAN DEFAULT FALSE,
    is_default BOOLEAN DEFAULT FALSE,

    -- Settings (JSONB for flexibility)
    chart_settings JSONB NOT NULL,
    layout_settings JSONB NOT NULL,
    trading_settings JSONB NOT NULL,
    indicator_settings JSONB NOT NULL,
    drawing_settings JSONB NOT NULL,
    alert_settings JSONB NOT NULL,
    notification_settings JSONB NOT NULL,
    appearance_settings JSONB NOT NULL,
    keyboard_settings JSONB NOT NULL,

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_used_at TIMESTAMPTZ
);

-- Audit log for profile changes
CREATE TABLE profile_audit_log (
    log_id BIGSERIAL PRIMARY KEY,
    profile_id UUID REFERENCES user_profiles(profile_id),
    action VARCHAR(50),          -- 'CREATED', 'UPDATED', 'ACTIVATED'
    changed_fields JSONB,        -- Which settings changed
    changed_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indicator templates (reusable)
CREATE TABLE indicator_templates (
    template_id UUID PRIMARY KEY,
    account_id BIGINT REFERENCES accounts(account_id),
    name VARCHAR(100) NOT NULL,
    indicators JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Drawing templates (reusable)
CREATE TABLE drawing_templates (
    template_id UUID PRIMARY KEY,
    account_id BIGINT REFERENCES accounts(account_id),
    name VARCHAR(100) NOT NULL,
    drawing_configs JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Price alerts
CREATE TABLE price_alerts (
    alert_id UUID PRIMARY KEY,
    account_id BIGINT REFERENCES accounts(account_id),
    symbol VARCHAR(20) NOT NULL,
    condition VARCHAR(50),       -- 'ABOVE', 'BELOW', 'CROSSES_UP'
    target_price DECIMAL(18,8),
    enabled BOOLEAN DEFAULT TRUE
);

-- Custom keyboard shortcuts
CREATE TABLE keyboard_shortcuts (
    shortcut_id UUID PRIMARY KEY,
    account_id BIGINT REFERENCES accounts(account_id),
    key_sequence VARCHAR(50) NOT NULL,
    action VARCHAR(100) NOT NULL,
    UNIQUE (account_id, key_sequence)
);
```

---

## API Endpoints

```
Profile Management
──────────────────
GET    /api/user/profiles              # List all profiles for account
POST   /api/user/profiles              # Create new profile
GET    /api/user/profiles/:id          # Get specific profile
PUT    /api/user/profiles/:id          # Update profile settings
DELETE /api/user/profiles/:id          # Delete profile
POST   /api/user/profiles/:id/activate # Make profile active
POST   /api/user/profiles/:id/clone    # Duplicate profile
GET    /api/user/profiles/:id/export   # Export as JSON file
POST   /api/user/profiles/import       # Import JSON file

Template Management
───────────────────
GET    /api/user/indicator-templates   # List indicator templates
POST   /api/user/indicator-templates   # Create indicator template
GET    /api/user/drawing-templates     # List drawing templates
POST   /api/user/drawing-templates     # Create drawing template
POST   /api/user/indicator-templates/:id/apply  # Apply to chart
POST   /api/user/drawing-templates/:id/apply    # Apply to chart

Alert Management
────────────────
GET    /api/user/alerts                # List all alerts
POST   /api/user/alerts                # Create price alert
PUT    /api/user/alerts/:id            # Update alert
DELETE /api/user/alerts/:id            # Delete alert
POST   /api/user/alerts/:id/test       # Test notification

Shortcut Management
───────────────────
GET    /api/user/shortcuts             # List all shortcuts
POST   /api/user/shortcuts             # Create custom shortcut
PUT    /api/user/shortcuts/:id         # Update shortcut
DELETE /api/user/shortcuts/:id         # Delete shortcut
```

---

## Current Gaps - What's NOT Persisted

| Feature | Status | Impact |
|---------|--------|--------|
| Indicator Configurations | ❌ Not Saved | Lost on refresh; can't save favorite setups |
| Drawing Annotations | ❌ Not Saved | Technical analysis work lost |
| Drawing Styles (color, width) | ❌ Not Saved | Manual re-styling every session |
| Indicator Templates | ❌ Not Saved | Can't reuse "EMA Cross 12/26" setup |
| Drawing Templates | ❌ Not Saved | Manual redraw of support/resistance |
| Keyboard Customization | ❌ Not Possible | Can't rebind hotkeys |
| Theme/Color Scheme | ⚠️ Type Only | Defined in types but no UI or persistence |
| Price Alerts | ⚠️ Backend Only | Not persisted; session-only |
| Multi-Profile Support | ❌ Not Supported | Can't switch between trading setups |
| Cross-Device Sync | ❌ Not Possible | Settings stuck on one browser |

---

## Benefits of Profile System

### For Traders
- **🚀 Fast Setup Switching** - Switch from scalping to swing trading in 1 click
- **💾 Save Work** - All indicators, drawings, alerts saved with profile
- **📱 Cloud Sync** - Access same setup from any device
- **🔄 Team Sharing** - Export/import trading templates with colleagues
- **📊 Better Organization** - Logical grouping instead of scattered settings
- **⏱️ Time Saving** - No more manual re-setup every session

### For Developers
- **🎯 Single API** - One endpoint for all settings (vs N localStorage keys)
- **🔒 Type Safety** - Strongly-typed TradingProfile interface
- **📝 Auditability** - Track all setting changes in audit log
- **🔌 Extensible** - Easy to add new setting categories
- **🚄 Performance** - Single JSONB read vs 20+ localStorage calls
- **🧪 Testable** - Standardized fixtures for testing

### For Business
- **📈 Analytics** - Which settings do traders prefer?
- **💰 Premium** - Paid profiles, template sharing marketplace
- **🎓 Support** - Diagnose issues with profile snapshots
- **🤖 Recommendations** - "Try this profile based on your trading style"

---

## Migration Path

### Phase 1: Prepare (Week 1)
- Design final TradingProfile schema
- Create database tables and migrations
- Build ProfileService backend API

### Phase 2: Backend APIs (Week 2)
- Implement profile CRUD endpoints
- Add profile activation logic
- Build export/import functionality
- Add audit logging

### Phase 3: Frontend Wiring (Week 3-4)
- Create ProfileManager service
- Migrate localStorage to profiles
- Add profile selector UI
- Implement profile switching

### Phase 4: Polish (Week 5)
- Profile templates UI
- Settings organization
- Documentation
- User education

### Quick Wins (Before Full Implementation)
```typescript
// Minimal changes to save indicator templates immediately
localStorage.setItem('indicator_templates', JSON.stringify([
  {
    name: "EMA Cross 12/26",
    indicators: [
      { type: 'EMA', period: 12, color: '#FF00FF' },
      { type: 'EMA', period: 26, color: '#00FF00' }
    ]
  }
]));
```

---

## File Locations Reference

**Frontend Settings Files**:
- `clients/desktop/src/store/useAppStore.ts` - Main app state
- `clients/desktop/src/services/windowManager.ts` - Layout persistence
- `clients/desktop/src/services/marketWatchActions.ts` - Market watch prefs
- `clients/desktop/src/store/toolbarState.ts` - Toolbar state (not persisted)
- `clients/desktop/src/services/drawingManager.ts` - Drawings (not persisted)
- `clients/desktop/src/services/indicatorManager.ts` - Indicators (not persisted)
- `clients/desktop/src/types/trading.ts` - Type definitions

**Backend Preferences Files**:
- `backend/notifications/preferences.go` - Notification preferences interface
- `backend/database/schema.sql` - Database schema (no profile table yet)

**Where to Add Profile Tables**:
- Create: `backend/migrations/008_add_user_profiles_tables.sql`
- Service: `backend/internal/api/handlers/profiles.go`
- Database Models: `backend/internal/models/profile.go`

---

## Summary

**Current State**: Settings fragmented across localStorage, in-memory stores, and scattered database tables. Many features (indicators, drawings, alerts) not persisted at all.

**Proposed State**: Unified profile system where each profile is a complete snapshot of UI/trading preferences, stored in database, with full CRUD APIs and profile-switching capability.

**Implementation**: 5-week roadmap to move from scattered settings to enterprise-grade profile management system.

**Value**: Better UX for traders, cleaner code for developers, foundation for premium features.

