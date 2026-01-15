---
must_haves:
  truths:
    - All passwords stored as bcrypt hashes (no plaintext fallback)
    - Plaintext password comparison code removed from codebase
    - Existing demo account password upgraded to bcrypt hash
    - Login only succeeds with valid bcrypt hash comparison
  artifacts:
    - Updated backend/auth/service.go with plaintext fallback removed
    - Updated backend/internal/core/engine.go CreateAccount to hash passwords
    - Test confirming plaintext passwords no longer work
  key_links:
    - backend/auth/service.go:81-92 (plaintext password fallback)
    - backend/internal/core/engine.go (CreateAccount function)
wave: 1
---

# Plan: Password Security Hardening

## Objective

Enforce bcrypt-only password authentication by removing the insecure plaintext password fallback. This plan eliminates the code path that allows plaintext password comparison, ensuring all passwords are stored and verified using cryptographically secure bcrypt hashes.

## Execution Context

**Requirements addressed:**
- SECURITY-04: Remove plaintext password fallback, enforce bcrypt

**Reference files:**
- backend/auth/service.go (lines 81-92: plaintext password fallback)
- backend/internal/core/engine.go (CreateAccount function)
- backend/cmd/server/main.go:76 (demo account creation with plaintext "password")

**Codebase context:**
- Go 1.24.0 backend with golang.org/x/crypto/bcrypt
- Authentication service in backend/auth/service.go
- Account creation in backend/internal/core/engine.go
- Demo account created on startup in main.go

## Context

**Current Security Issue:**

The `Login()` function in `backend/auth/service.go` lines 81-92 contains a dangerous plaintext password fallback:

```go
// 3. Verify Password
// We treat account.Password as a Hash.
err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
if err != nil {
    // Fallback: Check if it's plaintext "password" (Legacy Dev Mode)
    // ONLY if it doesn't look like a hash (bcrypt starts with $2)
    // And check if password is correct plaintext.
    if len(account.Password) > 0 && account.Password[0] != '$' && account.Password == password {
        log.Printf("[WARN] Auto-upgrading password for user %s to bcrypt", account.Username)
        newHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        s.engine.UpdatePassword(account.ID, string(newHash))
        // Proceed
    } else {
        log.Printf("[WARN] Login failed for user %s (invalid password)", username)
        return "", nil, errors.New("invalid credentials")
    }
}
```

**Why this is dangerous:**
1. Allows plaintext password storage (defeats bcrypt protection)
2. Enables timing attacks (different code paths for plaintext vs hash)
3. "Legacy Dev Mode" has no expiration - permanent security hole
4. Auto-upgrade masks the problem instead of fixing it
5. Violates security principle: fail early on misconfiguration

**Additional Issue:**

Demo account created in `backend/cmd/server/main.go` line 76:
```go
demoAccount := bbookEngine.CreateAccount("demo-user", "Demo User", "password", true)
```

The `CreateAccount` function likely accepts plaintext password without hashing it first.

**Required Fix:**

1. Remove plaintext password fallback from `Login()` function
2. Ensure `CreateAccount()` always hashes passwords with bcrypt
3. Update demo account creation to use pre-hashed password or ensure CreateAccount hashes
4. Verify all passwords stored as bcrypt hashes

## Tasks

### Task 1: Examine CreateAccount password handling
**Action:** Read engine.go to understand how CreateAccount handles passwords
**Files:**
- `backend/internal/core/engine.go` - Locate CreateAccount function

**Steps:**
1. Read `backend/internal/core/engine.go` to find `CreateAccount` function
2. Identify if it hashes passwords before storing
3. Determine if modification needed to enforce bcrypt hashing

**Verification:**
- CreateAccount function signature understood
- Password hashing behavior documented

---

### Task 2: Update CreateAccount to enforce bcrypt hashing
**Action:** Ensure CreateAccount always hashes passwords before storing
**Files:**
- `backend/internal/core/engine.go` - Update CreateAccount function

**Steps:**
1. Locate `CreateAccount` function in engine.go
2. Add bcrypt import if not present:
   ```go
   import (
       // ... existing imports
       "golang.org/x/crypto/bcrypt"
   )
   ```

3. Update CreateAccount to hash password parameter:
   ```go
   func (e *Engine) CreateAccount(username, name, password string, isDemo bool) *Account {
       e.mu.Lock()
       defer e.mu.Unlock()

       // Hash password with bcrypt before storing
       passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
       if err != nil {
           log.Printf("[ERROR] Failed to hash password for user %s: %v", username, err)
           return nil
       }

       // ... rest of function using string(passwordHash) instead of password
   ```

