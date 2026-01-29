package notifications

import (
	"context"
	"math"
	"sync"
	"time"
)

// RetryQueue manages failed notifications for retry
type RetryQueue struct {
	manager       *NotificationManager
	deliveryStore DeliveryStore
	queue         []*DeliveryRecord
	mu            sync.RWMutex
	config        RetryConfig
}

// NewRetryQueue creates a new retry queue
func NewRetryQueue(manager *NotificationManager, deliveryStore DeliveryStore) *RetryQueue {
	rq := &RetryQueue{
		manager:       manager,
		deliveryStore: deliveryStore,
		queue:         make([]*DeliveryRecord, 0),
		config:        DefaultRetryConfig(),
	}

	// Start background processor
	go rq.startProcessor()

	return rq
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:   5,
		InitialDelay:  1 * time.Second,
		MaxDelay:      5 * time.Minute,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"timeout",
			"connection refused",
			"temporary failure",
			"rate limit",
		},
	}
}

// Add adds a failed delivery record to the retry queue
func (rq *RetryQueue) Add(record *DeliveryRecord) {
	if record.Attempts >= rq.config.MaxAttempts {
		// Max attempts reached, don't retry
		return
	}

	rq.mu.Lock()
	defer rq.mu.Unlock()

	record.Status = StatusRetrying
	rq.queue = append(rq.queue, record)
}

// ProcessRetries processes all pending retries
func (rq *RetryQueue) ProcessRetries(ctx context.Context) error {
	rq.mu.Lock()
	pending := make([]*DeliveryRecord, len(rq.queue))
	copy(pending, rq.queue)
	rq.queue = rq.queue[:0] // Clear queue
	rq.mu.Unlock()

	for _, record := range pending {
		if err := rq.retryDelivery(ctx, record); err != nil {
			// If retry failed and attempts not exhausted, add back to queue
			if record.Attempts < rq.config.MaxAttempts {
				rq.Add(record)
			}
		}
	}

	return nil
}

// retryDelivery retries a single delivery
func (rq *RetryQueue) retryDelivery(ctx context.Context, record *DeliveryRecord) error {
	// Calculate backoff delay
	delay := rq.calculateBackoff(record.Attempts)

	// Wait for backoff duration
	if record.LastAttemptAt != nil {
		elapsed := time.Since(*record.LastAttemptAt)
		if elapsed < delay {
			time.Sleep(delay - elapsed)
		}
	}

	// Increment attempt counter
	record.Attempts++
	now := time.Now()
	record.LastAttemptAt = &now
	record.UpdatedAt = now

	// Get original notification (in production, fetch from database)
	// For now, we'll just update the delivery record
	// TODO: Implement notification retrieval and re-send logic

	// Update status in store
	if err := rq.deliveryStore.Save(ctx, record); err != nil {
		return err
	}

	return nil
}

// calculateBackoff calculates exponential backoff delay
func (rq *RetryQueue) calculateBackoff(attempts int) time.Duration {
	delay := float64(rq.config.InitialDelay) * math.Pow(rq.config.BackoffFactor, float64(attempts-1))

	if time.Duration(delay) > rq.config.MaxDelay {
		return rq.config.MaxDelay
	}

	return time.Duration(delay)
}

// startProcessor starts background retry processor
func (rq *RetryQueue) startProcessor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		rq.ProcessRetries(ctx)
	}
}

// GetQueueSize returns the current queue size
func (rq *RetryQueue) GetQueueSize() int {
	rq.mu.RLock()
	defer rq.mu.RUnlock()
	return len(rq.queue)
}

// Clear clears the retry queue
func (rq *RetryQueue) Clear() {
	rq.mu.Lock()
	defer rq.mu.Unlock()
	rq.queue = rq.queue[:0]
}
