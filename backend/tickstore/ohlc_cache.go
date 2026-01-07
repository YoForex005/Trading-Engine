package tickstore

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Timeframe represents a candle timeframe
type Timeframe string

const (
	TF_M1  Timeframe = "M1"
	TF_M5  Timeframe = "M5"
	TF_M15 Timeframe = "M15"
	TF_H1  Timeframe = "H1"
	TF_H4  Timeframe = "H4"
	TF_D1  Timeframe = "D1"
)

// TimeframeSeconds returns the duration of a timeframe in seconds
func TimeframeSeconds(tf Timeframe) int64 {
	switch tf {
	case TF_M1:
		return 60
	case TF_M5:
		return 300
	case TF_M15:
		return 900
	case TF_H1:
		return 3600
	case TF_H4:
		return 14400
	case TF_D1:
		return 86400
	default:
		return 60
	}
}

// OHLCCache manages pre-computed OHLC bars
type OHLCCache struct {
	mu          sync.RWMutex
	basePath    string
	bars        map[string]map[Timeframe][]OHLC // symbol -> timeframe -> bars
	currentBars map[string]map[Timeframe]*OHLC  // symbol -> timeframe -> current incomplete bar
	timeframes  []Timeframe
}

// NewOHLCCache creates a new OHLC cache
func NewOHLCCache(timeframes []Timeframe) *OHLCCache {
	if len(timeframes) == 0 {
		timeframes = []Timeframe{TF_M1, TF_M5, TF_H1, TF_D1}
	}

	cache := &OHLCCache{
		basePath:    "data/ohlc",
		bars:        make(map[string]map[Timeframe][]OHLC),
		currentBars: make(map[string]map[Timeframe]*OHLC),
		timeframes:  timeframes,
	}

	os.MkdirAll(cache.basePath, 0755)
	cache.loadAllCaches()

	go cache.persistPeriodically()

	log.Printf("[OHLCCache] Initialized with timeframes: %v", timeframes)
	return cache
}

// UpdateFromTick updates all timeframe bars from a new tick
func (c *OHLCCache) UpdateFromTick(symbol string, bid, ask float64, timestamp time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	price := (bid + ask) / 2
	ts := timestamp.Unix()

	// Initialize symbol if needed
	if c.bars[symbol] == nil {
		c.bars[symbol] = make(map[Timeframe][]OHLC)
	}
	if c.currentBars[symbol] == nil {
		c.currentBars[symbol] = make(map[Timeframe]*OHLC)
	}

	// Update each timeframe
	for _, tf := range c.timeframes {
		tfSecs := TimeframeSeconds(tf)
		candleTime := (ts / tfSecs) * tfSecs

		currentBar := c.currentBars[symbol][tf]

		if currentBar == nil || currentBar.Time != candleTime {
			// Finalize previous bar if exists
			if currentBar != nil {
				c.bars[symbol][tf] = append(c.bars[symbol][tf], *currentBar)
			}

			// Start new bar
			c.currentBars[symbol][tf] = &OHLC{
				Time:   candleTime,
				Open:   price,
				High:   price,
				Low:    price,
				Close:  price,
				Volume: 1,
			}
		} else {
			// Update existing bar
			if price > currentBar.High {
				currentBar.High = price
			}
			if price < currentBar.Low {
				currentBar.Low = price
			}
			currentBar.Close = price
			currentBar.Volume++
		}
	}
}

// GetBars returns OHLC bars for a symbol and timeframe
func (c *OHLCCache) GetBars(symbol string, tf Timeframe, limit int) []OHLC {
	c.mu.RLock()
	defer c.mu.RUnlock()

	symbolBars := c.bars[symbol]
	if symbolBars == nil {
		return []OHLC{}
	}

	bars := symbolBars[tf]
	if bars == nil {
		bars = []OHLC{}
	}

	// Include current bar if exists
	if c.currentBars[symbol] != nil && c.currentBars[symbol][tf] != nil {
		bars = append(bars, *c.currentBars[symbol][tf])
	}

	// Apply limit
	if limit > 0 && len(bars) > limit {
		bars = bars[len(bars)-limit:]
	}

	return bars
}

// GetSymbols returns all symbols with cached OHLC data
func (c *OHLCCache) GetSymbols() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	symbols := make([]string, 0, len(c.bars))
	for sym := range c.bars {
		symbols = append(symbols, sym)
	}
	sort.Strings(symbols)
	return symbols
}

// persistPeriodically saves cache to disk every 60 seconds
func (c *OHLCCache) persistPeriodically() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.PersistAll()
	}
}

