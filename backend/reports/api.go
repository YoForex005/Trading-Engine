package reports

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ReportAPI handles HTTP endpoints for reports
type ReportAPI struct {
	store     ReportStore
	scheduler *ReportScheduler
}

// NewReportAPI creates a new report API handler
func NewReportAPI(store ReportStore, scheduler *ReportScheduler) *ReportAPI {
	return &ReportAPI{
		store:     store,
		scheduler: scheduler,
	}
}

// validateAuthToken extracts and validates JWT from Authorization header
func validateAuthToken(r *http.Request) (*auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	tokenString := parts[1]
	return auth.ValidateTokenWithDefault(tokenString)
}

// HandleListReports returns a list of reports with optional filtering
// GET /admin/reports?type=daily_trading&limit=10
func (api *ReportAPI) HandleListReports(w http.ResponseWriter, r *http.Request) {
	// JWT authentication
	_, err := validateAuthToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	reportType := r.URL.Query().Get("type") // Optional filter by type
	limitStr := r.URL.Query().Get("limit")

	limit := 30 // Default limit
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get reports from store
	reports, err := api.store.ListReports(reportType, limit)
	if err != nil {
		log.Printf("[ReportAPI] Failed to list reports: %v", err)
		http.Error(w, "Failed to retrieve reports", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reports": reports,
		"count":   len(reports),
	})
}

