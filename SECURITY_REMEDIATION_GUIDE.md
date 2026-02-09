# Security Remediation Guide - Hardcoded Credentials

## CRITICAL ISSUE: Exposed Credentials in Source Code

**Severity**: 🔴 CRITICAL
**Location**: `backend/fix/gateway.go` lines 267-314
**Risk Level**: Active production accounts exposed

---

## Exposed Credentials

| Item | Current Value | Status |
|------|---------------|--------|
| YOFX1 Password | `Brand#143` | ⚠️ IN USE |
| YOFX2 Password | `Brand#143` | ⚠️ IN USE |
| YOFX Trading Account | `50153` | ⚠️ REAL ACCOUNT |
| Proxy Host | `81.29.145.69` | ⚠️ IN USE |
| Proxy Port | `49527` | ⚠️ IN USE |
| Proxy Username | `fGUqTcsdMsBZlms` | ⚠️ IN USE |
| Proxy Password | `3eo1qF91WA7Fyku` | ⚠️ IN USE |

---

## Immediate Remediation Steps (DO THIS NOW)

### Step 1: Rotate All Exposed Credentials

#### YoForex Credentials
1. Login to YoForex admin panel (account 50153)
2. Change YOFX1 password (replace `Brand#143`)
3. Change YOFX2 password (replace `Brand#143`)
4. Verify both sessions connect successfully with new passwords

#### Proxy Credentials
1. Contact proxy provider or reset proxy credentials
2. Generate new proxy username and password
3. Ensure new credentials work with YOFX connections

**Why**: Anyone with repository access has these credentials

---

### Step 2: Remove from Git History

**WARNING**: This is destructive and affects all users of the repository.

#### Option A: Using git-filter-repo (RECOMMENDED)

```bash
# Install git-filter-repo
pip install git-filter-repo

# Clone a fresh copy for history rewriting
cd /tmp
git clone --mirror C:\Users\s\ s\ laptop\ bazar\Trading-Engine2 trading-engine-mirror.git

# Filter the history
cd trading-engine-mirror.git
git filter-repo --replace-text ..//paths.txt
```

Create `paths.txt` with patterns to remove:
```
Brand#143
fGUqTcsdMsBZlms
3eo1qF91WA7Fyku
```

#### Option B: Using git reset (DANGEROUS)

```bash
# WARNING: This removes all history!
# Only use if repository is not shared

cd C:\Users\s\ s\ laptop\ bazar\Trading-Engine2

# Save current work
git stash

# Create new repository
git init
git add .
git commit -m "Initial commit - credentials removed"

# Force push (requires clearing remote)
git remote add origin <your-repo>
git push -f origin main
```

#### Option C: Using BFG Repo-Cleaner

```bash
# Download and use BFG
bfg --delete-files Brand#143 --replace-text ..//replacements.txt

# Then force push
git push -f origin main
```

---

### Step 3: Update Source Code

#### Current (INSECURE):
```go
// backend/fix/gateway.go:267-314
"YOFX1": {
    ID:              "YOFX1",
    Host:            getEnvOrDefault("YOFX_HOST", "23.106.238.138"),
    Port:            getEnvIntOrDefault("YOFX_PORT", 12336),
    SenderCompID:    getEnvOrDefault("YOFX1_SENDER_COMP_ID", "YOFX1"),
    TargetCompID:    getEnvOrDefault("YOFX_TARGET_COMP_ID", "YOFX"),
    Username:        getEnvOrDefault("YOFX1_USERNAME", "YOFX1"),
    Password:        getEnvOrDefault("YOFX1_PASSWORD", "Brand#143"),        // ❌ HARDCODED
    TradingAccount:  getEnvOrDefault("YOFX_TRADING_ACCOUNT", "50153"),     // ❌ HARDCODED
    UseProxy:        getEnvOrDefault("YOFX_USE_PROXY", "true") == "true",
    ProxyHost:       getEnvOrDefault("YOFX_PROXY_HOST", "81.29.145.69"),   // ❌ HARDCODED
    ProxyPort:       getEnvIntOrDefault("YOFX_PROXY_PORT", 49527),         // ❌ HARDCODED
    ProxyUsername:   getEnvOrDefault("YOFX_PROXY_USERNAME", "fGUqTcsdMsBZlms"),    // ❌ HARDCODED
    ProxyPassword:   getEnvOrDefault("YOFX_PROXY_PASSWORD", "3eo1qF91WA7Fyku"),    // ❌ HARDCODED
},
```

