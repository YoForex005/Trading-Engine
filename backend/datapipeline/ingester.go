package datapipeline

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"sync"
	"time"
)

// RawTick represents an unnormalized tick from various sources
type RawTick struct {
	Source       string      // LP name (OANDA, Binance, etc.)
	Symbol       string
	Bid          float64
	Ask          float64
	Timestamp    interface{} // Can be Unix int64, ISO8601 string, etc.
	RawData      interface{} // Original data for debugging
}

// NormalizedTick is the standardized tick format used internally
type NormalizedTick struct {
	Symbol       string    `json:"symbol"`
	Bid          float64   `json:"bid"`
	Ask          float64   `json:"ask"`
	Spread       float64   `json:"spread"`
	Timestamp    time.Time `json:"timestamp"`
	Source       string    `json:"source"`
	TickID       string    `json:"tick_id"` // For deduplication
	ReceivedAt   time.Time `json:"received_at"`
}

// DataIngester handles tick ingestion, normalization, and validation
type DataIngester struct {
	config           *PipelineConfig
	stats            *PipelineStats

	// Channels
	rawTicksChan     chan *RawTick
	normalizedTicks  chan *NormalizedTick

	// Deduplication
	mu               sync.RWMutex
	seenTicks        map[string]time.Time // tickID -> timestamp
	lastTickPerSymbol map[string]*NormalizedTick

	// Context
	ctx              context.Context
}

// NewDataIngester creates a new data ingester
func NewDataIngester(config *PipelineConfig, stats *PipelineStats) *DataIngester {
	return &DataIngester{
		config:           config,
		stats:            stats,
		rawTicksChan:     make(chan *RawTick, config.TickBufferSize),
		normalizedTicks:  make(chan *NormalizedTick, config.TickBufferSize),
		seenTicks:        make(map[string]time.Time),
		lastTickPerSymbol: make(map[string]*NormalizedTick),
	}
}

// Start starts the ingester workers
func (d *DataIngester) Start(ctx context.Context) error {
	d.ctx = ctx

	// Start worker pool for tick processing
	for i := 0; i < d.config.WorkerCount; i++ {
		go d.processWorker()
	}

	// Start cleanup goroutine for deduplication cache
	go d.cleanupDeduplicationCache()

	log.Printf("[Ingester] Started with %d workers", d.config.WorkerCount)
	return nil
}

// IngestTick queues a raw tick for processing
func (d *DataIngester) IngestTick(rawTick *RawTick) error {
	select {
	case d.rawTicksChan <- rawTick:
		d.stats.mu.Lock()
		d.stats.TicksReceived++
		d.stats.mu.Unlock()
		return nil
	default:
		d.stats.mu.Lock()
		d.stats.TicksDropped++
		d.stats.mu.Unlock()
		return fmt.Errorf("tick buffer full, dropping tick")
	}
}

// GetNormalizedTicks returns the normalized ticks channel
func (d *DataIngester) GetNormalizedTicks() <-chan *NormalizedTick {
	return d.normalizedTicks
}

// processWorker processes raw ticks from the queue
func (d *DataIngester) processWorker() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case rawTick := <-d.rawTicksChan:
			startTime := time.Now()

			// Normalize the tick
			normalizedTick, err := d.normalizeTick(rawTick)
			if err != nil {
				log.Printf("[Ingester] Failed to normalize tick from %s: %v", rawTick.Source, err)
				d.stats.mu.Lock()
				d.stats.TicksInvalid++
				d.stats.mu.Unlock()
				continue
			}

			// Validate the tick
			if !d.validateTick(normalizedTick) {
				d.stats.mu.Lock()
				d.stats.TicksInvalid++
				d.stats.mu.Unlock()
				continue
			}

			// Check for duplicates
			if d.config.EnableDeduplication && d.isDuplicate(normalizedTick) {
				d.stats.mu.Lock()
				d.stats.TicksDuplicate++
				d.stats.mu.Unlock()
				continue
			}

			// Check for out-of-order ticks
			if d.config.EnableOutOfOrderCheck && d.isOutOfOrder(normalizedTick) {
				d.stats.mu.Lock()
				d.stats.TicksOutOfOrder++
				d.stats.mu.Unlock()
				// Still process it, but log it
				log.Printf("[Ingester] Out-of-order tick detected for %s", normalizedTick.Symbol)
			}

			// Check price sanity
			if !d.checkPriceSanity(normalizedTick) {
				log.Printf("[Ingester] Abnormal price spike detected for %s: %.5f -> %.5f",
					normalizedTick.Symbol, d.getLastPrice(normalizedTick.Symbol), normalizedTick.Bid)
				d.stats.mu.Lock()
				d.stats.AbnormalSpikesDetected++
				d.stats.mu.Unlock()
				// Still process it (could be legitimate)
			}

			// Update last tick
			d.updateLastTick(normalizedTick)

			// Send to normalized channel
			select {
			case d.normalizedTicks <- normalizedTick:
				d.stats.mu.Lock()
				d.stats.TicksProcessed++
				d.stats.LastTickTime = time.Now()

				// Update latency
				latency := time.Since(startTime).Milliseconds()
				if d.stats.AvgTickLatencyMs == 0 {
					d.stats.AvgTickLatencyMs = float64(latency)
				} else {
					d.stats.AvgTickLatencyMs = (d.stats.AvgTickLatencyMs * 0.9) + (float64(latency) * 0.1)
				}
				d.stats.mu.Unlock()
			default:
				d.stats.mu.Lock()
				d.stats.TicksDropped++
				d.stats.mu.Unlock()
			}
		}
	}
}

