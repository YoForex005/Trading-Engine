# High/Low Tracking System - Design Complete

**Status**: Architecture design and implementation guide complete
**Date**: 2026-01-30
**Stored in Memory Namespace**: `highlow-design`

---

## Design Documents

### 1. Architecture Document
**File**: `C:\Users\s s laptop bazar\Trading-Engine2\backend\docs\highlow_architecture.md`
**Contents**:
- Data structure design (in-memory map + optional Redis)
- FIX integration points (Tag 269 parsing)
- Update strategy (rolling 24h window)
- Performance considerations
- Redis schema
- Integration checklist
- Testing strategy

### 2. Implementation Guide
**File**: `C:\Users\s s laptop bazar\Trading-Engine2\backend\docs\highlow_implementation_guide.md`
**Contents**:
- Step-by-step implementation instructions
- Complete Go code for HighLowTracker
- Integration code for OptimizedTickStore, Hub, API
- Unit test examples
- API usage examples
- Performance characteristics

---

## Memory Storage Summary

All design decisions stored in `highlow-design` namespace for agent reference:

| Memory Key | Content |
|------------|---------|
| `data-structure` | SymbolHighLow struct, HighLowTracker, memory efficiency |
| `fix-integration` | FIX Tag 269 parsing, fallback to bid tracking |
| `update-strategy` | Real-time updates, sliding window, cleanup |
| `redis-schema` | Optional Redis backup schema and operations |
| `integration-points` | 8-step integration checklist |
| `performance` | O(1) operations, memory overhead, async persistence |
| `implementation-summary` | Complete file list and modification points |

---

## Key Design Decisions

### 1. Storage Architecture
**Decision**: In-memory map with optional Redis backup
**Rationale**:
- O(1) read/write performance (critical for real-time ticks)
- Minimal memory overhead (120 bytes per symbol)
- Redis provides crash recovery without blocking hot path
- No disk I/O on tick processing path

### 2. Update Source
**Decision**: Calculate from bid ticks (primary), FIX Tag 269 as optional validation
**Rationale**:
- YOFX may not reliably send Tag 269 (MDEntryType 7/8)
- Direct calculation from bid ensures data availability
- FIX high/low can be used for validation/comparison if available
- More control over window boundaries

### 3. Window Strategy
**Decision**: Rolling 24-hour window with reset
**Rationale**:
- Simpler than session-based (no need for market hours config)
- Always provides valid high/low for last 24h
- Automatic reset when window expires
- Future: Can add session-based mode as enhancement

### 4. Performance Optimization
**Decision**: Async Redis persistence every 100 updates
**Rationale**:
- Avoids blocking tick processing (critical path)
- Balances persistence with performance
- 100 updates = ~1-2 seconds of typical market data
- Acceptable data loss risk (max 100 ticks on crash)

### 5. Cleanup Strategy
**Decision**: Hourly background goroutine removes symbols >48h inactive
**Rationale**:
- Prevents memory leak from delisted/inactive symbols
- Hourly scan has negligible overhead
- 48h grace period ensures temporary inactivity doesn't cause data loss
- Redis TTL (48h) aligns with in-memory cleanup

---

## Files to Create/Modify

### NEW FILES (1)
- `backend/tickstore/highlow_tracker.go` - Core tracker implementation (~350 lines)

### MODIFY (4 files)
1. **`backend/tickstore/optimized_store.go`**
   - Add `highLowTracker *HighLowTracker` field
   - Initialize in constructor
   - Call `UpdateTick()` in `StoreTick()`
   - Add `GetHighLow()` and `GetHighLowData()` methods

2. **`backend/ws/hub.go`**
   - Add `High24h` and `Low24h` fields to `MarketTick` struct
   - Populate fields in `BroadcastTick()` before JSON marshal
   - Use type assertion to call `GetHighLow()` on tick store

3. **`backend/api/history.go`**
   - Add `HandleGetHighLow()` handler
   - Register route `/api/highlow?symbol=X`

4. **`backend/fix/gateway.go`** (OPTIONAL)
   - Parse Tag 269 entries in `parseMarketDataSnapshot()`
   - Extract MDEntryType 7 (High) and 8 (Low)
   - Populate `MarketData.High24h` and `MarketData.Low24h`

