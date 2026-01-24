# ============================================================================
# WINDOWS TASK SCHEDULER SETUP FOR DATABASE ROTATION & COMPRESSION
# ============================================================================
# Purpose: Automatically create scheduled tasks for daily DB rotation and
#          weekly compression following the 6-month retention policy
# Run As: Administrator
# ============================================================================

param(
    [switch]$Uninstall,
    [switch]$DryRun,
    [string]$ProjectPath = "D:\Tading engine\Trading-Engine"
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
        default { "White" }
    }
    Write-Host "[$Level] $timestamp - $Message" -ForegroundColor $color
}

# Check admin privileges
function Test-AdminPrivileges {
    $isAdmin = [Security.Principal.WindowsIdentity]::GetCurrent().Groups -contains 'S-1-5-32-544'
    if (-not $isAdmin) {
        Write-Log "ERROR: This script requires Administrator privileges!" "ERROR"
        Write-Log "Please run PowerShell as Administrator and try again." "ERROR"
        exit 1
    }
    Write-Log "Admin privileges verified" "SUCCESS"
}

# Verify script paths exist
function Test-ScriptPaths {
    param([string]$ProjectPath)

    $rotatePath = Join-Path $ProjectPath "backend\schema\rotate_tick_db.ps1"
    $compressPath = Join-Path $ProjectPath "backend\schema\compress_old_dbs.sh"

    if (-not (Test-Path $rotatePath)) {
        Write-Log "ERROR: Rotation script not found: $rotatePath" "ERROR"
        return $false
    }

    Write-Log "Found rotation script: $rotatePath" "SUCCESS"
    return $true
}

# Create daily rotation task at midnight UTC
function New-RotationTask {
    param([string]$ProjectPath)

    $taskName = "Trading-Engine-DB-Rotation"
    $rotatePath = Join-Path $ProjectPath "backend\schema\rotate_tick_db.ps1"

    Write-Log "Creating scheduled task: $taskName" "INFO"

    # Check if task already exists
    $existingTask = Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue
    if ($existingTask) {
        Write-Log "Task already exists, removing..." "WARN"
        if (-not $DryRun) {
            Unregister-ScheduledTask -TaskName $taskName -Confirm:$false
        }
    }

    # Create trigger (Daily at midnight UTC)
    $trigger = New-ScheduledTaskTrigger -Daily -At "00:00"

    # Create action (PowerShell script execution)
    $action = New-ScheduledTaskAction `
        -Execute "powershell.exe" `
        -Argument "-NoProfile -ExecutionPolicy Bypass -File `"$rotatePath`" -Action rotate"

    # Create task settings
    $settings = New-ScheduledTaskSettingsSet `
        -AllowStartIfOnBatteries `
        -DontStopIfGoingOnBatteries `
        -StartWhenAvailable `
        -RunOnlyIfNetworkAvailable `
        -RunOnlyIfIdle:$false

    # Create principal (SYSTEM account)
    $principal = New-ScheduledTaskPrincipal `
        -UserId "SYSTEM" `
        -LogonType ServiceAccount `
        -RunLevel Highest

    # Register the task
    if ($DryRun) {
        Write-Log "[DRY RUN] Would create task with:" "INFO"
        Write-Log "  Name: $taskName" "INFO"
        Write-Log "  Trigger: Daily at 00:00 UTC" "INFO"
        Write-Log "  Script: $rotatePath" "INFO"
        return $true
    }

    try {
        Register-ScheduledTask `
            -TaskName $taskName `
            -Trigger $trigger `
            -Action $action `
            -Settings $settings `
            -Principal $principal `
            -Description "Daily database rotation for trading tick storage" `
            -Force | Out-Null

        Write-Log "Created rotation task: $taskName" "SUCCESS"
        Write-Log "  Schedule: Daily at 00:00 UTC" "INFO"
        Write-Log "  Run as: SYSTEM" "INFO"
        return $true
    }
    catch {
        Write-Log "Failed to create rotation task: $_" "ERROR"
        return $false
    }
}

