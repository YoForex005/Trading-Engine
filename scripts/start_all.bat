@echo off
setlocal enabledelayedexpansion
title Trading Engine - Full Project Launcher
color 0A

REM ============================================================
REM  TRADING ENGINE - ONE-CLICK PROJECT LAUNCHER
REM  Starts: Backend + Desktop + Broker Admin + Super Admin
REM ============================================================

REM Always run from project root
cd /d "%~dp0\.."
set "PROJECT_ROOT=%CD%"
echo [PROJECT ROOT] %PROJECT_ROOT%

REM ============================================================
REM  STEP 0: ASCII Banner
REM ============================================================
echo.
echo  ========================================================
echo   _____ ____      _    ____ ___ _   _  ____
echo  ^|_   _^|  _ \    / \  ^|  _ \_ _^| \ ^| ^|/ ___^|
echo    ^| ^| ^| ^|_) ^|  / _ \ ^| ^| ^| ^| ^|^|  \^| ^| ^|  _
echo    ^| ^| ^|  _ ^<  / ___ \^| ^|_^| ^| ^|^| ^|\  ^| ^|_^| ^|
echo    ^|_^| ^|_^| \_\/_/   \_\____^/_^|_^|_^| \_^|\____^|
echo.
echo     ENGINE - Full Platform Launcher v2.0
echo  ========================================================
echo.

REM ============================================================
REM  STEP 1: Check Prerequisites
REM ============================================================
echo [STEP 1/6] Checking prerequisites...
echo.

REM Check Node.js
where node >nul 2>&1
if %errorlevel% neq 0 (
    echo   [ERROR] Node.js is NOT installed!
    echo   Please install Node.js from https://nodejs.org
    echo.
    pause
    exit /b 1
)
for /f "tokens=*" %%i in ('node --version') do set NODE_VER=%%i
echo   [OK] Node.js: %NODE_VER%

REM Check npm
where npm >nul 2>&1
if %errorlevel% neq 0 (
    echo   [ERROR] npm is NOT installed!
    pause
    exit /b 1
)
for /f "tokens=*" %%i in ('npm --version') do set NPM_VER=%%i
echo   [OK] npm: v%NPM_VER%

REM Check Go (optional - uses pre-built exe if Go not available)
set "GO_AVAILABLE=0"
where go >nul 2>&1
if %errorlevel% equ 0 (
    set "GO_AVAILABLE=1"
    for /f "tokens=3" %%i in ('go version') do set GO_VER=%%i
    echo   [OK] Go: !GO_VER!
) else (
    REM Try common Go installation paths
    if exist "C:\Program Files\Go\bin\go.exe" (
        set "PATH=C:\Program Files\Go\bin;%PATH%"
        set "GO_AVAILABLE=1"
        echo   [OK] Go: Found at C:\Program Files\Go\bin
    ) else if exist "C:\Go\bin\go.exe" (
        set "PATH=C:\Go\bin;%PATH%"
        set "GO_AVAILABLE=1"
        echo   [OK] Go: Found at C:\Go\bin
    ) else (
        echo   [WARN] Go not found - will use pre-built server.exe
    )
)

echo.

REM ============================================================
REM  STEP 2: Kill any existing processes on our ports
REM ============================================================
echo [STEP 2/6] Cleaning up existing processes...

REM Kill existing server.exe instances
taskkill /F /IM server.exe >nul 2>&1 && echo   Stopped: existing server.exe || echo   No existing server.exe running

REM Kill processes on specific ports
for %%p in (7999 5173 3000 3001) do (
    for /f "tokens=5" %%a in ('netstat -aon ^| findstr ":%%p " ^| findstr "LISTENING" 2^>nul') do (
        if not "%%a"=="0" (
            taskkill /PID %%a /F >nul 2>&1
            echo   Freed port %%p ^(PID: %%a^)
        )
    )
)
echo   Ports cleared: 7999, 5173, 3000, 3001
echo.

REM ============================================================
REM  STEP 3: Build / Verify Backend
REM ============================================================
echo [STEP 3/6] Preparing Backend...

pushd backend

if "!GO_AVAILABLE!"=="1" (
    echo   Building from source with Go...
    go mod tidy >nul 2>&1
    go build -o server.exe cmd\server\main.go 2>&1
    if !errorlevel! neq 0 (
        echo   [WARN] Build failed. Checking for existing server.exe...
        if exist server.exe (
            echo   [OK] Using existing pre-built server.exe
        ) else (
            echo   [ERROR] No server.exe found and build failed!
            popd
            pause
            exit /b 1
        )
    ) else (
        echo   [OK] Backend built successfully
    )
) else (
    if exist server.exe (
        echo   [OK] Using pre-built server.exe ^(Go not installed^)
    ) else (
        echo   [ERROR] No server.exe found and Go is not installed!
        echo   Please install Go from https://go.dev/dl/ or provide server.exe
        popd
        pause
        exit /b 1
    )
)

popd
echo.

REM ============================================================
REM  STEP 4: Install npm dependencies (if missing)
REM ============================================================
echo [STEP 4/6] Installing dependencies ^(if needed^)...

REM Desktop Client
if not exist "clients\desktop\node_modules" (
    echo   Installing Desktop Client deps...
    pushd clients\desktop
    call npm install --no-audit --no-fund >nul 2>&1
    if !errorlevel! neq 0 (
        echo   [WARN] Desktop npm install had warnings, retrying...
        call npm install >nul 2>&1
    )
    popd
    echo   [OK] Desktop Client dependencies installed
) else (
    echo   [OK] Desktop Client deps already installed
)

