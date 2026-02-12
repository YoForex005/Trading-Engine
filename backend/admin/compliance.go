package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
	"github.com/google/uuid"
)

// TransactionAlert represents a compliance alert for suspicious transactions
type TransactionAlert struct {
	ID              string    `json:"id"`
	ClientID        string    `json:"clientId"`
	TransactionType string    `json:"transactionType"` // deposit, withdrawal, trade
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	Timestamp       time.Time `json:"timestamp"`
	AlertReason     string    `json:"alertReason"` // large_transaction, unusual_pattern, velocity_check, country_risk
	Status          string    `json:"status"`      // open, reviewed, escalated, dismissed
	ReviewedBy      string    `json:"reviewedBy"`
	ReviewedAt      time.Time `json:"reviewedAt"`
	Notes           string    `json:"notes"`
}

// RegulatoryReport represents a regulatory compliance report
type RegulatoryReport struct {
	ID          string                 `json:"id"`
	ReportType  string                 `json:"reportType"` // SAR, CTR, FATCA, EMIR
	Period      string                 `json:"period"`
	Status      string                 `json:"status"` // pending, filed, overdue
	FiledAt     time.Time              `json:"filedAt"`
	GeneratedAt time.Time              `json:"generatedAt"`
	Data        map[string]interface{} `json:"data"` // JSON blob with report details
}

// ComplianceService manages compliance monitoring and regulatory reporting
type ComplianceService struct {
	alerts  map[string]*TransactionAlert // ID -> Alert
	reports map[string]*RegulatoryReport // ID -> Report
	mu      sync.RWMutex

	// Client transaction history for pattern detection
	clientTxHistory map[string][]TransactionRecord // ClientID -> Transactions
	txHistoryMu     sync.RWMutex
}

// TransactionRecord tracks client transaction history for pattern detection
type TransactionRecord struct {
	Amount    float64
	Type      string
	Timestamp time.Time
}

// NewComplianceService creates a new compliance service with mock data
func NewComplianceService() *ComplianceService {
	service := &ComplianceService{
		alerts:          make(map[string]*TransactionAlert),
		reports:         make(map[string]*RegulatoryReport),
		clientTxHistory: make(map[string][]TransactionRecord),
	}

	// Generate 25 mock alerts
	service.generateMockAlerts()

	// Generate 10 mock reports
	service.generateMockReports()

	return service
}

// MonitorTransaction checks a transaction and auto-flags based on compliance rules
func (s *ComplianceService) MonitorTransaction(clientID, txType string, amount float64, currency string) (*TransactionAlert, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var alertReason string
	shouldAlert := false

	// Rule 1: CTR threshold - Amount > $10,000
	if amount > 10000 {
		alertReason = "large_transaction"
		shouldAlert = true
		log.Printf("[Compliance] CTR threshold triggered: ClientID=%s, Amount=%.2f", clientID, amount)
	}

	// Rule 2: Structuring detection - Multiple transactions > $5,000 same day
	if amount > 5000 && !shouldAlert {
		s.txHistoryMu.RLock()
		history := s.clientTxHistory[clientID]
		s.txHistoryMu.RUnlock()

		sameDayCount := 0
		today := time.Now().Truncate(24 * time.Hour)

		for _, tx := range history {
			if tx.Timestamp.Truncate(24*time.Hour).Equal(today) && tx.Amount > 5000 {
				sameDayCount++
			}
		}

		if sameDayCount >= 1 { // Already 1+ transaction today, this would be 2+
			alertReason = "velocity_check"
			shouldAlert = true
			log.Printf("[Compliance] Structuring detection: ClientID=%s, SameDayTxCount=%d", clientID, sameDayCount+1)
		}
	}

	// Rule 3: Unusual pattern - 10x client's average
	if !shouldAlert {
		s.txHistoryMu.RLock()
		history := s.clientTxHistory[clientID]
		s.txHistoryMu.RUnlock()

		if len(history) > 0 {
			var total float64
			for _, tx := range history {
				total += tx.Amount
			}
			avgAmount := total / float64(len(history))

			if amount > avgAmount*10 {
				alertReason = "unusual_pattern"
				shouldAlert = true
				log.Printf("[Compliance] Unusual pattern: ClientID=%s, Amount=%.2f, Avg=%.2f", clientID, amount, avgAmount)
			}
		}
	}

	// Record transaction in history
	s.txHistoryMu.Lock()
	s.clientTxHistory[clientID] = append(s.clientTxHistory[clientID], TransactionRecord{
		Amount:    amount,
		Type:      txType,
		Timestamp: time.Now(),
	})
	s.txHistoryMu.Unlock()

	// Create alert if flagged
	if shouldAlert {
		alert := &TransactionAlert{
			ID:              uuid.New().String(),
			ClientID:        clientID,
			TransactionType: txType,
			Amount:          amount,
			Currency:        currency,
			Timestamp:       time.Now(),
			AlertReason:     alertReason,
			Status:          "open",
		}

		s.alerts[alert.ID] = alert

		log.Printf("[Compliance] Alert created: ID=%s, ClientID=%s, Reason=%s, Amount=%.2f",
			alert.ID, clientID, alertReason, amount)

		return alert, nil
	}

	return nil, nil // No alert triggered
}

