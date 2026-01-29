# MT5 Visual Comparison & Requirements

**Visual Reference for Charting Implementation**

---

## Color Palette

### MT5 Standard Colors
```
Bullish Candle:  #14b8a6 (Teal)     ████████
Bearish Candle:  #ef4444 (Red)      ████████
Volume Bars:     #06b6d4 (Cyan)     ████████
Grid Lines:      #27272a (Gray)     ────────
BID Line:        #ef4444 (Red)      --------
ASK Line:        #14b8a6 (Teal)     --------
SL Line:         #ef4444 (Red)      ─ ─ ─ ─
TP Line:         #10b981 (Green)    ─ ─ ─ ─
Entry Line:      #3b82f6 (Blue)     --------
Pending Order:   #f59e0b (Orange)   --------
```

### Current System Colors (WRONG)
```
Bullish Candle:  #10b981 (Emerald)  ████████  ❌ Should be #14b8a6
Bearish Candle:  #ef4444 (Red)      ████████  ✅ Correct
Volume Bars:     N/A (Not rendered) ████████  ❌ Missing
```

---

## Layout Structure

### MT5 Chart Layout
```
┌─────────────────────────────────────────────────────────┐
│ Symbol: XAUUSD    Timeframe: M5                      [×]│
├─────────────────────────────────────────────────────────┤
│                                                          │
│  Price Scale                                             │
│  │                                                       │
│  ├─ 2720.80                                             │
│  │     │                                                │
│  │     │  ╔═╗                ╔═╗                       │
│  │     │  ║█║  ╔═╗           ║█║                       │
│  ├─────┼──╫█╫──╫█╫───────────╫█╫────── Grid (dotted)   │
│  │     │  ║█║  ║█║           ║█║                       │
│  │     │  ╚═╝  ╚═╝           ╚═╝                       │
│  ├─ 2720.40 ─────────────────────────── BID Line       │
│  │                                                       │
│  ├─ 2720.36 ─────────────────────────── ASK Line       │
│  │                                                       │
│  ├─ 2720.00  ········ BUY 0.05 @ 4607.33 ········       │ ← Trade Level
│  │           ─ ─ ─ ─  SL @ 4600.00  ─ ─ ─ ─            │ ← Stop Loss
│  │           ─ ─ ─ ─  TP @ 4650.00  ─ ─ ─ ─            │ ← Take Profit
│  │                                                       │
│  └─ 2719.60                                             │
│                                                          │
├─────────────────────────────────────────────────────────┤
│ Volume (20% of chart height)                            │
│    ▅  ▃  ▇  ▅  ▂  ▄  ▆  ▃  ▇  ▅  ▂  ▄  ▆  ▃  ▇       │ ← Cyan bars
│    │  │  │  │  │  │  │  │  │  │  │  │  │  │  │        │
├─────────────────────────────────────────────────────────┤
│ Time Scale: 10:00  10:05  10:10  10:15  10:20  10:25   │
└─────────────────────────────────────────────────────────┘
```

### Current System Layout (Missing Volume)
```
┌─────────────────────────────────────────────────────────┐
│ XAUUSD    M5                                          [×]│
├─────────────────────────────────────────────────────────┤
│                                                          │
│  Price Scale                                             │
│  │                                                       │
│  ├─ 2720.80                                             │
│  │     │                                                │
│  │     │  ╔═╗                ╔═╗                       │
│  │     │  ║█║  ╔═╗           ║█║                       │
│  ├─────┼──╫█╫──╫█╫───────────╫█╫────── Grid (solid)    │
│  │     │  ║█║  ║█║           ║█║                       │
│  │     │  ╚═╝  ╚═╝           ╚═╝                       │
│  │                                                       │
│  ├─ 2720.00  ········ #123 BUY 0.05 ········ (hover)   │ ← Shows ID on hover
│  │                                                       │
│  └─ 2719.60                                             │
│                                                          │
│                                                          │
│  ❌ NO VOLUME BARS ❌                                    │ ← Missing!
│                                                          │
│                                                          │
├─────────────────────────────────────────────────────────┤
│ Time Scale: 10:00  10:05  10:10  10:15  10:20  10:25   │
└─────────────────────────────────────────────────────────┘
```

---

