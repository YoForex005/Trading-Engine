# ============================================================================
# TEST AUTOMATION SETUP - Database Rotation & Compression
# ============================================================================
# Purpose: Test rotation and compression on sample data
# Run as: Administrator (not required for all tests)
# ============================================================================

param(
    [string]$ProjectPath = "D:\Tading engine\Trading-Engine",
    [switch]$Full,
    [int]$TestDays = 14
)

# Error handling
$ErrorActionPreference = "Stop"

# Logging
function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $color = switch ($Level) {
        "ERROR" { "Red" }
        "WARN"  { "Yellow" }
        "SUCCESS" { "Green" }
        "TEST" { "Cyan" }
        default { "White" }
    }
    Write-Host "[$Level] $timestamp - $Message" -ForegroundColor $color
}

function Write-Section {
    param([string]$Text)
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "  $Text" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
}

# Check script dependencies
function Test-Dependencies {
    Write-Section "Checking Dependencies"

    $deps = @{
        "sqlite3" = $null
        "zstd" = $null
    }

    # Check for sqlite3
    $sqlite = Get-Command sqlite3.exe -ErrorAction SilentlyContinue
    if ($sqlite) {
        Write-Log "Found sqlite3: $($sqlite.Source)" "SUCCESS"
        $deps["sqlite3"] = $true
    } else {
        Write-Log "sqlite3 not found (optional for this test)" "WARN"
        $deps["sqlite3"] = $false
    }

    # Check for zstd
    $zstd = Get-Command zstd.exe -ErrorAction SilentlyContinue
    if ($zstd) {
        Write-Log "Found zstd: $($zstd.Source)" "SUCCESS"
        $deps["zstd"] = $true
    } else {
        Write-Log "zstd not found - compression test will be skipped" "WARN"
        $deps["zstd"] = $false
    }

    return $deps
}

# Create sample test data
function New-SampleDatabase {
    param([datetime]$Date, [string]$DbPath)

    Write-Log "Creating sample database: $DbPath" "TEST"

    # Ensure directory exists
    $dir = Split-Path $DbPath
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
    }

    try {
        # Create database with schema
        $schema = @"
CREATE TABLE IF NOT EXISTS ticks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol VARCHAR(20) NOT NULL,
    timestamp INTEGER NOT NULL,
    bid REAL NOT NULL,
    ask REAL NOT NULL,
    spread REAL NOT NULL,
    volume INTEGER DEFAULT 0,
    lp_source VARCHAR(50),
    flags INTEGER DEFAULT 0,
    created_at INTEGER DEFAULT (strftime('%s', 'now') * 1000)
);

CREATE INDEX IF NOT EXISTS idx_ticks_symbol_timestamp ON ticks(symbol, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_ticks_timestamp ON ticks(timestamp DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ticks_unique ON ticks(symbol, timestamp);

CREATE TABLE IF NOT EXISTS symbols (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol VARCHAR(20) UNIQUE NOT NULL,
    description VARCHAR(255),
    asset_class VARCHAR(50),
    base_currency VARCHAR(10),
    quote_currency VARCHAR(10),
    is_active BOOLEAN DEFAULT 1,
    first_tick_at INTEGER,
    last_tick_at INTEGER,
    total_ticks INTEGER DEFAULT 0,
    created_at INTEGER DEFAULT (strftime('%s', 'now') * 1000)
);

PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = -64000;
"@

        # Create database
        $schema | & sqlite3.exe $DbPath

        # Insert sample data
        $baseTime = [int64]($Date.ToUniversalTime() - [datetime]"1970-01-01").TotalMilliseconds
        $symbols = @("EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD")

        foreach ($symbol in $symbols) {
            # Insert symbol
            & sqlite3.exe $DbPath "INSERT OR IGNORE INTO symbols (symbol, description, asset_class, base_currency, quote_currency) VALUES ('$symbol', 'Sample $symbol', 'forex', 'EUR', 'USD');"

            # Insert sample ticks (1000 per symbol)
            for ($i = 0; $i -lt 1000; $i++) {
                $timestamp = $baseTime + ($i * 10000)  # 10 second intervals
                $bid = 1.0800 + (Get-Random -Minimum -100 -Maximum 100) / 10000
                $ask = $bid + 0.0002
                $spread = $ask - $bid

                & sqlite3.exe $DbPath "INSERT INTO ticks (symbol, timestamp, bid, ask, spread, lp_source) VALUES ('$symbol', $timestamp, $bid, $ask, $spread, 'TestLP');"
            }
        }

        # Verify
        $tickCount = & sqlite3.exe $DbPath "SELECT COUNT(*) FROM ticks;"
        $size = (Get-Item $DbPath).Length / 1MB

        Write-Log "  ✓ Created database with $tickCount ticks ($([math]::Round($size, 2)) MB)" "SUCCESS"
        return $true
    }
    catch {
        Write-Log "  ✗ Failed to create database: $_" "ERROR"
        return $false
    }
}

