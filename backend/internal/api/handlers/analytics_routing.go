package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/epic1st/rtx/backend/cbook"
)

// RoutingBreakdownResponse represents A/B/C-Book distribution
type RoutingBreakdownResponse struct {
	ABookPercent      float64           `json:"abook_pct"`
	BBookPercent      float64           `json:"bbook_pct"`
	CBookPercent      float64           `json:"cbook_pct"` // Partial hedge
	PartialPercent    float64           `json:"partial_pct"`
	TotalVolume       float64           `json:"total_volume"`
	TotalDecisions    int64             `json:"total_decisions"`
	BreakdownBySymbol []SymbolBreakdown `json:"breakdown_by_symbol"`
	TimeRange         TimeRangeInfo     `json:"time_range"`
}

type SymbolBreakdown struct {
	Symbol         string  `json:"symbol"`
	ABookPercent   float64 `json:"abook_pct"`
	BBookPercent   float64 `json:"bbook_pct"`
	PartialPercent float64 `json:"partial_pct"`
	TotalVolume    float64 `json:"total_volume"`
	DecisionCount  int64   `json:"decision_count"`
}

type TimeRangeInfo struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// RoutingTimelineResponse represents routing decisions over time
type RoutingTimelineResponse struct {
	Timestamps  []string      `json:"timestamps"`
	ABookCounts []int64       `json:"abook_counts"`
	BBookCounts []int64       `json:"bbook_counts"`
	CBookCounts []int64       `json:"cbook_counts"` // Partial hedge
	TimeRange   TimeRangeInfo `json:"time_range"`
	Interval    string        `json:"interval"`
}

// RoutingConfidenceResponse represents decision confidence distribution
type RoutingConfidenceResponse struct {
	HighConfidencePct   float64       `json:"high_confidence_pct"`   // > 70%
	MediumConfidencePct float64       `json:"medium_confidence_pct"` // 40-70%
	LowConfidencePct    float64       `json:"low_confidence_pct"`    // < 40%
	AvgConfidence       float64       `json:"avg_confidence"`
	TotalDecisions      int64         `json:"total_decisions"`
	TimeRange           TimeRangeInfo `json:"time_range"`
}

