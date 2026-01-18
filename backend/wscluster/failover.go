package wscluster

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// FailoverStrategy determines how failover is handled
type FailoverStrategy string

const (
	// FailoverImmediate migrates connections immediately
	FailoverImmediate FailoverStrategy = "immediate"

	// FailoverGraceful allows in-flight operations to complete
	FailoverGraceful FailoverStrategy = "graceful"

	// FailoverDelayed waits for a period before migrating
	FailoverDelayed FailoverStrategy = "delayed"
)

// FailoverManager handles automatic failover when nodes fail
type FailoverManager struct {
	cluster  *Cluster
	affinity *SessionAffinity

	// Failover configuration
	strategy        FailoverStrategy
	gracePeriod     time.Duration
	maxRetries      int
	retryDelay      time.Duration

	// Active failovers
	activeFailovers map[string]*FailoverOperation
	failoverMu      sync.RWMutex

	// Context
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Metrics
	totalFailovers      int64
	successfulFailovers int64
	failedFailovers     int64
	metricsMu           sync.RWMutex
}

// FailoverOperation represents an ongoing failover
type FailoverOperation struct {
	ID              string                 `json:"id"`
	FailedNodeID    string                 `json:"failed_node_id"`
	TargetNodeID    string                 `json:"target_node_id"`
	StartTime       time.Time              `json:"start_time"`
	Status          string                 `json:"status"`
	SessionsMigrated int                   `json:"sessions_migrated"`
	SessionsFailed  int                    `json:"sessions_failed"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// NewFailoverManager creates a new failover manager
func NewFailoverManager(cluster *Cluster) *FailoverManager {
	ctx, cancel := context.WithCancel(context.Background())

	// Get or create session affinity
	affinity := NewSessionAffinity(cluster.config.RedisClient, ctx)

	return &FailoverManager{
		cluster:         cluster,
		affinity:        affinity,
		strategy:        FailoverGraceful,
		gracePeriod:     5 * time.Second,
		maxRetries:      3,
		retryDelay:      1 * time.Second,
		activeFailovers: make(map[string]*FailoverOperation),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// HandleNodeFailure initiates failover for a failed node
func (fm *FailoverManager) HandleNodeFailure(failedNode *Node) error {
	fm.metricsMu.Lock()
	fm.totalFailovers++
	fm.metricsMu.Unlock()

	// Select target node for migration
	targetNode := fm.selectTargetNode(failedNode)
	if targetNode == nil {
		fm.metricsMu.Lock()
		fm.failedFailovers++
		fm.metricsMu.Unlock()
		return fmt.Errorf("no available nodes for failover")
	}

	// Create failover operation
	operation := &FailoverOperation{
		ID:           fmt.Sprintf("failover-%d", time.Now().UnixNano()),
		FailedNodeID: failedNode.ID,
		TargetNodeID: targetNode.ID,
		StartTime:    time.Now(),
		Status:       "in_progress",
	}

	fm.failoverMu.Lock()
	fm.activeFailovers[operation.ID] = operation
	fm.failoverMu.Unlock()

	// Execute failover based on strategy
	switch fm.strategy {
	case FailoverImmediate:
		return fm.executeImmediateFailover(operation, failedNode, targetNode)
	case FailoverGraceful:
		return fm.executeGracefulFailover(operation, failedNode, targetNode)
	case FailoverDelayed:
		return fm.executeDelayedFailover(operation, failedNode, targetNode)
	default:
		return fm.executeGracefulFailover(operation, failedNode, targetNode)
	}
}

// executeImmediateFailover migrates all connections immediately
func (fm *FailoverManager) executeImmediateFailover(
	operation *FailoverOperation,
	failedNode *Node,
	targetNode *Node,
) error {
	// Migrate sessions
	migrated, err := fm.affinity.MigrateSessions(
		failedNode.ID,
		targetNode.ID,
		targetNode.Address,
	)

	operation.SessionsMigrated = migrated
	operation.Status = "completed"

	if err != nil {
		operation.Status = "failed"
		fm.metricsMu.Lock()
		fm.failedFailovers++
		fm.metricsMu.Unlock()
		return err
	}

	// Publish failover event
	event := ClusterEvent{
		Type:      EventTypeFailover,
		NodeID:    fm.cluster.localNode.ID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"operation_id":      operation.ID,
			"failed_node_id":    failedNode.ID,
			"target_node_id":    targetNode.ID,
			"sessions_migrated": migrated,
			"strategy":          "immediate",
		},
	}

	fm.cluster.pubsub.PublishEvent(&event)

	fm.metricsMu.Lock()
	fm.successfulFailovers++
	fm.metricsMu.Unlock()

	return nil
}

// executeGracefulFailover allows in-flight operations to complete
func (fm *FailoverManager) executeGracefulFailover(
	operation *FailoverOperation,
	failedNode *Node,
	targetNode *Node,
) error {
	// Wait for grace period
	time.Sleep(fm.gracePeriod)

	// Then execute immediate failover
	return fm.executeImmediateFailover(operation, failedNode, targetNode)
}

// executeDelayedFailover waits before migrating
func (fm *FailoverManager) executeDelayedFailover(
	operation *FailoverOperation,
	failedNode *Node,
	targetNode *Node,
) error {
	// Launch async failover
	fm.wg.Add(1)
	go func() {
		defer fm.wg.Done()

		// Wait for grace period
		select {
		case <-time.After(fm.gracePeriod):
		case <-fm.ctx.Done():
			return
		}

		// Execute failover
		fm.executeImmediateFailover(operation, failedNode, targetNode)
	}()

	return nil
}

// selectTargetNode chooses the best node for failover
func (fm *FailoverManager) selectTargetNode(failedNode *Node) *Node {
	nodes := fm.cluster.GetNodes()

	var bestNode *Node
	var minLoad float64 = 1.0

	for _, node := range nodes {
		// Skip failed node and unhealthy nodes
		if node.ID == failedNode.ID || node.Status != NodeStatusHealthy {
			continue
		}

		// Calculate load
		load := float64(node.Connections) / float64(node.MaxConnections)

		// Select node with lowest load
		if bestNode == nil || load < minLoad {
			bestNode = node
			minLoad = load
		}
	}

	return bestNode
}

// RebalanceConnections redistributes connections across healthy nodes
func (fm *FailoverManager) RebalanceConnections() error {
	nodes := fm.cluster.GetNodes()
	if len(nodes) <= 1 {
		return nil
	}

	// Calculate average load
	var totalConnections int64
	var totalCapacity int64

	for _, node := range nodes {
		if node.Status == NodeStatusHealthy {
			totalConnections += node.Connections
			totalCapacity += node.MaxConnections
		}
	}

	if totalCapacity == 0 {
		return fmt.Errorf("no capacity available")
	}

	targetLoad := float64(totalConnections) / float64(totalCapacity)

	// Identify overloaded and underloaded nodes
	var overloaded []*Node
	var underloaded []*Node

	for _, node := range nodes {
		if node.Status != NodeStatusHealthy {
			continue
		}

		load := float64(node.Connections) / float64(node.MaxConnections)

		if load > targetLoad+0.1 { // 10% threshold
			overloaded = append(overloaded, node)
		} else if load < targetLoad-0.1 {
			underloaded = append(underloaded, node)
		}
	}

	// Migrate connections from overloaded to underloaded nodes
	for _, srcNode := range overloaded {
		if len(underloaded) == 0 {
			break
		}

		dstNode := underloaded[0]

		// Calculate how many to migrate
		srcLoad := float64(srcNode.Connections) / float64(srcNode.MaxConnections)
		excessConnections := int64((srcLoad - targetLoad) * float64(srcNode.MaxConnections))

		if excessConnections <= 0 {
			continue
		}

		// Migrate sessions
		sessions, err := fm.affinity.GetSessionsForNode(srcNode.ID)
		if err != nil {
			continue
		}

		migrated := 0
		for _, session := range sessions {
			if int64(migrated) >= excessConnections {
				break
			}

			// Reassign session
			if err := fm.affinity.AssignSession(
				session.UserID,
				session.ConnectionID,
				dstNode.ID,
				dstNode.Address,
				session.Metadata,
			); err != nil {
				continue
			}

			migrated++
		}

		// Update underloaded nodes list
		dstLoad := float64(dstNode.Connections+int64(migrated)) / float64(dstNode.MaxConnections)
		if dstLoad >= targetLoad {
			underloaded = underloaded[1:]
		}
	}

	return nil
}

// GetActiveFailovers returns all active failover operations
func (fm *FailoverManager) GetActiveFailovers() []*FailoverOperation {
	fm.failoverMu.RLock()
	defer fm.failoverMu.RUnlock()

	operations := make([]*FailoverOperation, 0, len(fm.activeFailovers))
	for _, op := range fm.activeFailovers {
		operations = append(operations, op)
	}

	return operations
}

// GetMetrics returns failover metrics
func (fm *FailoverManager) GetMetrics() map[string]interface{} {
	fm.metricsMu.RLock()
	defer fm.metricsMu.RUnlock()

	successRate := float64(0)
	if fm.totalFailovers > 0 {
		successRate = float64(fm.successfulFailovers) / float64(fm.totalFailovers)
	}

	return map[string]interface{}{
		"total_failovers":      fm.totalFailovers,
		"successful_failovers": fm.successfulFailovers,
		"failed_failovers":     fm.failedFailovers,
		"success_rate":         successRate,
		"active_failovers":     len(fm.activeFailovers),
	}
}

// SetStrategy sets the failover strategy
func (fm *FailoverManager) SetStrategy(strategy FailoverStrategy) {
	fm.strategy = strategy
}

// SetGracePeriod sets the grace period for graceful failover
func (fm *FailoverManager) SetGracePeriod(period time.Duration) {
	fm.gracePeriod = period
}

// Stop gracefully shuts down the failover manager
func (fm *FailoverManager) Stop() error {
	fm.cancel()
	fm.wg.Wait()
	return nil
}

// RetryFailedOperation retries a failed failover operation
func (fm *FailoverManager) RetryFailedOperation(operationID string) error {
	fm.failoverMu.RLock()
	operation, exists := fm.activeFailovers[operationID]
	fm.failoverMu.RUnlock()

	if !exists {
		return fmt.Errorf("operation not found: %s", operationID)
	}

	if operation.Status != "failed" {
		return fmt.Errorf("operation is not in failed state")
	}

	// Get failed and target nodes
	var failedNode, targetNode *Node
	for _, node := range fm.cluster.GetNodes() {
		if node.ID == operation.FailedNodeID {
			failedNode = node
		}
		if node.ID == operation.TargetNodeID {
			targetNode = node
		}
	}

	if failedNode == nil || targetNode == nil {
		return fmt.Errorf("nodes not found")
	}

	// Retry with exponential backoff
	for i := 0; i < fm.maxRetries; i++ {
		if err := fm.executeImmediateFailover(operation, failedNode, targetNode); err == nil {
			return nil
		}

		if i < fm.maxRetries-1 {
			time.Sleep(fm.retryDelay * time.Duration(1<<uint(i)))
		}
	}

	return fmt.Errorf("retry failed after %d attempts", fm.maxRetries)
}
