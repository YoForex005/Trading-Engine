# Admin Panel Configuration Analysis Report
**Date:** 2026-01-18
**Component:** Admin Panel (AdminPanel.tsx) and Backend Configuration System
**Status:** Complete

---

## Executive Summary

The trading engine has **partial dynamic configuration** capabilities through the Admin Panel, but significant broker settings remain **hardcoded** and require code changes or server restarts to modify. The backend infrastructure for advanced configuration exists but is not fully exposed to the admin UI.

### Overall Assessment
- **Dynamic Configuration:** ~40% (basic broker settings, LP toggle, symbol enable/disable)
- **Hardcoded Configuration:** ~60% (symbol specs, commissions, routing rules, credentials)
- **Missing Admin Controls:** 14 critical settings not exposed

---

## 1. HARDCODED VALUES

### 1.1 Frontend Hardcoded Values
| Location | Value | Impact |
|----------|-------|--------|
| `AdminPanel.tsx:72,87,99,114,136,151` | `http://localhost:8080` | API base URL hardcoded (not env-configurable) |
| `AdminPanel.tsx:340` | Default leverage fallback `100` | UI defaults if API fails |
| `AdminPanel.tsx:168` | Message timeout `3000ms` | Toast notification duration |

### 1.2 Backend Hardcoded Values

#### Symbol Specifications (bbook/symbols.go)
All symbol specifications are **auto-generated** based on naming patterns, not persistently stored:

| Symbol Category | Hardcoded Settings |
|-----------------|-------------------|
| **Forex Majors** | Contract Size: 100,000 / Margin: 1% / Pip: 0.0001 / Pip Value: $10 |
| **Forex Minors** | Contract Size: 100,000 / Margin: 1% / Variable pip values |
| **Forex Exotics** | Contract Size: 100,000 / Margin: 3% / Pip Value: $5-10 |
| **Crypto** | Contract Size: 1 / Margin: 10% / BTC Pip: $1, ETH: $0.1 |
| **Metals** | XAU: 100 oz / Margin: 2% / XAG: 5,000 oz |
| **Commodities** | Contract Size: 1,000 / Margin: 5% / Pip: $10 |
| **Indices** | Contract Size: 1 / Margin: 5% / Pip: $0.1 |
| **Bonds** | Contract Size: 1,000 / Margin: 2% / Pip: $10 |

**Problem:** Admin cannot override these values without code changes.

#### Commission Rates (bbook/symbols.go)
- **SymbolSpec.CommissionPerLot** - Generated per symbol, not admin-configurable
- No volume-based commission schedules
- No client-tier commission structures

#### Other Backend Hardcoded Values
- **Server Port:** `7999` (config.go, but requires restart)
- **Demo Account Creation:** Logic in `main.go:86-92` (balance from config)
- **Pip Calculation Logic:** Embedded in `GenerateSymbolSpec()` function
- **Margin Calculation:** `(Volume * ContractSize * Price) / Leverage` - formula hardcoded

---

## 2. MISSING ADMIN CONTROLS

### 2.1 Critical Missing Features (14 Items)

| # | Missing Control | Backend Support | UI Gap |
|---|----------------|-----------------|--------|
| 1 | **LP Routing Priority** | ✅ `LPConfig.Priority` exists | ❌ No UI to set priority order |
| 2 | **LP Symbol Assignment** | ✅ `LPConfig.Symbols` array | ❌ No UI to assign symbols to LPs |
| 3 | **Symbol-Level Routing** | ⚠️ Partial (no routing rules table) | ❌ No UI for "EURUSD → OANDA" rules |
| 4 | **Commission Per Symbol** | ❌ Auto-generated only | ❌ No API endpoint to override |
| 5 | **Spread Markups** | ❌ Uses raw LP spreads | ❌ No spread adjustment controls |
| 6 | **Margin Requirements** | ❌ Category-based only | ❌ Cannot override per symbol |
| 7 | **Contract Sizes** | ❌ Auto-generated | ❌ Cannot edit contract sizes |
| 8 | **Pip Values/Sizes** | ❌ Auto-generated | ❌ Cannot customize pip values |
| 9 | **Volume Limits** | ✅ `MinVolume`, `MaxVolume` exist | ❌ No UI to edit limits |
| 10 | **LP Credentials** | ⚠️ Env vars only | ❌ No secure credential manager UI |
| 11 | **Per-Account Execution Mode** | ❌ Global ABOOK/BBOOK only | ❌ Cannot set A-Book for VIP accounts |
| 12 | **Swap/Rollover Rates** | ❌ Not implemented | ❌ No swap configuration |
| 13 | **Trading Hours** | ❌ Not implemented | ❌ No session time management |
| 14 | **Symbol Specs Persistence** | ❌ Regenerated on restart | ❌ No save/load symbol configs |

