#!/usr/bin/env python3
# ============================================================================
# DATABASE ROTATION & COMPRESSION MONITORING UTILITY
# ============================================================================
# Purpose: Monitor and manage database rotation and compression automation
# Features: Status monitoring, manual triggers, log analysis, health checks
# ============================================================================

import os
import sys
import json
import sqlite3
import subprocess
import argparse
import logging
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Dict, List, Optional, Tuple
import shutil

# Colors for terminal output
class Colors:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'

def print_header(text: str) -> None:
    """Print colored header text"""
    print(f"\n{Colors.HEADER}{Colors.BOLD}{'='*60}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}  {text}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{'='*60}{Colors.ENDC}\n")

def print_success(text: str) -> None:
    """Print success message"""
    print(f"{Colors.OKGREEN}[SUCCESS]{Colors.ENDC} {text}")

def print_warning(text: str) -> None:
    """Print warning message"""
    print(f"{Colors.WARNING}[WARN]{Colors.ENDC} {text}")

def print_error(text: str) -> None:
    """Print error message"""
    print(f"{Colors.FAIL}[ERROR]{Colors.ENDC} {text}")

def print_info(text: str) -> None:
    """Print info message"""
    print(f"{Colors.OKBLUE}[INFO]{Colors.ENDC} {text}")

