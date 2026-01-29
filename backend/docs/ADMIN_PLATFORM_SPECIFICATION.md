# Trading Platform Admin System - Comprehensive Specification

**Version**: 1.0
**Date**: 2026-01-18
**Status**: Research Complete - Ready for Implementation

---

## Executive Summary

This document provides a comprehensive specification for a professional-grade trading platform admin system, based on extensive research of industry leaders (MetaTrader 5 Manager, cTrader Admin) and regulatory requirements for 2026.

The admin platform enables complete lifecycle management of clients, funds, orders, risk, and compliance for forex/CFD brokers operating A-Book, B-Book, or hybrid C-Book execution models.

---

## 1. Client & User Management

### 1.1 Core Account Operations

**Create/Edit/Delete Accounts**
- Demo and live account types
- Swap-free (Islamic) accounts
- VIP and institutional accounts
- Account migration between groups
- Bulk account operations (CSV import)

**Account Details Dashboard**
```
Account View:
├── Basic Info: ID, Number, User, Email, Phone
├── Financial: Balance, Equity, Margin, Free Margin, Margin Level
├── Settings: Leverage, Group, Status, Type (Demo/Live)
├── Trading: Open Positions, Pending Orders, Total Volume, Total P&L
├── History: Last Login, Registration Date, Last Trade
└── KYC: Verification Status, Documents, Risk Score
```

**Account Status Management**
- `ACTIVE` - Normal trading enabled
- `SUSPENDED` - Temporary trading halt
- `DISABLED` - Permanent closure
- `RESTRICTED` - Limited operations (e.g., close-only)
- `PENDING_VERIFICATION` - KYC incomplete

**Security Features**
- Password reset and recovery
- Two-factor authentication (2FA) enforcement
- IP whitelisting per account
- Session management (view active sessions, force logout)
- Login history and audit trail

### 1.2 Client Classification & Profiling

**Automatic Classification System**

| Classification | Win Rate | Hold Time | Volume | B-Book % |
|---------------|----------|-----------|--------|----------|
| RETAIL | < 48% | Any | Small | 90% |
| SEMI_PRO | 48-52% | 5m-1h | Medium | 50% |
| PROFESSIONAL | > 52% | > 1h | Large | 20% |
| TOXIC | > 55% | < 1m | Any | 0% |

**Tracked Metrics**
- Win rate (% of profitable trades)
- Profit factor (gross profit / gross loss)
- Sharpe ratio (risk-adjusted returns)
- Maximum drawdown (peak to trough decline)
- Average trade size and hold time
- Order-to-fill ratio (cancellation rate)
- Instrument concentration (% in top symbol)
- Time-of-day patterns (trading hours distribution)
- Correlation with known toxic clients

**Toxicity Score (0-100)**
```go
toxicityScore = 0
IF win_rate > 55% THEN toxicityScore += 30
IF sharpe_ratio > 2.0 THEN toxicityScore += 25
IF avg_hold_time < 60s THEN toxicityScore += 20
IF cancel_rate > 50% THEN toxicityScore += 15
IF instrument_concentration > 80% THEN toxicityScore += 10

IF toxicityScore >= 70 THEN classification = TOXIC
```

**Risk Tier Assignment**
- `LOW` - Retail, safe for B-Book
- `MEDIUM` - Semi-pro, monitor closely
- `HIGH` - Professional, mostly A-Book
- `CRITICAL` - Toxic, 100% A-Book or reject

### 1.3 KYC/AML Integration

