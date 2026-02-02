# End-to-End Testing Script for Trading Engine
# Tests complete data flow from backend to frontend

Write-Host "==================================================================" -ForegroundColor Cyan
Write-Host "  TRADING ENGINE - END-TO-END TESTING" -ForegroundColor Cyan
Write-Host "==================================================================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:7999"
$wsUrl = "ws://localhost:7999/ws"
$testResults = @()

# Function to test HTTP endpoint
function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Url,
        [string]$Method = "GET",
        [string]$Body = $null
    )

    Write-Host "Testing: $Name..." -NoNewline

    try {
        $params = @{
            Uri = $Url
            Method = $Method
            TimeoutSec = 10
            ErrorAction = "Stop"
        }

        if ($Body) {
            $params.Body = $Body
            $params.ContentType = "application/json"
        }

        $response = Invoke-WebRequest @params

        if ($response.StatusCode -eq 200) {
            Write-Host " ✓ PASS" -ForegroundColor Green
            return @{ Name = $Name; Status = "PASS"; StatusCode = $response.StatusCode; Response = $response.Content }
        } else {
            Write-Host " ✗ FAIL (Status: $($response.StatusCode))" -ForegroundColor Red
            return @{ Name = $Name; Status = "FAIL"; StatusCode = $response.StatusCode; Error = "Unexpected status code" }
        }
    } catch {
        Write-Host " ✗ FAIL" -ForegroundColor Red
        Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Yellow
        return @{ Name = $Name; Status = "FAIL"; Error = $_.Exception.Message }
    }
}

Write-Host "PHASE 1: Backend Server Health Checks" -ForegroundColor Yellow
Write-Host "--------------------------------------" -ForegroundColor Yellow

# Test 1: Health endpoint
$testResults += Test-Endpoint -Name "Health Check" -Url "$baseUrl/health"

# Test 2: Config API
$testResults += Test-Endpoint -Name "Config API" -Url "$baseUrl/api/config"

# Test 3: FIX Status
$testResults += Test-Endpoint -Name "FIX Status" -Url "$baseUrl/admin/fix/status"

# Test 4: Market Data Diagnostics
$testResults += Test-Endpoint -Name "Market Data Diagnostics" -Url "$baseUrl/api/diagnostics/market-data"

# Test 5: Tick Data
$testResults += Test-Endpoint -Name "Tick Data Flow" -Url "$baseUrl/admin/fix/ticks"

# Test 6: Available Symbols
$testResults += Test-Endpoint -Name "Available Symbols" -Url "$baseUrl/api/symbols/available"

Write-Host ""
Write-Host "PHASE 2: Data Flow Analysis" -ForegroundColor Yellow
Write-Host "----------------------------" -ForegroundColor Yellow

# Analyze tick data
$ticksTest = $testResults | Where-Object { $_.Name -eq "Tick Data Flow" }
if ($ticksTest -and $ticksTest.Status -eq "PASS") {
    $tickData = $ticksTest.Response | ConvertFrom-Json
    Write-Host "  Total Ticks Received: $($tickData.totalTickCount)" -ForegroundColor Cyan
    Write-Host "  Active Symbols: $($tickData.symbolCount)" -ForegroundColor Cyan

    # Show sample of latest ticks
    if ($tickData.latestTicks) {
        $symbols = ($tickData.latestTicks | Get-Member -MemberType NoteProperty).Name | Select-Object -First 5
        Write-Host "  Sample Ticks:" -ForegroundColor Cyan
        foreach ($sym in $symbols) {
            $tick = $tickData.latestTicks.$sym
            Write-Host "    - $sym : Bid=$($tick.bid), Ask=$($tick.ask), LP=$($tick.lp)" -ForegroundColor Gray
        }
    }
}

