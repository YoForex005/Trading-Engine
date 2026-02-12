package admin

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ReportType represents the type of report
type ReportType string

const (
	ReportDailyPnL          ReportType = "daily_pnl"
	ReportWeeklySummary     ReportType = "weekly_summary"
	ReportMonthlyStatement  ReportType = "monthly_statement"
	ReportClientActivity    ReportType = "client_activity"
	ReportTradeVolume       ReportType = "trade_volume"
	ReportRiskExposure      ReportType = "risk_exposure"
	ReportCommission        ReportType = "commission_report"
	ReportDepositWithdrawal ReportType = "deposit_withdrawal"
	ReportComplianceAudit   ReportType = "compliance_audit"
	ReportCustom            ReportType = "custom"
)

// ReportFormat represents the output format
type ReportFormat string

const (
	FormatPDF  ReportFormat = "pdf"
	FormatCSV  ReportFormat = "csv"
	FormatXLSX ReportFormat = "xlsx"
	FormatJSON ReportFormat = "json"
)

// ReportSchedule represents the scheduling frequency
type ReportSchedule string

const (
	ScheduleDaily     ReportSchedule = "daily"
	ScheduleWeekly    ReportSchedule = "weekly"
	ScheduleMonthly   ReportSchedule = "monthly"
	ScheduleQuarterly ReportSchedule = "quarterly"
	ScheduleCustom    ReportSchedule = "custom_cron"
)

// ExecutionStatus represents the status of a report execution
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusCancelled ExecutionStatus = "cancelled"
)

// ScheduledReport represents a scheduled report configuration
type ScheduledReport struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        ReportType             `json:"type"`
	Format      ReportFormat           `json:"format"`
	Schedule    ReportSchedule         `json:"schedule"`
	CronExpr    string                 `json:"cron_expression,omitempty"`
	Recipients  []string               `json:"recipients"`
	Filters     map[string]interface{} `json:"filters"`
	LastRun     *time.Time             `json:"last_run,omitempty"`
	NextRun     time.Time              `json:"next_run"`
	Status      string                 `json:"status"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	IsActive    bool                   `json:"is_active"`
	Description string                 `json:"description"`
}

// ReportExecution represents a single execution of a scheduled report
type ReportExecution struct {
	ID          string          `json:"id"`
	ReportID    string          `json:"report_id"`
	ReportName  string          `json:"report_name"`
	StartTime   time.Time       `json:"start_time"`
	EndTime     *time.Time      `json:"end_time,omitempty"`
	Status      ExecutionStatus `json:"status"`
	FileSize    int64           `json:"file_size,omitempty"`
	DownloadURL string          `json:"download_url,omitempty"`
	Error       string          `json:"error,omitempty"`
	Duration    int             `json:"duration_seconds"`
	TriggeredBy string          `json:"triggered_by"`
}

// ScheduledReportTemplate represents a report template configuration
type ScheduledReportTemplate struct {
	Type              ReportType             `json:"type"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	AvailableFormats  []ReportFormat         `json:"available_formats"`
	AvailableFilters  []FilterDefinition     `json:"available_filters"`
	DefaultSchedule   ReportSchedule         `json:"default_schedule"`
	SampleConfig      map[string]interface{} `json:"sample_config"`
}

// FilterDefinition represents a configurable filter for a report
type FilterDefinition struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Default     string   `json:"default,omitempty"`
	Options     []string `json:"options,omitempty"`
	Description string   `json:"description"`
}

// ScheduledReportsStore manages scheduled reports and executions
type ScheduledReportsStore struct {
	mu         sync.RWMutex
	reports    []ScheduledReport
	executions []ReportExecution
	templates  []ScheduledReportTemplate
}

// NewScheduledReportsStore creates a new scheduled reports store with mock data
func NewScheduledReportsStore() *ScheduledReportsStore {
	store := &ScheduledReportsStore{
		templates: generateReportTemplates(),
	}
	store.reports = generateMockReports()
	store.executions = generateMockExecutions(store.reports)
	return store
}

