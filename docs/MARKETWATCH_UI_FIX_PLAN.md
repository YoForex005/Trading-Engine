# MarketWatch UI Comprehensive Fix Plan

## Executive Summary
The MarketWatch component currently lacks a proper context menu implementation. While the component has `addSymbol` and `removeSymbol` functions, they are not properly exposed through a context menu UI. This plan details all broken features and provides a structured repair strategy.

---

## 1. BROKEN FEATURES CATEGORIZATION

### 1.1 CRITICAL - Context Menu System (COMPLETELY MISSING)
**Status**: Not implemented
**Impact**: Users cannot access any right-click functionality
**Current State**: Right-click directly removes symbols (line 242-245)

**Missing Features**:
- Hide Symbol (remove from watchlist)
- Show All Symbols (restore hidden symbols)
- Symbol Properties/Details
- Add to Chart
- Column visibility toggles
- Symbols submenu
- Grid/Auto-arrange options

### 1.2 HIGH - Add Symbol Functionality (PARTIALLY WORKING)
**Status**: Modal exists but limited UX
**Current State**:
- Modal picker works (lines 291-329)
- "Click to add..." row triggers modal (lines 268-274)
- Search functionality works

**Missing Enhancements**:
- Context menu "Add Symbol" option
- Keyboard shortcuts (Ctrl+N)
- Recent symbols list
- Favorites/Quick access

### 1.3 HIGH - Column Management (NOT IMPLEMENTED)
**Status**: Columns are hardcoded, no visibility toggles
**Current State**: All columns always visible (lines 213-229)

**Missing**:
- Column visibility state management
- Context menu "Columns" submenu
- Individual column toggles with checkmarks
- Persist column preferences to localStorage
- Reorder columns (future enhancement)

### 1.4 MEDIUM - Symbol Hide/Remove (CRUDE IMPLEMENTATION)
**Status**: Works but poor UX
**Current State**: Right-click immediately removes (no confirmation, no context menu)

**Issues**:
- No "Hide" menu option
- No "Show All" restoration
- No hidden symbols tracking
- Removes immediately without confirmation

### 1.5 LOW - Additional Context Menu Features (NOT IMPLEMENTED)
**Missing**:
- Auto Scroll toggle
- Auto Arrange toggle
- Grid toggle
- Refresh All
- Export watchlist
- Symbol grouping

---

## 2. DETAILED FIX SPECIFICATIONS

### Fix 2.1: Implement Context Menu System
**File**: `D:\Tading engine\Trading-Engine\admin\broker-admin\src\components\dashboard\MarketWatch.tsx`
**Lines to Modify**: 242-245 (replace direct removeSymbol with context menu)

**Implementation**:

```typescript
// Add state for context menu (after line 41)
const [contextMenu, setContextMenu] = useState<{ x: number; y: number; symbol: string } | null>(null);

// Replace onContextMenu handler (lines 242-245)
onContextMenu={(e) => {
    e.preventDefault();
    setContextMenu({ x: e.clientX, y: e.clientY, symbol: data.symbol });
}}

// Add context menu actions function (after removeSymbol function, ~line 129)
const getContextMenuActions = (symbol: string): ContextAction[] => [
    {
        label: 'Hide Symbol',
        onClick: () => removeSymbol(symbol),
        shortcut: 'Del'
    },
    {
        label: 'Show All Symbols',
        onClick: () => setWatchlist(DEFAULT_WATCHLIST.concat(allSymbols.map(s => s.symbol).filter(s => !DEFAULT_WATCHLIST.includes(s)))),
        separator: true
    },
    {
        label: 'Symbol Properties',
        onClick: () => console.log('Properties', symbol)
    },
    {
        label: 'Add to Chart',
        onClick: () => console.log('Add to chart', symbol),
        separator: true
    },
    {
        label: 'Symbols',
        hasSubmenu: true,
        submenu: filteredSymbols.slice(0, 10).map(spec => ({
            label: spec.symbol,
            onClick: () => addSymbol(spec.symbol)
        })).concat([
            { separator: true },
            { label: 'More Symbols...', onClick: () => setShowSymbolPicker(true) }
        ])
    },
    {
        label: 'Columns',
        hasSubmenu: true,
        submenu: COLUMN_DEFINITIONS.map(col => ({
            label: col.label,
            checked: visibleColumns[col.key],
            onClick: () => toggleColumn(col.key)
        })),
        separator: true
    },
    { label: 'Auto Scroll', onClick: () => console.log('Auto Scroll'), checked: false },
    { label: 'Auto Arrange', onClick: () => console.log('Auto Arrange'), checked: true, shortcut: 'A' },
    { label: 'Grid', onClick: () => console.log('Grid'), checked: true, shortcut: 'G' },
];

// Add context menu render (before closing div, ~line 330)
{contextMenu && (
    <ContextMenu
        x={contextMenu.x}
        y={contextMenu.y}
        onClose={() => setContextMenu(null)}
        actions={getContextMenuActions(contextMenu.symbol)}
    />
)}
```

