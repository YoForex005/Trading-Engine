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
	engine *core.Engine
	// Admin hash: $2a$10$... (hash of 'password')

	adminHash []byte
}

func NewService(engine *core.Engine) *Service {

	// Generate hash for "password" on startup for the admin (Simulated Secure Config)
	// Cost 10 is decent for startup.
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	return &Service{
		engine:    engine,
		adminHash: hash,
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
		token, err := GenerateJWT(user)
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

	if account == nil {
		log.Printf("[WARN] Login failed: User %s not found", username)
		return "", nil, errors.New("invalid credentials")
		// "User Enumeration" prevention: return same error as password fail.
	}

	// 3. Verify Password (bcrypt only - no plaintext fallback)
	err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil {
		log.Printf("[WARN] Login failed for user %s (invalid password)", username)
		return "", nil, errors.New("invalid credentials")
	}

	// Password valid - continue to JWT generation

	// Success
	user := &User{
		ID:       strconv.FormatInt(account.ID, 10),
		Username: account.Username,
		Role:     "TRADER",
	}

	token, err := GenerateJWT(user)
	if err != nil {
		log.Printf("[CRITICAL] JWT Generation failed: %v", err)
		return "", nil, errors.New("system error")
	}

	return token, user, nil
}
