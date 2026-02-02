# User Guide - LP Tag and Real Data Verification

**Date**: January 30, 2026
**Version**: 1.0
**Audience**: Traders, Brokers, Operations Teams

---

## What is the LP Tag?

The **LP Tag** is a field in every market data tick that identifies the **Liquidity Provider** (data source) for that price.

### Old System (Removed)

Before January 30, 2026, the LP tag could show:
- ❌ `"YOFX"` - Real YOFX data
- ❌ `"MOCK"` - Simulated/artificial data
- ❌ `"OANDA"` - Alternative provider (if YOFX was down)
- ❌ `"SIMULATION"` - System generated fake data

**Problem**: You could never be sure if your prices were real or simulated!

### New System (Current)

As of January 30, 2026, the LP tag **ONLY shows**:
- ✅ `"YOFX"` - **Always real market data from YOFX**

**Guarantee**: If you see a tick, it's 100% real YOFX data. Period.

---

## How to Read Market Data

### In the Desktop Terminal

**Market Watch Panel**:
```
┌─────────────────────────────────────────────────────────┐
│ Symbol  │ Bid      │ Ask      │ Spread │ H24  │ L24  │ │
├─────────────────────────────────────────────────────────┤
│ EURUSD  │ 1.1045   │ 1.1046   │ 0.0001 │ 1.10 │ 1.10 │ │
│ GBPUSD  │ 1.2678   │ 1.2679   │ 0.0001 │ 1.27 │ 1.26 │ │
│ USDJPY  │ 100.45   │ 100.46   │ 0.01   │ 101  │ 100  │ │
└─────────────────────────────────────────────────────────┘

All prices shown = REAL YOFX data (LP tag = "YOFX")
```

**New Fields**:
- **H24** (High 24h): Highest price in last 24 hours from YOFX
- **L24** (Low 24h): Lowest price in last 24 hours from YOFX

### In WebSocket API

**Real-time Tick Message**:
```json
{
  "type": "tick",
  "symbol": "EURUSD",
  "bid": 1.1045156847,
  "ask": 1.1046156847,
  "spread": 0.0001,
  "high24h": 1.1051234567,
  "low24h": 1.1032456789,
  "dailyChange": 0.0234,
  "timestamp": 1769772266771,
  "lp": "YOFX",              ← LP Tag
  "source": "real"
}
```

**What it means**:
- ✅ `lp: "YOFX"` = Price is from real YOFX feed
- ✅ `high24h` = Real 24-hour high from YOFX
- ✅ `low24h` = Real 24-hour low from YOFX
- ✅ `source: "real"` = 100% real market data

---

## Understanding the Guarantees

### What You Get

| Aspect | Guarantee | How We Verify |
|--------|-----------|---------------|
| **Data Source** | Always YOFX | LP tag always "YOFX" |
| **No Simulation** | Never simulated | Verification endpoints |
| **Real High/Low** | 24h prices real | From YOFX FIX data |
| **Fresh Data** | <50ms latency | FIX sequence tracking |
| **Accurate Prices** | Real bid-ask | Direct from YOFX |

### Why It Matters

**Old System** (Removed):
```
You place order at: Bid 1.1045
Question: Was that REAL?
Answer: Maybe... could have been simulated 🤔
```

**New System** (Current):
```
You place order at: Bid 1.1045
Question: Was that REAL?
Answer: YES - LP tag shows "YOFX" ✅
```

---

## Data Verification

### How to Verify Real Data

#### 1. Check LP Tag in Trading Platform

**In your order ticket**:
```
Symbol: EURUSD
Entry Price: 1.1045
LP: YOFX          ← Confirms real price
Status: FILLED
```

#### 2. Use Admin Verification Endpoints

**Check current data quality** (requires admin access):
```bash
curl http://api.example.com/api/diagnostics/market-data \
  -H "Authorization: Bearer $TOKEN"

# Response shows:
{
  "dataQuality": {
    "realDataPercentage": 100.0,
    "simulationDataPercentage": 0.0,
    "mockDataPercentage": 0.0,
    "dataVerification": "passed"
  }
}
```

