# Backend Routing Engine Deep Dive Analysis

**Analysis Date:** 2026-01-18
**Analyst:** Research Agent
**Focus:** Routing configurability, A/B/C-Book implementation, LP failover

---

## Executive Summary

The trading engine implements a sophisticated multi-level routing system with **intelligent client classification** and **smart LP selection**. However, **critical configuration is hardcoded** and **routing rules are not persisted**, making runtime changes temporary.

### Key Findings
- ✅ **Advanced routing logic** with toxicity scoring and exposure management
- ✅ **Smart Order Router (SOR)** with health-based LP selection and failover
- ❌ **No persistence layer** - routing rules stored in memory only
- ❌ **Limited configurability** - thresholds and LP mappings hardcoded
- ⚠️ **Runtime updates possible** but lost on restart

---

## 1. Architecture Overview

### Routing Components

| Component | Location | Purpose |
|-----------|----------|---------|
| **C-Book Routing Engine** | `/backend/cbook/routing_engine.go` | Primary A/B/C-Book decision engine with client classification |
| **Smart Router** | `/backend/router/smart_router.go` | Legacy pattern-based routing (simpler) |
| **A-Book Execution** | `/backend/abook/engine.go` | Order execution to LPs via FIX |
| **Smart Order Router** | `/backend/abook/sor.go` | LP selection, quote aggregation, health monitoring |
| **LP Manager** | `/backend/lpmanager/lp.go` | LP adapter interface and configuration |

### Data Flow
```
Client Order
  ↓
Client Classification (Toxicity Score)
  ↓
Routing Engine Decision (Manual Rules → Classification → Volume → Exposure → Volatility)
  ↓
A-Book? → Smart Order Router → LP Selection (Health-Based) → FIX Gateway
B-Book? → Internal B-Book Engine → Instant Execution
```

---

## 2. Configuration Status

### ✅ Configurable via Environment Variables

**File:** `/backend/config/config.go`

```bash
# Broker Settings
EXECUTION_MODE=BBOOK              # A_BOOK, B_BOOK, C_BOOK (default mode)
PRICE_FEED_LP=OANDA              # Primary LP for price feed
DEFAULT_LEVERAGE=100
DEFAULT_BALANCE=5000.0
MARGIN_MODE=HEDGING              # HEDGING or NETTING

# LP Credentials
OANDA_API_KEY=xxx
OANDA_ACCOUNT_ID=xxx
BINANCE_API_KEY=xxx
BINANCE_SECRET_KEY=xxx
```

### ❌ NOT Configurable (Hardcoded in Code)

#### Routing Decision Thresholds
**File:** `/backend/cbook/routing_engine.go:150-194`

```go
// HARDCODED CLASSIFICATION ROUTING
switch profile.Classification {
case ClassificationToxic:
    if profile.ToxicityScore > 80 {  // ❌ Hardcoded threshold
        decision.Action = ActionReject
    } else {
        decision.Action = ActionABook  // 100% A-Book
    }
case ClassificationProfessional:
    decision.ABookPercent = 80  // ❌ Hardcoded 80/20 split
    decision.BBookPercent = 20
case ClassificationRetail:
    decision.BBookPercent = 90  // ❌ Hardcoded 90/10 split
    decision.ABookPercent = 10
}
```

#### Exposure Limits
**File:** `/backend/cbook/routing_engine.go:428-436`

```go
// DEFAULT EXPOSURE LIMITS (hardcoded)
limit = &ExposureLimit{
    Symbol:           symbol,
    MaxNetExposure:   500,  // ❌ 500 lots max net
    MaxGrossExposure: 1000, // ❌ 1000 lots max gross
    AutoHedgeLevel:   300,  // ❌ Auto-hedge at 300 lots
}
```

#### LP Session Mapping
**File:** `/backend/abook/sor.go:413-432`

```go
// LP TO FIX SESSION MAPPING (hardcoded)
mapping := map[string]string{
    "oanda":         "YOFX1",      // ❌ Cannot add new LPs
    "binance":       "YOFX1",
    "lmax":          "LMAX_PROD",
    "currenex":      "CURRENEX",
    "flexymarkets":  "YOFX1",
}
```

