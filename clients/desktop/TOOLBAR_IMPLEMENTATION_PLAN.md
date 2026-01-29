# RTX TERMINAL TOOLBAR - MT5 IMPLEMENTATION PLAN

## 🎯 MISSION: Achieve 100% MT5 Toolbar Parity

**Current State**: 73% of toolbar buttons are non-functional placeholders
**Target State**: Every button executes real logic with MT5-identical behavior
**Timeline**: 4 parallel agents, coordinated implementation

---

## 📊 TOOLBAR → ACTION MAPPING TABLE

| Button | MT5 Behavior | Required State | Command | Keyboard |
|--------|--------------|----------------|---------|----------|
| **New Order** | Opens order panel with current symbol | `selectedSymbol`, `currentPrice` | `OPEN_ORDER_PANEL` | F9 |
| **Bar Chart** | Switch to bar renderer | `chartType` | `SET_CHART_TYPE` | - |
| **Candlestick** | Switch to candlestick renderer | `chartType` | `SET_CHART_TYPE` | - |
| **Line Chart** | Switch to line renderer | `chartType` | `SET_CHART_TYPE` | - |
| **Indicators** | Open indicator navigator | `indicators[]` | `OPEN_INDICATOR_NAVIGATOR` | Ctrl+I |
| **Crosshair** | Toggle crosshair mode | `crosshairEnabled` | `TOGGLE_CROSSHAIR` | - |
| **Zoom In** | Increase candle density | `candleWidth` | `ZOOM_IN` | + |
| **Zoom Out** | Decrease candle density | `candleWidth` | `ZOOM_OUT` | - |
| **Tile Windows** | Arrange charts in grid | `layoutMode` | `TILE_WINDOWS` | - |
| **Cursor** | Cancel drawing mode | `activeTool` | `SELECT_CURSOR` | Esc |
| **Trendline** | Enter trendline drawing | `activeTool`, `drawingMode` | `SELECT_TRENDLINE` | - |
| **H-Line** | Enter horizontal line drawing | `activeTool`, `drawingMode` | `SELECT_HLINE` | - |
| **V-Line** | Enter vertical line drawing | `activeTool`, `drawingMode` | `SELECT_VLINE` | - |
| **Text Tool** | Enter text placement mode | `activeTool`, `drawingMode` | `SELECT_TEXT` | - |
| **M1-MN** | Load timeframe data | `timeframe`, `candleData[]` | `SET_TIMEFRAME` | - |

---

## 🏗️ COMMAND BUS SCHEMA

### **Central Command Dispatcher**

```typescript
// types/commands.ts
export type Command =
  | { type: 'OPEN_ORDER_PANEL'; payload: { symbol: string; price: { bid: number; ask: number } } }
  | { type: 'SET_CHART_TYPE'; payload: { chartType: ChartType } }
  | { type: 'OPEN_INDICATOR_NAVIGATOR'; payload: {} }
  | { type: 'TOGGLE_CROSSHAIR'; payload: {} }
  | { type: 'ZOOM_IN'; payload: {} }
  | { type: 'ZOOM_OUT'; payload: {} }
  | { type: 'TILE_WINDOWS'; payload: { mode: 'horizontal' | 'vertical' | 'grid' } }
  | { type: 'SELECT_TOOL'; payload: { tool: DrawingTool } }
  | { type: 'SET_TIMEFRAME'; payload: { timeframe: Timeframe } }
  | { type: 'ADD_INDICATOR'; payload: { name: string; params: any } }
  | { type: 'SAVE_DRAWING'; payload: { type: string; points: any[] } };

export interface CommandBus {
  dispatch(command: Command): void;
  subscribe(type: Command['type'], handler: (payload: any) => void): () => void;
}
```

### **Implementation**

```typescript
// services/commandBus.ts
class CommandBusImpl implements CommandBus {
  private handlers = new Map<string, Set<Function>>();

  dispatch(command: Command): void {
    const handlers = this.handlers.get(command.type) || new Set();
    handlers.forEach(handler => handler(command.payload));
  }

  subscribe(type: Command['type'], handler: Function): () => void {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);

    return () => {
      this.handlers.get(type)?.delete(handler);
    };
  }
}

export const commandBus = new CommandBusImpl();
```

---

## 🎮 STATE MACHINE DIAGRAM

