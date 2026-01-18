# Export and Reporting Capabilities Research
## Trading Engine Analytics Dashboard

**Research Date:** January 19, 2026
**Status:** Comprehensive Research Complete
**Scope:** Export formats, reporting features, data transformation, integrations, and audit trails

---

## Executive Summary

The trading engine has a **foundation for advanced reporting** with existing performance analytics, tax reporting, and drawdown analysis. However, **export and external reporting integrations are minimal**. This research identifies library recommendations, architectural patterns, and implementation strategies for enterprise-grade export and reporting capabilities.

### Current State Assessment

**Existing Capabilities:**
- Performance report generation (win/loss analysis, Sharpe ratio, Sortino ratio)
- Tax reporting with capital gains classification
- Drawdown analysis with recovery tracking
- Regulatory transaction reporting (MiFID II, EMIR, CAT)
- Real-time dashboard metrics (React-based UI)

**Gaps Identified:**
- No CSV/Excel export functionality
- No PDF report generation with visualizations
- No scheduled report generation (cron-based)
- No email delivery system
- No multi-format templating system
- No cloud storage integrations
- Limited data aggregation options (tick/minute/hour/day levels)
- No time zone handling for international clients
- No currency conversion support
- Missing GDPR-compliant data export
- No webhook notification system
- No SFTP upload for regulatory submissions

---

## 1. Export Format Implementation

### 1.1 CSV Export

**Recommended Library:** `encoding/csv` (Go stdlib) + `papaparse` (JavaScript)

#### Backend Implementation (Go)

```go
// File: backend/features/export.go

package features

import (
    "encoding/csv"
    "fmt"
    "io"
    "time"
)

// CSVExporter handles CSV export with custom columns
type CSVExporter struct {
    writer *csv.Writer
}

// ExportTradesCSV exports trades with configurable columns
func (e *CSVExporter) ExportTradesCSV(
    w io.Writer,
    trades []Trade,
    columns []string, // ["id", "symbol", "side", "profit", ...]
) error {
    csvWriter := csv.NewWriter(w)
    defer csvWriter.Flush()

    // Write headers
    if err := csvWriter.Write(columns); err != nil {
        return fmt.Errorf("failed to write CSV header: %w", err)
    }

    // Write trade rows
    for _, trade := range trades {
        row := e.tradeToCSVRow(trade, columns)
        if err := csvWriter.Write(row); err != nil {
            return fmt.Errorf("failed to write CSV row: %w", err)
        }
    }

    return nil
}

// ExportPerformanceReportCSV exports performance metrics
func (e *CSVExporter) ExportPerformanceReportCSV(
    w io.Writer,
    report *PerformanceReport,
) error {
    csvWriter := csv.NewWriter(w)
    defer csvWriter.Flush()

    // Metrics section
    headers := []string{"Metric", "Value"}
    csvWriter.Write(headers)

    metrics := map[string]interface{}{
        "Total Trades":      report.TotalTrades,
        "Win Rate":         fmt.Sprintf("%.2f%%", report.WinRate),
        "Net Profit":       fmt.Sprintf("$%.2f", report.NetProfit),
        "Sharpe Ratio":     fmt.Sprintf("%.2f", report.SharpeRatio),
        "Sortino Ratio":    fmt.Sprintf("%.2f", report.SortinoRatio),
        "Max Drawdown":     fmt.Sprintf("%.2f%%", report.MaxDrawdownPct),
        "Profit Factor":    fmt.Sprintf("%.2f", report.ProfitFactor),
    }

    for metric, value := range metrics {
        csvWriter.Write([]string{metric, fmt.Sprintf("%v", value)})
    }

    // Symbol breakdown section
    csvWriter.Write([]string{}) // Blank line
    csvWriter.Write([]string{"Symbol", "Profit", "Trades"})

    for symbol, profit := range report.ProfitBySymbol {
        trades := report.TradesBySymbol[symbol]
        csvWriter.Write([]string{
            symbol,
            fmt.Sprintf("$%.2f", profit),
            fmt.Sprintf("%d", trades),
        })
    }

    return nil
}

// Helper to convert trade to CSV row
func (e *CSVExporter) tradeToCSVRow(
    trade Trade,
    columns []string,
) []string {
    columnMap := map[string]string{
        "id":          trade.ID,
        "symbol":     trade.Symbol,
        "side":       trade.Side,
        "volume":     fmt.Sprintf("%.2f", trade.Volume),
        "openPrice":  fmt.Sprintf("%.5f", trade.OpenPrice),
        "closePrice": fmt.Sprintf("%.5f", trade.ClosePrice),
        "profit":     fmt.Sprintf("%.2f", trade.Profit),
        "commission": fmt.Sprintf("%.2f", trade.Commission),
        "openTime":   trade.OpenTime.Format(time.RFC3339),
        "closeTime":  trade.CloseTime.Format(time.RFC3339),
    }

    row := make([]string, len(columns))
    for i, col := range columns {
        row[i] = columnMap[col]
    }
    return row
}
```

#### Frontend Export Component (React)

```typescript
// File: clients/desktop/src/services/exportService.ts

import Papa from 'papaparse';

export interface ExportOptions {
  format: 'csv' | 'json' | 'excel';
  filename: string;
  includeHeaders: boolean;
  dateFormat?: string;
  timezone?: string;
}

export class ExportService {
  static async exportToCSV(
    data: unknown[],
    options: ExportOptions
  ): Promise<Blob> {
    const csv = Papa.unparse(data, {
      header: options.includeHeaders,
      quotes: true,
      quoteChar: '"',
      escapeChar: '"',
      dynamicTyping: false,
      skipEmptyLines: true,
    });

    return new Blob([csv], { type: 'text/csv;charset=utf-8;' });
  }

  static downloadFile(blob: Blob, filename: string): void {
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    link.setAttribute('href', url);
    link.setAttribute('download', filename);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }

  static generateFilename(prefix: string, format: string): string {
    const timestamp = new Date().toISOString().replace(/:/g, '-').slice(0, -5);
    return `${prefix}-${timestamp}.${format}`;
  }
}
```

### 1.2 Excel Export

**Recommended Library:** `excelize` (Go) + `exceljs` (JavaScript)

#### Backend Implementation (Go)