// PersistAll saves all cached bars to disk
func (c *OHLCCache) PersistAll() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for symbol, tfBars := range c.bars {
		symbolDir := filepath.Join(c.basePath, symbol)
		os.MkdirAll(symbolDir, 0755)

		for tf, bars := range tfBars {
			// Include current bar
			allBars := bars
			if c.currentBars[symbol] != nil && c.currentBars[symbol][tf] != nil {
				allBars = append(allBars, *c.currentBars[symbol][tf])
			}

			if len(allBars) == 0 {
				continue
			}

			filePath := filepath.Join(symbolDir, string(tf)+".json")
			tempPath := filePath + ".tmp"

			data, err := json.Marshal(allBars)
			if err != nil {
				continue
			}

			if err := os.WriteFile(tempPath, data, 0644); err != nil {
				continue
			}

			os.Rename(tempPath, filePath)
		}
	}

	// Count total bars
	total := 0
	for _, tfBars := range c.bars {
		for _, bars := range tfBars {
			total += len(bars)
		}
	}

	if total > 0 {
		log.Printf("[OHLCCache] Persisted %d bars across %d symbols", total, len(c.bars))
	}
}

// loadAllCaches loads all cached OHLC data from disk
func (c *OHLCCache) loadAllCaches() {
	dirs, err := os.ReadDir(c.basePath)
	if err != nil {
		return
	}

	totalLoaded := 0
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}

		symbol := d.Name()
		c.bars[symbol] = make(map[Timeframe][]OHLC)
		c.currentBars[symbol] = make(map[Timeframe]*OHLC)

		symbolDir := filepath.Join(c.basePath, symbol)
		for _, tf := range c.timeframes {
			filePath := filepath.Join(symbolDir, string(tf)+".json")

			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			var bars []OHLC
			if err := json.Unmarshal(data, &bars); err != nil {
				log.Printf("[OHLCCache] Error parsing %s: %v", filePath, err)
				continue
			}

			if len(bars) > 0 {
				// Last bar becomes current (may be incomplete)
				c.currentBars[symbol][tf] = &bars[len(bars)-1]
				c.bars[symbol][tf] = bars[:len(bars)-1]
				totalLoaded += len(bars)
			}
		}
	}

	if totalLoaded > 0 {
		log.Printf("[OHLCCache] Loaded %d bars across %d symbols", totalLoaded, len(c.bars))
	}
}

// MergeHistoricalBars merges historical bars from external source
func (c *OHLCCache) MergeHistoricalBars(symbol string, tf Timeframe, newBars []OHLC) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.bars[symbol] == nil {
		c.bars[symbol] = make(map[Timeframe][]OHLC)
	}

	existing := c.bars[symbol][tf]

	// Merge and sort
	merged := append(existing, newBars...)
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Time < merged[j].Time
	})

	// Remove duplicates (same timestamp)
	deduped := make([]OHLC, 0, len(merged))
	seenTimes := make(map[int64]bool)
	for _, bar := range merged {
		if !seenTimes[bar.Time] {
			seenTimes[bar.Time] = true
			deduped = append(deduped, bar)
		}
	}

	c.bars[symbol][tf] = deduped
	log.Printf("[OHLCCache] Merged %d bars for %s/%s, total: %d", len(newBars), symbol, tf, len(deduped))
}

// RebuildFromTicks rebuilds OHLC cache from tick data
func (c *OHLCCache) RebuildFromTicks(symbol string, ticks []Tick) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(ticks) == 0 {
		return
	}

	// Initialize
	if c.bars[symbol] == nil {
		c.bars[symbol] = make(map[Timeframe][]OHLC)
	}
	if c.currentBars[symbol] == nil {
		c.currentBars[symbol] = make(map[Timeframe]*OHLC)
	}

	// Sort ticks by timestamp
	sort.Slice(ticks, func(i, j int) bool {
		return ticks[i].Timestamp.Before(ticks[j].Timestamp)
	})

	// Build bars for each timeframe
	for _, tf := range c.timeframes {
		tfSecs := TimeframeSeconds(tf)
		barMap := make(map[int64]*OHLC)
		var candleTimes []int64

		for _, tick := range ticks {
			price := (tick.Bid + tick.Ask) / 2
			ts := tick.Timestamp.Unix()
			candleTime := (ts / tfSecs) * tfSecs

			if bar, exists := barMap[candleTime]; exists {
				if price > bar.High {
					bar.High = price
				}
				if price < bar.Low {
					bar.Low = price
				}
				bar.Close = price
				bar.Volume++
			} else {
				barMap[candleTime] = &OHLC{
					Time:   candleTime,
					Open:   price,
					High:   price,
					Low:    price,
					Close:  price,
					Volume: 1,
				}
				candleTimes = append(candleTimes, candleTime)
			}
		}

		// Convert to sorted slice
		sort.Slice(candleTimes, func(i, j int) bool {
			return candleTimes[i] < candleTimes[j]
		})

		bars := make([]OHLC, 0, len(candleTimes))
		for _, t := range candleTimes {
			bars = append(bars, *barMap[t])
		}

		c.bars[symbol][tf] = bars
	}

	log.Printf("[OHLCCache] Rebuilt cache for %s from %d ticks", symbol, len(ticks))
}
