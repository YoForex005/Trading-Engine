# Frontend LP Display Testing Guide

## Overview
This guide provides comprehensive testing procedures to verify that the Liquidity Provider (LP) column correctly displays "YOFX" for real market data in the Trading Engine desktop application.

## Prerequisites
- Trading Engine backend running on `http://localhost:7999`
- Desktop client running on `http://localhost:5173`
- YOFX FIX sessions (YOFX1, YOFX2) connected and logged in
- Browser DevTools open (F12) for network inspection

---

## Test 1: LP Column Display Verification

### Objective
Verify that the Market Watch displays the LP column with "YOFX" tags for all symbols.

### Steps
1. **Launch Desktop Client**
   - Open browser to `http://localhost:5173`
   - Login with credentials (if required)
   - Navigate to Market Watch panel

2. **Check LP Column Exists**
   - Verify Market Watch table has a column labeled "LP" or "Liquidity Provider"
   - Column should be visible by default (not hidden)

3. **Verify LP Values**
   - For each symbol in Market Watch, check the LP column value
   - **Expected**: "YOFX" for all symbols receiving real data
   - **Not Expected**: "SIM", "Simulated", or any simulation-related tags

4. **Document Results**

| Symbol | LP Column Value | Status | Notes |
|--------|----------------|--------|-------|
| EURUSD | YOFX           | ✅ PASS | Real data |
| GBPUSD | YOFX           | ✅ PASS | Real data |
| USDJPY | YOFX           | ✅ PASS | Real data |
| XAUUSD | YOFX           | ✅ PASS | Real data |

### Pass Criteria
- ✅ LP column is visible
- ✅ All symbols show "YOFX" as LP value
- ✅ No "SIM" or "Simulated" tags appear

### Fail Criteria
- ❌ LP column is missing or hidden
- ❌ Any symbol shows "SIM" or simulation-related LP tag
- ❌ LP column shows "undefined" or null values

---

## Test 2: Color Coding Validation

### Objective
Verify that LP tags are color-coded correctly (green for YOFX real data, red for simulation if any).

### Steps
1. **Inspect LP Cell Styling**
   - Right-click on an LP cell → Inspect Element
   - Check CSS classes applied to the cell

