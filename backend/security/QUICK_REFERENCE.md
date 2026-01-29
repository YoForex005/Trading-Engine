# Security Package - Quick Reference Card

## 🚀 Quick Start

```go
// 1. Initialize security
security, _ := security.InitializeSecurity()

// 2. Run checks
security.RunSecurityChecks()

// 3. Apply to routes
security.ApplyToRouter(mux)
```

## 📦 Core Components

| Component | Purpose | File |
|-----------|---------|------|
| WAF | Rate limiting, IP blocking, DDoS protection | `waf.go` |
| Encryption | AES-256-GCM data encryption | `encryption.go` |
| Audit Logger | Security event logging | `audit.go` |
| Compliance | OWASP Top 10 checks | `compliance.go` |
| Scanner | Vulnerability scanning | `scanner.go` |
| CSRF | Cross-site request forgery protection | `csrf.go` |
| Sessions | Session management | `session.go` |
| API Keys | Key rotation and management | `apikeys.go` |
| Middleware | HTTP security middleware | `middleware.go` |
| Testing | Security test framework | `testing.go` |

## 🔒 Common Operations

### WAF - Block/Unblock IP
```go
waf := security.NewWAF(config)

// Block IP temporarily
waf.BlockIP("1.2.3.4", 15*time.Minute)

// Block IP permanently
waf.BlockIPPermanent("5.6.7.8")

// Unblock IP
waf.UnblockIP("1.2.3.4")

// Get stats
stats := waf.GetStats()

// Get blocked IPs
blockedIPs := waf.GetBlockedIPs()
```

### Encryption - Encrypt/Decrypt Data
```go
enc := security.NewEncryptionService("master-key")

// Encrypt
encrypted, _ := enc.EncryptString("sensitive-data")

// Decrypt
decrypted, _ := enc.DecryptString(encrypted)
```

### Audit Logging
```go
logger, _ := security.NewAuditLogger("/var/log/rtx/audit")

// Log authentication
logger.LogAuthAttempt("user123", "1.2.3.4", true, "success")

// Log admin action
logger.LogAdminAction("admin", "1.2.3.4", "deposit", "acc-001", true, metadata)

// Log security incident
logger.LogSecurityIncident("auth", "brute_force", "1.2.3.4", "detected", metadata)

// Query logs
events, _ := logger.QueryLogs(startTime, endTime, security.AuditLevelSecurity, "")
```

### Compliance Checking
```go
checker := security.NewComplianceChecker(logger)

// Run all checks
report, _ := checker.RunAllChecks()

fmt.Printf("Status: %s, Passed: %d, Failed: %d\n",
    report.OverallStatus, report.Passed, report.Failed)

// Validate input
err := security.ValidateInput("test@example.com", "email")
err = security.ValidateInput("user123", "alphanumeric")

// Sanitize input (XSS prevention)
safe := security.SanitizeInput(userInput)
```

### Vulnerability Scanning
```go
scanner := security.NewVulnerabilityScanner(logger)

// Scan directory
report, _ := scanner.ScanDirectory("/path/to/code")

// Scan for secrets only
secrets, _ := scanner.ScanSecrets("/path/to/code")

// Generate report
reportText := scanner.GenerateReport(report)
```

### CSRF Protection
```go
csrf := security.NewCSRFProtection("secret")

// Generate token for session
token, _ := csrf.GenerateToken(sessionID)

// Validate token
err := csrf.ValidateToken(sessionID, token)

// Apply middleware
http.Handle("/api/", csrf.Middleware(handler))
```

### Session Management
```go
sessionMgr := security.NewSessionManager(config, logger)

// Create session
session, _ := sessionMgr.CreateSession("user123", "1.2.3.4")

// Validate session
validSession, _ := sessionMgr.ValidateSession(session.ID, "1.2.3.4")

// Destroy session
sessionMgr.DestroySession(session.ID)

// Destroy all user sessions
sessionMgr.DestroyUserSessions("user123")

// Get active sessions
count := sessionMgr.GetActiveSessions()
```

### API Key Management
```go
keyMgr := security.NewAPIKeyManager(90*24*time.Hour, logger, enc)

// Generate key
apiKey, _ := keyMgr.GenerateKey("service-name")

// Get current key
key, _ := keyMgr.GetKey("service-name")

// Rotate key
newKey, _ := keyMgr.RotateKey("service-name", "scheduled_rotation")

// Revoke key
keyMgr.RevokeKey("service-name", "security_breach")

// Get rotation status
status := keyMgr.GetRotationStatus()
```

### Security Middleware
```go
securityMw := security.NewSecurityMiddleware(waf, csrf, sessionMgr, logger)

// Protect endpoints
http.Handle("/api/", securityMw.Protect(apiHandler))

// Admin-only endpoints
http.Handle("/admin/", securityMw.AdminProtect(adminHandler))

// Add security headers
http.Handle("/", security.SecurityHeadersMiddleware(handler))

// CORS
http.Handle("/api/", security.CORSMiddleware(allowedOrigins)(handler))

// Input sanitization
http.Handle("/api/", security.InputSanitizationMiddleware(handler))
```

