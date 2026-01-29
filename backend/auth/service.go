package auth

import (
	"errors"
	"log"
	"strconv"

	"github.com/epic1st/rtx/backend/internal/core"
	"golang.org/x/crypto/bcrypt"
)

// User represents a system user
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Service handles authentication logic
type Service struct {
	engine    *core.Engine
	adminHash []byte
	jwtSecret []byte
}

// NewService creates authentication service with admin credentials and JWT secret
func NewService(engine *core.Engine, adminPasswordHash string, jwtSecret string) *Service {
	// Use provided hash or generate default (for development only)
	var hash []byte
	if adminPasswordHash != "" {
		hash = []byte(adminPasswordHash)
	} else {
		// WARNING: Development only - generate temporary hash for "password"
		log.Println("[SECURITY WARNING] No ADMIN_PASSWORD_HASH provided - using insecure default password")
		hash, _ = bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	}

	// Validate JWT secret
	secret := []byte(jwtSecret)
	if len(secret) == 0 {
		// WARNING: Development only - use insecure default
		log.Println("[SECURITY WARNING] No JWT_SECRET provided - using insecure default secret")
		secret = []byte("super_secret_dev_key_do_not_use_in_prod")
	}

	return &Service{
		engine:    engine,
		adminHash: hash,
		jwtSecret: secret,
	}
}

// Login validates credentials securely
func (s *Service) Login(username, password string) (string, *User, error) {
	// 1. Admin Login
	if username == "admin" {
		err := bcrypt.CompareHashAndPassword(s.adminHash, []byte(password))
		if err != nil {
			log.Printf("[WARN] Admin login failed (invalid password)")
			return "", nil, errors.New("invalid credentials")
		}

		log.Printf("[INFO] Admin logged in")
		user := &User{ID: "0", Username: "admin", Role: "ADMIN"}
		token, err := s.GenerateToken(user)
		if err != nil {
			log.Printf("[CRITICAL] JWT Generation failed: %v", err)
			return "", nil, errors.New("system error")
		}
		return token, user, nil
	}

	// 2. Client Login
	// We iterate accounts to find by username or ID
	// Attempt ID lookup
	var account *core.Account

	if id, err := strconv.ParseInt(username, 10, 64); err == nil {
		acc, ok := s.engine.GetAccount(id)
		if ok {
			account = acc
		}
	}

	// If not found by numeric ID, try looking up by string UserID
	if account == nil {
		accounts := s.engine.GetAccountByUser(username)
		if len(accounts) > 0 {
			account = accounts[0] // Use first account for this user
		}
	}

	if account == nil {
		log.Printf("[WARN] Login failed: User %s not found", username)
		return "", nil, errors.New("invalid credentials")
		// "User Enumeration" prevention: return same error as password fail.
	}

	// 3. Verify Password
	// We treat account.Password as a Hash.
	err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil {
		// Fallback: Check if it's plaintext "password" (Legacy Dev Mode)
		// ONLY if it doesn't look like a hash (bcrypt starts with $2)
		// And check if password is correct plaintext.
		if len(account.Password) > 0 && account.Password[0] != '$' && account.Password == password {
			log.Printf("[WARN] Auto-upgrading password for user %s to bcrypt", account.Username)
			newHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			s.engine.UpdatePassword(account.ID, string(newHash))
			// Proceed
		} else {
			log.Printf("[WARN] Login failed for user %s (invalid password)", username)
			return "", nil, errors.New("invalid credentials")
		}
	}

	// Success
	user := &User{
		ID:       strconv.FormatInt(account.ID, 10),
		Username: account.Username,
		Role:     "TRADER",
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		log.Printf("[CRITICAL] JWT Generation failed: %v", err)
		return "", nil, errors.New("system error")
	}

	return token, user, nil
}

// GenerateToken creates a JWT token for the given user using the service's secret
func (s *Service) GenerateToken(user *User) (string, error) {
	return GenerateJWTWithSecret(user, s.jwtSecret)
}

// ValidateToken validates a JWT token using the service's secret
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	return ValidateToken(tokenString, s.jwtSecret)
}