**Dependencies**: Import `ContextMenu` and `ContextAction` from `../ui/ContextMenu`

---

### Fix 2.2: Implement Column Visibility Management
**File**: `D:\Tading engine\Trading-Engine\admin\broker-admin\src\components\dashboard\MarketWatch.tsx`
**Lines to Add**: After line 41 (state declarations)

**Implementation**:

```typescript
// Define column configuration (after line 29)
const COLUMN_DEFINITIONS = [
    { key: 'symbol', label: 'Symbol', defaultVisible: true },
    { key: 'bid', label: 'Bid', defaultVisible: true },
    { key: 'ask', label: 'Ask', defaultVisible: true },
    { key: 'spread', label: 'Spread', defaultVisible: true },
    { key: 'dailyChange', label: 'Daily Change', defaultVisible: true },
    { key: 'lp', label: 'LP', defaultVisible: false },
    { key: 'time', label: 'Time', defaultVisible: false },
    { key: 'high', label: 'High', defaultVisible: false },
    { key: 'low', label: 'Low', defaultVisible: false },
];

// Add state for visible columns (after line 41)
const [visibleColumns, setVisibleColumns] = useState<Record<string, boolean>>(() => {
    const saved = localStorage.getItem('rtx_market_watch_columns');
    if (saved) {
        try {
            return JSON.parse(saved);
        } catch {
            return Object.fromEntries(COLUMN_DEFINITIONS.map(c => [c.key, c.defaultVisible]));
        }
    }
    return Object.fromEntries(COLUMN_DEFINITIONS.map(c => [c.key, c.defaultVisible]));
});

// Persist column visibility (add useEffect after line 62)
useEffect(() => {
    localStorage.setItem('rtx_market_watch_columns', JSON.stringify(visibleColumns));
}, [visibleColumns]);

// Toggle column function (after removeSymbol, ~line 129)
const toggleColumn = (key: string) => {
    setVisibleColumns(prev => ({ ...prev, [key]: !prev[key] }));
};

// Filter active columns
const activeColumns = COLUMN_DEFINITIONS.filter(c => visibleColumns[c.key]);

// Update header grid template (line 213)
// Change from hardcoded grid-cols to dynamic based on activeColumns
className={`grid gap-0 bg-[#1E2026] border-b border-[#383A42] text-[#888] text-[9px]`}
style={{ gridTemplateColumns: activeColumns.map(c => c.key === 'symbol' ? '1fr' : c.key === 'bid' || c.key === 'ask' ? '60px' : c.key === 'spread' ? '35px' : '55px').join(' ') }}

// Update row grid template similarly (line 241)
```

**Affected Lines**: 213-229 (headers), 233-265 (rows)

---

### Fix 2.3: Enhanced Symbol Management
**File**: `D:\Tading engine\Trading-Engine\admin\broker-admin\src\components\dashboard\MarketWatch.tsx`
**Lines to Modify**: Multiple sections

**Implementation**:

```typescript
// Add hidden symbols tracking (after line 35)
const [hiddenSymbols, setHiddenSymbols] = useState<string[]>([]);

// Load hidden symbols from localStorage (in existing useEffect, ~line 44)
useEffect(() => {
    const saved = localStorage.getItem(STORAGE_KEY);
    const savedHidden = localStorage.getItem('rtx_market_watch_hidden');
    if (saved) {
        try {
            setWatchlist(JSON.parse(saved));
        } catch {
            setWatchlist(DEFAULT_WATCHLIST);
        }
    } else {
        setWatchlist(DEFAULT_WATCHLIST);
    }
    if (savedHidden) {
        try {
            setHiddenSymbols(JSON.parse(savedHidden));
        } catch {
            setHiddenSymbols([]);
        }
    }
}, []);

// Save hidden symbols (add useEffect after line 62)
useEffect(() => {
    if (hiddenSymbols.length > 0) {
        localStorage.setItem('rtx_market_watch_hidden', JSON.stringify(hiddenSymbols));
    }
}, [hiddenSymbols]);

// Update removeSymbol to track hidden (replace line 126-128)
const hideSymbol = (symbol: string) => {
    setWatchlist(watchlist.filter(s => s !== symbol));
    setHiddenSymbols([...hiddenSymbols, symbol]);
};

// Add showAllSymbols function (after hideSymbol)
const showAllSymbols = () => {
    const allAvailable = allSymbols
        .filter(s => !s.disabled)
        .map(s => s.symbol);
    setWatchlist(allAvailable);
    setHiddenSymbols([]);
};

// Update context menu to use hideSymbol instead of removeSymbol
```

---

### Fix 2.4: Add Keyboard Shortcuts
**File**: `D:\Tading engine\Trading-Engine\admin\broker-admin\src\components\dashboard\MarketWatch.tsx`
**Lines to Add**: After WebSocket hook (~line 115)

**Implementation**:

```typescript
// Add keyboard shortcuts (after line 115)
useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
        // Ctrl+N: Add new symbol
        if (e.ctrlKey && e.key === 'n') {
            e.preventDefault();
            setShowSymbolPicker(true);
        }
        // Delete: Remove selected symbol (requires selection state)
        if (e.key === 'Delete' && selectedSymbol) {
            e.preventDefault();
            hideSymbol(selectedSymbol);
        }
        // A: Toggle auto arrange
        if (e.key === 'a' || e.key === 'A') {
            console.log('Toggle auto arrange');
        }
        // G: Toggle grid
        if (e.key === 'g' || e.key === 'G') {
            console.log('Toggle grid');
        }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
}, [selectedSymbol]);