**Verify no simulation**:
```bash
curl http://api.example.com/api/verify/no-simulation \
  -H "Authorization: Bearer $TOKEN"

# Response:
{
  "hasSimulation": false,
  "hasMockFallback": false,
  "verification": "PASSED"
}
```

#### 3. Monitor FIX Connection Status

**Check if YOFX is connected** (admin access):
```bash
curl http://api.example.com/admin/fix/status \
  -H "Authorization: Bearer $TOKEN"

# Response:
{
  "sessions": {
    "YOFX2": {
      "status": "LOGGED_IN",
      "inboundSeqNum": 123456
    }
  },
  "primarySession": "YOFX2",
  "overallStatus": "HEALTHY"
}
```

---

## Understanding High/Low 24h

### What Are High24h and Low24h?

These are **real prices from YOFX** over the last 24 hours:

**Example**:
```
EURUSD Today:
Opening price:   1.1040
Current bid:     1.1045
High (in 24h):   1.1051  ← Highest price today
Low (in 24h):    1.1032  ← Lowest price today
Close (now):     1.1045
```

### How It Helps You

**For Risk Management**:
- Know the daily trading range
- Plan your stops and targets
- Understand volatility

**For Technical Analysis**:
- See if price is near high/low
- Identify support/resistance
- Make trading decisions

**For Order Planning**:
```
If High24h = 1.1051 and Low24h = 1.1032:
Daily Range = 19 pips

You might set:
- Buy Stop: 1.1052 (above high)
- Sell Limit: 1.1031 (below low)
```

### Verification

**High/Low are real because**:
- ✅ Come directly from YOFX FIX feed
- ✅ Not calculated or estimated
- ✅ Updated every tick
- ✅ Always from same source (YOFX)

---

## Daily Change

### What is Daily Change?

**Daily Change** shows how much the price has moved since market open.

**Example**:
```
EURUSD:
Open:  1.1040
Now:   1.1045
Change: +5 pips
Daily Change: +0.0453% (shown as dailyChange: 0.0453)
```

### How It's Calculated

```
Daily Change = (High24h - Open) / Open * 100

Example:
High24h: 1.1051
Open: 1.1040
Daily Change = (1.1051 - 1.1040) / 1.1040 * 100 = 0.0996%
```

---

## Troubleshooting

### Problem: LP Tag Shows Something Other Than "YOFX"

**This should NEVER happen.**

**If it does**:
1. ✅ Screenshot the price
2. ✅ Note the exact time
3. ✅ Contact support immediately
4. ✅ Do NOT place orders until resolved

**What's likely wrong**:
- System needs restart
- YOFX connection issue
- Bug in data flow

### Problem: High/Low Seem Wrong

**Example**:
```
High24h: 1.1051
Current: 1.1045
But you saw 1.1055 earlier!

Question: Why isn't 1.1055 the high?
```

**Reasons**:
1. Time zone confusion (YOFX uses UTC)
2. Price update lag (check timestamp)
3. Different symbol (check symbol carefully)
4. Data reset at market open

**To investigate**:
```bash
# Get current market diagnostics
curl http://api.example.com/api/diagnostics/market-data \
  -H "Authorization: Bearer $TOKEN" | jq '.latestTicks.EURUSD'

# Check timestamp - is it today's data?
# Compare with your charting platform
```

### Problem: Prices Seem Different from Other Brokers

**This is normal** because:
- Different latency (your data vs. their data)
- Slight timing differences
- Rounding differences
- Network delays

**Real prices from YOFX should be within 1-2 pips of other major brokers.**

If difference is larger:
1. Check LP tag is "YOFX"
2. Check timestamp is recent
3. Verify symbol is correct
4. Contact support

---

## Best Practices

### 1. Always Check the LP Tag

**Before trading**:
- ✅ Verify LP tag = "YOFX"
- ✅ Note the exact price
- ✅ Check timestamp is recent