```go
// File: backend/features/export_excel.go

package features

import (
    "fmt"
    "github.com/xuri/excelize/v2"
    "time"
)

// ExcelExporter handles Excel export with formatting
type ExcelExporter struct {
    file *excelize.File
}

// NewExcelExporter creates a new Excel exporter
func NewExcelExporter() *ExcelExporter {
    return &ExcelExporter{
        file: excelize.NewFile(),
    }
}

// ExportPerformanceReportExcel creates formatted Excel report
func (e *ExcelExporter) ExportPerformanceReportExcel(
    report *PerformanceReport,
    trades []Trade,
) ([]byte, error) {
    // Remove default sheet
    e.file.DeleteSheet("Sheet1")

    // Create summary sheet
    e.createSummarySheet(report)

    // Create trades detail sheet
    e.createTradesDetailSheet(trades)

    // Create attribution sheet
    e.createAttributionSheet(report)

    // Create drawdown analysis sheet (if available)
    // e.createDrawdownSheet(analysis)

    // Write to buffer
    buf, err := e.file.WriteToBuffer()
    if err != nil {
        return nil, fmt.Errorf("failed to write Excel: %w", err)
    }

    return buf.Bytes(), nil
}

// createSummarySheet creates performance summary
func (e *ExcelExporter) createSummarySheet(report *PerformanceReport) {
    sheetName := "Summary"
    e.file.NewSheet(sheetName)

    // Headers
    e.file.SetCellValue(sheetName, "A1", "Performance Report Summary")
    e.file.SetCellValue(sheetName, "A2", "Generated")
    e.file.SetCellValue(sheetName, "B2", report.GeneratedAt.Format(time.RFC1123))

    // Metrics table
    row := 4
    metrics := []struct {
        label string
        value interface{}
    }{
        {"Total Trades", report.TotalTrades},
        {"Winning Trades", report.WinningTrades},
        {"Losing Trades", report.LosingTrades},
        {"Win Rate (%)", fmt.Sprintf("%.2f", report.WinRate)},
        {"Net Profit ($)", fmt.Sprintf("%.2f", report.NetProfit)},
        {"Profit Factor", fmt.Sprintf("%.2f", report.ProfitFactor)},
        {"Sharpe Ratio", fmt.Sprintf("%.2f", report.SharpeRatio)},
        {"Sortino Ratio", fmt.Sprintf("%.2f", report.SortinoRatio)},
        {"Max Drawdown (%)", fmt.Sprintf("%.2f", report.MaxDrawdownPct)},
        {"Average Trade ($)", fmt.Sprintf("%.2f", report.AverageTrade)},
    }

    for _, m := range metrics {
        cell := fmt.Sprintf("A%d", row)
        e.file.SetCellValue(sheetName, cell, m.label)
        e.file.SetCellValue(sheetName, fmt.Sprintf("B%d", row), m.value)
        row++
    }

    // Set column widths
    e.file.SetColWidth(sheetName, "A", "A", 25)
    e.file.SetColWidth(sheetName, "B", "B", 20)
}

// createTradesDetailSheet creates detailed trade list
func (e *ExcelExporter) createTradesDetailSheet(trades []Trade) {
    sheetName := "Trades"
    e.file.NewSheet(sheetName)

    // Headers
    headers := []string{
        "ID", "Symbol", "Side", "Volume", "Open Price",
        "Close Price", "Profit ($)", "Commission", "Swap",
        "Open Time", "Close Time", "Holding (Hours)",
    }

    for col, header := range headers {
        cell := excelize.ToCol(col + 1) + "1"
        e.file.SetCellValue(sheetName, cell, header)
    }

    // Trade data
    for i, trade := range trades {
        row := i + 2
        e.file.SetCellValue(sheetName, fmt.Sprintf("A%d", row), trade.ID)
        e.file.SetCellValue(sheetName, fmt.Sprintf("B%d", row), trade.Symbol)
        e.file.SetCellValue(sheetName, fmt.Sprintf("C%d", row), trade.Side)
        e.file.SetCellValue(sheetName, fmt.Sprintf("D%d", row), trade.Volume)
        e.file.SetCellValue(sheetName, fmt.Sprintf("E%d", row), trade.OpenPrice)
        e.file.SetCellValue(sheetName, fmt.Sprintf("F%d", row), trade.ClosePrice)
        e.file.SetCellValue(sheetName, fmt.Sprintf("G%d", row), trade.Profit)
        e.file.SetCellValue(sheetName, fmt.Sprintf("H%d", row), trade.Commission)
        e.file.SetCellValue(sheetName, fmt.Sprintf("I%d", row), trade.Swap)
        e.file.SetCellValue(sheetName, fmt.Sprintf("J%d", row), trade.OpenTime.Format(time.RFC3339))
        e.file.SetCellValue(sheetName, fmt.Sprintf("K%d", row), trade.CloseTime.Format(time.RFC3339))
        e.file.SetCellValue(sheetName, fmt.Sprintf("L%d", row), float64(trade.HoldingTime)/3600)
    }

    // Auto-fit columns
    for col := 1; col <= len(headers); col++ {
        e.file.SetColWidth(sheetName, excelize.ToCol(col), excelize.ToCol(col), 15)
    }
}

// createAttributionSheet creates symbol and time attribution
func (e *ExcelExporter) createAttributionSheet(report *PerformanceReport) {
    sheetName := "Attribution"
    e.file.NewSheet(sheetName)

    // Symbol attribution
    e.file.SetCellValue(sheetName, "A1", "Symbol Attribution")
    row := 2
    e.file.SetCellValue(sheetName, "A2", "Symbol")
    e.file.SetCellValue(sheetName, "B2", "Profit ($)")
    e.file.SetCellValue(sheetName, "C2", "Trades")

    row = 3
    for symbol, profit := range report.ProfitBySymbol {
        e.file.SetCellValue(sheetName, fmt.Sprintf("A%d", row), symbol)
        e.file.SetCellValue(sheetName, fmt.Sprintf("B%d", row), profit)
        e.file.SetCellValue(sheetName, fmt.Sprintf("C%d", row), report.TradesBySymbol[symbol])
        row++
    }

    // Time attribution
    e.file.SetCellValue(sheetName, "D1", "Hour Attribution")
    row = 2
    e.file.SetCellValue(sheetName, "D2", "Hour")
    e.file.SetCellValue(sheetName, "E2", "Profit ($)")

    row = 3
    for hour := 0; hour < 24; hour++ {
        if profit, ok := report.ProfitByHour[hour]; ok {
            e.file.SetCellValue(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("%02d:00", hour))
            e.file.SetCellValue(sheetName, fmt.Sprintf("E%d", row), profit)
            row++
        }
    }
}
```

### 1.3 PDF Export with Charts

**Recommended Libraries:**
- Go: `github.com/go-pdf/fpdf` (lightweight) or `gopdf`
- JavaScript: `pdfkit` or `jsPDF` with `Chart.js`

```typescript
// File: clients/desktop/src/services/pdfExportService.ts

import jsPDF from 'jspdf';
import { PerformanceReport } from '../types/trading';

export class PDFExportService {
  static generatePerformanceReportPDF(
    report: PerformanceReport,
    chartImages: { [key: string]: string } // base64 encoded charts
  ): jsPDF {
    const pdf = new jsPDF({
      orientation: 'portrait',
      unit: 'mm',
      format: 'a4',
    });

    const margin = 15;
    const pageWidth = pdf.internal.pageSize.getWidth();
    const pageHeight = pdf.internal.pageSize.getHeight();
    let yPosition = margin;

    // Title
    pdf.setFontSize(24);
    pdf.text('Performance Report', margin, yPosition);
    yPosition += 15;

    // Date range
    pdf.setFontSize(10);
    pdf.text(
      `Period: ${report.startDate.toLocaleDateString()} - ${report.endDate.toLocaleDateString()}`,
      margin,
      yPosition
    );
    yPosition += 10;

    // Summary metrics
    pdf.setFontSize(12);
    pdf.text('Key Metrics', margin, yPosition);
    yPosition += 8;

    const metrics = [
      [`Win Rate: ${report.winRate.toFixed(2)}%`, `Sharpe Ratio: ${report.sharpeRatio.toFixed(2)}`],
      [`Total Trades: ${report.totalTrades}`, `Profit Factor: ${report.profitFactor.toFixed(2)}`],
      [`Net Profit: $${report.netProfit.toFixed(2)}`, `Max Drawdown: ${report.maxDrawdownPct.toFixed(2)}%`],
    ];

    metrics.forEach((row) => {
      pdf.setFontSize(10);
      pdf.text(row[0], margin, yPosition);
      pdf.text(row[1], pageWidth / 2 + margin, yPosition);
      yPosition += 7;
    });

    yPosition += 5;

    // Add charts
    if (chartImages.equityCurve) {
      pdf.text('Equity Curve', margin, yPosition);
      yPosition += 5;
      pdf.addImage(chartImages.equityCurve, 'PNG', margin, yPosition, 170, 60);
      yPosition += 65;
    }

    if (chartImages.drawdown) {
      pdf.addPage();
      yPosition = margin;
      pdf.text('Drawdown Chart', margin, yPosition);
      yPosition += 5;
      pdf.addImage(chartImages.drawdown, 'PNG', margin, yPosition, 170, 60);
      yPosition += 65;
    }

    return pdf;
  }
}
```

