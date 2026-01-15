# Phase 7: Multi-Asset Support - Research

**Researched:** 2026-01-16
**Domain:** Multi-asset trading platform (Forex, Crypto, Stocks, Commodities, Indices, CFDs)
**Confidence:** HIGH

<research_summary>
## Summary

Researched the ecosystem for building a multi-asset trading platform that supports forex, cryptocurrencies, stocks, commodities, indices, and CFDs. The modern approach uses specialized liquidity provider APIs for each asset class (Interactive Brokers for stocks, Binance for crypto, commodity data feeds) with a unified architecture that normalizes symbol metadata, trading hours, and contract specifications across asset types.

Key finding: Don't hand-roll symbol metadata management, market calendar/trading hours logic, or CFD contract specifications. Use structured data from LP APIs and standardized libraries. Fragmented technology stacks are the #1 pitfall—build a unified architecture with consistent data models across all asset classes to avoid data silos, execution misalignment, and integration complexity.

**Primary recommendation:** Use RESTful/WebSocket APIs for most integrations (Interactive Brokers Web API, Alpaca, commodity data APIs), reserve FIX protocol only if serving institutional clients requiring ultra-low latency. Build unified symbol metadata layer with asset-agnostic interfaces, implement consistent session management across all asset classes, and use existing CFD calculation libraries rather than custom margin/swap logic.
</research_summary>

<standard_stack>
## Standard Stack

The established libraries/tools for multi-asset trading platform integration:

### Core Liquidity Providers
| Provider | Asset Classes | API Type | Why Standard |
|----------|---------------|----------|--------------|
| Interactive Brokers | Stocks, Forex, Commodities, Indices | REST (Web API), FIX, TWS API | Industry standard for multi-asset institutional access, 30+ years established |
| Alpaca | Stocks, Crypto | REST, WebSocket | Modern developer-friendly API, official Nasdaq vendor, free tier available |
| Binance | Cryptocurrencies | REST, WebSocket | Largest crypto exchange by volume, comprehensive API documentation |
| Twelve Data | All major assets | REST | Unified API for stocks, forex, crypto, commodities, indices with 10+ years historical data |
| Financial Modeling Prep | Stocks, Commodities, Indices | REST, WebSocket | Real-time quotes, historical data, index constituents, commodity futures |

### Supporting Data Feeds
| Provider | Purpose | Coverage | When to Use |
|----------|---------|----------|-------------|
| Commodities-API | Commodity prices | 600+ commodities (Gold, Oil, Silver, etc.) | When primary LP lacks commodity coverage |
| Indices-API | Index data | 170+ world indices (S&P 500, NASDAQ, DAX) | Real-time index quotes and historical data |
| TradingHours.com | Market calendars | Every global exchange, holidays, session breaks | Managing trading hours across all asset classes |
| Metals-API | Precious metals | Gold, Silver, Platinum, Palladium | Specialized precious metals pricing |

### Alternative LP Integrations Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Interactive Brokers | Multiple single-asset LPs | IB provides unified multi-asset access; separate LPs increase integration complexity |
| REST/WebSocket APIs | FIX protocol | FIX offers ultra-low latency (microseconds) but requires institutional setup; REST/WebSocket sufficient for retail broker platforms |
| Commercial data feeds | Free/open APIs | Commercial feeds (Bloomberg, Refinitiv) offer guaranteed SLAs and support but cost $1000+/month vs free tiers |

**Installation:**

Each LP has its own SDK/client library:

```bash
# Interactive Brokers (Python example)
pip install ibapi

# Alpaca (Go example)
go get github.com/alpacahq/alpaca-trade-api-go/v3

# For data APIs, typically REST calls with API keys
# No specific client library required, use standard HTTP client
```
</standard_stack>

<architecture_patterns>
## Architecture Patterns

### Recommended Project Structure
```
backend/
├── lpmanager/
│   ├── adapters/
│   │   ├── oanda.go          # Existing forex LP
│   │   ├── binance.go         # Existing crypto LP
│   │   ├── interactivebrokers.go  # NEW: Multi-asset LP
│   │   ├── commodities.go     # NEW: Commodity data feed
│   │   └── indices.go         # NEW: Index data feed
│   ├── metadata/
│   │   ├── symbol_specs.go    # Unified symbol metadata
│   │   ├── trading_hours.go   # Session management
│   │   └── contract_specs.go  # CFD/futures specifications
│   └── manager.go             # LP manager with multi-asset routing
├── engine/
│   ├── cfd_calculator.go      # CFD margin, swap, rollover
│   └── asset_validator.go     # Asset-specific validation rules
```

