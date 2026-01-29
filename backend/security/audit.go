package security

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditLevel defines the severity of audit events
type AuditLevel string

const (
	AuditLevelInfo     AuditLevel = "INFO"
	AuditLevelWarning  AuditLevel = "WARNING"
	AuditLevelCritical AuditLevel = "CRITICAL"
	AuditLevelSecurity AuditLevel = "SECURITY"
)

// AuditEvent represents a security audit event
type AuditEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       AuditLevel             `json:"level"`
	Category    string                 `json:"category"`
	Action      string                 `json:"action"`
	UserID      string                 `json:"user_id,omitempty"`
	IP          string                 `json:"ip,omitempty"`
	Resource    string                 `json:"resource,omitempty"`
	Success     bool                   `json:"success"`
	Message     string                 `json:"message"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AuditLogger handles security audit logging
type AuditLogger struct {
	logDir      string
	logFile     *os.File
	mu          sync.Mutex
	rotateSize  int64  // Rotate log when it reaches this size
	retainDays  int    // Number of days to retain logs
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logDir string) (*AuditLogger, error) {
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %w", err)
	}

	logger := &AuditLogger{
		logDir:     logDir,
		rotateSize: 100 * 1024 * 1024, // 100MB
		retainDays: 90,                 // 90 days retention
	}

	if err := logger.openLogFile(); err != nil {
		return nil, err
	}

	// Start log rotation and cleanup routine
	go logger.rotationRoutine()

	return logger, nil
}

// openLogFile opens or creates the current log file
func (a *AuditLogger) openLogFile() error {
	filename := filepath.Join(a.logDir, fmt.Sprintf("audit_%s.log", time.Now().Format("2006-01-02")))

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %w", err)
	}

	a.logFile = file
	return nil
}

// Log records an audit event
func (a *AuditLogger) Log(event AuditEvent) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	event.Timestamp = time.Now()

	// Serialize event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	// Write to file
	if _, err := a.logFile.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write audit event: %w", err)
	}

	// Also log to standard logger for critical events
	if event.Level == AuditLevelCritical || event.Level == AuditLevelSecurity {
		log.Printf("[AUDIT-%s] %s: %s", event.Level, event.Action, event.Message)
	}

	return nil
}

// LogAuthAttempt logs authentication attempts
func (a *AuditLogger) LogAuthAttempt(userID, ip string, success bool, message string) {
	a.Log(AuditEvent{
		Level:    AuditLevelSecurity,
		Category: "authentication",
		Action:   "login_attempt",
		UserID:   userID,
		IP:       ip,
		Success:  success,
		Message:  message,
	})
}

// LogTradeExecution logs trade execution events
func (a *AuditLogger) LogTradeExecution(userID, symbol string, volume float64, success bool, message string) {
	a.Log(AuditEvent{
		Level:    AuditLevelInfo,
		Category: "trading",
		Action:   "trade_execution",
		UserID:   userID,
		Resource: symbol,
		Success:  success,
		Message:  message,
		Metadata: map[string]interface{}{
			"volume": volume,
		},
	})
}

// LogAdminAction logs administrative actions
func (a *AuditLogger) LogAdminAction(adminID, ip, action, resource string, success bool, metadata map[string]interface{}) {
	a.Log(AuditEvent{
		Level:    AuditLevelCritical,
		Category: "admin",
		Action:   action,
		UserID:   adminID,
		IP:       ip,
		Resource: resource,
		Success:  success,
		Message:  fmt.Sprintf("Admin action: %s on %s", action, resource),
		Metadata: metadata,
	})
}

// LogSecurityIncident logs security incidents
func (a *AuditLogger) LogSecurityIncident(category, action, ip, message string, metadata map[string]interface{}) {
	a.Log(AuditEvent{
		Level:    AuditLevelSecurity,
		Category: category,
		Action:   action,
		IP:       ip,
		Success:  false,
		Message:  message,
		Metadata: metadata,
	})
}

// LogDataAccess logs sensitive data access
func (a *AuditLogger) LogDataAccess(userID, ip, resource string, success bool) {
	a.Log(AuditEvent{
		Level:    AuditLevelWarning,
		Category: "data_access",
		Action:   "sensitive_data_access",
		UserID:   userID,
		IP:       ip,
		Resource: resource,
		Success:  success,
		Message:  fmt.Sprintf("User %s accessed %s", userID, resource),
	})
}

// LogConfigChange logs configuration changes
func (a *AuditLogger) LogConfigChange(adminID, ip, setting, oldValue, newValue string) {
	a.Log(AuditEvent{
		Level:    AuditLevelCritical,
		Category: "configuration",
		Action:   "config_change",
		UserID:   adminID,
		IP:       ip,
		Resource: setting,
		Success:  true,
		Message:  fmt.Sprintf("Configuration changed: %s", setting),
		Metadata: map[string]interface{}{
			"old_value": oldValue,
			"new_value": newValue,
		},
	})
}

// LogAPIKeyRotation logs API key rotation events
func (a *AuditLogger) LogAPIKeyRotation(service, reason string, success bool) {
	a.Log(AuditEvent{
		Level:    AuditLevelSecurity,
		Category: "api_keys",
		Action:   "key_rotation",
		Resource: service,
		Success:  success,
		Message:  fmt.Sprintf("API key rotated for %s: %s", service, reason),
	})
}

// rotationRoutine periodically checks if log rotation is needed
func (a *AuditLogger) rotationRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		a.rotateIfNeeded()
		a.cleanupOldLogs()
	}
}

// rotateIfNeeded rotates the log file if it exceeds the size limit
func (a *AuditLogger) rotateIfNeeded() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.logFile == nil {
		return
	}

	// Check file size
	info, err := a.logFile.Stat()
	if err != nil {
		log.Printf("[AUDIT] Failed to stat log file: %v", err)
		return
	}

	if info.Size() >= a.rotateSize {
		// Close current file
		a.logFile.Close()

		// Rename with timestamp
		oldName := a.logFile.Name()
		newName := fmt.Sprintf("%s.%d", oldName, time.Now().Unix())
		os.Rename(oldName, newName)

		// Open new file
		if err := a.openLogFile(); err != nil {
			log.Printf("[AUDIT] Failed to rotate log file: %v", err)
		}
	}
}

// cleanupOldLogs removes audit logs older than the retention period
func (a *AuditLogger) cleanupOldLogs() {
	cutoff := time.Now().AddDate(0, 0, -a.retainDays)

	filepath.Walk(a.logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() && info.ModTime().Before(cutoff) {
			log.Printf("[AUDIT] Removing old audit log: %s", path)
			os.Remove(path)
		}

		return nil
	})
}

// Close closes the audit logger
func (a *AuditLogger) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.logFile != nil {
		return a.logFile.Close()
	}

	return nil
}

// QueryLogs searches audit logs for specific criteria
func (a *AuditLogger) QueryLogs(startTime, endTime time.Time, level AuditLevel, category string) ([]AuditEvent, error) {
	var events []AuditEvent

	err := filepath.Walk(a.logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// Read log file
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Parse JSON lines
		lines := string(data)
		for _, line := range splitLines(lines) {
			if line == "" {
				continue
			}

			var event AuditEvent
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				continue
			}

			// Apply filters
			if !event.Timestamp.After(startTime) || !event.Timestamp.Before(endTime) {
				continue
			}

			if level != "" && event.Level != level {
				continue
			}

			if category != "" && event.Category != category {
				continue
			}

			events = append(events, event)
		}

		return nil
	})

	return events, err
}

// Helper function to split lines
func splitLines(s string) []string {
	var lines []string
	start := 0

	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}

	if start < len(s) {
		lines = append(lines, s[start:])
	}

	return lines
}