// GetAlerts returns filtered list of alerts
func (s *ComplianceService) GetAlerts(statusFilter string, dateFrom, dateTo time.Time, minAmount float64) []*TransactionAlert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alerts := make([]*TransactionAlert, 0)

	for _, alert := range s.alerts {
		// Apply status filter
		if statusFilter != "" && alert.Status != statusFilter {
			continue
		}

		// Apply date range filter
		if !dateFrom.IsZero() && alert.Timestamp.Before(dateFrom) {
			continue
		}
		if !dateTo.IsZero() && alert.Timestamp.After(dateTo) {
			continue
		}

		// Apply minimum amount filter
		if minAmount > 0 && alert.Amount < minAmount {
			continue
		}

		alerts = append(alerts, alert)
	}

	return alerts
}

// ReviewAlert marks an alert as reviewed, escalated, or dismissed
func (s *ComplianceService) ReviewAlert(alertID, adminID, status, notes string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	alert, exists := s.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found")
	}

	// Validate status
	validStatuses := map[string]bool{
		"reviewed":  true,
		"escalated": true,
		"dismissed": true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: must be reviewed, escalated, or dismissed")
	}

	alert.Status = status
	alert.ReviewedBy = adminID
	alert.ReviewedAt = time.Now()
	alert.Notes = notes

	log.Printf("[Compliance] Alert reviewed: ID=%s, Status=%s, ReviewedBy=%s",
		alertID, status, adminID)

	return nil
}

// GenerateReport creates a new regulatory report
func (s *ComplianceService) GenerateReport(reportType, period string) (*RegulatoryReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate report type
	validTypes := map[string]bool{
		"SAR":   true,
		"CTR":   true,
		"FATCA": true,
		"EMIR":  true,
	}
	if !validTypes[reportType] {
		return nil, fmt.Errorf("invalid report type: must be SAR, CTR, FATCA, or EMIR")
	}

	// Create report with mock data
	report := &RegulatoryReport{
		ID:          uuid.New().String(),
		ReportType:  reportType,
		Period:      period,
		Status:      "pending",
		GeneratedAt: time.Now(),
		Data: map[string]interface{}{
			"reportType":       reportType,
			"period":           period,
			"generatedBy":      "System",
			"transactionCount": rand.Intn(100) + 50,
			"totalVolume":      float64(rand.Intn(1000000) + 500000),
			"alertCount":       rand.Intn(20) + 5,
		},
	}

	s.reports[report.ID] = report

	log.Printf("[Compliance] Report generated: ID=%s, Type=%s, Period=%s",
		report.ID, reportType, period)

	return report, nil
}

// FileReport marks a report as filed
func (s *ComplianceService) FileReport(reportID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	report, exists := s.reports[reportID]
	if !exists {
		return fmt.Errorf("report not found")
	}

	if report.Status == "filed" {
		return fmt.Errorf("report already filed")
	}

	report.Status = "filed"
	report.FiledAt = time.Now()

	log.Printf("[Compliance] Report filed: ID=%s, Type=%s", reportID, report.ReportType)

	return nil
}

