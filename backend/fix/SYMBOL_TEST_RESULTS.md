# YoFX Symbol Discovery Test Results

**Test Date**: 2026-01-19
**Test Time**: 11:49-11:52 UTC
**Server**: 23.106.238.138:12336
**Session**: YOFX2 (Market Data Feed)

## Summary

- **Total Symbols Tested**: 11
- **Subscriptions Accepted**: 9 (81%)
- **Subscriptions Rejected**: 2 (19%)
- **Symbols Delivering Data**: 0 (0%)

## Key Findings

### 1. Market Data Request Format Fixed ✅
- **Previous Issue**: Missing tag 267 (NoMDEntryTypes)
- **Solution**: Added tag 267 with BID (269=0) and OFFER (269=1) entry types
- **Result**: Server now accepts subscriptions

### 2. Accepted Symbols (But No Data Flowing)

These symbols accept the subscription without rejection, but do NOT deliver tick data:

| Symbol | Status | Notes |
|--------|--------|-------|
| EURUSD | Accepted, No Data | Most common FX pair |
| GBPUSD | Accepted, No Data | Major FX pair |
| USDJPY | Accepted, No Data | Major FX pair |
| AUDUSD | Accepted, No Data | Major FX pair |
| USDCHF | Accepted, No Data | Major FX pair |
| USDCAD | Accepted, No Data | Major FX pair |
| EURGBP | Accepted, No Data | Cross pair |
| EURJPY | Accepted, No Data | Cross pair |
| GBPJPY | Accepted, No Data | Cross pair |

### 3. Rejected Symbols (Unknown Symbol)

| Symbol | Status | Reason |
|--------|--------|--------|
| NZDUSD | REJECTED | "Unknown symbol 'NZDUSD'" |
| XAUUSD | REJECTED | "Unknown symbol 'XAUUSD'" (Gold) |

## Analysis

### Why No Data is Flowing?

Several possible reasons:

1. **Market Closed**: Tests run at 11:49-11:52 UTC on Sunday (2026-01-19)
   - Forex market typically opens Sunday 22:00 UTC (5pm EST)
   - Tests were conducted during market close hours

2. **Server Configuration**:
   - Server accepts subscriptions for valid symbols
   - But may not have active data feeds configured
   - Or data only flows during market hours

3. **Subscription Type**:
   - Using `263=1` (Snapshot + Updates)
   - Server might require different subscription parameters

4. **Account Permissions**:
   - YOFX2 account may have limited permissions
   - May need trading account (YOFX1) for market data

## Recommendations

### Immediate Next Steps

1. **Test During Market Hours**
   - Retry tests Sunday 22:00 UTC or later
   - Monday-Friday during active trading hours

2. **Try Alternative Request Format**
   ```
   263=0  (Snapshot only)
   264=1  (Top of book instead of full depth)
   265=1  (Incremental refresh)
   ```

3. **Request Security List**
   - Send Security List Request (MsgType=x)
   - Ask server which symbols are available
   - Use the exact symbol names from server response

4. **Check With YOFX1 Account**
   - YOFX2 is "Market Data Feed"
   - But YOFX1 is "Trading Account"
   - Trading account might have better data access

### Long-term Solutions

1. **Contact YoFX Support**
   - Ask for list of available symbols
   - Confirm market data permissions for YOFX2
   - Request documentation on market data subscriptions

2. **Implement Security List Request**
   - FIX MsgType=x (Security List Request)
   - Server will return all available instruments
   - Use exact symbol names from response

3. **Add Fallback Data Source**
   - If YoFX has limited symbol coverage
   - Consider backup data provider
   - Or implement data aggregation from multiple sources

## Next Test to Run

Create a test that:
1. Runs during market hours (Sunday 22:00+ UTC or Mon-Fri)
2. Sends Security List Request to get available symbols
3. Tests with those exact symbol names
4. Monitors for 60 seconds with verbose logging
5. Captures ALL message types (not just W/X/Y)

## FIX Message Format (Corrected)

```
Market Data Request (MsgType=V):
- 262: MDReqID (unique request ID)
- 263: SubscriptionRequestType (1=Snapshot+Updates)
- 264: MarketDepth (0=Full book)
- 267: NoMDEntryTypes (2=number of entry types)
- 269: MDEntryType (0=Bid, 1=Offer)
- 146: NoRelatedSym (1=one symbol)
- 55: Symbol (e.g., "EURUSD")
```

## Conclusion

**Current Status**: ⚠️ Partial Success

- ✅ Connection to YoFX server works
- ✅ Authentication successful (YOFX2)
- ✅ Market Data Request format corrected
- ✅ 9/11 symbols accept subscription
- ❌ No live data flowing (likely due to market hours or configuration)
- ❌ NZDUSD and XAUUSD not available on this server

**Next Action**: Retry during active market hours and request Security List from server.
