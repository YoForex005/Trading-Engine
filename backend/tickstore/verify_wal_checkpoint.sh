#!/bin/bash
# WAL Checkpoint Verification Script
# Tests that WAL checkpoints are working correctly

set -e

DB_DIR="../../data/ticks/2026/01"
DB_FILE="ticks_2026-01-20.db"
WAL_FILE="${DB_FILE}-wal"

echo "=== WAL Checkpoint Verification ==="
echo ""

# 1. Check if DB exists
echo "1. Checking database file..."
if [ -f "${DB_DIR}/${DB_FILE}" ]; then
    DB_SIZE=$(stat -c%s "${DB_DIR}/${DB_FILE}" 2>/dev/null || stat -f%z "${DB_DIR}/${DB_FILE}" 2>/dev/null || echo "unknown")
    echo "   ✓ Database exists: ${DB_FILE} (${DB_SIZE} bytes)"
else
    echo "   ✗ Database not found: ${DB_FILE}"
    exit 1
fi

# 2. Check WAL file status
echo ""
echo "2. Checking WAL file..."
if [ -f "${DB_DIR}/${WAL_FILE}" ]; then
    WAL_SIZE=$(stat -c%s "${DB_DIR}/${WAL_FILE}" 2>/dev/null || stat -f%z "${DB_DIR}/${WAL_FILE}" 2>/dev/null || echo "unknown")
    echo "   ⚠ WAL file exists: ${WAL_FILE} (${WAL_SIZE} bytes)"

    if [ "${WAL_SIZE}" == "0" ] || [ "${WAL_SIZE}" == "unknown" ]; then
        echo "   ✓ WAL file is empty (checkpoint successful)"
    else
        echo "   ⚠ WAL file has data (checkpoint may not have run)"
    fi
else
    echo "   ✓ No WAL file (checkpoint successful or WAL mode not enabled)"
fi

# 3. Check database integrity
echo ""
echo "3. Running integrity check..."
INTEGRITY=$(sqlite3 "${DB_DIR}/${DB_FILE}" "PRAGMA integrity_check;" 2>&1)
if [ "$INTEGRITY" == "ok" ]; then
    echo "   ✓ Database integrity: OK"
else
    echo "   ✗ Database integrity: FAILED"
    echo "   Error: $INTEGRITY"
    exit 1
fi

# 4. Check WAL mode
echo ""
echo "4. Checking journal mode..."
JOURNAL_MODE=$(sqlite3 "${DB_DIR}/${DB_FILE}" "PRAGMA journal_mode;" 2>&1)
echo "   Journal mode: ${JOURNAL_MODE}"
if [ "$JOURNAL_MODE" == "wal" ]; then
    echo "   ✓ WAL mode enabled"
else
    echo "   ⚠ WAL mode not enabled (expected: wal, got: ${JOURNAL_MODE})"
fi

# 5. Check synchronous mode
echo ""
echo "5. Checking synchronous mode..."
SYNC_MODE=$(sqlite3 "${DB_DIR}/${DB_FILE}" "PRAGMA synchronous;" 2>&1)
case $SYNC_MODE in
    0) SYNC_NAME="OFF" ;;
    1) SYNC_NAME="NORMAL" ;;
    2) SYNC_NAME="FULL" ;;
    3) SYNC_NAME="EXTRA" ;;
    *) SYNC_NAME="UNKNOWN" ;;
esac
echo "   Synchronous: ${SYNC_MODE} (${SYNC_NAME})"
if [ "$SYNC_MODE" == "1" ]; then
    echo "   ✓ NORMAL mode (balanced durability/performance)"
else
    echo "   ⚠ Not NORMAL mode"
fi

# 6. Check tick count
echo ""
echo "6. Checking tick data..."
TICK_COUNT=$(sqlite3 "${DB_DIR}/${DB_FILE}" "SELECT COUNT(*) FROM ticks;" 2>&1)
SYMBOL_COUNT=$(sqlite3 "${DB_DIR}/${DB_FILE}" "SELECT COUNT(DISTINCT symbol) FROM ticks;" 2>&1)
echo "   Total ticks: ${TICK_COUNT}"
echo "   Unique symbols: ${SYMBOL_COUNT}"

# 7. Sample recent ticks
echo ""
echo "7. Sample recent ticks (last 3)..."
sqlite3 "${DB_DIR}/${DB_FILE}" "SELECT symbol, datetime(timestamp/1000, 'unixepoch'), bid, ask FROM ticks ORDER BY timestamp DESC LIMIT 3;" 2>&1 | while read line; do
    echo "   $line"
done

echo ""
echo "=== Verification Complete ==="

# Summary
echo ""
echo "Summary:"
echo "  - Database integrity: OK"
echo "  - WAL checkpoint: $([ -f "${DB_DIR}/${WAL_FILE}" ] && [ "${WAL_SIZE}" != "0" ] && echo 'PENDING' || echo 'COMPLETE')"
echo "  - Total ticks stored: ${TICK_COUNT}"
echo ""
echo "To manually checkpoint WAL:"
echo "  sqlite3 ${DB_DIR}/${DB_FILE} 'PRAGMA wal_checkpoint(TRUNCATE);'"