## Candlestick Anatomy

### MT5 Candlestick (Bullish)
```
     High ─────┐
               │  ← Wick (thin, #14b8a6)
        ┌──────┤
        │      │  ← Body (filled, #14b8a6)
        │ OPEN │
        │      │
        │      │
        │ CLOSE│
        └──────┤
               │  ← Wick (thin, #14b8a6)
      Low ─────┘

Width: 80% of bar spacing
Wick: 1px line
Border: 1px outline
```

### Current System (Wrong Color)
```
     High ─────┐
               │  ← Wick (thin, #10b981) ❌ WRONG
        ┌──────┤
        │      │  ← Body (filled, #10b981) ❌ WRONG
        │ OPEN │
        │      │
        │ CLOSE│
        └──────┤
               │  ← Wick (thin, #10b981) ❌ WRONG
      Low ─────┘
```

---

## Volume Bars

### MT5 Volume Rendering
```
Volume Scale (left side):
200 ┐
    │     ▇
150 ┤  ▅  █       ▇
    │  █  █   ▃   █  ▅
100 ┤  █  █   █   █  █  ▂
    │  █  █   █   █  █  █  ▄
 50 ┤  █  █   █   █  █  █  █
    │  █  █   █   █  █  █  █
  0 └─────────────────────────
    10:00  10:05  10:10  10:15

Color Logic:
- If candle is bullish (close ≥ open): Teal (#14b8a6)
- If candle is bearish (close < open): Red (#ef4444)
- Alpha: 0.5 (50% transparency)
```

### Volume Data Structure
```json
{
  "time": 1737364800,
  "open": 2720.50,
  "high": 2720.80,
  "low": 2720.30,
  "close": 2720.60,
  "volume": 125  ← This field exists in backend!
}
```

---

## Trade Level Lines

### MT5 Trade Level Format
```
Entry Line (Blue, solid):
═════════════════════════════════════════
  BUY 0.05 @ 4607.33                     ← Always visible label

Stop Loss (Red, dashed):
─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
  SL @ 4600.00                           ← Dashed line

Take Profit (Green, dashed):
─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
  TP @ 4650.00                           ← Dashed line

Pending Order (Orange, dotted):
· · · · · · · · · · · · · · · · · · · · ·
  SELL STOP 0.10 @ 4720.00               ← Dotted line
```

### Current System Format
```
Entry Line (Blue, dotted):
· · · · · · · · · · · · · · · · · · · · ·
  #123 BUY 0.05                          ← Shows ID, only on hover ❌

Stop Loss (Red, dashed):
─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
  SL: 4600.00                            ← Only on hover ❌

Take Profit (Green, dashed):
─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
  TP: 4650.00                            ← Only on hover ❌
```

**Issues**:
1. Entry line should be solid, not dotted
2. Labels should always be visible, not just on hover
3. Should show "BUY 0.05 @ 4607.33" format, not "#123 BUY 0.05"

---

## Grid System

### MT5 Grid Pattern
```
Horizontal Lines (dotted):
┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄  ← 2720.80

┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄  ← 2720.60

┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄  ← 2720.40

Vertical Lines (dotted):
┆       ┆       ┆       ┆
↑       ↑       ↑       ↑
10:00  10:05  10:10  10:15

Style: LineStyle.Dotted (1)
Color: #27272a (dark gray)
Spacing: Auto-calculated by chart
```

### Current System (Solid Lines)
```
Horizontal Lines (solid):
────────────────────────────  ← 2720.80

────────────────────────────  ← 2720.60

────────────────────────────  ← 2720.40

Vertical Lines (solid):
│       │       │       │
↑       ↑       ↑       ↑
10:00  10:05  10:10  10:15

Style: Solid (default) ❌ Should be dotted
```

---

## Bid/Ask Price Lines

### MT5 Current Price Display
```
Price Scale on Right:

  2720.80
  2720.60
  2720.43  ← ASK (teal, dashed) ───────────────
  2720.40  ← BID (red, dashed)  ───────────────
  2720.20
  2720.00

Lines extend across entire chart
Update in real-time (every tick)
Labels show exact price with 5 decimals
```

