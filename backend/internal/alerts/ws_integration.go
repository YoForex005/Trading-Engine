package alerts

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/epic1st/rtx/backend/ws"
)

// WSAlertHub wraps the WebSocket hub to broadcast alerts
type WSAlertHub struct {
	hub *ws.Hub
	mu  sync.RWMutex
}

// NewWSAlertHub creates a WebSocket alert broadcaster
func NewWSAlertHub(hub *ws.Hub) *WSAlertHub {
	return &WSAlertHub{
		hub: hub,
	}
}

// BroadcastAlert sends alert to connected WebSocket clients
func (w *WSAlertHub) BroadcastAlert(alert *Alert) {
	if w.hub == nil {
		log.Println("[WSAlertHub] Hub not initialized, cannot broadcast alert")
		return
	}

	// Create alert message for WebSocket clients
	message := map[string]interface{}{
		"type":      "alert",
		"alert":     alert,
		"timestamp": alert.CreatedAt.Unix(),
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("[WSAlertHub] Failed to marshal alert: %v", err)
		return
	}

	// Broadcast to all connected clients
	// Note: In production, you'd filter by accountID to only send to relevant clients
	w.hub.BroadcastMessage(data)

	log.Printf("[WSAlertHub] Broadcasted alert %s (severity: %s) to WebSocket clients",
		alert.ID, alert.Severity)
}

// Note: We need to add BroadcastMessage method to ws.Hub
// This is a temporary workaround - add this method to ws/hub.go:
/*
func (h *Hub) BroadcastMessage(message []byte) {
	select {
	case h.broadcast <- message:
	default:
		log.Println("[Hub] Broadcast buffer full, message dropped")
	}
}
*/
