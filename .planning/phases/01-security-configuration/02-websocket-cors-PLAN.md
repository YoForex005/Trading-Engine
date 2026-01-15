---
must_haves:
  truths:
    - WebSocket connections validate origin against whitelist (not allowing all origins)
    - CORS validation configured via environment variable (ALLOWED_ORIGINS)
    - Invalid origins rejected with clear error logging
    - Localhost origins allowed for development environment
  artifacts:
    - Updated backend/ws/hub.go with CORS validation logic
    - Environment variable ALLOWED_ORIGINS documented in .env.example
    - Test cases confirming CORS validation works
  key_links:
    - backend/ws/hub.go:15-18 (CORS CheckOrigin function)
wave: 1
---

# Plan: WebSocket Security & CORS Validation

## Objective

Implement production-grade CORS validation for WebSocket connections to prevent unauthorized cross-origin access. This plan replaces the insecure "allow all origins" configuration with an environment-configurable origin whitelist.

## Execution Context

**Requirements addressed:**
- SECURITY-03: Implement CORS validation with origin whitelist for WebSocket

**Reference files:**
- backend/ws/hub.go (lines 15-18: insecure CheckOrigin function)
- .env.example (need to add ALLOWED_ORIGINS)

**Codebase context:**
- Go 1.24.0 backend with gorilla/websocket
- WebSocket upgrader at backend/ws/hub.go:15-18
- Environment variable system from Plan 01 (godotenv already added)

## Context

**Current Security Issue:**

The WebSocket upgrader in `backend/ws/hub.go` lines 15-18 allows **all origins**:

```go
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true  // SECURITY RISK: Allows any origin
    },
}
```

This allows any website to connect to the WebSocket endpoint, enabling:
- Unauthorized data access from malicious sites
- Cross-site WebSocket hijacking (CSWSH)
- Data exfiltration attacks

**Required Fix:**

Implement origin whitelist validation:
1. Load allowed origins from `ALLOWED_ORIGINS` environment variable
2. Parse comma-separated list of allowed origins
3. Validate incoming WebSocket connections against whitelist
4. Log and reject unauthorized origins
5. Support wildcard patterns for development (localhost variants)

**Environment Variable Format:**
```
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000,https://app.example.com
```

**gorilla/websocket CheckOrigin Signature:**
```go
CheckOrigin: func(r *http.Request) bool
```
- Returns `true` to allow connection
- Returns `false` to reject with 403 Forbidden
- Can inspect `r.Header.Get("Origin")`

## Tasks

### Task 1: Add ALLOWED_ORIGINS to environment configuration
**Action:** Document WebSocket CORS configuration in .env.example
**Files:**
- `.env.example` - Add ALLOWED_ORIGINS variable

**Steps:**
1. Add to `.env.example` in the WebSocket section:
   ```
   # WebSocket CORS (comma-separated origins)
   # Development: Include all local development server origins
   # Production: Only include your production frontend domain
   ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000,http://localhost:8081
   ```

2. Add helpful comment explaining security importance:
   ```
   # SECURITY: Only list trusted frontend origins
   # Any origin listed here can connect to your WebSocket and receive market data
   ```

3. Update local `.env` with development origins:
   ```bash
   echo "ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000,http://localhost:8081" >> .env
   ```

**Verification:**
- `.env.example` contains ALLOWED_ORIGINS with clear documentation
- `.env` contains ALLOWED_ORIGINS with development values

---

### Task 2: Implement origin whitelist parsing
**Action:** Create helper function to parse and validate allowed origins
**Files:**
- `backend/ws/hub.go` - Add origin parsing logic

**Steps:**
1. Add imports if not present:
   ```go
   import (
       // ... existing imports
       "os"
       "strings"
   )
   ```

2. Add global variable for parsed allowed origins (after `upgrader` declaration, around line 19):
   ```go
   var upgrader = websocket.Upgrader{
       CheckOrigin: func(r *http.Request) bool {
           return true  // Will be replaced in Task 3
       },
   }

   // Allowed origins for WebSocket CORS validation
   var allowedOrigins []string
   ```

3. Add `init()` function to parse ALLOWED_ORIGINS on startup:
   ```go
   func init() {
       // Load allowed origins from environment variable
       originsEnv := os.Getenv("ALLOWED_ORIGINS")
       if originsEnv == "" {
           log.Println("[WARN] ALLOWED_ORIGINS not set - WebSocket will reject all connections")
           log.Println("[WARN] Set ALLOWED_ORIGINS=http://localhost:5173,... in .env for development")
           allowedOrigins = []string{} // Empty whitelist = reject all
           return
       }

       // Parse comma-separated origins
       rawOrigins := strings.Split(originsEnv, ",")
       for _, origin := range rawOrigins {
           trimmed := strings.TrimSpace(origin)
           if trimmed != "" {
               allowedOrigins = append(allowedOrigins, trimmed)
           }
       }

       log.Printf("[WebSocket] CORS allowed origins: %v", allowedOrigins)
   }
   ```

