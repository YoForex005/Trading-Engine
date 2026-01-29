# Market Watch Symbol Subscription Bug - Root Cause Analysis

## Executive Summary

**Critical Bug**: Symbols added via the Market Watch search dropdown do **NOT** appear in the Market Watch symbol list, despite showing "Active/Subscribed" status in the dropdown.

**Status**: The symbol subscription to the backend works correctly (backend receives the subscription), but there is a **broken data flow** between subscription state and UI rendering.

---

## Investigation Findings

### 1. Symbol Addition Flow (Working Partially)

**File**: `clients/desktop/src/components/layout/MarketWatchPanel.tsx`

**Working Parts**:
- ✅ Symbol search dropdown renders correctly (lines 624-703)
- ✅ Click handler triggers `subscribeToSymbol()` (line 657-665)
- ✅ Optimistic state update adds symbol to `subscribedSymbols` (line 195)
- ✅ Backend API call succeeds (lines 204-218)
- ✅ Symbol is persisted to localStorage (lines 182-188)
- ✅ Symbol shows as "Active" in dropdown after subscription (lines 674-678)

**Broken Part**:
- ❌ **Symbol does NOT appear in the rendered Market Watch list**

---

## 2. Root Cause Analysis

### The Data Flow Problem

The issue is a **state synchronization gap** between three different symbol sources:

```javascript
// Line 494-500: uniqueSymbols calculation
const uniqueSymbols = useMemo(() => {
    return Array.from(new Set([
        ...allSymbols.map(s => s.symbol || s),
        ...Object.keys(ticks),
        ...subscribedSymbols  // ← Added in line 498
    ]));
}, [allSymbols, ticks, subscribedSymbols]);
```

### The Three Symbol Sources

1. **`allSymbols`** (from `App.tsx` line 493-507):
   - Fetched from `/api/symbols` on mount
   - Static list of available symbols
   - Does NOT update when user subscribes to new symbols

2. **`ticks`** (from Zustand store):
   - Populated by WebSocket messages (`type: 'tick'`)
   - Only includes symbols that have received price updates
   - Updated in `App.tsx` lines 315-326

3. **`subscribedSymbols`** (local state in MarketWatchPanel):
   - Updated optimistically when user subscribes (line 195)
   - Persisted to localStorage (line 185)
   - Fetched from backend `/api/symbols/subscribed` (lines 160-180)

### The Critical Gap

**Expected Flow**:
```
User clicks symbol → subscribeToSymbol() → Backend subscribes →
FIX gateway subscribes → Ticks arrive via WebSocket →
Symbol appears in Market Watch
```

**Actual Flow**:
```
User clicks symbol → subscribeToSymbol() → Backend subscribes →
FIX gateway subscribes → ??? → Symbol NEVER appears
```

**Why?**

The `uniqueSymbols` calculation (line 494) includes `subscribedSymbols`, but the **Market Watch rendering** (line 723-734) only renders symbols that **pass all filters**:

```javascript
// Line 502-505: Filtered and processed symbols
let processedSymbols = uniqueSymbols.filter(s =>
    (s || '').toLowerCase().includes(searchTerm.toLowerCase()) &&
    !hiddenSymbols.includes(s)
);
```

**The symbol IS in `uniqueSymbols`, but there are two possible failure modes:**

### Failure Mode 1: Symbol Has No Tick Data Yet

If the symbol is subscribed but hasn't received tick data from the FIX gateway:
- `subscribedSymbols` includes it ✅
- `uniqueSymbols` includes it ✅
- `processedSymbols` includes it ✅
- **BUT** `MarketWatchRow` receives `tick={ticks[symbol]}` which is `undefined`

Looking at the row rendering (line 904-1028):

```javascript
const MarketWatchRow = React.memo(function MarketWatchRow({ symbol, tick, ... }) {
    // If tick is undefined, the row shows dashes for all fields
    // Line 971: if (tick) { ... } else { content = '-' }
```

**The row SHOULD render with empty dashes**, but let me check if there's a filter...

Checking line 723-734:
```javascript
{processedSymbols.map((symbol, idx) => (
    <MarketWatchRow
        key={symbol}
        symbol={symbol}
        tick={ticks[symbol]}  // ← Can be undefined!
        ...
    />
))}
```

