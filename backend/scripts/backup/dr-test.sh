#!/bin/bash
# Automated Disaster Recovery Test
# Runs monthly to validate backup/restore procedures

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${BACKUP_CONFIG:-$SCRIPT_DIR/backup.config}"
LOG_DIR="${BACKUP_LOG_DIR:-/var/log/trading-engine/backup}"
TEST_ID="dr-test-$(date +%Y%m%d-%H%M%S)"

# Load configuration
if [[ -f "$CONFIG_FILE" ]]; then
    # shellcheck source=/dev/null
    source "$CONFIG_FILE"
fi

BACKUP_ROOT="${BACKUP_ROOT:-/var/backups/trading-engine}"
ALERT_EMAIL="${ALERT_EMAIL:-ops@trading-engine.local}"

mkdir -p "$LOG_DIR"
TEST_LOG="$LOG_DIR/dr-test.log"

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$TEST_LOG"
}

error() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*" | tee -a "$TEST_LOG" >&2
}

send_report() {
    local status="$1"
    local message="$2"

    if command -v mail >/dev/null 2>&1; then
        echo -e "$message" | mail -s "DR Test Report - $status" "$ALERT_EMAIL"
    fi
}

log "=========================================="
log "Disaster Recovery Test: $TEST_ID"
log "=========================================="

START_TIME=$(date +%s)
TEST_RESULTS=()
TEST_PASSED=0
TEST_FAILED=0

# Test 1: Verify backup exists
log "Test 1: Verify recent backup exists..."
LATEST_BACKUP=$(find "$BACKUP_ROOT" -name "full-*.tar.gz*" -type f | sort -r | head -1)

if [[ -n "$LATEST_BACKUP" ]] && [[ -f "$LATEST_BACKUP" ]]; then
    BACKUP_AGE_HOURS=$(( ($(date +%s) - $(stat -f %m "$LATEST_BACKUP" 2>/dev/null || stat -c %Y "$LATEST_BACKUP")) / 3600 ))
    log "✓ Latest backup found: $(basename "$LATEST_BACKUP") (${BACKUP_AGE_HOURS}h old)"
    TEST_RESULTS+=("✓ Backup exists (age: ${BACKUP_AGE_HOURS}h)")
    ((TEST_PASSED++))
else
    error "✗ No backup found"
    TEST_RESULTS+=("✗ No backup found")
    ((TEST_FAILED++))
    LATEST_BACKUP=""
fi

# Test 2: Verify backup integrity
if [[ -n "$LATEST_BACKUP" ]]; then
    log "Test 2: Verify backup integrity..."

    VERIFY_START=$(date +%s)
    if "$SCRIPT_DIR/backup-verify.sh" "$LATEST_BACKUP" >> "$TEST_LOG" 2>&1; then
        VERIFY_DURATION=$(($(date +%s) - VERIFY_START))
        log "✓ Backup verification passed (${VERIFY_DURATION}s)"
        TEST_RESULTS+=("✓ Backup integrity verified (${VERIFY_DURATION}s)")
        ((TEST_PASSED++))
    else
        error "✗ Backup verification failed"
        TEST_RESULTS+=("✗ Backup verification failed")
        ((TEST_FAILED++))
    fi
fi

# Test 3: Test restore time (RTO)
log "Test 3: Measure restore time (RTO)..."

