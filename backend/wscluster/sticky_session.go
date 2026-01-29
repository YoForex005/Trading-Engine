package wscluster

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// SessionAffinity manages sticky sessions for WebSocket connections
type SessionAffinity struct {
	client *redis.Client
	ctx    context.Context

	// Local cache for faster lookups
	localCache map[string]*SessionMapping
	cacheMu    sync.RWMutex
	cacheTTL   time.Duration

	// Session configuration
	sessionTTL time.Duration
	enableCache bool

	// Metrics
	cacheHits   int64
	cacheMisses int64
	metricsMu   sync.RWMutex
}

// SessionMapping maps a user/connection to a node
type SessionMapping struct {
	UserID       string    `json:"user_id"`
	ConnectionID string    `json:"connection_id"`
	NodeID       string    `json:"node_id"`
	NodeAddress  string    `json:"node_address"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccess   time.Time `json:"last_access"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// AffinityStrategy determines how sessions are assigned to nodes
type AffinityStrategy string

const (
	// AffinityByUserID assigns based on user ID (same user always goes to same node)
	AffinityByUserID AffinityStrategy = "user_id"

	// AffinityByConnection assigns based on connection ID
	AffinityByConnection AffinityStrategy = "connection_id"

	// AffinityByHash uses consistent hashing
	AffinityByHash AffinityStrategy = "hash"

	// AffinityByRegion assigns based on geographic region
	AffinityByRegion AffinityStrategy = "region"

	// AffinityNone allows any node (round-robin)
	AffinityNone AffinityStrategy = "none"
)

// NewSessionAffinity creates a new session affinity manager
func NewSessionAffinity(client *redis.Client, ctx context.Context) *SessionAffinity {
	return &SessionAffinity{
		client:      client,
		ctx:         ctx,
		localCache:  make(map[string]*SessionMapping),
		cacheTTL:    5 * time.Minute,
		sessionTTL:  30 * time.Minute,
		enableCache: true,
	}
}

// AssignSession assigns a user/connection to a node
func (sa *SessionAffinity) AssignSession(userID, connectionID, nodeID, nodeAddress string, metadata map[string]interface{}) error {
	mapping := &SessionMapping{
		UserID:       userID,
		ConnectionID: connectionID,
		NodeID:       nodeID,
		NodeAddress:  nodeAddress,
		CreatedAt:    time.Now(),
		LastAccess:   time.Now(),
		Metadata:     metadata,
	}

	// Store in Redis
	key := sa.getUserKey(userID)
	data, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal session mapping: %w", err)
	}

	if err := sa.client.Set(sa.ctx, key, data, sa.sessionTTL).Err(); err != nil {
		return fmt.Errorf("failed to store session mapping: %w", err)
	}

	// Also index by connection ID for faster lookups
	connKey := sa.getConnectionKey(connectionID)
	if err := sa.client.Set(sa.ctx, connKey, data, sa.sessionTTL).Err(); err != nil {
		return fmt.Errorf("failed to store connection mapping: %w", err)
	}

	// Update local cache
	if sa.enableCache {
		sa.cacheMu.Lock()
		sa.localCache[userID] = mapping
		sa.cacheMu.Unlock()
	}

	return nil
}

// GetNodeForUser returns the node assigned to a user
func (sa *SessionAffinity) GetNodeForUser(userID string) (*SessionMapping, error) {
	// Check local cache first
	if sa.enableCache {
		sa.cacheMu.RLock()
		if mapping, exists := sa.localCache[userID]; exists {
			sa.cacheMu.RUnlock()

			sa.metricsMu.Lock()
			sa.cacheHits++
			sa.metricsMu.Unlock()

			// Refresh TTL
			mapping.LastAccess = time.Now()
			return mapping, nil
		}
		sa.cacheMu.RUnlock()

		sa.metricsMu.Lock()
		sa.cacheMisses++
		sa.metricsMu.Unlock()
	}

	// Fetch from Redis
	key := sa.getUserKey(userID)
	data, err := sa.client.Get(sa.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No mapping exists
		}
		return nil, fmt.Errorf("failed to get session mapping: %w", err)
	}

	var mapping SessionMapping
	if err := json.Unmarshal([]byte(data), &mapping); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session mapping: %w", err)
	}

	// Update cache
	if sa.enableCache {
		sa.cacheMu.Lock()
		sa.localCache[userID] = &mapping
		sa.cacheMu.Unlock()
	}

	return &mapping, nil
}

