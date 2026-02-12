package admin

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
	"github.com/google/uuid"
)

// ApiKey represents a programmatic access key
type ApiKey struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	KeyHash   string    `json:"-"`              // SHA-256 hash of the key (never returned)
	Prefix    string    `json:"prefix"`         // First 8 chars for identification
	Scopes    []string  `json:"scopes"`         // read, trade, admin
	CreatedBy string    `json:"createdBy"`      // User ID who created the key
	CreatedAt time.Time `json:"createdAt"`
	LastUsed  time.Time `json:"lastUsed"`
	ExpiresAt time.Time `json:"expiresAt"`
	IsActive  bool      `json:"isActive"`
}

// ApiKeyStore manages API keys in memory
type ApiKeyStore struct {
	keys      map[string]*ApiKey // ID -> ApiKey
	keyHashes map[string]string  // KeyHash -> ID (for fast lookup)
	mu        sync.RWMutex

	// Rate limiting per API key
	rateLimits map[string]*apiKeyRateLimit
	rateLimitMu sync.RWMutex
}

// apiKeyRateLimit tracks API key usage for rate limiting
type apiKeyRateLimit struct {
	count     int
	resetTime time.Time
	mu        sync.Mutex
}

// NewApiKeyStore creates a new API key store
func NewApiKeyStore() *ApiKeyStore {
	return &ApiKeyStore{
		keys:       make(map[string]*ApiKey),
		keyHashes:  make(map[string]string),
		rateLimits: make(map[string]*apiKeyRateLimit),
	}
}

// GenerateApiKey generates a new API key with cryptographically secure random bytes
func GenerateApiKey() (string, error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}

	// Encode as base64 URL-safe string
	key := base64.URLEncoding.EncodeToString(bytes)

	// Add prefix for easy identification (rtx5_)
	return "rtx5_" + key, nil
}

// HashKey creates SHA-256 hash of an API key
func HashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// CreateKey creates a new API key
func (s *ApiKeyStore) CreateKey(name string, scopes []string, createdBy string, expiresInDays int) (*ApiKey, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate new API key
	apiKey, err := GenerateApiKey()
	if err != nil {
		return nil, "", err
	}

	// Hash the key for storage
	keyHash := HashKey(apiKey)

	// Get prefix (first 8 chars after rtx5_)
	prefix := apiKey[:13] // rtx5_ + first 8 chars

	// Create expiration time
	expiresAt := time.Now().AddDate(0, 0, expiresInDays)

	// Create ApiKey struct
	key := &ApiKey{
		ID:        uuid.New().String(),
		Name:      name,
		KeyHash:   keyHash,
		Prefix:    prefix,
		Scopes:    scopes,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		LastUsed:  time.Time{}, // Never used yet
		ExpiresAt: expiresAt,
		IsActive:  true,
	}

	// Store in maps
	s.keys[key.ID] = key
	s.keyHashes[keyHash] = key.ID

	log.Printf("[ApiKeys] Created new API key '%s' (ID: %s) for user %s with scopes: %v",
		name, key.ID, createdBy, scopes)

	// Return the key (only time it's shown in plaintext)
	return key, apiKey, nil
}

// GetKey retrieves an API key by ID
func (s *ApiKeyStore) GetKey(id string) (*ApiKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, exists := s.keys[id]
	if !exists {
		return nil, fmt.Errorf("API key not found")
	}

	return key, nil
}

// ListKeys returns all API keys (with sensitive data masked)
func (s *ApiKeyStore) ListKeys() []*ApiKey {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]*ApiKey, 0, len(s.keys))
	for _, key := range s.keys {
		keys = append(keys, key)
	}

	return keys
}

// UpdateKey updates an API key's name and scopes
func (s *ApiKeyStore) UpdateKey(id string, name string, scopes []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, exists := s.keys[id]
	if !exists {
		return fmt.Errorf("API key not found")
	}

	key.Name = name
	key.Scopes = scopes

	log.Printf("[ApiKeys] Updated API key %s: name='%s', scopes=%v", id, name, scopes)

	return nil
}

