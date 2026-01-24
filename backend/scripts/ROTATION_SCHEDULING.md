# Tick Data Rotation - Scheduling Guide

This guide provides instructions for scheduling automated tick data rotation using either Unix cron (Linux/macOS) or Windows Task Scheduler.

## Overview

The rotation scripts manage tick data lifecycle with three phases:

1. **Archival**: Files older than 30 days (configurable) are moved to `backend/data/archive/` with date-based subdirectories
2. **Deletion**: Files in archive older than 90 days (configurable) are permanently deleted
3. **Cleanup**: Empty archive subdirectories are automatically removed

## Quick Start

### Test Run (Dry-Run Mode)

Before scheduling, always test in dry-run mode to see what would happen:

**Linux/macOS:**
```bash
cd backend/scripts
./rotate_ticks.sh --dry-run
```

**Windows PowerShell:**
```powershell
cd backend\scripts
.\rotate_ticks.ps1 -DryRun
```

### Configuration

Edit `backend/config/retention.yaml` to customize thresholds:

```yaml
retention:
  archive_threshold_days: 30    # Archive files older than 30 days
  deletion_threshold_days: 90   # Delete archived files older than 90 days

operations:
  enable_archival: true         # Enable moving old files to archive
  enable_deletion: true         # Enable deleting expired files
  compress_archives: true       # Compress files when archiving
```

---

## Linux/macOS - Cron Scheduling

### Option 1: Daily Rotation (Recommended)

Run rotation daily at 2 AM:

```bash
# Edit crontab
crontab -e

# Add this line
0 2 * * * cd /path/to/backend/scripts && ./rotate_ticks.sh >> /tmp/rotation.log 2>&1
```

### Option 2: Weekly Rotation

Run rotation every Sunday at 3 AM:

```bash
0 3 * * 0 cd /path/to/backend/scripts && ./rotate_ticks.sh >> /tmp/rotation.log 2>&1
```

### Option 3: Every 6 Hours

For more frequent rotation (suitable for high-volume trading):

```bash
0 */6 * * * cd /path/to/backend/scripts && ./rotate_ticks.sh >> /tmp/rotation.log 2>&1
```

### Cron Variables and Syntax

```
# ┌───────────── minute (0 - 59)
# │ ┌───────────── hour (0 - 23)
# │ │ ┌───────────── day of month (1 - 31)
# │ │ │ ┌───────────── month (1 - 12)
# │ │ │ │ ┌───────────── day of week (0 - 6) (Sunday to Saturday)
# │ │ │ │ │
# │ │ │ │ │
# * * * * *
```

### Verify Cron Job

```bash
# List all scheduled jobs
crontab -l

# View system cron logs (varies by OS)
# macOS:
log stream --predicate 'process == "cron"' --level debug

# Linux:
tail -f /var/log/cron

# Or check rotation log:
tail -f backend/logs/rotation.log
```

### Disable Cron Job

```bash
# Remove the cron job
crontab -e
# Delete the line, save and exit
```

---

## Windows - Task Scheduler

### Option 1: Create Task via GUI

1. **Open Task Scheduler:**
   - Press `Win + R`
   - Type `taskschd.msc` and press Enter

2. **Create Basic Task:**
   - Right-click "Task Scheduler Library" → "Create Task..."
   - **General Tab:**
     - Name: `Tick Data Rotation`
     - Description: `Automated tick data archival and cleanup`
     - Run whether user is logged in or not: ✓
     - Run with highest privileges: ✓

   - **Triggers Tab:**
     - Click "New..."
     - Begin the task: `On a schedule`
     - Daily, 2:00 AM ✓
     - Click OK

   - **Actions Tab:**
     - Click "New..."
     - Action: `Start a program`
     - Program/script: `powershell.exe`
     - Arguments: `-NoProfile -ExecutionPolicy Bypass -File "D:\Tading engine\Trading-Engine\backend\scripts\rotate_ticks.ps1"`
     - Start in: `D:\Tading engine\Trading-Engine\backend\scripts`
     - Click OK

   - **Conditions Tab:**
     - Uncheck "Start the task only if the computer is on AC power"

   - **Settings Tab:**
     - Allow task to be run on demand: ✓
     - If the task fails, restart: ✓
     - Retry every: 5 minutes