# Create weekly compression task
function New-CompressionTask {
    param([string]$ProjectPath)

    $taskName = "Trading-Engine-DB-Compression"
    $compressPath = Join-Path $ProjectPath "backend\schema\compress_old_dbs.sh"

    Write-Log "Creating scheduled task: $taskName" "INFO"

    # Check if task already exists
    $existingTask = Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue
    if ($existingTask) {
        Write-Log "Task already exists, removing..." "WARN"
        if (-not $DryRun) {
            Unregister-ScheduledTask -TaskName $taskName -Confirm:$false
        }
    }

    # Create trigger (Weekly on Sunday at 02:00 UTC)
    $trigger = New-ScheduledTaskTrigger `
        -Weekly `
        -DaysOfWeek Sunday `
        -At "02:00"

    # For Windows, we need a PowerShell wrapper or use WSL if available
    # This creates a batch file wrapper that can call bash via Git Bash or WSL
    $wslAction = New-ScheduledTaskAction `
        -Execute "powershell.exe" `
        -Argument "-NoProfile -ExecutionPolicy Bypass -Command `"bash '$compressPath'`""

    # Create task settings
    $settings = New-ScheduledTaskSettingsSet `
        -AllowStartIfOnBatteries `
        -DontStopIfGoingOnBatteries `
        -StartWhenAvailable `
        -RunOnlyIfNetworkAvailable:$false `
        -RunOnlyIfIdle:$false

    # Create principal (SYSTEM account)
    $principal = New-ScheduledTaskPrincipal `
        -UserId "SYSTEM" `
        -LogonType ServiceAccount `
        -RunLevel Highest

    # Register the task
    if ($DryRun) {
        Write-Log "[DRY RUN] Would create task with:" "INFO"
        Write-Log "  Name: $taskName" "INFO"
        Write-Log "  Trigger: Weekly on Sunday at 02:00 UTC" "INFO"
        Write-Log "  Script: $compressPath" "INFO"
        Write-Log "  Note: Requires bash (WSL, Git Bash, or native bash)" "WARN"
        return $true
    }

    try {
        Register-ScheduledTask `
            -TaskName $taskName `
            -Trigger $trigger `
            -Action $wslAction `
            -Settings $settings `
            -Principal $principal `
            -Description "Weekly compression of databases older than 7 days" `
            -Force | Out-Null

        Write-Log "Created compression task: $taskName" "SUCCESS"
        Write-Log "  Schedule: Weekly on Sunday at 02:00 UTC" "INFO"
        Write-Log "  Run as: SYSTEM" "INFO"
        Write-Log "  Note: Requires bash (WSL, Git Bash, or native bash)" "WARN"
        return $true
    }
    catch {
        Write-Log "Failed to create compression task: $_" "ERROR"
        return $false
    }
}

# Create status check task (hourly)
function New-StatusCheckTask {
    param([string]$ProjectPath)

    $taskName = "Trading-Engine-DB-Status-Check"
    $rotatePath = Join-Path $ProjectPath "backend\schema\rotate_tick_db.ps1"

    Write-Log "Creating scheduled task: $taskName" "INFO"

    # Check if task already exists
    $existingTask = Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue
    if ($existingTask) {
        Write-Log "Task already exists, removing..." "WARN"
        if (-not $DryRun) {
            Unregister-ScheduledTask -TaskName $taskName -Confirm:$false
        }
    }

    # Create trigger (Hourly)
    $trigger = New-ScheduledTaskTrigger `
        -Once `
        -At (Get-Date) `
        -RepetitionInterval (New-TimeSpan -Hours 1) `
        -RepetitionDuration (New-TimeSpan -Days 30)

    # Create action
    $action = New-ScheduledTaskAction `
        -Execute "powershell.exe" `
        -Argument "-NoProfile -ExecutionPolicy Bypass -File `"$rotatePath`" -Action status"

    # Create task settings
    $settings = New-ScheduledTaskSettingsSet `
        -AllowStartIfOnBatteries `
        -DontStopIfGoingOnBatteries `
        -StartWhenAvailable `
        -RunOnlyIfNetworkAvailable:$false `
        -RunOnlyIfIdle:$false

    # Create principal (SYSTEM account)
    $principal = New-ScheduledTaskPrincipal `
        -UserId "SYSTEM" `
        -LogonType ServiceAccount `
        -RunLevel Highest

    # Register the task
    if ($DryRun) {
        Write-Log "[DRY RUN] Would create task with:" "INFO"
        Write-Log "  Name: $taskName" "INFO"
        Write-Log "  Trigger: Hourly" "INFO"
        Write-Log "  Script: $rotatePath -Action status" "INFO"
        return $true
    }

    try {
        Register-ScheduledTask `
            -TaskName $taskName `
            -Trigger $trigger `
            -Action $action `
            -Settings $settings `
            -Principal $principal `
            -Description "Hourly status check for database rotation" `
            -Force | Out-Null

        Write-Log "Created status check task: $taskName" "SUCCESS"
        Write-Log "  Schedule: Every hour" "INFO"
        Write-Log "  Run as: SYSTEM" "INFO"
        return $true
    }
    catch {
        Write-Log "Failed to create status check task: $_" "ERROR"
        return $false
    }
}

# List all tasks
function Get-AllTasks {
    Write-Log "Listing all Trading Engine database tasks:" "INFO"
    Write-Host ""

    $tasks = Get-ScheduledTask | Where-Object {$_.TaskName -like "*Trading-Engine-DB*"}

    if ($tasks.Count -eq 0) {
        Write-Log "No Trading Engine database tasks found" "WARN"
        return
    }

    foreach ($task in $tasks) {
        Write-Log "Task: $($task.TaskName)" "INFO"
        Write-Log "  State: $($task.State)" "INFO"
        Write-Log "  Last Run: $($task.LastRunTime)" "INFO"
        Write-Log "  Last Result: $($task.LastTaskResult)" "INFO"
        Write-Host ""
    }
}

# Uninstall all tasks
function Uninstall-AllTasks {
    Write-Log "Removing all Trading Engine database tasks..." "WARN"

    $tasks = Get-ScheduledTask | Where-Object {$_.TaskName -like "*Trading-Engine-DB*"}

    if ($tasks.Count -eq 0) {
        Write-Log "No tasks to remove" "INFO"
        return $true
    }

    if ($DryRun) {
        Write-Log "[DRY RUN] Would remove $($tasks.Count) task(s)" "INFO"
        foreach ($task in $tasks) {
            Write-Log "  - $($task.TaskName)" "INFO"
        }
        return $true
    }

    foreach ($task in $tasks) {
        try {
            Unregister-ScheduledTask -TaskName $task.TaskName -Confirm:$false
            Write-Log "Removed: $($task.TaskName)" "SUCCESS"
        }
        catch {
            Write-Log "Failed to remove $($task.TaskName): $_" "ERROR"
            return $false
        }
    }

    return $true
}

# Main function
function Main {
    Write-Host ""
    Write-Host "============================================================" -ForegroundColor Cyan
    Write-Host "  Windows Task Scheduler Setup for Database Automation" -ForegroundColor Cyan
    Write-Host "============================================================" -ForegroundColor Cyan
    Write-Host ""

    # Check admin privileges
    Test-AdminPrivileges

    # Verify script paths
    if (-not (Test-ScriptPaths -ProjectPath $ProjectPath)) {
        exit 1
    }

    Write-Host ""

    if ($Uninstall) {
        Write-Log "Uninstalling mode: Removing all scheduled tasks" "WARN"
        $success = Uninstall-AllTasks
    }
    else {
        Write-Log "Installation mode: Creating scheduled tasks" "INFO"
        Write-Host ""

        $success = $true

        # Create rotation task
        if (-not (New-RotationTask -ProjectPath $ProjectPath)) {
            $success = $false
        }
        Write-Host ""

        # Create compression task
        if (-not (New-CompressionTask -ProjectPath $ProjectPath)) {
            $success = $false
        }
        Write-Host ""

        # Create status check task
        if (-not (New-StatusCheckTask -ProjectPath $ProjectPath)) {
            $success = $false
        }
        Write-Host ""

        # List all tasks
        Get-AllTasks
    }

    Write-Host ""
    if ($success) {
        Write-Log "Setup completed successfully!" "SUCCESS"
        exit 0
    }
    else {
        Write-Log "Setup completed with errors" "ERROR"
        exit 1
    }
}

# Run main function
Main