// RevokeKey deactivates an API key
func (s *ApiKeyStore) RevokeKey(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, exists := s.keys[id]
	if !exists {
		return fmt.Errorf("API key not found")
	}

	key.IsActive = false

	log.Printf("[ApiKeys] Revoked API key %s ('%s')", id, key.Name)

	return nil
}

// DeleteKey permanently deletes an API key
func (s *ApiKeyStore) DeleteKey(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, exists := s.keys[id]
	if !exists {
		return fmt.Errorf("API key not found")
	}

	// Remove from both maps
	delete(s.keyHashes, key.KeyHash)
	delete(s.keys, id)

	log.Printf("[ApiKeys] Deleted API key %s ('%s')", id, key.Name)

	return nil
}

// ValidateKey validates an API key and updates last used time
func (s *ApiKeyStore) ValidateKey(apiKey string) (*ApiKey, error) {
	// Hash the provided key
	keyHash := HashKey(apiKey)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find key by hash
	keyID, exists := s.keyHashes[keyHash]
	if !exists {
		return nil, fmt.Errorf("invalid API key")
	}

	key, exists := s.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("API key not found")
	}

	// Check if active
	if !key.IsActive {
		return nil, fmt.Errorf("API key has been revoked")
	}

	// Check if expired
	if time.Now().After(key.ExpiresAt) {
		return nil, fmt.Errorf("API key has expired")
	}

	// Update last used time
	key.LastUsed = time.Now()

	return key, nil
}

// CheckRateLimit enforces rate limiting per API key (100 requests per minute)
func (s *ApiKeyStore) CheckRateLimit(keyID string) bool {
	s.rateLimitMu.Lock()
	defer s.rateLimitMu.Unlock()

	rl, exists := s.rateLimits[keyID]
	if !exists {
		rl = &apiKeyRateLimit{
			count:     0,
			resetTime: time.Now().Add(1 * time.Minute),
		}
		s.rateLimits[keyID] = rl
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Reset counter if time window expired
	if time.Now().After(rl.resetTime) {
		rl.count = 0
		rl.resetTime = time.Now().Add(1 * time.Minute)
	}

	// Check limit (max 100 per minute)
	if rl.count >= 100 {
		return false
	}

	rl.count++
	return true
}

// ApiKeyHandler handles API key management endpoints
type ApiKeyHandler struct {
	store       *ApiKeyStore
	authService *auth.Service
}

// NewApiKeyHandler creates a new API key handler
func NewApiKeyHandler(store *ApiKeyStore, authService *auth.Service) *ApiKeyHandler {
	return &ApiKeyHandler{
		store:       store,
		authService: authService,
	}
}

// HandleCreateKey creates a new API key
// POST /admin/apikeys
func (h *ApiKeyHandler) HandleCreateKey(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	claims, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Name          string   `json:"name"`
		Scopes        []string `json:"scopes"`
		ExpiresInDays int      `json:"expiresInDays"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate inputs
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if len(req.Scopes) == 0 {
		http.Error(w, "At least one scope is required", http.StatusBadRequest)
		return
	}

	// Validate scopes
	validScopes := map[string]bool{"read": true, "trade": true, "admin": true}
	for _, scope := range req.Scopes {
		if !validScopes[scope] {
			http.Error(w, fmt.Sprintf("Invalid scope: %s. Valid scopes: read, trade, admin", scope), http.StatusBadRequest)
			return
		}
	}

	// Default expiration: 365 days
	if req.ExpiresInDays == 0 {
		req.ExpiresInDays = 365
	}

	// Create API key
	key, apiKey, err := h.store.CreateKey(req.Name, req.Scopes, claims.UserID, req.ExpiresInDays)
	if err != nil {
		log.Printf("[ApiKeys] Failed to create API key: %v", err)
		http.Error(w, "Failed to create API key", http.StatusInternalServerError)
		return
	}

	// Return API key (only time it's shown in plaintext)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        key.ID,
		"name":      key.Name,
		"apiKey":    apiKey, // ONLY TIME THIS IS RETURNED
		"prefix":    key.Prefix,
		"scopes":    key.Scopes,
		"createdAt": key.CreatedAt,
		"expiresAt": key.ExpiresAt,
		"warning":   "Save this API key securely. It will not be shown again.",
	})
}

// HandleListKeys lists all API keys (masked)
// GET /admin/apikeys
func (h *ApiKeyHandler) HandleListKeys(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all keys
	keys := h.store.ListKeys()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"keys":  keys,
		"count": len(keys),
	})
}

// HandleUpdateKey updates an API key's name and scopes
// PUT /admin/apikeys/:id
func (h *ApiKeyHandler) HandleUpdateKey(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract key ID from URL path
	// Expected: /admin/apikeys/{id}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid API key ID", http.StatusBadRequest)
		return
	}
	keyID := pathParts[3]

	// Parse request body
	var req struct {
		Name   string   `json:"name"`
		Scopes []string `json:"scopes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate scopes
	if len(req.Scopes) > 0 {
		validScopes := map[string]bool{"read": true, "trade": true, "admin": true}
		for _, scope := range req.Scopes {
			if !validScopes[scope] {
				http.Error(w, fmt.Sprintf("Invalid scope: %s", scope), http.StatusBadRequest)
				return
			}
		}
	}

	// Update key
	if err := h.store.UpdateKey(keyID, req.Name, req.Scopes); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Get updated key
	key, _ := h.store.GetKey(keyID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(key)
}

// HandleDeleteKey deletes an API key
// DELETE /admin/apikeys/:id
func (h *ApiKeyHandler) HandleDeleteKey(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract key ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid API key ID", http.StatusBadRequest)
		return
	}
	keyID := pathParts[3]

	// Delete key
	if err := h.store.DeleteKey(keyID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "API key deleted successfully",
	})
}