2. **Verify Color Scheme**
   - **YOFX** should have:
     - Text color: Green (#10b981 or similar)
     - Background: Light green (#d1fae5) or transparent
     - Font weight: Bold or semibold

3. **Check for Warning Colors**
   - If any "SIM" tag appears (testing only):
     - Text color: Red (#ef4444)
     - Background: Light red (#fee2e2)
     - Warning icon should be present

4. **Screenshot Documentation**
   - Take screenshot showing LP column with color coding
   - Save as `screenshots/lp_color_coding_YOFX.png`

### Expected Appearance
```
┌─────────┬─────────┬─────────┬────────┐
│ Symbol  │ Bid     │ Ask     │ LP     │
├─────────┼─────────┼─────────┼────────┤
│ EURUSD  │ 1.10500 │ 1.10502 │ YOFX ✓ │ ← Green text
│ GBPUSD  │ 1.25000 │ 1.25003 │ YOFX ✓ │ ← Green text
│ XAUUSD  │ 2030.50 │ 2030.60 │ YOFX ✓ │ ← Green text
└─────────┴─────────┴─────────┴────────┘
```

### Pass Criteria
- ✅ YOFX tags display in green color
- ✅ Color scheme is consistent across all rows
- ✅ Text is readable with good contrast

### Fail Criteria
- ❌ LP tags have no color coding (black/white only)
- ❌ Incorrect colors (red for YOFX)
- ❌ Poor contrast making text unreadable

---

## Test 3: Tooltip and Warning Messages

### Objective
Verify that hovering over LP tags shows appropriate tooltips with LP information.

### Steps
1. **Hover Over YOFX LP Tag**
   - Move mouse over "YOFX" in LP column
   - Wait for tooltip to appear (usually 500ms delay)

2. **Check Tooltip Content**
   - Should show: "Liquidity Provider: YOFX"
   - May include: "Real market data feed"
   - May include: "Connected" status indicator

3. **Test Multiple Symbols**
   - Hover over LP tags for different symbols
   - Verify tooltips appear consistently

4. **Check for Warning Tooltips (if simulation exists)**
   - If "SIM" tag appears, tooltip should warn:
     - "⚠️ Simulated data - Not for live trading"
     - "Switch to real LP for production use"

### Expected Tooltip
```
┌─────────────────────────────────────┐
│ 🟢 Liquidity Provider: YOFX        │
│    Real market data feed            │
│    Status: Connected                │
│    Latency: 45ms                    │
└─────────────────────────────────────┘
```

### Pass Criteria
- ✅ Tooltips appear on hover
- ✅ Content accurately describes LP status
- ✅ Tooltips display for all LP cells

### Fail Criteria
- ❌ No tooltips appear
- ❌ Tooltip content is incorrect or misleading
- ❌ Tooltips are truncated or unreadable

---

## Test 4: Filter Functionality by LP Type

### Objective
Verify that users can filter Market Watch by LP type (YOFX only, or all).

### Steps
1. **Locate LP Filter Control**
   - Check Market Watch toolbar for "Filter by LP" dropdown
   - Or right-click column header → Filter options

2. **Test Filter Options**
   - Select "YOFX only" → Should show only YOFX symbols
   - Select "All LPs" → Should show all symbols regardless of LP
   - Select "Real Data Only" → Should exclude any simulated data

3. **Verify Filter Results**
   - Count symbols before and after filtering
   - Ensure filtered results match criteria

4. **Test Clear Filter**
   - Click "Clear Filter" or select "All"
   - Verify all symbols reappear

### Filter Test Matrix

| Filter Setting | Symbols Shown | Expected Count | Status |
|----------------|---------------|----------------|--------|
| All LPs        | All symbols   | 30+            | ✅ PASS |
| YOFX Only      | Real data     | 30+            | ✅ PASS |
| Simulated Only | SIM only      | 0 (production) | ✅ PASS |

### Pass Criteria
- ✅ Filter controls are accessible and functional
- ✅ Filtering by "YOFX" shows only real data symbols
- ✅ "Real Data Only" filter excludes simulation

### Fail Criteria
- ❌ Filter controls are missing or broken
- ❌ Filtering produces incorrect results
- ❌ Cannot clear filter after applying

---

## Test 5: Real-Time Data Updates

### Objective
Verify that LP tags remain "YOFX" as market data updates in real-time.

### Steps
1. **Monitor Market Watch for 30 Seconds**
   - Observe bid/ask prices updating
   - Watch LP column values

2. **Verify LP Tag Stability**
   - LP tags should remain "YOFX" during price updates
   - No flickering between "YOFX" and "SIM"
   - No blank LP values appearing

3. **Test Symbol Subscription**
   - Add a new symbol to Market Watch (right-click → Add Symbol)
   - Verify new symbol immediately shows "YOFX" LP tag

4. **Test Reconnection Scenario**
   - Simulate connection loss (stop backend briefly)
   - Reconnect backend
   - Verify LP tags restore to "YOFX" after reconnection

### Expected Behavior
- ✅ LP tags remain stable during price updates
- ✅ New symbols get "YOFX" tag immediately
- ✅ Reconnection restores "YOFX" tags

### Pass Criteria
- ✅ No LP tag flickering or changes during normal operation
- ✅ New symbols correctly tagged as "YOFX"
- ✅ Reconnection preserves LP tag integrity

### Fail Criteria
- ❌ LP tags change from "YOFX" to "SIM" during updates
- ❌ New symbols show undefined or wrong LP
- ❌ Reconnection causes LP tags to be lost

---

## Test 6: Browser DevTools Network Inspection

### Objective
Verify that WebSocket messages contain "YOFX" LP tags in the data payload.

### Steps
1. **Open Browser DevTools** (F12)
   - Navigate to Network tab
   - Filter by "WS" (WebSocket)

2. **Inspect WebSocket Connection**
   - Find connection to `ws://localhost:7999/ws`
   - Click on connection → Messages tab

3. **Examine Tick Messages**
   - Look for messages with `"type":"tick"`
   - Check `"LP"` field in JSON payload

4. **Sample Message Verification**
```json
{
  "type": "tick",
  "symbol": "EURUSD",
  "bid": 1.10500,
  "ask": 1.10502,
  "spread": 0.00002,
  "timestamp": 1706545200000,
  "LP": "YOFX",  // ← Should be "YOFX", NOT "SIM"
  "high24h": 1.10650,
  "low24h": 1.10320
}
```

5. **Search for "SIM" in Messages**
   - Use DevTools search (Ctrl+F)
   - Search for "SIM" in WebSocket messages
   - **Expected**: No results found
   - **Fail if**: Any "SIM" LP tags found

### Pass Criteria
- ✅ All tick messages have `"LP": "YOFX"`
- ✅ No "SIM" or simulation LP tags in WebSocket stream
- ✅ WebSocket connection is stable and receiving data

### Fail Criteria
- ❌ Any tick messages contain `"LP": "SIM"`
- ❌ WebSocket connection drops frequently
- ❌ No tick messages received

---

## Test 7: Configuration API Verification

### Objective
Verify that the backend configuration API reports YOFX as the price feed LP.

### Steps
1. **Call Configuration Endpoint**
```bash
curl -s http://localhost:7999/api/config | jq
```

2. **Check Response Fields**
```json
{
  "brokerName": "Trading Engine",
  "priceFeedLP": "YOFX",     // ← Should be "YOFX"
  "priceFeedName": "YOFX Market Data Feed",
  "executionMode": "BBOOK",
  "fixStatus": {
    "YOFX1": "LOGGED_IN",
    "YOFX2": "LOGGED_IN"
  }
}
```

3. **Verify Frontend Uses This Config**
   - Open DevTools → Application → Local Storage
   - Check if `priceFeedLP` is stored as "YOFX"
   - Frontend should display this in status bar or settings

### Pass Criteria
- ✅ Config API returns `"priceFeedLP": "YOFX"`
- ✅ FIX sessions show "LOGGED_IN" status
- ✅ Frontend displays YOFX as active LP

### Fail Criteria
- ❌ Config shows "SIM" or other LP
- ❌ FIX sessions not logged in
- ❌ Frontend shows wrong LP information

---

## Manual Testing Checklist

### Pre-Testing Setup
- [ ] Backend is running and accessible
- [ ] YOFX FIX sessions are connected
- [ ] Desktop client is built and running
- [ ] Browser DevTools open for inspection

### Visual Verification
- [ ] LP column is visible in Market Watch
- [ ] All symbols show "YOFX" LP tag
- [ ] LP tags are color-coded green
- [ ] No "SIM" tags visible anywhere
- [ ] Tooltips display correct LP information

### Functional Testing
- [ ] LP filter works correctly
- [ ] New symbols get YOFX tag automatically
- [ ] LP tags remain stable during price updates
- [ ] Reconnection preserves LP tags
- [ ] WebSocket messages contain "YOFX" only

### Performance Testing
- [ ] LP column renders without lag
- [ ] No performance issues with 30+ symbols
- [ ] Tooltips appear/disappear smoothly
- [ ] Filter operations are instant

### Edge Cases
- [ ] Test with unstable connection
- [ ] Test with rapid symbol add/remove
- [ ] Test with backend restart
- [ ] Test with multiple browser tabs

---

## Screenshot Documentation

### Required Screenshots
1. **Market Watch Overview** (`market_watch_yofx.png`)
   - Full Market Watch panel showing LP column with YOFX tags

2. **LP Color Coding** (`lp_color_green.png`)
   - Close-up of LP column showing green color coding

3. **Tooltip Display** (`lp_tooltip.png`)
   - Tooltip appearing on hover over YOFX tag

4. **WebSocket Messages** (`websocket_yofx_messages.png`)
   - DevTools showing tick messages with LP: "YOFX"

5. **Configuration API** (`config_api_yofx.png`)
   - Browser or Postman showing /api/config response

### Screenshot Storage
Save all screenshots to:
```
Trading-Engine2/
  docs/
    testing/
      screenshots/
        lp_display/
          *.png
```

---

## Troubleshooting

### Issue: LP Column Shows "undefined"
**Cause**: WebSocket tick messages missing `LP` field
**Solution**:
- Check backend code: `hub.BroadcastTick()` should include LP field
- Verify FIX gateway sets `tick.LP = "YOFX"`

### Issue: LP Shows "SIM" Instead of "YOFX"
**Cause**: Simulation fallback is active
**Solution**:
- Check FIX connection status: `curl http://localhost:7999/admin/fix/status`
- Ensure YOFX2 session is LOGGED_IN
- Review backend logs for FIX connection errors

### Issue: LP Color Coding Not Working
**Cause**: CSS classes not applied
**Solution**:
- Check React component renders LP cell with className
- Verify Tailwind CSS classes for green/red colors are defined
- Check browser console for CSS errors

### Issue: Tooltips Not Appearing
**Cause**: Tooltip component not implemented or broken
**Solution**:
- Verify tooltip library is installed (e.g., Radix UI, Headless UI)
- Check React component has tooltip wrapper
- Test with browser DevTools to check for JS errors

---

## Expected vs Actual Results Template

Use this template to document test results:

| Test Case | Expected Result | Actual Result | Status | Notes |
|-----------|----------------|---------------|--------|-------|
| LP column visible | Column shows "LP" header | ✅ Column visible | PASS | - |
| LP value is YOFX | All rows show "YOFX" | ✅ All YOFX | PASS | - |
| LP color is green | Green text (#10b981) | ✅ Green color | PASS | - |
| No SIM tags | Zero "SIM" occurrences | ✅ No SIM found | PASS | - |
| Tooltip appears | Shows "YOFX" details | ✅ Tooltip works | PASS | - |
| Filter by LP | Shows filtered results | ✅ Filter works | PASS | - |
| WebSocket LP field | `"LP": "YOFX"` in JSON | ✅ YOFX in messages | PASS | - |

---

## Test Sign-Off

### Tester Information
- **Name**: _________________
- **Date**: _________________
- **Environment**: Production / Staging / Development
- **Backend Version**: _________________
- **Frontend Version**: _________________

### Test Results Summary
- **Total Test Cases**: 7
- **Passed**: _____ / 7
- **Failed**: _____ / 7
- **Blocked**: _____ / 7

### Sign-Off
- [ ] All critical tests passed
- [ ] No "SIM" tags found in production environment
- [ ] LP display meets requirements
- [ ] Ready for production deployment

**Signature**: _____________________ **Date**: __________

---

## Appendix: Quick Reference

### Key Endpoints
- Market Data Diagnostics: `GET /api/diagnostics/market-data`
- FIX Status: `GET /admin/fix/status`
- Configuration: `GET /api/config`
- WebSocket: `ws://localhost:7999/ws`

### Expected LP Values
- **Production**: "YOFX" (real data)
- **Not Allowed**: "SIM", "Simulated", "Demo"

### Color Codes
- **YOFX (Real)**: Green (#10b981)
- **SIM (Never in Prod)**: Red (#ef4444)
- **Unknown/Error**: Yellow (#f59e0b)
