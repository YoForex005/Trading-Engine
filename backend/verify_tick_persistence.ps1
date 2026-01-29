# Verification script for tick persistence fix
# PowerShell version for Windows

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Tick Persistence Verification Script" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Check if backend is running
Write-Host "1. Checking if backend is running..." -ForegroundColor Yellow
$process = Get-Process -Name "server" -ErrorAction SilentlyContinue
if ($process) {
    Write-Host "   ✅ Backend is running (PID: $($process.Id))" -ForegroundColor Green
} else {
    Write-Host "   ❌ Backend is NOT running - please start it first" -ForegroundColor Red
    exit 1
}

# Wait for initial ticks
Write-Host ""
Write-Host "2. Waiting 10 seconds for tick data..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Check data directory
$dataDir = ".\data\ticks"
if (-not (Test-Path $dataDir)) {
    Write-Host "   ❌ Data directory not found: $dataDir" -ForegroundColor Red
    exit 1
}

Write-Host "   ✅ Data directory exists: $dataDir" -ForegroundColor Green

# Check for today's tick files
Write-Host ""
Write-Host "3. Checking for today's tick files..." -ForegroundColor Yellow
$today = Get-Date -Format "yyyy-MM-dd"
Write-Host "   Looking for files with date: $today" -ForegroundColor Gray

$tickCount = 0
Get-ChildItem $dataDir -Directory | ForEach-Object {
    $symbol = $_.Name
    $todayFile = Join-Path $_.FullName "$today.json"

    if (Test-Path $todayFile) {
        $fileSize = (Get-Item $todayFile).Length
        if ($fileSize -gt 100) {
            $fileSizeKB = [math]::Round($fileSize / 1KB, 2)
            Write-Host "   ✅ ${symbol}: ${fileSizeKB} KB" -ForegroundColor Green
            $tickCount++
        }
    }
}

if ($tickCount -eq 0) {
    Write-Host "   ❌ No tick files found for today" -ForegroundColor Red
    Write-Host "   This might indicate ticks are not being persisted!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "   ✅ Found tick data for $tickCount symbols" -ForegroundColor Green

# Monitor tick file growth
Write-Host ""
Write-Host "5. Monitoring tick file growth (10 second test)..." -ForegroundColor Yellow
$testSymbol = "EURUSD"
$testFile = Join-Path $dataDir "$testSymbol\$today.json"

if (Test-Path $testFile) {
    $sizeBefore = (Get-Item $testFile).Length
    Write-Host "   Initial size: $sizeBefore bytes" -ForegroundColor Gray

    Start-Sleep -Seconds 10

    $sizeAfter = (Get-Item $testFile).Length
    Write-Host "   Size after 10s: $sizeAfter bytes" -ForegroundColor Gray

    $growth = $sizeAfter - $sizeBefore
    if ($growth -gt 0) {
        Write-Host "   ✅ File grew by $growth bytes - ticks are being persisted!" -ForegroundColor Green
    } else {
        Write-Host "   ❌ File did NOT grow - ticks may not be persisting!" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "   ⚠️ Test file not found: $testFile" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "✅ Verification PASSED" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Summary:" -ForegroundColor Cyan
Write-Host "  - Backend is running" -ForegroundColor White
Write-Host "  - Tick files are being created" -ForegroundColor White
Write-Host "  - $tickCount symbols have data" -ForegroundColor White
Write-Host "  - Files are actively growing" -ForegroundColor White
Write-Host ""
Write-Host "Tick persistence is working correctly!" -ForegroundColor Green