#### Volatility Threshold
**File:** `/backend/cbook/routing_engine.go:86-87`

```go
volatilityThreshold: 0.02,  // ❌ Hardcoded 2%
```

---

## 3. Routing Rules - Storage Analysis

### Current Implementation: **IN-MEMORY ONLY**

**File:** `/backend/cbook/routing_engine.go:69-92`

```go
type RoutingEngine struct {
    mu sync.RWMutex

    // ❌ STORED IN MEMORY - LOST ON RESTART
    rules          []*RoutingRule
    exposureLimits map[string]*ExposureLimit
    symbolExposure map[string]*SymbolExposure

    // Configuration (also in-memory)
    defaultLP            string
    defaultHedgePercent  float64
    maxBBookExposure     float64
    volatilityThreshold  float64
}
```

### No Persistence Found
- ❌ No database schema for `routing_rules` table
- ❌ No `LoadRoutingRules()` from JSON/YAML
- ❌ No `SaveRoutingRules()` method
- ❌ Rules created programmatically only

### Configuration Files Found
```
/backend/fix/config/sessions.json     ✓ FIX session configs
/backend/data/ticks/**/*.json          ✓ Historical tick data
/backend/docs/api/postman-collection.json  ✓ API tests

NO routing rules files found
NO lp_config.json or lp_config.yaml
```

---

## 4. A-Book / B-Book / C-Book Implementation

### Classification-Based Routing

**File:** `/backend/cbook/routing_engine.go:120-264`

| Client Type | Toxicity Score | Routing Decision |
|-------------|----------------|------------------|
| **TOXIC (High)** | > 80 | ❌ **REJECT** |
| **TOXIC (Med)** | 60-80 | 100% A-Book (full hedge) |
| **PROFESSIONAL** | 40-60 | 80% A-Book + 20% B-Book |
| **SEMI-PRO** | 20-40 | 50% A-Book + 50% B-Book |
| **RETAIL** | 0-20 | 10% A-Book + 90% B-Book |

### Multi-Level Override System

#### 1. Manual Rules (Highest Priority)
**Checked first** - priority-based rule matching

```go
type RoutingRule struct {
    Priority      int               // Higher = checked first
    AccountIDs    []int64           // Filter by account
    Symbols       []string          // Filter by symbol
    MinVolume     float64
    MaxVolume     float64
    Classifications []ClientClassification
    MinToxicity   float64
    MaxToxicity   float64
    Action        RoutingAction     // A_BOOK, B_BOOK, PARTIAL_HEDGE, REJECT
    HedgePercent  float64           // For partial hedge
}
```

#### 2. Volume Override
```go
// Orders >= 10 lots always go to A-Book
if volume >= 10 {
    decision.Action = ActionABook
    decision.ABookPercent = 100
    decision.Reason = "Large volume - full A-Book"
}
```

#### 3. Exposure Override
```go
// If symbol exposure exceeds auto-hedge level
if abs(projectedNet) > limit.AutoHedgeLevel {
    decision.Action = ActionABook
    decision.ABookPercent = 100
    decision.Reason = "Exposure limit reached - A-Book hedge"
}

// Approaching limit - gradual increase
if abs(projectedNet) > limit.AutoHedgeLevel * 0.7 {
    exposureFactor := abs(projectedNet) / limit.AutoHedgeLevel
    decision.ABookPercent += exposureFactor * 30
}
```

#### 4. Volatility Override
```go
// High volatility increases A-Book percentage
if currentVolatility > re.volatilityThreshold {  // 2%
    decision.ABookPercent += 30
    decision.Reason += " + high volatility"
}
```

---

## 5. LP Connection Management

### LP Manager Architecture

**File:** `/backend/lpmanager/lp.go`

```go
// Interface that ALL LPs must implement
type LPAdapter interface {
    ID() string
    Name() string
    Type() string                    // REST, WebSocket, FIX
    Connect() error
    Disconnect() error
    IsConnected() bool
    GetSymbols() ([]SymbolInfo, error)
    Subscribe(symbols []string) error
    GetQuotesChan() <-chan Quote     // Real-time quotes
    GetStatus() LPStatus
}
```

