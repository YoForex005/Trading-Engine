package datapipeline

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// APIHandler provides HTTP endpoints for the data pipeline
type APIHandler struct {
	pipeline *MarketDataPipeline
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(pipeline *MarketDataPipeline) *APIHandler {
	return &APIHandler{
		pipeline: pipeline,
	}
}

// RegisterRoutes registers all API routes
func (h *APIHandler) RegisterRoutes(mux *http.ServeMux) {
	// Quote API
	mux.HandleFunc("/api/quotes/latest/", h.HandleGetLatestQuote)
	mux.HandleFunc("/api/quotes/stream", h.HandleQuoteStream)
	mux.HandleFunc("/api/quotes/history/", h.HandleGetQuoteHistory)

	// OHLC API
	mux.HandleFunc("/api/ohlc/", h.HandleGetOHLC)
	mux.HandleFunc("/api/ohlc/latest/", h.HandleGetLatestOHLC)

	// Ticks API (backward compatibility)
	mux.HandleFunc("/api/ticks/", h.HandleGetTicks)

	// Pipeline management
	mux.HandleFunc("/api/pipeline/stats", h.HandleGetStats)
	mux.HandleFunc("/api/pipeline/health", h.HandleGetHealth)
	mux.HandleFunc("/api/pipeline/feed-health", h.HandleGetFeedHealth)
	mux.HandleFunc("/api/pipeline/alerts", h.HandleGetAlerts)

	// Admin API
	mux.HandleFunc("/admin/pipeline/cleanup", h.HandleCleanupStorage)
}

// HandleGetLatestQuote returns the latest quote for a symbol
func (h *APIHandler) HandleGetLatestQuote(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract symbol from path
	symbol := r.URL.Path[len("/api/quotes/latest/"):]
	if symbol == "" {
		http.Error(w, "Missing symbol", http.StatusBadRequest)
		return
	}

	// Get from storage
	tick, err := h.pipeline.storage.GetLatestTick(symbol)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get quote: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tick)
}

// HandleGetQuoteHistory returns recent quotes for a symbol
func (h *APIHandler) HandleGetQuoteHistory(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	symbol := r.URL.Path[len("/api/quotes/history/"):]
	if symbol == "" {
		http.Error(w, "Missing symbol", http.StatusBadRequest)
		return
	}

	// Get limit from query
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Get from storage
	ticks, err := h.pipeline.storage.GetRecentTicks(symbol, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get history: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticks)
}

// HandleQuoteStream provides Server-Sent Events stream
func (h *APIHandler) HandleQuoteStream(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w)

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Get symbols from query
	symbols := r.URL.Query()["symbol"]
	if len(symbols) == 0 {
		http.Error(w, "No symbols specified", http.StatusBadRequest)
		return
	}

	// Create flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	log.Printf("[API] SSE stream started for symbols: %v", symbols)

	// Subscribe to Redis pub/sub for these symbols
	// This is a simplified version - in production, use proper Redis pub/sub
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			log.Println("[API] SSE client disconnected")
			return
		case <-ticker.C:
			// Get latest quotes for subscribed symbols
			for _, symbol := range symbols {
				tick, err := h.pipeline.storage.GetLatestTick(symbol)
				if err != nil {
					continue
				}

				data, _ := json.Marshal(tick)
				fmt.Fprintf(w, "data: %s\n\n", data)
			}
			flusher.Flush()
		}
	}
}

// HandleGetOHLC returns OHLC bars for a symbol
func (h *APIHandler) HandleGetOHLC(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	symbol := r.URL.Path[len("/api/ohlc/"):]
	if symbol == "" {
		http.Error(w, "Missing symbol", http.StatusBadRequest)
		return
	}

	// Parse timeframe
	timeframeStr := r.URL.Query().Get("timeframe")
	if timeframeStr == "" {
		timeframeStr = "1m"
	}

	timeframe := h.parseTimeframe(timeframeStr)

	// Get limit
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Get OHLC bars
	bars, err := h.pipeline.storage.GetOHLCBars(symbol, timeframe, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get OHLC: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bars)
}

// HandleGetLatestOHLC returns the current active OHLC bar
func (h *APIHandler) HandleGetLatestOHLC(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	symbol := r.URL.Path[len("/api/ohlc/latest/"):]
	if symbol == "" {
		http.Error(w, "Missing symbol", http.StatusBadRequest)
		return
	}

	timeframeStr := r.URL.Query().Get("timeframe")
	if timeframeStr == "" {
		timeframeStr = "1m"
	}

	timeframe := h.parseTimeframe(timeframeStr)

	// Get active bar from OHLC engine
	bar := h.pipeline.ohlcEngine.GetActiveBar(symbol, timeframe)
	if bar == nil {
		http.Error(w, "No active bar found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bar)
}

// HandleGetTicks provides backward compatibility
func (h *APIHandler) HandleGetTicks(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "Missing symbol parameter", http.StatusBadRequest)
		return
	}

	limit := 500
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ticks, err := h.pipeline.storage.GetRecentTicks(symbol, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get ticks: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticks)
}

// HandleGetStats returns pipeline statistics
func (h *APIHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w)

	stats := h.pipeline.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleGetHealth returns pipeline health status
func (h *APIHandler) HandleGetHealth(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w)

	health, err := h.pipeline.HealthCheck()
	if err != nil {
		http.Error(w, fmt.Sprintf("Health check failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// HandleGetFeedHealth returns feed health status
func (h *APIHandler) HandleGetFeedHealth(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w)

	feedHealth := h.pipeline.monitor.GetFeedHealth()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feedHealth)
}

// HandleGetAlerts returns recent alerts
func (h *APIHandler) HandleGetAlerts(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w)

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	alerts := h.pipeline.monitor.GetAlerts(limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// HandleCleanupStorage triggers storage cleanup
func (h *APIHandler) HandleCleanupStorage(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w)

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.pipeline.storage.CleanupOldData(); err != nil {
		http.Error(w, fmt.Sprintf("Cleanup failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Storage cleanup completed",
	})
}

// parseTimeframe converts string to Timeframe
func (h *APIHandler) parseTimeframe(s string) Timeframe {
	switch s {
	case "1m", "M1":
		return TF_M1
	case "5m", "M5":
		return TF_M5
	case "15m", "M15":
		return TF_M15
	case "1h", "H1":
		return TF_H1
	case "4h", "H4":
		return TF_H4
	case "1d", "D1":
		return TF_D1
	default:
		return TF_M1
	}
}

// setCORSHeaders sets CORS headers
func (h *APIHandler) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}
