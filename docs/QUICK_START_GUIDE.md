# Quick Start Guide - Trading Engine Testing

## System Status: ✅ OPERATIONAL

Backend is running and broadcasting market data via WebSocket. All systems tested and verified.

---

## 1. Access Test Tools

### WebSocket Test Page
Open in browser to see live tick data:
```bash
start backend/test_websocket.html
```

**Features:**
- Real-time tick visualization
- Connection monitoring
- Statistics dashboard
- 29 subscribed symbols updating every 200ms

### Run Full E2E Tests
Execute comprehensive health check:
```powershell
cd backend
powershell -ExecutionPolicy Bypass -File test_e2e.ps1
```

---

## 2. Quick API Tests

### Health Check
```bash
curl http://localhost:7999/health
# Response: OK
```

### Get Broker Configuration
```bash
curl http://localhost:7999/api/config
# Returns: broker name, execution mode, FIX status
```

### View Live Ticks
```bash
curl http://localhost:7999/admin/fix/ticks
# Returns: 30 symbols with latest bid/ask/spread
```

### Check Market Data Flow
```bash
curl http://localhost:7999/api/diagnostics/market-data
# Returns: tick count, active streams, latency stats
```

### List Available Symbols
```bash
curl http://localhost:7999/api/symbols/available
# Returns: all tradeable symbols with categories
```

### List Subscribed Symbols
```bash
curl http://localhost:7999/api/symbols/subscribed
# Returns: 29 currently subscribed symbols
```

---

## 3. Connect FIX Sessions (Optional)

Currently using simulated data. To enable live YOFX data:

### Connect YOFX1 (Trading Account)
```bash
curl -X POST http://localhost:7999/admin/fix/connect \
  -H "Content-Type: application/json" \
  -d "{\"sessionId\": \"YOFX1\"}"
```

### Connect YOFX2 (Market Data Feed)
```bash
curl -X POST http://localhost:7999/admin/fix/connect \
  -H "Content-Type: application/json" \
  -d "{\"sessionId\": \"YOFX2\"}"
```

### Check FIX Status
```bash
curl http://localhost:7999/admin/fix/status
# Shows: LOGGED_IN, CONNECTED, or DISCONNECTED for each session
```

---

## 4. Start Frontend (Desktop Client)

```bash
cd clients/desktop
npm install  # if not done already
npm run dev
```

**Access:** http://localhost:3000

**Expected:**
- WebSocket connects to ws://localhost:7999/ws
- Charts display real-time tick data
- Market Watch shows 30+ symbols
- Trading panel ready for orders

---

## 5. Key Endpoints Reference

### Public APIs
- `GET /health` - Health check
- `GET /api/config` - Broker configuration
- `GET /api/symbols/available` - Available symbols

### Market Data
- `GET /api/diagnostics/market-data` - Market data status
- `GET /admin/fix/ticks` - Current tick data
- `GET /api/symbols/subscribed` - Subscribed symbols
- `WS ws://localhost:7999/ws` - Real-time tick stream

### Admin APIs
- `GET /admin/fix/status` - FIX session status
- `POST /admin/fix/connect` - Connect FIX session
- `POST /admin/fix/disconnect` - Disconnect FIX session
- `POST /admin/fix/subscribe` - Subscribe to symbol

---

## 6. Current Configuration

### Backend
- **Port:** 7999
- **Process ID:** 11600 (check with `netstat -ano | findstr 7999`)
- **Execution Mode:** BBOOK (internal execution)
- **Market Data:** Simulation (LP=SIM)
- **Symbols:** 30 active (Forex + Metals + Crypto)

### Frontend
- **API URL:** http://localhost:7999 (default)
- **WebSocket:** ws://localhost:7999/ws
- **Dev Port:** 3000 (when running `npm run dev`)

### FIX Configuration
- **Config:** backend/fix/config/yofx1_session.cfg
- **Server:** 23.106.238.138:12336
- **Protocol:** FIX 4.4
- **Sessions:** YOFX1 (Trading), YOFX2 (Market Data)
- **Status:** Configured but disconnected

---

## 7. Troubleshooting

### Backend Not Running?
```bash
cd backend
go run cmd/server/main.go
# Should start on port 7999
```

