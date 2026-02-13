# Quick Start Guide - Trading Engine Desktop Client

## TL;DR - Get Started in 30 Seconds

```bash
# From project root directory
START.bat
```

**That's it!** The script will:
1. Check dependencies (Node.js, npm, Go)
2. Start backend on port 7999
3. Start desktop client on port 5173
4. Start admin panels on ports 3000 and 3001

**Login:** http://localhost:5173
- Username: `admin`
- Password: `Admin@123`

---

## What Was Fixed?

### Before (Broken) ❌
- API calls failed (wrong URLs: `localhost:8080` instead of `localhost:7999`)
- No authentication headers on requests
- WebSocket connections failed
- TypeScript compilation errors
- No centralized configuration

### After (Working) ✅
- All API calls use correct URLs (`http://localhost:7999`)
- JWT authentication on all requests
- WebSocket with token auth + auto-reconnect
- Zero TypeScript errors
- Centralized API configuration

---

## Architecture Overview

```
┌─────────────────────────────────────────────────┐
│  Desktop Client (React + TypeScript)            │
│  http://localhost:5173                          │
│                                                  │
│  - Login / Authentication                       │
│  - Market Data Display                          │
│  - Order Entry / Position Management            │
│  - Real-time WebSocket Updates                  │
└─────────────┬───────────────────────────────────┘
              │
              │ HTTP + WebSocket
              │
┌─────────────▼───────────────────────────────────┐
│  Backend API (Go + QuickFIX)                    │
│  http://localhost:7999                          │
│  ws://localhost:7999/ws                         │
│                                                  │
│  - REST API (400+ endpoints)                    │
│  - WebSocket Hub (real-time ticks)              │
│  - Authentication (JWT tokens)                  │
│  - Order Management System (OMS)                │
│  - Risk Engine                                  │
│  - FIX Gateway (LP connectivity)                │
└─────────────────────────────────────────────────┘
```

---

## What Works Now

### Core Trading Features ✅
- [x] User login with JWT authentication
- [x] Real-time market data (bid/ask prices)
- [x] Symbol list (EURUSD, GBPUSD, USDJPY, etc.)
- [x] Market order execution (Buy/Sell)
- [x] Position management (open/close)
- [x] Account information (balance, equity, margin)
- [x] WebSocket streaming (tick updates)

### Admin Features ✅
- [x] Broker Admin dashboard (port 3000)
- [x] Super Admin dashboard (port 3001)
- [x] Symbol management
- [x] Account management
- [x] Execution mode control (A-Book/B-Book)

### Technical Features ✅
- [x] TypeScript type safety (zero compilation errors)
- [x] Automatic token refresh on requests
- [x] 401 Unauthorized handling (auto-logout)
- [x] WebSocket reconnection with exponential backoff
- [x] Centralized API endpoint configuration
- [x] Environment variable support

---

## What Uses Mock Data (Non-Critical)

Some advanced features display placeholder data (see `docs/NON_FUNCTIONAL_FEATURES.md` for details):

**UI-Only Features (29 components):**
- Economic Calendar (shows placeholder events)
- News Impact Analyzer (random sentiment)
- Copy Trading Leaderboard (fake stats)
- Advanced Order Types (OCO, Trailing Stop) - UI exists, backend TODO
- Market Replay - uses simulated data

**Impact:** These don't affect core trading. You can still:
- Login and authenticate ✅
- View real market data ✅
- Execute orders ✅
- Manage positions ✅
- View account balance ✅

---

## File Changes Summary

### New Files Created (1)
- `clients/desktop/src/config/api.ts` - Centralized API configuration

### Files Enhanced (2)
- `clients/desktop/src/services/api.ts` - JWT auth + error handling
- `clients/desktop/src/services/websocket-enhanced.ts` - Token auth + reconnection

### Files Updated (28)
- All components now use centralized `API_ENDPOINTS`
- All stores updated to use correct URLs
- WebSocket connections use token authentication

### Total: 31 Files Modified

---

## Testing Checklist

### Basic Functionality Test (5 minutes)

1. **Start the application:**
   ```bash
   START.bat
   ```

2. **Open desktop client:**
   - Navigate to: http://localhost:5173
   - Enter credentials: `admin` / `Admin@123`
   - Click "Login"
   - ✅ Expected: Dashboard loads

3. **Check market data:**
   - Look at Market Watch panel (left side)
   - ✅ Expected: Symbol list loads (EURUSD, GBPUSD, etc.)
   - ✅ Expected: Bid/Ask prices update every second

