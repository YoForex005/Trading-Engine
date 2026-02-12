package auth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

// TOTPConfig stores 2FA configuration for a user
type TOTPConfig struct {
	Secret      string    `json:"-"` // Encrypted secret, never expose
	Enabled     bool      `json:"enabled"`
	BackupCodes []string  `json:"-"` // Hashed backup codes
	CreatedAt   time.Time `json:"createdAt"`
}

// TOTPStore manages TOTP configurations in memory
type TOTPStore struct {
	mu      sync.RWMutex
	configs map[string]*TOTPConfig // userID -> config
}

// NewTOTPStore creates a new TOTP store
func NewTOTPStore() *TOTPStore {
	return &TOTPStore{
		configs: make(map[string]*TOTPConfig),
	}
}

// EnableTOTP enables 2FA for a user
func (s *TOTPStore) EnableTOTP(userID, secret string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	config := s.configs[userID]
	if config == nil {
		config = &TOTPConfig{
			CreatedAt: time.Now(),
		}
		s.configs[userID] = config
	}

	config.Secret = secret
	config.Enabled = true

	return nil
}

// DisableTOTP disables 2FA for a user
func (s *TOTPStore) DisableTOTP(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	config := s.configs[userID]
	if config != nil {
		config.Enabled = false
		config.Secret = ""
		config.BackupCodes = nil
	}

	return nil
}

// GetTOTPConfig retrieves TOTP configuration for a user
func (s *TOTPStore) GetTOTPConfig(userID string) (*TOTPConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.configs[userID]
	return config, exists
}

// SetBackupCodes stores backup codes for a user
func (s *TOTPStore) SetBackupCodes(userID string, codes []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	config := s.configs[userID]
	if config == nil {
		return ErrTOTPNotEnabled
	}

	config.BackupCodes = codes
	return nil
}

// ValidateBackupCode validates and consumes a backup code (one-time use)
func (s *TOTPStore) ValidateBackupCode(userID, code string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	config := s.configs[userID]
	if config == nil || !config.Enabled {
		return false, ErrTOTPNotEnabled
	}

	// Check if code exists in backup codes
	for i, backupCode := range config.BackupCodes {
		if backupCode == code {
			// Remove used backup code (one-time use)
			config.BackupCodes = append(config.BackupCodes[:i], config.BackupCodes[i+1:]...)
			return true, nil
		}
	}

	return false, nil
}

// GenerateBackupCodes generates one-time backup codes
func GenerateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)

	for i := 0; i < count; i++ {
		// Generate 8-byte random code
		b := make([]byte, 8)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}

		// Encode as base64 and take first 12 characters
		code := base64.RawStdEncoding.EncodeToString(b)[:12]
		codes[i] = code
	}

	return codes, nil
}

// IsTOTPEnabled checks if 2FA is enabled for a user
func (s *TOTPStore) IsTOTPEnabled(userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.configs[userID]
	return exists && config.Enabled
}