---

## 3. DYNAMIC vs. STATIC CONFIGURATION

### 3.1 Currently Dynamic (Admin Can Change Without Code)
✅ **Broker Name** - Via `/api/config` POST
✅ **Execution Mode** - Global A-Book/B-Book toggle
✅ **Default Leverage** - New account default
✅ **Default Balance** - New account starting balance
✅ **Margin Mode** - HEDGING vs NETTING
✅ **LP Enable/Disable** - Via `/admin/lps/{id}/toggle`
✅ **Symbol Enable/Disable** - Via `/admin/symbols/toggle`

### 3.2 Currently Static (Requires Code/Restart)
❌ **Symbol Specifications** - Auto-generated from naming patterns
❌ **Commission Rates** - Per symbol type, not configurable
❌ **Margin Percentages** - By asset class only
❌ **LP Routing Rules** - No symbol-to-LP assignment UI
❌ **API Keys/Credentials** - Environment variables only
❌ **Server Configuration** - Port, URLs (env vars, restart required)
❌ **Swap Rates** - Not implemented
❌ **Trading Sessions** - No time-based controls

---

## 4. API ENDPOINTS ANALYSIS

### 4.1 Existing Admin API Endpoints
| Endpoint | Method | Purpose | Status |
|----------|--------|---------|--------|
| `/api/admin/config` | GET | Get broker config | ✅ Working |
| `/api/admin/config` | POST | Update broker config | ✅ Working |
| `/admin/lps` | GET | List LP configs | ✅ Working |
| `/admin/lps` | POST | Add new LP | ✅ Working |
| `/admin/lps/{id}` | PUT | Update LP | ✅ Working |
| `/admin/lps/{id}` | DELETE | Remove LP | ✅ Working |
| `/admin/lps/{id}/toggle` | POST | Enable/disable LP | ✅ Working |
| `/admin/lp-status` | GET | Get LP connection status | ✅ Working |
| `/admin/symbols` | GET | List all symbols | ✅ Working |
| `/admin/symbols/toggle` | POST | Enable/disable symbol | ✅ Working |
| `/admin/execution-mode` | GET/POST | A-Book/B-Book toggle | ✅ Working |

### 4.2 Missing API Endpoints (NEEDED)
| Endpoint | Method | Purpose | Priority |
|----------|--------|---------|----------|
| `/api/admin/lp/{id}/priority` | PUT | Set LP routing priority | HIGH |
| `/api/admin/lp/{id}/symbols` | PUT | Assign symbols to LP | HIGH |
| `/api/admin/routing/rules` | GET/POST | Symbol-to-LP routing table | HIGH |
| `/api/admin/symbols/{symbol}/commission` | PUT | Override commission | HIGH |
| `/api/admin/symbols/{symbol}/spread` | PUT | Set spread markup | MEDIUM |
| `/api/admin/symbols/{symbol}/margin` | PUT | Override margin % | MEDIUM |
| `/api/admin/symbols/{symbol}/specs` | PUT | Edit contract size, pip value | MEDIUM |
| `/api/admin/symbols/{symbol}/limits` | PUT | Edit volume min/max/step | LOW |
| `/api/admin/symbols/{symbol}/swap` | PUT | Configure swap rates | LOW |
| `/api/admin/symbols/{symbol}/hours` | PUT | Set trading hours | LOW |
| `/api/admin/account/{id}/execution-mode` | PUT | Per-account A-Book/B-Book | HIGH |
| `/api/admin/credentials/lp/{id}` | PUT | Manage LP API keys securely | MEDIUM |

---

## 5. ARCHITECTURE GAPS

### 5.1 Symbol Management Issues
**Problem:** Symbol specifications are **ephemeral** (regenerated on startup)
- Generated by `GenerateSymbolSpec()` based on naming patterns
- Not persisted to database or config file
- Admin changes are lost on restart

**Solution Needed:**
1. Add `symbols` table to database
2. Add `/api/admin/symbols/{symbol}/specs` endpoint
3. Save admin overrides persistently
4. Fall back to auto-generation if no override exists

### 5.2 LP Routing Issues
**Problem:** No granular symbol-to-LP routing
- `LPConfig.Symbols` array exists but is not used for routing
- `LPConfig.Priority` exists but no UI to manage it
- Cannot route EURUSD → OANDA and BTCUSD → Binance

**Solution Needed:**
1. Create `RoutingRule` struct (symbol, LP ID, priority)
2. Add routing rules table/config
3. Implement routing engine that respects rules
4. Build UI for drag-and-drop routing management

