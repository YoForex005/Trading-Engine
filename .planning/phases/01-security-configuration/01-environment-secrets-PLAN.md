---
must_haves:
  truths:
    - No hardcoded OANDA API credentials exist in codebase
    - JWT secret loaded from environment variable (not hardcoded fallback)
    - Platform starts successfully using .env configuration
    - .env.example template documents all required environment variables
  artifacts:
    - .env.example file in project root with all required variables
    - Updated backend/cmd/server/main.go with OANDA credentials from env
    - Updated backend/auth/token.go with JWT secret from env
    - .gitignore includes .env to prevent accidental commits
  key_links:
    - backend/cmd/server/main.go:23-24 (OANDA credentials)
    - backend/auth/token.go:10-18 (JWT secret)
wave: 1
---

# Plan: Environment Configuration & Secret Management

## Objective

Eliminate hardcoded credentials and implement production-grade environment variable configuration system. This plan rotates OANDA API credentials and JWT secret to environment variables, creates .env infrastructure, and ensures platform starts with secure configuration.

## Execution Context

**Requirements addressed:**
- SECURITY-01: Rotate hardcoded OANDA API credentials to environment variables
- SECURITY-02: Replace weak JWT secret with cryptographically random key
- SECURITY-05: Implement environment variable configuration system (.env support)

**Reference files:**
- backend/cmd/server/main.go (lines 23-24: hardcoded OANDA credentials)
- backend/auth/token.go (lines 10-18: weak JWT secret fallback)
- backend/lpmanager/adapters/oanda.go (likely constructor signature)

**Codebase context:**
- Go 1.24.0 backend
- No existing .env infrastructure
- Credentials currently as constants
- No .env library detected (need to add)

## Context

**Current Security Issues:**

1. **Hardcoded OANDA Credentials** (main.go:23-24):
   ```go
   const OANDA_API_KEY = "977e1a77e25bac3a688011d6b0e845dd-8e3ab3a7682d9351af4c33be65e89b70"
   const OANDA_ACCOUNT_ID = "101-004-37008470-002"
   ```
   Used at line 106: `adapters.NewOANDAAdapter(OANDA_API_KEY, OANDA_ACCOUNT_ID)`

2. **Weak JWT Secret** (token.go:10-18):
   ```go
   var jwtKey = []byte(os.Getenv("JWT_SECRET"))

   func init() {
       if len(jwtKey) == 0 {
           jwtKey = []byte("super_secret_dev_key_do_not_use_in_prod")
       }
   }
   ```

**Required Environment Variables:**
- `OANDA_API_KEY` - OANDA API authentication key
- `OANDA_ACCOUNT_ID` - OANDA account identifier
- `JWT_SECRET` - Cryptographically secure random key (minimum 32 bytes)
- `PORT` - Server port (default 8080)
- `ENVIRONMENT` - Runtime environment (development, staging, production)

**Go .env Library Options:**
- `github.com/joho/godotenv` - Most popular, simple API
- `github.com/kelseyhightower/envconfig` - Struct-based config
- Decision: Use godotenv (simple, well-established)

## Tasks

### Task 1: Add .env infrastructure
**Action:** Install godotenv library and create .env.example template
**Files:**
- `backend/go.mod` - Add godotenv dependency
- `.env.example` - Create template with all required variables
- `.gitignore` - Add `.env` to prevent credential commits

**Steps:**
1. Add godotenv to backend:
   ```bash
   cd backend && go get github.com/joho/godotenv
   ```
2. Create `.env.example` in project root with template:
   ```
   # Trading Engine Configuration

   # Server Configuration
   PORT=8080
   ENVIRONMENT=development

   # JWT Authentication
   JWT_SECRET=<generate_with_openssl_rand_base64_32>

   # OANDA Liquidity Provider
   OANDA_API_KEY=your_oanda_api_key_here
   OANDA_ACCOUNT_ID=your_oanda_account_id_here

   # WebSocket CORS (comma-separated origins)
   ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
   ```
3. Update `.gitignore` to include `.env` if not already present
4. Generate secure JWT_SECRET for development:
   ```bash
   openssl rand -base64 32
   ```

**Verification:**
- `backend/go.mod` contains `github.com/joho/godotenv` dependency
- `.env.example` exists in project root with all variables documented
- `.gitignore` contains `.env` entry
- Running `go mod tidy` in backend succeeds

---

### Task 2: Load environment variables in main.go
**Action:** Initialize godotenv at server startup before any configuration loading
**Files:**
- `backend/cmd/server/main.go` - Add .env loading in init() or early in main()

**Steps:**
1. Import godotenv at top of main.go:
   ```go
   import (
       // ... existing imports
       "github.com/joho/godotenv"
   )
   ```