// GetNodeForConnection returns the node assigned to a connection
func (sa *SessionAffinity) GetNodeForConnection(connectionID string) (*SessionMapping, error) {
	key := sa.getConnectionKey(connectionID)
	data, err := sa.client.Get(sa.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get connection mapping: %w", err)
	}

	var mapping SessionMapping
	if err := json.Unmarshal([]byte(data), &mapping); err != nil {
		return nil, fmt.Errorf("failed to unmarshal connection mapping: %w", err)
	}

	return &mapping, nil
}

// RemoveSession removes a user's session mapping
func (sa *SessionAffinity) RemoveSession(userID, connectionID string) error {
	// Remove from Redis
	userKey := sa.getUserKey(userID)
	connKey := sa.getConnectionKey(connectionID)

	pipe := sa.client.Pipeline()
	pipe.Del(sa.ctx, userKey)
	pipe.Del(sa.ctx, connKey)

	if _, err := pipe.Exec(sa.ctx); err != nil {
		return fmt.Errorf("failed to remove session mapping: %w", err)
	}

	// Remove from cache
	if sa.enableCache {
		sa.cacheMu.Lock()
		delete(sa.localCache, userID)
		sa.cacheMu.Unlock()
	}

	return nil
}

// RefreshSession updates the TTL for a session
func (sa *SessionAffinity) RefreshSession(userID, connectionID string) error {
	userKey := sa.getUserKey(userID)
	connKey := sa.getConnectionKey(connectionID)

	pipe := sa.client.Pipeline()
	pipe.Expire(sa.ctx, userKey, sa.sessionTTL)
	pipe.Expire(sa.ctx, connKey, sa.sessionTTL)

	if _, err := pipe.Exec(sa.ctx); err != nil {
		return fmt.Errorf("failed to refresh session: %w", err)
	}

	return nil
}

// GetSessionsForNode returns all sessions assigned to a node
func (sa *SessionAffinity) GetSessionsForNode(nodeID string) ([]*SessionMapping, error) {
	pattern := "ws:session:user:*"

	keys, err := sa.client.Keys(sa.ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to scan keys: %w", err)
	}

	var sessions []*SessionMapping

	for _, key := range keys {
		data, err := sa.client.Get(sa.ctx, key).Result()
		if err != nil {
			continue
		}

		var mapping SessionMapping
		if err := json.Unmarshal([]byte(data), &mapping); err != nil {
			continue
		}

		if mapping.NodeID == nodeID {
			sessions = append(sessions, &mapping)
		}
	}

	return sessions, nil
}

// MigrateSessions moves sessions from one node to another (for failover)
func (sa *SessionAffinity) MigrateSessions(fromNodeID, toNodeID, toNodeAddress string) (int, error) {
	sessions, err := sa.GetSessionsForNode(fromNodeID)
	if err != nil {
		return 0, fmt.Errorf("failed to get sessions: %w", err)
	}

	count := 0
	for _, session := range sessions {
		// Update node assignment
		session.NodeID = toNodeID
		session.NodeAddress = toNodeAddress
		session.LastAccess = time.Now()

		// Store updated mapping
		if err := sa.AssignSession(
			session.UserID,
			session.ConnectionID,
			toNodeID,
			toNodeAddress,
			session.Metadata,
		); err != nil {
			// Log error but continue with other sessions
			continue
		}

		count++
	}

	return count, nil
}

// SelectNodeByStrategy selects a node based on affinity strategy
func (sa *SessionAffinity) SelectNodeByStrategy(
	strategy AffinityStrategy,
	userID string,
	availableNodes []*Node,
) *Node {
	if len(availableNodes) == 0 {
		return nil
	}

	switch strategy {
	case AffinityByUserID:
		return sa.selectByHash(userID, availableNodes)

	case AffinityByHash:
		return sa.selectByHash(userID, availableNodes)

	case AffinityByRegion:
		// TODO: Implement region-based selection
		return availableNodes[0]

	case AffinityNone:
		// Select least loaded node
		return sa.selectLeastLoaded(availableNodes)

	default:
		return availableNodes[0]
	}
}

