package datapipeline

import (
	"context"
	"log"
	"sync"
	"time"
)

// Timeframe represents different OHLC timeframes
type Timeframe int

const (
	TF_M1  Timeframe = 60      // 1 minute
	TF_M5  Timeframe = 300     // 5 minutes
	TF_M15 Timeframe = 900     // 15 minutes
	TF_H1  Timeframe = 3600    // 1 hour
	TF_H4  Timeframe = 14400   // 4 hours
	TF_D1  Timeframe = 86400   // 1 day
)

// OHLCBar represents a complete OHLC candle
type OHLCBar struct {
	Symbol     string    `json:"symbol"`
	Timeframe  Timeframe `json:"timeframe"`
	OpenTime   time.Time `json:"open_time"`
	CloseTime  time.Time `json:"close_time"`
	Open       float64   `json:"open"`
	High       float64   `json:"high"`
	Low        float64   `json:"low"`
	Close      float64   `json:"close"`
	Volume     int64     `json:"volume"`
	TickCount  int64     `json:"tick_count"`
	IsClosed   bool      `json:"is_closed"`
}

// OHLCEngine handles real-time OHLC aggregation from ticks
type OHLCEngine struct {
	config         *PipelineConfig
	stats          *PipelineStats

	// Active bars (current incomplete bars)
	mu             sync.RWMutex
	activeBars     map[string]map[Timeframe]*OHLCBar // symbol -> timeframe -> bar

	// Output channel
	ohlcChannel    chan *OHLCBar

	// Supported timeframes
	timeframes     []Timeframe

	// Context
	ctx            context.Context
}

// NewOHLCEngine creates a new OHLC aggregation engine
func NewOHLCEngine(config *PipelineConfig, stats *PipelineStats) *OHLCEngine {
	return &OHLCEngine{
		config:      config,
		stats:       stats,
		activeBars:  make(map[string]map[Timeframe]*OHLCBar),
		ohlcChannel: make(chan *OHLCBar, config.OHLCBufferSize),
		timeframes:  []Timeframe{TF_M1, TF_M5, TF_M15, TF_H1, TF_H4, TF_D1},
	}
}

// Start starts the OHLC engine
func (o *OHLCEngine) Start(ctx context.Context) error {
	o.ctx = ctx

	// Start bar closing goroutine (checks for completed bars)
	go o.barClosingWorker()

	log.Println("[OHLCEngine] Started with timeframes: M1, M5, M15, H1, H4, D1")
	return nil
}

// ProcessTick processes a normalized tick and updates OHLC bars
func (o *OHLCEngine) ProcessTick(tick *NormalizedTick) {
	startTime := time.Now()

	o.mu.Lock()
	defer o.mu.Unlock()

	// Initialize symbol map if needed
	if _, exists := o.activeBars[tick.Symbol]; !exists {
		o.activeBars[tick.Symbol] = make(map[Timeframe]*OHLCBar)
	}

	// Use mid-price for OHLC
	price := (tick.Bid + tick.Ask) / 2.0

	// Update all timeframes
	for _, tf := range o.timeframes {
		bar := o.getOrCreateBar(tick.Symbol, tf, tick.Timestamp)

		// Update OHLC values
		if bar.Open == 0 {
			bar.Open = price
		}
		if price > bar.High || bar.High == 0 {
			bar.High = price
		}
		if price < bar.Low || bar.Low == 0 {
			bar.Low = price
		}
		bar.Close = price
		bar.Volume += int64(tick.Spread * 100000) // Approximate volume from spread
		bar.TickCount++
	}

	// Update latency stats
	o.stats.mu.Lock()
	latency := time.Since(startTime).Milliseconds()
	if o.stats.AvgOHLCLatencyMs == 0 {
		o.stats.AvgOHLCLatencyMs = float64(latency)
	} else {
		o.stats.AvgOHLCLatencyMs = (o.stats.AvgOHLCLatencyMs * 0.9) + (float64(latency) * 0.1)
	}
	o.stats.mu.Unlock()
}

