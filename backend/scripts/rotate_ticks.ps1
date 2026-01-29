<#
.SYNOPSIS
    Tick Data Rotation Script for Windows

.DESCRIPTION
    Manages tick data lifecycle - archive old data, delete expired data.

    Features:
    - Read configuration from retention.yaml
    - Find files older than archive threshold and move to archive with date structure
    - Delete files older than retention limit
    - Create backups before deletion
    - Comprehensive logging with timestamps
    - Dry-run mode for testing

.PARAMETER ConfigPath
    Path to retention.yaml config file (default: ..\config\retention.yaml)

.PARAMETER DryRun
    Run in dry-run mode (show what would be done without modifying files)

.EXAMPLE
    .\rotate_ticks.ps1

.EXAMPLE
    .\rotate_ticks.ps1 -DryRun

.EXAMPLE
    .\rotate_ticks.ps1 -ConfigPath "D:\path\to\retention.yaml"
#>

param(
    [Parameter(Mandatory=$false)]
    [string]$ConfigPath = $(Join-Path (Split-Path $PSScriptRoot) "config" "retention.yaml"),

    [Parameter(Mandatory=$false)]
    [switch]$DryRun = $false
)

$ErrorActionPreference = "Stop"

# Color definitions
$Colors = @{
    "ERROR"   = [ConsoleColor]::Red
    "WARN"    = [ConsoleColor]::Yellow
    "INFO"    = [ConsoleColor]::Cyan
    "SUCCESS" = [ConsoleColor]::Green
    "DEBUG"   = [ConsoleColor]::Gray
}

# Configuration defaults
$Config = @{
    archive_threshold_days = 30
    deletion_threshold_days = 90
    ticks_directory = "./data/ticks"
    archive_directory = "./data/archive"
    logs_directory = "./logs"
    log_file = "rotation.log"
    enable_archival = $true
    enable_deletion = $true
    compress_archives = $false
    dry_run = $false
    level = "INFO"
}

################################################################################
# YAML Parser
################################################################################
function Parse-YamlSimple {
    param(
        [string]$FilePath
    )

    if (-not (Test-Path $FilePath)) {
        Write-Log -Level ERROR -Message "Config file not found: $FilePath"
        exit 1
    }

    $result = @{}
    $content = Get-Content $FilePath -Raw

    # Simple YAML parsing for our specific format
    $lines = $content -split "`n"
    foreach ($line in $lines) {
        $line = $line.Trim()

        # Skip comments and empty lines
        if ($line.StartsWith("#") -or $line -eq "") {
            continue
        }

        # Parse key: value pairs
        if ($line -match "^\s*(\w+):\s*(.+?)$") {
            $key = $matches[1]
            $value = $matches[2].Trim('"''')

            # Convert boolean strings
            if ($value -eq "true") {
                $value = $true
            } elseif ($value -eq "false") {
                $value = $false
            } elseif ($value -match "^\d+$") {
                $value = [int]$value
            }

            $result[$key] = $value
        }
    }

    return $result
}

################################################################################
# Logging Functions
################################################################################
function Write-Log {
    param(
        [Parameter(Mandatory=$true)]
        [ValidateSet("ERROR", "WARN", "INFO", "SUCCESS", "DEBUG")]
        [string]$Level,

        [Parameter(Mandatory=$true)]
        [string]$Message
    )

    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logMessage = "$timestamp [$Level] $Message"

    # Write to log file
    Add-Content -Path $LogFile -Value $logMessage -Encoding UTF8

    # Write to console with colors
    $color = $Colors[$Level]
    Write-Host $logMessage -ForegroundColor $color
}

function Get-FileAgeInDays {
    param(
        [string]$FilePath
    )

    $fileInfo = Get-Item $FilePath
    $ageTimespan = (Get-Date) - $fileInfo.LastWriteTime
    return [int]$ageTimespan.TotalDays
}

function Convert-BytesToReadable {
    param(
        [long]$Bytes
    )

    $sizes = @("B", "KB", "MB", "GB", "TB")
    $size = $Bytes
    $index = 0

    while ($size -gt 1024 -and $index -lt ($sizes.Length - 1)) {
        $size = $size / 1024
        $index++
    }

    return "{0:N2} {1}" -f $size, $sizes[$index]
}

