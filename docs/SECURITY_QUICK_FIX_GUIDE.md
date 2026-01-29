# Security Quick Fix Guide - URGENT
**For**: Stability & Security Agent Deliverables
**Priority**: P0 - Fix Before Production

---

## 1. Path Traversal Fix (HIGH RISK)

**File**: `backend/scripts/rotate_ticks.sh`
**Lines**: 154-156

### Current Vulnerable Code:
```bash
local file_date=$(stat -f%Sb -t%Y-%m-%d "$file" 2>/dev/null || stat -c%y "$file" 2>/dev/null | cut -d' ' -f1)
local archive_subdir="${archive_directory}/${file_date}"
```

### Fixed Code:
```bash
# Extract date and SANITIZE
local file_date=$(stat -f%Sb -t%Y-%m-%d "$file" 2>/dev/null || stat -c%y "$file" 2>/dev/null | cut -d' ' -f1)

# SECURITY FIX: Sanitize file_date to prevent path traversal
file_date=$(echo "$file_date" | sed 's/[^0-9-]//g')

# SECURITY FIX: Validate date format (YYYY-MM-DD)
if [[ ! "$file_date" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; then
    log_msg WARN "Invalid date format: $file_date, skipping file $filename"
    continue
fi

# SECURITY FIX: Verify date is actually valid
if ! date -d "$file_date" &>/dev/null 2>&1; then
    log_msg WARN "Invalid date: $file_date, skipping file $filename"
    continue
fi

local archive_subdir="${archive_directory}/${file_date}"

# SECURITY FIX: Verify archive_subdir doesn't escape archive_directory
archive_subdir=$(realpath -m "$archive_subdir")
archive_base=$(realpath -m "$archive_directory")
if [[ ! "$archive_subdir" == "$archive_base"* ]]; then
    log_msg ERROR "Security violation: Path traversal detected - $archive_subdir"
    continue
fi
```

---

## 2. Command Injection Fix (HIGH RISK)

**File**: `backend/scripts/migrate-json-to-timescale.sh`
**Lines**: 116-117

### Current Vulnerable Code:
```bash
migrate_symbol() {
    local symbol_dir="$1"
    local symbol=$(basename "$symbol_dir")  # NO VALIDATION!
```

### Fixed Code:
```bash
migrate_symbol() {
    local symbol_dir="$1"
    local symbol=$(basename "$symbol_dir")

    # SECURITY FIX: Validate symbol format (6-10 uppercase alphanumeric)
    if [[ ! "$symbol" =~ ^[A-Z0-9]{6,10}$ ]]; then
        log_error "Invalid symbol format: $symbol (expected 6-10 uppercase alphanumeric)"
        return 1
    fi

    # SECURITY FIX: Prevent SQL injection via symbol name
    # Check for SQL metacharacters
    if [[ "$symbol" =~ [\'\"\;\`\$\(\)\{\}] ]]; then
        log_error "Invalid characters in symbol: $symbol"
        return 1
    fi

    log_info "Processing symbol: $symbol"
    # ... rest of function
}
```

### Also Fix psql Command (Line 109):
```bash
import_csv_to_database() {
    local csv_file="$1"
    local symbol="$2"

    # SECURITY FIX: Re-validate symbol before SQL operation
    if [[ ! "$symbol" =~ ^[A-Z0-9]{6,10}$ ]]; then
        log_error "Symbol validation failed: $symbol"
        return 1
    fi

    export PGPASSWORD="$DB_PASSWORD"

    # Use psql variable binding (safer than string interpolation)
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -v symbol="$symbol" \
        -c "COPY tick_history (timestamp, broker_id, symbol, bid, ask, spread, lp) FROM STDIN WITH (FORMAT CSV);" \
        < "$csv_file"
}
```

---

## 3. Input Validation Fix - admin_history.go (HIGH RISK)

**File**: `backend/api/admin_history.go`
**Lines**: 256-281

### Current Vulnerable Code:
```go
basePath := filepath.Join("data", "ticks", symbol)
```