**Verification:**
- `allowedOrigins` variable declared after `upgrader`
- `init()` function parses ALLOWED_ORIGINS env var
- Warning logged if ALLOWED_ORIGINS not set
- Startup logs show parsed origins list

---

### Task 3: Implement origin validation in CheckOrigin
**Action:** Replace insecure "return true" with whitelist validation
**Files:**
- `backend/ws/hub.go` - Update CheckOrigin function

**Steps:**
1. Replace the CheckOrigin function (lines 16-18):
   ```go
   var upgrader = websocket.Upgrader{
       CheckOrigin: func(r *http.Request) bool {
           origin := r.Header.Get("Origin")

           // Allow requests with no Origin header (same-origin, non-browser clients)
           if origin == "" {
               return true
           }

           // Check against whitelist
           for _, allowed := range allowedOrigins {
               if origin == allowed {
                   return true
               }
           }

           // Reject unauthorized origin
           log.Printf("[WebSocket] REJECTED connection from unauthorized origin: %s (client: %s)", origin, r.RemoteAddr)
           return false
       },
   }
   ```

2. Add detailed logging for accepted connections in `ServeWs` function (around line 238):
   ```go
   func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
       origin := r.Header.Get("Origin")
       log.Printf("[WS] Connection request from %s (origin: %s)", r.RemoteAddr, origin)

       conn, err := upgrader.Upgrade(w, r, nil)
       // ... rest of existing code
   ```

**Verification:**
- CheckOrigin function validates against `allowedOrigins` slice
- Empty Origin header allowed (same-origin requests)
- Unauthorized origins logged and rejected
- Authorized origins logged and accepted

---

### Task 4: Add development wildcard support (optional enhancement)
**Action:** Support localhost wildcard patterns for easier development
**Files:**
- `backend/ws/hub.go` - Enhance origin matching

**Steps:**
1. Add helper function for pattern matching (before `init()` function):
   ```go
   // isOriginAllowed checks if origin matches any allowed pattern
   func isOriginAllowed(origin string, allowedOrigins []string) bool {
       for _, allowed := range allowedOrigins {
           // Exact match
           if origin == allowed {
               return true
           }

           // Wildcard localhost pattern (e.g., "http://localhost:*")
           if strings.HasSuffix(allowed, ":*") {
               prefix := strings.TrimSuffix(allowed, ":*")
               if strings.HasPrefix(origin, prefix+":") {
                   return true
               }
           }
       }
       return false
   }
   ```

2. Update CheckOrigin to use new helper:
   ```go
   CheckOrigin: func(r *http.Request) bool {
       origin := r.Header.Get("Origin")

       if origin == "" {
           return true
       }

       if isOriginAllowed(origin, allowedOrigins) {
           return true
       }

       log.Printf("[WebSocket] REJECTED connection from unauthorized origin: %s (client: %s)", origin, r.RemoteAddr)
       return false
   },
   ```

3. Update `.env.example` to document wildcard support:
   ```
   # WebSocket CORS (comma-separated origins)
   # Supports exact matches and localhost wildcards
   # Example: http://localhost:* matches http://localhost:5173, http://localhost:3000, etc.
   ALLOWED_ORIGINS=http://localhost:*,https://app.example.com
   ```

**Verification:**
- Wildcard pattern `http://localhost:*` matches `http://localhost:5173`
- Wildcard pattern `http://localhost:*` matches `http://localhost:3000`
- Non-localhost origins require exact match
- Invalid patterns safely rejected

---

### Task 5: Test CORS validation with valid origin
**Action:** Verify authorized origins can connect successfully
**Files:**
- N/A (manual testing)

**Steps:**
1. Ensure `.env` has development origins:
   ```
   ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
   ```

2. Start backend server:
   ```bash
   cd backend && go run cmd/server/main.go
   ```

3. Check startup logs show allowed origins:
   ```
   [WebSocket] CORS allowed origins: [http://localhost:5173 http://localhost:3000]
   ```

4. Test WebSocket connection from allowed origin (using browser console or wscat):
   ```javascript
   // In browser console at http://localhost:5173
   const ws = new WebSocket('ws://localhost:8080/ws');
   ws.onopen = () => console.log('Connected!');
   ws.onerror = (err) => console.error('Error:', err);
   ```

5. Check backend logs show accepted connection:
   ```
   [WS] Connection request from 127.0.0.1:xxxxx (origin: http://localhost:5173)
   [WS] Upgrade SUCCESS for 127.0.0.1:xxxxx
   [Hub] Client connected. Total clients: 1
   ```

**Verification:**
- WebSocket connection succeeds from http://localhost:5173
- WebSocket connection succeeds from http://localhost:3000
- Backend logs show origin and successful upgrade
- No REJECTED messages in logs for valid origins

---

