# Secure Coding Patterns - Trading Engine

**Purpose:** Reusable security patterns for consistent implementation across all modules
**Audience:** Development team, security reviewers
**Maintenance:** Update on each security review

---

## Table of Contents
1. [Input Validation](#input-validation)
2. [Authentication Patterns](#authentication-patterns)
3. [Secure Database Access](#secure-database-access)
4. [Cryptographic Operations](#cryptographic-operations)
5. [Error Handling](#error-handling)
6. [Logging & Monitoring](#logging--monitoring)

---

## Input Validation

### Pattern 1: Zod-Based API Validation

**Use Case:** Validate all HTTP request bodies to prevent injection attacks

**Implementation:**
```go
package validation

import (
    "encoding/json"
    "net/http"
    "regexp"
    "errors"
)

type OrderRequest struct {
    Symbol string  `json:"symbol"`
    Side   string  `json:"side"`
    Volume float64 `json:"volume"`
    Price  float64 `json:"price,omitempty"`
}

type Validator struct {
    symbolPattern *regexp.Regexp
}

func NewValidator() *Validator {
    return &Validator{
        symbolPattern: regexp.MustCompile(`^[A-Z]{6}$`), // EURUSD, BTCUSD, etc.
    }
}

func (v *Validator) ValidateOrderRequest(r *http.Request) (*OrderRequest, error) {
    var req OrderRequest

    // Limit request size (prevent DoS)
    r.Body = http.MaxBytesReader(nil, r.Body, 1*1024*1024) // 1MB

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return nil, errors.New("invalid JSON format")
    }

    // Validate symbol format
    if !v.symbolPattern.MatchString(req.Symbol) {
        return nil, errors.New("invalid symbol format (must be 6 uppercase letters)")
    }

    // Validate side
    if req.Side != "BUY" && req.Side != "SELL" {
        return nil, errors.New("side must be BUY or SELL")
    }

    // Validate volume range
    if req.Volume <= 0 || req.Volume > 1000 {
        return nil, errors.New("volume must be between 0.01 and 1000")
    }

    // Validate price (if limit order)
    if req.Price != 0 && (req.Price <= 0 || req.Price > 1000000) {
        return nil, errors.New("price must be between 0 and 1,000,000")
    }

    return &req, nil
}

// Usage in handler
func (h *Handler) HandlePlaceOrder(w http.ResponseWriter, r *http.Request) {
    req, err := h.validator.ValidateOrderRequest(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // req is now safe to use
    order := h.orderService.PlaceOrder(req.Symbol, req.Side, req.Volume, req.Price)
    json.NewEncoder(w).Encode(order)
}
```

**Security Benefits:**
- Prevents SQL injection via validated inputs
- Prevents NoSQL injection in unstructured queries
- Rejects malformed requests early (DoS prevention)
- Type safety with strong validation rules

---

### Pattern 2: Path Sanitization

**Use Case:** Prevent path traversal attacks when handling file operations

**Implementation:**
```go
package security

import (
    "path/filepath"
    "strings"
    "errors"
)

type PathValidator struct {
    allowedPrefix string
}

func NewPathValidator(allowedPrefix string) *PathValidator {
    return &PathValidator{
        allowedPrefix: filepath.Clean(allowedPrefix),
    }
}

func (pv *PathValidator) ValidatePath(userPath string) (string, error) {
    // Clean the path (resolves .., ., etc.)
    cleaned := filepath.Clean(userPath)

    // Convert to absolute path
    absolute := filepath.Join(pv.allowedPrefix, cleaned)

    // Ensure it's within allowed directory
    if !strings.HasPrefix(absolute, pv.allowedPrefix) {
        return "", errors.New("path traversal detected")
    }

    return absolute, nil
}

// Usage example
func (s *TickStore) LoadTickFile(symbol, date string) error {
    validator := NewPathValidator("/var/data/ticks")

    // User-provided input
    requestedPath := fmt.Sprintf("%s/%s.json", symbol, date)

    safePath, err := validator.ValidatePath(requestedPath)
    if err != nil {
        return fmt.Errorf("invalid path: %w", err)
    }

    // safePath is guaranteed to be within /var/data/ticks
    data, err := os.ReadFile(safePath)
    // ...
}
```

**Dangerous Pattern (NEVER DO THIS):**
```go
// VULNERABLE: Direct concatenation allows path traversal
userInput := r.URL.Query().Get("file") // ../../../../etc/passwd
filePath := "/var/data/" + userInput
data, _ := os.ReadFile(filePath) // EXPLOITED!
```

---

### Pattern 3: SQL Injection Prevention

**Use Case:** All database queries with user input

**Implementation:**
```go
package repository

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
    pool *pgxpool.Pool
}

// SECURE: Prepared statement with parameterized query
func (r *OrderRepository) GetOrdersBySymbol(symbol string) ([]*Order, error) {
    query := `
        SELECT id, symbol, side, volume, price, status, created_at
        FROM orders
        WHERE symbol = $1 AND user_id = $2
        ORDER BY created_at DESC
        LIMIT 100
    `

    rows, err := r.pool.Query(context.Background(), query, symbol, getCurrentUserID())
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var orders []*Order
    for rows.Next() {
        var order Order
        err := rows.Scan(&order.ID, &order.Symbol, &order.Side, &order.Volume,
            &order.Price, &order.Status, &order.CreatedAt)
        if err != nil {
            return nil, err
        }
        orders = append(orders, &order)
    }

    return orders, nil
}
```

**Dangerous Pattern (NEVER DO THIS):**
```go
// VULNERABLE: String concatenation allows SQL injection
symbol := r.URL.Query().Get("symbol") // "EURUSD'; DROP TABLE orders; --"
query := "SELECT * FROM orders WHERE symbol = '" + symbol + "'" // EXPLOITED!
rows, _ := db.Query(query)
```

---

## Authentication Patterns

### Pattern 4: JWT Validation Middleware

**Use Case:** Protect all authenticated endpoints

**Implementation:**
```go
package middleware

import (
    "context"
    "net/http"
    "strings"
    "github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserContextKey contextKey = "user"

type Claims struct {
    UserID   string   `json:"sub"`
    Username string   `json:"username"`
    Role     string   `json:"role"`
    Perms    []string `json:"permissions"`
    jwt.RegisteredClaims
}

func AuthMiddleware(publicKey interface{}) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract Authorization header
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "Authorization header required", http.StatusUnauthorized)
                return
            }

            // Parse Bearer token
            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" {
                http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
                return
            }

            tokenString := parts[1]

            // Parse and validate JWT
            token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
                // Verify signing method
                if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
                    return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
                }
                return publicKey, nil
            })

            if err != nil || !token.Valid {
                http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
                return
            }

            claims, ok := token.Claims.(*Claims)
            if !ok {
                http.Error(w, "Invalid token claims", http.StatusUnauthorized)
                return
            }

            // Add claims to request context
            ctx := context.WithValue(r.Context(), UserContextKey, claims)

            // Log authentication for audit
            log.Printf("[AUTH] User=%s, Role=%s, IP=%s, Endpoint=%s",
                claims.Username, claims.Role, getClientIP(r), r.URL.Path)

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Helper: Extract user from context
func GetUser(r *http.Request) (*Claims, bool) {
    user, ok := r.Context().Value(UserContextKey).(*Claims)
    return user, ok
}

// Usage in handler
func (h *Handler) HandleClosePosition(w http.ResponseWriter, r *http.Request) {
    user, ok := GetUser(r)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Check permission
    if !hasPermission(user.Perms, "trading:position:close") {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // Proceed with authorized action
    // ...
}
```

---

### Pattern 5: Password Hashing

**Use Case:** Securely store user passwords

**Implementation:**
```go
package auth

import (
    "golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12 // Adjust based on performance requirements

// HashPassword hashes a plaintext password using bcrypt
func HashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
    if err != nil {
        return "", err
    }
    return string(hash), nil
}

// VerifyPassword compares plaintext password with hashed password
func VerifyPassword(hashedPassword, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Usage in account creation
func (s *Service) CreateAccount(username, password string) error {
    // Validate password strength
    if err := ValidatePasswordStrength(password); err != nil {
        return err
    }

    // Hash password
    hashedPassword, err := HashPassword(password)
    if err != nil {
        return err
    }

    // Store in database
    return s.repo.InsertAccount(username, hashedPassword)
}

// Usage in login
func (s *Service) Login(username, password string) (*User, error) {
    account, err := s.repo.GetAccountByUsername(username)
    if err != nil {
        return nil, errors.New("invalid credentials") // Generic error (prevent user enumeration)
    }

    // Verify password
    if err := VerifyPassword(account.PasswordHash, password); err != nil {
        return nil, errors.New("invalid credentials")
    }

    return &User{
        ID:       account.ID,
        Username: account.Username,
        Role:     account.Role,
    }, nil
}
```

**Dangerous Patterns (NEVER DO THIS):**
```go
// VULNERABLE: SHA-256 with static salt (not acceptable for passwords)
hashedPassword := sha256.Sum256([]byte(password + "static_salt"))

// VULNERABLE: Plaintext storage
account.Password = password

// VULNERABLE: Weak hashing (MD5, SHA-1)
hashedPassword := md5.Sum([]byte(password))
```

---

## Secure Database Access

### Pattern 6: Connection Pooling

**Use Case:** Prevent connection exhaustion and SQL injection

**Implementation:**
```go
package database

import (
    "context"
    "crypto/tls"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
    Host     string
    Port     int
    Database string
    User     string
    Password string
    SSLMode  string
    MaxConns int32
    MinConns int32
}

func NewPool(cfg *Config) (*pgxpool.Pool, error) {
    connString := fmt.Sprintf(
        "postgres://%s:%s@%s:%d/%s?sslmode=%s",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode,
    )

    config, err := pgxpool.ParseConfig(connString)
    if err != nil {
        return nil, err
    }

    // Connection pool configuration
    config.MaxConns = cfg.MaxConns           // Prevent exhaustion
    config.MinConns = cfg.MinConns           // Maintain minimum connections
    config.MaxConnLifetime = 1 * time.Hour   // Rotate connections
    config.MaxConnIdleTime = 30 * time.Minute

    // TLS configuration
    config.ConnConfig.TLSConfig = &tls.Config{
        MinVersion: tls.VersionTLS13,
        ServerName: cfg.Host,
    }

    // Connection timeout
    config.ConnConfig.ConnectTimeout = 10 * time.Second

    pool, err := pgxpool.NewWithConfig(context.Background(), config)
    if err != nil {
        return nil, err
    }

    // Verify connectivity
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("database ping failed: %w", err)
    }

    log.Println("[DATABASE] Connection pool initialized successfully")

    return pool, nil
}
```

---

### Pattern 7: Row-Level Security

**Use Case:** Multi-tenant data isolation

**Implementation:**
```sql
-- Enable RLS on accounts table
ALTER TABLE accounts ENABLE ROW LEVEL SECURITY;

-- Policy: Users can only see their own account
CREATE POLICY account_isolation ON accounts
    FOR SELECT
    USING (id = current_setting('app.user_id')::bigint);

-- Policy: Users can only update their own account
CREATE POLICY account_update ON accounts
    FOR UPDATE
    USING (id = current_setting('app.user_id')::bigint);

-- Policy: Admins can see all accounts
CREATE POLICY admin_access ON accounts
    FOR ALL
    TO admin_role
    USING (true);
```

```go
// Set user context before querying
func (r *Repository) SetUserContext(ctx context.Context, userID int64) error {
    query := "SET LOCAL app.user_id = $1"
    _, err := r.pool.Exec(ctx, query, userID)
    return err
}

// Usage in handler
func (h *Handler) HandleGetAccount(w http.ResponseWriter, r *http.Request) {
    user, _ := GetUser(r)

    ctx := context.Background()

    // Set RLS context
    if err := h.repo.SetUserContext(ctx, user.ID); err != nil {
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }

    // Query automatically filtered by RLS
    account, err := h.repo.GetAccount(ctx)
    if err != nil {
        http.Error(w, "Account not found", http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(account)
}
```

---

## Cryptographic Operations

### Pattern 8: Secure Random Generation

**Use Case:** Generate session IDs, API keys, CSRF tokens

**Implementation:**
```go
package security

import (
    "crypto/rand"
    "encoding/base64"
    "encoding/hex"
)

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateAPIKey generates a secure API key
func GenerateAPIKey() (string, error) {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return "rtx_" + hex.EncodeToString(bytes), nil
}

// GenerateSessionID generates a secure session ID
func GenerateSessionID() (string, error) {
    return GenerateSecureToken(32)
}
```

**Dangerous Patterns (NEVER DO THIS):**
```go
// VULNERABLE: math/rand is NOT cryptographically secure
import "math/rand"
sessionID := strconv.FormatInt(rand.Int63(), 36) // PREDICTABLE!

// VULNERABLE: Timestamp-based IDs
sessionID := strconv.FormatInt(time.Now().Unix(), 10) // PREDICTABLE!
```

---

### Pattern 9: AES-GCM Encryption

**Use Case:** Encrypt sensitive data at rest (SSN, credit cards, API keys)

**Implementation:**
```go
package encryption

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "io"
)

type AESEncryptor struct {
    gcm cipher.AEAD
}

func NewAESEncryptor(key []byte) (*AESEncryptor, error) {
    if len(key) != 32 {
        return nil, errors.New("key must be 32 bytes for AES-256")
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    return &AESEncryptor{gcm: gcm}, nil
}

func (e *AESEncryptor) Encrypt(plaintext string) (string, error) {
    // Generate random nonce
    nonce := make([]byte, e.gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    // Encrypt and authenticate
    ciphertext := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)

    // Encode to base64 for storage
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *AESEncryptor) Decrypt(ciphertext string) (string, error) {
    // Decode from base64
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", err
    }

    nonceSize := e.gcm.NonceSize()
    if len(data) < nonceSize {
        return "", errors.New("ciphertext too short")
    }

    // Extract nonce and ciphertext
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]

    // Decrypt and verify
    plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", errors.New("decryption failed (tampered data)")
    }

    return string(plaintext), nil
}

// Usage example
func (r *Repository) StoreAPIKey(userID int64, apiKey string) error {
    encryptionKey := getEncryptionKeyFromVault()
    encryptor, err := NewAESEncryptor(encryptionKey)
    if err != nil {
        return err
    }

    encryptedKey, err := encryptor.Encrypt(apiKey)
    if err != nil {
        return err
    }

    query := "UPDATE users SET api_key_encrypted = $1 WHERE id = $2"
    _, err = r.pool.Exec(context.Background(), query, encryptedKey, userID)
    return err
}
```

---

## Error Handling

### Pattern 10: Secure Error Messages

**Use Case:** Prevent information disclosure via error messages

**Implementation:**
```go
package errors

import (
    "log"
    "net/http"
)

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

// LogAndReturnGenericError logs detailed error internally but returns generic message to client
func LogAndReturnGenericError(w http.ResponseWriter, internalErr error, userMessage string) {
    // Log full error details for debugging (server-side only)
    log.Printf("[ERROR] Internal error: %v", internalErr)

    // Return generic message to client (prevent information disclosure)
    apiErr := APIError{
        Code:    "INTERNAL_ERROR",
        Message: userMessage,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(apiErr)
}

// Usage in handler
func (h *Handler) HandlePlaceOrder(w http.ResponseWriter, r *http.Request) {
    order, err := h.orderService.PlaceOrder(...)
    if err != nil {
        // SECURE: Generic message prevents leaking internal details
        LogAndReturnGenericError(w, err, "Unable to place order. Please try again.")
        return
    }

    json.NewEncoder(w).Encode(order)
}
```

**Dangerous Pattern (NEVER DO THIS):**
```go
// VULNERABLE: Exposes internal implementation details
if err != nil {
    http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
    // Leaks: "Database error: pq: password authentication failed for user 'admin'"
    return
}
```

---

## Logging & Monitoring

### Pattern 11: Audit Logging

**Use Case:** Track all security-relevant events for compliance

**Implementation:**
```go
package audit

import (
    "encoding/json"
    "time"
)

type AuditLog struct {
    EventType string                 `json:"event_type"`
    UserID    string                 `json:"user_id"`
    IP        string                 `json:"ip"`
    Action    string                 `json:"action"`
    Resource  string                 `json:"resource"`
    Status    string                 `json:"status"` // SUCCESS, FAILED
    Details   map[string]interface{} `json:"details"`
    Timestamp time.Time              `json:"timestamp"`
}

type Logger struct {
    repo *Repository
}

func NewLogger(repo *Repository) *Logger {
    return &Logger{repo: repo}
}

func (l *Logger) LogEvent(event *AuditLog) {
    event.Timestamp = time.Now()

    // Persist to database (immutable audit trail)
    if err := l.repo.InsertAuditLog(event); err != nil {
        log.Printf("[CRITICAL] Failed to write audit log: %v", err)
        // Alert security team
    }

    // Also log to stdout for SIEM ingestion
    jsonLog, _ := json.Marshal(event)
    log.Printf("[AUDIT] %s", string(jsonLog))
}

// Usage in handler
func (h *Handler) HandleClosePosition(w http.ResponseWriter, r *http.Request) {
    user, _ := GetUser(r)

    var req ClosePositionRequest
    json.NewDecoder(r.Body).Decode(&req)

    err := h.positionService.ClosePosition(req.PositionID)

    // Log security event
    h.auditLogger.LogEvent(&AuditLog{
        EventType: "POSITION_CLOSE",
        UserID:    user.ID,
        IP:        getClientIP(r),
        Action:    "close_position",
        Resource:  req.PositionID,
        Status:    statusFromError(err),
        Details: map[string]interface{}{
            "position_id": req.PositionID,
            "symbol":      req.Symbol,
            "volume":      req.Volume,
        },
    })

    if err != nil {
        http.Error(w, "Failed to close position", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
```

---

## Summary

### Critical Security Principles

1. **Never Trust User Input**
   - Validate all inputs (whitelist approach)
   - Use prepared statements for SQL queries
   - Sanitize file paths

2. **Defense in Depth**
   - Multiple layers of security (network, application, data)
   - Fail securely (deny by default)

3. **Least Privilege**
   - Users get minimum required permissions
   - Services run with minimum required access

4. **Secure by Default**
   - HTTPS/TLS required
   - Authentication required (unless explicitly public)
   - Encryption at rest

5. **Audit Everything**
   - Log all security events
   - Immutable audit trail
   - Monitor for anomalies

### Code Review Checklist

- [ ] All user inputs validated
- [ ] SQL queries use prepared statements
- [ ] Passwords hashed with bcrypt (cost 12+)
- [ ] Secrets stored in Vault (not code)
- [ ] TLS 1.3 enforced
- [ ] Error messages don't leak internal details
- [ ] Audit logging for sensitive operations
- [ ] Rate limiting on all endpoints
- [ ] Authentication middleware on protected routes
- [ ] Authorization checks before privileged operations

---

**Document Owner:** Security Architect
**Last Updated:** 2026-01-18
**Next Review:** 2026-02-18
