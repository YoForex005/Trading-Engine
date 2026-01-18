package notifications

import (
	"context"
	"sync"
	"time"
)

// BatchProcessor handles batching of notifications for efficiency
type BatchProcessor struct {
	manager      *NotificationManager
	batches      map[string]*batchBuffer
	mu           sync.RWMutex
	batchSize    int
	batchTimeout time.Duration
}

// batchBuffer holds notifications for a user+channel combination
type batchBuffer struct {
	userID        string
	channel       NotificationChannel
	notifications []Notification
	timer         *time.Timer
	mu            sync.Mutex
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(manager *NotificationManager) *BatchProcessor {
	return &BatchProcessor{
		manager:      manager,
		batches:      make(map[string]*batchBuffer),
		batchSize:    10,  // Batch up to 10 notifications
		batchTimeout: 5 * time.Minute, // Or send after 5 minutes
	}
}

// AddToBatch adds a notification to the batch
func (bp *BatchProcessor) AddToBatch(userID string, channel NotificationChannel, notif Notification) {
	bufferKey := userID + ":" + string(channel)

	bp.mu.Lock()
	buffer, exists := bp.batches[bufferKey]
	if !exists {
		buffer = &batchBuffer{
			userID:        userID,
			channel:       channel,
			notifications: make([]Notification, 0, bp.batchSize),
		}
		bp.batches[bufferKey] = buffer
	}
	bp.mu.Unlock()

	buffer.mu.Lock()
	defer buffer.mu.Unlock()

	buffer.notifications = append(buffer.notifications, notif)

	// Set timer for timeout if not already set
	if buffer.timer == nil {
		buffer.timer = time.AfterFunc(bp.batchTimeout, func() {
			bp.flushBuffer(bufferKey)
		})
	}

	// Flush if batch size reached
	if len(buffer.notifications) >= bp.batchSize {
		bp.flushBuffer(bufferKey)
	}
}

// flushBuffer sends all notifications in a buffer
func (bp *BatchProcessor) flushBuffer(bufferKey string) {
	bp.mu.Lock()
	buffer, exists := bp.batches[bufferKey]
	if !exists {
		bp.mu.Unlock()
		return
	}
	delete(bp.batches, bufferKey)
	bp.mu.Unlock()

	buffer.mu.Lock()
	defer buffer.mu.Unlock()

	if buffer.timer != nil {
		buffer.timer.Stop()
	}

	if len(buffer.notifications) == 0 {
		return
	}

	// Send batched notifications
	ctx := context.Background()
	bp.sendBatch(ctx, buffer)
}

// sendBatch sends a batch of notifications
func (bp *BatchProcessor) sendBatch(ctx context.Context, buffer *batchBuffer) {
	// For email, we can combine into a digest
	if buffer.channel == ChannelEmail {
		bp.sendEmailDigest(ctx, buffer)
		return
	}

	// For other channels, send individually but concurrently
	var wg sync.WaitGroup
	for _, notif := range buffer.notifications {
		wg.Add(1)
		go func(n Notification) {
			defer wg.Done()
			// Send individual notification
			// In production, get user contacts from database
		}(notif)
	}
	wg.Wait()
}

// sendEmailDigest sends multiple notifications as a single digest email
func (bp *BatchProcessor) sendEmailDigest(ctx context.Context, buffer *batchBuffer) {
	if len(buffer.notifications) == 0 {
		return
	}

	// Group notifications by type
	grouped := make(map[NotificationType][]Notification)
	for _, notif := range buffer.notifications {
		grouped[notif.Type] = append(grouped[notif.Type], notif)
	}

	// Build digest notification
	digest := &Notification{
		ID:       generateID(),
		UserID:   buffer.userID,
		Type:     "digest",
		Priority: PriorityNormal,
		Subject:  "Trading Activity Summary",
		Message:  bp.buildDigestMessage(grouped),
		Data: map[string]interface{}{
			"count":        len(buffer.notifications),
			"notification_types": grouped,
		},
		Channels:  []NotificationChannel{ChannelEmail},
		CreatedAt: time.Now(),
	}

	// Send digest (in production, get user email from database)
	// bp.manager.emailNotifier.Send(ctx, digest, userEmail)
	_ = digest // Placeholder
}

// buildDigestMessage builds a summary message from grouped notifications
func (bp *BatchProcessor) buildDigestMessage(grouped map[NotificationType][]Notification) string {
	message := "You have new trading activity:\n\n"

	for notifType, notifications := range grouped {
		message += string(notifType) + ": " + string(rune(len(notifications))) + " notifications\n"
	}

	return message
}

// ProcessBatch processes a batch of notifications
func (bp *BatchProcessor) ProcessBatch(ctx context.Context, notifications []Notification, contacts map[string]UserContacts) error {
	// Group by user and channel
	type batchKey struct {
		userID  string
		channel NotificationChannel
	}

	batches := make(map[batchKey][]Notification)

	for _, notif := range notifications {
		// Get channels for this notification
		channels, err := bp.manager.prefsManager.GetChannelsForNotification(ctx, notif.UserID, &notif)
		if err != nil {
			continue
		}

		for _, channel := range channels {
			key := batchKey{userID: notif.UserID, channel: channel}
			batches[key] = append(batches[key], notif)
		}
	}

	// Process each batch concurrently
	var wg sync.WaitGroup
	for key, batch := range batches {
		wg.Add(1)
		go func(k batchKey, b []Notification) {
			defer wg.Done()

			buffer := &batchBuffer{
				userID:        k.userID,
				channel:       k.channel,
				notifications: b,
			}

			bp.sendBatch(ctx, buffer)
		}(key, batch)
	}

	wg.Wait()
	return nil
}

// FlushAll flushes all pending batches
func (bp *BatchProcessor) FlushAll() {
	bp.mu.Lock()
	buffers := make(map[string]*batchBuffer)
	for k, v := range bp.batches {
		buffers[k] = v
	}
	bp.batches = make(map[string]*batchBuffer)
	bp.mu.Unlock()

	for key := range buffers {
		bp.flushBuffer(key)
	}
}
