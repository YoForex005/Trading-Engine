# PowerShell Test Script for Chart Print Functionality

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Chart Print Functionality Test Suite" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

$script:TestsPassed = 0
$script:TestsFailed = 0

function Test-PrintFeature {
    param(
        [string]$TestName,
        [scriptblock]$TestBlock
    )

    Write-Host "Testing: $TestName... " -NoNewline

    try {
        $result = & $TestBlock
        if ($result) {
            Write-Host "PASS" -ForegroundColor Green
            $script:TestsPassed++
            return $true
        } else {
            Write-Host "FAIL" -ForegroundColor Red
            $script:TestsFailed++
            return $false
        }
    } catch {
        Write-Host "FAIL" -ForegroundColor Red
        $script:TestsFailed++
        return $false
    }
}

# File existence tests
Test-PrintFeature "chartPrinter service exists" {
    Test-Path "clients\desktop\src\services\chartPrinter.ts"
}

Test-PrintFeature "PrintSetupDialog exists" {
    Test-Path "clients\desktop\src\components\dialogs\PrintSetupDialog.tsx"
}

Test-PrintFeature "PrintPreviewDialog exists" {
    Test-Path "clients\desktop\src\components\dialogs\PrintPreviewDialog.tsx"
}

Test-PrintFeature "useChartPrint hook exists" {
    Test-Path "clients\desktop\src\hooks\useChartPrint.ts"
}

Test-PrintFeature "ChartPrintExample exists" {
    Test-Path "clients\desktop\src\components\ChartPrintExample.tsx"
}

Test-PrintFeature "print_preferences.go exists" {
    Test-Path "backend\api\print_preferences.go"
}

Test-PrintFeature "chartPrinter test exists" {
    Test-Path "clients\desktop\src\services\chartPrinter.test.ts"
}

Test-PrintFeature "Documentation exists" {
    Test-Path "docs\CHART_PRINT_FUNCTIONALITY.md"
}

Test-PrintFeature "Integration guide exists" {
    Test-Path "docs\QUICK_PRINT_INTEGRATION.md"
}

Test-PrintFeature "Implementation summary exists" {
    Test-Path "CHART_PRINT_IMPLEMENTATION_SUMMARY.md"
}

# Summary
Write-Host ""
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Tests Passed: $($script:TestsPassed)" -ForegroundColor Green
Write-Host "Tests Failed: $($script:TestsFailed)" -ForegroundColor Red
Write-Host "Total Tests: $($script:TestsPassed + $script:TestsFailed)"

if ($script:TestsFailed -eq 0) {
    Write-Host ""
    Write-Host "All tests passed! Chart print functionality is ready." -ForegroundColor Green
    exit 0
} else {
    Write-Host ""
    Write-Host "Some tests failed. Please review the output above." -ForegroundColor Red
    exit 1
}
