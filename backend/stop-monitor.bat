@echo off
REM ============================================================================
REM Stop Backend Monitor
REM Stops the keep-alive monitor and server
REM ============================================================================

set "BACKEND_DIR=%~dp0"

echo Stopping Backend Monitor and Server...
echo.

REM Kill keep-alive.bat processes
taskkill /F /FI "WINDOWTITLE eq Trading Engine Backend Monitor*" >nul 2>&1

REM Kill server.exe
taskkill /F /IM server.exe >nul 2>&1

REM Kill processes on port 7999
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":7999" ^| findstr "LISTENING"') do (
    taskkill /F /PID %%a >nul 2>&1
)

REM Clean up PID file
if exist "%BACKEND_DIR%server.pid" (
    del "%BACKEND_DIR%server.pid"
)

echo.
echo Backend monitor and server stopped.
echo.

pause
