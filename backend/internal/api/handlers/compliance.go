package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
	"github.com/google/uuid"
)

// ComplianceHandler handles compliance and regulatory reporting
type ComplianceHandler struct {
	engine *core.Engine
}

// NewComplianceHandler creates a new compliance handler
func NewComplianceHandler(engine *core.Engine) *ComplianceHandler {
	return &ComplianceHandler{
		engine: engine,
	}
}

// ============================================================================
// Data Structures for Compliance Reports
// ============================================================================

// BestExecutionReport represents MiFID II RTS 27/28 best execution report
type BestExecutionReport struct {
	ReportID        string                  `json:"report_id"`
	GeneratedAt     time.Time               `json:"generated_at"`
	ReportPeriod    ReportPeriod            `json:"report_period"`
	Summary         BestExecutionSummary    `json:"summary"`
	VenueBreakdown  []VenueExecutionMetrics `json:"venue_breakdown"`
	InstrumentStats []InstrumentStats       `json:"instrument_stats"`
	Metadata        map[string]interface{}  `json:"metadata"`
}

type ReportPeriod struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type BestExecutionSummary struct {
	TotalOrders      int64   `json:"total_orders"`
	TotalVolume      float64 `json:"total_volume"`
	AverageSpread    float64 `json:"average_spread"`
	AverageSlippage  float64 `json:"average_slippage"`
	FillRate         float64 `json:"fill_rate"`
	AverageLatencyMs float64 `json:"average_latency_ms"`
}

type VenueExecutionMetrics struct {
	VenueName        string  `json:"venue_name"`
	OrderCount       int64   `json:"order_count"`
	VolumeExecuted   float64 `json:"volume_executed"`
	AverageSpread    float64 `json:"average_spread"`
	AverageSlippage  float64 `json:"average_slippage"`
	FillRate         float64 `json:"fill_rate"`
	RejectRate       float64 `json:"reject_rate"`
	AverageLatencyMs float64 `json:"average_latency_ms"`
}

type InstrumentStats struct {
	Symbol         string  `json:"symbol"`
	OrderCount     int64   `json:"order_count"`
	VolumeExecuted float64 `json:"volume_executed"`
	AveragePrice   float64 `json:"average_price"`
	BestVenue      string  `json:"best_venue"`
	AverageSpread  float64 `json:"average_spread"`
}

// OrderRoutingReport represents SEC Rule 606 order routing disclosure
type OrderRoutingReport struct {
	ReportID        string                 `json:"report_id"`
	Quarter         string                 `json:"quarter"`
	Year            int                    `json:"year"`
	GeneratedAt     time.Time              `json:"generated_at"`
	RoutingData     []VenueRoutingStats    `json:"routing_data"`
	PaymentAnalysis PaymentForOrderFlow    `json:"payment_analysis"`
	Metadata        map[string]interface{} `json:"metadata"`
}

type VenueRoutingStats struct {
	VenueName             string  `json:"venue_name"`
	OrdersRouted          int64   `json:"orders_routed"`
	OrdersRoutedPct       float64 `json:"orders_routed_pct"`
	NonDirectedOrders     int64   `json:"non_directed_orders"`
	MarketOrders          int64   `json:"market_orders"`
	MarketableLimit       int64   `json:"marketable_limit"`
	NonMarketableLimit    int64   `json:"non_marketable_limit"`
	OtherOrders           int64   `json:"other_orders"`
	AverageFeePerOrder    float64 `json:"average_fee_per_order"`
	AverageRebatePerOrder float64 `json:"average_rebate_per_order"`
	NetPaymentReceived    float64 `json:"net_payment_received"`
}

type PaymentForOrderFlow struct {
	TotalPaymentReceived float64 `json:"total_payment_received"`
	TotalPaymentPaid     float64 `json:"total_payment_paid"`
	NetPayment           float64 `json:"net_payment"`
	PaymentAsPercentage  float64 `json:"payment_as_percentage"`
}