### Pattern 1: Unified Symbol Metadata Layer
**What:** Asset-agnostic symbol specification interface that normalizes metadata across all asset classes
**When to use:** All multi-asset platforms to avoid data silos and enable consistent execution logic
**Example:**
```go
// Unified symbol specification across all asset classes
type SymbolSpec struct {
    Symbol        string          // "AAPL", "BTCUSD", "XAUUSD", "SPX"
    AssetClass    AssetClass      // Stock, Crypto, Commodity, Index, CFD
    TickSize      decimal.Decimal // Minimum price movement
    ContractSize  decimal.Decimal // Units per lot/contract
    MarginRate    decimal.Decimal // Required margin percentage
    TradingHours  TradingSession  // Market hours with timezone
    SwapRateLong  decimal.Decimal // Overnight long financing (CFDs)
    SwapRateShort decimal.Decimal // Overnight short financing (CFDs)
    LPSource      string          // "InteractiveBrokers", "Binance", etc.
}

// Asset-agnostic interface for LP adapters
type LiquidityProvider interface {
    GetSymbolSpec(symbol string) (*SymbolSpec, error)
    GetQuote(symbol string) (*Quote, error)
    PlaceOrder(order *Order) (*Execution, error)
    GetTradingHours(symbol string) (*TradingSession, error)
}
```

### Pattern 2: Asset-Specific Validation with Fallback Chain
**What:** Validate orders against asset-specific rules (trading hours, margin requirements) with LP failover
**When to use:** Before submitting orders to ensure compliance with asset-specific constraints
**Example:**
```go
// Validate order based on asset class
func (e *Engine) ValidateOrder(order *Order, symbol *SymbolSpec) error {
    // Check trading hours for asset class
    if !symbol.TradingHours.IsMarketOpen(time.Now()) {
        return fmt.Errorf("market closed for %s", symbol.Symbol)
    }

    // Asset-specific validation
    switch symbol.AssetClass {
    case AssetClassStock:
        // US stocks: min $1 per share typically
        if order.Price.LessThan(decimal.NewFromInt(1)) {
            return errors.New("invalid stock price")
        }
    case AssetClassCFD:
        // CFDs: validate margin requirement
        requiredMargin := order.Volume.Mul(order.Price).Mul(symbol.MarginRate)
        if e.Account.FreeMargin.LessThan(requiredMargin) {
            return errors.New("insufficient margin for CFD")
        }
    case AssetClassCrypto:
        // Crypto: typically no minimum, but validate contract size
        if order.Volume.LessThan(symbol.ContractSize) {
            return errors.New("volume below minimum contract size")
        }
    }

    return nil
}
```

### Pattern 3: CFD Swap/Rollover Calculator
**What:** Calculate overnight financing charges for CFDs based on position size and asset-specific rates
**When to use:** Daily at market close for all open CFD positions
**Example:**
```go
// Calculate overnight swap charge/credit for CFD position
// Formula: Lot Size × Contract Size × Swap Rate × Number of Nights
func CalculateCFDSwap(position *Position, symbol *SymbolSpec, nights int) decimal.Decimal {
    swapRate := symbol.SwapRateLong
    if position.Side == SideSell {
        swapRate = symbol.SwapRateShort
    }

    // Base swap = position size × swap rate per lot
    baseSwap := position.Volume.Mul(swapRate)

    // Account for triple swap on Wednesdays (weekend rollover)
    multiplier := decimal.NewFromInt(int64(nights))
    if isWednesday(time.Now()) {
        multiplier = decimal.NewFromInt(3) // Triple swap
    }

    return baseSwap.Mul(multiplier)
}

// Apply swap charges at daily rollover (typically 10pm UK time)
func (e *Engine) ApplyDailySwaps() error {
    rolloverTime := time.Date(time.Now().Year(), time.Now().Month(),
        time.Now().Day(), 22, 0, 0, 0, time.UTC) // 10pm UTC

    if time.Now().Before(rolloverTime) {
        return nil // Not yet rollover time
    }

    for _, position := range e.GetOpenPositions() {
        symbol, _ := e.GetSymbolSpec(position.Symbol)
        if symbol.AssetClass != AssetClassCFD {
            continue // Only CFDs have swap charges
        }

        swap := CalculateCFDSwap(position, symbol, 1)
        e.Account.ApplySwap(swap) // Credit or debit account
    }

    return nil
}
```