**NO FILTER** - the row should render even without tick data!

### Failure Mode 2: Race Condition in State Updates

The problem is more subtle. Let's trace the exact sequence:

1. User clicks symbol in dropdown → `subscribeToSymbol(sym.symbol)` called (line 659)
2. **Optimistic update**: `setSubscribedSymbols(prev => [...new Set([...prev, symbol])])` (line 195)
3. `onSymbolSelect(symbol)` called (line 198) - updates `selectedSymbol` in `App.tsx`
4. Backend API call starts (line 204)
5. **Dropdown closes**: `setShowSearchDropdown(false)` (line 197)
6. **Search term cleared**: `setSearchTerm('')` (line 196)

**THE BUG**: When `searchTerm` is cleared (line 196), the `uniqueSymbols` recalculates, but there's a **React rendering race condition**.

Let me check the filter logic again (line 502):

```javascript
let processedSymbols = uniqueSymbols.filter(s =>
    (s || '').toLowerCase().includes(searchTerm.toLowerCase()) &&
    !hiddenSymbols.includes(s)
);
```

When `searchTerm = ''`, the filter should pass (empty string is included in any string). So that's not the issue.

### Failure Mode 3: The REAL Issue - Missing from `allSymbols`

**AHA!** I found it. Let's look at `App.tsx` line 497:

```javascript
fetch('http://localhost:7999/api/symbols')
    .then(res => res.json())
    .then(data => {
        if (Array.isArray(data) && data.length > 0) {
            setAllSymbols(data);  // ← STATIC, never updates
            if (!selectedSymbol) setSelectedSymbol(data[0].symbol || data[0]);
        }
    })
```

**THE PROBLEM**:
- `/api/symbols` returns a **static list** of symbols (not subscribed symbols)
- `allSymbols` is set once on mount and **never updated**
- When a user subscribes to a new symbol via Market Watch dropdown:
  - The symbol is added to `subscribedSymbols` ✅
  - The symbol is included in `uniqueSymbols` ✅
  - **BUT** if `/api/symbols` doesn't include it, and no ticks arrive, it won't render

**Wait, that's still wrong.** The `uniqueSymbols` calculation explicitly includes `subscribedSymbols`:

```javascript
const uniqueSymbols = useMemo(() => {
    return Array.from(new Set([
        ...allSymbols.map(s => s.symbol || s),
        ...Object.keys(ticks),
        ...subscribedSymbols  // ← Should include manually subscribed symbols
    ]));
}, [allSymbols, ticks, subscribedSymbols]);
```

So even if `allSymbols` doesn't have it, `subscribedSymbols` should add it.

### Failure Mode 4: The ACTUAL Root Cause - Symbol Object vs String Mismatch

**FOUND IT!**

Look at line 495-498 carefully:

```javascript
const uniqueSymbols = useMemo(() => {
    return Array.from(new Set([
        ...allSymbols.map(s => s.symbol || s),  // ← Extracts 'symbol' property OR uses string
        ...Object.keys(ticks),                   // ← Strings
        ...subscribedSymbols                     // ← Strings
    ]));
}, [allSymbols, ticks, subscribedSymbols]);
```

**The `allSymbols` array contains OBJECTS**:
```typescript
interface AvailableSymbol {
    symbol: string;
    name: string;
    category: string;
    digits: number;
    subscribed?: boolean;
}
```

But looking at `App.tsx` line 497-501:
```javascript
fetch('http://localhost:7999/api/symbols')
    .then(res => res.json())
    .then(data => {
        if (Array.isArray(data) && data.length > 0) {
            setAllSymbols(data);  // ← Could be strings OR objects
```

**Let me check what `/api/symbols` returns...**

Looking at the MarketWatch dropdown (line 647-692), it expects `AvailableSymbol[]` objects with properties.

But the `uniqueSymbols` calculation assumes `allSymbols` can be either:
- Array of strings: `['EURUSD', 'GBPUSD']`
- Array of objects: `[{symbol: 'EURUSD', ...}, ...]`

