# FIX Provisioning Integration Guide

Quick guide to integrate the FIX provisioning system with your existing FIX gateway.

## Step 1: Initialize Provisioning Service

Add to your `main.go` or gateway initialization:

```go
package main

import (
    "log"
    "os"
    "backend/fix"
    "backend/admin"
)

func main() {
    // Get master password from environment
    masterPassword := os.Getenv("FIX_MASTER_PASSWORD")
    if masterPassword == "" {
        log.Fatal("FIX_MASTER_PASSWORD environment variable required")
    }

    // Initialize audit logger
    auditLogger := fix.NewSimpleAuditLogger()

    // Initialize provisioning service
    provisioningService, err := fix.NewProvisioningService(
        "./data/fix_credentials.json",
        masterPassword,
        auditLogger,
    )
    if err != nil {
        log.Fatalf("Failed to initialize provisioning: %v", err)
    }

    // Initialize admin manager (optional - for admin API)
    fixManager := admin.NewFIXManager(provisioningService)

    // Start admin HTTP server (optional)
    go startAdminServer(fixManager)

    // Your existing gateway initialization...
    gateway := NewGateway(provisioningService)
    gateway.Start()
}
```

## Step 2: Modify FIX Logon Handler

Update your logon handler to validate credentials:

```go
// In gateway.go - handleLogon function
func (gw *Gateway) handleLogon(msg string, conn net.Conn) error {
    // Extract FIX fields
    senderCompID := extractField(msg, 49)  // Tag 49 = SenderCompID
    password := extractField(msg, 96)      // Tag 96 = Password (or tag 554)

    // Get client IP
    ipAddress := conn.RemoteAddr().String()

    // Validate credentials using provisioning service
    creds, err := gw.provisioning.ValidateLogin(senderCompID, password, ipAddress)
    if err != nil {
        log.Printf("Login failed for %s from %s: %v", senderCompID, ipAddress, err)

        // Send FIX Logout (35=5) with reason
        gw.sendLogout(conn, fmt.Sprintf("Invalid credentials: %v", err))
        return err
    }

    // Create session
    session := &LPSession{
        ID:           generateSessionID(),
        SenderCompID: creds.SenderCompID,
        TargetCompID: creds.TargetCompID,
        conn:         conn,
        Status:       "LOGGED_IN",
    }

    // Register session with provisioning service
    err = gw.provisioning.RegisterSession(
        session.ID,
        creds.UserID,
        creds.SenderCompID,
        ipAddress,
    )
    if err != nil {
        log.Printf("Failed to register session: %v", err)
        return err
    }

    // Store session
    gw.sessions[session.ID] = session

    // Send Logon response (35=A)
    gw.sendLogonResponse(session)

    log.Printf("User %s logged in successfully from %s", creds.UserID, ipAddress)
    return nil
}
```

## Step 3: Add Rate Limiting to Message Handlers

Add rate limit checks before processing orders or messages:

```go
// In gateway.go - handleNewOrderSingle function
func (gw *Gateway) handleNewOrderSingle(msg string, session *LPSession) error {
    // Get user from session
    userID := gw.getUserFromSession(session.ID)

    // Check rate limit (isOrder = true)
    err := gw.provisioning.TrackMessage(userID, session.ID, true)
    if err != nil {
        log.Printf("Order rate limit exceeded for user %s: %v", userID, err)

        // Send Business Reject (35=j)
        gw.sendBusinessReject(session, "Rate limit exceeded")
        return err
    }

    // Process order normally
    return gw.processNewOrder(msg, session)
}

// For other messages (market data requests, etc.)
func (gw *Gateway) handleMarketDataRequest(msg string, session *LPSession) error {
    userID := gw.getUserFromSession(session.ID)

    // Check rate limit (isOrder = false)
    err := gw.provisioning.TrackMessage(userID, session.ID, false)
    if err != nil {
        log.Printf("Message rate limit exceeded for user %s: %v", userID, err)

        // Send Reject
        gw.sendReject(session, "Rate limit exceeded")
        return err
    }

    // Process request normally
    return gw.processMarketDataRequest(msg, session)
}
```

