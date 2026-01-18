package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
)

// ExposureSnapshot represents exposure at a point in time
type ExposureSnapshot struct {
	Timestamp     int64              `json:"timestamp"`
	SymbolData    map[string]float64 `json:"symbolData"`
	NetExposure   float64            `json:"netExposure"`
	TotalLong     float64            `json:"totalLong"`
	TotalShort    float64            `json:"totalShort"`
	PositionCount int                `json:"positionCount"`
}

// SymbolExposure represents current exposure for a symbol
type SymbolExposure struct {
	Symbol         string  `json:"symbol"`
	NetExposure    float64 `json:"net_exposure"`
	Long           float64 `json:"long"`
	Short          float64 `json:"short"`
	UtilizationPct float64 `json:"utilization_pct"`
	Limit          float64 `json:"limit"`
	Status         string  `json:"status"`
}

// ExposureTimeline represents exposure history for a symbol
type ExposureTimeline struct {
	Symbol   string                    `json:"symbol"`
	Timeline []ExposureTimelineEntry   `json:"timeline"`
}

// ExposureTimelineEntry represents a single point in timeline
type ExposureTimelineEntry struct {
	Timestamp      int64   `json:"timestamp"`
	NetExposure    float64 `json:"net_exposure"`
	Long           float64 `json:"long"`
	Short          float64 `json:"short"`
	UtilizationPct float64 `json:"utilization_pct"`
}

