@echo off
REM ============================================================================
REM Backend Service Uninstaller
REM Removes the Trading Engine Backend Windows Service
REM ============================================================================

setlocal

REM Check for admin privileges
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: This script requires Administrator privileges.
    echo Right-click and select "Run as Administrator"
    pause
    exit /b 1
)

set "BACKEND_DIR=%~dp0"
set "SERVICE_NAME=TradingEngineBackend"
set "NSSM_EXE=%BACKEND_DIR%nssm\win64\nssm.exe"

echo ============================================================================
echo Trading Engine Backend - Service Uninstaller
echo ============================================================================
echo.

REM Check if service exists
sc query "%SERVICE_NAME%" >nul 2>&1
if %errorlevel% neq 0 (
    echo Service "%SERVICE_NAME%" not found.
    echo Nothing to uninstall.
    pause
    exit /b 0
)

echo Service Name: %SERVICE_NAME%
echo.

set /p CONFIRM="Are you sure you want to remove the service? (Y/N): "
if /i not "%CONFIRM%"=="Y" (
    echo Uninstall cancelled.
    pause
    exit /b 0
)

echo.
echo Stopping service...
net stop "%SERVICE_NAME%" >nul 2>&1
timeout /t 2 /nobreak >nul

echo Removing service...
if exist "%NSSM_EXE%" (
    "%NSSM_EXE%" remove "%SERVICE_NAME%" confirm
) else (
    sc delete "%SERVICE_NAME%"
)

echo.
echo Service removed successfully!
echo.

set /p CLEANUP="Do you want to delete log files? (Y/N): "
if /i "%CLEANUP%"=="Y" (
    echo Cleaning up log files...
    del "%BACKEND_DIR%keep-alive.log" >nul 2>&1
    del "%BACKEND_DIR%service.log" >nul 2>&1
    del "%BACKEND_DIR%service-error.log" >nul 2>&1
    del "%BACKEND_DIR%server.pid" >nul 2>&1
    echo Log files deleted.
)

echo.
pause

endlocal