## Step 4: Clean Up on Logout

Unregister sessions when users logout:

```go
// In gateway.go - handleLogout function
func (gw *Gateway) handleLogout(session *LPSession) error {
    // Unregister session from provisioning
    err := gw.provisioning.UnregisterSession(session.ID)
    if err != nil {
        log.Printf("Warning: Failed to unregister session %s: %v", session.ID, err)
    }

    // Send Logout response (35=5)
    gw.sendLogoutResponse(session)

    // Close connection
    session.conn.Close()

    // Remove from sessions map
    delete(gw.sessions, session.ID)

    return nil
}
```

## Step 5: Add Admin HTTP Server (Optional)

```go
func startAdminServer(fixManager *admin.FIXManager) {
    mux := http.NewServeMux()

    // Register FIX provisioning endpoints
    fixManager.RegisterHTTPHandlers(mux)

    // Add authentication middleware (recommended)
    authMux := addAuthMiddleware(mux)

    log.Println("Starting admin server on :8080")
    if err := http.ListenAndServe(":8080", authMux); err != nil {
        log.Fatalf("Admin server failed: %v", err)
    }
}

// Simple auth middleware example
func addAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check API key or JWT token
        apiKey := r.Header.Get("X-API-Key")
        if apiKey != os.Getenv("ADMIN_API_KEY") {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

## Step 6: Helper Functions

Add these helper functions to your gateway:

```go
// Extract FIX field value
func extractField(msg string, tag int) string {
    tagStr := fmt.Sprintf("%d=", tag)
    start := strings.Index(msg, tagStr)
    if start == -1 {
        return ""
    }
    start += len(tagStr)

    end := strings.Index(msg[start:], "\x01")
    if end == -1 {
        return msg[start:]
    }

    return msg[start : start+end]
}

// Get user ID from session
func (gw *Gateway) getUserFromSession(sessionID string) string {
    // You'll need to store this mapping when session is created
    if session, exists := gw.sessions[sessionID]; exists {
        return session.UserID // Add UserID field to your LPSession struct
    }
    return ""
}

// Generate unique session ID
func generateSessionID() string {
    return fmt.Sprintf("session-%d", time.Now().UnixNano())
}
```

## Environment Variables

Set these environment variables:

```bash
# Required
export FIX_MASTER_PASSWORD="your-secure-master-password-here"

# Optional - for admin API
export ADMIN_API_KEY="your-admin-api-key"

# Optional - custom paths
export FIX_CREDENTIALS_PATH="./data/fix_credentials.json"
```

## Testing the Integration

### 1. Provision a Test User

```bash
curl -X POST http://localhost:8080/admin/fix/provision \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-admin-api-key" \
  -d '{
    "user_id": "testuser123",
    "rate_limit_tier": "standard",
    "max_sessions": 3,
    "account_balance": 5000,
    "trading_volume": 50000,
    "account_age_days": 90,
    "kyc_level": 2,
    "allowed_ips": ["127.0.0.1"]
  }'
```

Response:
```json
{
  "success": true,
  "credentials": {
    "sender_comp_id": "USER_XYZ123",
    "target_comp_id": "GATEWAY",
    "password": "abc123xyz..."
  }
}
```

### 2. Test FIX Login

Use the credentials to connect via FIX:

```
8=FIX.4.4|9=120|35=A|49=USER_XYZ123|56=GATEWAY|34=1|52=20260118-10:30:00|96=abc123xyz...|98=0|108=30|10=123|
```

### 3. Monitor Sessions

```bash
# View all active sessions
curl http://localhost:8080/admin/fix/sessions \
  -H "X-API-Key: your-admin-api-key"