// generateReportTemplates creates 5 report templates
func generateReportTemplates() []ScheduledReportTemplate {
	return []ScheduledReportTemplate{
		{
			Type:        ReportDailyPnL,
			Name:        "Daily P&L Report",
			Description: "Daily profit and loss summary for all accounts",
			AvailableFormats: []ReportFormat{FormatPDF, FormatCSV, FormatXLSX},
			AvailableFilters: []FilterDefinition{
				{Name: "account_type", Type: "select", Required: false, Options: []string{"all", "demo", "live"}, Description: "Filter by account type"},
				{Name: "min_pnl", Type: "number", Required: false, Description: "Minimum P&L threshold"},
			},
			DefaultSchedule: ScheduleDaily,
			SampleConfig: map[string]interface{}{
				"include_charts": true,
				"group_by":       "account",
			},
		},
		{
			Type:        ReportWeeklySummary,
			Name:        "Weekly Summary Report",
			Description: "Weekly trading activity and performance summary",
			AvailableFormats: []ReportFormat{FormatPDF, FormatXLSX},
			AvailableFilters: []FilterDefinition{
				{Name: "date_range", Type: "daterange", Required: true, Description: "Week to report"},
			},
			DefaultSchedule: ScheduleWeekly,
			SampleConfig: map[string]interface{}{
				"include_top_traders": true,
				"include_volume":      true,
			},
		},
		{
			Type:        ReportMonthlyStatement,
			Name:        "Monthly Account Statement",
			Description: "Comprehensive monthly statement for client accounts",
			AvailableFormats: []ReportFormat{FormatPDF, FormatCSV},
			AvailableFilters: []FilterDefinition{
				{Name: "account_id", Type: "text", Required: false, Description: "Specific account ID"},
				{Name: "include_trades", Type: "boolean", Required: false, Default: "true", Description: "Include trade details"},
			},
			DefaultSchedule: ScheduleMonthly,
			SampleConfig: map[string]interface{}{
				"detailed_trades": true,
				"include_fees":    true,
			},
		},
		{
			Type:        ReportRiskExposure,
			Name:        "Risk Exposure Report",
			Description: "Current risk exposure across all positions",
			AvailableFormats: []ReportFormat{FormatPDF, FormatCSV, FormatXLSX, FormatJSON},
			AvailableFilters: []FilterDefinition{
				{Name: "symbol_group", Type: "select", Required: false, Options: []string{"all", "forex", "metals", "crypto"}, Description: "Filter by symbol group"},
				{Name: "risk_threshold", Type: "number", Required: false, Description: "Show only risks above threshold"},
			},
			DefaultSchedule: ScheduleDaily,
			SampleConfig: map[string]interface{}{
				"include_var":          true,
				"include_concentration": true,
			},
		},
		{
			Type:        ReportComplianceAudit,
			Name:        "Compliance Audit Report",
			Description: "Regulatory compliance and audit trail report",
			AvailableFormats: []ReportFormat{FormatPDF, FormatXLSX},
			AvailableFilters: []FilterDefinition{
				{Name: "audit_type", Type: "select", Required: true, Options: []string{"kyc", "transactions", "admin_actions"}, Description: "Type of audit"},
				{Name: "date_range", Type: "daterange", Required: true, Description: "Audit period"},
			},
			DefaultSchedule: ScheduleMonthly,
			SampleConfig: map[string]interface{}{
				"include_violations": true,
				"detailed_logs":      true,
			},
		},
	}
}