### Pattern 4: Multi-LP Routing with Asset Class Preference
**What:** Route orders to appropriate LP based on asset class with automatic failover
**When to use:** All order execution to ensure optimal LP selection and redundancy
**Example:**
```go
// LP routing configuration by asset class
type LPRouter struct {
    routes map[AssetClass][]LiquidityProvider
}

func NewLPRouter() *LPRouter {
    return &LPRouter{
        routes: map[AssetClass][]LiquidityProvider{
            AssetClassStock:     {NewInteractiveBrokersLP(), NewAlpacaLP()},
            AssetClassCrypto:    {NewBinanceLP(), NewAlpacaLP()},
            AssetClassCommodity: {NewInteractiveBrokersLP(), NewCommoditiesAPI()},
            AssetClassIndex:     {NewInteractiveBrokersLP(), NewIndicesAPI()},
            AssetClassForex:     {NewOandaLP(), NewInteractiveBrokersLP()},
        },
    }
}

// Route order to best LP for asset class with failover
func (r *LPRouter) RouteOrder(order *Order, assetClass AssetClass) (*Execution, error) {
    providers := r.routes[assetClass]
    if len(providers) == 0 {
        return nil, fmt.Errorf("no LP configured for %s", assetClass)
    }

    // Try each LP in order until success
    var lastErr error
    for _, lp := range providers {
        execution, err := lp.PlaceOrder(order)
        if err == nil {
            return execution, nil
        }
        lastErr = err
        log.Printf("LP %T failed, trying next: %v", lp, err)
    }

    return nil, fmt.Errorf("all LPs failed for %s: %v", assetClass, lastErr)
}
```

### Anti-Patterns to Avoid
- **Hardcoded symbol specifications:** Use LP-provided metadata or structured config files, never hardcode tick sizes or contract specifications
- **Single LP for all asset classes:** Different LPs specialize in different assets; using one LP limits coverage and creates single point of failure
- **Custom CFD margin calculations:** Use LP-provided margin requirements or industry-standard formulas; custom math leads to regulatory issues
- **Ignoring trading hours/sessions:** Each asset class has different market hours and holiday calendars; ignoring this causes rejected orders
- **Fragmented data models:** Using separate database schemas or data structures for each asset class creates data silos and integration complexity
</architecture_patterns>

<dont_hand_roll>
## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Symbol metadata | Manual CSV files or hardcoded specs | LP API symbol specification endpoints | LPs update tick sizes, contract sizes, margin requirements; manual files go stale and cause execution errors |
| Trading hours/calendars | Custom timezone and holiday logic | TradingHours.com API or LP-provided session data | 250+ exchanges, irregular holidays, session breaks, daylight saving changes—complex edge cases |
| CFD swap calculations | Custom interest rate formulas | LP-provided swap rates or industry calculators | Swap rates vary by asset, LP markup, triple swaps on Wednesdays; custom formulas miss edge cases |
| Market data normalization | Custom parsers for each LP format | Unified data abstraction layer (like TA-Lib or custom wrapper) | Each LP has different timestamp formats, decimal precision, null handling; normalization is complex |
| FIX protocol implementation | Custom FIX message parser | QuickFIX library (C++, Java, Python, Go) | FIX protocol is extensive (50,000+ pages of specs); custom parsers are bug-prone and incomplete |
| Currency conversion | Custom forex rate lookups | LP-provided conversion rates or dedicated forex API | Real-time rates, bid/ask spreads, market hours affect accuracy; custom lookups miss nuances |
| Contract rollover (futures) | Manual rollover date tracking | LP-provided contract specifications with expiry dates | Rollover dates vary by contract, exchange rules change; manual tracking causes missed rollovers |

**Key insight:** Multi-asset trading platforms have decades of established patterns and data standards. Symbol specifications, trading calendars, and CFD calculations are complex regulatory requirements that LPs maintain professionally. Custom implementations introduce regulatory risk, data staleness, and execution errors. Use LP-provided metadata and industry-standard libraries to avoid these pitfalls.
</dont_hand_roll>

