package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/epic1st/rtx/backend/auth"
	"github.com/epic1st/rtx/backend/tickstore"
)

// AdminHistoryHandler handles admin endpoints for historical data management
type AdminHistoryHandler struct {
	tickStore   tickstore.TickStorageService
	authService *auth.Service
}

// StatsResponse contains statistics about historical data storage
type StatsResponse struct {
	TotalSymbols    int                      `json:"total_symbols"`
	TotalTicks      int64                    `json:"total_ticks"`
	TotalSizeBytes  int64                    `json:"total_size_bytes"`
	TotalSizeMB     float64                  `json:"total_size_mb"`
	OldestTick      time.Time                `json:"oldest_tick"`
	NewestTick      time.Time                `json:"newest_tick"`
	DaysOfData      int                      `json:"days_of_data"`
	SymbolStats     []SymbolStats            `json:"symbol_stats"`
	StorageHealth   string                   `json:"storage_health"`
}

// SymbolStats contains stats for a single symbol
type SymbolStats struct {
	Symbol      string    `json:"symbol"`
	TickCount   int64     `json:"tick_count"`
	SizeBytes   int64     `json:"size_bytes"`
	OldestTick  time.Time `json:"oldest_tick"`
	NewestTick  time.Time `json:"newest_tick"`
	AvgTickRate float64   `json:"avg_tick_rate"` // ticks per day
}

// ImportDataRequest is the request for bulk data import
type ImportDataRequest struct {
	Source    string                       `json:"source"` // "csv", "json", "external"
	Format    string                       `json:"format"` // "tick", "ohlc"
	Data      map[string][]tickstore.Tick  `json:"data"`   // symbol -> ticks
	Overwrite bool                         `json:"overwrite"`
}

// CleanupRequest is the request for cleaning up old data
type CleanupRequest struct {
	OlderThanDays int      `json:"older_than_days"`
	Symbols       []string `json:"symbols,omitempty"` // Empty = all symbols
	DryRun        bool     `json:"dry_run"`
}

// CompressRequest is the request for compressing historical data
type CompressRequest struct {
	Symbols []string `json:"symbols,omitempty"`
	Method  string   `json:"method"` // "gzip", "lz4", "zstd"
}

// BackupRequest is the request for backing up historical data
type BackupRequest struct {
	Symbols     []string `json:"symbol,omitempty"`
	Destination string   `json:"destination"`
	Compress    bool     `json:"compress"`
}

// MonitoringResponse contains real-time monitoring data
type MonitoringResponse struct {
	ActiveSymbols    int                    `json:"active_symbols"`
	TicksPerSecond   float64                `json:"ticks_per_second"`
	AvgLatency       float64                `json:"avg_latency_ms"`
	MemoryUsageMB    float64                `json:"memory_usage_mb"`
	DiskUsageMB      float64                `json:"disk_usage_mb"`
	LastTickReceived time.Time              `json:"last_tick_received"`
	Health           string                 `json:"health"`
	Alerts           []string               `json:"alerts"`
}

// NewAdminHistoryHandler creates a new admin history handler
func NewAdminHistoryHandler(ts tickstore.TickStorageService, authService *auth.Service) *AdminHistoryHandler {
	return &AdminHistoryHandler{
		tickStore:   ts,
		authService: authService,
	}
}