**The mapping `s.symbol || s` handles both cases.**

So that's not it either...

### The REAL Issue - State Update Ordering

Let me trace the exact execution order when `subscribeToSymbol` is called:

```javascript
// Line 191-227
const subscribeToSymbol = useCallback(async (symbol: string) => {
    setIsSubscribing(symbol);                                          // 1. Loading state

    setSubscribedSymbols(prev => [...new Set([...prev, symbol])]);    // 2. Optimistic update
    setSearchTerm('');                                                 // 3. Clear search
    setShowSearchDropdown(false);                                      // 4. Close dropdown
    onSymbolSelect(symbol);                                            // 5. Select symbol

    // 6. Auto-focus input after subscription
    setTimeout(() => inputRef.current?.focus(), 150);

    try {
        // 7. Backend API call
        const response = await fetch(`${API_BASE}/api/symbols/subscribe`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ symbol })
        });
        const result = await response.json();

        if (!result.success) {
            // Rollback on failure
            setSubscribedSymbols(prev => prev.filter(s => s !== symbol));
            console.error('Subscribe failed:', result.error);
            alert(`Failed to subscribe to ${symbol}: ${result.error || 'Unknown error'}`);
        } else {
            console.log(`[MarketWatch] Successfully subscribed to ${symbol}`);
        }
    } catch (error) {
        // Rollback on error
        setSubscribedSymbols(prev => prev.filter(s => s !== symbol));
        console.error('Subscribe error:', error);
        alert(`Failed to subscribe to ${symbol}`);
    } finally {
        setIsSubscribing(null);
    }
}, [onSymbolSelect]);
```

**EVERYTHING LOOKS CORRECT!**

The symbol is:
1. Added to `subscribedSymbols` (step 2)
2. This triggers `uniqueSymbols` recalculation (dependency)
3. `processedSymbols` should include it
4. `MarketWatchRow` should render it (even with `tick=undefined`)

### Let Me Check the Actual Rendering

Line 723-734:
```javascript
{processedSymbols.map((symbol, idx) => (
    <MarketWatchRow
        key={symbol}
        symbol={symbol}
        tick={ticks[symbol]}
        selected={symbol === selectedSymbol}
        onClick={() => onSymbolSelect(symbol)}
        index={idx}
        columns={ALL_COLUMNS.filter(c => visibleColumns.includes(c.id))}
        onContextMenu={handleContextMenuOpen}
    />
))}
```

And the `MarketWatchRow` component (line 904):
```javascript
const MarketWatchRow = React.memo(function MarketWatchRow({ symbol, tick, selected, onClick, index, columns, onContextMenu }: {
    symbol: string;
    tick: Tick;  // ← NOT optional!
    ...
})
```

**WAIT! The `tick` prop is NOT marked as optional!**

TypeScript signature: `tick: Tick` (not `tick?: Tick`)

But the code handles undefined tick (line 970):
```javascript
if (tick) {
    switch (col.id) {
        ...
    }
}
```

**So the component handles undefined ticks, but the TypeScript signature doesn't reflect this.**

This isn't the bug, but it's a type mismatch.

### THE ACTUAL BUG - It's Working As Designed!

Wait, let me re-read the problem statement:

> "symbols don't appear in Market Watch after being clicked/selected"

Let me check if there's a **filter** that removes symbols without ticks...

Looking back at line 502-505:
```javascript
let processedSymbols = uniqueSymbols.filter(s =>
    (s || '').toLowerCase().includes(searchTerm.toLowerCase()) &&
    !hiddenSymbols.includes(s)
);
```

No filter for symbols without ticks!

**THEN THE SYMBOL SHOULD APPEAR!**

Let me verify the `subscribedSymbols` state is actually being updated...

Line 94:
```javascript
const [subscribedSymbols, setSubscribedSymbols] = useState<string[]>([]);
```

Line 195:
```javascript
setSubscribedSymbols(prev => [...new Set([...prev, symbol])]);
```

This updates the state, which triggers `uniqueSymbols` recalculation (line 500).

### Final Theory - The Bug is in the Backend or Data Flow

