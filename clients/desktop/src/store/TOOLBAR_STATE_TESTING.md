# Toolbar State Manager - Testing Guide

## Implementation Status

**Agent 2 - COMPLETE**

All deliverables have been implemented:
- ✅ `clients/desktop/src/store/toolbarState.ts` - State machine with reducer
- ✅ `clients/desktop/src/hooks/useToolbarState.ts` - React hook with command bus integration
- ✅ `clients/desktop/src/services/keyboardShortcuts.ts` - Keyboard shortcut handlers
- ✅ `clients/desktop/src/components/layout/TopToolbar.tsx` - Modified with command bus integration

## Testing Instructions

### 1. Start the Development Server

```bash
cd clients/desktop
npm run dev
```

### 2. Open Browser DevTools

Press `F12` to open Chrome DevTools and navigate to the **Console** tab.

### 3. Test Toolbar Buttons

Click each button and verify console output:

#### Chart Type Buttons
- **Click Bar Chart** → Should log: `[COMMAND BUS] Command dispatched: { type: 'SET_CHART_TYPE', payload: { chartType: 'bar' } }`
- **Click Candlestick** → Should log: `[COMMAND BUS] Command dispatched: { type: 'SET_CHART_TYPE', payload: { chartType: 'candlestick' } }`
- **Click Line Chart** → Should log: `[COMMAND BUS] Command dispatched: { type: 'SET_CHART_TYPE', payload: { chartType: 'area' } }`

#### Tool Buttons
- **Click Crosshair** → Should log: `[COMMAND BUS] Command dispatched: { type: 'TOGGLE_CROSSHAIR', payload: {} }`
- **Click Zoom In** → Should log: `[COMMAND BUS] Command dispatched: { type: 'ZOOM_IN', payload: {} }`
- **Click Zoom Out** → Should log: `[COMMAND BUS] Command dispatched: { type: 'ZOOM_OUT', payload: {} }`
- **Click Templates** → Should log: `[COMMAND BUS] Command dispatched: { type: 'OPEN_TEMPLATE_MANAGER', payload: {} }`

#### Drawing Tools
- **Click Cursor** → Should log: `[COMMAND BUS] Command dispatched: { type: 'SELECT_TOOL', payload: { tool: 'cursor' } }`
- **Click Trendline** → Should log: `[COMMAND BUS] Command dispatched: { type: 'SELECT_TOOL', payload: { tool: 'trendline' } }`
- **Click Horizontal Line** → Should log: `[COMMAND BUS] Command dispatched: { type: 'SELECT_TOOL', payload: { tool: 'hline' } }`
- **Click Text Tool** → Should log: `[COMMAND BUS] Command dispatched: { type: 'SELECT_TOOL', payload: { tool: 'text' } }`
- **Click ChevronDown (More Tools)** → Should log: `[COMMAND BUS] Command dispatched: { type: 'OPEN_DRAWING_TOOLS_MENU', payload: {} }`

#### Trading Actions
- **Click Algo Trading** → Should log: `[COMMAND BUS] Command dispatched: { type: 'TOGGLE_ALGO_TRADING', payload: {} }`
- **Click New Order** → Should log: `[COMMAND BUS] Command dispatched: { type: 'OPEN_ORDER_PANEL', payload: { symbol: 'XAUUSD', price: { bid: 0, ask: 0 } } }`

#### Timeframe Buttons
- **Click M1** → Should log: `[COMMAND BUS] Command dispatched: { type: 'SET_TIMEFRAME', payload: { timeframe: 'm1' } }`
- **Click H1** → Should log: `[COMMAND BUS] Command dispatched: { type: 'SET_TIMEFRAME', payload: { timeframe: 'h1' } }`
- **Click D1** → Should log: `[COMMAND BUS] Command dispatched: { type: 'SET_TIMEFRAME', payload: { timeframe: 'd1' } }`

### 4. Test Keyboard Shortcuts

Press each key and verify console output:

#### F9 - Open Order Panel
- **Press F9** → Should log:
  ```
  [KEYBOARD] F9 pressed - Open Order Panel
  [COMMAND BUS] Command dispatched: { type: 'OPEN_ORDER_PANEL', payload: { symbol: 'XAUUSD', price: { bid: 0, ask: 0 } } }
  ```

#### Ctrl+I - Open Indicator Navigator
- **Press Ctrl+I** → Should log:
  ```
  [KEYBOARD] Ctrl+I pressed - Open Indicator Navigator
  [COMMAND BUS] Command dispatched: { type: 'OPEN_INDICATOR_NAVIGATOR', payload: {} }
  ```

#### Escape - Cancel Drawing Mode
- **Press Escape** → Should log:
  ```
  [KEYBOARD] Escape pressed - Cancel drawing mode
  [COMMAND BUS] Command dispatched: { type: 'SELECT_TOOL', payload: { tool: 'cursor' } }
  ```

#### + - Zoom In
- **Press +** or **=** → Should log:
  ```
  [KEYBOARD] + pressed - Zoom In
  [COMMAND BUS] Command dispatched: { type: 'ZOOM_IN', payload: {} }
  ```

#### - - Zoom Out
- **Press -** or **_** → Should log:
  ```
  [KEYBOARD] - pressed - Zoom Out
  [COMMAND BUS] Command dispatched: { type: 'ZOOM_OUT', payload: {} }
  ```