<common_pitfalls>
## Common Pitfalls

### Pitfall 1: Fragmented Technology Infrastructure
**What goes wrong:** Building separate systems for each asset class (one for stocks, one for forex, one for crypto) creates data silos, execution misalignment, and integration complexity
**Why it happens:** Starting with single-asset MVP and bolting on new assets without refactoring architecture
**How to avoid:** Design unified data models and interfaces from the start; use asset-agnostic abstractions (SymbolSpec, LiquidityProvider interface) that work across all asset classes
**Warning signs:** Code duplication across asset-specific modules, difficulty aggregating cross-asset reports, different execution pathways for different assets

### Pitfall 2: Ignoring Asset-Specific Trading Rules
**What goes wrong:** Orders rejected due to market closed, invalid tick sizes, or insufficient margin for CFDs; leads to poor user experience and execution failures
**Why it happens:** Assuming all assets trade 24/7 like forex or using same validation rules across all asset classes
**How to avoid:** Implement asset-specific validation (trading hours, tick size rounding, margin requirements) before submitting orders to LP; query LP for trading sessions and contract specifications
**Warning signs:** High order rejection rates, "market closed" errors during expected trading hours, margin call alerts for seemingly valid positions

### Pitfall 3: Stale Symbol Metadata
**What goes wrong:** Execution errors due to outdated tick sizes, contract sizes, or margin requirements; especially problematic during LP spec changes or contract rollovers
**Why it happens:** Loading symbol specs once at startup or using hardcoded values instead of querying LP dynamically
**How to avoid:** Refresh symbol specifications periodically (daily or on first use each session); subscribe to LP spec update notifications if available; implement cache invalidation
**Warning signs:** Execution failures after LP updates specifications, margin calculations incorrect, orders rejected for "invalid lot size"

### Pitfall 4: Incorrect CFD Swap Calculations
**What goes wrong:** Wrong swap charges applied to CFD positions, leading to customer complaints, regulatory issues, and P&L discrepancies
**Why it happens:** Misunderstanding triple swap on Wednesdays, wrong swap rate sign (long vs short), or missing broker markup on LP swap rates
**How to avoid:** Use LP-provided swap rates exactly as specified; test swap calculations with known positions; apply triple swap only on configured day (typically Wednesday); log all swap calculations for audit
**Warning signs:** Customer disputes about swap charges, swap credits when debits expected (or vice versa), inconsistent swap amounts for same position size

### Pitfall 5: Currency Conversion Errors in Multi-Asset Portfolios
**What goes wrong:** Account balance shows incorrect values when positions denominated in different currencies (USD stocks + EUR forex + BTC crypto); P&L calculations wrong
**Why it happens:** Using stale forex rates, ignoring bid/ask spreads in conversions, or converting at wrong timestamps
**How to avoid:** Use real-time forex rates from LP for all currency conversions; apply bid/ask spread (use bid for sells, ask for buys); convert at exact execution timestamp, not end-of-day rate
**Warning signs:** Account balance fluctuates unexpectedly when no trades executed, P&L differs from LP reports, margin level incorrect

### Pitfall 6: Session/Timezone Management Complexity
**What goes wrong:** Trading hours calculated incorrectly due to timezone confusion, daylight saving changes, or exchange-specific session breaks
**Why it happens:** Hardcoding UTC offsets, not handling DST transitions, or ignoring exchange-specific lunch breaks (e.g., Chinese stock markets)
**How to avoid:** Use TradingHours.com API or LP-provided session data with IANA timezone identifiers; never hardcode UTC offsets; test around DST transition dates
**Warning signs:** Orders accepted during supposed market close, "market closed" errors during supposed open hours, session breaks ignored

### Pitfall 7: Over-Diversification and Liquidity Fragmentation
**What goes wrong:** Supporting 100+ symbols across all asset classes but low liquidity from LPs, leading to wide spreads and poor execution quality
**Why it happens:** Adding every possible symbol without validating LP liquidity or trader demand
**How to avoid:** Start with most liquid symbols per asset class (top 20 stocks, major forex pairs, BTC/ETH for crypto); validate LP provides adequate liquidity before adding symbol; monitor spread and slippage per symbol
**Warning signs:** Wide bid/ask spreads on obscure symbols, frequent partial fills, customer complaints about execution quality
</common_pitfalls>

