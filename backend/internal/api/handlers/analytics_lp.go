package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// AnalyticsLPHandler handles LP performance analytics API requests
type AnalyticsLPHandler struct {
	db *sql.DB
}

// NewAnalyticsLPHandler creates a new analytics LP handler
func NewAnalyticsLPHandler() (*AnalyticsLPHandler, error) {
	connStr := getConnectionString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("[Analytics] LP analytics handler initialized with database connection")

	return &AnalyticsLPHandler{
		db: db,
	}, nil
}

// Close closes the database connection
func (h *AnalyticsLPHandler) Close() error {
	if h.db != nil {
		return h.db.Close()
	}
	return nil
}

// LPComparisonResponse represents the response for LP comparison
type LPComparisonResponse struct {
	LPs []LPMetrics `json:"lps"`
}

// LPMetrics represents performance metrics for a single LP
type LPMetrics struct {
	Name          string  `json:"name"`
	AvgLatencyMS  float64 `json:"avg_latency_ms"`
	FillRatePct   float64 `json:"fill_rate_pct"`
	SlippageBPS   int     `json:"slippage_bps"`
	Volume24h     float64 `json:"volume_24h"`
	Rank          int     `json:"rank"`
}

// LPPerformanceResponse represents detailed performance for a single LP
type LPPerformanceResponse struct {
	LPName   string              `json:"lp_name"`
	Metrics  LPDetailedMetrics   `json:"metrics"`
	Timeline []LPTimelineMetrics `json:"timeline"`
}

// LPDetailedMetrics represents detailed performance metrics
type LPDetailedMetrics struct {
	LatencyP50  float64 `json:"latency_p50"`
	LatencyP95  float64 `json:"latency_p95"`
	LatencyP99  float64 `json:"latency_p99"`
	FillRate    float64 `json:"fill_rate"`
	AvgSlippage int     `json:"avg_slippage"`
	UptimePct   float64 `json:"uptime_pct"`
}

// LPTimelineMetrics represents metrics over time
type LPTimelineMetrics struct {
	Timestamp   time.Time `json:"timestamp"`
	AvgLatency  float64   `json:"avg_latency"`
	FillRate    float64   `json:"fill_rate"`
	Volume      float64   `json:"volume"`
	OrderCount  int       `json:"order_count"`
}

// LPRankingResponse represents LP ranking by metric
type LPRankingResponse struct {
	Rankings []LPRanking `json:"rankings"`
}

// LPRanking represents a single LP ranking entry
type LPRanking struct {
	Rank       int     `json:"rank"`
	LPName     string  `json:"lp_name"`
	Value      float64 `json:"value"`
	Percentile float64 `json:"percentile"`
}

