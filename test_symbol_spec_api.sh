#!/bin/bash

# Test Symbol Specification API

echo "Testing Symbol Specification API"
echo "================================="
echo ""

# Test EURUSD
echo "1. Testing EURUSD specification:"
curl -s http://localhost:7999/api/symbols/EURUSD/spec | jq .
echo ""

# Test GBPUSD
echo "2. Testing GBPUSD specification:"
curl -s http://localhost:7999/api/symbols/GBPUSD/spec | jq .
echo ""

# Test XAUUSD (Gold)
echo "3. Testing XAUUSD (Gold) specification:"
curl -s http://localhost:7999/api/symbols/XAUUSD/spec | jq .
echo ""

# Test invalid symbol
echo "4. Testing invalid symbol (should return 404):"
curl -s -w "\nHTTP Status: %{http_code}\n" http://localhost:7999/api/symbols/INVALID/spec
echo ""

# Test invalid characters (should return 400)
echo "5. Testing symbol with invalid characters (should return 400):"
curl -s -w "\nHTTP Status: %{http_code}\n" http://localhost:7999/api/symbols/EUR-USD/spec
echo ""
