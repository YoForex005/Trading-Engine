# Database Rotation & Compression Automation - Quick Reference Guide

## Overview

This guide covers the automated database rotation and compression system that implements a **6-month retention policy** for tick data. The system automatically rotates daily databases and compresses old data while preserving historical access.

## Architecture

```
Data Flow:
  Active Trading (24h)
      ↓
  Daily Rotation at Midnight UTC
      ↓
  7-Day Hot Storage (Active DBs)
      ↓
  Weekly Compression on Sunday
      ↓
  30-Day Warm Storage (Compressed)
      ↓
  Monthly Archival
      ↓
  Cold Storage (180 days)
      ↓
  Automatic Deletion (>180 days)
```

## File Organization

```
data/ticks/db/
├── YYYY/
│   ├── MM/
│   │   ├── ticks_YYYY-MM-DD.db        # Active databases (< 7 days)
│   │   ├── ticks_YYYY-MM-DD.db.zst    # Compressed databases (7-180 days)
│   │   └── rotation_metadata.json
│   └── [other months]
├── backup/                              # Daily backups of previous day's DB
│   └── YYYY/MM/
│       └── ticks_YYYY-MM-DD.db
└── archive/                             # Monthly archives (>30 days)
    └── YYYY-MM/
        └── ticks_YYYY-MM-DD.db.zst
```

## Automation Scripts

### 1. Rotation Script

**What it does:**
- Creates new database at midnight UTC
- Closes previous database (WAL checkpoint)
- Backs up closed database
- Pre-creates tomorrow's database for faster rotation

**Files:**
- **Windows:** `backend/schema/rotate_tick_db.ps1`
- **Linux/macOS:** `backend/schema/rotate_tick_db.sh`

**Manual execution:**

```bash
# Windows (PowerShell)
.\backend\schema\rotate_tick_db.ps1 -Action rotate
.\backend\schema\rotate_tick_db.ps1 -Action status

# Dry run to see what would happen
.\backend\schema\rotate_tick_db.ps1 -Action rotate -DryRun

# Linux/macOS (Bash)
./backend/schema/rotate_tick_db.sh rotate
./backend/schema/rotate_tick_db.sh status

# Dry run
DRY_RUN=true ./backend/schema/rotate_tick_db.sh rotate
```

### 2. Compression Script

**What it does:**
- Finds databases older than 7 days
- Compresses with zstd (level 19)
- Achieves 4-5x compression ratio
- Maintains data integrity through checksums

**File:** `backend/schema/compress_old_dbs.sh`

**Manual execution:**

```bash
# Linux/macOS
./backend/schema/compress_old_dbs.sh compress

# Dry run
DRY_RUN=true ./backend/schema/compress_old_dbs.sh compress

# Decompress a specific database
./backend/schema/compress_old_dbs.sh decompress data/ticks/db/2026/01/ticks_2026-01-10.db.zst

# List all compressed databases
./backend/schema/compress_old_dbs.sh list
```

### 3. Monitoring Utility

**What it does:**
- Checks system setup status
- Inventory of databases
- Database analysis (tick counts, date ranges)
- Tests rotation/compression in dry-run mode

**File:** `backend/schema/automation_monitor.py`

**Usage:**

```bash
# Python 3 required
python3 automation_monitor.py status           # System status check
python3 automation_monitor.py inventory        # List all databases
python3 automation_monitor.py policy           # Show retention policy
python3 automation_monitor.py test-rotation --dry-run    # Test rotation
python3 automation_monitor.py test-compression --dry-run # Test compression
python3 automation_monitor.py analyze <db_path>         # Analyze database
```

## Setup Instructions

### Windows Setup (Automated)

1. **Open PowerShell as Administrator:**
   ```powershell
   # Start PowerShell as Administrator
   # Go to project directory
   cd "D:\Tading engine\Trading-Engine"
   ```

2. **Run setup script:**
   ```powershell
   .\backend\schema\setup_windows_scheduler.ps1
   ```