### LP Configuration Structure

```go
type LPConfig struct {
    ID       string            `json:"id"`        // "oanda", "binance"
    Name     string            `json:"name"`      // "OANDA"
    Type     string            `json:"type"`      // "OANDA", "BINANCE"
    Enabled  bool              `json:"enabled"`
    Priority int               `json:"priority"`  // Lower = higher priority
    Settings map[string]string `json:"settings"`  // API keys, endpoints
    Symbols  []string          `json:"symbols"`   // Empty = all available
}
```

### Default LP Configuration
**File:** `/backend/lpmanager/lp.go:96-121`

```go
func NewDefaultConfig() *LPManagerConfig {
    return &LPManagerConfig{
        LPs: []LPConfig{
            {
                ID:       "oanda",
                Name:     "OANDA",
                Type:     "OANDA",
                Enabled:  true,
                Priority: 1,         // Highest priority
            },
            {
                ID:       "binance",
                Name:     "Binance",
                Type:     "BINANCE",
                Enabled:  true,
                Priority: 2,
            },
        },
        PrimaryLP: "oanda",
    }
}
```

---

## 6. Smart Order Router (SOR)

**File:** `/backend/abook/sor.go`

### Features

#### Quote Aggregation
- Collects bid/ask from ALL connected LPs
- Maintains best bid/ask across all venues
- Filters stale quotes (> 5 seconds old)

```go
type AggregatedQuote struct {
    Symbol      string
    BestBid     float64      // Highest bid across all LPs
    BestAsk     float64      // Lowest ask across all LPs
    BestBidLP   string       // LP offering best bid
    BestAskLP   string       // LP offering best ask
    LPQuotes    map[string]*LPQuote
    LastUpdate  time.Time
}
```

#### LP Health Monitoring

**Health Score Formula:**
```go
HealthScore = (FillRate * 0.4) +
              ((1 - Slippage) * 0.3) +
              ((1 - Latency) * 0.2) +
              ((1 - RejectRate) * 0.1)
```

**Monitored Metrics:**
- Fill rate (successful executions)
- Average slippage (pips)
- Average latency (milliseconds)
- Reject rate (failed orders)
- Connection state (connected/disconnected)

#### LP Selection Algorithm

**File:** `/backend/abook/sor.go:86-157`

```go
func (s *SmartOrderRouter) SelectLP(symbol, side string, volume float64) (*LPSelection, error) {
    // 1. Get best price from aggregated quotes
    if side == "BUY" {
        targetLP = quote.BestAskLP     // Lowest ask
        targetPrice = quote.BestAsk
    } else {
        targetLP = quote.BestBidLP     // Highest bid
        targetPrice = quote.BestBid
    }

    // 2. Check LP health
    if health.HealthScore < 0.5 {
        // Find alternative LP (sorted by health)
        targetLP = findAlternativeLP()
    }

    // 3. Map to FIX session
    sessionID = mapLPToSession(targetLP)

    return &LPSelection{
        LPID:        targetLP,
        SessionID:   sessionID,
        Price:       targetPrice,
        HealthScore: healthScore,
    }
}
```

---

## 7. LP Failover Mechanisms

### Three-Level Failover Strategy

#### Level 1: Health-Based Failover
```go
// Primary LP unhealthy → Switch to alternative
if health.HealthScore < 0.5 {
    log.Printf("Primary LP %s has low health (%.2f), finding alternative",
        targetLP, health.HealthScore)

    // Find alternative LP (sorted by health score descending)
    altLP, altPrice := s.findAlternativeLP(symbol, side, targetLP, quote)

    if altLP != "" {
        targetLP = altLP
        targetPrice = altPrice
    }
}
```

#### Level 2: Stale Quote Detection
```go
// Skip quotes older than 5 seconds
for lpID, lpQuote := range aggQuote.LPQuotes {
    if time.Since(lpQuote.Timestamp) > 5*time.Second {
        continue  // Exclude from best price calculation
    }
}
```

