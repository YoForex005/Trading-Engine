package admin

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AuthService handles admin authentication and authorization
type AuthService struct {
	mu              sync.RWMutex
	admins          map[int64]*Admin
	adminsByUsername map[string]*Admin
	sessions        map[string]*AdminSession
	nextAdminID     int64
}

// NewAuthService creates a new admin auth service
func NewAuthService() *AuthService {
	svc := &AuthService{
		admins:           make(map[int64]*Admin),
		adminsByUsername: make(map[string]*Admin),
		sessions:         make(map[string]*AdminSession),
		nextAdminID:      1,
	}

	// Create default super admin
	superAdmin, err := svc.CreateAdmin("admin", "admin@rtx.local", "Admin@123", RoleSuperAdmin, nil, "SYSTEM")
	if err != nil {
		log.Printf("[AdminAuth] Failed to create default super admin: %v", err)
	} else {
		log.Printf("[AdminAuth] Default super admin created: %s", superAdmin.Username)
	}

	return svc
}

// CreateAdmin creates a new admin user
func (s *AuthService) CreateAdmin(username, email, password string, role AdminRole, ipWhitelist []string, createdBy string) (*Admin, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if username exists
	if _, exists := s.adminsByUsername[username]; exists {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	admin := &Admin{
		ID:           s.nextAdminID,
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		IPWhitelist:  ipWhitelist,
		Status:       "ACTIVE",
		CreatedAt:    time.Now(),
		CreatedBy:    createdBy,
	}

	s.nextAdminID++
	s.admins[admin.ID] = admin
	s.adminsByUsername[username] = admin

	log.Printf("[AdminAuth] Admin created: %s (%s) by %s", username, role, createdBy)
	return admin, nil
}

// Login authenticates an admin and creates a session
func (s *AuthService) Login(username, password, ipAddress, userAgent string) (*AdminSession, error) {
	s.mu.RLock()
	admin, exists := s.adminsByUsername[username]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		log.Printf("[AdminAuth] Failed login attempt for %s from %s", username, ipAddress)
		return nil, errors.New("invalid credentials")
	}

	// Check admin status
	if admin.Status != "ACTIVE" {
		return nil, fmt.Errorf("admin account is %s", admin.Status)
	}

	// Check IP whitelist
	if len(admin.IPWhitelist) > 0 {
		if !s.isIPWhitelisted(ipAddress, admin.IPWhitelist) {
			log.Printf("[AdminAuth] IP %s not whitelisted for %s", ipAddress, username)
			return nil, errors.New("IP address not authorized")
		}
	}

	// Generate session token
	sessionID, err := generateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session: %w", err)
	}

	// Create session
	now := time.Now()
	session := &AdminSession{
		SessionID:  sessionID,
		AdminID:    admin.ID,
		Username:   admin.Username,
		Role:       admin.Role,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		CreatedAt:  now,
		ExpiresAt:  now.Add(8 * time.Hour), // 8 hour sessions
		LastActive: now,
	}

	s.mu.Lock()
	s.sessions[sessionID] = session
	admin.LastLogin = now
	s.mu.Unlock()

	log.Printf("[AdminAuth] Admin logged in: %s from %s", username, ipAddress)
	return session, nil
}

// ValidateSession validates a session token and returns the admin
func (s *AuthService) ValidateSession(sessionID, ipAddress string) (*Admin, error) {
	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("invalid session")
	}

	// Check expiration
	if time.Now().After(session.ExpiresAt) {
		s.mu.Lock()
		delete(s.sessions, sessionID)
		s.mu.Unlock()
		return nil, errors.New("session expired")
	}

	// Verify IP hasn't changed (optional, configurable)
	if session.IPAddress != ipAddress {
		log.Printf("[AdminAuth] IP mismatch for session %s: %s vs %s", sessionID, session.IPAddress, ipAddress)
		// You can make this configurable - for now, just log
	}

	s.mu.RLock()
	admin, exists := s.admins[session.AdminID]
	s.mu.RUnlock()

	if !exists || admin.Status != "ACTIVE" {
		return nil, errors.New("admin account not active")
	}

	// Update last active
	s.mu.Lock()
	session.LastActive = time.Now()
	s.mu.Unlock()

	return admin, nil
}

// Logout terminates a session
func (s *AuthService) Logout(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.sessions[sessionID]; exists {
		delete(s.sessions, sessionID)
		log.Printf("[AdminAuth] Admin logged out: %s", session.Username)
		return nil
	}

	return errors.New("session not found")
}