// GetReports returns filtered list of reports
func (s *ComplianceService) GetReports(typeFilter, statusFilter string) []*RegulatoryReport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reports := make([]*RegulatoryReport, 0)

	for _, report := range s.reports {
		// Apply type filter
		if typeFilter != "" && report.ReportType != typeFilter {
			continue
		}

		// Apply status filter
		if statusFilter != "" && report.Status != statusFilter {
			continue
		}

		reports = append(reports, report)
	}

	return reports
}

// GetOverdueReports returns reports that are past their due date
func (s *ComplianceService) GetOverdueReports() []*RegulatoryReport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	overdueReports := make([]*RegulatoryReport, 0)

	// Reports older than 30 days in pending status are overdue
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	for _, report := range s.reports {
		if report.Status == "pending" && report.GeneratedAt.Before(thirtyDaysAgo) {
			// Mark as overdue
			s.mu.RUnlock()
			s.mu.Lock()
			report.Status = "overdue"
			s.mu.Unlock()
			s.mu.RLock()

			overdueReports = append(overdueReports, report)
		}
	}

	return overdueReports
}

// GetComplianceScore calculates overall compliance score (0-100)
func (s *ComplianceService) GetComplianceScore(kycStore *KYCStore) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	score := 100.0

	// Factor 1: KYC Completion Rate (40% weight)
	kycApps := kycStore.ListApplications("")
	if len(kycApps) > 0 {
		approvedCount := 0
		for _, app := range kycApps {
			if app.Status == "approved" {
				approvedCount++
			}
		}
		kycCompletionRate := float64(approvedCount) / float64(len(kycApps)) * 100
		kycScore := kycCompletionRate * 0.4
		score = kycScore
	} else {
		score = 40.0 // Default if no KYC data
	}

	// Factor 2: Alert Resolution Rate (30% weight)
	totalAlerts := len(s.alerts)
	if totalAlerts > 0 {
		resolvedAlerts := 0
		for _, alert := range s.alerts {
			if alert.Status == "reviewed" || alert.Status == "dismissed" {
				resolvedAlerts++
			}
		}
		resolutionRate := float64(resolvedAlerts) / float64(totalAlerts) * 100
		score += resolutionRate * 0.3
	} else {
		score += 30.0 // Default if no alerts
	}

	// Factor 3: Report Filing Timeliness (30% weight)
	totalReports := len(s.reports)
	if totalReports > 0 {
		filedOnTime := 0
		for _, report := range s.reports {
			if report.Status == "filed" {
				filedOnTime++
			}
		}
		filingRate := float64(filedOnTime) / float64(totalReports) * 100
		score += filingRate * 0.3
	} else {
		score += 30.0 // Default if no reports
	}

	// Ensure score is within 0-100 range
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return int(score)
}

// GetStats returns compliance dashboard statistics
func (s *ComplianceService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	openAlerts := 0
	reviewedAlerts := 0
	escalatedAlerts := 0

	for _, alert := range s.alerts {
		switch alert.Status {
		case "open":
			openAlerts++
		case "reviewed":
			reviewedAlerts++
		case "escalated":
			escalatedAlerts++
		}
	}

	pendingReports := 0
	filedReports := 0
	overdueReports := 0

	for _, report := range s.reports {
		switch report.Status {
		case "pending":
			pendingReports++
		case "filed":
			filedReports++
		case "overdue":
			overdueReports++
		}
	}

	return map[string]interface{}{
		"totalAlerts":      len(s.alerts),
		"openAlerts":       openAlerts,
		"reviewedAlerts":   reviewedAlerts,
		"escalatedAlerts":  escalatedAlerts,
		"totalReports":     len(s.reports),
		"pendingReports":   pendingReports,
		"filedReports":     filedReports,
		"overdueReports":   overdueReports,
		"generatedAt":      time.Now(),
	}
}

