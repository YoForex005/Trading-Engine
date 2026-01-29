# TradingDashboard Configuration Audit Summary

## Quick Reference: Hardcoded Values

| Category | Location | Hardcoded Value | Impact |
|----------|----------|-----------------|--------|
| **Logo** | Sidebar | "RT" (static text) | Cannot rebrand |
| **Title** | Header | "Trading Platform" | Cannot customize |
| **Sidebar Width** | Layout | `w-16` (64px) | Not responsive |
| **Header Height** | Layout | `h-14` (56px) | Fixed size |
| **Order Panel Width** | Layout | `w-80` (320px) | Not responsive |
| **Order Book Panel** | Layout | `w-96` (384px) | Not responsive |
| **Background Color** | Theme | `#09090b` | Cannot change theme |
| **Chart Background** | Theme | `#131722` | Hard to theme |
| **Accent Color** | Theme | `emerald-500` | Single accent only |
| **Default Volume** | OrderEntry | `0.01` lots | Cannot configure |
| **Risk %** | OrderEntry | `1%` | Hardcoded default |
| **SL Pips** | OrderEntry | `20 pips` | Hardcoded default |
| **Lot Size** | Calculation | `100,000 units` | Global constant |
| **Pip Multiplier** | OrderEntry | JPY=100, Others=10000 | Rule-based only |
| **Price Decimals** | Display | JPY=3, Others=5 | Limited flexibility |
| **API Base URL** | Endpoints | `http://localhost:8080` | Not configurable |
| **View Default** | State | `trading` (always) | Cannot change |
| **Nav Items** | Component | 5 hardcoded buttons | Cannot customize |

---

## Configuration Score: 15% (of 100%)

### What IS Configurable
- Chart type (via UI)
- Timeframe (via UI)
- Selected symbol (via market watch)
- View mode (via tabs)
- Account/position data (API-driven)

### What IS NOT Configurable
- **Layout/Dimensions** (5 hardcoded values)
- **Theme/Colors** (10+ hardcoded classes)
- **Branding** (logo, title)
- **Trading Defaults** (4 hardcoded values)
- **API Configuration** (1 hardcoded URL)
- **Navigation Menu** (5 hardcoded items)
- **View Visibility** (all always shown)

---

## Configuration Extraction Plan

### Phase 1: Core Configuration (4 hours)

Create `/src/config/dashboard.config.ts`:
```typescript
export const DASHBOARD_CONFIG = {
  branding: { logo: 'RT', title: 'Trading Platform' },
  layout: { sidebarWidth: 64, headerHeight: 56, orderPanelWidth: 320 },
  api: { baseUrl: 'http://localhost:8080' },
  trading: { lotSize: 100000, defaultVolume: 0.01, defaultRisk: 1, defaultSL: 20 },
};
```

### Phase 2: Theme System (6 hours)

Create `/src/theme/` with:
- `colors.ts` - All color definitions
- `ThemeProvider.tsx` - Context provider
- `useTheme.ts` - Hook for consumers

### Phase 3: View Registry (4 hours)

Create `/src/config/views.config.ts`:
```typescript
export const VIEW_REGISTRY = [
  { id: 'trading', label: 'Trading', component: TradingView, enabled: true },
  { id: 'positions', label: 'Positions', component: PositionsView, enabled: true },
  // ... etc
];
```

---

## Impact Analysis

### If NO Configuration Changes Made
- Dashboard locked to single branding
- Cannot adapt to different layouts
- Difficult to test
- Hard to customize for clients
- API URL hardcoded to localhost

### If Configuration Extracted
- Reusable component
- Multi-tenant support
- Easy testing
- Client customization
- Environment-based API URLs

---

## Files Analyzed

1. **TradingDashboard.tsx** (Main component)
   - 379 lines
   - 5 sub-views
   - 1 config interface (missing)

2. **OrderEntry.tsx** (Order form)
   - 379 lines
   - 4 hardcoded defaults
   - No props for customization

3. **App.tsx** (Root component)
   - 559 lines
   - 1 default symbol (EURUSD)
   - API hardcoded

4. **AdminPanel.tsx** (Admin interface)
   - Symbol configuration management
   - Routing mode configuration
   - LP management

---

## Next Steps

### Immediate (Day 1)
1. Extract all theme colors to config file
2. Extract layout dimensions to constants
3. Extract API base URL to environment variable

### Short-term (Week 1)
1. Create ThemeProvider context
2. Make views configurable
3. Extract trading defaults

### Medium-term (Month 1)
1. Build dashboard configuration UI
2. Add multi-theme support
3. Support per-symbol configuration

---

## Memory Key for Reference

**Key:** `trading-dashboard-analysis`  
**Namespace:** `patterns`  
**Size:** 2669 bytes  
**Retrievable:** Yes (HNSW indexed)

```bash
npx @claude-flow/cli@latest memory retrieve --key "trading-dashboard-analysis" --namespace patterns
```