### Fixed Code:
```go
// SECURITY FIX: Validate symbol format before filesystem operations
func isValidSymbol(s string) bool {
    // Must be 6-10 uppercase letters/numbers
    matched, _ := regexp.MatchString(`^[A-Z0-9]{6,10}$`, s)
    if !matched {
        return false
    }
    // Explicitly block path traversal sequences
    if strings.Contains(s, "..") || strings.Contains(s, "/") || strings.Contains(s, "\\") {
        return false
    }
    return true
}

// In HandleCleanupOldData function:
for _, symbol := range symbols {
    // SECURITY FIX: Validate symbol before path operations
    if !isValidSymbol(symbol) {
        log.Printf("[AdminHistory] Invalid symbol rejected: %s", symbol)
        continue
    }

    basePath := filepath.Join("data", "ticks", symbol)

    // SECURITY FIX: Ensure basePath doesn't escape data/ticks
    cleanPath := filepath.Clean(basePath)
    if !strings.HasPrefix(cleanPath, "data/ticks/") {
        log.Printf("[AdminHistory] Path traversal blocked: %s -> %s", symbol, cleanPath)
        continue
    }

    // Rest of cleanup logic...
}
```

### Add to top of file:
```go
import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "regexp"  // ADD THIS
    "strings" // ENSURE THIS
    "time"
)
```

---

## 4. Input Validation Fix - history.go (MEDIUM RISK)

**File**: `backend/api/history.go`
**Lines**: 196-206, 672-676

### Fix 1: Symbol Extraction (Line 196)
```go
// Current vulnerable code:
parts := strings.Split(r.URL.Path, "/")
symbol := ""
if len(parts) >= 5 {
    symbol = parts[4]  // NO VALIDATION
}

// SECURITY FIX:
parts := strings.Split(r.URL.Path, "/")
symbol := ""
if len(parts) >= 5 {
    symbol = filepath.Clean(parts[4])  // Basic sanitization

    // Validate format
    if !isValidSymbol(symbol) {
        http.Error(w, "Invalid symbol format", http.StatusBadRequest)
        return
    }
}
```

### Fix 2: Offset/Limit Validation (Line 672)
```go
// Current vulnerable code:
offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
if limit <= 0 {
    limit = 5000
}

// SECURITY FIX:
offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
if err != nil || offset < 0 {
    offset = 0  // Reset negative offsets to 0
}

limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
if err != nil || limit <= 0 {
    limit = 5000
}
if limit > 50000 {
    limit = 50000  // Cap at 50k
}

// SECURITY FIX: Bounds check before slicing
total := len(allTicks)
if offset >= total {
    allTicks = []tickstore.Tick{}  // Return empty if offset too high
} else {
    end := offset + limit
    if end > total {
        end = total
    }
    allTicks = allTicks[offset:end]  // Safe slicing
}
```

### Fix 3: page_size Validation (Line 222)
```go
// Current code:
pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
if pageSize < 1 || pageSize > 10000 {
    pageSize = 1000
}

// SECURITY FIX: Add explicit max check
pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
if err != nil || pageSize < 1 {
    pageSize = 1000
}
if pageSize > 10000 {
    pageSize = 10000  // Hard cap at 10k
}
```

---

## 5. Console Log Removal (MEDIUM RISK)

**Files**: 34 frontend files (151 total occurrences)

### Step 1: Create Logger Utility
**File**: `clients/desktop/src/utils/logger.ts`

```typescript
// Production-safe logger utility
const isDevelopment = process.env.NODE_ENV === 'development';

export const logger = {
  log: (...args: any[]) => {
    if (isDevelopment) {
      console.log(...args);
    }
  },

  error: (...args: any[]) => {
    // Always log errors (but sanitize sensitive data)
    console.error(...args);
  },

  warn: (...args: any[]) => {
    if (isDevelopment) {
      console.warn(...args);
    }
  },

  info: (...args: any[]) => {
    if (isDevelopment) {
      console.info(...args);
    }
  },

  debug: (...args: any[]) => {
    if (isDevelopment) {
      console.debug(...args);
    }
  }
};
```

### Step 2: Replace All console.log Calls
```bash
# Find and replace in all files
find clients/desktop/src -name "*.ts" -o -name "*.tsx" | xargs sed -i 's/console\.log/logger.log/g'
find clients/desktop/src -name "*.ts" -o -name "*.tsx" | xargs sed -i 's/console\.warn/logger.warn/g'
find clients/desktop/src -name "*.ts" -o -name "*.tsx" | xargs sed -i 's/console\.info/logger.info/g'
```

### Step 3: Add Import to Each File
```typescript
import { logger } from '@/utils/logger';

// Replace:
console.log('[WS] Connecting to:', url, 'with token:', token);

// With:
logger.log('[WS] Connecting to:', url);  // REMOVE TOKEN FROM LOG
```

### Priority Files (Most Occurrences):
1. `websocket-enhanced.ts` - 26 occurrences
2. `cache-manager.ts` - 12 occurrences
3. `websocket.ts` - 12 occurrences

