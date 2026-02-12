# How to Apply Security Middleware to cmd/server/main.go

**Quick Guide:** Apply authentication and CORS middleware to protect routes.

---

## Step 1: Add Imports

Add to imports section in `cmd/server/main.go`:

```go
import (
    // ... existing imports ...
    "github.com/epic1st/rtx/backend/internal/middleware"
)
```

---

## Step 2: Initialize Middleware (After authService creation)

Add after line ~200 where `authService` is created:

```go
// ============================================
// SECURITY MIDDLEWARE INITIALIZATION
// ============================================
log.Println("[SECURITY] Initializing security middleware...")

// CORS Middleware
corsMiddleware := middleware.NewCORSMiddleware(middleware.CORSConfig{
    AllowedOrigins: cfg.CORS.AllowedOrigins,
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders: []string{"Content-Type", "Authorization"},
})

// Rate Limiting Middleware
rateLimitConfig := middleware.RateLimitConfig{
    RequestsPerSecond: 10,
    RequestsPerMinute: 500,
    BurstSize:         20,
    CleanupInterval:   5 * time.Minute,
    ClientTimeout:     10 * time.Minute,
}
rateLimiter := middleware.NewRateLimiter(rateLimitConfig)

// Authentication Middleware
requireAuth := middleware.RequireAuth(authService)
requireAdmin := middleware.RequireAdmin(authService)

log.Println("[SECURITY] ✓ Security middleware initialized")
```

---

## Step 3: Wrap Routes with Middleware

### Pattern 1: Public Routes (CORS + Rate Limiting)
```go
// Health check - public
http.Handle("/health", corsMiddleware(rateLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}))))
```

### Pattern 2: Login Route (CORS + Stricter Rate Limiting)
```go
// Login - needs stricter rate limiting (5 req/min)
strictRateLimiter := middleware.NewRateLimiter(middleware.RateLimitConfig{
    RequestsPerSecond: 0.083, // ~5 requests per minute
    RequestsPerMinute: 5,
    BurstSize:         3,
    CleanupInterval:   5 * time.Minute,
    ClientTimeout:     10 * time.Minute,
})

http.Handle("/login", corsMiddleware(strictRateLimiter.Middleware(http.HandlerFunc(server.HandleLogin))))
```

### Pattern 3: User Routes (CORS + Rate Limiting + Auth)
```go
// User endpoints - require authentication
http.Handle("/api/positions", corsMiddleware(rateLimiter.Middleware(requireAuth(http.HandlerFunc(apiHandler.HandleGetPositions)))))
http.Handle("/api/orders", corsMiddleware(rateLimiter.Middleware(requireAuth(http.HandlerFunc(apiHandler.HandleGetOrders)))))
```

### Pattern 4: Admin Routes (CORS + Rate Limiting + Admin Auth)
```go
// Admin endpoints - require admin role
http.Handle("/api/routing/rules", corsMiddleware(rateLimiter.Middleware(requireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        apiHandler.HandleListRoutingRules(w, r)
    } else if r.Method == "POST" {
        apiHandler.HandleCreateRoutingRule(w, r)
    }
})))))

http.Handle("/api/compliance/", corsMiddleware(rateLimiter.Middleware(requireAdmin(complianceHandler))))
http.Handle("/api/analytics/", corsMiddleware(rateLimiter.Middleware(requireAdmin(analyticsHandler))))
```

---

## Step 4: Complete Example

Replace existing route registration with protected versions:

```go
// ============================================
// ROUTE REGISTRATION WITH SECURITY
// ============================================
log.Println("[ROUTES] Registering secured API endpoints...")

// Public routes (no auth required)
http.Handle("/health", corsMiddleware(rateLimiter.Middleware(http.HandlerFunc(healthHandler))))
http.Handle("/docs", corsMiddleware(rateLimiter.Middleware(http.HandlerFunc(docsHandler))))

// Authentication
http.Handle("/login", corsMiddleware(strictRateLimiter.Middleware(http.HandlerFunc(server.HandleLogin))))

// User-authenticated routes
userRoutes := []struct {
    pattern string
    handler http.HandlerFunc
}{
    {"/api/account/summary", apiHandler.HandleGetAccountSummary},
    {"/api/positions", apiHandler.HandleGetPositions},
    {"/api/positions/close", apiHandler.HandleClosePosition},
    {"/api/orders", apiHandler.HandleGetOrders},
    {"/api/orders/market", apiHandler.HandlePlaceMarketOrder},
    {"/api/symbols", apiHandler.HandleGetSymbols},
}

for _, route := range userRoutes {
    http.Handle(route.pattern, corsMiddleware(rateLimiter.Middleware(requireAuth(route.handler))))
}

// Admin-only routes
adminRoutes := []struct {
    pattern string
    handler http.Handler
}{
    {"/api/routing/rules", routingRulesHandler},
    {"/api/compliance/best-execution", http.HandlerFunc(complianceHandler.HandleBestExecution)},
    {"/api/compliance/audit-trail", http.HandlerFunc(complianceHandler.HandleAuditTrail)},
    {"/api/analytics/routing/breakdown", http.HandlerFunc(apiHandler.HandleRoutingBreakdown)},
}

for _, route := range adminRoutes {
    http.Handle(route.pattern, corsMiddleware(rateLimiter.Middleware(requireAdmin(route.handler))))
}

log.Println("[ROUTES] ✓ All routes secured with middleware")
```

