@echo off
echo ============================================
echo Market Watch Fix Verification
echo ============================================
echo.

echo [1/4] Checking backend compilation...
cd backend
go build -o verify_build.exe ./cmd/server
if %errorlevel% neq 0 (
    echo ERROR: Backend compilation failed!
    pause
    exit /b 1
)
echo OK: Backend compiles successfully
del verify_build.exe
cd ..
echo.

echo [2/4] Checking frontend dependencies...
cd clients\desktop
if not exist node_modules (
    echo Installing dependencies...
    call npm install
)
echo OK: Dependencies present
cd ..\..
echo.

echo [3/4] Verifying critical files...
if exist "clients\desktop\src\App.tsx" (
    echo OK: App.tsx found
) else (
    echo ERROR: App.tsx not found!
    pause
    exit /b 1
)

if exist "clients\desktop\src\store\useMarketDataStore.ts" (
    echo OK: useMarketDataStore.ts found
) else (
    echo ERROR: useMarketDataStore.ts not found!
    pause
    exit /b 1
)

if exist "backend\cmd\server\main.go" (
    echo OK: main.go found
) else (
    echo ERROR: main.go not found!
    pause
    exit /b 1
)
echo.

echo [4/4] Checking for fix markers...
findstr /C:"useMarketDataStore.getState().updateTick" "clients\desktop\src\App.tsx" >nul
if %errorlevel% equ 0 (
    echo OK: Fix applied - useMarketDataStore.updateTick() call found
) else (
    echo WARNING: Fix marker not found in App.tsx
    echo This might indicate the fix was not applied correctly
)
echo.

echo ============================================
echo Verification Complete!
echo ============================================
echo.
echo NEXT STEPS:
echo 1. Start backend: cd backend ^&^& go run ./cmd/server
echo 2. Wait for YOFX2 auto-connect and subscriptions
echo 3. Start frontend: cd clients\desktop ^&^& npm run dev
echo 4. Login and check MarketWatch panel
echo 5. Verify live prices are displayed
echo.
echo For detailed information, see:
echo   docs\market-watch-fix-report.md
echo.
pause
