package repository

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/compliance/models"
)

// ComplianceRepository handles data persistence for compliance system
// In production, this would use PostgreSQL or similar database
type ComplianceRepository struct {
	mu sync.RWMutex

	// In-memory storage (replace with DB in production)
	transactionReports  map[string]*models.TransactionReport
	bestExecutionReports map[string]*models.BestExecutionReport
	kycRecords          map[string]*models.KYCRecord
	amlAlerts           map[string]*models.AMLAlert
	positionReports     map[string]*models.PositionReport
	auditTrail          []*models.AuditTrailEntry
	leverageLimits      map[string]*models.LeverageLimit
	riskWarnings        map[string]*models.RiskWarning
	clientStatements    map[string]*models.ClientStatement
	complaints          map[string]*models.Complaint
	segregatedAccounts  map[string]*models.SegregatedAccount
	gdprConsents        map[string]*models.GDPRConsent
	executionMetrics    []map[string]interface{}
	regulatorySubmissions []map[string]interface{}
}

func NewComplianceRepository() *ComplianceRepository {
	return &ComplianceRepository{
		transactionReports:   make(map[string]*models.TransactionReport),
		bestExecutionReports: make(map[string]*models.BestExecutionReport),
		kycRecords:           make(map[string]*models.KYCRecord),
		amlAlerts:            make(map[string]*models.AMLAlert),
		positionReports:      make(map[string]*models.PositionReport),
		auditTrail:           make([]*models.AuditTrailEntry, 0),
		leverageLimits:       make(map[string]*models.LeverageLimit),
		riskWarnings:         make(map[string]*models.RiskWarning),
		clientStatements:     make(map[string]*models.ClientStatement),
		complaints:           make(map[string]*models.Complaint),
		segregatedAccounts:   make(map[string]*models.SegregatedAccount),
		gdprConsents:         make(map[string]*models.GDPRConsent),
		executionMetrics:     make([]map[string]interface{}, 0),
		regulatorySubmissions: make([]map[string]interface{}, 0),
	}
}

// Transaction Reporting

func (r *ComplianceRepository) SaveTransactionReport(report *models.TransactionReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transactionReports[report.ID] = report
	return nil
}

func (r *ComplianceRepository) GetTransactionReport(id string) (*models.TransactionReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	report, ok := r.transactionReports[id]
	if !ok {
		return nil, fmt.Errorf("transaction report not found")
	}
	return report, nil
}

func (r *ComplianceRepository) UpdateTransactionReportStatus(id, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if report, ok := r.transactionReports[id]; ok {
		report.Status = status
		return nil
	}
	return fmt.Errorf("report not found")
}

func (r *ComplianceRepository) GetPendingTransactionReports() ([]*models.TransactionReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var pending []*models.TransactionReport
	for _, report := range r.transactionReports {
		if report.Status == "PENDING" {
			pending = append(pending, report)
		}
	}
	return pending, nil
}

func (r *ComplianceRepository) GetTransactionReportsByDate(date time.Time) ([]*models.TransactionReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var reports []*models.TransactionReport
	for _, report := range r.transactionReports {
		if report.CreatedAt.Format("2006-01-02") == date.Format("2006-01-02") {
			reports = append(reports, report)
		}
	}
	return reports, nil
}

// KYC/AML

func (r *ComplianceRepository) SaveKYCRecord(record *models.KYCRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.kycRecords[record.ID] = record
	return nil
}

func (r *ComplianceRepository) GetKYCRecord(id string) (*models.KYCRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	record, ok := r.kycRecords[id]
	if !ok {
		return nil, fmt.Errorf("KYC record not found")
	}
	return record, nil
}

func (r *ComplianceRepository) UpdateKYCDocumentStatus(id string, verified bool, provider string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if record, ok := r.kycRecords[id]; ok {
		record.DocumentVerified = verified
		record.VerificationProvider = provider
		record.UpdatedAt = time.Now()
		return nil
	}
	return fmt.Errorf("KYC record not found")
}

func (r *ComplianceRepository) UpdateKYCPEPStatus(id, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if record, ok := r.kycRecords[id]; ok {
		record.PEPStatus = status
		record.UpdatedAt = time.Now()
		return nil
	}
	return fmt.Errorf("KYC record not found")
}

func (r *ComplianceRepository) UpdateKYCSanctionsStatus(id string, match bool, lists []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if record, ok := r.kycRecords[id]; ok {
		record.SanctionsMatch = match
		record.SanctionsLists = lists
		record.UpdatedAt = time.Now()
		return nil
	}
	return fmt.Errorf("KYC record not found")
}

func (r *ComplianceRepository) UpdateKYCRiskRating(id, rating string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if record, ok := r.kycRecords[id]; ok {
		record.RiskRating = rating
		record.UpdatedAt = time.Now()
		return nil
	}
	return fmt.Errorf("KYC record not found")
}

func (r *ComplianceRepository) UpdateKYCLastScreening(id string, timestamp time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if record, ok := r.kycRecords[id]; ok {
		record.LastScreening = timestamp
		record.UpdatedAt = time.Now()
		return nil
	}
	return fmt.Errorf("KYC record not found")
}

func (r *ComplianceRepository) GetKYCRecordsBeforeDate(date time.Time) ([]*models.KYCRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var records []*models.KYCRecord
	for _, record := range r.kycRecords {
		if record.LastScreening.Before(date) {
			records = append(records, record)
		}
	}
	return records, nil
}

