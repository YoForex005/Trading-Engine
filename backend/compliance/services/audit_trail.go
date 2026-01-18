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

// AuditTrailService provides immutable audit logging
type AuditTrailService struct {
	repo         *repository.ComplianceRepository
	previousHash string // For blockchain-style chaining
}

func NewAuditTrailService(repo *repository.ComplianceRepository) *AuditTrailService {
	return &AuditTrailService{
		repo:         repo,
		previousHash: "",
	}
}

// LogEvent creates an immutable audit trail entry
func (s *AuditTrailService) LogEvent(
	eventType, userID, userRole, clientID string,
	orderID, tradeID, symbol, action string,
	before, after interface{},
	ipAddress, userAgent string,
) (*models.AuditTrailEntry, error) {

	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)

	entry := &models.AuditTrailEntry{
		ID:        uuid.New().String(),
		EventType: eventType,
		UserID:    userID,
		UserRole:  userRole,
		ClientID:  clientID,
		OrderID:   orderID,
		TradeID:   tradeID,
		Symbol:    symbol,
		Action:    action,
		Before:    string(beforeJSON),
		After:     string(afterJSON),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Timestamp: time.Now(),
	}

	// Calculate hash for tamper detection
	entry.Hash = s.calculateHash(entry)
	entry.PreviousHash = s.previousHash
	s.previousHash = entry.Hash

	if err := s.repo.SaveAuditEntry(entry); err != nil {
		return nil, fmt.Errorf("failed to save audit entry: %w", err)
	}

	return entry, nil
}

// LogOrderPlaced logs order placement event
func (s *AuditTrailService) LogOrderPlaced(userID, clientID, orderID, symbol string, orderData interface{}, ip, ua string) error {
	_, err := s.LogEvent(
		"ORDER_PLACED",
		userID,
		"TRADER",
		clientID,
		orderID,
		"",
		symbol,
		"PLACE_ORDER",
		nil,
		orderData,
		ip,
		ua,
	)
	return err
}

// LogOrderModified logs order modification
func (s *AuditTrailService) LogOrderModified(userID, clientID, orderID, symbol string, before, after interface{}, ip, ua string) error {
	_, err := s.LogEvent(
		"ORDER_MODIFIED",
		userID,
		"TRADER",
		clientID,
		orderID,
		"",
		symbol,
		"MODIFY_ORDER",
		before,
		after,
		ip,
		ua,
	)
	return err
}

// LogOrderCancelled logs order cancellation
func (s *AuditTrailService) LogOrderCancelled(userID, clientID, orderID, symbol string, orderData interface{}, ip, ua string) error {
	_, err := s.LogEvent(
		"ORDER_CANCELLED",
		userID,
		"TRADER",
		clientID,
		orderID,
		"",
		symbol,
		"CANCEL_ORDER",
		orderData,
		nil,
		ip,
		ua,
	)
	return err
}

// LogTradeExecuted logs trade execution
func (s *AuditTrailService) LogTradeExecuted(userID, clientID, orderID, tradeID, symbol string, tradeData interface{}, ip, ua string) error {
	_, err := s.LogEvent(
		"TRADE_EXECUTED",
		userID,
		"TRADER",
		clientID,
		orderID,
		tradeID,
		symbol,
		"EXECUTE_TRADE",
		nil,
		tradeData,
		ip,
		ua,
	)
	return err
}

// LogPositionClosed logs position closure
func (s *AuditTrailService) LogPositionClosed(userID, clientID, tradeID, symbol string, positionData interface{}, ip, ua string) error {
	_, err := s.LogEvent(
		"POSITION_CLOSED",
		userID,
		"TRADER",
		clientID,
		"",
		tradeID,
		symbol,
		"CLOSE_POSITION",
		positionData,
		nil,
		ip,
		ua,
	)
	return err
}

// LogAccountModification logs account changes
func (s *AuditTrailService) LogAccountModification(userID, clientID, action string, before, after interface{}, ip, ua string) error {
	_, err := s.LogEvent(
		"ACCOUNT_MODIFIED",
		userID,
		"ADMIN",
		clientID,
		"",
		"",
		"",
		action,
		before,
		after,
		ip,
		ua,
	)
	return err
}

// LogWithdrawal logs withdrawal request
func (s *AuditTrailService) LogWithdrawal(userID, clientID string, amount float64, ip, ua string) error {
	data := map[string]interface{}{
		"amount":    amount,
		"timestamp": time.Now(),
	}
	_, err := s.LogEvent(
		"WITHDRAWAL",
		userID,
		"TRADER",
		clientID,
		"",
		"",
		"",
		"REQUEST_WITHDRAWAL",
		nil,
		data,
		ip,
		ua,
	)
	return err
}

