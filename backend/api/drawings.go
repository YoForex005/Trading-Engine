package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// Drawing represents a chart annotation
type Drawing struct {
	ID        string      `json:"id"`
	Symbol    string      `json:"symbol"`
	AccountID int         `json:"accountId"`
	Type      string      `json:"type"`
	Subtype   string      `json:"subtype,omitempty"`
	Points    []Point     `json:"points"`
	Text      string      `json:"text,omitempty"`
	Color     string      `json:"color,omitempty"`
	LineWidth int         `json:"lineWidth,omitempty"`
	LineStyle string      `json:"lineStyle,omitempty"`
	Locked    bool        `json:"locked,omitempty"`
}

type Point struct {
	Time  float64 `json:"time"`
	Price float64 `json:"price"`
}

// DrawingsHandler manages drawing persistence
type DrawingsHandler struct {
	store map[string]Drawing // Key: Drawing ID
	mutex sync.RWMutex
}

// NewDrawingsHandler creates a new handler
func NewDrawingsHandler() *DrawingsHandler {
	return &DrawingsHandler{
		store: make(map[string]Drawing),
	}
}

// RegisterRoutes registers API routes
func (h *DrawingsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/drawings", h.handleDrawings)
	mux.HandleFunc("/api/drawings/", h.handleSingleDrawing)
}

func (h *DrawingsHandler) handleDrawings(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == "GET" {
		h.handleGetDrawings(w, r)
		return
	}

	if r.Method == "POST" {
		h.handleSaveDrawing(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (h *DrawingsHandler) handleSingleDrawing(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == "DELETE" {
		h.handleDeleteDrawing(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (h *DrawingsHandler) handleGetDrawings(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var result []Drawing
	for _, d := range h.store {
		if d.Symbol == symbol {
			result = append(result, d)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *DrawingsHandler) handleSaveDrawing(w http.ResponseWriter, r *http.Request) {
	var drawing Drawing
	if err := json.NewDecoder(r.Body).Decode(&drawing); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	h.mutex.Lock()
	h.store[drawing.ID] = drawing
	h.mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(drawing)
}

func (h *DrawingsHandler) handleDeleteDrawing(w http.ResponseWriter, r *http.Request) {
	// Path: /api/drawings/{id}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	id := parts[3]

	h.mutex.Lock()
	delete(h.store, id)
	h.mutex.Unlock()

	w.WriteHeader(http.StatusOK)
}
