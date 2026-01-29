package wscluster

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// MetricsCollector collects and aggregates cluster metrics
type MetricsCollector struct {
	client *redis.Client
	nodeID string
	ctx    context.Context

	// Local metrics
	localMetrics *NodeMetrics
	metricsMu    sync.RWMutex

	// Historical data
	history      []*MetricsSnapshot
	historyMu    sync.RWMutex
	maxHistory   int

	// Real-time tracking
	requestLatencies []float64
	latencyMu        sync.Mutex
	latencyWindow    int
}

// NodeMetrics tracks detailed node-level metrics
type NodeMetrics struct {
	NodeID              string    `json:"node_id"`
	Timestamp           time.Time `json:"timestamp"`

	// Connection metrics
	ActiveConnections   int64     `json:"active_connections"`
	TotalConnectionsAll int64     `json:"total_connections_all_time"`
	ConnectionsOpened   int64     `json:"connections_opened_1m"`
	ConnectionsClosed   int64     `json:"connections_closed_1m"`
	FailedConnections   int64     `json:"failed_connections"`

	// Message metrics
	MessagesReceived    int64     `json:"messages_received"`
	MessagesSent        int64     `json:"messages_sent"`
	MessagesFailed      int64     `json:"messages_failed"`
	BytesReceived       int64     `json:"bytes_received"`
	BytesSent           int64     `json:"bytes_sent"`

	// Performance metrics
	AverageLatencyMs    float64   `json:"average_latency_ms"`
	P50LatencyMs        float64   `json:"p50_latency_ms"`
	P95LatencyMs        float64   `json:"p95_latency_ms"`
	P99LatencyMs        float64   `json:"p99_latency_ms"`

	// Resource metrics
	CPUUsage            float64   `json:"cpu_usage_percent"`
	MemoryUsage         float64   `json:"memory_usage_percent"`
	GoroutineCount      int       `json:"goroutine_count"`

	// Error metrics
	ErrorRate           float64   `json:"error_rate"`
	ReconnectionRate    float64   `json:"reconnection_rate"`
	TimeoutRate         float64   `json:"timeout_rate"`
}

// MetricsSnapshot captures metrics at a point in time
type MetricsSnapshot struct {
	Timestamp time.Time     `json:"timestamp"`
	Cluster   *ClusterMetrics `json:"cluster"`
	Nodes     []*NodeMetrics  `json:"nodes"`
}

// AlertRule defines conditions for triggering alerts
type AlertRule struct {
	Name        string  `json:"name"`
	Metric      string  `json:"metric"`
	Operator    string  `json:"operator"` // >, <, >=, <=, ==
	Threshold   float64 `json:"threshold"`
	Duration    time.Duration `json:"duration"`
	Severity    string  `json:"severity"` // critical, warning, info
	Enabled     bool    `json:"enabled"`
}

