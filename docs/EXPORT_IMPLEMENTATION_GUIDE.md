# Export & Reporting Implementation Guide
## Ready-to-Use Code Templates and Integration Steps

---

## Quick Reference: Library Installation

### Backend (Go)

```bash
# CSV (built-in, no install needed)

# Excel
go get github.com/xuri/excelize/v2

# PDF
go get github.com/go-pdf/fpdf

# Scheduling
go get github.com/robfig/cron/v3

# AWS S3
go get github.com/aws/aws-sdk-go-v2/...

# SFTP
go get github.com/pkg/sftp
go get golang.org/x/crypto/ssh

# Parquet (optional)
go get github.com/xitongsys/parquet-go
```

### Frontend (TypeScript/React)

```bash
# CSV parsing
npm install papaparse
npm install --save-dev @types/papaparse

# Excel generation
npm install exceljs

# PDF generation
npm install jspdf html2canvas

# Date utilities
npm install date-fns

# Form validation
npm install zod

# Charts (if needed)
npm install recharts
```

---

## Step 1: Setup Export Directory Structure

```bash
mkdir -p backend/features/export
mkdir -p backend/internal/api/handlers/export
mkdir -p clients/desktop/src/services/export
mkdir -p clients/desktop/src/components/export
```

---

## Step 2: Backend CSV Exporter Implementation

### File: `backend/features/export/csv_exporter.go`