// Add selected symbol state (after line 41)
const [selectedSymbol, setSelectedSymbol] = useState<string | null>(null);

// Add click handler to symbol rows (modify line 239)
onClick={() => setSelectedSymbol(data.symbol)}
```

---

### Fix 2.5: Import ContextMenu Component
**File**: `D:\Tading engine\Trading-Engine\admin\broker-admin\src\components\dashboard\MarketWatch.tsx`
**Lines to Add**: Line 2-3

**Implementation**:

```typescript
import ContextMenu, { ContextAction } from '../ui/ContextMenu';
```

---

## 3. IMPLEMENTATION STRATEGY

### 3.1 Parallel Implementation Groups

**Group A: Context Menu Foundation (Can work independently)**
- Fix 2.5: Import ContextMenu (5 min)
- Fix 2.1: Implement basic context menu with Hide/Show All (30 min)

**Group B: Column Management (Can work independently)**
- Fix 2.2: Implement column visibility state and toggles (45 min)
- Update grid templates to be dynamic (20 min)

**Group C: Enhanced Symbol Management (Depends on Group A)**
- Fix 2.3: Add hidden symbols tracking (30 min)
- Update context menu with enhanced actions (15 min)

**Group D: User Experience Enhancements (Depends on all)**
- Fix 2.4: Add keyboard shortcuts (20 min)
- Add visual feedback for selected symbols (10 min)

### 3.2 Sequential Dependencies

```
Fix 2.5 (Import)
    ↓
Fix 2.1 (Context Menu) ──→ Fix 2.3 (Symbol Management)
    ↓                            ↓
