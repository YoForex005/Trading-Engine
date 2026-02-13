# Quick Start - Backend Server

**Status:** ✅ Backend Operational
**Port:** 7999
**Time to Start:** < 30 seconds

---

## One-Command Startup

### Full Platform (Backend + Frontend)
```batch
START.bat
```
**Opens:** Backend + Desktop + Admin dashboards

---

### Backend Only
```batch
cd backend
server.exe
```

---

## Quick Troubleshooting

### Port Already in Use
```batch
# Kill process on port 7999
for /f "tokens=5" %a in ('netstat -ano ^| findstr :7999') do taskkill /F /PID %a

# Restart
server.exe
```

---

### Database Not Connected
```batch
# Start PostgreSQL
net start postgresql-x64-14

# Test connection
psql -U trading -d trading_engine
```

---

### Redis Not Connected
```batch
# Start Redis
net start Redis

# Test
redis-cli ping
```

---

### Server Won't Start
```batch
# Check .env exists
dir backend\.env

# Check Go version
go version

# Rebuild if needed
cd backend
go build -o server.exe cmd/server/main.go
server.exe
```

---

## Essential Checks

### ✅ Is Backend Running?
```batch
curl http://localhost:7999/health
```
**Expected:** `{"status":"healthy"}`

---

### ✅ Is Port Listening?
```batch
netstat -ano | findstr :7999
```
**Expected:** Line showing `LISTENING`

---

### ✅ Can Login?
```batch
curl -X POST http://localhost:7999/api/admin/auth/login ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"admin\",\"password\":\"Admin@123\"}"
```
**Expected:** JSON with token

---

## Emergency Fixes

### Clean Restart
```batch
# Kill everything
taskkill /F /IM server.exe

# Clear port
for /f "tokens=5" %a in ('netstat -ano ^| findstr :7999') do taskkill /F /PID %a

# Restart services
net start postgresql-x64-14
net start Redis

# Start backend
cd backend
server.exe
```

---

### Rebuild from Source
```batch
cd backend
go mod tidy
go mod download
go build -o server.exe cmd/server/main.go
server.exe
```

---

### Configuration Reset
```batch
# Backup current .env
copy backend\.env backend\.env.backup

# Copy example
copy backend\.env.example backend\.env

# Edit critical settings
notepad backend\.env

# Required minimum:
# PORT=7999
# DB_HOST=localhost
# DB_NAME=trading_engine
# DB_USER=trading
# DB_PASSWORD=trading_pass
# JWT_SECRET=your-secure-secret-here-min-32-bytes
```

---

## Success Indicators

### ✅ Backend Started Successfully

Look for these log messages:

```
[INIT] Loaded .env file
[GC] Set GOGC=50 for more frequent garbage collection
[ENVIRONMENT] Running in development mode
[DATABASE] Connected to PostgreSQL: localhost:5432/trading_engine
[REDIS] Connected to Redis: localhost:6379
[HTTP] Server listening on port 7999
[WEBSOCKET] WebSocket endpoint ready
```

**Server is ready when you see:**
```
Server listening on port 7999
```

---

### ✅ All Services Healthy

```batch
# Health check
curl http://localhost:7999/health

# Database
psql -U trading -d trading_engine -c "SELECT 1;"

# Redis
redis-cli ping

# Ports active
netstat -ano | findstr "7999 5173 3000"
```

---

## Common Errors Quick Fix

| Error | Fix |
|-------|-----|
| `Port already in use` | `for /f "tokens=5" %a in ('netstat -ano ^| findstr :7999') do taskkill /F /PID %a` |
| `Database connection failed` | `net start postgresql-x64-14` |
| `Redis connection failed` | `net start Redis` |
| `panic: TOTP_ENCRYPTION_KEY` | Add `TOTP_ENCRYPTION_KEY=0123...` to .env or disable 2FA |
| `Go not found` | Use pre-built `server.exe` |
| `.env not found` | `copy .env.example .env` |

---

## Quick Access URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| **Health Check** | http://localhost:7999/health | None |
| **Admin API** | http://localhost:7999/api/admin | admin / Admin@123 |
| **WebSocket** | ws://localhost:7999/ws | Token required |
| **Desktop Client** | http://localhost:5173 | admin / Admin@123 |
| **Admin Dashboard** | http://localhost:5174 | admin / Admin@123 |

---

## Monitoring Commands