#### Required (SECURE):
```go
// backend/fix/gateway.go:267-314
"YOFX1": {
    ID:              "YOFX1",
    Host:            getEnvRequired("YOFX_HOST"),                          // ✅ REQUIRED
    Port:            getEnvIntRequired("YOFX_PORT"),                       // ✅ REQUIRED
    SenderCompID:    getEnvRequired("YOFX1_SENDER_COMP_ID"),              // ✅ REQUIRED
    TargetCompID:    getEnvRequired("YOFX_TARGET_COMP_ID"),               // ✅ REQUIRED
    Username:        getEnvRequired("YOFX1_USERNAME"),                     // ✅ REQUIRED
    Password:        getEnvRequired("YOFX1_PASSWORD"),                     // ✅ REQUIRED (NO DEFAULT)
    TradingAccount:  getEnvRequired("YOFX_TRADING_ACCOUNT"),              // ✅ REQUIRED (NO DEFAULT)
    UseProxy:        getEnvAsBoolRequired("YOFX_USE_PROXY"),              // ✅ REQUIRED
    ProxyHost:       getEnvRequired("YOFX_PROXY_HOST"),                    // ✅ REQUIRED (NO DEFAULT)
    ProxyPort:       getEnvIntRequired("YOFX_PROXY_PORT"),                 // ✅ REQUIRED (NO DEFAULT)
    ProxyUsername:   getEnvRequired("YOFX_PROXY_USERNAME"),                // ✅ REQUIRED (NO DEFAULT)
    ProxyPassword:   getEnvRequired("YOFX_PROXY_PASSWORD"),                // ✅ REQUIRED (NO DEFAULT)
},
```

#### Add helper functions to `backend/fix/gateway.go`:

```go
// getEnvRequired returns the environment variable or fatally errors if not set
func getEnvRequired(key string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    log.Fatalf("FATAL: Required environment variable %s not set", key)
    return "" // Never reached
}

// getEnvIntRequired returns the environment variable as int or fatally errors
func getEnvIntRequired(key string) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
        log.Fatalf("FATAL: Environment variable %s is not a valid integer", key)
    }
    log.Fatalf("FATAL: Required environment variable %s not set", key)
    return 0
}

// getEnvAsBoolRequired returns the environment variable as bool or fatally errors
func getEnvAsBoolRequired(key string) bool {
    value := os.Getenv(key)
    if value == "" {
        log.Fatalf("FATAL: Required environment variable %s not set", key)
    }
    boolVal, err := strconv.ParseBool(value)
    if err != nil {
        log.Fatalf("FATAL: Environment variable %s is not a valid boolean", key)
    }
    return boolVal
}
```

---

### Step 4: Update .env Configuration