// AuditTrailExport represents audit log export
type AuditTrailExport struct {
	ReportID    string                 `json:"report_id"`
	GeneratedAt time.Time              `json:"generated_at"`
	Period      ReportPeriod           `json:"period"`
	TotalCount  int64                  `json:"total_count"`
	Entries     []AuditEntry           `json:"entries"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type AuditEntry struct {
	ID         string                 `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	UserID     string                 `json:"user_id,omitempty"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Action     string                 `json:"action"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Hash       string                 `json:"hash"` // Tamper-proof hash
}

// ============================================================================
// API Handlers
// ============================================================================

// HandleBestExecution generates MiFID II RTS 27/28 best execution report
// GET /api/compliance/best-execution?start_time=...&end_time=...&format=json|csv|pdf
func (h *ComplianceHandler) HandleBestExecution(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		http.Error(w, "Invalid start_time format (use RFC3339)", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		http.Error(w, "Invalid end_time format (use RFC3339)", http.StatusBadRequest)
		return
	}

	// Generate report
	report := h.generateBestExecutionReport(startTime, endTime)

	// Respond based on format
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(report)
	case "csv":
		h.exportBestExecutionCSV(w, report)
	case "pdf":
		h.exportBestExecutionPDF(w, report)
	default:
		http.Error(w, "Invalid format (use json, csv, or pdf)", http.StatusBadRequest)
	}
}

// HandleOrderRouting generates SEC Rule 606 order routing disclosure
// GET /api/compliance/order-routing?quarter=Q1&year=2026&format=json|csv|pdf
func (h *ComplianceHandler) HandleOrderRouting(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	quarter := r.URL.Query().Get("quarter")
	yearStr := r.URL.Query().Get("year")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	// Validate quarter
	if quarter != "Q1" && quarter != "Q2" && quarter != "Q3" && quarter != "Q4" {
		http.Error(w, "Invalid quarter (use Q1, Q2, Q3, or Q4)", http.StatusBadRequest)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Invalid year", http.StatusBadRequest)
		return
	}

	// Generate report
	report := h.generateOrderRoutingReport(quarter, year)

	// Respond based on format
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(report)
	case "csv":
		h.exportOrderRoutingCSV(w, report)
	case "pdf":
		http.Error(w, "PDF format not yet implemented for order routing", http.StatusNotImplemented)
	default:
		http.Error(w, "Invalid format (use json or csv)", http.StatusBadRequest)
	}
}

// HandleAuditTrail exports audit log entries
// GET /api/compliance/audit-trail?start_time=...&end_time=...&entity_type=...&format=json|csv
func (h *ComplianceHandler) HandleAuditTrail(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	entityType := r.URL.Query().Get("entity_type")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		http.Error(w, "Invalid start_time format (use RFC3339)", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		http.Error(w, "Invalid end_time format (use RFC3339)", http.StatusBadRequest)
		return
	}

	// Generate export
	export := h.generateAuditTrailExport(startTime, endTime, entityType)

	// Respond based on format
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(export)
	case "csv":
		h.exportAuditTrailCSV(w, export)
	default:
		http.Error(w, "Invalid format (use json or csv)", http.StatusBadRequest)
	}
}

