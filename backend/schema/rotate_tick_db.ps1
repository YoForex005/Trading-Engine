# ============================================================================
# ROTATE TICK DATABASE DAILY (PowerShell for Windows)
# ============================================================================
# Purpose: Create new daily database at midnight and close previous day's DB
# Schedule: Run at 00:00 UTC via Task Scheduler
# ============================================================================

param(
    [string]$Action = "rotate",
    [string]$DbDir = "data\ticks\db",
    [string]$SchemaFile = "backend\schema\ticks.sql",
    [string]$BackupDir = "data\ticks\backup",
    [switch]$EnableBackup = $true,
    [switch]$Verbose = $true,
    [switch]$DryRun = $false
)

# Error handling
$ErrorActionPreference = "Stop"

# Logging functions
function Write-Log {
    param([string]$Message, [string]$Level = "INFO")

    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $color = switch ($Level) {
        "ERROR" { "Red" }
        "WARN"  { "Yellow" }
        "SUCCESS" { "Green" }
        default { "White" }
    }

    if ($Verbose -or $Level -ne "INFO") {
        Write-Host "[$Level] $timestamp - $Message" -ForegroundColor $color
    }
}

# Check dependencies
function Test-Dependencies {
    # Check for SQLite
    $sqlitePath = (Get-Command sqlite3.exe -ErrorAction SilentlyContinue).Source
    if (-not $sqlitePath) {
        Write-Log "SQLite not found in PATH. Checking local installation..." "WARN"

        # Try common installation paths
        $commonPaths = @(
            "$env:ProgramFiles\SQLite\sqlite3.exe",
            "$env:ProgramFiles(x86)\SQLite\sqlite3.exe",
            ".\sqlite3.exe"
        )

        foreach ($path in $commonPaths) {
            if (Test-Path $path) {
                $script:SqlitePath = $path
                Write-Log "Found SQLite at: $path" "SUCCESS"
                return $true
            }
        }

        Write-Log "SQLite not found. Please install SQLite and add to PATH." "ERROR"
        Write-Log "Download from: https://www.sqlite.org/download.html" "ERROR"
        return $false
    }

    $script:SqlitePath = $sqlitePath
    Write-Log "SQLite found at: $sqlitePath" "SUCCESS"

    # Check schema file
    if (-not (Test-Path $SchemaFile)) {
        Write-Log "Schema file not found: $SchemaFile" "ERROR"
        return $false
    }

    return $true
}

# Get database path for a specific date
function Get-DbPath {
    param([datetime]$Date)

    $year = $Date.ToString("yyyy")
    $month = $Date.ToString("MM")
    $dateStr = $Date.ToString("yyyy-MM-dd")

    return "$DbDir\$year\$month\ticks_$dateStr.db"
}

# Create directory structure
function New-Directories {
    param([datetime]$Date)

    $year = $Date.ToString("yyyy")
    $month = $Date.ToString("MM")
    $dir = "$DbDir\$year\$month"

    if ($DryRun) {
        Write-Log "[DRY RUN] Would create directory: $dir" "INFO"
        return
    }

    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
        Write-Log "Created directory: $dir" "SUCCESS"
    }

    if ($EnableBackup) {
        $backupPath = "$BackupDir\$year\$month"
        if (-not (Test-Path $backupPath)) {
            New-Item -ItemType Directory -Path $backupPath -Force | Out-Null
            Write-Log "Created backup directory: $backupPath" "SUCCESS"
        }
    }
}

# Create new database with schema
function New-Database {
    param([string]$DbPath)

    if (Test-Path $DbPath) {
        Write-Log "Database already exists: $DbPath" "WARN"
        return $true
    }

    if ($DryRun) {
        Write-Log "[DRY RUN] Would create database: $DbPath" "INFO"
        return $true
    }

    Write-Log "Creating database: $DbPath" "INFO"

    try {
        # Create database and apply schema
        Get-Content $SchemaFile | & $SqlitePath $DbPath
        Write-Log "  Schema applied successfully" "SUCCESS"

        # Set performance pragmas
        $pragmas = @(
            "PRAGMA journal_mode = WAL;",
            "PRAGMA synchronous = NORMAL;",
            "PRAGMA cache_size = -64000;",
            "PRAGMA temp_store = MEMORY;",
            "PRAGMA mmap_size = 268435456;"
        )

        foreach ($pragma in $pragmas) {
            & $SqlitePath $DbPath $pragma | Out-Null
        }
        Write-Log "  Performance pragmas configured" "SUCCESS"

        # Verify database
        $integrityCheck = & $SqlitePath $DbPath "PRAGMA integrity_check;"
        if ($integrityCheck -ne "ok") {
            Write-Log "  Database integrity check failed" "ERROR"
            return $false
        }
        Write-Log "  Database integrity verified" "SUCCESS"

        # Get file size
        $size = (Get-Item $DbPath).Length / 1KB
        Write-Log "  Database created successfully (size: $([math]::Round($size, 2)) KB)" "SUCCESS"

        return $true
    }
    catch {
        Write-Log "  Failed to create database: $_" "ERROR"
        return $false
    }
}

