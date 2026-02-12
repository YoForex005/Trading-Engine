package admin

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ServerNode represents a server node in the cluster
type ServerNode struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Hostname        string    `json:"hostname"`
	IPAddress       string    `json:"ip_address"`
	Port            int       `json:"port"`
	Role            string    `json:"role"` // primary, secondary, failover, read_replica
	Status          string    `json:"status"` // online, offline, maintenance, degraded
	Region          string    `json:"region"`
	CPU             float64   `json:"cpu"`
	MemoryUsed      float64   `json:"memory_used"`
	MemoryTotal     float64   `json:"memory_total"`
	DiskUsed        float64   `json:"disk_used"`
	DiskTotal       float64   `json:"disk_total"`
	Connections     int       `json:"connections"`
	MaxConnections  int       `json:"max_connections"`
	Uptime          string    `json:"uptime"`
	Version         string    `json:"version"`
	LastHealthCheck time.Time `json:"last_health_check"`
	StartedAt       time.Time `json:"started_at"`
}

// LoadBalancerConfig represents load balancer configuration
type LoadBalancerConfig struct {
	Algorithm           string `json:"algorithm"` // round_robin, least_connections, weighted, ip_hash
	StickySession       bool   `json:"sticky_session"`
	HealthCheckInterval int    `json:"health_check_interval"` // seconds
	FailoverThresholdMs int    `json:"failover_threshold_ms"`
	MaxRetries          int    `json:"max_retries"`
	EnableCompression   bool   `json:"enable_compression"`
	Timeout             int    `json:"timeout"` // seconds
}

// ClusterHealth represents overall cluster health summary
type ClusterHealth struct {
	TotalNodes      int      `json:"total_nodes"`
	OnlineNodes     int      `json:"online_nodes"`
	TotalConnections int     `json:"total_connections"`
	AvgCPU          float64  `json:"avg_cpu"`
	AvgMemory       float64  `json:"avg_memory"`
	AvgLatencyMs    float64  `json:"avg_latency_ms"`
	Alerts          []string `json:"alerts"`
	Status          string   `json:"status"` // healthy, degraded, critical
	LastUpdated     time.Time `json:"last_updated"`
}

// NodeMetric represents time-series metrics for a node
type NodeMetric struct {
	NodeID         int64     `json:"node_id"`
	Timestamp      time.Time `json:"timestamp"`
	CPU            float64   `json:"cpu"`
	Memory         float64   `json:"memory"`
	Connections    int       `json:"connections"`
	RequestsPerSec float64   `json:"requests_per_sec"`
	AvgResponseMs  float64   `json:"avg_response_ms"`
	ErrorRate      float64   `json:"error_rate"`
}

// ClusterStats represents overall cluster statistics
type ClusterStats struct {
	TotalCapacity      int     `json:"total_capacity"` // max connections
	CurrentUtilization float64 `json:"current_utilization"` // percentage
	AvgUptime          string  `json:"avg_uptime"`
	TotalRequests      int64   `json:"total_requests"`
	TotalErrors        int64   `json:"total_errors"`
	SuccessRate        float64 `json:"success_rate"`
	DataTransferred    int64   `json:"data_transferred"` // bytes
	RegionDistribution map[string]int `json:"region_distribution"`
}

// UpdateNodeRequest represents request to update node configuration
type UpdateNodeRequest struct {
	Role           string `json:"role"`
	Status         string `json:"status"`
	MaxConnections int    `json:"max_connections"`
}

// ServerClusterService manages server cluster operations
type ServerClusterService struct {
	mu          sync.RWMutex
	nodes       map[int64]*ServerNode
	lbConfig    *LoadBalancerConfig
	metrics     map[int64][]NodeMetric // nodeId -> metrics
	nodeIDSeq   int64
}

// NewServerClusterService creates a new server cluster service with mock data
func NewServerClusterService() *ServerClusterService {
	s := &ServerClusterService{
		nodes:    make(map[int64]*ServerNode),
		metrics:  make(map[int64][]NodeMetric),
		nodeIDSeq: 1,
	}
	s.initializeMockData()
	return s
}