---

## 6. SQLite WAL Checkpoint Fix (MEDIUM RISK)

**File**: `backend/tickstore/sqlite_store.go`
**Lines**: 441-446

### Current Code:
```go
if len(batch) >= 450 {
    if _, err := db.Exec("PRAGMA wal_checkpoint(PASSIVE)"); err != nil {
        log.Printf("[SQLiteStore] PASSIVE checkpoint deferred (DB busy): %v", err)
    }
}
```

### Fixed Code:
```go
// STABILITY FIX: Track batch counter for periodic FULL checkpoint
// Add to struct:
// batchCounter int64

// In writeBatch function:
batchCounter := atomic.AddInt64(&s.batchCounter, 1)

// Run FULL checkpoint every 10 batches (5000 ticks)
if batchCounter % 10 == 0 {
    log.Printf("[SQLiteStore] Running FULL checkpoint (batch %d)...", batchCounter)

    s.mu.RLock()
    db := s.db
    s.mu.RUnlock()

    if db != nil {
        // FULL blocks writers until WAL is merged
        if _, err := db.Exec("PRAGMA wal_checkpoint(FULL)"); err != nil {
            log.Printf("[SQLiteStore] FULL checkpoint failed: %v", err)

            // Escalate to TRUNCATE if FULL fails
            if _, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
                log.Printf("[SQLiteStore] CRITICAL: TRUNCATE checkpoint failed: %v", err)
                // TODO: Alert operations team
            }
        } else {
            log.Printf("[SQLiteStore] FULL checkpoint completed successfully")
        }
    }
} else if len(batch) >= 450 {
    // Try PASSIVE for smaller batches (non-blocking)
    if _, err := db.Exec("PRAGMA wal_checkpoint(PASSIVE)"); err != nil {
        // Expected if busy - will checkpoint on next FULL
    }
}
```

### Add to struct (Line 105):
```go
type SQLiteStore struct {
    // ... existing fields
    batchCounter int64  // ADD THIS for checkpoint tracking
}
```

---

## 7. Testing Commands

```bash
# 1. Test path traversal fix
echo "Testing path traversal protection..."
touch "backend/data/ticks/EURUSD/2026-01-01/../../../etc/passwd.json"
bash backend/scripts/rotate_ticks.sh --dry-run
# Expected: "Invalid date format" or "Security violation" error

# 2. Test command injection fix
echo "Testing command injection protection..."
mkdir -p "backend/data/ticks/TEST'; DROP TABLE tick_history; --"
bash backend/scripts/migrate-json-to-timescale.sh 2>&1 | grep -i "invalid"
# Expected: "Invalid symbol format" error

# 3. Test input validation
echo "Testing API input validation..."
curl "http://localhost:8080/api/history/ticks/../../etc/passwd"
# Expected: 400 Bad Request

curl "http://localhost:8080/api/history/ticks?symbol=EURUSD&offset=-1&limit=100"
# Expected: Returns data with offset=0 (no panic)

# 4. Check console.log removal
echo "Checking console.log statements..."
grep -r "console\.log" clients/desktop/src --exclude-dir=node_modules
# Expected: Only logger.log or development-only logs

# 5. Monitor SQLite WAL size
echo "Monitoring SQLite WAL file size..."
watch -n 5 'ls -lh backend/data/ticks/*.db-wal'
# Expected: WAL files < 100MB, periodic shrinking
```

---

## 8. Deployment Checklist

Before deploying to production:

- [ ] Path traversal fix applied and tested
- [ ] Command injection fix applied and tested
- [ ] Input validation added to all API endpoints
- [ ] All 151 console.log statements replaced with logger
- [ ] SQLite WAL checkpoint strategy implemented
- [ ] Security tests pass (see section 7)
- [ ] Code review completed
- [ ] Penetration testing scheduled
- [ ] Monitoring alerts configured
- [ ] Runbook updated with security procedures

---

## 9. Emergency Rollback

If issues arise after deployment:

```bash
# 1. Revert to previous version
git revert HEAD

# 2. Disable affected endpoints
# In main.go or server.go:
# Comment out vulnerable routes temporarily

# 3. Enable emergency mode
export EMERGENCY_MODE=true
# This should disable non-essential features

# 4. Monitor error rates
tail -f /var/log/trading-engine/errors.log

# 5. Notify team
# Send alert to #security-incidents Slack channel
```

---

**End of Quick Fix Guide**
