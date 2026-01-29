package crm

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// WebhookHandler processes CRM webhooks
type WebhookHandler struct {
	secret string
	events []WebhookEvent
}

type WebhookEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Source    string                 `json:"source"`
}

type WebhookEventLog struct {
	Event     WebhookEvent
	Processed bool
	Error     string
	Timestamp time.Time
}

func NewWebhookHandler(secret string) *WebhookHandler {
	return &WebhookHandler{
		secret: secret,
		events: make([]WebhookEvent, 0),
	}
}

// ValidateSignature verifies webhook signature
func (h *WebhookHandler) ValidateSignature(payload []byte, signature string) bool {
	expectedSignature := h.calculateSignature(payload)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// calculateSignature generates HMAC signature
func (h *WebhookHandler) calculateSignature(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// HandleWebhook processes incoming webhook
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Read body
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	payload := buf.Bytes()

	// Verify signature
	signature := r.Header.Get("X-Webhook-Signature")
	if signature == "" {
		http.Error(w, "Missing signature", http.StatusUnauthorized)
		return
	}

	if !h.ValidateSignature(payload, signature) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse event
	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Process event
	h.events = append(h.events, event)

	// Return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "received",
		"id": event.ID,
	})
}

func (h *WebhookHandler) GetEvents() []WebhookEvent {
	return h.events
}

func (h *WebhookHandler) Reset() {
	h.events = make([]WebhookEvent, 0)
}

// Test cases
func TestWebhookSignatureValidation(t *testing.T) {
	t.Run("ValidSignature", func(t *testing.T) {
		handler := NewWebhookHandler("secret-key")

		payload := []byte(`{"id":"123","type":"contact.created","timestamp":1234567890}`)
		signature := handler.calculateSignature(payload)

		if !handler.ValidateSignature(payload, signature) {
			t.Error("Valid signature was rejected")
		}
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		handler := NewWebhookHandler("secret-key")

		payload := []byte(`{"id":"123","type":"contact.created","timestamp":1234567890}`)
		invalidSignature := "invalid-signature"

		if handler.ValidateSignature(payload, invalidSignature) {
			t.Error("Invalid signature was accepted")
		}
	})

	t.Run("DifferentSecret", func(t *testing.T) {
		handler1 := NewWebhookHandler("secret-key-1")
		handler2 := NewWebhookHandler("secret-key-2")

		payload := []byte(`{"id":"123","type":"contact.created"}`)
		signature1 := handler1.calculateSignature(payload)

		if handler2.ValidateSignature(payload, signature1) {
			t.Error("Signature from different secret was accepted")
		}
	})
}

