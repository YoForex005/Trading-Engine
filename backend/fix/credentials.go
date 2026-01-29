package fix

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// Credential status
	CredentialStatusActive   = "active"
	CredentialStatusRevoked  = "revoked"
	CredentialStatusExpired  = "expired"
	CredentialStatusSuspended = "suspended"

	// Encryption settings
	encryptionKeyIterations = 100000
	encryptionKeySalt       = "fix-credentials-salt-v1"
)

// FIXCredentials represents a user's FIX API access credentials
type FIXCredentials struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	SenderCompID    string    `json:"sender_comp_id"`
	TargetCompID    string    `json:"target_comp_id"`
	Password        string    `json:"password"` // Encrypted
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	RateLimitTier   string    `json:"rate_limit_tier"` // basic, standard, premium
	AllowedIPs      []string  `json:"allowed_ips,omitempty"`
	MaxSessions     int       `json:"max_sessions"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// CredentialStore manages FIX credentials with encryption
type CredentialStore struct {
	storePath      string
	encryptionKey  []byte
	credentials    map[string]*FIXCredentials // userID -> credentials
	mu             sync.RWMutex
	auditLogger    AuditLogger
}

// AuditLogger interface for credential operations
type AuditLogger interface {
	LogCredentialOperation(operation, userID, details string, success bool)
}

// NewCredentialStore creates a new credential store
func NewCredentialStore(storePath, masterPassword string, auditLogger AuditLogger) (*CredentialStore, error) {
	// Derive encryption key from master password
	encKey := pbkdf2.Key([]byte(masterPassword), []byte(encryptionKeySalt), encryptionKeyIterations, 32, sha256.New)

	store := &CredentialStore{
		storePath:     storePath,
		encryptionKey: encKey,
		credentials:   make(map[string]*FIXCredentials),
		auditLogger:   auditLogger,
	}

	// Load existing credentials
	if err := store.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	return store, nil
}

// GenerateCredentials creates new FIX credentials for a user
func (cs *CredentialStore) GenerateCredentials(userID, rateLimitTier string, maxSessions int, expiresIn *time.Duration) (*FIXCredentials, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Check if user already has active credentials
	if existing, exists := cs.credentials[userID]; exists && existing.Status == CredentialStatusActive {
		cs.auditLogger.LogCredentialOperation("generate_failed", userID, "user already has active credentials", false)
		return nil, errors.New("user already has active credentials")
	}

	// Generate unique IDs
	senderCompID := fmt.Sprintf("USER_%s", generateRandomID(8))
	targetCompID := "GATEWAY" // Your gateway's target comp ID
	password := generateSecurePassword(32)

	// Encrypt password
	encryptedPassword, err := cs.encrypt(password)
	if err != nil {
		cs.auditLogger.LogCredentialOperation("generate_failed", userID, "encryption failed", false)
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}

	now := time.Now()
	creds := &FIXCredentials{
		ID:            generateRandomID(16),
		UserID:        userID,
		SenderCompID:  senderCompID,
		TargetCompID:  targetCompID,
		Password:      encryptedPassword,
		Status:        CredentialStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
		RateLimitTier: rateLimitTier,
		MaxSessions:   maxSessions,
		Metadata:      make(map[string]interface{}),
	}

	if expiresIn != nil {
		expiresAt := now.Add(*expiresIn)
		creds.ExpiresAt = &expiresAt
	}

	cs.credentials[userID] = creds

	// Persist to disk
	if err := cs.save(); err != nil {
		delete(cs.credentials, userID)
		cs.auditLogger.LogCredentialOperation("generate_failed", userID, "save failed", false)
		return nil, fmt.Errorf("failed to save credentials: %w", err)
	}

	cs.auditLogger.LogCredentialOperation("generate_success", userID, fmt.Sprintf("sender=%s", senderCompID), true)

	// Return credentials with decrypted password (only time it's shown in plaintext)
	result := *creds
	result.Password = password
	return &result, nil
}

// GetCredentials retrieves credentials for a user
func (cs *CredentialStore) GetCredentials(userID string) (*FIXCredentials, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	creds, exists := cs.credentials[userID]
	if !exists {
		return nil, errors.New("credentials not found")
	}

	return creds, nil
}

// ValidateCredentials checks if credentials are valid for login
func (cs *CredentialStore) ValidateCredentials(senderCompID, password string) (*FIXCredentials, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	// Find credentials by SenderCompID
	for _, creds := range cs.credentials {
		if creds.SenderCompID == senderCompID {
			// Check status
			if creds.Status != CredentialStatusActive {
				cs.auditLogger.LogCredentialOperation("validate_failed", creds.UserID, fmt.Sprintf("status=%s", creds.Status), false)
				return nil, fmt.Errorf("credentials are %s", creds.Status)
			}

			// Check expiration
			if creds.ExpiresAt != nil && time.Now().After(*creds.ExpiresAt) {
				cs.auditLogger.LogCredentialOperation("validate_failed", creds.UserID, "credentials expired", false)
				return nil, errors.New("credentials have expired")
			}

			// Decrypt and compare password
			decryptedPassword, err := cs.decrypt(creds.Password)
			if err != nil {
				cs.auditLogger.LogCredentialOperation("validate_failed", creds.UserID, "decryption failed", false)
				return nil, errors.New("failed to validate credentials")
			}

			if decryptedPassword != password {
				cs.auditLogger.LogCredentialOperation("validate_failed", creds.UserID, "password mismatch", false)
				return nil, errors.New("invalid credentials")
			}

			// Update last used timestamp
			now := time.Now()
			creds.LastUsedAt = &now

			cs.auditLogger.LogCredentialOperation("validate_success", creds.UserID, senderCompID, true)
			return creds, nil
		}
	}

	cs.auditLogger.LogCredentialOperation("validate_failed", "unknown", fmt.Sprintf("sender=%s", senderCompID), false)
	return nil, errors.New("invalid credentials")
}

// RevokeCredentials revokes a user's FIX access
func (cs *CredentialStore) RevokeCredentials(userID, reason string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	creds, exists := cs.credentials[userID]
	if !exists {
		return errors.New("credentials not found")
	}

	now := time.Now()
	creds.Status = CredentialStatusRevoked
	creds.RevokedAt = &now
	creds.UpdatedAt = now

	if err := cs.save(); err != nil {
		cs.auditLogger.LogCredentialOperation("revoke_failed", userID, reason, false)
		return fmt.Errorf("failed to save revocation: %w", err)
	}

	cs.auditLogger.LogCredentialOperation("revoke_success", userID, reason, true)
	return nil
}

// RegeneratePassword generates a new password for existing credentials
func (cs *CredentialStore) RegeneratePassword(userID string) (string, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	creds, exists := cs.credentials[userID]
	if !exists {
		cs.auditLogger.LogCredentialOperation("regenerate_failed", userID, "not found", false)
		return "", errors.New("credentials not found")
	}

	// Generate new password
	newPassword := generateSecurePassword(32)
	encryptedPassword, err := cs.encrypt(newPassword)
	if err != nil {
		cs.auditLogger.LogCredentialOperation("regenerate_failed", userID, "encryption failed", false)
		return "", fmt.Errorf("failed to encrypt password: %w", err)
	}

	creds.Password = encryptedPassword
	creds.UpdatedAt = time.Now()

	if err := cs.save(); err != nil {
		cs.auditLogger.LogCredentialOperation("regenerate_failed", userID, "save failed", false)
		return "", fmt.Errorf("failed to save new password: %w", err)
	}

	cs.auditLogger.LogCredentialOperation("regenerate_success", userID, "password regenerated", true)
	return newPassword, nil
}

// SuspendCredentials temporarily suspends access
func (cs *CredentialStore) SuspendCredentials(userID, reason string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	creds, exists := cs.credentials[userID]
	if !exists {
		return errors.New("credentials not found")
	}

	creds.Status = CredentialStatusSuspended
	creds.UpdatedAt = time.Now()

	if err := cs.save(); err != nil {
		cs.auditLogger.LogCredentialOperation("suspend_failed", userID, reason, false)
		return fmt.Errorf("failed to save suspension: %w", err)
	}

	cs.auditLogger.LogCredentialOperation("suspend_success", userID, reason, true)
	return nil
}

// ReactivateCredentials reactivates suspended credentials
func (cs *CredentialStore) ReactivateCredentials(userID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	creds, exists := cs.credentials[userID]
	if !exists {
		return errors.New("credentials not found")
	}

	if creds.Status != CredentialStatusSuspended {
		return fmt.Errorf("cannot reactivate credentials with status: %s", creds.Status)
	}

	creds.Status = CredentialStatusActive
	creds.UpdatedAt = time.Now()

	if err := cs.save(); err != nil {
		cs.auditLogger.LogCredentialOperation("reactivate_failed", userID, "save failed", false)
		return fmt.Errorf("failed to save reactivation: %w", err)
	}

	cs.auditLogger.LogCredentialOperation("reactivate_success", userID, "credentials reactivated", true)
	return nil
}

// ListAllCredentials returns all credentials (for admin)
func (cs *CredentialStore) ListAllCredentials() []*FIXCredentials {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	result := make([]*FIXCredentials, 0, len(cs.credentials))
	for _, creds := range cs.credentials {
		result = append(result, creds)
	}

	return result
}

// encrypt encrypts data using AES-GCM
func (cs *CredentialStore) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(cs.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts data using AES-GCM
func (cs *CredentialStore) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(cs.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// save persists credentials to disk (encrypted)
func (cs *CredentialStore) save() error {
	// Ensure directory exists
	dir := filepath.Dir(cs.storePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cs.credentials, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file first, then rename (atomic operation)
	tempPath := cs.storePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return err
	}

	return os.Rename(tempPath, cs.storePath)
}

// load reads credentials from disk
func (cs *CredentialStore) load() error {
	data, err := os.ReadFile(cs.storePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &cs.credentials)
}

// generateRandomID generates a random alphanumeric ID
func generateRandomID(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b)
}

// generateSecurePassword generates a cryptographically secure password
func generateSecurePassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b)
}
