package auth

import (
	"errors"
)

// User represents a system user
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Service handles authentication logic
type Service struct{}

func NewService() *Service {
	return &Service{}
}

// Login validates credentials and returns a token (mocked)
func (s *Service) Login(username, password string) (string, *User, error) {
	// Mock Database Lookup
	// Mock Database Lookup
	if username == "admin" && password == "password" {
		return "mock_jwt_token_12345", &User{
			ID:       "1", // Default admin account is 1
			Username: "admin",
			Role:     "ADMIN",
		}, nil
	}
	
	// DEV MODE: Allow any username if password is "password"
	// This allows Account Switching by entering Account ID as "username"
	if password == "password" {
		return "mock_jwt_token_universal", &User{
			ID:       username, // Use the input username as the User ID (Account ID)
			Username: username,
			Role:     "TRADER",
		}, nil
	}

	return "", nil, errors.New("invalid credentials")
}
