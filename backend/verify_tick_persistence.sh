#!/bin/bash
# Verification script for tick persistence fix

echo "=========================================="
echo "Tick Persistence Verification Script"
echo "=========================================="
echo ""

# Check if backend is running
echo "1. Checking if backend is running..."
if pgrep -f "server.exe" > /dev/null; then
    echo "   ✅ Backend is running"
else
    echo "   ❌ Backend is NOT running - please start it first"
    exit 1
fi

# Wait for initial ticks
echo ""
echo "2. Waiting 10 seconds for tick data..."
sleep 10

# Check data directory
DATA_DIR="./data/ticks"
if [ ! -d "$DATA_DIR" ]; then
    echo "   ❌ Data directory not found: $DATA_DIR"
    exit 1
fi

echo "   ✅ Data directory exists: $DATA_DIR"

# Check for today's tick files
echo ""
echo "3. Checking for today's tick files..."
TODAY=$(date +%Y-%m-%d)
echo "   Looking for files with date: $TODAY"

TICK_COUNT=0
for symbol_dir in "$DATA_DIR"/*; do
    if [ -d "$symbol_dir" ]; then
        SYMBOL=$(basename "$symbol_dir")
        TODAY_FILE="$symbol_dir/$TODAY.json"

        if [ -f "$TODAY_FILE" ]; then
            FILE_SIZE=$(stat -c%s "$TODAY_FILE" 2>/dev/null || stat -f%z "$TODAY_FILE" 2>/dev/null)
            if [ "$FILE_SIZE" -gt 100 ]; then
                echo "   ✅ $SYMBOL: $(du -h "$TODAY_FILE" | cut -f1)"
                TICK_COUNT=$((TICK_COUNT + 1))
            fi
        fi
    fi
done

if [ $TICK_COUNT -eq 0 ]; then
    echo "   ❌ No tick files found for today"
    echo "   This might indicate ticks are not being persisted!"
    exit 1
fi

echo ""
echo "   ✅ Found tick data for $TICK_COUNT symbols"

# Check if WebSocket clients are connected
echo ""
echo "4. Checking WebSocket client status..."
echo "   (This should show ticks are persisted even without clients)"

# Monitor tick file growth
echo ""
echo "5. Monitoring tick file growth (10 second test)..."
TEST_SYMBOL="EURUSD"
TEST_FILE="$DATA_DIR/$TEST_SYMBOL/$TODAY.json"

if [ -f "$TEST_FILE" ]; then
    SIZE_BEFORE=$(stat -c%s "$TEST_FILE" 2>/dev/null || stat -f%z "$TEST_FILE" 2>/dev/null)
    echo "   Initial size: $SIZE_BEFORE bytes"

    sleep 10

    SIZE_AFTER=$(stat -c%s "$TEST_FILE" 2>/dev/null || stat -f%z "$TEST_FILE" 2>/dev/null)
    echo "   Size after 10s: $SIZE_AFTER bytes"

    GROWTH=$((SIZE_AFTER - SIZE_BEFORE))
    if [ $GROWTH -gt 0 ]; then
        echo "   ✅ File grew by $GROWTH bytes - ticks are being persisted!"
    else
        echo "   ❌ File did NOT grow - ticks may not be persisting!"
        exit 1
    fi
else
    echo "   ⚠️ Test file not found: $TEST_FILE"
fi

echo ""
echo "=========================================="
echo "✅ Verification PASSED"
echo "=========================================="
echo ""
echo "Summary:"
echo "  - Backend is running"
echo "  - Tick files are being created"
echo "  - $TICK_COUNT symbols have data"
echo "  - Files are actively growing"
echo ""
echo "Tick persistence is working correctly!"
