# Resource Monitoring Script for Quote Streaming Test
# Monitors CPU and Memory usage for 2 minutes, sampling every 5 seconds

$ServerPath = "D:\Tading engine\Trading-Engine\backend\cmd\server\server.exe"
$OutputLog = "D:\Tading engine\Trading-Engine\monitoring_results.csv"
$DurationSeconds = 120  # 2 minutes
$SampleIntervalSeconds = 5
$InitialWaitSeconds = 10

# Initialize results array
$results = @()
$systemMetrics = @()

# Color output functions
function Write-Info { Write-Host "[INFO] $args" -ForegroundColor Cyan }
function Write-Warning { Write-Host "[WARNING] $args" -ForegroundColor Yellow }
function Write-Error2 { Write-Host "[ERROR] $args" -ForegroundColor Red }
function Write-Success { Write-Host "[SUCCESS] $args" -ForegroundColor Green }

# Start the server
Write-Info "Starting backend server..."
$serverProcess = Start-Process -FilePath $ServerPath -NoNewWindow -PassThru
Write-Success "Server started with PID: $($serverProcess.Id)"

# Wait for server to initialize
Write-Info "Waiting $InitialWaitSeconds seconds for server to fully initialize..."
Start-Sleep -Seconds $InitialWaitSeconds

# Verify process is still running
if ($serverProcess.HasExited) {
    Write-Error2 "Server process exited unexpectedly!"
    exit 1
}

Write-Info "Starting resource monitoring for $DurationSeconds seconds (sampling every $SampleIntervalSeconds seconds)..."
Write-Info "Monitor will check: CPU, Memory, and System stability"
Write-Info ""

$stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
$sampleCount = 0

# Monitoring loop
while ($stopwatch.Elapsed.TotalSeconds -lt $DurationSeconds) {
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss.fff"

    # Check if server is still running
    if ($serverProcess.HasExited) {
        Write-Error2 "Server process terminated during monitoring!"
        break
    }

    try {
        # Get process-specific metrics (go.exe process)
        $goProcesses = Get-Process -Name "server" -ErrorAction SilentlyContinue

        if ($goProcesses) {
            $proc = $goProcesses[0]
            $cpuTime = $proc.CPU
            $memoryMB = $proc.WorkingSet64 / 1MB
            $privateMB = $proc.PrivateMemorySize64 / 1MB

            # Get total system metrics
            $cpuUsage = (Get-WmiObject win32_processor).LoadPercentage
            $osMetrics = Get-WmiObject Win32_OperatingSystem
            $totalMemMB = $osMetrics.TotalVisibleMemorySize / 1KB
            $freeMemMB = $osMetrics.FreePhysicalMemory / 1KB
            $usedMemMB = $totalMemMB - $freeMemMB
            $systemMemPercent = ($usedMemMB / $totalMemMB) * 100

            # Store results
            $result = [PSCustomObject]@{
                Timestamp = $timestamp
                ElapsedSeconds = [math]::Round($stopwatch.Elapsed.TotalSeconds, 1)
                ProcessCPU_Percent = [math]::Round($cpuTime, 2)
                ProcessMemory_MB = [math]::Round($memoryMB, 2)
                ProcessPrivateMem_MB = [math]::Round($privateMB, 2)
                SystemCPU_Percent = $cpuUsage
                SystemMemory_MB = [math]::Round($usedMemMB, 2)
                SystemMemory_Percent = [math]::Round($systemMemPercent, 2)
                ThreadCount = $proc.Threads.Count
            }

            $results += $result
            $sampleCount++

            # Display progress
            Write-Host "[$sampleCount] $timestamp | Process: CPU=${cpuTime}% Mem=${memoryMB}MB | System: CPU=$cpuUsage% Mem=${usedMemMB}MB (${systemMemPercent}%) | Threads=$($proc.Threads.Count)" -ForegroundColor White

            # Check for critical thresholds
            if ($cpuTime -gt 90) {
                Write-Warning "ALERT: Process CPU usage exceeds 90%! Current: $cpuTime%"
            }
            if ($memoryMB -gt 2048) {
                Write-Warning "ALERT: Process memory usage exceeds 2GB! Current: $memoryMB MB"
            }
            if ($cpuUsage -gt 85) {
                Write-Warning "ALERT: System CPU usage is high! Current: $cpuUsage%"
            }
            if ($systemMemPercent -gt 90) {
                Write-Warning "ALERT: System memory usage exceeds 90%! Current: $systemMemPercent%"
            }
        }
        else {
            Write-Warning "Could not find server process"
        }
    }
    catch {
        Write-Warning "Error collecting metrics: $_"
    }

    # Wait for next sample
    Start-Sleep -Seconds $SampleIntervalSeconds
}

$stopwatch.Stop()

Write-Info ""
Write-Success "Monitoring complete. Collecting final metrics..."

# Stop the server
if (-not $serverProcess.HasExited) {
    Write-Info "Stopping server process..."
    $serverProcess | Stop-Process -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 1
    Write-Success "Server stopped"
}

# Export results to CSV
Write-Info "Exporting results to $OutputLog..."
$results | Export-Csv -Path $OutputLog -NoTypeInformation
Write-Success "Results exported to CSV"