# Analyze FIX status
$fixTest = $testResults | Where-Object { $_.Name -eq "FIX Status" }
if ($fixTest -and $fixTest.Status -eq "PASS") {
    $fixData = $fixTest.Response | ConvertFrom-Json
    Write-Host ""
    Write-Host "  FIX Session Status:" -ForegroundColor Cyan
    foreach ($session in $fixData.sessions.PSObject.Properties) {
        $status = $session.Value
        $color = if ($status -eq "LOGGED_IN") { "Green" } elseif ($status -eq "CONNECTED") { "Yellow" } else { "Red" }
        Write-Host "    - $($session.Name): $status" -ForegroundColor $color
    }
}

Write-Host ""
Write-Host "PHASE 3: Configuration Verification" -ForegroundColor Yellow
Write-Host "------------------------------------" -ForegroundColor Yellow

# Check frontend environment
$frontendEnvPath = "C:\Users\s s laptop bazar\Trading-Engine2\clients\desktop\.env"
if (Test-Path $frontendEnvPath) {
    Write-Host "  Frontend .env: EXISTS ✓" -ForegroundColor Green
    $envContent = Get-Content $frontendEnvPath
    $apiUrl = $envContent | Where-Object { $_ -match "VITE_API_URL" }
    if ($apiUrl) {
        Write-Host "    $apiUrl" -ForegroundColor Gray
    }
} else {
    Write-Host "  Frontend .env: MISSING ✗" -ForegroundColor Red
    Write-Host "    Note: Will use default http://localhost:7999" -ForegroundColor Yellow
}

# Check backend environment
$backendEnvPath = "C:\Users\s s laptop bazar\Trading-Engine2\backend\.env"
if (Test-Path $backendEnvPath) {
    Write-Host "  Backend .env: EXISTS ✓" -ForegroundColor Green
} else {
    Write-Host "  Backend .env: MISSING (using defaults)" -ForegroundColor Yellow
}

# Check FIX sequence numbers
$yofx1SeqPath = "C:\Users\s s laptop bazar\Trading-Engine2\backend\fixstore\YOFX1.seqnums"
$yofx2SeqPath = "C:\Users\s s laptop bazar\Trading-Engine2\backend\fixstore\YOFX2.seqnums"

if (Test-Path $yofx1SeqPath) {
    $seq = Get-Content $yofx1SeqPath
    Write-Host "  YOFX1 Sequence: $seq" -ForegroundColor Gray
}
if (Test-Path $yofx2SeqPath) {
    $seq = Get-Content $yofx2SeqPath
    Write-Host "  YOFX2 Sequence: $seq" -ForegroundColor Gray
}

Write-Host ""
Write-Host "PHASE 4: Port and Network Checks" -ForegroundColor Yellow
Write-Host "---------------------------------" -ForegroundColor Yellow

# Check if port 7999 is listening
$port7999 = Get-NetTCPConnection -LocalPort 7999 -State Listen -ErrorAction SilentlyContinue
if ($port7999) {
    Write-Host "  Port 7999: LISTENING ✓" -ForegroundColor Green
    Write-Host "    Process ID: $($port7999[0].OwningProcess)" -ForegroundColor Gray
} else {
    Write-Host "  Port 7999: NOT LISTENING ✗" -ForegroundColor Red
}

# Check FIX server connectivity
Write-Host ""
Write-Host "  Testing FIX Server Connectivity..." -ForegroundColor Cyan
$fixServer = "23.106.238.138"
$fixPort = 12336

try {
    $tcpTest = Test-NetConnection -ComputerName $fixServer -Port $fixPort -WarningAction SilentlyContinue
    if ($tcpTest.TcpTestSucceeded) {
        Write-Host "    FIX Server ($fixServer`:$fixPort): REACHABLE ✓" -ForegroundColor Green
    } else {
        Write-Host "    FIX Server ($fixServer`:$fixPort): UNREACHABLE ✗" -ForegroundColor Red
    }
} catch {
    Write-Host "    FIX Server Test: FAILED ($($_.Exception.Message))" -ForegroundColor Red
}

