# FIX Connection and LP Integration Analysis Report

## Executive Summary

The Trading Engine is configured to connect to **real liquidity providers (LPs)** and uses **production FIX protocol connections**. The system is NOT using simulated or test modes for the core trading connections, though it does support both production and demo configurations.

**Critical Finding**: The codebase contains hardcoded credentials and connection parameters for real FIX sessions, which poses significant security risks when committed to repositories.

---

## 1. FIX Connection Architecture

### 1.1 FIX Gateway Overview

**Location**: `backend/fix/gateway.go`

The FIX Gateway manages connections to Liquidity Providers through the FIX (Financial Information Exchange) protocol. It is the central hub for all LP communications.

#### Key Components:
- **LPSession struct**: Represents individual connections to an LP
- **Multiple concurrent sessions**: Each session maintains separate sequence numbers, message stores, and connection state
- **Channel-based communication**: Market data, execution reports, positions, and trades flow through dedicated channels
- **Automatic sequence number persistence**: Sequence numbers are saved to disk for crash recovery

### 1.2 Configured FIX Sessions

The system is initialized with **4 distinct FIX sessions** (from `backend/fix/gateway.go:236-314`):

#### Session 1: LMAX_PROD (Production)
```go
"LMAX_PROD": {
    ID:           "LMAX_PROD",
    Name:         "LMAX Exchange",
    Host:         "fix.lmax.com",
    Port:         443,
    SenderCompID: "RTX_BROKER",
    TargetCompID: "LMAX",
    BeginString:  "FIX.4.4",
    SSL:          true,
    Status:       "DISCONNECTED",
}
```
- **Type**: Real production connection to LMAX Exchange
- **Protocol**: FIX 4.4
- **SSL**: Enabled
- **Purpose**: Production trading via LMAX (tier-1 LP)

#### Session 2: LMAX_DEMO
```go
"LMAX_DEMO": {
    ID:           "LMAX_DEMO",
    Name:         "LMAX Demo",
    Host:         "demo-fix.lmax.com",
    Port:         443,
    SenderCompID: "RTX_BROKER_DEMO",
    TargetCompID: "LMAX_DEMO",
    BeginString:  "FIX.4.4",
    SSL:          true,
}
```
- **Type**: Demo/test connection to LMAX
- **Purpose**: Testing and development without real money

#### Session 3: YOFX1 (Production - Trading)
```go
"YOFX1": {
    ID:              "YOFX1",
    Name:            "YOFX Trading Account",
    Host:            getEnvOrDefault("YOFX_HOST", "23.106.238.138"),
    Port:            getEnvIntOrDefault("YOFX_PORT", 12336),
    SenderCompID:    getEnvOrDefault("YOFX1_SENDER_COMP_ID", "YOFX1"),
    TargetCompID:    getEnvOrDefault("YOFX_TARGET_COMP_ID", "YOFX"),
    Username:        getEnvOrDefault("YOFX1_USERNAME", "YOFX1"),
    Password:        getEnvOrDefault("YOFX1_PASSWORD", "Brand#143"),        // ⚠️ HARDCODED
    TradingAccount:  getEnvOrDefault("YOFX_TRADING_ACCOUNT", "50153"),     // ⚠️ HARDCODED
    BeginString:     "FIX.4.4",
    SSL:             false,
    UseProxy:        true,
    ProxyHost:       getEnvOrDefault("YOFX_PROXY_HOST", "81.29.145.69"),   // ⚠️ HARDCODED
    ProxyPort:       getEnvIntOrDefault("YOFX_PROXY_PORT", 49527),         // ⚠️ HARDCODED
    ProxyUsername:   getEnvOrDefault("YOFX_PROXY_USERNAME", "fGUqTcsdMsBZlms"),    // ⚠️ HARDCODED
    ProxyPassword:   getEnvOrDefault("YOFX_PROXY_PASSWORD", "3eo1qF91WA7Fyku"),    // ⚠️ HARDCODED
    Status:          "DISCONNECTED",
    ResetSeqNumFlag: getEnvOrDefault("FIX_RESET_SEQ", "false") == "true",
}
```
- **Type**: Real connection to YoForex (market data and trading)
- **Routing**: Uses proxy tunnel (81.29.145.69:49527)
- **Account**: Account number 50153
- **SSL**: False (connects through proxy)
- **⚠️ SECURITY ISSUE**: Credentials are hardcoded in source code