3. **Verify installation:**
   ```powershell
   # List scheduled tasks
   Get-ScheduledTask | Where-Object {$_.TaskName -like "*Trading-Engine-DB*"}
   ```

4. **Uninstall (if needed):**
   ```powershell
   .\backend\schema\setup_windows_scheduler.ps1 -Uninstall
   ```

### Linux/macOS Setup (Automated)

1. **Run setup script as root:**
   ```bash
   cd /path/to/Trading-Engine
   sudo ./backend/schema/setup_linux_cron.sh install
   ```

2. **Verify installation:**
   ```bash
   crontab -l  # Show current cron jobs
   ```

3. **Check logs:**
   ```bash
   tail -f /var/log/trading-engine/rotation.log
   tail -f /var/log/trading-engine/compression.log
   ```

4. **Uninstall (if needed):**
   ```bash
   sudo ./backend/schema/setup_linux_cron.sh uninstall
   ```

### Manual Setup (If automation tools unavailable)

#### Windows Task Scheduler

```powershell
# Create daily rotation task at midnight UTC
$trigger = New-ScheduledTaskTrigger -Daily -At "00:00"
$action = New-ScheduledTaskAction -Execute "powershell.exe" `
    -Argument "-File 'D:\Tading engine\Trading-Engine\backend\schema\rotate_tick_db.ps1' -Action rotate"
$principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -RunLevel Highest
Register-ScheduledTask -TaskName "Trading-DB-Rotation" -Trigger $trigger `
    -Action $action -Principal $principal
```

#### Linux/macOS Cron

```bash
# Edit crontab
sudo crontab -e

# Add these lines:
# Daily rotation at midnight UTC
0 0 * * * /path/to/Trading-Engine/backend/schema/rotate_tick_db.sh rotate >> /var/log/tick-rotation.log 2>&1

# Weekly compression on Sunday at 2 AM UTC
0 2 * * 0 /path/to/Trading-Engine/backend/schema/compress_old_dbs.sh compress >> /var/log/tick-compression.log 2>&1

# Hourly status check
0 * * * * /path/to/Trading-Engine/backend/schema/rotate_tick_db.sh status >> /var/log/tick-status.log 2>&1
```

## Automation Schedule

### Daily Tasks (Automatic)

| Time | Task | Script | Duration |
|------|------|--------|----------|
| 00:00 UTC | Database Rotation | `rotate_tick_db.*` | ~10-30s |
| Every hour | Status Check | `rotate_tick_db.* status` | ~5s |

### Weekly Tasks (Automatic)

| Time | Task | Script | Duration | Typical Size Reduction |
|------|------|--------|----------|----------------------|
| Sunday 02:00 UTC | Compression | `compress_old_dbs.sh` | ~1-5m per DB | 4-5x (20% original) |

### Monthly Tasks (Manual recommended)

| Time | Task | Notes |
|------|------|-------|
| 1st of month | Archival | Move 30+ day old databases to archive storage |
| End of month | Cleanup | Delete 180+ day old databases |

## Retention Policy Details

### Hot Tier (0-7 days)
- **Storage:** Uncompressed SQLite databases
- **Performance:** Real-time access, full query support
- **Access pattern:** Active trading queries
- **Size per symbol:** ~5-10 MB per day (typical)
- **Total space:** 100 symbols = ~500-1000 MB/week

### Warm Tier (7-30 days)
- **Storage:** Compressed with zstd (level 19)
- **Compression ratio:** 4-5x typical
- **Size per symbol:** ~1-2 MB per day compressed
- **Access pattern:** Historical analysis, overnight jobs
- **Total space:** 100 symbols = ~100-200 MB/month compressed

### Cold Tier (30-180 days)
- **Storage:** Archived to external/cold storage
- **Access:** On-demand decompression required
- **Total space:** Depends on archive strategy
- **Use case:** Regulatory compliance, long-term analysis

### Deletion (>180 days)
- **Action:** Automatic deletion or archival to deep storage
- **Retention:** Configurable (default: 180 days ≈ 6 months)
- **Compliance:** Adjust per regulatory requirements