#### Tool Selection Shortcuts
- **Press C** → Toggle Crosshair
- **Press T** → Select Trendline Tool
- **Press H** → Select Horizontal Line Tool
- **Press V** → Select Vertical Line Tool (not yet implemented in UI)
- **Press X** → Select Text Tool

#### Other Shortcuts (Future Implementation)
- **Press Ctrl+Z** → Undo
- **Press Ctrl+Y** → Redo
- **Press Delete** → Delete selected drawing
- **Press Ctrl+A** → Select all drawings

### 5. Test State Machine Logic

#### Only One Tool Active at a Time
1. Click **Trendline** button
2. Verify it becomes highlighted
3. Click **Horizontal Line** button
4. Verify **Trendline** is no longer highlighted
5. Verify **Horizontal Line** is highlighted

#### Cursor Deselects All Tools
1. Click **Trendline** button
2. Verify it becomes highlighted
3. Click **Cursor** button
4. Verify **Trendline** is no longer highlighted
5. Verify no drawing tools are highlighted

#### Escape Cancels Drawing Mode
1. Click **Text Tool** button
2. Verify it becomes highlighted
3. Press **Escape**
4. Verify **Text Tool** is no longer highlighted

### 6. Test Visual Feedback

#### Active State Styling
- **Active Chart Type**: Should have blue background (`bg-blue-600`)
- **Active Tool**: Should have blue glow (`bg-blue-500/10 border border-blue-500/20`)
- **Active Timeframe**: Should have blue background (`bg-blue-600`)

### 7. Test Crosshair Toggle
1. Click **Crosshair** button
2. Verify button becomes highlighted
3. Click **Crosshair** button again
4. Verify button is no longer highlighted
5. Verify state toggles correctly in console

## Expected Console Output on Page Load

```
[KEYBOARD] Registering keyboard shortcuts with input check
[TOOLBAR STATE] Subscribed to command bus
[COMMAND BUS] Subscribed
```

## State Machine Validation

The toolbar state reducer handles these action types:

1. **SELECT_TOOL** - Only one drawing tool active at a time
2. **SET_CHART_TYPE** - Updates chart type (bar, candlestick, area)
3. **SET_TIMEFRAME** - Updates active timeframe
4. **TOGGLE_CROSSHAIR** - Toggles crosshair on/off
5. **ZOOM** - Adjusts candle width (min: 2px, max: 50px)
6. **ADD_INDICATOR** - Adds indicator to chart (prevents duplicates)
7. **REMOVE_INDICATOR** - Removes indicator by name
8. **ADD_DRAWING** - Adds drawing to chart
9. **DELETE_DRAWING** - Removes drawing by ID
10. **CLEAR_DRAWINGS** - Clears all drawings

## Integration with Command Bus

**Current Status**: Using placeholder command bus implementation

**Placeholder Implementation**:
- `useCommandBus()` hook in `TopToolbar.tsx`
- Logs all commands to console
- Returns `dispatch` and `subscribe` functions

**When Agent 1 Completes**:
1. Replace placeholder `useCommandBus()` in `TopToolbar.tsx` with import from Agent 1's implementation
2. Replace placeholder in `useToolbarState.ts` with import from Agent 1's implementation
3. Update type imports to use proper command types

## Acceptance Criteria Checklist

- ✅ Only one drawing tool can be active at a time
- ✅ Cursor tool deselects all drawing tools
- ✅ F9 opens order panel (command dispatched)
- ✅ Ctrl+I triggers indicator navigator (command dispatched)
- ✅ Escape cancels drawing mode
- ✅ +/- zoom in/out (commands dispatched)
- ✅ All button clicks dispatch commands (no direct function calls)
- ✅ Visual active state for selected tools
- ✅ Keyboard shortcuts respect input focus (won't trigger while typing)

## Known Limitations

1. **Vertical Line Tool**: Keyboard shortcut (V) implemented but no UI button yet
2. **Command Bus**: Using placeholder until Agent 1 completes
3. **Order Panel**: Command dispatches but panel doesn't exist yet
4. **Indicator Navigator**: Command dispatches but navigator doesn't exist yet
5. **Template Manager**: Command dispatches but manager doesn't exist yet
6. **Undo/Redo**: Commands dispatch but no implementation yet

## Next Steps

**Waiting for Agent 1**:
- Command bus implementation (`clients/desktop/src/services/commandBus.ts`)
- Replace placeholder implementations in `TopToolbar.tsx` and `useToolbarState.ts`

**Future Agents**:
- Agent 3: Order Panel component
- Agent 4: Indicator Navigator component
- Agent 5: Template Manager component
- Agent 6: Drawing tools implementation (canvas interaction)
- Agent 7: Undo/Redo history manager

## File Locations

```
clients/desktop/src/
├── store/
│   └── toolbarState.ts (NEW - State machine and reducer)
├── hooks/
│   └── useToolbarState.ts (NEW - React hook with command bus integration)
├── services/
│   └── keyboardShortcuts.ts (NEW - Keyboard shortcut handlers)
└── components/
    └── layout/
        └── TopToolbar.tsx (MODIFIED - Command bus integration)
```

## Performance Notes

- State updates are batched by React
- Keyboard event listeners registered once on mount
- Cleanup functions properly unregister listeners on unmount
- Console logging can be disabled in production builds