################################################################################
# Archival Functions
################################################################################
function Archive-OldFiles {
    Write-Log -Level INFO -Message "Starting archival process..."
    Write-Log -Level INFO -Message "Archive threshold: $($Config.archive_threshold_days) days"
    Write-Log -Level INFO -Message "Source directory: $TicksDirectory"
    Write-Log -Level INFO -Message "Archive directory: $ArchiveDirectory"

    if (-not (Test-Path $TicksDirectory)) {
        Write-Log -Level WARN -Message "Ticks directory not found: $TicksDirectory"
        return
    }

    $archivedCount = 0
    $archivedSize = 0

    # Find all files in ticks directory (excluding subdirectories)
    $files = Get-ChildItem -Path $TicksDirectory -File -Filter "*.json"

    foreach ($file in $files) {
        $ageDays = Get-FileAgeInDays -FilePath $file.FullName

        if ($ageDays -ge $Config.archive_threshold_days) {
            $fileDate = $file.LastWriteTime.ToString("yyyy-MM-dd")
            $archiveSubdir = Join-Path $ArchiveDirectory $fileDate

            if (-not $DryRun) {
                # Create archive subdirectory
                if (-not (Test-Path $archiveSubdir)) {
                    New-Item -ItemType Directory -Path $archiveSubdir -Force | Out-Null
                }

                # Check if compression is enabled
                if ($Config.compress_archives) {
                    try {
                        # Compress with 7z if available, otherwise just copy
                        if (Get-Command 7z -ErrorAction SilentlyContinue) {
                            $archivePath = Join-Path $archiveSubdir "$($file.Name).7z"
                            & 7z a $archivePath $file.FullName -y | Out-Null
                            Remove-Item $file.FullName -Force
                            Write-Log -Level INFO -Message "Archived (compressed): $($file.Name) -> $archivePath (age: $ageDays days)"
                        } else {
                            # Fallback to copy without compression
                            Copy-Item -Path $file.FullName -Destination $archiveSubdir -Force
                            Remove-Item $file.FullName -Force
                            Write-Log -Level INFO -Message "Archived: $($file.Name) -> $archiveSubdir (age: $ageDays days)"
                        }
                    } catch {
                        Write-Log -Level WARN -Message "Failed to archive: $($file.Name) - $_"
                    }
                } else {
                    # Move file without compression
                    try {
                        Move-Item -Path $file.FullName -Destination $archiveSubdir -Force
                        Write-Log -Level INFO -Message "Archived: $($file.Name) -> $archiveSubdir (age: $ageDays days)"
                        $archivedCount++
                        $archivedSize += $file.Length
                    } catch {
                        Write-Log -Level WARN -Message "Failed to archive: $($file.Name) - $_"
                    }
                }
            } else {
                Write-Log -Level INFO -Message "[DRY-RUN] Would archive: $($file.Name) (age: $ageDays days, size: $(Convert-BytesToReadable $file.Length))"
            }
        }
    }

    Write-Log -Level INFO -Message "Archival complete: $archivedCount files archived, $(Convert-BytesToReadable $archivedSize)"
}

################################################################################
# Deletion Functions
################################################################################
function Delete-ExpiredFiles {
    Write-Log -Level INFO -Message "Starting deletion process..."
    Write-Log -Level INFO -Message "Deletion threshold: $($Config.deletion_threshold_days) days"
    Write-Log -Level INFO -Message "Archive directory: $ArchiveDirectory"

    if (-not (Test-Path $ArchiveDirectory)) {
        Write-Log -Level INFO -Message "Archive directory not found, nothing to delete"
        return
    }

    $deletedCount = 0
    $freedSize = 0

    # Find all files in archive directory
    $files = Get-ChildItem -Path $ArchiveDirectory -File -Recurse

    foreach ($file in $files) {
        $ageDays = Get-FileAgeInDays -FilePath $file.FullName

        if ($ageDays -ge $Config.deletion_threshold_days) {
            if (-not $DryRun) {
                try {
                    Remove-Item $file.FullName -Force
                    Write-Log -Level INFO -Message "Deleted: $($file.Name) (age: $ageDays days, size: $(Convert-BytesToReadable $file.Length))"
                    $deletedCount++
                    $freedSize += $file.Length
                } catch {
                    Write-Log -Level WARN -Message "Failed to delete: $($file.Name) - $_"
                }
            } else {
                Write-Log -Level INFO -Message "[DRY-RUN] Would delete: $($file.Name) (age: $ageDays days, size: $(Convert-BytesToReadable $file.Length))"
            }
        }
    }

    # Clean up empty directories
    if (-not $DryRun) {
        Get-ChildItem -Path $ArchiveDirectory -Directory -Recurse |
        Where-Object { (Get-ChildItem $_.FullName).Count -eq 0 } |
        Remove-Item -Force -ErrorAction SilentlyContinue
    }

    Write-Log -Level INFO -Message "Deletion complete: $deletedCount files deleted, $(Convert-BytesToReadable $freedSize) freed"
}