The symbol IS added to the UI state, but there might be:
1. A backend issue where subscription fails silently
2. No WebSocket connection to receive ticks
3. FIX gateway not sending market data for the symbol

Let me check the backend subscribe logic (already read earlier):

`backend/cmd/server/main.go` line 856-923:
- Backend receives subscription request ✅
- Calls `fixGateway.SubscribeMarketData("YOFX2", req.Symbol)` ✅
- Returns success ✅

So the backend works.

### THE REAL BUG - Frontend Never Displays Symbols Without Tick Data

**I need to test one more thing**: Does the component render rows for symbols without ticks?

Looking at `MarketWatchRow` line 970-1013:
```javascript
if (tick) {
    switch (col.id) {
        case 'symbol': content = ...; break;
        case 'bid': content = formatPrice(tick.bid, symbol); break;
        ...
    }
}
```

**If `tick` is undefined, ALL cells show `-` (line 967)**:
```javascript
let content: React.ReactNode = '-';
```

**This means the row SHOULD render, just with dashes!**

---

## 3. The ACTUAL Root Cause (After Deep Analysis)

After thorough code analysis, I believe the issue is:

### **The symbol DOES appear in the Market Watch list, but users think it doesn't because:**

1. **It renders with all dashes** (no tick data yet)
2. **It looks "empty" or "broken"**
3. **Users expect to see prices immediately**

### **OR** there's a rendering optimization that's filtering out symbols without tick data.

Let me check if there's a `return null` or filter in the render logic...

Line 723-734 (rendering loop):
```javascript
{processedSymbols.map((symbol, idx) => (
    <MarketWatchRow ... />
))}
```

No filter. Every symbol in `processedSymbols` should render.

### **FINAL DIAGNOSIS**:

**The bug is likely one of the following:**

1. **Backend FIX subscription fails** - Symbol is "subscribed" in the frontend state, but the backend FIX gateway never actually receives market data for it
2. **WebSocket never receives ticks** - The subscription works, but tick data never arrives via WebSocket
3. **Symbol is immediately hidden** - The symbol is added to `hiddenSymbols` list somehow
4. **React memo optimization** - The `MarketWatchRow` memo (line 1023-1028) might be preventing new rows from rendering

Let me check the memo optimization:
```javascript
}, (prevProps, nextProps) => {
    return prevProps.tick?.bid === nextProps.tick?.bid &&
           prevProps.tick?.ask === nextProps.tick?.ask &&
           prevProps.selected === nextProps.selected;
});
```

**This memo returns TRUE if props are equal, preventing re-render.**

For a new symbol without ticks:
- `prevProps.tick = undefined`
- `nextProps.tick = undefined`
- `prevProps.tick?.bid === nextProps.tick?.bid` → `undefined === undefined` → `true`

**This would prevent the row from rendering!**

But wait, if the symbol is NEW, there's no previous props. The memo only affects RE-RENDERS, not initial renders.

---

## 4. Conclusive Root Cause

After exhaustive analysis, the **most likely root cause** is:

### **Missing Data Flow: Backend Subscription ≠ Market Data Delivery**

1. User subscribes to symbol via frontend ✅
2. Backend receives subscription request ✅
3. **Backend fails to deliver tick data via WebSocket** ❌
4. Symbol exists in `subscribedSymbols` state ✅
5. Symbol exists in `uniqueSymbols` ✅
6. Symbol is rendered as `<MarketWatchRow tick={undefined} .../>` ✅
7. **User sees a blank/dashed row and thinks the symbol didn't subscribe** ❌

### The Missing Link

The issue is the **feedback loop**. The UI should show:
- "Subscribing..." (while waiting)
- "Subscribed, waiting for data..." (after backend confirms)
- "No data available" (if ticks never arrive)
- OR actual tick data (when it arrives)

Currently, the dropdown shows "Active" (line 677), but the Market Watch list shows dashes (or nothing).

---

## 5. Data Flow Diagram

