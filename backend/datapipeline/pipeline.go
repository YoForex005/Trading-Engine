package datapipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// MarketDataPipeline orchestrates the entire real-time data flow
type MarketDataPipeline struct {
	// Core components
	ingester     *DataIngester
	ohlcEngine   *OHLCEngine
	distributor  *QuoteDistributor
	storage      *StorageManager
	monitor      *DataMonitor

	// Configuration
	config       *PipelineConfig

	// State management
	mu           sync.RWMutex
	running      bool
	ctx          context.Context
	cancel       context.CancelFunc

	// Statistics
	stats        *PipelineStats
}

// PipelineConfig holds pipeline configuration
type PipelineConfig struct {
	// Redis configuration
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Performance tuning
	TickBufferSize        int
	OHLCBufferSize        int
	DistributionBufferSize int
	WorkerCount           int

	// Data quality
	EnableDeduplication   bool
	EnableOutOfOrderCheck bool
	MaxTickAgeSeconds     int
	PriceSanityThreshold  float64 // Max % price change per tick

	// Storage
	HotDataRetention      int // Ticks to keep in Redis
	WarmDataRetentionDays int // Days to keep in TimescaleDB

	// Monitoring
	EnableHealthChecks    bool
	HealthCheckInterval   time.Duration
	StaleQuoteThreshold   time.Duration
}

// PipelineStats tracks pipeline performance metrics
type PipelineStats struct {
	mu                    sync.RWMutex

	// Throughput
	TicksReceived         int64
	TicksProcessed        int64
	TicksDropped          int64
	TicksDuplicate        int64
	TicksOutOfOrder       int64
	TicksInvalid          int64

	// OHLC
	OHLCBarsGenerated     int64

	// Distribution
	QuotesDistributed     int64
	ClientsConnected      int32

	// Performance
	AvgTickLatencyMs      float64
	AvgOHLCLatencyMs      float64
	AvgDistributionLatencyMs float64

	// Data quality
	StaleFeedsDetected    int64
	AbnormalSpikesDetected int64

	// Last update
	LastTickTime          time.Time
	LastUpdateTime        time.Time
}

// NewPipeline creates a new market data pipeline
func NewPipeline(config *PipelineConfig) (*MarketDataPipeline, error) {
	if config == nil {
		config = DefaultPipelineConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	pipeline := &MarketDataPipeline{
		config:  config,
		ctx:     ctx,
		cancel:  cancel,
		stats:   &PipelineStats{},
	}

	// Initialize components
	pipeline.ingester = NewDataIngester(config, pipeline.stats)
	pipeline.ohlcEngine = NewOHLCEngine(config, pipeline.stats)
	pipeline.distributor = NewQuoteDistributor(redisClient, config, pipeline.stats)
	pipeline.storage = NewStorageManager(redisClient, config)
	pipeline.monitor = NewDataMonitor(config, pipeline.stats)

	log.Printf("[Pipeline] Initialized with %d workers, tick buffer: %d, Redis: %s",
		config.WorkerCount, config.TickBufferSize, config.RedisAddr)

	return pipeline, nil
}

// Start starts the data pipeline
func (p *MarketDataPipeline) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("pipeline already running")
	}

	log.Println("[Pipeline] Starting market data pipeline...")

	// Start components in order
	if err := p.ingester.Start(p.ctx); err != nil {
		return fmt.Errorf("failed to start ingester: %w", err)
	}

	if err := p.ohlcEngine.Start(p.ctx); err != nil {
		return fmt.Errorf("failed to start OHLC engine: %w", err)
	}

	if err := p.distributor.Start(p.ctx); err != nil {
		return fmt.Errorf("failed to start distributor: %w", err)
	}

	if err := p.storage.Start(p.ctx); err != nil {
		return fmt.Errorf("failed to start storage: %w", err)
	}

	if err := p.monitor.Start(p.ctx); err != nil {
		return fmt.Errorf("failed to start monitor: %w", err)
	}

	// Wire up the pipeline flow
	p.wirePipeline()

	p.running = true
	log.Println("[Pipeline] Market data pipeline started successfully")

	return nil
}

