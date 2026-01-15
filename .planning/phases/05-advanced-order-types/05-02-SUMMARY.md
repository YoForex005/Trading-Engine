---
phase: 05-advanced-order-types
plan: 02
subsystem: trading-engine
tags: [trailing-stop, order-monitoring, dynamic-orders, risk-management]

# Dependency graph
requires:
  - phase: 05-01
    provides: Stop-loss/take-profit order monitoring infrastructure
provides:
  - Trailing stop order creation and management
  - Dynamic trigger price adjustment algorithm
  - Trailing stop UI controls

affects: [06-risk-management, future-phases-requiring-trailing-stops]

# Tech tracking
tech-stack:
  added: []
  patterns: [dynamic-order-adjustment, trailing-stop-algorithm]

key-files:
  created: []
  modified:
    - backend/internal/core/order_monitor.go
    - backend/internal/database/repository/order.go
    - backend/internal/core/engine.go
    - backend/internal/api/handlers/positions.go
    - backend/cmd/server/main.go
    - clients/desktop/src/components/BottomDock.tsx
    - clients/desktop/src/App.tsx

key-decisions:
  - "Trailing stop triggers move with favorable price, never against trader"
  - "BUY positions: trigger = current_ask - delta (moves up with rising price)"
  - "SELL positions: trigger = current_bid + delta (moves down with falling price)"
  - "Trailing delta validated to max 10% of position open price"
  - "UpdateTriggerPrice() persists trigger changes before checking for execution"

patterns-established:
  - "Dynamic order monitoring: check and update trailing stops before trigger execution"
  - "Price-based order adjustment: orders can modify their own triggers based on market movement"

# Metrics
duration: ~90 min
completed: 2026-01-16
---

# Phase 5 Plan 02: Trailing Stop Orders - Execution Summary

**Trailing stop orders dynamically follow favorable price movements, locking in profits while allowing positions to capture extended trends**

## Performance

- **Duration:** ~90 min
- **Started:** 2026-01-16 (afternoon)
- **Completed:** 2026-01-16 (evening)
- **Tasks:** 2/2 completed
- **Files modified:** 7

## Accomplishments

- Implemented trailing stop price adjustment algorithm in OrderMonitor
- Created UpdateTriggerPrice() repository method for persisting dynamic trigger changes
- Added CreateTrailingStop() engine method for creating trailing stop orders
- Built HandleSetTrailingStop API endpoint with validation
- Created trailing stop UI in BottomDock with "Trail" column and delta input
- Trailing stops adjust automatically as market moves favorably
- Trigger prices never move against trader's position

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement trailing stop price adjustment algorithm** - `e4d1afa` (feat)
2. **Task 2a: Add trailing stop API endpoint** - `43441ac` (feat)
3. **Task 2b: Add trailing stop frontend UI** - `8cc36c1` (feat)

**Plan metadata:** (will be committed with SUMMARY)

## Files Created/Modified

**Backend:**
- `backend/internal/core/order_monitor.go` - Added updateTrailingStops() method, calls before checkPriceTriggers()
- `backend/internal/database/repository/order.go` - Added UpdateTriggerPrice() for persisting trigger changes
- `backend/internal/core/engine.go` - Added CreateTrailingStop() method for order creation
- `backend/internal/api/handlers/positions.go` - Added HandleSetTrailingStop() endpoint
- `backend/cmd/server/main.go` - Registered /api/positions/trailing-stop route

**Frontend:**
- `clients/desktop/src/components/BottomDock.tsx` - Added "Trail" column, trailing stop input, visual controls
- `clients/desktop/src/App.tsx` - Added setTrailingStop() handler, wired to BottomDock

## Decisions Made

1. **Trigger Price Calculation**: For BUY positions, trigger = current_ask - delta (moves up with price). For SELL positions, trigger = current_bid + delta (moves down with price). This ensures trailing stops follow favorable movement only.

2. **Update Before Check Pattern**: updateTrailingStops() runs BEFORE checkPriceTriggers() in monitoring loop. This ensures trailing stops adjust to latest market price before evaluating if trigger hit.

3. **Max Delta Validation**: Trailing delta limited to 10% of position open price to prevent unreasonable distances.

4. **Repository Method for Updates**: Added UpdateTriggerPrice() instead of generic Update() to maintain single-responsibility principle in repository pattern.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - implementation proceeded smoothly using existing Order Monitor infrastructure from Plan 01.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Trailing stop infrastructure complete
- OrderMonitor handles dynamic order adjustments
- Ready for OCO (One-Cancels-Other) orders implementation
- Ready for pending order expiry checking
- Ready for Phase 6: Risk Management

## Verification Results

✅ **Build Status:**
- `go build ./backend/...` - SUCCESS
- `bun run typecheck` - PASS (only pre-existing unused variable warnings)

✅ **Must-Haves Verification:**

| Truth | Status | Evidence |
|-------|--------|----------|
| Trader can set trailing stop on position | ✅ | UI "Trail" column + API endpoint + CreateTrailingStop() |
| Trailing stop follows favorable price movement | ✅ | updateTrailingStops() adjusts trigger as market moves |
| Trailing stop triggers when price reverses by delta | ✅ | checkPriceTriggers() executes when trigger hit |
| Trailing delta persists across server restarts | ✅ | OrderRepository.UpdateTriggerPrice() persists to database |

| Artifact | Status | Evidence |
|----------|--------|----------|
| backend/internal/core/order_monitor.go contains updateTrailingStop | ✅ | Line 169: updateTrailingStops() method |
| backend/internal/core/engine.go contains trailing_delta | ✅ | Line 1211: TrailingDelta field in order creation |
| clients/desktop/src/components/BottomDock.tsx contains trailingStop | ✅ | Line 264: trailingStopId state, Line 383: Trail column |

| Key Link | Status | Evidence |
|----------|--------|----------|
| order_monitor.go → orderRepo.UpdateTriggerPrice via trigger update | ✅ | Line 214: orderRepo.UpdateTriggerPrice(ctx, order.ID, newTrigger) |

## Algorithm Example

For BUY position with trailing_delta = 0.0020:

1. **Initial**: Position opens at 1.0970, trailing stop created at 1.0950 (1.0970 - 0.0020)
2. **Price rises to 1.0990**: Trigger updates to 1.0970 (1.0990 - 0.0020)
3. **Price rises to 1.1010**: Trigger updates to 1.0990 (1.1010 - 0.0020)
4. **Price reverses to 1.0985**: Trigger stays at 1.0990, position closes when bid ≤ 1.0990

The trailing stop "locked in" 0.0020 profit by following price upward.

## Technical Implementation Details

**OrderMonitor Flow:**
1. Fetch all pending orders with trigger prices
2. Call updateTrailingStops(orders) - adjusts TRAILING_STOP order triggers
3. Call checkPriceTriggers() - executes orders whose triggers are hit
4. Repeat every 100ms

**UpdateTrailingStops Algorithm:**
```
For each TRAILING_STOP order:
  Get current market price (bid/ask)

  If BUY position:
    newTrigger = ask - trailingDelta
    Update only if newTrigger > oldTrigger (moves up)

  If SELL position:
    newTrigger = bid + trailingDelta
    Update only if newTrigger < oldTrigger (moves down)

  Persist newTrigger to database
```

**API Endpoint:**
- `PATCH /api/positions/trailing-stop`
- Request: `{ positionId, trailingDelta }`
- Validates delta > 0 and < 10% of position price
- Creates TRAILING_STOP order via engine.CreateTrailingStop()
- Returns order details with initial trigger price

---
*Phase: 05-advanced-order-types*
*Completed: 2026-01-16*