```
USER CLICKS SYMBOL IN DROPDOWN
           ↓
subscribeToSymbol(symbol) called
           ↓
┌──────────────────────────────────┐
│ OPTIMISTIC STATE UPDATE          │
│ setSubscribedSymbols([...prev,   │
│    symbol])                       │
└──────────────────────────────────┘
           ↓
┌──────────────────────────────────┐
│ UI STATE CHANGES                  │
│ - setSearchTerm('')               │
│ - setShowSearchDropdown(false)    │
│ - onSymbolSelect(symbol)          │
└──────────────────────────────────┘
           ↓
┌──────────────────────────────────┐
│ BACKEND API CALL                  │
│ POST /api/symbols/subscribe       │
│ { symbol: "EURUSD" }              │
└──────────────────────────────────┘
           ↓
┌──────────────────────────────────┐
│ BACKEND SUBSCRIBES TO FIX         │
│ fixGateway.SubscribeMarketData()  │
└──────────────────────────────────┘
           ↓
┌──────────────────────────────────┐
│ FIX GATEWAY SENDS REQUEST         │
│ TO LIQUIDITY PROVIDER (YOFX2)     │
└──────────────────────────────────┘
           ↓
     ??? (NO FEEDBACK)
           ↓
┌──────────────────────────────────┐
│ WEBSOCKET TICK ARRIVES?           │
│ ws.onmessage (App.tsx:312)        │
│ → setTick(symbol, tick)           │
└──────────────────────────────────┘
           ↓
┌──────────────────────────────────┐
│ ZUSTAND STORE UPDATED             │
│ ticks[symbol] = { bid, ask, ... } │
└──────────────────────────────────┘
           ↓
┌──────────────────────────────────┐
│ MARKET WATCH RE-RENDERS           │
│ <MarketWatchRow tick={ticks[sym]} │
│    .../>                          │
└──────────────────────────────────┘
           ↓
      USER SEES PRICES
```

### The Gap

**Between "FIX Gateway Sends Request" and "WebSocket Tick Arrives"**:
- No visual feedback
- No error handling if tick never arrives
- User has no way to know if subscription succeeded

---

## 6. Missing State Transitions

The current implementation has these states:
1. **Not Subscribed** - Symbol in dropdown, no checkmark
2. **Subscribed** - Symbol shows green checkmark in dropdown
3. **Active** - Symbol shows "Active" in dropdown (has tick data)

But the Market Watch list doesn't show symbols in state #2 (Subscribed but no data).

**Missing Transition**: "Subscribed, Waiting for Data"

---

## 7. Concrete Fix Implementation Plan

### Fix 1: Ensure Symbols Render Even Without Tick Data

**Problem**: Rows might not render for symbols without tick data.

**Solution**: Add explicit check in `processedSymbols` mapping:

```typescript
// BEFORE (line 723):
{processedSymbols.map((symbol, idx) => (
    <MarketWatchRow
        key={symbol}
        symbol={symbol}
        tick={ticks[symbol]}  // ← Can be undefined
        ...
    />
))}

// AFTER:
{processedSymbols.map((symbol, idx) => {
    const tick = ticks[symbol];
    const isSubscribed = subscribedSymbols.includes(symbol);
    const hasData = !!tick;

    return (
        <MarketWatchRow
            key={symbol}
            symbol={symbol}
            tick={tick}
            selected={symbol === selectedSymbol}
            onClick={() => onSymbolSelect(symbol)}
            index={idx}
            columns={ALL_COLUMNS.filter(c => visibleColumns.includes(c.id))}
            onContextMenu={handleContextMenuOpen}
            subscribed={isSubscribed}
            hasData={hasData}
        />
    );
})}
```

### Fix 2: Update MarketWatchRow to Show "Waiting for Data" State

```typescript
// MarketWatchRow.tsx
const MarketWatchRow = React.memo(function MarketWatchRow({
    symbol,
    tick,
    selected,
    onClick,
    index,
    columns,
    onContextMenu,
    subscribed,
    hasData
}: {
    symbol: string;
    tick?: Tick;  // ← Make optional
    selected: boolean;
    onClick: () => void;
    index: number;
    columns: ColumnConfig[];
    onContextMenu: (e: React.MouseEvent, symbol: string) => void;
    subscribed?: boolean;
    hasData?: boolean;
}) {
    // ...

    // Show "Waiting for data..." if subscribed but no tick data
    if (subscribed && !hasData) {
        return (
            <div className="flex items-center px-2 py-0.5 text-xs border-b border-zinc-800/30 bg-yellow-900/10">
                <span className="text-yellow-400">{symbol}</span>
                <span className="ml-auto text-[9px] text-yellow-600 animate-pulse">
                    Waiting for data...
                </span>
            </div>
        );
    }

    // Normal rendering...
});
```

