@echo off
setlocal

REM MANDATORY FIX: Always run from Project Root
cd /d %~dp0
echo Current Working Directory: %CD%

echo Starting Trading Engine...

REM ==========================================
REM BACKEND
REM ==========================================
echo [1/2] Checking Backend...

echo [1/2] Building Backend...
REM Force kill any running server instance to release file lock
taskkill /F /IM server.exe >nul 2>&1

pushd backend
echo     - Resolving dependencies...
go mod tidy
echo     - Building server.exe from source...
go build -o server.exe cmd/server/main.go
if errorlevel 1 (
    echo [ERROR] Backend build failed!
    pause
    exit /b 1
)
popd
echo     - Build successful.

echo     - Starting Backend Service (From Project Root)...
REM EXECUTE FROM ROOT, pointing to the exe in backend folder
start "Trading Engine Backend" cmd /k "backend\server.exe"

REM ==========================================
REM FRONTEND
REM ==========================================
echo [2/2] Checking Frontend...
pushd clients\desktop

REM Check dependencies
if not exist node_modules (
    echo     - Dependencies missing.
    echo     - Installing...
    call npm install
)

echo     - Starting Desktop Client...
start "Trading Engine Desktop" cmd /k "npm run dev -- --host"
popd

echo ===================================================
echo Services launched!
echo Backend: http://localhost:7999
echo Frontend: http://localhost:5173
echo ===================================================
pause
