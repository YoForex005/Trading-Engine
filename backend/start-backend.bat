@echo off
REM ========================================
REM Trading Engine Backend Startup Script
REM ========================================
REM This script starts the backend server with proper error handling

echo.
echo ========================================
echo Trading Engine Backend - Startup
echo ========================================
echo.

REM Change to backend directory
cd /d "%~dp0"

REM Check if .env file exists
if not exist ".env" (
    echo [ERROR] .env file not found!
    echo Please copy .env.example to .env and configure it.
    pause
    exit /b 1
)

REM Check if server.exe exists
if not exist "server.exe" (
    echo [WARNING] server.exe not found. Building from source...
    echo.

    REM Check if Go is installed
    where go >nul 2>nul
    if errorlevel 1 (
        echo [ERROR] Go is not installed or not in PATH!
        echo Please install Go from https://go.dev/dl/
        pause
        exit /b 1
    )

    echo [BUILD] Compiling backend...
    go build -o server.exe ./cmd/server

    if errorlevel 1 (
        echo [ERROR] Build failed!
        pause
        exit /b 1
    )

    echo [SUCCESS] Build completed successfully
    echo.
)

REM Check if port 7999 is already in use
netstat -ano | findstr ":7999" | findstr "LISTENING" >nul 2>nul
if not errorlevel 1 (
    echo [WARNING] Port 7999 is already in use!
    echo.
    echo Do you want to kill the existing process and restart?
    choice /C YN /M "Kill existing server"

    if errorlevel 2 (
        echo Startup cancelled.
        pause
        exit /b 0
    )

    REM Get PID of process using port 7999
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":7999" ^| findstr "LISTENING"') do (
        echo Killing process %%a...
        taskkill //F //PID %%a >nul 2>nul
    )

    echo Waiting 2 seconds for port to be released...
    timeout /t 2 /nobreak >nul
)

REM Start the server
echo [START] Starting backend server on port 7999...
echo.
echo Press Ctrl+C to stop the server
echo ========================================
echo.

server.exe

REM If server stops, show exit code
if errorlevel 1 (
    echo.
    echo [ERROR] Server exited with error code %errorlevel%
    pause
)