### Security Testing
```go
testSuite := security.NewSecurityTestSuite("http://localhost:8080")
testSuite.RegisterDefaultTests()

// Run all tests
report := testSuite.RunAll()

fmt.Printf("Passed: %d, Failed: %d\n", report.Passed, report.Failed)

// Add custom test
testSuite.AddTest(security.SecurityTest{
    Name: "Custom Test",
    TestFunc: func() (bool, string) {
        // Your test logic
        return true, "Test passed"
    },
})
```

## 🎯 Common Patterns

### Login Handler with Security
```go
func handleLogin(w http.ResponseWriter, r *http.Request) {
    // Get IP
    clientIP := extractIP(r)

    // Parse credentials
    var creds struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    json.NewDecoder(r.Body).Decode(&creds)

    // Validate & sanitize
    creds.Username = security.SanitizeInput(creds.Username)

    // Authenticate
    user, err := authenticateUser(creds.Username, creds.Password)

    // Log attempt
    auditLogger.LogAuthAttempt(creds.Username, clientIP, err == nil, "login")

    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Create session
    session, _ := sessionManager.CreateSession(user.ID, clientIP)

    // Generate CSRF token
    csrfToken, _ := csrf.GenerateToken(session.ID)

    // Return token
    json.NewEncoder(w).Encode(map[string]string{
        "session_id": session.ID,
        "csrf_token": csrfToken,
    })
}
```

### Protected API Endpoint
```go
func handleTransfer(w http.ResponseWriter, r *http.Request) {
    // Extract session
    sessionID := r.Header.Get("Authorization")
    clientIP := extractIP(r)

    // Validate session
    session, err := sessionManager.ValidateSession(sessionID, clientIP)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // CSRF token already validated by middleware

    // Process transfer
    // ...

    // Log action
    auditLogger.Log(security.AuditEvent{
        Level:    security.AuditLevelCritical,
        Category: "trading",
        Action:   "transfer",
        UserID:   session.UserID,
        IP:       clientIP,
        Success:  true,
        Message:  "Transfer completed",
    })
}
```

### Emergency Lockdown
```go
func emergencyLockdown(security *SecuritySetup) {
    log.Println("🚨 EMERGENCY LOCKDOWN INITIATED")

    // 1. Enable aggressive rate limiting
    security.WAF.config.MaxRequestsPerIP = 5

    // 2. Rotate all API keys
    services := []string{"oanda", "binance"}
    for _, service := range services {
        security.APIKeyManager.RotateKey(service, "emergency")
    }

    // 3. Log incident
    security.AuditLogger.LogSecurityIncident(
        "emergency", "lockdown", "", "Emergency lockdown", nil,
    )

    log.Println("✅ Lockdown complete")
}
```

## 📊 Security Metrics

```go
// WAF stats
wafStats := waf.GetStats()
// -> blocked_ips, tracked_ips, active_connections

// Session count
activeSessions := sessionManager.GetActiveSessions()

// API key status
rotationStatus := keyManager.GetRotationStatus()

// Compliance status
complianceReport, _ := complianceChecker.RunAllChecks()
// -> overall_status, passed, failed

// Vulnerability count
scanReport, _ := scanner.ScanDirectory(".")
// -> critical, high, medium, low
```

## ⚡ Performance

| Operation | Latency | Notes |
|-----------|---------|-------|
| WAF check | ~1-2ms | Per request |
| CSRF validation | ~0.5ms | HMAC verify |
| Session validation | ~0.5ms | Map lookup |
| Encryption | ~2-3ms | AES-256-GCM |
| Audit log | ~1ms | Async write |

## 🔧 Configuration

### Production Settings
```go
// WAF
wafConfig := &security.WAFConfig{
    MaxRequestsPerIP:     50,
    MaxConcurrentConns:   20000,
    BlockDuration:        30 * time.Minute,
    AdminWhitelist:       []string{"10.0.0.1"},
    WhitelistEnabled:     true,
}

// Sessions
sessionConfig := &security.SessionConfig{
    Timeout:              30 * time.Minute,
    MaxConcurrentPerUser: 3,
    RequireIP:            true,
}

// API Keys
keyRotation := 90 * 24 * time.Hour // 90 days
```

## 🆘 Emergency Commands

```bash
# Run security checks
go run scripts/run_security_checks.go

# Scan for vulnerabilities
go run scripts/vulnerability_scan.go

# Run compliance audit
go run scripts/compliance_check.go

# Block IP manually
# (Add to WAF blocklist in code)

# Rotate all keys
# (Call RotateKey for each service)
```

## 📞 Support

**Security Issues:** security@rtx-trading.com
**Incident Hotline:** [24/7 hotline]
**Documentation:** `/backend/security/README.md`

---

*Quick Reference v1.0 - January 2026*