// CheckPermission verifies admin has permission for an action
func (s *AuthService) CheckPermission(admin *Admin, action string) bool {
	// Super admin has all permissions
	if admin.Role == RoleSuperAdmin {
		return true
	}

	// Regular admin permissions
	if admin.Role == RoleAdmin {
		switch action {
		case "view_users", "view_orders", "view_funds", "view_groups",
			"modify_user", "fund_deposit", "fund_withdraw", "modify_order",
			"close_position", "modify_group":
			return true
		case "create_admin", "delete_admin", "system_config":
			return false
		default:
			return false
		}
	}

	// Support role - read-only mostly
	if admin.Role == RoleSupport {
		switch action {
		case "view_users", "view_orders", "view_funds", "view_groups":
			return true
		default:
			return false
		}
	}

	return false
}

// UpdatePassword changes an admin's password
func (s *AuthService) UpdatePassword(adminID int64, oldPassword, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	admin, exists := s.admins[adminID]
	if !exists {
		return errors.New("admin not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("incorrect current password")
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	admin.PasswordHash = string(hash)
	log.Printf("[AdminAuth] Password updated for admin: %s", admin.Username)
	return nil
}

// ResetPassword resets an admin's password (super admin only)
func (s *AuthService) ResetPassword(adminID int64, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	admin, exists := s.admins[adminID]
	if !exists {
		return errors.New("admin not found")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	admin.PasswordHash = string(hash)
	log.Printf("[AdminAuth] Password reset for admin: %s", admin.Username)
	return nil
}

// SetIPWhitelist updates IP whitelist for an admin
func (s *AuthService) SetIPWhitelist(adminID int64, ips []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	admin, exists := s.admins[adminID]
	if !exists {
		return errors.New("admin not found")
	}

	admin.IPWhitelist = ips
	log.Printf("[AdminAuth] IP whitelist updated for %s: %v", admin.Username, ips)
	return nil
}

// GetAdmin returns an admin by ID
func (s *AuthService) GetAdmin(adminID int64) (*Admin, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	admin, exists := s.admins[adminID]
	if !exists {
		return nil, errors.New("admin not found")
	}

	return admin, nil
}

// ListAdmins returns all admins
func (s *AuthService) ListAdmins() []*Admin {
	s.mu.RLock()
	defer s.mu.RUnlock()

	admins := make([]*Admin, 0, len(s.admins))
	for _, admin := range s.admins {
		admins = append(admins, admin)
	}

	return admins
}

// DisableAdmin disables an admin account
func (s *AuthService) DisableAdmin(adminID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	admin, exists := s.admins[adminID]
	if !exists {
		return errors.New("admin not found")
	}

	admin.Status = "DISABLED"
	log.Printf("[AdminAuth] Admin disabled: %s", admin.Username)

	// Terminate all sessions for this admin
	for sessionID, session := range s.sessions {
		if session.AdminID == adminID {
			delete(s.sessions, sessionID)
		}
	}

	return nil
}

// EnableAdmin enables an admin account
func (s *AuthService) EnableAdmin(adminID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	admin, exists := s.admins[adminID]
	if !exists {
		return errors.New("admin not found")
	}

	admin.Status = "ACTIVE"
	log.Printf("[AdminAuth] Admin enabled: %s", admin.Username)
	return nil
}

// Helper functions

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *AuthService) isIPWhitelisted(ip string, whitelist []string) bool {
	userIP := net.ParseIP(ip)
	if userIP == nil {
		return false
	}

	for _, whitelisted := range whitelist {
		// Check if it's a CIDR range
		if _, cidr, err := net.ParseCIDR(whitelisted); err == nil {
			if cidr.Contains(userIP) {
				return true
			}
		} else {
			// Direct IP comparison
			whitelistedIP := net.ParseIP(whitelisted)
			if whitelistedIP != nil && subtle.ConstantTimeCompare(userIP, whitelistedIP) == 1 {
				return true
			}
		}
	}

	return false
}

// GetActiveSessions returns all active sessions
func (s *AuthService) GetActiveSessions() []*AdminSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*AdminSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		if time.Now().Before(session.ExpiresAt) {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// CleanupExpiredSessions removes expired sessions (should be run periodically)
func (s *AuthService) CleanupExpiredSessions() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	now := time.Now()
	for sessionID, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, sessionID)
			count++
		}
	}

	if count > 0 {
		log.Printf("[AdminAuth] Cleaned up %d expired sessions", count)
	}

	return count
}
