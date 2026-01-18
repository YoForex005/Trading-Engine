# FIX API Provisioning System

Complete FIX API provisioning system with credentials management, rules engine, rate limiting, admin controls, and audit logging.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Admin API Layer                             │
│                   (backend/admin/fix_manager.go)                │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────┐
│              Provisioning Service (provisioning.go)             │
│  - Request validation                                           │
│  - User provisioning workflow                                   │
│  - Session management                                           │
│  - Audit logging coordination                                   │
└──────┬───────────┬──────────────┬───────────────┬───────────────┘
       │           │              │               │
       ▼           ▼              ▼               ▼
┌─────────┐ ┌───────────┐ ┌──────────────┐ ┌──────────────┐
│Credential│ │  Rules    │ │ Rate Limiter │ │Audit Logger  │
│  Store  │ │  Engine   │ │              │ │              │
└─────────┘ └───────────┘ └──────────────┘ └──────────────┘
```

## Components

### 1. Credential Store (`credentials.go`)

Manages FIX credentials with encryption.

**Features:**
- AES-GCM encryption for passwords
- PBKDF2 key derivation
- Credential lifecycle (generate, revoke, regenerate, suspend)
- Secure password generation
- Atomic file operations

**Key Functions:**
```go
// Generate new credentials
creds, err := store.GenerateCredentials(userID, tier, maxSessions, expiresIn)

// Validate login
creds, err := store.ValidateCredentials(senderCompID, password)

// Revoke access
err := store.RevokeCredentials(userID, reason)

// Regenerate password
newPassword, err := store.RegeneratePassword(userID)
```

### 2. Rules Engine (`rules_engine.go`)

Evaluates access rules for FIX API provisioning.

**Built-in Rules:**
- Minimum account balance
- Minimum trading volume (30 days)
- Account age requirement
- KYC verification level
- Group membership
- Custom rules (extensible)

**Key Functions:**
```go
// Evaluate user access
result := engine.EvaluateAccess(userContext)

// Add/remove rules
err := engine.AddRule(rule)
err := engine.RemoveRule(ruleID)

// User-specific rules
err := engine.SetUserRules(userID, ruleIDs)
```

**Default Rules:**
| Rule | Type | Default Value |
|------|------|---------------|
| Minimum Balance | `min_balance` | $1,000 |
| Minimum Volume | `min_volume` | $10,000 (30d) |
| Account Age | `account_age` | 30 days |
| KYC Level | `kyc_level` | Level 2 |

### 3. Rate Limiter (`rate_limiter.go`)

Token bucket algorithm for rate limiting.

**Rate Limit Tiers:**
| Tier | Orders/sec | Messages/sec | Max Sessions | Burst Size |
|------|-----------|--------------|--------------|------------|
| Basic | 5 | 20 | 1 | 10 |
| Standard | 20 | 100 | 3 | 40 |
| Premium | 100 | 500 | 10 | 200 |
| Unlimited | 10,000 | 50,000 | 100 | 10,000 |

**Key Functions:**
```go
// Initialize user rate limiting
err := limiter.InitializeUser(userID, "standard")

// Check limits
allowed, err := limiter.CheckOrderLimit(userID)
allowed, err := limiter.CheckMessageLimit(userID)
allowed, err := limiter.CheckSessionLimit(userID)

// Update tier
err := limiter.UpdateUserTier(userID, "premium")
```

### 4. Provisioning Service (`provisioning.go`)

Orchestrates the provisioning workflow.

**Workflow:**
1. Evaluate access rules
2. Generate credentials
3. Initialize rate limiting
4. Configure IP whitelist (optional)
5. Audit log all operations

**Key Functions:**
```go
// Provision FIX access
resp := service.ProvisionAccess(req)

// Validate login
creds, err := service.ValidateLogin(senderCompID, password, ipAddress)

// Register session
err := service.RegisterSession(sessionID, userID, senderCompID, ip)

// Track message (with rate limiting)
err := service.TrackMessage(userID, sessionID, isOrder)
```

### 5. Admin Manager (`admin/fix_manager.go`)

Admin controls and HTTP API.

**Admin Operations:**
- Provision/revoke user access
- Suspend/reactivate users
- Regenerate passwords
- Manage access rules
- Configure rate limits
- Monitor sessions
- View system statistics

**HTTP Endpoints:**
```
POST   /admin/fix/provision              - Provision FIX access
GET    /admin/fix/credentials            - List all credentials
GET    /admin/fix/credentials/user       - Get user credentials
GET    /admin/fix/sessions               - List active sessions
GET    /admin/fix/sessions/user          - Get user sessions
POST   /admin/fix/sessions/kill          - Kill session
GET    /admin/fix/rate-limits            - List rate limits
POST   /admin/fix/rate-limits/update     - Update rate limit tier
GET    /admin/fix/rules                  - List access rules
POST   /admin/fix/rules/add              - Add rule
GET    /admin/fix/stats                  - System statistics
```

## Usage Examples

### Provision User

```go
// Initialize services
auditLogger := fix.NewSimpleAuditLogger()
service, _ := fix.NewProvisioningService(
    "./data/credentials.json",
    "master-password",
    auditLogger,
)
manager := admin.NewFIXManager(service)

