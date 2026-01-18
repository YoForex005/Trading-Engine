package wscluster

import (
	"fmt"
	"math/rand"
	"sync"
)

// LoadBalancingAlgorithm determines how connections are distributed
type LoadBalancingAlgorithm string

const (
	// AlgorithmRoundRobin distributes connections evenly in rotation
	AlgorithmRoundRobin LoadBalancingAlgorithm = "round-robin"

	// AlgorithmLeastConnections sends to node with fewest connections
	AlgorithmLeastConnections LoadBalancingAlgorithm = "least-connections"

	// AlgorithmWeighted uses weighted distribution based on capacity
	AlgorithmWeighted LoadBalancingAlgorithm = "weighted"

	// AlgorithmIPHash uses IP-based hashing for consistency
	AlgorithmIPHash LoadBalancingAlgorithm = "ip-hash"

	// AlgorithmRandom selects a random node
	AlgorithmRandom LoadBalancingAlgorithm = "random"

	// AlgorithmAdaptive uses dynamic selection based on real-time metrics
	AlgorithmAdaptive LoadBalancingAlgorithm = "adaptive"
)

// LoadBalancer distributes WebSocket connections across cluster nodes
type LoadBalancer struct {
	cluster   *Cluster
	algorithm LoadBalancingAlgorithm

	// Round-robin state
	rrIndex int
	rrMu    sync.Mutex

	// Weighted distribution
	weights map[string]int
	weightsMu sync.RWMutex

	// Adaptive thresholds
	cpuThreshold    float64
	memoryThreshold float64
	latencyThreshold float64

	// Metrics
	totalAssignments int64
	failedAssignments int64
	metricsMu sync.RWMutex
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(cluster *Cluster) *LoadBalancer {
	return &LoadBalancer{
		cluster:          cluster,
		algorithm:        AlgorithmAdaptive, // Default to adaptive
		weights:          make(map[string]int),
		cpuThreshold:     0.85,    // 85% CPU
		memoryThreshold:  0.85,    // 85% memory
		latencyThreshold: 100.0,   // 100ms
	}
}

// SelectNode selects the best node for a new connection
func (lb *LoadBalancer) SelectNode(clientIP string, userID string) (*Node, error) {
	nodes := lb.cluster.GetNodes()

	// Filter healthy nodes
	healthyNodes := lb.filterHealthyNodes(nodes)
	if len(healthyNodes) == 0 {
		lb.metricsMu.Lock()
		lb.failedAssignments++
		lb.metricsMu.Unlock()
		return nil, fmt.Errorf("no healthy nodes available")
	}

	var selectedNode *Node
	var err error

	switch lb.algorithm {
	case AlgorithmRoundRobin:
		selectedNode = lb.selectRoundRobin(healthyNodes)
	case AlgorithmLeastConnections:
		selectedNode = lb.selectLeastConnections(healthyNodes)
	case AlgorithmWeighted:
		selectedNode = lb.selectWeighted(healthyNodes)
	case AlgorithmIPHash:
		selectedNode = lb.selectIPHash(clientIP, healthyNodes)
	case AlgorithmRandom:
		selectedNode = lb.selectRandom(healthyNodes)
	case AlgorithmAdaptive:
		selectedNode = lb.selectAdaptive(healthyNodes)
	default:
		selectedNode = lb.selectAdaptive(healthyNodes)
	}

	if selectedNode == nil {
		lb.metricsMu.Lock()
		lb.failedAssignments++
		lb.metricsMu.Unlock()
		return nil, fmt.Errorf("failed to select node")
	}

	lb.metricsMu.Lock()
	lb.totalAssignments++
	lb.metricsMu.Unlock()

	return selectedNode, err
}

// filterHealthyNodes returns only healthy nodes with capacity
func (lb *LoadBalancer) filterHealthyNodes(nodes []*Node) []*Node {
	healthy := make([]*Node, 0, len(nodes))

	for _, node := range nodes {
		// Check health status
		if node.Status != NodeStatusHealthy {
			continue
		}

		// Check if node has capacity
		if node.Connections >= node.MaxConnections {
			continue
		}

		// Check resource thresholds (for adaptive)
		if lb.algorithm == AlgorithmAdaptive {
			if node.CPUUsage > lb.cpuThreshold {
				continue
			}
			if node.MemoryUsage > lb.memoryThreshold {
				continue
			}
		}

		healthy = append(healthy, node)
	}

	return healthy
}

// selectRoundRobin selects node using round-robin
func (lb *LoadBalancer) selectRoundRobin(nodes []*Node) *Node {
	if len(nodes) == 0 {
		return nil
	}

	lb.rrMu.Lock()
	defer lb.rrMu.Unlock()

	index := lb.rrIndex % len(nodes)
	lb.rrIndex++

	return nodes[index]
}

// selectLeastConnections selects node with fewest connections
func (lb *LoadBalancer) selectLeastConnections(nodes []*Node) *Node {
	if len(nodes) == 0 {
		return nil
	}

	minNode := nodes[0]
	minConnections := nodes[0].Connections

	for _, node := range nodes[1:] {
		if node.Connections < minConnections {
			minNode = node
			minConnections = node.Connections
		}
	}

	return minNode
}

// selectWeighted selects node based on weighted capacity
func (lb *LoadBalancer) selectWeighted(nodes []*Node) *Node {
	if len(nodes) == 0 {
		return nil
	}

	// Calculate total weight
	lb.weightsMu.RLock()
	totalWeight := 0
	for _, node := range nodes {
		weight := lb.weights[node.ID]
		if weight == 0 {
			weight = 1 // Default weight
		}

		// Adjust weight based on available capacity
		availableCapacity := node.MaxConnections - node.Connections
		effectiveWeight := int(float64(weight) * (float64(availableCapacity) / float64(node.MaxConnections)))

		totalWeight += effectiveWeight
	}
	lb.weightsMu.RUnlock()

	if totalWeight == 0 {
		return lb.selectLeastConnections(nodes)
	}

	// Select random point in weight range
	randPoint := rand.Intn(totalWeight)

	// Find node at that point
	currentWeight := 0
	for _, node := range nodes {
		lb.weightsMu.RLock()
		weight := lb.weights[node.ID]
		lb.weightsMu.RUnlock()

		if weight == 0 {
			weight = 1
		}

		availableCapacity := node.MaxConnections - node.Connections
		effectiveWeight := int(float64(weight) * (float64(availableCapacity) / float64(node.MaxConnections)))

		currentWeight += effectiveWeight
		if currentWeight > randPoint {
			return node
		}
	}

	return nodes[0]
}

// selectIPHash selects node based on client IP hash
func (lb *LoadBalancer) selectIPHash(clientIP string, nodes []*Node) *Node {
	if len(nodes) == 0 {
		return nil
	}

	// Use session affinity's hash function
	affinity := NewSessionAffinity(lb.cluster.config.RedisClient, lb.cluster.ctx)
	return affinity.selectByHash(clientIP, nodes)
}

// selectRandom selects a random node
func (lb *LoadBalancer) selectRandom(nodes []*Node) *Node {
	if len(nodes) == 0 {
		return nil
	}

	return nodes[rand.Intn(len(nodes))]
}

// selectAdaptive uses dynamic selection based on real-time metrics
func (lb *LoadBalancer) selectAdaptive(nodes []*Node) *Node {
	if len(nodes) == 0 {
		return nil
	}

	// Score each node based on multiple factors
	type nodeScore struct {
		node  *Node
		score float64
	}

	scores := make([]nodeScore, 0, len(nodes))

	for _, node := range nodes {
		score := lb.calculateNodeScore(node)
		scores = append(scores, nodeScore{node: node, score: score})
	}

	// Select node with highest score
	best := scores[0]
	for _, ns := range scores[1:] {
		if ns.score > best.score {
			best = ns
		}
	}

	return best.node
}

// calculateNodeScore calculates a comprehensive score for a node
func (lb *LoadBalancer) calculateNodeScore(node *Node) float64 {
	// Factors:
	// 1. Available capacity (40% weight)
	// 2. CPU usage (20% weight)
	// 3. Memory usage (20% weight)
	// 4. Error rate (10% weight)
	// 5. Message throughput (10% weight)

	// Capacity score (higher is better)
	capacityScore := float64(node.MaxConnections-node.Connections) / float64(node.MaxConnections)

	// CPU score (lower usage is better)
	cpuScore := 1.0 - node.CPUUsage

	// Memory score (lower usage is better)
	memoryScore := 1.0 - node.MemoryUsage

	// Error rate score (lower is better)
	errorScore := 1.0 - node.ErrorRate

	// Throughput score (normalize to 0-1 range)
	throughputScore := 0.5 // Default if no throughput data
	if node.MessagesPerSec > 0 {
		// Assume 10000 msg/s is max
		throughputScore = float64(node.MessagesPerSec) / 10000.0
		if throughputScore > 1.0 {
			throughputScore = 1.0
		}
	}

	// Calculate weighted score
	score := (capacityScore * 0.4) +
		(cpuScore * 0.2) +
		(memoryScore * 0.2) +
		(errorScore * 0.1) +
		(throughputScore * 0.1)

	return score
}

// SetAlgorithm sets the load balancing algorithm
func (lb *LoadBalancer) SetAlgorithm(algorithm LoadBalancingAlgorithm) {
	lb.algorithm = algorithm
}

// SetNodeWeight sets the weight for a specific node (for weighted algorithm)
func (lb *LoadBalancer) SetNodeWeight(nodeID string, weight int) {
	lb.weightsMu.Lock()
	defer lb.weightsMu.Unlock()

	lb.weights[nodeID] = weight
}

// SetThresholds sets adaptive algorithm thresholds
func (lb *LoadBalancer) SetThresholds(cpu, memory, latency float64) {
	lb.cpuThreshold = cpu
	lb.memoryThreshold = memory
	lb.latencyThreshold = latency
}

// GetMetrics returns load balancer metrics
func (lb *LoadBalancer) GetMetrics() map[string]interface{} {
	lb.metricsMu.RLock()
	defer lb.metricsMu.RUnlock()

	successRate := float64(0)
	if lb.totalAssignments > 0 {
		successRate = float64(lb.totalAssignments-lb.failedAssignments) / float64(lb.totalAssignments)
	}

	return map[string]interface{}{
		"total_assignments":  lb.totalAssignments,
		"failed_assignments": lb.failedAssignments,
		"success_rate":       successRate,
		"algorithm":          string(lb.algorithm),
	}
}

// GetNodeDistribution returns connection distribution across nodes
func (lb *LoadBalancer) GetNodeDistribution() map[string]int64 {
	nodes := lb.cluster.GetNodes()
	distribution := make(map[string]int64)

	for _, node := range nodes {
		distribution[node.ID] = node.Connections
	}

	return distribution
}

// CalculateLoadImbalance calculates the coefficient of variation for load distribution
func (lb *LoadBalancer) CalculateLoadImbalance() float64 {
	nodes := lb.cluster.GetNodes()
	if len(nodes) == 0 {
		return 0
	}

	// Calculate mean load
	var totalLoad float64
	for _, node := range nodes {
		load := float64(node.Connections) / float64(node.MaxConnections)
		totalLoad += load
	}
	meanLoad := totalLoad / float64(len(nodes))

	// Calculate standard deviation
	var variance float64
	for _, node := range nodes {
		load := float64(node.Connections) / float64(node.MaxConnections)
		diff := load - meanLoad
		variance += diff * diff
	}
	stdDev := variance / float64(len(nodes))

	// Coefficient of variation (lower is better)
	if meanLoad == 0 {
		return 0
	}

	return stdDev / meanLoad
}

// RebalanceRecommendation suggests if rebalancing is needed
func (lb *LoadBalancer) RebalanceRecommendation() bool {
	imbalance := lb.CalculateLoadImbalance()

	// If coefficient of variation > 0.3, recommend rebalancing
	return imbalance > 0.3
}

// GetAlgorithmInfo returns information about the current algorithm
func (lb *LoadBalancer) GetAlgorithmInfo() map[string]interface{} {
	info := map[string]interface{}{
		"algorithm": string(lb.algorithm),
	}

	switch lb.algorithm {
	case AlgorithmWeighted:
		lb.weightsMu.RLock()
		info["weights"] = lb.weights
		lb.weightsMu.RUnlock()

	case AlgorithmAdaptive:
		info["cpu_threshold"] = lb.cpuThreshold
		info["memory_threshold"] = lb.memoryThreshold
		info["latency_threshold"] = lb.latencyThreshold
	}

	return info
}