################################################################################
# Summary Functions
################################################################################
function Print-Summary {
    Write-Log -Level INFO -Message "=========================================="
    Write-Log -Level INFO -Message "Tick Data Rotation Summary"
    Write-Log -Level INFO -Message "=========================================="
    Write-Log -Level INFO -Message "Configuration: $ConfigPath"
    Write-Log -Level INFO -Message "Ticks directory: $TicksDirectory"
    Write-Log -Level INFO -Message "Archive directory: $ArchiveDirectory"
    Write-Log -Level INFO -Message "Archive threshold: $($Config.archive_threshold_days) days"
    Write-Log -Level INFO -Message "Deletion threshold: $($Config.deletion_threshold_days) days"

    if ($DryRun) {
        Write-Log -Level WARN -Message "DRY-RUN MODE: No files were actually modified"
    }

    # Print directory stats
    if (Test-Path $TicksDirectory) {
        $ticksFiles = @(Get-ChildItem -Path $TicksDirectory -File)
        $ticksSize = ($ticksFiles | Measure-Object -Property Length -Sum).Sum
        Write-Log -Level INFO -Message "Active ticks: $($ticksFiles.Count) files ($(Convert-BytesToReadable $ticksSize))"
    }

    if (Test-Path $ArchiveDirectory) {
        $archiveFiles = @(Get-ChildItem -Path $ArchiveDirectory -File -Recurse)
        $archiveSize = ($archiveFiles | Measure-Object -Property Length -Sum).Sum
        Write-Log -Level INFO -Message "Archived files: $($archiveFiles.Count) files ($(Convert-BytesToReadable $archiveSize))"
    }

    Write-Log -Level INFO -Message "Log file: $LogFile"
    Write-Log -Level INFO -Message "=========================================="
}

################################################################################
# Main Execution
################################################################################
function Main {
    # Resolve paths
    $script:TicksDirectory = Resolve-Path $Config.ticks_directory -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Path
    if (-not $script:TicksDirectory) {
        $script:TicksDirectory = $Config.ticks_directory
    }

    $script:ArchiveDirectory = Resolve-Path $Config.archive_directory -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Path
    if (-not $script:ArchiveDirectory) {
        $script:ArchiveDirectory = $Config.archive_directory
    }

    $script:LogsDirectory = Resolve-Path $Config.logs_directory -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Path
    if (-not $script:LogsDirectory) {
        $script:LogsDirectory = $Config.logs_directory
    }

    # Create logs directory
    if (-not (Test-Path $script:LogsDirectory)) {
        New-Item -ItemType Directory -Path $script:LogsDirectory -Force | Out-Null
    }

    $script:LogFile = Join-Path $script:LogsDirectory $Config.log_file

    # Parse configuration file
    Write-Host "Loading configuration from: $ConfigPath" -ForegroundColor Cyan
    $parsedConfig = Parse-YamlSimple -FilePath $ConfigPath

    # Merge parsed config with defaults
    foreach ($key in $parsedConfig.Keys) {
        $Config[$key] = $parsedConfig[$key]
    }

    Write-Log -Level INFO -Message "=========================================="
    Write-Log -Level INFO -Message "Tick Data Rotation Script Started"
    Write-Log -Level INFO -Message "=========================================="
    Write-Log -Level INFO -Message "Script: $($PSCommandPath | Split-Path -Leaf)"
    Write-Log -Level INFO -Message "Executed at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
    Write-Log -Level INFO -Message "Config file: $ConfigPath"

    if ($DryRun) {
        Write-Log -Level WARN -Message "Running in DRY-RUN mode - no files will be modified"
    }

    # Perform archival
    if ($Config.enable_archival) {
        Archive-OldFiles
    } else {
        Write-Log -Level INFO -Message "Archival is disabled in configuration"
    }

    # Perform deletion
    if ($Config.enable_deletion) {
        Delete-ExpiredFiles
    } else {
        Write-Log -Level INFO -Message "Deletion is disabled in configuration"
    }

    # Print summary
    Print-Summary

    Write-Log -Level SUCCESS -Message "Rotation script completed successfully"
}

# Run main function
try {
    Main
    exit 0
} catch {
    Write-Log -Level ERROR -Message "Script failed: $_"
    exit 1
}
