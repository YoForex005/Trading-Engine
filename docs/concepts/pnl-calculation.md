# Profit & Loss (P&L) Calculation Methodology

## Overview

This document explains how profit and loss (P&L) is calculated in the RTX Trading Engine for different asset classes, order types, and scenarios.

## Key Concepts

### Realized vs Unrealized P&L

**Realized P&L**
- Profit or loss from closed positions
- Permanently added to/subtracted from account balance
- Recorded in ledger as transaction

**Unrealized P&L** (Floating P&L)
- Profit or loss from open positions
- Changes with market price movements
- Not added to balance until position closes
- Used for margin calculations (equity = balance + unrealized P&L)

### Important Variables

| Variable | Description | Example |
|----------|-------------|---------|
| **Entry Price** | Price at which position was opened | 1.0900 |
| **Exit Price** | Price at which position closes | 1.0950 |
| **Current Price** | Latest market price (for unrealized P&L) | 1.0925 |
| **Volume** | Position size in lots | 0.1 |
| **Contract Size** | Units per lot | 100,000 |
| **Pip Size** | Minimum price movement | 0.0001 |
| **Pip Value** | Value of 1 pip move | $10 |
| **Side** | BUY or SELL | BUY |

## P&L Calculation Formulas

### 1. Forex P&L Calculation

#### Buy Position
```
Price Difference = Exit Price - Entry Price
Pips = Price Difference / Pip Size
P&L = Pips × Pip Value × Volume
```

**Example: EURUSD Buy**
```
Entry:     1.0900
Exit:      1.0950
Volume:    0.1 lots
Pip Size:  0.0001
Pip Value: $10

Price Diff = 1.0950 - 1.0900 = 0.0050
Pips       = 0.0050 / 0.0001 = 50 pips
P&L        = 50 × $10 × 0.1 = $50
```

#### Sell Position
```
Price Difference = Entry Price - Exit Price
Pips = Price Difference / Pip Size
P&L = Pips × Pip Value × Volume
```

**Example: EURUSD Sell**
```
Entry:     1.0950
Exit:      1.0900
Volume:    0.1 lots
Pip Size:  0.0001
Pip Value: $10

Price Diff = 1.0950 - 1.0900 = 0.0050
Pips       = 0.0050 / 0.0001 = 50 pips
P&L        = 50 × $10 × 0.1 = $50
```

### 2. JPY Pairs P&L Calculation

For JPY pairs, pip size is 0.01 (not 0.0001):

**Example: USDJPY Buy**
```
Entry:     147.50
Exit:      148.00
Volume:    0.1 lots
Pip Size:  0.01
Pip Value: $9.09 (approximate)

Price Diff = 148.00 - 147.50 = 0.50
Pips       = 0.50 / 0.01 = 50 pips
P&L        = 50 × $9.09 × 0.1 = $45.45
```

### 3. Crypto P&L Calculation

For crypto (BTCUSD, ETHUSD), contract size is 1:

**Example: BTCUSD Buy**
```
Entry:         95,000
Exit:          96,000
Volume:        0.01 (lots = contracts)
Contract Size: 1
Pip Size:      0.01
Pip Value:     0.01

Price Diff = 96,000 - 95,000 = 1,000
Pips       = 1,000 / 0.01 = 100,000 pips
P&L        = 100,000 × 0.01 × 0.01 = $10

Or simply: (96,000 - 95,000) × 0.01 = $10
```

### 4. Code Implementation

```go
// From internal/core/engine.go
func (e *Engine) calculatePnL(pos *Position, currentPrice, volume float64, spec *SymbolSpec) float64 {
    if spec == nil {
        return 0
    }

    var priceDiff float64
    if pos.Side == "BUY" {
        priceDiff = currentPrice - pos.OpenPrice
    } else {
        priceDiff = pos.OpenPrice - currentPrice
    }

    // P&L = (PriceDiff / PipSize) * PipValue * Volume
    pips := priceDiff / spec.PipSize
    return pips * spec.PipValue * volume
}
```

## Unrealized P&L (Floating P&L)

### Real-time Calculation

For open positions, unrealized P&L is calculated continuously:

**Buy Position**
```
Current P&L = (Current Bid - Entry Price) / Pip Size × Pip Value × Volume
```

**Sell Position**
```
Current P&L = (Entry Price - Current Ask) / Pip Size × Pip Value × Volume
```

