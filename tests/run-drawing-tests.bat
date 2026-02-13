@echo off
REM Drawing Tools Test Execution Script
REM This script helps execute the drawing tools verification tests
REM Date: 2026-02-13

echo ========================================
echo  Drawing Tools Test Execution
echo ========================================
echo.

REM Check if desktop client is running
echo [1/4] Checking if desktop client is running...
curl -s http://localhost:5173 >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Desktop client not running at http://localhost:5173
    echo Please start the client first with: npm run dev
    echo.
    pause
    exit /b 1
) else (
    echo [OK] Desktop client is running
)
echo.

REM Check if backend is running
echo [2/4] Checking if backend is running...
curl -s http://localhost:8080/health >nul 2>&1
if errorlevel 1 (
    echo [WARNING] Backend might not be running at http://localhost:8080
    echo Tests will rely on localStorage fallback
) else (
    echo [OK] Backend is running
)
echo.

REM Display test options
echo [3/4] Test Execution Options:
echo.
echo   1. Quick Smoke Test (5 minutes - 8 core tests)
echo   2. Full Test Suite (30-60 minutes - 20 tests)
echo   3. Open Automated Test Runner (Browser-based)
echo   4. View Test Documentation
echo   5. Exit
echo.

set /p choice="Select option (1-5): "

if "%choice%"=="1" goto quick_test
if "%choice%"=="2" goto full_test
if "%choice%"=="3" goto automated_test
if "%choice%"=="4" goto view_docs
if "%choice%"=="5" goto end

echo Invalid choice. Exiting.
goto end

:quick_test
echo.
echo ========================================
echo  QUICK SMOKE TEST (8 Core Tests)
echo ========================================
echo.
echo Opening Quick Verification Checklist...
start "" "%CD%\QUICK_VERIFICATION_CHECKLIST.md"
echo.
echo Follow these steps:
echo   1. Ensure chart is loaded in browser
echo   2. Complete the 8 core tests in the checklist
echo   3. Mark each test as Pass/Fail
echo   4. If all pass, system is functional
echo.
echo Test Checklist:
echo   [ ] 1. Horizontal Line
echo   [ ] 2. Trendline
echo   [ ] 3. Rectangle
echo   [ ] 4. Fibonacci
echo   [ ] 5. Selection ^& Handles
echo   [ ] 6. Drag to Move
echo   [ ] 7. Delete
echo   [ ] 8. Persistence
echo.
pause
goto end

:full_test
echo.
echo ========================================
echo  FULL TEST SUITE (20 Tests)
echo ========================================
echo.
echo Opening Full Verification Plan...
start "" "%CD%\DRAWING_TOOLS_VERIFICATION_PLAN.md"
echo.
echo This will take 30-60 minutes to complete.
echo.
echo Follow these steps:
echo   1. Open the verification plan document
echo   2. Execute each of the 20 tests sequentially
echo   3. Mark Pass/Fail for each test
echo   4. Document failures with screenshots
echo   5. Complete the summary table at the end
echo.
echo Tip: Open browser console (F12) to monitor for errors
echo.
pause
goto end

:automated_test
echo.
echo ========================================
echo  AUTOMATED TEST RUNNER
echo ========================================
echo.
echo Opening browser-based test interface...
start "" "%CD%\automated-drawing-test.html"
echo.
echo The automated test runner provides:
echo   - Visual test progress tracking
echo   - Real-time pass/fail indicators
echo   - Comprehensive logging
echo   - JSON export of results
echo.
echo Once opened:
echo   1. Click "Run All Tests" button
echo   2. Watch automated execution
echo   3. Review pass/fail results
echo   4. Export results to JSON if needed
echo.
echo Note: This is a demo interface. Actual verification
echo       requires manual testing in the live application.
echo.
pause
goto end

:view_docs
echo.
echo ========================================
echo  TEST DOCUMENTATION
echo ========================================
echo.
echo Available test documents:
echo.
echo   1. QUICK_VERIFICATION_CHECKLIST.md
echo      - 5-minute smoke test
echo      - 8 critical path tests
echo.
echo   2. DRAWING_TOOLS_VERIFICATION_PLAN.md
echo      - Complete 20-test suite
echo      - Detailed step-by-step instructions
echo      - Troubleshooting guides
echo.
echo   3. automated-drawing-test.html
echo      - Interactive test runner
echo      - Visual progress tracking
echo.
echo   4. DRAWING_TOOLS_TEST_SUMMARY.md
echo      - Executive summary
echo      - Implementation analysis
echo      - Expected results
echo.
echo Opening all documents...
start "" "%CD%\QUICK_VERIFICATION_CHECKLIST.md"
timeout /t 1 >nul
start "" "%CD%\DRAWING_TOOLS_VERIFICATION_PLAN.md"
timeout /t 1 >nul
start "" "%CD%\DRAWING_TOOLS_TEST_SUMMARY.md"
echo.
pause
goto end

:end
echo.
echo ========================================
echo  Test Execution Complete
echo ========================================
echo.
echo Next Steps:
echo   - Document test results
echo   - Report any failures
echo   - Update memory with findings
echo.
echo To store results in memory:
echo   npx @claude-flow/cli@latest memory store --namespace "drawing-tools-tests" --key "test-results-YYYY-MM-DD" --value "Your results here"
echo.
pause