#### Session 4: YOFX2 (Production - Market Data)
```go
"YOFX2": {
    ID:              "YOFX2",
    Name:            "YOFX Market Data Feed",
    Host:            "23.106.238.138",
    Port:            12336,
    SenderCompID:    getEnvOrDefault("YOFX2_SENDER_COMP_ID", "YOFX2"),
    TargetCompID:    getEnvOrDefault("YOFX_TARGET_COMP_ID", "YOFX"),
    Username:        getEnvOrDefault("YOFX2_USERNAME", "YOFX2"),
    Password:        getEnvOrDefault("YOFX2_PASSWORD", "Brand#143"),        // ⚠️ HARDCODED
    TradingAccount:  getEnvOrDefault("YOFX_TRADING_ACCOUNT", "50153"),
    BeginString:     "FIX.4.4",
    SSL:             false,
    UseProxy:        true,
    ProxyHost:       "81.29.145.69",
    ProxyPort:       49527,
    ProxyUsername:   "fGUqTcsdMsBZlms",
    ProxyPassword:   "3eo1qF91WA7Fyku",
    Status:          "DISCONNECTED",
}
```
- **Type**: Real connection for market data streaming
- **Purpose**: Price feeds, symbols discovery, market hours
- **⚠️ SECURITY ISSUE**: Credentials are hardcoded in source code

---

## 2. FIX Protocol Implementation Details

### 2.1 Session Management

**Connection Manager** (`backend/fix/connection_manager.go`):
- Implements automatic reconnection with exponential backoff
- Default retry strategy: 5 seconds initial delay, 5 minute max delay, 2x multiplier
- Configurable auto-reconnect per session
- Health monitoring via heartbeat tracking

**Key Features**:
```go
type ReconnectConfig struct {
    SessionID         string
    Enabled           bool
    MaxRetries        int           // 0 = unlimited
    InitialDelay      time.Duration // 5 seconds
    MaxDelay          time.Duration // 5 minutes
    BackoffMultiplier float64       // 2.0
    AutoReconnect     bool
}
```

### 2.2 Sequence Number Management

- **Persistent storage**: Sequence numbers saved to `fixstore/` directory
- **Files**: `YOFX1.seqnums`, `YOFX2.seqnums`, `LMAX_PROD.seqnums`, `LMAX_DEMO.seqnums`
- **Format**: `OutSeqNum:InSeqNum` on single line
- **Recovery**: Automatically loaded on gateway initialization
- **Reset capability**: Can reset via environment variable `FIX_RESET_SEQ`

### 2.3 Message Store

- **In-memory store**: Maps sequence number to FIX message
- **Disk persistence**: Messages stored in `fixstore/` with `.msgs` extension
- **Purpose**: Gap recovery and resend requests
- **Retention**: Configurable per session

### 2.4 FIX Protocol Messages Supported

| Category | Message Types |
|----------|---------------|
| **Session** | Logon (A), Logout (5), Heartbeat (0), TestRequest (1) |
| **Trading** | NewOrderSingle (D), OrderCancelRequest (F), ExecutionReport (8), OrderStatusRequest (H) |
| **Market Data** | MarketDataRequest (V), MarketDataSnapshot (W), MarketDataIncremental (X), MarketDataReject (Y) |
| **Positions** | RequestForPositions (AN), PositionReport (AP) |
| **Security** | SecurityListRequest (x), SecurityList (y), SecurityDefinition (d) |
| **Trades** | TradeCaptureReport (AE), TradeCaptureRequest (AD) |