// getOrCreateBar returns existing bar or creates new one with aligned timestamps
func (o *OHLCEngine) getOrCreateBar(symbol string, tf Timeframe, timestamp time.Time) *OHLCBar {
	// Calculate aligned bar start time
	openTime := o.alignTimestamp(timestamp, tf)
	closeTime := openTime.Add(time.Duration(tf) * time.Second)

	// Check if we have an active bar for this period
	if bar, exists := o.activeBars[symbol][tf]; exists {
		if bar.OpenTime.Equal(openTime) {
			return bar
		}

		// Bar period has changed, close the old bar
		if !bar.IsClosed {
			bar.IsClosed = true
			bar.CloseTime = closeTime.Add(-1 * time.Second)

			// Send completed bar
			select {
			case o.ohlcChannel <- bar:
				o.stats.mu.Lock()
				o.stats.OHLCBarsGenerated++
				o.stats.mu.Unlock()
			default:
				// Buffer full, drop bar
			}
		}
	}

	// Create new bar
	newBar := &OHLCBar{
		Symbol:    symbol,
		Timeframe: tf,
		OpenTime:  openTime,
		CloseTime: closeTime,
		IsClosed:  false,
	}

	o.activeBars[symbol][tf] = newBar
	return newBar
}

// alignTimestamp aligns timestamp to bar boundary
func (o *OHLCEngine) alignTimestamp(t time.Time, tf Timeframe) time.Time {
	seconds := int64(tf)
	unix := t.Unix()
	aligned := (unix / seconds) * seconds

	return time.Unix(aligned, 0).UTC()
}

// barClosingWorker periodically checks and closes completed bars
func (o *OHLCEngine) barClosingWorker() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-o.ctx.Done():
			return
		case now := <-ticker.C:
			o.closeCompletedBars(now)
		}
	}
}

// closeCompletedBars checks all active bars and closes those that have ended
func (o *OHLCEngine) closeCompletedBars(now time.Time) {
	o.mu.Lock()
	defer o.mu.Unlock()

	for symbol, tfBars := range o.activeBars {
		for tf, bar := range tfBars {
			if !bar.IsClosed && now.After(bar.CloseTime) {
				// Close the bar
				bar.IsClosed = true

				log.Printf("[OHLCEngine] Closing bar: %s %v [%s - %s] O:%.5f H:%.5f L:%.5f C:%.5f V:%d",
					symbol, tf, bar.OpenTime.Format("15:04:05"), bar.CloseTime.Format("15:04:05"),
					bar.Open, bar.High, bar.Low, bar.Close, bar.Volume)

				// Send to channel
				select {
				case o.ohlcChannel <- bar:
					o.stats.mu.Lock()
					o.stats.OHLCBarsGenerated++
					o.stats.mu.Unlock()
				default:
					log.Printf("[OHLCEngine] WARN: OHLC buffer full, dropping bar for %s", symbol)
				}

				// Remove from active bars (will be recreated on next tick)
				delete(o.activeBars[symbol], tf)
			}
		}
	}
}

// GetOHLCChannel returns the OHLC output channel
func (o *OHLCEngine) GetOHLCChannel() <-chan *OHLCBar {
	return o.ohlcChannel
}

// GetActiveBar returns the current active bar for a symbol/timeframe
func (o *OHLCEngine) GetActiveBar(symbol string, tf Timeframe) *OHLCBar {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if tfBars, exists := o.activeBars[symbol]; exists {
		if bar, exists := tfBars[tf]; exists {
			// Return a copy
			barCopy := *bar
			return &barCopy
		}
	}

	return nil
}

// GetAllActiveBars returns all active bars for a symbol
func (o *OHLCEngine) GetAllActiveBars(symbol string) []*OHLCBar {
	o.mu.RLock()
	defer o.mu.RUnlock()

	bars := make([]*OHLCBar, 0)

	if tfBars, exists := o.activeBars[symbol]; exists {
		for _, bar := range tfBars {
			barCopy := *bar
			bars = append(bars, &barCopy)
		}
	}

	return bars
}

// BackfillOHLC rebuilds OHLC from historical ticks
func (o *OHLCEngine) BackfillOHLC(symbol string, ticks []*NormalizedTick) {
	log.Printf("[OHLCEngine] Backfilling OHLC for %s from %d ticks", symbol, len(ticks))

	for _, tick := range ticks {
		o.ProcessTick(tick)
	}

	log.Printf("[OHLCEngine] Backfill complete for %s", symbol)
}