func (s *ServerClusterService) initializeMockData() {
	rand.Seed(time.Now().UnixNano())

	// Initialize load balancer config
	s.lbConfig = &LoadBalancerConfig{
		Algorithm:           "round_robin",
		StickySession:       true,
		HealthCheckInterval: 30,
		FailoverThresholdMs: 5000,
		MaxRetries:          3,
		EnableCompression:   true,
		Timeout:             60,
	}

	// Create 5 server nodes across 3 regions
	regions := []string{"us-east-1", "us-west-2", "eu-west-1"}
	roles := []string{"primary", "secondary", "secondary", "failover", "read_replica"}
	statuses := []string{"online", "online", "online", "online", "degraded"}

	for i := 0; i < 5; i++ {
		nodeID := s.nodeIDSeq
		s.nodeIDSeq++

		startedAt := time.Now().Add(-time.Duration(30+rand.Intn(300)) * 24 * time.Hour)
		uptime := time.Since(startedAt)

		memoryTotal := 32.0 + float64(rand.Intn(32)) // 32-64 GB
		memoryUsed := memoryTotal * (0.3 + rand.Float64()*0.4)
		diskTotal := 500.0 + float64(rand.Intn(500)) // 500-1000 GB
		diskUsed := diskTotal * (0.2 + rand.Float64()*0.5)

		cpu := 10.0 + rand.Float64()*70.0
		if statuses[i] == "degraded" {
			cpu = 85.0 + rand.Float64()*10.0
		}

		maxConn := 1000 + rand.Intn(4000)
		currentConn := int(float64(maxConn) * (0.1 + rand.Float64()*0.5))

		node := &ServerNode{
			ID:              nodeID,
			Name:            fmt.Sprintf("rtx5-node-%d", nodeID),
			Hostname:        fmt.Sprintf("node%d.rtx5.internal", nodeID),
			IPAddress:       fmt.Sprintf("10.0.%d.%d", (i/10)+1, (i%10)+1),
			Port:            7999 + i,
			Role:            roles[i],
			Status:          statuses[i],
			Region:          regions[i%len(regions)],
			CPU:             cpu,
			MemoryUsed:      memoryUsed,
			MemoryTotal:     memoryTotal,
			DiskUsed:        diskUsed,
			DiskTotal:       diskTotal,
			Connections:     currentConn,
			MaxConnections:  maxConn,
			Uptime:          formatDuration(uptime),
			Version:         fmt.Sprintf("v1.%d.%d", 2+rand.Intn(3), rand.Intn(10)),
			LastHealthCheck: time.Now().Add(-time.Duration(rand.Intn(60)) * time.Second),
			StartedAt:       startedAt,
		}

		s.nodes[nodeID] = node

		// Generate 24-hour metric history (hourly snapshots)
		metrics := make([]NodeMetric, 24)
		for hour := 0; hour < 24; hour++ {
			timestamp := time.Now().Add(-time.Duration(23-hour) * time.Hour)

			cpuMetric := 20.0 + rand.Float64()*60.0
			memoryMetric := 40.0 + rand.Float64()*40.0
			connMetric := 100 + rand.Intn(currentConn)
			reqPerSec := 50.0 + rand.Float64()*450.0
			avgResp := 10.0 + rand.Float64()*90.0
			errRate := rand.Float64() * 0.02 // 0-2% error rate

			if statuses[i] == "degraded" && hour >= 20 {
				cpuMetric = 85.0 + rand.Float64()*10.0
				avgResp = 100.0 + rand.Float64()*200.0
				errRate = 0.05 + rand.Float64()*0.05
			}

			metrics[hour] = NodeMetric{
				NodeID:         nodeID,
				Timestamp:      timestamp,
				CPU:            cpuMetric,
				Memory:         memoryMetric,
				Connections:    connMetric,
				RequestsPerSec: reqPerSec,
				AvgResponseMs:  avgResp,
				ErrorRate:      errRate,
			}
		}

		s.metrics[nodeID] = metrics
	}
}

// GetAllNodes returns all server nodes
func (s *ServerClusterService) GetAllNodes() []*ServerNode {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nodes := make([]*ServerNode, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodes = append(nodes, node)
	}

	return nodes
}

// GetNodeByID returns a specific node with its metrics
func (s *ServerClusterService) GetNodeByID(nodeID int64) (*ServerNode, []NodeMetric) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	node := s.nodes[nodeID]
	metrics := s.metrics[nodeID]

	return node, metrics
}

// UpdateNode updates node configuration
func (s *ServerClusterService) UpdateNode(nodeID int64, req UpdateNodeRequest) (*ServerNode, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, exists := s.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node not found")
	}

	if req.Role != "" {
		validRoles := map[string]bool{
			"primary": true, "secondary": true, "failover": true, "read_replica": true,
		}
		if !validRoles[req.Role] {
			return nil, fmt.Errorf("invalid role")
		}
		node.Role = req.Role
	}

	if req.Status != "" {
		validStatuses := map[string]bool{
			"online": true, "offline": true, "maintenance": true, "degraded": true,
		}
		if !validStatuses[req.Status] {
			return nil, fmt.Errorf("invalid status")
		}
		node.Status = req.Status
	}

	if req.MaxConnections > 0 {
		node.MaxConnections = req.MaxConnections
	}

	return node, nil
}

