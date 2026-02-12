package reports

import (
	"log"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
	"github.com/epic1st/rtx/backend/notifications"
)

// ReportScheduler manages automated report generation
type ReportScheduler struct {
	generator    *ReportGenerator
	store        ReportStore
	emailService notifications.EmailService
	adminEmail   string // Email address for report delivery

	mu       sync.RWMutex
	running  bool
	stopChan chan struct{}
}

// NewReportScheduler creates a new report scheduler
func NewReportScheduler(engine *core.Engine, store ReportStore, emailService notifications.EmailService, adminEmail string) *ReportScheduler {
	return &ReportScheduler{
		generator:    NewReportGenerator(engine),
		store:        store,
		emailService: emailService,
		adminEmail:   adminEmail,
		running:      false,
		stopChan:     make(chan struct{}),
	}
}

// Start begins the scheduled report generation
func (s *ReportScheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		log.Println("[ReportScheduler] Already running")
		return
	}
	s.running = true
	s.mu.Unlock()

	log.Println("[ReportScheduler] Starting automated report scheduler")

	// Start goroutine for scheduled checks
	go s.schedulerLoop()
}

// Stop stops the scheduler gracefully
func (s *ReportScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		log.Println("[ReportScheduler] Not running")
		return
	}

	log.Println("[ReportScheduler] Stopping scheduler...")
	s.running = false
	close(s.stopChan)
}

// schedulerLoop is the main loop that checks time and triggers reports
func (s *ReportScheduler) schedulerLoop() {
	// Check every minute for scheduled times
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	lastDailyRun := time.Time{}
	lastWeeklyRun := time.Time{}

	for {
		select {
		case <-s.stopChan:
			log.Println("[ReportScheduler] Scheduler stopped")
			return

		case now := <-ticker.C:
			// Check if it's time for daily reports (02:00 UTC)
			if s.shouldRunDailyReports(now, lastDailyRun) {
				log.Printf("[ReportScheduler] Triggering daily reports at %s", now.Format("2006-01-02 15:04:05 MST"))
				lastDailyRun = now

				// Run in separate goroutine (non-blocking)
				go s.runDailyReports(now)
			}

			// Check if it's time for weekly reports (Monday 02:00 UTC)
			if s.shouldRunWeeklyReports(now, lastWeeklyRun) {
				log.Printf("[ReportScheduler] Triggering weekly reports at %s", now.Format("2006-01-02 15:04:05 MST"))
				lastWeeklyRun = now

				// Run in separate goroutine (non-blocking)
				go s.runWeeklyReports(now)
			}
		}
	}
}

// shouldRunDailyReports checks if it's time to run daily reports (02:00 UTC)
func (s *ReportScheduler) shouldRunDailyReports(now time.Time, lastRun time.Time) bool {
	nowUTC := now.UTC()

	// Check if it's 02:00 UTC (hour 2, minute 0-59)
	if nowUTC.Hour() != 2 {
		return false
	}

	// Prevent running multiple times in the same hour
	if !lastRun.IsZero() {
		// If last run was today at 02:00, skip
		if lastRun.UTC().Year() == nowUTC.Year() &&
			lastRun.UTC().YearDay() == nowUTC.YearDay() &&
			lastRun.UTC().Hour() == 2 {
			return false
		}
	}

	return true
}

// shouldRunWeeklyReports checks if it's time to run weekly reports (Monday 02:00 UTC)
func (s *ReportScheduler) shouldRunWeeklyReports(now time.Time, lastRun time.Time) bool {
	nowUTC := now.UTC()

	// Check if it's Monday
	if nowUTC.Weekday() != time.Monday {
		return false
	}

	// Check if it's 02:00 UTC
	if nowUTC.Hour() != 2 {
		return false
	}

	// Prevent running multiple times in the same week
	if !lastRun.IsZero() {
		year1, week1 := lastRun.UTC().ISOWeek()
		year2, week2 := nowUTC.ISOWeek()

		if year1 == year2 && week1 == week2 {
			return false
		}
	}

	return true
}

// runDailyReports generates and sends daily reports
func (s *ReportScheduler) runDailyReports(now time.Time) {
	// Generate reports for yesterday (reports are generated for the completed day)
	yesterday := now.AddDate(0, 0, -1)

	log.Printf("[ReportScheduler] Generating daily trading summary for %s", yesterday.Format("2006-01-02"))

	// Generate daily trading summary
	tradingReport := s.generator.DailyTradingSummary(yesterday)
	if err := s.store.SaveReport(tradingReport); err != nil {
		log.Printf("[ReportScheduler] Failed to save daily trading report: %v", err)
	} else {
		log.Printf("[ReportScheduler] Daily trading report saved: %s", tradingReport.ID)
		// Send email notification
		go s.sendReportEmail(tradingReport)
	}

	log.Printf("[ReportScheduler] Generating daily risk report for %s", yesterday.Format("2006-01-02"))

	// Generate daily risk report
	riskReport := s.generator.DailyRiskReport(yesterday)
	if err := s.store.SaveReport(riskReport); err != nil {
		log.Printf("[ReportScheduler] Failed to save daily risk report: %v", err)
	} else {
		log.Printf("[ReportScheduler] Daily risk report saved: %s", riskReport.ID)
		// Send email notification
		go s.sendReportEmail(riskReport)
	}
}

