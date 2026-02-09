package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/epic1st/rtx/backend/fix"
)

// FIXConnectionHandler provides admin endpoints for FIX connection management
type FIXConnectionHandler struct {
	gateway   *fix.FIXGateway
	connMgr   *fix.ConnectionManager
}

// NewFIXConnectionHandler creates a new FIX connection handler
func NewFIXConnectionHandler(gateway *fix.FIXGateway, connMgr *fix.ConnectionManager) *FIXConnectionHandler {
	return &FIXConnectionHandler{
		gateway: gateway,
		connMgr: connMgr,
	}
}

// ConnectionRequest represents a connection request
type ConnectionRequest struct {
	SessionID string `json:"session_id"`
}

// ReconnectRequest represents a reconnection request
type ReconnectRequest struct {
	SessionID      string `json:"session_id"`
	EnableAutoRetry bool  `json:"enable_auto_retry"`
}

// ConnectionStatusResponse provides detailed connection status
type ConnectionStatusResponse struct {
	SessionID           string    `json:"session_id"`
	SessionName         string    `json:"session_name"`
	Status              string    `json:"status"`
	Host                string    `json:"host"`
	Port                int       `json:"port"`
	SSL                 bool      `json:"ssl"`
	UseProxy            bool      `json:"use_proxy"`
	LastHeartbeat       time.Time `json:"last_heartbeat,omitempty"`
	LastHeartbeatAgo    string    `json:"last_heartbeat_ago,omitempty"`
	ConnectionStats     *fix.ConnectionStats `json:"connection_stats,omitempty"`
	OutSeqNum           int       `json:"out_seq_num"`
	InSeqNum            int       `json:"in_seq_num"`
	Diagnostics         string    `json:"diagnostics"`
}

// RegisterRoutes registers HTTP routes for FIX connection management
func (h *FIXConnectionHandler) RegisterRoutes(mux *http.ServeMux) {
	// Connection management
	mux.HandleFunc("/admin/fix/connect", h.handleConnect)
	mux.HandleFunc("/admin/fix/disconnect", h.handleDisconnect)
	mux.HandleFunc("/admin/fix/reconnect", h.handleReconnect)
	mux.HandleFunc("/admin/fix/status", h.handleStatus)
	mux.HandleFunc("/admin/fix/detailed-status", h.handleDetailedStatus)

	// Connection manager controls
	mux.HandleFunc("/admin/fix/enable-auto-reconnect", h.handleEnableAutoReconnect)
	mux.HandleFunc("/admin/fix/disable-auto-reconnect", h.handleDisableAutoReconnect)
	mux.HandleFunc("/admin/fix/connection-stats", h.handleConnectionStats)
	mux.HandleFunc("/admin/fix/diagnostics", h.handleDiagnostics)
}

// handleConnect initiates a FIX connection
func (h *FIXConnectionHandler) handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	// Attempt connection
	err := h.gateway.Connect(req.SessionID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"session_id": req.SessionID,
		})
		return
	}

	// Wait a moment to check if login succeeded
	time.Sleep(2 * time.Second)

	status := h.gateway.GetStatus()
	sessionStatus := status[req.SessionID]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    sessionStatus == "LOGGED_IN" || sessionStatus == "CONNECTING",
		"session_id": req.SessionID,
		"status":     sessionStatus,
		"message":    fmt.Sprintf("Connection initiated for %s", req.SessionID),
	})
}

// handleDisconnect disconnects a FIX session
func (h *FIXConnectionHandler) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	err := h.gateway.Disconnect(req.SessionID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Disconnected from %s", req.SessionID),
	})
}

// handleReconnect forces a reconnection
func (h *FIXConnectionHandler) handleReconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ReconnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	// Enable auto-retry if requested
	if req.EnableAutoRetry {
		h.connMgr.EnableAutoReconnect(req.SessionID)
	}

	// Force reconnect
	err := h.connMgr.ForceReconnect(req.SessionID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"message":       fmt.Sprintf("Reconnection initiated for %s", req.SessionID),
		"auto_retry_enabled": req.EnableAutoRetry,
	})
}

// handleStatus returns basic status for all sessions
func (h *FIXConnectionHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := h.gateway.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": status,
		"timestamp": time.Now(),
	})
}

// handleDetailedStatus returns detailed status including connection stats
func (h *FIXConnectionHandler) handleDetailedStatus(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")

	if sessionID == "" {
		// Return all sessions
		allStats := h.connMgr.GetAllConnectionStats()

		detailedStatus := make(map[string]*ConnectionStatusResponse)
		for id, stats := range allStats {
			detailedStatus[id] = h.buildDetailedStatus(id, stats)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sessions":  detailedStatus,
			"timestamp": time.Now(),
		})
		return
	}

	// Return single session
	stats := h.connMgr.GetConnectionStats(sessionID)
	response := h.buildDetailedStatus(sessionID, stats)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// buildDetailedStatus builds a detailed status response