4. Update the Account struct field assignment:
   ```go
   account := &Account{
       // ... other fields
       Password: string(passwordHash),  // Store hash, not plaintext
       // ... other fields
   }
   ```

5. Add logging to confirm hashing:
   ```go
   log.Printf("[Account] Created account %s with bcrypt-hashed password", username)
   ```

**Verification:**
- CreateAccount function hashes password with bcrypt.GenerateFromPassword
- Password parameter never stored directly (always hashed)
- Error handling for hash generation failure
- Log message confirms bcrypt hashing

---

### Task 3: Remove plaintext password fallback from Login
**Action:** Delete insecure plaintext password comparison code
**Files:**
- `backend/auth/service.go` - Remove lines 81-92 fallback

**Steps:**
1. Locate the `Login()` function in service.go (around line 40)

2. Find the bcrypt comparison block (around line 79-92)

3. **Delete the entire fallback block** (lines 81-92):
   ```go
   // DELETE THIS ENTIRE BLOCK:
   /*
   if err != nil {
       // Fallback: Check if it's plaintext "password" (Legacy Dev Mode)
       // ONLY if it doesn't look like a hash (bcrypt starts with $2)
       // And check if password is correct plaintext.
       if len(account.Password) > 0 && account.Password[0] != '$' && account.Password == password {
           log.Printf("[WARN] Auto-upgrading password for user %s to bcrypt", account.Username)
           newHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
           s.engine.UpdatePassword(account.ID, string(newHash))
           // Proceed
       } else {
           log.Printf("[WARN] Login failed for user %s (invalid password)", username)
           return "", nil, errors.New("invalid credentials")
       }
   }
   */
   ```

4. **Replace with clean bcrypt-only validation:**
   ```go
   // 3. Verify Password (bcrypt only - no plaintext fallback)
   err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
   if err != nil {
       log.Printf("[WARN] Login failed for user %s (invalid password)", username)
       return "", nil, errors.New("invalid credentials")
   }

   // Password valid - continue to JWT generation
   ```

5. Simplify the code flow - password validation is now one clean block with no fallback logic

**Verification:**
- Lines 81-92 plaintext fallback code deleted
- Only bcrypt.CompareHashAndPassword used for password verification
- Login fails immediately on bcrypt comparison failure
- No auto-upgrade logic remains

---

### Task 4: Update demo account creation to use hashed password
**Action:** Ensure demo account created with bcrypt hash (via CreateAccount fix)
**Files:**
- `backend/cmd/server/main.go` - Verify demo account creation

**Steps:**
1. Locate demo account creation in main.go (around line 76):
   ```go
   demoAccount := bbookEngine.CreateAccount("demo-user", "Demo User", "password", true)
   ```

2. Verify this calls the updated CreateAccount function (which now hashes passwords)

3. No changes needed to main.go if CreateAccount was updated in Task 2

4. **Optional:** Add comment explaining password is hashed internally:
   ```go
   // Create demo account with password "password" (hashed internally by CreateAccount)
   demoAccount := bbookEngine.CreateAccount("demo-user", "Demo User", "password", true)
   ```

5. After changes, demo account will have bcrypt hash stored automatically

**Verification:**
- Demo account creation unchanged in main.go (CreateAccount does hashing)
- Comment added explaining password handling
- No plaintext password storage in demo account

---

### Task 5: Test login with bcrypt-hashed password
**Action:** Verify authentication works with bcrypt hashes only
**Files:**
- N/A (integration test)

**Steps:**
1. Clean any existing account data (if using file-based storage):
   ```bash
   # Backup and clean existing accounts if needed
   # (depends on storage implementation)
   ```

2. Start backend server:
   ```bash
   cd backend && go run cmd/server/main.go
   ```

3. Check logs confirm demo account created with bcrypt hash:
   ```
   [Account] Created account demo-user with bcrypt-hashed password
   [B-Book] Demo account created: DEMO-XXX | Balance: $5000.00
   ```

4. Test login with correct password:
   ```bash
   curl -X POST http://localhost:8080/login \
     -H "Content-Type: application/json" \
     -d '{"username":"demo-user","password":"password"}'
   ```

5. Expected response: JWT token returned
   ```json
   {
     "token": "eyJhbGciOi...",
     "user": {
       "id": "...",
       "username": "demo-user",
       "role": "TRADER"
     }
   }
   ```

6. Check logs show successful login:
   ```
   [INFO] Login successful for user demo-user
   ```

**Verification:**
- Login succeeds with correct password
- JWT token returned
- No auto-upgrade warnings in logs
- Bcrypt comparison works correctly

