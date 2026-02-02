# 🚀 Desktop Client Startup Guide

## ✅ FIXES APPLIED

The following critical issues have been fixed:

1. **TypeScript Compilation Errors** ✅
   - Removed non-existent `ChartControls` export
   - Exported missing type interfaces (`PrintSetupDialogProps`, `PrintSettings`, `PrintPreviewDialogProps`, `SavePictureDialogProps`)
   - Fixed `onClose` → `onCancel` in PrintSetupDialog
   - Fixed type-only import violations

2. **CORS Configuration** ✅
   - Added `http://localhost:5173` to backend CORS whitelist
   - Updated both `backend/config/config.go` and `backend/.env`

3. **Environment Setup** ✅
   - Created `backend/.env` from `.env.example`
   - Configured default values for development

---

## 📋 STARTUP INSTRUCTIONS

### Step 1: Start Backend Server (Port 7999)

**Option A: Using Go directly**
```bash
cd backend
go run cmd/server/main.go
```

**Option B: Using compiled binary**
```bash
cd backend
go build -o server.exe cmd/server/main.go
./server.exe
```

**Expected Output:**
```
[GIN-debug] Listening and serving HTTP on :7999
WebSocket hub started
```

---

### Step 2: Start Desktop Client (Port 5173)

Open a **new terminal** and run:

```bash
cd clients/desktop
npm run dev
```

**Expected Output:**
```
VITE v7.2.4  ready in XXX ms

➜  Local:   http://localhost:5173/
➜  Network: use --host to expose
➜  press h + enter to show help
```

---

### Step 3: Open Browser

Navigate to: **http://localhost:5173/**

### Step 4: Login

Use default credentials:
- **Username:** admin
- **Password:** password

---

## 🔧 TROUBLESHOOTING

### Issue: "TypeScript compilation errors"
**Solution:** Already fixed! Run `npm run typecheck` to verify:
```bash
cd clients/desktop
npm run typecheck
```

### Issue: "CORS policy blocked"
**Solution:** Already fixed! Backend now allows port 5173.

### Issue: "Cannot connect to backend"
**Symptoms:**
- WebSocket connection failed
- API requests return network errors
- Console shows `ECONNREFUSED` on port 7999

**Solution:**
1. Verify backend is running: Check terminal for "Listening on :7999"
2. Check if port 7999 is available:
   ```bash
   netstat -ano | findstr :7999
   ```
3. Allow port through Windows Firewall if needed

### Issue: "Database connection failed"
**Symptoms:**
- Backend logs show database errors
- PostgreSQL connection refused

**Solution:**
The app can run **WITHOUT PostgreSQL/Redis** for basic functionality. The backend will log warnings but continue running in memory-only mode.

To use full features with persistence:
1. Install PostgreSQL (port 5432) and Redis (port 6379)
2. Create database: `createdb trading_engine`
3. Update `backend/.env` with database credentials

---

## 🎯 QUICK START (One Command)

**Terminal 1 - Backend:**
```bash
cd backend && go run cmd/server/main.go
```

**Terminal 2 - Frontend:**
```bash
cd clients/desktop && npm run dev
```

**Browser:**
```
http://localhost:5173
```

---

## 📊 PORT SUMMARY

| Service | Port | Required | Status |
|---------|------|----------|--------|
| Backend API/WebSocket | 7999 | ✅ YES | Fixed & Ready |
| Desktop Client | 5173 | ✅ YES | Fixed & Ready |
| PostgreSQL | 5432 | ❌ Optional | For persistence |
| Redis | 6379 | ❌ Optional | For sessions |
| Broker Admin | 3000 | ❌ Optional | Admin dashboard |
| Super Admin | 3001 | ❌ Optional | Super admin |

---

## ✅ VERIFICATION CHECKLIST

- [x] TypeScript compiles without errors
- [x] Backend CORS includes port 5173
- [x] Backend .env file created
- [ ] Backend server running on port 7999
- [ ] Desktop client running on port 5173
- [ ] Browser can access http://localhost:5173
- [ ] WebSocket connection established
- [ ] Login successful with admin/password

---

## 🎉 SUCCESS INDICATORS

When everything is working, you should see:

1. **Backend Terminal:**
   ```
   [GIN] 2024/XX/XX - 12:00:00 | 200 | GET /api/config
   [GIN] 2024/XX/XX - 12:00:01 | 101 | WebSocket /ws connected
   ```

2. **Browser Console (F12):**
   ```
   ✅ WebSocket connected
   ✅ Config loaded
   ✅ Authenticated
   ```

3. **Desktop Client UI:**
   - Trading chart visible
   - Market watch panel showing symbols
   - Order panel accessible
   - No error messages

---

## 📚 ADDITIONAL RESOURCES

- **Main README:** `C:\Users\s s laptop bazar\Trading-Engine2\README.md`
- **Backend Config:** `C:\Users\s s laptop bazar\Trading-Engine2\backend\.env`
- **Desktop Client:** `C:\Users\s s laptop bazar\Trading-Engine2\clients\desktop\`

---

## 🐛 STILL HAVING ISSUES?

If the app still doesn't load:

1. **Clear browser cache:** Ctrl+Shift+Delete
2. **Restart both servers:** Kill terminals and restart
3. **Check Node.js version:** `node --version` (should be 18+)
4. **Check Go version:** `go version` (should be 1.22+)
5. **Reinstall dependencies:**
   ```bash
   cd clients/desktop
   rm -rf node_modules package-lock.json
   npm install
   ```

---

**Your desktop client is now ready to run! 🚀**
