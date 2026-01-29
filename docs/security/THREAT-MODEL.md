# Threat Model - RTX Trading Engine

**Classification:** CONFIDENTIAL
**Version:** 1.0
**Last Updated:** 2026-01-18
**Methodology:** STRIDE (Microsoft Threat Modeling)

---

## Executive Summary

This document identifies potential security threats to the RTX Trading Engine based on STRIDE methodology, assesses their risk levels, and documents mitigation strategies. The trading engine handles sensitive financial data and real-time trading operations, requiring comprehensive threat analysis.

**Key Findings:**
- **8 CVE-level vulnerabilities** identified (CRITICAL priority)
- **12 HIGH-severity threats** requiring immediate mitigation
- **15 MEDIUM-severity threats** requiring Phase 2 implementation
- Attack surface spans API endpoints, WebSocket connections, database, and external LP integrations

---

## System Overview

### Architecture Components

```
┌─────────────────────────────────────────────────────────────┐
│                    EXTERNAL ACTORS                           │
│  - Traders (authenticated users)                            │
│  - Admins (broker operators)                                │
│  - Attackers (malicious actors)                             │
│  - Liquidity Providers (OANDA, YoFx, Binance)              │
└─────────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────────┐
│                    TRUST BOUNDARY 1                          │
│                    (Public Internet)                         │
└─────────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────────┐
│  COMPONENT 1: HTTP API Server (Go)                          │
│  - Authentication (/login)                                   │
│  - Trading endpoints (/api/orders, /api/positions)          │
│  - Admin endpoints (/admin/deposit, /admin/withdraw)        │
│  - Market data (/ticks, /ohlc)                              │
└─────────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────────┐
│  COMPONENT 2: WebSocket Hub (Real-time prices)              │
│  - Price feed aggregation from multiple LPs                 │
│  - Client connections (ws://localhost:7999/ws)              │
│  - Tick distribution to subscribers                         │
└─────────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────────┐
│  COMPONENT 3: B-Book Engine (In-Memory)                     │
│  - Order execution                                           │
│  - Position management                                       │
│  - Balance/equity calculation                                │
│  - Ledger (transaction log)                                 │
└─────────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────────┐
│  COMPONENT 4: External LP Adapters                          │
│  - OANDA (REST API)                                          │
│  - YoFx (FIX 4.4)                                            │
│  - Binance (WebSocket)                                       │
└─────────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────────┐
│                    TRUST BOUNDARY 2                          │
│                (Internal Data Storage)                       │
└─────────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────────┐
│  COMPONENT 5: Data Persistence (Future)                     │
│  - PostgreSQL (accounts, orders, positions, trades)         │
│  - Redis (cache, session store)                             │
│  - File system (tick data JSON files)                       │
└─────────────────────────────────────────────────────────────┘
```

---

## STRIDE Threat Analysis

### S - Spoofing Identity

#### Threat S1: JWT Token Forgery
**Description:** Attacker crafts fake JWT token to impersonate legitimate user

**Attack Vector:**
1. Attacker discovers JWT secret from source code (CVE-1)
2. Attacker uses jwt.io to generate token with admin role
3. Attacker accesses admin endpoints to manipulate funds

**CVSS Score:** 9.8 (Critical)

**Current Vulnerability:**
```go
// VULNERABLE: Weak default JWT secret
var jwtKey = []byte("super_secret_dev_key_do_not_use_in_prod")
```

**Mitigation:**
- Use RS256 asymmetric signing (private key signs, public key verifies)
- Rotate keys regularly (every 90 days)
- Store private key in HashiCorp Vault
- Include IP binding in JWT claims
- Implement token revocation list (Redis)

**Implementation Priority:** CRITICAL (Week 1, Day 2)

---

#### Threat S2: Session Hijacking
**Description:** Attacker steals session token via XSS or network sniffing

**Attack Vector:**
1. Attacker injects JavaScript via reflected XSS in error messages
2. Script steals JWT from localStorage
3. Attacker uses stolen token to place unauthorized trades

**CVSS Score:** 8.1 (High)

**Mitigation:**
- Use httpOnly, secure, SameSite cookies (not localStorage)
- Implement Content Security Policy (CSP)
- Enforce TLS 1.3 (no plaintext transmission)
- Bind sessions to IP address and User-Agent
- Implement anomaly detection (unusual IP/location)