```batch
# Real-time logs
Get-Content backend.log -Tail 50 -Wait

# Monitor errors only
Get-Content backend.log -Wait | Select-String "ERROR|FATAL|panic"

# Check resource usage
tasklist /FI "IMAGENAME eq server.exe" /FO TABLE

# Watch port activity
netstat -ano | findstr :7999

# Database connections
psql -U trading -d trading_engine -c "SELECT count(*) FROM pg_stat_activity;"
```

---

## Auto-Restart Monitor

### Enable Monitoring
```batch
cd backend
start-monitor.bat
```

**Features:**
- Auto-restarts if server crashes
- Monitors port 7999 every 30 seconds
- Logs to `keep-alive.log`

**To stop:** Close "Trading Engine Backend Monitor" window

---

## Critical Environment Variables

**Minimum required in `.env`:**

```env
PORT=7999
DB_HOST=localhost
DB_NAME=trading_engine
DB_USER=trading
DB_PASSWORD=trading_pass
JWT_SECRET=your-secure-secret-here-min-32-bytes
BROKER_NAME=RTX Trading
EXECUTION_MODE=BBOOK
```

**Check current values:**
```batch
findstr "PORT= DB_HOST= JWT_SECRET=" backend\.env
```

---

## Testing After Startup

### 1. Health Check
```batch
curl http://localhost:7999/health
```
**Expected:** `{"status":"healthy",...}`

### 2. Login Test
```batch
curl -X POST http://localhost:7999/api/admin/auth/login ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"admin\",\"password\":\"Admin@123\"}"
```
**Expected:** Token in response

### 3. WebSocket Test (Browser Console)
```javascript
const ws = new WebSocket('ws://localhost:7999/ws');
ws.onopen = () => console.log('✓ Connected');
ws.onmessage = (e) => console.log('Message:', e.data);
```
**Expected:** "✓ Connected"

### 4. Database Test
```batch
psql -U trading -d trading_engine -c "SELECT version();"
```
**Expected:** PostgreSQL version info

### 5. Redis Test
```batch
redis-cli PING
```
**Expected:** `PONG`

---

## Next Steps After Backend Starts

### 1. Start Frontend
```batch
# Desktop client
cd clients\desktop
npm run dev
# Opens: http://localhost:5173

# Admin dashboard
cd clients\admin-dashboard
npm run dev
# Opens: http://localhost:5174
```

### 2. Access Platform
- Open browser: http://localhost:5173
- Login: admin / Admin@123
- Start trading

### 3. Check Logs
```batch
# View backend logs
Get-Content backend.log -Tail 50

# Monitor errors
Get-Content backend.log -Wait | Select-String ERROR
```

---

## Stopping Backend

### Graceful Stop
```batch
# Press Ctrl+C in server window
# Or close the server window
```

### Force Stop
```batch
taskkill /F /IM server.exe
```

### Stop All Services
```batch
# If started with START.bat
# Press any key in launcher window

# Or manually
taskkill /F /IM server.exe
taskkill /F /IM node.exe
net stop postgresql-x64-14
net stop Redis
```

---

## Quick Reference Card

```
┌─────────────────────────────────────────────────────┐
│ RTX BACKEND - QUICK REFERENCE                       │
├─────────────────────────────────────────────────────┤
│ START:      cd backend && server.exe                │
│ FULL STACK: START.bat                               │
│ MONITOR:    start-monitor.bat                       │
│                                                      │
│ HEALTH:     curl localhost:7999/health              │
│ STOP:       taskkill /F /IM server.exe              │
│                                                      │
│ PORT:       7999                                     │
│ LOGIN:      admin / Admin@123                       │
│                                                      │
│ LOGS:       Get-Content backend.log -Tail 50 -Wait  │
│ DATABASE:   psql -U trading -d trading_engine       │
│ REDIS:      redis-cli ping                          │
│                                                      │
│ DOCS:       docs/BACKEND_STARTUP_GUIDE.md           │
│ ISSUES:     BACKEND_STARTUP_ISSUE_REPORT.md         │
└─────────────────────────────────────────────────────┘
```

---

## Help & Support

- **Full Guide:** `docs/BACKEND_STARTUP_GUIDE.md`
- **Issue Report:** `BACKEND_STARTUP_ISSUE_REPORT.md`
- **Production Status:** `backend/PRODUCTION_STATUS.md`
- **Configuration:** `backend/docs/CONFIGURATION_GUIDE.md`

---

## System Status

✅ **Backend:** Fully operational
✅ **Build:** Zero errors
✅ **Database:** PostgreSQL ready
✅ **Cache:** Redis ready
✅ **Monitoring:** Available
✅ **Documentation:** Complete

**Ready to start trading!**

---

**Quick Start Guide - RTX Trading Engine Backend**
**Last Updated:** 2026-02-13