// generateMockReports creates 20 scheduled reports
func generateMockReports() []ScheduledReport {
	reports := make([]ScheduledReport, 0, 20)
	rand.Seed(time.Now().UnixNano())

	reportTypes := []ReportType{
		ReportDailyPnL, ReportWeeklySummary, ReportMonthlyStatement,
		ReportClientActivity, ReportTradeVolume, ReportRiskExposure,
		ReportCommission, ReportDepositWithdrawal, ReportComplianceAudit,
	}

	formats := []ReportFormat{FormatPDF, FormatCSV, FormatXLSX, FormatJSON}
	schedules := []ReportSchedule{ScheduleDaily, ScheduleWeekly, ScheduleMonthly, ScheduleQuarterly}

	creators := []string{"admin@rtx5.com", "manager@rtx5.com", "compliance@rtx5.com", "finance@rtx5.com"}

	for i := 0; i < 20; i++ {
		reportType := reportTypes[i%len(reportTypes)]
		format := formats[rand.Intn(len(formats))]
		schedule := schedules[rand.Intn(len(schedules))]

		daysAgo := rand.Intn(90)
		createdAt := time.Now().Add(-time.Duration(daysAgo) * 24 * time.Hour)

		var lastRun *time.Time
		if daysAgo > 1 {
			lr := createdAt.Add(time.Duration(rand.Intn(daysAgo)) * 24 * time.Hour)
			lastRun = &lr
		}

		nextRun := calculateNextRun(schedule, lastRun)

		recipients := []string{
			fmt.Sprintf("user%d@example.com", rand.Intn(10)+1),
			fmt.Sprintf("manager%d@example.com", rand.Intn(5)+1),
		}

		filters := make(map[string]interface{})
		if reportType == ReportDailyPnL {
			filters["account_type"] = []string{"all"}[rand.Intn(1)]
		} else if reportType == ReportRiskExposure {
			filters["symbol_group"] = []string{"forex", "metals", "crypto"}[rand.Intn(3)]
		}

		isActive := rand.Float64() > 0.1

		status := "active"
		if !isActive {
			status = "inactive"
		}

		report := ScheduledReport{
			ID:          fmt.Sprintf("RPT%d", 1000+i),
			Name:        fmt.Sprintf("%s Report %d", reportType, i+1),
			Type:        reportType,
			Format:      format,
			Schedule:    schedule,
			Recipients:  recipients,
			Filters:     filters,
			LastRun:     lastRun,
			NextRun:     nextRun,
			Status:      status,
			CreatedBy:   creators[rand.Intn(len(creators))],
			CreatedAt:   createdAt,
			IsActive:    isActive,
			Description: fmt.Sprintf("Automated %s report in %s format", reportType, format),
		}

		if schedule == ScheduleCustom {
			report.CronExpr = "0 9 * * 1-5"
		}

		reports = append(reports, report)
	}

	return reports
}

// generateMockExecutions creates 100 execution history entries
func generateMockExecutions(reports []ScheduledReport) []ReportExecution {
	executions := make([]ReportExecution, 0, 100)
	rand.Seed(time.Now().UnixNano())

	statuses := []ExecutionStatus{StatusCompleted, StatusCompleted, StatusCompleted, StatusFailed, StatusCancelled}

	executionID := 1
	for _, report := range reports {
		numExecutions := rand.Intn(7) + 1
		if !report.IsActive {
			numExecutions = rand.Intn(3)
		}

		for i := 0; i < numExecutions && executionID <= 100; i++ {
			daysAgo := rand.Intn(int(time.Since(report.CreatedAt).Hours() / 24))
			startTime := time.Now().Add(-time.Duration(daysAgo*24+rand.Intn(24)) * time.Hour)

			status := statuses[rand.Intn(len(statuses))]

			var endTime *time.Time
			var fileSize int64
			var downloadURL string
			var errorMsg string
			duration := 0

			if status == StatusCompleted {
				et := startTime.Add(time.Duration(5+rand.Intn(55)) * time.Second)
				endTime = &et
				duration = int(et.Sub(startTime).Seconds())
				fileSize = int64(1024 * (50 + rand.Intn(500)))
				downloadURL = fmt.Sprintf("/api/reports/download/%s/exec_%d", report.ID, executionID)
			} else if status == StatusFailed {
				et := startTime.Add(time.Duration(2+rand.Intn(10)) * time.Second)
				endTime = &et
				duration = int(et.Sub(startTime).Seconds())
				errors := []string{
					"Database connection timeout",
					"Insufficient data for period",
					"File generation error",
					"Network timeout",
					"Memory limit exceeded",
				}
				errorMsg = errors[rand.Intn(len(errors))]
			}

			triggeredBy := "scheduler"
			if rand.Float64() < 0.2 {
				triggeredBy = "admin@rtx5.com"
			}

			execution := ReportExecution{
				ID:          fmt.Sprintf("EXEC%d", 10000+executionID),
				ReportID:    report.ID,
				ReportName:  report.Name,
				StartTime:   startTime,
				EndTime:     endTime,
				Status:      status,
				FileSize:    fileSize,
				DownloadURL: downloadURL,
				Error:       errorMsg,
				Duration:    duration,
				TriggeredBy: triggeredBy,
			}

			executions = append(executions, execution)
			executionID++
		}
	}

	sort.Slice(executions, func(i, j int) bool {
		return executions[i].StartTime.After(executions[j].StartTime)
	})

	return executions
}