### 1.4 JSON Export (API Integration)

```go
// File: backend/features/export_json.go

package features

import (
    "encoding/json"
    "fmt"
)

// JSONExportResponse wrapper for API responses
type JSONExportResponse struct {
    Status    string      `json:"status"`
    Timestamp time.Time   `json:"timestamp"`
    Data      interface{} `json:"data"`
    Metadata  ExportMetadata `json:"metadata"`
}

type ExportMetadata struct {
    Format    string `json:"format"`
    Records   int    `json:"records"`
    Version   string `json:"version"`
}

// ExportTradesJSON exports trades as JSON
func ExportTradesJSON(trades []Trade) ([]byte, error) {
    response := JSONExportResponse{
        Status:    "success",
        Timestamp: time.Now(),
        Data:      trades,
        Metadata: ExportMetadata{
            Format:  "application/json",
            Records: len(trades),
            Version: "1.0",
        },
    }

    return json.MarshalIndent(response, "", "  ")
}

// Handler for JSON API export
func (h *FeatureHandlers) HandleExportJSON(w http.ResponseWriter, r *http.Request) {
    accountID := r.URL.Query().Get("accountId")
    format := r.URL.Query().Get("format") // "trades", "performance", "drawdown"

    startDate := parseDate(r.URL.Query().Get("startDate"))
    endDate := parseDate(r.URL.Query().Get("endDate"))

    trades, err := h.reportService.getTrades(accountID, startDate, endDate)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    data, err := ExportTradesJSON(trades)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Content-Disposition", `attachment; filename="export.json"`)
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Write(data)
}
```

### 1.5 Parquet Export (Big Data Analysis)

**Recommended Library:** `github.com/xitongsys/parquet-go`

```go
// File: backend/features/export_parquet.go

package features

import (
    "fmt"
    "github.com/xitongsys/parquet-go-source/local"
    "github.com/xitongsys/parquet-go/writer"
)

// ParquetExporter exports to Parquet format for big data tools
type ParquetExporter struct{}

// ExportTradesToParquet exports trades to Parquet
func (e *ParquetExporter) ExportTradesToParquet(
    filePath string,
    trades []Trade,
) error {
    fw, err := local.NewLocalFileWriter(filePath)
    if err != nil {
        return fmt.Errorf("failed to create file writer: %w", err)
    }
    defer fw.Close()

    pw, err := writer.NewCSVWriter(trades, fw, 128)
    if err != nil {
        return fmt.Errorf("failed to create Parquet writer: %w", err)
    }

    if err = pw.Write(trades); err != nil {
        return fmt.Errorf("failed to write Parquet: %w", err)
    }

    if err = pw.WriteStop(); err != nil {
        return fmt.Errorf("failed to finalize Parquet: %w", err)
    }

    return nil
}

// This enables integration with tools like:
// - Apache Spark for distributed analysis
// - DuckDB for OLAP queries
// - Pandas for Python analysis
// - Presto/Trino for SQL queries
```

---

## 2. Reporting Features

### 2.1 Scheduled Report Generation

**Recommended Libraries:**
- Go: `robfig/cron` (v3)
- JavaScript: `node-cron`

```go
// File: backend/features/scheduled_reports.go

package features

import (
    "fmt"
    "log"
    "sync"
    "time"
    "github.com/robfig/cron/v3"
)

// ScheduledReportManager manages scheduled report generation
type ScheduledReportManager struct {
    cron              *cron.Cron
    reportService     *ReportService
    reportStorage     ReportStorage
    emailService      EmailService
    mu                sync.RWMutex
    activeSchedules   map[string]cron.EntryID
}

type ScheduleFrequency string

const (
    Daily   ScheduleFrequency = "daily"
    Weekly  ScheduleFrequency = "weekly"
    Monthly ScheduleFrequency = "monthly"
)

type ScheduledReport struct {
    ID           string
    AccountID    string
    ReportType   string // "performance", "tax", "drawdown"
    Frequency    ScheduleFrequency
    Recipients   []string // email addresses
    Format       string   // "pdf", "excel", "csv"
    CreatedAt    time.Time
    LastRun      *time.Time
    NextRun      time.Time
    IsActive     bool
    IncludeCharts bool
}

// NewScheduledReportManager creates a new manager
func NewScheduledReportManager(
    reportService *ReportService,
    storage ReportStorage,
    emailService EmailService,
) *ScheduledReportManager {
    return &ScheduledReportManager{
        cron:            cron.New(),
        reportService:   reportService,
        reportStorage:   storage,
        emailService:    emailService,
        activeSchedules: make(map[string]cron.EntryID),
    }
}

// CreateSchedule creates a new scheduled report
func (m *ScheduledReportManager) CreateSchedule(
    schedule *ScheduledReport,
) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // Determine cron expression
    cronExpr := m.frequencyToCron(schedule.Frequency)

    // Create cron job
    entryID, err := m.cron.AddFunc(
        cronExpr,
        func() {
            m.generateAndSend(schedule)
        },
    )
    if err != nil {
        return fmt.Errorf("failed to add cron job: %w", err)
    }

    // Store schedule
    if err := m.reportStorage.SaveSchedule(schedule); err != nil {
        return fmt.Errorf("failed to save schedule: %w", err)
    }

    m.activeSchedules[schedule.ID] = entryID
    log.Printf("[ScheduledReports] Created schedule %s for account %s", schedule.ID, schedule.AccountID)

    return nil
}

// generateAndSend generates and sends report
func (m *ScheduledReportManager) generateAndSend(schedule *ScheduledReport) {
    log.Printf("[ScheduledReports] Generating %s report for %s", schedule.ReportType, schedule.AccountID)

    // Generate report based on type
    var reportData []byte
    var filename string
    var err error

    now := time.Now()
    startDate := m.getStartDate(now, schedule.Frequency)

    switch schedule.ReportType {
    case "performance":
        report, err := m.reportService.GeneratePerformanceReport(
            schedule.AccountID,
            startDate,
            now,
        )
        if err != nil {
            log.Printf("[Error] Failed to generate performance report: %v", err)
            return
        }

        reportData, filename, err = m.formatReport(report, schedule.Format, "performance")

    case "tax":
        report, err := m.reportService.GenerateTaxReport(schedule.AccountID, now.Year())
        if err != nil {
            log.Printf("[Error] Failed to generate tax report: %v", err)
            return
        }

        reportData, filename, err = m.formatReport(report, schedule.Format, "tax")

    case "drawdown":
        // Generate drawdown report
        // Implementation...
    }

    if err != nil {
        log.Printf("[Error] Failed to format report: %v", err)
        return
    }

    // Send to recipients
    for _, recipient := range schedule.Recipients {
        if err := m.emailService.SendReportEmail(
            recipient,
            filename,
            reportData,
            schedule.ReportType,
        ); err != nil {
            log.Printf("[Error] Failed to send report to %s: %v", recipient, err)
        }
    }

    // Update last run
    now = time.Now()
    schedule.LastRun = &now
    m.reportStorage.UpdateSchedule(schedule)

    log.Printf("[ScheduledReports] Report sent to %d recipients", len(schedule.Recipients))
}

// frequencyToCron converts frequency to cron expression
func (m *ScheduledReportManager) frequencyToCron(freq ScheduleFrequency) string {
    switch freq {
    case Daily:
        return "0 9 * * *"           // 9 AM daily
    case Weekly:
        return "0 9 ? * MON"         // 9 AM Monday
    case Monthly:
        return "0 9 1 * *"           // 9 AM 1st of month
    default:
        return "0 9 * * *"
    }
}

// getStartDate calculates report period start
func (m *ScheduledReportManager) getStartDate(now time.Time, freq ScheduleFrequency) time.Time {
    switch freq {
    case Daily:
        return now.AddDate(0, 0, -1)
    case Weekly:
        return now.AddDate(0, 0, -7)
    case Monthly:
        return now.AddDate(0, -1, 0)
    default:
        return now.AddDate(0, 0, -1)
    }
}

// Start starts the cron scheduler
func (m *ScheduledReportManager) Start() {
    m.cron.Start()
    log.Println("[ScheduledReports] Scheduler started")
}

// Stop stops the cron scheduler
func (m *ScheduledReportManager) Stop() {
    m.cron.Stop()
    log.Println("[ScheduledReports] Scheduler stopped")
}
```