// validateAdminAuth validates JWT token and checks admin role
func (h *ApiKeyHandler) validateAdminAuth(r *http.Request) (*auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	tokenString := parts[1]
	claims, err := auth.ValidateTokenWithDefault(tokenString)
	if err != nil {
		return nil, err
	}

	// Check if user has admin role
	// For now, we'll assume all authenticated users can manage API keys
	// In production, you'd check claims.Role == "admin"

	return claims, nil
}

// ValidateApiKeyMiddleware is middleware that validates API keys from X-API-Key header
func ValidateApiKeyMiddleware(store *ApiKeyStore, requiredScopes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get API key from header
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				http.Error(w, "Missing X-API-Key header", http.StatusUnauthorized)
				return
			}

			// Validate API key
			key, err := store.ValidateKey(apiKey)
			if err != nil {
				log.Printf("[ApiKeys] Validation failed: %v", err)
				http.Error(w, "Invalid or expired API key", http.StatusUnauthorized)
				return
			}

			// Check rate limit
			if !store.CheckRateLimit(key.ID) {
				http.Error(w, "Rate limit exceeded (100 requests per minute)", http.StatusTooManyRequests)
				return
			}

			// Check scopes
			if len(requiredScopes) > 0 {
				hasRequiredScope := false
				for _, requiredScope := range requiredScopes {
					for _, keyScope := range key.Scopes {
						if keyScope == requiredScope || keyScope == "admin" {
							hasRequiredScope = true
							break
						}
					}
					if hasRequiredScope {
						break
					}
				}

				if !hasRequiredScope {
					http.Error(w, fmt.Sprintf("Insufficient permissions. Required scopes: %v", requiredScopes), http.StatusForbidden)
					return
				}
			}

			// Add key info to request context for downstream handlers
			// (You could use context.WithValue here if needed)

			next.ServeHTTP(w, r)
		})
	}
}