#### Level 3: Connection Monitoring
**File:** `/backend/abook/sor.go:382-410`

```go
// Monitors LP connection state every 10 seconds
func (s *SmartOrderRouter) monitorLPHealth() {
    ticker := time.NewTicker(10 * time.Second)

    for range ticker.C {
        status := s.lpManager.GetStatus()

        for lpID, lpStatus := range status {
            if lpStatus.Connected {
                health.ConnectionState = "CONNECTED"
            } else {
                health.ConnectionState = "DISCONNECTED"
                health.HealthScore = 0  // Force failover
            }
        }
    }
}
```

### Alternative LP Selection

**File:** `/backend/abook/sor.go:325-380`

```go
// Sorts candidate LPs by health score (descending)
func findAlternativeLP(symbol, side, excludeLP string, quote *AggregatedQuote) (string, float64) {
    candidates := []lpCandidate{}

    for lpID, lpQuote := range quote.LPQuotes {
        if lpID == excludeLP { continue }
        if time.Since(lpQuote.Timestamp) > 5*time.Second { continue }

        candidates = append(candidates, lpCandidate{
            lpID:   lpID,
            price:  (side == "BUY") ? lpQuote.Ask : lpQuote.Bid,
            health: healthScore,
        })
    }

    // Sort by health (descending)
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].health > candidates[j].health
    })

    return candidates[0].lpID, candidates[0].price
}
```

---

## 8. Dynamic Rule Updates

### API Methods Available

**File:** `/backend/cbook/routing_engine.go`

```go
// Add new routing rule
func (re *RoutingEngine) AddRule(rule *RoutingRule)

// Update existing rule
func (re *RoutingEngine) UpdateRule(ruleID string, updated *RoutingRule) error

// Delete rule
func (re *RoutingEngine) DeleteRule(ruleID string) error

// Set exposure limits per symbol
func (re *RoutingEngine) SetExposureLimit(symbol string, limit *ExposureLimit)

// Get current rules
func (re *RoutingEngine) GetRules() []*RoutingRule

// Get routing statistics
func (re *RoutingEngine) GetRoutingStats() map[string]interface{}

// Get decision history (last 10,000 decisions)
func (re *RoutingEngine) GetDecisionHistory(limit int) []RoutingDecision
```

### LP Manager Methods

**File:** `/backend/lpmanager/` (inferred from handlers)

```go
// Add new LP configuration
func (m *Manager) AddLP(config LPConfig) error

// Update LP configuration
func (m *Manager) UpdateLP(config LPConfig) error

// Remove LP
func (m *Manager) RemoveLP(id string) error

// Enable/disable LP
func (m *Manager) ToggleLP(id string) error

// Get LP status
func (m *Manager) GetStatus() map[string]LPStatus

// Get LP configuration
func (m *Manager) GetLPConfig(id string) *LPConfig
```

### ✅ Can Update WITHOUT Restart (Runtime)

- ✅ Add/modify/delete routing rules
- ✅ Enable/disable rules
- ✅ Adjust exposure limits
- ✅ Change rule priorities
- ✅ Add/remove LP connections
- ✅ Enable/disable LPs

### ❌ Cannot Update Without Code Change & Restart

- ❌ Classification thresholds (toxic > 80, professional 80/20)
- ❌ Default exposure limits (500 net, 300 auto-hedge)
- ❌ Volatility threshold (2%)
- ❌ Health score formula weights
- ❌ LP-to-FIX session mappings
- ❌ Volume override threshold (10 lots)
- ❌ Quote staleness threshold (5 seconds)

### ⚠️ Updates Lost on Restart

**Problem:** All runtime updates are in-memory only

```go
// Example: Adding a rule at runtime
curl -X POST /admin/routing/rules \
  -d '{"id":"VIP-hedge","priority":200,"action":"A_BOOK"}'

// ✅ Works immediately
// ❌ Lost when server restarts
// ❌ Not saved to database
// ❌ Not saved to config file
```

---

