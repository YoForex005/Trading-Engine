# Security Hardening Package

Comprehensive security hardening features for the RTX Trading Engine backend, implementing OWASP Top 10 compliance and financial industry security best practices.

## Features

### 1. Web Application Firewall (WAF)
**File:** `waf.go`

- **Rate Limiting**: Configurable requests per IP/minute
- **IP Blocking**: Automatic blocking after failed attempts
- **DDoS Protection**: Connection limits and slow loris prevention
- **Request Validation**: Size limits for body, URL, and headers
- **IP Whitelisting**: Admin endpoint protection
- **Geo-blocking**: Country-based traffic filtering

**Usage:**
```go
wafConfig := security.DefaultWAFConfig()
waf := security.NewWAF(wafConfig)

// Apply WAF middleware
http.Handle("/api/", waf.Middleware(apiHandler))

// Admin-only endpoints
http.Handle("/admin/", waf.AdminOnlyMiddleware(adminHandler))
```

### 2. Data Encryption
**File:** `encryption.go`

- **AES-256-GCM**: Industry-standard encryption
- **Argon2**: Key derivation from passphrase
- **Sensitive Data Protection**: API keys, credentials, bank details

**Usage:**
```go
encryptionService := security.NewEncryptionService("your-master-passphrase")

// Encrypt sensitive data
encrypted, _ := encryptionService.EncryptString("sensitive-api-key")

// Decrypt when needed
decrypted, _ := encryptionService.DecryptString(encrypted)
```

### 3. Audit Logging
**File:** `audit.go`

- **Security Events**: Authentication, authorization, data access
- **Compliance**: Transaction audit trail, admin actions
- **Log Rotation**: Automatic rotation at 100MB
- **Retention**: 90-day default retention
- **Queryable**: Search logs by time, level, category

**Usage:**
```go
auditLogger, _ := security.NewAuditLogger("/var/log/rtx/audit")

// Log authentication attempt
auditLogger.LogAuthAttempt("user123", "192.168.1.1", true, "Login successful")

// Log admin action
auditLogger.LogAdminAction("admin", "10.0.0.1", "deposit", "account-001", true, metadata)
```

### 4. Compliance Checker
**File:** `compliance.go`

- **OWASP Top 10**: All 10 categories covered
- **Financial Compliance**: PCI DSS, KYC/AML checks
- **Input Validation**: Email, alphanumeric, numeric, symbol
- **XSS Prevention**: Input sanitization
- **Path Security**: Traversal attack prevention

**Usage:**
```go
complianceChecker := security.NewComplianceChecker(auditLogger)

// Run all compliance checks
report, _ := complianceChecker.RunAllChecks()

fmt.Printf("Status: %s, Passed: %d, Failed: %d\n",
    report.OverallStatus, report.Passed, report.Failed)
```

### 5. Vulnerability Scanner
**File:** `scanner.go`

- **Code Scanning**: Detects 12+ vulnerability patterns
- **Secret Detection**: Hardcoded API keys, passwords, JWT secrets
- **SQL Injection**: String concatenation in queries
- **Weak Crypto**: MD5, SHA1, insecure random
- **Command Injection**: Unsafe exec calls
- **Path Traversal**: Unsafe file operations

**Usage:**
```go
scanner := security.NewVulnerabilityScanner(auditLogger)

// Scan entire codebase
report, _ := scanner.ScanDirectory("/path/to/backend")

fmt.Printf("Files: %d, Vulnerabilities: %d (Critical: %d)\n",
    report.FilesScanned, report.Vulnerabilities, report.Critical)
```

### 6. CSRF Protection
**File:** `csrf.go`

- **Token-based**: HMAC-signed tokens
- **Session-bound**: Tokens tied to sessions
- **Auto-validation**: Middleware for state-changing operations
- **24-hour validity**: Configurable token lifetime