Fix 2.2 (Columns)     ──────→ Fix 2.4 (Shortcuts)
```

### 3.3 Priority Order (Critical First)

1. **CRITICAL** - Fix 2.5 + Fix 2.1: Context menu system (35 min)
2. **HIGH** - Fix 2.2: Column management (65 min)
3. **HIGH** - Fix 2.3: Enhanced symbol management (45 min)
4. **MEDIUM** - Fix 2.4: Keyboard shortcuts (30 min)

**Total Estimated Time**: 2.9 hours

---

## 4. TESTING CHECKLIST

### 4.1 Context Menu Tests

#### Test: Hide Symbol
- [ ] Right-click on any symbol row
- [ ] Context menu appears at cursor position
- [ ] Click "Hide Symbol"
- [ ] Symbol is removed from watchlist
- [ ] Symbol is added to hidden list
- [ ] Change persists on page reload

#### Test: Show All Symbols
- [ ] Right-click on any symbol
- [ ] Click "Show All Symbols"
- [ ] All available (non-disabled) symbols appear in watchlist
- [ ] Hidden symbols list is cleared
- [ ] Change persists on page reload

#### Test: Context Menu Position
- [ ] Right-click on symbol near top of screen
- [ ] Menu appears below cursor
- [ ] Right-click on symbol near bottom of screen
- [ ] Menu appears above cursor (if space limited)
- [ ] Right-click near right edge
- [ ] Menu appears to left of cursor (if needed)

#### Test: Context Menu Close
- [ ] Open context menu
- [ ] Click anywhere outside menu
- [ ] Menu closes
- [ ] Click on another symbol row
- [ ] Previous menu closes, new menu does NOT open

#### Test: Symbols Submenu
- [ ] Right-click on any symbol
- [ ] Hover over "Symbols" menu item
- [ ] Submenu appears to the right
- [ ] Shows first 10 available symbols
- [ ] Click any symbol in submenu
- [ ] Symbol is added to watchlist
- [ ] Both menus close

### 4.2 Column Visibility Tests

#### Test: Toggle Column Visibility
- [ ] Right-click on any symbol
- [ ] Hover over "Columns" menu item
- [ ] Submenu shows all column definitions
- [ ] Visible columns have checkmarks
- [ ] Click "LP" to hide it
- [ ] LP column disappears from both header and rows
- [ ] Click "LP" again to show it
- [ ] LP column reappears
- [ ] Checkmark toggles correctly

#### Test: Column State Persistence
- [ ] Toggle several columns on/off
- [ ] Reload the page
- [ ] Column visibility matches previous state
- [ ] localStorage contains correct JSON

#### Test: All Columns Hidden Edge Case
- [ ] Hide all columns except Symbol
- [ ] Table still renders with Symbol column
- [ ] Try to hide Symbol column
- [ ] (Optional: prevent hiding last column)

#### Test: Dynamic Grid Layout
- [ ] Show all columns
- [ ] Verify grid layout is correct
- [ ] Hide "Bid" column
- [ ] Grid adjusts, remaining columns fill space
- [ ] Hide "Ask" column
- [ ] Grid adjusts again
- [ ] No layout breaks or overlaps

### 4.3 Symbol Management Tests

#### Test: Add Symbol via Modal
- [ ] Click "click to add..." row at bottom
- [ ] Modal appears centered
- [ ] Search box is auto-focused
- [ ] Type "EUR" in search
- [ ] Only EURUSD, EURGBP, etc. appear
- [ ] Click any symbol
- [ ] Symbol is added to watchlist
- [ ] Modal closes
- [ ] Search query is cleared

#### Test: Add Symbol via Context Menu
- [ ] Right-click any symbol
- [ ] Hover over "Symbols" submenu
- [ ] Click a symbol not in watchlist
- [ ] Symbol is added to watchlist
- [ ] Menu closes

#### Test: Hidden Symbols Tracking
- [ ] Note the current watchlist
- [ ] Hide 3 symbols
- [ ] Verify they're removed from view
- [ ] Check localStorage for 'rtx_market_watch_hidden'
- [ ] Verify JSON contains the 3 symbols
- [ ] Reload page
- [ ] Symbols remain hidden

#### Test: Show All Restoration
- [ ] Hide 5 symbols
- [ ] Right-click and select "Show All Symbols"
- [ ] All symbols from allSymbols (non-disabled) appear
- [ ] Hidden list is cleared
- [ ] localStorage 'rtx_market_watch_hidden' is empty or removed

### 4.4 Keyboard Shortcut Tests

#### Test: Ctrl+N Add Symbol
- [ ] Focus on MarketWatch component
- [ ] Press Ctrl+N
- [ ] Symbol picker modal opens
- [ ] Default browser "New Window" is prevented

#### Test: Delete Key Hide Symbol
- [ ] Click on a symbol row to select it
- [ ] Row highlights (selection state)
- [ ] Press Delete key
- [ ] Symbol is removed from watchlist
- [ ] No browser default action occurs

#### Test: A Key Auto Arrange
- [ ] Press 'A' or 'a' key
- [ ] Console logs "Toggle auto arrange"
- [ ] (Future: actual auto-arrange logic)

#### Test: G Key Grid Toggle
- [ ] Press 'G' or 'g' key
- [ ] Console logs "Toggle grid"
- [ ] (Future: actual grid visual change)

### 4.5 Integration Tests

#### Test: Multi-Symbol Workflow
- [ ] Start with default watchlist
- [ ] Hide 2 symbols
- [ ] Add 3 new symbols via modal
- [ ] Add 1 symbol via context menu
- [ ] Toggle 2 columns off
- [ ] Reload page
- [ ] All states persist correctly
- [ ] Watchlist matches expectations
- [ ] Column visibility matches
- [ ] Hidden symbols remain hidden

#### Test: WebSocket Data Flow
- [ ] Verify WebSocket connection (green dot)
- [ ] Add a symbol to watchlist
- [ ] Verify ticks update for new symbol
- [ ] Hide the symbol
- [ ] Verify ticks stop displaying (symbol removed)
- [ ] Show All Symbols
- [ ] Verify ticks resume for restored symbols

#### Test: Empty Watchlist Edge Case
- [ ] Hide all symbols (or start with empty watchlist)
- [ ] Verify "click to add..." row still renders
- [ ] Click to add a symbol
- [ ] Watchlist now has 1 symbol
- [ ] Data displays correctly

#### Test: All Symbols Added
- [ ] Add all available symbols to watchlist
- [ ] Open "Add Symbol" modal
- [ ] Search box should show "No symbols found" (all already added)
- [ ] Verify no duplicate additions possible

### 4.6 Visual & UX Tests

#### Test: Context Menu Styling
- [ ] Right-click on symbol
- [ ] Menu background is `#1B1B1B`
- [ ] Menu border is `#666`
- [ ] Menu shadow is visible
- [ ] Text is white on dark background
- [ ] Hover highlights item with `#3399FF`