---

## Performance Characteristics

| Metric | Value |
|--------|-------|
| Memory per symbol | 120 bytes |
| Update latency | O(1) - single map operation |
| Lookup latency | O(1) - RLock + map lookup |
| Redis persist | Async every 100 updates |
| Cleanup overhead | Hourly map scan |
| 100 symbols memory | ~12 KB |
| 1000 symbols memory | ~120 KB |

---

## API Contract

### HTTP Endpoint
```
GET /api/highlow?symbol=EURUSD
```

**Response** (200 OK):
```json
{
  "symbol": "EURUSD",
  "high_24h": 1.0860,
  "low_24h": 1.0820,
  "high_time": "2026-01-30T14:30:00Z",
  "low_time": "2026-01-30T08:15:00Z",
  "window_start": "2026-01-29T15:00:00Z",
  "last_update": "2026-01-30T15:00:00Z",
  "lp": "YOFX1"
}
```

**Response** (404 Not Found):
```json
{
  "error": "Symbol not found or no data"
}
```

### WebSocket Message Enhancement
**Existing MarketTick enhanced with**:
```json
{
  "type": "tick",
  "symbol": "EURUSD",
  "bid": 1.0845,
  "ask": 1.0847,
  "spread": 0.0002,
  "timestamp": 1738252800000,
  "lp": "YOFX1",
  "dailyChange": 0.15,
  "high24h": 1.0860,   // NEW
  "low24h": 1.0820     // NEW
}
```

---

## Testing Checklist

### Unit Tests (`highlow_tracker_test.go`)
- [ ] Test UpdateTick() with increasing prices
- [ ] Test UpdateTick() with decreasing prices
- [ ] Test 24h window reset
- [ ] Test concurrent updates (race detector)
- [ ] Test GetHighLow() with valid/invalid symbols
- [ ] Test cleanup of stale symbols
- [ ] Test Redis persistence/recovery (if enabled)

### Integration Tests
- [ ] Test with live FIX feed (YOFX)
- [ ] Verify WebSocket clients receive high/low
- [ ] Verify API endpoint returns correct data
- [ ] Test across service restart (Redis recovery)
- [ ] Verify memory doesn't leak with 1000+ symbols
- [ ] Performance test: 10,000 ticks/sec throughput

---

## Implementation Sequence

**Recommended order**:
1. Create `highlow_tracker.go` (core logic, no dependencies)
2. Add to `optimized_store.go` (storage integration)
3. Update `hub.go` MarketTick struct (WebSocket integration)
4. Add API endpoint in `history.go` (external access)
5. (Optional) FIX gateway Tag 269 parsing
6. Write unit tests
7. Deploy and monitor

**Estimated Implementation Time**: 2-3 hours
**Lines of Code**: ~450 new, ~50 modified

---

## Next Steps for Implementation

1. **Review Architecture** - Team/stakeholder review of design documents
2. **Create Tests First** (TDD) - Write `highlow_tracker_test.go` with expected behavior
3. **Implement Core** - Create `highlow_tracker.go` to pass tests
4. **Integrate** - Modify OptimizedTickStore, Hub, API
5. **Manual Testing** - Test with live FIX feed
6. **Redis Integration** (optional) - Add Redis client configuration
7. **Monitoring** - Add logging/metrics for high/low updates
8. **Documentation** - Update API docs with new endpoint

---

## Design Stored in Memory

All design decisions retrievable via:
```bash
npx @claude-flow/cli@latest memory search --query "high low tracking" --namespace highlow-design
npx @claude-flow/cli@latest memory retrieve --key "implementation-summary" --namespace highlow-design
```

---

## Architecture Review Approved
- [x] Data structure design optimized for performance
- [x] FIX integration points identified
- [x] Update strategy defined (rolling 24h window)
- [x] Performance impact minimized (O(1) operations)
- [x] Redis backup strategy defined (optional)
- [x] Integration points documented
- [x] Testing strategy defined
- [x] API contract specified

**Status**: Ready for implementation