---

## 3. Liquidity Provider (LP) Integration

### 3.1 LP Manager Architecture

**Location**: `backend/lpmanager/`

The LP Manager orchestrates all liquidity provider connections and provides a unified interface to the application.

#### Core Components:

1. **LPAdapter Interface** (`backend/lpmanager/lp.go`):
```go
type LPAdapter interface {
    ID() string
    Name() string
    Type() string // REST, WebSocket, FIX
    Connect() error
    Disconnect() error
    IsConnected() bool
    GetSymbols() ([]SymbolInfo, error)
    Subscribe(symbols []string) error
    Unsubscribe(symbols []string) error
    GetQuotesChan() <-chan Quote
    GetStatus() LPStatus
}
```

2. **Manager** (`backend/lpmanager/manager.go`):
- Manages multiple LP adapters
- Configuration loaded from `data/lp_config.json`
- Provides unified quote aggregation
- Quote channel: `chan Quote` (buffer size 1000)

### 3.2 Supported LP Adapters

#### 1. OANDA (Forex)
- **Type**: REST API
- **Purpose**: Forex price feeds
- **Authentication**: API Key + Account ID
- **Location**: `backend/lpmanager/adapters/oanda.go`
- **Config Source**: `cfg.LP.OandaAPIKey`, `cfg.LP.OandaAccountID`

#### 2. Binance (Cryptocurrency)
- **Type**: REST/WebSocket
- **Purpose**: Crypto price feeds
- **Authentication**: API Key + Secret
- **Location**: `backend/lpmanager/adapters/binance.go`
- **Config Source**: `cfg.LP.BinanceAPIKey`, `cfg.LP.BinanceSecretKey`

#### 3. YoForex (FIX)
- **Type**: FIX Protocol (native integration)
- **Purpose**: Forex and commodities (XAUUSD, etc.)
- **Sessions**: YOFX1 (trading), YOFX2 (market data)
- **Connection**: Real production connection
- **Credentials**: Hardcoded in `backend/fix/gateway.go`

### 3.3 LP Configuration Structure

**Default Config** (`backend/lpmanager/lp.go:96-121`):
```go
{
    "lps": [
        {
            "id": "oanda",
            "name": "OANDA",
            "type": "OANDA",
            "enabled": true,
            "priority": 1,
            "settings": {},
            "symbols": []  // Empty = all available
        },
        {
            "id": "binance",
            "name": "Binance",
            "type": "BINANCE",
            "enabled": true,
            "priority": 2,
            "settings": {},
            "symbols": []
        }
    ],
    "primaryLp": "oanda",
    "lastModified": timestamp
}
```

### 3.4 Data Flow: LP to Application

```
LPs (OANDA/Binance/YoForex)
    ↓
LP Manager (aggregation)
    ↓
Quote Channel
    ↓
WebSocket Hub (broadcast to UI)
    ↓
Tick Store (persistence)
    ↓
B-Book Engine (pricing)
```

**Integration Point** (`backend/cmd/server/main.go:232-257`):
```go
go func() {
    for quote := range lpMgr.GetQuotesChan() {
        tick := &ws.MarketTick{
            Symbol:    quote.Symbol,
            Bid:       quote.Bid,
            Ask:       quote.Ask,
            Timestamp: quote.Timestamp,
            LP:        quote.LP,
        }
        hub.BroadcastTick(tick)           // WebSocket broadcast
        tickStore.StoreTick(...)           // Persistence
    }
}()
```

---

## 4. Connection Status and Monitoring

### 4.1 Admin Connection Management

**Location**: `backend/admin/fix_connection.go`

