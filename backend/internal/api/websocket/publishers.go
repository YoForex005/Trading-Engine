package websocket

import (
	"time"
)

// RoutingMetrics represents real-time routing decision data
type RoutingMetrics struct {
	Timestamp       string  `json:"timestamp"`
	Symbol          string  `json:"symbol"`
	Side            string  `json:"side"`
	Volume          float64 `json:"volume"`
	RoutingDecision string  `json:"routingDecision"` // "BBOOK" or "ABOOK"
	LPSelected      string  `json:"lpSelected,omitempty"`
	ExecutionTime   int64   `json:"executionTimeMs"`
	Spread          float64 `json:"spread"`
	Slippage        float64 `json:"slippage,omitempty"`
}

// LPPerformanceMetrics represents LP performance data
type LPPerformanceMetrics struct {
	Timestamp        string  `json:"timestamp"`
	LPName           string  `json:"lpName"`
	Status           string  `json:"status"` // "connected", "disconnected", "degraded"
	AvgSpread        float64 `json:"avgSpread"`
	ExecutionQuality float64 `json:"executionQuality"` // 0-1 score
	Latency          int64   `json:"latencyMs"`
	QuotesPerSecond  int     `json:"quotesPerSecond"`
	RejectRate       float64 `json:"rejectRate"`
	Uptime           float64 `json:"uptime"` // percentage
}

// ExposureUpdate represents real-time exposure changes
type ExposureUpdate struct {
	Timestamp      string             `json:"timestamp"`
	TotalExposure  float64            `json:"totalExposure"`
	BySymbol       map[string]float64 `json:"bySymbol"`
	ByLP           map[string]float64 `json:"byLP"`
	NetExposure    float64            `json:"netExposure"`
	ExposureLimit  float64            `json:"exposureLimit"`
	UtilizationPct float64            `json:"utilizationPct"`
	RiskLevel      string             `json:"riskLevel"` // "low", "medium", "high", "critical"
}

// Alert represents a real-time alert
type Alert struct {
	ID          string `json:"id"`
	Timestamp   string `json:"timestamp"`
	Severity    string `json:"severity"` // "info", "warning", "error", "critical"
	Category    string `json:"category"` // "exposure", "lp", "routing", "system"
	Title       string `json:"title"`
	Message     string `json:"message"`
	Source      string `json:"source,omitempty"`
	ActionItems []string `json:"actionItems,omitempty"`
}

// PublishRoutingMetrics broadcasts routing decision metrics
func (h *AnalyticsHub) PublishRoutingMetrics(metrics *RoutingMetrics) {
	if metrics.Timestamp == "" {
		metrics.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	h.Broadcast(ChannelRoutingMetrics, "routing-decision", metrics)
}

// PublishLPPerformance broadcasts LP performance metrics
func (h *AnalyticsHub) PublishLPPerformance(metrics *LPPerformanceMetrics) {
	if metrics.Timestamp == "" {
		metrics.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	h.Broadcast(ChannelLPPerformance, "lp-metrics", metrics)
}

// PublishExposureUpdate broadcasts exposure changes
func (h *AnalyticsHub) PublishExposureUpdate(update *ExposureUpdate) {
	if update.Timestamp == "" {
		update.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	h.Broadcast(ChannelExposureUpdates, "exposure-change", update)
}

// PublishAlert broadcasts a real-time alert
func (h *AnalyticsHub) PublishAlert(alert *Alert) {
	if alert.Timestamp == "" {
		alert.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	h.Broadcast(ChannelAlerts, "alert", alert)
}

// Example usage functions (for integration reference)

// OnOrderRouted is called when an order routing decision is made
func (h *AnalyticsHub) OnOrderRouted(symbol, side string, volume float64, decision, lp string, execTime int64, spread, slippage float64) {
	h.PublishRoutingMetrics(&RoutingMetrics{
		Symbol:          symbol,
		Side:            side,
		Volume:          volume,
		RoutingDecision: decision,
		LPSelected:      lp,
		ExecutionTime:   execTime,
		Spread:          spread,
		Slippage:        slippage,
	})
}

// OnLPStatusChange is called when LP connection status or performance changes
func (h *AnalyticsHub) OnLPStatusChange(lpName, status string, avgSpread, execQuality float64, latency int64, qps int, rejectRate, uptime float64) {
	h.PublishLPPerformance(&LPPerformanceMetrics{
		LPName:           lpName,
		Status:           status,
		AvgSpread:        avgSpread,
		ExecutionQuality: execQuality,
		Latency:          latency,
		QuotesPerSecond:  qps,
		RejectRate:       rejectRate,
		Uptime:           uptime,
	})
}

// OnExposureChange is called when net exposure changes significantly
func (h *AnalyticsHub) OnExposureChange(totalExposure, netExposure, exposureLimit float64, bySymbol, byLP map[string]float64) {
	utilizationPct := (totalExposure / exposureLimit) * 100

	var riskLevel string
	switch {
	case utilizationPct >= 90:
		riskLevel = "critical"
	case utilizationPct >= 75:
		riskLevel = "high"
	case utilizationPct >= 50:
		riskLevel = "medium"
	default:
		riskLevel = "low"
	}

	h.PublishExposureUpdate(&ExposureUpdate{
		TotalExposure:  totalExposure,
		BySymbol:       bySymbol,
		ByLP:           byLP,
		NetExposure:    netExposure,
		ExposureLimit:  exposureLimit,
		UtilizationPct: utilizationPct,
		RiskLevel:      riskLevel,
	})
}

// EmitAlert creates and broadcasts an alert
func (h *AnalyticsHub) EmitAlert(severity, category, title, message, source string, actionItems []string) {
	h.PublishAlert(&Alert{
		ID:          generateAlertID(),
		Severity:    severity,
		Category:    category,
		Title:       title,
		Message:     message,
		Source:      source,
		ActionItems: actionItems,
	})
}

// generateAlertID creates a unique alert ID
func generateAlertID() string {
	return time.Now().Format("20060102-150405") + "-" + randomString(6)
}

// randomString generates a random string (simple implementation)
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, n)
	for i := range result {
		result[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(result)
}