### 2.2 Email Delivery Service

```go
// File: backend/features/email_service.go

package features

import (
    "fmt"
    "net/smtp"
    "mime"
    "mime/multipart"
    "strings"
)

// EmailService handles report delivery
type EmailService interface {
    SendReportEmail(
        recipient string,
        filename string,
        data []byte,
        reportType string,
    ) error
    SendNotification(
        recipient string,
        subject string,
        body string,
    ) error
}

// SMTPEmailService implements EmailService
type SMTPEmailService struct {
    host     string
    port     string
    username string
    password string
    from     string
}

// NewSMTPEmailService creates SMTP email service
func NewSMTPEmailService(
    host, port, username, password, from string,
) *SMTPEmailService {
    return &SMTPEmailService{
        host:     host,
        port:     port,
        username: username,
        password: password,
        from:     from,
    }
}

// SendReportEmail sends report as email attachment
func (s *SMTPEmailService) SendReportEmail(
    recipient string,
    filename string,
    data []byte,
    reportType string,
) error {
    subject := fmt.Sprintf("Trading Report - %s", reportType)
    body := fmt.Sprintf(`
Dear Trader,

Please find attached your %s report.

Generated: %s

Best regards,
Trading Engine
`, reportType, time.Now().Format(time.RFC1123))

    // Create multipart message with attachment
    boundary := "boundary123456789"
    message := fmt.Sprintf(
        `From: %s
To: %s
Subject: %s
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary=%s

--%s
Content-Type: text/plain; charset=utf-8

%s

--%s
Content-Type: application/octet-stream
Content-Disposition: attachment; filename="%s"
Content-Transfer-Encoding: base64

%s
--%s--
`,
        s.from,
        recipient,
        subject,
        boundary,
        boundary,
        body,
        boundary,
        filename,
        base64.StdEncoding.EncodeToString(data),
        boundary,
    )

    // Send email
    auth := smtp.PlainAuth("", s.username, s.password, s.host)
    addr := fmt.Sprintf("%s:%s", s.host, s.port)

    if err := smtp.SendMail(addr, auth, s.from, []string{recipient}, []byte(message)); err != nil {
        return fmt.Errorf("failed to send email: %w", err)
    }

    return nil
}

// SendNotification sends a simple notification
func (s *SMTPEmailService) SendNotification(
    recipient string,
    subject string,
    body string,
) error {
    message := fmt.Sprintf(
        `From: %s
To: %s
Subject: %s

%s
`,
        s.from,
        recipient,
        subject,
        body,
    )

    auth := smtp.PlainAuth("", s.username, s.password, s.host)
    addr := fmt.Sprintf("%s:%s", s.host, s.port)

    if err := smtp.SendMail(addr, auth, s.from, []string{recipient}, []byte(message)); err != nil {
        return fmt.Errorf("failed to send notification: %w", err)
    }

    return nil
}
```

### 2.3 Custom Report Templates

```go
// File: backend/features/report_templates.go

package features

import (
    "bytes"
    "fmt"
    "text/template"
    "time"
)

// ReportTemplate represents a customizable report template
type ReportTemplate struct {
    ID            string
    Name          string
    Description   string
    ReportType    string // "performance", "tax", "drawdown"
    TemplateHTML  string // HTML template with {{.Field}} placeholders
    TemplateText  string // Text template
    Sections      []TemplateSection
    CreatedAt     time.Time
    LastModified  time.Time
}

type TemplateSection struct {
    ID       string
    Name     string
    Enabled  bool
    Position int
}

// TemplateRenderer renders templates with data
type TemplateRenderer struct {
    templates map[string]*template.Template
}

// RenderPerformanceReport renders performance report with template
func (r *TemplateRenderer) RenderPerformanceReport(
    templateID string,
    report *PerformanceReport,
    trades []Trade,
) (string, error) {
    tmpl, exists := r.templates[templateID]
    if !exists {
        return "", fmt.Errorf("template not found: %s", templateID)
    }

    // Prepare data
    data := map[string]interface{}{
        "Report":           report,
        "Trades":          trades,
        "GeneratedAt":     time.Now(),
        "TotalProfit":     report.NetProfit,
        "WinRate":         fmt.Sprintf("%.2f%%", report.WinRate),
        "SharpeRatio":     fmt.Sprintf("%.2f", report.SharpeRatio),
        "MaxDrawdown":     fmt.Sprintf("%.2f%%", report.MaxDrawdownPct),
        "ProfitFactor":    fmt.Sprintf("%.2f", report.ProfitFactor),
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", fmt.Errorf("failed to render template: %w", err)
    }

    return buf.String(), nil
}

// Example template HTML
const DefaultPerformanceTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { border-bottom: 2px solid #333; padding-bottom: 10px; }
        .section { margin: 20px 0; }
        table { width: 100%; border-collapse: collapse; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .positive { color: green; }
        .negative { color: red; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Performance Report</h1>
        <p>Generated: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</p>
        <p>Period: {{.Report.StartDate.Format "2006-01-02"}} - {{.Report.EndDate.Format "2006-01-02"}}</p>
    </div>

    <div class="section">
        <h2>Summary Metrics</h2>
        <table>
            <tr>
                <th>Metric</th>
                <th>Value</th>
            </tr>
            <tr>
                <td>Total Trades</td>
                <td>{{.Report.TotalTrades}}</td>
            </tr>
            <tr>
                <td>Win Rate</td>
                <td class="positive">{{.WinRate}}</td>
            </tr>
            <tr>
                <td>Net Profit</td>
                <td class="{{if ge .Report.NetProfit 0}}positive{{else}}negative{{end}}">${{.TotalProfit}}</td>
            </tr>
            <tr>
                <td>Sharpe Ratio</td>
                <td>{{.SharpeRatio}}</td>
            </tr>
            <tr>
                <td>Max Drawdown</td>
                <td class="negative">{{.MaxDrawdown}}</td>
            </tr>
        </table>
    </div>
</body>
</html>
`
```

---

## 3. Data Transformation & Aggregation

### 3.1 Multi-Level Aggregation (Tick → Minute → Hour → Day)

```go
// File: backend/features/data_aggregation.go

package features

import (
    "fmt"
    "time"
)

type AggregationLevel string

const (
    Tick   AggregationLevel = "tick"
    Minute AggregationLevel = "minute"
    Hour   AggregationLevel = "hour"
    Day    AggregationLevel = "day"
)

// AggregatedTrade represents aggregated trade data
type AggregatedTrade struct {
    Symbol        string
    Timestamp     time.Time
    Open          float64
    High          float64
    Low           float64
    Close         float64
    Volume        float64
    Count         int
    Level         AggregationLevel
}