**Location**: `backend/.env` (create if doesn't exist)

```bash
# YOFX Configuration - REQUIRED
YOFX_HOST=23.106.238.138
YOFX_PORT=12336
YOFX1_SENDER_COMP_ID=YOFX1
YOFX2_SENDER_COMP_ID=YOFX2
YOFX_TARGET_COMP_ID=YOFX
YOFX1_USERNAME=YOFX1
YOFX1_PASSWORD=<YOUR_NEW_PASSWORD>              # NEW PASSWORD FROM STEP 1
YOFX2_USERNAME=YOFX2
YOFX2_PASSWORD=<YOUR_NEW_PASSWORD>              # NEW PASSWORD FROM STEP 1
YOFX_TRADING_ACCOUNT=50153
YOFX_USE_PROXY=true
YOFX_PROXY_HOST=<YOUR_PROXY_HOST>               # NEW FROM STEP 1
YOFX_PROXY_PORT=<YOUR_PROXY_PORT>               # NEW FROM STEP 1
YOFX_PROXY_USERNAME=<YOUR_PROXY_USERNAME>       # NEW FROM STEP 1
YOFX_PROXY_PASSWORD=<YOUR_PROXY_PASSWORD>       # NEW FROM STEP 1
```

**Location**: `backend/.env.example`

```bash
# YOFX Configuration - REQUIRED
# DO NOT commit real credentials - use .env file instead
YOFX_HOST=23.106.238.138
YOFX_PORT=12336
YOFX1_SENDER_COMP_ID=YOFX1
YOFX2_SENDER_COMP_ID=YOFX2
YOFX_TARGET_COMP_ID=YOFX
YOFX1_USERNAME=YOFX1
YOFX1_PASSWORD=<SET_YOUR_PASSWORD>
YOFX2_USERNAME=YOFX2
YOFX2_PASSWORD=<SET_YOUR_PASSWORD>
YOFX_TRADING_ACCOUNT=<SET_YOUR_ACCOUNT_NUMBER>
YOFX_USE_PROXY=true
YOFX_PROXY_HOST=<SET_YOUR_PROXY_HOST>
YOFX_PROXY_PORT=<SET_YOUR_PROXY_PORT>
YOFX_PROXY_USERNAME=<SET_YOUR_PROXY_USERNAME>
YOFX_PROXY_PASSWORD=<SET_YOUR_PROXY_PASSWORD>
```

---

### Step 5: Update .gitignore

**Location**: `backend/.gitignore` (add if not present)

```bash
# Never commit credentials
.env
.env.local
.env.production
.env.*.local

# Credentials and keys
*.key
*.pem
*.cert
credentials.json
secret*.txt

# FIX store sensitive data
fixstore/*
data/fix_credentials/*

# Sensitive configuration
config/*.secret
config/production.yml

# IDE secrets
.idea/aws.xml
.idea/dataSources.xml
.vscode/settings.json

# OS
.DS_Store
Thumbs.db
```

---

### Step 6: Verify Changes

```bash
# Check that no credentials are in the code
grep -r "Brand#143" backend/
grep -r "fGUqTcsdMsBZlms" backend/
grep -r "3eo1qF91WA7Fyku" backend/
grep -r "getEnvOrDefault.*PASSWORD" backend/fix/gateway.go
grep -r "getEnvOrDefault.*ACCOUNT" backend/fix/gateway.go
grep -r "getEnvOrDefault.*PROXY" backend/fix/gateway.go

# All should return empty (no matches)
```

---

## Deployment Steps

### For Development

```bash
cd backend

# Create .env file with new credentials
cp .env.example .env
# Edit .env with actual credentials

# Test with new credentials
go run ./cmd/server/main.go

# Verify connections
curl http://localhost:7999/admin/fix/status
```

### For Production

```bash
# Using environment variables instead of .env
export YOFX_HOST="23.106.238.138"
export YOFX_PORT="12336"
export YOFX1_PASSWORD="<new_password>"
export YOFX2_PASSWORD="<new_password>"
export YOFX_TRADING_ACCOUNT="<account>"
export YOFX_PROXY_HOST="<proxy_host>"
export YOFX_PROXY_PORT="<proxy_port>"
export YOFX_PROXY_USERNAME="<proxy_user>"
export YOFX_PROXY_PASSWORD="<proxy_pass>"
# ... set all required variables

./server
```

### Using Docker

```dockerfile
# Dockerfile
FROM golang:1.21

WORKDIR /app
COPY backend .

# Credentials injected at runtime via environment
RUN go build -o server ./cmd/server/main.go

CMD ["./server"]
```

Run with:
```bash
docker run \
  -e YOFX_HOST="23.106.238.138" \
  -e YOFX_PORT="12336" \
  -e YOFX1_PASSWORD="<new_password>" \
  -e YOFX2_PASSWORD="<new_password>" \
  -e YOFX_TRADING_ACCOUNT="<account>" \
  -e YOFX_PROXY_HOST="<proxy_host>" \
  -e YOFX_PROXY_PORT="<proxy_port>" \
  -e YOFX_PROXY_USERNAME="<proxy_user>" \
  -e YOFX_PROXY_PASSWORD="<proxy_pass>" \
  trading-engine:latest
```

---

## Credential Management Best Practices

### Option 1: Environment Variables (Simple)

**Pros**: Simple, no additional tools
**Cons**: Not encrypted in transit

```bash
export YOFX1_PASSWORD="secure_password_here"
go run ./cmd/server/main.go
```

### Option 2: .env Files (Local Only)

**Pros**: Local development convenience
**Cons**: .env files can be accidentally committed

```bash
# .env (in .gitignore)
YOFX1_PASSWORD=secure_password_here
YOFX_PROXY_PASSWORD=secure_password_here
```

**Critical**: Add `.env` to `.gitignore`

### Option 3: HashiCorp Vault (Enterprise)

**Pros**: Secure, encrypted, audit logging, rotation policies
**Cons**: Additional infrastructure

```go
// Example using Vault
client, _ := api.NewClient(...)
secret, _ := client.Logical().Read("secret/yofx/credentials")
password := secret.Data["password"].(string)
```

### Option 4: AWS Secrets Manager (Cloud)

**Pros**: AWS-native, KMS encryption, rotation
**Cons**: AWS dependency

```go
client := secretsmanager.New(...)
output, _ := client.GetSecretValue(...)
password := output.SecretString
```

### Option 5: Kubernetes Secrets (K8s)

**Pros**: Native K8s solution, integrated
**Cons**: K8s deployment required

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: yofx-credentials
type: Opaque
stringData:
  username: YOFX1
  password: <encoded>
  proxy-username: <encoded>
  proxy-password: <encoded>
```

---

## Post-Remediation Checklist

- [ ] All credentials rotated at source (YoForex, Proxy)
- [ ] Git history cleaned (old credentials removed)
- [ ] Source code updated (no hardcoded values)
- [ ] Helper functions added (getEnvRequired, etc.)
- [ ] .gitignore updated (prevents future commits)
- [ ] .env.example created (templates without secrets)
- [ ] .env file created locally with new credentials
- [ ] YOFX1 session connects successfully
- [ ] YOFX2 session connects successfully
- [ ] Market data flows through to WebSocket
- [ ] Admin endpoints work with `curl /admin/fix/status`
- [ ] Repository pushed with cleaned history

---

## Testing After Changes

### Test 1: Verify Server Starts

```bash
cd backend
go run ./cmd/server/main.go
# Should start without errors about missing env vars
```

### Test 2: Check FIX Connections

```bash
# In another terminal
curl http://localhost:7999/admin/fix/status

# Expected output:
{
  "sessions": {
    "YOFX1": "LOGGED_IN",
    "YOFX2": "LOGGED_IN",
    "LMAX_PROD": "DISCONNECTED",
    "LMAX_DEMO": "DISCONNECTED"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Test 3: Market Data Flow

```bash
# Check tick endpoint
curl "http://localhost:7999/api/v1/ticks/EURUSD?limit=5"

# Should return recent quotes from FIX connection
[
  {
    "symbol": "EURUSD",
    "bid": 1.0850,
    "ask": 1.0852,
    "timestamp": "2024-01-15T10:30:15Z",
    "lp": "yoforex"
  },
  ...
]
```

### Test 4: Admin Diagnostics

```bash
curl "http://localhost:7999/admin/fix/diagnostics?session_id=YOFX1"

# Should show detailed connection information
# without exposing credentials in output
```

---

## Emergency Actions

### If Credentials Are Still Exposed

1. **Immediately revoke old credentials** at YoForex
2. **Force push cleaned repository** (requires force push permissions)
3. **Notify all developers** to pull latest changes
4. **Rotate all credentials again** to be safe
5. **Update Docker/deployment configs** with new secrets

### If Unauthorized Access Suspected

1. **Disconnect all FIX sessions** immediately
2. **Review YoForex account activity** for suspicious trades
3. **File incident report** with YoForex
4. **Check git logs** for unauthorized changes
5. **Audit all deployed instances** for data exfiltration
6. **Restore from known-good backup** if needed
7. **File incident report** with security team

---

## Long-Term Security Improvements

### 1. Automated Secret Scanning

Add pre-commit hook to prevent credential commits:

```bash
# .git/hooks/pre-commit
#!/bin/bash
if git diff --cached | grep -E 'Brand#143|password.*=.*[^$]|api.?key|secret.?key'; then
    echo "ERROR: Detected potential credentials in commit"
    exit 1
fi
```

### 2. CI/CD Secret Scanning

Use tools like `git-secrets` or `detect-secrets`:

```yaml
# .github/workflows/security.yml
name: Secret Scanning
on: [push, pull_request]
jobs:
  secrets:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: trufflesecurity/trufflehog@main
```

### 3. Credential Rotation Policy

- Rotate YoForex credentials quarterly
- Rotate proxy credentials quarterly
- Implement automated rotation via Vault/Secrets Manager
- Monitor for unused credentials

### 4. Access Control

- Limit repository access to authenticated users
- Use branch protection rules
- Require code review for backend changes
- Audit who accesses environment variables

### 5. Monitoring & Alerting

- Alert on failed FIX connection attempts
- Alert on credential access
- Log all connection attempts
- Monitor for unusual trading patterns

---

## References

- [OWASP: Secrets Management](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [Git Docs: Removing Sensitive Data](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/removing-sensitive-data-from-a-repository)
- [HashiCorp Vault](https://www.vaultproject.io/)
- [FIX Protocol Specification](https://www.fixtrading.org/)