**Implementation Priority:** HIGH (Week 1, Day 4)

---

### T - Tampering with Data

#### Threat T1: Order Manipulation
**Description:** Attacker modifies order parameters in transit or in database

**Attack Vector:**
1. Attacker intercepts HTTP request (no TLS)
2. Modifies order from BUY 0.01 lots to BUY 1000 lots
3. System executes manipulated order

**CVSS Score:** 8.8 (High)

**Mitigation:**
- Enforce TLS 1.3 for all communications
- Implement HMAC signatures on order messages
- Database transaction integrity checks
- Order parameter validation (volume limits)
- Write-ahead logging for audit trail

**Implementation Priority:** CRITICAL (Week 1, Day 4)

---

#### Threat T2: SQL Injection
**Description:** Attacker injects malicious SQL to modify database records

**Attack Vector:**
```go
// VULNERABLE CODE
symbol := r.URL.Query().Get("symbol")
query := "SELECT * FROM orders WHERE symbol = '" + symbol + "'"
// Attacker sends: symbol=EURUSD'; UPDATE accounts SET balance=1000000 WHERE id=1; --
```

**CVSS Score:** 9.1 (Critical)

**Mitigation:**
- Use parameterized queries exclusively (pgx $1, $2)
- Input validation on all user-provided data
- Least privilege database user (no DROP, ALTER)
- Database query logging and monitoring

**Implementation Priority:** CRITICAL (Week 1, Day 1)

---

### R - Repudiation

#### Threat R1: Trade Denial
**Description:** User denies placing losing trade to avoid losses

**Attack Vector:**
1. User places high-risk trade manually
2. Trade loses money
3. User claims "I never placed that trade"
4. No audit trail to prove otherwise

**CVSS Score:** 7.5 (High)

**Mitigation:**
- Immutable audit log for all trades
- Cryptographic signatures on order messages
- Timestamp verification via NTP
- User IP address, device fingerprint logging
- Video recording for admin actions (optional)
- Non-repudiation for high-value trades (>$10k)

**Implementation Priority:** HIGH (Week 2, Day 1)

---

### I - Information Disclosure

#### Threat I1: API Key Exposure
**Description:** Hardcoded API keys leaked via Git history

**Attack Vector:**
1. Attacker clones public GitHub repository
2. Searches Git history for API keys
3. Finds: `const OANDA_API_KEY = "977e1a77e25bac3a688011d6b0e845dd..."`
4. Uses key to access OANDA account and drain funds

**CVSS Score:** 9.8 (Critical) - **CVE-1**

**Mitigation:**
- Remove all hardcoded secrets from source code
- Use environment variables loaded from .env file
- Store production secrets in HashiCorp Vault
- Rewrite Git history to remove exposed secrets
- Rotate compromised API keys immediately
- Implement secret scanning in CI/CD (Trufflehog)

**Implementation Priority:** CRITICAL (Week 1, Day 1)

---

#### Threat I2: Sensitive Data in Logs
**Description:** PII, passwords, API keys logged to stdout/files

**Attack Vector:**
1. Attacker gains read access to log files
2. Searches for keywords: "password", "api_key", "secret"
3. Extracts sensitive data from logs

**CVSS Score:** 7.5 (High)

**Mitigation:**
- Sanitize logs (redact passwords, API keys, SSN)
- Use structured logging (JSON format)
- Implement log retention policies (90 days)
- Encrypt log files at rest
- Restrict log file permissions (0600)
- NEVER log request bodies containing credentials

**Example:**
```go
// VULNERABLE
log.Printf("Login attempt: username=%s, password=%s", username, password)

// SECURE
log.Printf("Login attempt: username=%s", username)
```

**Implementation Priority:** HIGH (Week 1, Day 3)

---

### D - Denial of Service

#### Threat D1: Rate Limit Bypass
**Description:** Attacker overwhelms system with unlimited requests

**Attack Vector:**
1. Attacker writes script to send 10,000 requests/second
2. No rate limiting implemented
3. Server CPU/memory exhausted
4. Legitimate users cannot trade

**CVSS Score:** 7.5 (High) - **CVE-7**

**Mitigation:**
- Implement token bucket rate limiting
- Adaptive limits based on user tier:
  - TRADER: 200 req/min
  - ADMIN: 500 req/min
  - SUPER_ADMIN: 1000 req/min
  - ANONYMOUS: 10 req/min
