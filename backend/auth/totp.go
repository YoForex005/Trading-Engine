package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPService handles TOTP generation and validation
type TOTPService struct {
	encryptionKey []byte
	issuer        string
}

// NewTOTPService creates a new TOTP service
func NewTOTPService() *TOTPService {
	// Get encryption key from environment
	keyStr := os.Getenv("TOTP_ENCRYPTION_KEY")
	if keyStr == "" {
		panic("TOTP_ENCRYPTION_KEY environment variable is required for 2FA")
	}

	// Key must be 32 bytes for AES-256
	key := []byte(keyStr)
	if len(key) != 32 {
		panic("TOTP_ENCRYPTION_KEY must be exactly 32 bytes for AES-256-GCM")
	}

	issuer := os.Getenv("TOTP_ISSUER")
	if issuer == "" {
		issuer = "RTX5 Trading Engine"
	}

	return &TOTPService{
		encryptionKey: key,
		issuer:        issuer,
	}
}

// GenerateTOTPSecret generates a new TOTP secret for a user
func (s *TOTPService) GenerateTOTPSecret(userID string) (secret, qrCodeURL string, err error) {
	// Generate TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: userID,
		SecretSize:  32, // 256-bit secret
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// Encrypt the secret before storing
	encryptedSecret, err := s.encryptSecret(key.Secret())
	if err != nil {
		return "", "", fmt.Errorf("failed to encrypt secret: %w", err)
	}

	// Generate QR code URL for authenticator apps
	qrCodeURL = key.URL()

	return encryptedSecret, qrCodeURL, nil
}

// ValidateTOTPCode validates a 6-digit TOTP code against the secret
func (s *TOTPService) ValidateTOTPCode(encryptedSecret, code string) (bool, error) {
	// Decrypt the secret
	secret, err := s.decryptSecret(encryptedSecret)
	if err != nil {
		return false, fmt.Errorf("failed to decrypt secret: %w", err)
	}

	// Validate the code with 1 period grace (30 seconds before/after)
	valid := totp.Validate(code, secret)
	return valid, nil
}

// ValidateTOTPCodeWithSkew validates with custom time skew tolerance
func (s *TOTPService) ValidateTOTPCodeWithSkew(encryptedSecret, code string, skew uint) (bool, error) {
	// Decrypt the secret
	secret, err := s.decryptSecret(encryptedSecret)
	if err != nil {
		return false, fmt.Errorf("failed to decrypt secret: %w", err)
	}

	// Validate with custom skew (number of 30-second periods to check)
	opts := totp.ValidateOpts{
		Period:    30,
		Skew:      skew, // Allow codes from +/- skew periods
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	}

	valid, err := totp.ValidateCustom(code, secret, time.Now(), opts)
	if err != nil {
		return false, fmt.Errorf("validation error: %w", err)
	}

	return valid, nil
}

// encryptSecret encrypts the TOTP secret using AES-256-GCM
func (s *TOTPService) encryptSecret(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt and prepend nonce
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Base64 encode for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptSecret decrypts the TOTP secret using AES-256-GCM
func (s *TOTPService) decryptSecret(encryptedBase64 string) (string, error) {
	// Base64 decode
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
