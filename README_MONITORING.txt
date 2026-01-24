================================================================================
                    QUOTE STREAMING RESOURCE MONITORING
                              Complete Setup
================================================================================

OVERVIEW
========
This package provides comprehensive resource monitoring for quote streaming
load tests. It automatically tracks CPU, memory, threads, and system stability.

FILES INCLUDED
==============

1. QUICK_START.bat
   - Interactive menu to run tests easily
   - No command-line knowledge required
   - RUN THIS FIRST for easiest experience

2. monitor_resources.ps1
   - PowerShell monitoring script
   - Starts server and collects metrics
   - 2-minute test with 5-second samples
   - Generates CSV results and console report

3. load_test_quotes.ps1
   - Quote streaming simulator
   - Sends HTTP requests to server
   - Tests with multiple currency pairs
   - Run separately or with monitoring

4. run_load_test.ps1
   - Master orchestration script
   - Coordinates monitoring and load test
   - Advanced users only

5. LOAD_TEST_GUIDE.txt
   - Comprehensive testing guide
   - Explains all metrics
   - Troubleshooting section
   - CSV analysis tips

6. TEST_REPORT_TEMPLATE.txt
   - Post-test reporting template
   - Document findings
   - Track improvements over time
   - Retest planning

7. README_MONITORING.txt
   - This file
   - Quick reference and setup

QUICK START (EASIEST WAY)
=========================

Step 1: Open QUICK_START.bat
   - Double-click the file or run: QUICK_START.bat
   - Displays interactive menu

Step 2: Select Option 1 (Full Test)
   - Server starts automatically
   - Resources monitored for 2 minutes
   - Results displayed in console
   - CSV file saved with results

Step 3: Review Results
   - Console shows immediate summary
   - monitoring_results.csv contains detailed data
   - Open in Excel for analysis

SYSTEM REQUIREMENTS
===================

Hardware:
  - Windows 7+ or Windows Server 2008+
  - At least 2GB RAM available
  - Dual-core processor minimum
  - 100MB free disk space

Software:
  - PowerShell 3.0+ (built into Windows 7+)
  - Go runtime (for building/running backend)
  - Backend server.exe must exist
  - Port 8080 available (for server)

Network:
  - No external network required
  - localhost/127.0.0.1 connectivity only

INSTALLATION
=============

No installation required! Just:

1. Ensure backend/cmd/server/server.exe exists
   - If missing, rebuild with: go build -o server.exe ./cmd/server

2. Copy all scripts to Trading-Engine root directory
   - Scripts should be in: D:\Tading engine\Trading-Engine\

3. Run QUICK_START.bat
   - No administrator privileges required

RUNNING THE TEST
=================

METHOD 1: Using QUICK_START Menu (RECOMMENDED)
   1. Double-click QUICK_START.bat
   2. Press 1 for Full Test
   3. Wait 2 minutes
   4. Review console output

METHOD 2: PowerShell Command Line
   powershell -ExecutionPolicy Bypass -File monitor_resources.ps1

METHOD 3: From PowerShell IDE
   1. Open PowerShell ISE
   2. Open monitor_resources.ps1
   3. Click Run Script button
   4. Select Full Test option

UNDERSTANDING THE OUTPUT
=========================