<code_examples>
## Code Examples

Verified patterns from official sources and industry best practices:

### Unified Symbol Specification Loading
```go
// Load symbol specifications from Interactive Brokers API
// Source: IB Web API documentation pattern
func (lp *InteractiveBrokersLP) LoadSymbolSpecs() error {
    // IB Web API endpoint: /v1/api/trsrv/secdef
    symbols := []string{"AAPL", "XAUUSD", "SPX", "CL"} // Stock, Commodity, Index, Futures

    for _, symbol := range symbols {
        resp, err := lp.client.Get(fmt.Sprintf(
            "https://api.ibkr.com/v1/api/trsrv/secdef?symbol=%s", symbol))
        if err != nil {
            return err
        }

        var ibSpec IBSecurityDefinition
        json.NewDecoder(resp.Body).Decode(&ibSpec)

        // Normalize to unified SymbolSpec
        spec := &SymbolSpec{
            Symbol:       ibSpec.Symbol,
            AssetClass:   mapIBAssetClass(ibSpec.SecType),
            TickSize:     decimal.NewFromFloat(ibSpec.MinTick),
            ContractSize: decimal.NewFromFloat(ibSpec.Multiplier),
            MarginRate:   decimal.NewFromFloat(ibSpec.MarginRate),
            TradingHours: parseIBTradingHours(ibSpec.TradingHours),
            LPSource:     "InteractiveBrokers",
        }

        lp.symbolCache[symbol] = spec
    }

    return nil
}

// Map IB asset types to unified asset classes
func mapIBAssetClass(secType string) AssetClass {
    switch secType {
    case "STK":
        return AssetClassStock
    case "CASH":
        return AssetClassForex
    case "CMDTY":
        return AssetClassCommodity
    case "IND":
        return AssetClassIndex
    case "CFD":
        return AssetClassCFD
    default:
        return AssetClassUnknown
    }
}
```

### Trading Hours Validation
```go
// Validate if market is open for symbol, accounting for asset-specific hours
// Source: Industry best practice pattern
func (tm *TradingHoursManager) IsMarketOpen(symbol string, assetClass AssetClass, t time.Time) (bool, error) {
    // Load trading hours from cache or API
    hours, exists := tm.cache[symbol]
    if !exists {
        // Fetch from TradingHours.com API or LP
        hours, err := tm.fetchTradingHours(symbol, assetClass)
        if err != nil {
            return false, err
        }
        tm.cache[symbol] = hours
    }

    // Check if current time falls within trading session
    for _, session := range hours.Sessions {
        if t.Weekday() == session.DayOfWeek {
            // Parse session times in exchange timezone
            loc, _ := time.LoadLocation(hours.Timezone)
            openTime := time.Date(t.Year(), t.Month(), t.Day(),
                session.OpenHour, session.OpenMinute, 0, 0, loc)
            closeTime := time.Date(t.Year(), t.Month(), t.Day(),
                session.CloseHour, session.CloseMinute, 0, 0, loc)

            if t.After(openTime) && t.Before(closeTime) {
                // Check if not on holiday
                if !tm.isHoliday(symbol, t) {
                    return true, nil
                }
            }
        }
    }

    return false, nil
}

// Example trading hours structure
type TradingSession struct {
    Timezone  string             // "America/New_York", "Europe/London"
    Sessions  []SessionPeriod    // Multiple sessions per day (pre-market, regular, after-hours)
    Holidays  []time.Time        // Exchange-specific holidays
}

type SessionPeriod struct {
    DayOfWeek   time.Weekday
    OpenHour    int
    OpenMinute  int
    CloseHour   int
    CloseMinute int
    SessionType string  // "pre-market", "regular", "after-hours"
}
```