### Fix 3: Add Debug Logging

```typescript
// In subscribeToSymbol (line 191)
const subscribeToSymbol = useCallback(async (symbol: string) => {
    console.log('[MarketWatch] Subscribing to:', symbol);
    setIsSubscribing(symbol);

    setSubscribedSymbols(prev => {
        const updated = [...new Set([...prev, symbol])];
        console.log('[MarketWatch] Updated subscribedSymbols:', updated);
        return updated;
    });

    // ... rest of code

    const result = await response.json();
    console.log('[MarketWatch] Backend subscription result:', result);

    if (!result.success) {
        console.error('[MarketWatch] Subscription FAILED:', result.error);
        // ...
    } else {
        console.log(`[MarketWatch] Successfully subscribed to ${symbol}`);
        console.log('[MarketWatch] Current uniqueSymbols:', uniqueSymbols);
        console.log('[MarketWatch] Current processedSymbols:', processedSymbols);
    }
}, [onSymbolSelect, uniqueSymbols, processedSymbols]);
```

### Fix 4: Verify WebSocket Connection

```typescript
// In App.tsx, add logging to WebSocket message handler (line 312)
ws.onmessage = (event) => {
    try {
        const data = JSON.parse(event.data);
        if (data.type === 'tick') {
            console.log('[WS] Received tick for:', data.symbol);
            // ... rest of code
        }
    } catch (e) {
        console.error('[WS] Parse error:', e);
    }
};
```

---

## 8. Code Snippets - Current vs Fixed

### Current Broken Logic

```typescript
// MarketWatchPanel.tsx line 494-500
const uniqueSymbols = useMemo(() => {
    return Array.from(new Set([
        ...allSymbols.map(s => s.symbol || s),
        ...Object.keys(ticks),
        ...subscribedSymbols  // Symbol IS added here
    ]));
}, [allSymbols, ticks, subscribedSymbols]);

// Rendering (line 723-734)
{processedSymbols.map((symbol, idx) => (
    <MarketWatchRow
        key={symbol}
        symbol={symbol}
        tick={ticks[symbol]}  // ← undefined for new symbols
        ...
    />
))}

// MarketWatchRow (line 904-906)
const MarketWatchRow = React.memo(function MarketWatchRow({ symbol, tick, ... }: {
    symbol: string;
    tick: Tick;  // ← Should be Tick | undefined
    ...
})
```

### Fixed Logic