// runWeeklyReports generates and sends weekly reports
func (s *ReportScheduler) runWeeklyReports(now time.Time) {
	// Generate report for last week (Monday-Sunday)
	// If today is Monday, last week ended yesterday (Sunday)
	endDate := now.AddDate(0, 0, -1) // Yesterday (Sunday)
	startDate := endDate.AddDate(0, 0, -6) // 6 days before (Monday)

	log.Printf("[ReportScheduler] Generating weekly revenue report for %s to %s",
		startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Generate weekly revenue report
	revenueReport := s.generator.WeeklyRevenueReport(startDate, endDate)
	if err := s.store.SaveReport(revenueReport); err != nil {
		log.Printf("[ReportScheduler] Failed to save weekly revenue report: %v", err)
	} else {
		log.Printf("[ReportScheduler] Weekly revenue report saved: %s", revenueReport.ID)
		// Send email notification
		go s.sendReportEmail(revenueReport)
	}
}

// sendReportEmail sends an email notification when a report is generated
func (s *ReportScheduler) sendReportEmail(report *Report) {
	if s.emailService == nil {
		log.Println("[ReportScheduler] Email service not configured - skipping email")
		return
	}

	if s.adminEmail == "" {
		log.Println("[ReportScheduler] Admin email not configured - skipping email")
		return
	}

	// Render email template
	templateData := map[string]interface{}{
		"reportTitle":  report.Title,
		"reportType":   report.Type,
		"reportPeriod": report.Period,
		"reportId":     report.ID,
		"generatedAt":  report.GeneratedAt.Format("2006-01-02 15:04:05 MST"),
	}

	// Add report-specific data
	switch report.Type {
	case "daily_trading":
		if totalTrades, ok := report.Data["totalTrades"].(int); ok {
			templateData["totalTrades"] = totalTrades
		}
		if totalVolume, ok := report.Data["totalVolume"].(float64); ok {
			templateData["totalVolume"] = totalVolume
		}
		if totalPnL, ok := report.Data["totalPnL"].(float64); ok {
			templateData["totalPnL"] = totalPnL
		}

	case "daily_risk":
		if avgMarginUtilization, ok := report.Data["avgMarginUtilization"].(float64); ok {
			templateData["avgMarginUtilization"] = avgMarginUtilization
		}
		if openPositions, ok := report.Data["openPositions"].(int); ok {
			templateData["openPositions"] = openPositions
		}

	case "weekly_revenue":
		if totalRevenue, ok := report.Data["totalRevenue"].(float64); ok {
			templateData["totalRevenue"] = totalRevenue
		}
		if spreadRevenue, ok := report.Data["spreadRevenue"].(float64); ok {
			templateData["spreadRevenue"] = spreadRevenue
		}
		if commissionRevenue, ok := report.Data["commissionRevenue"].(float64); ok {
			templateData["commissionRevenue"] = commissionRevenue
		}
	}

	// Render HTML body using template
	htmlBody, err := notifications.RenderTemplate("report_generated", templateData)
	if err != nil {
		log.Printf("[ReportScheduler] Failed to render email template: %v", err)
		// Fallback to plain text
		htmlBody = "A new report has been generated: " + report.Title
	}

	subject := "RTX5 Automated Report: " + report.Title

	if err := s.emailService.SendEmail(s.adminEmail, subject, htmlBody); err != nil {
		log.Printf("[ReportScheduler] Failed to send report email: %v", err)
	} else {
		log.Printf("[ReportScheduler] Report email sent to %s for report %s", s.adminEmail, report.ID)
	}
}

// GenerateNow manually triggers report generation (for testing/manual generation)
func (s *ReportScheduler) GenerateNow(reportType string, date time.Time) (*Report, error) {
	var report *Report

	switch reportType {
	case "daily_trading":
		report = s.generator.DailyTradingSummary(date)

	case "daily_risk":
		report = s.generator.DailyRiskReport(date)

	case "weekly_revenue":
		// For weekly, date is the end date - calculate start date (7 days prior)
		startDate := date.AddDate(0, 0, -6)
		report = s.generator.WeeklyRevenueReport(startDate, date)

	default:
		log.Printf("[ReportScheduler] Unknown report type: %s", reportType)
		return nil, nil
	}

	if report != nil {
		if err := s.store.SaveReport(report); err != nil {
			log.Printf("[ReportScheduler] Failed to save manually generated report: %v", err)
			return nil, err
		}
		log.Printf("[ReportScheduler] Manually generated report saved: %s", report.ID)

		// Send email notification
		go s.sendReportEmail(report)
	}

	return report, nil
}
