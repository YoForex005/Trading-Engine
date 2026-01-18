package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/epic1st/rtx/backend/compliance/models"
	"github.com/epic1st/rtx/backend/compliance/services"
)

// ComplianceHandler handles HTTP requests for compliance operations
type ComplianceHandler struct {
	transactionService *services.TransactionReportingService
	kycService         *services.KYCAMLService
	auditService       *services.AuditTrailService
	bestExecService    *services.BestExecutionService
	leverageService    *services.LeverageLimitsService
}

func NewComplianceHandler(
	transactionService *services.TransactionReportingService,
	kycService *services.KYCAMLService,
	auditService *services.AuditTrailService,
	bestExecService *services.BestExecutionService,
	leverageService *services.LeverageLimitsService,
) *ComplianceHandler {
	return &ComplianceHandler{
		transactionService: transactionService,
		kycService:         kycService,
		auditService:       auditService,
		bestExecService:    bestExecService,
		leverageService:    leverageService,
	}
}

// Transaction Reporting Endpoints

func (h *ComplianceHandler) HandleGetPendingReports(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	reports, err := h.transactionService.GetPendingReports()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}

func (h *ComplianceHandler) HandleSubmitReport(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		ReportID string `json:"reportId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.transactionService.SubmitReport(req.ReportID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "reportId": req.ReportID})
}

func (h *ComplianceHandler) HandleDailyReport(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	summary, err := h.transactionService.GenerateDailyReport()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// KYC/AML Endpoints

func (h *ComplianceHandler) HandleCreateKYC(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		ClientID string `json:"clientId"`
		FullName string `json:"fullName"`
		DOB      string `json:"dob"` // YYYY-MM-DD
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	dob, err := time.Parse("2006-01-02", req.DOB)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	record, err := h.kycService.CreateKYCRecord(req.ClientID, req.FullName, dob)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

func (h *ComplianceHandler) HandleScreenPEP(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		KYCID    string `json:"kycId"`
		FullName string `json:"fullName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	status, err := h.kycService.ScreenPEP(req.KYCID, req.FullName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": status})
}

func (h *ComplianceHandler) HandleScreenSanctions(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		KYCID       string `json:"kycId"`
		FullName    string `json:"fullName"`
		Nationality string `json:"nationality"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hasMatch, lists, err := h.kycService.ScreenSanctions(req.KYCID, req.FullName, req.Nationality)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"match": hasMatch,
		"lists": lists,
	})
}

func (h *ComplianceHandler) HandleFileSAR(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		AlertID   string `json:"alertId"`
		Narrative string `json:"narrative"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.kycService.FileSAR(req.AlertID, req.Narrative); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// Audit Trail Endpoints

func (h *ComplianceHandler) HandleGetAuditHistory(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	clientID := r.URL.Query().Get("clientId")
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)

	entries, err := h.auditService.GetAuditHistory(clientID, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (h *ComplianceHandler) HandleVerifyAuditIntegrity(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)

	isValid, tamperedEntries, err := h.auditService.VerifyIntegrity(start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":           isValid,
		"tamperedEntries": tamperedEntries,
	})
}

func (h *ComplianceHandler) HandleExportAuditTrail(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	format := r.URL.Query().Get("format") // JSON, CSV, XML

	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)

	data, err := h.auditService.ExportAuditTrail(start, end, format)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// Best Execution Endpoints

func (h *ComplianceHandler) HandleGenerateRTS27(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Year            int    `json:"year"`
		Quarter         string `json:"quarter"`
		InstrumentClass string `json:"instrumentClass"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	report, err := h.bestExecService.GenerateRTS27Report(req.Year, req.Quarter, req.InstrumentClass)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *ComplianceHandler) HandleGetExecutionQuality(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	lpName := r.URL.Query().Get("lp")
	symbol := r.URL.Query().Get("symbol")
	hours, _ := strconv.Atoi(r.URL.Query().Get("hours"))

	if hours == 0 {
		hours = 24 // Default to last 24 hours
	}

	score, err := h.bestExecService.CalculateExecutionQuality(lpName, symbol, hours)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"lp":      lpName,
		"symbol":  symbol,
		"score":   score,
		"hours":   hours,
		"rating":  getQualityRating(score),
	})
}

// Leverage Limits Endpoints

func (h *ComplianceHandler) HandleValidateLeverage(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	var req struct {
		Jurisdiction      string `json:"jurisdiction"`
		ClientClass       string `json:"clientClass"`
		Symbol            string `json:"symbol"`
		InstrumentClass   string `json:"instrumentClass"`
		RequestedLeverage int    `json:"requestedLeverage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	jurisdiction := models.Jurisdiction(req.Jurisdiction)
	clientClass := models.ClientClassification(req.ClientClass)

	isValid, maxLeverage, err := h.leverageService.ValidateLeverage(
		jurisdiction,
		clientClass,
		req.Symbol,
		req.InstrumentClass,
		req.RequestedLeverage,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":       isValid,
		"maxLeverage": maxLeverage,
		"warning":     h.leverageService.DisplayLeverageWarning(clientClass, req.RequestedLeverage, "en"),
	})
}

func (h *ComplianceHandler) HandleGetESMALimits(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	clientClass := r.URL.Query().Get("clientClass")
	limits := h.leverageService.GetESMALimits(models.ClientClassification(clientClass))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(limits)
}

// Helper functions

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
}

func getQualityRating(score float64) string {
	if score >= 90 {
		return "EXCELLENT"
	} else if score >= 75 {
		return "GOOD"
	} else if score >= 60 {
		return "FAIR"
	}
	return "POOR"
}