# Close previous database (checkpoint WAL)
function Close-Database {
    param([string]$DbPath)

    if (-not (Test-Path $DbPath)) {
        Write-Log "Database not found: $DbPath" "WARN"
        return $true
    }

    if ($DryRun) {
        Write-Log "[DRY RUN] Would close database: $DbPath" "INFO"
        return $true
    }

    Write-Log "Closing database: $DbPath" "INFO"

    try {
        # Checkpoint WAL to merge into main database
        & $SqlitePath $DbPath "PRAGMA wal_checkpoint(TRUNCATE);" | Out-Null
        Write-Log "  WAL checkpoint completed" "SUCCESS"

        # Verify integrity
        $integrityCheck = & $SqlitePath $DbPath "PRAGMA integrity_check;"
        if ($integrityCheck -ne "ok") {
            Write-Log "  Final integrity check failed" "ERROR"
            return $false
        }
        Write-Log "  Final integrity check passed" "SUCCESS"

        # Get statistics
        $tickCount = & $SqlitePath $DbPath "SELECT COUNT(*) FROM ticks;"
        $symbolCount = & $SqlitePath $DbPath "SELECT COUNT(DISTINCT symbol) FROM ticks;"
        $size = (Get-Item $DbPath).Length / 1MB

        Write-Log "  Database statistics:" "INFO"
        Write-Log "    - Total ticks: $tickCount" "INFO"
        Write-Log "    - Symbols: $symbolCount" "INFO"
        Write-Log "    - Size: $([math]::Round($size, 2)) MB" "INFO"

        return $true
    }
    catch {
        Write-Log "  Failed to close database: $_" "ERROR"
        return $false
    }
}

# Backup database
function Backup-Database {
    param(
        [string]$DbPath,
        [datetime]$Date
    )

    if (-not $EnableBackup) {
        Write-Log "Backup disabled, skipping" "INFO"
        return $true
    }

    if (-not (Test-Path $DbPath)) {
        Write-Log "Database not found for backup: $DbPath" "WARN"
        return $true
    }

    $year = $Date.ToString("yyyy")
    $month = $Date.ToString("MM")
    $dateStr = $Date.ToString("yyyy-MM-dd")
    $backupPath = "$BackupDir\$year\$month\ticks_$dateStr.db"

    if ($DryRun) {
        Write-Log "[DRY RUN] Would backup to: $backupPath" "INFO"
        return $true
    }

    Write-Log "Backing up database to: $backupPath" "INFO"

    try {
        # Use SQLite backup command for safe copy
        & $SqlitePath $DbPath ".backup '$backupPath'" | Out-Null

        if (Test-Path $backupPath) {
            $size = (Get-Item $backupPath).Length / 1MB
            Write-Log "  Backup completed (size: $([math]::Round($size, 2)) MB)" "SUCCESS"
            return $true
        }
        else {
            Write-Log "  Backup failed - file not created" "ERROR"
            return $false
        }
    }
    catch {
        Write-Log "  Backup failed: $_" "ERROR"
        return $false
    }
}

# Update rotation metadata
function Update-Metadata {
    param(
        [datetime]$Today,
        [datetime]$Yesterday
    )

    $metadataFile = "$DbDir\rotation_metadata.json"

    if ($DryRun) {
        Write-Log "[DRY RUN] Would update metadata: $metadataFile" "INFO"
        return
    }

    $metadata = @{
        last_rotation = (Get-Date -Format "yyyy-MM-dd HH:mm:ss UTC")
        current_date = $Today.ToString("yyyy-MM-dd")
        previous_date = $Yesterday.ToString("yyyy-MM-dd")
        current_db = Get-DbPath -Date $Today
        previous_db = Get-DbPath -Date $Yesterday
    }

    $metadata | ConvertTo-Json | Set-Content -Path $metadataFile
    Write-Log "Metadata updated: $metadataFile" "SUCCESS"
}

# Main rotation process
function Invoke-Rotation {
    $today = Get-Date
    $yesterday = $today.AddDays(-1)

    Write-Log "=== Daily Database Rotation ===" "INFO"
    Write-Log "Today: $($today.ToString('yyyy-MM-dd'))" "INFO"
    Write-Log "Yesterday: $($yesterday.ToString('yyyy-MM-dd'))" "INFO"
    Write-Log "" "INFO"

    # Step 1: Create directories
    Write-Log "Step 1: Creating directory structure" "INFO"
    New-Directories -Date $today
    Write-Log "" "INFO"

    # Step 2: Close yesterday's database
    Write-Log "Step 2: Closing yesterday's database" "INFO"
    $yesterdayDb = Get-DbPath -Date $yesterday
    if (-not (Close-Database -DbPath $yesterdayDb)) {
        Write-Log "Failed to close yesterday's database" "ERROR"
        return $false
    }
    Write-Log "" "INFO"

    # Step 3: Backup yesterday's database
    Write-Log "Step 3: Backing up yesterday's database" "INFO"
    if (-not (Backup-Database -DbPath $yesterdayDb -Date $yesterday)) {
        Write-Log "Failed to backup yesterday's database" "ERROR"
        return $false
    }
    Write-Log "" "INFO"

    # Step 4: Create today's database
    Write-Log "Step 4: Creating today's database" "INFO"
    $todayDb = Get-DbPath -Date $today
    if (-not (New-Database -DbPath $todayDb)) {
        Write-Log "Failed to create today's database" "ERROR"
        return $false
    }
    Write-Log "" "INFO"

    # Step 5: Update metadata
    Write-Log "Step 5: Updating rotation metadata" "INFO"
    Update-Metadata -Today $today -Yesterday $yesterday
    Write-Log "" "INFO"

    # Step 6: Pre-create tomorrow's database
    $tomorrow = $today.AddDays(1)
    Write-Log "Step 6: Pre-creating tomorrow's database ($($tomorrow.ToString('yyyy-MM-dd')))" "INFO"
    New-Directories -Date $tomorrow
    $tomorrowDb = Get-DbPath -Date $tomorrow
    New-Database -DbPath $tomorrowDb | Out-Null
    Write-Log "" "INFO"

    Write-Log "=== Rotation Complete ===" "SUCCESS"
    Write-Log "Successfully rotated to: $todayDb" "SUCCESS"

    return $true
}

