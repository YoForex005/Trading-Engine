# MT5 Parity Implementation - Integration Test Plan

## Status: WAITING FOR AGENTS 1, 2, 3

**Current State (as of 2026-01-20):**
- ❌ Agent 1: Backend throttling config - NOT STARTED
- ❌ Agent 2: Flash animations - NOT STARTED
- ❌ Agent 3: State consolidation - NOT STARTED
- ⏳ Agent 4: Waiting for other agents to complete

---

## Integration Test Strategy

### Phase 1: Backend Verification (Agent 1)

**File to Check:** `D:\Tading engine\Trading-Engine\backend\ws\hub.go`

#### Expected Changes:
```go
// NEW: MT5 mode flag (around line 47)
type Hub struct {
    // ... existing fields ...
    mt5Mode bool // Broadcast all ticks without throttling
}

// NEW: Environment variable support (in NewHub or main.go)
func NewHub() *Hub {
    h := &Hub{
        // ... existing initialization ...
        mt5Mode: os.Getenv("MT5_MODE") == "true",
    }
    return h
}

// MODIFIED: BroadcastTick method (around line 141)
func (h *Hub) BroadcastTick(tick *MarketTick) {
    atomic.AddInt64(&h.ticksReceived, 1)

    // Storage happens first (unchanged)
    if h.tickStore != nil {
        h.tickStore.StoreTick(...)
    }

    // NEW: Skip throttling if MT5 mode enabled
    if h.mt5Mode {
        // Broadcast all ticks immediately
        data, err := json.Marshal(tick)
        if err != nil {
            return
        }
        select {
        case h.broadcast <- data:
            atomic.AddInt64(&h.ticksBroadcast, 1)
        default:
            // Buffer full - drop tick
        }
        return
    }

    // Existing throttling logic continues for normal mode
    // ...
}
```

#### Test Cases:

**Test 1.1: Normal Mode (Default)**
```bash
# Start server without MT5_MODE
cd backend
go run cmd/server/main.go

# Expected: Throttling active (60-80% reduction)
# Check logs for: "throttled=%d (60-80%)"
```

**Test 1.2: MT5 Mode (Enabled)**
```bash
# Set environment variable and restart
export MT5_MODE=true  # Linux/Mac
# OR
set MT5_MODE=true     # Windows CMD
$env:MT5_MODE="true"  # Windows PowerShell

go run cmd/server/main.go

# Expected: More ticks broadcast (closer to 100% of ticks received)
# Check logs for: "throttled=%d (0-10%)"
```

**Test 1.3: API Configuration Endpoint (Optional)**
```bash
# Check if MT5 mode is exposed via API
curl http://localhost:7999/api/config

# Expected response:
{
  "brokerName": "RTX Trading System",
  "mt5Mode": true,  // NEW FIELD
  "priceFeedLP": "YOFX",
  "executionMode": "B-Book (Internal)",
  // ...
}
```

---

### Phase 2: Frontend Flash Animation Verification (Agent 2)

**File to Check:** `D:\Tading engine\Trading-Engine\clients\desktop\src\components\layout\MarketWatchPanel.tsx`

#### Expected Changes:

```typescript
// NEW: Flash state map (add near line 76)
const [flashStates, setFlashStates] = useState<Record<string, 'up' | 'down' | null>>({});

// NEW: useEffect to detect price changes and trigger flashes
useEffect(() => {
    Object.entries(ticks).forEach(([symbol, tick]) => {
        if (tick.prevBid !== undefined && tick.bid !== tick.prevBid) {
            const direction = tick.bid > tick.prevBid ? 'up' : 'down';

            // Set flash state
            setFlashStates(prev => ({ ...prev, [symbol]: direction }));

            // Clear flash after 200ms
            setTimeout(() => {
                setFlashStates(prev => ({ ...prev, [symbol]: null }));
            }, 200);
        }
    });
}, [ticks]);

// MODIFIED: MarketWatchRow component (around line 836)
function MarketWatchRow({ symbol, tick, selected, onClick, index, columns, onContextMenu }: {
    symbol: string;
    tick: Tick;
    selected: boolean;
    onClick: () => void;
    index: number;
    columns: ColumnConfig[];
    onContextMenu: (e: React.MouseEvent, symbol: string) => void;
}) {
    // NEW: Get flash state
    const flashState = flashStates[symbol];

    // NEW: Flash CSS classes
    const flashClass = flashState === 'up'
        ? 'animate-flash-green'
        : flashState === 'down'
        ? 'animate-flash-red'
        : '';

    return (
        <div
            onClick={onClick}
            onContextMenu={(e) => onContextMenu(e, symbol)}
            className={`flex items-center px-2 py-0.5 cursor-pointer text-xs font-medium border-b border-zinc-800/30
                ${selected ? 'bg-[#2d3436] text-white' : index % 2 === 0 ? 'bg-[#1e1e1e]' : 'bg-[#232323]'}
                hover:bg-[#2d3436]/80 hover:text-white transition-colors
                ${flashClass}  // NEW: Flash animation
            `}
        >
            {/* ... rest of row ... */}
        </div>
    );
}
```

#### Expected CSS (in globals.css or component):
```css
@keyframes flash-green {
    0% { background-color: rgba(74, 222, 128, 0.5); }
    100% { background-color: transparent; }
}

@keyframes flash-red {
    0% { background-color: rgba(248, 113, 113, 0.5); }
    100% { background-color: transparent; }
}

.animate-flash-green {
    animation: flash-green 200ms ease-out;
}

.animate-flash-red {
    animation: flash-red 200ms ease-out;
}
```

#### Test Cases:

**Test 2.1: Visual Flash Detection**
1. Open `http://localhost:5174`
2. Select an active symbol (EURUSD)
3. Watch bid/ask prices change
4. **Expected:** Green flash on price increase, red flash on price decrease
5. **Duration:** Flash should last ~200ms

**Test 2.2: Performance Check**
```bash
# Open Chrome DevTools
# Navigate to Performance tab
# Start recording
# Watch ticks update for 10 seconds
# Stop recording

# Expected:
# - 60 FPS maintained (16.67ms frame time)
# - No layout thrashing
# - Smooth animations without jank
```

**Test 2.3: Flash Color Accuracy**
- Green flash (`rgba(74, 222, 128, 0.5)`) when bid increases
- Red flash (`rgba(248, 113, 113, 0.5)`) when bid decreases
- No flash when price unchanged

---

### Phase 3: State Consolidation Verification (Agent 3)

**Files to Check:**
1. `D:\Tading engine\Trading-Engine\clients\desktop\src\App.tsx`
2. `D:\Tading engine\Trading-Engine\clients\desktop\src\components\layout\MarketWatchPanel.tsx`

#### Expected Changes in App.tsx:

```typescript
// REMOVED: Line 63 (local ticks state)
// const [ticks, setTicks] = useState<Record<string, Tick>>({});

// MODIFIED: Line 162-207 (WebSocket flush logic)
const flushTicks = () => {
    const buffer = tickBuffer.current;
    if (Object.keys(buffer).length > 0) {
        // REMOVED: setTicks(prev => ({ ...prev, ...buffer }));

        // KEPT: Update Zustand store only
        Object.entries(buffer).forEach(([symbol, tick]) => {
            useAppStore.getState().setTick(symbol, tick);
        });

        tickBuffer.current = {};
    }
    rafId = requestAnimationFrame(flushTicks);
};

// REMOVED: ticks prop from MarketWatchPanel (around line 348)
<MarketWatchPanel
    className="flex-1 min-h-0"
    // ticks={ticks} // REMOVED - now uses Zustand
    allSymbols={allSymbols}
    selectedSymbol={selectedSymbol}
    onSymbolSelect={setSelectedSymbol}
/>
```

#### Expected Changes in MarketWatchPanel.tsx:

```typescript
// MODIFIED: Props interface (around line 34)
interface MarketWatchPanelProps {
    // ticks: Record<string, Tick>;  // REMOVED
    allSymbols: any[];
    selectedSymbol: string;
    onSymbolSelect: (symbol: string) => void;
    className?: string;
}

// NEW: Import Zustand store (around line 4)
import { useAppStore } from '../../store/useAppStore';

// NEW: Get ticks from Zustand (around line 76)
export const MarketWatchPanel: React.FC<MarketWatchPanelProps> = ({
    // ticks,  // REMOVED from props
    allSymbols,
    selectedSymbol,
    onSymbolSelect,
    className
}) => {
    // NEW: Get ticks from global store
    const ticks = useAppStore(state => state.ticks);

    // ... rest of component unchanged ...
};
```

#### Test Cases:

**Test 3.1: Compile Check**
```bash
cd clients/desktop
npm run build

# Expected:
# - No TypeScript errors
# - No unused prop warnings
# - Clean build output
```

**Test 3.2: Runtime Verification**
```bash
# Open React DevTools
# Navigate to Components tab
# Select <App> component

# Expected:
# - NO "ticks" state in App component
# - MarketWatchPanel receives NO "ticks" prop
```

**Test 3.3: State Sync Verification**
```bash
# Open React DevTools
# Navigate to Components tab
# Find Zustand store provider

# Expected:
# - Ticks object populated with symbols
# - Updates in real-time
# - No duplicate state in App
```

**Test 3.4: Functional Test**
1. Open http://localhost:5174
2. Login
3. Select symbol (EURUSD)
4. **Expected:** Prices update normally
5. **Verify:** No console errors
6. **Check:** Flash animations still work (dependent on Agent 2)

---

## Phase 4: Full Integration Testing

### Test 4.1: End-to-End MT5 Mode

**Scenario:** Enable MT5 mode and verify complete pipeline

```bash
# 1. Start backend with MT5 mode
cd backend
export MT5_MODE=true
go run cmd/server/main.go

# 2. Start frontend
cd ../clients/desktop
npm run dev

# 3. Open browser to http://localhost:5174
# 4. Login
# 5. Select active symbol (EURUSD)

# Expected Results:
# ✅ More frequent tick updates (less throttling)
# ✅ Flash animations appear on every price change
# ✅ No duplicate state (single source in Zustand)
# ✅ 60 FPS maintained
# ✅ Backend logs show low throttle rate (0-10%)
```

### Test 4.2: Performance Metrics

**Backend Metrics:**
```bash
# Check server logs for stats (printed every 60 seconds)
# Example output:
[Hub] Stats: received=10000, broadcast=9500, throttled=500 (5.0% reduction), clients=1

# MT5 Mode: Expect throttle < 10%
# Normal Mode: Expect throttle > 60%
```

**Frontend Metrics:**
```javascript
// Open Chrome DevTools Console
// Measure tick update rate
let tickCount = 0;
const startTime = Date.now();

// Hook into Zustand updates (run in console)
const unsubscribe = useAppStore.subscribe(
    state => state.ticks,
    (ticks) => {
        tickCount++;
        const elapsed = (Date.now() - startTime) / 1000;
        console.log(`Ticks/sec: ${(tickCount / elapsed).toFixed(2)}`);
    }
);

// Let run for 30 seconds, then:
// unsubscribe();

// Expected:
// - Normal mode: 10-20 ticks/sec per symbol
// - MT5 mode: 30-60 ticks/sec per symbol
```

### Test 4.3: Memory & CPU Check

**Tools:**
- Chrome DevTools Performance tab
- Task Manager (Windows) / Activity Monitor (Mac)

**Procedure:**
1. Record 30 seconds of activity
2. Check memory usage (should be stable, no leaks)
3. Check CPU usage (should be < 30% for single client)
4. Check FPS (should maintain 60 FPS)

**Expected Results:**
- Memory: Stable (no continuous growth)
- CPU: < 30% average
- FPS: 60 FPS (16.67ms frame time)
- Network: WebSocket active, no reconnects

---

## Phase 5: MT5 Parity Assessment

### Feature Checklist

| Feature | Priority | Status | MT5 Parity |
|---------|----------|--------|------------|
| Dynamic spread calculation | P0 | ✅ DONE (Agent 4 - prev) | 100% |
| Symbol-aware spread formatting | P0 | ✅ DONE (Agent 4 - prev) | 100% |
| Optimistic UI updates | P0 | ✅ DONE (Agent 3 - prev) | 100% |
| Keyboard navigation | P0 | ✅ DONE (Agent 1 - prev) | 100% |
| 60 FPS tick rendering | P1 | ✅ DONE (Agent 2 - prev) | 100% |
| Reactive symbol list | P0 | ✅ DONE (Agent 3 - prev) | 100% |
| Backend throttling config | P2 | ⏳ PENDING (Agent 1) | TBD |
| Flash animations | P2 | ⏳ PENDING (Agent 2) | TBD |
| Single source of truth | P2 | ⏳ PENDING (Agent 3) | TBD |

### Parity Calculation

**Current Status:**
- Completed features: 6/9 (67%)
- Pending features: 3/9 (33%)
- Previous parity estimate: ~85%

**Expected After Integration:**
- All 9 features complete: 9/9 (100%)
- **Final parity estimate: 90-95%**

**Remaining gaps (outside scope):**
- Advanced order types (pending orders)
- Chart drawing tools
- Custom indicators
- Expert Advisors (EA) support
- Strategy tester

---

## Phase 6: Known Issues & Recommendations

### Potential Issues

**Issue 1: Flash Animation Overhead**
- **Symptom:** FPS drops below 60 when many symbols flash simultaneously
- **Solution:** Limit flash animations to visible rows only (virtual scrolling)

**Issue 2: MT5 Mode Memory Usage**
- **Symptom:** Higher memory usage with MT5 mode enabled
- **Solution:** Implement tick buffer size limit (keep last 1000 ticks per symbol)

**Issue 3: WebSocket Buffer Overflow**
- **Symptom:** Buffer full warnings in backend logs
- **Solution:** Increase buffer size or implement backpressure

### Recommendations

**Short Term (Next Sprint):**
1. Add MT5 mode toggle in UI (Settings panel)
2. Add performance metrics dashboard
3. Implement virtual scrolling for large symbol lists

**Medium Term (Next Release):**
1. Add tick history playback (for testing flash animations)
2. Implement tick rate limiter (client-side)
3. Add comprehensive E2E tests

**Long Term (Future):**
1. Migrate to WebSocket binary protocol (reduce bandwidth)
2. Implement tick compression (reduce memory)
3. Add server-side symbol filtering (reduce unnecessary broadcasts)

---

## Acceptance Criteria

### Must Have (Blocker)
- [ ] All 3 agents completed successfully
- [ ] Backend compiles without errors
- [ ] Frontend compiles without errors
- [ ] No TypeScript/Go errors in code
- [ ] WebSocket connection stable
- [ ] Ticks update in real-time

### Should Have (Important)
- [ ] MT5 mode configurable via environment variable
- [ ] Flash animations visible on price changes
- [ ] Single state source verified (no duplicate ticks)
- [ ] Performance acceptable (60 FPS, <30% CPU)

### Nice to Have (Optional)
- [ ] MT5 mode toggle in UI
- [ ] Performance metrics exposed via API
- [ ] Automated E2E tests

---

## Deliverables

1. **Integration Test Results** (this document, updated after testing)
2. **Final MT5 Parity Report** (with percentage and breakdown)
3. **Performance Metrics Summary** (tick rate, FPS, CPU, memory)
4. **Known Issues & Recommendations** (for product backlog)

---

## Next Steps

**For Agent 4 (Me):**
1. Wait for Agents 1, 2, 3 to complete
2. Run all integration tests documented above
3. Generate final parity report
4. Document any issues found
5. Create summary for stakeholders

**For Other Agents:**
- **Agent 1:** Implement backend throttling config
- **Agent 2:** Implement flash animations
- **Agent 3:** Remove duplicate state, consolidate to Zustand

---

**Status:** ⏳ WAITING FOR AGENTS 1, 2, 3 TO COMPLETE

**Last Updated:** 2026-01-20

**Agent 4 (Integration & Testing Specialist)**
