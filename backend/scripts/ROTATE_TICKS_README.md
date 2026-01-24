# Tick Data Rotation System

Automated management of tick data lifecycle with archival and deletion capabilities.

## Quick Start

### Test the Scripts

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

### Run Immediately

**Linux/macOS:**
```bash
./rotate_ticks.sh
```

**Windows PowerShell:**
```powershell
.\rotate_ticks.ps1
```

---

## What It Does

The rotation system manages tick data in three phases:

### 1. **Archival** (Default: 30 days)
- Scans `backend/data/ticks/` for JSON files
- Moves files older than 30 days to `backend/data/archive/YYYY-MM-DD/`
- Creates date-based subdirectories for organization
- Optionally compresses files (gzip/7z)
- Logs all operations with timestamps

### 2. **Deletion** (Default: 90 days)
- Scans archived files in `backend/data/archive/`
- Permanently deletes files older than 90 days
- Cleans up empty subdirectories
- Prevents disk space overflow

### 3. **Logging**
- All operations logged to `backend/logs/rotation.log`
- Timestamped entries with operation details
- Tracks file sizes, ages, and status
- Useful for auditing and troubleshooting

---

## Files Included

### Scripts

| File | Type | Purpose |
|------|------|---------|
| `rotate_ticks.sh` | Bash/Shell | Linux/macOS execution (341 lines, 12.8 KB) |
| `rotate_ticks.ps1` | PowerShell | Windows execution (391 lines, 14.4 KB) |

### Configuration

| File | Type | Purpose |
|------|------|---------|
| `../config/retention.yaml` | YAML | Thresholds, paths, and operation flags (72 lines, 1.8 KB) |

### Documentation

| File | Type | Purpose |
|------|------|---------|
| `ROTATION_SCHEDULING.md` | Markdown | Cron/Task Scheduler setup (438 lines) |
| `ROTATE_TICKS_README.md` | Markdown | This file - quick reference |

---

## Configuration

Edit `backend/config/retention.yaml`:

```yaml
retention:
  archive_threshold_days: 30    # Archive files older than this
  deletion_threshold_days: 90   # Delete archived files older than this

paths:
  ticks_directory: "./data/ticks"
  archive_directory: "./data/archive"
  logs_directory: "./logs"

operations:
  enable_archival: true         # Enable file archival
  enable_deletion: true         # Enable file deletion
  compress_archives: true       # Compress files before archival
  dry_run: false               # Set to true for testing
```

---

## Features

### Core Features
- **YAML Configuration**: Easy-to-read config file with sensible defaults
- **Configurable Thresholds**: Archive at 30 days, delete at 90 days (customizable)
- **Date-Based Organization**: Archived files organized by date (YYYY-MM-DD)
- **Optional Compression**: Reduce storage with gzip (bash) or 7z (PowerShell)
- **Comprehensive Logging**: Timestamped logs to `backend/logs/rotation.log`

### Safety Features
- **Dry-Run Mode**: Preview changes before executing
- **Path Validation**: Handles both absolute and relative paths
- **Error Handling**: Graceful failure with detailed error messages
- **Empty Directory Cleanup**: Removes empty subdirectories automatically

### Cross-Platform
- **Bash Version**: Linux, macOS, WSL
- **PowerShell Version**: Windows (native)
- **Same Features**: Both scripts provide identical functionality

---

## Command-Line Usage

### Bash Script

```bash
./rotate_ticks.sh [OPTIONS]

Options:
  --config CONFIG_FILE    Path to retention.yaml config file
  --dry-run              Run in dry-run mode (show what would happen)
  --help                 Show help message

Examples:
  ./rotate_ticks.sh                          # Run with defaults
  ./rotate_ticks.sh --dry-run               # Preview changes
  ./rotate_ticks.sh --config /path/to/config.yaml
```

### PowerShell Script

```powershell
.\rotate_ticks.ps1 [-ConfigPath <path>] [-DryRun]

Parameters:
  -ConfigPath <path>    Path to retention.yaml config file
  -DryRun              Run in dry-run mode (show what would happen)

Examples:
  .\rotate_ticks.ps1                          # Run with defaults
  .\rotate_ticks.ps1 -DryRun                 # Preview changes
  .\rotate_ticks.ps1 -ConfigPath "D:\path\to\config.yaml"
```

---

## Scheduling

### Linux/macOS - Cron

Add to crontab (`crontab -e`):

```bash
# Daily at 2 AM
0 2 * * * cd /path/to/backend/scripts && ./rotate_ticks.sh >> /tmp/rotation.log 2>&1

# Every 6 hours (for high-volume trading)
0 */6 * * * cd /path/to/backend/scripts && ./rotate_ticks.sh >> /tmp/rotation.log 2>&1

# Weekly on Sunday at 3 AM
0 3 * * 0 cd /path/to/backend/scripts && ./rotate_ticks.sh >> /tmp/rotation.log 2>&1
```

