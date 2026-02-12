package admin

// INTEGRATION NOTES:
//
// 1. LOGIN FLOW INTEGRATION (auth.go):
//    - In Login() method (after JWT token generation):
//      token := generateJWT(admin.ID, admin.Role, cfg.JWT.Secret)
//      sessionStore.CreateSession(strconv.FormatInt(admin.ID, 10), token, ipAddress, userAgent)
//
// 2. AUTH MIDDLEWARE INTEGRATION:
//    - After JWT validation, check session validity:
//      tokenHash := hashToken(jwtToken)
//      if !sessionStore.IsSessionValid(tokenHash) {
//          return errors.New("session revoked or expired")
//      }
//      sessionStore.UpdateLastActive(tokenHash)
//
// 3. LOGOUT FLOW:
//    - RevokeSession(sessionID, adminID) when admin logs out
//
// 4. ENV VAR CONFIGURATION:
//    - SESSION_MAX_CONCURRENT (default: 3) - max concurrent sessions per user

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

// Session represents an admin session with revocation support
type Session struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`     // Admin ID as string
	TokenHash   string    `json:"-"`          // SHA-256 hash of JWT token (hidden from JSON)
	IP          string    `json:"ip"`
	UserAgent   string    `json:"userAgent"`
	CreatedAt   time.Time `json:"createdAt"`
	LastActive  time.Time `json:"lastActive"`
	IsRevoked   bool      `json:"isRevoked"`
	RevokedAt   *time.Time `json:"revokedAt,omitempty"`
	RevokedBy   string    `json:"revokedBy,omitempty"`
}

// SessionStore manages admin sessions with revocation and limits
type SessionStore struct {
	mu                 sync.RWMutex
	sessions           map[string]*Session  // sessionID -> Session
	userSessions       map[string][]string  // userID -> []sessionID
	tokenHashToSession map[string]string    // tokenHash -> sessionID
	maxConcurrent      int                  // Max concurrent sessions per user
	cleanupTicker      *time.Ticker
	stopCleanup        chan bool
}

// NewSessionStore creates a new session store
func NewSessionStore() *SessionStore {
	// Read max concurrent sessions from env (default 3)
	maxConcurrent := 3
	if maxStr := os.Getenv("SESSION_MAX_CONCURRENT"); maxStr != "" {
		if max, err := strconv.Atoi(maxStr); err == nil && max > 0 {
			maxConcurrent = max
		}
	}

	store := &SessionStore{
		sessions:           make(map[string]*Session),
		userSessions:       make(map[string][]string),
		tokenHashToSession: make(map[string]string),
		maxConcurrent:      maxConcurrent,
		stopCleanup:        make(chan bool),
	}

	// Start cleanup goroutine (runs every 30 minutes)
	store.startCleanup()

	log.Printf("[SessionStore] Initialized with max %d concurrent sessions per user", maxConcurrent)
	return store
}

// CreateSession creates a new session for a user
// Automatically revokes oldest session if concurrent limit exceeded
func (s *SessionStore) CreateSession(userID string, token string, ip string, userAgent string) (*Session, error) {
	if userID == "" || token == "" {
		return nil, errors.New("userID and token are required")
	}

	// Generate session ID and token hash
	sessionID := generateSessionID()
	tokenHash := hashToken(token)

	session := &Session{
		ID:         sessionID,
		UserID:     userID,
		TokenHash:  tokenHash,
		IP:         ip,
		UserAgent:  userAgent,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		IsRevoked:  false,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Add to sessions map
	s.sessions[sessionID] = session
	s.tokenHashToSession[tokenHash] = sessionID

	// Add to user sessions
	if s.userSessions[userID] == nil {
		s.userSessions[userID] = []string{}
	}
	s.userSessions[userID] = append(s.userSessions[userID], sessionID)

	// Check concurrent session limit
	activeSessions := s.getActiveSessionsForUser(userID)
	if len(activeSessions) > s.maxConcurrent {
		// Find oldest session and revoke it
		var oldestSession *Session
		var oldestSessionID string

		for _, sid := range activeSessions {
			session := s.sessions[sid]
			if session != nil && !session.IsRevoked {
				if oldestSession == nil || session.CreatedAt.Before(oldestSession.CreatedAt) {
					oldestSession = session
					oldestSessionID = sid
				}
			}
		}

		if oldestSession != nil {
			now := time.Now()
			oldestSession.IsRevoked = true
			oldestSession.RevokedAt = &now
			oldestSession.RevokedBy = "SYSTEM_AUTO_LIMIT"
			log.Printf("[SessionStore] Auto-revoked oldest session %s for user %s (limit: %d)", oldestSessionID, userID, s.maxConcurrent)
		}
	}

	log.Printf("[SessionStore] Session created: %s for user %s from %s", sessionID, userID, ip)
	return session, nil
}

// GetUserSessions returns all sessions for a user (active and revoked)
func (s *SessionStore) GetUserSessions(userID string) []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessionIDs := s.userSessions[userID]
	if sessionIDs == nil {
		return []*Session{}
	}

	sessions := make([]*Session, 0, len(sessionIDs))
	for _, sid := range sessionIDs {
		if session, exists := s.sessions[sid]; exists {
			// Return copy to prevent external modification
			sessionCopy := *session
			sessions = append(sessions, &sessionCopy)
		}
	}

	// Sort by created date descending (newest first)
	for i := 0; i < len(sessions)-1; i++ {
		for j := i + 1; j < len(sessions); j++ {
			if sessions[i].CreatedAt.Before(sessions[j].CreatedAt) {
				sessions[i], sessions[j] = sessions[j], sessions[i]
			}
		}
	}

	return sessions
}

// GetActiveUserSessions returns only active (not revoked) sessions for a user
func (s *SessionStore) GetActiveUserSessions(userID string) []*Session {
	allSessions := s.GetUserSessions(userID)
	
	activeSessions := make([]*Session, 0)
	for _, session := range allSessions {
		if !session.IsRevoked {
			activeSessions = append(activeSessions, session)
		}
	}

	return activeSessions
}

// RevokeSession marks a session as revoked
func (s *SessionStore) RevokeSession(sessionID string, revokedBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return errors.New("session not found")
	}

	if session.IsRevoked {
		return errors.New("session already revoked")
	}

	now := time.Now()
	session.IsRevoked = true
	session.RevokedAt = &now
	session.RevokedBy = revokedBy

	log.Printf("[SessionStore] Session revoked: %s by %s", sessionID, revokedBy)
	return nil
}

// RevokeAllUserSessions revokes all sessions for a user (force logout everywhere)
func (s *SessionStore) RevokeAllUserSessions(userID string, revokedBy string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionIDs := s.userSessions[userID]
	if sessionIDs == nil {
		return 0
	}

	count := 0
	now := time.Now()

	for _, sid := range sessionIDs {
		session := s.sessions[sid]
		if session != nil && !session.IsRevoked {
			session.IsRevoked = true
			session.RevokedAt = &now
			session.RevokedBy = revokedBy
			count++
		}
	}

	log.Printf("[SessionStore] Revoked %d sessions for user %s by %s", count, userID, revokedBy)
	return count
}

// IsSessionValid checks if a session is valid (exists, not revoked, not expired)
func (s *SessionStore) IsSessionValid(tokenHash string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessionID, exists := s.tokenHashToSession[tokenHash]
	if !exists {
		return false
	}

	session := s.sessions[sessionID]
	if session == nil {
		return false
	}

	// Check if revoked
	if session.IsRevoked {
		return false
	}

	// Check if inactive for more than 24 hours
	if time.Since(session.LastActive) > 24*time.Hour {
		return false
	}

	return true
}

// UpdateLastActive updates the last active timestamp for a session
func (s *SessionStore) UpdateLastActive(tokenHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID, exists := s.tokenHashToSession[tokenHash]
	if !exists {
		return errors.New("session not found")
	}

	session := s.sessions[sessionID]
	if session == nil {
		return errors.New("session not found")
	}

	session.LastActive = time.Now()
	return nil
}

// GetSession retrieves a session by ID
func (s *SessionStore) GetSession(sessionID string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}

	// Return copy
	sessionCopy := *session
	return &sessionCopy, nil
}

// GetStats returns session statistics
func (s *SessionStore) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalSessions := len(s.sessions)
	activeCount := 0
	revokedCount := 0
	inactiveCount := 0

	now := time.Now()
	for _, session := range s.sessions {
		if session.IsRevoked {
			revokedCount++
		} else if now.Sub(session.LastActive) > 24*time.Hour {
			inactiveCount++
		} else {
			activeCount++
		}
	}

	return map[string]interface{}{
		"total":    totalSessions,
		"active":   activeCount,
		"revoked":  revokedCount,
		"inactive": inactiveCount,
		"maxConcurrentPerUser": s.maxConcurrent,
	}
}

// Cleanup removes sessions inactive for more than 24 hours
func (s *SessionStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	removed := 0

	for sessionID, session := range s.sessions {
		// Remove if inactive for more than 24 hours
		if now.Sub(session.LastActive) > 24*time.Hour {
			// Remove from tokenHash map
			delete(s.tokenHashToSession, session.TokenHash)

			// Remove from user sessions
			if userSessionIDs := s.userSessions[session.UserID]; userSessionIDs != nil {
				filtered := []string{}
				for _, sid := range userSessionIDs {
					if sid != sessionID {
						filtered = append(filtered, sid)
					}
				}
				s.userSessions[session.UserID] = filtered
			}

			// Remove from sessions map
			delete(s.sessions, sessionID)
			removed++
		}
	}

	if removed > 0 {
		log.Printf("[SessionStore] Cleanup: removed %d inactive sessions (>24h)", removed)
	}
}

// startCleanup starts the background cleanup goroutine
func (s *SessionStore) startCleanup() {
	s.cleanupTicker = time.NewTicker(30 * time.Minute)

	go func() {
		for {
			select {
			case <-s.cleanupTicker.C:
				s.cleanup()
			case <-s.stopCleanup:
				s.cleanupTicker.Stop()
				return
			}
		}
	}()

	log.Println("[SessionStore] Cleanup goroutine started (runs every 30 minutes)")
}

// Stop stops the cleanup goroutine
func (s *SessionStore) Stop() {
	close(s.stopCleanup)
	log.Println("[SessionStore] Cleanup goroutine stopped")
}

// Helper functions

// getActiveSessionsForUser returns active session IDs for a user (internal, must hold lock)
func (s *SessionStore) getActiveSessionsForUser(userID string) []string {
	sessionIDs := s.userSessions[userID]
	if sessionIDs == nil {
		return []string{}
	}

	active := []string{}
	for _, sid := range sessionIDs {
		session := s.sessions[sid]
		if session != nil && !session.IsRevoked {
			active = append(active, sid)
		}
	}

	return active
}

// hashToken creates a SHA-256 hash of the token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// generateSessionID generates a random session ID
func generateSessionID() string {
	// Reuse the existing generateSessionToken function from auth.go
	// Or implement a simple one here
	return "sess_" + time.Now().Format("20060102150405") + "_" + randomString(16)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