### 5.3 Per-Account Execution Mode
**Problem:** Global A-Book/B-Book toggle affects all accounts
- Cannot have VIP accounts on A-Book and retail on B-Book
- `Account` struct doesn't have `ExecutionMode` field

**Solution Needed:**
1. Add `ExecutionMode` field to `Account` struct
2. Update order execution logic to check account-level mode
3. Add `/api/admin/account/{id}/execution-mode` endpoint
4. Update Admin Panel to show per-account execution mode

### 5.4 Credential Management
**Problem:** LP credentials are environment variables only
- Admin must edit `.env` file and restart server
- No secure in-app credential manager
- No credential rotation workflow

**Solution Needed:**
1. Encrypt credentials in database
2. Add secure credential update endpoint
3. Hot-reload LP connections on credential change
4. Build credential management UI with masked inputs

---

## 6. DETAILED FINDINGS

### 6.1 AdminPanel Component (clients/desktop/src/components/AdminPanel.tsx)

**Structure:**
- 3 tabs: Broker Config, Liquidity Providers, Symbols
- 539 lines total
- Uses direct `fetch()` calls (not centralized API service)

**Broker Config Tab:**
✅ Configurable:
- Broker Name
- Routing Mode (A-Book/B-Book/Hybrid) - **UI shows 3 options but backend only supports ABOOK/BBOOK**
- Execution Mode (INSTANT/MARKET/REQUOTE)
- Default Leverage
- Margin Mode (HEDGED/NETTING)
- Price Feed LP (text input)
- Execution LP (text input)
- Slippage (pips)
- Commission (USD/lot)

❌ Issues:
- "Hybrid" routing mode shown but not implemented in backend
- LP assignment is text input, not dropdown from available LPs
- No validation on LP names
- No per-symbol routing configuration

**LP Tab:**
✅ Working:
- List all LPs with status (ACTIVE/INACTIVE/ERROR)
- Show latency, uptime, symbol count
- Enable/Disable toggle

❌ Missing:
- Add new LP (form exists in backend, not in UI)
- Edit LP priority
- Assign symbols to LP
- Configure LP credentials
- Test LP connection

**Symbols Tab:**
✅ Working:
- List all symbols
- Enable/Disable per symbol
- Show current spread, commission, leverage

❌ Missing:
- Edit spread, commission, leverage per symbol
- Edit volume limits (min/max)
- Edit contract size, pip value
- Assign symbol to specific LP
- Configure swap rates
- Set trading hours

### 6.2 Backend Configuration System (backend/config/config.go)

**Well-Designed Structure:**
```go
type Config struct {
    Broker         BrokerConfig
    LP             LPConfig
    Database       DatabaseConfig
    Admin          AdminConfig
    JWT            JWTConfig
    Encryption     EncryptionConfig
    FIX            FIXConfig
}
```

✅ **Strengths:**
- Comprehensive config structure
- Environment variable loading with defaults
- Validation in production mode
- Supports multiple LP providers

❌ **Weaknesses:**
- All config requires server restart
- No hot-reload mechanism
- Credentials in env vars (not encrypted)
- No config versioning/history

### 6.3 LP Manager (backend/lpmanager/)

**Architecture:**
- `LPAdapter` interface for pluggable providers
- `LPConfig` with Priority and Symbols array
- Config stored in `data/lp_config.json`
- Supports OANDA, Binance adapters

✅ **Strengths:**
- Priority-based routing (field exists)
- Per-LP symbol filtering (field exists)
- Hot-reloadable config (JSON file)
- Admin API endpoints implemented

❌ **Gaps:**
- Priority not used in routing logic
- Symbols array not enforced
- No routing rules engine
- No fallback/failover logic

### 6.4 Symbol System (backend/bbook/symbols.go)

**Auto-Generation Logic:**
1. Detect category from symbol name (e.g., "BTC" → CRYPTO)
2. Apply category-based defaults
3. Generate `SymbolSpec` dynamically
4. No persistence (lost on restart)

**Categories Detected:**
- FOREX_MAJOR (7 pairs)
- FOREX_MINOR (cross pairs)
- FOREX_EXOTIC (TRY, ZAR, MXN, etc.)
- CRYPTO (BTC, ETH, BNB, etc.)
- METALS (XAU, XAG, XPT, XPD, XCU)
- COMMODITIES (BCO, WTICO, NATGAS, etc.)
- INDICES (US30, NAS100, SPX500, etc.)
- BONDS (USB, UK10YB, etc.)

**Strengths:**
✅ Automatic symbol discovery
✅ Smart category detection
✅ Sensible defaults per category

**Weaknesses:**
❌ No admin override capability
❌ No persistence (ephemeral)
❌ Cannot handle non-standard symbols
❌ Commission rates cannot be customized

---

## 7. RECOMMENDATIONS