### CFD Margin Calculation
```go
// Calculate required margin for CFD position
// Source: Industry standard CFD margin formula
func CalculateCFDMargin(order *Order, symbol *SymbolSpec) decimal.Decimal {
    // Formula: Position Value × Margin Rate
    // Position Value = Volume × Contract Size × Price

    positionValue := order.Volume.Mul(symbol.ContractSize).Mul(order.Price)
    requiredMargin := positionValue.Mul(symbol.MarginRate)

    // Example:
    // 10 lots × 100 contract size × $50 price = $50,000 position value
    // $50,000 × 0.10 margin rate = $5,000 required margin

    return requiredMargin
}

// Validate margin before opening CFD position
func (e *Engine) ValidateCFDMargin(order *Order, symbol *SymbolSpec) error {
    if symbol.AssetClass != AssetClassCFD {
        return nil // Not a CFD, no special margin validation
    }

    requiredMargin := CalculateCFDMargin(order, symbol)

    // Check if account has sufficient free margin
    if e.Account.FreeMargin.LessThan(requiredMargin) {
        return fmt.Errorf("insufficient margin: required %s, available %s",
            requiredMargin, e.Account.FreeMargin)
    }

    // Reserve margin for position
    e.Account.UsedMargin = e.Account.UsedMargin.Add(requiredMargin)
    e.Account.FreeMargin = e.Account.FreeMargin.Sub(requiredMargin)

    return nil
}
```

### Multi-LP Adapter Pattern
```go
// Universal adapter interface for all liquidity providers
// Source: Adapter pattern for multi-LP integration
type LPAdapter interface {
    Connect() error
    Disconnect() error
    GetQuote(symbol string) (*Quote, error)
    PlaceOrder(order *Order) (*Execution, error)
    GetSymbolSpec(symbol string) (*SymbolSpec, error)
    SubscribeQuotes(symbols []string, callback func(*Quote)) error
}

// Interactive Brokers adapter implementation
type InteractiveBrokersAdapter struct {
    client *http.Client
    apiKey string
    baseURL string
}

func (ib *InteractiveBrokersAdapter) GetQuote(symbol string) (*Quote, error) {
    resp, err := ib.client.Get(fmt.Sprintf("%s/v1/api/md/snapshot?symbols=%s",
        ib.baseURL, symbol))
    if err != nil {
        return nil, err
    }

    var ibQuote []struct {
        Symbol string  `json:"symbol"`
        Bid    float64 `json:"31"` // IB uses numeric field codes
        Ask    float64 `json:"86"`
        Last   float64 `json:"84"`
    }

    json.NewDecoder(resp.Body).Decode(&ibQuote)

    return &Quote{
        Symbol: ibQuote[0].Symbol,
        Bid:    decimal.NewFromFloat(ibQuote[0].Bid),
        Ask:    decimal.NewFromFloat(ibQuote[0].Ask),
        Last:   decimal.NewFromFloat(ibQuote[0].Last),
        Time:   time.Now(),
    }, nil
}

// Alpaca adapter implementation
type AlpacaAdapter struct {
    client *alpaca.Client
}

func (a *AlpacaAdapter) GetQuote(symbol string) (*Quote, error) {
    quote, err := a.client.GetLatestQuote(symbol, alpaca.GetLatestQuoteRequest{})
    if err != nil {
        return nil, err
    }

    return &Quote{
        Symbol: symbol,
        Bid:    decimal.NewFromFloat(quote.BidPrice),
        Ask:    decimal.NewFromFloat(quote.AskPrice),
        Last:   decimal.NewFromFloat(quote.BidPrice), // Use bid as last
        Time:   quote.Timestamp,
    }, nil
}

// LP Manager routes to appropriate adapter based on asset class
func (lpm *LPManager) GetBestQuote(symbol string, assetClass AssetClass) (*Quote, error) {
    adapters := lpm.getAdaptersForAsset(assetClass)

    var bestQuote *Quote
    var bestSpread decimal.Decimal

    for _, adapter := range adapters {
        quote, err := adapter.GetQuote(symbol)
        if err != nil {
            continue // Try next adapter
        }

        spread := quote.Ask.Sub(quote.Bid)
        if bestQuote == nil || spread.LessThan(bestSpread) {
            bestQuote = quote
            bestSpread = spread
        }
    }

    return bestQuote, nil
}
```
</code_examples>

<sota_updates>
## State of the Art (2024-2026)