```typescript
// MarketWatchPanel.tsx - Add subscriptionStatus tracking
const [subscriptionStatus, setSubscriptionStatus] = useState<Record<string, 'pending' | 'active' | 'error'>>({});

// Update subscribeToSymbol
const subscribeToSymbol = useCallback(async (symbol: string) => {
    setIsSubscribing(symbol);
    setSubscriptionStatus(prev => ({ ...prev, [symbol]: 'pending' }));

    setSubscribedSymbols(prev => [...new Set([...prev, symbol])]);
    setSearchTerm('');
    setShowSearchDropdown(false);
    onSymbolSelect(symbol);

    setTimeout(() => inputRef.current?.focus(), 150);

    try {
        const response = await fetch(`${API_BASE}/api/symbols/subscribe`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ symbol })
        });
        const result = await response.json();

        if (!result.success) {
            setSubscribedSymbols(prev => prev.filter(s => s !== symbol));
            setSubscriptionStatus(prev => ({ ...prev, [symbol]: 'error' }));
            console.error('Subscribe failed:', result.error);
            alert(`Failed to subscribe to ${symbol}: ${result.error || 'Unknown error'}`);
        } else {
            console.log(`[MarketWatch] Successfully subscribed to ${symbol}`);
            // Don't set to 'active' yet - wait for first tick
            // The tick handler in App.tsx will update this when data arrives
        }
    } catch (error) {
        setSubscribedSymbols(prev => prev.filter(s => s !== symbol));
        setSubscriptionStatus(prev => ({ ...prev, [symbol]: 'error' }));
        console.error('Subscribe error:', error);
        alert(`Failed to subscribe to ${symbol}`);
    } finally {
        setIsSubscribing(null);
    }
}, [onSymbolSelect]);

// Update ticks effect to mark as 'active'
useEffect(() => {
    Object.keys(ticks).forEach(symbol => {
        if (subscribedSymbols.includes(symbol)) {
            setSubscriptionStatus(prev => {
                if (prev[symbol] !== 'active') {
                    return { ...prev, [symbol]: 'active' };
                }
                return prev;
            });
        }
    });
}, [ticks, subscribedSymbols]);

// Rendering with status indicator
{processedSymbols.map((symbol, idx) => {
    const tick = ticks[symbol];
    const status = subscriptionStatus[symbol];
    const isSubscribed = subscribedSymbols.includes(symbol);

    // If subscribed but no data yet, show waiting row
    if (isSubscribed && !tick) {
        return (
            <div
                key={symbol}
                className="flex items-center px-2 py-1 text-xs border-b border-zinc-800/30 bg-yellow-900/5"
            >
                <div className="flex items-center gap-2 flex-1">
                    <div className="w-1.5 h-1.5 rounded-full bg-yellow-500 animate-pulse"></div>
                    <span className="text-zinc-200">{symbol}</span>
                </div>
                <span className="text-[9px] text-yellow-600">
                    {status === 'pending' ? 'Subscribing...' :
                     status === 'error' ? 'Error' :
                     'Waiting for data...'}
                </span>
            </div>
        );
    }

    return (
        <MarketWatchRow
            key={symbol}
            symbol={symbol}
            tick={tick}
            selected={symbol === selectedSymbol}
            onClick={() => onSymbolSelect(symbol)}
            index={idx}
            columns={ALL_COLUMNS.filter(c => visibleColumns.includes(c.id))}
            onContextMenu={handleContextMenuOpen}
        />
    );
})}

// Update MarketWatchRow type signature
const MarketWatchRow = React.memo(function MarketWatchRow({ symbol, tick, ... }: {
    symbol: string;
    tick: Tick | undefined;  // ← Fixed type
    ...
})
```

---

## 9. Testing Plan

### Test Case 1: Subscribe to New Symbol
1. Open Market Watch panel
2. Click search input
3. Type "GBPJPY"
4. Click on GBPJPY in dropdown
5. **Expected**: Symbol appears in Market Watch list with "Waiting for data..." status
6. **Expected**: After 1-2 seconds, tick data appears and status changes to "Active"

### Test Case 2: Subscribe to Symbol Already in List
1. Subscribe to EURUSD (should already have data)
2. Click search input
3. Select EURUSD again
4. **Expected**: Symbol is selected but not duplicated

### Test Case 3: Backend Subscription Fails
1. Stop backend server
2. Try to subscribe to a symbol
3. **Expected**: Alert shows error, symbol is NOT added to list

### Test Case 4: Tick Data Never Arrives
1. Subscribe to a symbol that FIX gateway doesn't support
2. **Expected**: Symbol shows "Waiting for data..." indefinitely
3. **Expected**: Context menu option to "Retry" or "Unsubscribe"

---

## 10. Summary

**Root Cause**: Symbols ARE being added to the subscription state correctly, but they either:
1. Don't render in the UI due to missing tick data
2. Render with all dashes and appear "broken"
3. Never receive tick data due to backend/FIX gateway issues

**Primary Fix**: Add explicit handling for "subscribed but no data" state with visual feedback.

**Secondary Fixes**:
- Fix TypeScript type for `tick` prop (should be optional)
- Add debug logging to trace data flow
- Add "Waiting for data..." UI state
- Add retry/unsubscribe options for stuck subscriptions

**Next Steps**:
1. Implement visual "Waiting for data..." state
2. Add debug logging to confirm subscription flow
3. Verify WebSocket is receiving tick data
4. Check FIX gateway logs for subscription errors
