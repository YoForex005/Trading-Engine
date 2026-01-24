# Market Data Pipeline Verification Script for Windows
# Purpose: Verify pipeline health and data flow
# Usage: ./verify_pipeline.ps1

param(
    [string]$BackendURL = "http://localhost:8080",
    [string]$RedisPort = "6379",
    [string]$Duration = "30"  # seconds
)

$ErrorActionPreference = "Continue"

function Write-Status {
    param(
        [string]$Status,
        [string]$Message,
        [switch]$Success,
        [switch]$Warning,
        [switch]$Error
    )

    $Color = "White"
    if ($Success) { $Color = "Green" }
    elseif ($Warning) { $Color = "Yellow" }
    elseif ($Error) { $Color = "Red" }

    Write-Host "[$Status] $Message" -ForegroundColor $Color
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Market Data Pipeline Verification" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Backend URL: $BackendURL"
Write-Host "Redis Port: $RedisPort"
Write-Host "Test Duration: ${Duration}s"
Write-Host ""

# 1. Check Backend Service
Write-Host "1. Backend Service Status:" -ForegroundColor Cyan

$BackendRunning = Get-Process | Where-Object { $_.ProcessName -like "*go*" -and $_.CommandLine -like "*server*" }

if ($BackendRunning) {
    Write-Status "✓" "Backend running (PID: $($BackendRunning.Id))" -Success
    $BackendOK = $true
} else {
    Write-Status "✗" "Backend NOT running" -Error
    Write-Host "  Start with: cd backend\cmd\server && go run main.go" -ForegroundColor Yellow
    $BackendOK = $false
}

# 2. Check Redis
Write-Host ""
Write-Host "2. Redis Status:" -ForegroundColor Cyan

try {
    $RedisCheck = Invoke-WebRequest -Uri "http://localhost:$RedisPort/ping" -ErrorAction SilentlyContinue

    # Try direct connection check via TCP
    $TcpClient = New-Object System.Net.Sockets.TcpClient
    $Result = $TcpClient.BeginConnect("127.0.0.1", $RedisPort, $null, $null)
    $Result.AsyncWaitHandle.WaitOne(3000) | Out-Null

    if ($TcpClient.Connected) {
        Write-Status "✓" "Redis online (Port $RedisPort)" -Success
        $RedisOK = $true
        $TcpClient.Close()
    } else {
        Write-Status "✗" "Redis NOT responding" -Error
        $RedisOK = $false
    }
} catch {
    Write-Status "✗" "Redis NOT running" -Error
    Write-Host "  Start with: redis-server" -ForegroundColor Yellow
    $RedisOK = $false
}

if ($BackendOK) {
    # 3. Check FIX Connection
    Write-Host ""
    Write-Host "3. FIX Gateway Status:" -ForegroundColor Cyan

    try {
        $FixStatus = Invoke-RestMethod -Uri "$BackendURL/api/admin/fix-status" -Method Get -TimeoutSec 5 -ErrorAction Stop
        $YOFX2Status = $FixStatus.YOFX2

        if ($YOFX2Status -eq "connected") {
            Write-Status "✓" "FIX connected (YOFX2)" -Success
            $FixOK = $true
        } elseif ($YOFX2Status -eq "connecting") {
            Write-Status "⏳" "FIX connecting (YOFX2)" -Warning
            $FixOK = $true
        } else {
            Write-Status "⚠" "FIX status: $YOFX2Status" -Warning
            $FixOK = $true
        }
    } catch {
        Write-Status "✗" "Cannot reach FIX status endpoint" -Error
        $FixOK = $false
    }

    # 4. Check Pipeline Stats
    Write-Host ""
    Write-Host "4. Pipeline Statistics:" -ForegroundColor Cyan

    try {
        $PipelineStats = Invoke-RestMethod -Uri "$BackendURL/api/admin/pipeline-stats" -Method Get -TimeoutSec 5 -ErrorAction Stop

        $TicksReceived = $PipelineStats.data.ticks_received
        $TicksProcessed = $PipelineStats.data.ticks_processed
        $AvgLatency = [Math]::Round($PipelineStats.data.avg_latency_ms, 2)
        $TicksDropped = $PipelineStats.data.ticks_dropped
        $ClientsConnected = $PipelineStats.data.clients_connected

        Write-Host "   Ticks received: $TicksReceived"
        Write-Host "   Ticks processed: $TicksProcessed"
        Write-Host "   Avg latency: ${AvgLatency}ms"
        Write-Host "   Dropped: $TicksDropped"
        Write-Host "   Connected clients: $ClientsConnected"

        # Check latency
        if ($AvgLatency -lt 10) {
            Write-Status "✓" "Latency acceptable (< 10ms)" -Success
            $LatencyOK = $true
        } elseif ($AvgLatency -lt 20) {
            Write-Status "⚠" "Latency elevated (${AvgLatency}ms)" -Warning
            $LatencyOK = $true
        } else {
            Write-Status "✗" "Latency HIGH (${AvgLatency}ms > 20ms)" -Error
            $LatencyOK = $false
        }

        # Check dropped
        if ($TicksDropped -eq 0) {
            Write-Status "✓" "No dropped ticks" -Success
            $DroppedOK = $true
        } else {
            Write-Status "⚠" "Dropped ticks: $TicksDropped" -Warning
            $DroppedOK = $false
        }
    } catch {
        Write-Status "✗" "Cannot reach pipeline stats: $_" -Error
        $LatencyOK = $false
        $DroppedOK = $false
    }

    # 5. Check Redis Data Volume
    Write-Host ""
    Write-Host "5. Market Data Storage:" -ForegroundColor Cyan

    # Parse FIX messages if available
    $FixMsgPath = ".\backend\fixstore\YOFX2.msgs"
    if (Test-Path $FixMsgPath) {
        $FIXLineCount = (Get-Content $FixMsgPath | Measure-Object -Line).Lines
        Write-Host "   FIX messages logged: $FIXLineCount"

        # Count market data messages (MsgType D)
        $MarketDataCount = (Get-Content $FixMsgPath | Select-String "35=D" | Measure-Object).Count
        if ($MarketDataCount -gt 0) {
            Write-Status "✓" "Market data quotes: $MarketDataCount" -Success
            $StorageOK = $true
        } else {
            Write-Status "⚠" "No market data quotes in FIX log" -Warning
            $StorageOK = $false
        }
    } else {
        Write-Status "⚠" "FIX log not found: $FixMsgPath" -Warning
        $StorageOK = $true
    }

    # 6. WebSocket Status
    Write-Host ""
    Write-Host "6. WebSocket Status:" -ForegroundColor Cyan

    if ($ClientsConnected -gt 0) {
        Write-Status "✓" "WebSocket clients connected: $ClientsConnected" -Success
        $WebSocketOK = $true
    } else {
        Write-Status "⏳" "No WebSocket clients (but service running)" -Warning
        $WebSocketOK = $true
    }
} else {
    Write-Host "Skipping remaining checks (backend not running)" -ForegroundColor Yellow
    $FixOK = $false
    $LatencyOK = $false
    $DroppedOK = $false
    $StorageOK = $false
    $WebSocketOK = $false
}

# Summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Health Check Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

if ($BackendOK) {
    Write-Status "✓" "Backend Service" -Success
} else {
    Write-Status "✗" "Backend Service" -Error
}

if ($RedisOK) {
    Write-Status "✓" "Redis" -Success
} else {
    Write-Status "✗" "Redis" -Error
}

if ($FixOK) {
    Write-Status "✓" "FIX Gateway" -Success
} else {
    Write-Status "⚠" "FIX Gateway" -Warning
}

if ($LatencyOK) {
    Write-Status "✓" "Pipeline Latency" -Success
} else {
    Write-Status "✗" "Pipeline Latency" -Error
}

if ($DroppedOK) {
    Write-Status "✓" "No Dropped Ticks" -Success
} else {
    Write-Status "⚠" "Dropped Ticks" -Warning
}

if ($StorageOK) {
    Write-Status "✓" "Data Storage" -Success
} else {
    Write-Status "⚠" "Data Storage" -Warning
}

if ($WebSocketOK) {
    Write-Status "✓" "WebSocket" -Success
} else {
    Write-Status "⚠" "WebSocket" -Warning
}

Write-Host ""

$AllOK = $BackendOK -and $RedisOK -and $LatencyOK -and $DroppedOK
if ($AllOK) {
    Write-Host "OVERALL STATUS: HEALTHY" -ForegroundColor Green
    exit 0
} else {
    Write-Host "OVERALL STATUS: ISSUES DETECTED" -ForegroundColor Red
    exit 1
}