Console Output Format:
   [Sample#] Timestamp | Process: CPU=X% Mem=YMB | System: CPU=Z% Mem=WMB | Threads=N

Example:
   [1] 2026-01-19 14:30:15.123 | Process: CPU=12% Mem=125MB | System: CPU=45% Mem=8450MB | Threads=24

Meaning:
   [1]                              - 1st sample
   2026-01-19 14:30:15.123         - When collected
   CPU=12%                          - Process using 12% CPU
   Mem=125MB                        - Process using 125MB memory
   CPU=45%                          - System at 45% CPU
   Mem=8450MB                       - System memory usage
   Threads=24                       - 24 active threads

FINAL REPORT
============

After test completes, you'll see:

Process Metrics:
  Peak CPU Usage:              X%
  Average CPU Usage:           Y%
  Peak Memory Usage:           Z MB
  Average Memory Usage:        W MB
  Memory Growth:               V MB
  Memory Growth Rate:          U MB/minute

System Metrics:
  Peak System CPU:             X%
  Peak System Memory:          Y MB
  Average System Memory %:     Z%

Stability Assessment:
  ✓ All metrics within ranges  (GOOD)
  ✗ Issues detected (BAD)

Recommendation:
  System is stable - test can continue

CSV RESULTS
===========

File: monitoring_results.csv

Contains:
  - Timestamp of each sample
  - Process CPU usage
  - Process memory (working set and private)
  - System CPU usage
  - System memory usage
  - Thread count

Open in Excel:
  1. Double-click monitoring_results.csv
  2. Data imports automatically
  3. Create charts to visualize trends

MEMORY LEAK DETECTION
=====================

Signs of Memory Leak:
  - Memory continuously increases
  - Memory grows faster than normal
  - Memory never returns to baseline
  - Growth rate > 10 MB/minute

How to Check:
  1. Look at "Memory Growth" in report
  2. Check "Memory Growth Rate" - should be < 10 MB/min
  3. Look at CSV data - plot memory over time
  4. Memory should plateau, not always increase

If Leak Detected:
  1. Document peak memory reached
  2. Note at what point it increased fastest
  3. Check application logs for clues
  4. Review recent code changes
  5. Consider: caches, connections, goroutines

PERFORMANCE THRESHOLDS
======================

ACCEPTABLE:
  - Process CPU:        10-50%
  - Process Memory:     50-500MB
  - Memory Growth:      0-10 MB/minute
  - System CPU:         < 60%
  - System Memory Used: < 80%

WARNING:
  - Process CPU:        50-80%
  - Process Memory:     500-1500MB
  - Memory Growth:      10-50 MB/minute
  - System CPU:         60-85%
  - System Memory Used: 80-90%

CRITICAL (Stop Test):
  - Process CPU:        > 90%
  - Process Memory:     > 2GB
  - System unresponsive

TROUBLESHOOTING
===============

Q: "Cannot reach server at localhost:8080"
A: - Backend server not running
   - Port 8080 in use
   - Firewall blocking port
   - Server crashed on startup

Q: "High memory usage detected"
A: - Normal for high tick volumes
   - Check if memory stabilizes
   - If growing, likely memory leak
   - Review cache settings

Q: "High CPU usage"
A: - Normal during load test
   - If sustained > 80%, investigate
   - May need optimization
   - Try lower request rate

Q: "Script won't run"
A: - Check PowerShell execution policy:
     powershell -ExecutionPolicy Bypass
   - Verify paths are correct
   - Ensure server.exe exists
   - Run as Administrator if needed

Q: "Process keeps crashing"
A: - Check server logs
   - Verify sufficient memory
   - Review backend code for panics
   - Try rebuilding server

ADVANCED USAGE
==============

Custom Duration (Modify monitor_resources.ps1):
   Line: $DurationSeconds = 120     # Change to desired seconds
   Line: $SampleIntervalSeconds = 5 # Change sample frequency

Custom Request Rate (Modify load_test_quotes.ps1):
   Line: [int]$RequestsPerSecond = 10  # Increase for more stress

Run Multiple Tests:
   1. Run first test
   2. Check results
   3. Wait for system to stabilize
   4. Run second test
   5. Compare results

NEXT STEPS
===========

1. Run the test:
   - Double-click QUICK_START.bat
   - Or run monitor_resources.ps1

2. Review the results:
   - Check console output for issues
   - Look at memory growth rate
   - Verify CPU stays below 80%

3. Analyze CSV data:
   - Open monitoring_results.csv
   - Create charts
   - Look for trends

4. Document findings:
   - Use TEST_REPORT_TEMPLATE.txt
   - Save results with timestamp
   - Track improvements over time

5. Optimize if needed:
   - Address memory leaks
   - Improve performance
   - Retest to verify fixes

SUPPORT
=======

For detailed information, see:
  - LOAD_TEST_GUIDE.txt      - Complete guide with examples
  - TEST_REPORT_TEMPLATE.txt - Report format and analysis
  - Backend logs             - Server diagnostics

For code issues:
  - Check backend/cmd/server/main.go
  - Review WebSocket handlers
  - Check database queries
  - Look for goroutine leaks

CONTACT
=======

Questions? Check:
  1. LOAD_TEST_GUIDE.txt (comprehensive guide)
  2. Console output (error messages)
  3. monitoring_results.csv (detailed metrics)
  4. Backend logs (server errors)

================================================================================
Last Updated: 2026-01-19
Version: 1.0
================================================================================