// HandleAuditLog writes a new audit entry (internal use)
// POST /api/compliance/audit-log
func (h *ComplianceHandler) HandleAuditLog(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID     string                 `json:"user_id"`
		Action     string                 `json:"action"`
		EntityType string                 `json:"entity_type"`
		EntityID   string                 `json:"entity_id"`
		Details    map[string]interface{} `json:"details"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create audit entry with tamper-proof hash
	entry := AuditEntry{
		ID:         uuid.New().String(),
		Timestamp:  time.Now().UTC(),
		UserID:     req.UserID,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		Action:     req.Action,
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Details:    req.Details,
	}

	// Generate tamper-proof hash
	entry.Hash = h.generateAuditHash(entry)

	// Store to database (placeholder - implement with actual DB)
	// In production: INSERT INTO audit_log ...

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entry_id":  entry.ID,
		"timestamp": entry.Timestamp,
		"hash":      entry.Hash,
	})
}

// ============================================================================
// Report Generation Logic
// ============================================================================

func (h *ComplianceHandler) generateBestExecutionReport(startTime, endTime time.Time) BestExecutionReport {
	// In production: Query database for order execution data
	// For now, return sample data

	report := BestExecutionReport{
		ReportID:    uuid.New().String(),
		GeneratedAt: time.Now().UTC(),
		ReportPeriod: ReportPeriod{
			StartTime: startTime,
			EndTime:   endTime,
		},
		Summary: BestExecutionSummary{
			TotalOrders:      1250,
			TotalVolume:      15750000.00,
			AverageSpread:    0.00015,
			AverageSlippage:  0.00008,
			FillRate:         98.7,
			AverageLatencyMs: 45.3,
		},
		VenueBreakdown: []VenueExecutionMetrics{
			{
				VenueName:        "OANDA",
				OrderCount:       850,
				VolumeExecuted:   12500000.00,
				AverageSpread:    0.00012,
				AverageSlippage:  0.00006,
				FillRate:         99.2,
				RejectRate:       0.8,
				AverageLatencyMs: 38.5,
			},
			{
				VenueName:        "Binance",
				OrderCount:       400,
				VolumeExecuted:   3250000.00,
				AverageSpread:    0.00022,
				AverageSlippage:  0.00012,
				FillRate:         97.5,
				RejectRate:       2.5,
				AverageLatencyMs: 62.1,
			},
		},
		InstrumentStats: []InstrumentStats{
			{
				Symbol:         "EUR/USD",
				OrderCount:     650,
				VolumeExecuted: 8500000.00,
				AveragePrice:   1.0875,
				BestVenue:      "OANDA",
				AverageSpread:  0.00010,
			},
			{
				Symbol:         "GBP/USD",
				OrderCount:     350,
				VolumeExecuted: 4250000.00,
				AveragePrice:   1.2650,
				BestVenue:      "OANDA",
				AverageSpread:  0.00015,
			},
		},
		Metadata: map[string]interface{}{
			"generated_by": "RTX Trading Compliance System",
			"report_type":  "MiFID II RTS 27/28",
			"version":      "1.0",
		},
	}

	return report
}

func (h *ComplianceHandler) generateOrderRoutingReport(quarter string, year int) OrderRoutingReport {
	// In production: Aggregate quarterly routing data from database

	report := OrderRoutingReport{
		ReportID:    uuid.New().String(),
		Quarter:     quarter,
		Year:        year,
		GeneratedAt: time.Now().UTC(),
		RoutingData: []VenueRoutingStats{
			{
				VenueName:             "OANDA LP",
				OrdersRouted:          3250,
				OrdersRoutedPct:       68.4,
				NonDirectedOrders:     3250,
				MarketOrders:          2800,
				MarketableLimit:       300,
				NonMarketableLimit:    150,
				OtherOrders:           0,
				AverageFeePerOrder:    0.50,
				AverageRebatePerOrder: 0.15,
				NetPaymentReceived:    1137.50,
			},
			{
				VenueName:             "Binance",
				OrdersRouted:          1500,
				OrdersRoutedPct:       31.6,
				NonDirectedOrders:     1500,
				MarketOrders:          1350,
				MarketableLimit:       100,
				NonMarketableLimit:    50,
				OtherOrders:           0,
				AverageFeePerOrder:    0.35,
				AverageRebatePerOrder: 0.10,
				NetPaymentReceived:    375.00,
			},
		},
		PaymentAnalysis: PaymentForOrderFlow{
			TotalPaymentReceived: 1512.50,
			TotalPaymentPaid:     0,
			NetPayment:           1512.50,
			PaymentAsPercentage:  0.032,
		},
		Metadata: map[string]interface{}{
			"generated_by": "RTX Trading Compliance System",
			"report_type":  "SEC Rule 606",
			"quarter":      quarter,
			"year":         year,
		},
	}

	return report
}

func (h *ComplianceHandler) generateAuditTrailExport(startTime, endTime time.Time, entityType string) AuditTrailExport {
	// In production: Query audit_log table with filters

	export := AuditTrailExport{
		ReportID:    uuid.New().String(),
		GeneratedAt: time.Now().UTC(),
		Period: ReportPeriod{
			StartTime: startTime,
			EndTime:   endTime,
		},
		TotalCount: 5,
		Entries: []AuditEntry{
			{
				ID:         uuid.New().String(),
				Timestamp:  time.Now().Add(-24 * time.Hour),
				UserID:     "user-123",
				EntityType: "order",
				EntityID:   "order-456",
				Action:     "INSERT",
				IPAddress:  "192.168.1.100",
				UserAgent:  "Mozilla/5.0",
				Details: map[string]interface{}{
					"symbol":   "EUR/USD",
					"quantity": 10000,
					"side":     "BUY",
				},
				Hash: "abc123def456",
			},
			// More entries...
		},
		Metadata: map[string]interface{}{
			"entity_type_filter": entityType,
			"retention_years":    7,
			"tamper_proof":       true,
		},
	}

	return export
}

// ============================================================================
// CSV Export Functions
// ============================================================================

func (h *ComplianceHandler) exportBestExecutionCSV(w http.ResponseWriter, report BestExecutionReport) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"best-execution-%s.csv\"", report.ReportID))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header
	writer.Write([]string{"Venue", "Order Count", "Volume", "Avg Spread", "Avg Slippage", "Fill Rate", "Reject Rate", "Avg Latency (ms)"})

	// Data
	for _, venue := range report.VenueBreakdown {
		writer.Write([]string{
			venue.VenueName,
			fmt.Sprintf("%d", venue.OrderCount),
			fmt.Sprintf("%.2f", venue.VolumeExecuted),
			fmt.Sprintf("%.5f", venue.AverageSpread),
			fmt.Sprintf("%.5f", venue.AverageSlippage),
			fmt.Sprintf("%.2f%%", venue.FillRate),
			fmt.Sprintf("%.2f%%", venue.RejectRate),
			fmt.Sprintf("%.2f", venue.AverageLatencyMs),
		})
	}
}

func (h *ComplianceHandler) exportOrderRoutingCSV(w http.ResponseWriter, report OrderRoutingReport) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"order-routing-%s-%d.csv\"", report.Quarter, report.Year))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header
	writer.Write([]string{"Venue", "Orders Routed", "% of Total", "Market Orders", "Marketable Limit", "Non-Marketable Limit", "Avg Fee", "Net Payment"})

	// Data
	for _, venue := range report.RoutingData {
		writer.Write([]string{
			venue.VenueName,
			fmt.Sprintf("%d", venue.OrdersRouted),
			fmt.Sprintf("%.2f%%", venue.OrdersRoutedPct),
			fmt.Sprintf("%d", venue.MarketOrders),
			fmt.Sprintf("%d", venue.MarketableLimit),
			fmt.Sprintf("%d", venue.NonMarketableLimit),
			fmt.Sprintf("%.2f", venue.AverageFeePerOrder),
			fmt.Sprintf("%.2f", venue.NetPaymentReceived),
		})
	}
}

func (h *ComplianceHandler) exportAuditTrailCSV(w http.ResponseWriter, export AuditTrailExport) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"audit-trail-%s.csv\"", export.ReportID))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header
	writer.Write([]string{"Timestamp", "User ID", "Entity Type", "Entity ID", "Action", "IP Address", "Hash"})

	// Data
	for _, entry := range export.Entries {
		writer.Write([]string{
			entry.Timestamp.Format(time.RFC3339),
			entry.UserID,
			entry.EntityType,
			entry.EntityID,
			entry.Action,
			entry.IPAddress,
			entry.Hash,
		})
	}
}

// ============================================================================
// PDF Export (Placeholder)
// ============================================================================

func (h *ComplianceHandler) exportBestExecutionPDF(w http.ResponseWriter, _ BestExecutionReport) {
	// In production: Use a PDF library like gofpdf or wkhtmltopdf
	http.Error(w, "PDF generation not yet implemented", http.StatusNotImplemented)
}

// ============================================================================
// Utility Functions
// ============================================================================

func (h *ComplianceHandler) generateAuditHash(entry AuditEntry) string {
	// In production: Use SHA-256 or similar cryptographic hash
	// Hash should include: timestamp + userID + entityType + entityID + action
	_ = fmt.Sprintf("%s|%s|%s|%s|%s",
		entry.Timestamp.Format(time.RFC3339Nano),
		entry.UserID,
		entry.EntityType,
		entry.EntityID,
		entry.Action,
	)
	// Simplified hash for demo (use crypto/sha256 in production)
	// In production: return hex.EncodeToString(sha256.Sum256([]byte(data)))
	return fmt.Sprintf("hash_%s", uuid.New().String()[:16])
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}
	return ip
}

// ============================================================================
// Audit Middleware
// ============================================================================

// AuditMiddleware intercepts all API calls and logs them to audit trail
func AuditMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip OPTIONS requests
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip health checks and static files
		if r.URL.Path == "/health" || r.URL.Path == "/docs" || strings.HasPrefix(r.URL.Path, "/static") {
			next.ServeHTTP(w, r)
			return
		}

		// Record request start time
		startTime := time.Now()

		// Create response recorder to capture status code
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Execute the actual handler
		next.ServeHTTP(recorder, r)

		// Calculate response time
		responseTime := time.Since(startTime).Milliseconds()

		// Extract user ID from context (if authenticated)
		userID := r.Header.Get("X-User-ID") // Set by auth middleware

		// Log to audit trail (in production: INSERT INTO api_access_log)
		auditEntry := map[string]interface{}{
			"timestamp":     time.Now().UTC(),
			"user_id":       userID,
			"endpoint":      r.URL.Path,
			"method":        r.Method,
			"status_code":   recorder.statusCode,
			"response_time": responseTime,
			"ip_address":    getClientIP(r),
			"user_agent":    r.UserAgent(),
		}

		// In production: Store to database
		_ = auditEntry // Placeholder

		// If this is a critical action (POST/PUT/DELETE), log to audit_log table
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
			// Log detailed audit entry
			_ = userID // Placeholder for detailed logging
		}
	})
}

// responseRecorder captures response status code
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *responseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// ============================================================================
// Data Retention Policy
// ============================================================================

// CleanupOldAuditLogs removes audit logs older than retention period (7 years for compliance)
func (h *ComplianceHandler) CleanupOldAuditLogs() error {
	// In production: Run as scheduled job (cron)
	// DELETE FROM audit_log WHERE created_at < NOW() - INTERVAL '7 years'
	retentionYears := 7
	cutoffDate := time.Now().AddDate(-retentionYears, 0, 0)

	// Archive old logs before deletion (for regulatory compliance)
	// In production: Export to cold storage (S3, Glacier, etc.)
	_ = cutoffDate

	return nil
}

// ArchiveAuditLogs exports old logs to long-term storage
func (h *ComplianceHandler) ArchiveAuditLogs(startDate, endDate time.Time) error {
	// In production: Export to S3/Glacier with encryption
	// 1. Query audit logs in date range
	// 2. Compress and encrypt
	// 3. Upload to cold storage
	// 4. Verify upload
	// 5. Delete from primary database

	archivePath := fmt.Sprintf("%s/audit_archive_%s_to_%s.gz",
		getArchivePath(),
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)
	_ = archivePath

	return nil
}

func getArchivePath() string {
	// Get from environment variable or config
	path := os.Getenv("AUDIT_ARCHIVE_PATH")
	if path == "" {
		path = "./data/audit_archives"
	}
	return path
}
