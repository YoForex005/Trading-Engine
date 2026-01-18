package datapipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// QuoteDistributor handles distribution of quotes via Redis pub/sub and WebSocket
type QuoteDistributor struct {
	redis          *redis.Client
	config         *PipelineConfig
	stats          *PipelineStats

	// Subscription management
	mu             sync.RWMutex
	subscriptions  map[string]map[string]bool // clientID -> symbol -> subscribed

	// Throttling (send updates only if price changed significantly)
	lastSentPrices map[string]float64
	throttleMu     sync.RWMutex

	// Rate limiting per client
	clientRateLimits map[string]*RateLimiter

	// Channels
	tickQueue      chan *NormalizedTick
	ohlcQueue      chan *OHLCBar

	// Context
	ctx            context.Context

	// Connected clients counter
	clientCount    int32
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	mu            sync.Mutex
	tokens        float64
	maxTokens     float64
	refillRate    float64 // tokens per second
	lastRefill    time.Time
}

// NewQuoteDistributor creates a new quote distributor
func NewQuoteDistributor(redis *redis.Client, config *PipelineConfig, stats *PipelineStats) *QuoteDistributor {
	return &QuoteDistributor{
		redis:            redis,
		config:           config,
		stats:            stats,
		subscriptions:    make(map[string]map[string]bool),
		lastSentPrices:   make(map[string]float64),
		clientRateLimits: make(map[string]*RateLimiter),
		tickQueue:        make(chan *NormalizedTick, config.DistributionBufferSize),
		ohlcQueue:        make(chan *OHLCBar, config.OHLCBufferSize),
	}
}

// Start starts the distributor
func (d *QuoteDistributor) Start(ctx context.Context) error {
	d.ctx = ctx

	// Start workers for distribution
	for i := 0; i < d.config.WorkerCount; i++ {
		go d.tickDistributionWorker()
		go d.ohlcDistributionWorker()
	}

	log.Println("[Distributor] Started quote distribution workers")
	return nil
}

// DistributeTick queues a tick for distribution
func (d *QuoteDistributor) DistributeTick(tick *NormalizedTick) {
	select {
	case d.tickQueue <- tick:
	default:
		// Buffer full, drop tick
		d.stats.mu.Lock()
		d.stats.TicksDropped++
		d.stats.mu.Unlock()
	}
}

// DistributeOHLC queues an OHLC bar for distribution
func (d *QuoteDistributor) DistributeOHLC(bar *OHLCBar) {
	select {
	case d.ohlcQueue <- bar:
	default:
		// Buffer full, drop bar
	}
}

// tickDistributionWorker processes ticks for distribution
func (d *QuoteDistributor) tickDistributionWorker() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case tick := <-d.tickQueue:
			startTime := time.Now()

			// Check if price changed significantly (throttling)
			if d.shouldThrottle(tick) {
				continue
			}

			// Publish to Redis pub/sub
			if err := d.publishToRedis("quotes", tick); err != nil {
				log.Printf("[Distributor] Failed to publish to Redis: %v", err)
			}

			// Publish symbol-specific channel
			if err := d.publishToRedis(fmt.Sprintf("quotes:%s", tick.Symbol), tick); err != nil {
				log.Printf("[Distributor] Failed to publish symbol quote: %v", err)
			}

			// Update stats
			d.stats.mu.Lock()
			d.stats.QuotesDistributed++
			latency := time.Since(startTime).Milliseconds()
			if d.stats.AvgDistributionLatencyMs == 0 {
				d.stats.AvgDistributionLatencyMs = float64(latency)
			} else {
				d.stats.AvgDistributionLatencyMs = (d.stats.AvgDistributionLatencyMs * 0.9) + (float64(latency) * 0.1)
			}
			d.stats.mu.Unlock()

			// Update last sent price
			d.throttleMu.Lock()
			d.lastSentPrices[tick.Symbol] = tick.Bid
			d.throttleMu.Unlock()
		}
	}
}

// ohlcDistributionWorker processes OHLC bars for distribution
func (d *QuoteDistributor) ohlcDistributionWorker() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case bar := <-d.ohlcQueue:
			// Publish to Redis pub/sub
			if err := d.publishToRedis("ohlc", bar); err != nil {
				log.Printf("[Distributor] Failed to publish OHLC: %v", err)
			}

			// Publish symbol+timeframe specific channel
			channel := fmt.Sprintf("ohlc:%s:%d", bar.Symbol, bar.Timeframe)
			if err := d.publishToRedis(channel, bar); err != nil {
				log.Printf("[Distributor] Failed to publish OHLC to %s: %v", channel, err)
			}
		}
	}
}

