package services

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/epic1st/rtx/backend/compliance/models"
	"github.com/epic1st/rtx/backend/compliance/repository"
	"github.com/google/uuid"
)

// TransactionReportingService handles regulatory transaction reporting
type TransactionReportingService struct {
	repo *repository.ComplianceRepository
}

func NewTransactionReportingService(repo *repository.ComplianceRepository) *TransactionReportingService {
	return &TransactionReportingService{
		repo: repo,
	}
}

// CreateTransactionReport generates regulatory transaction report
func (s *TransactionReportingService) CreateTransactionReport(
	jurisdiction models.Jurisdiction,
	orderID, executionID, clientID, symbol string,
	side string, quantity, price float64,
	lpName string, classification models.ClientClassification,
) (*models.TransactionReport, error) {

	reportType := s.determineReportType(jurisdiction)

	report := &models.TransactionReport{
		ID:                   uuid.New().String(),
		ReportType:           reportType,
		Jurisdiction:         jurisdiction,
		TransactionID:        uuid.New().String(),
		OrderID:              orderID,
		ExecutionID:          executionID,
		ClientID:             clientID,
		Symbol:               symbol,
		Side:                 side,
		Quantity:             quantity,
		Price:                price,
		ExecutionTimestamp:   time.Now(),
		TradingVenue:         "OTC",
		LiquidityProvider:    lpName,
		Currency:             "USD",
		ClientClassification: classification,
		CreatedAt:            time.Now(),
		Status:               "PENDING",
	}

	// Add MiFID II specific fields
	if jurisdiction == models.JurisdictionEU || jurisdiction == models.JurisdictionUK {
		report.TransmissionOfOrder = true
		report.InvestmentDecisionMaker = clientID
		report.ExecutingTrader = "AUTOMATED"
		report.ShortSellingIndicator = side == "SELL"
	}

	if err := s.repo.SaveTransactionReport(report); err != nil {
		return nil, fmt.Errorf("failed to save transaction report: %w", err)
	}

	return report, nil
}

// SubmitReport submits transaction report to regulator
func (s *TransactionReportingService) SubmitReport(reportID string) error {
	report, err := s.repo.GetTransactionReport(reportID)
	if err != nil {
		return fmt.Errorf("report not found: %w", err)
	}

	// Format report based on type
	var formattedReport interface{}
	switch report.ReportType {
	case "MiFID_II":
		formattedReport = s.formatMiFIDII(report)
	case "EMIR":
		formattedReport = s.formatEMIR(report)
	case "CAT":
		formattedReport = s.formatCAT(report)
	}

	// Submit to trade repository or regulatory system
	if err := s.submitToRegulator(report.Jurisdiction, formattedReport); err != nil {
		report.Status = "FAILED"
		s.repo.UpdateTransactionReportStatus(reportID, "FAILED")
		return fmt.Errorf("submission failed: %w", err)
	}

	now := time.Now()
	report.Status = "SUBMITTED"
	report.SubmittedAt = &now
	s.repo.UpdateTransactionReportStatus(reportID, "SUBMITTED")

	return nil
}

// BatchSubmit submits multiple reports at once
func (s *TransactionReportingService) BatchSubmit(reportIDs []string) error {
	for _, id := range reportIDs {
		if err := s.SubmitReport(id); err != nil {
			// Log error but continue with others
			fmt.Printf("Failed to submit report %s: %v\n", id, err)
		}
	}
	return nil
}

// GetPendingReports returns all pending reports for submission
func (s *TransactionReportingService) GetPendingReports() ([]*models.TransactionReport, error) {
	return s.repo.GetPendingTransactionReports()
}

// determineReportType determines which regulatory report type is needed
func (s *TransactionReportingService) determineReportType(jurisdiction models.Jurisdiction) string {
	switch jurisdiction {
	case models.JurisdictionEU, models.JurisdictionUK:
		return "MiFID_II"
	case models.JurisdictionUS:
		return "CAT"
	default:
		return "STANDARD"
	}
}

