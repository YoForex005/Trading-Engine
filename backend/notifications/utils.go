package notifications

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// generateID generates a unique ID for notifications and delivery records
func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// InMemoryDeliveryStore implements an in-memory delivery store
type InMemoryDeliveryStore struct {
	records map[string]*DeliveryRecord
	mu      sync.RWMutex
}

// NewInMemoryDeliveryStore creates a new in-memory delivery store
func NewInMemoryDeliveryStore() *InMemoryDeliveryStore {
	return &InMemoryDeliveryStore{
		records: make(map[string]*DeliveryRecord),
	}
}

// Save saves a delivery record
func (s *InMemoryDeliveryStore) Save(ctx context.Context, record *DeliveryRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record.UpdatedAt = time.Now()
	s.records[record.ID] = record
	return nil
}

// Get retrieves a delivery record by ID
func (s *InMemoryDeliveryStore) Get(ctx context.Context, id string) (*DeliveryRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, exists := s.records[id]
	if !exists {
		return nil, fmt.Errorf("delivery record not found: %s", id)
	}

	return record, nil
}

// GetByNotification retrieves all delivery records for a notification
func (s *InMemoryDeliveryStore) GetByNotification(ctx context.Context, notificationID string) ([]*DeliveryRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var records []*DeliveryRecord
	for _, record := range s.records {
		if record.NotificationID == notificationID {
			records = append(records, record)
		}
	}

	return records, nil
}

// GetByUser retrieves delivery records for a user
func (s *InMemoryDeliveryStore) GetByUser(ctx context.Context, userID string, limit int) ([]*DeliveryRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var records []*DeliveryRecord
	for _, record := range s.records {
		if record.UserID == userID {
			records = append(records, record)
			if len(records) >= limit {
				break
			}
		}
	}

	return records, nil
}

// UpdateStatus updates the status of a delivery record
func (s *InMemoryDeliveryStore) UpdateStatus(ctx context.Context, id string, status DeliveryStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, exists := s.records[id]
	if !exists {
		return fmt.Errorf("delivery record not found: %s", id)
	}

	record.Status = status
	record.UpdatedAt = time.Now()

	return nil
}