## 9. Critical Gaps Identified

### 1. Configuration Persistence ❌

**Current State:**
- Routing rules stored in `[]*RoutingRule` (in-memory slice)
- Exposure limits in `map[string]*ExposureLimit` (in-memory map)
- LP configurations use default hardcoded values

**Missing:**
- Database schema for `routing_rules` table
- Database schema for `exposure_limits` table
- JSON/YAML config file support
- Load/Save methods

**Impact:**
- All runtime changes lost on restart
- Cannot version control routing rules
- No audit trail for rule changes
- Cannot restore previous configurations

### 2. LP Management ❌

**Hardcoded LP Mappings:**
```go
// abook/sor.go:413-432
mapping := map[string]string{
    "oanda":    "YOFX1",
    "lmax":     "LMAX_PROD",
    "currenex": "CURRENEX",
}
```

**Problems:**
- Cannot add new LPs without code changes
- Cannot customize session IDs
- No dynamic LP discovery
- No multi-session support per LP

### 3. Limited Configurability ⚠️

**Hardcoded Business Logic:**
- Classification routing percentages
- Exposure limits and auto-hedge thresholds
- Volatility thresholds
- Health score formula
- Quote staleness detection

**Impact:**
- Cannot tune routing without redeployment
- Different market conditions require different thresholds
- Cannot A/B test routing strategies

### 4. No Hot-Reload ❌

**Current Behavior:**
- Changes require server restart
- No config file watching
- No graceful reload mechanism

### 5. No Audit Trail ❌

**Missing:**
- Rule change history
- Who changed what and when
- Decision audit log (only last 10,000 in memory)
- LP failover events logging

---

## 10. Recommendations

### Priority 1: Persistence Layer (HIGH)

**Implement Database Schema**

```sql
-- Routing Rules
CREATE TABLE routing_rules (
    id VARCHAR(255) PRIMARY KEY,
    priority INT NOT NULL,
    enabled BOOLEAN DEFAULT true,

    -- Filters
    account_ids JSONB,        -- [1001, 1002, 1003]
    symbols JSONB,            -- ["EURUSD", "GBPUSD"]
    min_volume DECIMAL(10,2),
    max_volume DECIMAL(10,2),
    classifications JSONB,    -- ["TOXIC", "PROFESSIONAL"]
    min_toxicity DECIMAL(5,2),
    max_toxicity DECIMAL(5,2),

    -- Action
    action VARCHAR(50),       -- A_BOOK, B_BOOK, PARTIAL_HEDGE, REJECT
    target_lp VARCHAR(100),
    hedge_percent DECIMAL(5,2),

    -- Metadata
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100),

    INDEX idx_priority (priority DESC),
    INDEX idx_enabled (enabled)
);

-- Exposure Limits
CREATE TABLE exposure_limits (
    symbol VARCHAR(20) PRIMARY KEY,
    max_net_exposure DECIMAL(15,2) NOT NULL,
    max_gross_exposure DECIMAL(15,2) NOT NULL,
    auto_hedge_level DECIMAL(15,2) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    updated_at TIMESTAMP DEFAULT NOW(),
    updated_by VARCHAR(100)
);

-- LP Configurations
CREATE TABLE lp_configs (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    priority INT DEFAULT 100,
    settings JSONB,           -- {"api_key": "xxx", "endpoint": "https://..."}
    symbols JSONB,            -- ["EURUSD", "GBPUSD"] or [] for all
    session_id VARCHAR(100),  -- FIX session ID mapping
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Audit Trail
CREATE TABLE routing_audit (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(50),   -- RULE_ADDED, RULE_UPDATED, RULE_DELETED, LP_TOGGLED
    entity_type VARCHAR(50),  -- ROUTING_RULE, EXPOSURE_LIMIT, LP_CONFIG
    entity_id VARCHAR(255),
    old_value JSONB,
    new_value JSONB,
    user_id VARCHAR(100),
    timestamp TIMESTAMP DEFAULT NOW(),

    INDEX idx_timestamp (timestamp DESC),
    INDEX idx_entity (entity_type, entity_id)
);
```