```go
package export

import (
    "encoding/csv"
    "fmt"
    "io"
    "time"
)

// CSVExporter provides CSV export functionality
type CSVExporter struct {
    writer *csv.Writer
}

// NewCSVExporter creates a new CSV exporter
func NewCSVExporter() *CSVExporter {
    return &CSVExporter{}
}

// ExportOptions configures CSV export behavior
type ExportOptions struct {
    Delimiter      rune
    UseCRLF        bool
    IncludeHeaders bool
    DateFormat     string
}

// DefaultExportOptions returns default CSV options
func DefaultExportOptions() ExportOptions {
    return ExportOptions{
        Delimiter:      ',',
        UseCRLF:        false,
        IncludeHeaders: true,
        DateFormat:     time.RFC3339,
    }
}

// ExportTrades exports trade records to CSV
func (e *CSVExporter) ExportTrades(
    w io.Writer,
    trades []Trade,
    opts ExportOptions,
) error {
    csvWriter := csv.NewWriter(w)
    csvWriter.Comma = opts.Delimiter
    csvWriter.UseCRLF = opts.UseCRLF
    defer csvWriter.Flush()

    if opts.IncludeHeaders {
        headers := []string{
            "ID", "Symbol", "Side", "Volume",
            "Open Price", "Close Price", "Profit",
            "Commission", "Swap", "Open Time", "Close Time",
            "Holding Time (Seconds)", "MAE", "MFE", "Exit Reason",
        }
        if err := csvWriter.Write(headers); err != nil {
            return fmt.Errorf("failed to write CSV headers: %w", err)
        }
    }

    for _, trade := range trades {
        row := []string{
            trade.ID,
            trade.Symbol,
            trade.Side,
            fmt.Sprintf("%.2f", trade.Volume),
            fmt.Sprintf("%.5f", trade.OpenPrice),
            fmt.Sprintf("%.5f", trade.ClosePrice),
            fmt.Sprintf("%.2f", trade.Profit),
            fmt.Sprintf("%.2f", trade.Commission),
            fmt.Sprintf("%.2f", trade.Swap),
            trade.OpenTime.Format(opts.DateFormat),
            trade.CloseTime.Format(opts.DateFormat),
            fmt.Sprintf("%d", trade.HoldingTime),
            fmt.Sprintf("%.5f", trade.MAE),
            fmt.Sprintf("%.5f", trade.MFE),
            trade.ExitReason,
        }
        if err := csvWriter.Write(row); err != nil {
            return fmt.Errorf("failed to write CSV row: %w", err)
        }
    }

    return nil
}

// ExportPerformanceMetrics exports performance metrics
func (e *CSVExporter) ExportPerformanceMetrics(
    w io.Writer,
    report *PerformanceReport,
    opts ExportOptions,
) error {
    csvWriter := csv.NewWriter(w)
    csvWriter.Comma = opts.Delimiter
    defer csvWriter.Flush()

    // Write headers
    csvWriter.Write([]string{"Metric", "Value"})

    metrics := [][]string{
        {"Period Start", report.StartDate.Format(opts.DateFormat)},
        {"Period End", report.EndDate.Format(opts.DateFormat)},
        {"Total Trades", fmt.Sprintf("%d", report.TotalTrades)},
        {"Winning Trades", fmt.Sprintf("%d", report.WinningTrades)},
        {"Losing Trades", fmt.Sprintf("%d", report.LosingTrades)},
        {"Win Rate (%)", fmt.Sprintf("%.2f", report.WinRate)},
        {"Net Profit ($)", fmt.Sprintf("%.2f", report.NetProfit)},
        {"Total Profit ($)", fmt.Sprintf("%.2f", report.TotalProfit)},
        {"Total Loss ($)", fmt.Sprintf("%.2f", report.TotalLoss)},
        {"Average Trade ($)", fmt.Sprintf("%.2f", report.AverageTrade)},
        {"Average Win ($)", fmt.Sprintf("%.2f", report.AverageWin)},
        {"Average Loss ($)", fmt.Sprintf("%.2f", report.AverageLoss)},
        {"Largest Win ($)", fmt.Sprintf("%.2f", report.LargestWin)},
        {"Largest Loss ($)", fmt.Sprintf("%.2f", report.LargestLoss)},
        {"Profit Factor", fmt.Sprintf("%.2f", report.ProfitFactor)},
        {"Sharpe Ratio", fmt.Sprintf("%.2f", report.SharpeRatio)},
        {"Sortino Ratio", fmt.Sprintf("%.2f", report.SortinoRatio)},
        {"Calmar Ratio", fmt.Sprintf("%.2f", report.CalmarRatio)},
        {"Max Drawdown ($)", fmt.Sprintf("%.2f", report.MaxDrawdown)},
        {"Max Drawdown (%)", fmt.Sprintf("%.2f", report.MaxDrawdownPct)},
        {"Recovery Factor", fmt.Sprintf("%.2f", report.RecoveryFactor)},
        {"Longest Win Streak", fmt.Sprintf("%d", report.LongestWinStreak)},
        {"Longest Loss Streak", fmt.Sprintf("%d", report.LongestLossStreak)},
        {"Avg Holding Time (hrs)", fmt.Sprintf("%.2f", float64(report.AverageHoldingTime)/3600)},
    }

    for _, row := range metrics {
        if err := csvWriter.Write(row); err != nil {
            return fmt.Errorf("failed to write metrics row: %w", err)
        }
    }

    // Symbol breakdown
    csvWriter.Write([]string{})
    csvWriter.Write([]string{"Symbol Attribution"})
    csvWriter.Write([]string{"Symbol", "Profit ($)", "Trades", "Win %"})

    for symbol, profit := range report.ProfitBySymbol {
        trades := report.TradesBySymbol[symbol]
        // Calculate win % for this symbol
        winPct := 0.0 // would need to calculate from trades
        row := []string{
            symbol,
            fmt.Sprintf("%.2f", profit),
            fmt.Sprintf("%d", trades),
            fmt.Sprintf("%.2f", winPct),
        }
        csvWriter.Write(row)
    }

    return nil
}
```

---

## Step 3: Backend Excel Exporter Implementation

### File: `backend/features/export/excel_exporter.go`

