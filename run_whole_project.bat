@echo off
echo ===================================================
echo Starting FULL Trading Engine Project
echo ===================================================

echo 1. Starting Backend...
REM Using absolute path to Go because it was just installed and PATH might not be updated yet
start "BACKEND" cmd /k "cd backend && "C:\Program Files\Go\bin\go.exe" run cmd/server/main.go"

echo 2. Starting Desktop Client...
start "DESKTOP CLIENT" cmd /k "cd clients/desktop && npm run dev"

echo 3. Starting Broker Admin...
start "BROKER ADMIN" cmd /k "cd admin/broker-admin && npm run dev"

echo 4. Starting Super Admin...
start "SUPER ADMIN" cmd /k "cd admin/super-admin && npm run dev"

echo ===================================================
echo All services are starting in separate windows.
echo - Desktop: http://localhost:5173
echo - Broker Admin: http://localhost:3000
echo - Super Admin: http://localhost:3001
echo - Backend: http://localhost:8080 (API)
echo ===================================================
pause