**Document Verification**
- Government-issued ID (passport, driver's license)
- Proof of address (utility bill, bank statement < 3 months)
- Source of funds (bank statements, payslips)
- Selfie with ID (liveness detection)

**KYC Provider Integration**
- **Onfido** - AI-powered ID verification
- **Jumio** - Real-time biometric authentication
- **Trulioo** - Global identity verification
- **ComplyAdvantage** - Sanctions and PEP screening

**PEP (Politically Exposed Person) Screening**
- Database: World-Check, Dow Jones, LexisNexis
- Screening frequency: On registration, then every 90 days
- Risk levels: Low, Medium, High
- Enhanced due diligence for high-risk PEPs

**Sanctions List Screening**
- OFAC (US), UN, EU, UK HMT
- Real-time screening on registration and withdrawals
- Automatic blocking of sanctioned individuals/countries
- Alert notifications to compliance team

**AML Monitoring**
- Transaction monitoring (deposits, withdrawals, transfers)
- Pattern detection (structuring, smurfing, round-tripping)
- Suspicious activity thresholds (configurable)
- SAR (Suspicious Activity Report) filing workflow
- Case management system for investigations

**Ongoing Monitoring**
- Re-screening every 90 days (automated)
- Client activity profiling (normal vs abnormal)
- Large transaction alerts (> $10,000)
- Country risk scoring (high-risk jurisdictions)

---

## 2. Fund Operations

### 2.1 Deposit Management

**Deposit Methods**
- Bank transfer (wire, ACH, SEPA)
- Credit/debit cards (Visa, Mastercard)
- E-wallets (Skrill, Neteller, PayPal)
- Cryptocurrency (BTC, ETH, USDT)
- Local payment methods (regional)

**Payment Gateway Integration**
- **Stripe** - Cards, bank transfers, wallets
- **PayPal** - PayPal balances and cards
- **Adyen** - Global payment processing
- **Coinbase Commerce** - Crypto payments
- **PrimeXM** - Institutional payments

**Deposit Workflow**
```
1. Client initiates deposit
2. Payment gateway processes payment
3. Webhook notification to admin system
4. Verify payment (amount, method, account)
5. Compliance check (AML, sanctions)
6. Auto-credit account OR manual approval
7. Email/SMS confirmation to client
8. Audit log entry
```

**Manual Deposit Processing**
- View pending deposits
- Approve or reject deposits
- Add notes and reason codes
- Upload proof of payment
- Adjust amount (if discrepancy)
- Refund processing

**Reconciliation**
- Daily reconciliation with payment provider
- Identify discrepancies (missing, duplicate)
- Auto-match by reference ID
- Manual matching interface
- Export reconciliation reports

**Multi-Currency Support**
- Account base currency (USD, EUR, GBP, etc.)
- Deposit currency conversion
- FX rate source (ECB, Reuters)
- Conversion fee configuration
- Historical rate tracking

### 2.2 Withdrawal Management

**Withdrawal Request Handling**
```
Withdrawal Request
├── KYC Verification (must be approved)
├── AML Checks (transaction monitoring)
├── Balance Validation (sufficient funds)
├── Withdrawal Limits (daily, monthly)
├── Approval Workflow (auto or manual)
├── Payment Processing (to bank/wallet)
└── Confirmation (email/SMS)
```

**Approval Workflow**
- Auto-approve: Verified clients, < $1,000, low-risk
- Manual review: New clients, > $1,000, high-risk
- Two-factor approval: > $10,000 (requires 2 admins)
- Rejection reasons: Incomplete KYC, bonus wagering, fraud

**Withdrawal Limits**
- Daily limit per account (configurable)
- Monthly limit per account
- Limits by client tier (retail, VIP)
- Limits by payment method
- Override capability (admin approval)

**Priority Processing**
- Standard: 1-3 business days
- Fast: Same-day (higher fee)
- VIP: Priority queue (free for VIP clients)
- Batch processing (scheduled payouts)

**Payout Methods**
- Same method as deposit (regulatory requirement)
- Bank transfer (same account as registered)
- E-wallet (verified wallet address)
- Cryptocurrency (verified crypto address)
- Check (for large amounts, rare)

**Withdrawal Fees**
- Free withdrawals: 1 per month (configurable)
- Fee structure: Fixed + percentage
- Fee by method (bank = $20, crypto = 0.5%)
- Fee waiver for VIP clients

### 2.3 Credit/Debit Adjustments

**Manual Balance Adjustments**
```go
type Adjustment struct {
    AccountID   int64
    Type        string // CREDIT, DEBIT
    Amount      float64
    Reason      string // Required
    Category    string // COMPENSATION, CORRECTION, GOODWILL, ERROR
    AdminID     int64
    ApprovedBy  int64  // For large amounts (>$1000)
    Reference   string // Ticket ID, trade ID
    Notes       string
}
```

**Use Cases**
1. **Compensation**: Spread error, platform downtime, LP issues
2. **P&L Correction**: Wrong price executed, trade reversal
3. **Goodwill**: Client retention, complaint resolution
4. **Error Correction**: Accounting errors, duplicate transactions
5. **Margin Credit**: Temporary credit for large trades
6. **Bonus Application**: Manual bonus addition

**Approval Requirements**
- < $100: Single admin approval
- $100-$1,000: Senior admin approval
- > $1,000: Two admin approvals + reason code
- > $10,000: Compliance officer + CFO approval

**Audit Trail**
- Admin ID and username
- Before/after balance
- Reason and category
- Approval chain
- Timestamp (UTC)
- IP address
- Related tickets/trades

**Restrictions**
- Cannot adjust during open positions (unless approved)
- Cannot create negative balance (unless NBP override)
- Requires written client consent (for debits)
- Regulatory reporting (large adjustments)

### 2.4 Bonus & Promotion Management

**Bonus Types**

| Type | Description | Withdrawable | Wagering |
|------|-------------|--------------|----------|
| Welcome Bonus | First deposit bonus | No | 3x volume |
| Deposit Bonus | % match on deposit | No | 5x volume |
| Cashback | Rebate on volume | Yes | None |
| Credit Bonus | Trading credit | No | 10x volume |
| No-Deposit Bonus | Free credit | No | 20x volume |

**Bonus Configuration**
```go
type Bonus struct {
    ID              string
    Name            string
    Type            string // WELCOME, DEPOSIT, CASHBACK, CREDIT
    Amount          float64 // Fixed amount
    Percentage      float64 // % of deposit
    MaxBonus        float64 // Cap
    MinDeposit      float64 // Minimum deposit to qualify
    Withdrawable    bool
    WageringMultiplier float64 // 3x, 5x, 10x
    ExpirationDays  int
    TargetGroups    []int64 // Client groups eligible
    StartDate       time.Time
    EndDate         time.Time
    AutoApply       bool // Auto or manual
    TermsURL        string
}
```

**Auto-Apply Rules**
- Trigger: First deposit, deposit >= $X, specific group
- Conditions: KYC approved, no active bonus, country allowed
- Application: Auto-credit on deposit or after wagering
- Notification: Email with bonus details and T&Cs

**Wagering Requirements**
```
Wagering = Bonus Amount × Multiplier

Example:
Bonus: $100
Multiplier: 5x
Required Volume: $500 (in lots)

Track:
- Current volume: $320
- Remaining: $180
- Progress: 64%
```

**Bonus Expiration**
- Auto-removal after X days
- Warning notifications (7 days before, 1 day before)
- Manual extension (admin approval)
- Forfeiture on withdrawal attempt (if wagering incomplete)

**Bonus Abuse Detection**
- Multiple accounts from same IP
- Same payment method across accounts
- Deposit → Bonus → Withdrawal pattern
- High-frequency bonus claims
- Bonus hunting (deposit only when bonus available)

**Actions on Abuse**
- Flag account for review
- Void bonus and winnings
- Suspend account
- Report to compliance
- Blacklist IP/payment method

### 2.5 Segregated Account Management

**Regulatory Requirements**
- Client funds must be segregated from company funds
- Daily reconciliation required
- Trustee account at licensed bank
- Regulatory reporting (monthly)

**Segregated Account Tracking**
```
Daily Reconciliation:
├── Client Balances (Total): $10,000,000
├── Segregated Account (Bank): $10,500,000
├── Buffer: $500,000 (5%)
└── Status: COMPLIANT
```

**Discrepancy Handling**
- Auto-detect discrepancies > 1%
- Alert compliance team immediately
- Investigation workflow
- Corrective action (transfer funds)
- Regulatory notification (if required)

**Reporting**
- Daily balance report
- Monthly reconciliation statement
- Quarterly external audit
- Annual regulatory submission

---

## 3. Order Management

### 3.1 View Orders & Positions

**Real-Time Order Book**
```
Order View:
├── Pending Orders
│   ├── Limit Orders (buy/sell, price, volume)
│   ├── Stop Orders (stop-loss, take-profit)
│   ├── OCO (One-Cancels-Other)
│   └── Trailing Stops
├── Open Positions
│   ├── Position ID, Symbol, Side, Volume
│   ├── Open Price, Current Price, P&L
│   ├── SL/TP levels
│   └── Swap/Commission
└── Order History
    ├── Filled, Cancelled, Rejected, Expired
    └── Execution reports
```

**Search & Filter**
- By account ID, username, account number
- By symbol, order ID, position ID
- By order type (market, limit, stop)
- By side (buy, sell)
- By status (pending, filled, cancelled)
- By date range
- By P&L range (winning, losing)

**Export Capabilities**
- CSV, Excel, PDF formats
- Bulk export (all accounts)
- Scheduled exports (daily, weekly)
- Email delivery to admins

### 3.2 Modify Orders

**Order Parameter Editing**
```go
type OrderModification struct {
    OrderID         int64
    NewPrice        *float64 // For limit/stop orders
    NewVolume       *float64 // Increase/decrease size
    NewSL           *float64 // Stop-loss level
    NewTP           *float64 // Take-profit level
    NewExpiration   *time.Time // Good-Till-Date
    Reason          string // Required
    AdminID         int64
}
```

**Manual Execution**
- Execute pending order at current market price
- Execute at specific price (admin override)
- Partial execution (fill portion of volume)
- Reject order (with reason)

**Routing Override**
- Force A-Book execution (bypass routing rules)
- Force B-Book internalization
- Specify target LP (LMAX, YOFX, etc.)
- Override for single trade or all future trades

**Audit Trail**
- Before/after order parameters
- Admin ID and reason
- Timestamp and IP address
- Client notification (if applicable)

### 3.3 Reverse/Cancel Orders

**Cancel Pending Orders**
- Single order cancellation
- Bulk cancellation (all orders for account)
- Cancellation by symbol (e.g., cancel all EURUSD orders)
- Reason code required

**Reverse Executed Trades**
```
Trade Reversal Workflow:
1. Admin requests reversal (reason required)
2. Compliance approval (for large trades)
3. Reverse trade execution
   - Create opposite trade (same price, volume)
   - Credit/debit account P&L
   - Reverse commissions/swaps
4. Adjust account balance
5. Notify client
6. Audit log entry
```

**Reversal Conditions**
- Error trade (wrong price, wrong symbol)
- Platform malfunction
- LP pricing error
- Client dispute resolution
- Time limit: Within 24 hours (configurable)

**Close Positions**
- Close single position (full or partial)
- Close all positions for account
- Close all positions for symbol
- Market close vs limit close
- Slippage tolerance

**Force Close Scenarios**
- Margin call (80% margin level)
- Stop-out (50% margin level)
- Risk limit breach (admin override)
- Account closure request
- Regulatory requirement

### 3.4 Delete Orders (Historical)

**Soft Delete**
- Mark order as deleted (status = DELETED)
- Retain in database (compliance requirement)
- Exclude from reports (unless specified)
- Reversible (can be undeleted)

**Hard Delete**
- Permanent removal from database
- Only after retention period (7 years)
- Requires super admin approval
- Regulatory approval (if applicable)
- Backup before deletion (mandatory)

**Bulk Deletion**
- Delete orders by date range
- Delete orders for closed accounts
- Delete cancelled orders (after retention)
- Dry-run mode (preview before delete)

---

## 4. Group Management

### 4.1 Trader Group Configuration

**Group Hierarchy**
```
├── Default (All Clients)
├── Retail
│   ├── Standard
│   ├── Mini
│   └── Swap-Free
├── VIP
│   ├── Silver
│   ├── Gold
│   └── Platinum
├── Institutional
│   ├── Hedge Funds
│   └── Prop Firms
└── Inactive (Dormant/Closed)
```

**Group Settings**
```go
type UserGroup struct {
    ID              int64
    Name            string
    Description     string

    // Execution
    ExecutionMode   string // BBOOK, ABOOK, HYBRID
    TargetLP        string // Default LP for A-Book

    // Financial
    Markup          float64 // Spread markup (pips)
    Commission      float64 // Per lot
    DefaultBalance  float64 // For new accounts

    // Leverage & Risk
    MaxLeverage     float64
    MarginMode      string // HEDGING, NETTING
    MarginCallLevel float64
    StopOutLevel    float64

    // Symbols
    EnabledSymbols  []string
    SymbolSettings  map[string]SymbolGroupSettings

    // Trading Hours
    TradingHours    TradingHours

    // Swap
    SwapMode        string // NORMAL, SWAP_FREE, CUSTOM
    SwapMultiplier  float64

    // Status
    Status          string // ACTIVE, DISABLED
    CreatedAt       time.Time
    CreatedBy       string
}
```

**Execution Mode**
- `BBOOK` - All trades internalized (market making)
- `ABOOK` - All trades routed to LP (STP/ECN)
- `HYBRID` - Dynamic routing based on client classification

**Default Account Settings**
- Balance: $10,000 (demo), $0 (live)
- Leverage: 1:100 (retail), 1:500 (pro)
- Margin mode: Hedging (allow opposite positions)
- Currency: USD, EUR, GBP, etc.

**Enabled Symbols**
- Symbol whitelist (only listed symbols tradeable)
- Symbol blacklist (exclude specific symbols)
- Asset class filters (Forex, Metals, Indices, Crypto)

**Trading Hours**
- 24/5 (Forex), 9:30-16:00 (Indices)
- Pre-market, market, post-market sessions
- Holiday calendar integration
- Auto-close positions before weekends (optional)

**Swap Settings**
- `NORMAL` - Standard swap rates from LP
- `SWAP_FREE` - Zero swap (Islamic accounts)
- `CUSTOM` - Admin-defined swap rates per symbol

### 4.2 Markup & Spread Configuration

**Group-Level Markup**
```
Base Spread (LP): 1.2 pips
Group Markup: 0.5 pips
Client Spread: 1.7 pips
```

**Per-Symbol Markup Override**
```go
type SymbolGroupSettings struct {
    Symbol         string
    Markup         float64 // Override group markup
    Commission     float64 // Override group commission
    MaxVolume      float64 // Max position size
    MinVolume      float64 // Min position size
    Disabled       bool    // Disable trading for symbol
}

// Example
EURUSD: markup = 0.3 pips (lower for major pair)
GBPJPY: markup = 1.0 pips (higher for volatile pair)
XAUUSD: markup = 0.05 pips (absolute value, not pips)
```

**Commission Structure**
- **Per Lot**: Fixed amount per standard lot (e.g., $7/lot)
- **Per Side**: Charge on open only
- **Round Turn**: Charge on open + close
- **Percentage**: % of trade value
- **Volume Tiers**: Lower commission for higher volume

**Spread Type**
- **Fixed Spread**: Always 1.5 pips (simple, predictable)
- **Variable Spread**: LP spread + markup (market-driven)
- **Raw Spread + Commission**: 0.2 pips + $7/lot (ECN model)

**Slippage Control**
- Max allowed slippage: 3 pips
- Reject order if slippage > limit
- Re-quote if slippage > threshold
- Positive slippage (pass to client vs keep)

### 4.3 Commission Management

**IB (Introducing Broker) Commissions**
```
Multi-Level Structure:
├── Master IB (50% commission share)
│   ├── Sub-IB Level 1 (30% share)
│   │   └── Sub-IB Level 2 (20% share)
│   └── Sub-IB Level 1 (30% share)
└── Direct Clients (0% share)
```

**Commission Calculation**
```
Trade Volume: 10 lots
Commission: $7/lot
Total Commission: $70

Master IB: $70 × 50% = $35
Sub-IB L1: $70 × 30% = $21
Sub-IB L2: $70 × 20% = $14
Broker: $70 - $35 - $21 - $14 = $0 (or broker keeps X%)
```

**Volume-Based Rebates**
```go
type CommissionTier struct {
    MinVolume    float64 // In lots/month
    MaxVolume    float64
    Commission   float64 // Per lot
    Rebate       float64 // Cashback %
}

// Example
0-100 lots: $7/lot, 0% rebate
101-500 lots: $6/lot, 5% rebate
500+ lots: $5/lot, 10% rebate
```

**Commission Reports**
- Per IB (monthly earnings)
- Per group (total commissions paid)
- Commission payout schedule (weekly, monthly)
- Export to Excel/PDF

**Commission Payout Automation**
- Auto-calculate monthly commissions
- Generate payout report
- Send to finance for processing
- Email notification to IBs
- Track payment status

**Custom Commission Formulas**
```go
// Example: Progressive commission
if volume < 100 {
    commission = volume * 7.0
} else if volume < 500 {
    commission = (100 * 7.0) + ((volume - 100) * 6.0)
} else {
    commission = (100 * 7.0) + (400 * 6.0) + ((volume - 500) * 5.0)
}
```

### 4.4 Group Migration

**Move Accounts Between Groups**
```go
func MigrateAccount(accountID int64, targetGroupID int64) {
    // 1. Validate migration
    // 2. Apply new group settings
    // 3. Notify client
    // 4. Audit log
}
```

**Bulk Migration**
- CSV import (Account ID, Target Group)
- Dry-run mode (preview changes)
- Batch processing (1000 accounts at a time)
- Progress tracking
- Error handling (rollback on failure)

**Group Settings Inheritance**
- Markup, commission, leverage updated immediately
- Swap rates applied on next rollover
- Enabled symbols updated (close disabled symbols)
- Trading hours enforced on next order

**Transition Grace Period**
- 24-hour notice before migration (email to client)
- Option to keep old settings for X days
- Client consent required (for downgrades)

---

## 5. FIX API Provisioning

### 5.1 FIX Credentials Management

**Generate FIX Credentials**
```go
type FIXCredentials struct {
    AccountID       int64
    SenderCompID    string // CLIENT_12345
    TargetCompID    string // BROKER
    BeginString     string // FIX.4.4
    Username        string
    Password        string // Encrypted

    // Sessions
    TradingSession  FIXSession
    MarketDataSession FIXSession

    // Security
    IPWhitelist     []string
    MaxConnections  int

    // Status
    Status          string // ACTIVE, DISABLED, EXPIRED
    CreatedAt       time.Time
    ExpiresAt       *time.Time
}

type FIXSession struct {
    Host            string // fix.broker.com
    Port            int    // 12336
    SSL             bool
    HeartbeatInterval int // 30 seconds
}
```

**Credential Lifecycle**
- **Generate**: On client request or admin creation
- **Activate**: After client confirms email
- **Rotate**: Forced password change every 90 days
- **Expire**: Auto-disable after 180 days of inactivity
- **Revoke**: Admin-initiated, immediate effect

**Separate Sessions**
- **Trading Session**: NewOrderSingle, OrderCancelRequest, etc.
- **Market Data Session**: MarketDataRequest, QuoteRequest
- Allows independent rate limits and IP whitelists

**IP Whitelisting**
- Required for all FIX connections
- Up to 5 IPs per session
- Auto-reject connections from unlisted IPs
- IP change request workflow (email verification)

**Session Limits**
- Max concurrent connections: 2 per session
- Auto-disconnect oldest session on new connection
- Reconnection cooldown: 5 seconds (prevent hammering)

### 5.2 Conditional Access Rules

**Trading Hours Restrictions**
```go
type TradingHoursRule struct {
    SessionID       string
    AllowedHours    []TimeRange // 24/5, 9:30-16:00, etc.
    Timezone        string
    Exceptions      []Date // Holidays
    Action          string // REJECT_ORDERS, CLOSE_ONLY
}
```

**Symbol Access Control**
- Whitelist: Only allow specific symbols
- Blacklist: Block specific symbols
- Asset class filter: Forex only, no crypto
- Dynamic updates (push to client on symbol disable)

**Volume Limits**
```go
type VolumeLimit struct {
    SessionID       string
    MaxOrderVolume  float64 // Max single order size
    MaxDailyVolume  float64 // Max total volume per day
    MaxPositions    int     // Max concurrent positions
    Action          string  // REJECT, WARN, ALLOW_WITH_APPROVAL
}
```

**Order Type Restrictions**
- Market orders only (no limits/stops)
- Limit orders only (no market execution)
- No stop-loss allowed (risky)
- No hedging (no opposite positions)

**Rate Limiting**
```go
type RateLimit struct {
    SessionID           string
    MaxOrdersPerSecond  int
    MaxOrdersPerMinute  int
    MaxOrdersPerHour    int
    BurstAllowance      int // Allow short bursts
    ThrottleAction      string // REJECT, DELAY, DISCONNECT
}
```

**Geographic Restrictions**
- Country whitelist (only allow US, EU)
- Country blacklist (block sanctioned countries)
- IP geolocation validation
- VPN detection (optional)

### 5.3 FIX Rules Engine

**Pre-Trade Validation Rules**
```go
type FIXRule struct {
    ID              string
    Name            string
    Priority        int // 1 = highest
    Enabled         bool

    // Conditions
    Conditions      []Condition

    // Actions
    Action          string // ALLOW, REJECT, MODIFY, ROUTE_TO_ADMIN
    Reason          string
    NotifyClient    bool
    NotifyAdmin     bool
}

type Condition struct {
    Field           string // Symbol, Volume, Price, etc.
    Operator        string // EQUALS, GT, LT, IN, NOT_IN
    Value           interface{}
}
```

**Example Rules**
```yaml
Rule 1: Block Large Orders
  Conditions:
    - Volume > 100 lots
    - Symbol IN [EURUSD, GBPUSD]
  Action: REJECT
  Reason: "Exceeds max volume limit"

Rule 2: Route Crypto to A-Book
  Conditions:
    - Symbol CONTAINS "USD" AND Symbol STARTS_WITH "BTC|ETH|XRP"
  Action: ROUTE_ABOOK
  TargetLP: "LMAX"

Rule 3: Block During News
  Conditions:
    - Time IN news_blackout_window
    - Symbol = EURUSD
  Action: REJECT
  Reason: "Trading halted during news event"
```

**Order Rejection Triggers**
- Insufficient margin (< required margin)
- Symbol disabled (maintenance, delisted)
- Outside trading hours
- Excessive volume (> max limit)
- Price too far from market (> 1% deviation)
- Account suspended/restricted

**Automated Notifications**
- Email to client on rule violation
- SMS alert to admin on repeated violations
- Dashboard alert (real-time)
- Audit log entry

**Rule Prioritization**
- Rules evaluated in priority order (1, 2, 3, ...)
- First matching rule wins
- Override rules for VIP clients (priority 1)

**Custom Rule Scripting**
```lua
-- Lua script example
function evaluate(order)
  if order.volume > 50 and order.symbol == "XAUUSD" then
    if order.account.group == "Retail" then
      return {action = "REJECT", reason = "Gold volume too high"}
    else
      return {action = "ALLOW"}
    end
  end
  return {action = "ALLOW"}
end
```

### 5.4 Session Monitoring

**Active FIX Sessions Dashboard**
```
Session View:
├── Session ID: YOFX1_12345
├── Status: CONNECTED
├── Uptime: 4h 23m
├── IP Address: 203.0.113.42
├── Last Heartbeat: 5s ago
├── Messages:
│   ├── Sent: 1,234
│   ├── Received: 1,189
│   ├── Rejected: 3
│   └── Sequence Numbers: In=567, Out=578
└── Performance:
    ├── Latency: 45ms (avg), 78ms (p95)
    ├── Fill Rate: 98.3%
    ├── Reject Rate: 0.2%
```

**Connection Status**
- `CONNECTED` - Active, heartbeat received
- `DISCONNECTED` - Intentional disconnect
- `TIMEOUT` - No heartbeat for 2× interval
- `REJECTED` - Login failed (bad credentials)
- `THROTTLED` - Rate limit exceeded

**Message Throughput Metrics**
- Messages per second (current, avg, peak)
- Message types breakdown (35=D, 35=8, 35=F, etc.)
- Bandwidth usage (KB/s)
- Queue depth (pending messages)

**Sequence Number Tracking**
```
Sequence Numbers:
├── Inbound: Expected=123, Received=123 ✓
├── Outbound: Next=124
├── Gap Detected: No
└── Resend Requests: 2 (today)
```

**Gap Fill Requests**
- Auto-request missing messages (GapFill, 35=4)
- Resend request (35=2)
- Message store (persistent, 7 days)
- Sequence reset on session restart

**Session-Level Analytics**
```
Performance Metrics (24h):
├── Latency
│   ├── Average: 45ms
│   ├── Median (p50): 38ms
│   ├── p95: 78ms
│   ├── p99: 125ms
│   └── Max: 234ms
├── Fill Rate: 98.3% (1,189/1,210 orders filled)
├── Reject Rate: 0.2% (3/1,210 orders rejected)
├── Message Counts
│   ├── NewOrderSingle (35=D): 1,210
│   ├── ExecutionReport (35=8): 2,398
│   ├── OrderCancelRequest (35=F): 45
│   ├── OrderCancelReject (35=9): 2
│   └── Heartbeat (35=0): 576
└── Errors: 5 (sequence gaps, parsing errors)
```

---

## 6. CRM Integration

### 6.1 Client Communication

**Email Templates**
- Welcome email (registration)
- Email verification
- Deposit confirmation
- Withdrawal confirmation
- Margin call warning
- Account statements (monthly)
- Promotional campaigns

**SMS Notifications**
- Login from new device
- Large withdrawal request
- Password reset
- 2FA codes
- Urgent alerts (margin call)

**Push Notifications**
- Mobile app notifications
- Position closed (SL/TP hit)
- Order filled
- Deposit/withdrawal status
- Market news alerts

**In-App Messaging**
- Client portal announcements
- Promotional banners
- Support chat integration
- Ticket updates

**Client Announcements**
- System-wide (all clients)
- Targeted (specific groups, segments)
- Scheduled (future delivery)
- Urgency levels (info, warning, critical)

**Email Campaign Management**
- Drip campaigns (onboarding, re-engagement)
- A/B testing (subject lines, content)
- Open/click rate tracking
- Unsubscribe management
- GDPR compliance (consent required)

**Communication History**
- All emails/SMS sent to client
- Read receipts (if available)
- Click tracking (links in emails)
- Response tracking (reply emails)
- Opt-out preferences

### 6.2 CRM Data Sync

**Bi-Directional Sync**
```
Trading Platform ←→ CRM
├── Client Registration → Create Lead/Contact
├── Deposit → Update Deal Stage
├── KYC Status → Update Lead Status
├── Trade Activity → Update Engagement Score
└── CRM Update → Update Client Info
```

**REST API for CRM Integration**
```
Endpoints:
POST /api/crm/sync/client
POST /api/crm/sync/deposit
POST /api/crm/sync/kyc
GET  /api/crm/client/:id
PUT  /api/crm/client/:id
```

**Webhook Notifications**
```go
type WebhookEvent struct {
    Event       string // client.registered, deposit.completed
    Timestamp   time.Time
    ClientID    int64
    Data        map[string]interface{}
    Signature   string // HMAC signature for verification
}

// Events
- client.registered
- client.kyc_approved
- deposit.completed
- withdrawal.completed
- trade.executed
- position.closed
```

**Custom Field Mapping**
```yaml
Platform → CRM:
  account_id → External_ID
  email → Email
  phone → Mobile
  balance → Account_Balance
  total_pnl → Lifetime_PnL
  status → Status
```

**Real-Time Data Push**
- WebSocket connection to CRM
- Event-driven updates (deposit, withdrawal, trade)
- Batching (reduce API calls)
- Retry logic (exponential backoff)
- Dead letter queue (failed events)

### 6.3 Lead Management

**Lead Capture**
- Website form submissions
- Landing page conversions
- Live chat interactions
- Phone inquiries
- Social media leads

**Lead Scoring**
```go
type LeadScore struct {
    Demographics   int // 0-25 (country, age, income)
    Engagement     int // 0-25 (email opens, clicks, visits)
    Intent         int // 0-30 (demo account, deposit amount)
    Fit            int // 0-20 (matches ICP)
    Total          int // 0-100
}

// Example
Country: US (+10)
Engagement: High (+20)
Deposit Intent: $5,000 (+25)
Total: 55 (Medium Priority)
```

**Lead Qualification**
- BANT: Budget, Authority, Need, Timeline
- Auto-qualify based on score (>70 = qualified)
- Manual review for borderline leads
- Disqualification reasons (no budget, competitor, etc.)

**Sales Funnel Tracking**
```
Funnel Stages:
├── Lead (website visitor)
├── Contacted (sales call)
├── Demo Account (registered)
├── Deposit (funded account)
├── Active (first trade)
└── Retained (3+ months)
```

**Conversion Analytics**
- Conversion rate per stage
- Average time in stage
- Drop-off analysis
- Attribution (source, campaign)
- ROI per channel

**Lead Assignment**
- Round-robin (distribute evenly)
- Territory-based (by country/region)
- Skill-based (by language, product)
- Ownership rules (sales rep capacity)
- Auto-assignment on qualification

**Follow-Up Automation**
- Drip campaigns (nurture leads)
- Task reminders (follow-up in 3 days)
- Escalation (if no response in 7 days)
- Re-engagement (dormant leads)

### 6.4 Support Ticket Integration

**Client Tickets**
```go
type Ticket struct {
    ID          int64
    AccountID   int64
    Username    string
    Email       string
    Subject     string
    Category    string // TECHNICAL, BILLING, GENERAL
    Priority    string // LOW, MEDIUM, HIGH, URGENT
    Status      string // OPEN, IN_PROGRESS, RESOLVED, CLOSED
    AssignedTo  int64 // Admin ID
    CreatedAt   time.Time
    UpdatedAt   time.Time
    ResolvedAt  *time.Time
    Messages    []TicketMessage
}
```

**Ticket Status Workflow**
```
NEW → OPEN → IN_PROGRESS → RESOLVED → CLOSED
     ↓
     ESCALATED
```

**Priority Levels**
- `LOW` - General inquiry, 48h SLA
- `MEDIUM` - Account issue, 24h SLA
- `HIGH` - Trading issue, 4h SLA
- `URGENT` - Critical issue (platform down), 1h SLA

**Auto-Assignment**
- By category (billing → finance team)
- By language (Spanish → Spanish support)
- By product (MT5 → MT5 specialist)
- By availability (online agents only)

**SLA Compliance**
```
SLA Tracking:
├── First Response Time: 15 minutes (SLA: 1 hour) ✓
├── Resolution Time: 3 hours (SLA: 4 hours) ✓
├── Customer Satisfaction: 4.5/5
└── Escalation: No
```

**Internal Notes**
- Admin-only notes (not visible to client)
- Collaboration (tag other admins)
- Investigation findings
- Resolution steps

---

## Implementation Roadmap

### Phase 1: MVP (Months 1-3)
**Priority**: Core functionality for broker operations

**Features**:
1. Client management (create, edit, view accounts)
2. Fund operations (deposits, withdrawals, adjustments)
3. Order viewing and basic modifications
4. Group management (basic settings, markup, leverage)
5. Risk controls (pre-trade checks, margin monitoring)
6. Audit logging (basic trail for compliance)
7. Admin authentication (login, 2FA, sessions)

**Deliverables**:
- Admin web interface (React)
- Backend API (Go)
- PostgreSQL database
- Basic reporting (clients, deposits, orders)

**Timeline**: 12 weeks

---

### Phase 2: Core Features (Months 4-6)
**Priority**: Compliance and advanced operations

**Features**:
8. KYC/AML integration (Onfido, sanctions screening)
9. FIX API provisioning (credentials, access rules)
10. B-Book/A-Book routing (basic classification)
11. Financial reporting (P&L, commission, statements)
12. CRM integration (basic sync with Salesforce/HubSpot)
13. Real-time dashboards (admin overview, risk dashboard)
14. Circuit breakers (volatility, daily loss, news events)

**Deliverables**:
- KYC provider integration
- FIX session management
- Routing engine (basic rules)
- Dashboard widgets (real-time)

**Timeline**: 12 weeks

---

### Phase 3: Advanced Features (Months 7-12)
**Priority**: Intelligence and optimization

**Features**:
15. C-Book hybrid routing (ML-based client profiling)
16. Advanced analytics (client profiling, routing effectiveness)
17. Regulatory reporting (MiFID II, EMIR, best execution)
18. White-label management (multi-brand support)
19. Custom report builder (drag-drop interface)
20. API rate limiting and security hardening

**Deliverables**:
- ML model for client classification
- Regulatory reporting automation
- White-label infrastructure
- Advanced analytics dashboard

**Timeline**: 24 weeks

---

### Phase 4: Optimization (Months 13+)
**Priority**: Scalability and performance

**Features**:
21. Performance tuning (horizontal scaling, caching)
22. Advanced ML models (deep learning, LSTM)
23. Real-time stress testing (VaR, scenario analysis)
24. Multi-LP smart order routing (price comparison)
25. Cross-client netting and optimization

**Deliverables**:
- Kubernetes deployment
- Advanced ML pipeline
- Stress testing framework
- Multi-LP SOR

**Timeline**: Ongoing

---

## Technology Stack

### Backend
- **Language**: Go (high performance, concurrency)
- **Framework**: Gin (REST API), Gorilla WebSocket
- **Database**: PostgreSQL (primary), Redis (cache)
- **Message Queue**: Kafka (event streaming)
- **Search**: Elasticsearch (audit logs, analytics)

### Frontend
- **Framework**: React + TypeScript
- **UI Library**: Material-UI (admin interface)
- **State Management**: Redux Toolkit
- **Charting**: TradingView Charting Library
- **Real-Time**: WebSocket (live updates)

### Infrastructure
- **Cloud**: AWS (ECS, RDS, ElastiCache, S3)
- **Containers**: Docker + Kubernetes
- **CI/CD**: GitHub Actions
- **Monitoring**: Prometheus + Grafana
- **Logging**: ELK Stack (Elasticsearch, Logstash, Kibana)

---

## Sources & References

### Industry Research
- [MetaTrader 5 for hedge fund manager](https://www.metatrader5.com/en/hedge-funds/owner)
- [MT5 Manager API Solutions](https://brokeret.com/api/mt5-api)
- [cTrader Admin platform](https://www.spotware.com/ctrader/brokers/admin/)
- [cTrader FIX API Specification](https://help.ctrader.com/fix/specification/)
- [A-Book vs B-Book Brokers](https://b2broker.com/news/a-book-vs-b-book-brokers-whats-the-difference/)
- [Brokerage Models Comparison](https://www.effectivesoft.com/blog/a-book-b-book-hybrid-brokerage-models-comparison.html)

### CRM & Admin Platforms
- [Forex CRM Solutions 2026](https://newyorkcityservers.com/blog/forex-crm-solutions)
- [B2Core Trader's Room](https://b2broker.com/products/b2core-traders-room/)
- [Match-Trade CRM](https://match-trade.com/products/client-office-app-with-forex-crm/)
- [Top Broker Management Software](https://www.crmone.com/broker-management-software)

### Compliance & Regulation
- [Brokerage Setup: CRM for Brokers](https://centroidsol.com/brokerage-setup-series-crm-for-financial-brokers/)
- [Risk Management in Brokerage](https://www.soft-fx.com/blog/risk-management-in-brokerage-business/)
- [Understanding Brokerage Models](https://devexperts.com/blog/a-book-b-book-and-hybrid-models-in-forex-brokerage/)

---

**Document Version**: 1.0
**Last Updated**: 2026-01-18
**Next Review**: 2026-02-18
**Owner**: Trading Platform Development Team
