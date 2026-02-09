# Documentation Index - Trading Engine Profile & Storage Architecture

This directory contains comprehensive documentation of the Trading Engine's profile data architecture and state management system.

## Documents Overview

### 1. **PROFILE_DATA_ARCHITECTURE.md** (Main Reference)
**Comprehensive, 18-section deep dive into the entire storage and state management system**

- Executive summary
- 4-tier architecture overview
- Zustand store details (useAppStore, useMarketDataStore)
- 13+ localStorage keys explained
- Singleton service managers (WindowManager, IndicatorManager, DrawingManager)
- React Context & event bus
- Keyboard shortcuts storage
- Complete data flow diagram
- Data persistence patterns (Zustand vs manual vs event-based)
- Type definitions
- Current limitations & missing features
- Storage statistics & scalability
- Integration points between components
- Recommendations for enhancement (Profiles, Workspaces, Themes)
- Summary table of all features
- Code examples for reading/writing/clearing profile data

**Use this for**: Understanding the full architecture, making design decisions

---

### 2. **STORAGE_ARCHITECTURE_DIAGRAM.txt** (Visual Guide)
**ASCII diagrams and flow charts showing the complete storage hierarchy**

- Overall data flow (9 layers)
- Zustand store persistence lifecycle
- Singleton service managers architecture
- Manual persistence patterns
- Market Watch configuration storage
- Authentication & session management
- Complete page reload lifecycle
- Performance optimization details
- Missing/not persisted features
- Storage quotas and scalability limits
- Storage mechanisms summary table (13 different key patterns)

**Use this for**: Visual understanding, presentations, whiteboarding

---

### 3. **PROFILE_STORAGE_QUICK_REFERENCE.md** (Developer Handbook)
**Quick lookup guide for practical development work**

- All localStorage keys (alphabetically sorted)
- Where persisted data is used (startup, interaction, changes)
- File paths (absolute paths to all storage-related files)
- Architecture summary
- Data types reference (Tick, Position, IndicatorConfig, Drawing, etc)
- Key operations (reading, writing, clearing)
- Keyboard shortcuts table
- Custom events reference
- Common issues & solutions
- Browser DevTools tips
- Migration path (if implementing profiles)
- Zustand store methods (all actions & selectors)
- Performance notes
- Common code patterns
- Testing tips

**Use this for**: Daily development, quick lookups, debugging

---

### 4. **This File (DOCS_INDEX.md)** (Navigation)
**Overview and navigation guide for all documentation**

---

## Quick Navigation

### For Different Roles

#### Product Manager / Architect
→ Read: **PROFILE_DATA_ARCHITECTURE.md** (Sections 1, 11-16)
Focus on:
- Current architecture (Section 2)
- Data persistence patterns (Section 11)
- Current limitations (Section 13)
- Recommendations for enhancement (Section 16)

#### Frontend Developer
→ Read: **PROFILE_STORAGE_QUICK_REFERENCE.md** (All sections)
Reference:
- Data types
- Key operations
- Code patterns
- Testing tips
- Common issues

#### System Architect
→ Read: **STORAGE_ARCHITECTURE_DIAGRAM.txt** (All sections)
Then: **PROFILE_DATA_ARCHITECTURE.md** (Sections 2-5, 16)
Focus on:
- Architecture layers
- Integration points
- Scalability limits
- Migration path

#### QA / Tester
→ Read: **PROFILE_STORAGE_QUICK_REFERENCE.md** (Sections: Common Issues, Testing Tips, Browser DevTools)
Then: **PROFILE_DATA_ARCHITECTURE.md** (Section 13 - Limitations)

---

## Key Concepts Summary

### Three Types of Storage

1. **Zustand Stores** (Reactive state management)
   - `useAppStore` - Trading configuration, auto-persisted
   - `useMarketDataStore` - Live market data, not persisted
   - Automatic hydration on page load
   - Middleware: `persist` + `devtools`

2. **Singleton Services** (Programmatic persistence)
   - `windowManager` - Layout management
   - `indicatorManager` - Chart indicators
   - `drawingManager` - Chart annotations
   - Manual `saveToStorage()` / `loadFromStorage()` calls

