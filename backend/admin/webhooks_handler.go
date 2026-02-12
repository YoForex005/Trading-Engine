package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// WebhooksHandler provides HTTP handlers for webhook management
type WebhooksHandler struct {
	webhookSvc *WebhookService
	authSvc    *AuthService
}

// NewWebhooksHandler creates a new webhooks handler
func NewWebhooksHandler(webhookSvc *WebhookService, authSvc *AuthService) *WebhooksHandler {
	return &WebhooksHandler{
		webhookSvc: webhookSvc,
		authSvc:    authSvc,
	}
}

// ListWebhooks handles GET /admin/webhooks - returns all webhooks
func (wh *WebhooksHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := wh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("[WebhooksHandler] Admin %s listing webhooks", admin.Username)

	webhooks := wh.webhookSvc.ListWebhooks()
	respondJSON(w, map[string]interface{}{
		"success":  true,
		"webhooks": webhooks,
		"count":    len(webhooks),
	})
}

// GetWebhook handles GET /admin/webhooks/:id - returns a specific webhook
func (wh *WebhooksHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := wh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract webhook ID from path
	webhookID, err := wh.extractWebhookID(r)
	if err != nil {
		respondError(w, "Invalid webhook ID", http.StatusBadRequest)
		return
	}

	webhook, err := wh.webhookSvc.GetWebhook(webhookID)
	if err != nil {
		respondError(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("[WebhooksHandler] Admin %s retrieved webhook %d", admin.Username, webhookID)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"webhook": webhook,
	})
}

// CreateWebhook handles POST /admin/webhooks - creates a new webhook
func (wh *WebhooksHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := wh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		URL        string         `json:"url"`
		Events     []WebhookEvent `json:"events"`
		Secret     string         `json:"secret,omitempty"`
		RetryCount int            `json:"retryCount,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.URL == "" {
		respondError(w, "URL is required", http.StatusBadRequest)
		return
	}

	if len(req.Events) == 0 {
		respondError(w, "At least one event is required", http.StatusBadRequest)
		return
	}

	webhook, err := wh.webhookSvc.CreateWebhook(
		req.URL,
		req.Events,
		req.Secret,
		admin.Username,
	)

	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update retry count if provided
	if req.RetryCount > 0 {
		wh.webhookSvc.UpdateWebhook(webhook.ID, nil, nil, nil, nil, &req.RetryCount)
	}

	log.Printf("[WebhooksHandler] Admin %s created webhook %d (%s)", admin.Username, webhook.ID, webhook.URL)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"webhook": webhook,
		"message": "Webhook created successfully",
	})
}

// UpdateWebhook handles PUT /admin/webhooks/:id - updates an existing webhook
func (wh *WebhooksHandler) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := wh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract webhook ID from path
	webhookID, err := wh.extractWebhookID(r)
	if err != nil {
		respondError(w, "Invalid webhook ID", http.StatusBadRequest)
		return
	}

	var req struct {
		URL        *string        `json:"url,omitempty"`
		Events     []WebhookEvent `json:"events,omitempty"`
		Secret     *string        `json:"secret,omitempty"`
		IsActive   *bool          `json:"isActive,omitempty"`
		RetryCount *int           `json:"retryCount,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = wh.webhookSvc.UpdateWebhook(
		webhookID,
		req.URL,
		req.Events,
		req.Secret,
		req.IsActive,
		req.RetryCount,
	)

	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch updated webhook
	webhook, _ := wh.webhookSvc.GetWebhook(webhookID)

	log.Printf("[WebhooksHandler] Admin %s updated webhook %d", admin.Username, webhookID)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"webhook": webhook,
		"message": "Webhook updated successfully",
	})
}

// DeleteWebhook handles DELETE /admin/webhooks/:id - deletes a webhook
func (wh *WebhooksHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodDelete {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := wh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract webhook ID from path
	webhookID, err := wh.extractWebhookID(r)
	if err != nil {
		respondError(w, "Invalid webhook ID", http.StatusBadRequest)
		return
	}

	err = wh.webhookSvc.DeleteWebhook(webhookID)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[WebhooksHandler] Admin %s deleted webhook %d", admin.Username, webhookID)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Webhook deleted successfully",
	})
}

// GetWebhookDeliveries handles GET /admin/webhooks/:id/deliveries - returns delivery history
func (wh *WebhooksHandler) GetWebhookDeliveries(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := wh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract webhook ID from path
	webhookID, err := wh.extractWebhookID(r)
	if err != nil {
		respondError(w, "Invalid webhook ID", http.StatusBadRequest)
		return
	}

	// Verify webhook exists
	_, err = wh.webhookSvc.GetWebhook(webhookID)
	if err != nil {
		respondError(w, "Webhook not found", http.StatusNotFound)
		return
	}

	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	deliveries := wh.webhookSvc.GetDeliveries(webhookID, limit)

	log.Printf("[WebhooksHandler] Admin %s retrieved %d deliveries for webhook %d",
		admin.Username, len(deliveries), webhookID)

	respondJSON(w, map[string]interface{}{
		"success":    true,
		"webhookId":  webhookID,
		"deliveries": deliveries,
		"count":      len(deliveries),
	})
}

// TestWebhook handles POST /admin/webhooks/:id/test - sends a test webhook
func (wh *WebhooksHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := wh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract webhook ID from path
	webhookID, err := wh.extractWebhookID(r)
	if err != nil {
		respondError(w, "Invalid webhook ID", http.StatusBadRequest)
		return
	}

	// Verify webhook exists
	webhook, err := wh.webhookSvc.GetWebhook(webhookID)
	if err != nil {
		respondError(w, "Webhook not found", http.StatusNotFound)
		return
	}

	// Send test webhook
	testData := map[string]interface{}{
		"test":    true,
		"message": "This is a test webhook",
		"sentBy":  admin.Username,
	}

	// Manually trigger the first event from the webhook's events list
	if len(webhook.Events) > 0 {
		wh.webhookSvc.TriggerEvent(webhook.Events[0], testData)
	}

	log.Printf("[WebhooksHandler] Admin %s sent test webhook %d", admin.Username, webhookID)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Test webhook sent",
	})
}

// Helper methods

func (wh *WebhooksHandler) authenticate(r *http.Request) (*Admin, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, http.ErrNoCookie
	}

	// Extract Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, http.ErrNoCookie
	}

	sessionID := parts[1]
	ipAddress := getIPAddress(r)

	admin, err := wh.authSvc.ValidateSession(sessionID, ipAddress)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

func (wh *WebhooksHandler) extractWebhookID(r *http.Request) (int64, error) {
	// Extract webhook ID from URL path
	// Expected path: /admin/webhooks/:id or /admin/webhooks/:id/deliveries
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Find the index of "webhooks" and get the next part
	for i, part := range parts {
		if part == "webhooks" && i+1 < len(parts) {
			idStr := parts[i+1]
			// Skip if it's "deliveries" or "test" (for subresources)
			if idStr == "deliveries" || idStr == "test" {
				continue
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return 0, err
			}
			return id, nil
		}
	}

	return 0, http.ErrNoCookie
}
