@echo off
REM Test script for FIX connection management endpoints

echo ======================================
echo Testing FIX Connection Management API
echo ======================================
echo.

set BASE_URL=http://localhost:8080

echo [1] Checking FIX Status...
curl -s %BASE_URL%/admin/fix/status | jq .
echo.
echo.

echo [2] Getting Detailed Status for YOFX1...
curl -s "%BASE_URL%/admin/fix/detailed-status?session_id=YOFX1" | jq .
echo.
echo.

echo [3] Getting Connection Stats...
curl -s %BASE_URL%/admin/fix/connection-stats | jq .
echo.
echo.

echo [4] Getting Diagnostics for YOFX1...
curl -s "%BASE_URL%/admin/fix/diagnostics?session_id=YOFX1"
echo.
echo.

echo [5] Testing Reconnect (if needed)...
echo Do you want to test reconnect? (y/n)
choice /c yn /n /m "Press Y for Yes, N for No: "
if errorlevel 2 goto skip_reconnect

echo Sending reconnect request...
curl -X POST %BASE_URL%/admin/fix/reconnect ^
  -H "Content-Type: application/json" ^
  -d "{\"session_id\": \"YOFX1\", \"enable_auto_retry\": true}" | jq .
echo.

:skip_reconnect
echo.
echo ======================================
echo Test Complete
echo ======================================
echo.
echo Use these commands for manual testing:
echo.
echo   Status:       curl %BASE_URL%/admin/fix/status
echo   Details:      curl "%BASE_URL%/admin/fix/detailed-status?session_id=YOFX1"
echo   Stats:        curl %BASE_URL%/admin/fix/connection-stats
echo   Diagnostics:  curl "%BASE_URL%/admin/fix/diagnostics?session_id=YOFX1"
echo   Reconnect:    curl -X POST %BASE_URL%/admin/fix/reconnect -H "Content-Type: application/json" -d "{\"session_id\":\"YOFX1\"}"
echo.
pause