HTTP endpoints for managing FIX connections:

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/admin/fix/connect` | POST | Initiate connection |
| `/admin/fix/disconnect` | POST | Terminate connection |
| `/admin/fix/reconnect` | POST | Force reconnection |
| `/admin/fix/status` | GET | Get status of all sessions |
| `/admin/fix/detailed-status` | GET | Detailed stats including sequence numbers |
| `/admin/fix/enable-auto-reconnect` | POST | Enable auto-reconnect |
| `/admin/fix/disable-auto-reconnect` | POST | Disable auto-reconnect |
| `/admin/fix/connection-stats` | GET | Connection statistics |
| `/admin/fix/diagnostics` | GET | Full diagnostics (host, port, errors, etc.) |

### 4.2 Connection Statistics

**ConnectionStats struct**:
```go
type ConnectionStats struct {
    SessionID           string    // e.g., "YOFX1"
    Status              string    // "LOGGED_IN", "CONNECTING", "DISCONNECTED"
    TotalReconnects     int       // Number of reconnection attempts
    ConsecutiveFailures int       // Current failure streak
    LastError           string    // Error message from last failure
    LastSuccess         time.Time // When last login succeeded
    NextRetry           time.Time // When next retry is scheduled
    HealthStatus        string    // "HEALTHY", "DEGRADED", "UNHEALTHY"
    LastHeartbeat       time.Time // From FIX heartbeat
    AutoReconnectActive bool      // Is auto-reconnect enabled?
}
```

### 4.3 Real-World Connection Status Example

From test file `backend/cmd/connect_yofx2/main.go`:
- Connects to real YOFX2 session
- Waits for FIX logon completion
- Checks session status before operations
- Verifies sequence numbers
- Monitors heartbeat intervals

---

## 5. Credential Management and Security

### 5.1 Credential Storage

**Provisioning Service** (`backend/fix/provisioning.go`):
- AES-GCM encryption for stored credentials
- PBKDF2 key derivation (100,000 iterations)
- Encrypted storage location: `data/fix_credentials`
- Master password protection

**FIXCredentials struct**:
```go
type FIXCredentials struct {
    ID            string
    UserID        string
    SenderCompID  string
    TargetCompID  string
    Password      string    // Encrypted
    Status        string    // active, revoked, expired, suspended
    CreatedAt     time.Time
    ExpiresAt     *time.Time
    RateLimitTier string    // basic, standard, premium
    MaxSessions   int
}
```

### 5.2 CRITICAL SECURITY FINDINGS

#### Issue 1: Hardcoded Credentials in Source Code

**Location**: `backend/fix/gateway.go:267-314`

The following credentials are hardcoded in the source code:

| Item | Value | Risk Level |
|------|-------|-----------|
| YOFX1 Password | `Brand#143` | 🔴 CRITICAL |
| YOFX2 Password | `Brand#143` | 🔴 CRITICAL |
| YOFX Account | `50153` | 🔴 CRITICAL |
| Proxy Host | `81.29.145.69` | 🔴 CRITICAL |
| Proxy Port | `49527` | 🔴 CRITICAL |
| Proxy Username | `fGUqTcsdMsBZlms` | 🔴 CRITICAL |
| Proxy Password | `3eo1qF91WA7Fyku` | 🔴 CRITICAL |

**Why This is Critical**:
1. Git repository contains plaintext credentials
2. Repository may be public or shared
3. Credentials can be extracted via git history
4. Anyone with repository access has LP credentials
5. Credentials cannot be easily rotated
6. FIX sessions use these for real trading accounts

#### Issue 2: Environment Variable Fallback to Hardcoded Values

```go
Password: getEnvOrDefault("YOFX1_PASSWORD", "Brand#143"),
```

The code attempts to load from environment variables but falls back to hardcoded defaults if not set. The fallback values are the security problem.

#### Issue 3: Git History Exposure

Even if removed from current code, git history contains:
- All previous credentials
- All previous configuration values
- Complete git log showing access patterns

### 5.3 Recommended Security Mitigations