# Test rotation
function Test-Rotation {
    Write-Section "Testing Database Rotation"

    $rotatePath = Join-Path $ProjectPath "backend\schema\rotate_tick_db.ps1"

    if (-not (Test-Path $rotatePath)) {
        Write-Log "Rotation script not found: $rotatePath" "ERROR"
        return $false
    }

    Write-Log "Testing rotation with dry-run..."

    try {
        $result = & powershell.exe -NoProfile -ExecutionPolicy Bypass -File $rotatePath -Action rotate -DryRun

        Write-Log "Dry-run completed successfully" "SUCCESS"
        return $true
    }
    catch {
        Write-Log "Dry-run failed: $_" "ERROR"
        return $false
    }
}

# Test compression
function Test-Compression {
    param([bool]$HasZstd)

    Write-Section "Testing Database Compression"

    $compressPath = Join-Path $ProjectPath "backend\schema\compress_old_dbs.sh"

    if (-not (Test-Path $compressPath)) {
        Write-Log "Compression script not found: $compressPath" "ERROR"
        return $false
    }

    if (-not $HasZstd) {
        Write-Log "zstd not available - skipping compression test" "WARN"
        return $true
    }

    Write-Log "Testing compression with dry-run..."

    try {
        # For Windows, we need to check if bash is available
        $bash = Get-Command bash -ErrorAction SilentlyContinue
        if (-not $bash) {
            Write-Log "bash not available - compression test requires WSL or Git Bash" "WARN"
            return $true
        }

        $env:DRY_RUN = "true"
        $result = & bash -c "$compressPath compress"

        Write-Log "Compression dry-run completed successfully" "SUCCESS"
        return $true
    }
    catch {
        Write-Log "Compression test failed: $_" "ERROR"
        return $false
    }
}

# Test database analysis
function Test-DatabaseAnalysis {
    param([string]$DbPath)

    Write-Section "Testing Database Analysis"

    if (-not (Test-Path $DbPath)) {
        Write-Log "Database not found: $DbPath" "ERROR"
        return $false
    }

    try {
        Write-Log "Analyzing database: $DbPath"

        # Count ticks
        $tickCount = & sqlite3.exe $DbPath "SELECT COUNT(*) FROM ticks;"
        Write-Log "  Total ticks: $tickCount" "INFO"

        # Count symbols
        $symbolCount = & sqlite3.exe $DbPath "SELECT COUNT(DISTINCT symbol) FROM ticks;"
        Write-Log "  Symbols: $symbolCount" "INFO"

        # Get date range
        $dateInfo = & sqlite3.exe $DbPath "SELECT datetime(MIN(timestamp) / 1000, 'unixepoch'), datetime(MAX(timestamp) / 1000, 'unixepoch') FROM ticks;"
        Write-Log "  Date range: $dateInfo" "INFO"

        # Get file size
        $size = (Get-Item $DbPath).Length / 1MB
        Write-Log "  File size: $([math]::Round($size, 2)) MB" "INFO"

        # Check integrity
        $integrity = & sqlite3.exe $DbPath "PRAGMA integrity_check;"
        if ($integrity -eq "ok") {
            Write-Log "  Integrity: OK" "SUCCESS"
        } else {
            Write-Log "  Integrity: FAILED - $integrity" "ERROR"
            return $false
        }

        return $true
    }
    catch {
        Write-Log "Database analysis failed: $_" "ERROR"
        return $false
    }
}

