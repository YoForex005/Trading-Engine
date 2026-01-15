package auth

import (
	"log"
	"os"
	"time"

	_ "github.com/epic1st/rtx/backend/config" // Load .env before reading JWT_SECRET
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// init validates JWT_SECRET is set and cryptographically secure
func init() {
	if len(jwtKey) == 0 {
		log.Fatal("[CRITICAL] JWT_SECRET environment variable not set. Generate with: openssl rand -base64 32")
	}
	if len(jwtKey) < 32 {
		log.Fatal("[CRITICAL] JWT_SECRET too short (minimum 32 bytes required)")
	}
	log.Printf("[Auth] JWT secret loaded (%d bytes)", len(jwtKey))
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a new token for a user
func GenerateJWT(user *User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "rtx-trading-engine",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