```go
package export

import (
    "fmt"
    "bytes"
    "time"
    "github.com/xuri/excelize/v2"
)

// ExcelExporter generates Excel files
type ExcelExporter struct {
    file *excelize.File
}

// NewExcelExporter creates a new Excel exporter
func NewExcelExporter() *ExcelExporter {
    return &ExcelExporter{
        file: excelize.NewFile(),
    }
}

// ExportPerformanceReport generates a complete performance report Excel
func (e *ExcelExporter) ExportPerformanceReport(
    report *PerformanceReport,
    trades []Trade,
) (*bytes.Buffer, error) {
    f := excelize.NewFile()
    defer f.Close()

    // Remove default sheet
    f.DeleteSheet("Sheet1")

    // Create sheets
    if err := e.createSummarySheet(f, report); err != nil {
        return nil, err
    }

    if err := e.createTradesSheet(f, trades); err != nil {
        return nil, err
    }

    if err := e.createAttributionSheet(f, report); err != nil {
        return nil, err
    }

    if err := e.createStatisticsSheet(f, report); err != nil {
        return nil, err
    }

    // Write to buffer
    buf, err := f.WriteToBuffer()
    if err != nil {
        return nil, fmt.Errorf("failed to write Excel buffer: %w", err)
    }

    return buf, nil
}

// createSummarySheet creates the summary sheet
func (e *ExcelExporter) createSummarySheet(
    f *excelize.File,
    report *PerformanceReport,
) error {
    sheetName := "Summary"
    f.NewSheet(sheetName)

    // Title
    f.SetCellValue(sheetName, "A1", "Performance Report Summary")
    f.SetCellFont(sheetName, "A1", &excelize.Font{
        Bold: true,
        Size: 14,
    })

    // Period
    f.SetCellValue(sheetName, "A2", "Period:")
    f.SetCellValue(sheetName, "B2",
        fmt.Sprintf("%s to %s",
            report.StartDate.Format("2006-01-02"),
            report.EndDate.Format("2006-01-02")))

    f.SetCellValue(sheetName, "A3", "Generated:")
    f.SetCellValue(sheetName, "B3", time.Now().Format("2006-01-02 15:04:05"))

    // Metrics table
    row := 5
    f.SetCellValue(sheetName, "A5", "Key Metrics")
    f.SetCellFont(sheetName, "A5", &excelize.Font{Bold: true})

    row = 6
    f.SetCellValue(sheetName, "A6", "Metric")
    f.SetCellValue(sheetName, "B6", "Value")
    f.SetCellStyle(sheetName, "A6", "B6", e.getHeaderStyle(f))

    metrics := []struct {
        label string
        value interface{}
    }{
        {"Total Trades", report.TotalTrades},
        {"Winning Trades", report.WinningTrades},
        {"Losing Trades", report.LosingTrades},
        {"Win Rate (%)", fmt.Sprintf("%.2f", report.WinRate)},
        {"Net Profit ($)", fmt.Sprintf("%.2f", report.NetProfit)},
        {"Gross Profit ($)", fmt.Sprintf("%.2f", report.TotalProfit)},
        {"Gross Loss ($)", fmt.Sprintf("%.2f", report.TotalLoss)},
        {"Average Trade ($)", fmt.Sprintf("%.2f", report.AverageTrade)},
        {"Largest Win ($)", fmt.Sprintf("%.2f", report.LargestWin)},
        {"Largest Loss ($)", fmt.Sprintf("%.2f", report.LargestLoss)},
        {"Profit Factor", fmt.Sprintf("%.2f", report.ProfitFactor)},
        {"Sharpe Ratio", fmt.Sprintf("%.2f", report.SharpeRatio)},
        {"Sortino Ratio", fmt.Sprintf("%.2f", report.SortinoRatio)},
        {"Calmar Ratio", fmt.Sprintf("%.2f", report.CalmarRatio)},
        {"Max Drawdown (%)", fmt.Sprintf("%.2f", report.MaxDrawdownPct)},
        {"Recovery Factor", fmt.Sprintf("%.2f", report.RecoveryFactor)},
        {"Longest Win Streak", report.LongestWinStreak},
        {"Longest Loss Streak", report.LongestLossStreak},
    }

    for i, m := range metrics {
        row = 6 + i + 1
        f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), m.label)
        f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), m.value)
    }

    // Set column widths
    f.SetColWidth(sheetName, "A", "A", 25)
    f.SetColWidth(sheetName, "B", "B", 20)

    return nil
}

// createTradesSheet creates detailed trades sheet
func (e *ExcelExporter) createTradesSheet(
    f *excelize.File,
    trades []Trade,
) error {
    sheetName := "Trades"
    f.NewSheet(sheetName)

    // Headers
    headers := []string{
        "ID", "Symbol", "Side", "Volume",
        "Open Price", "Close Price", "Profit",
        "Commission", "Swap", "Open Time", "Close Time",
        "Holding (Hours)", "MAE", "MFE", "Exit Reason",
    }

    for col, header := range headers {
        cellRef := fmt.Sprintf("%s1", excelize.ToCol(col+1))
        f.SetCellValue(sheetName, cellRef, header)
        f.SetCellStyle(sheetName, cellRef, cellRef, e.getHeaderStyle(f))
    }

    // Data rows
    for i, trade := range trades {
        row := i + 2
        f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), trade.ID)
        f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), trade.Symbol)
        f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), trade.Side)
        f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), trade.Volume)
        f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), trade.OpenPrice)
        f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), trade.ClosePrice)
        f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), trade.Profit)
        f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), trade.Commission)
        f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), trade.Swap)
        f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), trade.OpenTime.Format(time.RFC3339))
        f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), trade.CloseTime.Format(time.RFC3339))
        f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), float64(trade.HoldingTime)/3600)
        f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), trade.MAE)
        f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), trade.MFE)
        f.SetCellValue(sheetName, fmt.Sprintf("O%d", row), trade.ExitReason)
    }

    // Auto-fit columns
    for col := 1; col <= len(headers); col++ {
        f.SetColWidth(sheetName, excelize.ToCol(col), excelize.ToCol(col), 15)
    }

    return nil
}

// createAttributionSheet creates symbol and time attribution
func (e *ExcelExporter) createAttributionSheet(
    f *excelize.File,
    report *PerformanceReport,
) error {
    sheetName := "Attribution"
    f.NewSheet(sheetName)

    // Symbol attribution
    f.SetCellValue(sheetName, "A1", "Symbol Attribution")
    f.SetCellFont(sheetName, "A1", &excelize.Font{Bold: true})

    f.SetCellValue(sheetName, "A2", "Symbol")
    f.SetCellValue(sheetName, "B2", "Profit")
    f.SetCellValue(sheetName, "C2", "Trades")
    f.SetCellStyle(sheetName, "A2", "C2", e.getHeaderStyle(f))

    row := 3
    for symbol, profit := range report.ProfitBySymbol {
        f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), symbol)
        f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), profit)
        f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), report.TradesBySymbol[symbol])
        row++
    }

    // Time attribution (hourly)
    f.SetCellValue(sheetName, "E1", "Hourly Attribution")
    f.SetCellFont(sheetName, "E1", &excelize.Font{Bold: true})

    f.SetCellValue(sheetName, "E2", "Hour")
    f.SetCellValue(sheetName, "F2", "Profit")
    f.SetCellStyle(sheetName, "E2", "F2", e.getHeaderStyle(f))

    row = 3
    for hour := 0; hour < 24; hour++ {
        if profit, ok := report.ProfitByHour[hour]; ok {
            f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("%02d:00", hour))
            f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), profit)
            row++
        }
    }

    return nil
}

// createStatisticsSheet creates detailed statistics
func (e *ExcelExporter) createStatisticsSheet(
    f *excelize.File,
    report *PerformanceReport,
) error {
    sheetName := "Statistics"
    f.NewSheet(sheetName)

    // Risk metrics
    f.SetCellValue(sheetName, "A1", "Risk Analysis")
    f.SetCellFont(sheetName, "A1", &excelize.Font{Bold: true})

    f.SetCellValue(sheetName, "A2", "Metric")
    f.SetCellValue(sheetName, "B2", "Value")

    stats := []struct {
        label string
        value interface{}
    }{
        {"Average Win", fmt.Sprintf("%.2f", report.AverageWin)},
        {"Average Loss", fmt.Sprintf("%.2f", report.AverageLoss)},
        {"Win/Loss Ratio", fmt.Sprintf("%.2f", report.AverageWin/(-1*report.AverageLoss))},
        {"Average MAE", fmt.Sprintf("%.5f", report.AverageMAE)},
        {"Average MFE", fmt.Sprintf("%.5f", report.AverageMFE)},
        {"MAE/MFE Ratio", fmt.Sprintf("%.2f", report.MAEMFERatio)},
        {"Avg R-Multiple", fmt.Sprintf("%.2f", report.AverageRMultiple)},
        {"Median R-Multiple", fmt.Sprintf("%.2f", report.MedianRMultiple)},
    }

    for i, s := range stats {
        row := 2 + i + 1
        f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), s.label)
        f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), s.value)
    }

    return nil
}

// Helper method to get header style
func (e *ExcelExporter) getHeaderStyle(f *excelize.File) int {
    style, _ := f.NewStyle(&excelize.Style{
        Fill: excelize.Fill{
            Type:    "pattern",
            Color:   []string{"D3D3D3"},
            Pattern: 1,
        },
        Font: &excelize.Font{
            Bold: true,
        },
    })
    return style
}
```