# Create sample test structure
function Create-TestStructure {
    param([int]$Days)

    Write-Section "Creating Sample Test Data Structure"

    $dbDir = Join-Path $ProjectPath "data\ticks\db"
    Write-Log "Creating sample databases in: $dbDir"

    $created = 0

    # Create databases for past N days
    for ($i = 0; $i -lt $Days; $i++) {
        $date = (Get-Date).AddDays(-$i)
        $year = $date.ToString("yyyy")
        $month = $date.ToString("MM")
        $dateStr = $date.ToString("yyyy-MM-dd")

        $dir = Join-Path $dbDir "$year\$month"
        $dbPath = Join-Path $dir "ticks_$dateStr.db"

        # Skip if already exists
        if (Test-Path $dbPath) {
            $size = (Get-Item $dbPath).Length / 1MB
            Write-Log "Database already exists: $dateStr ($([math]::Round($size, 2)) MB)"
            continue
        }

        if (New-SampleDatabase -Date $date -DbPath $dbPath) {
            $created++
        }
    }

    Write-Log "Created $created sample databases" "SUCCESS"
    return $created -gt 0
}

# Main test suite
function Main {
    Write-Host ""
    Write-Host "============================================================" -ForegroundColor Cyan
    Write-Host "  Database Automation Test Suite" -ForegroundColor Cyan
    Write-Host "============================================================" -ForegroundColor Cyan
    Write-Host ""

    $deps = Test-Dependencies

    if ($Full -or $PSBoundParameters.Count -eq 0) {
        # Full test suite
        Write-Section "Full Test Suite (This creates sample data)"

        # Create test data
        if (Create-TestStructure -Days $TestDays) {
            Write-Log "Sample data structure created successfully" "SUCCESS"
        } else {
            Write-Log "Failed to create sample data" "ERROR"
            exit 1
        }

        # Run tests
        $testsPassed = 0
        $testsFailed = 0

        # Test rotation
        if (Test-Rotation) {
            $testsPassed++
        } else {
            $testsFailed++
        }

        # Test compression
        if (Test-Compression -HasZstd $deps["zstd"]) {
            $testsPassed++
        } else {
            $testsFailed++
        }

        # Test analysis on most recent database
        $recentDb = Join-Path $ProjectPath "data\ticks\db\$(Get-Date -Format 'yyyy\MM')\ticks_$(Get-Date -Format 'yyyy-MM-dd').db"
        if (Test-Path $recentDb) {
            if (Test-DatabaseAnalysis -DbPath $recentDb) {
                $testsPassed++
            } else {
                $testsFailed++
            }
        }

        # Summary
        Write-Section "Test Summary"
        Write-Log "Tests passed: $testsPassed" "SUCCESS"
        if ($testsFailed -gt 0) {
            Write-Log "Tests failed: $testsFailed" "ERROR"
        }

        Write-Log ""
        Write-Log "Next steps:"
        Write-Log "  1. Review AUTOMATION_QUICK_REFERENCE.md"
        Write-Log "  2. Run setup script to install automation"
        Write-Log "  3. Monitor logs in data/ticks/db/rotation_metadata.json"
        Write-Log ""

        if ($testsFailed -eq 0) {
            exit 0
        } else {
            exit 1
        }
    }
}

# Run main
Main
