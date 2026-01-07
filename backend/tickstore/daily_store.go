package tickstore

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// DailyStore handles daily file rotation for tick storage
type DailyStore struct {
	mu          sync.RWMutex
	basePath    string
	brokerID    string
	currentDay  string
	todayTicks  map[string][]Tick // symbol -> today's ticks
	maxDaysKeep int               // Number of days to keep
}

// NewDailyStore creates a new daily tick store
func NewDailyStore(brokerID string, maxDaysKeep int) *DailyStore {
	ds := &DailyStore{
		basePath:    "data/ticks",
		brokerID:    brokerID,
		currentDay:  time.Now().Format("2006-01-02"),
		todayTicks:  make(map[string][]Tick),
		maxDaysKeep: maxDaysKeep,
	}

	// Ensure base directory exists
	os.MkdirAll(ds.basePath, 0755)

	// Load today's ticks for all symbols
	ds.loadToday()

	// Start background jobs
	go ds.persistPeriodically()
	go ds.rotateDaily()

	log.Printf("[DailyStore] Initialized for broker '%s', keeping %d days", brokerID, maxDaysKeep)
	return ds
}

// StoreTick stores a tick for a symbol
func (ds *DailyStore) StoreTick(symbol string, tick Tick) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Check for day rotation
	today := time.Now().Format("2006-01-02")
	if today != ds.currentDay {
		ds.rotateToPersist()
		ds.currentDay = today
	}

	ds.todayTicks[symbol] = append(ds.todayTicks[symbol], tick)
}

// GetHistory returns historical ticks for a symbol across multiple days
func (ds *DailyStore) GetHistory(symbol string, limit int, daysBack int) []Tick {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var allTicks []Tick

	// Get ticks from past days (from files)
	today := time.Now()
	for i := daysBack; i >= 1; i-- {
		date := today.AddDate(0, 0, -i).Format("2006-01-02")
		ticks := ds.loadDayForSymbol(symbol, date)
		allTicks = append(allTicks, ticks...)
	}

	// Add today's ticks
	allTicks = append(allTicks, ds.todayTicks[symbol]...)

	// Apply limit
	if limit > 0 && len(allTicks) > limit {
		allTicks = allTicks[len(allTicks)-limit:]
	}

	return allTicks
}

// GetSymbols returns all symbols with stored data
func (ds *DailyStore) GetSymbols() []string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	symbolMap := make(map[string]bool)

	// Add symbols from today
	for sym := range ds.todayTicks {
		symbolMap[sym] = true
	}

	// Add symbols from files
	dirs, _ := os.ReadDir(ds.basePath)
	for _, d := range dirs {
		if d.IsDir() {
			symbolMap[d.Name()] = true
		}
	}

	symbols := make([]string, 0, len(symbolMap))
	for sym := range symbolMap {
		symbols = append(symbols, sym)
	}
	sort.Strings(symbols)
	return symbols
}

// loadToday loads today's ticks for all symbols
func (ds *DailyStore) loadToday() {
	dirs, err := os.ReadDir(ds.basePath)
	if err != nil {
		return
	}

	totalLoaded := 0
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		symbol := d.Name()
		ticks := ds.loadDayForSymbol(symbol, ds.currentDay)
		if len(ticks) > 0 {
			ds.todayTicks[symbol] = ticks
			totalLoaded += len(ticks)
		}
	}

	if totalLoaded > 0 {
		log.Printf("[DailyStore] Loaded %d ticks for today across %d symbols", totalLoaded, len(ds.todayTicks))
	}
}

// loadDayForSymbol loads ticks for a specific symbol and day
func (ds *DailyStore) loadDayForSymbol(symbol, date string) []Tick {
	filePath := filepath.Join(ds.basePath, symbol, date+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	var ticks []Tick
	if err := json.Unmarshal(data, &ticks); err != nil {
		log.Printf("[DailyStore] Error parsing %s: %v", filePath, err)
		return nil
	}

	return ticks
}

// persistPeriodically saves current day's ticks every 30 seconds
func (ds *DailyStore) persistPeriodically() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ds.persistToday()
	}
}

// persistToday saves today's ticks to files
func (ds *DailyStore) persistToday() {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if len(ds.todayTicks) == 0 {
		return
	}

	totalPersisted := 0
	for symbol, ticks := range ds.todayTicks {
		if len(ticks) == 0 {
			continue
		}

		// Ensure symbol directory exists
		symbolDir := filepath.Join(ds.basePath, symbol)
		os.MkdirAll(symbolDir, 0755)

		filePath := filepath.Join(symbolDir, ds.currentDay+".json")

		// Atomic write: temp file + rename
		tempPath := filePath + ".tmp"
		data, err := json.Marshal(ticks)
		if err != nil {
			log.Printf("[DailyStore] Marshal error for %s: %v", symbol, err)
			continue
		}

		if err := os.WriteFile(tempPath, data, 0644); err != nil {
			log.Printf("[DailyStore] Write error for %s: %v", symbol, err)
			continue
		}

		if err := os.Rename(tempPath, filePath); err != nil {
			log.Printf("[DailyStore] Rename error for %s: %v", symbol, err)
			os.Remove(tempPath)
			continue
		}

		totalPersisted += len(ticks)
	}

	if totalPersisted > 0 {
		log.Printf("[DailyStore] Persisted %d ticks across %d symbols", totalPersisted, len(ds.todayTicks))
	}
}

