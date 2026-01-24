#!/bin/bash
# Quick verification script for historical data simulation

echo "================================================"
echo "Historical Data Simulation Verification"
echo "================================================"
echo ""

# Check if server is running
if ! curl -s http://localhost:7999/admin/fix/ticks > /dev/null 2>&1; then
    echo "ERROR: Server is not running on port 7999"
    echo "Please start the server with: ./server.exe"
    exit 1
fi

echo "✓ Server is running"
echo ""

# Wait a bit for ticks to arrive
echo "Fetching tick data..."
sleep 2

# Get tick data
RESPONSE=$(curl -s http://localhost:7999/admin/fix/ticks)

# Check symbol count
SYMBOL_COUNT=$(echo "$RESPONSE" | grep -o '"symbolCount":[0-9]*' | grep -o '[0-9]*')
echo "✓ Symbols loaded: $SYMBOL_COUNT"

# Check LP label
LP_LABEL=$(echo "$RESPONSE" | grep -o '"lp":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "✓ LP Label: $LP_LABEL"

# Show sample prices
echo ""
echo "Sample Prices:"
echo "=============="

# Extract EURUSD
EURUSD_BID=$(echo "$RESPONSE" | grep -o '"EURUSD"[^}]*"bid":[0-9.]*' | grep -o '[0-9.]*$')
EURUSD_ASK=$(echo "$RESPONSE" | grep -o '"EURUSD"[^}]*"ask":[0-9.]*' | grep -o '[0-9.]*$')
echo "EURUSD: Bid=$EURUSD_BID Ask=$EURUSD_ASK"

# Extract GBPUSD
GBPUSD_BID=$(echo "$RESPONSE" | grep -o '"GBPUSD"[^}]*"bid":[0-9.]*' | grep -o '[0-9.]*$')
GBPUSD_ASK=$(echo "$RESPONSE" | grep -o '"GBPUSD"[^}]*"ask":[0-9.]*' | grep -o '[0-9.]*$')
echo "GBPUSD: Bid=$GBPUSD_BID Ask=$GBPUSD_ASK"

# Extract USDJPY
USDJPY_BID=$(echo "$RESPONSE" | grep -o '"USDJPY"[^}]*"bid":[0-9.]*' | grep -o '[0-9.]*$')
USDJPY_ASK=$(echo "$RESPONSE" | grep -o '"USDJPY"[^}]*"ask":[0-9.]*' | grep -o '[0-9.]*$')
echo "USDJPY: Bid=$USDJPY_BID Ask=$USDJPY_ASK"

echo ""
echo "Verifying price movement (checking 3 times over 4 seconds)..."
echo "=============================================================="

for i in {1..3}; do
    PRICE=$(curl -s http://localhost:7999/admin/fix/ticks | grep -o '"EURUSD"[^}]*"bid":[0-9.]*' | grep -o '[0-9.]*$')
    echo "Sample $i: EURUSD Bid = $PRICE"
    sleep 2
done

echo ""
echo "================================================"
echo "✓ Verification Complete!"
echo "================================================"
echo ""
echo "Expected Results:"
echo "- Symbol Count: 16"
echo "- LP Label: OANDA-HISTORICAL"
echo "- Prices should be realistic (EURUSD ~1.05-1.20)"
echo "- Prices should change slightly between samples"
echo ""