// HandleGetReport returns a single report by ID
// GET /admin/reports/{id}
func (api *ReportAPI) HandleGetReport(w http.ResponseWriter, r *http.Request) {
	// JWT authentication
	_, err := validateAuthToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract report ID from path
	// Expected format: /admin/reports/{id}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid report ID", http.StatusBadRequest)
		return
	}
	reportID := pathParts[3]

	// Get report from store
	report, err := api.store.GetReport(reportID)
	if err != nil {
		log.Printf("[ReportAPI] Report not found: %s - %v", reportID, err)
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// HandleGenerateReport manually generates a report on demand
// POST /admin/reports/generate
// Body: {"type": "daily_trading", "date": "2024-01-15"}
func (api *ReportAPI) HandleGenerateReport(w http.ResponseWriter, r *http.Request) {
	// JWT authentication
	_, err := validateAuthToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Type string `json:"type"` // daily_trading, daily_risk, weekly_revenue
		Date string `json:"date"` // YYYY-MM-DD format
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate report type
	if req.Type != "daily_trading" && req.Type != "daily_risk" && req.Type != "weekly_revenue" {
		http.Error(w, "Invalid report type. Must be: daily_trading, daily_risk, or weekly_revenue", http.StatusBadRequest)
		return
	}

	// Parse date
	var targetDate time.Time
	if req.Date == "" {
		// Default to yesterday
		targetDate = time.Now().AddDate(0, 0, -1)
	} else {
		parsedDate, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		targetDate = parsedDate
	}

	// Generate report using scheduler
	report, err := api.scheduler.GenerateNow(req.Type, targetDate)
	if err != nil {
		log.Printf("[ReportAPI] Failed to generate report: %v", err)
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}

	if report == nil {
		http.Error(w, "Report generation failed", http.StatusInternalServerError)
		return
	}

	log.Printf("[ReportAPI] Manually generated report: %s (%s)", report.ID, report.Type)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(report)
}

// HandleDownloadReportCSV downloads a report as CSV
// GET /admin/reports/{id}/csv
func (api *ReportAPI) HandleDownloadReportCSV(w http.ResponseWriter, r *http.Request) {
	// JWT authentication
	_, err := validateAuthToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract report ID from path
	// Expected format: /admin/reports/{id}/csv
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid report ID", http.StatusBadRequest)
		return
	}
	reportID := pathParts[3]

	// Get report from store
	report, err := api.store.GetReport(reportID)
	if err != nil {
		log.Printf("[ReportAPI] Report not found for CSV download: %s - %v", reportID, err)
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	// Convert report to CSV
	csvData, err := api.convertReportToCSV(report)
	if err != nil {
		log.Printf("[ReportAPI] Failed to convert report to CSV: %v", err)
		http.Error(w, "Failed to generate CSV", http.StatusInternalServerError)
		return
	}

	// Set headers for CSV download
	filename := fmt.Sprintf("%s_%s.csv", report.Type, report.Period)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	w.Write([]byte(csvData))
}

// convertReportToCSV converts a Report to CSV format
func (api *ReportAPI) convertReportToCSV(report *Report) (string, error) {
	var records [][]string

	switch report.Type {
	case "daily_trading":
		records = api.dailyTradingToCSV(report)
	case "daily_risk":
		records = api.dailyRiskToCSV(report)
	case "weekly_revenue":
		records = api.weeklyRevenueToCSV(report)
	default:
		return "", fmt.Errorf("unsupported report type for CSV: %s", report.Type)
	}

	// Write CSV to string
	var csvBuilder strings.Builder
	csvWriter := csv.NewWriter(&csvBuilder)

	for _, record := range records {
		if err := csvWriter.Write(record); err != nil {
			return "", err
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return "", err
	}

	return csvBuilder.String(), nil
}

// dailyTradingToCSV converts daily trading report to CSV records
func (api *ReportAPI) dailyTradingToCSV(report *Report) [][]string {
	records := [][]string{
		{"Daily Trading Summary - " + report.Period},
		{},
		{"Metric", "Value"},
		{"Total Trades", fmt.Sprintf("%v", report.Data["totalTrades"])},
		{"Total Volume", fmt.Sprintf("%.2f", report.Data["totalVolume"])},
		{"Total P&L", fmt.Sprintf("%.2f", report.Data["totalPnL"])},
		{"Active Accounts", fmt.Sprintf("%v", report.Data["activeAccounts"])},
		{"New Accounts", fmt.Sprintf("%v", report.Data["newAccounts"])},
		{"Total Accounts", fmt.Sprintf("%v", report.Data["totalAccounts"])},
		{},
		{"Top Symbols by Volume"},
		{"Symbol", "Volume"},
	}

	// Add top symbols
	if topSymbols, ok := report.Data["topSymbols"].([]interface{}); ok {
		for _, item := range topSymbols {
			if symbolMap, ok := item.(map[string]interface{}); ok {
				symbol := fmt.Sprintf("%v", symbolMap["Symbol"])
				volume := fmt.Sprintf("%.2f", symbolMap["Volume"])
				records = append(records, []string{symbol, volume})
			}
		}
	}

	return records
}

// dailyRiskToCSV converts daily risk report to CSV records
func (api *ReportAPI) dailyRiskToCSV(report *Report) [][]string {
	records := [][]string{
		{"Daily Risk Report - " + report.Period},
		{},
		{"Metric", "Value"},
		{"Average Margin Utilization", fmt.Sprintf("%.2f%%", report.Data["avgMarginUtilization"])},
		{"Total Margin Used", fmt.Sprintf("%.2f", report.Data["totalMarginUsed"])},
		{"Total Equity", fmt.Sprintf("%.2f", report.Data["totalEquity"])},
		{"Open Positions", fmt.Sprintf("%v", report.Data["openPositions"])},
		{},
		{"Exposure by Symbol"},
		{"Symbol", "Exposure"},
	}

	// Add exposure by symbol
	if exposureMap, ok := report.Data["exposureBySymbol"].(map[string]interface{}); ok {
		for symbol, exposure := range exposureMap {
			records = append(records, []string{symbol, fmt.Sprintf("%.2f", exposure)})
		}
	}

	return records
}

// weeklyRevenueToCSV converts weekly revenue report to CSV records
func (api *ReportAPI) weeklyRevenueToCSV(report *Report) [][]string {
	records := [][]string{
		{"Weekly Revenue Report - " + report.Period},
		{},
		{"Period", fmt.Sprintf("%s to %s", report.Data["startDate"], report.Data["endDate"])},
		{},
		{"Revenue Breakdown"},
		{"Type", "Amount"},
		{"Spread Revenue", fmt.Sprintf("%.2f", report.Data["spreadRevenue"])},
		{"Commission Revenue", fmt.Sprintf("%.2f", report.Data["commissionRevenue"])},
		{"Swap Revenue", fmt.Sprintf("%.2f", report.Data["swapRevenue"])},
		{"Total Revenue", fmt.Sprintf("%.2f", report.Data["totalRevenue"])},
		{},
		{"Revenue by Symbol"},
		{"Symbol", "Revenue"},
	}

	// Add revenue by symbol
	if revenueMap, ok := report.Data["revenueBySymbol"].(map[string]interface{}); ok {
		for symbol, revenue := range revenueMap {
			records = append(records, []string{symbol, fmt.Sprintf("%.2f", revenue)})
		}
	}

	return records
}
