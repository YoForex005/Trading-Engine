# Security Architecture - Trading Engine v3.0

**Classification:** CONFIDENTIAL
**Version:** 3.0.0
**Last Updated:** 2026-01-18
**Owner:** Security Architect

## Table of Contents
1. [Executive Summary](#executive-summary)
2. [Threat Model](#threat-model)
3. [Security Boundaries](#security-boundaries)
4. [Defense in Depth](#defense-in-depth)
5. [Authentication & Authorization](#authentication--authorization)
6. [Data Protection](#data-protection)
7. [Network Security](#network-security)
8. [Monitoring & Incident Response](#monitoring--incident-response)

---

## Executive Summary

The RTX Trading Engine implements a comprehensive security architecture based on defense-in-depth principles, zero-trust networking, and industry best practices for financial systems. This document outlines the complete security model covering all attack surfaces from API endpoints to database persistence.

### Security Objectives
- **Confidentiality**: Protect customer financial data and trading credentials
- **Integrity**: Ensure accuracy of trades, balances, and audit logs
- **Availability**: Maintain 99.99% uptime with DDoS protection
- **Compliance**: Meet PCI-DSS, SOC 2, GDPR, MiFID II requirements

### Key Security Controls
- **Authentication**: JWT with RS256, 2FA/MFA, OAuth2/OIDC
- **Authorization**: Role-Based Access Control (RBAC) with 5 roles
- **Encryption**: TLS 1.3 in transit, AES-256-GCM at rest
- **Rate Limiting**: Adaptive token bucket per user tier
- **Audit Logging**: Immutable audit trail for regulatory compliance

---

## Threat Model

### STRIDE Analysis

#### Spoofing
**Threat**: Attacker impersonates legitimate user to access trading account

**Controls**:
- JWT with RS256 asymmetric signing (prevents token forgery)
- 2FA/MFA for high-value accounts
- Device fingerprinting for anomaly detection
- IP-based session binding

#### Tampering
**Threat**: Attacker modifies order data to manipulate trades

**Controls**:
- HMAC signatures on all trade messages
- Database transaction integrity checks
- Write-ahead logging for crash recovery
- Blockchain-based audit trail (optional)

#### Repudiation
**Threat**: User denies placing losing trade

**Controls**:
- Immutable audit log with cryptographic signatures
- Timestamp verification via NTP
- Non-repudiation signatures for high-value trades
- Video recording of admin actions (optional)

#### Information Disclosure
**Threat**: Sensitive data leaked via API or database

**Controls**:
- TLS 1.3 for all communications
- Field-level encryption for PII
- Data masking in logs
- Secure key management via Vault

#### Denial of Service
**Threat**: Attacker overwhelms system to prevent trading

**Controls**:
- Rate limiting per IP and user
- Connection pooling limits
- Circuit breakers for external services
- DDoS protection via Cloudflare/AWS Shield

#### Elevation of Privilege
**Threat**: Trader gains admin access to manipulate accounts

**Controls**:
- Least privilege RBAC
- Privilege escalation logging
- Admin actions require 2FA
- Separation of duties for critical operations

---

## Security Boundaries

```
┌─────────────────────────────────────────────────────────────┐
│                      INTERNET                                │
│                         ↕                                    │
│              ┌──────────────────────┐                        │
│              │   WAF + DDoS Shield  │                        │
│              └──────────────────────┘                        │
│                         ↕                                    │
│              ┌──────────────────────┐                        │
│              │   HTTPS/WSS (TLS)    │                        │
│              └──────────────────────┘                        │
└─────────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────────┐
│                 DMZ (Demilitarized Zone)                     │
│              ┌──────────────────────┐                        │
│              │   Load Balancer      │                        │
│              └──────────────────────┘                        │
│                         ↕                                    │
│              ┌──────────────────────┐                        │
│              │  API Gateway (Auth)  │                        │
│              └──────────────────────┘                        │
└─────────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────────┐
│               APPLICATION TIER (Private Subnet)              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  API Server │  │ WebSocket   │  │  B-Book     │         │
│  │  (HTTP)     │  │  Hub        │  │  Engine     │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│         ↕                ↕                 ↕                 │
│  ┌─────────────────────────────────────────────┐            │
│  │        Middleware Layer (Security)          │            │
│  │  - Auth/JWT Validation                      │            │
│  │  - Rate Limiting                            │            │
│  │  - Input Validation                         │            │
│  │  - Audit Logging                            │            │
│  └─────────────────────────────────────────────┘            │
└─────────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────────┐
│                DATA TIER (Private Subnet)                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ PostgreSQL  │  │    Redis    │  │   Vault     │         │
│  │ (Encrypted) │  │   (Cache)   │  │  (Secrets)  │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│         ↕                ↕                 ↕                 │
│  ┌─────────────────────────────────────────────┐            │
│  │    Data Protection Layer                    │            │
│  │  - Column-level encryption                  │            │
│  │  - Row-level security (RLS)                 │            │
│  │  - Backup encryption                        │            │
│  └─────────────────────────────────────────────┘            │
└─────────────────────────────────────────────────────────────┘
```

### Trust Zones

1. **Public Zone** (Internet)
   - Untrusted, all traffic
   - Controls: WAF, DDoS protection

2. **DMZ** (Load Balancer, API Gateway)
   - Semi-trusted, authenticated traffic only
   - Controls: TLS termination, rate limiting

3. **Application Zone** (API servers, services)
   - Trusted, internal communication
   - Controls: Mutual TLS, service mesh

4. **Data Zone** (Databases, secrets)
   - Highly trusted, encrypted at rest
   - Controls: Network isolation, encryption

---

## Defense in Depth

### Layer 1: Network Security

```go
// Firewall rules (iptables/nftables)
// Allow HTTPS (443), WSS (443), Admin SSH (2222)
iptables -A INPUT -p tcp --dport 443 -j ACCEPT
iptables -A INPUT -p tcp --dport 2222 -s 10.0.0.0/8 -j ACCEPT
iptables -A INPUT -j DROP # Default deny

// VPC Security Groups (AWS/GCP)
SecurityGroup:
  Ingress:
    - Port: 443, Source: 0.0.0.0/0 (HTTPS)
    - Port: 443, Source: 0.0.0.0/0 (WSS)
    - Port: 2222, Source: 10.0.0.0/8 (Admin SSH)
  Egress:
    - Port: 5432, Destination: db-subnet (PostgreSQL)
    - Port: 6379, Destination: cache-subnet (Redis)
    - Port: 443, Destination: 0.0.0.0/0 (HTTPS outbound)
```

### Layer 2: Application Security

```go
package middleware

// Security middleware chain
func SecurityChain(next http.Handler) http.Handler {
    // Order matters: most restrictive first
    chain := next
    chain = RateLimitMiddleware(chain)
    chain = AuthMiddleware(chain)
    chain = InputValidationMiddleware(chain)
    chain = AuditLogMiddleware(chain)
    chain = SecurityHeadersMiddleware(chain)
    chain = CORSMiddleware(chain)

    return chain
}

// Input validation middleware
func InputValidationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Validate content type
        if r.Method == "POST" || r.Method == "PUT" {
            ct := r.Header.Get("Content-Type")
            if ct != "application/json" {
                http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
                return
            }
        }

        // Validate request size (prevent DoS)
        r.Body = http.MaxBytesReader(w, r.Body, 1*1024*1024) // 1MB max

        // Validate path parameters (prevent path traversal)
        if strings.Contains(r.URL.Path, "..") {
            http.Error(w, "Invalid path", http.StatusBadRequest)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

### Layer 3: Data Security

```go
package database

import (
    "github.com/jackc/pgx/v5/pgxpool"
    "crypto/aes"
    "crypto/cipher"
)

// Encrypted database connection
func NewSecurePool(connString string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(connString)
    if err != nil {
        return nil, err
    }

    // Enforce SSL/TLS
    config.ConnConfig.TLSConfig = &tls.Config{
        MinVersion: tls.VersionTLS13,
        ServerName: "postgres.example.com",
    }

    // Connection pooling limits (prevent exhaustion)
    config.MaxConns = 100
    config.MinConns = 10
    config.MaxConnLifetime = 1 * time.Hour
    config.MaxConnIdleTime = 30 * time.Minute

    return pgxpool.NewWithConfig(context.Background(), config)
}

// Field-level encryption for sensitive data
type EncryptedField struct {
    gcm cipher.AEAD
}

func NewEncryptedField(key []byte) (*EncryptedField, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    return &EncryptedField{gcm: gcm}, nil
}

func (e *EncryptedField) Encrypt(plaintext string) (string, error) {
    nonce := make([]byte, e.gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    ciphertext := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *EncryptedField) Decrypt(ciphertext string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", err
    }

    nonceSize := e.gcm.NonceSize()
    if len(data) < nonceSize {
        return "", errors.New("ciphertext too short")
    }

    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }

    return string(plaintext), nil
}
```

---

## Authentication & Authorization

### Authentication Flow

```
┌─────────┐                                    ┌─────────┐
│ Client  │                                    │  Server │
└────┬────┘                                    └────┬────┘
     │                                              │
     │  POST /login {username, password}            │
     │─────────────────────────────────────────────>│
     │                                              │
     │                         1. Validate password │
     │                         2. Check 2FA enabled │
     │                                              │
     │<─────────────────────────────────────────────│
     │  200 OK {twoFactorRequired: true}            │
     │                                              │
     │  POST /login/2fa {code}                      │
     │─────────────────────────────────────────────>│
     │                                              │
     │                          1. Validate TOTP    │
     │                          2. Generate JWT     │
     │                          3. Create session   │
     │                                              │
     │<─────────────────────────────────────────────│
     │  200 OK {token, refreshToken}                │
     │                                              │
     │  GET /api/positions                          │
     │  Authorization: Bearer <JWT>                 │
     │─────────────────────────────────────────────>│
     │                                              │
     │                         1. Validate JWT      │
     │                         2. Check permissions │
     │                         3. Fetch positions   │
     │                                              │
     │<─────────────────────────────────────────────│
     │  200 OK {positions: [...]}                   │
     │                                              │
```

### JWT Structure

```json
{
  "header": {
    "alg": "RS256",
    "typ": "JWT",
    "kid": "key-2026-01"
  },
  "payload": {
    "sub": "12345",
    "username": "trader1",
    "role": "TRADER",
    "permissions": [
      "trading:order:create",
      "trading:position:read",
      "trading:position:close"
    ],
    "iat": 1705536000,
    "exp": 1705622400,
    "iss": "rtx-trading-engine",
    "aud": "rtx-api",
    "jti": "unique-token-id",
    "ip": "203.0.113.42",
    "device": "Chrome/macOS"
  },
  "signature": "..."
}
```

### RBAC Permission Matrix

| Role | Permissions | Description |
|------|-------------|-------------|
| SUPER_ADMIN | `admin:*`, `system:*` | Full system access, can modify execution mode |
| ADMIN | `admin:accounts:*`, `admin:deposit`, `admin:withdraw` | Broker management, fund operations |
| BROKER | `account:read`, `account:write`, `trading:read` | Client account management |
| TRADER | `trading:order:*`, `trading:position:*`, `account:read` | Trading operations only |
| VIEWER | `trading:read`, `account:read` | Read-only access |

### Implementation

```go
package auth

import (
    "crypto/rsa"
    "github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
    privateKey *rsa.PrivateKey
    publicKey  *rsa.PublicKey
    issuer     string
}

func NewJWTService(privateKeyPath, publicKeyPath string) (*JWTService, error) {
    privateKeyData, err := os.ReadFile(privateKeyPath)
    if err != nil {
        return nil, err
    }

    privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
    if err != nil {
        return nil, err
    }

    publicKeyData, err := os.ReadFile(publicKeyPath)
    if err != nil {
        return nil, err
    }

    publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
    if err != nil {
        return nil, err
    }

    return &JWTService{
        privateKey: privateKey,
        publicKey:  publicKey,
        issuer:     "rtx-trading-engine",
    }, nil
}

func (s *JWTService) GenerateToken(user *User, ip, device string) (string, error) {
    now := time.Now()
    claims := jwt.MapClaims{
        "sub":         user.ID,
        "username":    user.Username,
        "role":        user.Role,
        "permissions": Permissions[user.Role],
        "iat":         now.Unix(),
        "exp":         now.Add(24 * time.Hour).Unix(),
        "iss":         s.issuer,
        "aud":         "rtx-api",
        "jti":         generateJTI(),
        "ip":          ip,
        "device":      device,
    }

    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    token.Header["kid"] = "key-2026-01" // Key rotation support

    return token.SignedString(s.privateKey)
}

func (s *JWTService) ValidateToken(tokenString string) (*jwt.Token, error) {
    return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Validate signing method
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }

        // Return public key for verification
        return s.publicKey, nil
    })
}
```

---

## Data Protection

### Encryption Standards

1. **In Transit** (TLS 1.3)
   - Cipher Suite: `TLS_AES_256_GCM_SHA384` (preferred)
   - Key Exchange: X25519 (ECDHE)
   - Certificate: RSA 4096 or ECDSA P-384

2. **At Rest** (AES-256-GCM)
   - Database: PostgreSQL native encryption + pgcrypto
   - Backups: Encrypted with GPG (RSA 4096)
   - Secrets: HashiCorp Vault with auto-rotation

3. **In Use** (Memory Protection)
   - Sensitive data cleared after use
   - No secrets in logs
   - Core dumps disabled for production

### Key Management

```go
package keymanagement

import (
    "github.com/hashicorp/vault/api"
)

type KeyManager struct {
    vault *api.Client
    path  string
}

func NewKeyManager(vaultAddr, token string) (*KeyManager, error) {
    config := api.DefaultConfig()
    config.Address = vaultAddr

    client, err := api.NewClient(config)
    if err != nil {
        return nil, err
    }

    client.SetToken(token)

    return &KeyManager{
        vault: client,
        path:  "secret/trading-engine",
    }, nil
}

func (km *KeyManager) GetEncryptionKey(purpose string) ([]byte, error) {
    secret, err := km.vault.Logical().Read(km.path + "/encryption-keys")
    if err != nil {
        return nil, err
    }

    key, ok := secret.Data[purpose].(string)
    if !ok {
        return nil, fmt.Errorf("key not found: %s", purpose)
    }

    return base64.StdEncoding.DecodeString(key)
}

func (km *KeyManager) RotateKey(purpose string) error {
    // Generate new AES-256 key
    newKey := make([]byte, 32)
    if _, err := rand.Read(newKey); err != nil {
        return err
    }

    // Store in Vault
    data := map[string]interface{}{
        purpose: base64.StdEncoding.EncodeToString(newKey),
    }

    _, err := km.vault.Logical().Write(km.path+"/encryption-keys", data)
    if err != nil {
        return err
    }

    log.Printf("[SECURITY] Encryption key rotated: %s", purpose)
    return nil
}
```

### Database Security

```sql
-- Row-Level Security (RLS) for multi-tenancy
ALTER TABLE accounts ENABLE ROW LEVEL SECURITY;

CREATE POLICY account_isolation ON accounts
    USING (user_id = current_setting('app.current_user_id')::bigint);

-- Column-level encryption for PII
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Encrypted column for SSN
ALTER TABLE users ADD COLUMN ssn_encrypted BYTEA;

-- Encrypt function
CREATE OR REPLACE FUNCTION encrypt_ssn(ssn TEXT) RETURNS BYTEA AS $$
BEGIN
    RETURN pgp_sym_encrypt(ssn, current_setting('app.encryption_key'));
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Decrypt function (restricted access)
CREATE OR REPLACE FUNCTION decrypt_ssn(encrypted BYTEA) RETURNS TEXT AS $$
BEGIN
    RETURN pgp_sym_decrypt(encrypted, current_setting('app.encryption_key'));
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Audit trigger
CREATE OR REPLACE FUNCTION audit_trigger() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO audit_log (
        table_name, action, user_id, old_data, new_data, timestamp
    ) VALUES (
        TG_TABLE_NAME, TG_OP, current_setting('app.current_user_id')::bigint,
        row_to_json(OLD), row_to_json(NEW), NOW()
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER accounts_audit
    AFTER INSERT OR UPDATE OR DELETE ON accounts
    FOR EACH ROW EXECUTE FUNCTION audit_trigger();
```

---

## Network Security

### DDoS Protection

```go
package ddos

import (
    "sync"
    "time"
)

type DDoSProtector struct {
    requests map[string]*RequestTracker
    mu       sync.RWMutex
    window   time.Duration
    threshold int
}

type RequestTracker struct {
    count     int
    firstSeen time.Time
    blocked   bool
}

func NewDDoSProtector(window time.Duration, threshold int) *DDoSProtector {
    dp := &DDoSProtector{
        requests:  make(map[string]*RequestTracker),
        window:    window,
        threshold: threshold,
    }

    go dp.cleanup()

    return dp
}

func (dp *DDoSProtector) Check(ip string) bool {
    dp.mu.Lock()
    defer dp.mu.Unlock()

    tracker, exists := dp.requests[ip]
    if !exists {
        dp.requests[ip] = &RequestTracker{
            count:     1,
            firstSeen: time.Now(),
            blocked:   false,
        }
        return true
    }

    // Reset counter if window expired
    if time.Since(tracker.firstSeen) > dp.window {
        tracker.count = 1
        tracker.firstSeen = time.Now()
        tracker.blocked = false
        return true
    }

    // Increment counter
    tracker.count++

    // Block if over threshold
    if tracker.count > dp.threshold {
        tracker.blocked = true
        log.Printf("[SECURITY] DDoS protection triggered for IP: %s (%d requests)", ip, tracker.count)
        return false
    }

    return true
}

func (dp *DDoSProtector) cleanup() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        dp.mu.Lock()
        now := time.Now()
        for ip, tracker := range dp.requests {
            if now.Sub(tracker.firstSeen) > dp.window*2 {
                delete(dp.requests, ip)
            }
        }
        dp.mu.Unlock()
    }
}
```

---

## Monitoring & Incident Response

### Security Monitoring

```go
package monitoring

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    failedLoginAttempts = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "auth_failed_login_attempts_total",
            Help: "Total number of failed login attempts",
        },
        []string{"username", "ip"},
    )

    rateLimitExceeded = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "rate_limit_exceeded_total",
            Help: "Total number of rate limit violations",
        },
        []string{"endpoint", "ip"},
    )

    suspiciousActivity = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "security_suspicious_activity_total",
            Help: "Total number of suspicious activities detected",
        },
        []string{"type", "severity"},
    )
)

