package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2 parameters for key derivation
	argon2Time    = 1
	argon2Memory  = 64 * 1024
	argon2Threads = 4
	argon2KeyLen  = 32
)

// EncryptionService handles data encryption at rest
type EncryptionService struct {
	masterKey []byte
}

// NewEncryptionService creates a new encryption service
func NewEncryptionService(passphrase string) *EncryptionService {
	// Derive master key from passphrase using Argon2
	salt := sha256.Sum256([]byte("rtx-trading-engine-salt-v1"))
	masterKey := argon2.IDKey(
		[]byte(passphrase),
		salt[:],
		argon2Time,
		argon2Memory,
		argon2Threads,
		argon2KeyLen,
	)

	return &EncryptionService{
		masterKey: masterKey,
	}
}

// Encrypt encrypts plaintext using AES-256-GCM
func (e *EncryptionService) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (e *EncryptionService) Decrypt(ciphertext string) ([]byte, error) {
	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// EncryptString encrypts a string
func (e *EncryptionService) EncryptString(plaintext string) (string, error) {
	return e.Encrypt([]byte(plaintext))
}

// DecryptString decrypts to a string
func (e *EncryptionService) DecryptString(ciphertext string) (string, error) {
	plaintext, err := e.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// EncryptSensitiveData encrypts sensitive trading data
type SensitiveData struct {
	APIKeys      map[string]string
	Credentials  map[string]string
	BankDetails  map[string]string
}

// EncryptSensitiveData encrypts all sensitive data fields
func (e *EncryptionService) EncryptSensitiveData(data *SensitiveData) (*SensitiveData, error) {
	encrypted := &SensitiveData{
		APIKeys:     make(map[string]string),
		Credentials: make(map[string]string),
		BankDetails: make(map[string]string),
	}

	// Encrypt API keys
	for k, v := range data.APIKeys {
		encV, err := e.EncryptString(v)
		if err != nil {
			return nil, err
		}
		encrypted.APIKeys[k] = encV
	}

	// Encrypt credentials
	for k, v := range data.Credentials {
		encV, err := e.EncryptString(v)
		if err != nil {
			return nil, err
		}
		encrypted.Credentials[k] = encV
	}

	// Encrypt bank details
	for k, v := range data.BankDetails {
		encV, err := e.EncryptString(v)
		if err != nil {
			return nil, err
		}
		encrypted.BankDetails[k] = encV
	}

	return encrypted, nil
}

// DecryptSensitiveData decrypts all sensitive data fields
func (e *EncryptionService) DecryptSensitiveData(encrypted *SensitiveData) (*SensitiveData, error) {
	data := &SensitiveData{
		APIKeys:     make(map[string]string),
		Credentials: make(map[string]string),
		BankDetails: make(map[string]string),
	}

	// Decrypt API keys
	for k, v := range encrypted.APIKeys {
		decV, err := e.DecryptString(v)
		if err != nil {
			return nil, err
		}
		data.APIKeys[k] = decV
	}

	// Decrypt credentials
	for k, v := range encrypted.Credentials {
		decV, err := e.DecryptString(v)
		if err != nil {
			return nil, err
		}
		data.Credentials[k] = decV
	}

	// Decrypt bank details
	for k, v := range encrypted.BankDetails {
		decV, err := e.DecryptString(v)
		if err != nil {
			return nil, err
		}
		data.BankDetails[k] = decV
	}

	return data, nil
}

// GenerateRandomKey generates a cryptographically secure random key
func GenerateRandomKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HashSHA256 creates a SHA-256 hash of the input
func HashSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// SecureCompare performs constant-time comparison to prevent timing attacks
func SecureCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	result := byte(0)
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}