- IP blacklisting after threshold violations
- DDoS protection via Cloudflare/AWS Shield
- Connection limits on WebSocket (100 per IP)

**Implementation Priority:** HIGH (Week 2, Day 2)

---

#### Threat D2: Memory Exhaustion
**Description:** Unbounded data structures grow until OOM crash

**Attack Vector:**
1. Attacker opens 10,000 WebSocket connections
2. Hub.clients map grows unbounded
3. Server runs out of memory and crashes

**CVSS Score:** 6.5 (Medium) - **CVE-6**

**Mitigation:**
- Bounded collections (LRU cache with max size)
- Context-based cleanup of goroutines
- Memory limits via Go runtime (debug.SetMemoryLimit)
- Connection limits per IP
- Graceful degradation under load
- Memory monitoring and alerting

**Implementation Priority:** MEDIUM (Week 2, Day 1)

---

### E - Elevation of Privilege

#### Threat E1: Horizontal Privilege Escalation
**Description:** Trader accesses another trader's account/positions

**Attack Vector:**
1. Trader logs in with valid credentials
2. Manipulates account ID in URL: `/api/positions?accountId=12345`
3. Views another trader's positions
4. No authorization check to verify ownership

**CVSS Score:** 8.1 (High)

**Mitigation:**
- Implement authorization checks on all endpoints
- Verify user owns requested resource
- Use database Row-Level Security (RLS)
- Set current_user_id in database context
- Log all cross-account access attempts

**Example:**
```go
func (h *Handler) GetPositions(w http.ResponseWriter, r *http.Request) {
    user, _ := GetUser(r)
    requestedAccountID := r.URL.Query().Get("accountId")

    // CRITICAL: Verify ownership
    if !h.authService.OwnsAccount(user.ID, requestedAccountID) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    positions := h.positionRepo.GetByAccount(requestedAccountID)
    json.NewEncoder(w).Encode(positions)
}
```

**Implementation Priority:** CRITICAL (Week 1, Day 3)

---

#### Threat E2: Vertical Privilege Escalation
**Description:** Trader gains admin privileges to manipulate funds

**Attack Vector:**
1. Trader discovers JWT role claim is not validated server-side
2. Modifies JWT payload to set `role: "SUPER_ADMIN"`
3. Accesses `/admin/deposit` endpoint
4. Adds unlimited funds to own account

**CVSS Score:** 9.1 (Critical)