if [[ -n "$LATEST_BACKUP" ]]; then
    TEST_DB_NAME="dr_test_${TEST_ID}"

    RESTORE_START=$(date +%s)

    # Extract and restore to test database
    TEMP_DIR="/tmp/dr-test-${TEST_ID}"
    mkdir -p "$TEMP_DIR"

    # Decrypt
    if [[ "$LATEST_BACKUP" == *.gpg ]]; then
        gpg --decrypt --output "$TEMP_DIR/backup.tar.gz" "$LATEST_BACKUP" 2>/dev/null
        BACKUP_ARCHIVE="$TEMP_DIR/backup.tar.gz"
    else
        BACKUP_ARCHIVE="$LATEST_BACKUP"
    fi

    # Extract
    tar -xzf "$BACKUP_ARCHIVE" -C "$TEMP_DIR" 2>/dev/null

    # Find PostgreSQL dump
    POSTGRES_DUMP=$(find "$TEMP_DIR" -name "postgres-*.dump" | head -1)

    if [[ -f "$POSTGRES_DUMP" ]]; then
        # Create test database
        PGPASSWORD="${DB_PASSWORD}" createdb \
            -h "${DB_HOST:-localhost}" \
            -p "${DB_PORT:-5432}" \
            -U "${DB_USER:-postgres}" \
            "$TEST_DB_NAME" 2>/dev/null || true

        # Restore
        PGPASSWORD="${DB_PASSWORD}" pg_restore \
            -h "${DB_HOST:-localhost}" \
            -p "${DB_PORT:-5432}" \
            -U "${DB_USER:-postgres}" \
            -d "$TEST_DB_NAME" \
            "$POSTGRES_DUMP" >> "$TEST_LOG" 2>&1

        RESTORE_DURATION=$(($(date +%s) - RESTORE_START))
        RESTORE_MINUTES=$((RESTORE_DURATION / 60))
        RESTORE_SECONDS=$((RESTORE_DURATION % 60))

        # Verify table count
        TABLE_COUNT=$(PGPASSWORD="${DB_PASSWORD}" psql \
            -h "${DB_HOST:-localhost}" \
            -p "${DB_PORT:-5432}" \
            -U "${DB_USER:-postgres}" \
            -d "$TEST_DB_NAME" \
            -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | xargs)

        # Check RTO target (15 minutes)
        if [[ $RESTORE_DURATION -le 900 ]]; then
            log "✓ RTO target met: ${RESTORE_MINUTES}m ${RESTORE_SECONDS}s < 15m"
            TEST_RESULTS+=("✓ RTO: ${RESTORE_MINUTES}m ${RESTORE_SECONDS}s (target: 15m)")
            ((TEST_PASSED++))
        else
            error "✗ RTO target exceeded: ${RESTORE_MINUTES}m ${RESTORE_SECONDS}s > 15m"
            TEST_RESULTS+=("✗ RTO: ${RESTORE_MINUTES}m ${RESTORE_SECONDS}s (target: 15m)")
            ((TEST_FAILED++))
        fi

        # Verify data
        if [[ $TABLE_COUNT -gt 0 ]]; then
            log "✓ Data restored: $TABLE_COUNT tables"
            TEST_RESULTS+=("✓ Data integrity: $TABLE_COUNT tables")
            ((TEST_PASSED++))
        else
            error "✗ No tables restored"
            TEST_RESULTS+=("✗ No tables restored")
            ((TEST_FAILED++))
        fi

        # Cleanup test database
        PGPASSWORD="${DB_PASSWORD}" dropdb \
            -h "${DB_HOST:-localhost}" \
            -p "${DB_PORT:-5432}" \
            -U "${DB_USER:-postgres}" \
            "$TEST_DB_NAME" 2>/dev/null || true
    else
        error "✗ PostgreSQL dump not found in backup"
        TEST_RESULTS+=("✗ PostgreSQL dump not found")
        ((TEST_FAILED++))
    fi

    # Cleanup temp files
    rm -rf "$TEMP_DIR"
fi

# Test 4: Verify S3 backup sync
if [[ "${ENABLE_S3:-false}" == "true" ]] && command -v aws >/dev/null 2>&1; then
    log "Test 4: Verify S3 backup sync..."

    S3_COUNT=$(aws s3 ls "s3://${S3_BUCKET}/backups/full/" 2>/dev/null | wc -l || echo 0)

    if [[ $S3_COUNT -gt 0 ]]; then
        log "✓ S3 backups available: $S3_COUNT"
        TEST_RESULTS+=("✓ S3 sync working: $S3_COUNT backups")
        ((TEST_PASSED++))
    else
        error "✗ No S3 backups found"
        TEST_RESULTS+=("✗ No S3 backups found")
        ((TEST_FAILED++))
    fi
