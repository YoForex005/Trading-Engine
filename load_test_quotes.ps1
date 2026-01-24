# Quote Streaming Load Test Script
# Sends quote requests to the backend server for 2 minutes

param(
    [string]$ServerURL = "http://localhost:8080",
    [int]$DurationSeconds = 120,
    [int]$RequestsPerSecond = 10
)

function Write-Info { Write-Host "[INFO] $args" -ForegroundColor Cyan }
function Write-Success { Write-Host "[SUCCESS] $args" -ForegroundColor Green }
function Write-Error2 { Write-Host "[ERROR] $args" -ForegroundColor Red }

# Quote symbols to test
$symbols = @("EURUSD", "GBPUSD", "AUDUSD", "NZDUSD", "USDCAD", "USDCHF", "USDJPY", "EURGBP")

Write-Info "Quote Streaming Load Test"
Write-Info "Target: $ServerURL"
Write-Info "Duration: $DurationSeconds seconds"
Write-Info "Rate: $RequestsPerSecond requests/second"
Write-Info "Symbols: $($symbols -join ', ')"
Write-Info ""

# Verify server is reachable
Write-Info "Testing server connectivity..."
try {
    $response = Invoke-WebRequest -Uri "$ServerURL/health" -ErrorAction Stop -TimeoutSec 5
    Write-Success "Server is reachable"
}
catch {
    Write-Error2 "Cannot reach server at $ServerURL. Is it running?"
    exit 1
}

Write-Info "Starting load test..."
$stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
$successCount = 0
$errorCount = 0
$totalRequests = 0

while ($stopwatch.Elapsed.TotalSeconds -lt $DurationSeconds) {
    for ($i = 0; $i -lt $RequestsPerSecond; $i++) {
        $symbol = $symbols | Get-Random

        try {
            $uri = "$ServerURL/api/quote/$symbol"
            $response = Invoke-WebRequest -Uri $uri -ErrorAction Stop -TimeoutSec 5 -UseBasicParsing
            $successCount++
            $totalRequests++

            if (($totalRequests % 50) -eq 0) {
                Write-Info "Requests sent: $totalRequests (Success: $successCount, Errors: $errorCount)"
            }
        }
        catch {
            $errorCount++
            $totalRequests++
            Write-Host "Request failed: $_" -ForegroundColor Yellow
        }
    }

    # Maintain timing
    $elapsed = $stopwatch.Elapsed.TotalSeconds
    $expectedRequests = $elapsed * $RequestsPerSecond
    if ($totalRequests -lt $expectedRequests) {
        $sleepTime = [math]::Max(10, 1000 / $RequestsPerSecond - (([System.Diagnostics.Stopwatch]::StartNew().Elapsed).TotalMilliseconds))
        Start-Sleep -Milliseconds $sleepTime
    }
}

$stopwatch.Stop()

Write-Info ""
Write-Success "Load test completed"
Write-Host ""
Write-Host "====== LOAD TEST RESULTS =======" -ForegroundColor Cyan
Write-Host "Duration:              $([math]::Round($stopwatch.Elapsed.TotalSeconds, 1)) seconds"
Write-Host "Total Requests:        $totalRequests"
Write-Host "Successful:            $successCount"
Write-Host "Failed:                $errorCount"
Write-Host "Success Rate:          $([math]::Round(($successCount / $totalRequests) * 100, 2))%"
Write-Host "Requests/Second:       $([math]::Round($totalRequests / $stopwatch.Elapsed.TotalSeconds, 2))"
Write-Host "=====================================" -ForegroundColor Cyan
