package admin

import (
	"log"
	"sync"
	"time"
)

// AuditLog provides audit trail functionality
type AuditLog struct {
	mu       sync.RWMutex
	entries  []AuditEntry
	nextID   int64
	maxSize  int // Maximum entries to keep in memory
}

// NewAuditLog creates a new audit log
func NewAuditLog(maxSize int) *AuditLog {
	if maxSize <= 0 {
		maxSize = 10000 // Default 10k entries
	}

	return &AuditLog{
		entries: make([]AuditEntry, 0, maxSize),
		nextID:  1,
		maxSize: maxSize,
	}
}

// Log records an audit entry
func (a *AuditLog) Log(adminID int64, adminName, action, entityType string, entityID int64, changes interface{}, reason, ipAddress, userAgent, status, errorMsg string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	entry := AuditEntry{
		ID:         a.nextID,
		AdminID:    adminID,
		AdminName:  adminName,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Changes:    changes,
		Reason:     reason,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Status:     status,
		ErrorMsg:   errorMsg,
		CreatedAt:  time.Now(),
	}

	a.nextID++
	a.entries = append(a.entries, entry)

	// Trim old entries if exceeded max size
	if len(a.entries) > a.maxSize {
		a.entries = a.entries[len(a.entries)-a.maxSize:]
	}

	// Log to console for immediate visibility
	if status == "FAILED" {
		log.Printf("[AUDIT] FAILED: %s by %s on %s #%d: %s", action, adminName, entityType, entityID, errorMsg)
	} else {
		log.Printf("[AUDIT] %s by %s on %s #%d", action, adminName, entityType, entityID)
	}
}

// GetEntries returns audit entries with optional filters
func (a *AuditLog) GetEntries(adminID *int64, action *string, entityType *string, entityID *int64, fromDate, toDate *time.Time, limit int) []AuditEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var filtered []AuditEntry

	for i := len(a.entries) - 1; i >= 0; i-- {
		entry := a.entries[i]

		// Apply filters
		if adminID != nil && entry.AdminID != *adminID {
			continue
		}
		if action != nil && entry.Action != *action {
			continue
		}
		if entityType != nil && entry.EntityType != *entityType {
			continue
		}
		if entityID != nil && entry.EntityID != *entityID {
			continue
		}
		if fromDate != nil && entry.CreatedAt.Before(*fromDate) {
			continue
		}
		if toDate != nil && entry.CreatedAt.After(*toDate) {
			continue
		}

		filtered = append(filtered, entry)

		if limit > 0 && len(filtered) >= limit {
			break
		}
	}

	return filtered
}

// GetEntriesByAdmin returns all entries for a specific admin
func (a *AuditLog) GetEntriesByAdmin(adminID int64, limit int) []AuditEntry {
	return a.GetEntries(&adminID, nil, nil, nil, nil, nil, limit)
}

// GetEntriesByAction returns all entries for a specific action
func (a *AuditLog) GetEntriesByAction(action string, limit int) []AuditEntry {
	return a.GetEntries(nil, &action, nil, nil, nil, nil, limit)
}

// GetEntriesByEntity returns all entries for a specific entity
func (a *AuditLog) GetEntriesByEntity(entityType string, entityID int64, limit int) []AuditEntry {
	return a.GetEntries(nil, nil, &entityType, &entityID, nil, nil, limit)
}

// GetRecentEntries returns the most recent audit entries
func (a *AuditLog) GetRecentEntries(limit int) []AuditEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	start := len(a.entries) - limit
	if start < 0 {
		start = 0
	}

	entries := make([]AuditEntry, len(a.entries)-start)
	copy(entries, a.entries[start:])

	// Reverse to get newest first
	for i := 0; i < len(entries)/2; i++ {
		j := len(entries) - 1 - i
		entries[i], entries[j] = entries[j], entries[i]
	}

	return entries
}

// GetStats returns audit statistics
func (a *AuditLog) GetStats() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	actionCounts := make(map[string]int)
	adminCounts := make(map[string]int)
	successCount := 0
	failedCount := 0

	for _, entry := range a.entries {
		actionCounts[entry.Action]++
		adminCounts[entry.AdminName]++

		if entry.Status == "SUCCESS" {
			successCount++
		} else if entry.Status == "FAILED" {
			failedCount++
		}
	}

	return map[string]interface{}{
		"totalEntries": len(a.entries),
		"successCount": successCount,
		"failedCount":  failedCount,
		"actionCounts": actionCounts,
		"adminCounts":  adminCounts,
	}
}

// GetTodayEntries returns entries from today
func (a *AuditLog) GetTodayEntries() []AuditEntry {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return a.GetEntries(nil, nil, nil, nil, &startOfDay, &endOfDay, 0)
}

// SearchEntries performs a text search across all fields
func (a *AuditLog) SearchEntries(query string, limit int) []AuditEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var results []AuditEntry

	for i := len(a.entries) - 1; i >= 0; i-- {
		entry := a.entries[i]

		// Simple substring search across text fields
		if contains(entry.AdminName, query) ||
			contains(entry.Action, query) ||
			contains(entry.EntityType, query) ||
			contains(entry.Reason, query) ||
			contains(entry.IPAddress, query) ||
			contains(entry.ErrorMsg, query) {
			results = append(results, entry)

			if limit > 0 && len(results) >= limit {
				break
			}
		}
	}

	return results
}

// Helper function for case-insensitive substring search
func contains(str, substr string) bool {
	if len(str) == 0 || len(substr) == 0 {
		return false
	}

	// Simple case-insensitive check
	strLower := toLower(str)
	substrLower := toLower(substr)

	for i := 0; i <= len(strLower)-len(substrLower); i++ {
		if strLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}

	return false
}

func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// Count returns the total number of audit entries
func (a *AuditLog) Count() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.entries)
}

// Clear removes all audit entries (use with caution)
func (a *AuditLog) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.entries = make([]AuditEntry, 0, a.maxSize)
	log.Println("[AUDIT] Audit log cleared")
}
