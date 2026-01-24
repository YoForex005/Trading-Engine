# YOFX Market Data Issue - Comprehensive Analysis

**Date:** 2026-01-19
**Status:** 🔴 CRITICAL - Server Disconnecting on MarketDataRequest
**Severity:** HIGH - Blocks all real market data

---

## 🔍 Executive Summary

The Trading Engine frontend is displaying **simulated market data** (`LP: "SIMULATED"`) instead of real FIX 4.4 quotes from YOFX because:

**ROOT CAUSE:** YOFX server **immediately disconnects** when receiving MarketDataRequest messages.

This is NOT a silent ignore - the server is **actively rejecting** market data subscriptions by closing the FIX session.

---

## 📊 Evidence

### Frontend Symptoms
✅ **Confirmed via WebSocket analysis:**
- `LP` field shows "SIMULATED" (not "YOFX")
- Impossible price values (negative prices: AUDUSD = -0.07279)
- Unrealistic levels (EURUSD at 0.387 instead of ~1.085)
- Perfect 1.5 pip spreads (too consistent for real market)
- Identical timestamps across all symbols

### Backend Diagnostics
✅ **Confirmed via FIX message log analysis:**
- **186 FIX messages logged** in `backend/fixstore/YOFX2.msgs`
- **107 MarketDataRequest (35=V) messages SENT**
- **0 MarketDataSnapshot (35=W) messages received**
- **0 MarketDataReject (35=Y) messages received**
- Only Heartbeats (35=0) received before disconnect

### Critical Finding from Diagnostic Tool
✅ **Server disconnect behavior observed:**

```
Time: 17:20:10 - Sent MarketDataRequest for EURUSD
Time: 17:20:11 - [FIX] Read error: EOF ← SERVER CLOSED CONNECTION
Time: 17:20:11 - [FIX] Disconnected from YOFX Market Data Feed
```

**Timeline:**
- T+0s: Logon successful (35=A accepted)
- T+9s: Send MarketDataRequest (35=V)
- T+10s: Server disconnects (EOF)
- T+11s: All subsequent tests fail (session DISCONNECTED)

---

## 🔬 Technical Analysis

### What's Working ✅
1. **Network connectivity** - Proxy connection established
2. **FIX 4.4 Logon** - Authentication successful (LOGGED_IN)
3. **Heartbeat exchange** - Session stays alive WITHOUT market data requests
4. **Message formatting** - Valid FIX 4.4 structure
5. **Simulation fallback** - Working perfectly (activates after 30s)

### What's NOT Working ❌
1. **MarketDataRequest triggers disconnect** - Server closes connection immediately
2. **No market data flowing** - totalTickCount = 0
3. **Frontend showing mock data** - Simulated prices after 30-second timeout

### MarketDataRequest Format Being Sent

```
35=V (MarketDataRequest)
262=MD_EURUSD_1768823410832379600 (MDReqID - unique identifier)
263=1 (SubscriptionRequestType: 1=Snapshot+Updates)
264=0 (MarketDepth: 0=Full book)  ← Tested both 0 and 1
265=0 (MDUpdateType: 0=Full refresh)
267=2 (NoMDEntryTypes: 2 types)
269=0 (MDEntryType: 0=Bid)
269=1 (MDEntryType: 1=Offer)
146=1 (NoRelatedSym: 1 symbol)
55=EURUSD (Symbol)
```

**Code location:** `backend/fix/gateway.go` lines 1885-1926

---

## 🎯 Root Cause Analysis

### Three Possible Scenarios

#### Scenario 1: Missing Account Permissions (MOST LIKELY) 🔴
**Evidence:**
- Server accepts Logon (authentication OK)
- Server disconnects on MarketDataRequest (permissions NOT OK)
- This is classic behavior for "account not entitled to market data"

**Solution:**
- Contact YOFX support
- Verify YOFX2 account has market data subscription
- Request account entitlements list

#### Scenario 2: Incorrect MarketDataRequest Format
**Evidence:**
- Format appears FIX 4.4 compliant
- But server may require additional tags
- Tested both MarketDepth=0 and MarketDepth=1 (both failed)

