# Backend Quick Start Guide

## Start the Backend (30 seconds)

### Windows
```batch
cd C:\Users\Yofor\Desktop\Trading-Engine\backend
start-backend.bat
```

### Linux/Mac
```bash
cd backend
./server.exe
```

---

## Verify It's Running

### Check Port
```bash
netstat -ano | findstr :7999
```

### Test Health Endpoint
```bash
curl http://localhost:7999/health
```

### Expected Output
```
2026/02/13 13:19:34   SERVER READY - B-BOOK TRADING ENGINE
2026/02/13 13:19:34 Starting server on port :7999
```

---

## Common Issues & Fixes

### ❌ "Multiple main() declaration"
**Fix:** Test files in wrong location
```bash
cd backend
mkdir -p tests/network
mv test_*.go tests/network/
go build -o server.exe ./cmd/server
```

### ❌ "panic: invalid argument to Intn"
**Fix:** Already fixed in `admin/paper_trading.go`
```bash
go build -o server.exe ./cmd/server
```

### ❌ "Port 7999 already in use"
**Fix:** Kill old process
```bash
# Find PID
netstat -ano | findstr :7999 | findstr LISTENING

# Kill it (replace <PID> with actual PID)
taskkill //F //PID <PID>
```

### ❌ ".env file not found"
**Fix:** Copy from example
```bash
cd backend
copy .env.example .env
# Edit .env with your configuration
```

---

## Configuration

### Required Settings (.env)
```bash
PORT=7999
ENVIRONMENT=development
DB_HOST=localhost
DB_PORT=5432
DB_NAME=trading_engine
```

### Default Login
- **Username:** admin
- **Password:** password
- ⚠️ Change in production!

---

## Frontend Connection

Update frontend `.env`:
```bash
VITE_API_URL=http://localhost:7999
VITE_WS_URL=ws://localhost:7999/ws
```

---

## Troubleshooting

### Check Logs
```bash
# Real-time logs
tail -f backend.log

# Last 50 lines
tail -n 50 backend.log
```

### Rebuild from Source
```bash
cd backend
go mod download
go build -o server.exe ./cmd/server
./server.exe
```

### Full Reset
```bash
# Kill server
taskkill //F //IM server.exe

# Clean build
cd backend
del server.exe
go clean -cache
go build -o server.exe ./cmd/server

# Restart
start-backend.bat
```

---

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /health` | Health check |
| `POST /api/auth/login` | Login |
| `GET /api/symbols` | List symbols |
| `GET /api/accounts` | List accounts |
| `POST /api/orders` | Place order |
| `GET /ws` | WebSocket connection |

---

## Performance

- **Startup Time:** ~5 seconds
- **Memory Usage:** ~500MB
- **Port:** 7999
- **WebSocket:** Enabled
- **Concurrent Users:** 1000+

---

**For full details, see:** `BACKEND_FIX_REPORT.md`