3. **Click OK and enter administrator credentials**

### Option 2: PowerShell Script to Create Task

Save as `create_rotation_task.ps1`:

```powershell
# Run as Administrator
$taskName = "Tick Data Rotation"
$scriptPath = "D:\Tading engine\Trading-Engine\backend\scripts\rotate_ticks.ps1"
$workingDir = "D:\Tading engine\Trading-Engine\backend\scripts"

# Create trigger (daily at 2 AM)
$trigger = New-ScheduledTaskTrigger -Daily -At 2:00AM

# Create action
$action = New-ScheduledTaskAction `
    -Execute "powershell.exe" `
    -Argument "-NoProfile -ExecutionPolicy Bypass -File `"$scriptPath`"" `
    -WorkingDirectory $workingDir

# Create settings
$settings = New-ScheduledTaskSettingsSet `
    -AllowStartIfOnBatteries `
    -Compatibility Win8 `
    -ExecutionTimeLimit 0:30:00 `
    -RestartCount 3 `
    -RestartInterval "00:05:00"

# Register the task
$principal = New-ScheduledTaskPrincipal `
    -UserID "NT AUTHORITY\SYSTEM" `
    -LogonType ServiceAccount `
    -RunLevel Highest

Register-ScheduledTask `
    -TaskName $taskName `
    -Trigger $trigger `
    -Action $action `
    -Settings $settings `
    -Principal $principal `
    -Force

Write-Host "Task '$taskName' created successfully!"
Write-Host "Scheduled to run daily at 2:00 AM"
```

Run with administrator privileges:

```powershell
# Open PowerShell as Administrator
cd D:\Tading engine\Trading-Engine\backend\scripts
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process -Force
.\create_rotation_task.ps1
```

### Option 3: Batch File with Task Scheduler

Save as `create_rotation_task.bat`:

```batch
@echo off
REM Run as Administrator to create the task

set TASK_NAME=Tick Data Rotation
set SCRIPT_PATH=D:\Tading engine\Trading-Engine\backend\scripts\rotate_ticks.ps1
set WORK_DIR=D:\Tading engine\Trading-Engine\backend\scripts

REM Delete existing task if it exists
taskkill /FI "TASKNAME eq Tick Data Rotation*" /IM powershell.exe /F 2>nul
schtasks /delete /tn "%TASK_NAME%" /f 2>nul

REM Create new task - runs daily at 2:00 AM
schtasks /create /tn "%TASK_NAME%" ^
    /tr "powershell -NoProfile -ExecutionPolicy Bypass -File \"%SCRIPT_PATH%\"" ^
    /sc daily /st 02:00 ^
    /ru SYSTEM /rp ^
    /rl HIGHEST

echo.
echo Task '%TASK_NAME%' has been created
echo Scheduled to run daily at 2:00 AM
pause
```

### Verify Task Creation

```powershell
# List all tasks with "Rotation" in the name
Get-ScheduledTask | Where-Object {$_.TaskName -like "*Rotation*"}

# Get full task details
Get-ScheduledTask -TaskName "Tick Data Rotation" | Select-Object * | Format-List

# Check task history
Get-ScheduledTaskInfo -TaskName "Tick Data Rotation"
```

### Run Task Manually

```powershell
# Start the task immediately
Start-ScheduledTask -TaskName "Tick Data Rotation"

# Monitor the task
Get-ScheduledTaskInfo -TaskName "Tick Data Rotation"
```

### Disable/Delete Task

```powershell
# Disable the task
Disable-ScheduledTask -TaskName "Tick Data Rotation"

