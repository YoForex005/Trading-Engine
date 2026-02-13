@echo off
REM ============================================================================
REM Backend Status Check
REM Shows current status of backend and monitor
REM ============================================================================

setlocal enabledelayedexpansion

set "BACKEND_DIR=%~dp0"
set "PORT=7999"

echo ============================================================================
echo Trading Engine Backend - Status
echo ============================================================================
echo.

REM Check if server.exe is running
tasklist /FI "IMAGENAME eq server.exe" 2>NUL | find /I /N "server.exe">NUL
if "%ERRORLEVEL%"=="0" (
    echo [OK] server.exe is running

    REM Get PID
    for /f "tokens=2" %%a in ('tasklist /FI "IMAGENAME eq server.exe" /NH') do (
        echo     Process ID: %%a
        goto :FOUNDPID
    )
    :FOUNDPID
) else (
    echo [DOWN] server.exe is NOT running
)

echo.

REM Check if port is listening
netstat -ano | findstr ":%PORT%" | findstr "LISTENING" >nul 2>&1
if !errorlevel! equ 0 (
    echo [OK] Port %PORT% is listening

    REM Show connection details
    for /f "tokens=2,5" %%a in ('netstat -ano ^| findstr ":%PORT%" ^| findstr "LISTENING"') do (
        echo     Address: %%a
        echo     PID: %%b
    )
) else (
    echo [DOWN] Port %PORT% is NOT listening
)

echo.

REM Check if monitor is running
tasklist /FI "IMAGENAME eq cmd.exe" /V 2>NUL | find /I "Trading Engine Backend Monitor">NUL
if "%ERRORLEVEL%"=="0" (
    echo [OK] Keep-Alive Monitor is running
) else (
    echo [DOWN] Keep-Alive Monitor is NOT running
)

echo.

REM Check if service is installed
sc query "TradingEngineBackend" >nul 2>&1
if %errorlevel% equ 0 (
    echo [INFO] Windows Service is installed
    echo.
    sc query "TradingEngineBackend" | findstr "STATE"
) else (
    echo [INFO] Windows Service is NOT installed
)

echo.

REM Show recent log entries
if exist "%BACKEND_DIR%keep-alive.log" (
    echo Recent Monitor Log (last 10 lines):
    echo ----------------------------------------------------------------------------
    powershell -Command "& {Get-Content '%BACKEND_DIR%keep-alive.log' -Tail 10}"
    echo.
)

REM Show server log if exists
if exist "%BACKEND_DIR%server.log" (
    echo Recent Server Log (last 5 lines):
    echo ----------------------------------------------------------------------------
    powershell -Command "& {Get-Content '%BACKEND_DIR%server.log' -Tail 5}"
    echo.
)

echo ============================================================================
echo.

pause

endlocal