3. **Direct localStorage** (Direct API usage)
   - Market Watch configuration
   - UI preferences
   - Authentication tokens
   - No reactivity, manual sync required

### Key Storage Patterns

| Pattern | Storage | Persistence | Example |
|---------|---------|-------------|---------|
| Zustand Automatic | Zustand | Auto on change | `useAppStore.setState({...})` |
| Service Manual | localStorage | Explicit call | `indicatorManager.saveToStorage(symbol)` |
| Function-based | localStorage | Direct API | `hideSymbol('EURUSD')` |
| useState + useEffect | localStorage | useEffect listener | `MarketWatchPanel` state |

---

## File Locations (Absolute Paths)

### State Management Files
- `C:\...\clients\desktop\src\store\useAppStore.ts`
- `C:\...\clients\desktop\src\store\useMarketDataStore.ts`
- `C:\...\clients\desktop\src\store\toolbarState.ts`

### Service Managers
- `C:\...\clients\desktop\src\services\windowManager.ts`
- `C:\...\clients\desktop\src\services\indicatorManager.ts`
- `C:\...\clients\desktop\src\services\drawingManager.ts`
- `C:\...\clients\desktop\src\services\marketWatchActions.ts`

### UI Components
- `C:\...\clients\desktop\src\App.tsx` (main app, dockHeight)
- `C:\...\clients\desktop\src\components\layout\MarketWatchPanel.tsx`
- `C:\...\clients\desktop\src\components\professional\MarketWatch.tsx`
- `C:\...\clients\desktop\src\components\settings\OptionsDialog.tsx`

### Type Definitions
- `C:\...\clients\desktop\src\types\trading.ts`

### WebSocket & Network
- `C:\...\clients\desktop\src\services\websocket.ts`
- `C:\...\clients\desktop\src\services\websocket-enhanced.ts`

---

## All localStorage Keys Reference

**13 key patterns (some per-symbol):**

1. `chart-indicators-{SYMBOL}` - Indicator configs per symbol
2. `dockHeight` - UI state
3. `drawings-{SYMBOL}` - Drawing objects per symbol
4. `MarketDataStore` - Market data (Zustand)
5. `rtx5_hidden_symbols` - Market Watch hidden list
6. `rtx5_marketwatch_cols` - Market Watch columns
7. `rtx5_marketwatch_lp_filter` - LP filter toggle
8. `rtx5_marketwatch_options` - Market Watch system options
9. `rtx5_subscribed_symbols` - Current subscriptions
10. `rtx5_symbol_sets` - Custom symbol groups (workspaces)
11. `rtx_token` - Authentication token
12. `trading-app-storage` - Main Zustand store
13. `windowManagerState` - Layout configuration

---

## Current Limitations

The application currently lacks:

- ❌ **Profile Switching** - No profile selection UI
- ❌ **Profile Export/Import** - Cannot save/restore configurations
- ❌ **Multiple Workspaces** - Only one workspace supported
- ❌ **Chart Tab Persistence** - Open tabs reset on reload
- ❌ **Theme Customization** - Dark theme hardcoded
- ❌ **Keyboard Shortcut Remapping** - Hardcoded shortcuts
- ❌ **Multi-Account Support** - Single account only
- ❌ **Chart Templates** - Cannot save indicator/drawing combinations

---

## Code Examples

### Reading Profile Data
```typescript
const appState = useAppStore.getState();
const { chartType, timeframe, orderVolume } = appState;

const indicators = indicatorManager.getIndicators();
const layoutMode = windowManager.getLayoutMode();
```

### Writing Profile Data
```typescript
// Zustand (auto-persisted)
useAppStore.setState({ chartType: 'line', timeframe: '5m' });

// Services (manual)
indicatorManager.addIndicator(config);
indicatorManager.saveToStorage(symbol);

// Direct localStorage
hideSymbol('EURUSD');
```

### Clearing Profile Data
```typescript
useAppStore.getState().reset();
windowManager.reset();
localStorage.clear();
```

---

## Data Flow on Page Reload