4. **Open DevTools (F12):**
   - Go to Console tab
   - ✅ Expected: `[WS] Connected successfully`
   - ✅ Expected: No red error messages
   - Go to Network tab
   - ✅ Expected: Requests to `localhost:7999`
   - ✅ Expected: Response status 200 (not 404 or CORS errors)

5. **Test order entry (optional):**
   - Click on a symbol (e.g., EURUSD)
   - Open Order Entry panel
   - Enter volume: `0.01`
   - Click "Buy Market"
   - ✅ Expected: Position appears in Positions panel

### Backend Health Check
```bash
curl http://localhost:7999/health
# Expected: {"status":"ok"}
```

---

## Troubleshooting

### Problem: "Cannot start backend - Go not installed"

**Solution 1:** Use pre-built executable (already exists)
```bash
cd backend
.\server.exe
```

**Solution 2:** Install Go
- Download from: https://go.dev/dl/
- Install and restart terminal
- Re-run `START.bat`

---

### Problem: "Port 7999 already in use"

**Solution:** Kill existing process
```bash
# Windows
netstat -ano | findstr :7999
taskkill /PID <PID> /F

# Or just re-run START.bat (it kills old processes automatically)
```

---

### Problem: Login fails with "Network Error"

**Checklist:**
- [ ] Backend is running: Check backend window for logs
- [ ] Port 7999 is accessible: `curl http://localhost:7999/health`
- [ ] No firewall blocking: Temporarily disable and retry
- [ ] Check backend logs: Look for errors in backend console

---

### Problem: WebSocket won't connect

**Checklist:**
- [ ] Backend WebSocket is running
- [ ] Token exists: Open DevTools → Application → Local Storage → Check `rtx_token`
- [ ] Check browser console: Look for `[WS]` messages
- [ ] Try clearing cache: `Ctrl + Shift + R`

---

### Problem: TypeScript errors in IDE

**Solution:**
```bash
cd clients/desktop
npm install
# VS Code: Ctrl+Shift+P → "TypeScript: Restart TS Server"
```

---

## Environment Variables (Optional)

Create `clients/desktop/.env`:

```env
# Override backend URL (default: http://localhost:7999)
VITE_API_URL=http://your-backend-url:7999

# Override WebSocket URL (default: ws://localhost:7999/ws)
VITE_WS_URL=ws://your-backend-url:7999/ws
```

**Note:** Environment variables require restart of Vite dev server.

---

## Key Files Reference

### Configuration
- **API Config:** `clients/desktop/src/config/api.ts`
- **Backend Config:** `backend/config.json` (auto-generated)
- **Vite Config:** `clients/desktop/vite.config.ts`

### Services
- **API Service:** `clients/desktop/src/services/api.ts`
- **WebSocket Service:** `clients/desktop/src/services/websocket-enhanced.ts`

### Main App
- **App Entry:** `clients/desktop/src/App.tsx`
- **Login Component:** `clients/desktop/src/components/Login.tsx`

### Backend
- **Main Entry:** `backend/cmd/server/main.go`
- **Executable:** `backend/server.exe`

---

## URL Reference

| Service | URL | Purpose |
|---------|-----|---------|
| Desktop Client | http://localhost:5173 | Main trading terminal |
| Backend API | http://localhost:7999 | REST API endpoints |
| WebSocket | ws://localhost:7999/ws | Real-time market data |
| Backend Health | http://localhost:7999/health | Health check |
| Broker Admin | http://localhost:3000 | Broker management |
| Super Admin | http://localhost:3001 | Platform control |

---

## Login Credentials

**Desktop Client:**
- Username: `admin`
- Password: `Admin@123`

**Alternative account:**
- Username: `trader`
- Password: `password`

---

## Next Steps

1. **Start the app:** Run `START.bat`
2. **Test core features:** Login, market data, orders
3. **Review detailed report:** See `DESKTOP_CLIENT_FIX_VERIFICATION_REPORT.md`
4. **Implement missing features:** Check `docs/NON_FUNCTIONAL_FEATURES.md` for TODOs

---

## Support

**Detailed Documentation:**
- Full verification report: `DESKTOP_CLIENT_FIX_VERIFICATION_REPORT.md`
- Non-functional features: `docs/NON_FUNCTIONAL_FEATURES.md`
- Backend documentation: `backend/README_BUILD.md`
- Project README: `README.md`

**Health Checks:**
```bash
# Backend
curl http://localhost:7999/health

# Frontend (check if Vite is running)
curl http://localhost:5173
```

---

**Last Updated:** 2026-02-12
**Status:** ✅ Ready for Testing