// Provision user
req := &admin.ProvisionUserRequest{
    UserID:         "user123",
    RateLimitTier:  "standard",
    MaxSessions:    3,
    ExpiresInDays:  365,
    AllowedIPs:     []string{"192.168.1.100"},
    AccountBalance: 5000.0,
    TradingVolume:  50000.0,
    AccountAgeDays: 90,
    KYCLevel:       2,
}

resp, _ := manager.ProvisionUser(req)
if resp.Success {
    fmt.Printf("SenderCompID: %s\n", resp.Credentials.SenderCompID)
    fmt.Printf("Password: %s\n", resp.Credentials.Password)
}
```

### Validate Login

```go
creds, err := service.ValidateLogin(
    "USER_ABC123",
    "password",
    "192.168.1.100",
)
if err != nil {
    log.Printf("Login failed: %v", err)
} else {
    // Register session
    service.RegisterSession(
        "session-001",
        creds.UserID,
        creds.SenderCompID,
        "192.168.1.100",
    )
}
```

### Add Custom Rule

```go
rule := &fix.AccessRule{
    ID:          "rule_vip",
    Name:        "VIP Only",
    Type:        fix.RuleTypeGroupMembership,
    Enabled:     true,
    Priority:    100,
    Value:       []string{"vip"},
    ErrorMsg:    "VIP membership required",
}

manager.AddAccessRule(rule)
```

### HTTP API Usage

```bash
# Provision user
curl -X POST http://localhost:8080/admin/fix/provision \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "rate_limit_tier": "standard",
    "max_sessions": 3,
    "account_balance": 5000,
    "trading_volume": 50000,
    "account_age_days": 90,
    "kyc_level": 2
  }'

# Get system stats
curl http://localhost:8080/admin/fix/stats

# List active sessions
curl http://localhost:8080/admin/fix/sessions

# Kill session
curl -X POST "http://localhost:8080/admin/fix/sessions/kill?session_id=session-001"
```

## Integration with Existing FIX Gateway

### 1. Add to Gateway Initialization

```go
// In gateway.go or main.go
provisioning, err := fix.NewProvisioningService(
    "./data/fix_credentials.json",
    os.Getenv("FIX_MASTER_PASSWORD"),
    fix.NewSimpleAuditLogger(),
)
if err != nil {
    log.Fatalf("Failed to initialize provisioning: %v", err)
}
```

### 2. Validate Login on FIX Logon (35=A)

```go
func (gw *Gateway) handleLogon(msg string, session *LPSession) {
    senderCompID := extractField(msg, 49)
    password := extractField(msg, 96) // Tag 96 = Password
    ipAddress := session.conn.RemoteAddr().String()

    // Validate credentials
    creds, err := provisioning.ValidateLogin(senderCompID, password, ipAddress)
    if err != nil {
        sendReject(session, "Invalid credentials")
        return
    }

    // Register session
    provisioning.RegisterSession(
        session.ID,
        creds.UserID,
        creds.SenderCompID,
        ipAddress,
    )

    // Continue with normal logon flow
    acceptLogon(session)
}
```

### 3. Rate Limit Checks on Messages

```go
func (gw *Gateway) handleNewOrder(msg string, session *LPSession) {
    // Get user from session
    userID := getUserFromSession(session.ID)

    // Check rate limit
    err := provisioning.TrackMessage(userID, session.ID, true) // true = is order
    if err != nil {
        sendBusinessReject(session, "Rate limit exceeded")
        return
    }

    // Process order normally
    processOrder(msg, session)
}
```

### 4. Session Cleanup

```go
func (gw *Gateway) handleLogout(session *LPSession) {
    // Unregister session
    provisioning.UnregisterSession(session.ID)

    // Continue with normal logout flow
    sendLogoutResponse(session)
}
```

## Security Considerations

1. **Master Password**: Store in environment variable or secret manager
2. **Encryption**: AES-256-GCM with PBKDF2 key derivation
3. **Password Display**: Only shown once during generation
4. **Audit Logging**: All operations logged with timestamps
5. **IP Whitelist**: Optional IP-based access control
6. **Session Limits**: Prevent credential sharing

## File Structure

```
backend/
├── fix/
│   ├── credentials.go          # Credential management
│   ├── rules_engine.go         # Access rules
│   ├── rate_limiter.go         # Rate limiting
│   ├── provisioning.go         # Main service
│   └── examples/
│       └── provisioning_example.go
└── admin/
    └── fix_manager.go          # Admin API
```

## Testing

```go
// Run example
go run backend/fix/examples/provisioning_example.go

// Unit tests
go test ./backend/fix -v
go test ./backend/admin -v
```

## Future Enhancements

- [ ] Database persistence (currently JSON file)
- [ ] Advanced audit logging (database, external service)
- [ ] OAuth2/JWT for admin API
- [ ] Real-time monitoring dashboard
- [ ] Email notifications for credential events
- [ ] Multi-factor authentication for high-tier users
- [ ] Geolocation-based access control
- [ ] Integration with compliance systems

## License

See project LICENSE file.