REM Broker Admin
if not exist "admin\broker-admin\node_modules" (
    echo   Installing Broker Admin deps...
    pushd admin\broker-admin
    call npm install --no-audit --no-fund >nul 2>&1
    popd
    echo   [OK] Broker Admin dependencies installed
) else (
    echo   [OK] Broker Admin deps already installed
)

REM Super Admin
if not exist "admin\super-admin\node_modules" (
    echo   Installing Super Admin deps...
    pushd admin\super-admin
    call npm install --no-audit --no-fund >nul 2>&1
    popd
    echo   [OK] Super Admin dependencies installed
) else (
    echo   [OK] Super Admin deps already installed
)

echo.

REM ============================================================
REM  STEP 5: Start All Services
REM ============================================================
echo [STEP 5/6] Launching all services...
echo.

REM --- Start Backend ---
echo   Starting Backend ^(port 7999^)...
start "TRADING-ENGINE-BACKEND" /D "%PROJECT_ROOT%\backend" cmd /k server.exe
timeout /t 4 /nobreak >nul

REM --- Verify Backend Health ---
curl -s http://localhost:7999/health >nul 2>&1
if %errorlevel% equ 0 (
    echo   [OK] Backend started and healthy
) else (
    echo   [WAIT] Backend still starting, waiting 3 more seconds...
    timeout /t 3 /nobreak >nul
    curl -s http://localhost:7999/health >nul 2>&1
    if !errorlevel! equ 0 (
        echo   [OK] Backend started and healthy
    ) else (
        echo   [WARN] Backend may still be initializing - check its window
    )
)

REM --- Start Desktop Client ---
echo   Starting Desktop Client ^(port 5173^)...
start "TRADING-ENGINE-DESKTOP" /D "%PROJECT_ROOT%\clients\desktop" cmd /k "npx vite --host"
timeout /t 2 /nobreak >nul
echo   [STARTED] Desktop Client

REM --- Start Broker Admin ---
echo   Starting Broker Admin ^(port 3000^)...
start "TRADING-ENGINE-BROKER-ADMIN" /D "%PROJECT_ROOT%\admin\broker-admin" cmd /k "npx next dev -p 3000"
timeout /t 2 /nobreak >nul
echo   [STARTED] Broker Admin

REM --- Start Super Admin ---
echo   Starting Super Admin ^(port 3001^)...
start "TRADING-ENGINE-SUPER-ADMIN" /D "%PROJECT_ROOT%\admin\super-admin" cmd /k "npx next dev -p 3001"
timeout /t 2 /nobreak >nul
echo   [STARTED] Super Admin

echo.

REM ============================================================
REM  STEP 6: Health Check Summary
REM ============================================================
echo [STEP 6/6] Verifying services...
timeout /t 5 /nobreak >nul

set "ALL_OK=1"

REM Check Backend
curl -s http://localhost:7999/health >nul 2>&1
if %errorlevel% equ 0 (
    echo   [OK] Backend API:     http://localhost:7999
) else (
    echo   [!!] Backend API:     http://localhost:7999  ^(check backend window^)
    set "ALL_OK=0"
)

REM Check Desktop
curl -s -o nul http://localhost:5173/ >nul 2>&1
if %errorlevel% equ 0 (
    echo   [OK] Desktop Client:  http://localhost:5173
) else (
    echo   [..] Desktop Client:  http://localhost:5173  ^(still starting^)
)

REM Check Broker Admin
curl -s -o nul http://localhost:3000/ >nul 2>&1
if %errorlevel% equ 0 (
    echo   [OK] Broker Admin:    http://localhost:3000
) else (
    echo   [..] Broker Admin:    http://localhost:3000  ^(still compiling^)
)

REM Check Super Admin
curl -s -o nul http://localhost:3001/ >nul 2>&1
if %errorlevel% equ 0 (
    echo   [OK] Super Admin:     http://localhost:3001
) else (
    echo   [..] Super Admin:     http://localhost:3001  ^(still compiling^)
)

echo.
echo   WebSocket:       ws://localhost:7999/ws
echo   Health Check:    http://localhost:7999/health
echo.

REM ============================================================
REM  ALL SERVICES LAUNCHED
REM ============================================================
echo  ========================================================
echo   ALL SERVICES LAUNCHED!
echo  ========================================================
echo.
echo   Login Credentials:
echo     Username: admin
echo     Password: Admin@123
echo.
echo   4 service windows are open.
echo   Close them or press any key here to STOP ALL.
echo  ========================================================
echo.
pause

REM ============================================================
REM  SHUTDOWN
REM ============================================================
echo.
echo Stopping all services...

REM Kill by window title
taskkill /FI "WINDOWTITLE eq TRADING-ENGINE-BACKEND*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq TRADING-ENGINE-DESKTOP*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq TRADING-ENGINE-BROKER-ADMIN*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq TRADING-ENGINE-SUPER-ADMIN*" /F >nul 2>&1

REM Also kill by process name as backup
taskkill /F /IM server.exe >nul 2>&1

REM Kill node processes on our ports
for %%p in (7999 5173 3000 3001) do (
    for /f "tokens=5" %%a in ('netstat -aon ^| findstr ":%%p " ^| findstr "LISTENING" 2^>nul') do (
        if not "%%a"=="0" (
            taskkill /PID %%a /F >nul 2>&1
        )
    )
)

echo.
echo All services stopped. Goodbye!
timeout /t 3 /nobreak >nul
exit /b 0