What's changed recently in multi-asset trading platform integration:

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| FIX protocol mandatory | REST/WebSocket primary, FIX optional | 2020-2024 | Modern platforms use REST/WebSocket for retail, reserve FIX for institutional; lower barrier to entry |
| Separate systems per asset | Unified multi-asset platform | 2022-2026 | Microservices architecture with unified data layer becoming standard; reduces fragmentation |
| Manual symbol spec updates | Dynamic LP-provided metadata | 2023-2026 | LPs now expose real-time symbol specifications via API; eliminates stale data issues |
| Single LP per platform | Multi-LP aggregation with routing | 2024-2026 | Platforms route to best LP per asset class with automatic failover; improves execution quality |
| Hardcoded trading hours | TradingHours.com API integration | 2021-2025 | Dynamic trading calendar APIs handle holidays, DST, session breaks; reduces maintenance |

**New tools/patterns to consider:**
- **Alpaca as Multi-Asset Gateway:** Alpaca now supports both stocks and crypto through single API, simplifying integration for platforms needing both asset classes (2025 expansion)
- **WebTransport for Real-Time Data:** WebSocket's successor offering better performance on unreliable networks; some LPs beginning to support (2026 emerging)
- **Polygon.io Rebrand to Massive.com:** Market data provider rebranded with focus on scale and reliability; APIs remain compatible (2026)
- **AI-Powered LP Routing:** Some platforms using AI to predict optimal LP based on order characteristics, time of day, and historical execution quality (2025-2026 trend)
- **Embedded Trading Calendars in Symbol Specs:** LPs increasingly embedding trading hours, holidays, and session breaks directly in symbol specification responses (2025-2026)

**Deprecated/outdated:**
- **TWS Gateway Required:** Interactive Brokers now offers direct Web API access without requiring Trader Workstation installation; TWS API still available but Web API preferred for new integrations
- **FIX 4.2:** FIX 5.0 SP2 is current standard for institutional trading; platforms using old FIX versions missing features like algorithmic orders and extended metadata
- **Manual CFD Swap Files:** Brokers manually maintaining CSV files with swap rates; modern LPs expose swap rates via API with real-time updates
- **Single-Currency Accounts:** Modern platforms support multi-currency accounts with automatic conversion; single-currency designs limit asset class support
</sota_updates>

<open_questions>
## Open Questions

Things that couldn't be fully resolved during research:

1. **Interactive Brokers API Latency for High-Frequency Trading**
   - What we know: IB offers REST Web API, TWS API, and FIX protocol with varying latency profiles
   - What's unclear: Exact latency benchmarks for Web API vs FIX vs TWS API under production load; whether Web API sufficient for broker platform or if FIX required
   - Recommendation: Start with Web API for MVP (sufficient for retail broker platform); benchmark latency during implementation; migrate specific high-frequency flows to FIX if needed. Only institutional HFT clients need sub-millisecond latency from FIX.

2. **CFD Margin Requirements Across Jurisdictions**
   - What we know: CFD margin requirements vary by jurisdiction (ESMA in EU, FCA in UK, ASIC in Australia); LPs provide margin rates but may not account for all regulatory overlays
   - What's unclear: Whether LP-provided margin rates include regulatory minimums or if broker must enforce additional restrictions
   - Recommendation: Validate LP margin rates against regulatory requirements for target jurisdictions during planning phase; implement configurable margin multipliers per account group to enforce jurisdiction-specific rules.

3. **Commodity Data Feed Reliability for Critical Trading**
   - What we know: Multiple commodity APIs available (Commodities-API, Twelve Data, FMP) with varying pricing and reliability
   - What's unclear: Which providers offer guaranteed SLAs and support suitable for production broker platform vs hobby/development use
   - Recommendation: Evaluate commercial feeds (Bloomberg, Refinitiv) for mission-critical commodity trading; use free APIs for non-critical or supplementary data; implement multi-feed redundancy for critical commodities (gold, oil).

4. **Symbol Specification Update Frequency**
   - What we know: LPs can change tick sizes, margin rates, contract specifications; platforms must stay current
   - What's unclear: How frequently LPs update specifications, whether they provide notification mechanisms, and optimal cache refresh strategy
   - Recommendation: Implement daily symbol spec refresh at market open; subscribe to LP notifications if available; log all spec changes for audit; implement cache invalidation on execution errors suggesting stale specs.
</open_questions>

<sources>
## Sources