// LogDeposit logs deposit event
func (s *AuditTrailService) LogDeposit(userID, clientID string, amount float64, ip, ua string) error {
	data := map[string]interface{}{
		"amount":    amount,
		"timestamp": time.Now(),
	}
	_, err := s.LogEvent(
		"DEPOSIT",
		userID,
		"TRADER",
		clientID,
		"",
		"",
		"",
		"DEPOSIT_FUNDS",
		nil,
		data,
		ip,
		ua,
	)
	return err
}

// VerifyIntegrity verifies the audit trail integrity
func (s *AuditTrailService) VerifyIntegrity(startDate, endDate time.Time) (bool, []string, error) {
	entries, err := s.repo.GetAuditEntriesDateRange(startDate, endDate)
	if err != nil {
		return false, nil, err
	}

	var tamperedEntries []string
	previousHash := ""

	for _, entry := range entries {
		// Verify hash chain
		if entry.PreviousHash != previousHash {
			tamperedEntries = append(tamperedEntries, fmt.Sprintf("Entry %s: hash chain broken", entry.ID))
		}

		// Verify entry hash
		calculatedHash := s.calculateHash(entry)
		if calculatedHash != entry.Hash {
			tamperedEntries = append(tamperedEntries, fmt.Sprintf("Entry %s: hash mismatch", entry.ID))
		}

		previousHash = entry.Hash
	}

	return len(tamperedEntries) == 0, tamperedEntries, nil
}

// GetAuditHistory retrieves audit history for regulatory audit
func (s *AuditTrailService) GetAuditHistory(clientID string, startDate, endDate time.Time) ([]*models.AuditTrailEntry, error) {
	return s.repo.GetAuditEntriesForClient(clientID, startDate, endDate)
}

// GetEventsByType retrieves specific event types
func (s *AuditTrailService) GetEventsByType(eventType string, startDate, endDate time.Time) ([]*models.AuditTrailEntry, error) {
	return s.repo.GetAuditEntriesByType(eventType, startDate, endDate)
}

// ExportAuditTrail exports audit trail for regulatory submission
func (s *AuditTrailService) ExportAuditTrail(startDate, endDate time.Time, format string) (interface{}, error) {
	entries, err := s.repo.GetAuditEntriesDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	switch format {
	case "JSON":
		return entries, nil
	case "CSV":
		return s.convertToCSV(entries), nil
	case "XML":
		return s.convertToXML(entries), nil
	default:
		return entries, nil
	}
}

// RetentionPolicy enforces 5-7 year retention requirement
func (s *AuditTrailService) RetentionPolicy() error {
	// Regulatory requirement: minimum 5 years, many jurisdictions require 7 years
	retentionYears := 7
	cutoffDate := time.Now().AddDate(-retentionYears, 0, 0)

	// Archive old entries to cold storage instead of deleting
	// Regulatory compliance requires retention, not deletion
	oldEntries, err := s.repo.GetAuditEntriesBeforeDate(cutoffDate)
	if err != nil {
		return err
	}

	// Move to archive storage
	for _, entry := range oldEntries {
		if err := s.repo.ArchiveAuditEntry(entry); err != nil {
			return fmt.Errorf("failed to archive entry %s: %w", entry.ID, err)
		}
	}

	return nil
}

// calculateHash generates SHA256 hash of audit entry
func (s *AuditTrailService) calculateHash(entry *models.AuditTrailEntry) string {
	data := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%d",
		entry.ID,
		entry.EventType,
		entry.UserID,
		entry.ClientID,
		entry.OrderID,
		entry.Action,
		entry.Before,
		entry.After,
		entry.IPAddress,
		entry.PreviousHash,
		entry.Timestamp.Unix(),
	)

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func (s *AuditTrailService) convertToCSV(entries []*models.AuditTrailEntry) string {
	// Simplified CSV conversion
	csv := "ID,EventType,UserID,ClientID,Action,Timestamp\n"
	for _, e := range entries {
		csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s\n",
			e.ID, e.EventType, e.UserID, e.ClientID, e.Action, e.Timestamp.Format(time.RFC3339))
	}
	return csv
}

func (s *AuditTrailService) convertToXML(entries []*models.AuditTrailEntry) string {
	// Simplified XML conversion
	xml := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<AuditTrail>\n"
	for _, e := range entries {
		xml += fmt.Sprintf("  <Entry id=\"%s\" event=\"%s\" user=\"%s\" timestamp=\"%s\"/>\n",
			e.ID, e.EventType, e.UserID, e.Timestamp.Format(time.RFC3339))
	}
	xml += "</AuditTrail>"
	return xml
}
