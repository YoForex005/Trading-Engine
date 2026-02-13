@echo off
REM Real-time Log Monitoring for Trading Engine
REM Monitors: Backend logs, npm/vite output, startup errors

setlocal enabledelayedexpansion
set "ProjectRoot=C:\Users\Yofor\Desktop\Trading-Engine"
set "LogDir=%ProjectRoot%\logs"
set "Timestamp=%date% %time%"

if not exist "%LogDir%" mkdir "%LogDir%"

echo.
echo ================================================================================
echo Real-Time Log Monitoring System
echo Started: %Timestamp%
echo ================================================================================
echo.

REM Check System Requirements
echo Checking system requirements...
echo.

node --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Node.js NOT found - Install from nodejs.org
) else (
    for /f "tokens=*" %%i in ('node --version') do set "NodeVer=%%i"
    echo [OK] Node.js: !NodeVer!
)

npm --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] npm NOT found
) else (
    for /f "tokens=*" %%i in ('npm --version') do set "NpmVer=%%i"
    echo [OK] npm: !NpmVer!
)

go version >nul 2>&1
if errorlevel 1 (
    echo [WARN] Go NOT found - needed for backend
) else (
    for /f "tokens=*" %%i in ('go version') do set "GoVer=%%i"
    echo [OK] Go installed
)

echo.
echo ================================================================================
echo BACKEND STATUS
echo ================================================================================
if exist "%ProjectRoot%\backend.log" (
    echo [OK] Backend log file found
    echo Last 20 lines of backend.log:
    echo.
    for /f "tokens=*" %%i in ('powershell -Command "Get-Content '%ProjectRoot%\backend.log' -Tail 20"') do (
        echo %%i
    )
) else (
    echo [WARN] Backend log file not found: %ProjectRoot%\backend.log
)

if exist "%ProjectRoot%\server.exe" (
    echo [OK] Backend executable found (server.exe)
) else (
    echo [INFO] Backend executable not found
)

echo.
echo ================================================================================
echo DESKTOP CLIENT STATUS
echo ================================================================================
if exist "%ProjectRoot%\clients\desktop\package.json" (
    echo [OK] Desktop client package.json found
) else (
    echo [ERROR] Desktop client package.json NOT found
)

if exist "%ProjectRoot%\clients\desktop\node_modules" (
    echo [OK] Desktop node_modules found
) else (
    echo [WARN] Desktop node_modules NOT found - Run: npm install in clients/desktop
)

if exist "%ProjectRoot%\clients\desktop\vite.config.ts" (
    echo [OK] Vite config found
)

if exist "%ProjectRoot%\clients\desktop\dist" (
    echo [OK] Desktop build output found
) else (
    echo [INFO] Desktop build output not found
)

echo.
echo ================================================================================
echo ADMIN DASHBOARD STATUS
echo ================================================================================
if exist "%ProjectRoot%\clients\admin-dashboard\package.json" (
    echo [OK] Admin dashboard package.json found
) else (
    echo [WARN] Admin dashboard package.json NOT found
)

if exist "%ProjectRoot%\clients\admin-dashboard\node_modules" (
    echo [OK] Admin node_modules found
) else (
    echo [WARN] Admin node_modules NOT found - Run: npm install in clients/admin-dashboard
)

echo.
echo ================================================================================
echo MOBILE CLIENT STATUS
echo ================================================================================
if exist "%ProjectRoot%\clients\mobile\package.json" (
    echo [OK] Mobile client package.json found
) else (
    echo [WARN] Mobile client package.json NOT found
)

if exist "%ProjectRoot%\clients\mobile\node_modules" (
    echo [OK] Mobile node_modules found
) else (
    echo [WARN] Mobile node_modules NOT found - Run: npm install in clients/mobile
)

echo.
echo ================================================================================
echo ERROR SCAN
echo ================================================================================
echo Scanning for critical errors...
echo.

setlocal enabledelayedexpansion
set ErrorCount=0

if exist "%ProjectRoot%\backend.log" (
    for /f "tokens=*" %%i in ('powershell -Command "Select-String -Path '%ProjectRoot%\backend.log' -Pattern 'panic|ERROR|FATAL' | Select-Object -Last 5"') do (
        echo [CRITICAL] %%i
        set /a ErrorCount+=1
    )
)

if !ErrorCount! equ 0 (
    echo [OK] No critical errors found
)

echo.
echo ================================================================================
echo MONITORING SUMMARY
echo ================================================================================
echo Log directory: %LogDir%
echo Project root: %ProjectRoot%
echo.
echo To run with specific options:
echo   - Monitor backend only: MONITOR_LOGS.bat
echo   - Monitor desktop: set Component=desktop ^& MONITOR_LOGS.bat
echo   - Monitor all: set Component=all ^& MONITOR_LOGS.bat
echo.
echo Next steps:
echo   1. Run START.bat to initialize all services
echo   2. Check individual client logs as they start
echo   3. Report critical errors from this monitoring output
echo.
pause
