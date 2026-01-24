@echo off
REM Quick Start Script for Quote Streaming Load Test
REM This script provides easy access to the monitoring tools

setlocal enabledelayedexpansion

:menu
cls
echo.
echo ================================================================================
echo                    TRADING ENGINE - LOAD TEST QUICK START
echo ================================================================================
echo.
echo Select test mode:
echo.
echo   1) Run Full Test (Monitor + Server + Load Test) [RECOMMENDED]
echo   2) Monitor Only (Server + Resource Monitoring)
echo   3) View Test Results
echo   4) View Load Test Guide
echo   5) Exit
echo.
set /p choice="Enter your choice (1-5): "

if "%choice%"=="1" goto full_test
if "%choice%"=="2" goto monitor_only
if "%choice%"=="3" goto view_results
if "%choice%"=="4" goto view_guide
if "%choice%"=="5" goto exit_menu
goto menu

:full_test
cls
echo.
echo Starting Full Load Test...
echo - Server will start automatically
echo - Resources will be monitored for 2 minutes
echo - Load test will run in background
echo - Results will be saved to monitoring_results.csv
echo.
echo Press any key to start...
pause >nul

powershell -ExecutionPolicy Bypass -File "monitor_resources.ps1"
goto menu

:monitor_only
cls
echo.
echo Starting Resource Monitoring Only...
echo - Server will start automatically
echo - Resources will be monitored for 2 minutes
echo - Results will be saved to monitoring_results.csv
echo.
echo Press any key to start...
pause >nul

powershell -ExecutionPolicy Bypass -File "monitor_resources.ps1"
goto menu

:view_results
cls
echo.
echo Monitoring Results (CSV):
echo.
if exist "monitoring_results.csv" (
    echo File found: monitoring_results.csv
    echo.
    echo Opening with default CSV application...
    start excel "monitoring_results.csv"
) else (
    echo No monitoring results found yet.
    echo Run the test first to generate results.
)
echo.
pause
goto menu

:view_guide
cls
echo.
echo Opening Load Test Guide...
echo.
if exist "LOAD_TEST_GUIDE.txt" (
    start notepad "LOAD_TEST_GUIDE.txt"
) else (
    echo Guide file not found.
)
goto menu

:exit_menu
cls
echo.
echo Exiting...
echo.
timeout /t 2 /nobreak >nul
exit /b 0