### 7.1 HIGH PRIORITY (Production Critical)
1. **Persist Symbol Specifications**
   - Create `symbols` database table
   - Store admin overrides
   - Fall back to auto-generation if no override

2. **Implement Symbol-to-LP Routing**
   - Build routing rules table
   - Expose `/api/admin/routing/rules` endpoint
   - Add UI for drag-and-drop routing management

3. **Per-Account Execution Mode**
   - Add `ExecutionMode` to `Account` struct
   - Allow VIP accounts on A-Book, retail on B-Book
   - Update Admin Panel to show per-account mode

4. **LP Priority Management**
   - Expose `LPConfig.Priority` in Admin UI
   - Implement priority-based routing in execution logic
   - Add failover to backup LP if primary fails

5. **Symbol Configuration UI**
   - Add edit modals for commission, spread, margin
   - Expose volume limits (min/max/step)
   - Allow override of contract size, pip value

### 7.2 MEDIUM PRIORITY (Enhanced Control)
6. **Credential Management**
   - Encrypt LP credentials in database
   - Build secure credential update UI
   - Hot-reload LP connections on change

7. **Spread Markup Configuration**
   - Add per-symbol spread markup settings
   - Apply markup on top of LP raw spread
   - Track markup in separate field

8. **Volume-Based Commission Tiers**
   - Implement tiered commission structure
   - Configure per account or per symbol
   - Support volume rebates

### 7.3 LOW PRIORITY (Nice to Have)
9. **Swap Rate Configuration**
   - Add swap calculation engine
   - Configure long/short swap per symbol
   - Apply swap at rollover (5pm EST)

10. **Trading Hours Management**
    - Configure session times per symbol
    - Auto-disable trading outside hours
    - Show market status in UI

11. **Config Versioning**
    - Track config changes with timestamps
    - Allow rollback to previous config
    - Audit log for admin actions

---

## 8. IMPLEMENTATION ROADMAP

### Phase 1: Symbol Persistence (Week 1)
- Create `symbols` table schema
- Add CRUD endpoints for symbol specs
- Update engine to load from DB
- Add edit UI in Admin Panel

### Phase 2: LP Routing (Week 2)
- Create routing rules data structure
- Implement routing engine
- Add priority-based selection logic
- Build routing UI (drag-and-drop)

### Phase 3: Account-Level Controls (Week 3)
- Add `ExecutionMode` to accounts
- Update order execution logic
- Add per-account admin controls
- Test VIP A-Book routing

### Phase 4: Advanced Configuration (Week 4)
- Credential management UI
- Spread markup configuration
- Commission tier system
- Config audit logging

---

## 9. CODE QUALITY NOTES

### Strengths
✅ Clean separation: Frontend (AdminPanel.tsx) → API → Backend handlers
✅ Modular LP adapter architecture
✅ Comprehensive config system (config.go)
✅ Auto-symbol generation is clever (bbook/symbols.go)

### Areas for Improvement
⚠️ **Type Mismatch:** AdminPanel shows "Hybrid" mode not supported in backend
⚠️ **Hardcoded URLs:** API base URL not env-configurable in frontend
⚠️ **No Validation:** LP names are text input without validation
⚠️ **Ephemeral Specs:** Symbol specs lost on restart
⚠️ **No Error Handling:** Missing fallback UI for API failures
⚠️ **No Optimistic Updates:** UI doesn't update until API confirms

### Security Concerns
🔒 **Credentials in .env:** Should be encrypted in database
🔒 **No API Auth:** Admin endpoints lack authentication/authorization
🔒 **CORS Wide Open:** `Access-Control-Allow-Origin: *`
🔒 **No Rate Limiting:** Admin API vulnerable to abuse

---

## 10. CONCLUSION

The trading engine has a **solid foundation** for dynamic configuration but is **60% incomplete** for production broker needs. The backend infrastructure (LP Manager, Config System) supports advanced features but lacks:

1. **UI Exposure** - Many backend fields not in admin UI
2. **Persistence** - Symbol specs and routing rules are ephemeral
3. **Granularity** - Cannot configure per-symbol or per-account
4. **Security** - Credentials and auth need hardening

**Estimated Effort to Production-Ready Admin Panel:**
- Phase 1-2 (Symbol + Routing): **2-3 weeks**
- Phase 3-4 (Account + Advanced): **2 weeks**
- **Total: 4-5 weeks** for full dynamic configuration

**Business Impact:**
- With these fixes, broker can **reconfigure live** without code deploys
- Enables **A-Book for VIPs, B-Book for retail** (dual-mode brokerage)
- **Symbol-level routing** reduces dependency on single LP
- **Credential hot-reload** improves operational agility

---

**END OF REPORT**