---

## Step 4: API Handler Implementation

### File: `backend/internal/api/handlers/export/export.go`

```go
package export

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/epic1st/rtx/backend/features"
    exportpkg "github.com/epic1st/rtx/backend/features/export"
)

// Handler contains export handlers
type Handler struct {
    reportService *features.ReportService
    auditLog      *features.ExportAuditLog
}

// New creates a new export handler
func New(
    reportService *features.ReportService,
    auditLog *features.ExportAuditLog,
) *Handler {
    return &Handler{
        reportService: reportService,
        auditLog:      auditLog,
    }
}

// ExportTradesCSV exports trades as CSV
// GET /api/v1/export/trades?format=csv&start=2024-01-01&end=2024-01-31
func (h *Handler) ExportTradesCSV(w http.ResponseWriter, r *http.Request) {
    accountID := r.Header.Get("X-Account-ID")
    if accountID == "" {
        http.Error(w, "missing account ID", http.StatusBadRequest)
        return
    }

    // Parse dates
    startStr := r.URL.Query().Get("start")
    endStr := r.URL.Query().Get("end")

    startDate, err := time.Parse("2006-01-02", startStr)
    if err != nil {
        http.Error(w, "invalid start date format", http.StatusBadRequest)
        return
    }

    endDate, err := time.Parse("2006-01-02", endStr)
    if err != nil {
        http.Error(w, "invalid end date format", http.StatusBadRequest)
        return
    }

    // Get trades
    trades, err := h.reportService.getTrades(accountID, startDate, endDate)
    if err != nil {
        log.Printf("[Export Error] Failed to get trades: %v", err)
        http.Error(w, "failed to retrieve trades", http.StatusInternalServerError)
        return
    }

    // Export to CSV
    exporter := exportpkg.NewCSVExporter()
    opts := exportpkg.DefaultExportOptions()

    var buf bytes.Buffer
    if err := exporter.ExportTrades(&buf, trades, opts); err != nil {
        log.Printf("[Export Error] Failed to export CSV: %v", err)
        http.Error(w, "failed to export trades", http.StatusInternalServerError)
        return
    }

    // Record audit log
    timestamp := time.Now().Format("20060102_150405")
    auditID := h.auditLog.RecordExport(
        accountID,
        r.Header.Get("X-User-ID"),
        "export",
        "trades",
        "csv",
        len(trades),
        getClientIP(r),
        r.UserAgent(),
    )

    // Send response
    filename := fmt.Sprintf("trades_%s.csv", timestamp)
    w.Header().Set("Content-Type", "text/csv")
    w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
    w.Header().Set("X-Audit-ID", auditID)
    w.Write(buf.Bytes())

    log.Printf("[Export Success] CSV export: %s (%d trades, %d bytes)", filename, len(trades), buf.Len())
}

// ExportTradesExcel exports trades as Excel
// GET /api/v1/export/trades?format=excel&start=2024-01-01&end=2024-01-31
func (h *Handler) ExportTradesExcel(w http.ResponseWriter, r *http.Request) {
    accountID := r.Header.Get("X-Account-ID")
    if accountID == "" {
        http.Error(w, "missing account ID", http.StatusBadRequest)
        return
    }

    // Parse dates
    startStr := r.URL.Query().Get("start")
    endStr := r.URL.Query().Get("end")

    startDate, _ := time.Parse("2006-01-02", startStr)
    endDate, _ := time.Parse("2006-01-02", endStr)

    // Get trades and report
    trades, _ := h.reportService.getTrades(accountID, startDate, endDate)
    report, _ := h.reportService.GeneratePerformanceReport(accountID, startDate, endDate)

    // Export to Excel
    exporter := exportpkg.NewExcelExporter()
    buf, err := exporter.ExportPerformanceReport(report, trades)
    if err != nil {
        log.Printf("[Export Error] Failed to export Excel: %v", err)
        http.Error(w, "failed to export report", http.StatusInternalServerError)
        return
    }

    // Record audit
    timestamp := time.Now().Format("20060102_150405")
    auditID := h.auditLog.RecordExport(
        accountID,
        r.Header.Get("X-User-ID"),
        "export",
        "performance_report",
        "excel",
        len(trades),
        getClientIP(r),
        r.UserAgent(),
    )

    // Send response
    filename := fmt.Sprintf("report_%s.xlsx", timestamp)
    w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
    w.Header().Set("X-Audit-ID", auditID)
    w.Write(buf.Bytes())

    log.Printf("[Export Success] Excel export: %s", filename)
}

// ExportTradesJSON exports trades as JSON
// GET /api/v1/export/trades?format=json&start=2024-01-01&end=2024-01-31
func (h *Handler) ExportTradesJSON(w http.ResponseWriter, r *http.Request) {
    accountID := r.Header.Get("X-Account-ID")
    startStr := r.URL.Query().Get("start")
    endStr := r.URL.Query().Get("end")

    startDate, _ := time.Parse("2006-01-02", startStr)
    endDate, _ := time.Parse("2006-01-02", endStr)

    trades, _ := h.reportService.getTrades(accountID, startDate, endDate)

    // Record audit
    h.auditLog.RecordExport(
        accountID,
        r.Header.Get("X-User-ID"),
        "export",
        "trades",
        "json",
        len(trades),
        getClientIP(r),
        r.UserAgent(),
    )

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(trades)
}

// Helper to get client IP
func getClientIP(r *http.Request) string {
    if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
        return ip
    }
    return r.RemoteAddr
}
```