#### Immediate Actions:
1. **Rotate all exposed credentials**:
   - YOFX account passwords
   - Proxy credentials
   - All API keys

2. **Remove from git history**:
   - Use `git filter-branch` or `git filter-repo`
   - Rewrite entire history
   - Force push to origin (requires access reset)

3. **Implement credential management**:
   - Use HashiCorp Vault or similar
   - Use AWS Secrets Manager / Azure Key Vault
   - Load from environment variables only
   - Remove all hardcoded values

#### Long-term Changes:
```go
// ❌ WRONG - Current implementation
Password: getEnvOrDefault("YOFX1_PASSWORD", "Brand#143")

// ✅ CORRECT - Required environment variable
if password := os.Getenv("YOFX1_PASSWORD"); password != "" {
    session.Password = password
} else {
    return fmt.Errorf("YOFX1_PASSWORD environment variable required")
}
```

---

## 6. Real vs. Test/Demo Mode

### 6.1 Mode Configuration

The system runs in **REAL MODE by default** with these production connections:

```
YOFX1   → Real trading account #50153 @ YoForex
YOFX2   → Real market data feed @ YoForex
LMAX_PROD → Production LMAX exchange
LMAX_DEMO → Optional demo/testing
```

### 6.2 Environment Variable Control

```bash
# Production settings (actual LP connections)
YOFX_HOST=23.106.238.138           # Real YoForex IP
YOFX_PORT=12336                    # Real FIX port
YOFX_USE_PROXY=true                # Route through proxy
ENVIRONMENT=production             # Production mode

# Or test settings (via .env)
ENVIRONMENT=development            # Development mode
FIX_RESET_SEQ=false               # Preserve sequence numbers
```

### 6.3 Quote Aggregation

Real market data flows from:
1. **OANDA API** (if credentials configured) → Forex quotes
2. **Binance API** (if credentials configured) → Crypto quotes
3. **YoForex FIX** (always attempted) → Forex + Commodities
4. **Aggregated into WebSocket stream** → Frontend receives real prices

### 6.4 Simulation vs. Real Trading

The system uses **B-Book (Market Maker) engine** with real LP pricing:

```go
// B-Book engine gets REAL prices from LPs
bbookEngine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
    tick := hub.GetLatestPrice(symbol)  // From real LP feeds
    if tick != nil {
        return tick.Bid, tick.Ask, true
    }
    return 0, 0, false
})
```

**No simulation layer**: Prices are live market data, not simulated.

---

## 7. Auto-Reconnect and Failover Logic

### 7.1 Connection Manager Lifecycle

```
Session Status Flow:
DISCONNECTED
    ↓ (Connect called)
CONNECTING
    ↓ (TCP connects)
CONNECTED
    ↓ (FIX Logon sent)
LOGGED_IN
    ↓ (Ready for trading)
DISCONNECTED (on failure/disconnect)
```

### 7.2 Exponential Backoff Strategy

```go
// Initial configuration
InitialDelay: 5 seconds
MaxDelay: 5 minutes
BackoffMultiplier: 2.0

// Example retry sequence:
Attempt 1: Delay 5s   → Try connect
Attempt 2: Delay 10s  → Try connect
Attempt 3: Delay 20s  → Try connect
Attempt 4: Delay 40s  → Try connect
Attempt 5: Delay 80s  → Try connect
... (capped at 5 minutes)
```

### 7.3 Health Monitoring

- **Heartbeat Timeout**: 90 seconds (3x heartbeat interval)
- **Health Check Interval**: 10 seconds
- **Status Levels**: HEALTHY, DEGRADED, UNHEALTHY
- **Automatic Recovery**: Triggers reconnect on heartbeat timeout

---

## 8. Data Persistence

### 8.1 Sequence Number Storage

**Directory**: `backend/fixstore/`

