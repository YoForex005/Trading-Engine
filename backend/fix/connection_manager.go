package fix

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// ConnectionManager handles automatic reconnection and health monitoring for FIX sessions
type ConnectionManager struct {
	gateway           *FIXGateway
	mu                sync.RWMutex
	reconnectConfigs  map[string]*ReconnectConfig
	healthCheckers    map[string]*HealthChecker
	stopChan          chan struct{}
	running           bool
}

// ReconnectConfig defines reconnection behavior for a session
type ReconnectConfig struct {
	SessionID          string
	Enabled            bool
	MaxRetries         int           // 0 = unlimited
	RetryCount         int
	InitialDelay       time.Duration // Start with this delay
	MaxDelay           time.Duration // Cap exponential backoff
	BackoffMultiplier  float64       // Multiply delay by this after each failure
	CurrentDelay       time.Duration
	LastAttempt        time.Time
	LastSuccess        time.Time
	LastError          error
	NextRetry          time.Time
	AutoReconnect      bool
}

// HealthChecker monitors connection health
type HealthChecker struct {
	SessionID            string
	Enabled              bool
	LastHeartbeatTimeout time.Time
	LastHealthCheck      time.Time
	HeartbeatInterval    time.Duration
	HeartbeatTimeout     time.Duration // If no heartbeat for this long, trigger reconnect
	ConsecutiveFailures  int
	HealthStatus         string // HEALTHY, DEGRADED, UNHEALTHY
}