# Delete the task
Unregister-ScheduledTask -TaskName "Tick Data Rotation" -Confirm:$false
```

---

## Monitoring and Logging

### View Logs

**All platforms:**
```
backend/logs/rotation.log
```

**Follow logs in real-time (Linux/macOS):**
```bash
tail -f backend/logs/rotation.log
```

**Follow logs in real-time (Windows PowerShell):**
```powershell
Get-Content backend\logs\rotation.log -Wait
```

### Log Format

Each line includes timestamp, level, and message:

```
2024-01-20 02:00:15 [INFO] Starting archival process...
2024-01-20 02:00:15 [INFO] Archive threshold: 30 days
2024-01-20 02:00:16 [INFO] Archived: ticks_2023-12-15.json (age: 36 days)
2024-01-20 02:00:20 [SUCCESS] Rotation script completed successfully
```

### Alert on Failures

**Linux/macOS - Email notifications:**

Add to crontab:
```bash
0 2 * * * cd /path/to/backend/scripts && ./rotate_ticks.sh >> /tmp/rotation.log 2>&1 || mail -s "Rotation Failed" admin@example.com < /tmp/rotation.log
```

**Windows - Event Log monitoring:**

Check Windows Event Viewer → Task Scheduler → History

---

## Optimization Tips

### High-Volume Trading

If you generate many ticks per day, increase rotation frequency:

**Every 6 hours (4 times daily):**
```bash
# Cron
0 */6 * * * cd /path/to/backend/scripts && ./rotate_ticks.sh

# PowerShell Task - Set 4 daily triggers at: 12:00 AM, 6:00 AM, 12:00 PM, 6:00 PM
```

### Large File Archival

For very large tick files, enable compression:

```yaml
operations:
  compress_archives: true
```

This reduces storage by 80-90% but takes slightly longer.

### Backup Strategy

Before enabling automatic deletion, ensure backups are in place:

```yaml
operations:
  create_backup: true          # Backup before deletion (if implemented)
  backup_before_delete: true
```

### Separate Archive Volumes

For production, consider moving archive to a different disk:

```yaml
paths:
  archive_directory: "/mnt/archive/ticks"  # Different volume
```

---

## Troubleshooting

### Script Not Running

**Linux/macOS:**
```bash
# Check if script has execute permission
ls -la backend/scripts/rotate_ticks.sh

# Make executable if needed
chmod +x backend/scripts/rotate_ticks.sh

# Test script directly
./backend/scripts/rotate_ticks.sh --help
```

**Windows:**
```powershell
# Check execution policy
Get-ExecutionPolicy

# Set to Bypass if needed (requires admin)
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope CurrentUser
```

### Permission Denied

**Linux/macOS:**
```bash
# Run with sudo
sudo crontab -e

# Or adjust file permissions
sudo chmod 755 backend/scripts/rotate_ticks.sh
sudo chmod 755 backend/logs
```

**Windows:**
- Ensure PowerShell runs as Administrator
- Task Scheduler must have "Run with highest privileges" enabled

### Files Not Archiving

1. Check log file: `backend/logs/rotation.log`
2. Run in dry-run mode to see what would happen
3. Verify paths in `backend/config/retention.yaml`
4. Check file timestamps: `ls -la backend/data/ticks/`

### Out of Disk Space

1. Lower `archive_threshold_days` to archive files sooner
2. Enable `compress_archives: true` for smaller files
3. Lower `deletion_threshold_days` to delete files sooner
4. Increase rotation frequency

---

## Additional Resources

- **Cron Reference:** https://crontab.guru
- **Windows Task Scheduler:** https://docs.microsoft.com/en-us/windows/win32/taskschd/task-scheduler-start-page
- **PowerShell ScheduledTask:** https://docs.microsoft.com/en-us/powershell/module/scheduledtasks/

---

## Support

For issues or questions:
1. Check logs: `backend/logs/rotation.log`
2. Review configuration: `backend/config/retention.yaml`
3. Run dry-run test: `./rotate_ticks.sh --dry-run`
4. Test permissions and paths manually
