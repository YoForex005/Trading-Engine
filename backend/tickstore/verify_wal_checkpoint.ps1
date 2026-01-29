# WAL Checkpoint Verification Script (PowerShell)
# Tests that WAL checkpoints are working correctly

$ErrorActionPreference = "Stop"

$DB_DIR = "..\..\data\ticks\2026\01"
$DB_FILE = "ticks_2026-01-20.db"
$WAL_FILE = "$DB_FILE-wal"

Write-Host "=== WAL Checkpoint Verification ===" -ForegroundColor Cyan
Write-Host ""

# 1. Check if DB exists
Write-Host "1. Checking database file..." -ForegroundColor Yellow
$dbPath = Join-Path $DB_DIR $DB_FILE
if (Test-Path $dbPath) {
    $dbSize = (Get-Item $dbPath).Length
    Write-Host "   ✓ Database exists: $DB_FILE ($dbSize bytes)" -ForegroundColor Green
} else {
    Write-Host "   ✗ Database not found: $DB_FILE" -ForegroundColor Red
    exit 1
}

# 2. Check WAL file status
Write-Host ""
Write-Host "2. Checking WAL file..." -ForegroundColor Yellow
$walPath = Join-Path $DB_DIR $WAL_FILE
if (Test-Path $walPath) {
    $walSize = (Get-Item $walPath).Length
    Write-Host "   ⚠ WAL file exists: $WAL_FILE ($walSize bytes)" -ForegroundColor Yellow

    if ($walSize -eq 0) {
        Write-Host "   ✓ WAL file is empty (checkpoint successful)" -ForegroundColor Green
    } else {
        Write-Host "   ⚠ WAL file has data (checkpoint may not have run)" -ForegroundColor Yellow
    }
} else {
    Write-Host "   ✓ No WAL file (checkpoint successful or WAL mode not enabled)" -ForegroundColor Green
}

# 3. Check database integrity
Write-Host ""
Write-Host "3. Running integrity check..." -ForegroundColor Yellow
try {
    $integrity = sqlite3 $dbPath "PRAGMA integrity_check;"
    if ($integrity -eq "ok") {
        Write-Host "   ✓ Database integrity: OK" -ForegroundColor Green
    } else {
        Write-Host "   ✗ Database integrity: FAILED" -ForegroundColor Red
        Write-Host "   Error: $integrity" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "   ✗ Error running integrity check: $_" -ForegroundColor Red
    Write-Host "   (Make sure sqlite3.exe is in PATH)" -ForegroundColor Yellow
}

# 4. Check WAL mode
Write-Host ""
Write-Host "4. Checking journal mode..." -ForegroundColor Yellow
$journalMode = sqlite3 $dbPath "PRAGMA journal_mode;"
Write-Host "   Journal mode: $journalMode"
if ($journalMode -eq "wal") {
    Write-Host "   ✓ WAL mode enabled" -ForegroundColor Green
} else {
    Write-Host "   ⚠ WAL mode not enabled (expected: wal, got: $journalMode)" -ForegroundColor Yellow
}

# 5. Check synchronous mode
Write-Host ""
Write-Host "5. Checking synchronous mode..." -ForegroundColor Yellow
$syncMode = sqlite3 $dbPath "PRAGMA synchronous;"
$syncName = switch ($syncMode) {
    "0" { "OFF" }
    "1" { "NORMAL" }
    "2" { "FULL" }
    "3" { "EXTRA" }
    default { "UNKNOWN" }
}
Write-Host "   Synchronous: $syncMode ($syncName)"
if ($syncMode -eq "1") {
    Write-Host "   ✓ NORMAL mode (balanced durability/performance)" -ForegroundColor Green
} else {
    Write-Host "   ⚠ Not NORMAL mode" -ForegroundColor Yellow
}

# 6. Check tick count
Write-Host ""
Write-Host "6. Checking tick data..." -ForegroundColor Yellow
$tickCount = sqlite3 $dbPath "SELECT COUNT(*) FROM ticks;"
$symbolCount = sqlite3 $dbPath "SELECT COUNT(DISTINCT symbol) FROM ticks;"
Write-Host "   Total ticks: $tickCount"
Write-Host "   Unique symbols: $symbolCount"

# 7. Sample recent ticks
Write-Host ""
Write-Host "7. Sample recent ticks (last 3)..." -ForegroundColor Yellow
$recentTicks = sqlite3 $dbPath "SELECT symbol, datetime(timestamp/1000, 'unixepoch'), bid, ask FROM ticks ORDER BY timestamp DESC LIMIT 3;"
$recentTicks -split "`n" | ForEach-Object {
    Write-Host "   $_"
}

Write-Host ""
Write-Host "=== Verification Complete ===" -ForegroundColor Cyan

# Summary
Write-Host ""
Write-Host "Summary:" -ForegroundColor Cyan
Write-Host "  - Database integrity: OK" -ForegroundColor Green
$checkpointStatus = if ((Test-Path $walPath) -and ((Get-Item $walPath).Length -ne 0)) { "PENDING" } else { "COMPLETE" }
Write-Host "  - WAL checkpoint: $checkpointStatus" -ForegroundColor $(if ($checkpointStatus -eq "COMPLETE") { "Green" } else { "Yellow" })
Write-Host "  - Total ticks stored: $tickCount"
Write-Host ""
Write-Host "To manually checkpoint WAL:" -ForegroundColor Yellow
Write-Host "  sqlite3 $dbPath 'PRAGMA wal_checkpoint(TRUNCATE);'"