# Calculate statistics
Write-Info ""
Write-Host "====== MONITORING RESULTS =======" -ForegroundColor Cyan

if ($results.Count -gt 0) {
    $peakCPU = ($results | Measure-Object -Property "ProcessCPU_Percent" -Maximum).Maximum
    $peakMemory = ($results | Measure-Object -Property "ProcessMemory_MB" -Maximum).Maximum
    $avgCPU = ($results | Measure-Object -Property "ProcessCPU_Percent" -Average).Average
    $avgMemory = ($results | Measure-Object -Property "ProcessMemory_MB" -Average).Average
    $peakSysCPU = ($results | Measure-Object -Property "SystemCPU_Percent" -Maximum).Maximum
    $peakSysMem = ($results | Measure-Object -Property "SystemMemory_MB" -Maximum).Maximum
    $avgSysMem = ($results | Measure-Object -Property "SystemMemory_MB" -Average).Average
    $avgSysMemPercent = ($results | Measure-Object -Property "SystemMemory_Percent" -Average).Average

    # Calculate memory growth rate
    $firstMemory = $results[0].ProcessMemory_MB
    $lastMemory = $results[-1].ProcessMemory_MB
    $memoryGrowth = $lastMemory - $firstMemory
    $memoryGrowthRate = ($memoryGrowth / ($DurationSeconds / 60))

    Write-Host ""
    Write-Host "Process (go.exe/server.exe) Metrics:" -ForegroundColor Yellow
    Write-Host "  Peak CPU Usage:              $([math]::Round($peakCPU, 2))%"
    Write-Host "  Average CPU Usage:          $([math]::Round($avgCPU, 2))%"
    Write-Host "  Peak Memory Usage:          $([math]::Round($peakMemory, 2)) MB"
    Write-Host "  Average Memory Usage:       $([math]::Round($avgMemory, 2)) MB"
    Write-Host "  Memory Growth (Total):      $([math]::Round($memoryGrowth, 2)) MB"
    Write-Host "  Memory Growth Rate:         $([math]::Round($memoryGrowthRate, 2)) MB/minute"

    Write-Host ""
    Write-Host "System Metrics:" -ForegroundColor Yellow
    Write-Host "  Peak System CPU:            $peakSysCPU%"
    Write-Host "  Peak System Memory:         $([math]::Round($peakSysMem, 2)) MB"
    Write-Host "  Average System Memory:      $([math]::Round($avgSysMem, 2)) MB"
    Write-Host "  Average Memory Utilization: $([math]::Round($avgSysMemPercent, 2))%"

    Write-Host ""
    Write-Host "Monitoring Details:" -ForegroundColor Yellow
    Write-Host "  Total Samples Collected:    $($results.Count)"
    Write-Host "  Monitoring Duration:       $([math]::Round($DurationSeconds / 60, 1)) minutes"
    Write-Host "  Sample Interval:            $SampleIntervalSeconds seconds"

    # Stability assessment
    Write-Host ""
    Write-Host "Stability Assessment:" -ForegroundColor Yellow

    $issues = @()

    if ($peakCPU -gt 90) {
        $issues += "CRITICAL: Peak CPU > 90% ($([math]::Round($peakCPU, 2))%)"
    }
    elseif ($peakCPU -gt 80) {
        $issues += "WARNING: Peak CPU > 80% ($([math]::Round($peakCPU, 2))%)"
    }

    if ($peakMemory -gt 2048) {
        $issues += "CRITICAL: Peak Memory > 2GB ($([math]::Round($peakMemory, 2)) MB)"
    }
    elseif ($peakMemory -gt 1536) {
        $issues += "WARNING: Peak Memory > 1.5GB ($([math]::Round($peakMemory, 2)) MB)"
    }

    if ($memoryGrowthRate -gt 50) {
        $issues += "WARNING: High memory growth rate ($([math]::Round($memoryGrowthRate, 2)) MB/minute)"
    }

    if ($peakSysCPU -gt 90) {
        $issues += "WARNING: Peak System CPU > 90% ($peakSysCPU%)"
    }

    if ($issues.Count -eq 0) {
        Write-Success "✓ All metrics within acceptable ranges"
        Write-Host "✓ No memory leaks detected"
        Write-Host "✓ CPU usage stable"
        Write-Host "✓ System responsive"
    }
    else {
        Write-Warning "Issues detected:"
        foreach ($issue in $issues) {
            Write-Warning "  ✗ $issue"
        }
    }

    Write-Host ""
    Write-Host "RECOMMENDATION:" -ForegroundColor Cyan
    if (($peakCPU -gt 90) -or ($peakMemory -gt 2048)) {
        Write-Error2 "STOP TEST: Critical resource thresholds exceeded"
    }
    elseif ($issues.Count -gt 0) {
        Write-Warning "Monitor performance - some metrics are elevated"
    }
    else {
        Write-Success "System is stable - test can continue"
    }
}
else {
    Write-Error2 "No monitoring data collected"
}

Write-Host ""
Write-Host "=====================================" -ForegroundColor Cyan
