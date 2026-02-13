@echo off
REM ============================================================================
REM Backend Service Installer
REM Installs backend as a Windows Service using NSSM (Non-Sucking Service Manager)
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
set "NSSM_URL=https://nssm.cc/release/nssm-2.24.zip"
set "NSSM_DIR=%BACKEND_DIR%nssm"
set "NSSM_EXE=%NSSM_DIR%\win64\nssm.exe"

echo ============================================================================
echo Trading Engine Backend - Service Installer
echo ============================================================================
echo.

REM Check if NSSM is installed
if not exist "%NSSM_EXE%" (
    echo NSSM not found. Installing NSSM...
    echo.
    echo Downloading NSSM from %NSSM_URL%...

    REM Download using PowerShell
    powershell -Command "& {[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri '%NSSM_URL%' -OutFile '%BACKEND_DIR%nssm.zip'}"

    if exist "%BACKEND_DIR%nssm.zip" (
        echo Extracting NSSM...
        powershell -Command "& {Expand-Archive -Path '%BACKEND_DIR%nssm.zip' -DestinationPath '%NSSM_DIR%' -Force}"

        REM Move files to correct location
        if exist "%NSSM_DIR%\nssm-2.24" (
            xcopy /E /Y "%NSSM_DIR%\nssm-2.24\*" "%NSSM_DIR%\" >nul
            rmdir /S /Q "%NSSM_DIR%\nssm-2.24"
        )

        del "%BACKEND_DIR%nssm.zip"
        echo NSSM installed successfully.
    ) else (
        echo ERROR: Failed to download NSSM.
        echo Please download manually from https://nssm.cc/download
        pause
        exit /b 1
    )
)

echo.
echo Service Name: %SERVICE_NAME%
echo Backend Directory: %BACKEND_DIR%
echo.

REM Check if service already exists
sc query "%SERVICE_NAME%" >nul 2>&1
if %errorlevel% equ 0 (
    echo Service already exists. Removing old service...
    "%NSSM_EXE%" stop "%SERVICE_NAME%" >nul 2>&1
    "%NSSM_EXE%" remove "%SERVICE_NAME%" confirm >nul 2>&1
    timeout /t 2 /nobreak >nul
)

REM Install service
echo Installing service...
"%NSSM_EXE%" install "%SERVICE_NAME%" "%BACKEND_DIR%keep-alive.bat"

REM Configure service
echo Configuring service...
"%NSSM_EXE%" set "%SERVICE_NAME%" AppDirectory "%BACKEND_DIR%"
"%NSSM_EXE%" set "%SERVICE_NAME%" DisplayName "Trading Engine Backend"
"%NSSM_EXE%" set "%SERVICE_NAME%" Description "Trading Engine Backend with auto-restart monitoring"
"%NSSM_EXE%" set "%SERVICE_NAME%" Start SERVICE_AUTO_START
"%NSSM_EXE%" set "%SERVICE_NAME%" AppStdout "%BACKEND_DIR%service.log"
"%NSSM_EXE%" set "%SERVICE_NAME%" AppStderr "%BACKEND_DIR%service-error.log"
"%NSSM_EXE%" set "%SERVICE_NAME%" AppRotateFiles 1
"%NSSM_EXE%" set "%SERVICE_NAME%" AppRotateOnline 1
"%NSSM_EXE%" set "%SERVICE_NAME%" AppRotateBytes 10485760

REM Set restart policy
"%NSSM_EXE%" set "%SERVICE_NAME%" AppExit Default Restart
"%NSSM_EXE%" set "%SERVICE_NAME%" AppRestartDelay 5000

echo.
echo Service installed successfully!
echo.
echo To start the service now, run:
echo   net start %SERVICE_NAME%
echo.
echo To start the service:
echo   sc start %SERVICE_NAME%
echo   OR
echo   net start %SERVICE_NAME%
echo.
echo To stop the service:
echo   sc stop %SERVICE_NAME%
echo   OR
echo   net stop %SERVICE_NAME%
echo.
echo To remove the service:
echo   sc stop %SERVICE_NAME%
echo   "%NSSM_EXE%" remove %SERVICE_NAME% confirm
echo.
echo To view service status:
echo   sc query %SERVICE_NAME%
echo.

set /p START_NOW="Do you want to start the service now? (Y/N): "
if /i "%START_NOW%"=="Y" (
    echo Starting service...
    net start "%SERVICE_NAME%"
    echo.
    echo Service started. Check logs at:
    echo   - %BACKEND_DIR%keep-alive.log
    echo   - %BACKEND_DIR%service.log
)

echo.
pause

endlocal