### Port 7999 Already in Use?
```bash
# Find process using port
netstat -ano | findstr 7999

# Kill process (replace PID)
taskkill /PID <process_id> /F
```

### WebSocket Not Connecting?
1. Check backend is running: `curl http://localhost:7999/health`
2. Verify port: `netstat -ano | findstr 7999`
3. Check browser console for errors
4. Open test page: `backend/test_websocket.html`

### No Tick Data?
1. Check diagnostics: `curl http://localhost:7999/api/diagnostics/market-data`
2. Verify subscriptions: `curl http://localhost:7999/api/symbols/subscribed`
3. Should see 29 symbols and 50,000+ ticks
4. Simulation runs automatically (LP=SIM)

### FIX Sessions Won't Connect?
1. Check network: `Test-NetConnection 23.106.238.138 -Port 12336`
2. Verify credentials in `backend/fix/config/yofx1_session.cfg`
3. Check sequence numbers in `backend/fixstore/*.seqnums`
4. Review backend logs for FIX errors

---

## 8. Development Workflow

### Typical Development Session
```bash
# 1. Start backend (if not running)
cd backend
go run cmd/server/main.go

# 2. Start frontend
cd ../clients/desktop
npm run dev

# 3. Open browser
start http://localhost:3000

# 4. Monitor backend
curl http://localhost:7999/api/diagnostics/market-data
```

### Testing Changes
```bash
# Test specific endpoint
curl http://localhost:7999/api/config

# Check WebSocket
start backend/test_websocket.html

# Run full E2E tests
cd backend
powershell -ExecutionPolicy Bypass -File test_e2e.ps1
```

---

## 9. Data Flow Diagram

```
FIX Gateway (Disconnected)
        │
        ├─ YOFX1: Trading Account
        └─ YOFX2: Market Data Feed
                │
                ▼
        ┌───────────────┐
        │  Simulation   │◄── Active (when FIX disconnected)
        │    Engine     │
        └───────────────┘
                │
                │ Generates 30 symbols @ 200ms
                ▼
        ┌───────────────┐
        │  LP Manager   │
        │  (Aggregates) │
        └───────────────┘
                │
                ▼
        ┌───────────────┐
        │ WebSocket Hub │
        │ Port: 7999    │
        └───────────────┘
                │
                ▼
        ws://localhost:7999/ws
                │
                ▼
        ┌───────────────┐
        │ Desktop Client│
        │ Port: 3000    │
        └───────────────┘
```

---

## 10. Quick Commands Cheatsheet

```bash
# Backend
curl http://localhost:7999/health                    # Health
curl http://localhost:7999/api/config                # Config
curl http://localhost:7999/admin/fix/ticks           # Ticks
curl http://localhost:7999/api/diagnostics/market-data  # Diagnostics

# Frontend
cd clients/desktop && npm run dev                    # Start dev server
start http://localhost:3000                          # Open browser

# Testing
start backend/test_websocket.html                    # WebSocket test
powershell -File backend/test_e2e.ps1                # Full E2E tests

# FIX Management
curl -X POST http://localhost:7999/admin/fix/connect -d '{"sessionId":"YOFX1"}'  # Connect
curl http://localhost:7999/admin/fix/status          # Status
```

---

## 11. Test Results Summary

✅ **Backend Server:** Running on port 7999
✅ **WebSocket:** Broadcasting at ws://localhost:7999/ws
✅ **Market Data:** 51,457+ ticks, 30 symbols
✅ **API Endpoints:** All 6 tested endpoints responding
✅ **Data Flow:** Complete pipeline verified
⚠️ **FIX Sessions:** Disconnected (simulation active)
ℹ️ **Environment:** Using defaults (no .env needed)

**System Status:** FULLY OPERATIONAL for development and testing.

---

## Support Files

- **E2E Test Report:** `docs/E2E_TEST_REPORT.md`
- **Test Script:** `backend/test_e2e.ps1`
- **WebSocket Test:** `backend/test_websocket.html`
- **FIX Config:** `backend/fix/config/yofx1_session.cfg`

---

**Last Updated:** 2026-01-30
**System Version:** Trading Engine v3.0
