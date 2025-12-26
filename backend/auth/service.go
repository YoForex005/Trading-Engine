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
	if username == "admin" && password == "password" {
		return "mock_jwt_token_12345", &User{
			ID:       "user_001",
			Username: "admin",
			Role:     "ADMIN",
		}, nil
	}
	
	if username == "trader" && password == "password" {
		return "mock_jwt_token_67890", &User{
			ID:       "user_002",
			Username: "trader",
			Role:     "TRADER",
		}, nil
	}

	return "", nil, errors.New("invalid credentials")
}