// rotateToPersist saves today's ticks and clears for new day
func (ds *DailyStore) rotateToPersist() {
	log.Printf("[DailyStore] Day rotation: %s -> %s", ds.currentDay, time.Now().Format("2006-01-02"))
	ds.persistTodayLocked()
	ds.todayTicks = make(map[string][]Tick)
}

// persistTodayLocked saves today's ticks (caller must hold lock)
func (ds *DailyStore) persistTodayLocked() {
	for symbol, ticks := range ds.todayTicks {
		if len(ticks) == 0 {
			continue
		}

		symbolDir := filepath.Join(ds.basePath, symbol)
		os.MkdirAll(symbolDir, 0755)

		filePath := filepath.Join(symbolDir, ds.currentDay+".json")
		tempPath := filePath + ".tmp"

		data, _ := json.Marshal(ticks)
		if err := os.WriteFile(tempPath, data, 0644); err == nil {
			os.Rename(tempPath, filePath)
		}
	}
}

// rotateDaily checks at midnight for day change and cleans old files
func (ds *DailyStore) rotateDaily() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ds.cleanOldFiles()
	}
}

// cleanOldFiles removes files older than maxDaysKeep
func (ds *DailyStore) cleanOldFiles() {
	if ds.maxDaysKeep <= 0 {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -ds.maxDaysKeep)
	cutoffStr := cutoff.Format("2006-01-02")

	dirs, _ := os.ReadDir(ds.basePath)
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}

		symbolDir := filepath.Join(ds.basePath, d.Name())
		files, _ := os.ReadDir(symbolDir)

		for _, f := range files {
			if f.IsDir() {
				continue
			}

			// Extract date from filename (e.g., "2026-01-01.json")
			name := f.Name()
			if !strings.HasSuffix(name, ".json") {
				continue
			}

			dateStr := strings.TrimSuffix(name, ".json")
			if dateStr < cutoffStr {
				filePath := filepath.Join(symbolDir, name)
				if err := os.Remove(filePath); err == nil {
					log.Printf("[DailyStore] Cleaned old file: %s", filePath)
				}
			}
		}
	}
}

// GetTodayTickCount returns tick count for today
func (ds *DailyStore) GetTodayTickCount(symbol string) int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return len(ds.todayTicks[symbol])
}

// GetAvailableDates returns available dates for a symbol
func (ds *DailyStore) GetAvailableDates(symbol string) []string {
	symbolDir := filepath.Join(ds.basePath, symbol)
	files, err := os.ReadDir(symbolDir)
	if err != nil {
		return nil
	}

	var dates []string
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		date := strings.TrimSuffix(f.Name(), ".json")
		dates = append(dates, date)
	}

	sort.Strings(dates)
	return dates
}

// MergeHistoricalData merges ticks from an external source (e.g., imported data)
func (ds *DailyStore) MergeHistoricalData(symbol string, ticks []Tick) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Group ticks by date
	ticksByDate := make(map[string][]Tick)
	for _, tick := range ticks {
		date := tick.Timestamp.Format("2006-01-02")
		ticksByDate[date] = append(ticksByDate[date], tick)
	}

	// Merge each date
	for date, dateTicks := range ticksByDate {
		// Load existing
		existing := ds.loadDayForSymbol(symbol, date)

		// Merge (existing + new, sorted by timestamp)
		merged := append(existing, dateTicks...)
		sort.Slice(merged, func(i, j int) bool {
			return merged[i].Timestamp.Before(merged[j].Timestamp)
		})

		// Remove duplicates (same timestamp)
		deduped := make([]Tick, 0, len(merged))
		seen := make(map[int64]bool)
		for _, t := range merged {
			ts := t.Timestamp.UnixNano()
			if !seen[ts] {
				seen[ts] = true
				deduped = append(deduped, t)
			}
		}

		// Save
		symbolDir := filepath.Join(ds.basePath, symbol)
		os.MkdirAll(symbolDir, 0755)

		filePath := filepath.Join(symbolDir, date+".json")
		data, _ := json.Marshal(deduped)

		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to save merged data for %s/%s: %w", symbol, date, err)
		}

		log.Printf("[DailyStore] Merged %d ticks for %s on %s", len(deduped), symbol, date)
	}

	return nil
}
