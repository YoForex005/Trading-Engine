# Master Load Test Orchestration Script
# Runs resource monitoring and quote streaming simultaneously

$ScriptDir = "D:\Tading engine\Trading-Engine"
$MonitorScript = Join-Path $ScriptDir "monitor_resources.ps1"
$LoadTestScript = Join-Path $ScriptDir "load_test_quotes.ps1"

function Write-Info { Write-Host "[INFO] $args" -ForegroundColor Cyan }
function Write-Success { Write-Host "[SUCCESS] $args" -ForegroundColor Green }
function Write-Error2 { Write-Host "[ERROR] $args" -ForegroundColor Red }

Write-Info "======================================"
Write-Info "QUOTE STREAMING LOAD TEST"
Write-Info "======================================"
Write-Info ""
Write-Info "This script will:"
Write-Info "1. Start the backend server"
Write-Info "2. Monitor system resources for 2 minutes"
Write-Info "3. Send quote requests to test the system"
Write-Info "4. Generate resource usage report"
Write-Info ""
Write-Info "Monitoring will track:"
Write-Info "  - CPU usage (process & system)"
Write-Info "  - Memory usage (process & system)"
Write-Info "  - Memory leak detection"
Write-Info "  - Thread count"
Write-Info "  - System stability"
Write-Info ""

# Verify scripts exist
if (-not (Test-Path $MonitorScript)) {
    Write-Error2 "Monitoring script not found: $MonitorScript"
    exit 1
}

Write-Info "Starting orchestration..."
Write-Info ""

# Run monitoring script (which starts the server internally)
Write-Info "Executing resource monitoring and load test..."
& powershell -ExecutionPolicy Bypass -File $MonitorScript

Write-Info ""
Write-Info "======================================"
Write-Success "Load test and monitoring complete"
Write-Info "======================================"
Write-Info ""
Write-Info "Results saved to:"
Write-Info "  CSV: $(Join-Path $ScriptDir 'monitoring_results.csv')"
Write-Info ""
Write-Info "To view results:"
Write-Info "  Import-Csv '$(Join-Path $ScriptDir 'monitoring_results.csv')' | Format-Table"