# Status check
function Get-RotationStatus {
    $today = Get-Date
    $todayDb = Get-DbPath -Date $today

    Write-Log "=== Database Rotation Status ===" "INFO"
    Write-Log "Current date: $($today.ToString('yyyy-MM-dd'))" "INFO"
    Write-Log "Expected database: $todayDb" "INFO"

    if (Test-Path $todayDb) {
        Write-Log "Today's database exists" "SUCCESS"

        try {
            # Check if database is accessible
            $testQuery = & $SqlitePath $todayDb "SELECT 1;"
            Write-Log "Database is accessible" "SUCCESS"

            # Get statistics
            $tickCount = & $SqlitePath $todayDb "SELECT COUNT(*) FROM ticks;"
            $symbolCount = & $SqlitePath $todayDb "SELECT COUNT(DISTINCT symbol) FROM ticks;"

            Write-Log "Statistics:" "INFO"
            Write-Log "  - Ticks: $tickCount" "INFO"
            Write-Log "  - Symbols: $symbolCount" "INFO"
        }
        catch {
            Write-Log "Database is not accessible: $_" "ERROR"
            return $false
        }
    }
    else {
        Write-Log "Today's database does not exist" "ERROR"
        Write-Log "Run rotation: .\rotate_tick_db.ps1 -Action rotate" "WARN"
        return $false
    }

    # Check metadata
    $metadataFile = "$DbDir\rotation_metadata.json"
    if (Test-Path $metadataFile) {
        Write-Log "" "INFO"
        Write-Log "Last rotation metadata:" "INFO"
        Get-Content $metadataFile | Write-Host
    }

    return $true
}

# Main entry point
function Main {
    Write-Host ""
    Write-Host "============================================================" -ForegroundColor Cyan
    Write-Host "  Tick Database Rotation Tool (PowerShell)" -ForegroundColor Cyan
    Write-Host "============================================================" -ForegroundColor Cyan
    Write-Host ""

    # Check dependencies
    if (-not (Test-Dependencies)) {
        exit 1
    }

    Write-Host ""

    # Execute action
    $success = switch ($Action.ToLower()) {
        "rotate" {
            Invoke-Rotation
        }
        "status" {
            Get-RotationStatus
        }
        "help" {
            Write-Host @"
Tick Database Rotation Tool (PowerShell)

USAGE:
    .\rotate_tick_db.ps1 [-Action <action>] [options]

ACTIONS:
    rotate       Perform daily database rotation (default)
    status       Check rotation status
    help         Show this help message

OPTIONS:
    -DbDir          Database directory (default: data\ticks\db)
    -SchemaFile     Schema SQL file (default: backend\schema\ticks.sql)
    -BackupDir      Backup directory (default: data\ticks\backup)
    -EnableBackup   Enable backups (default: true)
    -Verbose        Verbose output (default: true)
    -DryRun         Dry run mode (default: false)

EXAMPLES:
    # Perform rotation
    .\rotate_tick_db.ps1 -Action rotate

    # Check status
    .\rotate_tick_db.ps1 -Action status

    # Dry run
    .\rotate_tick_db.ps1 -Action rotate -DryRun

TASK SCHEDULER:
    # Create scheduled task for daily rotation at midnight
    `$trigger = New-ScheduledTaskTrigger -Daily -At "00:00"
    `$action = New-ScheduledTaskAction -Execute "powershell.exe" -Argument "-File C:\path\to\rotate_tick_db.ps1 -Action rotate"
    Register-ScheduledTask -TaskName "TickDBRotation" -Trigger `$trigger -Action `$action -User "SYSTEM"

"@
            $true
        }
        default {
            Write-Log "Unknown action: $Action" "ERROR"
            Write-Log "Run '.\rotate_tick_db.ps1 -Action help' for usage" "ERROR"
            $false
        }
    }

    Write-Host ""

    if ($success) {
        exit 0
    }
    else {
        exit 1
    }
}

# Run main function
Main