### Why Bid for Buy, Ask for Sell?

- **Buy positions close at Bid**: When you sell, you get the Bid price
- **Sell positions close at Ask**: When you buy back, you pay the Ask price

### Example: Real-time P&L Update

```go
// Position opened
Position: BUY 0.1 EURUSD @ 1.0900

// Market moves
Tick 1: Bid 1.0905, Ask 1.0907
  → Unrealized P&L = (1.0905 - 1.0900) / 0.0001 × $10 × 0.1 = $5

Tick 2: Bid 1.0910, Ask 1.0912
  → Unrealized P&L = (1.0910 - 1.0900) / 0.0001 × $10 × 0.1 = $10

Tick 3: Bid 1.0895, Ask 1.0897
  → Unrealized P&L = (1.0895 - 1.0900) / 0.0001 × $10 × 0.1 = -$5
```

## Account Equity Calculation

```
Equity = Balance + Unrealized P&L (all positions)
```

**Example**
```
Account Balance: $5,000

Open Positions:
  Position 1: EURUSD BUY 0.1 @ 1.0900, Current: 1.0910 → +$10
  Position 2: GBPUSD SELL 0.2 @ 1.2600, Current: 1.2610 → -$20
  Position 3: USDJPY BUY 0.1 @ 147.50, Current: 148.00 → +$45.45

Total Unrealized P&L = $10 - $20 + $45.45 = $35.45
Equity = $5,000 + $35.45 = $5,035.45
```

## Fees and Adjustments

### 1. Commission

Deducted when position opens:

```go
commission := spec.CommissionPerLot * volume
account.Balance -= commission
```

**Example**
```
Commission: $5 per lot
Volume: 0.5 lots
Total Commission: $5 × 0.5 = $2.50

Opening Balance: $5,000.00
After Commission: $4,997.50
```

### 2. Swap (Overnight Fee)

Applied daily for positions held overnight:

**Positive Swap** (earn interest)
```
Position: BUY EURUSD
Daily Swap: +$0.50
Balance += $0.50
```

**Negative Swap** (pay interest)
```
Position: SELL GBPUSD
Daily Swap: -$1.20
Balance -= $1.20
```

### 3. Net P&L Calculation

```
Net P&L = Gross P&L - Commission - Swap
```

**Complete Example**
```
Entry: BUY 0.1 EURUSD @ 1.0900
Exit:  Close @ 1.0950
Held: 2 days

Gross P&L:    +$50.00
Commission:   -$2.50  ($5/lot × 0.1 × 2 trades)
Swap (2 days): -$1.00 (-$0.50/day × 2)

Net P&L = $50.00 - $2.50 - $1.00 = $46.50
```

## Partial Position Closing

When closing part of a position:

```go
closeVolume := 0.05  // Closing half of 0.1 lot position
remainingVolume := 0.05

// Calculate P&L for closed portion only
realizedPnL := calculatePnL(position, closePrice, closeVolume, spec)

// Update position
position.Volume -= closeVolume
account.Balance += realizedPnL
```

**Example**
```
Original Position: BUY 0.1 EURUSD @ 1.0900
Current Price: 1.0950
Close 0.05 lots:

Realized P&L = (1.0950 - 1.0900) / 0.0001 × $10 × 0.05 = $25
Balance += $25

Remaining Position: BUY 0.05 EURUSD @ 1.0900
```

## Multiple Position Scenarios

### Hedging Mode (Multiple Positions per Symbol)

```
Account allows multiple positions on same symbol:

Position 1: BUY 0.1 EURUSD @ 1.0900  → Unrealized P&L: +$10
Position 2: SELL 0.1 EURUSD @ 1.0920 → Unrealized P&L: -$5
Position 3: BUY 0.2 EURUSD @ 1.0880  → Unrealized P&L: +$60

Total Unrealized P&L = $10 - $5 + $60 = $65
```

Each position tracked independently.

### Netting Mode (Single Net Position per Symbol)

```
Trades are netted into single position:

Trade 1: BUY 0.1 EURUSD @ 1.0900
Trade 2: BUY 0.2 EURUSD @ 1.0920
  → Net Position: BUY 0.3 EURUSD @ weighted average price

Weighted Entry = (0.1 × 1.0900 + 0.2 × 1.0920) / 0.3
               = (0.109 + 0.2184) / 0.3
               = 1.0913333

Current Price: 1.0950
Unrealized P&L = (1.0950 - 1.0913333) / 0.0001 × $10 × 0.3 = $110
```