// ConnectionStats provides detailed connection statistics
type ConnectionStats struct {
	SessionID           string    `json:"session_id"`
	Status              string    `json:"status"`
	ConnectedSince      time.Time `json:"connected_since,omitempty"`
	DisconnectedSince   time.Time `json:"disconnected_since,omitempty"`
	TotalReconnects     int       `json:"total_reconnects"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
	LastError           string    `json:"last_error,omitempty"`
	LastSuccess         time.Time `json:"last_success,omitempty"`
	NextRetry           time.Time `json:"next_retry,omitempty"`
	HealthStatus        string    `json:"health_status"`
	LastHeartbeat       time.Time `json:"last_heartbeat,omitempty"`
	AutoReconnectActive bool      `json:"auto_reconnect_active"`
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(gateway *FIXGateway) *ConnectionManager {
	return &ConnectionManager{
		gateway:          gateway,
		reconnectConfigs: make(map[string]*ReconnectConfig),
		healthCheckers:   make(map[string]*HealthChecker),
		stopChan:         make(chan struct{}),
		running:          false,
	}
}

// EnableAutoReconnect enables automatic reconnection for a session
func (cm *ConnectionManager) EnableAutoReconnect(sessionID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.reconnectConfigs[sessionID]; !exists {
		cm.reconnectConfigs[sessionID] = &ReconnectConfig{
			SessionID:         sessionID,
			Enabled:           true,
			MaxRetries:        0, // Unlimited
			InitialDelay:      5 * time.Second,
			MaxDelay:          5 * time.Minute,
			BackoffMultiplier: 2.0,
			CurrentDelay:      5 * time.Second,
			AutoReconnect:     true,
		}
	} else {
		cm.reconnectConfigs[sessionID].Enabled = true
		cm.reconnectConfigs[sessionID].AutoReconnect = true
	}

	// Also enable health checking
	if _, exists := cm.healthCheckers[sessionID]; !exists {
		cm.healthCheckers[sessionID] = &HealthChecker{
			SessionID:         sessionID,
			Enabled:           true,
			HeartbeatInterval: 30 * time.Second,
			HeartbeatTimeout:  90 * time.Second, // 3x heartbeat interval
			HealthStatus:      "UNKNOWN",
		}
	} else {
		cm.healthCheckers[sessionID].Enabled = true
	}

	log.Printf("[ConnectionManager] Auto-reconnect enabled for %s (initial delay: %v, max delay: %v)",
		sessionID, cm.reconnectConfigs[sessionID].InitialDelay, cm.reconnectConfigs[sessionID].MaxDelay)
}

// DisableAutoReconnect disables automatic reconnection for a session
func (cm *ConnectionManager) DisableAutoReconnect(sessionID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if config, exists := cm.reconnectConfigs[sessionID]; exists {
		config.Enabled = false
		config.AutoReconnect = false
	}

	if checker, exists := cm.healthCheckers[sessionID]; exists {
		checker.Enabled = false
	}

	log.Printf("[ConnectionManager] Auto-reconnect disabled for %s", sessionID)
}

// Start begins the connection management background tasks
func (cm *ConnectionManager) Start() {
	cm.mu.Lock()
	if cm.running {
		cm.mu.Unlock()
		return
	}
	cm.running = true
	cm.mu.Unlock()

	log.Println("[ConnectionManager] Starting connection manager...")

	// Start reconnection monitor
	go cm.reconnectMonitor()

	// Start health checker
	go cm.healthMonitor()

	log.Println("[ConnectionManager] Connection manager started")
}

// Stop halts the connection manager
func (cm *ConnectionManager) Stop() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.running {
		return
	}

	cm.running = false
	close(cm.stopChan)
	log.Println("[ConnectionManager] Connection manager stopped")
}

// reconnectMonitor periodically checks for disconnected sessions and attempts reconnection
func (cm *ConnectionManager) reconnectMonitor() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cm.stopChan:
			return
		case <-ticker.C:
			cm.checkAndReconnect()
		}
	}
}

// checkAndReconnect checks all sessions and reconnects if needed
func (cm *ConnectionManager) checkAndReconnect() {
	cm.mu.RLock()
	configs := make([]*ReconnectConfig, 0, len(cm.reconnectConfigs))
	for _, config := range cm.reconnectConfigs {
		configs = append(configs, config)
	}
	cm.mu.RUnlock()

	for _, config := range configs {
		if !config.Enabled || !config.AutoReconnect {
			continue
		}

		// Check if session is disconnected
		status := cm.gateway.GetStatus()
		sessionStatus := status[config.SessionID]

		if sessionStatus == "LOGGED_IN" || sessionStatus == "CONNECTING" {
			// Reset retry count on successful connection
			if config.RetryCount > 0 {
				cm.mu.Lock()
				config.RetryCount = 0
				config.CurrentDelay = config.InitialDelay
				config.LastSuccess = time.Now()
				cm.mu.Unlock()
			}
			continue
		}

		// Session is disconnected, check if we should retry
		if time.Now().Before(config.NextRetry) {
			continue
		}

		// Check max retries
		if config.MaxRetries > 0 && config.RetryCount >= config.MaxRetries {
			log.Printf("[ConnectionManager] Max retries (%d) reached for %s. Auto-reconnect disabled.",
				config.MaxRetries, config.SessionID)
			cm.mu.Lock()
			config.Enabled = false
			cm.mu.Unlock()
			continue
		}

		// Attempt reconnection
		cm.attemptReconnect(config)
	}
}

// attemptReconnect attempts to reconnect a session
func (cm *ConnectionManager) attemptReconnect(config *ReconnectConfig) {
	cm.mu.Lock()
	config.RetryCount++
	config.LastAttempt = time.Now()
	cm.mu.Unlock()

	log.Printf("[ConnectionManager] Attempting reconnect for %s (attempt %d, delay: %v)",
		config.SessionID, config.RetryCount, config.CurrentDelay)

	err := cm.gateway.Connect(config.SessionID)

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err != nil {
		config.LastError = err
		// Exponential backoff
		config.CurrentDelay = time.Duration(float64(config.CurrentDelay) * config.BackoffMultiplier)
		if config.CurrentDelay > config.MaxDelay {
			config.CurrentDelay = config.MaxDelay
		}
		config.NextRetry = time.Now().Add(config.CurrentDelay)

		log.Printf("[ConnectionManager] Reconnect failed for %s: %v (next retry in %v)",
			config.SessionID, err, config.CurrentDelay)
	} else {
		config.LastSuccess = time.Now()
		config.CurrentDelay = config.InitialDelay
		config.RetryCount = 0
		log.Printf("[ConnectionManager] Reconnect successful for %s", config.SessionID)
	}
}

// healthMonitor monitors connection health via heartbeat tracking
func (cm *ConnectionManager) healthMonitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cm.stopChan:
			return
		case <-ticker.C:
			cm.checkHealth()
		}
	}
}

// checkHealth checks health of all monitored sessions
func (cm *ConnectionManager) checkHealth() {
	cm.mu.RLock()
	checkers := make([]*HealthChecker, 0, len(cm.healthCheckers))
	for _, checker := range cm.healthCheckers {
		checkers = append(checkers, checker)
	}
	cm.mu.RUnlock()

	for _, checker := range checkers {
		if !checker.Enabled {
			continue
		}

		// Get session status
		status := cm.gateway.GetStatus()
		sessionStatus := status[checker.SessionID]

		if sessionStatus != "LOGGED_IN" {
			cm.mu.Lock()
			checker.HealthStatus = "DISCONNECTED"
			checker.LastHealthCheck = time.Now()
			cm.mu.Unlock()
			continue
		}

		// Check last heartbeat time
		session := cm.gateway.GetSession(checker.SessionID)
		if session == nil {
			continue
		}

		timeSinceHeartbeat := time.Since(session.LastHeartbeat)

		cm.mu.Lock()
		checker.LastHealthCheck = time.Now()

		if timeSinceHeartbeat > checker.HeartbeatTimeout {
			// Heartbeat timeout detected
			checker.ConsecutiveFailures++
			checker.LastHeartbeatTimeout = time.Now()
			checker.HealthStatus = "UNHEALTHY"

			log.Printf("[ConnectionManager] Heartbeat timeout for %s (last heartbeat: %v ago, failures: %d)",
				checker.SessionID, timeSinceHeartbeat, checker.ConsecutiveFailures)

			// Trigger reconnection if configured
			if config, exists := cm.reconnectConfigs[checker.SessionID]; exists && config.Enabled {
				cm.mu.Unlock()
				log.Printf("[ConnectionManager] Triggering reconnect due to heartbeat timeout for %s", checker.SessionID)
				// Disconnect first, then let reconnectMonitor handle it
				cm.gateway.Disconnect(checker.SessionID)
				continue
			}
		} else if timeSinceHeartbeat > checker.HeartbeatInterval*2 {
			checker.HealthStatus = "DEGRADED"
		} else {
			checker.HealthStatus = "HEALTHY"
			checker.ConsecutiveFailures = 0
		}
		cm.mu.Unlock()
	}
}

// GetConnectionStats returns detailed connection statistics
func (cm *ConnectionManager) GetConnectionStats(sessionID string) *ConnectionStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	status := cm.gateway.GetStatus()
	session := cm.gateway.GetSession(sessionID)

	stats := &ConnectionStats{
		SessionID: sessionID,
		Status:    status[sessionID],
	}

	if session != nil {
		stats.LastHeartbeat = session.LastHeartbeat
	}

	if config, exists := cm.reconnectConfigs[sessionID]; exists {
		stats.TotalReconnects = config.RetryCount
		stats.ConsecutiveFailures = config.RetryCount
		stats.LastSuccess = config.LastSuccess
		stats.NextRetry = config.NextRetry
		stats.AutoReconnectActive = config.Enabled && config.AutoReconnect

		if config.LastError != nil {
			stats.LastError = config.LastError.Error()
		}
	}

	if checker, exists := cm.healthCheckers[sessionID]; exists {
		stats.HealthStatus = checker.HealthStatus
	}

	return stats
}

// GetAllConnectionStats returns stats for all sessions
func (cm *ConnectionManager) GetAllConnectionStats() map[string]*ConnectionStats {
	cm.mu.RLock()
	sessionIDs := make([]string, 0, len(cm.reconnectConfigs))
	for id := range cm.reconnectConfigs {
		sessionIDs = append(sessionIDs, id)
	}
	cm.mu.RUnlock()

	result := make(map[string]*ConnectionStats)
	for _, id := range sessionIDs {
		result[id] = cm.GetConnectionStats(id)
	}
	return result
}

// ForceReconnect manually triggers a reconnection
func (cm *ConnectionManager) ForceReconnect(sessionID string) error {
	log.Printf("[ConnectionManager] Manual reconnect triggered for %s", sessionID)

	// Disconnect first
	if err := cm.gateway.Disconnect(sessionID); err != nil {
		log.Printf("[ConnectionManager] Error during disconnect: %v", err)
	}

	// Wait a moment
	time.Sleep(1 * time.Second)

	// Attempt reconnect
	err := cm.gateway.Connect(sessionID)
	if err != nil {
		return fmt.Errorf("reconnect failed: %v", err)
	}

	// Reset retry counters
	cm.mu.Lock()
	if config, exists := cm.reconnectConfigs[sessionID]; exists {
		config.RetryCount = 0
		config.CurrentDelay = config.InitialDelay
		config.LastSuccess = time.Now()
	}
	cm.mu.Unlock()

	return nil
}