// HandleGetStats returns comprehensive statistics about historical data
func (h *AdminHistoryHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	h.setCORS(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	symbols := h.tickStore.GetSymbols()

	stats := StatsResponse{
		TotalSymbols: len(symbols),
		SymbolStats:  make([]SymbolStats, 0, len(symbols)),
	}

	var totalTicks int64
	var totalSize int64
	var oldestTick, newestTick time.Time

	for _, symbol := range symbols {
		// Get tick metadata (limited query for performance)
		ticks := h.tickStore.GetHistory(symbol, 1000)
		tickCount := int64(len(ticks))
		totalTicks += tickCount

		if len(ticks) > 0 {
			oldest := ticks[0].Timestamp
			newest := ticks[len(ticks)-1].Timestamp

			if oldestTick.IsZero() || oldest.Before(oldestTick) {
				oldestTick = oldest
			}
			if newest.After(newestTick) {
				newestTick = newest
			}

			estimatedSize := tickCount * 100

			daysDiff := newest.Sub(oldest).Hours() / 24
			avgTickRate := 0.0
			if daysDiff > 0 {
				avgTickRate = float64(tickCount) / daysDiff
			}

			stats.SymbolStats = append(stats.SymbolStats, SymbolStats{
				Symbol:      symbol,
				TickCount:   tickCount,
				SizeBytes:   estimatedSize,
				OldestTick:  oldest,
				NewestTick:  newest,
				AvgTickRate: avgTickRate,
			})

			totalSize += estimatedSize
		}
	}

	stats.TotalTicks = totalTicks
	stats.TotalSizeBytes = totalSize
	stats.TotalSizeMB = float64(totalSize) / 1024 / 1024
	stats.OldestTick = oldestTick
	stats.NewestTick = newestTick

	if !oldestTick.IsZero() && !newestTick.IsZero() {
		stats.DaysOfData = int(newestTick.Sub(oldestTick).Hours() / 24)
	}

	if stats.TotalSizeMB > 10000 {
		stats.StorageHealth = "warning"
	} else if stats.TotalSizeMB > 50000 {
		stats.StorageHealth = "critical"
	} else {
		stats.StorageHealth = "healthy"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)

	log.Printf("[AdminHistory] Stats: %d symbols, %d ticks, %.2f MB",
		stats.TotalSymbols, stats.TotalTicks, stats.TotalSizeMB)
}

// HandleImportData handles bulk data import
func (h *AdminHistoryHandler) HandleImportData(w http.ResponseWriter, r *http.Request) {
	h.setCORS(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ImportDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	totalImported := 0

	for symbol, ticks := range req.Data {
		if ts, ok := h.tickStore.(*tickstore.TickStore); ok {
			dailyStore := ts.GetDailyStore()
			if err := dailyStore.MergeHistoricalData(symbol, ticks); err != nil {
				log.Printf("[AdminHistory] Failed to import %s: %v", symbol, err)
				continue
			}
			totalImported += len(ticks)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"symbols":        len(req.Data),
		"total_imported": totalImported,
		"source":         req.Source,
	})

	log.Printf("[AdminHistory] Imported %d ticks from %s", totalImported, req.Source)
}

// HandleCleanupOldData handles cleanup of old historical data
func (h *AdminHistoryHandler) HandleCleanupOldData(w http.ResponseWriter, r *http.Request) {
	h.setCORS(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CleanupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.OlderThanDays <= 0 {
		http.Error(w, "older_than_days must be positive", http.StatusBadRequest)
		return
	}

	cutoffDate := time.Now().AddDate(0, 0, -req.OlderThanDays)
	cutoffStr := cutoffDate.Format("2006-01-02")

	symbols := req.Symbols
	if len(symbols) == 0 {
		symbols = h.tickStore.GetSymbols()
	}

	filesDeleted := 0
	totalSize := int64(0)

	for _, symbol := range symbols {
		// Validate symbol to prevent path traversal
		if !isValidSymbol(symbol) {
			log.Printf("[AdminHistory] Invalid symbol '%s' (skipping)", symbol)
			continue
		}

		basePath := filepath.Join("data", "ticks", symbol)

		files, err := os.ReadDir(basePath)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
				continue
			}

			fileName := file.Name()
			dateStr := fileName[:len(fileName)-5]

			if dateStr < cutoffStr {
				if !req.DryRun {
					filePath := filepath.Join(basePath, fileName)
					info, _ := file.Info()
					if info != nil {
						totalSize += info.Size()
					}
					os.Remove(filePath)
				}
				filesDeleted++
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"files_deleted": filesDeleted,
		"size_freed_mb": float64(totalSize) / 1024 / 1024,
		"cutoff_date":   cutoffStr,
		"dry_run":       req.DryRun,
	})

	action := "Dry run"
	if !req.DryRun {
		action = "Cleaned up"
	}
	log.Printf("[AdminHistory] %s: %d files (%.2f MB) older than %s",
		action, filesDeleted, float64(totalSize)/1024/1024, cutoffStr)
}

// HandleCompressData handles compression of historical data
func (h *AdminHistoryHandler) HandleCompressData(w http.ResponseWriter, r *http.Request) {
	h.setCORS(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CompressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": "Compression feature not yet implemented",
		"method":  req.Method,
	})
}

// HandleBackup handles backing up historical data
func (h *AdminHistoryHandler) HandleBackup(w http.ResponseWriter, r *http.Request) {
	h.setCORS(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Destination == "" {
		http.Error(w, "destination is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     false,
		"message":     "Backup feature not yet implemented",
		"destination": req.Destination,
	})
}

// HandleGetMonitoring returns real-time monitoring data
func (h *AdminHistoryHandler) HandleGetMonitoring(w http.ResponseWriter, r *http.Request) {
	h.setCORS(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	symbols := h.tickStore.GetSymbols()

	var diskUsage int64
	basePath := "data/ticks"
	filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			diskUsage += info.Size()
		}
		return nil
	})

	monitoring := MonitoringResponse{
		ActiveSymbols:    len(symbols),
		TicksPerSecond:   0,
		AvgLatency:       0,
		MemoryUsageMB:    0,
		DiskUsageMB:      float64(diskUsage) / 1024 / 1024,
		LastTickReceived: time.Now(),
		Health:           "healthy",
		Alerts:           []string{},
	}

	if monitoring.DiskUsageMB > 10000 {
		monitoring.Alerts = append(monitoring.Alerts, "Disk usage exceeds 10GB")
		monitoring.Health = "warning"
	}

	if len(symbols) == 0 {
		monitoring.Alerts = append(monitoring.Alerts, "No active symbols")
		monitoring.Health = "warning"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(monitoring)
}

// Helper: setCORS sets CORS headers
func (h *AdminHistoryHandler) setCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
}

// Helper: isValidSymbol validates symbol format to prevent path traversal
func isValidSymbol(symbol string) bool {
	// Only allow alphanumeric characters (A-Z, 0-9)
	// Prevents path traversal attacks like "../../../etc/passwd"
	for _, c := range symbol {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return len(symbol) > 0 && len(symbol) <= 20
}