**Example**:
```
"Order placed at EURUSD 1.1045"
"LP: YOFX" ✅ Confirmed real price
"Timestamp: 2026-01-30T12:34:56.123Z" ✅ Fresh
```

### 2. Monitor Connection Status

**As a broker admin**:
- ✅ Check YOFX connection status regularly
- ✅ Set up alerts for disconnections
- ✅ Have fallback procedures documented

```bash
# Add to your monitoring
curl http://api.example.com/admin/fix/status | jq '.overallStatus'
# Alert if not "HEALTHY"
```

### 3. Use High/Low for Analytics

**Build trading reports using real data**:
```
Symbol    Open     High     Low    Close   Change
EURUSD    1.1040   1.1051   1.1032 1.1045  +0.045%
GBPUSD    1.2675   1.2695   1.2634 1.2678  +0.024%
```

### 4. Document Your Price Evidence

**Keep records of**:
- Price and LP tag
- Exact timestamp
- Order execution price
- High/low at that time

**Helps with**:
- Dispute resolution
- Audit trails
- Performance analysis
- Compliance verification

---

## FAQs

### Q: Can the system switch back to simulation if YOFX goes down?

**A**: No. The system will NOT fall back to simulation.

If YOFX disconnects:
- You will NOT receive fake prices
- Trading will be disabled
- You will see an error message
- We will work to restore YOFX connection

This is a feature, not a bug - it protects your trading integrity.

### Q: How often is High/Low 24h updated?

**A**: Every tick (typically 100-500 times per second).

As market moves:
- If new high is reached, `high24h` updates immediately
- If new low is reached, `low24h` updates immediately
- You always have current extremes

### Q: Is High/Low 24h reset at midnight?

**A**: It depends on your market session.

**Forex Markets**:
- 24h high/low is rolling (last 24 hours)
- Resets at midnight UTC (market open)
- NOT tied to your local midnight

**Example**:
```
Monday 5:00 PM UTC: Day starts
Tuesday 4:59 PM UTC: Still same day's high/low
Tuesday 5:00 PM UTC: Reset - new day starts
```

### Q: What if my order uses a price when YOFX is down?

**A**: Order will NOT be placed.

If YOFX connection lost:
- All new orders rejected
- Error: "Market data unavailable"
- You cannot trade
- This is intentional - protects your account

### Q: Can I see historical high/low from other days?

**A**: Yes, through the API.

```bash
# Get OHLC data for past days
curl http://api.example.com/api/market/ohlc/EURUSD/D1 \
  -H "Authorization: Bearer $TOKEN"

# Returns: [
#   {"timestamp": 1769686800000, "high": 1.1051, "low": 1.1032},
#   {"timestamp": 1769773200000, "high": 1.1055, "low": 1.1028}
# ]
```

---

## Additional Resources

### Documentation

- **API Documentation**: `API_REAL_DATA_DOCUMENTATION.md`
- **Deployment Guide**: `DEPLOYMENT_GUIDE.md`
- **Change Log**: `SIMULATION_REMOVAL_CHANGELOG.md`
- **Compliance**: `COMPLIANCE_AND_DATA_INTEGRITY.md`

### Support Channels

- **Technical Issues**: technical@example.com
- **Trading Questions**: trading@example.com
- **Compliance Questions**: compliance@example.com
- **Emergency Support**: 24/7 hotline

---

## Summary

**The LP Tag tells you where your data comes from:**

| LP Tag | Meaning | Status |
|--------|---------|--------|
| YOFX | Real YOFX market data | ✅ Current |
| MOCK | Simulated data | ❌ Removed |
| OANDA | Alternative provider | ❌ Removed |
| SIMULATION | Fake data | ❌ Removed |

**Key Takeaway**: If you see `lp: "YOFX"`, you're trading on real market prices.

---

## Document Information

- **Version**: 1.0
- **Date**: January 30, 2026
- **Status**: ✅ Ready for end-user distribution
- **Last Updated**: January 30, 2026