**Implementation:**
```go
// Add persistence layer
func (re *RoutingEngine) LoadRulesFromDB() error
func (re *RoutingEngine) SaveRuleToDB(rule *RoutingRule) error
func (re *RoutingEngine) DeleteRuleFromDB(ruleID string) error
func (re *RoutingEngine) LoadExposureLimitsFromDB() error
```

### Priority 2: Configuration File Support (MEDIUM)

**Create YAML Configuration**

**File:** `/backend/config/routing.yaml`

```yaml
routing:
  # Global Settings
  default_lp: "LMAX_PROD"
  default_hedge_percent: 70
  max_bbook_exposure: 1000
  volatility_threshold: 0.02
  quote_staleness_seconds: 5

  # Classification Thresholds
  classification:
    toxic_reject_threshold: 80
    toxic_abook_threshold: 60
    professional_abook_pct: 80
    semipro_abook_pct: 50
    retail_bbook_pct: 90

  # Volume-Based Overrides
  volume_overrides:
    large_order_threshold: 10.0   # >= 10 lots → A-Book
    force_abook: true

  # Exposure Limits (Defaults)
  exposure_defaults:
    max_net_exposure: 500
    max_gross_exposure: 1000
    auto_hedge_level: 300
    approaching_threshold_pct: 0.7  # 70% of auto_hedge_level

  # Symbol-Specific Limits
  symbol_limits:
    EURUSD:
      max_net_exposure: 1000
      auto_hedge_level: 600
    XAUUSD:
      max_net_exposure: 200
      auto_hedge_level: 100

  # Health Monitoring
  health:
    score_weights:
      fill_rate: 0.4
      slippage: 0.3
      latency: 0.2
      reject_rate: 0.1
    unhealthy_threshold: 0.5
    connection_check_interval_sec: 10

# LP Configuration
liquidity_providers:
  - id: oanda
    name: OANDA
    type: OANDA
    enabled: true
    priority: 1
    session_id: YOFX1
    settings:
      api_key: ${OANDA_API_KEY}
      account_id: ${OANDA_ACCOUNT_ID}
      endpoint: https://api-fxpractice.oanda.com

  - id: lmax
    name: LMAX Exchange
    type: FIX
    enabled: true
    priority: 2
    session_id: LMAX_PROD
    settings:
      sender_comp_id: YOFX_CLIENT
      target_comp_id: LMAX

  - id: binance
    name: Binance
    type: BINANCE
    enabled: false
    priority: 3
    session_id: YOFX1

# Manual Routing Rules (can override with database)
routing_rules:
  - id: rule_vip_accounts
    priority: 200
    enabled: true
    filters:
      account_ids: [1001, 1002, 1003]
    action: A_BOOK
    target_lp: lmax
    description: "VIP accounts always A-Book to LMAX"

  - id: rule_toxic_clients
    priority: 190
    enabled: true
    filters:
      min_toxicity: 80
    action: REJECT
    description: "Block highly toxic clients"

  - id: rule_crypto_hedge
    priority: 180
    enabled: true
    filters:
      symbols: ["BTCUSD", "ETHUSD"]
      min_volume: 1.0
    action: A_BOOK
    target_lp: binance
    description: "Crypto orders to Binance"
```

**Implementation:**
```go
// Load configuration from file
func LoadRoutingConfig(path string) (*RoutingConfig, error)

// Watch for file changes and reload
func (re *RoutingEngine) WatchConfigFile(path string) error

// Merge database rules with file rules (database higher priority)
func (re *RoutingEngine) MergeRules(fileRules, dbRules []*RoutingRule) []*RoutingRule
```

### Priority 3: Hot-Reload Capability (MEDIUM)