#### Test: Checkmark Display
- [ ] Open Columns submenu
- [ ] Visible columns show checkmark (✓)
- [ ] Hidden columns show no checkmark
- [ ] Checkmark is small (10px) and bold

#### Test: Separator Lines
- [ ] Context menu has separators in correct positions
- [ ] Separator is 1px height, `#444` color
- [ ] Separators have 1px margin

#### Test: Submenu Arrow
- [ ] Hover over "Symbols" or "Columns"
- [ ] ChevronRight icon appears on right side
- [ ] Icon is `#CCC` color
- [ ] Icon is small (10px)

### 4.7 Regression Tests

#### Test: Symbol Sorting
- [ ] Add multiple symbols
- [ ] Click "Symbol" column header
- [ ] Symbols sort alphabetically A-Z
- [ ] Click again
- [ ] Symbols sort Z-A
- [ ] Sort indicator (▲/▼) updates

#### Test: Bid/Ask Sorting
- [ ] Click "Bid" column header
- [ ] Symbols sort by bid price ascending
- [ ] Click again
- [ ] Symbols sort descending
- [ ] Verify numerical sort (not string sort)

#### Test: Spread/Daily Change Sorting
- [ ] Sort by Spread column
- [ ] Verify correct numerical order
- [ ] Sort by Daily Change column
- [ ] Verify correct order with negatives

#### Test: WebSocket Connection States
- [ ] Disconnect WebSocket (stop server)
- [ ] Red dot appears in header
- [ ] Tooltip shows "Disconnected" or error
- [ ] Reconnect WebSocket (start server)
- [ ] Green dot appears
- [ ] Tooltip shows "Connected"

#### Test: Symbol Picker Search
- [ ] Open symbol picker
- [ ] Search "XAU"
- [ ] Only XAUUSD appears
- [ ] Clear search
- [ ] All symbols appear
- [ ] Search "ZZZ" (non-existent)
- [ ] "No symbols found" message