// RestartNode simulates node restart
func (s *ServerClusterService) RestartNode(nodeID int64) (*ServerNode, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, exists := s.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node not found")
	}

	node.Status = "maintenance"
	node.StartedAt = time.Now()
	node.Uptime = "0m"
	node.Connections = 0

	// Simulate restart completion after delay
	go func() {
		time.Sleep(3 * time.Second)
		s.mu.Lock()
		defer s.mu.Unlock()
		node.Status = "online"
	}()

	return node, nil
}

// GetClusterHealth returns overall cluster health
func (s *ServerClusterService) GetClusterHealth() *ClusterHealth {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalNodes := len(s.nodes)
	onlineNodes := 0
	totalConnections := 0
	totalCPU := 0.0
	totalMemory := 0.0
	alerts := make([]string, 0)

	for _, node := range s.nodes {
		if node.Status == "online" {
			onlineNodes++
		}
		totalConnections += node.Connections
		totalCPU += node.CPU

		memoryPercent := (node.MemoryUsed / node.MemoryTotal) * 100
		totalMemory += memoryPercent

		// Generate alerts
		if node.CPU > 80 {
			alerts = append(alerts, fmt.Sprintf("Node %s CPU usage critical: %.1f%%", node.Name, node.CPU))
		}
		if memoryPercent > 85 {
			alerts = append(alerts, fmt.Sprintf("Node %s memory usage critical: %.1f%%", node.Name, memoryPercent))
		}
		if node.Status == "degraded" {
			alerts = append(alerts, fmt.Sprintf("Node %s is in degraded state", node.Name))
		}
		if node.Status == "offline" {
			alerts = append(alerts, fmt.Sprintf("Node %s is offline", node.Name))
		}
	}

	avgCPU := totalCPU / float64(totalNodes)
	avgMemory := totalMemory / float64(totalNodes)

	// Calculate average latency from recent metrics
	avgLatency := 0.0
	latencyCount := 0
	for _, metrics := range s.metrics {
		if len(metrics) > 0 {
			avgLatency += metrics[len(metrics)-1].AvgResponseMs
			latencyCount++
		}
	}
	if latencyCount > 0 {
		avgLatency /= float64(latencyCount)
	}

	status := "healthy"
	if len(alerts) > 5 || onlineNodes < totalNodes-1 {
		status = "critical"
	} else if len(alerts) > 0 {
		status = "degraded"
	}

	return &ClusterHealth{
		TotalNodes:       totalNodes,
		OnlineNodes:      onlineNodes,
		TotalConnections: totalConnections,
		AvgCPU:           math.Round(avgCPU*100) / 100,
		AvgMemory:        math.Round(avgMemory*100) / 100,
		AvgLatencyMs:     math.Round(avgLatency*100) / 100,
		Alerts:           alerts,
		Status:           status,
		LastUpdated:      time.Now(),
	}
}

// GetLoadBalancerConfig returns current load balancer configuration
func (s *ServerClusterService) GetLoadBalancerConfig() *LoadBalancerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lbConfig
}

// UpdateLoadBalancerConfig updates load balancer configuration
func (s *ServerClusterService) UpdateLoadBalancerConfig(config *LoadBalancerConfig) *LoadBalancerConfig {
	s.mu.Lock()
	defer s.mu.Unlock()

	if config.Algorithm != "" {
		s.lbConfig.Algorithm = config.Algorithm
	}
	if config.HealthCheckInterval > 0 {
		s.lbConfig.HealthCheckInterval = config.HealthCheckInterval
	}
	if config.FailoverThresholdMs > 0 {
		s.lbConfig.FailoverThresholdMs = config.FailoverThresholdMs
	}
	if config.MaxRetries > 0 {
		s.lbConfig.MaxRetries = config.MaxRetries
	}
	if config.Timeout > 0 {
		s.lbConfig.Timeout = config.Timeout
	}

	s.lbConfig.StickySession = config.StickySession
	s.lbConfig.EnableCompression = config.EnableCompression

	return s.lbConfig
}

