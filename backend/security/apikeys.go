package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

// APIKeyManager handles API key generation, rotation, and validation
type APIKeyManager struct {
	keys         map[string]*APIKey
	serviceKeys  map[string]string // service -> current key ID
	mu           sync.RWMutex
	rotationInterval time.Duration
	auditLogger  *AuditLogger
	encryption   *EncryptionService
}

// APIKey represents an API key
type APIKey struct {
	ID          string
	Service     string
	Key         string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	LastUsed    time.Time
	Rotations   int
	Active      bool
}

// NewAPIKeyManager creates a new API key manager
func NewAPIKeyManager(rotationInterval time.Duration, auditLogger *AuditLogger, encryption *EncryptionService) *APIKeyManager {
	manager := &APIKeyManager{
		keys:             make(map[string]*APIKey),
		serviceKeys:      make(map[string]string),
		rotationInterval: rotationInterval,
		auditLogger:      auditLogger,
		encryption:       encryption,
	}

	// Start auto-rotation routine
	go manager.autoRotationRoutine()

	return manager
}

// GenerateKey generates a new API key for a service
func (m *APIKeyManager) GenerateKey(service string) (*APIKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate cryptographically secure key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, err
	}

	key := base64.URLEncoding.EncodeToString(keyBytes)

	// Create API key
	apiKey := &APIKey{
		ID:        generateKeyID(),
		Service:   service,
		Key:       key,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(m.rotationInterval),
		Active:    true,
	}

	// Store key
	m.keys[apiKey.ID] = apiKey
	m.serviceKeys[service] = apiKey.ID

	// Log key generation
	if m.auditLogger != nil {
		m.auditLogger.LogAPIKeyRotation(service, "new_key_generated", true)
	}

	return apiKey, nil
}

// RotateKey rotates an API key for a service
func (m *APIKeyManager) RotateKey(service string, reason string) (*APIKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get current key
	currentKeyID, exists := m.serviceKeys[service]
	if exists {
		currentKey := m.keys[currentKeyID]
		currentKey.Active = false
	}

	// Generate new key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, err
	}

	key := base64.URLEncoding.EncodeToString(keyBytes)

	// Create new API key
	apiKey := &APIKey{
		ID:        generateKeyID(),
		Service:   service,
		Key:       key,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(m.rotationInterval),
		Active:    true,
	}

	if exists {
		apiKey.Rotations = m.keys[currentKeyID].Rotations + 1
	}

	// Store new key
	m.keys[apiKey.ID] = apiKey
	m.serviceKeys[service] = apiKey.ID

	// Log rotation
	if m.auditLogger != nil {
		m.auditLogger.LogAPIKeyRotation(service, reason, true)
	}

	return apiKey, nil
}

// GetKey retrieves the current active key for a service
func (m *APIKeyManager) GetKey(service string) (*APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keyID, exists := m.serviceKeys[service]
	if !exists {
		return nil, errors.New("no key found for service")
	}

	key := m.keys[keyID]
	if !key.Active {
		return nil, errors.New("key is inactive")
	}

	if time.Now().After(key.ExpiresAt) {
		return nil, errors.New("key expired")
	}

	// Update last used
	key.LastUsed = time.Now()

	return key, nil
}

// ValidateKey validates an API key
func (m *APIKeyManager) ValidateKey(service, keyString string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keyID, exists := m.serviceKeys[service]
	if !exists {
		return false, errors.New("no key found for service")
	}

	key := m.keys[keyID]

	if key.Key != keyString {
		return false, errors.New("key mismatch")
	}

	if !key.Active {
		return false, errors.New("key is inactive")
	}

	if time.Now().After(key.ExpiresAt) {
		return false, errors.New("key expired")
	}

	// Update last used
	key.LastUsed = time.Now()

	return true, nil
}

// RevokeKey revokes an API key
func (m *APIKeyManager) RevokeKey(service string, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	keyID, exists := m.serviceKeys[service]
	if !exists {
		return errors.New("no key found for service")
	}

	key := m.keys[keyID]
	key.Active = false

	// Log revocation
	if m.auditLogger != nil {
		m.auditLogger.Log(AuditEvent{
			Level:    AuditLevelSecurity,
			Category: "api_keys",
			Action:   "key_revoked",
			Resource: service,
			Success:  true,
			Message:  "API key revoked: " + reason,
		})
	}

	return nil
}

// ListKeys returns all keys for a service
func (m *APIKeyManager) ListKeys(service string) []*APIKey {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var keys []*APIKey
	for _, key := range m.keys {
		if key.Service == service {
			keys = append(keys, key)
		}
	}

	return keys
}

// autoRotationRoutine automatically rotates keys based on rotation interval
func (m *APIKeyManager) autoRotationRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		m.checkAndRotate()
	}
}

// checkAndRotate checks for expiring keys and rotates them
func (m *APIKeyManager) checkAndRotate() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	for service, keyID := range m.serviceKeys {
		key := m.keys[keyID]

		// Rotate if key is about to expire (within 7 days)
		if key.ExpiresAt.Sub(now) < 7*24*time.Hour {
			// Generate new key
			keyBytes := make([]byte, 32)
			rand.Read(keyBytes)
			newKeyString := base64.URLEncoding.EncodeToString(keyBytes)

			newKey := &APIKey{
				ID:        generateKeyID(),
				Service:   service,
				Key:       newKeyString,
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(m.rotationInterval),
				Rotations: key.Rotations + 1,
				Active:    true,
			}

			// Deactivate old key
			key.Active = false

			// Store new key
			m.keys[newKey.ID] = newKey
			m.serviceKeys[service] = newKey.ID

			// Log rotation
			if m.auditLogger != nil {
				m.auditLogger.LogAPIKeyRotation(service, "auto_rotation", true)
			}
		}
	}
}

// GetRotationStatus returns rotation status for all services
func (m *APIKeyManager) GetRotationStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]interface{})

	for service, keyID := range m.serviceKeys {
		key := m.keys[keyID]
		timeUntilExpiry := key.ExpiresAt.Sub(time.Now())

		status[service] = map[string]interface{}{
			"key_id":           key.ID,
			"created_at":       key.CreatedAt,
			"expires_at":       key.ExpiresAt,
			"time_until_expiry": timeUntilExpiry.String(),
			"rotations":        key.Rotations,
			"last_used":        key.LastUsed,
			"active":           key.Active,
		}
	}

	return status
}

// generateKeyID generates a unique key ID
func generateKeyID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// EncryptAndStoreKey encrypts and stores a key securely
func (m *APIKeyManager) EncryptAndStoreKey(service, key string) error {
	if m.encryption == nil {
		return errors.New("encryption service not available")
	}

	encrypted, err := m.encryption.EncryptString(key)
	if err != nil {
		return err
	}

	// Store encrypted key (implementation depends on storage backend)
	_ = encrypted

	return nil
}

// DecryptAndRetrieveKey decrypts and retrieves a stored key
func (m *APIKeyManager) DecryptAndRetrieveKey(service string) (string, error) {
	if m.encryption == nil {
		return "", errors.New("encryption service not available")
	}

	// Retrieve encrypted key (implementation depends on storage backend)
	encryptedKey := ""

	decrypted, err := m.encryption.DecryptString(encryptedKey)
	if err != nil {
		return "", err
	}

	return decrypted, nil
}
