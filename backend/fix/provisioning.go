package fix

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// ProvisioningService manages FIX API provisioning for users
type ProvisioningService struct {
	credentialStore *CredentialStore
	rulesEngine     *RulesEngine
	rateLimiter     *RateLimiter
	activeSessions  map[string][]*ActiveFIXSession // userID -> sessions
	sessionMu       sync.RWMutex
	auditLogger     AuditLogger
}

// ActiveFIXSession tracks active FIX sessions
type ActiveFIXSession struct {
	SessionID    string    `json:"session_id"`
	UserID       string    `json:"user_id"`
	SenderCompID string    `json:"sender_comp_id"`
	IPAddress    string    `json:"ip_address"`
	ConnectedAt  time.Time `json:"connected_at"`
	LastActivity time.Time `json:"last_activity"`
	MessageCount int       `json:"message_count"`
	OrderCount   int       `json:"order_count"`
}

// ProvisioningRequest represents a request for FIX API access
type ProvisioningRequest struct {
	UserID        string         `json:"user_id"`
	UserContext   *UserContext   `json:"user_context"`
	RateLimitTier string         `json:"rate_limit_tier"`
	MaxSessions   int            `json:"max_sessions"`
	ExpiresIn     *time.Duration `json:"expires_in,omitempty"`
	AllowedIPs    []string       `json:"allowed_ips,omitempty"`
}

// ProvisioningResponse represents the result of provisioning
type ProvisioningResponse struct {
	Success      bool            `json:"success"`
	Credentials  *FIXCredentials `json:"credentials,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty"`
	FailedRules  []string        `json:"failed_rules,omitempty"`
}

// NewProvisioningService creates a new provisioning service
func NewProvisioningService(
	storePath string,
	masterPassword string,
	auditLogger AuditLogger,
) (*ProvisioningService, error) {
	// Initialize credential store
	credStore, err := NewCredentialStore(storePath, masterPassword, auditLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential store: %w", err)
	}

	// Initialize rules engine
	rulesEngine := NewRulesEngine(auditLogger)

	// Initialize rate limiter
	rateLimiter := NewRateLimiter(auditLogger)

	return &ProvisioningService{
		credentialStore: credStore,
		rulesEngine:     rulesEngine,
		rateLimiter:     rateLimiter,
		activeSessions:  make(map[string][]*ActiveFIXSession),
		auditLogger:     auditLogger,
	}, nil
}

// ProvisionAccess provisions FIX API access for a user
func (ps *ProvisioningService) ProvisionAccess(req *ProvisioningRequest) *ProvisioningResponse {
	// Step 1: Evaluate access rules
	ruleResult := ps.rulesEngine.EvaluateAccess(req.UserContext)
	if !ruleResult.Allowed {
		ps.auditLogger.LogCredentialOperation("provision_denied", req.UserID,
			fmt.Sprintf("failed_rules=%v", ruleResult.FailedRules), false)

		return &ProvisioningResponse{
			Success:      false,
			ErrorMessage: "Access denied: does not meet requirements",
			FailedRules:  ruleResult.FailedRules,
		}
	}

	// Step 2: Generate credentials
	creds, err := ps.credentialStore.GenerateCredentials(
		req.UserID,
		req.RateLimitTier,
		req.MaxSessions,
		req.ExpiresIn,
	)
	if err != nil {
		ps.auditLogger.LogCredentialOperation("provision_failed", req.UserID,
			fmt.Sprintf("error=%v", err), false)

		return &ProvisioningResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to generate credentials: %v", err),
		}
	}

	// Step 3: Initialize rate limiting
	if err := ps.rateLimiter.InitializeUser(req.UserID, req.RateLimitTier); err != nil {
		// Rollback credential generation
		_ = ps.credentialStore.RevokeCredentials(req.UserID, "rate limiter initialization failed")

		return &ProvisioningResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to initialize rate limiting: %v", err),
		}
	}

	// Step 4: Set allowed IPs if specified
	if len(req.AllowedIPs) > 0 {
		creds.AllowedIPs = req.AllowedIPs
	}

	ps.auditLogger.LogCredentialOperation("provision_success", req.UserID,
		fmt.Sprintf("sender=%s tier=%s", creds.SenderCompID, req.RateLimitTier), true)

	return &ProvisioningResponse{
		Success:     true,
		Credentials: creds,
	}
}

// ValidateLogin validates a FIX login attempt
func (ps *ProvisioningService) ValidateLogin(senderCompID, password, ipAddress string) (*FIXCredentials, error) {
	// Validate credentials
	creds, err := ps.credentialStore.ValidateCredentials(senderCompID, password)
	if err != nil {
		return nil, err
	}

	// Check IP whitelist if configured
	if len(creds.AllowedIPs) > 0 {
		allowed := false
		for _, allowedIP := range creds.AllowedIPs {
			if allowedIP == ipAddress {
				allowed = true
				break
			}
		}
		if !allowed {
			ps.auditLogger.LogCredentialOperation("login_denied", creds.UserID,
				fmt.Sprintf("ip=%s not in whitelist", ipAddress), false)
			return nil, fmt.Errorf("IP address %s not allowed", ipAddress)
		}
	}

	// Check session limit
	canConnect, err := ps.rateLimiter.CheckSessionLimit(creds.UserID)
	if err != nil || !canConnect {
		ps.auditLogger.LogCredentialOperation("login_denied", creds.UserID,
			"session limit exceeded", false)
		return nil, errors.New("session limit exceeded")
	}

	return creds, nil
}