// AML Alerts

func (r *ComplianceRepository) SaveAMLAlert(alert *models.AMLAlert) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.amlAlerts[alert.ID] = alert
	return nil
}

func (r *ComplianceRepository) GetAMLAlert(id string) (*models.AMLAlert, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	alert, ok := r.amlAlerts[id]
	if !ok {
		return nil, fmt.Errorf("AML alert not found")
	}
	return alert, nil
}

func (r *ComplianceRepository) UpdateAMLAlertSAR(id string, filed bool, timestamp time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if alert, ok := r.amlAlerts[id]; ok {
		alert.SARFiled = filed
		alert.SARFiledAt = &timestamp
		return nil
	}
	return fmt.Errorf("AML alert not found")
}

func (r *ComplianceRepository) GetRecentTransactions(clientID string, duration time.Duration) ([]map[string]interface{}, error) {
	// Placeholder - would query actual transaction database
	return []map[string]interface{}{}, nil
}

// Audit Trail

func (r *ComplianceRepository) SaveAuditEntry(entry *models.AuditTrailEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.auditTrail = append(r.auditTrail, entry)
	return nil
}

func (r *ComplianceRepository) GetAuditEntriesDateRange(start, end time.Time) ([]*models.AuditTrailEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var entries []*models.AuditTrailEntry
	for _, entry := range r.auditTrail {
		if entry.Timestamp.After(start) && entry.Timestamp.Before(end) {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func (r *ComplianceRepository) GetAuditEntriesForClient(clientID string, start, end time.Time) ([]*models.AuditTrailEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var entries []*models.AuditTrailEntry
	for _, entry := range r.auditTrail {
		if entry.ClientID == clientID && entry.Timestamp.After(start) && entry.Timestamp.Before(end) {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func (r *ComplianceRepository) GetAuditEntriesByType(eventType string, start, end time.Time) ([]*models.AuditTrailEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var entries []*models.AuditTrailEntry
	for _, entry := range r.auditTrail {
		if entry.EventType == eventType && entry.Timestamp.After(start) && entry.Timestamp.Before(end) {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func (r *ComplianceRepository) GetAuditEntriesBeforeDate(date time.Time) ([]*models.AuditTrailEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var entries []*models.AuditTrailEntry
	for _, entry := range r.auditTrail {
		if entry.Timestamp.Before(date) {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func (r *ComplianceRepository) ArchiveAuditEntry(entry *models.AuditTrailEntry) error {
	// In production, move to archive storage (S3, cold storage, etc.)
	return nil
}

// Best Execution

func (r *ComplianceRepository) SaveBestExecutionReport(report *models.BestExecutionReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bestExecutionReports[report.ID] = report
	return nil
}

func (r *ComplianceRepository) GetBestExecutionReport(id string) (*models.BestExecutionReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	report, ok := r.bestExecutionReports[id]
	if !ok {
		return nil, fmt.Errorf("best execution report not found")
	}
	return report, nil
}

func (r *ComplianceRepository) UpdateBestExecutionReportPublished(id string, timestamp time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if report, ok := r.bestExecutionReports[id]; ok {
		report.PublishedAt = &timestamp
		return nil
	}
	return fmt.Errorf("report not found")
}

func (r *ComplianceRepository) SaveExecutionMetrics(metrics map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.executionMetrics = append(r.executionMetrics, metrics)
	return nil
}

func (r *ComplianceRepository) GetExecutionMetrics(instrumentClass string, start, end time.Time) ([]map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var filtered []map[string]interface{}
	for _, m := range r.executionMetrics {
		// Simplified filtering
		filtered = append(filtered, m)
	}
	return filtered, nil
}

func (r *ComplianceRepository) GetExecutionMetricsByLP(lpName, symbol string, start, end time.Time) ([]map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var filtered []map[string]interface{}
	for _, m := range r.executionMetrics {
		if m["lp_name"] == lpName && m["symbol"] == symbol {
			filtered = append(filtered, m)
		}
	}
	return filtered, nil
}

func (r *ComplianceRepository) GetVenueStatistics(instrumentClass string, start, end time.Time) ([]interface{}, error) {
	// Placeholder - would aggregate venue statistics
	return []interface{}{}, nil
}

// Leverage Limits

func (r *ComplianceRepository) GetLeverageLimit(jurisdiction models.Jurisdiction, instrumentClass string, clientClass models.ClientClassification) (*models.LeverageLimit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := fmt.Sprintf("%s_%s_%s", jurisdiction, instrumentClass, clientClass)
	limit, ok := r.leverageLimits[key]
	if !ok {
		return nil, nil // No specific limit found
	}
	return limit, nil
}

// Regulatory Submissions

func (r *ComplianceRepository) SaveRegulatorySubmission(submission map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.regulatorySubmissions = append(r.regulatorySubmissions, submission)
	return nil
}

// Helper methods for JSON marshaling

func (r *ComplianceRepository) MarshalKYCRecord(record *models.KYCRecord) ([]byte, error) {
	return json.Marshal(record)
}

func (r *ComplianceRepository) UnmarshalKYCRecord(data []byte) (*models.KYCRecord, error) {
	var record models.KYCRecord
	err := json.Unmarshal(data, &record)
	return &record, err
}
