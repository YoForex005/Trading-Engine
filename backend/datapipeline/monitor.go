package datapipeline

import (
	"context"
	"log"
	"sync"
	"time"
)

// DataMonitor monitors data quality and pipeline health
type DataMonitor struct {
	config         *PipelineConfig
	stats          *PipelineStats

	// Feed health tracking
	mu             sync.RWMutex
	lastTickTime   map[string]time.Time // symbol -> last tick time
	feedHealth     map[string]FeedHealth

	// Alerts
	alerts         []Alert

	ctx            context.Context
}

// FeedHealth represents the health status of a data feed
type FeedHealth struct {
	Symbol         string
	LastTickTime   time.Time
	TickCount      int64
	IsStale        bool
	StaleSeconds   float64
	Status         string // "healthy", "stale", "dead"
}

// Alert represents a data quality alert
type Alert struct {
	Timestamp      time.Time
	Level          string // "info", "warning", "critical"
	Type           string // "stale_feed", "abnormal_spike", "missing_data"
	Symbol         string
	Message        string
	Details        map[string]interface{}
}

// NewDataMonitor creates a new data monitor
func NewDataMonitor(config *PipelineConfig, stats *PipelineStats) *DataMonitor {
	return &DataMonitor{
		config:       config,
		stats:        stats,
		lastTickTime: make(map[string]time.Time),
		feedHealth:   make(map[string]FeedHealth),
		alerts:       make([]Alert, 0, 1000),
	}
}

// Start starts the data monitor
func (m *DataMonitor) Start(ctx context.Context) error {
	m.ctx = ctx

	if m.config.EnableHealthChecks {
		go m.healthCheckWorker()
	}

	log.Println("[Monitor] Data quality monitoring started")
	return nil
}

// UpdateTickTime updates the last tick time for a symbol
func (m *DataMonitor) UpdateTickTime(symbol string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastTickTime[symbol] = time.Now()

	// Update feed health
	if health, exists := m.feedHealth[symbol]; exists {
		health.TickCount++
		health.LastTickTime = time.Now()
		health.IsStale = false
		health.Status = "healthy"
		m.feedHealth[symbol] = health
	} else {
		m.feedHealth[symbol] = FeedHealth{
			Symbol:       symbol,
			LastTickTime: time.Now(),
			TickCount:    1,
			IsStale:      false,
			Status:       "healthy",
		}
	}
}

// healthCheckWorker periodically checks feed health
func (m *DataMonitor) healthCheckWorker() {
	ticker := time.NewTicker(m.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkFeedHealth()
		}
	}
}

// checkFeedHealth checks all feeds for staleness
func (m *DataMonitor) checkFeedHealth() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	for symbol, lastTime := range m.lastTickTime {
		elapsed := now.Sub(lastTime)

		health := m.feedHealth[symbol]
		health.StaleSeconds = elapsed.Seconds()

		if elapsed > m.config.StaleQuoteThreshold {
			if !health.IsStale {
				// Feed just went stale
				health.IsStale = true
				health.Status = "stale"

				// Create alert
				alert := Alert{
					Timestamp: now,
					Level:     "warning",
					Type:      "stale_feed",
					Symbol:    symbol,
					Message:   "Feed has not received data for extended period",
					Details: map[string]interface{}{
						"last_tick":     lastTime,
						"stale_seconds": elapsed.Seconds(),
					},
				}
				m.addAlert(alert)

				log.Printf("[Monitor] STALE FEED: %s has not received data for %.1f seconds",
					symbol, elapsed.Seconds())

				m.stats.mu.Lock()
				m.stats.StaleFeedsDetected++
				m.stats.mu.Unlock()
			}

			// Mark as dead if too long
			if elapsed > 5*m.config.StaleQuoteThreshold {
				health.Status = "dead"
			}
		} else {
			if health.IsStale {
				// Feed recovered
				health.IsStale = false
				health.Status = "healthy"

				alert := Alert{
					Timestamp: now,
					Level:     "info",
					Type:      "feed_recovered",
					Symbol:    symbol,
					Message:   "Feed has recovered",
				}
				m.addAlert(alert)

				log.Printf("[Monitor] RECOVERED: %s feed is healthy again", symbol)
			}
		}

		m.feedHealth[symbol] = health
	}
}

// addAlert adds an alert to the queue
func (m *DataMonitor) addAlert(alert Alert) {
	m.alerts = append(m.alerts, alert)

	// Keep only last 1000 alerts
	if len(m.alerts) > 1000 {
		m.alerts = m.alerts[len(m.alerts)-1000:]
	}
}

// GetFeedHealth returns health status for all feeds
func (m *DataMonitor) GetFeedHealth() map[string]FeedHealth {
	m.mu.RLock()
	defer m.mu.RUnlock()

	healthCopy := make(map[string]FeedHealth)
	for symbol, health := range m.feedHealth {
		healthCopy[symbol] = health
	}

	return healthCopy
}

// GetAlerts returns recent alerts
func (m *DataMonitor) GetAlerts(limit int) []Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.alerts) {
		limit = len(m.alerts)
	}

	alertsCopy := make([]Alert, limit)
	copy(alertsCopy, m.alerts[len(m.alerts)-limit:])

	return alertsCopy
}

// GetHealthSummary returns a summary of overall health
func (m *DataMonitor) GetHealthSummary() HealthSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := HealthSummary{
		Timestamp:    time.Now(),
		TotalFeeds:   len(m.feedHealth),
		HealthyFeeds: 0,
		StaleFeeds:   0,
		DeadFeeds:    0,
	}

	for _, health := range m.feedHealth {
		switch health.Status {
		case "healthy":
			summary.HealthyFeeds++
		case "stale":
			summary.StaleFeeds++
		case "dead":
			summary.DeadFeeds++
		}
	}

	// Overall status
	if summary.DeadFeeds > 0 {
		summary.OverallStatus = "critical"
	} else if summary.StaleFeeds > 0 {
		summary.OverallStatus = "degraded"
	} else {
		summary.OverallStatus = "healthy"
	}

	return summary
}

// DetectAbnormalSpike checks for abnormal price movements
func (m *DataMonitor) DetectAbnormalSpike(symbol string, oldPrice, newPrice float64) bool {
	change := ((newPrice - oldPrice) / oldPrice)
	if change < 0 {
		change = -change
	}

	if change > m.config.PriceSanityThreshold {
		alert := Alert{
			Timestamp: time.Now(),
			Level:     "warning",
			Type:      "abnormal_spike",
			Symbol:    symbol,
			Message:   "Abnormal price spike detected",
			Details: map[string]interface{}{
				"old_price":      oldPrice,
				"new_price":      newPrice,
				"change_percent": change * 100,
			},
		}

		m.mu.Lock()
		m.addAlert(alert)
		m.mu.Unlock()

		return true
	}

	return false
}

// HealthSummary represents overall pipeline health
type HealthSummary struct {
	Timestamp     time.Time `json:"timestamp"`
	OverallStatus string    `json:"overall_status"`
	TotalFeeds    int       `json:"total_feeds"`
	HealthyFeeds  int       `json:"healthy_feeds"`
	StaleFeeds    int       `json:"stale_feeds"`
	DeadFeeds     int       `json:"dead_feeds"`
}
