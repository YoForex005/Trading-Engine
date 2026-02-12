package notifications

import (
	"encoding/json"
	"log"
)

// WSHub interface represents the WebSocket hub from backend/ws package
type WSHub interface {
	BroadcastMessage(message []byte)
	BroadcastToUser(userID string, message []byte) // User-specific messages
}

// WSPublisher implements WebSocketPublisher using the existing WS hub
type WSPublisher struct {
	hub WSHub
}

// NewWSPublisher creates a new WebSocket publisher
func NewWSPublisher(hub WSHub) *WSPublisher {
	return &WSPublisher{
		hub: hub,
	}
}

// PublishNotification sends a notification to a specific user via WebSocket
func (p *WSPublisher) PublishNotification(userID string, notification *Notification) error {
	// Create WebSocket message with notification
	message := map[string]interface{}{
		"type":         "notification",
		"notification": notification,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("[WSPublisher] Failed to marshal notification: %v", err)
		return err
	}

	// Send to specific user's authenticated connections only
	p.hub.BroadcastToUser(userID, data)

	return nil
}
