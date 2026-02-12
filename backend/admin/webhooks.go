package admin

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// WebhookEvent defines supported webhook events
type WebhookEvent string

const (
	EventTradeOpened    WebhookEvent = "trade.opened"
	EventTradeClosed    WebhookEvent = "trade.closed"
	EventDeposit        WebhookEvent = "deposit"
	EventWithdrawal     WebhookEvent = "withdrawal"
	EventMarginCall     WebhookEvent = "margin_call"
	EventStopOut        WebhookEvent = "stop_out"
)

// WebhookConfig represents a webhook configuration
type WebhookConfig struct {
	ID         int64          `json:"id"`
	URL        string         `json:"url"`
	Events     []WebhookEvent `json:"events"`
	Secret     string         `json:"secret"`     // HMAC-SHA256 signing secret
	IsActive   bool           `json:"isActive"`
	RetryCount int            `json:"retryCount"` // Max retries (default: 3)
	CreatedAt  time.Time      `json:"createdAt"`
	CreatedBy  string         `json:"createdBy"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID           int64        `json:"id"`
	WebhookID    int64        `json:"webhookId"`
	Event        WebhookEvent `json:"event"`
	Payload      interface{}  `json:"payload"`
	Status       string       `json:"status"` // PENDING, SUCCESS, FAILED
	StatusCode   int          `json:"statusCode,omitempty"`
	ErrorMessage string       `json:"errorMessage,omitempty"`
	AttemptCount int          `json:"attemptCount"`
	NextRetryAt  *time.Time   `json:"nextRetryAt,omitempty"`
	CreatedAt    time.Time    `json:"createdAt"`
	CompletedAt  *time.Time   `json:"completedAt,omitempty"`
}

// WebhookPayload is the structure sent to webhook endpoints
type WebhookPayload struct {
	Event     WebhookEvent `json:"event"`
	Timestamp time.Time    `json:"timestamp"`
	Data      interface{}  `json:"data"`
}

// WebhookService manages webhook configurations and delivery
type WebhookService struct {
	mu            sync.RWMutex
	webhooks      map[int64]*WebhookConfig
	deliveries    map[int64]*WebhookDelivery
	nextWebhookID int64
	nextDeliveryID int64
	deliveryQueue chan *WebhookDelivery
	stopWorker    chan bool
	workerRunning bool
}

// NewWebhookService creates a new webhook service
func NewWebhookService() *WebhookService {
	svc := &WebhookService{
		webhooks:      make(map[int64]*WebhookConfig),
		deliveries:    make(map[int64]*WebhookDelivery),
		nextWebhookID: 1,
		nextDeliveryID: 1,
		deliveryQueue: make(chan *WebhookDelivery, 1000),
		stopWorker:    make(chan bool),
	}

	// Start delivery worker
	svc.startDeliveryWorker()

	log.Println("[Webhooks] Webhook service initialized with delivery worker")
	return svc
}

// CreateWebhook creates a new webhook configuration
func (ws *WebhookService) CreateWebhook(url string, events []WebhookEvent, secret string, createdBy string) (*WebhookConfig, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Validate URL
	if url == "" {
		return nil, errors.New("webhook URL is required")
	}

	// Validate events
	if len(events) == 0 {
		return nil, errors.New("at least one event is required")
	}

	// Generate secret if not provided
	if secret == "" {
		secret = generateWebhookSecret()
	}

	webhook := &WebhookConfig{
		ID:         ws.nextWebhookID,
		URL:        url,
		Events:     events,
		Secret:     secret,
		IsActive:   true,
		RetryCount: 3, // Default: 3 retries
		CreatedAt:  time.Now(),
		CreatedBy:  createdBy,
		UpdatedAt:  time.Now(),
	}

	ws.nextWebhookID++
	ws.webhooks[webhook.ID] = webhook

	log.Printf("[Webhooks] Webhook created: ID=%d, URL=%s, Events=%v", webhook.ID, webhook.URL, webhook.Events)
	return webhook, nil
}

// UpdateWebhook updates an existing webhook configuration
func (ws *WebhookService) UpdateWebhook(id int64, url *string, events []WebhookEvent, secret *string, isActive *bool, retryCount *int) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	webhook, exists := ws.webhooks[id]
	if !exists {
		return errors.New("webhook not found")
	}

	if url != nil && *url != "" {
		webhook.URL = *url
	}

	if len(events) > 0 {
		webhook.Events = events
	}

	if secret != nil && *secret != "" {
		webhook.Secret = *secret
	}

	if isActive != nil {
		webhook.IsActive = *isActive
	}

	if retryCount != nil && *retryCount >= 0 {
		webhook.RetryCount = *retryCount
	}

	webhook.UpdatedAt = time.Now()

	log.Printf("[Webhooks] Webhook updated: ID=%d", id)
	return nil
}

// DeleteWebhook deletes a webhook configuration
func (ws *WebhookService) DeleteWebhook(id int64) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if _, exists := ws.webhooks[id]; !exists {
		return errors.New("webhook not found")
	}

	delete(ws.webhooks, id)

	log.Printf("[Webhooks] Webhook deleted: ID=%d", id)
	return nil
}

// GetWebhook retrieves a webhook configuration
func (ws *WebhookService) GetWebhook(id int64) (*WebhookConfig, error) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	webhook, exists := ws.webhooks[id]
	if !exists {
		return nil, errors.New("webhook not found")
	}

	return webhook, nil
}

// ListWebhooks returns all webhook configurations
func (ws *WebhookService) ListWebhooks() []*WebhookConfig {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	webhooks := make([]*WebhookConfig, 0, len(ws.webhooks))
	for _, webhook := range ws.webhooks {
		webhooks = append(webhooks, webhook)
	}

	return webhooks
}

// GetDeliveries returns delivery history for a webhook
func (ws *WebhookService) GetDeliveries(webhookID int64, limit int) []*WebhookDelivery {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	deliveries := make([]*WebhookDelivery, 0)
	for _, delivery := range ws.deliveries {
		if delivery.WebhookID == webhookID {
			deliveries = append(deliveries, delivery)
			if limit > 0 && len(deliveries) >= limit {
				break
			}
		}
	}

	return deliveries
}

// TriggerEvent triggers a webhook event for all active webhooks subscribed to it
func (ws *WebhookService) TriggerEvent(event WebhookEvent, data interface{}) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	for _, webhook := range ws.webhooks {
		if !webhook.IsActive {
			continue
		}

		// Check if webhook is subscribed to this event
		subscribed := false
		for _, subscribedEvent := range webhook.Events {
			if subscribedEvent == event {
				subscribed = true
				break
			}
		}

		if !subscribed {
			continue
		}

		// Create delivery record
		delivery := &WebhookDelivery{
			ID:           ws.nextDeliveryID,
			WebhookID:    webhook.ID,
			Event:        event,
			Payload:      data,
			Status:       "PENDING",
			AttemptCount: 0,
			CreatedAt:    time.Now(),
		}

		ws.nextDeliveryID++
		ws.deliveries[delivery.ID] = delivery

		// Queue for delivery
		go func(d *WebhookDelivery) {
			ws.deliveryQueue <- d
		}(delivery)

		log.Printf("[Webhooks] Event triggered: %s -> Webhook %d (delivery %d)", event, webhook.ID, delivery.ID)
	}
}

// startDeliveryWorker starts the background worker that processes webhook deliveries
func (ws *WebhookService) startDeliveryWorker() {
	ws.mu.Lock()
	if ws.workerRunning {
		ws.mu.Unlock()
		return
	}
	ws.workerRunning = true
	ws.mu.Unlock()

	go func() {
		log.Println("[Webhooks] Delivery worker started")

		// Retry ticker for handling exponential backoff
		retryTicker := time.NewTicker(1 * time.Second)
		defer retryTicker.Stop()

		for {
			select {
			case delivery := <-ws.deliveryQueue:
				ws.processDelivery(delivery)

			case <-retryTicker.C:
				// Check for deliveries that need retry
				ws.processRetries()

			case <-ws.stopWorker:
				log.Println("[Webhooks] Delivery worker stopped")
				ws.mu.Lock()
				ws.workerRunning = false
				ws.mu.Unlock()
				return
			}
		}
	}()
}

// processDelivery sends a webhook delivery
func (ws *WebhookService) processDelivery(delivery *WebhookDelivery) {
	ws.mu.RLock()
	webhook, exists := ws.webhooks[delivery.WebhookID]
	ws.mu.RUnlock()

	if !exists || !webhook.IsActive {
		ws.mu.Lock()
		delivery.Status = "FAILED"
		delivery.ErrorMessage = "Webhook not found or inactive"
		now := time.Now()
		delivery.CompletedAt = &now
		ws.mu.Unlock()
		return
	}

	// Create payload
	payload := WebhookPayload{
		Event:     delivery.Event,
		Timestamp: time.Now(),
		Data:      delivery.Payload,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ws.mu.Lock()
		delivery.Status = "FAILED"
		delivery.ErrorMessage = fmt.Sprintf("Failed to marshal payload: %v", err)
		now := time.Now()
		delivery.CompletedAt = &now
		ws.mu.Unlock()
		return
	}

	// Generate HMAC-SHA256 signature
	signature := generateHMACSignature(payloadBytes, webhook.Secret)

	// Send HTTP request
	req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		ws.scheduleRetry(delivery, fmt.Sprintf("Failed to create request: %v", err))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Event", string(delivery.Event))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		ws.scheduleRetry(delivery, fmt.Sprintf("Request failed: %v", err))
		return
	}
	defer resp.Body.Close()

	ws.mu.Lock()
	delivery.AttemptCount++
	delivery.StatusCode = resp.StatusCode

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Success
		delivery.Status = "SUCCESS"
		now := time.Now()
		delivery.CompletedAt = &now
		log.Printf("[Webhooks] Delivery success: ID=%d, Webhook=%d, Event=%s, StatusCode=%d",
			delivery.ID, delivery.WebhookID, delivery.Event, resp.StatusCode)
	} else {
		// HTTP error - schedule retry
		ws.mu.Unlock()
		ws.scheduleRetry(delivery, fmt.Sprintf("HTTP %d", resp.StatusCode))
		return
	}

	ws.mu.Unlock()
}

// scheduleRetry schedules a delivery for retry with exponential backoff
func (ws *WebhookService) scheduleRetry(delivery *WebhookDelivery, errorMsg string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	webhook, exists := ws.webhooks[delivery.WebhookID]
	if !exists {
		delivery.Status = "FAILED"
		delivery.ErrorMessage = "Webhook not found"
		now := time.Now()
		delivery.CompletedAt = &now
		return
	}

	delivery.AttemptCount++
	delivery.ErrorMessage = errorMsg

	if delivery.AttemptCount >= webhook.RetryCount {
		// Max retries reached
		delivery.Status = "FAILED"
		now := time.Now()
		delivery.CompletedAt = &now
		log.Printf("[Webhooks] Delivery failed after %d attempts: ID=%d, Error=%s",
			delivery.AttemptCount, delivery.ID, errorMsg)
		return
	}

	// Calculate next retry time with exponential backoff
	// Retry delays: 1s, 5s, 30s, 5min
	retryDelays := []time.Duration{
		1 * time.Second,
		5 * time.Second,
		30 * time.Second,
		5 * time.Minute,
	}

	delayIndex := delivery.AttemptCount - 1
	if delayIndex >= len(retryDelays) {
		delayIndex = len(retryDelays) - 1
	}

	nextRetry := time.Now().Add(retryDelays[delayIndex])
	delivery.NextRetryAt = &nextRetry
	delivery.Status = "PENDING"

	log.Printf("[Webhooks] Delivery retry scheduled: ID=%d, Attempt=%d/%d, NextRetry=%s",
		delivery.ID, delivery.AttemptCount, webhook.RetryCount, nextRetry.Format(time.RFC3339))
}

// processRetries checks for deliveries that need to be retried
func (ws *WebhookService) processRetries() {
	ws.mu.RLock()
	now := time.Now()
	retriesToProcess := make([]*WebhookDelivery, 0)

	for _, delivery := range ws.deliveries {
		if delivery.Status == "PENDING" && delivery.NextRetryAt != nil && now.After(*delivery.NextRetryAt) {
			retriesToProcess = append(retriesToProcess, delivery)
		}
	}
	ws.mu.RUnlock()

	// Process retries
	for _, delivery := range retriesToProcess {
		ws.mu.Lock()
		delivery.NextRetryAt = nil // Clear next retry time
		ws.mu.Unlock()

		go func(d *WebhookDelivery) {
			ws.deliveryQueue <- d
		}(delivery)
	}
}

// StopWorker stops the delivery worker
func (ws *WebhookService) StopWorker() {
	ws.mu.Lock()
	if !ws.workerRunning {
		ws.mu.Unlock()
		return
	}
	ws.mu.Unlock()

	close(ws.stopWorker)
}

// Helper functions

// generateWebhookSecret generates a random secret for webhook signing
func generateWebhookSecret() string {
	// Generate 32-byte random secret
	secret := make([]byte, 32)
	// Use timestamp as seed for deterministic generation in demo
	timestamp := time.Now().UnixNano()
	for i := range secret {
		secret[i] = byte((timestamp >> (i * 8)) & 0xFF)
	}
	return hex.EncodeToString(secret)
}

// generateHMACSignature generates HMAC-SHA256 signature for payload
func generateHMACSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := mac.Sum(nil)
	return hex.EncodeToString(signature)
}

// VerifyWebhookSignature verifies the HMAC signature of a webhook payload
func VerifyWebhookSignature(payload []byte, signature string, secret string) bool {
	expectedSignature := generateHMACSignature(payload, secret)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
