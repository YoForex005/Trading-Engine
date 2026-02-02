package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// DataSourceVerificationHandler handles data source verification endpoints
type DataSourceVerificationHandler struct {
	db *sql.DB
}

// NewDataSourceVerificationHandler creates a new data source verification handler
func NewDataSourceVerificationHandler(db *sql.DB) *DataSourceVerificationHandler {
	return &DataSourceVerificationHandler{
		db: db,
	}
}

// HandleDataSourceHealth provides a health check endpoint for data source verification
func (h *DataSourceVerificationHandler) HandleDataSourceHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	response := h.checkDataSources()

	// Return 500 if simulation detected in production
	if response["has_simulation"].(bool) && os.Getenv("ENVIRONMENT") == "production" {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}

// checkDataSources performs comprehensive data source verification
func (h *DataSourceVerificationHandler) checkDataSources() map[string]interface{} {
	response := map[string]interface{}{
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
		"environment":     os.Getenv("ENVIRONMENT"),
		"allow_simulation": os.Getenv("ALLOW_SIMULATION"),
		"status":          "healthy",
	}

	// Check if database is available
	if h.db == nil {
		response["status"] = "degraded"
		response["error"] = "Database not available"
		response["checks"] = []string{}
		return response
	}

	checks := make([]map[string]interface{}, 0)

	// Check 1: Scan for simulation data
	simCheck := h.checkForSimulationData()
	checks = append(checks, simCheck)

	if simCheck["has_simulation"].(bool) {
		response["has_simulation"] = true
		response["status"] = "critical"
		if os.Getenv("ENVIRONMENT") == "production" {
			response["alert"] = "PRODUCTION ALERT: Simulation data detected in database"
		}
	} else {
		response["has_simulation"] = false
	}

	// Check 2: Verify real LP data exists
	lpCheck := h.checkRealLPData()
	checks = append(checks, lpCheck)

	// Check 3: Get LP distribution
	lpDist := h.getLPDistribution()
	response["lp_distribution"] = lpDist

	// Check 4: Check data freshness
	freshnessCheck := h.checkDataFreshness()
	checks = append(checks, freshnessCheck)

	response["checks"] = checks

	return response
}

// checkForSimulationData scans database for simulation ticks
func (h *DataSourceVerificationHandler) checkForSimulationData() map[string]interface{} {
	check := map[string]interface{}{
		"name":   "simulation_data_check",
		"status": "pass",
	}

	var simCount int64
	err := h.db.QueryRow(`
		SELECT COUNT(*)
		FROM ticks
		WHERE lp_source = 'SIM' OR lp_source = 'SIMULATION'
	`).Scan(&simCount)

	if err != nil {
		check["status"] = "error"
		check["error"] = err.Error()
		check["has_simulation"] = false
		return check
	}

	check["simulation_tick_count"] = simCount
	check["has_simulation"] = simCount > 0

	if simCount > 0 {
		check["status"] = "fail"
		check["message"] = fmt.Sprintf("Found %d simulation ticks in database", simCount)

		// Get latest simulation tick
		var latestSimTimestamp int64
		err = h.db.QueryRow(`
			SELECT MAX(timestamp)
			FROM ticks
			WHERE lp_source = 'SIM' OR lp_source = 'SIMULATION'
		`).Scan(&latestSimTimestamp)

		if err == nil {
			check["latest_sim_tick"] = time.Unix(0, latestSimTimestamp*int64(time.Millisecond)).Format(time.RFC3339)
		}
	} else {
		check["message"] = "No simulation data found - production safe"
	}

	return check
}

// checkRealLPData verifies real LP data exists
func (h *DataSourceVerificationHandler) checkRealLPData() map[string]interface{} {
	check := map[string]interface{}{
		"name":   "real_lp_data_check",
		"status": "pass",
	}

	var realLPCount int64
	err := h.db.QueryRow(`
		SELECT COUNT(*)
		FROM ticks
		WHERE lp_source != 'SIM' AND lp_source != 'SIMULATION' AND lp_source IS NOT NULL
	`).Scan(&realLPCount)

	if err != nil {
		check["status"] = "error"
		check["error"] = err.Error()
		return check
	}

	check["real_lp_tick_count"] = realLPCount

	if realLPCount == 0 {
		check["status"] = "warn"
		check["message"] = "No real LP data found in database"
	} else {
		check["message"] = fmt.Sprintf("Found %d real LP ticks", realLPCount)
	}

	return check
}

// getLPDistribution returns tick count by LP source
func (h *DataSourceVerificationHandler) getLPDistribution() []map[string]interface{} {
	rows, err := h.db.Query(`
		SELECT
			lp_source,
			COUNT(*) as tick_count,
			MIN(timestamp) as first_tick,
			MAX(timestamp) as last_tick
		FROM ticks
		GROUP BY lp_source
		ORDER BY tick_count DESC
		LIMIT 20
	`)
	if err != nil {
		log.Printf("[DataSourceVerification] Error getting LP distribution: %v", err)
		return []map[string]interface{}{}
	}
	defer rows.Close()

	distribution := make([]map[string]interface{}, 0)

	for rows.Next() {
		var lpSource string
		var tickCount int64
		var firstTick, lastTick int64

		if err := rows.Scan(&lpSource, &tickCount, &firstTick, &lastTick); err != nil {
			continue
		}

		distribution = append(distribution, map[string]interface{}{
			"lp_source":   lpSource,
			"tick_count":  tickCount,
			"first_tick":  time.Unix(0, firstTick*int64(time.Millisecond)).Format(time.RFC3339),
			"last_tick":   time.Unix(0, lastTick*int64(time.Millisecond)).Format(time.RFC3339),
		})
	}

	return distribution
}