// formatMiFIDII formats transaction report for MiFID II (27 required fields)
func (s *TransactionReportingService) formatMiFIDII(report *models.TransactionReport) map[string]interface{} {
	return map[string]interface{}{
		"trading_venue":           report.TradingVenue,
		"transaction_id":          report.TransactionID,
		"execution_timestamp":     report.ExecutionTimestamp.Format(time.RFC3339),
		"instrument_id":           report.Symbol,
		"buy_sell_indicator":      report.Side,
		"quantity":                report.Quantity,
		"price":                   report.Price,
		"currency":                report.Currency,
		"client_id":               report.ClientID,
		"investment_decision":     report.InvestmentDecisionMaker,
		"executing_trader":        report.ExecutingTrader,
		"transmission_of_order":   report.TransmissionOfOrder,
		"short_selling_indicator": report.ShortSellingIndicator,
		"waiver_indicator":        report.WaiverIndicator,
		"liquidity_provider":      report.LiquidityProvider,
		// ... additional 12 fields as required by MiFID II
	}
}

// formatEMIR formats for EMIR derivatives reporting
func (s *TransactionReportingService) formatEMIR(report *models.TransactionReport) map[string]interface{} {
	return map[string]interface{}{
		"counterparty_1":      report.ClientID,
		"counterparty_2":      report.LiquidityProvider,
		"product_type":        "DERIVATIVE",
		"notional_amount":     report.Quantity * report.Price,
		"execution_timestamp": report.ExecutionTimestamp.Format(time.RFC3339),
	}
}

// formatCAT formats for US Consolidated Audit Trail
func (s *TransactionReportingService) formatCAT(report *models.TransactionReport) map[string]interface{} {
	return map[string]interface{}{
		"order_id":            report.OrderID,
		"execution_id":        report.ExecutionID,
		"symbol":              report.Symbol,
		"quantity":            report.Quantity,
		"price":               report.Price,
		"side":                report.Side,
		"timestamp":           report.ExecutionTimestamp.Format(time.RFC3339),
		"routing_destination": report.LiquidityProvider,
	}
}

// submitToRegulator submits formatted report to regulatory system
func (s *TransactionReportingService) submitToRegulator(jurisdiction models.Jurisdiction, report interface{}) error {
	// In production, this would integrate with actual regulatory submission systems
	// For now, we'll log and store the submission

	data, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	hash := sha256.Sum256(data)

	// Store submission proof
	submission := map[string]interface{}{
		"jurisdiction": jurisdiction,
		"data":         string(data),
		"hash":         fmt.Sprintf("%x", hash),
		"timestamp":    time.Now(),
	}

	return s.repo.SaveRegulatorySubmission(submission)
}

// GenerateDailyReport generates end-of-day transaction report summary
func (s *TransactionReportingService) GenerateDailyReport() (map[string]interface{}, error) {
	reports, err := s.repo.GetTransactionReportsByDate(time.Now())
	if err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"date":           time.Now().Format("2006-01-02"),
		"total_reports":  len(reports),
		"submitted":      0,
		"pending":        0,
		"failed":         0,
		"by_jurisdiction": make(map[string]int),
	}

	for _, report := range reports {
		switch report.Status {
		case "SUBMITTED":
			summary["submitted"] = summary["submitted"].(int) + 1
		case "PENDING":
			summary["pending"] = summary["pending"].(int) + 1
		case "FAILED":
			summary["failed"] = summary["failed"].(int) + 1
		}

		jKey := string(report.Jurisdiction)
		if count, ok := summary["by_jurisdiction"].(map[string]int)[jKey]; ok {
			summary["by_jurisdiction"].(map[string]int)[jKey] = count + 1
		} else {
			summary["by_jurisdiction"].(map[string]int)[jKey] = 1
		}
	}

	return summary, nil
}
