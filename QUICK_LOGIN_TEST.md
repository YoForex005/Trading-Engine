# Quick Login Test Procedure

**5-Minute Test Guide for RTX Trading Engine Login**

---

## 🚀 Quick Start (Windows)

### 1. Start Everything

Double-click `START.bat` in project root.

**Wait 10-15 seconds** while services start.

### 2. Verify Backend

Open browser: http://localhost:7999/health

**Should see:**
```json
{"status":"healthy","timestamp":"..."}
```

### 3. Open Desktop Client

Navigate to: http://localhost:5173

### 4. Login

**Enter these credentials:**
- **Account ID:** `1`
- **Password:** `password`
- **Server:** `localhost:7999` (pre-filled)

Click **"Connect to Gate"**

### 5. Success!

You should see:
- Trading terminal interface
- Account balance: $10,000
- Symbol list with prices
- Live market data streaming

---

## 📋 Manual Start (If START.bat Doesn't Work)

### Option 1: Using Pre-built Backend

```bash
# Terminal 1: Start Backend
cd backend
server.exe
```

```bash
# Terminal 2: Start Frontend
cd clients\desktop
npm run dev
```

### Option 2: Build from Source

```bash
# Terminal 1: Build & Start Backend
cd backend
go build -o server.exe cmd\server\main.go
server.exe
```

```bash
# Terminal 2: Start Frontend
cd clients\desktop
npm install
npm run dev
```

---

## ✅ Expected Results

### Backend Console (Terminal 1)

```
╔═══════════════════════════════════════════════════════════╗
║          RTX Trading - Backend v3.0                       ║
║        BBOOK Mode + OANDA LP                              ║
╚═══════════════════════════════════════════════════════════╝

[B-Book] Demo account created: demo_001 with $10000.00
[AdminAuth] Default super admin created: admin
[WebSocket] Hub started
[HTTP] Server listening on :7999
```

### Frontend Terminal (Terminal 2)

```
  VITE v7.2.4  ready in 523 ms

  ➜  Local:   http://localhost:5173/
  ➜  Network: use --host to expose
  ➜  press h + enter to show help
```

### Browser After Login

- **URL:** http://localhost:5173 (main terminal)
- **Console:** No errors (F12 → Console tab)
- **Network:** WebSocket connected (F12 → Network → WS tab)
- **UI:**
  - Top bar shows account balance
  - Left panel shows symbol list
  - Chart displays
  - Order panel active

---

## 🧪 CLI Test (Alternative)

### Test Login Endpoint Directly

```bash
curl -X POST http://localhost:7999/login ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"1\",\"password\":\"password\"}"
```

**Expected Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "1",
    "username": "demo-user",
    "role": "TRADER"
  }
}
```

### Test Health Check

```bash
curl http://localhost:7999/health
```

**Expected:**
```json
{"status":"healthy","timestamp":"2026-02-13T..."}
```

---

## ❌ Troubleshooting Quick Fixes

### Problem: Backend Won't Start

**Fix:**
```bash
# Kill existing backend
taskkill /F /IM server.exe

# Check port is free
netstat -ano | findstr :7999

# Restart backend
cd backend
server.exe
```

### Problem: Frontend Shows "Server Not Responding"

**Check 1:** Is backend running?
```bash
curl http://localhost:7999/health
```

**Check 2:** Is server URL correct?
- Should be: `localhost:7999`
- NOT: `http://localhost:7999` (remove http://)

**Check 3:** Wait longer (backend might still be starting)

### Problem: "Invalid Credentials"

**Try these credentials:**

| Type | Username | Password |
|------|----------|----------|
| Client | `1` | `password` |
| Admin | `admin` | `Admin@123` |

### Problem: CORS Error in Browser

**Check backend/.env has:**
```env
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000,http://localhost:3001
```

**Then restart backend.**

### Problem: Timeout Error

**Causes:**
- Backend not running
- Port blocked
- Firewall issue

**Fix:**
```bash
# 1. Verify backend health
curl http://localhost:7999/health

# 2. Check firewall allows port 7999

# 3. Restart both services
```

---

## 🔑 Test Credentials Reference

### Desktop Client Login

```
Account ID: 1
Password: password
Server: localhost:7999
```

### Admin Panel Login

```
Username: admin
Password: Admin@123
URL: http://localhost:3000 (if broker-admin is running)
```

---

## 🌐 Service URLs

| Service | URL | Port |
|---------|-----|------|
| Backend API | http://localhost:7999 | 7999 |
| Desktop Client | http://localhost:5173 | 5173 |
| Broker Admin | http://localhost:3000 | 3000 |
| Super Admin | http://localhost:3001 | 3001 |
| WebSocket | ws://localhost:7999/ws | 7999 |
| Health Check | http://localhost:7999/health | 7999 |

---

## 📊 Success Indicators

### Backend Healthy

- ✅ Console shows "Server listening on :7999"
- ✅ /health endpoint returns 200 OK
- ✅ Demo account created log appears
- ✅ No panic/crash errors

### Frontend Healthy

- ✅ Vite dev server running
- ✅ No TypeScript errors
- ✅ Page loads at localhost:5173
- ✅ Login form displays

### Login Successful

- ✅ No timeout error
- ✅ No "invalid credentials" message
- ✅ Browser redirects to terminal
- ✅ Token in localStorage (DevTools → Application → Local Storage)
- ✅ WebSocket connected (DevTools → Network → WS)
- ✅ Market data streaming

---

## 🛑 Stop Services

### Graceful Shutdown

Press `Ctrl+C` in each terminal window.

### Force Stop All

```bash
# Windows
taskkill /F /IM server.exe
taskkill /F /IM node.exe

# Or use START.bat's built-in shutdown
# (Press any key when prompted)
```

---

## 📞 Need Help?

### Check Detailed Guide

See `docs/LOGIN_TERMINAL_TROUBLESHOOTING.md` for:
- Complete login flow diagram
- Step-by-step debugging
- Common issues and solutions
- Environment setup
- Database configuration

### Check Backend Logs

Look in the backend terminal window for errors.

### Check Browser Console

Press F12 → Console tab to see frontend errors.

### Verify Prerequisites

```bash
node --version   # Should be v16+
npm --version    # Should be v9+
go version       # Should be go1.20+ (or use server.exe)
```

---

## 🎯 One-Line Test

```bash
START.bat && timeout /t 15 && curl http://localhost:7999/health && start http://localhost:5173
```

This will:
1. Start all services
2. Wait 15 seconds
3. Check backend health
4. Open frontend in browser

**Then login with:** Account ID `1`, Password `password`

---

**Last Updated:** February 13, 2026
**Quick Reference Version:** 1.0
