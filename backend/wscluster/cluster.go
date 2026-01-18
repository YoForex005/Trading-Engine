package wscluster

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// NodeStatus represents the health status of a cluster node
type NodeStatus string

const (
	NodeStatusHealthy   NodeStatus = "healthy"
	NodeStatusDegraded  NodeStatus = "degraded"
	NodeStatusUnhealthy NodeStatus = "unhealthy"
	NodeStatusOffline   NodeStatus = "offline"
)

// Node represents a WebSocket server instance in the cluster
type Node struct {
	ID              string                 `json:"id"`
	Address         string                 `json:"address"`
	Port            int                    `json:"port"`
	Status          NodeStatus             `json:"status"`
	Connections     int64                  `json:"connections"`
	MaxConnections  int64                  `json:"max_connections"`
	LastHeartbeat   time.Time              `json:"last_heartbeat"`
	Version         string                 `json:"version"`
	Region          string                 `json:"region,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CPUUsage        float64                `json:"cpu_usage"`
	MemoryUsage     float64                `json:"memory_usage"`
	MessagesPerSec  int64                  `json:"messages_per_sec"`
	BytesPerSec     int64                  `json:"bytes_per_sec"`
	ErrorRate       float64                `json:"error_rate"`
}

// ClusterConfig holds cluster-wide configuration
type ClusterConfig struct {
	// Redis connection
	RedisClient *redis.Client

	// Node configuration
	NodeID          string
	NodeAddress     string
	NodePort        int
	MaxConnections  int64
	Region          string

	// Timing configuration
	HeartbeatInterval    time.Duration
	NodeTimeout          time.Duration
	HealthCheckInterval  time.Duration

	// Scaling configuration
	ScaleUpThreshold     float64 // Trigger scale-up at this % of max connections
	ScaleDownThreshold   float64 // Trigger scale-down at this % of max connections
	MinNodes             int
	MaxNodes             int

	// Performance tuning
	MessageBufferSize    int
	EnableCompression    bool
	EnableBatching       bool
	BatchInterval        time.Duration
	MaxBatchSize         int
}

// Cluster manages the WebSocket server cluster
type Cluster struct {
	config        *ClusterConfig
	localNode     *Node
	nodes         map[string]*Node
	nodesMu       sync.RWMutex

	pubsub        *PubSubManager
	loadBalancer  *LoadBalancer
	failover      *FailoverManager

	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup

	// Metrics
	metrics       *ClusterMetrics
	metricsMu     sync.RWMutex

	// Callbacks
	onNodeJoin    func(*Node)
	onNodeLeave   func(*Node)
	onNodeUpdate  func(*Node)
}

// ClusterMetrics tracks cluster-wide metrics
type ClusterMetrics struct {
	TotalConnections    int64     `json:"total_connections"`
	TotalNodes          int       `json:"total_nodes"`
	HealthyNodes        int       `json:"healthy_nodes"`
	TotalMessagesPerSec int64     `json:"total_messages_per_sec"`
	TotalBytesPerSec    int64     `json:"total_bytes_per_sec"`
	AverageLatency      float64   `json:"average_latency_ms"`
	ErrorRate           float64   `json:"error_rate"`
	LastUpdate          time.Time `json:"last_update"`
}

// NewCluster creates a new cluster instance
func NewCluster(config *ClusterConfig) (*Cluster, error) {
	if config.NodeID == "" {
		config.NodeID = uuid.New().String()
	}

	// Set defaults
	if config.HeartbeatInterval == 0 {
		config.HeartbeatInterval = 5 * time.Second
	}
	if config.NodeTimeout == 0 {
		config.NodeTimeout = 15 * time.Second
	}
	if config.HealthCheckInterval == 0 {
		config.HealthCheckInterval = 10 * time.Second
	}
	if config.MaxConnections == 0 {
		config.MaxConnections = 10000 // Default per node
	}
	if config.ScaleUpThreshold == 0 {
		config.ScaleUpThreshold = 0.8 // 80%
	}
	if config.ScaleDownThreshold == 0 {
		config.ScaleDownThreshold = 0.3 // 30%
	}
	if config.MinNodes == 0 {
		config.MinNodes = 2
	}
	if config.MaxNodes == 0 {
		config.MaxNodes = 20
	}
	if config.MessageBufferSize == 0 {
		config.MessageBufferSize = 1000
	}
	if config.BatchInterval == 0 {
		config.BatchInterval = 100 * time.Millisecond
	}
	if config.MaxBatchSize == 0 {
		config.MaxBatchSize = 100
	}

	ctx, cancel := context.WithCancel(context.Background())

	cluster := &Cluster{
		config: config,
		localNode: &Node{
			ID:             config.NodeID,
			Address:        config.NodeAddress,
			Port:           config.NodePort,
			Status:         NodeStatusHealthy,
			Connections:    0,
			MaxConnections: config.MaxConnections,
			LastHeartbeat:  time.Now(),
			Version:        "1.0.0",
			Region:         config.Region,
			Metadata:       make(map[string]interface{}),
		},
		nodes:   make(map[string]*Node),
		ctx:     ctx,
		cancel:  cancel,
		metrics: &ClusterMetrics{},
	}

	// Initialize PubSub
	var err error
	cluster.pubsub, err = NewPubSubManager(config.RedisClient, config.NodeID)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create pubsub manager: %w", err)
	}

	// Initialize load balancer
	cluster.loadBalancer = NewLoadBalancer(cluster)

	// Initialize failover manager
	cluster.failover = NewFailoverManager(cluster)

	return cluster, nil
}

// Start begins cluster operations
func (c *Cluster) Start() error {
	// Register local node
	if err := c.registerNode(c.localNode); err != nil {
		return fmt.Errorf("failed to register local node: %w", err)
	}

	// Start PubSub
	if err := c.pubsub.Start(c.ctx); err != nil {
		return fmt.Errorf("failed to start pubsub: %w", err)
	}

	// Start background tasks
	c.wg.Add(4)
	go c.heartbeatLoop()
	go c.healthCheckLoop()
	go c.metricsUpdateLoop()
	go c.nodeDiscoveryLoop()

	return nil
}

// Stop gracefully shuts down the cluster
func (c *Cluster) Stop() error {
	// Cancel context
	c.cancel()

	// Deregister node
	if err := c.deregisterNode(c.localNode.ID); err != nil {
		return fmt.Errorf("failed to deregister node: %w", err)
	}

	// Stop PubSub
	if err := c.pubsub.Stop(); err != nil {
		return fmt.Errorf("failed to stop pubsub: %w", err)
	}

	// Wait for goroutines
	c.wg.Wait()

	return nil
}

// registerNode adds a node to Redis
func (c *Cluster) registerNode(node *Node) error {
	key := fmt.Sprintf("ws:cluster:node:%s", node.ID)

	data, err := json.Marshal(node)
	if err != nil {
		return err
	}

	return c.config.RedisClient.Set(c.ctx, key, data, c.config.NodeTimeout).Err()
}

// deregisterNode removes a node from Redis
func (c *Cluster) deregisterNode(nodeID string) error {
	key := fmt.Sprintf("ws:cluster:node:%s", nodeID)
	return c.config.RedisClient.Del(c.ctx, key).Err()
}

// heartbeatLoop sends periodic heartbeats
func (c *Cluster) heartbeatLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.localNode.LastHeartbeat = time.Now()

			if err := c.registerNode(c.localNode); err != nil {
				// Log error but continue
				continue
			}

			// Publish heartbeat event
			event := ClusterEvent{
				Type:      EventTypeHeartbeat,
				NodeID:    c.localNode.ID,
				Timestamp: time.Now(),
				Data:      c.localNode,
			}

			if err := c.pubsub.PublishEvent(&event); err != nil {
				// Log error but continue
			}
		}
	}
}

// healthCheckLoop monitors cluster health
func (c *Cluster) healthCheckLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.checkClusterHealth()
		}
	}
}

// checkClusterHealth verifies all nodes are healthy
func (c *Cluster) checkClusterHealth() {
	pattern := "ws:cluster:node:*"

	keys, err := c.config.RedisClient.Keys(c.ctx, pattern).Result()
	if err != nil {
		return
	}

	now := time.Now()
	activeNodes := make(map[string]*Node)

	for _, key := range keys {
		data, err := c.config.RedisClient.Get(c.ctx, key).Result()
		if err != nil {
			continue
		}

		var node Node
		if err := json.Unmarshal([]byte(data), &node); err != nil {
			continue
		}

		// Check if node is still alive
		if now.Sub(node.LastHeartbeat) > c.config.NodeTimeout {
			// Node is dead, trigger failover
			c.failover.HandleNodeFailure(&node)

			// Remove from Redis
			c.deregisterNode(node.ID)

			if c.onNodeLeave != nil {
				c.onNodeLeave(&node)
			}

			continue
		}

		activeNodes[node.ID] = &node
	}

	// Update local node map
	c.nodesMu.Lock()
	oldNodes := c.nodes
	c.nodes = activeNodes
	c.nodesMu.Unlock()

	// Detect new nodes
	for id, node := range activeNodes {
		if _, exists := oldNodes[id]; !exists && id != c.localNode.ID {
			if c.onNodeJoin != nil {
				c.onNodeJoin(node)
			}
		}
	}
}

// metricsUpdateLoop aggregates cluster metrics
func (c *Cluster) metricsUpdateLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.updateMetrics()
		}
	}
}

// updateMetrics calculates cluster-wide metrics
func (c *Cluster) updateMetrics() {
	c.nodesMu.RLock()
	nodes := c.nodes
	c.nodesMu.RUnlock()

	var totalConnections int64
	var totalMessagesPerSec int64
	var totalBytesPerSec int64
	var healthyNodes int
	var totalErrorRate float64

	for _, node := range nodes {
		totalConnections += node.Connections
		totalMessagesPerSec += node.MessagesPerSec
		totalBytesPerSec += node.BytesPerSec
		totalErrorRate += node.ErrorRate

		if node.Status == NodeStatusHealthy {
			healthyNodes++
		}
	}

	metrics := &ClusterMetrics{
		TotalConnections:    totalConnections,
		TotalNodes:          len(nodes),
		HealthyNodes:        healthyNodes,
		TotalMessagesPerSec: totalMessagesPerSec,
		TotalBytesPerSec:    totalBytesPerSec,
		LastUpdate:          time.Now(),
	}

	if len(nodes) > 0 {
		metrics.ErrorRate = totalErrorRate / float64(len(nodes))
	}

	c.metricsMu.Lock()
	c.metrics = metrics
	c.metricsMu.Unlock()

	// Publish metrics to Redis for monitoring
	c.publishMetrics(metrics)

	// Check if auto-scaling is needed
	c.checkAutoScaling(metrics)
}

// publishMetrics stores metrics in Redis
func (c *Cluster) publishMetrics(metrics *ClusterMetrics) {
	key := "ws:cluster:metrics"

	data, err := json.Marshal(metrics)
	if err != nil {
		return
	}

	c.config.RedisClient.Set(c.ctx, key, data, 10*time.Second)
}

// checkAutoScaling determines if cluster should scale
func (c *Cluster) checkAutoScaling(metrics *ClusterMetrics) {
	if metrics.TotalNodes == 0 {
		return
	}

	// Calculate average utilization
	totalCapacity := int64(metrics.TotalNodes) * c.config.MaxConnections
	utilization := float64(metrics.TotalConnections) / float64(totalCapacity)

	// Scale up if utilization is high
	if utilization > c.config.ScaleUpThreshold && metrics.TotalNodes < c.config.MaxNodes {
		event := ClusterEvent{
			Type:      EventTypeScaleUp,
			NodeID:    c.localNode.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"current_nodes":  metrics.TotalNodes,
				"utilization":    utilization,
				"threshold":      c.config.ScaleUpThreshold,
			},
		}
		c.pubsub.PublishEvent(&event)
	}

	// Scale down if utilization is low
	if utilization < c.config.ScaleDownThreshold && metrics.TotalNodes > c.config.MinNodes {
		event := ClusterEvent{
			Type:      EventTypeScaleDown,
			NodeID:    c.localNode.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"current_nodes":  metrics.TotalNodes,
				"utilization":    utilization,
				"threshold":      c.config.ScaleDownThreshold,
			},
		}
		c.pubsub.PublishEvent(&event)
	}
}

// nodeDiscoveryLoop handles service discovery
func (c *Cluster) nodeDiscoveryLoop() {
	defer c.wg.Done()

	// Subscribe to cluster events
	eventChan := c.pubsub.Subscribe("cluster")

	for {
		select {
		case <-c.ctx.Done():
			return
		case event := <-eventChan:
			c.handleClusterEvent(event)
		}
	}
}

// handleClusterEvent processes cluster events
func (c *Cluster) handleClusterEvent(event *ClusterEvent) {
	switch event.Type {
	case EventTypeNodeJoin:
		// Handle new node
		if node, ok := event.Data.(*Node); ok && c.onNodeJoin != nil {
			c.onNodeJoin(node)
		}
	case EventTypeNodeLeave:
		// Handle node departure
		if node, ok := event.Data.(*Node); ok && c.onNodeLeave != nil {
			c.onNodeLeave(node)
		}
	case EventTypeHeartbeat:
		// Update node information
		if node, ok := event.Data.(*Node); ok {
			c.nodesMu.Lock()
			c.nodes[node.ID] = node
			c.nodesMu.Unlock()

			if c.onNodeUpdate != nil {
				c.onNodeUpdate(node)
			}
		}
	}
}

// GetMetrics returns current cluster metrics
func (c *Cluster) GetMetrics() *ClusterMetrics {
	c.metricsMu.RLock()
	defer c.metricsMu.RUnlock()

	// Return a copy
	metrics := *c.metrics
	return &metrics
}

// GetNodes returns all active nodes
func (c *Cluster) GetNodes() []*Node {
	c.nodesMu.RLock()
	defer c.nodesMu.RUnlock()

	nodes := make([]*Node, 0, len(c.nodes))
	for _, node := range c.nodes {
		nodeCopy := *node
		nodes = append(nodes, &nodeCopy)
	}

	return nodes
}

// UpdateLocalNode updates local node statistics
func (c *Cluster) UpdateLocalNode(connections int64, messagesPerSec int64, bytesPerSec int64) {
	c.localNode.Connections = connections
	c.localNode.MessagesPerSec = messagesPerSec
	c.localNode.BytesPerSec = bytesPerSec
	c.localNode.LastHeartbeat = time.Now()
}

// SetCallbacks sets event callbacks
func (c *Cluster) SetCallbacks(onJoin, onLeave, onUpdate func(*Node)) {
	c.onNodeJoin = onJoin
	c.onNodeLeave = onLeave
	c.onNodeUpdate = onUpdate
}

// GetLoadBalancer returns the load balancer
func (c *Cluster) GetLoadBalancer() *LoadBalancer {
	return c.loadBalancer
}

// GetPubSub returns the PubSub manager
func (c *Cluster) GetPubSub() *PubSubManager {
	return c.pubsub
}

// GetFailover returns the failover manager
func (c *Cluster) GetFailover() *FailoverManager {
	return c.failover
}