// checkDataFreshness verifies recent data flow
func (h *DataSourceVerificationHandler) checkDataFreshness() map[string]interface{} {
	check := map[string]interface{}{
		"name":   "data_freshness_check",
		"status": "pass",
	}

	var latestTimestamp int64
	err := h.db.QueryRow(`
		SELECT MAX(timestamp)
		FROM ticks
	`).Scan(&latestTimestamp)

	if err != nil {
		check["status"] = "error"
		check["error"] = err.Error()
		return check
	}

	latestTime := time.Unix(0, latestTimestamp*int64(time.Millisecond))
	check["latest_tick_time"] = latestTime.Format(time.RFC3339)

	age := time.Since(latestTime)
	check["data_age_seconds"] = int64(age.Seconds())

	// Warn if data is older than 5 minutes
	if age > 5*time.Minute {
		check["status"] = "warn"
		check["message"] = fmt.Sprintf("Latest tick is %v old", age.Round(time.Second))
	} else {
		check["message"] = fmt.Sprintf("Data is fresh (age: %v)", age.Round(time.Second))
	}

	return check
}

// HandleCleanupSimulationData provides endpoint to identify and tag simulation data
func (h *DataSourceVerificationHandler) HandleCleanupSimulationData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req struct {
		DryRun bool `json:"dryRun"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result := h.cleanupSimulationData(req.DryRun)

	json.NewEncoder(w).Encode(result)
}

// cleanupSimulationData tags or removes simulation data
func (h *DataSourceVerificationHandler) cleanupSimulationData(dryRun bool) map[string]interface{} {
	result := map[string]interface{}{
		"dry_run":   dryRun,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// Count simulation ticks
	var simCount int64
	err := h.db.QueryRow(`
		SELECT COUNT(*)
		FROM ticks
		WHERE lp_source = 'SIM' OR lp_source = 'SIMULATION'
	`).Scan(&simCount)

	if err != nil {
		result["error"] = err.Error()
		return result
	}

	result["simulation_ticks_found"] = simCount

	if simCount == 0 {
		result["message"] = "No simulation data found"
		return result
	}

	if dryRun {
		result["message"] = fmt.Sprintf("DRY RUN: Would tag %d simulation ticks", simCount)
		result["action"] = "none"
	} else {
		// Tag simulation data with a flag
		_, err := h.db.Exec(`
			UPDATE ticks
			SET flags = flags | 1  -- Set bit 0 to indicate simulation data
			WHERE (lp_source = 'SIM' OR lp_source = 'SIMULATION')
			AND (flags & 1) = 0    -- Only if not already flagged
		`)

		if err != nil {
			result["error"] = err.Error()
			return result
		}

		result["message"] = fmt.Sprintf("Tagged %d simulation ticks (flags |= 1)", simCount)
		result["action"] = "tagged"

		log.Printf("[DataSourceVerification] Tagged %d simulation ticks in database", simCount)
	}

	return result
}

// HandleGetSimulationReport provides detailed report on simulation data
func (h *DataSourceVerificationHandler) HandleGetSimulationReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	report := h.generateSimulationReport()
	json.NewEncoder(w).Encode(report)
}

// generateSimulationReport creates a detailed report of simulation data
func (h *DataSourceVerificationHandler) generateSimulationReport() map[string]interface{} {
	report := map[string]interface{}{
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"environment": os.Getenv("ENVIRONMENT"),
	}

	// Get simulation data by symbol
	rows, err := h.db.Query(`
		SELECT
			symbol,
			COUNT(*) as tick_count,
			MIN(timestamp) as first_tick,
			MAX(timestamp) as last_tick
		FROM ticks
		WHERE lp_source = 'SIM' OR lp_source = 'SIMULATION'
		GROUP BY symbol
		ORDER BY tick_count DESC
	`)
	if err != nil {
		report["error"] = err.Error()
		return report
	}
	defer rows.Close()

	symbols := make([]map[string]interface{}, 0)
	totalSimTicks := int64(0)

	for rows.Next() {
		var symbol string
		var tickCount int64
		var firstTick, lastTick int64

		if err := rows.Scan(&symbol, &tickCount, &firstTick, &lastTick); err != nil {
			continue
		}

		totalSimTicks += tickCount

		symbols = append(symbols, map[string]interface{}{
			"symbol":     symbol,
			"tick_count": tickCount,
			"first_tick": time.Unix(0, firstTick*int64(time.Millisecond)).Format(time.RFC3339),
			"last_tick":  time.Unix(0, lastTick*int64(time.Millisecond)).Format(time.RFC3339),
		})
	}

	report["total_simulation_ticks"] = totalSimTicks
	report["symbols_with_simulation"] = symbols
	report["symbol_count"] = len(symbols)

	// Production warning
	if os.Getenv("ENVIRONMENT") == "production" && totalSimTicks > 0 {
		report["production_alert"] = fmt.Sprintf("CRITICAL: %d simulation ticks found in PRODUCTION database", totalSimTicks)
	}

	return report
}