---

### Task 6: Test login rejection with plaintext password (negative test)
**Action:** Verify plaintext passwords are rejected (no fallback)
**Files:**
- N/A (negative test)

**Steps:**
1. Manually create test account with intentional plaintext password (simulate old data):
   ```bash
   # This requires temporarily modifying CreateAccount or direct data manipulation
   # Skip this test if CreateAccount now enforces hashing (desired behavior)
   # Instead, test with wrong password to confirm rejection works
   ```

2. Test login with wrong password:
   ```bash
   curl -X POST http://localhost:8080/login \
     -H "Content-Type: application/json" \
     -d '{"username":"demo-user","password":"wrongpassword"}'
   ```

3. Expected response: 401 Unauthorized
   ```json
   {
     "error": "invalid credentials"
   }
   ```

4. Check logs show login failure:
   ```
   [WARN] Login failed for user demo-user (invalid password)
   ```

5. Confirm NO auto-upgrade messages appear (plaintext fallback removed)

**Verification:**
- Wrong password rejected immediately
- No plaintext fallback triggered
- No auto-upgrade warnings
- Error message does not reveal if username exists (security)

---

### Task 7: Code audit for remaining plaintext password risks
**Action:** Search codebase for any remaining plaintext password storage or comparison
**Files:**
- All backend files (grep search)

**Steps:**
1. Search for potential plaintext password comparisons:
   ```bash
   cd backend
   # Search for direct password string comparisons (potential risk)
   grep -r "\.Password\s*==" . --include="*.go"
   grep -r "password\s*==" . --include="*.go" | grep -v "test"
   ```

2. Search for password storage without bcrypt:
   ```bash
   # Search for Password field assignments not using bcrypt
   grep -r "\.Password\s*=" . --include="*.go" | grep -v bcrypt
   ```

3. Review any findings:
   - If bcrypt used → OK
   - If plaintext → flag for fixing

4. Document findings in this task's output

5. Fix any additional issues found (if any)

**Verification:**
- No direct password comparisons using `==` operator
- All Password field assignments use bcrypt hash
- No plaintext password storage found
- Codebase audit clean

## Success Criteria

**Must be TRUE after completion:**

1. ✓ All passwords stored as bcrypt hashes
   - `CreateAccount()` function hashes passwords with bcrypt before storing
   - No code path stores plaintext passwords
   - Demo account has bcrypt-hashed password

2. ✓ Plaintext password comparison code removed
   - Lines 81-92 fallback deleted from `backend/auth/service.go`
   - Only `bcrypt.CompareHashAndPassword()` used for verification
   - No `account.Password == password` comparisons exist

3. ✓ Login only succeeds with valid bcrypt hash
   - Correct password authenticates successfully
   - Wrong password rejected immediately
   - No auto-upgrade or fallback behavior

4. ✓ No plaintext password risks remain in codebase
   - Code audit finds no plaintext password storage
   - Code audit finds no plaintext password comparison
   - All password handling uses bcrypt

## Verification

**Automated checks:**
```bash
# 1. No plaintext password fallback in Login
! grep -A 10 "bcrypt.CompareHashAndPassword" backend/auth/service.go | grep -q "account.Password == password"

# 2. No plaintext password comparisons anywhere
! grep -r "\.Password\s*==" backend/ --include="*.go" | grep -v test

# 3. CreateAccount uses bcrypt
grep -A 20 "func.*CreateAccount" backend/internal/core/engine.go | grep -q "bcrypt.GenerateFromPassword"

# 4. No "Auto-upgrading password" log messages
! grep -r "Auto-upgrading password" backend/ --include="*.go"
```

**Manual verification:**
1. Review `backend/auth/service.go` Login() - no plaintext fallback
2. Review `backend/internal/core/engine.go` CreateAccount() - uses bcrypt
3. Start server and test login with correct password - succeeds
4. Test login with wrong password - fails immediately
5. Check logs - no auto-upgrade warnings

## Output

**Modified files:**
- `backend/internal/core/engine.go` - Updated CreateAccount to hash passwords with bcrypt
- `backend/auth/service.go` - Removed plaintext password fallback (lines 81-92 deleted)
- `backend/cmd/server/main.go` - Added comment explaining password hashing (optional)

**Deleted code:**
- Plaintext password fallback logic (service.go lines 81-92)
- Auto-upgrade password mechanism
- Legacy dev mode password handling

**Security improvements:**
- Enforce bcrypt-only password authentication
- Eliminate plaintext password attack vector
- Remove timing attack risk from dual code paths
- Consistent password security across all accounts