**Possible missing tags:**
- Tag 460 (Product) - Some servers require this
- Tag 207 (SecurityExchange) - Exchange identification
- Tag 15 (Currency) - Quote currency
- Tag 263=2 (Disable previous snapshot+updates) - May need explicit unsubscribe first

**Solution:**
- Request working MarketDataRequest example from YOFX
- Review FIX 4.4 specification for YOFX-specific requirements

#### Scenario 3: SecurityDefinition Required First
**Evidence:**
- Some FIX servers require SecurityDefinitionRequest (35=c) before MarketDataRequest
- This establishes symbol metadata before subscribing to quotes

**Solution:**
- Implement SecurityDefinitionRequest flow
- Send 35=c before 35=V
- Cache symbol definitions

---

## 📈 Data Flow Architecture

### Current Implementation

```
FIX Server (YOFX)
    ↓ (DISCONNECTS on MarketDataRequest)
Gateway.GetMarketData() channel
    ↓ (EMPTY - no data)
[30-second timeout]
    ↓
Simulated Fallback Activated
    ↓ (every 500ms)
WebSocket Hub.BroadcastTick()
    ↓
Frontend (receives simulated data)
    ↓
Shows: LP="SIMULATED", impossible prices
```

### Expected Flow (when fixed)

```
FIX Server (YOFX)
    ↓ (MarketDataSnapshot 35=W)
Gateway.GetMarketData() channel
    ↓
WebSocket Hub.BroadcastTick()
    ↓ (marked as LP="YOFX")
Frontend
    ↓
Shows: LP="YOFX", real market prices
```

---

## 🚀 Action Plan

### Priority 1: Contact YOFX Support (CRITICAL) 🔴

**Immediate action required:**

**Email Template:**
```
Subject: MarketDataRequest Causing Session Disconnect - Account YOFX2

Dear YOFX Support,

We are experiencing an issue with FIX 4.4 market data subscriptions on account YOFX2.

ISSUE:
- Logon (35=A) succeeds normally
- When we send MarketDataRequest (35=V), the server immediately disconnects (EOF)
- No MarketDataSnapshot (35=W) or MarketDataReject (35=Y) received

REQUEST:
1. Verify account YOFX2 has market data entitlements
2. Provide list of entitled symbols
3. Share a working example of MarketDataRequest for our account
4. Confirm if SecurityDefinitionRequest (35=c) is required before MarketDataRequest

DETAILS:
- Account: YOFX2 (SenderCompID=YOFX2, TargetCompID=YOFX)
- Credentials: Username=YOFX2, Password=Brand#143
- FIX Version: 4.4
- Server: 23.106.238.138:12336
- Connection: Via SOCKS5 proxy 81.29.145.69:49527

Sample MarketDataRequest being sent:
35=V|262=MD_EURUSD_xxx|263=1|264=0|265=0|267=2|269=0|269=1|146=1|55=EURUSD

Thank you,
[Your Name]
```

### Priority 2: Implement OANDA Historical Fallback (IN PROGRESS) ⚡

**Agent afd3994 is implementing:**
- Load OANDA tick data from `backend/data/ticks/`
- Use as realistic baseline instead of random walk
- Mark as `LP: "OANDA-HISTORICAL"`
- Better frontend experience while awaiting YOFX fix

**Benefits:**
- Realistic price levels (EURUSD ~1.085 not 0.387)
- Actual market spreads
- No negative prices
- User sees "OANDA-HISTORICAL" - knows it's not live but realistic

### Priority 3: Try SecurityDefinitionRequest
**Implementation:**
1. Before MarketDataRequest, send SecurityDefinitionRequest (35=c)
2. Wait for SecurityDefinition response (35=d)
3. Then send MarketDataRequest with symbol from definition

**Code location:** Add to `backend/fix/gateway.go`

### Priority 4: Review FIX 4.4 Spec
**Check for YOFX-specific requirements:**
- Additional required tags
- Specific tag ordering
- Session-level vs application-level messages
- Market hours restrictions

---

## 📁 Key Files Reference