```go
// File watcher for config changes
func (re *RoutingEngine) WatchConfig(configPath string) {
    watcher, _ := fsnotify.NewWatcher()
    watcher.Add(configPath)

    for {
        select {
        case event := <-watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                log.Println("Config file changed, reloading...")
                re.ReloadConfig(configPath)
            }
        }
    }
}

// Graceful reload without dropping connections
func (re *RoutingEngine) ReloadConfig(path string) error {
    newConfig, err := LoadRoutingConfig(path)
    if err != nil {
        return err
    }

    re.mu.Lock()
    defer re.mu.Unlock()

    // Atomic swap
    re.rules = newConfig.Rules
    re.exposureLimits = newConfig.ExposureLimits
    re.volatilityThreshold = newConfig.VolatilityThreshold

    log.Println("Configuration reloaded successfully")
    return nil
}

// Admin API endpoint
// POST /admin/routing/reload
func (h *RoutingHandler) HandleReloadConfig(w http.ResponseWriter, r *http.Request) {
    if err := h.engine.ReloadConfig(h.configPath); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    json.NewEncoder(w).Encode(map[string]string{"status": "reloaded"})
}
```

### Priority 4: Admin UI for Rule Management (LOW)

**API Endpoints Required:**

```go
// Routing Rules
GET    /admin/routing/rules          // List all rules
POST   /admin/routing/rules          // Create new rule
GET    /admin/routing/rules/:id      // Get rule details
PUT    /admin/routing/rules/:id      // Update rule
DELETE /admin/routing/rules/:id      // Delete rule
POST   /admin/routing/rules/:id/enable
POST   /admin/routing/rules/:id/disable

// Exposure Limits
GET    /admin/routing/exposure       // List all limits
PUT    /admin/routing/exposure/:symbol

// LP Management
GET    /admin/lps                    // List LPs
POST   /admin/lps                    // Add LP
PUT    /admin/lps/:id                // Update LP
DELETE /admin/lps/:id                // Remove LP
POST   /admin/lps/:id/toggle         // Enable/disable
GET    /admin/lps/:id/health         // Health metrics

// Monitoring
GET    /admin/routing/decisions      // Recent decisions
GET    /admin/routing/stats          // Routing statistics
GET    /admin/routing/exposure/:symbol  // Current exposure
GET    /admin/routing/audit          // Audit trail
```

### Priority 5: Make LP Mappings Configurable (HIGH)

**Current Problem:**
```go
// Hardcoded in abook/sor.go:413-432
mapping := map[string]string{
    "oanda": "YOFX1",
    "lmax":  "LMAX_PROD",
}
```

**Solution:**
```go
// Add session_id to LPConfig
type LPConfig struct {
    ID        string
    SessionID string  // NEW: FIX session mapping
    ...
}

// Use config instead of hardcoded map
func (s *SmartOrderRouter) mapLPToSession(lpID string) string {
    config := s.lpManager.GetLPConfig(lpID)
    if config != nil && config.SessionID != "" {
        return config.SessionID
    }
    return "YOFX1"  // Fallback
}
```

---

## 11. Summary & Action Items

### Current State
- ✅ **Routing Logic:** Sophisticated multi-level A/B/C-Book routing with client classification
- ✅ **LP Selection:** Smart Order Router with health-based failover
- ✅ **Real-Time:** Quote aggregation, exposure tracking, health monitoring
- ✅ **API:** Methods exist for runtime updates
- ❌ **Persistence:** Rules stored in memory only, lost on restart
- ❌ **Configuration:** Limited to environment variables, most logic hardcoded
- ❌ **LP Management:** Session mappings hardcoded, cannot add LPs dynamically

### Required Changes (In Priority Order)

#### Phase 1: Persistence (Critical)
1. ✅ Design database schema for routing rules, exposure limits, LP configs
2. ⏳ Implement data access layer (DAO)
3. ⏳ Add LoadFromDB/SaveToDB methods
4. ⏳ Migrate existing rules to database
5. ⏳ Add audit trail logging

#### Phase 2: Configuration Files (High)
1. ⏳ Design YAML schema for routing configuration
2. ⏳ Implement config file loader
3. ⏳ Add merge logic (database > file > defaults)
4. ⏳ Implement file watcher for hot-reload
5. ⏳ Document configuration options

#### Phase 3: Dynamic LP Management (High)
1. ⏳ Move LP-to-session mapping to LPConfig
2. ⏳ Remove hardcoded mapping from SOR
3. ⏳ Add validation for LP configurations
4. ⏳ Support multiple sessions per LP