fi

# Test 5: Verify encryption/decryption
log "Test 5: Verify encryption/decryption..."

if [[ "${ENABLE_ENCRYPTION:-true}" == "true" ]] && [[ -n "$LATEST_BACKUP" ]] && [[ "$LATEST_BACKUP" == *.gpg ]]; then
    TEST_FILE="/tmp/test-encryption-${TEST_ID}.txt"
    echo "DR test data" > "$TEST_FILE"

    # Encrypt
    gpg --encrypt --recipient "${GPG_RECIPIENT}" --output "$TEST_FILE.gpg" "$TEST_FILE" 2>/dev/null

    # Decrypt
    gpg --decrypt --output "$TEST_FILE.decrypted" "$TEST_FILE.gpg" 2>/dev/null

    # Compare
    if diff "$TEST_FILE" "$TEST_FILE.decrypted" >/dev/null 2>&1; then
        log "✓ Encryption/decryption working"
        TEST_RESULTS+=("✓ Encryption verified")
        ((TEST_PASSED++))
    else
        error "✗ Encryption/decryption failed"
        TEST_RESULTS+=("✗ Encryption failed")
        ((TEST_FAILED++))
    fi

    rm -f "$TEST_FILE" "$TEST_FILE.gpg" "$TEST_FILE.decrypted"
else
    log "- Encryption test skipped"
fi

# Test 6: Verify monitoring/alerting
log "Test 6: Verify monitoring/alerting..."

if [[ -f "$LOG_DIR/metrics.txt" ]]; then
    HEALTH_STATUS=$(grep "backup_health_status" "$LOG_DIR/metrics.txt" | awk '{print $2}')

    if [[ "$HEALTH_STATUS" == "1" ]]; then
        log "✓ Backup monitoring healthy"
        TEST_RESULTS+=("✓ Monitoring active")
        ((TEST_PASSED++))
    else
        error "✗ Backup monitoring unhealthy: $HEALTH_STATUS"
        TEST_RESULTS+=("✗ Monitoring unhealthy")
        ((TEST_FAILED++))
    fi
else
    log "- Monitoring metrics not found"
fi

# Calculate total duration
END_TIME=$(date +%s)
TOTAL_DURATION=$((END_TIME - START_TIME))
TOTAL_MINUTES=$((TOTAL_DURATION / 60))
TOTAL_SECONDS=$((TOTAL_DURATION % 60))

# Generate report
log "=========================================="
log "DR Test Results"
log "=========================================="
log "Test ID: $TEST_ID"
log "Duration: ${TOTAL_MINUTES}m ${TOTAL_SECONDS}s"
log "Tests Passed: $TEST_PASSED"
log "Tests Failed: $TEST_FAILED"
log ""
log "Detailed Results:"
for result in "${TEST_RESULTS[@]}"; do
    log "  $result"
done
log "=========================================="

# Determine overall status
if [[ $TEST_FAILED -eq 0 ]]; then
    OVERALL_STATUS="PASSED"
    log "✓ DR Test PASSED"
    EXIT_CODE=0
else
    OVERALL_STATUS="FAILED"
    error "✗ DR Test FAILED"
    EXIT_CODE=1
fi

# Send report
REPORT_MESSAGE="Disaster Recovery Test: $OVERALL_STATUS\n\n"
REPORT_MESSAGE+="Test ID: $TEST_ID\n"
REPORT_MESSAGE+="Duration: ${TOTAL_MINUTES}m ${TOTAL_SECONDS}s\n"
REPORT_MESSAGE+="Tests Passed: $TEST_PASSED\n"
REPORT_MESSAGE+="Tests Failed: $TEST_FAILED\n\n"
REPORT_MESSAGE+="Results:\n"
for result in "${TEST_RESULTS[@]}"; do
    REPORT_MESSAGE+="  $result\n"
done
REPORT_MESSAGE+="\nFull log: $TEST_LOG"

send_report "$OVERALL_STATUS" "$REPORT_MESSAGE"

exit $EXIT_CODE