type SecurityMonitor struct {
    alertChannel chan SecurityAlert
}

type SecurityAlert struct {
    Type     string
    Severity string
    Message  string
    IP       string
    UserID   string
    Time     time.Time
}

func (sm *SecurityMonitor) DetectAnomalies(user *User, ip string) {
    // Detect login from unusual location
    if !isKnownIP(user.ID, ip) {
        sm.alertChannel <- SecurityAlert{
            Type:     "unusual_login_location",
            Severity: "medium",
            Message:  fmt.Sprintf("Login from unknown IP: %s", ip),
            IP:       ip,
            UserID:   user.ID,
            Time:     time.Now(),
        }
    }

    // Detect unusual trading volume
    if isUnusualVolume(user.ID) {
        sm.alertChannel <- SecurityAlert{
            Type:     "unusual_trading_volume",
            Severity: "high",
            Message:  "Trading volume exceeds normal pattern",
            UserID:   user.ID,
            Time:     time.Now(),
        }
    }
}

func (sm *SecurityMonitor) ProcessAlerts() {
    for alert := range sm.alertChannel {
        // Log to SIEM
        log.Printf("[SECURITY ALERT] Type=%s, Severity=%s, IP=%s, User=%s, Message=%s",
            alert.Type, alert.Severity, alert.IP, alert.UserID, alert.Message)

        // Send to alerting system (PagerDuty, Slack, etc.)
        if alert.Severity == "critical" || alert.Severity == "high" {
            sendPagerDutyAlert(alert)
        }

        // Store in security events table
        storeSecurityEvent(alert)
    }
}
```

### Incident Response Plan

1. **Detection** (0-5 minutes)
   - Automated alerts via Prometheus/Grafana
   - Security event correlation in SIEM

2. **Triage** (5-15 minutes)
   - On-call security engineer notified
   - Assess severity and scope
   - Activate incident response team

3. **Containment** (15-30 minutes)
   - Block malicious IPs
   - Revoke compromised tokens
   - Isolate affected services

4. **Eradication** (30-60 minutes)
   - Remove malware/backdoors
   - Patch vulnerabilities
   - Reset credentials

5. **Recovery** (1-4 hours)
   - Restore from clean backups
   - Verify system integrity
   - Resume normal operations

6. **Lessons Learned** (24-48 hours)
   - Root cause analysis
   - Update runbooks
   - Implement preventive controls

---

**Document Classification:** CONFIDENTIAL
**Distribution:** Security team, DevOps, Compliance
**Retention:** 7 years (regulatory requirement)