**Usage:**
```go
csrf := security.NewCSRFProtection("csrf-secret-key")

// Generate token for session
token, _ := csrf.GenerateToken(sessionID)

// Apply middleware to protect endpoints
http.Handle("/api/transfer", csrf.Middleware(transferHandler))
```

### 7. Session Management
**File:** `session.go`

- **Timeouts**: 30-minute default inactivity timeout
- **Concurrent Limits**: Max 3 sessions per user
- **IP Validation**: Optional IP binding
- **Auto-cleanup**: Expired session removal

**Usage:**
```go
sessionConfig := security.DefaultSessionConfig()
sessionManager := security.NewSessionManager(sessionConfig, auditLogger)

// Create session
session, _ := sessionManager.CreateSession("user123", "192.168.1.1")

// Validate session
validSession, _ := sessionManager.ValidateSession(session.ID, "192.168.1.1")
```

### 8. API Key Rotation
**File:** `apikeys.go`

- **Auto-rotation**: Configurable rotation interval
- **Lifecycle Management**: Generate, rotate, revoke
- **Expiration Tracking**: Keys expire automatically
- **Audit Trail**: All operations logged

**Usage:**
```go
keyManager := security.NewAPIKeyManager(
    90*24*time.Hour, // 90-day rotation
    auditLogger,
    encryptionService,
)

// Generate key for service
apiKey, _ := keyManager.GenerateKey("oanda")

// Auto-rotation every 90 days
// Manual rotation
newKey, _ := keyManager.RotateKey("oanda", "scheduled_rotation")
```

### 9. Security Middleware
**File:** `middleware.go`

- **Comprehensive Protection**: WAF + CSRF + Sessions
- **Security Headers**: HSTS, CSP, X-Frame-Options, etc.
- **CORS**: Configurable allowed origins
- **Input Sanitization**: XSS prevention
- **Request Validation**: Content-Type, User-Agent checks

**Usage:**
```go
securityMiddleware := security.NewSecurityMiddleware(waf, csrf, sessionManager, auditLogger)

// Protect all endpoints
http.Handle("/api/", securityMiddleware.Protect(apiHandler))

// Admin endpoints with extra protection
http.Handle("/admin/", securityMiddleware.AdminProtect(adminHandler))

// Add security headers
http.Handle("/", security.SecurityHeadersMiddleware(handler))
```

### 10. Security Testing
**File:** `testing.go`

- **Automated Tests**: 12+ security test cases
- **OWASP Coverage**: SQL injection, XSS, CSRF
- **Authentication Tests**: Weak credentials, brute force
- **Header Validation**: Security header presence
- **Penetration Testing**: Path traversal, error disclosure

**Usage:**
```go
testSuite := security.NewSecurityTestSuite("http://localhost:7999")
testSuite.RegisterDefaultTests()

// Run all tests
report := testSuite.RunAll()

fmt.Printf("Passed: %d, Failed: %d\n", report.Passed, report.Failed)
```

## OWASP Top 10 Compliance

| OWASP | Vulnerability | Coverage |
|-------|---------------|----------|
| A1 | Injection | ✅ SQL injection prevention, input validation |
| A2 | Broken Authentication | ✅ bcrypt, JWT, session management |
| A3 | Sensitive Data Exposure | ✅ AES-256-GCM encryption, secure headers |
| A4 | XML External Entities | ✅ No XML parsing (JSON only) |
| A5 | Broken Access Control | ✅ RBAC, session validation, CSRF |
| A6 | Security Misconfiguration | ✅ Secure defaults, security headers |
| A7 | Cross-Site Scripting | ✅ CSP, input sanitization |
| A8 | Insecure Deserialization | ✅ Safe JSON parsing |
| A9 | Using Components with Known Vulnerabilities | ✅ Dependency scanner |
| A10 | Insufficient Logging & Monitoring | ✅ Comprehensive audit logging |

## Security Checklist

