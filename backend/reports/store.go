package reports

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Report represents a generated report
type Report struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // daily_trading, daily_risk, weekly_revenue
	Title       string                 `json:"title"`
	GeneratedAt time.Time              `json:"generatedAt"`
	Period      string                 `json:"period"` // e.g., "2024-01-15" or "2024-W03"
	Data        map[string]interface{} `json:"data"`
	Format      string                 `json:"format"` // json, csv
}

// ReportStore interface for storing and retrieving reports
type ReportStore interface {
	SaveReport(report *Report) error
	GetReport(id string) (*Report, error)
	ListReports(reportType string, limit int) ([]*Report, error)
	DeleteOldReports(olderThan time.Time) int
}

// InMemoryReportStore stores reports in memory (last 30 reports)
type InMemoryReportStore struct {
	reports   map[string]*Report
	reportIDs []string // Ordered list for LRU management
	mu        sync.RWMutex
	maxSize   int
}

// NewInMemoryReportStore creates a new in-memory report store
func NewInMemoryReportStore(maxSize int) *InMemoryReportStore {
	return &InMemoryReportStore{
		reports:   make(map[string]*Report),
		reportIDs: make([]string, 0, maxSize),
		maxSize:   maxSize,
	}
}

// SaveReport saves a report to the store
func (s *InMemoryReportStore) SaveReport(report *Report) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate ID if not set
	if report.ID == "" {
		report.ID = uuid.New().String()
	}

	// Add to map
	s.reports[report.ID] = report

	// Add to ordered list
	s.reportIDs = append(s.reportIDs, report.ID)

	// Enforce max size (LRU eviction)
	if len(s.reportIDs) > s.maxSize {
		// Remove oldest
		oldestID := s.reportIDs[0]
		delete(s.reports, oldestID)
		s.reportIDs = s.reportIDs[1:]
	}

	return nil
}

// GetReport retrieves a report by ID
func (s *InMemoryReportStore) GetReport(id string) (*Report, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	report, exists := s.reports[id]
	if !exists {
		return nil, fmt.Errorf("report not found: %s", id)
	}

	return report, nil
}

// ListReports lists reports, optionally filtered by type, with a limit
func (s *InMemoryReportStore) ListReports(reportType string, limit int) ([]*Report, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*Report

	// Iterate in reverse order (newest first)
	for i := len(s.reportIDs) - 1; i >= 0; i-- {
		id := s.reportIDs[i]
		report := s.reports[id]

		// Filter by type if specified
		if reportType != "" && report.Type != reportType {
			continue
		}

		results = append(results, report)

		// Apply limit
		if limit > 0 && len(results) >= limit {
			break
		}
	}

	return results, nil
}

// DeleteOldReports deletes reports older than the specified time
func (s *InMemoryReportStore) DeleteOldReports(olderThan time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	deletedCount := 0
	newReportIDs := make([]string, 0, len(s.reportIDs))

	for _, id := range s.reportIDs {
		report := s.reports[id]
		if report.GeneratedAt.Before(olderThan) {
			delete(s.reports, id)
			deletedCount++
		} else {
			newReportIDs = append(newReportIDs, id)
		}
	}

	s.reportIDs = newReportIDs
	return deletedCount
}
