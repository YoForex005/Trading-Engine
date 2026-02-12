//go:build rtx_legacy_admin
// +build rtx_legacy_admin

package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// AuditHandler handles audit log API endpoints
type AuditHandler struct {
	logger      *AuditLogger
	authService AuthTokenValidator
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(logger *AuditLogger, authService AuthTokenValidator) *AuditHandler {
	return &AuditHandler{
		logger:      logger,
		authService: authService,
	}
}

// HandleListLogs - GET /admin/audit/logs
// Query params: ?admin=X&action=Y&resource=Z&from=TIMESTAMP&to=TIMESTAMP&success=true/false&limit=50&offset=0
func (h *AuditHandler) HandleListLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	filters := AuditFilters{
		AdminID:  r.URL.Query().Get("admin"),
		Action:   AuditAction(r.URL.Query().Get("action")),
		Resource: r.URL.Query().Get("resource"),
		FromTime: parseTimestamp(r.URL.Query().Get("from")),
		ToTime:   parseTimestamp(r.URL.Query().Get("to")),
		Limit:    parseIntParam(r.URL.Query().Get("limit"), 50),
		Offset:   parseIntParam(r.URL.Query().Get("offset"), 0),
	}

	// Parse success filter (optional)
	if successStr := r.URL.Query().Get("success"); successStr != "" {
		success := successStr == "true"
		filters.Success = &success
	}

	// Query logs
	logs := h.logger.Query(filters)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
		"total": h.logger.Count(),
	})
}

// HandleGetLog - GET /admin/audit/logs/{id}
func (h *AuditHandler) HandleGetLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid log ID", http.StatusBadRequest)
		return
	}
	logID := parts[4]

	// Get log entry
	entry := h.logger.GetByID(logID)
	if entry == nil {
		http.Error(w, "Log entry not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// HandleGetStats - GET /admin/audit/stats
func (h *AuditHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := h.logger.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// parseTimestamp parses a timestamp string to Unix timestamp
func parseTimestamp(ts string) int64 {
	if ts == "" {
		return 0
	}
	timestamp, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return 0
	}
	return timestamp
}

// parseIntParam parses an integer parameter with default value
func parseIntParam(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}
