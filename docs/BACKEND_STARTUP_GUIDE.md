# Backend Startup Guide - RTX Trading Engine

**Version:** 1.0
**Last Updated:** 2026-02-13
**Platform:** Windows (WSL/Linux compatible)
**Go Version:** 1.26.0

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Environment Setup](#environment-setup)
3. [Step-by-Step Startup Instructions](#step-by-step-startup-instructions)
4. [Verification & Health Checks](#verification--health-checks)
5. [Common Issues & Troubleshooting](#common-issues--troubleshooting)
6. [Advanced Configuration](#advanced-configuration)
7. [Monitoring & Logging](#monitoring--logging)
8. [Shutdown Procedures](#shutdown-procedures)

---

## Prerequisites

### Required Software

| Software | Version | Purpose | Installation |
|----------|---------|---------|--------------|
| **Go** | 1.20+ | Backend compilation | https://go.dev/dl/ |
| **PostgreSQL** | 12+ | Database | https://www.postgresql.org/download/ |
| **Redis** | 6+ | Cache & sessions | https://redis.io/download |
| **Node.js** | 18+ | Frontend development | https://nodejs.org/ |

### Optional Software

| Software | Version | Purpose |
|----------|---------|---------|
| **curl** | Any | API testing |
| **Postman** | Any | API testing |
| **pgAdmin** | Any | Database management |

---

### Prerequisites Checklist

Before starting the backend, verify:

```batch
# Check Go installation
go version
# Expected: go version go1.26.0 windows/amd64

# Check Node.js
node --version
# Expected: v18.x.x or higher

# Check PostgreSQL
psql --version
# Expected: psql (PostgreSQL) 12.x or higher

# Check Redis (Windows)
redis-server --version
# Expected: Redis server v6.x.x or higher

# Check if backend executable exists
dir C:\Users\Yofor\Desktop\Trading-Engine\backend\server.exe
# Expected: File exists (if previously built)
```

**If any are missing, install them before proceeding.**

---

## Environment Setup

### Step 1: Configure Database

#### PostgreSQL Setup

```batch
# Start PostgreSQL service
net start postgresql-x64-14

# Create database and user
psql -U postgres
```

```sql
-- Create database
CREATE DATABASE trading_engine;

-- Create user
CREATE USER trading WITH PASSWORD 'trading_pass';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE trading_engine TO trading;

-- Connect to database
\c trading_engine

-- Verify connection
\dt
```

**Exit psql:** Type `\q` and press Enter

---

### Step 2: Configure Redis

```batch
# Start Redis service
net start Redis

# Test connection
redis-cli ping
# Expected: PONG

# Check Redis is listening
netstat -ano | findstr :6379
# Expected: Line showing LISTENING on port 6379
```

---

### Step 3: Configure Environment Variables

#### Create .env File

Navigate to backend directory:
```batch
cd C:\Users\Yofor\Desktop\Trading-Engine\backend
```

The `.env` file should already exist. Verify critical settings:

```env
# ===========================================
# SERVER CONFIGURATION
# ===========================================
PORT=7999
ENVIRONMENT=development

# ===========================================
# DATABASE (PostgreSQL)
# ===========================================
DB_HOST=localhost
DB_PORT=5432
DB_NAME=trading_engine
DB_USER=trading
DB_PASSWORD=trading_pass
DB_SSL_MODE=disable

# ===========================================
# REDIS CACHE
# ===========================================
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=rtx_redis
REDIS_MAX_RETRIES=3
REDIS_POOL_SIZE=10

# ===========================================
# JWT AUTHENTICATION
# ===========================================
JWT_SECRET=your-secure-secret-here-min-32-bytes
JWT_EXPIRY=24h

# ===========================================
# BROKER SETTINGS
# ===========================================
BROKER_NAME=RTX Trading
BROKER_DISPLAY_NAME=YoForex
EXECUTION_MODE=BBOOK
DEFAULT_LEVERAGE=100
DEFAULT_BALANCE=5000.0
MARGIN_MODE=HEDGING

# ===========================================
# CORS
# ===========================================
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000,http://localhost:3001,http://localhost:3002
```

#### Validate Environment

```batch
# Check if .env exists
dir .env
# Expected: File exists

# Verify critical variables are set
findstr "PORT=" .env
findstr "DB_NAME=" .env
findstr "JWT_SECRET=" .env
```

---

### Step 4: Database Migration (Optional)

If using migration system:

```batch
cd C:\Users\Yofor\Desktop\Trading-Engine\backend

# Initialize schema
go run cmd/migrate/main.go -init

# Run migrations
go run cmd/migrate/main.go -up

# Check migration status
go run cmd/migrate/main.go -status
```

**Note:** Skip if database schema already exists.

---

## Step-by-Step Startup Instructions

### Method 1: One-Click Startup (Recommended for Full Stack)

```batch
# From project root
cd C:\Users\Yofor\Desktop\Trading-Engine
START.bat
```

**What this does:**
1. Checks prerequisites (Go, Node.js, npm)
2. Kills existing processes on ports 7999, 5173, 3000, 3001
3. Builds backend or uses pre-built `server.exe`
4. Installs frontend dependencies (if missing)
5. Starts all services in separate windows:
   - Backend API (port 7999)
   - Desktop Client (port 5173)
   - Broker Admin (port 3000)
   - Super Admin (port 3001)
6. Verifies all services are healthy

**Windows opened:**
- `TRADING-ENGINE-BACKEND` - Backend server
- `TRADING-ENGINE-DESKTOP` - Desktop client
- `TRADING-ENGINE-BROKER-ADMIN` - Broker admin
- `TRADING-ENGINE-SUPER-ADMIN` - Super admin

**To stop:** Press any key in the main launcher window.

---

### Method 2: Backend Only (Development)

```batch
# Navigate to backend
cd C:\Users\Yofor\Desktop\Trading-Engine\backend

# Start server directly
server.exe
```

**Expected output:**
```
[INIT] Loaded .env file. YOFX_PROXY_HOST=81.29.145.69
[GC] Set GOGC=50 for more frequent garbage collection
[GC] Set GOMEMLIMIT=2GiB to prevent OOM crashes
[ENVIRONMENT] Running in development mode
[DATABASE] Connected to PostgreSQL: localhost:5432/trading_engine
[REDIS] Connected to Redis: localhost:6379
[FIX] Initialized YOFX1 session
[FIX] Initialized YOFX2 session
[HTTP] Server listening on port 7999
[WEBSOCKET] WebSocket endpoint ready at ws://localhost:7999/ws
✓ Backend started successfully
```

**Server is ready when you see:** `Server listening on port 7999`

---

### Method 3: With Auto-Restart Monitoring

```batch
# Navigate to backend
cd C:\Users\Yofor\Desktop\Trading-Engine\backend

# Start with monitoring
start-monitor.bat
```

**Features:**
- Automatically restarts server if it crashes
- Monitors port 7999 every 30 seconds
- Logs restart events to `keep-alive.log`
- Tracks server PID for process management
- Runs in minimized background window

**To stop:** Close the "Trading Engine Backend Monitor" window

**View logs:**
```batch
type keep-alive.log
```

---

### Method 4: Build from Source

```batch
# Navigate to backend
cd C:\Users\Yofor\Desktop\Trading-Engine\backend

# Download dependencies
go mod download
go mod verify

# Build executable
go build -o server.exe cmd/server/main.go

# Verify build
dir server.exe
# Expected: File exists (~30-40 MB)

# Run server
server.exe
```

**Build time:** 10-15 seconds (first time), 3-5 seconds (incremental)

---

## Verification & Health Checks

### 1. Check Server is Running

```batch
# Method 1: Check port
netstat -ano | findstr :7999
# Expected: Line showing LISTENING on 0.0.0.0:7999 or [::]:7999

# Method 2: Check process
tasklist | findstr server.exe
# Expected: server.exe     [PID]    Console    1    [Memory] K
```

---

### 2. Health Check Endpoint

```batch
# Using curl
curl http://localhost:7999/health

# Expected response:
{
  "status": "healthy",
  "timestamp": "2026-02-13T12:00:00Z",
  "version": "1.0.0",
  "uptime": "120s"
}
```

**In browser:** Navigate to http://localhost:7999/health

---

### 3. Test API Endpoints

#### Admin Login
```batch
curl -X POST http://localhost:7999/api/admin/auth/login ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"admin\",\"password\":\"Admin@123\"}"

# Expected: JSON with token
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "admin": {
    "id": 1,
    "username": "admin",
    "role": "SuperAdmin"
  }
}
```

#### List Symbols
```batch
curl http://localhost:7999/api/symbols

# Expected: Array of trading symbols
[
  {"symbol": "EURUSD", "bid": 1.0850, "ask": 1.0852},
  {"symbol": "GBPUSD", "bid": 1.2650, "ask": 1.2652},
  ...
]
```

---

### 4. WebSocket Connection Test

```batch
# Using curl
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" ^
  -H "Sec-WebSocket-Version: 13" ^
  -H "Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==" ^
  http://localhost:7999/ws

# Expected: HTTP 101 Switching Protocols
```

**Alternative:** Use browser console:
```javascript
const ws = new WebSocket('ws://localhost:7999/ws');
ws.onopen = () => console.log('Connected');
ws.onmessage = (e) => console.log('Message:', e.data);
```

---

### 5. Database Connection Test

```batch
# From backend directory
go run cmd/test_e2e/main.go

# Or connect directly
psql -U trading -d trading_engine -h localhost

# Run query
SELECT COUNT(*) FROM users;
```

---

### 6. Redis Connection Test

```batch
redis-cli

# Test set/get
127.0.0.1:6379> SET test_key "Hello Redis"
127.0.0.1:6379> GET test_key
# Expected: "Hello Redis"

127.0.0.1:6379> PING
# Expected: PONG

# Exit
127.0.0.1:6379> EXIT
```

---

## Common Issues & Troubleshooting

### Issue 1: Port 7999 Already in Use

**Error:**
```
[FATAL] Failed to start HTTP server: listen tcp :7999: bind: address already in use
```

**Solution:**
```batch
# Find process using port 7999
netstat -ano | findstr :7999
# Example output: TCP 0.0.0.0:7999  0.0.0.0:0  LISTENING  12345

# Kill the process (replace 12345 with actual PID)
taskkill /F /PID 12345

# Or kill all server.exe instances
taskkill /F /IM server.exe

# Restart server
server.exe
```

---

### Issue 2: Database Connection Failed

**Error:**
```
[FATAL] Failed to connect to database: dial tcp [::1]:5432: connectex: No connection could be made
```

**Solutions:**

#### Check PostgreSQL Service
```batch
# Check if running
sc query postgresql-x64-14

# If stopped, start it
net start postgresql-x64-14
```

#### Verify Database Exists
```batch
psql -U postgres -l
# Look for trading_engine in the list
```

#### Check Credentials
```batch
# Test connection manually
psql -U trading -d trading_engine -h localhost -W
# Enter password: trading_pass

# If successful, check .env settings
findstr "DB_" backend\.env
```

#### Check .env Settings
```env
DB_HOST=localhost       # NOT 127.0.0.1 if PostgreSQL uses IPv6
DB_PORT=5432           # Default PostgreSQL port
DB_NAME=trading_engine # Must exist
DB_USER=trading        # Must have privileges
DB_PASSWORD=trading_pass
DB_SSL_MODE=disable    # For local development
```

---

### Issue 3: Redis Connection Failed

**Error:**
```
[FATAL] Failed to connect to Redis: dial tcp [::1]:6379: connectex: No connection could be made
```

**Solutions:**

#### Check Redis Service
```batch
# Check if running
sc query Redis

# If stopped, start it
net start Redis

# Test connection
redis-cli ping
# Expected: PONG
```

#### Check Redis Configuration
```batch
# Check Redis config file
type "C:\Program Files\Redis\redis.windows-service.conf"

# Verify bind address (should include localhost)
# bind 127.0.0.1 ::1

# Verify port
# port 6379
```

---

### Issue 4: Panic - TOTP_ENCRYPTION_KEY Required

**Error:**
```
panic: TOTP_ENCRYPTION_KEY environment variable is required for 2FA
```

**Solutions:**

#### Option 1: Disable 2FA (Development Only)
```batch
# Edit .env and set:
# (Comment out or don't enable TOTP features)
```

#### Option 2: Generate Encryption Key
```batch
# Generate 64-character hex key (32 bytes)
# On Windows PowerShell:
-join ((1..32) | ForEach-Object { "{0:x2}" -f (Get-Random -Minimum 0 -Maximum 256) })

# Add to .env:
TOTP_ENCRYPTION_KEY=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
```

---

### Issue 5: Go Not Installed

**Error:**
```
'go' is not recognized as an internal or external command
```

**Solution:**

1. Download Go: https://go.dev/dl/go1.26.0.windows-amd64.msi
2. Run installer (installs to C:\Program Files\Go)
3. Verify installation:
   ```batch
   # Close and reopen terminal
   go version
   # Expected: go version go1.26.0 windows/amd64
   ```
4. If not in PATH, add manually:
   ```batch
   setx PATH "%PATH%;C:\Program Files\Go\bin"
   # Restart terminal
   ```

**Alternative:** Use pre-built `server.exe` (no Go required)
```batch
cd backend
server.exe
```

---

### Issue 6: Build Failures

**Error:**
```
go build: undefined references, import errors, etc.
```

**Solutions:**

#### Clean and Rebuild
```batch
cd backend

# Clean module cache
go clean -modcache

# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify

# Rebuild
go build -o server.exe cmd/server/main.go
```

#### Check for Code Issues
```batch
# Run static analysis
go vet ./...

# Check formatting
gofmt -l .

# Run tests
go test ./...
```

---

### Issue 7: Frontend Can't Connect (CORS Errors)

**Error in Browser Console:**
```
Access to XMLHttpRequest at 'http://localhost:7999/api/...' from origin
'http://localhost:5173' has been blocked by CORS policy
```

**Solutions:**

#### Verify Backend is Running
```batch
curl http://localhost:7999/health
```

#### Check CORS Configuration in .env
```env
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000,http://localhost:3001,http://localhost:3002
```

#### Verify Frontend API Configuration
```batch
# Check clients/desktop/src/config/api.ts
type clients\desktop\src\config\api.ts

# Should have:
export const API_BASE_URL = 'http://localhost:7999';
```

---

### Issue 8: High Memory Usage

**Symptom:** Server uses >2GB memory

**Solutions:**

#### Check GC Settings (Already configured in main.go)
```go
// These are set automatically on startup:
os.Setenv("GOGC", "50")          // More frequent GC
os.Setenv("GOMEMLIMIT", "2GiB")  // Hard memory cap
```

#### Monitor Memory
```batch
# Check current usage
tasklist /FI "IMAGENAME eq server.exe" /FO TABLE

# View real-time stats
typeperf "\Process(server)\Private Bytes" -si 5
```

#### Restart Server
```batch
# Graceful restart (if monitoring enabled)
# Just close server window, monitor will restart it

# Or manual restart
taskkill /F /IM server.exe
server.exe
```

---

## Advanced Configuration

### Production Environment Variables

For production deployment, update `.env`:

```env
# Environment
ENVIRONMENT=production

# Security
JWT_SECRET=<generate-strong-secret-256-bits>
MASTER_ENCRYPTION_KEY=<generate-aes-256-key>

# Database (production credentials)
DB_HOST=prod-db-server.example.com
DB_PORT=5432
DB_NAME=trading_engine_prod
DB_USER=prod_user
DB_PASSWORD=<strong-password>
DB_SSL_MODE=require

# Redis (production instance)
REDIS_HOST=prod-redis.example.com
REDIS_PORT=6379
REDIS_PASSWORD=<strong-password>

# CORS (production domains)
ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com

# Monitoring
SENTRY_DSN=<your-sentry-dsn>
LOG_LEVEL=warn

# Email/SMS
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASS=<sendgrid-api-key>
```

---

### Custom Port Configuration

To run on a different port:

```env
# .env
PORT=8080
```

Then update frontend configuration:
```typescript
// clients/desktop/src/config/api.ts
export const API_BASE_URL = 'http://localhost:8080';
```

---

### Enable FIX Protocol Trading

```env
# .env
FIX_PROVISIONING_ENABLED=true
FIX_PROVISIONING_STORE_PATH=./data/fix_credentials
FIX_MASTER_PASSWORD=<secure-password>

# YOFX Configuration
YOFX_HOST=23.106.238.138
YOFX_PORT=12336
YOFX1_PASSWORD=Brand#143
YOFX2_PASSWORD=Brand#143
YOFX_TRADING_ACCOUNT=50153

# Proxy (Required for YOFX)
YOFX_USE_PROXY=true
YOFX_PROXY_HOST=81.29.145.69
YOFX_PROXY_PORT=49527
YOFX_PROXY_USERNAME=fGUqTcsdMsBZlms
YOFX_PROXY_PASSWORD=3eo1qF91WA7Fyku
```

---

## Monitoring & Logging

### View Real-Time Logs

```batch
# If logging to file (configure in code)
Get-Content backend.log -Tail 50 -Wait

# Filter for errors
Get-Content backend.log -Wait | Select-String "ERROR|FATAL|panic"

# View in separate window
start powershell -Command "Get-Content backend.log -Tail 50 -Wait"
```

---

### Monitor Server Status

```batch
# Check if server is responding
curl http://localhost:7999/health

# Check uptime
curl http://localhost:7999/api/system/stats

# Check active connections
netstat -ano | findstr :7999
```

---

### Resource Monitoring

```batch
# CPU and Memory usage
typeperf "\Process(server)\% Processor Time" "\Process(server)\Private Bytes" -si 5 -sc 60

# Active database connections
psql -U trading -d trading_engine -c "SELECT count(*) FROM pg_stat_activity;"

# Redis memory usage
redis-cli INFO memory
```

---

### Log Levels

Configure in `.env`:
```env
LOG_LEVEL=debug   # Verbose (development)
LOG_LEVEL=info    # Normal (default)
LOG_LEVEL=warn    # Warnings only (production)
LOG_LEVEL=error   # Errors only (production)
```

---

## Shutdown Procedures

### Graceful Shutdown

```batch
# Method 1: Press Ctrl+C in server window
# Server will cleanup and exit gracefully

# Method 2: Close server window
# Monitoring script will handle cleanup
```

---

### Force Stop

```batch
# Kill server process
taskkill /F /IM server.exe

# Kill by PID (get PID from netstat)
netstat -ano | findstr :7999
taskkill /F /PID [PID]
```

---

### Stop All Services (Full Stack)

```batch
# If started with START.bat, press any key in launcher window

# Or manually kill all
taskkill /F /IM server.exe
taskkill /F /IM node.exe
```

---

### Cleanup

```batch
# Stop services
net stop postgresql-x64-14
net stop Redis

# Remove PID file
del backend\server.pid

# Archive logs
move backend\keep-alive.log backend\logs\keep-alive-%date%.log
```

---

## Quick Reference

### Essential Commands

```batch
# Start backend only
cd backend && server.exe

# Start with monitoring
cd backend && start-monitor.bat

# Start full stack
START.bat

# Check health
curl http://localhost:7999/health

# View logs
Get-Content backend.log -Tail 50 -Wait

# Stop server
taskkill /F /IM server.exe

# Check port
netstat -ano | findstr :7999

# Test database
psql -U trading -d trading_engine

# Test Redis
redis-cli ping
```

---

### Directory Structure

```
Trading-Engine/
├── backend/
│   ├── server.exe           # Compiled backend executable
│   ├── .env                 # Environment configuration
│   ├── cmd/server/main.go   # Main server entry point
│   ├── keep-alive.bat       # Auto-restart monitor
│   ├── start-monitor.bat    # Start monitoring
│   └── keep-alive.log       # Monitor logs
├── START.bat                # One-click full stack launcher
└── docs/
    └── BACKEND_STARTUP_GUIDE.md  # This file
```

---

## Next Steps

1. ✅ Backend is running
2. Start frontend clients (see [MONITORING_INSTRUCTIONS.txt](../MONITORING_INSTRUCTIONS.txt))
3. Access trading platform at http://localhost:5173
4. Access admin dashboard at http://localhost:5174
5. Review [PRODUCTION_STATUS.md](../backend/PRODUCTION_STATUS.md) for deployment

---

## Support

- **Configuration:** `backend/docs/CONFIGURATION_GUIDE.md`
- **Build Issues:** `backend/BUILD_STATUS.md`
- **Production:** `backend/PRODUCTION_STATUS.md`
- **Quick Fixes:** `backend/QUICK_FIX_GUIDE.md`

---

**Backend Startup Guide - RTX Trading Engine**
**Status:** ✅ Complete and Production-Ready