class DatabaseRotationMonitor:
    """Monitor database rotation and compression automation"""

    def __init__(self, project_path: str, db_dir: str = None):
        """Initialize monitor"""
        self.project_path = Path(project_path)
        self.db_dir = Path(db_dir) if db_dir else self.project_path / "data" / "ticks" / "db"
        self.schema_dir = self.project_path / "backend" / "schema"
        self.log_dir = Path("/var/log/trading-engine") if os.name != "nt" else self.project_path / "logs"
        self.metadata_file = self.db_dir / "rotation_metadata.json"

    def check_system_status(self) -> bool:
        """Check if automation is properly set up"""
        print_header("System Status Check")

        checks_passed = 0
        checks_total = 0

        # Check 1: Script files exist
        checks_total += 1
        rotate_script = self.schema_dir / ("rotate_tick_db.ps1" if os.name == "nt" else "rotate_tick_db.sh")
        compress_script = self.schema_dir / "compress_old_dbs.sh"

        if rotate_script.exists():
            print_success(f"Rotation script exists: {rotate_script}")
            checks_passed += 1
        else:
            print_error(f"Rotation script missing: {rotate_script}")

        if compress_script.exists():
            print_success(f"Compression script exists: {compress_script}")
            checks_passed += 1
        else:
            print_error(f"Compression script missing: {compress_script}")

        # Check 2: DB directory exists
        checks_total += 1
        if self.db_dir.exists():
            print_success(f"Database directory exists: {self.db_dir}")
            checks_passed += 1
        else:
            print_warning(f"Database directory does not exist yet: {self.db_dir}")

        # Check 3: Metadata file
        checks_total += 1
        if self.metadata_file.exists():
            print_success(f"Rotation metadata found: {self.metadata_file}")
            checks_passed += 1
        else:
            print_warning("Rotation metadata not found (will be created on first rotation)")

        # Check 4: Cron/Task Scheduler
        checks_total += 1
        if os.name == "nt":
            result = subprocess.run(
                ["powershell.exe", "-Command", "Get-ScheduledTask | Where-Object {$_.TaskName -like '*Trading-Engine-DB*'} | Measure-Object | Select-Object -ExpandProperty Count"],
                capture_output=True,
                text=True
            )
            if result.returncode == 0 and result.stdout.strip() and int(result.stdout.strip()) > 0:
                print_success("Windows Task Scheduler tasks found")
                checks_passed += 1
            else:
                print_warning("Windows Task Scheduler tasks not found")
        else:
            result = subprocess.run(
                ["crontab", "-l"],
                capture_output=True,
                text=True,
                stderr=subprocess.DEVNULL
            )
            if result.returncode == 0 and "rotate_tick_db" in result.stdout:
                print_success("Cron jobs configured")
                checks_passed += 1
            else:
                print_warning("Cron jobs not configured")

        # Check 5: Disk space
        checks_total += 1
        if self.db_dir.exists():
            stat = shutil.disk_usage(self.db_dir)
            usage_percent = (stat.used / stat.total) * 100
            if usage_percent < 80:
                print_success(f"Disk usage: {usage_percent:.1f}% ({stat.used / (1024**3):.1f}GB / {stat.total / (1024**3):.1f}GB)")
                checks_passed += 1
            else:
                print_error(f"Disk usage high: {usage_percent:.1f}% - Consider archiving old databases")

        print(f"\n{Colors.BOLD}Summary: {checks_passed}/{checks_total} checks passed{Colors.ENDC}")
        return checks_passed == checks_total

    def get_database_inventory(self) -> Dict:
        """Get inventory of all databases"""
        print_header("Database Inventory")

        inventory = {
            "active": [],
            "compressed": [],
            "total_size_bytes": 0,
            "compressed_size_bytes": 0,
        }

        if not self.db_dir.exists():
            print_warning(f"Database directory does not exist: {self.db_dir}")
            return inventory

        # Find all databases
        active_dbs = sorted(self.db_dir.glob("*/*/*.db"))
        compressed_dbs = sorted(self.db_dir.glob("*/*/*.db.zst"))

        print_info(f"Active databases: {len(active_dbs)}")
        print_info(f"Compressed databases: {len(compressed_dbs)}")
        print()

        # Analyze active databases
        if active_dbs:
            print(f"{Colors.OKBLUE}Active Databases (last 10):{Colors.ENDC}")
            for db_file in active_dbs[-10:]:
                size_mb = db_file.stat().st_size / (1024 * 1024)
                mod_time = datetime.fromtimestamp(db_file.stat().st_mtime)
                inventory["active"].append({
                    "path": str(db_file),
                    "size_mb": size_mb,
                    "mtime": mod_time.isoformat()
                })
                inventory["total_size_bytes"] += db_file.stat().st_size
                print(f"  {db_file.name:30} {size_mb:8.2f} MB  {mod_time.strftime('%Y-%m-%d %H:%M:%S')}")

        # Analyze compressed databases
        if compressed_dbs:
            print(f"\n{Colors.OKBLUE}Compressed Databases (last 10):{Colors.ENDC}")
            for db_file in compressed_dbs[-10:]:
                size_mb = db_file.stat().st_size / (1024 * 1024)
                mod_time = datetime.fromtimestamp(db_file.stat().st_mtime)
                inventory["compressed"].append({
                    "path": str(db_file),
                    "size_mb": size_mb,
                    "mtime": mod_time.isoformat()
                })
                inventory["compressed_size_bytes"] += db_file.stat().st_size
                print(f"  {db_file.name:30} {size_mb:8.2f} MB  {mod_time.strftime('%Y-%m-%d %H:%M:%S')}")

        # Summary
        print()
        total_mb = inventory["total_size_bytes"] / (1024 * 1024)
        compressed_mb = inventory["compressed_size_bytes"] / (1024 * 1024)
        print(f"{Colors.BOLD}Storage Summary:{Colors.ENDC}")
        print(f"  Active databases:     {total_mb:10.2f} MB")
        print(f"  Compressed databases: {compressed_mb:10.2f} MB")
        print(f"  Total storage:        {(total_mb + compressed_mb):10.2f} MB")

        return inventory

    def get_rotation_metadata(self) -> Optional[Dict]:
        """Get rotation metadata"""
        if not self.metadata_file.exists():
            return None

        try:
            with open(self.metadata_file, 'r') as f:
                return json.load(f)
        except Exception as e:
            print_error(f"Failed to read metadata: {e}")
            return None

    def show_rotation_status(self) -> None:
        """Show rotation status"""
        print_header("Rotation Status")

        metadata = self.get_rotation_metadata()
        if metadata:
            print(f"{Colors.OKGREEN}Last Rotation:{Colors.ENDC}")
            print(f"  Time:        {metadata.get('last_rotation', 'N/A')}")
            print(f"  Current DB:  {metadata.get('current_date', 'N/A')}")
            print(f"  Previous DB: {metadata.get('previous_date', 'N/A')}")
        else:
            print_warning("No rotation metadata found - rotation may not have run yet")

    def analyze_database(self, db_path: str) -> Dict:
        """Analyze a specific database"""
        print_header(f"Database Analysis: {db_path}")

        analysis = {
            "valid": False,
            "tick_count": 0,
            "symbol_count": 0,
            "size_mb": 0,
            "date_range": None,
        }

        db_file = Path(db_path)
        if not db_file.exists():
            print_error(f"Database not found: {db_path}")
            return analysis

        analysis["size_mb"] = db_file.stat().st_size / (1024 * 1024)
        print_info(f"File size: {analysis['size_mb']:.2f} MB")

        try:
            conn = sqlite3.connect(db_path)
            cursor = conn.cursor()

            # Check integrity
            cursor.execute("PRAGMA integrity_check;")
            integrity = cursor.fetchone()[0]
            if integrity == "ok":
                print_success("Database integrity check passed")
                analysis["valid"] = True
            else:
                print_error(f"Database integrity check failed: {integrity}")

            # Count ticks
            cursor.execute("SELECT COUNT(*) FROM ticks;")
            analysis["tick_count"] = cursor.fetchone()[0]
            print_info(f"Total ticks: {analysis['tick_count']:,}")

            # Count symbols
            cursor.execute("SELECT COUNT(DISTINCT symbol) FROM ticks;")
            analysis["symbol_count"] = cursor.fetchone()[0]
            print_info(f"Symbols: {analysis['symbol_count']}")

            # Get date range
            cursor.execute("""
                SELECT
                    datetime(MIN(timestamp) / 1000, 'unixepoch') as first,
                    datetime(MAX(timestamp) / 1000, 'unixepoch') as last
                FROM ticks;
            """)
            first, last = cursor.fetchone()
            analysis["date_range"] = {"first": first, "last": last}
            print_info(f"Date range: {first} to {last}")

            # Get symbols
            cursor.execute("SELECT DISTINCT symbol FROM ticks ORDER BY symbol;")
            symbols = [row[0] for row in cursor.fetchall()]
            print_info(f"Symbols in database: {', '.join(symbols[:10])}" + ("..." if len(symbols) > 10 else ""))

            conn.close()
        except Exception as e:
            print_error(f"Failed to analyze database: {e}")

        return analysis

    def run_rotation_test(self, dry_run: bool = True) -> bool:
        """Test rotation script"""
        print_header("Database Rotation Test")

        rotate_script = self.schema_dir / ("rotate_tick_db.ps1" if os.name == "nt" else "rotate_tick_db.sh")

        if not rotate_script.exists():
            print_error(f"Rotation script not found: {rotate_script}")
            return False

        try:
            if os.name == "nt":
                cmd = [
                    "powershell.exe",
                    "-NoProfile",
                    "-ExecutionPolicy", "Bypass",
                    "-File", str(rotate_script),
                    "-Action", "rotate",
                ]
                if dry_run:
                    cmd.extend(["-DryRun"])
            else:
                cmd = [str(rotate_script), "rotate"]
                env = os.environ.copy()
                env["DRY_RUN"] = "true" if dry_run else "false"

            print_info(f"Running: {' '.join(cmd)}")
            result = subprocess.run(cmd, capture_output=True, text=True, env=env if os.name != "nt" else None)

            if result.returncode == 0:
                print_success("Rotation test passed")
                if result.stdout:
                    print(result.stdout)
                return True
            else:
                print_error("Rotation test failed")
                if result.stderr:
                    print(result.stderr)
                return False
        except Exception as e:
            print_error(f"Failed to run rotation test: {e}")
            return False

    def run_compression_test(self, dry_run: bool = True) -> bool:
        """Test compression script"""
        print_header("Database Compression Test")

        compress_script = self.schema_dir / "compress_old_dbs.sh"

        if not compress_script.exists():
            print_error(f"Compression script not found: {compress_script}")
            return False

        try:
            env = os.environ.copy()
            env["DRY_RUN"] = "true" if dry_run else "false"

            cmd = [str(compress_script), "compress"]
            print_info(f"Running: {' '.join(cmd)} (DRY_RUN={env['DRY_RUN']})")

            result = subprocess.run(cmd, capture_output=True, text=True, env=env)

            if result.returncode == 0:
                print_success("Compression test passed")
                if result.stdout:
                    print(result.stdout)
                return True
            else:
                print_error("Compression test failed")
                if result.stderr:
                    print(result.stderr)
                return False
        except Exception as e:
            print_error(f"Failed to run compression test: {e}")
            return False

    def show_retention_policy(self) -> None:
        """Show retention policy summary"""
        print_header("6-Month Retention Policy")

        print(f"{Colors.BOLD}Daily Rotation:{Colors.ENDC}")
        print("  • Time: Midnight UTC (00:00)")
        print("  • Action: Create new database for current day")
        print("  • Previous day: Closed and backed up")
        print()

        print(f"{Colors.BOLD}Weekly Compression:{Colors.ENDC}")
        print("  • Time: Sunday 02:00 UTC")
        print("  • Threshold: Databases older than 7 days")
        print("  • Method: zstd compression level 19")
        print("  • Compression ratio: Typically 4-5x")
        print()

        print(f"{Colors.BOLD}Monthly Archival:{Colors.ENDC}")
        print("  • Time: 1st of month at 03:00 UTC")
        print("  • Threshold: Databases older than 30 days")
        print("  • Action: Archive to cold storage")
        print()

        print(f"{Colors.BOLD}Retention Timeline:{Colors.ENDC}")
        print("  • 0-7 days:   Active (uncompressed) - Hot tier")
        print("  • 7-30 days:  Compressed - Warm tier")
        print("  • 30-180 days: Archived - Cold tier")
        print("  • 180+ days:  Deleted")