Write-Host ""
Write-Host "==================================================================" -ForegroundColor Cyan
Write-Host "  TEST SUMMARY" -ForegroundColor Cyan
Write-Host "==================================================================" -ForegroundColor Cyan

$passCount = ($testResults | Where-Object { $_.Status -eq "PASS" }).Count
$failCount = ($testResults | Where-Object { $_.Status -eq "FAIL" }).Count
$totalTests = $testResults.Count

Write-Host ""
Write-Host "  Total Tests: $totalTests" -ForegroundColor White
Write-Host "  Passed: $passCount" -ForegroundColor Green
Write-Host "  Failed: $failCount" -ForegroundColor Red
Write-Host ""

if ($failCount -gt 0) {
    Write-Host "FAILED TESTS:" -ForegroundColor Red
    $testResults | Where-Object { $_.Status -eq "FAIL" } | ForEach-Object {
        Write-Host "  - $($_.Name): $($_.Error)" -ForegroundColor Yellow
    }
    Write-Host ""
}

Write-Host "KEY FINDINGS:" -ForegroundColor Cyan
Write-Host "-------------" -ForegroundColor Cyan

# Analyze configuration
$configTest = $testResults | Where-Object { $_.Name -eq "Config API" }
if ($configTest -and $configTest.Status -eq "PASS") {
    $config = $configTest.Response | ConvertFrom-Json
    Write-Host "  1. Broker: $($config.brokerDisplayName)" -ForegroundColor White
    Write-Host "  2. Execution Mode: $($config.executionMode)" -ForegroundColor White
    Write-Host "  3. Price Feed: $($config.priceFeedName) ($($config.priceFeedLP))" -ForegroundColor White
}

# Analyze market data
$mdTest = $testResults | Where-Object { $_.Name -eq "Market Data Diagnostics" }
if ($mdTest -and $mdTest.Status -eq "PASS") {
    $md = $mdTest.Response | ConvertFrom-Json
    Write-Host "  4. Market Data Status: $($md.status.ToUpper())" -ForegroundColor White
    Write-Host "  5. Active Streams: $($md.activeStreams)" -ForegroundColor White
    Write-Host "  6. Total Ticks: $($md.totalTicksReceived)" -ForegroundColor White
}

Write-Host ""
Write-Host "RECOMMENDATIONS:" -ForegroundColor Yellow
Write-Host "----------------" -ForegroundColor Yellow

# Check if FIX is connected
$fixConnected = $false
if ($fixTest -and $fixTest.Status -eq "PASS") {
    $fixData = $fixTest.Response | ConvertFrom-Json
    $fixConnected = ($fixData.sessions.YOFX1 -eq "LOGGED_IN" -or $fixData.sessions.YOFX2 -eq "LOGGED_IN")
}

if (-not $fixConnected) {
    Write-Host "  ⚠ FIX Sessions are DISCONNECTED" -ForegroundColor Yellow
    Write-Host "    - Market data is being SIMULATED (LP=SIM)" -ForegroundColor Gray
    Write-Host "    - To connect FIX:" -ForegroundColor Gray
    Write-Host "      POST http://localhost:7999/admin/fix/connect" -ForegroundColor Gray
    Write-Host "      Body: {`"sessionId`": `"YOFX1`"}" -ForegroundColor Gray
} else {
    Write-Host "  ✓ FIX Sessions connected - receiving live data" -ForegroundColor Green
}

Write-Host ""
Write-Host "  WebSocket Endpoint: $wsUrl" -ForegroundColor Cyan
Write-Host "  Frontend should connect to this endpoint for real-time ticks" -ForegroundColor Gray
Write-Host ""

# Export results to JSON
$resultsPath = "C:\Users\s s laptop bazar\Trading-Engine2\backend\e2e_test_results.json"
$testResults | ConvertTo-Json -Depth 5 | Out-File -FilePath $resultsPath -Encoding UTF8
Write-Host "  Test results exported to: $resultsPath" -ForegroundColor Cyan
Write-Host ""
Write-Host "==================================================================" -ForegroundColor Cyan
