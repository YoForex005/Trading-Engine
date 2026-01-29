package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

// SessionConfig holds session management configuration
type SessionConfig struct {
	Timeout            time.Duration
	MaxSessions        int
	MaxConcurrentPerUser int
	RequireIP          bool
}

// DefaultSessionConfig returns secure default session configuration
func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		Timeout:            30 * time.Minute,
		MaxSessions:        10000,
		MaxConcurrentPerUser: 3,
		RequireIP:          true,
	}
}

// Session represents a user session
type Session struct {
	ID         string
	UserID     string
	IP         string
	CreatedAt  time.Time
	LastAccess time.Time
	Data       map[string]interface{}
}

// SessionManager handles user session management
type SessionManager struct {
	config      *SessionConfig
	sessions    map[string]*Session
	userSessions map[string][]string // userID -> sessionIDs
	mu          sync.RWMutex
	auditLogger *AuditLogger
}

// NewSessionManager creates a new session manager
func NewSessionManager(config *SessionConfig, auditLogger *AuditLogger) *SessionManager {
	if config == nil {
		config = DefaultSessionConfig()
	}

	sm := &SessionManager{
		config:      config,
		sessions:    make(map[string]*Session),
		userSessions: make(map[string][]string),
		auditLogger: auditLogger,
	}

	// Start cleanup routine
	go sm.cleanupRoutine()

	return sm
}

// CreateSession creates a new session for a user
func (sm *SessionManager) CreateSession(userID, ip string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check max concurrent sessions per user
	if userSessions, exists := sm.userSessions[userID]; exists {
		if len(userSessions) >= sm.config.MaxConcurrentPerUser {
			// Remove oldest session
			sm.removeOldestSession(userID)
		}
	}

	// Check global session limit
	if len(sm.sessions) >= sm.config.MaxSessions {
		return nil, errors.New("maximum sessions reached")
	}

	// Generate session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:         sessionID,
		UserID:     userID,
		IP:         ip,
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
		Data:       make(map[string]interface{}),
	}

	// Store session
	sm.sessions[sessionID] = session
	sm.userSessions[userID] = append(sm.userSessions[userID], sessionID)

	// Log session creation
	if sm.auditLogger != nil {
		sm.auditLogger.Log(AuditEvent{
			Level:    AuditLevelInfo,
			Category: "session",
			Action:   "session_created",
			UserID:   userID,
			IP:       ip,
			Success:  true,
			Message:  "New session created",
		})
	}

	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}

	// Check if session expired
	if time.Since(session.LastAccess) > sm.config.Timeout {
		return nil, errors.New("session expired")
	}

	return session, nil
}

// ValidateSession validates a session and updates last access time
func (sm *SessionManager) ValidateSession(sessionID, ip string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}

	// Check if session expired
	if time.Since(session.LastAccess) > sm.config.Timeout {
		sm.destroySessionInternal(sessionID)
		return nil, errors.New("session expired")
	}

	// Check IP if required
	if sm.config.RequireIP && session.IP != ip {
		sm.auditLogger.LogSecurityIncident(
			"session",
			"ip_mismatch",
			ip,
			"Session IP mismatch detected",
			map[string]interface{}{
				"session_ip": session.IP,
				"request_ip": ip,
				"user_id":    session.UserID,
			},
		)
		return nil, errors.New("session IP mismatch")
	}

	// Update last access
	session.LastAccess = time.Now()

	return session, nil
}

// DestroySession destroys a session
func (sm *SessionManager) DestroySession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	return sm.destroySessionInternal(sessionID)
}

// destroySessionInternal destroys a session (internal, no lock)
func (sm *SessionManager) destroySessionInternal(sessionID string) error {
	session, exists := sm.sessions[sessionID]
	if !exists {
		return errors.New("session not found")
	}

	// Remove from sessions map
	delete(sm.sessions, sessionID)

	// Remove from user sessions
	if userSessions, exists := sm.userSessions[session.UserID]; exists {
		for i, sid := range userSessions {
			if sid == sessionID {
				sm.userSessions[session.UserID] = append(userSessions[:i], userSessions[i+1:]...)
				break
			}
		}
	}

	// Log session destruction
	if sm.auditLogger != nil {
		sm.auditLogger.Log(AuditEvent{
			Level:    AuditLevelInfo,
			Category: "session",
			Action:   "session_destroyed",
			UserID:   session.UserID,
			Success:  true,
			Message:  "Session destroyed",
		})
	}

	return nil
}

// DestroyUserSessions destroys all sessions for a user
func (sm *SessionManager) DestroyUserSessions(userID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionIDs, exists := sm.userSessions[userID]
	if !exists {
		return nil
	}

	for _, sessionID := range sessionIDs {
		delete(sm.sessions, sessionID)
	}

	delete(sm.userSessions, userID)

	// Log session destruction
	if sm.auditLogger != nil {
		sm.auditLogger.Log(AuditEvent{
			Level:    AuditLevelWarning,
			Category: "session",
			Action:   "all_sessions_destroyed",
			UserID:   userID,
			Success:  true,
			Message:  "All user sessions destroyed",
		})
	}

	return nil
}

// removeOldestSession removes the oldest session for a user
func (sm *SessionManager) removeOldestSession(userID string) {
	sessionIDs := sm.userSessions[userID]
	if len(sessionIDs) == 0 {
		return
	}

	// Find oldest session
	oldestID := sessionIDs[0]
	oldestTime := sm.sessions[oldestID].CreatedAt

	for _, sid := range sessionIDs[1:] {
		if sm.sessions[sid].CreatedAt.Before(oldestTime) {
			oldestID = sid
			oldestTime = sm.sessions[sid].CreatedAt
		}
	}

	sm.destroySessionInternal(oldestID)
}

// cleanupRoutine periodically removes expired sessions
func (sm *SessionManager) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.cleanup()
	}
}

// cleanup removes expired sessions
func (sm *SessionManager) cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	expiredSessions := make([]string, 0)

	for sessionID, session := range sm.sessions {
		if now.Sub(session.LastAccess) > sm.config.Timeout {
			expiredSessions = append(expiredSessions, sessionID)
		}
	}

	for _, sessionID := range expiredSessions {
		sm.destroySessionInternal(sessionID)
	}
}

// GetActiveSessions returns the number of active sessions
func (sm *SessionManager) GetActiveSessions() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.sessions)
}

// GetUserSessions returns the number of sessions for a user
func (sm *SessionManager) GetUserSessions(userID string) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sessions, exists := sm.userSessions[userID]; exists {
		return len(sessions)
	}

	return 0
}

// generateSessionID generates a cryptographically secure session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