2. Add .env loading at very start of `main()` function (before line 52):
   ```go
   func main() {
       // Load .env file (ignore error in production where env vars are set directly)
       _ = godotenv.Load()

       // Existing code continues...
       log.Println("╔═══════════════════════════════════════════════════════════╗")
   ```
3. Add helpful error logging for missing critical env vars later in code

**Verification:**
- `godotenv.Load()` called before any `os.Getenv()` usage
- Server starts without error when .env file exists
- Server starts without error when .env file missing (env vars set directly)

---

### Task 3: Rotate OANDA credentials to environment variables
**Action:** Replace hardcoded OANDA constants with env var loading
**Files:**
- `backend/cmd/server/main.go` - Remove constants, load from env

**Steps:**
1. **Delete lines 22-24** (hardcoded constants):
   ```go
   // REMOVE THESE LINES:
   // LP API Keys - In production, use environment variables
   const OANDA_API_KEY = "977e1a77e25bac3a688011d6b0e845dd-8e3ab3a7682d9351af4c33be65e89b70"
   const OANDA_ACCOUNT_ID = "101-004-37008470-002"
   ```

2. **Add env var loading** after godotenv.Load() in main():
   ```go
   func main() {
       _ = godotenv.Load()

       // Load OANDA credentials from environment
       oandaAPIKey := os.Getenv("OANDA_API_KEY")
       oandaAccountID := os.Getenv("OANDA_ACCOUNT_ID")

       // Validate critical credentials
       if oandaAPIKey == "" {
           log.Println("[WARN] OANDA_API_KEY not set - OANDA adapter will fail to connect")
       }
       if oandaAccountID == "" {
           log.Println("[WARN] OANDA_ACCOUNT_ID not set - OANDA adapter will fail to connect")
       }

       // Continue with existing startup...
   ```

3. **Update adapter registration** (around line 106):
   ```go
   // Replace:
   // lpMgr.RegisterAdapter(adapters.NewOANDAAdapter(OANDA_API_KEY, OANDA_ACCOUNT_ID))

   // With:
   if oandaAPIKey != "" && oandaAccountID != "" {
       lpMgr.RegisterAdapter(adapters.NewOANDAAdapter(oandaAPIKey, oandaAccountID))
       log.Println("[LP Manager] OANDA adapter registered")
   } else {
       log.Println("[LP Manager] OANDA adapter skipped (credentials not configured)")
   }
   ```

**Verification:**
- Search codebase for string `"977e1a77e25bac3a688011d6b0e845dd"` returns 0 results
- Search codebase for string `"101-004-37008470-002"` returns 0 results
- No `const OANDA_API_KEY` or `const OANDA_ACCOUNT_ID` definitions exist
- Server starts with env vars set
- Warning logged when env vars missing

---

### Task 4: Enforce secure JWT secret loading
**Action:** Remove weak fallback, require JWT_SECRET environment variable
**Files:**
- `backend/auth/token.go` - Update initialization logic

**Steps:**
1. Update `init()` function in token.go (lines 12-20):
   ```go
   func init() {
       if len(jwtKey) == 0 {
           log.Fatal("[CRITICAL] JWT_SECRET environment variable not set. Generate with: openssl rand -base64 32")
       }
       if len(jwtKey) < 32 {
           log.Fatal("[CRITICAL] JWT_SECRET too short (minimum 32 bytes required)")
       }
       log.Printf("[Auth] JWT secret loaded (%d bytes)", len(jwtKey))
   }
   ```

2. Remove the weak fallback entirely (delete line 18):
   ```go
   // DELETE THIS LINE:
   // jwtKey = []byte("super_secret_dev_key_do_not_use_in_prod")
   ```

3. Add a comment explaining the security requirement:
   ```go
   var jwtKey = []byte(os.Getenv("JWT_SECRET"))

   // init validates JWT_SECRET is set and cryptographically secure
   func init() {
       // ... validation code from step 1
   }
   ```

**Verification:**
- Search codebase for string `"super_secret_dev_key_do_not_use_in_prod"` returns 0 results
- Server panics on startup if `JWT_SECRET` not set (expected behavior)
- Server panics if `JWT_SECRET` shorter than 32 bytes
- Server starts successfully with valid `JWT_SECRET` set
- Log message confirms JWT secret loaded with byte count

---

### Task 5: Create local .env file for development
**Action:** Create actual .env file with secure values for local testing
**Files:**
- `.env` - Create in project root (NOT committed to git)