---

## Step 5: Remove Old CORS Headers

Search and remove all inline CORS headers:

**Find:**
```go
w.Header().Set("Access-Control-Allow-Origin", "*")
w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

if r.Method == "OPTIONS" {
    w.WriteHeader(http.StatusOK)
    return
}
```

**Replace with:** (nothing - middleware handles it)

---

## Step 6: Test the Changes

### Test 1: Verify CORS Works
```bash
# Should succeed (localhost allowed by default)
curl -H "Origin: http://localhost:3000" \
     http://localhost:7999/api/config

# Should fail (unauthorized origin)
curl -H "Origin: https://evil.com" \
     http://localhost:7999/api/config
```

### Test 2: Verify Authentication
```bash
# Should return 401 (no token)
curl http://localhost:7999/api/positions

# Login first
TOKEN=$(curl -X POST http://localhost:7999/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"password"}' \
     | jq -r '.token')

# Should succeed (with token)
curl -H "Authorization: Bearer $TOKEN" \
     http://localhost:7999/api/positions
```

### Test 3: Verify Rate Limiting
```bash
# Send 100 requests rapidly
for i in {1..100}; do
    curl http://localhost:7999/api/symbols &
done

# Should see "429 Too Many Requests" after ~20 requests
```

### Test 4: Verify Admin Protection
```bash
# Login as regular user
USER_TOKEN=$(curl -X POST http://localhost:7999/login \
     -H "Content-Type: application/json" \
     -d '{"username":"trader1","password":"password"}' \
     | jq -r '.token')

# Should return 403 Forbidden (not admin)
curl -H "Authorization: Bearer $USER_TOKEN" \
     http://localhost:7999/api/routing/rules

# Login as admin
ADMIN_TOKEN=$(curl -X POST http://localhost:7999/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"password"}' \
     | jq -r '.token')

# Should succeed (admin token)
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
     http://localhost:7999/api/routing/rules
```

---

## Common Issues & Solutions

### Issue 1: "CORS middleware not working"
**Solution:** Check that `ALLOWED_ORIGINS` is set in `.env`:
```bash
export ALLOWED_ORIGINS="http://localhost:3000,http://localhost:5173"
```

### Issue 2: "Rate limiting too strict"
**Solution:** Adjust rate limit config:
```go
rateLimitConfig := middleware.RateLimitConfig{
    RequestsPerSecond: 50,  // Increase if needed
    BurstSize:         100, // Increase burst size
}
```

### Issue 3: "Authentication fails with valid token"
**Solution:** Check JWT_SECRET is consistent:
```bash
# Ensure JWT_SECRET in .env matches what was used to generate token
echo $JWT_SECRET
```

### Issue 4: "OPTIONS preflight fails"
**Solution:** CORS middleware handles OPTIONS automatically. Ensure you're using the middleware.

---

## Checklist

- [ ] Imports added to `cmd/server/main.go`
- [ ] Middleware initialized after `authService`
- [ ] `/login` route uses strict rate limiter
- [ ] User routes wrapped with `requireAuth`
- [ ] Admin routes wrapped with `requireAdmin`
- [ ] All routes wrapped with `corsMiddleware`
- [ ] All routes wrapped with `rateLimiter.Middleware`
- [ ] Old inline CORS headers removed
- [ ] Tested CORS with curl
- [ ] Tested authentication with curl
- [ ] Tested rate limiting with curl
- [ ] Tested admin protection with curl

---

## Performance Impact

**Middleware Overhead:**
- CORS: ~0.1ms per request
- Rate Limiting: ~0.05ms per request
- Authentication: ~0.5ms per request (JWT validation)

**Total:** ~0.65ms per protected request (negligible for trading platform)

---

## Quick Reference: Middleware Chain Order

**Order matters!** Apply middleware in this sequence:

```
Request → CORS → Rate Limiting → Authentication → Handler
```

**Why this order?**
1. **CORS first** - Reject invalid origins immediately
2. **Rate Limiting second** - Prevent abuse before expensive operations
3. **Authentication last** - Validate credentials only for allowed, rate-limited requests

**Correct:**
```go
http.Handle("/api/", corsMiddleware(rateLimiter.Middleware(requireAuth(handler))))
```

**Incorrect:**
```go
http.Handle("/api/", requireAuth(corsMiddleware(rateLimiter.Middleware(handler))))
// Auth runs before rate limiting - bad!
```

---

## Done!

After applying these changes:
1. Restart the server
2. Run all 4 tests above
3. Monitor logs for `[SECURITY]` messages
4. Review `SECURITY_AUDIT_REPORT.md` for additional hardening

**Questions?** See `SECURITY_FIXES_APPLIED.md` for detailed information.