```
┌─────────────────────────────────────────────────────────────┐
│                   TOOLBAR STATE MACHINE                      │
└─────────────────────────────────────────────────────────────┘

┌──────────┐     SELECT_CURSOR      ┌──────────┐
│  CURSOR  │ ◄─────────────────────  │ DRAWING  │
│   MODE   │                         │   MODE   │
└──────────┘   SELECT_TRENDLINE     └──────────┘
     │            SELECT_HLINE            ▲
     │            SELECT_VLINE            │
     │            SELECT_TEXT             │
     └────────────────────────────────────┘

Chart Type State:
┌──────────┐     SET_CHART_TYPE     ┌──────────┐
│   BAR    │ ◄────────────────────► │ CANDLE   │
└──────────┘                        └──────────┘
     ▲                                   ▲
     │         SET_CHART_TYPE             │
     └───────────┐            ┌───────────┘
                 ▼            ▼
             ┌──────────┐
             │   LINE   │
             └──────────┘

Crosshair State:
┌──────────┐     TOGGLE_CROSSHAIR   ┌──────────┐
│ ENABLED  │ ◄────────────────────► │ DISABLED │
└──────────┘                        └──────────┘

Zoom State:
┌──────────┐     ZOOM_IN/ZOOM_OUT   ┌──────────┐
│ CANDLE   │ ─────────────────────► │ CANDLE   │
│ WIDTH: n │                        │ WIDTH:n±1│
└──────────┘                        └──────────┘
```

---

## 🐛 BUG LIST (TOOLBAR-SPECIFIC)

### **CRITICAL BUGS**

1. **No Command Bus** - All buttons call functions directly, causing tight coupling
2. **No Drawing State** - Drawing tools don't track active tool or persist drawings
3. **Crosshair Always On** - No toggle functionality, always active
4. **Zoom = CSS Scale** - Current implementation likely uses CSS transform (wrong)
5. **No Keyboard Shortcuts** - F9, Ctrl+I, Esc, +/- not implemented
6. **No Indicator System** - Indicator button completely missing
7. **No Order Panel Integration** - New Order button has no handler
8. **No Window Management** - Tile button has no layout manager

### **HIGH PRIORITY BUGS**

9. **Chart Type Mutation** - Likely destroys chart instance instead of switching series
10. **No Tool State Persistence** - Drawings disappear on chart type change
11. **No Drawing Deletion** - No way to remove placed drawings
12. **No Drawing Edit** - Can't modify existing drawings
13. **Text Tool Missing** - No text placement logic
14. **V-Line Missing** - Vertical line button doesn't exist

---

## 🔧 STEP-BY-STEP FIX PLAN

### **Phase 1: Command Bus Foundation** (Agent 1)
**Files to Create:**
- `src/types/commands.ts` - Command type definitions
- `src/services/commandBus.ts` - Central dispatcher
- `src/hooks/useCommandBus.ts` - React hook for command dispatch

**Tasks:**
1. Define all toolbar command types
2. Implement CommandBus class with subscribe/dispatch
3. Create React context provider
4. Add event replay capability for debugging

**Acceptance Criteria:**
- All toolbar buttons dispatch commands
- No direct function calls from buttons
- Command log visible in DevTools

---

### **Phase 2: Toolbar State Management** (Agent 2)
**Files to Create:**
- `src/store/toolbarState.ts` - Toolbar state machine
- `src/hooks/useToolbarState.ts` - State management hook
- `src/services/keyboardShortcuts.ts` - Shortcut handler

**Files to Modify:**
- `src/components/layout/TopToolbar.tsx` - Connect to command bus

**Tasks:**
1. Create toolbar state reducer
2. Implement tool selection logic (cursor, trendline, h-line, v-line, text)
3. Add keyboard shortcut handlers (F9, Ctrl+I, Esc, +, -)
4. Create active tool indicator in UI

**Acceptance Criteria:**
- Only one drawing tool active at a time
- Cursor tool deselects all drawing tools
- Keyboard shortcuts work globally
- Active tool has visual highlight

---

### **Phase 3: Chart Integration** (Agent 3)
**Files to Create:**
- `src/services/chartManager.ts` - Chart state coordinator
- `src/services/drawingManager.ts` - Drawing tool logic
- `src/services/indicatorManager.ts` - Indicator attachment
- `src/components/IndicatorNavigator.tsx` - Indicator picker dialog

**Files to Modify:**
- `src/components/TradingChart.tsx` - Add drawing layer, crosshair toggle, zoom controls

**Tasks:**
1. **Crosshair Toggle**
   - Add crosshair enable/disable logic
   - Subscribe to TOGGLE_CROSSHAIR command
   - Update lightweight-charts config

2. **Zoom Controls**
   - Implement candle width adjustment (not CSS scale)
   - Subscribe to ZOOM_IN/ZOOM_OUT commands
   - Update timeScale with new bar spacing

3. **Drawing Tools**
   - Create drawing overlay layer
   - Implement trendline drawing (2 click points)
   - Implement h-line drawing (1 click point + drag)
   - Implement v-line drawing (1 click point)
   - Add drawing deletion (right-click menu)
   - Add drawing edit mode (click to select, drag to move)