// Alert represents a triggered alert
type Alert struct {
	Rule        *AlertRule  `json:"rule"`
	NodeID      string      `json:"node_id"`
	Value       float64     `json:"value"`
	TriggeredAt time.Time   `json:"triggered_at"`
	Message     string      `json:"message"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(client *redis.Client, nodeID string, ctx context.Context) *MetricsCollector {
	return &MetricsCollector{
		client:         client,
		nodeID:         nodeID,
		ctx:            ctx,
		localMetrics:   &NodeMetrics{NodeID: nodeID},
		history:        make([]*MetricsSnapshot, 0, 100),
		maxHistory:     100,
		requestLatencies: make([]float64, 0, 1000),
		latencyWindow:  1000,
	}
}

// RecordLatency records a request latency
func (mc *MetricsCollector) RecordLatency(latencyMs float64) {
	mc.latencyMu.Lock()
	defer mc.latencyMu.Unlock()

	mc.requestLatencies = append(mc.requestLatencies, latencyMs)

	// Keep only last N latencies
	if len(mc.requestLatencies) > mc.latencyWindow {
		mc.requestLatencies = mc.requestLatencies[1:]
	}

	// Update metrics
	mc.metricsMu.Lock()
	mc.localMetrics.AverageLatencyMs = mc.calculateAverage(mc.requestLatencies)
	mc.localMetrics.P50LatencyMs = mc.calculatePercentile(mc.requestLatencies, 0.50)
	mc.localMetrics.P95LatencyMs = mc.calculatePercentile(mc.requestLatencies, 0.95)
	mc.localMetrics.P99LatencyMs = mc.calculatePercentile(mc.requestLatencies, 0.99)
	mc.metricsMu.Unlock()
}

// RecordConnection records connection event
func (mc *MetricsCollector) RecordConnection(opened bool) {
	mc.metricsMu.Lock()
	defer mc.metricsMu.Unlock()

	if opened {
		mc.localMetrics.ActiveConnections++
		mc.localMetrics.TotalConnectionsAll++
		mc.localMetrics.ConnectionsOpened++
	} else {
		mc.localMetrics.ActiveConnections--
		mc.localMetrics.ConnectionsClosed++
	}
}

// RecordMessage records message event
func (mc *MetricsCollector) RecordMessage(sent bool, bytes int64, failed bool) {
	mc.metricsMu.Lock()
	defer mc.metricsMu.Unlock()

	if failed {
		mc.localMetrics.MessagesFailed++
		return
	}

	if sent {
		mc.localMetrics.MessagesSent++
		mc.localMetrics.BytesSent += bytes
	} else {
		mc.localMetrics.MessagesReceived++
		mc.localMetrics.BytesReceived += bytes
	}
}

// RecordError records an error
func (mc *MetricsCollector) RecordError() {
	mc.metricsMu.Lock()
	defer mc.metricsMu.Unlock()

	totalMessages := mc.localMetrics.MessagesSent + mc.localMetrics.MessagesReceived
	if totalMessages > 0 {
		mc.localMetrics.ErrorRate = float64(mc.localMetrics.MessagesFailed) / float64(totalMessages)
	}
}

// UpdateResources updates resource usage metrics
func (mc *MetricsCollector) UpdateResources(cpuPercent, memoryPercent float64, goroutines int) {
	mc.metricsMu.Lock()
	defer mc.metricsMu.Unlock()

	mc.localMetrics.CPUUsage = cpuPercent
	mc.localMetrics.MemoryUsage = memoryPercent
	mc.localMetrics.GoroutineCount = goroutines
}

// GetMetrics returns current node metrics
func (mc *MetricsCollector) GetMetrics() *NodeMetrics {
	mc.metricsMu.RLock()
	defer mc.metricsMu.RUnlock()

	// Return a copy
	metrics := *mc.localMetrics
	metrics.Timestamp = time.Now()
	return &metrics
}

// PublishMetrics publishes metrics to Redis
func (mc *MetricsCollector) PublishMetrics() error {
	metrics := mc.GetMetrics()

	key := fmt.Sprintf("ws:metrics:node:%s", mc.nodeID)
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	return mc.client.Set(mc.ctx, key, data, 1*time.Minute).Err()
}

// GetClusterMetrics aggregates metrics from all nodes
func (mc *MetricsCollector) GetClusterMetrics() (*ClusterMetrics, error) {
	pattern := "ws:metrics:node:*"
	keys, err := mc.client.Keys(mc.ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics keys: %w", err)
	}

	var totalConnections int64
	var totalMessages int64
	var totalBytes int64
	var totalLatency float64
	var nodeCount int
	var healthyNodes int

	for _, key := range keys {
		data, err := mc.client.Get(mc.ctx, key).Result()
		if err != nil {
			continue
		}

		var metrics NodeMetrics
		if err := json.Unmarshal([]byte(data), &metrics); err != nil {
			continue
		}

		totalConnections += metrics.ActiveConnections
		totalMessages += metrics.MessagesSent + metrics.MessagesReceived
		totalBytes += metrics.BytesSent + metrics.BytesReceived
		totalLatency += metrics.AverageLatencyMs
		nodeCount++

		if metrics.ErrorRate < 0.01 { // Less than 1% error rate
			healthyNodes++
		}
	}

	avgLatency := float64(0)
	if nodeCount > 0 {
		avgLatency = totalLatency / float64(nodeCount)
	}

	return &ClusterMetrics{
		TotalConnections:    totalConnections,
		TotalNodes:          nodeCount,
		HealthyNodes:        healthyNodes,
		TotalMessagesPerSec: totalMessages / 60, // Approximate
		TotalBytesPerSec:    totalBytes / 60,
		AverageLatency:      avgLatency,
		LastUpdate:          time.Now(),
	}, nil
}

// SnapshotMetrics captures current state
func (mc *MetricsCollector) SnapshotMetrics() (*MetricsSnapshot, error) {
	clusterMetrics, err := mc.GetClusterMetrics()
	if err != nil {
		return nil, err
	}

	// Get all node metrics
	pattern := "ws:metrics:node:*"
	keys, err := mc.client.Keys(mc.ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	nodeMetrics := make([]*NodeMetrics, 0, len(keys))
	for _, key := range keys {
		data, err := mc.client.Get(mc.ctx, key).Result()
		if err != nil {
			continue
		}

		var metrics NodeMetrics
		if err := json.Unmarshal([]byte(data), &metrics); err != nil {
			continue
		}

		nodeMetrics = append(nodeMetrics, &metrics)
	}

	snapshot := &MetricsSnapshot{
		Timestamp: time.Now(),
		Cluster:   clusterMetrics,
		Nodes:     nodeMetrics,
	}

	// Store in history
	mc.historyMu.Lock()
	mc.history = append(mc.history, snapshot)
	if len(mc.history) > mc.maxHistory {
		mc.history = mc.history[1:]
	}
	mc.historyMu.Unlock()

	return snapshot, nil
}

// GetHistory returns historical metrics
func (mc *MetricsCollector) GetHistory(duration time.Duration) []*MetricsSnapshot {
	mc.historyMu.RLock()
	defer mc.historyMu.RUnlock()

	cutoff := time.Now().Add(-duration)
	filtered := make([]*MetricsSnapshot, 0)

	for _, snapshot := range mc.history {
		if snapshot.Timestamp.After(cutoff) {
			filtered = append(filtered, snapshot)
		}
	}

	return filtered
}

// CheckAlerts checks if any alert rules are triggered
func (mc *MetricsCollector) CheckAlerts(rules []*AlertRule) []*Alert {
	metrics := mc.GetMetrics()
	alerts := make([]*Alert, 0)

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		value := mc.extractMetricValue(metrics, rule.Metric)
		triggered := mc.evaluateCondition(value, rule.Operator, rule.Threshold)

		if triggered {
			alert := &Alert{
				Rule:        rule,
				NodeID:      mc.nodeID,
				Value:       value,
				TriggeredAt: time.Now(),
				Message:     fmt.Sprintf("%s %s %.2f (threshold: %.2f)", rule.Metric, rule.Operator, value, rule.Threshold),
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// extractMetricValue extracts a specific metric value by name
func (mc *MetricsCollector) extractMetricValue(metrics *NodeMetrics, name string) float64 {
	switch name {
	case "cpu_usage":
		return metrics.CPUUsage
	case "memory_usage":
		return metrics.MemoryUsage
	case "error_rate":
		return metrics.ErrorRate
	case "average_latency":
		return metrics.AverageLatencyMs
	case "p99_latency":
		return metrics.P99LatencyMs
	case "active_connections":
		return float64(metrics.ActiveConnections)
	default:
		return 0
	}
}

// evaluateCondition evaluates an alert condition
func (mc *MetricsCollector) evaluateCondition(value float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return value > threshold
	case "<":
		return value < threshold
	case ">=":
		return value >= threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	default:
		return false
	}
}

// calculateAverage calculates average of values
func (mc *MetricsCollector) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := float64(0)
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}

// calculatePercentile calculates percentile of values
func (mc *MetricsCollector) calculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Simple percentile calculation (not efficient for large datasets)
	// For production, use a proper percentile library

	// Copy and sort
	sorted := make([]float64, len(values))
	copy(sorted, values)

	// Bubble sort (simple implementation)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)) * percentile)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// ExportMetrics exports metrics in Prometheus format
func (mc *MetricsCollector) ExportMetrics() string {
	metrics := mc.GetMetrics()

	return fmt.Sprintf(`# HELP ws_active_connections Number of active WebSocket connections
# TYPE ws_active_connections gauge
ws_active_connections{node="%s"} %d

# HELP ws_messages_total Total number of messages
# TYPE ws_messages_total counter
ws_messages_sent_total{node="%s"} %d
ws_messages_received_total{node="%s"} %d

# HELP ws_latency_ms WebSocket message latency in milliseconds
# TYPE ws_latency_ms summary
ws_latency_ms{node="%s",quantile="0.5"} %.2f
ws_latency_ms{node="%s",quantile="0.95"} %.2f
ws_latency_ms{node="%s",quantile="0.99"} %.2f

# HELP ws_error_rate WebSocket error rate
# TYPE ws_error_rate gauge
ws_error_rate{node="%s"} %.4f

# HELP ws_cpu_usage CPU usage percentage
# TYPE ws_cpu_usage gauge
ws_cpu_usage{node="%s"} %.2f

# HELP ws_memory_usage Memory usage percentage
# TYPE ws_memory_usage gauge
ws_memory_usage{node="%s"} %.2f
`,
		mc.nodeID, metrics.ActiveConnections,
		mc.nodeID, metrics.MessagesSent,
		mc.nodeID, metrics.MessagesReceived,
		mc.nodeID, metrics.P50LatencyMs,
		mc.nodeID, metrics.P95LatencyMs,
		mc.nodeID, metrics.P99LatencyMs,
		mc.nodeID, metrics.ErrorRate,
		mc.nodeID, metrics.CPUUsage,
		mc.nodeID, metrics.MemoryUsage,
	)
}