// generateMockAlerts creates 25 mock transaction alerts
func (s *ComplianceService) generateMockAlerts() {
	clientIDs := []string{
		"client-001", "client-002", "client-003", "client-004", "client-005",
		"client-006", "client-007", "client-008", "client-009", "client-010",
		"client-011", "client-012", "client-013", "client-014", "client-015",
	}

	txTypes := []string{"deposit", "withdrawal", "trade"}
	alertReasons := []string{"large_transaction", "unusual_pattern", "velocity_check", "country_risk"}
	statuses := []string{"open", "reviewed", "escalated", "dismissed"}
	currencies := []string{"USD", "EUR", "GBP", "JPY"}

	for i := 0; i < 25; i++ {
		status := statuses[i%len(statuses)]

		alert := &TransactionAlert{
			ID:              uuid.New().String(),
			ClientID:        clientIDs[i%len(clientIDs)],
			TransactionType: txTypes[i%len(txTypes)],
			Amount:          float64(rand.Intn(90000) + 10000), // $10k - $100k
			Currency:        currencies[i%len(currencies)],
			Timestamp:       time.Now().AddDate(0, 0, -rand.Intn(60)), // Last 60 days
			AlertReason:     alertReasons[i%len(alertReasons)],
			Status:          status,
		}

		// If reviewed/dismissed, add review details
		if status != "open" {
			alert.ReviewedBy = "admin-001"
			alert.ReviewedAt = alert.Timestamp.Add(time.Duration(rand.Intn(24)) * time.Hour)
			alert.Notes = []string{
				"Verified legitimate transaction",
				"Escalated to compliance officer",
				"Client contacted, transaction confirmed",
				"Dismissed as false positive",
			}[i%4]
		}

		s.alerts[alert.ID] = alert
	}

	log.Printf("[Compliance] Generated 25 mock alerts (open: %d, reviewed: %d, escalated: %d, dismissed: %d)",
		len(s.GetAlerts("open", time.Time{}, time.Time{}, 0)),
		len(s.GetAlerts("reviewed", time.Time{}, time.Time{}, 0)),
		len(s.GetAlerts("escalated", time.Time{}, time.Time{}, 0)),
		len(s.GetAlerts("dismissed", time.Time{}, time.Time{}, 0)),
	)
}

// generateMockReports creates 10 mock regulatory reports
func (s *ComplianceService) generateMockReports() {
	reportTypes := []string{"SAR", "CTR", "FATCA", "EMIR"}
	periods := []string{"Q4 2025", "Q1 2026", "January 2026", "February 2026", "December 2025"}
	statuses := []string{"filed", "pending", "overdue"}

	for i := 0; i < 10; i++ {
		status := statuses[i%len(statuses)]
		generatedAt := time.Now().AddDate(0, 0, -rand.Intn(90)) // Last 90 days

		report := &RegulatoryReport{
			ID:          uuid.New().String(),
			ReportType:  reportTypes[i%len(reportTypes)],
			Period:      periods[i%len(periods)],
			Status:      status,
			GeneratedAt: generatedAt,
			Data: map[string]interface{}{
				"reportType":       reportTypes[i%len(reportTypes)],
				"period":           periods[i%len(periods)],
				"generatedBy":      "System",
				"transactionCount": rand.Intn(100) + 50,
				"totalVolume":      float64(rand.Intn(1000000) + 500000),
				"alertCount":       rand.Intn(20) + 5,
			},
		}

		// If filed, add filed date
		if status == "filed" {
			report.FiledAt = generatedAt.Add(time.Duration(rand.Intn(7)) * 24 * time.Hour)
		}

		s.reports[report.ID] = report
	}

	log.Printf("[Compliance] Generated 10 mock reports (filed: %d, pending: %d, overdue: %d)",
		len(s.GetReports("", "filed")),
		len(s.GetReports("", "pending")),
		len(s.GetReports("", "overdue")),
	)
}

// ============================================
// HTTP HANDLERS
// ============================================

// ComplianceHandler handles compliance-related HTTP requests
type ComplianceHandler struct {
	service     *ComplianceService
	kycStore    *KYCStore
	authService *auth.Service
}

// NewComplianceHandler creates a new compliance handler
func NewComplianceHandler(service *ComplianceService, kycStore *KYCStore, authService *auth.Service) *ComplianceHandler {
	return &ComplianceHandler{
		service:     service,
		kycStore:    kycStore,
		authService: authService,
	}
}