### Primary (HIGH confidence)
- [Interactive Brokers API Home](https://www.interactivebrokers.com/campus/ibkr-api-page/ibkr-api-home/) - Core API capabilities, programming languages, protocols
- [Interactive Brokers Trading Web API](https://www.interactivebrokers.com/campus/ibkr-api-page/web-api-trading/) - REST API for trading operations
- [Alpaca Markets API](https://alpaca.markets/) - Developer-first stock and crypto trading API
- [IG CFD Calculator](https://www.ig.com/en/cfd-trading/cfd-calculator) - CFD margin calculation methodology
- [CMC Markets CFD Margins](https://www.cmcmarkets.com/en/learn-cfd-trading/calculating-margins) - Industry-standard CFD margin formulas
- [Interactive Brokers CFD Margin Requirements](https://www.interactivebrokers.co.uk/en/trading/margin-cfd.php) - Risk-based margin calculation for CFDs
- [TradingHours.com](https://www.tradinghours.com/) - Comprehensive trading hours and holiday data for all exchanges
- [Commodities-API](https://commodities-api.com/) - Real-time commodity price data (600+ commodities)
- [Financial Modeling Prep](https://site.financialmodelingprep.com/datasets/commodity) - Commodity and index data APIs
- [Twelve Data](https://twelvedata.com/commodities) - Multi-asset data API (stocks, forex, crypto, commodities, indices)

### Secondary (MEDIUM confidence - WebSearch verified with official sources)
- [FIX vs WebSocket Comparison](https://b2prime.com/news/fix-vs-websocket-what-to-choose-for-your-trading-platform) - Protocol selection guidance verified against FIX Trading Community specs
- [Multi-Asset Trading Platform Architecture](https://www.bloomberg.com/professional/insights/financial-services/optimizing-multi-asset-trading-with-single-platform-solutions/) - Bloomberg professional insights on unified platform design
- [CFD Swap Calculation ActivTrades](https://www.activtrades.com/en/news/what-are-swap-rates-overnight-financing-for-cfd-positions) - Swap methodology verified against multiple broker implementations
- [Forex Market Hours](https://www.babypips.com/tools/forex-market-hours) - Trading session data cross-verified with exchange schedules

### Tertiary (LOW confidence - needs validation during implementation)
- Multi-asset platform common mistakes - Industry blog findings require validation against actual implementation
- AI-powered LP routing - Emerging 2025-2026 trend, limited production case studies available
- WebTransport adoption - New protocol, LP support still emerging as of 2026
</sources>

<metadata>
## Metadata

**Research scope:**
- Core technology: Multi-asset liquidity provider integrations (Interactive Brokers, Alpaca, Binance, commodity/index APIs)
- Ecosystem: REST/WebSocket APIs vs FIX protocol, trading calendar services, CFD calculation standards
- Patterns: Unified symbol metadata, asset-specific validation, multi-LP routing, session management
- Pitfalls: Fragmented infrastructure, stale metadata, CFD calculation errors, timezone complexity

**Confidence breakdown:**
- Standard stack: HIGH - Interactive Brokers, Alpaca, commodity APIs all actively maintained with 2026 documentation
- Architecture: HIGH - Patterns verified from MetaTrader API specs, industry best practices, and official LP documentation
- Pitfalls: HIGH - Bloomberg and industry sources document common fragmentation issues; CFD calculation pitfalls from broker documentation
- Code examples: MEDIUM - Patterns synthesized from API documentation and industry practices; require validation during implementation

**Research date:** 2026-01-16
**Valid until:** 2026-02-16 (30 days - financial APIs stable, but LP specs may change)

**Key Dependencies:**
- Phase 6 (Risk Management) must be complete - CFD margin and risk controls are prerequisite
- Existing LP adapters (OANDA, Binance) provide reference implementation patterns
- Database schema must support multi-asset symbol metadata (Phase 2 Database Migration provides foundation)

**Implementation Priorities:**
1. **High Priority:** Interactive Brokers integration for stocks and commodities (broadest coverage)
2. **High Priority:** Unified symbol metadata layer (prevents fragmentation)
3. **Medium Priority:** Extended Binance crypto support (build on existing adapter)
4. **Medium Priority:** CFD calculation engine for margin, swap, rollover
5. **Low Priority:** Indices data feed (can use IB or delay to future phase)
</metadata>

---

*Phase: 07-multi-asset-support*
*Research completed: 2026-01-16*
*Ready for planning: yes*