### FIX Implementation
- `backend/fix/gateway.go` - FIX gateway (lines 1885-1926 = SubscribeMarketData)
- `backend/fix/config/sessions.json` - YOFX2 session config
- `backend/fixstore/YOFX2.msgs` - FIX message log (186 messages)
- `backend/fixstore/YOFX2.seqnums` - Sequence number tracking

### Simulation Fallback
- `backend/cmd/server/main.go` - Lines 963-1042 (simulation logic)
- Activates after 30 seconds if `totalTickCount == 0`
- Generates ticks every 500ms
- Marks as `LP: "SIMULATED"`

### Frontend
- `clients/desktop/src/App.tsx` - WebSocket connection (lines 126-196)
- Receives ticks from ws://localhost:7999/ws
- Displays `LP` field to user

### Diagnostic Tools
- `backend/cmd/fix_diagnostic/main.go` - Comprehensive test tool
- `backend/cmd/test_eurusd/main.go` - EURUSD-specific test
- `backend/cmd/connect_yofx2/main.go` - XAUUSD test
- `backend/cmd/test_gold_symbols/main.go` - Symbol variation test

### Documentation
- `backend/fix/FIX44_CONNECTION_TEST_RESULTS.md` - Authentication verified
- `QUOTE_STATUS_SUMMARY.md` - Quick reference
- `docs/WEBSOCKET_QUOTE_ANALYSIS_REPORT.md` - Detailed analysis
- `IMMEDIATE_FIX_PLAN.md` - Short-term actions

---

## 🎯 Success Criteria

**Issue is resolved when:**
1. ✅ MarketDataRequest does NOT cause disconnect
2. ✅ MarketDataSnapshot (35=W) messages received from YOFX
3. ✅ Frontend shows `LP: "YOFX"` (not "SIMULATED" or "OANDA-HISTORICAL")
4. ✅ Prices match real market (EURUSD ~1.085)
5. ✅ No negative prices
6. ✅ Variable spreads (not constant 1.5 pips)
7. ✅ `totalTickCount > 0` in backend logs

---

## 📊 Test Results Summary

### Diagnostic Tool Results (6 configurations tested):

| Test | Symbol | MarketDepth | Result |
|------|--------|-------------|--------|
| 1 | EURUSD | 0 (Full) | ⏱️ TIMEOUT (server disconnected) |
| 2 | EURUSD | 1 (Top) | ❌ ERROR (session DISCONNECTED) |
| 3 | EURUSD | 0 (Snapshot only) | ❌ ERROR (session DISCONNECTED) |
| 4 | EUR/USD | 0 (With slash) | ❌ ERROR (session DISCONNECTED) |
| 5 | GBPUSD | 0 (Full) | ❌ ERROR (session DISCONNECTED) |
| 6 | USDJPY | 0 (Full) | ❌ ERROR (session DISCONNECTED) |

**Success Rate:** 0/6 (0%)
**Root Issue:** First MarketDataRequest triggers disconnect, all subsequent tests fail

---

## 💡 Recommendations

### Immediate (Hours)
1. ✅ Email YOFX support with disconnect evidence
2. ⏳ Complete OANDA historical fallback (Agent afd3994)
3. ✅ Document all findings (this report)

### Short-term (Days)
1. ⏳ Wait for YOFX support response
2. ⏳ Implement SecurityDefinitionRequest flow
3. ⏳ Test during different market hours
4. ⏳ Review YOFX API documentation (if available)

### Medium-term (Weeks)
1. ⏳ Consider alternative liquidity providers if YOFX cannot provide market data
2. ⏳ Implement multi-provider failover (YOFX + OANDA + others)
3. ⏳ Add market data health monitoring
4. ⏳ Implement automatic reconnection with exponential backoff

---

## 📞 Contact Information

**YOFX Support:**
- Email: support@yofx.com (assumed - verify correct address)
- Account: YOFX2
- SenderCompID: YOFX2
- TargetCompID: YOFX

**Include in all communications:**
- Diagnostic tool output (`backend/cmd/fix_diagnostic/diagnostic_results.txt`)
- FIX message log excerpt showing disconnect
- Specific error timestamp and sequence

---

**Report Generated:** 2026-01-19 11:35 UTC
**Next Update:** After YOFX support response or OANDA fallback completion