## Monitoring and Health Checks

### Check Rotation Status

```bash
# PowerShell (Windows)
.\backend\schema\rotate_tick_db.ps1 -Action status

# Bash (Linux/macOS)
./backend/schema/rotate_tick_db.sh status

# Python (All platforms)
python3 backend/schema/automation_monitor.py status
```

### View Database Inventory

```bash
python3 backend/schema/automation_monitor.py inventory
```

### Analyze Specific Database

```bash
python3 backend/schema/automation_monitor.py analyze data/ticks/db/2026/01/ticks_2026-01-20.db
```

### Check Disk Usage

```bash
# Linux/macOS
du -sh data/ticks/db/*/
df -h data/ticks/db/

# Windows PowerShell
Get-ChildItem "data\ticks\db" -Recurse -File | Measure-Object -Property Length -Sum
```

## Testing and Troubleshooting

### Test Rotation (Dry Run)

```bash
# PowerShell
.\backend\schema\rotate_tick_db.ps1 -Action rotate -DryRun

# Bash
DRY_RUN=true ./backend/schema/rotate_tick_db.sh rotate

# Python
python3 backend/schema/automation_monitor.py test-rotation --dry-run
```

### Test Compression (Dry Run)

```bash
# Bash
DRY_RUN=true ./backend/schema/compress_old_dbs.sh compress

# Python
python3 backend/schema/automation_monitor.py test-compression --dry-run
```

### View Recent Logs

```bash
# Windows Event Viewer
eventvwr.msc
# Look for "Trading-Engine-DB-Rotation" tasks

# Linux/macOS
tail -50 /var/log/trading-engine/rotation.log
tail -50 /var/log/trading-engine/compression.log
tail -50 /var/log/trading-engine/status.log
```

### Common Issues

#### Issue: Database rotation not running

**Symptoms:**
- Same database file for multiple days
- Rotation task not executing

**Solutions:**
1. Check if task/cron job is enabled:
   ```powershell
   # Windows
   Get-ScheduledTask "Trading-Engine-DB-Rotation" | Select-Object State
   ```
   ```bash
   # Linux
   crontab -l | grep rotate_tick_db
   ```

2. Check disk space:
   ```bash
   df -h  # Linux/macOS
   ```

3. Verify script permissions:
   ```bash
   ls -l backend/schema/rotate_tick_db.sh
   chmod +x backend/schema/rotate_tick_db.sh
   ```

#### Issue: Compression not working

**Symptoms:**
- Old databases not being compressed
- High disk usage

**Solutions:**
1. Check zstd is installed:
   ```bash
   which zstd  # Linux/macOS
   zstd --version
   ```

2. Install if missing:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install zstd

   # macOS
   brew install zstd

   # Windows: Download from https://github.com/facebook/zstd/releases
   ```

3. Test compression manually:
   ```bash
   DRY_RUN=true ./backend/schema/compress_old_dbs.sh compress
   ```

#### Issue: Tasks/Crons not executing automatically

**Symptoms:**
- Manual execution works, but scheduled tasks don't run
- No log files created

**Solutions:**
1. Check service/daemon is running:
   ```bash
   # Linux
   sudo systemctl status cron
   sudo systemctl restart cron

   # macOS
   sudo launchctl load /System/Library/LaunchDaemons/com.openssh.sshd.plist
   ```

2. Check task history:
   ```powershell
   # Windows
   Get-ScheduledTaskInfo -TaskName "Trading-Engine-DB-Rotation"
   ```

3. Run manual test:
   ```bash
   python3 backend/schema/automation_monitor.py test-rotation --dry-run
   ```

## Performance Benchmarks

### Rotation Performance

- **Database creation:** ~10-30 seconds
- **WAL checkpoint:** ~5-10 seconds
- **Backup copy:** ~5-15 seconds (depends on DB size)
- **Total time:** ~30 seconds for typical 100-200 MB database

### Compression Performance

- **Per database:** ~1-5 minutes (depends on size)
- **Throughput:** ~50-200 MB/sec compression
- **Decompression:** ~100-300 MB/sec
- **Typical compression ratio:** 4-5x (20% original size)

### Storage Impact

- **7-day hot storage:** 500-1000 MB (100 symbols)
- **3-week warm storage:** 100-200 MB compressed
- **6-month total:** ~1-2 GB active + archive

## Advanced Configuration

### Change Compression Threshold

```bash
# Compress databases older than 14 days instead of 7
DAYS_BEFORE_COMPRESS=14 ./backend/schema/compress_old_dbs.sh compress
```

### Change Compression Level

```bash
# Use faster compression (level 10 instead of 19)
COMPRESSION_LEVEL=10 ./backend/schema/compress_old_dbs.sh compress

