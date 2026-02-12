//go:build rtx_legacy_admin
// +build rtx_legacy_admin

package admin

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// AuditAction represents the type of action performed
type AuditAction string

const (
	ActionLogin              AuditAction = "LOGIN"
	ActionLogout             AuditAction = "LOGOUT"
	ActionCreateUser         AuditAction = "CREATE_USER"
	ActionModifyUser         AuditAction = "MODIFY_USER"
	ActionDisableUser        AuditAction = "DISABLE_USER"
	ActionEnableUser         AuditAction = "ENABLE_USER"
	ActionModifySymbol       AuditAction = "MODIFY_SYMBOL"
	ActionApproveWithdrawal  AuditAction = "APPROVE_WITHDRAWAL"
	ActionRejectWithdrawal   AuditAction = "REJECT_WITHDRAWAL"
	ActionModifyRiskLimit    AuditAction = "MODIFY_RISK_LIMIT"
	ActionExportData         AuditAction = "EXPORT_DATA"
	ActionViewSensitiveData  AuditAction = "VIEW_SENSITIVE_DATA"
	ActionCreateAccount      AuditAction = "CREATE_ACCOUNT"
	ActionModifyAccount      AuditAction = "MODIFY_ACCOUNT"
	ActionDeposit            AuditAction = "DEPOSIT"
	ActionWithdraw           AuditAction = "WITHDRAW"
	ActionAdjustBalance      AuditAction = "ADJUST_BALANCE"
	ActionToggleSymbol       AuditAction = "TOGGLE_SYMBOL"
	ActionUpdateSpread       AuditAction = "UPDATE_SPREAD"
)

// AuditLogEntry represents a single audit log entry
type AuditLogEntry struct {
	ID          string                 `json:"id"`
	AdminID     string                 `json:"adminId"`
	AdminEmail  string                 `json:"adminEmail"`
	Action      AuditAction            `json:"action"`
	Resource    string                 `json:"resource"`    // e.g., "user", "symbol", "account"
	ResourceID  string                 `json:"resourceId"`  // ID of affected resource
	Details     map[string]interface{} `json:"details"`     // Additional context
	IPAddress   string                 `json:"ipAddress"`
	Timestamp   int64                  `json:"timestamp"`   // Unix timestamp
	Success     bool                   `json:"success"`
	ErrorMsg    string                 `json:"errorMsg,omitempty"`
}

// AuditLogger manages audit log entries
type AuditLogger struct {
	entries    []AuditLogEntry
	maxEntries int
	mu         sync.RWMutex
}

// AuditFilters for querying audit logs
type AuditFilters struct {
	AdminID    string
	Action     AuditAction
	Resource   string
	FromTime   int64 // Unix timestamp
	ToTime     int64 // Unix timestamp
	Success    *bool // nil = all, true = success only, false = failures only
	Limit      int
	Offset     int
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(maxEntries int) *AuditLogger {
	if maxEntries <= 0 {
		maxEntries = 10000 // Default max entries
	}
	return &AuditLogger{
		entries:    make([]AuditLogEntry, 0, maxEntries),
		maxEntries: maxEntries,
	}
}

// Log adds an audit entry to the log
func (al *AuditLogger) Log(entry AuditLogEntry) {
	al.mu.Lock()
	defer al.mu.Unlock()

	// Generate ID if not provided
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	// Set timestamp if not provided
	if entry.Timestamp == 0 {
		entry.Timestamp = time.Now().Unix()
	}

	// Add entry
	al.entries = append(al.entries, entry)

	// FIFO eviction if exceeds max
	if len(al.entries) > al.maxEntries {
		// Remove oldest entries (keep last maxEntries)
		al.entries = al.entries[len(al.entries)-al.maxEntries:]
	}
}

// Query retrieves audit entries based on filters
func (al *AuditLogger) Query(filters AuditFilters) []AuditLogEntry {
	al.mu.RLock()
	defer al.mu.RUnlock()

	// Default limit
	if filters.Limit <= 0 {
		filters.Limit = 50
	}

	var results []AuditLogEntry

	// Filter entries
	for _, entry := range al.entries {
		// Apply filters
		if filters.AdminID != "" && entry.AdminID != filters.AdminID {
			continue
		}
		if filters.Action != "" && entry.Action != filters.Action {
			continue
		}
		if filters.Resource != "" && entry.Resource != filters.Resource {
			continue
		}
		if filters.FromTime > 0 && entry.Timestamp < filters.FromTime {
			continue
		}
		if filters.ToTime > 0 && entry.Timestamp > filters.ToTime {
			continue
		}
		if filters.Success != nil && entry.Success != *filters.Success {
			continue
		}

		results = append(results, entry)
	}

	// Sort by timestamp descending (newest first)
	for i := 0; i < len(results)/2; i++ {
		j := len(results) - 1 - i
		results[i], results[j] = results[j], results[i]
	}

	// Apply pagination
	if filters.Offset >= len(results) {
		return []AuditLogEntry{}
	}

	start := filters.Offset
	end := start + filters.Limit
	if end > len(results) {
		end = len(results)
	}

	return results[start:end]
}

// GetByID retrieves a specific audit entry by ID
func (al *AuditLogger) GetByID(id string) *AuditLogEntry {
	al.mu.RLock()
	defer al.mu.RUnlock()

	for i := len(al.entries) - 1; i >= 0; i-- {
		if al.entries[i].ID == id {
			entry := al.entries[i]
			return &entry
		}
	}
	return nil
}

// GetStats returns audit statistics for the last 24 hours
func (al *AuditLogger) GetStats() AuditStats {
	al.mu.RLock()
	defer al.mu.RUnlock()

	now := time.Now().Unix()
	cutoff := now - 86400 // 24 hours ago

	stats := AuditStats{
		TopActions: make(map[string]int),
	}

	uniqueAdmins := make(map[string]bool)
	actionCounts := make(map[string]int)

	for i := len(al.entries) - 1; i >= 0; i-- {
		entry := al.entries[i]

		// Only count entries from last 24h
		if entry.Timestamp < cutoff {
			break // Entries are sorted by time, so we can break early
		}

		stats.TotalActions24h++
		uniqueAdmins[entry.AdminID] = true
		actionCounts[string(entry.Action)]++

		if !entry.Success {
			stats.FailedAttempts24h++
		}
	}

	stats.UniqueAdmins24h = len(uniqueAdmins)

	// Get top 5 actions
	type actionCount struct {
		action string
		count  int
	}
	var actions []actionCount
	for action, count := range actionCounts {
		actions = append(actions, actionCount{action, count})
	}

	// Simple bubble sort for top actions
	for i := 0; i < len(actions); i++ {
		for j := i + 1; j < len(actions); j++ {
			if actions[j].count > actions[i].count {
				actions[i], actions[j] = actions[j], actions[i]
			}
		}
	}

	// Take top 5
	for i := 0; i < len(actions) && i < 5; i++ {
		stats.TopActions[actions[i].action] = actions[i].count
	}

	return stats
}

// AuditStats represents audit statistics
type AuditStats struct {
	TotalActions24h    int            `json:"totalActions24h"`
	UniqueAdmins24h    int            `json:"uniqueAdmins24h"`
	TopActions         map[string]int `json:"topActions"`
	FailedAttempts24h  int            `json:"failedAttempts24h"`
}

// Count returns total number of audit entries
func (al *AuditLogger) Count() int {
	al.mu.RLock()
	defer al.mu.RUnlock()
	return len(al.entries)
}