Files:
- `YOFX1.seqnums` - YOFX1 session sequence numbers
- `YOFX2.seqnums` - YOFX2 session sequence numbers
- `YOFX1.msgs` - Message store (for gap recovery)
- `YOFX2.msgs` - Message store (for gap recovery)

### 8.2 Market Data Storage

**Directory**: `backend/data/`

Via Tick Store:
- Quote history per symbol
- OHLC calculations
- Real-time and historical access
- SQLite backend with daily rotation

---

## 9. FIX API Provisioning

### 9.1 Credential Provisioning System

**Status**: Implemented but disabled by default

```go
FIXConfig {
    ProvisioningEnabled: false,  // Feature flag
    ProvisioningStorePath: "./data/fix_credentials",
    MasterPassword: "...",       // Required if enabled
}
```

**Usage**: For issuing FIX API credentials to third-party applications

### 9.2 Provisioning Flow

1. User requests FIX credentials
2. System evaluates access rules
3. Credentials generated with:
   - Unique SenderCompID
   - Encrypted password
   - Rate limiting tier
   - Session limits
   - IP whitelist (optional)
4. Credentials stored encrypted
5. One-time plaintext display to user

---

## 10. Compliance and Monitoring

### 10.1 Compliance Configuration

```go
ComplianceConfig {
    Enabled: true,
    AuditRetentionYears: 7,
    ReportArchivePath: "./data/compliance_reports",
    AutoArchiveEnabled: true,
    TamperProofEnabled: true,
    AdminOnlyAccess: true,
    MiFIDIIEnabled: true,
    SECRule606Enabled: true,
}
```

### 10.2 Audit Logging

The system logs:
- All credential operations (generate, revoke, regenerate)
- Connection attempts (success/failure)
- Message flow (per session)
- Trade executions
- Admin actions

---

## 11. Technical Summary Table

| Aspect | Value | Notes |
|--------|-------|-------|
| **FIX Protocol Version** | FIX.4.4 | Standard for forex/commodities |
| **Primary LP** | YoForex | YOFX1/YOFX2 sessions |
| **Secondary LPs** | OANDA, Binance | Via REST/WebSocket |
| **Connection Mode** | Real/Production | Not simulated |
| **Proxy Usage** | Yes | Routes through 81.29.145.69 |
| **SSL/TLS** | Partial | Proxy tunnel instead of SSL on YOFX |
| **Sequence Numbers** | Persistent | Stored in `fixstore/` |
| **Auto-Reconnect** | Yes | Exponential backoff (5s-5m) |
| **Message Storage** | In-memory + disk | For gap recovery |
| **Health Check** | 10s interval | Heartbeat-based |
| **Credential Encryption** | AES-GCM | PBKDF2 derived key |
| **Audit Logging** | Yes | Per-operation logging |

---

## 12. Recommendations

### Critical Priority

1. **Remove hardcoded credentials** from `backend/fix/gateway.go`
2. **Rotate all exposed credentials** (YOFX, proxy)
3. **Clean git history** to remove credential exposure
4. **Implement vault/secrets manager** for credential storage
5. **Force environment variables** for all LP credentials

### High Priority

1. Implement credential rotation policies
2. Add audit logging for all FIX operations
3. Implement TLS/SSL for all connections (not proxy tunneling)
4. Add circuit breaker pattern for failed LP connections
5. Implement rate limiting per session

### Medium Priority

1. Add monitoring/alerting for connection health
2. Implement graceful degradation when LPs fail
3. Add metrics/observability for FIX message flow
4. Document all FIX session configurations
5. Implement automated backup of sequence numbers

---

## Conclusion

The Trading Engine is configured for **real production connections** to liquidity providers through FIX protocol. The system successfully handles multiple concurrent sessions, automatic reconnection, and data persistence. However, **critical security vulnerabilities** exist due to hardcoded credentials in the source code. Immediate remediation is required before production deployment.

**Key Takeaway**: The system is production-ready from a technical standpoint but requires urgent security fixes before handling real trading activity.
