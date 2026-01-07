package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/epic1st/rtx/backend/lpmanager"
)

// LPHandler handles LP management API requests
type LPHandler struct {
	manager *lpmanager.Manager
}

// NewLPHandler creates a new LP handler
func NewLPHandler(manager *lpmanager.Manager) *LPHandler {
	return &LPHandler{manager: manager}
}

// HandleListLPs returns all LP configurations
func (h *LPHandler) HandleListLPs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	config := h.manager.GetConfig()
	if config == nil {
		json.NewEncoder(w).Encode([]lpmanager.LPConfig{})
		return
	}

	json.NewEncoder(w).Encode(config.LPs)
}

// HandleAddLP adds a new LP
func (h *LPHandler) HandleAddLP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	var config lpmanager.LPConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if config.ID == "" || config.Name == "" || config.Type == "" {
		http.Error(w, `{"error":"id, name, and type are required"}`, http.StatusBadRequest)
		return
	}

	if config.Settings == nil {
		config.Settings = make(map[string]string)
	}

	if err := h.manager.AddLP(config); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	log.Printf("[Admin] Added LP: %s (%s)", config.ID, config.Type)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "LP added successfully",
		"lp":      config,
	})
}

// HandleUpdateLP updates an existing LP
func (h *LPHandler) HandleUpdateLP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	// Extract ID from path: /admin/lps/{id}
	path := strings.TrimPrefix(r.URL.Path, "/admin/lps/")
	id := strings.Split(path, "/")[0]

	if id == "" {
		http.Error(w, `{"error":"LP ID required"}`, http.StatusBadRequest)
		return
	}

	var config lpmanager.LPConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	config.ID = id // Ensure ID matches path

	if err := h.manager.UpdateLP(config); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	log.Printf("[Admin] Updated LP: %s", id)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "LP updated successfully",
	})
}

// HandleDeleteLP removes an LP
func (h *LPHandler) HandleDeleteLP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/admin/lps/")
	id := strings.Split(path, "/")[0]

	if id == "" {
		http.Error(w, `{"error":"LP ID required"}`, http.StatusBadRequest)
		return
	}

	if err := h.manager.RemoveLP(id); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	log.Printf("[Admin] Removed LP: %s", id)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "LP removed successfully",
	})
}

// HandleToggleLP enables/disables an LP
func (h *LPHandler) HandleToggleLP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	// Extract ID from path: /admin/lps/{id}/toggle
	path := strings.TrimPrefix(r.URL.Path, "/admin/lps/")
	parts := strings.Split(path, "/")
	id := parts[0]

	if id == "" {
		http.Error(w, `{"error":"LP ID required"}`, http.StatusBadRequest)
		return
	}

	if err := h.manager.ToggleLP(id); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	// Get updated state
	lpConfig := h.manager.GetLPConfig(id)
	enabled := false
	if lpConfig != nil {
		enabled = lpConfig.Enabled
	}

	log.Printf("[Admin] Toggled LP %s: enabled=%v", id, enabled)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      id,
		"enabled": enabled,
		"message": "LP toggled successfully",
	})
}

// HandleLPStatus returns status of all LPs
func (h *LPHandler) HandleLPStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	status := h.manager.GetStatus()
	json.NewEncoder(w).Encode(status)
}

// HandleLPSymbols returns available symbols for an LP
func (h *LPHandler) HandleLPSymbols(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Extract ID from path: /admin/lps/{id}/symbols
	path := strings.TrimPrefix(r.URL.Path, "/admin/lps/")
	parts := strings.Split(path, "/")
	id := parts[0]

	if id == "" {
		http.Error(w, `{"error":"LP ID required"}`, http.StatusBadRequest)
		return
	}

	adapter, exists := h.manager.GetAdapter(id)
	if !exists {
		http.Error(w, `{"error":"LP adapter not found"}`, http.StatusNotFound)
		return
	}

	symbols, err := adapter.GetSymbols()
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"count":   len(symbols),
		"symbols": symbols,
	})
}