4. **Indicator System**
   - Create indicator navigator dialog
   - Implement indicator categories (Trend, Oscillators, Volumes, Bill Williams)
   - Add indicator attachment to chart
   - Persist indicators across chart type changes

**Acceptance Criteria:**
- Crosshair toggles on/off
- Zoom changes candle density, not CSS scale
- Drawings persist across chart type changes
- Drawings are deletable and editable
- Indicators attach to chart and render correctly

---

### **Phase 4: Order Panel & Window Management** (Agent 4)
**Files to Create:**
- `src/components/OrderPanelDialog.tsx` - MT5-style order panel
- `src/services/windowManager.ts` - Multi-chart layout manager

**Files to Modify:**
- `src/App.tsx` - Add order panel state, window layout state

**Tasks:**
1. **New Order Button (F9)**
   - Create order panel dialog component
   - Pre-fill with current symbol, bid/ask
   - Remember last lot size
   - Subscribe to OPEN_ORDER_PANEL command
   - Add F9 keyboard shortcut

2. **Tile Windows**
   - Implement horizontal/vertical/grid layouts
   - Subscribe to TILE_WINDOWS command
   - Support multiple chart windows
   - Persist layout preference

**Acceptance Criteria:**
- F9 opens order panel
- Order panel pre-fills with current symbol/price
- Tile Windows arranges multiple charts
- Layout persists on reload

---

## ✅ REGRESSION CHECKLIST

### **Before Deployment**
- [ ] All 15 toolbar buttons have real handlers
- [ ] Command bus logs all actions
- [ ] Keyboard shortcuts work (F9, Ctrl+I, Esc, +, -)
- [ ] Only one drawing tool active at a time
- [ ] Crosshair toggles on/off
- [ ] Zoom changes candle density (not CSS)
- [ ] Drawings persist across chart type changes
- [ ] Drawings are deletable
- [ ] Drawings are editable
- [ ] Indicators attach to chart
- [ ] Order panel opens with F9
- [ ] Tile windows arranges charts

### **Visual Parity with MT5**
- [ ] Button spacing matches MT5
- [ ] Active tool has blue highlight
- [ ] Timeframe buttons have blue active state
- [ ] Chart type buttons have blue active state
- [ ] Crosshair shows OHLC tooltip
- [ ] Zoom is smooth and incremental

### **Edge Cases**
- [ ] Switching chart type preserves drawings
- [ ] Switching timeframe preserves drawings
- [ ] Deleting all drawings works
- [ ] Multiple charts tile correctly
- [ ] Order panel handles missing price data
- [ ] Keyboard shortcuts work with multiple charts

---

## 🚀 AGENT ASSIGNMENTS

### **Agent 1: Command Bus Architect**
**Scope**: `src/types/commands.ts`, `src/services/commandBus.ts`, `src/hooks/useCommandBus.ts`
**Deliverable**: Central command dispatcher with full type safety

### **Agent 2: Toolbar State Manager**
**Scope**: `src/store/toolbarState.ts`, `src/hooks/useToolbarState.ts`, `src/services/keyboardShortcuts.ts`
**Deliverable**: Tool selection state machine + keyboard shortcuts

### **Agent 3: Chart Integration Engineer**
**Scope**: `src/services/chartManager.ts`, `src/services/drawingManager.ts`, `src/components/IndicatorNavigator.tsx`
**Deliverable**: Crosshair toggle, zoom controls, drawing tools, indicator system

### **Agent 4: Order Panel & Window Management**
**Scope**: `src/components/OrderPanelDialog.tsx`, `src/services/windowManager.ts`
**Deliverable**: F9 order panel, tile windows layout

---

## 📐 ARCHITECTURE CONSTRAINTS

1. **No Direct Mutations** - All state changes via command bus
2. **Deterministic Behavior** - Command replay must produce identical state
3. **Type Safety** - All commands fully typed
4. **Single Source of Truth** - Toolbar state in one reducer
5. **Event-Driven** - Components subscribe to commands, don't call each other
6. **Performance** - Drawing layer uses canvas, not DOM elements
7. **Persistence** - Drawings survive chart type/timeframe changes
8. **MT5 Parity** - Visual and functional behavior identical to MT5

---

## 🎯 SUCCESS METRICS

- **Button Functionality**: 15/15 buttons working (currently 4/15)
- **Keyboard Shortcuts**: 5 shortcuts working (F9, Ctrl+I, Esc, +, -)
- **Drawing Tools**: 4 tools working (trendline, h-line, v-line, text)
- **Code Quality**: 100% command bus usage, zero direct calls
- **User Experience**: Zero no-op buttons, zero fake handlers

**MISSION COMPLETE WHEN**: Every toolbar button executes real logic with MT5-identical behavior.

---

**Generated**: 2026-01-20
**Status**: Ready for Swarm Deployment
**Priority**: CRITICAL - Trading terminal core functionality