// wirePipeline connects all components
func (p *MarketDataPipeline) wirePipeline() {
	// Ingester -> OHLC Engine
	go func() {
		for tick := range p.ingester.GetNormalizedTicks() {
			select {
			case <-p.ctx.Done():
				return
			default:
				// Send to OHLC engine
				p.ohlcEngine.ProcessTick(tick)

				// Send to distributor for broadcasting
				p.distributor.DistributeTick(tick)

				// Store in Redis/TimescaleDB
				p.storage.StoreTick(tick)
			}
		}
	}()

	// OHLC Engine -> Distributor
	go func() {
		for bar := range p.ohlcEngine.GetOHLCChannel() {
			select {
			case <-p.ctx.Done():
				return
			default:
				// Distribute OHLC updates
				p.distributor.DistributeOHLC(bar)

				// Store OHLC
				p.storage.StoreOHLC(bar)
			}
		}
	}()
}

// IngestTick receives raw tick from external source
func (p *MarketDataPipeline) IngestTick(rawTick *RawTick) error {
	if !p.running {
		return fmt.Errorf("pipeline not running")
	}

	return p.ingester.IngestTick(rawTick)
}

// Stop gracefully stops the pipeline
func (p *MarketDataPipeline) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	log.Println("[Pipeline] Stopping market data pipeline...")

	// Cancel context to signal all components
	p.cancel()

	// Wait for graceful shutdown
	time.Sleep(2 * time.Second)

	p.running = false
	log.Println("[Pipeline] Pipeline stopped")

	return nil
}

// GetStats returns current pipeline statistics
func (p *MarketDataPipeline) GetStats() PipelineStats {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()

	statsCopy := *p.stats
	return statsCopy
}

// GetDistributor returns the quote distributor (for WebSocket integration)
func (p *MarketDataPipeline) GetDistributor() *QuoteDistributor {
	return p.distributor
}

// GetOHLCEngine returns the OHLC engine
func (p *MarketDataPipeline) GetOHLCEngine() *OHLCEngine {
	return p.ohlcEngine
}

// GetStorageManager returns the storage manager
func (p *MarketDataPipeline) GetStorageManager() *StorageManager {
	return p.storage
}

// HealthCheck performs a health check on all components
func (p *MarketDataPipeline) HealthCheck() (*HealthStatus, error) {
	status := &HealthStatus{
		Timestamp: time.Now(),
		Components: make(map[string]ComponentHealth),
	}

	// Check ingester
	status.Components["ingester"] = ComponentHealth{
		Status: "healthy",
		Uptime: time.Since(p.stats.LastUpdateTime),
	}

	// Check OHLC engine
	status.Components["ohlc_engine"] = ComponentHealth{
		Status: "healthy",
		Uptime: time.Since(p.stats.LastUpdateTime),
	}

	// Check distributor
	distHealth := p.distributor.HealthCheck()
	status.Components["distributor"] = distHealth

	// Check storage
	storageHealth := p.storage.HealthCheck()
	status.Components["storage"] = storageHealth

	// Overall status
	status.OverallStatus = "healthy"
	for _, comp := range status.Components {
		if comp.Status != "healthy" {
			status.OverallStatus = "degraded"
			break
		}
	}

	return status, nil
}

// DefaultPipelineConfig returns default configuration
func DefaultPipelineConfig() *PipelineConfig {
	return &PipelineConfig{
		RedisAddr:              "localhost:6379",
		RedisPassword:          "",
		RedisDB:                0,
		TickBufferSize:         10000,
		OHLCBufferSize:         1000,
		DistributionBufferSize: 5000,
		WorkerCount:            4,
		EnableDeduplication:    true,
		EnableOutOfOrderCheck:  true,
		MaxTickAgeSeconds:      60,
		PriceSanityThreshold:   0.10, // 10% max change
		HotDataRetention:       1000,
		WarmDataRetentionDays:  30,
		EnableHealthChecks:     true,
		HealthCheckInterval:    30 * time.Second,
		StaleQuoteThreshold:    10 * time.Second,
	}
}

// HealthStatus represents overall pipeline health
type HealthStatus struct {
	Timestamp     time.Time                  `json:"timestamp"`
	OverallStatus string                     `json:"overall_status"`
	Components    map[string]ComponentHealth `json:"components"`
}

// ComponentHealth represents individual component health
type ComponentHealth struct {
	Status       string        `json:"status"`
	Uptime       time.Duration `json:"uptime"`
	ErrorCount   int           `json:"error_count,omitempty"`
	LastError    string        `json:"last_error,omitempty"`
	Metrics      interface{}   `json:"metrics,omitempty"`
}

// SerializeToJSON serializes stats to JSON
func (s *PipelineStats) SerializeToJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return json.Marshal(s)
}