### Pre-Deployment
- [ ] Run compliance checker: `complianceChecker.RunAllChecks()`
- [ ] Scan for vulnerabilities: `scanner.ScanDirectory()`
- [ ] Run security tests: `testSuite.RunAll()`
- [ ] Review audit logs for anomalies
- [ ] Verify all API keys are rotated
- [ ] Confirm HTTPS enforcement
- [ ] Test WAF rate limiting
- [ ] Validate CSRF protection on state-changing endpoints

### Production Configuration
- [ ] Set strong master encryption passphrase
- [ ] Configure WAF with production limits
- [ ] Enable IP whitelisting for admin endpoints
- [ ] Set appropriate session timeouts
- [ ] Configure API key rotation intervals
- [ ] Enable audit logging with secure storage
- [ ] Set up log monitoring and alerting
- [ ] Configure CORS with specific allowed origins

### Ongoing Maintenance
- [ ] Weekly: Review audit logs
- [ ] Monthly: Run vulnerability scans
- [ ] Quarterly: Full compliance audit
- [ ] Quarterly: API key rotation verification
- [ ] Annually: Penetration testing
- [ ] Continuous: Monitor security alerts

## Performance Impact

| Component | Latency Added | Memory Overhead |
|-----------|---------------|-----------------|
| WAF | ~1-2ms | ~10MB |
| CSRF | ~0.5ms | ~5MB |
| Session Management | ~0.5ms | ~15MB |
| Audit Logging | ~1ms (async) | ~20MB |
| Encryption | ~2-3ms | ~5MB |
| **Total** | **~5-8ms** | **~55MB** |

## Integration Example

```go
package main

import (
    "log"
    "net/http"
    "time"

    "github.com/epic1st/rtx/backend/security"
)

func main() {
    // Initialize security components
    auditLogger, _ := security.NewAuditLogger("/var/log/rtx/audit")
    defer auditLogger.Close()

    encryptionService := security.NewEncryptionService("production-master-key")

    wafConfig := security.DefaultWAFConfig()
    waf := security.NewWAF(wafConfig)

    csrf := security.NewCSRFProtection("csrf-secret")

    sessionConfig := security.DefaultSessionConfig()
    sessionManager := security.NewSessionManager(sessionConfig, auditLogger)

    keyManager := security.NewAPIKeyManager(90*24*time.Hour, auditLogger, encryptionService)

    // Create security middleware
    securityMiddleware := security.NewSecurityMiddleware(waf, csrf, sessionManager, auditLogger)

    // Apply to routes
    mux := http.NewServeMux()

    // Public endpoints
    mux.Handle("/login", security.SecurityHeadersMiddleware(loginHandler))

    // Protected API endpoints
    mux.Handle("/api/", securityMiddleware.Protect(apiHandler))

    // Admin endpoints with extra protection
    mux.Handle("/admin/", securityMiddleware.AdminProtect(adminHandler))

    // Start server
    log.Println("Server starting with comprehensive security...")
    http.ListenAndServe(":8080", mux)
}
```

## Incident Response

### Detected Attack
1. Check audit logs: `auditLogger.QueryLogs()`
2. Review blocked IPs: `waf.GetBlockedIPs()`
3. Analyze pattern: Security incident logged automatically
4. Take action: Manual IP block if needed

### Data Breach
1. Rotate all API keys immediately
2. Invalidate all sessions: `sessionManager.DestroyUserSessions()`
3. Review audit trail for compromised accounts
4. Generate compliance report
5. Notify affected users per regulations

### Vulnerability Discovered
1. Run vulnerability scanner
2. Review compliance checker results
3. Apply fixes
4. Re-run security tests
5. Document in audit log

## Monitoring Dashboard

Key metrics to monitor:
- Active sessions: `sessionManager.GetActiveSessions()`
- Blocked IPs: `waf.GetStats()`
- Failed auth attempts: Query audit logs
- API key rotation status: `keyManager.GetRotationStatus()`
- Security test results: Run weekly

## Support

For security issues, contact: security@rtx-trading.com

**Do not** publicly disclose security vulnerabilities.
