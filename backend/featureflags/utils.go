package featureflags

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// generateID generates a unique ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Helper function to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}

// Helper function to create float64 pointers
func float64Ptr(f float64) *float64 {
	return &f
}