// DataAggregator handles multi-level aggregation
type DataAggregator struct {
    trades []Trade
}

// AggregateByLevel aggregates trades to specified level
func (a *DataAggregator) AggregateByLevel(
    symbol string,
    startDate, endDate time.Time,
    level AggregationLevel,
) ([]AggregatedTrade, error) {

    // Filter trades for symbol
    filtered := make([]Trade, 0)
    for _, t := range a.trades {
        if t.Symbol == symbol &&
            t.CloseTime.After(startDate) &&
            t.CloseTime.Before(endDate) {
            filtered = append(filtered, t)
        }
    }

    switch level {
    case Hour:
        return a.aggregateByHour(symbol, filtered)
    case Day:
        return a.aggregateByDay(symbol, filtered)
    case Minute:
        return a.aggregateByMinute(symbol, filtered)
    default:
        return filtered, nil // Tick level
    }
}

// aggregateByHour aggregates to hourly candles
func (a *DataAggregator) aggregateByHour(
    symbol string,
    trades []Trade,
) ([]AggregatedTrade, error) {
    groups := make(map[time.Time][]Trade)

    for _, trade := range trades {
        // Round time to hour
        hour := trade.CloseTime.Truncate(time.Hour)
        groups[hour] = append(groups[hour], trade)
    }

    result := make([]AggregatedTrade, 0, len(groups))

    for hour, groupTrades := range groups {
        agg := a.createAggregation(symbol, hour, groupTrades, Hour)
        result = append(result, agg)
    }

    return result, nil
}

// aggregateByDay aggregates to daily candles
func (a *DataAggregator) aggregateByDay(
    symbol string,
    trades []Trade,
) ([]AggregatedTrade, error) {
    groups := make(map[string][]Trade)

    for _, trade := range trades {
        day := trade.CloseTime.Format("2006-01-02")
        groups[day] = append(groups[day], trade)
    }

    result := make([]AggregatedTrade, 0, len(groups))

    for dayStr, groupTrades := range groups {
        t, _ := time.Parse("2006-01-02", dayStr)
        agg := a.createAggregation(symbol, t, groupTrades, Day)
        result = append(result, agg)
    }

    return result, nil
}

// createAggregation creates aggregated data
func (a *DataAggregator) createAggregation(
    symbol string,
    timestamp time.Time,
    trades []Trade,
    level AggregationLevel,
) AggregatedTrade {
    agg := AggregatedTrade{
        Symbol:    symbol,
        Timestamp: timestamp,
        Open:      trades[0].OpenPrice,
        Close:     trades[len(trades)-1].ClosePrice,
        High:      trades[0].ClosePrice,
        Low:       trades[0].ClosePrice,
        Volume:    0,
        Count:     len(trades),
        Level:     level,
    }

    for _, t := range trades {
        if t.ClosePrice > agg.High {
            agg.High = t.ClosePrice
        }
        if t.ClosePrice < agg.Low {
            agg.Low = t.ClosePrice
        }
        agg.Volume += t.Volume
    }

    return agg
}
```

### 3.2 Time Zone Handling

```go
// File: backend/features/timezone_handler.go

package features

import (
    "fmt"
    "time"
)

// TimezoneConverter handles time zone conversions
type TimezoneConverter struct {
    defaultTZ string // Server timezone
}

// NewTimezoneConverter creates a new converter
func NewTimezoneConverter(defaultTZ string) *TimezoneConverter {
    return &TimezoneConverter{
        defaultTZ: defaultTZ,
    }
}

// ConvertToUserTimezone converts time to user timezone
func (tc *TimezoneConverter) ConvertToUserTimezone(
    t time.Time,
    userTZ string,
) (time.Time, error) {
    loc, err := time.LoadLocation(userTZ)
    if err != nil {
        return t, fmt.Errorf("invalid timezone: %w", err)
    }

    return t.In(loc), nil
}

// ConvertFromUserTimezone converts from user timezone to UTC
func (tc *TimezoneConverter) ConvertFromUserTimezone(
    t time.Time,
    userTZ string,
) (time.Time, error) {
    loc, err := time.LoadLocation(userTZ)
    if err != nil {
        return t, fmt.Errorf("invalid timezone: %w", err)
    }

    // Assume input time is in user timezone
    return t.In(loc).UTC(), nil
}

// ValidateTimezone validates a timezone string
func (tc *TimezoneConverter) ValidateTimezone(tz string) error {
    _, err := time.LoadLocation(tz)
    return err
}

// GetSupportedTimezones returns list of supported timezones
func (tc *TimezoneConverter) GetSupportedTimezones() []string {
    return []string{
        "UTC",
        "America/New_York",
        "America/Chicago",
        "America/Los_Angeles",
        "Europe/London",
        "Europe/Paris",
        "Europe/Tokyo",
        "Asia/Shanghai",
        "Asia/Hong_Kong",
        "Australia/Sydney",
    }
}
```

### 3.3 Currency Conversion

```go
// File: backend/features/currency_converter.go

package features

import (
    "fmt"
    "sync"
    "time"
)

// ExchangeRateProvider provides current exchange rates
type ExchangeRateProvider interface {
    GetRate(from, to string) (float64, error)
    GetHistoricalRate(from, to string, date time.Time) (float64, error)
}

// CurrencyConverter handles currency conversion
type CurrencyConverter struct {
    provider    ExchangeRateProvider
    cache       map[string]float64
    cacheMutex  sync.RWMutex
    cacheExpiry time.Time
}

// NewCurrencyConverter creates a converter
func NewCurrencyConverter(provider ExchangeRateProvider) *CurrencyConverter {
    return &CurrencyConverter{
        provider: provider,
        cache:    make(map[string]float64),
    }
}

// Convert converts amount from one currency to another
func (cc *CurrencyConverter) Convert(
    amount float64,
    from, to string,
) (float64, error) {
    if from == to {
        return amount, nil
    }

    rate, err := cc.getRate(from, to)
    if err != nil {
        return 0, err
    }

    return amount * rate, nil
}

// ConvertTrade converts all amounts in a trade
func (cc *CurrencyConverter) ConvertTrade(
    trade *Trade,
    targetCurrency string,
) (*Trade, error) {
    if trade.Currency == targetCurrency {
        return trade, nil
    }

    converted := *trade

    rate, err := cc.getRate(trade.Currency, targetCurrency)
    if err != nil {
        return nil, err
    }

    converted.OpenPrice *= rate
    converted.ClosePrice *= rate
    converted.Profit *= rate
    converted.Commission *= rate
    converted.Swap *= rate

    return &converted, nil
}

// getRate gets exchange rate with caching
func (cc *CurrencyConverter) getRate(from, to string) (float64, error) {
    key := fmt.Sprintf("%s_%s", from, to)

    cc.cacheMutex.RLock()
    if rate, ok := cc.cache[key]; ok && time.Now().Before(cc.cacheExpiry) {
        cc.cacheMutex.RUnlock()
        return rate, nil
    }
    cc.cacheMutex.RUnlock()

    rate, err := cc.provider.GetRate(from, to)
    if err != nil {
        return 0, err
    }

    cc.cacheMutex.Lock()
    cc.cache[key] = rate
    cc.cacheExpiry = time.Now().Add(1 * time.Hour)
    cc.cacheMutex.Unlock()

    return rate, nil
}
```

---

## 4. Integration Capabilities

### 4.1 REST API for Export

```go
// File: backend/internal/api/handlers/export.go

package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type ExportHandler struct {
    reportService *features.ReportService
    csvExporter   *features.CSVExporter
    excelExporter *features.ExcelExporter
}

