#!/bin/bash

# RTX Trading Engine - Development Startup Script
# This script starts all services for local development

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║          RTX Trading Engine - Development Mode            ║"
echo "╚═══════════════════════════════════════════════════════════╝"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "⚠️  Go is not installed. Please install with: brew install go"
    echo "   Backend will not start."
else
    echo -e "${GREEN}✓${NC} Starting Backend on :8080..."
    cd backend
    go run main.go &
    BACKEND_PID=$!
    cd ..
fi

# Start Desktop Terminal
echo -e "${BLUE}✓${NC} Starting Desktop Terminal on :5173..."
cd clients/desktop
npm run dev &
DESKTOP_PID=$!
cd ../..

# Start Broker Admin
echo -e "${BLUE}✓${NC} Starting Broker Admin on :3000..."
cd admin/broker-admin
npm run dev -- -p 3000 &
BROKER_PID=$!
cd ../..

# Start Super Admin
echo -e "${BLUE}✓${NC} Starting Super Admin on :3001..."
cd admin/super-admin
npm run dev -- -p 3001 &
SUPER_PID=$!
cd ../..

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "Services started:"
echo "  • Backend API:        http://localhost:8080"
echo "  • Desktop Terminal:   http://localhost:5173"
echo "  • Broker Admin:       http://localhost:3000"
echo "  • Super Admin:        http://localhost:3001"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "Press Ctrl+C to stop all services"

# Wait for Ctrl+C
trap "echo 'Stopping all services...'; kill $BACKEND_PID $DESKTOP_PID $BROKER_PID $SUPER_PID 2>/dev/null; exit" SIGINT

wait