// calculateNextRun calculates the next run time based on schedule
func calculateNextRun(schedule ReportSchedule, lastRun *time.Time) time.Time {
	base := time.Now()
	if lastRun != nil {
		base = *lastRun
	}

	switch schedule {
	case ScheduleDaily:
		return base.Add(24 * time.Hour)
	case ScheduleWeekly:
		return base.Add(7 * 24 * time.Hour)
	case ScheduleMonthly:
		return base.AddDate(0, 1, 0)
	case ScheduleQuarterly:
		return base.AddDate(0, 3, 0)
	default:
		return base.Add(24 * time.Hour)
	}
}

// ScheduledReportsHandler handles scheduled reports requests
type ScheduledReportsHandler struct {
	store       *ScheduledReportsStore
	authService *auth.Service
}

// NewScheduledReportsHandler creates a new scheduled reports handler
func NewScheduledReportsHandler(store *ScheduledReportsStore, authService *auth.Service) *ScheduledReportsHandler {
	return &ScheduledReportsHandler{
		store:       store,
		authService: authService,
	}
}

// HandleListReports handles GET /admin/reports/scheduled
func (h *ScheduledReportsHandler) HandleListReports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	reportType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")

	filtered := make([]ScheduledReport, 0)
	for _, report := range h.store.reports {
		if reportType != "" && string(report.Type) != reportType {
			continue
		}
		if status == "active" && !report.IsActive {
			continue
		}
		if status == "inactive" && report.IsActive {
			continue
		}
		filtered = append(filtered, report)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    filtered,
		"count":   len(filtered),
	})
}

// HandleCreateReport handles POST /admin/reports/scheduled
func (h *ScheduledReportsHandler) HandleCreateReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	adminID, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name        string                 `json:"name"`
		Type        ReportType             `json:"type"`
		Format      ReportFormat           `json:"format"`
		Schedule    ReportSchedule         `json:"schedule"`
		CronExpr    string                 `json:"cron_expression"`
		Recipients  []string               `json:"recipients"`
		Filters     map[string]interface{} `json:"filters"`
		Description string                 `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Type == "" {
		http.Error(w, "Name and type are required", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	reportID := fmt.Sprintf("RPT%d", 1000+len(h.store.reports))
	now := time.Now()

	report := ScheduledReport{
		ID:          reportID,
		Name:        req.Name,
		Type:        req.Type,
		Format:      req.Format,
		Schedule:    req.Schedule,
		CronExpr:    req.CronExpr,
		Recipients:  req.Recipients,
		Filters:     req.Filters,
		NextRun:     calculateNextRun(req.Schedule, nil),
		Status:      "active",
		CreatedBy:   fmt.Sprintf("admin-%d", adminID),
		CreatedAt:   now,
		IsActive:    true,
		Description: req.Description,
	}

	h.store.reports = append(h.store.reports, report)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Scheduled report created successfully",
		"data":    report,
	})
}

// HandleGetReport handles GET /admin/reports/scheduled/:id
func (h *ScheduledReportsHandler) HandleGetReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Report ID required", http.StatusBadRequest)
		return
	}
	reportID := pathParts[4]

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	var report *ScheduledReport
	for i := range h.store.reports {
		if h.store.reports[i].ID == reportID {
			report = &h.store.reports[i]
			break
		}
	}

	if report == nil {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	recentExecutions := make([]ReportExecution, 0)
	for _, exec := range h.store.executions {
		if exec.ReportID == reportID {
			recentExecutions = append(recentExecutions, exec)
			if len(recentExecutions) >= 10 {
				break
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"report":      report,
		"executions":  recentExecutions,
		"exec_count":  len(recentExecutions),
	})
}

// HandleUpdateReport handles PUT /admin/reports/scheduled/:id
func (h *ScheduledReportsHandler) HandleUpdateReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Report ID required", http.StatusBadRequest)
		return
	}
	reportID := pathParts[4]

	var req struct {
		Name        string                 `json:"name"`
		Format      ReportFormat           `json:"format"`
		Schedule    ReportSchedule         `json:"schedule"`
		Recipients  []string               `json:"recipients"`
		Filters     map[string]interface{} `json:"filters"`
		IsActive    *bool                  `json:"is_active"`
		Description string                 `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	var report *ScheduledReport
	for i := range h.store.reports {
		if h.store.reports[i].ID == reportID {
			report = &h.store.reports[i]
			break
		}
	}

	if report == nil {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	if req.Name != "" {
		report.Name = req.Name
	}
	if req.Format != "" {
		report.Format = req.Format
	}
	if req.Schedule != "" {
		report.Schedule = req.Schedule
		report.NextRun = calculateNextRun(req.Schedule, report.LastRun)
	}
	if len(req.Recipients) > 0 {
		report.Recipients = req.Recipients
	}
	if len(req.Filters) > 0 {
		report.Filters = req.Filters
	}
	if req.IsActive != nil {
		report.IsActive = *req.IsActive
		if *req.IsActive {
			report.Status = "active"
		} else {
			report.Status = "inactive"
		}
	}
	if req.Description != "" {
		report.Description = req.Description
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Report updated successfully",
		"data":    report,
	})
}