// HandleExposureHeatmap returns exposure heatmap data
func (h *APIHandler) HandleExposureHeatmap(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse query parameters
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	interval := r.URL.Query().Get("interval")
	symbolsParam := r.URL.Query().Get("symbols")

	// Default to last 24 hours if not specified
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	if startTimeStr != "" {
		if ts, err := strconv.ParseInt(startTimeStr, 10, 64); err == nil {
			startTime = time.Unix(ts, 0)
		}
	}

	if endTimeStr != "" {
		if ts, err := strconv.ParseInt(endTimeStr, 10, 64); err == nil {
			endTime = time.Unix(ts, 0)
		}
	}

	// Default interval: 1h
	if interval == "" {
		interval = "1h"
	}

	// Parse symbols filter
	var symbolsFilter []string
	if symbolsParam != "" {
		symbolsFilter = strings.Split(symbolsParam, ",")
	}

	// Get all positions
	allPositions := h.engine.GetAllPositions()

	// Build timeline data
	snapshots := h.buildExposureTimeline(allPositions, startTime, endTime, interval, symbolsFilter)

	// Extract symbols and timestamps
	symbolsSet := make(map[string]bool)
	var timestamps []int64
	maxExposure := 0.0
	minExposure := 0.0

	for _, snapshot := range snapshots {
		timestamps = append(timestamps, snapshot.Timestamp)
		for symbol, exposure := range snapshot.SymbolData {
			symbolsSet[symbol] = true
			if exposure > maxExposure {
				maxExposure = exposure
			}
			if exposure < minExposure {
				minExposure = exposure
			}
		}
	}

	// Convert symbols set to sorted array
	var symbols []string
	for symbol := range symbolsSet {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

	// Build 2D data array [timestamp][symbol]
	data := make([][]float64, len(timestamps))
	for i := range timestamps {
		data[i] = make([]float64, len(symbols))
		for j, symbol := range symbols {
			exposure, exists := snapshots[i].SymbolData[symbol]
			if exists {
				data[i][j] = exposure
			} else {
				data[i][j] = 0
			}
		}
	}

	response := map[string]interface{}{
		"timestamps":   timestamps,
		"symbols":      symbols,
		"data":         data,
		"max_exposure": maxExposure,
		"min_exposure": minExposure,
		"interval":     interval,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleCurrentExposure returns current exposure by symbol
func (h *APIHandler) HandleCurrentExposure(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get all positions
	allPositions := h.engine.GetAllPositions()

	// Aggregate by symbol
	symbolExposures := h.calculateSymbolExposures(allPositions)

	// Sort by absolute net exposure (largest first)
	sort.Slice(symbolExposures, func(i, j int) bool {
		absI := symbolExposures[i].NetExposure
		if absI < 0 {
			absI = -absI
		}
		absJ := symbolExposures[j].NetExposure
		if absJ < 0 {
			absJ = -absJ
		}
		return absI > absJ
	})

	response := map[string]interface{}{
		"symbols": symbolExposures,
		"timestamp": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleExposureHistory returns exposure timeline for a specific symbol
func (h *APIHandler) HandleExposureHistory(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract symbol from URL path: /api/analytics/exposure/history/{symbol}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 {
		http.Error(w, "Symbol parameter required", http.StatusBadRequest)
		return
	}
	symbol := parts[5]

	// Parse query parameters
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	interval := r.URL.Query().Get("interval")

	// Default to last 7 days if not specified
	endTime := time.Now()
	startTime := endTime.Add(-7 * 24 * time.Hour)

	if startTimeStr != "" {
		if ts, err := strconv.ParseInt(startTimeStr, 10, 64); err == nil {
			startTime = time.Unix(ts, 0)
		}
	}

	if endTimeStr != "" {
		if ts, err := strconv.ParseInt(endTimeStr, 10, 64); err == nil {
			endTime = time.Unix(ts, 0)
		}
	}

	// Default interval: 1h
	if interval == "" {
		interval = "1h"
	}

	// Get all positions for this symbol
	allPositions := h.engine.GetAllPositions()
	var symbolPositions []*core.Position
	for _, pos := range allPositions {
		if pos.Symbol == symbol {
			symbolPositions = append(symbolPositions, pos)
		}
	}

	// Build timeline
	snapshots := h.buildExposureTimeline(symbolPositions, startTime, endTime, interval, []string{symbol})

	// Extract timeline entries
	var timeline []ExposureTimelineEntry
	for _, snapshot := range snapshots {
		exposure, exists := snapshot.SymbolData[symbol]
		if !exists {
			exposure = 0
		}

		// Calculate long/short from current positions at this time
		var long, short float64
		for _, pos := range symbolPositions {
			if pos.OpenTime.Before(time.Unix(snapshot.Timestamp, 0)) {
				notional := h.calculateNotionalValue(pos)
				if pos.Side == "BUY" {
					long += notional
				} else {
					short += notional
				}
			}
		}

		// Calculate utilization (using a default limit of 1,000,000 for now)
		limit := 1000000.0
		utilizationPct := 0.0
		if limit > 0 {
			absExposure := exposure
			if absExposure < 0 {
				absExposure = -absExposure
			}
			utilizationPct = (absExposure / limit) * 100
		}

		timeline = append(timeline, ExposureTimelineEntry{
			Timestamp:      snapshot.Timestamp,
			NetExposure:    exposure,
			Long:           long,
			Short:          short,
			UtilizationPct: utilizationPct,
		})
	}

	response := ExposureTimeline{
		Symbol:   symbol,
		Timeline: timeline,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions

func (h *APIHandler) buildExposureTimeline(positions []*core.Position, startTime, endTime time.Time, interval string, symbolsFilter []string) []ExposureSnapshot {
	// Parse interval
	intervalDuration := h.parseInterval(interval)

	var snapshots []ExposureSnapshot
	current := startTime

	for current.Before(endTime) || current.Equal(endTime) {
		snapshot := ExposureSnapshot{
			Timestamp:  current.Unix(),
			SymbolData: make(map[string]float64),
		}

		var totalLong, totalShort float64
		positionCount := 0

		// Calculate exposure at this timestamp
		for _, pos := range positions {
			// Only include positions that were open at this time
			if pos.OpenTime.After(current) {
				continue
			}

			// Filter by symbols if specified
			if len(symbolsFilter) > 0 {
				found := false
				for _, s := range symbolsFilter {
					if s == pos.Symbol {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// Calculate notional value
			notional := h.calculateNotionalValue(pos)

			if pos.Side == "BUY" {
				snapshot.SymbolData[pos.Symbol] += notional
				totalLong += notional
			} else {
				snapshot.SymbolData[pos.Symbol] -= notional
				totalShort += notional
			}
			positionCount++
		}

		snapshot.TotalLong = totalLong
		snapshot.TotalShort = totalShort
		snapshot.NetExposure = totalLong - totalShort
		snapshot.PositionCount = positionCount

		snapshots = append(snapshots, snapshot)
		current = current.Add(intervalDuration)
	}

	return snapshots
}

func (h *APIHandler) calculateSymbolExposures(positions []*core.Position) []SymbolExposure {
	exposureMap := make(map[string]*SymbolExposure)

	for _, pos := range positions {
		if _, exists := exposureMap[pos.Symbol]; !exists {
			exposureMap[pos.Symbol] = &SymbolExposure{
				Symbol: pos.Symbol,
				Limit:  1000000.0, // Default limit, can be made configurable
			}
		}

		notional := h.calculateNotionalValue(pos)

		if pos.Side == "BUY" {
			exposureMap[pos.Symbol].Long += notional
			exposureMap[pos.Symbol].NetExposure += notional
		} else {
			exposureMap[pos.Symbol].Short += notional
			exposureMap[pos.Symbol].NetExposure -= notional
		}
	}

	// Calculate utilization and status
	var result []SymbolExposure
	for _, exp := range exposureMap {
		absExposure := exp.NetExposure
		if absExposure < 0 {
			absExposure = -absExposure
		}

		if exp.Limit > 0 {
			exp.UtilizationPct = (absExposure / exp.Limit) * 100
		}

		// Determine status
		if exp.UtilizationPct > 90 {
			exp.Status = "critical"
		} else if exp.UtilizationPct > 75 {
			exp.Status = "warning"
		} else {
			exp.Status = "normal"
		}

		result = append(result, *exp)
	}

	return result
}

func (h *APIHandler) calculateNotionalValue(pos *core.Position) float64 {
	// Get current price
	price := pos.CurrentPrice
	if price == 0 {
		price = pos.OpenPrice
	}

	// Notional = Volume * Price * ContractSize
	// For forex, contract size is typically 100,000 (1 lot)
	contractSize := 100000.0

	// Try to get spec from engine's symbols
	symbols := h.engine.GetSymbols()
	for _, spec := range symbols {
		if spec.Symbol == pos.Symbol && spec.ContractSize > 0 {
			contractSize = spec.ContractSize
			break
		}
	}

	return pos.Volume * price * contractSize
}

func (h *APIHandler) parseInterval(interval string) time.Duration {
	switch interval {
	case "15m":
		return 15 * time.Minute
	case "1h":
		return 1 * time.Hour
	case "4h":
		return 4 * time.Hour
	case "1d":
		return 24 * time.Hour
	default:
		return 1 * time.Hour
	}
}