**Steps:**
1. Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```

2. Generate secure JWT secret:
   ```bash
   openssl rand -base64 32
   ```

3. Fill in `.env` with actual values:
   ```
   PORT=8080
   ENVIRONMENT=development
   JWT_SECRET=<paste_generated_secret_here>
   OANDA_API_KEY=977e1a77e25bac3a688011d6b0e845dd-8e3ab3a7682d9351af4c33be65e89b70
   OANDA_ACCOUNT_ID=101-004-37008470-002
   ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
   ```

4. Verify `.env` is gitignored:
   ```bash
   git status  # Should NOT show .env file
   ```

**Verification:**
- `.env` file exists in project root
- `.env` contains all required variables with valid values
- `git status` does not list `.env` as untracked file
- JWT_SECRET is at least 32 bytes when base64-decoded

---

### Task 6: Test server startup with environment configuration
**Action:** Verify platform starts successfully with .env-based configuration
**Files:**
- N/A (integration test)

**Steps:**
1. Stop any running backend server
2. Start backend with .env configuration:
   ```bash
   cd backend && go run cmd/server/main.go
   ```
3. Check startup logs for:
   - ✓ godotenv loaded (no error)
   - ✓ JWT secret loaded with byte count
   - ✓ OANDA adapter registered (or warning if skipped)
   - ✓ Server starts on configured PORT
   - ✓ No hardcoded credential warnings

4. Test basic API endpoint to confirm server functional:
   ```bash
   curl http://localhost:8080/health
   # Should return: OK
   ```

5. Test with missing JWT_SECRET (negative test):
   ```bash
   # Temporarily rename .env
   mv .env .env.backup
   go run cmd/server/main.go
   # Should FAIL with JWT_SECRET error
   mv .env.backup .env
   ```

**Verification:**
- Server starts without errors with valid .env
- Server fails fast with clear error when JWT_SECRET missing
- Health endpoint responds correctly
- No hardcoded credentials visible in logs
- OANDA adapter initializes with env credentials

## Success Criteria

**Must be TRUE after completion:**

1. ✓ No hardcoded credentials exist in codebase
   - Search for `"977e1a77e25bac3a688011d6b0e845dd"` returns 0 results
   - Search for `"101-004-37008470-002"` returns 0 results
   - Search for `"super_secret_dev_key_do_not_use_in_prod"` returns 0 results

2. ✓ JWT secret loaded from environment variable
   - `backend/auth/token.go` reads from `os.Getenv("JWT_SECRET")`
   - No weak fallback exists
   - Server fails fast if JWT_SECRET not set or too short

3. ✓ OANDA credentials loaded from environment variables
   - `backend/cmd/server/main.go` reads from `OANDA_API_KEY` and `OANDA_ACCOUNT_ID`
   - No hardcoded constants remain
   - Graceful handling when credentials not configured

4. ✓ Platform starts successfully using .env configuration
   - `.env` file in project root
   - `godotenv` loads environment variables on startup
   - All services initialize with env-based config

5. ✓ .env.example template exists
   - Documents all required environment variables
   - Includes comments explaining each variable
   - Committed to git as reference

6. ✓ .env excluded from git
   - `.gitignore` contains `.env` entry
   - `git status` never shows `.env` file

## Verification

**Automated checks:**
```bash
# 1. No hardcoded secrets in codebase
! grep -r "977e1a77e25bac3a688011d6b0e845dd" backend/
! grep -r "101-004-37008470-002" backend/
! grep -r "super_secret_dev_key_do_not_use_in_prod" backend/

# 2. Environment configuration files exist
test -f .env.example
test -f .env
grep -q "^.env$" .gitignore

# 3. godotenv dependency added
grep -q "github.com/joho/godotenv" backend/go.mod

# 4. Server starts with .env
cd backend && timeout 5s go run cmd/server/main.go > /tmp/startup.log 2>&1 || true
grep -q "JWT secret loaded" /tmp/startup.log
! grep -q "super_secret_dev_key" /tmp/startup.log
```

**Manual verification:**
1. Review `backend/cmd/server/main.go` - no credential constants
2. Review `backend/auth/token.go` - no weak fallback
3. Review `.env.example` - all variables documented
4. Start server and confirm startup logs show env-based configuration
5. Confirm `/health` endpoint responds

## Output

**Modified files:**
- `backend/go.mod` - Added godotenv dependency
- `backend/cmd/server/main.go` - Removed hardcoded OANDA credentials, added env loading
- `backend/auth/token.go` - Removed weak JWT fallback, enforce env var
- `.env.example` - Created with all required variables documented
- `.gitignore` - Added `.env` exclusion
- `.env` - Created for local development (not committed)

**Deleted code:**
- Hardcoded OANDA_API_KEY constant
- Hardcoded OANDA_ACCOUNT_ID constant
- Weak JWT_SECRET fallback

**New capabilities:**
- Environment-based configuration system
- Secure credential management
- Development/production configuration separation
- Fail-fast validation for missing secrets