## Stop Loss and Take Profit Execution

### Stop Loss (SL)

Automatically closes position when price reaches SL level:

**Buy Position with SL**
```
Entry:  BUY 0.1 EURUSD @ 1.0900
SL:     1.0850
Trigger: When Bid ≤ 1.0850

Realized P&L when triggered:
= (1.0850 - 1.0900) / 0.0001 × $10 × 0.1 = -$50
```

**Sell Position with SL**
```
Entry:  SELL 0.1 EURUSD @ 1.0900
SL:     1.0950
Trigger: When Ask ≥ 1.0950

Realized P&L when triggered:
= (1.0900 - 1.0950) / 0.0001 × $10 × 0.1 = -$50
```

### Take Profit (TP)

Automatically closes position when profit target reached:

**Buy Position with TP**
```
Entry:  BUY 0.1 EURUSD @ 1.0900
TP:     1.0950
Trigger: When Bid ≥ 1.0950

Realized P&L when triggered:
= (1.0950 - 1.0900) / 0.0001 × $10 × 0.1 = +$50
```

### Code Implementation

```go
// From internal/core/engine.go
func (e *Engine) UpdatePrice(symbol string, bid, ask float64) {
    for _, pos := range e.positions {
        if pos.Symbol != symbol || pos.Status != "OPEN" {
            continue
        }

        var currentPrice float64
        if pos.Side == "BUY" {
            currentPrice = bid
        } else {
            currentPrice = ask
        }

        // Check Stop Loss
        if pos.SL > 0 {
            if (pos.Side == "BUY" && currentPrice <= pos.SL) ||
               (pos.Side == "SELL" && currentPrice >= pos.SL) {
                e.ClosePosition(pos.ID, pos.Volume)
            }
        }

        // Check Take Profit
        if pos.TP > 0 {
            if (pos.Side == "BUY" && currentPrice >= pos.TP) ||
               (pos.Side == "SELL" && currentPrice <= pos.TP) {
                e.ClosePosition(pos.ID, pos.Volume)
            }
        }
    }
}
```

## Ledger Recording

All P&L changes are recorded in the ledger:

```go
type LedgerEntry struct {
    ID         int64
    AccountID  int64
    Type       string  // REALIZED_PNL, COMMISSION, SWAP, etc.
    Amount     float64
    Balance    float64  // Balance after transaction
    Reference  string   // Trade ID or Position ID
    Timestamp  int64
}
```

**Example Ledger Entries**
```
Entry 1:
  Type: COMMISSION
  Amount: -$2.50
  Balance: $4,997.50
  Reference: Trade #1234

Entry 2:
  Type: REALIZED_PNL
  Amount: +$50.00
  Balance: $5,047.50
  Reference: Position #5678 closed

Entry 3:
  Type: SWAP
  Amount: -$0.50
  Balance: $5,047.00
  Reference: Position #5679 (overnight)
```

## Summary

### Key Formulas

**Forex P&L (Buy)**
```
P&L = ((Exit - Entry) / PipSize) × PipValue × Volume
```

**Forex P&L (Sell)**
```
P&L = ((Entry - Exit) / PipSize) × PipValue × Volume
```

**Unrealized P&L (Buy)**
```
P&L = ((CurrentBid - Entry) / PipSize) × PipValue × Volume
```

**Unrealized P&L (Sell)**
```
P&L = ((Entry - CurrentAsk) / PipSize) × PipValue × Volume
```

**Equity**
```
Equity = Balance + Sum(Unrealized P&L)
```

**Net P&L**
```
Net P&L = Gross P&L - Commission - Swap
```

### Important Notes

1. **Buy positions close at Bid**, **Sell positions close at Ask**
2. Unrealized P&L updates in real-time with every tick
3. Commission is deducted immediately when position opens
4. Swap is applied daily for positions held overnight
5. Realized P&L permanently affects account balance
6. Equity (not balance) is used for margin calculations
7. SL/TP are monitored continuously and execute automatically

This P&L calculation methodology ensures accurate, real-time tracking of account performance and is the foundation for risk management and margin calculations.