# View specific user sessions
curl "http://localhost:8080/admin/fix/sessions/user?user_id=testuser123" \
  -H "X-API-Key: your-admin-api-key"
```

### 4. Check Rate Limits

```bash
curl "http://localhost:8080/admin/fix/rate-limits" \
  -H "X-API-Key: your-admin-api-key"
```

## Common Integration Patterns

### Pattern 1: Pre-Validate Before FIX Connection

```go
// Check if user is allowed to connect before accepting socket
func (gw *Gateway) acceptConnection(conn net.Conn) {
    // Read initial logon message
    msg := readFIXMessage(conn)

    senderCompID := extractField(msg, 49)
    password := extractField(msg, 96)
    ipAddress := conn.RemoteAddr().String()

    // Validate immediately
    _, err := gw.provisioning.ValidateLogin(senderCompID, password, ipAddress)
    if err != nil {
        // Reject connection immediately
        sendLogout(conn, "Access denied")
        conn.Close()
        return
    }

    // Continue with normal logon process
    gw.handleLogon(msg, conn)
}
```

### Pattern 2: Real-Time Rate Limit Updates

```go
// Allow admins to update rate limits without restart
func (gw *Gateway) updateUserRateLimit(userID, newTier string) error {
    return gw.provisioning.GetRateLimiter().UpdateUserTier(userID, newTier)
}
```

### Pattern 3: Emergency Session Kill

```go
// Kill all sessions for a user (e.g., suspicious activity)
func (gw *Gateway) killUserSessions(userID string) error {
    sessions := gw.provisioning.GetUserSessions(userID)

    for _, session := range sessions {
        // Find and close connection
        if lpSession, exists := gw.sessions[session.SessionID]; exists {
            gw.sendLogout(lpSession, "Session terminated by admin")
            lpSession.conn.Close()
        }

        // Unregister from provisioning
        gw.provisioning.UnregisterSession(session.SessionID)
    }

    return nil
}
```

## Monitoring & Observability

Add metrics collection:

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    loginAttempts = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "fix_login_attempts_total",
            Help: "Total FIX login attempts",
        },
        []string{"status"}, // success, failed
    )

    rateLimitExceeded = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "fix_rate_limit_exceeded_total",
            Help: "Total rate limit violations",
        },
        []string{"user_id", "type"}, // type: order, message
    )
)

func init() {
    prometheus.MustRegister(loginAttempts)
    prometheus.MustRegister(rateLimitExceeded)
}
```

## Security Best Practices

1. **Never log passwords**: Only log hashed or redacted values
2. **Use HTTPS**: Secure admin API with TLS
3. **API authentication**: Implement JWT or API key auth
4. **IP whitelisting**: Use for sensitive admin endpoints
5. **Audit trail**: Store all provisioning operations
6. **Password rotation**: Implement regular password regeneration
7. **Session timeouts**: Auto-logout inactive sessions

## Troubleshooting

### Issue: Login fails with "credentials not found"
- Check if user was provisioned
- Verify SenderCompID matches provisioned value
- Check credential status (active vs revoked/suspended)

### Issue: Rate limit exceeded immediately
- Check user's rate limit tier
- Verify tier configuration is correct
- Look for burst size limitations

### Issue: Session not registered
- Ensure RegisterSession is called after successful validation
- Check for session ID conflicts
- Verify provisioning service is initialized

### Issue: IP whitelist blocking legitimate users
- Verify allowed_ips includes user's IP
- Check for NAT/proxy IP address changes
- Consider using IP ranges instead of specific IPs

## Next Steps

1. Add database persistence for credentials
2. Implement advanced audit logging
3. Add email notifications for credential events
4. Create admin dashboard UI
5. Add compliance reporting features
6. Integrate with monitoring systems (Prometheus, Grafana)

## Support

For questions or issues:
- Check README_PROVISIONING.md for detailed API docs
- Review example code in examples/provisioning_example.go
- Examine audit logs for troubleshooting