// selectByHash uses consistent hashing to select a node
func (sa *SessionAffinity) selectByHash(key string, nodes []*Node) *Node {
	if len(nodes) == 0 {
		return nil
	}

	// Calculate MD5 hash
	hash := md5.Sum([]byte(key))
	_ = hex.EncodeToString(hash[:]) // hashStr for debugging if needed

	// Convert to integer and mod by number of nodes
	var hashInt int64
	for i := 0; i < 8; i++ {
		hashInt = (hashInt << 8) | int64(hash[i])
	}

	index := int(hashInt) % len(nodes)
	if index < 0 {
		index = -index
	}

	return nodes[index]
}

// selectLeastLoaded selects the node with lowest connection count
func (sa *SessionAffinity) selectLeastLoaded(nodes []*Node) *Node {
	if len(nodes) == 0 {
		return nil
	}

	minNode := nodes[0]
	minLoad := float64(minNode.Connections) / float64(minNode.MaxConnections)

	for _, node := range nodes[1:] {
		load := float64(node.Connections) / float64(node.MaxConnections)
		if load < minLoad {
			minLoad = load
			minNode = node
		}
	}

	return minNode
}

// CleanupExpiredSessions removes expired sessions from cache
func (sa *SessionAffinity) CleanupExpiredSessions() {
	if !sa.enableCache {
		return
	}

	sa.cacheMu.Lock()
	defer sa.cacheMu.Unlock()

	now := time.Now()
	for userID, mapping := range sa.localCache {
		if now.Sub(mapping.LastAccess) > sa.cacheTTL {
			delete(sa.localCache, userID)
		}
	}
}

// GetMetrics returns session affinity metrics
func (sa *SessionAffinity) GetMetrics() map[string]interface{} {
	sa.metricsMu.RLock()
	defer sa.metricsMu.RUnlock()

	sa.cacheMu.RLock()
	cacheSize := len(sa.localCache)
	sa.cacheMu.RUnlock()

	hitRate := float64(0)
	if sa.cacheHits+sa.cacheMisses > 0 {
		hitRate = float64(sa.cacheHits) / float64(sa.cacheHits+sa.cacheMisses)
	}

	return map[string]interface{}{
		"cache_hits":    sa.cacheHits,
		"cache_misses":  sa.cacheMisses,
		"cache_size":    cacheSize,
		"hit_rate":      hitRate,
		"cache_enabled": sa.enableCache,
	}
}

// SetCacheConfig configures caching behavior
func (sa *SessionAffinity) SetCacheConfig(enabled bool, ttl time.Duration) {
	sa.enableCache = enabled
	if ttl > 0 {
		sa.cacheTTL = ttl
	}
}

// SetSessionTTL configures session TTL
func (sa *SessionAffinity) SetSessionTTL(ttl time.Duration) {
	sa.sessionTTL = ttl
}

// getUserKey generates a Redis key for user session
func (sa *SessionAffinity) getUserKey(userID string) string {
	return fmt.Sprintf("ws:session:user:%s", userID)
}

// getConnectionKey generates a Redis key for connection
func (sa *SessionAffinity) getConnectionKey(connectionID string) string {
	return fmt.Sprintf("ws:session:conn:%s", connectionID)
}

// GetAllSessions returns all active sessions
func (sa *SessionAffinity) GetAllSessions() ([]*SessionMapping, error) {
	pattern := "ws:session:user:*"

	keys, err := sa.client.Keys(sa.ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to scan keys: %w", err)
	}

	var sessions []*SessionMapping

	for _, key := range keys {
		data, err := sa.client.Get(sa.ctx, key).Result()
		if err != nil {
			continue
		}

		var mapping SessionMapping
		if err := json.Unmarshal([]byte(data), &mapping); err != nil {
			continue
		}

		sessions = append(sessions, &mapping)
	}

	return sessions, nil
}
