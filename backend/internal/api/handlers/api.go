package handlers

import (
	"net/http"

	"github.com/epic1st/rtx/backend/internal/core"
	"github.com/epic1st/rtx/backend/ws"
)

// APIHandler handles B-Book API requests
type APIHandler struct {
	engine    *core.Engine
	pnlEngine *core.PnLEngine
	hub       *ws.Hub
}

// NewAPIHandler creates API handlers for B-Book
func NewAPIHandler(engine *core.Engine, pnlEngine *core.PnLEngine) *APIHandler {
	return &APIHandler{
		engine:    engine,
		pnlEngine: pnlEngine,
	}
}

// SetHub sets the WebSocket hub reference
func (h *APIHandler) SetHub(hub *ws.Hub) {
	h.hub = hub
}

// cors adds CORS headers
func cors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
}