---

## Step 5: Frontend Export Component

### File: `clients/desktop/src/components/export/ExportDialog.tsx`

```typescript
import React, { useState } from 'react';
import { Download, FileText, BarChart3 } from 'lucide-react';

interface ExportDialogProps {
  accountId: string;
  onClose: () => void;
}

export const ExportDialog: React.FC<ExportDialogProps> = ({
  accountId,
  onClose,
}) => {
  const [format, setFormat] = useState<'csv' | 'excel' | 'json'>('csv');
  const [reportType, setReportType] = useState<'trades' | 'performance'>('trades');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleExport = async () => {
    if (!startDate || !endDate) {
      alert('Please select both start and end dates');
      return;
    }

    setIsLoading(true);
    try {
      const queryParams = new URLSearchParams({
        start: startDate,
        end: endDate,
        format,
      });

      const endpoint = `/api/v1/export/${reportType}?${queryParams}`;
      const response = await fetch(endpoint, {
        method: 'GET',
        headers: {
          'X-Account-ID': accountId,
          'X-User-ID': localStorage.getItem('userId') || '',
        },
      });

      if (!response.ok) {
        throw new Error(`Export failed: ${response.statusText}`);
      }

      // Get filename from header or generate
      const filename = extractFilename(response.headers) ||
        `export_${Date.now()}.${getFileExtension(format)}`;

      // Download file
      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);

      onClose();
    } catch (error) {
      console.error('Export error:', error);
      alert('Failed to export data');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 max-w-md w-full">
        <h2 className="text-xl font-bold text-white mb-4">Export Data</h2>

        {/* Report Type */}
        <div className="mb-4">
          <label className="block text-sm text-zinc-400 mb-2">Report Type</label>
          <select
            value={reportType}
            onChange={(e) => setReportType(e.target.value as any)}
            className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-white"
          >
            <option value="trades">Trades</option>
            <option value="performance">Performance Report</option>
          </select>
        </div>

        {/* Format Selection */}
        <div className="mb-4">
          <label className="block text-sm text-zinc-400 mb-2">Format</label>
          <div className="grid grid-cols-3 gap-2">
            {['csv', 'excel', 'json'].map((fmt) => (
              <button
                key={fmt}
                onClick={() => setFormat(fmt as any)}
                className={`p-2 rounded border ${
                  format === fmt
                    ? 'bg-blue-600 border-blue-500'
                    : 'bg-zinc-800 border-zinc-700'
                } text-white text-sm font-medium`}
              >
                {fmt.toUpperCase()}
              </button>
            ))}
          </div>
        </div>

        {/* Date Range */}
        <div className="mb-4">
          <label className="block text-sm text-zinc-400 mb-2">Start Date</label>
          <input
            type="date"
            value={startDate}
            onChange={(e) => setStartDate(e.target.value)}
            className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-white"
          />
        </div>

        <div className="mb-6">
          <label className="block text-sm text-zinc-400 mb-2">End Date</label>
          <input
            type="date"
            value={endDate}
            onChange={(e) => setEndDate(e.target.value)}
            className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-white"
          />
        </div>

        {/* Actions */}
        <div className="flex gap-3">
          <button
            onClick={onClose}
            className="flex-1 px-4 py-2 bg-zinc-800 border border-zinc-700 rounded text-white hover:bg-zinc-700"
          >
            Cancel
          </button>
          <button
            onClick={handleExport}
            disabled={isLoading}
            className="flex-1 px-4 py-2 bg-blue-600 border border-blue-500 rounded text-white hover:bg-blue-700 disabled:opacity-50 flex items-center justify-center gap-2"
          >
            <Download className="w-4 h-4" />
            {isLoading ? 'Exporting...' : 'Export'}
          </button>
        </div>
      </div>
    </div>
  );
};

function getFileExtension(format: string): string {
  const extensions: Record<string, string> = {
    csv: 'csv',
    excel: 'xlsx',
    json: 'json',
  };
  return extensions[format] || format;
}

function extractFilename(headers: Headers): string | null {
  const disposition = headers.get('Content-Disposition');
  if (!disposition) return null;

  const match = disposition.match(/filename="?([^"]+)"?/);
  return match ? match[1] : null;
}
```