// publishToRedis publishes data to Redis pub/sub
func (d *QuoteDistributor) publishToRedis(channel string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return d.redis.Publish(d.ctx, channel, jsonData).Err()
}

// shouldThrottle checks if we should throttle this tick
func (d *QuoteDistributor) shouldThrottle(tick *NormalizedTick) bool {
	d.throttleMu.RLock()
	lastPrice, exists := d.lastSentPrices[tick.Symbol]
	d.throttleMu.RUnlock()

	if !exists {
		return false // First tick, don't throttle
	}

	// Calculate price change %
	change := ((tick.Bid - lastPrice) / lastPrice)
	if change < 0 {
		change = -change
	}

	// Only send if changed by more than 0.001% (0.01 pips for most pairs)
	if change < 0.00001 {
		return true // Throttle
	}

	return false
}

// Subscribe subscribes a client to a symbol
func (d *QuoteDistributor) Subscribe(clientID, symbol string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.subscriptions[clientID]; !exists {
		d.subscriptions[clientID] = make(map[string]bool)
		atomic.AddInt32(&d.clientCount, 1)

		// Create rate limiter for this client (100 msgs/sec)
		d.clientRateLimits[clientID] = NewRateLimiter(100, 100)
	}

	d.subscriptions[clientID][symbol] = true
	log.Printf("[Distributor] Client %s subscribed to %s", clientID, symbol)

	return nil
}

// Unsubscribe unsubscribes a client from a symbol
func (d *QuoteDistributor) Unsubscribe(clientID, symbol string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if subs, exists := d.subscriptions[clientID]; exists {
		delete(subs, symbol)
		log.Printf("[Distributor] Client %s unsubscribed from %s", clientID, symbol)
	}

	return nil
}

// DisconnectClient removes all subscriptions for a client
func (d *QuoteDistributor) DisconnectClient(clientID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.subscriptions[clientID]; exists {
		delete(d.subscriptions, clientID)
		delete(d.clientRateLimits, clientID)
		atomic.AddInt32(&d.clientCount, -1)
		log.Printf("[Distributor] Client %s disconnected", clientID)
	}
}

// IsSubscribed checks if a client is subscribed to a symbol
func (d *QuoteDistributor) IsSubscribed(clientID, symbol string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if subs, exists := d.subscriptions[clientID]; exists {
		return subs[symbol]
	}

	return false
}

// CheckRateLimit checks if client can send message
func (d *QuoteDistributor) CheckRateLimit(clientID string) bool {
	d.mu.RLock()
	limiter, exists := d.clientRateLimits[clientID]
	d.mu.RUnlock()

	if !exists {
		return false
	}

	return limiter.Allow()
}

// GetClientCount returns the number of connected clients
func (d *QuoteDistributor) GetClientCount() int32 {
	return atomic.LoadInt32(&d.clientCount)
}

// HealthCheck performs health check
func (d *QuoteDistributor) HealthCheck() ComponentHealth {
	clientCount := atomic.LoadInt32(&d.clientCount)

	// Check Redis connection
	if err := d.redis.Ping(d.ctx).Err(); err != nil {
		return ComponentHealth{
			Status:     "unhealthy",
			LastError:  fmt.Sprintf("Redis ping failed: %v", err),
			Metrics: map[string]interface{}{
				"clients_connected": clientCount,
			},
		}
	}

	return ComponentHealth{
		Status: "healthy",
		Metrics: map[string]interface{}{
			"clients_connected": clientCount,
			"subscriptions":     len(d.subscriptions),
		},
	}
}

// NewRateLimiter creates a new token bucket rate limiter
func NewRateLimiter(maxTokens, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if an action is allowed (consumes 1 token)
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(r.lastRefill).Seconds()
	r.tokens += elapsed * r.refillRate

	if r.tokens > r.maxTokens {
		r.tokens = r.maxTokens
	}

	r.lastRefill = now

	// Check if we have tokens
	if r.tokens >= 1.0 {
		r.tokens -= 1.0
		return true
	}

	return false
}