func (h *FIXConnectionHandler) buildDetailedStatus(sessionID string, stats *fix.ConnectionStats) *ConnectionStatusResponse {
	session := h.gateway.GetSession(sessionID)

	response := &ConnectionStatusResponse{
		SessionID:       sessionID,
		Status:          stats.Status,
		ConnectionStats: stats,
	}

	if session != nil {
		response.SessionName = session.Name
		response.Host = session.Host
		response.Port = session.Port
		response.SSL = session.SSL
		response.UseProxy = session.UseProxy
		response.LastHeartbeat = session.LastHeartbeat
		response.OutSeqNum = session.OutSeqNum
		response.InSeqNum = session.InSeqNum

		if !session.LastHeartbeat.IsZero() {
			response.LastHeartbeatAgo = time.Since(session.LastHeartbeat).String()
		}

		// Build diagnostics
		response.Diagnostics = h.buildDiagnostics(session, stats)
	}

	return response
}

// buildDiagnostics builds diagnostic information
func (h *FIXConnectionHandler) buildDiagnostics(session *fix.LPSession, stats *fix.ConnectionStats) string {
	var diag string

	diag += fmt.Sprintf("Session: %s (%s)\n", session.ID, session.Name)
	diag += fmt.Sprintf("Target: %s:%d (SSL: %v)\n", session.Host, session.Port, session.SSL)
	diag += fmt.Sprintf("Status: %s\n", stats.Status)
	diag += fmt.Sprintf("Health: %s\n", stats.HealthStatus)

	if session.UseProxy {
		diag += fmt.Sprintf("Proxy: %s:%d\n", session.ProxyHost, session.ProxyPort)
	}

	diag += fmt.Sprintf("\nFIX Protocol:\n")
	diag += fmt.Sprintf("  BeginString: %s\n", session.BeginString)
	diag += fmt.Sprintf("  SenderCompID: %s\n", session.SenderCompID)
	diag += fmt.Sprintf("  TargetCompID: %s\n", session.TargetCompID)
	diag += fmt.Sprintf("  OutSeqNum: %d\n", session.OutSeqNum)
	diag += fmt.Sprintf("  InSeqNum: %d\n", session.InSeqNum)

	if !session.LastHeartbeat.IsZero() {
		diag += fmt.Sprintf("\nLast Heartbeat: %v (%v ago)\n",
			session.LastHeartbeat, time.Since(session.LastHeartbeat))
	}

	if stats.LastError != "" {
		diag += fmt.Sprintf("\nLast Error: %s\n", stats.LastError)
	}

	diag += fmt.Sprintf("\nConnection Stats:\n")
	diag += fmt.Sprintf("  Total Reconnects: %d\n", stats.TotalReconnects)
	diag += fmt.Sprintf("  Consecutive Failures: %d\n", stats.ConsecutiveFailures)
	diag += fmt.Sprintf("  Auto-Reconnect: %v\n", stats.AutoReconnectActive)

	if !stats.NextRetry.IsZero() {
		diag += fmt.Sprintf("  Next Retry: %v (in %v)\n",
			stats.NextRetry, time.Until(stats.NextRetry))
	}

	return diag
}

// handleEnableAutoReconnect enables automatic reconnection
func (h *FIXConnectionHandler) handleEnableAutoReconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id query parameter is required", http.StatusBadRequest)
		return
	}

	h.connMgr.EnableAutoReconnect(sessionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"session_id": sessionID,
		"message":    "Auto-reconnect enabled",
	})
}

// handleDisableAutoReconnect disables automatic reconnection
func (h *FIXConnectionHandler) handleDisableAutoReconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id query parameter is required", http.StatusBadRequest)
		return
	}

	h.connMgr.DisableAutoReconnect(sessionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"session_id": sessionID,
		"message":    "Auto-reconnect disabled",
	})
}

// handleConnectionStats returns connection statistics
func (h *FIXConnectionHandler) handleConnectionStats(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")

	if sessionID == "" {
		// Return all stats
		allStats := h.connMgr.GetAllConnectionStats()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(allStats)
		return
	}

	// Return single session stats
	stats := h.connMgr.GetConnectionStats(sessionID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleDiagnostics returns detailed diagnostics
func (h *FIXConnectionHandler) handleDiagnostics(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id query parameter is required", http.StatusBadRequest)
		return
	}

	stats := h.connMgr.GetConnectionStats(sessionID)
	session := h.gateway.GetSession(sessionID)

	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	diagnostics := h.buildDiagnostics(session, stats)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(diagnostics))
}