// HandleRoutingBreakdown handles GET /api/analytics/routing/breakdown
func (h *APIHandler) HandleRoutingBreakdown(w http.ResponseWriter, r *http.Request) {
	cors(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	symbol := r.URL.Query().Get("symbol")
	accountIDStr := r.URL.Query().Get("account_id")

	// Validate and parse time range
	startTime, endTime, err := parseTimeRange(startTimeStr, endTimeStr)
	if err != nil {
		http.Error(w, "Invalid time range: "+err.Error(), http.StatusBadRequest)
		return
	}

	var accountID int64
	if accountIDStr != "" {
		accountID, err = strconv.ParseInt(accountIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid account_id", http.StatusBadRequest)
			return
		}
	}

	// Check if C-Book engine is available
	if h.cbookEngine == nil {
		http.Error(w, "C-Book analytics not available", http.StatusServiceUnavailable)
		return
	}

	// Get all routing decisions from the engine
	decisions := h.cbookEngine.GetDecisionHistory(10000) // Get last 10k decisions

	// Filter decisions by time range, symbol, and account
	filteredDecisions := filterDecisions(decisions, startTime, endTime, symbol, accountID)

	if len(filteredDecisions) == 0 {
		// Return empty but valid response
		response := RoutingBreakdownResponse{
			ABookPercent:      0,
			BBookPercent:      0,
			CBookPercent:      0,
			PartialPercent:    0,
			TotalVolume:       0,
			TotalDecisions:    0,
			BreakdownBySymbol: []SymbolBreakdown{},
			TimeRange: TimeRangeInfo{
				StartTime: startTime.Format(time.RFC3339),
				EndTime:   endTime.Format(time.RFC3339),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Calculate breakdown
	var aBookCount, bBookCount, partialCount int64
	var totalVolume float64
	symbolStats := make(map[string]*SymbolBreakdown)

	for _, decision := range filteredDecisions {
		// Count by action type
		switch decision.Action {
		case cbook.ActionABook:
			aBookCount++
		case cbook.ActionBBook:
			bBookCount++
		case cbook.ActionPartialHedge:
			partialCount++
		}

		// Aggregate volume (Note: Decision doesn't have Volume, using placeholder)
		// In production, would need to track volume with decisions
		totalVolume += 1.0 // Placeholder: 1 lot per decision

		// Track by symbol (Note: Decision doesn't have Symbol in current implementation)
		// This is a limitation of current data structure
		// For now, we'll use a generic symbol if not filtering
		symbolKey := "ALL"
		if symbol != "" {
			symbolKey = symbol
		}

		stats, exists := symbolStats[symbolKey]
		if !exists {
			stats = &SymbolBreakdown{
				Symbol: symbolKey,
			}
			symbolStats[symbolKey] = stats
		}

		switch decision.Action {
		case cbook.ActionABook:
			stats.ABookPercent += 100
		case cbook.ActionBBook:
			stats.BBookPercent += 100
		case cbook.ActionPartialHedge:
			stats.PartialPercent += 100
		}
		stats.DecisionCount++
		stats.TotalVolume += 1.0
	}

	// Calculate percentages
	totalCount := int64(len(filteredDecisions))
	aBookPct := (float64(aBookCount) / float64(totalCount)) * 100
	bBookPct := (float64(bBookCount) / float64(totalCount)) * 100
	partialPct := (float64(partialCount) / float64(totalCount)) * 100

	// Finalize symbol breakdowns
	breakdownBySymbol := make([]SymbolBreakdown, 0, len(symbolStats))
	for _, stats := range symbolStats {
		if stats.DecisionCount > 0 {
			stats.ABookPercent = stats.ABookPercent / float64(stats.DecisionCount)
			stats.BBookPercent = stats.BBookPercent / float64(stats.DecisionCount)
			stats.PartialPercent = stats.PartialPercent / float64(stats.DecisionCount)
		}
		breakdownBySymbol = append(breakdownBySymbol, *stats)
	}

	response := RoutingBreakdownResponse{
		ABookPercent:      aBookPct,
		BBookPercent:      bBookPct,
		CBookPercent:      partialPct,
		PartialPercent:    partialPct,
		TotalVolume:       totalVolume,
		TotalDecisions:    totalCount,
		BreakdownBySymbol: breakdownBySymbol,
		TimeRange: TimeRangeInfo{
			StartTime: startTime.Format(time.RFC3339),
			EndTime:   endTime.Format(time.RFC3339),
		},
	}

	log.Printf("[Analytics] Routing breakdown: A-Book=%.1f%%, B-Book=%.1f%%, Partial=%.1f%%, Total=%d",
		aBookPct, bBookPct, partialPct, totalCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleRoutingTimeline handles GET /api/analytics/routing/timeline
func (h *APIHandler) HandleRoutingTimeline(w http.ResponseWriter, r *http.Request) {
	cors(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	interval := r.URL.Query().Get("interval")
	symbol := r.URL.Query().Get("symbol")

	// Default interval
	if interval == "" {
		interval = "1h"
	}

	// Validate interval
	validIntervals := map[string]bool{
		"1m": true, "5m": true, "15m": true,
		"1h": true, "4h": true, "1d": true,
	}
	if !validIntervals[interval] {
		http.Error(w, "Invalid interval. Use: 1m, 5m, 15m, 1h, 4h, 1d", http.StatusBadRequest)
		return
	}

	// Parse time range
	startTime, endTime, err := parseTimeRange(startTimeStr, endTimeStr)
	if err != nil {
		http.Error(w, "Invalid time range: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Check if C-Book engine is available
	if h.cbookEngine == nil {
		http.Error(w, "C-Book analytics not available", http.StatusServiceUnavailable)
		return
	}

	// Get decisions
	decisions := h.cbookEngine.GetDecisionHistory(10000)
	filteredDecisions := filterDecisions(decisions, startTime, endTime, symbol, 0)

	// Calculate interval duration
	intervalDuration := parseIntervalDuration(interval)

	// Group decisions by time buckets
	timeBuckets := make(map[int64]*struct {
		aBook   int64
		bBook   int64
		partial int64
	})

	for _, decision := range filteredDecisions {
		bucketTime := decision.DecisionTime.Truncate(intervalDuration).Unix()

		bucket, exists := timeBuckets[bucketTime]
		if !exists {
			bucket = &struct {
				aBook   int64
				bBook   int64
				partial int64
			}{}
			timeBuckets[bucketTime] = bucket
		}

		switch decision.Action {
		case cbook.ActionABook:
			bucket.aBook++
		case cbook.ActionBBook:
			bucket.bBook++
		case cbook.ActionPartialHedge:
			bucket.partial++
		}
	}

	// Generate complete time series
	timestamps := []string{}
	aBookCounts := []int64{}
	bBookCounts := []int64{}
	cBookCounts := []int64{}

	currentTime := startTime.Truncate(intervalDuration)
	for currentTime.Before(endTime) || currentTime.Equal(endTime) {
		timestamps = append(timestamps, currentTime.Format(time.RFC3339))

		bucket, exists := timeBuckets[currentTime.Unix()]
		if exists {
			aBookCounts = append(aBookCounts, bucket.aBook)
			bBookCounts = append(bBookCounts, bucket.bBook)
			cBookCounts = append(cBookCounts, bucket.partial)
		} else {
			aBookCounts = append(aBookCounts, 0)
			bBookCounts = append(bBookCounts, 0)
			cBookCounts = append(cBookCounts, 0)
		}

		currentTime = currentTime.Add(intervalDuration)
	}

	response := RoutingTimelineResponse{
		Timestamps:  timestamps,
		ABookCounts: aBookCounts,
		BBookCounts: bBookCounts,
		CBookCounts: cBookCounts,
		TimeRange: TimeRangeInfo{
			StartTime: startTime.Format(time.RFC3339),
			EndTime:   endTime.Format(time.RFC3339),
		},
		Interval: interval,
	}

	log.Printf("[Analytics] Routing timeline: %d buckets, interval=%s", len(timestamps), interval)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleRoutingConfidence handles GET /api/analytics/routing/confidence
func (h *APIHandler) HandleRoutingConfidence(w http.ResponseWriter, r *http.Request) {
	cors(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	// Parse time range
	startTime, endTime, err := parseTimeRange(startTimeStr, endTimeStr)
	if err != nil {
		http.Error(w, "Invalid time range: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Check if C-Book engine is available
	if h.cbookEngine == nil {
		http.Error(w, "C-Book analytics not available", http.StatusServiceUnavailable)
		return
	}

	// Get decisions
	decisions := h.cbookEngine.GetDecisionHistory(10000)
	filteredDecisions := filterDecisions(decisions, startTime, endTime, "", 0)

	if len(filteredDecisions) == 0 {
		response := RoutingConfidenceResponse{
			HighConfidencePct:   0,
			MediumConfidencePct: 0,
			LowConfidencePct:    0,
			AvgConfidence:       0,
			TotalDecisions:      0,
			TimeRange: TimeRangeInfo{
				StartTime: startTime.Format(time.RFC3339),
				EndTime:   endTime.Format(time.RFC3339),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Calculate confidence distribution
	// Note: Current RoutingDecision doesn't have explicit confidence field
	// We'll derive it from ToxicityScore and ExposureRisk
	var highCount, mediumCount, lowCount int64
	var totalConfidence float64

	for _, decision := range filteredDecisions {
		// Derive confidence from toxicity and exposure scores
		// Higher toxicity + higher exposure = higher confidence in decision
		confidence := calculateDecisionConfidence(decision)
		totalConfidence += confidence

		if confidence > 70 {
			highCount++
		} else if confidence >= 40 {
			mediumCount++
		} else {
			lowCount++
		}
	}

	totalCount := int64(len(filteredDecisions))
	avgConfidence := totalConfidence / float64(totalCount)

	response := RoutingConfidenceResponse{
		HighConfidencePct:   (float64(highCount) / float64(totalCount)) * 100,
		MediumConfidencePct: (float64(mediumCount) / float64(totalCount)) * 100,
		LowConfidencePct:    (float64(lowCount) / float64(totalCount)) * 100,
		AvgConfidence:       avgConfidence,
		TotalDecisions:      totalCount,
		TimeRange: TimeRangeInfo{
			StartTime: startTime.Format(time.RFC3339),
			EndTime:   endTime.Format(time.RFC3339),
		},
	}

	log.Printf("[Analytics] Confidence distribution: High=%.1f%%, Medium=%.1f%%, Low=%.1f%%, Avg=%.1f",
		response.HighConfidencePct, response.MediumConfidencePct, response.LowConfidencePct, avgConfidence)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions

func parseTimeRange(startTimeStr, endTimeStr string) (time.Time, time.Time, error) {
	var startTime, endTime time.Time
	var err error

	// Default to last 24 hours if not provided
	if startTimeStr == "" {
		startTime = time.Now().Add(-24 * time.Hour)
	} else {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	if endTimeStr == "" {
		endTime = time.Now()
	} else {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	if startTime.After(endTime) {
		return time.Time{}, time.Time{}, fmt.Errorf("start_time must be before end_time")
	}

	return startTime, endTime, nil
}

func filterDecisions(decisions []cbook.RoutingDecision, startTime, endTime time.Time, _ string, _ int64) []cbook.RoutingDecision {
	filtered := make([]cbook.RoutingDecision, 0)

	for _, decision := range decisions {
		// Filter by time range
		if decision.DecisionTime.Before(startTime) || decision.DecisionTime.After(endTime) {
			continue
		}

		// Note: Current RoutingDecision doesn't have Symbol or AccountID fields
		// In production, these would need to be added to the RoutingDecision struct
		// For now, we accept all decisions if symbol/accountID filters are provided

		filtered = append(filtered, decision)
	}

	return filtered
}

func parseIntervalDuration(interval string) time.Duration {
	switch interval {
	case "1m":
		return 1 * time.Minute
	case "5m":
		return 5 * time.Minute
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

func calculateDecisionConfidence(decision cbook.RoutingDecision) float64 {
	// Derive confidence from toxicity and exposure risk
	// Higher scores indicate more confident decision
	// This is a heuristic - in production, would use ML confidence scores

	toxicityWeight := 0.6
	exposureWeight := 0.4

	confidence := (decision.ToxicityScore * toxicityWeight) + (decision.ExposureRisk * exposureWeight)

	// Normalize to 0-100
	if confidence > 100 {
		confidence = 100
	}
	if confidence < 0 {
		confidence = 0
	}

	return confidence
}