// normalizeTick converts a raw tick to normalized format
func (d *DataIngester) normalizeTick(raw *RawTick) (*NormalizedTick, error) {
	// Parse timestamp
	timestamp, err := d.parseTimestamp(raw.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}

	// Validate prices
	if raw.Bid <= 0 || raw.Ask <= 0 {
		return nil, fmt.Errorf("invalid prices: bid=%f, ask=%f", raw.Bid, raw.Ask)
	}

	if raw.Bid > raw.Ask {
		return nil, fmt.Errorf("bid > ask: bid=%f, ask=%f", raw.Bid, raw.Ask)
	}

	// Calculate spread
	spread := raw.Ask - raw.Bid

	// Generate tick ID for deduplication
	tickID := d.generateTickID(raw.Source, raw.Symbol, timestamp, raw.Bid, raw.Ask)

	normalized := &NormalizedTick{
		Symbol:     raw.Symbol,
		Bid:        raw.Bid,
		Ask:        raw.Ask,
		Spread:     spread,
		Timestamp:  timestamp,
		Source:     raw.Source,
		TickID:     tickID,
		ReceivedAt: time.Now(),
	}

	return normalized, nil
}

// parseTimestamp handles various timestamp formats
func (d *DataIngester) parseTimestamp(ts interface{}) (time.Time, error) {
	switch v := ts.(type) {
	case int64:
		// Unix timestamp in seconds or milliseconds
		if v > 1e12 {
			// Milliseconds
			return time.Unix(0, v*int64(time.Millisecond)), nil
		}
		// Seconds
		return time.Unix(v, 0), nil
	case float64:
		// Unix timestamp (seconds)
		return time.Unix(int64(v), 0), nil
	case string:
		// Try RFC3339 / ISO8601
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			return t, nil
		}
		// Try other common formats
		formats := []string{
			"2006-01-02T15:04:05.000Z",
			"2006-01-02 15:04:05",
			time.RFC3339Nano,
		}
		for _, format := range formats {
			t, err := time.Parse(format, v)
			if err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("unparseable timestamp string: %s", v)
	case time.Time:
		return v, nil
	default:
		return time.Time{}, fmt.Errorf("unsupported timestamp type: %T", ts)
	}
}

// generateTickID creates a unique ID for deduplication
func (d *DataIngester) generateTickID(source, symbol string, timestamp time.Time, bid, ask float64) string {
	h := sha256.New()
	h.Write([]byte(source))
	h.Write([]byte(symbol))

	tsBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(tsBytes, uint64(timestamp.Unix()))
	h.Write(tsBytes)

	bidBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bidBytes, uint64(bid*100000))
	h.Write(bidBytes)

	askBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(askBytes, uint64(ask*100000))
	h.Write(askBytes)

	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

// validateTick performs basic validation
func (d *DataIngester) validateTick(tick *NormalizedTick) bool {
	// Check age
	age := time.Since(tick.Timestamp)
	if age > time.Duration(d.config.MaxTickAgeSeconds)*time.Second {
		log.Printf("[Ingester] Tick too old: %s, age=%v", tick.Symbol, age)
		return false
	}

	// Check spread
	if tick.Spread < 0 {
		return false
	}

	// Future timestamps
	if tick.Timestamp.After(time.Now().Add(1 * time.Minute)) {
		log.Printf("[Ingester] Future timestamp detected: %s", tick.Symbol)
		return false
	}

	return true
}

// isDuplicate checks if we've seen this exact tick before
func (d *DataIngester) isDuplicate(tick *NormalizedTick) bool {
	d.mu.RLock()
	lastSeen, exists := d.seenTicks[tick.TickID]
	d.mu.RUnlock()

	if exists {
		// Consider duplicate if seen within last 5 seconds
		if time.Since(lastSeen) < 5*time.Second {
			return true
		}
	}

	// Mark as seen
	d.mu.Lock()
	d.seenTicks[tick.TickID] = time.Now()
	d.mu.Unlock()

	return false
}

// isOutOfOrder checks if tick is older than the last tick for this symbol
func (d *DataIngester) isOutOfOrder(tick *NormalizedTick) bool {
	d.mu.RLock()
	lastTick, exists := d.lastTickPerSymbol[tick.Symbol]
	d.mu.RUnlock()

	if !exists {
		return false
	}

	return tick.Timestamp.Before(lastTick.Timestamp)
}

// checkPriceSanity detects abnormal price spikes
func (d *DataIngester) checkPriceSanity(tick *NormalizedTick) bool {
	lastPrice := d.getLastPrice(tick.Symbol)
	if lastPrice == 0 {
		return true // First tick for symbol
	}

	// Calculate % change
	change := ((tick.Bid - lastPrice) / lastPrice)
	if change < 0 {
		change = -change
	}

	if change > d.config.PriceSanityThreshold {
		return false
	}

	return true
}

// getLastPrice returns the last bid price for a symbol
func (d *DataIngester) getLastPrice(symbol string) float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if lastTick, exists := d.lastTickPerSymbol[symbol]; exists {
		return lastTick.Bid
	}
	return 0
}

// updateLastTick updates the last tick for a symbol
func (d *DataIngester) updateLastTick(tick *NormalizedTick) {
	d.mu.Lock()
	d.lastTickPerSymbol[tick.Symbol] = tick
	d.mu.Unlock()
}

// cleanupDeduplicationCache periodically cleans old entries
func (d *DataIngester) cleanupDeduplicationCache() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.mu.Lock()
			now := time.Now()
			for tickID, seenTime := range d.seenTicks {
				if now.Sub(seenTime) > 5*time.Minute {
					delete(d.seenTicks, tickID)
				}
			}
			d.mu.Unlock()
		}
	}
}