// GetClusterStats returns overall cluster statistics
func (s *ServerClusterService) GetClusterStats() *ClusterStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalCapacity := 0
	currentConnections := 0
	totalUptimeSeconds := int64(0)
	regionDist := make(map[string]int)

	for _, node := range s.nodes {
		totalCapacity += node.MaxConnections
		currentConnections += node.Connections

		uptime := time.Since(node.StartedAt)
		totalUptimeSeconds += int64(uptime.Seconds())

		regionDist[node.Region]++
	}

	utilization := 0.0
	if totalCapacity > 0 {
		utilization = (float64(currentConnections) / float64(totalCapacity)) * 100
	}

	avgUptimeSeconds := totalUptimeSeconds / int64(len(s.nodes))
	avgUptime := formatDuration(time.Duration(avgUptimeSeconds) * time.Second)

	// Calculate total requests and errors from metrics
	totalRequests := int64(0)
	totalErrors := int64(0)

	for _, metrics := range s.metrics {
		for _, m := range metrics {
			requests := int64(m.RequestsPerSec * 3600) // hourly
			errors := int64(float64(requests) * m.ErrorRate)
			totalRequests += requests
			totalErrors += errors
		}
	}

	successRate := 100.0
	if totalRequests > 0 {
		successRate = (1.0 - float64(totalErrors)/float64(totalRequests)) * 100
	}

	// Simulate data transferred (bytes)
	dataTransferred := totalRequests * (1024 + int64(rand.Intn(4096)))

	return &ClusterStats{
		TotalCapacity:      totalCapacity,
		CurrentUtilization: math.Round(utilization*100) / 100,
		AvgUptime:          avgUptime,
		TotalRequests:      totalRequests,
		TotalErrors:        totalErrors,
		SuccessRate:        math.Round(successRate*100) / 100,
		DataTransferred:    dataTransferred,
		RegionDistribution: regionDist,
	}
}

// ServerClusterHandler handles HTTP requests for server cluster API
type ServerClusterHandler struct {
	service     *ServerClusterService
	authService *AuthService
}

// NewServerClusterHandler creates a new server cluster handler
func NewServerClusterHandler(service *ServerClusterService, authService *AuthService) *ServerClusterHandler {
	return &ServerClusterHandler{
		service:     service,
		authService: authService,
	}
}

// HandleGetAllNodes returns all server nodes
func (h *ServerClusterHandler) HandleGetAllNodes(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	nodes := h.service.GetAllNodes()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(nodes)
}

// HandleGetNodeByID returns node detail with 24h metrics
func (h *ServerClusterHandler) HandleGetNodeByID(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Path[len("/admin/cluster/nodes/"):]
	if slashIdx := strings.Index(idStr, "/"); slashIdx != -1 {
		idStr = idStr[:slashIdx]
	}

	nodeID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid node ID", http.StatusBadRequest)
		return
	}

	node, metrics := h.service.GetNodeByID(nodeID)
	if node == nil {
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"node":    node,
		"metrics": metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
}

// HandleUpdateNode updates node configuration
func (h *ServerClusterHandler) HandleUpdateNode(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Path[len("/admin/cluster/nodes/"):]
	if slashIdx := strings.Index(idStr, "/"); slashIdx != -1 {
		idStr = idStr[:slashIdx]
	}

	nodeID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid node ID", http.StatusBadRequest)
		return
	}

	var req UpdateNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	node, err := h.service.UpdateNode(nodeID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(node)
}

// HandleRestartNode triggers node restart
func (h *ServerClusterHandler) HandleRestartNode(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	nodeID, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		http.Error(w, "Invalid node ID", http.StatusBadRequest)
		return
	}

	node, err := h.service.RestartNode(nodeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Node %s restart initiated", node.Name),
		"node":    node,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
}

// HandleGetClusterHealth returns cluster health summary
func (h *ServerClusterHandler) HandleGetClusterHealth(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	health := h.service.GetClusterHealth()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(health)
}

// HandleGetLoadBalancerConfig returns current LB config
func (h *ServerClusterHandler) HandleGetLoadBalancerConfig(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	config := h.service.GetLoadBalancerConfig()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(config)
}

// HandleUpdateLoadBalancerConfig updates LB configuration
func (h *ServerClusterHandler) HandleUpdateLoadBalancerConfig(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var config LoadBalancerConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updatedConfig := h.service.UpdateLoadBalancerConfig(&config)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(updatedConfig)
}

// HandleGetClusterStats returns cluster statistics
func (h *ServerClusterHandler) HandleGetClusterStats(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetClusterStats()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(stats)
}
