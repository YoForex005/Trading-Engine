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

// HandleLPSymbols returns available symbols for an LP or updates subscriptions
func (h *LPHandler) HandleLPSymbols(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	// Extract ID from path: /admin/lps/{id}/symbols
	path := strings.TrimPrefix(r.URL.Path, "/admin/lps/")
	parts := strings.Split(path, "/")
	id := parts[0]

	if id == "" {
		http.Error(w, `{"error":"LP ID required"}`, http.StatusBadRequest)
		return
	}

	// GET - Retrieve available symbols for an LP
	if r.Method == "GET" {
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

		lpConfig := h.manager.GetLPConfig(id)
		currentSubs := []string{}
		if lpConfig != nil && lpConfig.Symbols != nil {
			currentSubs = lpConfig.Symbols
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":                   id,
			"availableCount":       len(symbols),
			"availableSymbols":     symbols,
			"currentSubscriptions": currentSubs,
		})
		return
	}

	// PUT - Update symbol subscriptions for an LP
	if r.Method == "PUT" {
		var req struct {
			Symbols []string `json:"symbols"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}

		lpConfig := h.manager.GetLPConfig(id)
		if lpConfig == nil {
			http.Error(w, `{"error":"LP not found"}`, http.StatusNotFound)
			return
		}

		// Update subscriptions
		lpConfig.Symbols = req.Symbols

		// Update the LP configuration
		if err := h.manager.UpdateLP(*lpConfig); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}

		// Attempt to subscribe in the adapter if it supports Subscribe
		adapter, exists := h.manager.GetAdapter(id)
		if exists {
			if wsAdapter, ok := adapter.(interface{ Subscribe([]string) error }); ok {
				if err := wsAdapter.Subscribe(req.Symbols); err != nil {
					log.Printf("[LP] Warning: Failed to subscribe %s to symbols: %v", id, err)
					// Continue anyway - config was saved
				}
			}
		}

		log.Printf("[Admin] Updated LP %s subscriptions: %d symbols", id, len(req.Symbols))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      id,
			"message": "Subscriptions updated successfully",
			"symbols": req.Symbols,
			"count":   len(req.Symbols),
		})
		return
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

// HandleAdminLiquidityProviders returns all LPs with detailed status
func (h *LPHandler) HandleAdminLiquidityProviders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	if r.Method != "GET" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	config := h.manager.GetConfig()
	status := h.manager.GetStatus()

	// Combine configuration and status for comprehensive view
	type LPDetail struct {
		ID                string   `json:"id"`
		Name              string   `json:"name"`
		Type              string   `json:"type"`
		Enabled           bool     `json:"enabled"`
		Connected         bool     `json:"connected"`
		Priority          int      `json:"priority"`
		SymbolCount       int      `json:"symbolCount"`
		SubscribedSymbols int      `json:"subscribedSymbols"`
		SubscribedTo      []string `json:"subscribedTo"`
		LastTick          string   `json:"lastTick"`
		ErrorMessage      string   `json:"errorMessage,omitempty"`
	}

	lps := make([]LPDetail, 0)
	if config != nil {
		for _, lpCfg := range config.LPs {
			lpStatus := status[lpCfg.ID]

			subscribedCount := 0
			subscribedTo := []string{}
			if len(lpCfg.Symbols) > 0 {
				subscribedCount = len(lpCfg.Symbols)
				subscribedTo = lpCfg.Symbols
			}

			detail := LPDetail{
				ID:                lpCfg.ID,
				Name:              lpCfg.Name,
				Type:              lpCfg.Type,
				Enabled:           lpCfg.Enabled,
				Connected:         lpStatus.Connected,
				Priority:          lpCfg.Priority,
				SymbolCount:       lpStatus.SymbolCount,
				SubscribedSymbols: subscribedCount,
				SubscribedTo:      subscribedTo,
				LastTick:          lpStatus.LastTick.Format("2006-01-02T15:04:05Z07:00"),
				ErrorMessage:      lpStatus.ErrorMessage,
			}

			lps = append(lps, detail)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"count": len(lps),
		"lps":   lps,
	})
}

// HandleToggleLPByName enables/disables an LP by name
func (h *LPHandler) HandleToggleLPByName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract name from path: /api/admin/lp/{name}/toggle
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/lp/")
	parts := strings.Split(path, "/")
	name := parts[0]

	if name == "" {
		http.Error(w, `{"error":"LP name required"}`, http.StatusBadRequest)
		return
	}

	// Find LP by name (case-insensitive)
	config := h.manager.GetConfig()
	var lpID string
	if config != nil {
		for _, lp := range config.LPs {
			if strings.EqualFold(lp.Name, name) || strings.EqualFold(lp.ID, name) {
				lpID = lp.ID
				break
			}
		}
	}

	if lpID == "" {
		http.Error(w, `{"error":"LP not found"}`, http.StatusNotFound)
		return
	}

	if err := h.manager.ToggleLP(lpID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	lpConfig := h.manager.GetLPConfig(lpID)
	enabled := false
	if lpConfig != nil {
		enabled = lpConfig.Enabled
	}

	log.Printf("[Admin] Toggled LP by name: %s (ID: %s), enabled=%v", name, lpID, enabled)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"name":    name,
		"id":      lpID,
		"enabled": enabled,
		"message": "LP toggled successfully",
	})
}

// HandleGetLPSubscriptions returns current symbol subscriptions for an LP
func (h *LPHandler) HandleGetLPSubscriptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	if r.Method != "GET" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract name from path: /api/admin/lp/{name}/subscriptions
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/lp/")
	parts := strings.Split(path, "/")
	name := parts[0]

	if name == "" {
		http.Error(w, `{"error":"LP name required"}`, http.StatusBadRequest)
		return
	}

	// Find LP by name (case-insensitive)
	config := h.manager.GetConfig()
	var lpID string
	var lpConfig *lpmanager.LPConfig
	if config != nil {
		for i := range config.LPs {
			if strings.EqualFold(config.LPs[i].Name, name) || strings.EqualFold(config.LPs[i].ID, name) {
				lpID = config.LPs[i].ID
				lpConfig = &config.LPs[i]
				break
			}
		}
	}

	if lpID == "" {
		http.Error(w, `{"error":"LP not found"}`, http.StatusNotFound)
		return
	}

	subscriptions := []string{}
	if lpConfig != nil && lpConfig.Symbols != nil {
		subscriptions = lpConfig.Symbols
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":           name,
		"id":             lpID,
		"count":          len(subscriptions),
		"subscriptions":  subscriptions,
		"subscribeToAll": len(subscriptions) == 0,
	})
}

// HandleUpdateLPSubscriptions updates symbol subscriptions for an LP by name
func (h *LPHandler) HandleUpdateLPSubscriptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	if r.Method != "PUT" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract name from path: /api/admin/lp/{name}/subscriptions
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/lp/")
	parts := strings.Split(path, "/")
	name := parts[0]

	if name == "" {
		http.Error(w, `{"error":"LP name required"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Symbols []string `json:"symbols"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Find LP by name (case-insensitive)
	config := h.manager.GetConfig()
	var lpID string
	var lpConfig *lpmanager.LPConfig
	if config != nil {
		for i := range config.LPs {
			if strings.EqualFold(config.LPs[i].Name, name) || strings.EqualFold(config.LPs[i].ID, name) {
				lpID = config.LPs[i].ID
				lpConfig = &config.LPs[i]
				break
			}
		}
	}

	if lpID == "" {
		http.Error(w, `{"error":"LP not found"}`, http.StatusNotFound)
		return
	}

	// Update subscriptions
	lpConfig.Symbols = req.Symbols

	// Persist the change
	if err := h.manager.UpdateLP(*lpConfig); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// Attempt to subscribe in the adapter if it supports Subscribe
	adapter, exists := h.manager.GetAdapter(lpID)
	if exists {
		if wsAdapter, ok := adapter.(interface{ Subscribe([]string) error }); ok {
			if err := wsAdapter.Subscribe(req.Symbols); err != nil {
				log.Printf("[LP] Warning: Failed to subscribe %s to symbols: %v", lpID, err)
				// Continue anyway - config was saved
			}
		}
	}

	log.Printf("[Admin] Updated LP subscriptions by name: %s (ID: %s), %d symbols", name, lpID, len(req.Symbols))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"name":         name,
		"id":           lpID,
		"message":      "Subscriptions updated successfully",
		"symbols":      req.Symbols,
		"count":        len(req.Symbols),
		"subscribeAll": len(req.Symbols) == 0,
	})
}