// POST /api/v1/export/trades
func (h *ExportHandler) ExportTrades(w http.ResponseWriter, r *http.Request) {
    accountID := r.Header.Get("X-Account-ID")
    format := r.URL.Query().Get("format")        // csv, json, excel
    startDate := parseDate(r.URL.Query().Get("start"))
    endDate := parseDate(r.URL.Query().Get("end"))

    trades, err := h.reportService.getTrades(accountID, startDate, endDate)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Set response headers
    timestamp := time.Now().Format("20060102_150405")
    filename := fmt.Sprintf("trades_%s.%s", timestamp, format)

    switch format {
    case "csv":
        var buf bytes.Buffer
        if err := h.csvExporter.ExportTradesCSV(&buf, trades, defaultColumns); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "text/csv")
        w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
        w.Write(buf.Bytes())

    case "json":
        data, err := json.MarshalIndent(trades, "", "  ")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
        w.Write(data)

    case "excel":
        exporter := features.NewExcelExporter()
        data, err := exporter.ExportPerformanceReportExcel(report, trades)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
        w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
        w.Write(data)
    }
}

// POST /api/v1/export/schedule
func (h *ExportHandler) CreateScheduledExport(w http.ResponseWriter, r *http.Request) {
    var req struct {
        AccountID  string   `json:"accountId"`
        ReportType string   `json:"reportType"` // "performance", "tax"
        Frequency  string   `json:"frequency"`  // "daily", "weekly", "monthly"
        Format     string   `json:"format"`     // "pdf", "excel", "csv"
        Recipients []string `json:"recipients"` // email addresses
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    schedule := &features.ScheduledReport{
        AccountID:  req.AccountID,
        ReportType: req.ReportType,
        Frequency:  features.ScheduleFrequency(req.Frequency),
        Format:     req.Format,
        Recipients: req.Recipients,
        IsActive:   true,
        CreatedAt:  time.Now(),
    }

    if err := h.scheduledReportManager.CreateSchedule(schedule); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(schedule)
}
```

### 4.2 Webhook Notifications

```go
// File: backend/features/webhooks.go

package features

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

// WebhookEvent represents a reportable event
type WebhookEvent struct {
    EventType  string                 `json:"eventType"`
    Timestamp  time.Time              `json:"timestamp"`
    AccountID  string                 `json:"accountId"`
    Data       map[string]interface{} `json:"data"`
}

// WebhookEndpoint configuration
type WebhookEndpoint struct {
    ID         string
    URL        string
    Secret     string
    Events     []string // ["report:generated", "report:failed"]
    Active     bool
    CreatedAt  time.Time
}

// WebhookManager manages webhook delivery
type WebhookManager struct {
    endpoints map[string]*WebhookEndpoint
    client    *http.Client
}

// SendWebhook sends webhook notification
func (wm *WebhookManager) SendWebhook(
    event *WebhookEvent,
    endpoint *WebhookEndpoint,
) error {
    // Check if endpoint subscribes to this event
    subscribed := false
    for _, e := range endpoint.Events {
        if e == event.EventType {
            subscribed = true
            break
        }
    }
    if !subscribed {
        return nil
    }

    // Marshal event
    payload, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal webhook: %w", err)
    }

    // Create request
    req, err := http.NewRequest("POST", endpoint.URL, bytes.NewReader(payload))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    // Add headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Webhook-Secret", endpoint.Secret)
    req.Header.Set("X-Timestamp", time.Now().Format(time.RFC3339))

    // Send request with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    resp, err := wm.client.Do(req.WithContext(ctx))
    if err != nil {
        return fmt.Errorf("failed to send webhook: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return fmt.Errorf("webhook returned status %d", resp.StatusCode)
    }

    return nil
}

// Events to trigger
const (
    EventReportGenerated = "report:generated"
    EventReportFailed    = "report:failed"
    EventReportScheduled = "report:scheduled"
    EventExportStarted   = "export:started"
    EventExportCompleted = "export:completed"
)
```

### 4.3 S3 Cloud Storage Integration

```go
// File: backend/features/cloud_storage.go

package features

import (
    "bytes"
    "fmt"
    "context"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "time"
)

// CloudStorageProvider interface for different cloud providers
type CloudStorageProvider interface {
    Upload(ctx context.Context, key string, data []byte) error
    Download(ctx context.Context, key string) ([]byte, error)
    Delete(ctx context.Context, key string) error
    ListObjects(ctx context.Context, prefix string) ([]string, error)
}

// S3StorageProvider implements CloudStorageProvider for AWS S3
type S3StorageProvider struct {
    client *s3.Client
    bucket string
}

// NewS3StorageProvider creates S3 storage provider
func NewS3StorageProvider(bucket string) (*S3StorageProvider, error) {
    cfg, err := config.LoadDefaultConfig(context.Background())
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }

    return &S3StorageProvider{
        client: s3.NewFromConfig(cfg),
        bucket: bucket,
    }, nil
}

// Upload uploads file to S3
func (s *S3StorageProvider) Upload(
    ctx context.Context,
    key string,
    data []byte,
) error {
    _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(s.bucket),
        Key:         aws.String(key),
        Body:        bytes.NewReader(data),
        ContentType: aws.String(getMimeType(key)),
    })

    if err != nil {
        return fmt.Errorf("failed to upload to S3: %w", err)
    }

    return nil
}

// Download downloads file from S3
func (s *S3StorageProvider) Download(
    ctx context.Context,
    key string,
) ([]byte, error) {
    result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),
    })

    if err != nil {
        return nil, fmt.Errorf("failed to download from S3: %w", err)
    }
    defer result.Body.Close()

    return io.ReadAll(result.Body)
}

// ListObjects lists objects with prefix
func (s *S3StorageProvider) ListObjects(
    ctx context.Context,
    prefix string,
) ([]string, error) {
    result, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
        Bucket: aws.String(s.bucket),
        Prefix: aws.String(prefix),
    })

    if err != nil {
        return nil, fmt.Errorf("failed to list S3 objects: %w", err)
    }

    keys := make([]string, len(result.Contents))
    for i, obj := range result.Contents {
        keys[i] = *obj.Key
    }

    return keys, nil
}

// StorageBackedReportService combines report generation with cloud storage
type StorageBackedReportService struct {
    reportService *ReportService
    storage       CloudStorageProvider
}

// GenerateAndStore generates report and stores in cloud
func (s *StorageBackedReportService) GenerateAndStore(
    ctx context.Context,
    accountID string,
    reportType string,
    format string,
) (string, error) {
    // Generate report
    report, err := s.reportService.GeneratePerformanceReport(
        accountID,
        time.Now().AddDate(0, -1, 0),
        time.Now(),
    )
    if err != nil {
        return "", err
    }

    // Format report
    var data []byte
    switch format {
    case "json":
        data, _ = json.Marshal(report)
    case "excel":
        exporter := NewExcelExporter()
        trades, _ := s.reportService.getTrades(accountID, report.StartDate, report.EndDate)
        data, _ = exporter.ExportPerformanceReportExcel(report, trades)
    }

    // Store in cloud
    timestamp := time.Now().Format("20060102_150405")
    key := fmt.Sprintf("reports/%s/%s/%s.%s", accountID, reportType, timestamp, format)

    if err := s.storage.Upload(ctx, key, data); err != nil {
        return "", err
    }

    return key, nil
}

func getMimeType(filename string) string {
    switch {
    case ends(filename, ".csv"):
        return "text/csv"
    case ends(filename, ".json"):
        return "application/json"
    case ends(filename, ".xlsx"):
        return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
    case ends(filename, ".pdf"):
        return "application/pdf"
    default:
        return "application/octet-stream"
    }
}