// RegisterSession registers a new FIX session
func (ps *ProvisioningService) RegisterSession(sessionID, userID, senderCompID, ipAddress string) error {
	ps.sessionMu.Lock()
	defer ps.sessionMu.Unlock()

	session := &ActiveFIXSession{
		SessionID:    sessionID,
		UserID:       userID,
		SenderCompID: senderCompID,
		IPAddress:    ipAddress,
		ConnectedAt:  time.Now(),
		LastActivity: time.Now(),
		MessageCount: 0,
		OrderCount:   0,
	}

	ps.activeSessions[userID] = append(ps.activeSessions[userID], session)
	ps.rateLimiter.IncrementSessionCount(userID)

	ps.auditLogger.LogCredentialOperation("session_registered", userID,
		fmt.Sprintf("session=%s ip=%s", sessionID, ipAddress), true)

	return nil
}

// UnregisterSession unregisters a FIX session
func (ps *ProvisioningService) UnregisterSession(sessionID string) error {
	ps.sessionMu.Lock()
	defer ps.sessionMu.Unlock()

	for userID, sessions := range ps.activeSessions {
		for i, session := range sessions {
			if session.SessionID == sessionID {
				// Remove session from slice
				ps.activeSessions[userID] = append(sessions[:i], sessions[i+1:]...)
				ps.rateLimiter.DecrementSessionCount(userID)

				ps.auditLogger.LogCredentialOperation("session_unregistered", userID,
					fmt.Sprintf("session=%s", sessionID), true)

				return nil
			}
		}
	}

	return errors.New("session not found")
}

// TrackMessage tracks a message sent by a user
func (ps *ProvisioningService) TrackMessage(userID, sessionID string, isOrder bool) error {
	// Check message rate limit
	allowed, err := ps.rateLimiter.CheckMessageLimit(userID)
	if err != nil || !allowed {
		return errors.New("message rate limit exceeded")
	}

	// If it's an order, check order rate limit too
	if isOrder {
		allowed, err := ps.rateLimiter.CheckOrderLimit(userID)
		if err != nil || !allowed {
			return errors.New("order rate limit exceeded")
		}
	}

	// Update session stats
	ps.sessionMu.Lock()
	defer ps.sessionMu.Unlock()

	if sessions, exists := ps.activeSessions[userID]; exists {
		for _, session := range sessions {
			if session.SessionID == sessionID {
				session.MessageCount++
				session.LastActivity = time.Now()
				if isOrder {
					session.OrderCount++
				}
				break
			}
		}
	}

	return nil
}

// GetUserSessions returns all active sessions for a user
func (ps *ProvisioningService) GetUserSessions(userID string) []*ActiveFIXSession {
	ps.sessionMu.RLock()
	defer ps.sessionMu.RUnlock()

	return ps.activeSessions[userID]
}

// GetAllSessions returns all active sessions
func (ps *ProvisioningService) GetAllSessions() map[string][]*ActiveFIXSession {
	ps.sessionMu.RLock()
	defer ps.sessionMu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string][]*ActiveFIXSession)
	for userID, sessions := range ps.activeSessions {
		result[userID] = append([]*ActiveFIXSession{}, sessions...)
	}

	return result
}

// KillSession forcefully terminates a session
func (ps *ProvisioningService) KillSession(sessionID string) error {
	ps.auditLogger.LogCredentialOperation("session_kill", "admin",
		fmt.Sprintf("session=%s", sessionID), true)

	return ps.UnregisterSession(sessionID)
}

// RevokeUserAccess revokes all FIX access for a user
func (ps *ProvisioningService) RevokeUserAccess(userID, reason string) error {
	// Revoke credentials
	if err := ps.credentialStore.RevokeCredentials(userID, reason); err != nil {
		return err
	}

	// Kill all active sessions
	ps.sessionMu.Lock()
	sessionIDs := make([]string, 0)
	if sessions, exists := ps.activeSessions[userID]; exists {
		for _, session := range sessions {
			sessionIDs = append(sessionIDs, session.SessionID)
		}
	}
	ps.sessionMu.Unlock()

	for _, sessionID := range sessionIDs {
		_ = ps.KillSession(sessionID)
	}

	ps.auditLogger.LogCredentialOperation("access_revoked", userID, reason, true)
	return nil
}

// GetCredentialStore returns the credential store (for admin operations)
func (ps *ProvisioningService) GetCredentialStore() *CredentialStore {
	return ps.credentialStore
}

// GetRulesEngine returns the rules engine (for admin operations)
func (ps *ProvisioningService) GetRulesEngine() *RulesEngine {
	return ps.rulesEngine
}

// GetRateLimiter returns the rate limiter (for admin operations)
func (ps *ProvisioningService) GetRateLimiter() *RateLimiter {
	return ps.rateLimiter
}

// SimpleAuditLogger is a basic implementation of AuditLogger
type SimpleAuditLogger struct {
	mu sync.Mutex
}

// LogCredentialOperation logs credential operations
func (sal *SimpleAuditLogger) LogCredentialOperation(operation, userID, details string, success bool) {
	sal.mu.Lock()
	defer sal.mu.Unlock()

	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}

	log.Printf("[AUDIT] [%s] operation=%s user=%s details=%s",
		status, operation, userID, details)
}

// NewSimpleAuditLogger creates a basic audit logger
func NewSimpleAuditLogger() *SimpleAuditLogger {
	return &SimpleAuditLogger{}
}
