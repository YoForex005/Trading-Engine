@echo off
REM ============================================================================
REM Backend Keep-Alive Monitor
REM Automatically restarts server.exe if it crashes or port 7999 is not listening
REM ============================================================================

setlocal enabledelayedexpansion

REM Configuration
set "BACKEND_DIR=%~dp0"
set "SERVER_EXE=%BACKEND_DIR%server.exe"
set "LOG_FILE=%BACKEND_DIR%keep-alive.log"
set "PID_FILE=%BACKEND_DIR%server.pid"
set "CHECK_INTERVAL=30"
set "PORT=7999"
set "MAX_RETRIES=3"
set "RETRY_DELAY=5"

REM Initialize log file
echo [%date% %time%] Keep-Alive Monitor Started > "%LOG_FILE%"
echo [%date% %time%] Monitoring port %PORT% every %CHECK_INTERVAL% seconds >> "%LOG_FILE%"
echo ============================================================================ >> "%LOG_FILE%"

:MONITOR_LOOP
    REM Check if server.exe exists
    if not exist "%SERVER_EXE%" (
        echo [%date% %time%] ERROR: server.exe not found at %SERVER_EXE% >> "%LOG_FILE%"
        echo ERROR: server.exe not found. Please build the backend first.
        timeout /t 60 /nobreak >nul
        goto MONITOR_LOOP
    )

    REM Check if port is listening
    netstat -ano | findstr ":%PORT%" | findstr "LISTENING" >nul 2>&1

    if !errorlevel! equ 0 (
        REM Port is listening - all good
        echo [%date% %time%] Status: Backend healthy on port %PORT% >> "%LOG_FILE%"
    ) else (
        REM Port not listening - attempt restart
        echo [%date% %time%] WARNING: Port %PORT% not listening - attempting restart >> "%LOG_FILE%"
        echo [%date% %time%] WARNING: Backend down on port %PORT% - restarting...

        call :KILL_SERVER
        call :START_SERVER
    )

    REM Wait before next check
    timeout /t %CHECK_INTERVAL% /nobreak >nul
    goto MONITOR_LOOP

:KILL_SERVER
    echo [%date% %time%] Killing existing server processes... >> "%LOG_FILE%"

    REM Kill by PID file if exists
    if exist "%PID_FILE%" (
        set /p SERVER_PID=<"%PID_FILE%"
        taskkill /F /PID !SERVER_PID! >nul 2>&1
        del "%PID_FILE%" >nul 2>&1
    )

    REM Kill all server.exe processes
    taskkill /F /IM server.exe >nul 2>&1

    REM Kill processes using port 7999
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%PORT%" ^| findstr "LISTENING"') do (
        echo [%date% %time%] Killing process %%a on port %PORT% >> "%LOG_FILE%"
        taskkill /F /PID %%a >nul 2>&1
    )

    timeout /t 2 /nobreak >nul
    exit /b 0

:START_SERVER
    echo [%date% %time%] Starting server.exe... >> "%LOG_FILE%"

    set "RETRY_COUNT=0"

    :RETRY_START
        if !RETRY_COUNT! geq %MAX_RETRIES% (
            echo [%date% %time%] ERROR: Failed to start server after %MAX_RETRIES% attempts >> "%LOG_FILE%"
            echo ERROR: Failed to start server after %MAX_RETRIES% attempts
            exit /b 1
        )

        REM Start server in background
        cd /d "%BACKEND_DIR%"
        start /B "" "%SERVER_EXE%" >> "%LOG_FILE%" 2>&1

        REM Wait for server to start
        timeout /t %RETRY_DELAY% /nobreak >nul

        REM Check if port is now listening
        netstat -ano | findstr ":%PORT%" | findstr "LISTENING" >nul 2>&1

        if !errorlevel! equ 0 (
            echo [%date% %time%] SUCCESS: Server started successfully >> "%LOG_FILE%"
            echo [%date% %time%] Server restarted successfully

            REM Save PID
            for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%PORT%" ^| findstr "LISTENING"') do (
                echo %%a > "%PID_FILE%"
                echo [%date% %time%] Server PID: %%a >> "%LOG_FILE%"
            )

            exit /b 0
        ) else (
            set /a RETRY_COUNT+=1
            echo [%date% %time%] WARNING: Start attempt !RETRY_COUNT! failed, retrying... >> "%LOG_FILE%"
            timeout /t %RETRY_DELAY% /nobreak >nul
            goto RETRY_START
        )

    exit /b 1

endlocal
