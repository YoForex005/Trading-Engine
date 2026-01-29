package websocket

import (
	"log"
	"net/http"

	"github.com/epic1st/rtx/backend/auth"
)

// Integration holds the analytics hub instance for global access
var globalAnalyticsHub *AnalyticsHub

// InitializeAnalyticsHub creates and starts the analytics WebSocket hub
func InitializeAnalyticsHub(authService *auth.Service) *AnalyticsHub {
	hub := NewAnalyticsHub(authService)
	globalAnalyticsHub = hub

	// Start the hub in a goroutine
	go hub.Run()

	log.Println("[AnalyticsWS] Analytics WebSocket hub initialized and running")
	return hub
}

// GetAnalyticsHub returns the global analytics hub instance
func GetAnalyticsHub() *AnalyticsHub {
	return globalAnalyticsHub
}

// RegisterAnalyticsRoutes registers the WebSocket endpoint
func RegisterAnalyticsRoutes(hub *AnalyticsHub, mux *http.ServeMux) {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	// WebSocket endpoint for analytics
	mux.HandleFunc("/ws/analytics", func(w http.ResponseWriter, r *http.Request) {
		// Enable CORS for WebSocket upgrade
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		hub.ServeAnalyticsWs(w, r)
	})

	log.Println("[AnalyticsWS] Analytics WebSocket route registered at /ws/analytics")
}

// Example integration functions for other modules

// NotifyRoutingDecision should be called from the routing engine after each decision
func NotifyRoutingDecision(symbol, side string, volume float64, decision, lp string, execTime int64, spread, slippage float64) {
	hub := GetAnalyticsHub()
	if hub != nil {
		hub.OnOrderRouted(symbol, side, volume, decision, lp, execTime, spread, slippage)
	}
}

// NotifyLPStatusChange should be called when LP status changes
func NotifyLPStatusChange(lpName, status string, avgSpread, execQuality float64, latency int64, qps int, rejectRate, uptime float64) {
	hub := GetAnalyticsHub()
	if hub != nil {
		hub.OnLPStatusChange(lpName, status, avgSpread, execQuality, latency, qps, rejectRate, uptime)
	}
}

// NotifyExposureChange should be called when exposure changes significantly
func NotifyExposureChange(totalExposure, netExposure, exposureLimit float64, bySymbol, byLP map[string]float64) {
	hub := GetAnalyticsHub()
	if hub != nil {
		hub.OnExposureChange(totalExposure, netExposure, exposureLimit, bySymbol, byLP)
	}
}

// NotifyAlert should be called to broadcast an alert
func NotifyAlert(severity, category, title, message, source string, actionItems []string) {
	hub := GetAnalyticsHub()
	if hub != nil {
		hub.EmitAlert(severity, category, title, message, source, actionItems)
	}
}

// ShutdownAnalyticsHub gracefully shuts down the analytics hub
func ShutdownAnalyticsHub() {
	hub := GetAnalyticsHub()
	if hub != nil {
		hub.Stop()
		log.Println("[AnalyticsWS] Analytics WebSocket hub shut down")
	}
}