#### Test: LocalStorage Limits
- [ ] Add maximum symbols (all available)
- [ ] Verify localStorage doesn't exceed limits
- [ ] Hide many symbols
- [ ] Verify hidden list saves correctly
- [ ] Toggle all columns
- [ ] Verify column state saves correctly

---

## 5. RISK ASSESSMENT

### 5.1 Low Risk Changes
- Fix 2.5: Import statement (no logic change)
- Fix 2.4: Keyboard shortcuts (additive feature)

### 5.2 Medium Risk Changes
- Fix 2.2: Column management (changes grid layout, could break styling)
- Fix 2.3: Symbol management (modifies state management)

### 5.3 High Risk Changes
- Fix 2.1: Context menu (replaces existing behavior, user-facing change)

### 5.4 Mitigation Strategies

**For Fix 2.1 (Context Menu)**:
- Test on all screen sizes
- Test with long symbol lists
- Test submenu positioning near screen edges
- Verify no z-index conflicts with other UI elements

**For Fix 2.2 (Columns)**:
- Test with minimum columns (1)
- Test with maximum columns (all)
- Verify grid doesn't break responsive layout
- Test column persistence doesn't corrupt localStorage

**For Fix 2.3 (Symbol Management)**:
- Verify hidden symbols don't interfere with WebSocket updates
- Test edge case: hide all symbols, then show all
- Verify no memory leaks from symbol state tracking

---

## 6. ACCEPTANCE CRITERIA

### 6.1 Must Have (MVP)
- [ ] Right-click context menu appears on symbol rows
- [ ] "Hide Symbol" removes symbol from watchlist
- [ ] "Show All Symbols" restores hidden symbols
- [ ] "Columns" submenu toggles column visibility
- [ ] Column visibility persists on reload
- [ ] Watchlist state persists on reload
- [ ] No console errors
- [ ] No layout breaks

### 6.2 Should Have (V1.1)
- [ ] Keyboard shortcut Ctrl+N opens symbol picker
- [ ] Delete key hides selected symbol
- [ ] Symbols submenu shows quick-add list
- [ ] Visual feedback for selected symbol
- [ ] Auto-arrange and Grid toggles functional

### 6.3 Nice to Have (Future)
- [ ] Drag-to-reorder columns
- [ ] Custom column widths
- [ ] Export/import watchlist
- [ ] Symbol favorites/groups
- [ ] Recent symbols list
- [ ] Symbol search in context menu

---

## 7. ROLLBACK PLAN

### 7.1 Git Checkpoint
Before starting fixes:
```bash
git add admin/broker-admin/src/components/dashboard/MarketWatch.tsx
git commit -m "checkpoint: MarketWatch before context menu fixes"
```

### 7.2 Rollback Commands
If fixes cause critical issues:
```bash
# Rollback MarketWatch only
git checkout HEAD~1 -- admin/broker-admin/src/components/dashboard/MarketWatch.tsx

# Full rollback
git revert HEAD
```

### 7.3 Feature Flags (Future Enhancement)
```typescript
const ENABLE_CONTEXT_MENU = true;
const ENABLE_COLUMN_MANAGEMENT = true;
const ENABLE_KEYBOARD_SHORTCUTS = true;

// Wrap features in flags for easy disable
{ENABLE_CONTEXT_MENU && contextMenu && (
    <ContextMenu ... />
)}
```

---

## 8. PERFORMANCE CONSIDERATIONS

### 8.1 Context Menu Rendering
- Context menu only renders when active (conditional rendering)
- Uses React portals if needed for z-index management
- Closes on any outside click (single event listener)

### 8.2 Column Management
- Grid template recalculated on column toggle (negligible cost)
- localStorage writes are debounced (useEffect batches)
- No re-renders of rows when only headers change

### 8.3 Symbol Management
- Hidden symbols stored separately (no filter overhead on every render)
- Watchlist filter uses useMemo for performance
- WebSocket tick processing unchanged (no new overhead)

---

## 9. DOCUMENTATION UPDATES

### 9.1 User Documentation (Future)
- Add section: "Managing Symbols in Market Watch"
- Add section: "Customizing Columns"
- Add section: "Keyboard Shortcuts Reference"

