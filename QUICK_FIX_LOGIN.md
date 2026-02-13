# Quick Fix: Desktop Client Login Issue

## Problem
Desktop client login fails with "Connection Failed" error.

## Root Cause
Backend server not running on port 7999.

## Quick Fix (30 seconds)

### Step 1: Start Backend Server
```bash
# Option A: Use startup script (RECOMMENDED)
START.bat

# Option B: Manual start
cd backend
./server.exe
```

### Step 2: Verify Server is Running
```bash
# Check if port 7999 is listening
netstat -ano | findstr "7999"

# Expected output:
# TCP    0.0.0.0:7999           0.0.0.0:0              LISTENING
```

### Step 3: Test Login
```bash
curl -X POST http://localhost:7999/login \
  -H "Content-Type: application/json" \
  -d '{"username":"1","password":"password"}'

# Expected response:
# {"token":"eyJ...", "user":{"id":"1","username":"Demo User","role":"TRADER"}}
```

### Step 4: Login to Desktop Client
1. Open desktop client: http://localhost:5173
2. Use default credentials:
   - **Username:** `1` or `demo-user`
   - **Password:** `password`
3. Click "Connect to Gate"

## Success Indicators
- ✅ Backend console shows: `[INFO] Admin logged in` or similar
- ✅ Desktop client redirects to trading interface
- ✅ Market watch panel shows symbols
- ✅ WebSocket connected (green dot in status bar)

## If Login Still Fails

### Check Backend Logs
```bash
cd backend
tail -f server.log
tail -f server_error.log
```

### Common Issues

1. **Port Already in Use**
   ```bash
   # Find process on port 7999
   netstat -ano | findstr "7999"
   # Kill process
   taskkill /PID <process_id> /F
   ```

2. **CORS Error**
   - Check `.env` file: `ALLOWED_ORIGINS` includes your client URL
   - Default: `http://localhost:5173`

3. **Invalid Credentials**
   - Default demo account: username=`1`, password=`password`
   - Default admin: username=`admin`, password from `.env` (default: `password`)

4. **Database Not Initialized**
   ```bash
   cd backend
   rm -rf data/  # Reset database (WARNING: deletes all data)
   ./server.exe  # Creates fresh database
   ```

## Default Credentials Reference

| Account | Username | Password | Role | Balance |
|---------|----------|----------|------|---------|
| Demo | `1` or `demo-user` | `password` | TRADER | $5,000 |
| Admin | `admin` | `password` | ADMIN | N/A |

## Architecture Overview

```
Desktop Client (Port 5173)
         ↓ HTTP/WebSocket
Backend Server (Port 7999)
         ↓
   B-Book Engine
   Auth Service
   WebSocket Hub
   Market Data
```

## Permanent Solution

### Setup Auto-Start Service

#### Windows (NSSM)
```bash
# Install NSSM
choco install nssm

# Create service
nssm install RTX-Backend "C:\Path\To\Trading-Engine\backend\server.exe"
nssm set RTX-Backend AppDirectory "C:\Path\To\Trading-Engine\backend"
nssm start RTX-Backend
```

#### Linux/Mac (systemd)
```bash
# Create service file
sudo nano /etc/systemd/system/rtx-backend.service

# Add content:
[Unit]
Description=RTX Trading Backend
After=network.target

[Service]
Type=simple
User=youruser
WorkingDirectory=/path/to/backend
ExecStart=/path/to/backend/server
Restart=always

[Install]
WantedBy=multi-user.target

# Enable and start
sudo systemctl enable rtx-backend
sudo systemctl start rtx-backend
```

## Support
For detailed investigation report, see: `docs/LOGIN_INVESTIGATION_REPORT.md`
