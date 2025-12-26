package auth

import (
	"errors"
	"strconv"
	"github.com/epic1st/rtx/backend/bbook"
)

// User represents a system user
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Service handles authentication logic
type Service struct{
	engine *bbook.Engine
}

func NewService(engine *bbook.Engine) *Service {
	return &Service{
		engine: engine,
	}
}

// Login validates credentials
func (s *Service) Login(username, password string) (string, *User, error) {
	// 1. Admin Login (Hardcoded for now)
	if username == "admin" && password == "password" {
		return "mock_jwt_token_admin", &User{
			ID:       "0", 
			Username: "admin",
			Role:     "ADMIN",
		}, nil
	}
	
	// 2. Client Login
	// User Requirement: "username login should always be in number"
	// So we attempt to parse 'username' as Account ID.
	if accountID, err := strconv.ParseInt(username, 10, 64); err == nil {
		account, found := s.engine.GetAccount(accountID)
		if found {
			// Check if account has a custom password set
			if account.Password != "" {
				if account.Password == password {
					return "mock_jwt_token_client", &User{
						ID:       username,
						Username: account.Username, // Return the custom username if set
						Role:     "TRADER",
					}, nil
				} else {
					return "", nil, errors.New("invalid password")
				}
			} 
			
			// Legacy/Dev fallback: If no password set, allow "password"
			if password == "password" {
				return "mock_jwt_token_universal", &User{
					ID:       username, 
					Username: account.Username,
					Role:     "TRADER",
				}, nil
			}
		}
	}
	
	// Fallback for "Universal Dev Mode" - if account ID not found or valid but no password set?
	// Actually, strictly enforce that account MUST exist.
	// If password is "password" but account doesn't exist, Create it on the fly? 
	// The user asked for Admin to create account. So we should probably Fail if not found.
	// However, for smoother dev exp, I'll keep the universal backdoor ONLY if password is "password" AND it's a number, creating it if missing?
	// NO, user wants Admin control. Let's return error if not found.
	
	return "", nil, errors.New("invalid credentials or account not found")
}
