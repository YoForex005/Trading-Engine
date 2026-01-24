# WebSocket Quote Analysis Tool
# Analyzes ticks from ws://localhost:7999/ws to determine if they are real or simulated

$WS_URL = "ws://localhost:7999/ws"
$CAPTURE_COUNT = 20
$TIMEOUT = 30

Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "🔌 WebSocket Quote Analysis Tool" -ForegroundColor Cyan
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "Connecting to: $WS_URL"
Write-Host "Target: Capture $CAPTURE_COUNT ticks per symbol"
Write-Host "Timeout: ${TIMEOUT}s"
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan

# Data collection
$capturedTicks = @{}
$totalTickCount = 0
$lpSources = @{}
$symbols = @()
$evidence = @()
$startTime = Get-Date

try {
    # Create WebSocket client
    $ws = New-Object System.Net.WebSockets.ClientWebSocket
    $ct = New-Object System.Threading.CancellationToken

    # Connect
    $uri = [System.Uri]::new($WS_URL)
    $connectTask = $ws.ConnectAsync($uri, $ct)
    $connectTask.Wait()

    Write-Host "✅ Connected to WebSocket" -ForegroundColor Green
    Write-Host "📊 Capturing ticks...`n" -ForegroundColor Yellow

    $buffer = New-Object byte[] 8192
    $segment = New-Object System.ArraySegment[byte] -ArgumentList @(,$buffer)

    $captureStart = Get-Date

    while ($ws.State -eq 'Open') {
        # Check timeout
        $elapsed = (Get-Date) - $captureStart
        if ($elapsed.TotalSeconds -gt $TIMEOUT) {
            Write-Host "`n⏰ Analysis timeout reached" -ForegroundColor Yellow
            break
        }

        # Receive message
        try {
            $receiveTask = $ws.ReceiveAsync($segment, $ct)

            # Wait for message with timeout
            $completed = $receiveTask.Wait(5000)

            if (-not $completed) {
                continue
            }

            $result = $receiveTask.Result

            if ($result.MessageType -eq 'Text') {
                $message = [System.Text.Encoding]::UTF8.GetString($buffer, 0, $result.Count)
                $tick = $message | ConvertFrom-Json

                if ($tick.type -eq 'tick') {
                    $totalTickCount++
                    $symbol = $tick.symbol

                    # Track symbols
                    if ($symbols -notcontains $symbol) {
                        $symbols += $symbol
                    }

                    # Track LP sources
                    $lp = if ($tick.lp) { $tick.lp } else { "UNKNOWN" }
                    if ($lpSources.ContainsKey($lp)) {
                        $lpSources[$lp]++
                    } else {
                        $lpSources[$lp] = 1
                    }

                    # Store tick
                    if (-not $capturedTicks.ContainsKey($symbol)) {
                        $capturedTicks[$symbol] = @()
                    }

                    if ($capturedTicks[$symbol].Count -lt $CAPTURE_COUNT) {
                        $capturedTicks[$symbol] += $tick
                    }

                    # Log first few ticks
                    if ($totalTickCount -le 5) {
                        $timestamp = ([DateTimeOffset]::FromUnixTimeSeconds($tick.timestamp)).DateTime.ToString("yyyy-MM-ddTHH:mm:ss")
                        Write-Host "[$totalTickCount] $symbol | Bid: $($tick.bid.ToString('F5')) | Ask: $($tick.ask.ToString('F5')) | LP: $lp | Time: $timestamp"
                    }

                    # Check if we have enough data
                    $symbolsWithEnough = 0
                    foreach ($key in $capturedTicks.Keys) {
                        if ($capturedTicks[$key].Count -ge $CAPTURE_COUNT) {
                            $symbolsWithEnough++
                        }
                    }

                    if ($symbolsWithEnough -ge [Math]::Min(3, $symbols.Count) -and $totalTickCount -ge $CAPTURE_COUNT) {
                        break
                    }
                }
            }
        }
        catch {
            # Timeout or error, continue
        }
    }

    # Close connection
    if ($ws.State -eq 'Open') {
        $closeTask = $ws.CloseAsync([System.Net.WebSockets.WebSocketCloseStatus]::NormalClosure, "Analysis complete", $ct)
        $closeTask.Wait()
    }

    Write-Host "`n🔌 WebSocket disconnected" -ForegroundColor Yellow

    # Check if we captured data
    if ($totalTickCount -eq 0) {
        Write-Host "`n❌ No ticks captured. Server may not be sending data." -ForegroundColor Red
        exit 1
    }

    # ANALYSIS
    Write-Host "`n📊 LP SOURCE ANALYSIS:" -ForegroundColor Cyan
    Write-Host ("═" * 60) -ForegroundColor Cyan

    foreach ($lp in $lpSources.Keys) {
        $count = $lpSources[$lp]
        $percentage = ($count / $totalTickCount) * 100
        Write-Host "  $lp`: $count ticks ($($percentage.ToString('F1'))%)"

        if ($lp -eq 'SIMULATED') {
            $evidence += "❌ LP field shows `"SIMULATED`" - clear simulation marker"
        } elseif ($lp -eq 'YOFX') {
            $evidence += "✅ LP field shows `"YOFX`" - indicates real FIX gateway data"
        }
    }

    # Price Pattern Analysis
    Write-Host "`n📈 PRICE PATTERN ANALYSIS:" -ForegroundColor Cyan
    Write-Host ("═" * 60) -ForegroundColor Cyan

    foreach ($symbol in $capturedTicks.Keys) {
        $ticks = $capturedTicks[$symbol]

        if ($ticks.Count -lt 3) {
            continue
        }

        # Calculate price changes
        $priceChanges = @()
        for ($i = 1; $i -lt $ticks.Count; $i++) {
            $change = $ticks[$i].bid - $ticks[$i-1].bid
            $priceChanges += $change
        }

        # Check regularity
        $uniqueChanges = ($priceChanges | ForEach-Object { [Math]::Round($_, 6) } | Sort-Object -Unique).Count
        $regularityRatio = if ($priceChanges.Count -gt 0) { $uniqueChanges / $priceChanges.Count } else { 0 }

        Write-Host "  $symbol`:"
        Write-Host "    - Price changes: $($priceChanges.Count)"
        Write-Host "    - Unique changes: $uniqueChanges"
        Write-Host "    - Regularity ratio: $($regularityRatio.ToString('F3')) (lower = more regular)"

        if ($regularityRatio -lt 0.5) {
            $evidence += "⚠️  $symbol`: Low regularity ratio ($($regularityRatio.ToString('F3'))) suggests simulated data"
        } else {
            $evidence += "✅ $symbol`: High price variation ($($regularityRatio.ToString('F3'))) suggests real market data"
        }

        # Sample changes
        $sampleChanges = ($priceChanges | Select-Object -First 5 | ForEach-Object { $_.ToString('F6') }) -join ', '
        Write-Host "    - Sample changes: [$sampleChanges]"
    }

    # FINAL VERDICT
    Write-Host "`n`n$("═" * 60)" -ForegroundColor Cyan
    Write-Host "🔍 FINAL VERDICT" -ForegroundColor Cyan
    Write-Host ("═" * 60) -ForegroundColor Cyan

    Write-Host "`nEVIDENCE SUMMARY:"
    foreach ($e in $evidence) {
        Write-Host "  $e"
    }

    $simulatedIndicators = ($evidence | Where-Object { $_ -match '⚠️|❌' }).Count
    $realIndicators = ($evidence | Where-Object { $_ -match '✅' }).Count

    Write-Host "`n$("─" * 60)"
    Write-Host "Simulated indicators: $simulatedIndicators"
    Write-Host "Real data indicators: $realIndicators"
    Write-Host ("─" * 60)

    # Determine verdict
    if ($lpSources.ContainsKey('SIMULATED')) {
        Write-Host "`n🎭 VERDICT: SIMULATED DATA" -ForegroundColor Red
        Write-Host "   The LP field explicitly shows `"SIMULATED`""
    } elseif ($lpSources.ContainsKey('YOFX') -and $realIndicators -gt $simulatedIndicators) {
        Write-Host "`n✅ VERDICT: REAL MARKET DATA" -ForegroundColor Green
        Write-Host "   LP field shows `"YOFX`" (FIX gateway)"
        Write-Host "   Price patterns and timestamps show real market characteristics"
    } elseif ($simulatedIndicators -gt $realIndicators) {
        Write-Host "`n⚠️  VERDICT: LIKELY SIMULATED" -ForegroundColor Yellow
        Write-Host "   Patterns suggest simulation despite LP field"
    } else {
        Write-Host "`n❓ VERDICT: INCONCLUSIVE" -ForegroundColor Yellow
        Write-Host "   Mixed indicators - need more data"
    }

    $duration = ((Get-Date) - $startTime).TotalSeconds
    Write-Host "`n$("═" * 60)" -ForegroundColor Cyan
    Write-Host "Total ticks analyzed: $totalTickCount"
    Write-Host "Symbols detected: $($symbols -join ', ')"
    Write-Host "Duration: $($duration.ToString('F1'))s"
    Write-Host ("═" * 60) -ForegroundColor Cyan

} catch {
    Write-Host "`n❌ Error: $_" -ForegroundColor Red
    exit 1
} finally {
    if ($ws) {
        $ws.Dispose()
    }
}