**Mitigation:**
- Validate JWT signature server-side (prevent tampering)
- Use RS256 asymmetric signing (can't forge without private key)
- Role checks in middleware before handler execution
- Audit log all admin actions with elevated permissions
- Require 2FA for admin operations
- Separate admin portal on different subdomain

**Implementation Priority:** CRITICAL (Week 1, Day 2)

---

## Attack Surface Analysis

### 1. HTTP API Endpoints

| Endpoint | Auth Required | Input Validation | Rate Limit | Risk Level |
|----------|---------------|------------------|------------|------------|
| `/login` | No | ❌ Missing | ❌ None | CRITICAL |
| `/api/orders/market` | ❌ No | ❌ Missing | ❌ None | CRITICAL |
| `/api/positions/close` | ❌ No | ❌ Missing | ❌ None | CRITICAL |
| `/admin/deposit` | ❌ No | ❌ Missing | ❌ None | CRITICAL |
| `/admin/withdraw` | ❌ No | ❌ Missing | ❌ None | CRITICAL |
| `/ticks` | No | ✅ Symbol validation | ❌ None | MEDIUM |
| `/health` | No | N/A | ❌ None | LOW |

**Total Exposed Endpoints:** 42
**Unauthenticated Endpoints:** 40 (95%)
**Missing Input Validation:** 38 (90%)
**Missing Rate Limiting:** 42 (100%)

---

### 2. WebSocket Connections

**Current Implementation:**
```go
// VULNERABLE: No authentication, no origin validation
http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
    ws.ServeWs(hub, w, r)
})
```

**Vulnerabilities:**
- No authentication (anyone can connect)
- No origin validation (CSRF possible)
- No connection limits (DoS via 10k connections)
- No message rate limiting
- Plaintext WebSocket (ws:// not wss://)

**Mitigation:**
- Require JWT token in WebSocket handshake
- Validate Origin header against whitelist
- Limit connections per IP (100 max)
- Enforce WSS (WebSocket Secure over TLS)
- Implement message rate limiting (10 msg/sec)

---

### 3. Database Layer

**Current State:** No database (in-memory only) - **CVE-5**

**Future Vulnerabilities:**
- Connection string hardcoded (should be in .env)
- No TLS enforcement (plaintext credentials)
- No prepared statements (SQL injection risk)
- No connection pooling (exhaustion risk)
- No Row-Level Security (cross-account access)

**Mitigation:**
- Use pgx with TLS 1.3 (sslmode=require)
- Parameterized queries exclusively
- Connection pooling (max 100, min 10)
- RLS policies on all tables
- Column-level encryption for PII
- Backup encryption at rest

---

### 4. External Integrations

#### OANDA API
**Hardcoded Credentials:** `const OANDA_API_KEY = "977..."`
**Risk:** API key theft → unauthorized trading on real account
**Mitigation:** Environment variables + Vault

#### YoFx FIX Gateway
**Hardcoded Password:** `Password: getEnvOrDefault("YOFX_PASSWORD", "Brand#143")`
**Risk:** Default password accepted if env var missing
**Mitigation:** Require env var (panic if missing)

#### Binance WebSocket
**No Authentication:** Public endpoint (low risk)
**Risk:** Rate limiting by Binance (IP ban)
**Mitigation:** Implement retry with exponential backoff

---

## Risk Assessment Matrix

| Threat ID | Threat | Likelihood | Impact | Risk Score | Priority |
|-----------|--------|------------|--------|------------|----------|
| I1 | API Key Exposure | High | Critical | 9.8 | CRITICAL |
| T2 | SQL Injection | High | Critical | 9.1 | CRITICAL |
| S1 | JWT Forgery | High | Critical | 9.8 | CRITICAL |
| E2 | Vertical Privilege Escalation | High | Critical | 9.1 | CRITICAL |
| T1 | Order Manipulation | Medium | High | 8.8 | HIGH |
| S2 | Session Hijacking | Medium | High | 8.1 | HIGH |
| E1 | Horizontal Privilege Escalation | High | High | 8.1 | HIGH |
| R1 | Trade Denial | Medium | High | 7.5 | HIGH |
| I2 | Sensitive Data in Logs | Medium | High | 7.5 | HIGH |
| D1 | Rate Limit Bypass | Medium | High | 7.5 | HIGH |
| D2 | Memory Exhaustion | Medium | Medium | 6.5 | MEDIUM |

**Total Threats:** 45
**CRITICAL:** 4
**HIGH:** 12
**MEDIUM:** 15
**LOW:** 14

---

## Mitigation Roadmap

### Phase 1: CRITICAL (Week 1-2)
- [ ] CVE-1: Remove hardcoded API keys
- [ ] CVE-2: Implement JWT authentication
- [ ] CVE-3: Enable TLS 1.3
- [ ] CVE-4: Force password change on first login
- [ ] CVE-7: Implement rate limiting

### Phase 2: HIGH (Week 3-4)
- [ ] Audit logging for all trades
- [ ] Input validation on all endpoints
- [ ] Authorization checks (RBAC)
- [ ] Session management
- [ ] Error message sanitization

### Phase 3: MEDIUM (Week 5-6)
- [ ] CVE-5: Database migration (PostgreSQL)
- [ ] CVE-6: Bounded collections
- [ ] Memory monitoring
- [ ] DDoS protection
- [ ] Automated security testing

---

## Security Testing Plan

### 1. Static Analysis
- **Tool:** gosec, semgrep
- **Frequency:** Every commit (CI/CD)
- **Pass Criteria:** 0 high/critical issues

### 2. Dynamic Analysis
- **Tool:** OWASP ZAP, Burp Suite
- **Frequency:** Weekly
- **Pass Criteria:** 0 high-risk findings

### 3. Penetration Testing
- **Vendor:** External firm (annual)
- **Scope:** Full application stack
- **Pass Criteria:** No critical vulnerabilities

### 4. Bug Bounty
- **Platform:** HackerOne
- **Scope:** Production environment
- **Rewards:** $100-$10,000 per severity

---

**Document Owner:** Security Architect
**Approved By:** CTO, CISO
**Next Review:** 2026-02-18 (monthly)