### Current System (Missing)
```
Price Scale on Right:

  2720.80
  2720.60
  2720.40  ← Static price labels only
  2720.20
  2720.00

❌ No real-time bid/ask lines
❌ No dynamic price tracking
```

---

## Timeframe Selector

### MT5 Timeframe Bar
```
┌───┬───┬────┬────┬───┬───┬───┬───┬────┐
│M1 │M5 │M15 │M30 │H1 │H4 │D1 │W1 │MN  │
└───┴───┴────┴────┴───┴───┴───┴───┴────┘
 ▲ Active (highlighted in teal)
```

### Current System
```
┌───┬───┬────┬───┬───┬───┐
│M1 │M5 │M15 │H1 │H4 │D1 │
└───┴───┴────┴───┴───┴───┘
 ▲ Active (highlighted in emerald)

❌ Missing: W1 (weekly), MN (monthly)
❌ Missing: M30 (30-minute)
```

---

## Data Requirements

### OHLC Data Structure (Backend → Frontend)
```typescript
interface OHLCBar {
  time: number;      // Unix timestamp
  open: number;      // Opening price
  high: number;      // Highest price
  low: number;       // Lowest price
  close: number;     // Closing price
  volume: number;    // ✅ Available but not used
}
```

### Volume Histogram Data (lightweight-charts)
```typescript
interface HistogramData {
  time: Time;        // Same as OHLC
  value: number;     // Volume amount
  color?: string;    // Optional: teal or red based on candle
}
```

---

## Implementation Checklist

### Colors
- [ ] Bullish candles: `#14b8a6` (teal)
- [ ] Bearish candles: `#ef4444` (red)
- [ ] Volume bars: `#06b6d4` with 50% opacity
- [ ] Grid lines: `#27272a` (dotted)

### Volume
- [ ] Histogram series created
- [ ] Bottom 20% of chart reserved for volume
- [ ] Volume data mapped from OHLC
- [ ] Colors match candle direction
- [ ] Updates in real-time

### Grid
- [ ] Horizontal lines dotted
- [ ] Vertical lines dotted
- [ ] Toggle grid on/off option

### Trade Levels
- [ ] Entry lines always visible
- [ ] Labels show "BUY/SELL volume @ price"
- [ ] SL/TP lines dashed
- [ ] Pending orders different color

### Price Lines
- [ ] Bid line (red, dashed)
- [ ] Ask line (teal, dashed)
- [ ] Update on every tick
- [ ] Price labels visible

### Timeframes
- [ ] M1, M5, M15, M30, H1, H4, D1 working
- [ ] W1, MN added (optional)

---

## Visual Testing

### Quick Visual Check
Open chart and verify:
1. **Candle at 10:00** is teal (bullish)
2. **Volume bar at 10:00** is teal, height = 125
3. **Grid lines** are dotted (not solid)
4. **BID line** shows red dashed line at current bid
5. **Trade label** shows "BUY 0.05 @ 4607.33" (always visible)

### Side-by-Side Comparison
| Element | MT5 Screenshot | Current System | Match? |
|---------|----------------|----------------|--------|
| Candle color | Teal | Emerald | ❌ |
| Volume bars | Cyan | None | ❌ |
| Grid style | Dotted | Solid | ❌ |
| Trade labels | Always visible | Hover only | ❌ |
| Bid/Ask lines | Dynamic | Static | ❌ |

**Target**: All checks ✅ green

---

## Performance Metrics

### Before Optimization
- Render time: ~150ms
- Memory: ~60MB
- FPS: 60

### After Volume + Fixes (Expected)
- Render time: ~170ms (+20ms acceptable)
- Memory: ~70MB (+10MB acceptable)
- FPS: 60 (no degradation)

### Red Flags
- Render > 300ms: Reduce volume bar count
- Memory > 150MB: Limit OHLC history
- FPS < 30: Disable volume or reduce update frequency

---

## Reference Links

- Lightweight Charts Histogram: https://tradingview.github.io/lightweight-charts/tutorials/demos/histogram-series
- Line Styles: https://tradingview.github.io/lightweight-charts/docs/api#linestyle
- Price Lines: https://tradingview.github.io/lightweight-charts/docs/api/interfaces/IPriceLine

---

**This document serves as the visual specification for MT5 parity.**
All implementations should match the layouts and colors shown here.
