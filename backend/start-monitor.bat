@echo off
REM ============================================================================
REM Start Backend Monitor (Non-Service Mode)
REM Run this to start the keep-alive monitor in a separate window
REM ============================================================================

set "BACKEND_DIR=%~dp0"

echo Starting Backend Keep-Alive Monitor...
echo.
echo Monitor window will open. Do not close it.
echo Close this window if you want to stop monitoring later.
echo.

start "Trading Engine Backend Monitor" /min "%BACKEND_DIR%keep-alive.bat"

echo.
echo Monitor started in background.
echo.
echo To view logs: %BACKEND_DIR%keep-alive.log
echo To stop monitor: Close the "Trading Engine Backend Monitor" window
echo.

pause