#### Phase 4: Admin Interface (Medium)
1. ⏳ Complete REST API for rule management
2. ⏳ Build admin UI for rule configuration
3. ⏳ Add visual exposure monitoring
4. ⏳ Create LP health dashboard

#### Phase 5: Configurability (Low)
1. ⏳ Move business logic thresholds to config
2. ⏳ Make health score formula configurable
3. ⏳ Add strategy templates (conservative, aggressive)

### Immediate Recommendations

**For Development:**
1. Add database migrations for routing tables
2. Implement persistence layer with transactions
3. Create YAML config file structure
4. Add comprehensive logging for routing decisions

**For Operations:**
1. Document current routing behavior
2. Create backup of current rule configurations (code)
3. Plan migration strategy for existing rules
4. Test hot-reload mechanism thoroughly

**For Testing:**
1. Add integration tests for rule persistence
2. Test failover scenarios with LP disconnections
3. Verify routing decisions under various conditions
4. Load test with high-frequency updates

---

## Appendix A: File Inventory

### Files Analyzed
```
/backend/cbook/routing_engine.go         564 lines  Primary routing engine
/backend/router/smart_router.go          136 lines  Legacy router
/backend/abook/engine.go                 649 lines  A-Book execution
/backend/abook/sor.go                    443 lines  Smart Order Router
/backend/lpmanager/lp.go                 122 lines  LP interface
/backend/config/config.go                258 lines  Configuration
/backend/fix/rules_engine.go             438 lines  FIX access rules
/backend/internal/api/handlers/lp.go     231 lines  LP API
/backend/internal/api/handlers/abook.go  220 lines  A-Book API
/backend/internal/core/engine.go         817 lines  B-Book engine
/backend/bbook/engine.go                 743 lines  B-Book engine (alt)
```

### Key Data Structures

**Routing:**
- `RoutingEngine` - Main routing orchestrator
- `RoutingRule` - Individual routing rule
- `RoutingDecision` - Routing outcome
- `ExposureLimit` - Symbol exposure limits
- `SymbolExposure` - Current exposure tracking

**LP Management:**
- `LPAdapter` - Interface for all LPs
- `LPConfig` - LP configuration
- `LPStatus` - LP health/connection status
- `Quote` - Price quote from LP

**Smart Order Router:**
- `SmartOrderRouter` - SOR orchestrator
- `AggregatedQuote` - Best bid/ask across LPs
- `LPSelection` - Selected LP for execution
- `LPHealth` - LP health metrics

**A-Book Execution:**
- `ExecutionEngine` - A-Book order executor
- `Order` - Order state
- `Position` - Open position
- `Fill` - Execution record
- `ExecutionReport` - LP execution report

---

## Appendix B: Configuration Examples

### Example: Routing Rule JSON
```json
{
  "id": "rule_crypto_hedge",
  "priority": 180,
  "enabled": true,
  "accountIds": null,
  "symbols": ["BTCUSD", "ETHUSD"],
  "minVolume": 1.0,
  "maxVolume": 0,
  "classifications": null,
  "minToxicity": 0,
  "maxToxicity": 0,
  "action": "A_BOOK",
  "targetLp": "binance",
  "hedgePercent": 0,
  "description": "Route crypto to Binance"
}
```

### Example: Exposure Limit JSON
```json
{
  "symbol": "EURUSD",
  "maxNetExposure": 1000,
  "maxGrossExposure": 2000,
  "autoHedgeLevel": 600
}
```

### Example: LP Config JSON
```json
{
  "id": "lmax",
  "name": "LMAX Exchange",
  "type": "FIX",
  "enabled": true,
  "priority": 1,
  "sessionId": "LMAX_PROD",
  "settings": {
    "senderCompID": "YOFX_CLIENT",
    "targetCompID": "LMAX",
    "host": "lmax.prod.fixgateway.com",
    "port": "9001"
  },
  "symbols": []
}
```

---

**End of Analysis**