func TestWebhookEventProcessing(t *testing.T) {
	t.Run("ProcessContactCreatedEvent", func(t *testing.T) {
		handler := NewWebhookHandler("secret-key")

		event := WebhookEvent{
			ID:        "evt-123",
			Type:      "contact.created",
			Timestamp: time.Now().Unix(),
			Data: map[string]interface{}{
				"contact_id": "cnt-456",
				"email": "test@example.com",
				"name": "John Doe",
			},
			Source: "hubspot",
		}

		payload, _ := json.Marshal(event)
		signature := handler.calculateSignature(payload)

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
		request.Header.Set("X-Webhook-Signature", signature)

		handler.HandleWebhook(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", recorder.Code)
		}

		events := handler.GetEvents()
		if len(events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(events))
		}

		if events[0].Type != "contact.created" {
			t.Errorf("Expected type 'contact.created', got '%s'", events[0].Type)
		}
	})

	t.Run("ProcessContactUpdatedEvent", func(t *testing.T) {
		handler := NewWebhookHandler("secret-key")

		event := WebhookEvent{
			ID:        "evt-124",
			Type:      "contact.updated",
			Timestamp: time.Now().Unix(),
			Data: map[string]interface{}{
				"contact_id": "cnt-456",
				"email": "newemail@example.com",
			},
			Source: "salesforce",
		}

		payload, _ := json.Marshal(event)
		signature := handler.calculateSignature(payload)

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
		request.Header.Set("X-Webhook-Signature", signature)

		handler.HandleWebhook(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("ProcessContactDeletedEvent", func(t *testing.T) {
		handler := NewWebhookHandler("secret-key")

		event := WebhookEvent{
			ID:        "evt-125",
			Type:      "contact.deleted",
			Timestamp: time.Now().Unix(),
			Data: map[string]interface{}{
				"contact_id": "cnt-456",
			},
			Source: "zoho",
		}

		payload, _ := json.Marshal(event)
		signature := handler.calculateSignature(payload)

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
		request.Header.Set("X-Webhook-Signature", signature)

		handler.HandleWebhook(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", recorder.Code)
		}
	})
}

func TestWebhookMissingSignature(t *testing.T) {
	handler := NewWebhookHandler("secret-key")

	event := WebhookEvent{
		ID:   "evt-123",
		Type: "contact.created",
	}

	payload, _ := json.Marshal(event)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	// No signature header

	handler.HandleWebhook(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", recorder.Code)
	}
}

func TestWebhookInvalidSignature(t *testing.T) {
	handler := NewWebhookHandler("secret-key")

	event := WebhookEvent{
		ID:   "evt-123",
		Type: "contact.created",
	}

	payload, _ := json.Marshal(event)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	request.Header.Set("X-Webhook-Signature", "invalid-signature")

	handler.HandleWebhook(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", recorder.Code)
	}
}

func TestWebhookInvalidJSON(t *testing.T) {
	handler := NewWebhookHandler("secret-key")

	invalidPayload := []byte(`{invalid json}`)
	signature := handler.calculateSignature(invalidPayload)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/webhook", bytes.NewReader(invalidPayload))
	request.Header.Set("X-Webhook-Signature", signature)

	handler.HandleWebhook(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", recorder.Code)
	}
}

func TestWebhookMultipleEvents(t *testing.T) {
	handler := NewWebhookHandler("secret-key")

	// Process multiple events
	events := []WebhookEvent{
		{
			ID:   "evt-1",
			Type: "contact.created",
			Data: map[string]interface{}{"contact_id": "c1"},
		},
		{
			ID:   "evt-2",
			Type: "contact.updated",
			Data: map[string]interface{}{"contact_id": "c2"},
		},
		{
			ID:   "evt-3",
			Type: "contact.deleted",
			Data: map[string]interface{}{"contact_id": "c3"},
		},
	}

	for _, event := range events {
		payload, _ := json.Marshal(event)
		signature := handler.calculateSignature(payload)

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
		request.Header.Set("X-Webhook-Signature", signature)

		handler.HandleWebhook(recorder, request)
	}

	processedEvents := handler.GetEvents()
	if len(processedEvents) != 3 {
		t.Errorf("Expected 3 events, got %d", len(processedEvents))
	}
}

func TestWebhookEventTypes(t *testing.T) {
	eventTypes := []string{
		"contact.created",
		"contact.updated",
		"contact.deleted",
		"deal.created",
		"deal.updated",
		"deal.closed",
		"company.created",
		"company.updated",
	}

	t.Run("SupportedEventTypes", func(t *testing.T) {
		handler := NewWebhookHandler("secret-key")

		for _, eventType := range eventTypes {
			event := WebhookEvent{
				ID:        "evt-test",
				Type:      eventType,
				Timestamp: time.Now().Unix(),
				Data:      map[string]interface{}{},
				Source:    "test",
			}

			payload, _ := json.Marshal(event)
			signature := handler.calculateSignature(payload)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
			request.Header.Set("X-Webhook-Signature", signature)

			handler.HandleWebhook(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Errorf("Failed to process event type '%s': status %d", eventType, recorder.Code)
			}
		}
	})
}

func TestWebhookEventIDUniqueness(t *testing.T) {
	handler := NewWebhookHandler("secret-key")

	// Try to process event with duplicate ID
	event := WebhookEvent{
		ID:   "evt-duplicate",
		Type: "contact.created",
	}

	for i := 0; i < 2; i++ {
		payload, _ := json.Marshal(event)
		signature := handler.calculateSignature(payload)

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
		request.Header.Set("X-Webhook-Signature", signature)

		handler.HandleWebhook(recorder, request)
	}

	events := handler.GetEvents()
	if len(events) != 2 {
		t.Errorf("Expected 2 events (no deduplication in mock), got %d", len(events))
	}
}

func TestWebhookTimestampValidation(t *testing.T) {
	t.Run("ValidTimestamp", func(t *testing.T) {
		handler := NewWebhookHandler("secret-key")

		event := WebhookEvent{
			ID:        "evt-123",
			Type:      "contact.created",
			Timestamp: time.Now().Unix(),
			Data:      map[string]interface{}{},
		}

		payload, _ := json.Marshal(event)
		signature := handler.calculateSignature(payload)

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
		request.Header.Set("X-Webhook-Signature", signature)

		handler.HandleWebhook(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Errorf("Failed to process valid timestamp: status %d", recorder.Code)
		}
	})

	t.Run("OldTimestamp", func(t *testing.T) {
		handler := NewWebhookHandler("secret-key")

		// Timestamp from 1 hour ago
		oldTimestamp := time.Now().Add(-1 * time.Hour).Unix()

		event := WebhookEvent{
			ID:        "evt-old",
			Type:      "contact.created",
			Timestamp: oldTimestamp,
			Data:      map[string]interface{}{},
		}

		payload, _ := json.Marshal(event)
		signature := handler.calculateSignature(payload)

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
		request.Header.Set("X-Webhook-Signature", signature)

		// In real scenario, this might be rejected as too old
		handler.HandleWebhook(recorder, request)

		// Should still process but may log warning
		t.Logf("Processed old event: status %d", recorder.Code)
	})
}

func TestWebhookSourceTracking(t *testing.T) {
	handler := NewWebhookHandler("secret-key")

	sources := []string{"hubspot", "salesforce", "zoho"}

	for _, source := range sources {
		event := WebhookEvent{
			ID:     fmt.Sprintf("evt-%s", source),
			Type:   "contact.created",
			Source: source,
			Data:   map[string]interface{}{},
		}

		payload, _ := json.Marshal(event)
		signature := handler.calculateSignature(payload)

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
		request.Header.Set("X-Webhook-Signature", signature)

		handler.HandleWebhook(recorder, request)
	}

	events := handler.GetEvents()
	if len(events) != len(sources) {
		t.Errorf("Expected %d events, got %d", len(sources), len(events))
	}

	// Verify sources are tracked
	sourceMap := make(map[string]int)
	for _, event := range events {
		sourceMap[event.Source]++
	}

	if len(sourceMap) != 3 {
		t.Errorf("Expected 3 unique sources, got %d", len(sourceMap))
	}
}