### 9.2 Developer Documentation
- Update component props interface
- Document localStorage schema:
  ```
  rtx_market_watchlist: string[] (active symbols)
  rtx_market_watch_hidden: string[] (hidden symbols)
  rtx_market_watch_columns: Record<string, boolean> (column visibility)
  ```

---

## 10. COMPLETION CHECKLIST

- [ ] All fixes implemented
- [ ] All tests in section 4 passing
- [ ] No console errors or warnings
- [ ] No TypeScript compilation errors
- [ ] Code follows existing style conventions
- [ ] localStorage schema documented
- [ ] Git commit with descriptive message
- [ ] Pull request created (if applicable)
- [ ] QA review requested
- [ ] Product owner sign-off

---

## APPENDIX A: Code Reference Patterns

### Pattern 1: Context Menu from AccountsView (WORKING REFERENCE)
File: `D:\Tading engine\Trading-Engine\admin\broker-admin\src\components\dashboard\AccountsView.tsx`

Key elements to replicate:
- State: `const [contextMenu, setContextMenu] = useState<{ x: number; y: number; account: any } | null>(null);` (line 40)
- Handler: `const handleRightClick = (e: React.MouseEvent, acc: any) => { e.preventDefault(); ... }` (line 89-96)
- Render: `{contextMenu && <ContextMenu x={contextMenu.x} y={contextMenu.y} ... />}` (line 242-248)
- Actions: `const getMenuActions = (acc: any): ContextAction[] => [...]` (line 102-158)
- Columns submenu pattern (line 148-157)

### Pattern 2: Column Toggle from AccountsView
- State: `const [visibleColumns, setVisibleColumns] = useState<Record<string, boolean>>({...})` (line 43-49)
- Toggle: `const toggleColumn = (key: string) => { setVisibleColumns(prev => ({ ...prev, [key]: !prev[key] })); }` (line 98-100)
- Filter: `const activeColumns = COLUMNS_DEF.filter(c => visibleColumns[c.key]);` (line 75)
- Submenu: Map columns to checkboxes with `checked` property (line 152-156)

### Pattern 3: localStorage Persistence
```typescript
// Load on mount
useEffect(() => {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved) {
        try {
            setState(JSON.parse(saved));
        } catch {
            setState(DEFAULT_STATE);
        }
    }
}, []);

// Save on change
useEffect(() => {
    if (state.length > 0) {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
    }
}, [state]);
```

---

## APPENDIX B: File Locations Quick Reference

| Component | Path |
|-----------|------|
| MarketWatch (broken) | `D:\Tading engine\Trading-Engine\admin\broker-admin\src\components\dashboard\MarketWatch.tsx` |
| ContextMenu (working) | `D:\Tading engine\Trading-Engine\admin\broker-admin\src\components\ui\ContextMenu.tsx` |
| AccountsView (reference) | `D:\Tading engine\Trading-Engine\admin\broker-admin\src\components\dashboard\AccountsView.tsx` |
| OrdersView (reference) | `D:\Tading engine\Trading-Engine\admin\broker-admin\src\components\dashboard\OrdersView.tsx` |
| HistoryView (reference) | `D:\Tading engine\Trading-Engine\admin\broker-admin\src\components\dashboard\HistoryView.tsx` |

---

## APPENDIX C: ContextMenu Interface Reference

```typescript
export interface ContextAction {
    label: string;              // Menu item text
    onClick?: () => void;       // Click handler
    separator?: boolean;        // Add separator after item
    shortcut?: string;          // Keyboard shortcut display
    hasSubmenu?: boolean;       // Has child menu
    submenu?: ContextAction[];  // Child menu items
    danger?: boolean;           // Red text for destructive actions
    checked?: boolean;          // Checkmark for toggles
}

interface ContextMenuProps {
    x: number;                  // X position (clientX)
    y: number;                  // Y position (clientY)
    onClose: () => void;        // Close handler
    actions: ContextAction[];   // Menu items array
    level?: number;             // Recursion level (internal)
}
```

---

**END OF PLAN**

**Next Steps**: Implement fixes in priority order, following test-driven approach. Start with Fix 2.5 → Fix 2.1 → Fix 2.2 → Fix 2.3 → Fix 2.4.