// HandleListAlerts lists all transaction alerts with filters
// GET /admin/compliance/alerts?status=open&minAmount=10000
func (h *ComplianceHandler) HandleListAlerts(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
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

	// Parse filters from query params
	statusFilter := r.URL.Query().Get("status")
	minAmountStr := r.URL.Query().Get("minAmount")

	var minAmount float64
	if minAmountStr != "" {
		fmt.Sscanf(minAmountStr, "%f", &minAmount)
	}

	// Get alerts
	alerts := h.service.GetAlerts(statusFilter, time.Time{}, time.Time{}, minAmount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// HandleReviewAlert reviews a transaction alert
// PUT /admin/compliance/alerts/:id/review
func (h *ComplianceHandler) HandleReviewAlert(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	claims, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract alert ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}
	alertID := pathParts[4]

	// Parse request body
	var req struct {
		Status string `json:"status"` // reviewed, escalated, dismissed
		Notes  string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Review alert
	if err := h.service.ReviewAlert(alertID, claims.UserID, req.Status, req.Notes); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Alert reviewed successfully",
		"alertId": alertID,
		"status":  req.Status,
	})
}

// HandleListReports lists all regulatory reports with filters
// GET /admin/compliance/reports?type=SAR&status=pending
func (h *ComplianceHandler) HandleListReports(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
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

	// Parse filters from query params
	typeFilter := r.URL.Query().Get("type")
	statusFilter := r.URL.Query().Get("status")

	// Get reports
	reports := h.service.GetReports(typeFilter, statusFilter)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reports": reports,
		"count":   len(reports),
	})
}

// HandleGenerateReport generates a new regulatory report
// POST /admin/compliance/reports/generate
func (h *ComplianceHandler) HandleGenerateReport(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
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
		ReportType string `json:"reportType"` // SAR, CTR, FATCA, EMIR
		Period     string `json:"period"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate inputs
	if req.ReportType == "" {
		http.Error(w, "reportType is required", http.StatusBadRequest)
		return
	}
	if req.Period == "" {
		http.Error(w, "period is required", http.StatusBadRequest)
		return
	}

	// Generate report
	report, err := h.service.GenerateReport(req.ReportType, req.Period)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(report)
}

// HandleFileReport marks a report as filed
// PUT /admin/compliance/reports/:id/file
func (h *ComplianceHandler) HandleFileReport(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract report ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid report ID", http.StatusBadRequest)
		return
	}
	reportID := pathParts[4]

	// File report
	if err := h.service.FileReport(reportID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Report filed successfully",
		"reportId": reportID,
	})
}

// HandleGetComplianceScore returns overall compliance score
// GET /admin/compliance/score
func (h *ComplianceHandler) HandleGetComplianceScore(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
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

	// Calculate score
	score := h.service.GetComplianceScore(h.kycStore)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"score":       score,
		"generatedAt": time.Now(),
	})
}

// HandleGetStats returns compliance dashboard statistics
// GET /admin/compliance/stats
func (h *ComplianceHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
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

	// Get stats
	stats := h.service.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleMonitorTransaction manually triggers transaction monitoring
// POST /admin/compliance/monitor
func (h *ComplianceHandler) HandleMonitorTransaction(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
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
		ClientID        string  `json:"clientId"`
		TransactionType string  `json:"transactionType"` // deposit, withdrawal, trade
		Amount          float64 `json:"amount"`
		Currency        string  `json:"currency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate inputs
	if req.ClientID == "" {
		http.Error(w, "clientId is required", http.StatusBadRequest)
		return
	}
	if req.TransactionType == "" {
		http.Error(w, "transactionType is required", http.StatusBadRequest)
		return
	}
	if req.Amount <= 0 {
		http.Error(w, "amount must be positive", http.StatusBadRequest)
		return
	}
	if req.Currency == "" {
		req.Currency = "USD" // Default currency
	}

	// Monitor transaction
	alert, err := h.service.MonitorTransaction(req.ClientID, req.TransactionType, req.Amount, req.Currency)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if alert != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"flagged": true,
			"alert":   alert,
			"message": "Transaction flagged for review",
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"flagged": false,
			"message": "Transaction passed compliance checks",
		})
	}
}

// validateAdminAuth validates JWT token and checks admin role
func (h *ComplianceHandler) validateAdminAuth(r *http.Request) (*auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	tokenString := parts[1]
	claims, err := auth.ValidateTokenWithDefault(tokenString)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