---

## Step 6: Integration in Main Server

### File: `backend/cmd/server/main.go` (relevant sections)

```go
// In setupRoutes or similar function
func setupExportRoutes(mux *http.ServeMux, services *Services) {
    exportHandler := export.New(
        services.ReportService,
        services.AuditLog,
    )

    // Export endpoints
    mux.HandleFunc("GET /api/v1/export/trades", exportHandler.ExportTradesCSV)
    mux.HandleFunc("GET /api/v1/export/report", exportHandler.ExportTradesExcel)
    mux.HandleFunc("GET /api/v1/export/json", exportHandler.ExportTradesJSON)

    log.Println("[Routes] Export endpoints registered")
}

// In main()
func main() {
    // ... existing setup ...

    // Initialize audit log
    auditLog := features.NewExportAuditLog()

    // Initialize report service
    reportService := features.NewReportService()

    services := &Services{
        ReportService: reportService,
        AuditLog:      auditLog,
    }

    // Setup routes
    setupExportRoutes(mux, services)

    // ... rest of setup ...
}
```

---

## Step 7: Frontend Integration

### File: `clients/desktop/src/components/AccountInfoDashboard.tsx` (add export button)

```typescript
import { ExportDialog } from './export/ExportDialog';

export const AccountInfoDashboard = () => {
  const [showExportDialog, setShowExportDialog] = useState(false);
  const { account } = useAppStore();

  // ... existing code ...

  return (
    <div className="space-y-4">
      {/* Existing content ... */}

      {/* Export Button */}
      <button
        onClick={() => setShowExportDialog(true)}
        className="w-full px-4 py-2 bg-blue-600 border border-blue-500 rounded text-white hover:bg-blue-700 flex items-center justify-center gap-2"
      >
        <Download className="w-4 h-4" />
        Export Data
      </button>

      {/* Export Dialog */}
      {showExportDialog && (
        <ExportDialog
          accountId={account?.id || ''}
          onClose={() => setShowExportDialog(false)}
        />
      )}
    </div>
  );
};
```