```
Browser Loads App
  ↓
Read localStorage Keys
  ├─ trading-app-storage (Zustand)
  ├─ windowManagerState (WindowManager)
  ├─ rtx5_marketwatch_* (Market Watch)
  ├─ chart-indicators-* (IndicatorManager)
  └─ drawings-* (DrawingManager)
  ↓
Zustand Hydration (useAppStore)
  ├─ Restore trading preferences
  └─ Set initial state
  ↓
Component Mounting
  ├─ WindowManager loads layout
  ├─ MarketWatchPanel loads settings
  ├─ TradingChart loads indicators/drawings
  └─ App connects to WebSocket
  ↓
App Ready
  ├─ WebSocket streaming market data
  ├─ API fetching positions/orders/account
  └─ UI shows previous configuration
```

---

## Performance Characteristics

| Operation | Latency | Notes |
|-----------|---------|-------|
| localStorage.setItem() | 1-10ms | Synchronous |
| localStorage.getItem() | <1ms | Fast read |
| Zustand setState() | <1ms | In-memory update |
| Indicator calculation | 50-100ms | Offloaded to Web Worker |
| Chart render | 16-33ms | 30-60 FPS |

**Total startup time impact**: localStorage reads add ~10-50ms to app load

---

## Testing Strategy

### Unit Tests
- Test Zustand store actions
- Test service methods
- Test localStorage serialization/deserialization

### Integration Tests
- Test data persistence across page reload
- Test indicator/drawing persistence per symbol
- Test market watch configuration persistence

### E2E Tests
- Test complete user flow: login → configure → reload → verify
- Test export/import (if implemented)
- Test profile switching (if implemented)

---

## Migration & Upgrade Path

### Current State (Distributed)
Single localStorage keys for each feature:
```
localStorage: {
  'trading-app-storage': {...},
  'windowManagerState': {...},
  'chart-indicators-EURUSD': {...}
}
```

### Proposed State (Unified Profiles)
Named profiles with nested configuration:
```
localStorage: {
  'profile:scalping': {
    appStore: {...},
    windowManager: {...},
    indicators: {...},
    drawings: {...},
    marketWatch: {...}
  }
}
```

### Benefits
- Single export/import per profile
- Easy profile switching
- Profile comparison
- Profile templates
- Better organization

---

## Glossary

- **Tick** - Single market price update (bid, ask, timestamp)
- **OHLCV** - Open, High, Low, Close, Volume candle data
- **Indicator** - Technical analysis tool (SMA, EMA, RSI, MACD)
- **Drawing** - Chart annotation (trendline, support/resistance, etc)
- **Layout Mode** - Multi-chart arrangement (single, horizontal, vertical, grid)
- **Workspace** - Saved configuration (currently via rtx5_symbol_sets)
- **Profile** - Complete user configuration (not yet implemented)
- **Widget** - UI component with configurable state
- **Hydration** - Loading persisted state on app startup

---

## Support & Questions

### Architecture Questions
→ See: **PROFILE_DATA_ARCHITECTURE.md** (Sections 1-10)

### Storage Key Questions
→ See: **PROFILE_STORAGE_QUICK_REFERENCE.md** (All Storage Keys section)

### Implementation Questions
→ See: **PROFILE_STORAGE_QUICK_REFERENCE.md** (Code Patterns section)

### Problem Solving
→ See: **PROFILE_STORAGE_QUICK_REFERENCE.md** (Common Issues section)

---

## Document Maintenance

**Last Updated**: 2026-02-02
**Coverage**: Entire desktop application storage layer
**Scope**: All state management, persistence, and configuration

---

## Related Files in Repository

- `PROFILE_DATA_ARCHITECTURE.md` - Main reference document
- `STORAGE_ARCHITECTURE_DIAGRAM.txt` - Visual architecture
- `PROFILE_STORAGE_QUICK_REFERENCE.md` - Developer handbook
- `DOCS_INDEX.md` - This file (navigation)

---

## Next Steps for Enhancement

1. **Implement Profile System**
   - Add profile CRUD operations
   - Add profile selector UI
   - Add profile export/import

2. **Implement Workspace Persistence**
   - Save/restore chart tabs
   - Save/restore panel layouts
   - Save/restore keyboard shortcuts

3. **Add Theme Support**
   - Persist theme colors
   - Add theme selector
   - Support theme import/export

4. **Improve Data Integrity**
   - Add migration utilities
   - Add data validation
   - Add backup/restore

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-02-02 | Initial documentation of current architecture |