def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(
        description="Monitor database rotation and compression automation",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Check system status
  python automation_monitor.py status

  # Show database inventory
  python automation_monitor.py inventory

  # Analyze specific database
  python automation_monitor.py analyze data/ticks/db/2026/01/ticks_2026-01-20.db

  # Test rotation (dry run)
  python automation_monitor.py test-rotation --dry-run

  # Test compression (dry run)
  python automation_monitor.py test-compression --dry-run

  # Show retention policy
  python automation_monitor.py policy
        """
    )

    parser.add_argument(
        "command",
        choices=["status", "inventory", "analyze", "test-rotation", "test-compression", "policy"],
        help="Command to execute"
    )

    parser.add_argument(
        "--project-path",
        default=".",
        help="Project root path (default: current directory)"
    )

    parser.add_argument(
        "--db-path",
        help="Database path (for analyze command)"
    )

    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Run tests in dry-run mode"
    )

    args = parser.parse_args()

    # Create monitor
    monitor = DatabaseRotationMonitor(args.project_path)

    # Execute command
    if args.command == "status":
        monitor.check_system_status()
        monitor.show_rotation_status()
    elif args.command == "inventory":
        monitor.get_database_inventory()
    elif args.command == "analyze":
        if not args.db_path:
            print_error("Database path required for analyze command")
            sys.exit(1)
        monitor.analyze_database(args.db_path)
    elif args.command == "test-rotation":
        success = monitor.run_rotation_test(dry_run=args.dry_run or True)
        sys.exit(0 if success else 1)
    elif args.command == "test-compression":
        success = monitor.run_compression_test(dry_run=args.dry_run or True)
        sys.exit(0 if success else 1)
    elif args.command == "policy":
        monitor.show_retention_policy()

if __name__ == "__main__":
    main()