---

## Step 8: Testing Export Functionality

### File: `backend/features/export/export_test.go`

```go
package export

import (
    "bytes"
    "testing"
    "time"
)

func TestCSVExport(t *testing.T) {
    trades := []Trade{
        {
            ID:         "1",
            Symbol:     "EURUSD",
            Side:       "BUY",
            Volume:     1.0,
            OpenPrice:  1.0950,
            ClosePrice: 1.0975,
            Profit:     25.0,
            OpenTime:   time.Now(),
            CloseTime:  time.Now().Add(1 * time.Hour),
        },
    }

    exporter := NewCSVExporter()
    opts := DefaultExportOptions()

    var buf bytes.Buffer
    err := exporter.ExportTrades(&buf, trades, opts)

    if err != nil {
        t.Fatalf("ExportTrades failed: %v", err)
    }

    if buf.Len() == 0 {
        t.Fatal("Expected non-empty CSV output")
    }

    csv := buf.String()
    if !contains(csv, "EURUSD") {
        t.Fatal("CSV missing symbol")
    }

    if !contains(csv, "BUY") {
        t.Fatal("CSV missing side")
    }
}

func TestExcelExport(t *testing.T) {
    trades := []Trade{
        {
            ID:         "1",
            Symbol:     "EURUSD",
            Side:       "BUY",
            Volume:     1.0,
            OpenPrice:  1.0950,
            ClosePrice: 1.0975,
            Profit:     25.0,
            OpenTime:   time.Now(),
            CloseTime:  time.Now(),
        },
    }

    report := &PerformanceReport{
        TotalTrades:    1,
        WinningTrades:  1,
        NetProfit:      25.0,
        ProfitBySymbol: map[string]float64{"EURUSD": 25.0},
        TradesBySymbol: map[string]int{"EURUSD": 1},
    }

    exporter := NewExcelExporter()
    buf, err := exporter.ExportPerformanceReport(report, trades)

    if err != nil {
        t.Fatalf("ExportPerformanceReport failed: %v", err)
    }

    if buf.Len() == 0 {
        t.Fatal("Expected non-empty Excel output")
    }
}

func contains(s, substr string) bool {
    return len(s) > 0 && len(substr) > 0 &&
           (s == substr || len(s) > len(substr) &&
            bytes.Contains([]byte(s), []byte(substr)))
}
```