**Verify:**
```bash
crontab -l                    # List scheduled jobs
tail -f backend/logs/rotation.log  # Monitor logs
```

### Windows - Task Scheduler

#### Option A: GUI (Recommended)

1. Open Task Scheduler (`Win+R` → `taskschd.msc`)
2. Create Basic Task → Name: "Tick Data Rotation"
3. Trigger: Daily at 2:00 AM
4. Action: Start program
   - Program: `powershell.exe`
   - Arguments: `-NoProfile -ExecutionPolicy Bypass -File "D:\Tading engine\Trading-Engine\backend\scripts\rotate_ticks.ps1"`
5. Conditions: Uncheck "Start only on AC power"
6. Settings: Allow task to be run on demand ✓

**Verify:**
```powershell
Get-ScheduledTask -TaskName "Tick Data Rotation"
Get-ScheduledTaskInfo -TaskName "Tick Data Rotation"
```

#### Option B: PowerShell Script

Run as Administrator:

```powershell
$trigger = New-ScheduledTaskTrigger -Daily -At 2:00AM
$action = New-ScheduledTaskAction `
    -Execute "powershell.exe" `
    -Argument '-NoProfile -ExecutionPolicy Bypass -File "D:\Tading engine\Trading-Engine\backend\scripts\rotate_ticks.ps1"'
$settings = New-ScheduledTaskSettingsSet -ExecutionTimeLimit 0:30:00
$principal = New-ScheduledTaskPrincipal -UserID "NT AUTHORITY\SYSTEM" -LogonType ServiceAccount -RunLevel Highest

Register-ScheduledTask -TaskName "Tick Data Rotation" -Trigger $trigger -Action $action -Settings $settings -Principal $principal
```

#### Option C: Batch File

Save as `create_rotation_task.bat` and run as Administrator:

```batch
schtasks /create /tn "Tick Data Rotation" ^
    /tr "powershell -NoProfile -ExecutionPolicy Bypass -File \"D:\Tading engine\Trading-Engine\backend\scripts\rotate_ticks.ps1\"" ^
    /sc daily /st 02:00 ^
    /ru SYSTEM /rl HIGHEST
```

**Verify:**
```cmd
tasklist /FI "IMAGENAME eq svchost.exe"
schtasks /query /tn "Tick Data Rotation" /fo list /v
```

---

## Monitoring

### View Logs

```bash
# Stream logs (Linux/macOS)
tail -f backend/logs/rotation.log

# View logs (Windows PowerShell)
Get-Content backend\logs\rotation.log -Wait
```

### Log Format

```
2024-01-20 02:00:15 [INFO] Starting archival process...
2024-01-20 02:00:15 [INFO] Archive threshold: 30 days
2024-01-20 02:00:16 [INFO] Archived: ticks_2023-12-15.json -> ./data/archive/2023-12-15 (age: 36 days)
2024-01-20 02:00:20 [INFO] Archival complete: 5 files archived
2024-01-20 02:00:20 [INFO] Starting deletion process...
2024-01-20 02:00:25 [INFO] Deleted: archive/2023-09-01/ticks_old.json (age: 142 days)
2024-01-20 02:00:25 [SUCCESS] Rotation script completed successfully
```

### Directory Structure

```
backend/
├── data/
│   ├── ticks/                    # Active tick data
│   │   ├── 2024-01-20.json       # Recent files (not archived)
│   │   └── 2024-01-19.json
│   └── archive/                  # Archived tick data
│       ├── 2023-12-01/           # Date-based subdirectories
│       │   ├── ticks_2023-11-15.json.gz
│       │   └── ticks_2023-11-14.json.gz
│       └── 2023-11-01/
│           └── ticks_2023-10-01.json.gz
└── logs/
    └── rotation.log              # Operation log
```

---

## Performance Considerations

### Optimal Settings for Different Scenarios

#### Small Volume (< 100MB/day ticks)
```yaml
retention:
  archive_threshold_days: 60      # Archive after 2 months
  deletion_threshold_days: 180    # Delete after 6 months
operations:
  compress_archives: false        # No need to compress
```

**Frequency:** Daily at 2 AM

#### Medium Volume (100MB-1GB/day ticks)
```yaml
retention:
  archive_threshold_days: 30      # Archive after 1 month
  deletion_threshold_days: 90     # Delete after 3 months
operations:
  compress_archives: true         # Compress for storage
```

**Frequency:** Daily at 2 AM

#### High Volume (> 1GB/day ticks)
```yaml
retention:
  archive_threshold_days: 14      # Archive after 2 weeks
  deletion_threshold_days: 60     # Delete after 2 months
operations:
  compress_archives: true         # Important for storage
```