// HandleLPComparison handles GET /api/analytics/lp/comparison
func (h *AnalyticsLPHandler) HandleLPComparison(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	symbol := r.URL.Query().Get("symbol")
	metric := r.URL.Query().Get("metric")

	// Default to last 24 hours if not specified
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	if startTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = parsed
		}
	}

	if endTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = parsed
		}
	}

	// Default metric is latency
	if metric == "" {
		metric = "latency"
	}

	// Validate metric
	validMetrics := map[string]bool{
		"latency":    true,
		"fill_rate":  true,
		"slippage":   true,
	}

	if !validMetrics[metric] {
		http.Error(w, `{"error":"invalid metric. Must be one of: latency, fill_rate, slippage"}`, http.StatusBadRequest)
		return
	}

	// Build query
	query := `
		WITH lp_aggregates AS (
			SELECT
				lp.name,
				AVG(lor.execution_latency_ms) AS avg_latency_ms,
				ROUND(
					(SUM(CASE WHEN lor.success = true THEN 1 ELSE 0 END)::DECIMAL /
					NULLIF(COUNT(*), 0) * 100)::NUMERIC,
					2
				) AS fill_rate_pct,
				COALESCE(AVG(ABS(lor.actual_slippage)), 0) AS avg_slippage_bps,
				COALESCE(SUM(CASE WHEN o.status = 'FILLED' THEN o.volume ELSE 0 END), 0) AS volume_24h
			FROM liquidity_providers lp
			LEFT JOIN lp_order_routing lor ON lp.id = lor.lp_id
			LEFT JOIN orders o ON lor.order_id = o.id
			WHERE
				lor.created_at >= $1
				AND lor.created_at <= $2
				` + buildSymbolFilter(symbol) + `
			GROUP BY lp.id, lp.name
		)
		SELECT
			name,
			COALESCE(avg_latency_ms, 0) AS avg_latency_ms,
			COALESCE(fill_rate_pct, 0) AS fill_rate_pct,
			COALESCE(avg_slippage_bps, 0) AS slippage_bps,
			COALESCE(volume_24h, 0) AS volume_24h,
			` + buildRankExpression(metric) + ` AS rank
		FROM lp_aggregates
		ORDER BY ` + buildOrderExpression(metric)

	rows, err := h.db.Query(query, startTime, endTime)
	if err != nil {
		log.Printf("[Analytics] Query error: %v", err)
		http.Error(w, fmt.Sprintf(`{"error":"database query failed: %v"}`, err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	lps := make([]LPMetrics, 0)
	for rows.Next() {
		var lp LPMetrics
		var slippageBPS float64

		err := rows.Scan(
			&lp.Name,
			&lp.AvgLatencyMS,
			&lp.FillRatePct,
			&slippageBPS,
			&lp.Volume24h,
			&lp.Rank,
		)
		if err != nil {
			log.Printf("[Analytics] Scan error: %v", err)
			continue
		}

		lp.SlippageBPS = int(slippageBPS)
		lps = append(lps, lp)
	}

	response := LPComparisonResponse{
		LPs: lps,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleLPPerformance handles GET /api/analytics/lp/performance/{lp_name}
func (h *AnalyticsLPHandler) HandleLPPerformance(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract LP name from path
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/analytics/lp/performance/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, `{"error":"LP name is required"}`, http.StatusBadRequest)
		return
	}
	lpName := pathParts[0]

	// Parse query parameters
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	symbol := r.URL.Query().Get("symbol")

	// Default to last 24 hours if not specified
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	if startTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = parsed
		}
	}

	if endTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = parsed
		}
	}

	// Query for detailed metrics with percentiles
	metricsQuery := `
		WITH lp_data AS (
			SELECT
				lp.name,
				lor.execution_latency_ms,
				lor.success,
				lor.actual_slippage
			FROM liquidity_providers lp
			LEFT JOIN lp_order_routing lor ON lp.id = lor.lp_id
			LEFT JOIN orders o ON lor.order_id = o.id
			WHERE
				lp.name = $1
				AND lor.created_at >= $2
				AND lor.created_at <= $3
				` + buildSymbolFilter(symbol) + `
		),
		percentiles AS (
			SELECT
				PERCENTILE_CONT(0.50) WITHIN GROUP (ORDER BY execution_latency_ms) AS p50,
				PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY execution_latency_ms) AS p95,
				PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY execution_latency_ms) AS p99
			FROM lp_data
			WHERE execution_latency_ms IS NOT NULL
		),
		aggregates AS (
			SELECT
				ROUND(
					(SUM(CASE WHEN success = true THEN 1 ELSE 0 END)::DECIMAL /
					NULLIF(COUNT(*), 0) * 100)::NUMERIC,
					2
				) AS fill_rate,
				COALESCE(AVG(ABS(actual_slippage)), 0) AS avg_slippage
			FROM lp_data
		),
		uptime AS (
			SELECT
				COALESCE(
					(SUM(uptime_seconds)::DECIMAL /
					NULLIF(SUM(uptime_seconds + downtime_seconds), 0) * 100)::NUMERIC,
					100
				) AS uptime_pct
			FROM lp_performance_metrics lpm
			JOIN liquidity_providers lp ON lpm.lp_id = lp.id
			WHERE
				lp.name = $1
				AND lpm.time_bucket >= $2
				AND lpm.time_bucket <= $3
		)
		SELECT
			COALESCE(p.p50, 0) AS latency_p50,
			COALESCE(p.p95, 0) AS latency_p95,
			COALESCE(p.p99, 0) AS latency_p99,
			COALESCE(a.fill_rate, 0) AS fill_rate,
			COALESCE(a.avg_slippage, 0) AS avg_slippage,
			COALESCE(u.uptime_pct, 100) AS uptime_pct
		FROM percentiles p
		CROSS JOIN aggregates a
		CROSS JOIN uptime u
	`

	var metrics LPDetailedMetrics
	var avgSlippage float64

	err := h.db.QueryRow(metricsQuery, lpName, startTime, endTime).Scan(
		&metrics.LatencyP50,
		&metrics.LatencyP95,
		&metrics.LatencyP99,
		&metrics.FillRate,
		&avgSlippage,
		&metrics.UptimePct,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"LP not found or no data available"}`, http.StatusNotFound)
		} else {
			log.Printf("[Analytics] Metrics query error: %v", err)
			http.Error(w, fmt.Sprintf(`{"error":"database query failed: %v"}`, err), http.StatusInternalServerError)
		}
		return
	}

	metrics.AvgSlippage = int(avgSlippage)

	// Query for timeline data
	timelineQuery := `
		SELECT
			lpm.time_bucket,
			COALESCE(lpm.avg_fill_time_ms, 0) AS avg_latency,
			ROUND(
				(lpm.successful_orders::DECIMAL /
				NULLIF(lpm.total_orders, 0) * 100)::NUMERIC,
				2
			) AS fill_rate,
			COALESCE(lpm.total_volume, 0) AS volume,
			lpm.total_orders
		FROM lp_performance_metrics lpm
		JOIN liquidity_providers lp ON lpm.lp_id = lp.id
		WHERE
			lp.name = $1
			AND lpm.time_bucket >= $2
			AND lpm.time_bucket <= $3
			` + buildSymbolFilter(symbol) + `
		ORDER BY lpm.time_bucket ASC
	`

	timelineRows, err := h.db.Query(timelineQuery, lpName, startTime, endTime)
	if err != nil {
		log.Printf("[Analytics] Timeline query error: %v", err)
		http.Error(w, fmt.Sprintf(`{"error":"timeline query failed: %v"}`, err), http.StatusInternalServerError)
		return
	}
	defer timelineRows.Close()

	timeline := make([]LPTimelineMetrics, 0)
	for timelineRows.Next() {
		var t LPTimelineMetrics
		err := timelineRows.Scan(
			&t.Timestamp,
			&t.AvgLatency,
			&t.FillRate,
			&t.Volume,
			&t.OrderCount,
		)
		if err != nil {
			log.Printf("[Analytics] Timeline scan error: %v", err)
			continue
		}
		timeline = append(timeline, t)
	}

	response := LPPerformanceResponse{
		LPName:   lpName,
		Metrics:  metrics,
		Timeline: timeline,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleLPRanking handles GET /api/analytics/lp/ranking
func (h *AnalyticsLPHandler) HandleLPRanking(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	metric := r.URL.Query().Get("metric")
	limitStr := r.URL.Query().Get("limit")

	// Default to last 24 hours if not specified
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	if startTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = parsed
		}
	}

	if endTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = parsed
		}
	}

	// Default metric is latency
	if metric == "" {
		metric = "latency"
	}

	// Validate metric
	validMetrics := map[string]bool{
		"latency":    true,
		"fill_rate":  true,
		"slippage":   true,
		"volume":     true,
	}

	if !validMetrics[metric] {
		http.Error(w, `{"error":"invalid metric. Must be one of: latency, fill_rate, slippage, volume"}`, http.StatusBadRequest)
		return
	}

	// Parse limit (default 10)
	limit := 10
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// Build query with ranking and percentile calculation
	query := `
		WITH lp_metrics AS (
			SELECT
				lp.name,
				AVG(lor.execution_latency_ms) AS avg_latency,
				ROUND(
					(SUM(CASE WHEN lor.success = true THEN 1 ELSE 0 END)::DECIMAL /
					NULLIF(COUNT(*), 0) * 100)::NUMERIC,
					2
				) AS fill_rate,
				COALESCE(AVG(ABS(lor.actual_slippage)), 0) AS avg_slippage,
				COALESCE(SUM(CASE WHEN o.status = 'FILLED' THEN o.volume ELSE 0 END), 0) AS total_volume
			FROM liquidity_providers lp
			LEFT JOIN lp_order_routing lor ON lp.id = lor.lp_id
			LEFT JOIN orders o ON lor.order_id = o.id
			WHERE
				lor.created_at >= $1
				AND lor.created_at <= $2
			GROUP BY lp.id, lp.name
		),
		ranked AS (
			SELECT
				name,
				` + buildMetricValueExpression(metric) + ` AS value,
				` + buildRankExpression(metric) + ` AS rank,
				COUNT(*) OVER() AS total_count
			FROM lp_metrics
		)
		SELECT
			rank,
			name,
			COALESCE(value, 0) AS value,
			ROUND((100.0 - (rank::DECIMAL / total_count * 100))::NUMERIC, 2) AS percentile
		FROM ranked
		ORDER BY rank ASC
		LIMIT $3
	`

	rows, err := h.db.Query(query, startTime, endTime, limit)
	if err != nil {
		log.Printf("[Analytics] Ranking query error: %v", err)
		http.Error(w, fmt.Sprintf(`{"error":"database query failed: %v"}`, err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	rankings := make([]LPRanking, 0)
	for rows.Next() {
		var r LPRanking
		err := rows.Scan(
			&r.Rank,
			&r.LPName,
			&r.Value,
			&r.Percentile,
		)
		if err != nil {
			log.Printf("[Analytics] Ranking scan error: %v", err)
			continue
		}
		rankings = append(rankings, r)
	}

	response := LPRankingResponse{
		Rankings: rankings,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func buildSymbolFilter(symbol string) string {
	if symbol != "" {
		return fmt.Sprintf(" AND o.symbol = '%s'", symbol)
	}
	return ""
}

func buildRankExpression(metric string) string {
	switch metric {
	case "latency":
		return "ROW_NUMBER() OVER (ORDER BY avg_latency_ms ASC NULLS LAST)"
	case "fill_rate":
		return "ROW_NUMBER() OVER (ORDER BY fill_rate_pct DESC NULLS LAST)"
	case "slippage":
		return "ROW_NUMBER() OVER (ORDER BY avg_slippage_bps ASC NULLS LAST)"
	case "volume":
		return "ROW_NUMBER() OVER (ORDER BY volume_24h DESC NULLS LAST)"
	default:
		return "ROW_NUMBER() OVER (ORDER BY avg_latency_ms ASC NULLS LAST)"
	}
}

func buildOrderExpression(metric string) string {
	switch metric {
	case "latency":
		return "avg_latency_ms ASC NULLS LAST"
	case "fill_rate":
		return "fill_rate_pct DESC NULLS LAST"
	case "slippage":
		return "slippage_bps ASC NULLS LAST"
	case "volume":
		return "volume_24h DESC NULLS LAST"
	default:
		return "avg_latency_ms ASC NULLS LAST"
	}
}

func buildMetricValueExpression(metric string) string {
	switch metric {
	case "latency":
		return "avg_latency"
	case "fill_rate":
		return "fill_rate"
	case "slippage":
		return "avg_slippage"
	case "volume":
		return "total_volume"
	default:
		return "avg_latency"
	}
}

func getConnectionString() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "trading_engine")
	sslmode := getEnv("DB_SSLMODE", "disable")

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