---

## Quick Checklist

- [x] Backend CSV exporter implemented
- [x] Backend Excel exporter implemented
- [x] API handlers for export endpoints
- [x] Frontend export dialog component
- [x] Export audit logging
- [ ] Scheduled report generation (use robfig/cron)
- [ ] Email delivery integration (SMTP)
- [ ] S3 cloud storage integration
- [ ] PDF generation
- [ ] Webhook notifications
- [ ] GDPR compliance features

---

## Next Steps

1. **Test export endpoints locally:**
   ```bash
   curl "http://localhost:8080/api/v1/export/trades?start=2024-01-01&end=2024-01-31" \
     -H "X-Account-ID: test-account" \
     -H "X-User-ID: test-user" \
     -o trades.csv
   ```

2. **Add export to navigation menu**

3. **Implement email scheduling** using `robfig/cron`

4. **Add cloud storage integration** with S3

5. **Set up webhook notifications**

6. **Add GDPR data export endpoint**

---

## Troubleshooting

**CSV not displaying correctly:**
- Ensure UTF-8 encoding
- Check delimiter settings
- Verify date format

**Excel file is blank:**
- Verify sheet name is set
- Check cell references
- Ensure data is properly written

**Export timeout:**
- Implement pagination for large datasets
- Use streaming instead of buffering
- Add progress tracking

**Memory issues with large exports:**
- Use io.Writer instead of bytes.Buffer
- Process trades in batches
- Stream directly to HTTP response