### Task 6: Test CORS validation with invalid origin
**Action:** Verify unauthorized origins are rejected
**Files:**
- N/A (manual testing)

**Steps:**
1. Attempt WebSocket connection from unauthorized origin using curl with custom Origin header:
   ```bash
   curl -i -N \
     -H "Connection: Upgrade" \
     -H "Upgrade: websocket" \
     -H "Origin: http://evil.com" \
     -H "Sec-WebSocket-Version: 13" \
     -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
     http://localhost:8080/ws
   ```

2. Expected response:
   ```
   HTTP/1.1 403 Forbidden
   ```

3. Check backend logs show rejection:
   ```
   [WebSocket] REJECTED connection from unauthorized origin: http://evil.com (client: 127.0.0.1:xxxxx)
   ```

4. Test with another invalid origin:
   ```bash
   # Use wscat with custom origin (install: npm install -g wscat)
   wscat -c ws://localhost:8080/ws --origin http://attacker.com
   ```

5. Confirm 403 Forbidden response and rejection logged

**Verification:**
- Unauthorized origins receive 403 Forbidden
- REJECTED log message appears for each invalid origin
- Connection does not upgrade to WebSocket
- Server remains stable after rejection

---

### Task 7: Test empty ALLOWED_ORIGINS behavior
**Action:** Verify fail-safe behavior when ALLOWED_ORIGINS not configured
**Files:**
- N/A (negative test)

**Steps:**
1. Temporarily remove ALLOWED_ORIGINS from `.env`:
   ```bash
   cp .env .env.backup
   grep -v "ALLOWED_ORIGINS" .env.backup > .env
   ```

2. Restart backend:
   ```bash
   cd backend && go run cmd/server/main.go
   ```

3. Check startup logs show warning:
   ```
   [WARN] ALLOWED_ORIGINS not set - WebSocket will reject all connections
   [WARN] Set ALLOWED_ORIGINS=http://localhost:5173,... in .env for development
   ```

4. Attempt connection from any origin (should be rejected):
   ```javascript
   // In browser console
   const ws = new WebSocket('ws://localhost:8080/ws');
   // Should fail with 403
   ```

5. Restore `.env`:
   ```bash
   mv .env.backup .env
   ```

**Verification:**
- Warning logged when ALLOWED_ORIGINS not set
- All WebSocket connections rejected (empty whitelist = deny all)
- Server doesn't crash, fails safely
- Clear guidance provided in warning message

## Success Criteria

**Must be TRUE after completion:**

1. ✓ WebSocket connections validate origin against whitelist
   - `backend/ws/hub.go` CheckOrigin function checks `allowedOrigins` slice
   - No more `return true` (allow all origins)
   - Valid origins accepted, invalid origins rejected

2. ✓ CORS validation configured via environment variable
   - `ALLOWED_ORIGINS` environment variable parsed on startup
   - Comma-separated list of allowed origins
   - Documented in `.env.example`

3. ✓ Invalid origins rejected with clear error logging
   - REJECTED log message includes origin and client IP
   - 403 Forbidden response returned
   - Connection does not upgrade to WebSocket

4. ✓ Localhost origins allowed for development
   - Development `.env` includes `http://localhost:5173` and other dev ports
   - Optional wildcard support for localhost patterns
   - Production configuration uses only production domain

5. ✓ Fail-safe behavior when misconfigured
   - Missing ALLOWED_ORIGINS logs warning and rejects all connections
   - Empty ALLOWED_ORIGINS rejects all connections
   - Server doesn't crash, handles gracefully

## Verification

**Automated checks:**
```bash
# 1. CheckOrigin no longer allows all origins
! grep -A 3 "CheckOrigin.*func" backend/ws/hub.go | grep -q "return true$"

# 2. ALLOWED_ORIGINS documented
grep -q "ALLOWED_ORIGINS" .env.example

# 3. allowedOrigins variable exists
grep -q "var allowedOrigins" backend/ws/hub.go

# 4. init() function parses origins
grep -q "func init()" backend/ws/hub.go
```

**Manual verification:**
1. Start server and confirm allowed origins logged
2. Connect from http://localhost:5173 - should succeed
3. Connect from http://evil.com - should fail with 403
4. Check logs show both accepted and rejected connections
5. Test with ALLOWED_ORIGINS unset - all connections rejected

## Output

**Modified files:**
- `backend/ws/hub.go` - Replaced insecure CheckOrigin with whitelist validation
- `.env.example` - Added ALLOWED_ORIGINS documentation
- `.env` - Added ALLOWED_ORIGINS with development values

**New code:**
- `allowedOrigins` global variable
- `init()` function for parsing ALLOWED_ORIGINS
- `isOriginAllowed()` helper function (optional wildcard support)
- Secure CheckOrigin implementation

**Security improvements:**
- WebSocket CORS validation prevents unauthorized cross-origin access
- Clear logging for security audit trail
- Fail-safe behavior (deny by default)
- Environment-based configuration for dev/prod separation