// HandleDeleteReport handles DELETE /admin/reports/scheduled/:id
func (h *ScheduledReportsHandler) HandleDeleteReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Report ID required", http.StatusBadRequest)
		return
	}
	reportID := pathParts[4]

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	found := false
	for i := range h.store.reports {
		if h.store.reports[i].ID == reportID {
			h.store.reports[i].IsActive = false
			h.store.reports[i].Status = "inactive"
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Report deactivated successfully",
	})
}

// HandleRunReport handles POST /admin/reports/scheduled/:id/run
func (h *ScheduledReportsHandler) HandleRunReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	adminID, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Report ID required", http.StatusBadRequest)
		return
	}
	reportID := pathParts[4]

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	var report *ScheduledReport
	for i := range h.store.reports {
		if h.store.reports[i].ID == reportID {
			report = &h.store.reports[i]
			break
		}
	}

	if report == nil {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	executionID := fmt.Sprintf("EXEC%d", 10000+len(h.store.executions))
	startTime := time.Now()
	endTime := startTime.Add(time.Duration(10+rand.Intn(50)) * time.Second)

	execution := ReportExecution{
		ID:          executionID,
		ReportID:    report.ID,
		ReportName:  report.Name,
		StartTime:   startTime,
		EndTime:     &endTime,
		Status:      StatusCompleted,
		FileSize:    int64(1024 * (100 + rand.Intn(400))),
		DownloadURL: fmt.Sprintf("/api/reports/download/%s/%s", report.ID, executionID),
		Duration:    int(endTime.Sub(startTime).Seconds()),
		TriggeredBy: fmt.Sprintf("admin-%d", adminID),
	}

	h.store.executions = append([]ReportExecution{execution}, h.store.executions...)
	now := startTime
	report.LastRun = &now
	report.NextRun = calculateNextRun(report.Schedule, &now)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"message":   "Report execution triggered successfully",
		"execution": execution,
	})
}

// HandleGetExecutions handles GET /admin/reports/executions
func (h *ScheduledReportsHandler) HandleGetExecutions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	status := r.URL.Query().Get("status")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 50
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	filtered := make([]ReportExecution, 0)
	for _, exec := range h.store.executions {
		if status != "" && string(exec.Status) != status {
			continue
		}
		filtered = append(filtered, exec)
	}

	totalCount := len(filtered)
	start := (page - 1) * limit
	end := start + limit

	if start >= totalCount {
		filtered = []ReportExecution{}
	} else {
		if end > totalCount {
			end = totalCount
		}
		filtered = filtered[start:end]
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    filtered,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       totalCount,
			"total_pages": int(math.Ceil(float64(totalCount) / float64(limit))),
		},
	})
}

// HandleGetTemplates handles GET /admin/reports/templates
func (h *ScheduledReportsHandler) HandleGetTemplates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	templates := h.store.templates
	h.store.mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"templates": templates,
		"count":     len(templates),
	})
}
