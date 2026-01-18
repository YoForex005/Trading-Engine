package services

import (
	"fmt"
	"time"

	"github.com/epic1st/rtx/backend/compliance/models"
	"github.com/epic1st/rtx/backend/compliance/repository"
	"github.com/google/uuid"
)

// KYCAMLService handles Know Your Customer and Anti-Money Laundering
type KYCAMLService struct {
	repo *repository.ComplianceRepository
}

func NewKYCAMLService(repo *repository.ComplianceRepository) *KYCAMLService {
	return &KYCAMLService{
		repo: repo,
	}
}

// CreateKYCRecord initiates KYC process for client
func (s *KYCAMLService) CreateKYCRecord(clientID, fullName string, dob time.Time) (*models.KYCRecord, error) {
	record := &models.KYCRecord{
		ID:                   uuid.New().String(),
		ClientID:             clientID,
		FullName:             fullName,
		DateOfBirth:          dob,
		DocumentVerified:     false,
		AddressVerified:      false,
		PEPStatus:            "NOT_PEP",
		SanctionsMatch:       false,
		RiskRating:           "MEDIUM",
		OngoingMonitoring:    true,
		LastScreening:        time.Now(),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := s.repo.SaveKYCRecord(record); err != nil {
		return nil, fmt.Errorf("failed to save KYC record: %w", err)
	}

	return record, nil
}

// VerifyDocument verifies client identity document
func (s *KYCAMLService) VerifyDocument(kycID, documentType, documentNumber string, provider string) error {
	// Integration with KYC providers (Onfido, Jumio, Trulioo)
	verified, err := s.callKYCProvider(provider, documentType, documentNumber)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	if verified {
		return s.repo.UpdateKYCDocumentStatus(kycID, true, provider)
	}

	return fmt.Errorf("document verification failed")
}

// ScreenPEP screens client against Politically Exposed Persons lists
func (s *KYCAMLService) ScreenPEP(kycID, fullName string) (string, error) {
	// Integration with PEP screening services
	status, err := s.checkPEPStatus(fullName)
	if err != nil {
		return "", fmt.Errorf("PEP screening failed: %w", err)
	}

	if err := s.repo.UpdateKYCPEPStatus(kycID, status); err != nil {
		return "", err
	}

	// If PEP detected, increase risk rating
	if status == "PEP" || status == "CLOSE_ASSOCIATE" {
		s.repo.UpdateKYCRiskRating(kycID, "HIGH")
	}

	return status, nil
}

// ScreenSanctions screens against sanctions lists (OFAC, UN, EU)
func (s *KYCAMLService) ScreenSanctions(kycID, fullName string, nationality string) (bool, []string, error) {
	lists := []string{"OFAC", "UN", "EU"}

	matches := []string{}
	for _, list := range lists {
		match, err := s.checkSanctionsList(list, fullName, nationality)
		if err != nil {
			return false, nil, fmt.Errorf("sanctions check failed for %s: %w", list, err)
		}
		if match {
			matches = append(matches, list)
		}
	}

	hasMatch := len(matches) > 0

	if err := s.repo.UpdateKYCSanctionsStatus(kycID, hasMatch, matches); err != nil {
		return hasMatch, matches, err
	}

	// Auto-escalate if sanctions match
	if hasMatch {
		s.repo.UpdateKYCRiskRating(kycID, "CRITICAL")
		s.CreateAMLAlert(kycID, "SANCTIONS_MATCH", fmt.Sprintf("Matched on: %v", matches), "CRITICAL")
	}

	return hasMatch, matches, nil
}

// CalculateRiskRating calculates overall AML risk rating
func (s *KYCAMLService) CalculateRiskRating(kycID string) (string, error) {
	record, err := s.repo.GetKYCRecord(kycID)
	if err != nil {
		return "", err
	}

	score := 0

	// PEP status
	if record.PEPStatus == "PEP" {
		score += 50
	} else if record.PEPStatus == "CLOSE_ASSOCIATE" {
		score += 30
	}

	// Sanctions match
	if record.SanctionsMatch {
		score += 100 // Auto critical
	}

	// High-risk countries
	highRiskCountries := []string{"IR", "KP", "SY"} // Simplified
	for _, country := range highRiskCountries {
		if record.ResidenceCountry == country || record.Nationality == country {
			score += 40
		}
	}

	// Determine rating
	rating := "LOW"
	if score >= 100 {
		rating = "CRITICAL"
	} else if score >= 50 {
		rating = "HIGH"
	} else if score >= 25 {
		rating = "MEDIUM"
	}

	s.repo.UpdateKYCRiskRating(kycID, rating)
	return rating, nil
}

// CreateAMLAlert creates suspicious activity alert
func (s *KYCAMLService) CreateAMLAlert(clientID, alertType, description, severity string) (*models.AMLAlert, error) {
	alert := &models.AMLAlert{
		ID:          uuid.New().String(),
		ClientID:    clientID,
		AlertType:   alertType,
		Description: description,
		Severity:    severity,
		DetectedAt:  time.Now(),
		Status:      "PENDING",
		SARFiled:    false,
	}

	if err := s.repo.SaveAMLAlert(alert); err != nil {
		return nil, fmt.Errorf("failed to save AML alert: %w", err)
	}

	return alert, nil
}

// MonitorTransactionPatterns monitors for suspicious patterns
func (s *KYCAMLService) MonitorTransactionPatterns(clientID string, transactionIDs []string, amount float64) error {
	// Check for large deposits
	if amount >= 10000 {
		s.CreateAMLAlert(clientID, "LARGE_DEPOSIT", fmt.Sprintf("Large deposit: $%.2f", amount), "MEDIUM")
	}

	// Check for rapid turnover
	transactions, err := s.repo.GetRecentTransactions(clientID, 24*time.Hour)
	if err != nil {
		return err
	}

	if len(transactions) > 50 { // More than 50 transactions in 24h
		s.CreateAMLAlert(clientID, "RAPID_TURNOVER", "Unusually high trading frequency", "MEDIUM")
	}

	return nil
}

// FileSAR files Suspicious Activity Report with regulator
func (s *KYCAMLService) FileSAR(alertID, narrative string) error {
	alert, err := s.repo.GetAMLAlert(alertID)
	if err != nil {
		return err
	}

	// Generate SAR report
	sarReport := map[string]interface{}{
		"alert_id":    alertID,
		"client_id":   alert.ClientID,
		"alert_type":  alert.AlertType,
		"narrative":   narrative,
		"detected_at": alert.DetectedAt,
		"filed_at":    time.Now(),
	}

	// Submit to FinCEN (US) or FIU (other jurisdictions)
	if err := s.submitSAR(sarReport); err != nil {
		return fmt.Errorf("SAR submission failed: %w", err)
	}

	now := time.Now()
	alert.SARFiled = true
	alert.SARFiledAt = &now

	return s.repo.UpdateAMLAlertSAR(alertID, true, now)
}

// OngoingMonitoring performs periodic re-screening
func (s *KYCAMLService) OngoingMonitoring() error {
	// Get all KYC records that need re-screening (> 90 days since last check)
	cutoff := time.Now().AddDate(0, 0, -90)
	records, err := s.repo.GetKYCRecordsBeforeDate(cutoff)
	if err != nil {
		return err
	}

	for _, record := range records {
		// Re-screen PEP
		s.ScreenPEP(record.ID, record.FullName)

		// Re-screen sanctions
		s.ScreenSanctions(record.ID, record.FullName, record.Nationality)

		// Update last screening date
		s.repo.UpdateKYCLastScreening(record.ID, time.Now())
	}

	return nil
}

// Helper methods for external integrations

func (s *KYCAMLService) callKYCProvider(provider, docType, docNumber string) (bool, error) {
	// Placeholder for actual KYC provider integration
	// In production, integrate with Onfido, Jumio, Trulioo APIs
	return true, nil
}

func (s *KYCAMLService) checkPEPStatus(fullName string) (string, error) {
	// Placeholder for PEP screening service
	// In production, integrate with WorldCheck, Dow Jones, etc.
	return "NOT_PEP", nil
}

func (s *KYCAMLService) checkSanctionsList(list, fullName, nationality string) (bool, error) {
	// Placeholder for sanctions list checking
	// In production, integrate with OFAC, UN, EU sanctions databases
	return false, nil
}

func (s *KYCAMLService) submitSAR(report map[string]interface{}) error {
	// Placeholder for SAR submission
	// In production, integrate with FinCEN BSA E-Filing System or equivalent
	return nil
}
