@echo off
echo ========================================
echo Chart Drawing Toolbar Verification
echo ========================================
echo.

echo Testing drawing toolbar implementation...
echo.

echo [1/5] Checking DrawingManager types...
findstr /C:"'rectangle' | 'ellipse' | 'arrow' | 'pitchfork'" "C:\Users\Yofor\Desktop\Trading-Engine\clients\desktop\src\services\drawingManager.ts" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [OK] New drawing types added to DrawingManager
) else (
    echo [FAIL] Drawing types not found in DrawingManager
)

echo [2/5] Checking TradingChart command subscriptions...
findstr /C:"'rectangle', 'ellipse', 'arrow', 'pitchfork'" "C:\Users\Yofor\Desktop\Trading-Engine\clients\desktop\src\components\TradingChart.tsx" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [OK] TradingChart subscribes to new tool commands
) else (
    echo [FAIL] Command subscriptions not updated
)

echo [3/5] Checking TopToolbar buttons...
findstr /C:"Square" "C:\Users\Yofor\Desktop\Trading-Engine\clients\desktop\src\components\layout\TopToolbar.tsx" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [OK] New toolbar buttons added
) else (
    echo [FAIL] Toolbar buttons not found
)

echo [4/5] Checking CSS styles...
findstr /C:"chart-drawing-overlay" "C:\Users\Yofor\Desktop\Trading-Engine\clients\desktop\src\index.css" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [OK] Drawing overlay CSS present
) else (
    echo [FAIL] CSS styles missing
)

echo [5/5] Checking documentation...
if exist "C:\Users\Yofor\Desktop\Trading-Engine\docs\DRAWING_TOOLBAR_IMPLEMENTATION.md" (
    echo [OK] Implementation documentation created
) else (
    echo [FAIL] Documentation not found
)

echo.
echo ========================================
echo Verification Complete!
echo ========================================
echo.
echo To use the drawing tools:
echo 1. Start the desktop client: cd clients\desktop ^&^& npm run dev
echo 2. Open chart in browser: http://localhost:5173
echo 3. Click any drawing tool button in top toolbar
echo 4. Click on chart to draw (cursor becomes crosshair)
echo.
echo See docs\DRAWING_TOOLBAR_IMPLEMENTATION.md for full guide
echo.
pause