**Frequency:** Every 6 hours (4x daily)

### Storage Savings

With compression enabled:
- JSON tick files typically reduce by 80-90%
- Example: 1GB of ticks → ~100-200MB compressed

---

## Troubleshooting

### Files Not Being Archived

1. **Check log file:**
   ```bash
   tail -100 backend/logs/rotation.log
   ```

2. **Run dry-run to see what would happen:**
   ```bash
   ./rotate_ticks.sh --dry-run
   ```

3. **Verify file timestamps:**
   ```bash
   ls -la backend/data/ticks/
   stat backend/data/ticks/*.json
   ```

4. **Check paths in config:**
   ```yaml
   paths:
     ticks_directory: "./data/ticks"      # Correct relative path
     archive_directory: "./data/archive"
   ```

### Permission Denied

**Linux/macOS:**
```bash
# Make script executable
chmod +x backend/scripts/rotate_ticks.sh

# Check directory permissions
ls -la backend/data/
chmod 755 backend/data/archive
```

**Windows:**
```powershell
# Run PowerShell as Administrator
# Or set execution policy:
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope CurrentUser
```

### Cron Job Not Running

**Linux/macOS:**
```bash
# Verify cron is running
sudo systemctl status cron

# Check cron logs
sudo journalctl -u cron --since today

# Or check system logs
log stream --predicate 'process == "cron"' --level debug
```

### Task Scheduler Not Running

**Windows:**
```powershell
# Check task status
Get-ScheduledTaskInfo -TaskName "Tick Data Rotation"

# View event viewer logs
Get-WinEvent -FilterHashtable @{LogName='System'; ID=103}

# Run task manually to test
Start-ScheduledTask -TaskName "Tick Data Rotation"
```

### Out of Disk Space

1. Lower `archive_threshold_days` (archive sooner):
   ```yaml
   archive_threshold_days: 7       # Archive after 1 week instead of 30 days
   ```

2. Lower `deletion_threshold_days` (delete sooner):
   ```yaml
   deletion_threshold_days: 30     # Delete after 1 month instead of 90 days
   ```

3. Enable compression:
   ```yaml
   compress_archives: true
   ```

4. Increase rotation frequency:
   - Daily → Every 6 hours
   - Cron: `0 */6 * * * ...`
   - Task: Create 4 daily triggers

---

## Advanced Usage

### Custom Configuration Files

```bash
# Use alternate config file
./rotate_ticks.sh --config ./custom-retention.yaml
```

### Dry-Run Before Deployment

```bash
# Test without modifying files
./rotate_ticks.sh --dry-run

# Example output:
# [INFO] [DRY-RUN] Would archive: ticks_2023-12-15.json (age: 36 days)
# [INFO] [DRY-RUN] Would delete: archive/2023-09-01/ticks_old.json (age: 142 days)
```

### Integration with Monitoring

**Alert on failures:**
```bash
# If script fails, send email
./rotate_ticks.sh || mail -s "Rotation Failed" admin@example.com < backend/logs/rotation.log
```

### Backup Before Deletion

For additional safety, backup archive directory periodically:

```bash
# Weekly backup of archive
0 1 * * 0 tar -czf /backup/archive-backup-$(date +\%Y\%m\%d).tar.gz backend/data/archive/
```

---

## Support & Resources

### Documentation Files
- **ROTATION_SCHEDULING.md**: Detailed scheduling setup for cron and Task Scheduler
- **retention.yaml**: Configuration file with all available options
- **rotate_ticks.sh**: Bash script source code (well-commented)
- **rotate_ticks.ps1**: PowerShell script source code (well-commented)

### External Resources
- **Cron Syntax**: https://crontab.guru
- **Windows Task Scheduler**: https://docs.microsoft.com/en-us/windows/win32/taskschd/
- **PowerShell ScheduledTasks**: https://docs.microsoft.com/en-us/powershell/module/scheduledtasks/

### Testing Checklist

- [ ] Run scripts in dry-run mode
- [ ] Verify logs are generated
- [ ] Test with actual tick data
- [ ] Verify archived file organization
- [ ] Schedule for production environment
- [ ] Monitor first run
- [ ] Verify log file growth (normal)

---

## Version History

### v1.0 (2024-01-20)
- Initial release with bash and PowerShell scripts
- YAML configuration support
- Archive/deletion functionality
- Comprehensive logging
- Dry-run mode
- Scheduling documentation

---

## License & Disclaimer

These scripts are provided as-is for tick data management. Always test in a non-production environment first.

**Important:**
- Backup critical tick data before enabling automatic deletion
- Review logs regularly to monitor operations
- Test scheduling on a development system first
- Archive important historical data before deletion thresholds are reached
