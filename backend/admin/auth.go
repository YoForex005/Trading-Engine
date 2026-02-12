package admin

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Minimal audit types for current admin auth flow.
// Full audit logger lives behind legacy build tags.
type AuditLogEntry struct {
	AdminID    string
	AdminEmail string
	Action     string
	Resource   string
	ResourceID string
	Details    map[string]interface{}
	IPAddress  string
	Success    bool
	ErrorMsg   string
}

type AuditLogger struct{}

func (l *AuditLogger) Log(_ AuditLogEntry) {}

const ActionLogin = "LOGIN"

// AuthService handles admin authentication and authorization
type AuthService struct {
	mu              sync.RWMutex
	admins          map[int64]*Admin
	adminsByUsername map[string]*Admin
	sessions        map[string]*AdminSession
	nextAdminID     int64
	auditLogger     *AuditLogger // Audit logger for tracking admin actions
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

// SetAuditLogger sets the audit logger for tracking admin actions
func (s *AuthService) SetAuditLogger(logger *AuditLogger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.auditLogger = logger
	log.Println("[AdminAuth] Audit logger attached to auth service")
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
// If 2FA is enabled, returns a partial session that requires TOTP validation
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

		// Log failed login attempt
		if s.auditLogger != nil {
			s.auditLogger.Log(AuditLogEntry{
				AdminID:    fmt.Sprintf("%d", admin.ID),
				AdminEmail: admin.Email,
				Action:     ActionLogin,
				Resource:   "auth",
				ResourceID: username,
				Details: map[string]interface{}{
					"reason": "invalid_password",
				},
				IPAddress: ipAddress,
				Success:   false,
				ErrorMsg:  "Invalid credentials",
			})
		}

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

	// If 2FA is enabled, return partial session requiring TOTP validation
	if admin.TwoFactorEnabled {
		log.Printf("[AdminAuth] 2FA required for admin: %s from %s", username, ipAddress)

		// Create a partial session with short TTL (5 minutes)
		now := time.Now()
		partialSession := &AdminSession{
			SessionID:  "", // Will be set after 2FA validation
			AdminID:    admin.ID,
			Username:   admin.Username,
			Role:       admin.Role,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
			CreatedAt:  now,
			ExpiresAt:  now.Add(5 * time.Minute), // Short TTL for 2FA step
			LastActive: now,
		}

		return partialSession, errors.New("2FA_REQUIRED")
	}

	// Generate session token
	sessionID, err := generateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session: %w", err)
	}

	// Create full session (no 2FA)
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

	// Log successful login
	if s.auditLogger != nil {
		s.auditLogger.Log(AuditLogEntry{
			AdminID:    fmt.Sprintf("%d", admin.ID),
			AdminEmail: admin.Email,
			Action:     ActionLogin,
			Resource:   "auth",
			ResourceID: username,
			Details: map[string]interface{}{
				"userAgent":   userAgent,
				"sessionID":   sessionID,
				"twoFactorEnabled": admin.TwoFactorEnabled,
			},
			IPAddress: ipAddress,
			Success:   true,
		})
	}

	log.Printf("[AdminAuth] Admin logged in: %s (role: %s) from %s", username, admin.Role, ipAddress)
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

// ValidateAdminToken validates a bearer/session token from request headers.
// It returns the authenticated admin ID for compatibility with handlers.
func (s *AuthService) ValidateAdminToken(r *http.Request) (int64, error) {
	token := ""

	if authHeader := strings.TrimSpace(r.Header.Get("Authorization")); authHeader != "" {
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = strings.TrimSpace(authHeader[7:])
		} else {
			token = authHeader
		}
	}

	if token == "" {
		token = strings.TrimSpace(r.Header.Get("X-Session-ID"))
	}
	if token == "" {
		token = strings.TrimSpace(r.Header.Get("X-Admin-Session"))
	}
	if token == "" {
		return 0, errors.New("missing admin token")
	}

	admin, err := s.ValidateSession(token, getIPAddress(r))
	if err != nil {
		return 0, err
	}

	return admin.ID, nil
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

// UpdateAdmin updates admin details (email, role, status, password)
func (s *AuthService) UpdateAdmin(adminID int64, email string, role AdminRole, status string, password string) (*Admin, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	admin, exists := s.admins[adminID]
	if !exists {
		return nil, errors.New("admin not found")
	}

	// Update email if provided
	if email != "" && email != admin.Email {
		admin.Email = email
	}

	// Update role if provided
	if role != "" {
		admin.Role = role
	}

	// Update status if provided
	if status != "" {
		admin.Status = status
	}

	// Update password if provided
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		admin.PasswordHash = string(hash)
	}

	log.Printf("[AdminAuth] Admin updated: %s (email=%s, role=%s, status=%s)", admin.Username, admin.Email, admin.Role, admin.Status)
	return admin, nil
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
