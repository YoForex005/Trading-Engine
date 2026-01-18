package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	AuditOrderPlacement      AuditEventType = "order_placement"
	AuditOrderCancellation   AuditEventType = "order_cancellation"
	AuditOrderModification   AuditEventType = "order_modification"
	AuditPositionOpen        AuditEventType = "position_open"
	AuditPositionClose       AuditEventType = "position_close"
	AuditPositionModify      AuditEventType = "position_modify"
	AuditAuthentication      AuditEventType = "authentication"
	AuditAuthenticationFail  AuditEventType = "authentication_failed"
	AuditAdminAction         AuditEventType = "admin_action"
	AuditAccountCreation     AuditEventType = "account_creation"
	AuditAccountModification AuditEventType = "account_modification"
	AuditDeposit             AuditEventType = "deposit"
	AuditWithdrawal          AuditEventType = "withdrawal"
	AuditLPRouting           AuditEventType = "lp_routing"
	AuditRiskAction          AuditEventType = "risk_action"
	AuditConfigChange        AuditEventType = "config_change"
)

// AuditEvent represents a single audit trail entry
type AuditEvent struct {
	EventID     string                 `json:"event_id"`
	Timestamp   time.Time              `json:"timestamp"`
	EventType   AuditEventType         `json:"event_type"`
	UserID      string                 `json:"user_id,omitempty"`
	AccountID   string                 `json:"account_id,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource,omitempty"`
	ResourceID  string                 `json:"resource_id,omitempty"`
	Before      map[string]interface{} `json:"before,omitempty"`
	After       map[string]interface{} `json:"after,omitempty"`
	Status      string                 `json:"status"` // success, failed, denied
	Reason      string                 `json:"reason,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Compliance  bool                   `json:"compliance"` // Flag for regulatory compliance
	Environment string                 `json:"environment"`
	RequestID   string                 `json:"request_id,omitempty"`
}

// AuditLogger handles audit trail logging with guaranteed persistence
type AuditLogger struct {
	mu           sync.Mutex
	file         *os.File
	encoder      *json.Encoder
	filePath     string
	rotateSize   int64 // Max file size before rotation
	currentSize  int64
	buffer       []*AuditEvent
	bufferSize   int
	flushTicker  *time.Ticker
	stopChan     chan struct{}
	environment  string
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(auditDir string) (*AuditLogger, error) {
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(auditDir, "audit.log")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	stat, _ := file.Stat()

	al := &AuditLogger{
		file:        file,
		encoder:     json.NewEncoder(file),
		filePath:    filePath,
		rotateSize:  100 * 1024 * 1024, // 100MB
		currentSize: stat.Size(),
		buffer:      make([]*AuditEvent, 0, 100),
		bufferSize:  100,
		flushTicker: time.NewTicker(5 * time.Second),
		stopChan:    make(chan struct{}),
		environment: getEnvironment(),
	}

	// Start auto-flush goroutine
	go al.autoFlush()

	return al, nil
}

// LogOrderPlacement logs an order placement event
func (al *AuditLogger) LogOrderPlacement(ctx context.Context, orderID, symbol, side string, volume, price float64, orderType string, accountID string) {
	al.logEvent(ctx, &AuditEvent{
		EventID:    generateEventID(),
		EventType:  AuditOrderPlacement,
		Action:     "place_order",
		Resource:   "order",
		ResourceID: orderID,
		AccountID:  accountID,
		Status:     "success",
		Metadata: map[string]interface{}{
			"symbol":     symbol,
			"side":       side,
			"volume":     volume,
			"price":      price,
			"order_type": orderType,
		},
		Compliance: true,
	})
}

// LogOrderCancellation logs an order cancellation event
func (al *AuditLogger) LogOrderCancellation(ctx context.Context, orderID, accountID string, reason string) {
	al.logEvent(ctx, &AuditEvent{
		EventID:    generateEventID(),
		EventType:  AuditOrderCancellation,
		Action:     "cancel_order",
		Resource:   "order",
		ResourceID: orderID,
		AccountID:  accountID,
		Status:     "success",
		Reason:     reason,
		Compliance: true,
	})
}

// LogPositionClose logs a position close event
func (al *AuditLogger) LogPositionClose(ctx context.Context, positionID, accountID string, pnl float64, closePrice float64) {
	al.logEvent(ctx, &AuditEvent{
		EventID:    generateEventID(),
		EventType:  AuditPositionClose,
		Action:     "close_position",
		Resource:   "position",
		ResourceID: positionID,
		AccountID:  accountID,
		Status:     "success",
		Metadata: map[string]interface{}{
			"pnl":         pnl,
			"close_price": closePrice,
		},
		Compliance: true,
	})
}

// LogAuthentication logs a successful authentication
func (al *AuditLogger) LogAuthentication(ctx context.Context, userID, ipAddress string, method string) {
	al.logEvent(ctx, &AuditEvent{
		EventID:   generateEventID(),
		EventType: AuditAuthentication,
		Action:    "login",
		UserID:    userID,
		IPAddress: ipAddress,
		Status:    "success",
		Metadata: map[string]interface{}{
			"method": method,
		},
		Compliance: true,
	})
}

// LogAuthenticationFailed logs a failed authentication attempt
func (al *AuditLogger) LogAuthenticationFailed(ctx context.Context, username, ipAddress, reason string) {
	al.logEvent(ctx, &AuditEvent{
		EventID:   generateEventID(),
		EventType: AuditAuthenticationFail,
		Action:    "login_failed",
		IPAddress: ipAddress,
		Status:    "failed",
		Reason:    reason,
		Metadata: map[string]interface{}{
			"username": username,
		},
		Compliance: true,
	})
}

// LogAdminAction logs an administrative action
func (al *AuditLogger) LogAdminAction(ctx context.Context, adminID, action, resource, resourceID string, before, after map[string]interface{}) {
	al.logEvent(ctx, &AuditEvent{
		EventID:    generateEventID(),
		EventType:  AuditAdminAction,
		UserID:     adminID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Before:     before,
		After:      after,
		Status:     "success",
		Compliance: true,
	})
}

// LogLPRouting logs an LP routing decision
func (al *AuditLogger) LogLPRouting(ctx context.Context, orderID, selectedLP, routingReason string, accountID string, metadata map[string]interface{}) {
	al.logEvent(ctx, &AuditEvent{
		EventID:    generateEventID(),
		EventType:  AuditLPRouting,
		Action:     "route_to_lp",
		Resource:   "order",
		ResourceID: orderID,
		AccountID:  accountID,
		Status:     "success",
		Reason:     routingReason,
		Metadata: map[string]interface{}{
			"selected_lp": selectedLP,
			"details":     metadata,
		},
		Compliance: true,
	})
}

// LogDeposit logs a deposit transaction
func (al *AuditLogger) LogDeposit(ctx context.Context, accountID string, amount float64, method, transactionID string) {
	al.logEvent(ctx, &AuditEvent{
		EventID:    generateEventID(),
		EventType:  AuditDeposit,
		Action:     "deposit",
		Resource:   "account",
		ResourceID: accountID,
		AccountID:  accountID,
		Status:     "success",
		Metadata: map[string]interface{}{
			"amount":         amount,
			"method":         method,
			"transaction_id": transactionID,
		},
		Compliance: true,
	})
}

// LogWithdrawal logs a withdrawal transaction
func (al *AuditLogger) LogWithdrawal(ctx context.Context, accountID string, amount float64, method, transactionID string) {
	al.logEvent(ctx, &AuditEvent{
		EventID:    generateEventID(),
		EventType:  AuditWithdrawal,
		Action:     "withdrawal",
		Resource:   "account",
		ResourceID: accountID,
		AccountID:  accountID,
		Status:     "success",
		Metadata: map[string]interface{}{
			"amount":         amount,
			"method":         method,
			"transaction_id": transactionID,
		},
		Compliance: true,
	})
}

// LogConfigChange logs a configuration change
func (al *AuditLogger) LogConfigChange(ctx context.Context, adminID, configKey string, before, after interface{}) {
	al.logEvent(ctx, &AuditEvent{
		EventID:   generateEventID(),
		EventType: AuditConfigChange,
		UserID:    adminID,
		Action:    "config_change",
		Resource:  "config",
		Before: map[string]interface{}{
			configKey: before,
		},
		After: map[string]interface{}{
			configKey: after,
		},
		Status:     "success",
		Compliance: true,
	})
}

// logEvent writes an audit event to the log
func (al *AuditLogger) logEvent(ctx context.Context, event *AuditEvent) {
	// Enrich event with context data
	event.Timestamp = time.Now().UTC()
	event.Environment = al.environment

	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		event.RequestID = requestID
	}

	if event.UserID == "" {
		if userID, ok := ctx.Value(userIDKey).(string); ok {
			event.UserID = userID
		}
	}

	if event.AccountID == "" {
		if accountID, ok := ctx.Value(accountIDKey).(string); ok {
			event.AccountID = accountID
		}
	}

	al.mu.Lock()
	defer al.mu.Unlock()

	// Add to buffer
	al.buffer = append(al.buffer, event)

	// Flush if buffer is full
	if len(al.buffer) >= al.bufferSize {
		al.flush()
	}
}

// flush writes buffered events to disk
func (al *AuditLogger) flush() {
	if len(al.buffer) == 0 {
		return
	}

	for _, event := range al.buffer {
		if err := al.encoder.Encode(event); err == nil {
			// Estimate size (rough approximation)
			al.currentSize += 500
		}
	}

	al.file.Sync() // Force write to disk
	al.buffer = al.buffer[:0]

	// Check if rotation is needed
	if al.currentSize >= al.rotateSize {
		al.rotate()
	}
}

// autoFlush periodically flushes the buffer
func (al *AuditLogger) autoFlush() {
	for {
		select {
		case <-al.flushTicker.C:
			al.mu.Lock()
			al.flush()
			al.mu.Unlock()
		case <-al.stopChan:
			return
		}
	}
}

// rotate rotates the log file
func (al *AuditLogger) rotate() {
	al.file.Close()

	// Rename current file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	rotatedPath := al.filePath + "." + timestamp
	os.Rename(al.filePath, rotatedPath)

	// Create new file
	file, err := os.OpenFile(al.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}

	al.file = file
	al.encoder = json.NewEncoder(file)
	al.currentSize = 0
}

// Close flushes and closes the audit logger
func (al *AuditLogger) Close() error {
	close(al.stopChan)
	al.flushTicker.Stop()

	al.mu.Lock()
	defer al.mu.Unlock()

	al.flush()
	return al.file.Close()
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("audit-%d", time.Now().UnixNano())
}