func ends(s, suffix string) bool {
    return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
```

### 4.4 SFTP Upload for Regulatory Submissions

```go
// File: backend/features/sftp_uploader.go

package features

import (
    "fmt"
    "github.com/pkg/sftp"
    "golang.org/x/crypto/ssh"
    "io"
)

// SFTPUploader handles SFTP file uploads
type SFTPUploader struct {
    host     string
    port     string
    username string
    password string
    path     string
}

// NewSFTPUploader creates SFTP uploader
func NewSFTPUploader(
    host, port, username, password, path string,
) *SFTPUploader {
    return &SFTPUploader{
        host:     host,
        port:     port,
        username: username,
        password: password,
        path:     path,
    }
}

// UploadRegulatory uploads regulatory report to SFTP
func (su *SFTPUploader) UploadRegulatory(
    filename string,
    data []byte,
    jurisdiction string,
) error {
    // Create SSH config
    config := &ssh.ClientConfig{
        User: su.username,
        Auth: []ssh.AuthMethod{
            ssh.Password(su.password),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }

    // Connect to SFTP server
    addr := fmt.Sprintf("%s:%s", su.host, su.port)
    conn, err := ssh.Dial("tcp", addr, config)
    if err != nil {
        return fmt.Errorf("failed to connect to SFTP: %w", err)
    }
    defer conn.Close()

    // Create SFTP client
    client, err := sftp.NewClient(conn)
    if err != nil {
        return fmt.Errorf("failed to create SFTP client: %w", err)
    }
    defer client.Close()

    // Create jurisdictional folder
    folderPath := fmt.Sprintf("%s/%s", su.path, jurisdiction)
    client.Mkdir(folderPath)

    // Upload file
    filePath := fmt.Sprintf("%s/%s", folderPath, filename)
    f, err := client.Create(filePath)
    if err != nil {
        return fmt.Errorf("failed to create remote file: %w", err)
    }
    defer f.Close()

    _, err = io.Copy(f, bytes.NewReader(data))
    if err != nil {
        return fmt.Errorf("failed to upload file: %w", err)
    }

    return nil
}
```

---

## 5. Audit Trails & Compliance

### 5.1 Export Audit Log

```go
// File: backend/features/export_audit.go

package features

import (
    "fmt"
    "log"
    "sync"
    "time"
)

// ExportAuditLog records all export activities
type ExportAuditLog struct {
    mu      sync.RWMutex
    entries []AuditEntry
}

// AuditEntry represents a single audit log entry
type AuditEntry struct {
    ID            string
    Timestamp     time.Time
    AccountID     string
    UserID        string
    Action        string // "export", "report_generated", "email_sent"
    ResourceType  string // "trades", "performance_report", "tax_report"
    Format        string // "csv", "json", "excel", "pdf"
    RecordCount   int
    IPAddress     string
    UserAgent     string
    Status        string // "success", "failed"
    ErrorMessage  string
    Duration      time.Duration
    DataHash      string // SHA256 hash of exported data
    RetentionDays int
    ExpiresAt     time.Time
}

// RecordExport records an export action
func (log *ExportAuditLog) RecordExport(
    accountID, userID, action, resourceType, format string,
    recordCount int,
    ipAddress, userAgent string,
) string {
    entry := AuditEntry{
        ID:           generateID(),
        Timestamp:    time.Now(),
        AccountID:    accountID,
        UserID:       userID,
        Action:       action,
        ResourceType: resourceType,
        Format:       format,
        RecordCount:  recordCount,
        IPAddress:    ipAddress,
        UserAgent:    userAgent,
        Status:       "success",
        RetentionDays: 365, // 1 year retention
        ExpiresAt:    time.Now().AddDate(1, 0, 0),
    }

    log.mu.Lock()
    defer log.mu.Unlock()

    log.entries = append(log.entries, entry)

    return entry.ID
}

// GetAuditTrail retrieves audit entries
func (log *ExportAuditLog) GetAuditTrail(
    accountID string,
    startDate, endDate time.Time,
) []AuditEntry {
    log.mu.RLock()
    defer log.mu.RUnlock()

    var result []AuditEntry
    for _, entry := range log.entries {
        if entry.AccountID == accountID &&
            entry.Timestamp.After(startDate) &&
            entry.Timestamp.Before(endDate) {
            result = append(result, entry)
        }
    }

    return result
}

// ComplianceReport generates compliance report
type ComplianceReport struct {
    Period        DateRange
    TotalExports  int
    ExportsByType map[string]int
    ExportsByUser map[string]int
    FailedExports int
    AverageSize   float64
}

// GenerateComplianceReport generates GDPR compliance report
func (log *ExportAuditLog) GenerateComplianceReport(
    startDate, endDate time.Time,
) *ComplianceReport {
    log.mu.RLock()
    defer log.mu.RUnlock()

    report := &ComplianceReport{
        Period:        DateRange{Start: startDate, End: endDate},
        ExportsByType: make(map[string]int),
        ExportsByUser: make(map[string]int),
    }

    totalSize := 0

    for _, entry := range log.entries {
        if entry.Timestamp.After(startDate) && entry.Timestamp.Before(endDate) {
            report.TotalExports++
            report.ExportsByType[entry.Format]++
            report.ExportsByUser[entry.UserID]++

            if entry.Status == "failed" {
                report.FailedExports++
            }

            totalSize += entry.RecordCount
        }
    }

    if report.TotalExports > 0 {
        report.AverageSize = float64(totalSize) / float64(report.TotalExports)
    }

    return report
}

// EnforceDataRetention enforces data retention policy
func (log *ExportAuditLog) EnforceDataRetention() {
    log.mu.Lock()
    defer log.mu.Unlock()

    now := time.Now()
    preserved := make([]AuditEntry, 0)

    for _, entry := range log.entries {
        if now.Before(entry.ExpiresAt) {
            preserved = append(preserved, entry)
        } else {
            log.LogRetention(entry)
        }
    }

    log.entries = preserved
}

// LogRetention logs deleted entries (for compliance)
func (log *ExportAuditLog) LogRetention(entry AuditEntry) {
    log.Printf("[RETENTION] Deleted audit entry %s (expired %s)",
        entry.ID, entry.ExpiresAt)
}

// GDPR Right to Erasure
func (log *ExportAuditLog) EraseUserData(userID string) error {
    log.mu.Lock()
    defer log.mu.Unlock()

    preserved := make([]AuditEntry, 0)

    for _, entry := range log.entries {
        if entry.UserID != userID {
            preserved = append(preserved, entry)
        } else {
            log.Printf("[GDPR] Erased audit entry for user %s", userID)
        }
    }

    log.entries = preserved
    return nil
}
```

### 5.2 GDPR Compliance Export

```go
// File: backend/features/gdpr_export.go

package features

import (
    "encoding/json"
    "fmt"
    "time"
)

// GDPRDataExport represents user's personal data export
type GDPRDataExport struct {
    RequestID     string          `json:"requestId"`
    UserID        string          `json:"userId"`
    RequestedAt   time.Time       `json:"requestedAt"`
    ExpiresAt     time.Time       `json:"expiresAt"`
    PersonalData  PersonalData    `json:"personalData"`
    Trades        []Trade         `json:"trades"`
    Reports       []interface{}   `json:"reports"`
    AuditLog      []AuditEntry    `json:"auditLog"`
    DownloadToken string          `json:"downloadToken"`
}

type PersonalData struct {
    AccountID    string    `json:"accountId"`
    Email        string    `json:"email"`
    Country      string    `json:"country"`
    CreatedAt    time.Time `json:"createdAt"`
    LastLogin    time.Time `json:"lastLogin"`
    Preferences  map[string]string `json:"preferences"`
}

// GDPRExporter handles GDPR requests
type GDPRExporter struct {
    reportService *ReportService
    auditLog      *ExportAuditLog
}

// ExportUserData generates GDPR compliant export
func (ge *GDPRExporter) ExportUserData(
    userID string,
    accountID string,
) (*GDPRDataExport, error) {
    export := &GDPRDataExport{
        RequestID:   generateID(),
        UserID:      userID,
        RequestedAt: time.Now(),
        ExpiresAt:   time.Now().AddDate(0, 0, 30), // 30 day download window
    }

    // Get all user trades
    trades, err := ge.reportService.getTrades(
        accountID,
        time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
        time.Now(),
    )
    if err != nil {
        return nil, err
    }
    export.Trades = trades

    // Get audit trail
    export.AuditLog = ge.auditLog.GetAuditTrail(
        accountID,
        time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
        time.Now(),
    )

    // Generate download token
    export.DownloadToken = generateSecureToken()

    return export, nil
}

// EncryptGDPRExport encrypts sensitive export data
func (ge *GDPRExporter) EncryptGDPRExport(
    export *GDPRDataExport,
    publicKey string,
) ([]byte, error) {
    // Implementation using RSA encryption
    // Allows user to decrypt with private key
    return nil, nil
}

// Consent tracking
type ConsentRecord struct {
    UserID     string
    ConsentType string // "data_processing", "marketing", "analytics"
    Given      bool
    Timestamp  time.Time
    ExpiresAt  time.Time
}

type ConsentManager struct {
    consents map[string][]ConsentRecord
}

// CheckConsent checks if user has given consent
func (cm *ConsentManager) CheckConsent(
    userID, consentType string,
) bool {
    records, exists := cm.consents[userID]
    if !exists {
        return false
    }

    for _, record := range records {
        if record.ConsentType == consentType &&
            record.Given &&
            time.Now().Before(record.ExpiresAt) {
            return true
        }
    }

    return false
}
```

---

## 6. Library Recommendations Summary

### Backend (Go)

| Category | Library | Use Case | Notes |
|----------|---------|----------|-------|
| **CSV** | `encoding/csv` (stdlib) | CSV export | Built-in, no dependencies |
| **Excel** | `github.com/xuri/excelize` | Excel generation | Feature-rich, supports charts |
| **PDF** | `github.com/go-pdf/fpdf` | PDF reports | Lightweight, supports text/images |
| **Scheduling** | `github.com/robfig/cron` | Scheduled reports | v3 with context support |
| **Email** | `net/smtp` (stdlib) | Email delivery | Built-in, supports MIME |
| **Cloud Storage** | `github.com/aws/aws-sdk-go-v2` | S3 integration | Official AWS SDK v2 |
| **SFTP** | `github.com/pkg/sftp` | Regulatory uploads | Standard SFTP client |
| **Parquet** | `github.com/xitongsys/parquet-go` | Big data export | For Spark/DuckDB integration |

### Frontend (TypeScript/React)

| Category | Library | Use Case | Notes |
|----------|---------|----------|-------|
| **CSV** | `papaparse` | CSV parsing/generation | Client-side CSV handling |
| **Excel** | `exceljs` | Excel generation | Comprehensive Excel support |
| **PDF** | `jspdf` + `html2canvas` | PDF export with charts | Client-side PDF generation |
| **Charts** | `recharts` | Chart rendering | React-native charts |
| **Date/Time** | `date-fns` | Timezone handling | Lightweight date utilities |
| **Validation** | `zod` | Export options validation | TypeScript-first validation |

---

## 7. Implementation Roadmap

### Phase 1: Core Export (2-3 weeks)
1. CSV/JSON export endpoints
2. Basic Excel generation
3. Export audit logging
4. Frontend export UI

### Phase 2: Advanced Features (3-4 weeks)
1. PDF generation with charts
2. Scheduled report generation
3. Email delivery integration
4. Custom report templates

### Phase 3: Enterprise Integration (3-4 weeks)
1. S3 cloud storage integration
2. SFTP regulatory uploads
3. Webhook notifications
4. Multi-currency support

### Phase 4: Compliance (2-3 weeks)
1. GDPR data export
2. Timezone handling
3. Data retention policies
4. Audit trail dashboard

---

## 8. Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                   Trading Analytics Dashboard               │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
    ┌─────────┐          ┌─────────┐          ┌─────────┐
    │ Frontend│          │ Backend │          │Database │
    │ Export  │          │ Services│          │ /Cache  │
    └─────────┘          └─────────┘          └─────────┘
        │                     │                     │
    ┌─────────┐          ┌─────────────────┐       │
    │CSV/JSON │          │ReportService    │       │
    │  Export │          │ExcelExporter    │───────┼─ Trades
    │ UI      │          │PDFExporter      │       │
    │Component│          │ParquetExporter  │───────┼─ Reports
    └─────────┘          └─────────────────┘       │
        │                     │                 ┌───┘
        │              ┌──────┴─────┐           │
        │              │            │           │
    ┌───┴────────┐ ┌────────┐ ┌──────────┐ ┌──────┐
    │  Download  │ │Scheduler│ │Email     │ │Audit │
    │  Handler   │ │(cron)   │ │Service   │ │ Log  │
    └────────────┘ └────────┘ └──────────┘ └──────┘
         │              │            │          │
    ┌────┴──────────────┼────────────┼──────────┘
    │                   │            │
┌───┴─────┐        ┌────┴────┐  ┌───┴──────┐
│Cloud     │        │Webhooks │  │SMTP      │
│Storage   │        │Events   │  │Email     │
│(S3)      │        │SFTP     │  │Server    │
└──────────┘        └─────────┘  └──────────┘
```

---

## 9. Key Considerations

### Security
- Hash exported data for tamper detection
- Implement role-based access for exports
- Encrypt sensitive data in transit and at rest
- Rate limit export endpoints
- Validate file formats before processing

### Performance
- Use streaming for large exports (>100MB)
- Implement pagination for trade lists
- Cache exchange rates for currency conversion
- Use goroutines for parallel processing
- Compress files before sending (gzip)

### Scalability
- Use message queues (Redis) for async report generation
- Implement distributed storage (S3) for large files
- Use connection pooling for email/SFTP
- Cache frequently accessed reports
- Monitor resource usage on scheduled jobs

### Compliance
- Maintain detailed audit logs (immutable)
- Implement data retention policies (GDPR)
- Support right to erasure
- Validate regulatory submission formats
- Encrypt PII in exports

---

## 10. Code Examples

### Quick Start: CSV Export Handler

```go
// File: backend/cmd/server/main.go (integration example)

func setupExportHandlers(router *http.ServeMux, services *Services) {
    exportHandler := &handlers.ExportHandler{
        reportService:    services.ReportService,
        csvExporter:      features.NewCSVExporter(),
        excelExporter:    features.NewExcelExporter(),
        pdfExporter:      features.NewPDFExporter(),
    }

    router.HandleFunc("POST /api/v1/export/trades", exportHandler.ExportTrades)
    router.HandleFunc("POST /api/v1/export/report", exportHandler.ExportReport)
    router.HandleFunc("POST /api/v1/export/schedule", exportHandler.CreateScheduledExport)
}
```

---

## Conclusion

This research provides a comprehensive roadmap for implementing enterprise-grade export and reporting capabilities. The recommended libraries are battle-tested, actively maintained, and widely used in production systems. Implementation should follow the phased approach, starting with core export functionality before adding advanced features and compliance requirements.

Key success factors:
1. Strong audit logging from day one
2. Security-first approach to data exports
3. Async processing for large reports
4. Clear separation between scheduled/on-demand reports
5. Support for multiple output formats
6. Regulatory compliance built-in (not added later)