# Use maximum compression (level 22, slower)
COMPRESSION_LEVEL=22 ./backend/schema/compress_old_dbs.sh compress
```

### Disable Backups (not recommended)

```bash
# Skip backup during rotation
ENABLE_BACKUP=false ./backend/schema/rotate_tick_db.sh rotate
```

### Verbose Output

```bash
# Enable detailed logging
VERBOSE=true ./backend/schema/rotate_tick_db.sh rotate
```

## Disaster Recovery

### Restore from Backup

```bash
# Find backup for date
ls -la data/ticks/backup/2026/01/ticks_2026-01-20.db

# Restore to active location
cp data/ticks/backup/2026/01/ticks_2026-01-20.db \
   data/ticks/db/2026/01/ticks_2026-01-20.db.restored
```

### Decompress Archived Database

```bash
# Decompress a zstd file
./backend/schema/compress_old_dbs.sh decompress \
  data/ticks/db/2026/01/ticks_2026-01-10.db.zst
```

### Verify Database Integrity

```bash
# Check database consistency
sqlite3 data/ticks/db/2026/01/ticks_2026-01-20.db "PRAGMA integrity_check;"

# Vacuum and optimize
sqlite3 data/ticks/db/2026/01/ticks_2026-01-20.db "VACUUM; ANALYZE;"
```

## Compliance and Auditing

### Retention Proof

```bash
# List all databases with dates
find data/ticks/db -name "ticks_*.db*" | sort | \
  xargs ls -lh | awk '{print $9, $5, $6, $7, $8}'
```

### Compression Audit

```bash
# Show compression ratios
find data/ticks/db -name "*.db.zst" | while read f; do
  uncompressed="${f%.zst}"
  if [ -f "$uncompressed" ]; then
    orig=$(stat -f%z "$uncompressed" 2>/dev/null || stat -c%s "$uncompressed" 2>/dev/null)
    comp=$(stat -f%z "$f" 2>/dev/null || stat -c%s "$f" 2>/dev/null)
    ratio=$(echo "scale=2; $orig / $comp" | bc)
    echo "$(basename $f): ${ratio}x compression"
  fi
done
```

### Audit Log

```bash
# Show rotation history
cat data/ticks/db/rotation_metadata.json | python3 -m json.tool

# Show compression log
tail -100 /var/log/trading-engine/compression.log
```

## Support and Documentation

- **Schema documentation:** `backend/schema/ticks.sql`
- **Rotation script help:** `./rotate_tick_db.sh help`
- **Compression script help:** `./compress_old_dbs.sh help`
- **Monitor help:** `python3 automation_monitor.py --help`

## Summary Checklist

- [ ] Scripts are executable (`chmod +x *.sh`)
- [ ] Automation setup complete (Windows Tasks or cron)
- [ ] Initial test rotation successful (dry-run passes)
- [ ] Initial test compression successful (dry-run passes)
- [ ] Log directories created and writable
- [ ] Disk space verified (at least 1-2 GB free)
- [ ] zstd installed (for compression)
- [ ] First rotation has completed and metadata created
- [ ] Compression ran on first scheduled time
- [ ] Monitoring utility can connect and check status
- [ ] Retention policy understood by operations team
- [ ] Disaster recovery procedures documented
