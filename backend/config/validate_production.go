package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// ProductionValidator provides comprehensive production environment validation
type ProductionValidator struct {
	config *Config
	errors []string
	warnings []string
}

// NewProductionValidator creates a new production validator
func NewProductionValidator(cfg *Config) *ProductionValidator {
	return &ProductionValidator{
		config: cfg,
		errors: make([]string, 0),
		warnings: make([]string, 0),
	}
}

// ValidateProduction performs comprehensive production validation
// Returns error if any critical checks fail
func (v *ProductionValidator) ValidateProduction() error {
	log.Println("[ProductionValidator] Starting production environment validation...")

	// Check 1: Environment must be production
	v.checkEnvironment()

	// Check 2: Simulation code must not be compiled
	v.checkNoSimulationCode()

	// Check 3: Environment variables must prohibit simulation
	v.checkSimulationEnvironmentVars()

	// Check 4: FIX sessions must be configured
	v.checkFIXConfiguration()

	// Check 5: Real LP configuration must exist
	v.checkLPConfiguration()

	// Check 6: Database must not contain simulation data
	v.checkDatabaseForSimulation()

	// Check 7: Critical security settings
	v.checkSecuritySettings()

	// Print summary
	v.printValidationSummary()

	// Fail if any errors
	if len(v.errors) > 0 {
		return fmt.Errorf("production validation failed with %d critical errors", len(v.errors))
	}

	log.Println("[ProductionValidator] ✓ All production validation checks passed")
	return nil
}

// checkEnvironment verifies production environment
func (v *ProductionValidator) checkEnvironment() {
	env := os.Getenv("ENVIRONMENT")
	if env != "production" {
		v.errors = append(v.errors, fmt.Sprintf("ENVIRONMENT must be 'production', got '%s'", env))
	}
}

// checkNoSimulationCode ensures simulation code is not compiled in
func (v *ProductionValidator) checkNoSimulationCode() {
	// This will be checked via build tags at compile time
	// At runtime, we check that ALLOW_SIMULATION is explicitly false or unset
	allowSim := os.Getenv("ALLOW_SIMULATION")
	if strings.ToLower(allowSim) == "true" || allowSim == "1" {
		v.errors = append(v.errors, "ALLOW_SIMULATION=true is FORBIDDEN in production")
	}
}

// checkSimulationEnvironmentVars ensures simulation is disabled
func (v *ProductionValidator) checkSimulationEnvironmentVars() {
	// Verify ALLOW_SIMULATION is explicitly set to false or unset
	allowSim := os.Getenv("ALLOW_SIMULATION")
	if allowSim != "" && allowSim != "false" && allowSim != "0" {
		v.errors = append(v.errors, fmt.Sprintf("ALLOW_SIMULATION must be unset or 'false', got '%s'", allowSim))
	}

	// Verify no simulation mode flags
	if os.Getenv("SIMULATION_MODE") != "" {
		v.errors = append(v.errors, "SIMULATION_MODE environment variable must not be set in production")
	}
}

// checkFIXConfiguration validates FIX sessions are configured
func (v *ProductionValidator) checkFIXConfiguration() {
	// Check for FIX session configuration
	yofxHost := os.Getenv("YOFX_HOST")
	yofxPort := os.Getenv("YOFX_PORT")

	if yofxHost == "" {
		v.warnings = append(v.warnings, "YOFX_HOST not configured - FIX market data will not be available")
	}
	if yofxPort == "" {
		v.warnings = append(v.warnings, "YOFX_PORT not configured - FIX market data will not be available")
	}

	// Verify FIX credentials exist
	fixUser := os.Getenv("YOFX_SENDER_COMP_ID")
	if fixUser == "" {
		v.warnings = append(v.warnings, "YOFX_SENDER_COMP_ID not configured - FIX sessions may fail")
	}
}

// checkLPConfiguration validates real LP configuration
func (v *ProductionValidator) checkLPConfiguration() {
	// At least one real LP should be configured
	hasRealLP := false

	if v.config.LP.OandaAPIKey != "" && v.config.LP.OandaAccountID != "" {
		hasRealLP = true
		log.Println("[ProductionValidator] ✓ OANDA LP configured")
	}

	if v.config.LP.BinanceAPIKey != "" {
		hasRealLP = true
		log.Println("[ProductionValidator] ✓ Binance LP configured")
	}

	yofxHost := os.Getenv("YOFX_HOST")
	if yofxHost != "" {
		hasRealLP = true
		log.Println("[ProductionValidator] ✓ YOFX FIX LP configured")
	}

	if !hasRealLP {
		v.errors = append(v.errors, "No real liquidity providers configured - production requires at least one real LP (OANDA, Binance, or YOFX)")
	}
}

// checkDatabaseForSimulation scans database for simulation data
func (v *ProductionValidator) checkDatabaseForSimulation() {
	// Check if database connection is available
	dbHost := v.config.Database.Host
	if dbHost == "" {
		v.warnings = append(v.warnings, "Database not configured - skipping simulation data check")
		return
	}

	// Try to connect to database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		v.config.Database.Host,
		v.config.Database.Port,
		v.config.Database.User,
		v.config.Database.Password,
		v.config.Database.Name,
		v.config.Database.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		v.warnings = append(v.warnings, fmt.Sprintf("Could not connect to database: %v", err))
		return
	}
	defer db.Close()

	// Set connection timeout
	db.SetConnMaxLifetime(5 * time.Second)

	// Ping database
	if err := db.Ping(); err != nil {
		v.warnings = append(v.warnings, fmt.Sprintf("Database ping failed: %v - skipping simulation check", err))
		return
	}

	// Check for simulation data in ticks table
	v.scanTicksTableForSimulation(db)
}

// scanTicksTableForSimulation checks ticks table for LP='SIM' entries
func (v *ProductionValidator) scanTicksTableForSimulation(db *sql.DB) {
	// Check if ticks table exists
	var tableExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'ticks'
		)
	`).Scan(&tableExists)

	if err != nil || !tableExists {
		log.Println("[ProductionValidator] Ticks table does not exist yet - skipping simulation check")
		return
	}

	// Count simulation ticks
	var simCount int64
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM ticks
		WHERE lp_source = 'SIM' OR lp_source = 'SIMULATION'
	`).Scan(&simCount)

	if err != nil {
		v.warnings = append(v.warnings, fmt.Sprintf("Could not scan ticks table: %v", err))
		return
	}

	if simCount > 0 {
		v.errors = append(v.errors, fmt.Sprintf("CRITICAL: Found %d simulation ticks in production database (LP='SIM')", simCount))
		log.Printf("[ProductionValidator] ✗ CRITICAL: %d simulation ticks found in database", simCount)
	} else {
		log.Println("[ProductionValidator] ✓ No simulation data found in ticks table")
	}

	// Get count by LP source
	rows, err := db.Query(`
		SELECT lp_source, COUNT(*) as count
		FROM ticks
		GROUP BY lp_source
		ORDER BY count DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		log.Println("[ProductionValidator] Liquidity provider distribution:")
		for rows.Next() {
			var lpSource string
			var count int64
			if err := rows.Scan(&lpSource, &count); err == nil {
				log.Printf("  - %s: %d ticks", lpSource, count)
			}
		}
	}
}

// checkSecuritySettings validates critical security configuration
func (v *ProductionValidator) checkSecuritySettings() {
	// JWT secret must be set
	if v.config.JWT.Secret == "" {
		v.errors = append(v.errors, "JWT_SECRET must be set in production")
	} else if len(v.config.JWT.Secret) < 32 {
		v.errors = append(v.errors, "JWT_SECRET must be at least 32 characters in production")
	}

	// Admin password must be set
	if v.config.Admin.Password == "" {
		v.errors = append(v.errors, "ADMIN_PASSWORD_HASH must be set in production")
	}

	// Master encryption key must be set
	if v.config.Encryption.MasterKey == "" {
		v.errors = append(v.errors, "MASTER_ENCRYPTION_KEY must be set in production")
	} else if len(v.config.Encryption.MasterKey) < 32 {
		v.errors = append(v.errors, "MASTER_ENCRYPTION_KEY must be at least 32 characters")
	}

	// Database SSL should be enabled in production
	if v.config.Database.SSLMode == "disable" {
		v.warnings = append(v.warnings, "Database SSL is disabled - consider enabling for production")
	}
}

// printValidationSummary prints the validation results
func (v *ProductionValidator) printValidationSummary() {
	log.Println("\n╔════════════════════════════════════════════════════════════════╗")
	log.Println("║         PRODUCTION ENVIRONMENT VALIDATION SUMMARY              ║")
	log.Println("╚════════════════════════════════════════════════════════════════╝")

	// Print errors
	if len(v.errors) > 0 {
		log.Printf("\n⚠️  CRITICAL ERRORS (%d):", len(v.errors))
		for i, err := range v.errors {
			log.Printf("  %d. %s", i+1, err)
		}
	} else {
		log.Println("\n✓ No critical errors found")
	}

	// Print warnings
	if len(v.warnings) > 0 {
		log.Printf("\n⚠️  WARNINGS (%d):", len(v.warnings))
		for i, warn := range v.warnings {
			log.Printf("  %d. %s", i+1, warn)
		}
	} else {
		log.Println("✓ No warnings")
	}

	log.Println("\n════════════════════════════════════════════════════════════════\n")
}

// GetDataSourceMetrics returns metrics about data sources
func GetDataSourceMetrics(db *sql.DB) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Count ticks by LP source
	rows, err := db.Query(`
		SELECT
			lp_source,
			COUNT(*) as tick_count,
			MIN(timestamp) as first_tick,
			MAX(timestamp) as last_tick
		FROM ticks
		GROUP BY lp_source
		ORDER BY tick_count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lpSources := make([]map[string]interface{}, 0)
	hasSimulation := false

	for rows.Next() {
		var lpSource string
		var tickCount int64
		var firstTick, lastTick int64

		if err := rows.Scan(&lpSource, &tickCount, &firstTick, &lastTick); err != nil {
			return nil, err
		}

		if lpSource == "SIM" || lpSource == "SIMULATION" {
			hasSimulation = true
		}

		lpSources = append(lpSources, map[string]interface{}{
			"lp_source":   lpSource,
			"tick_count":  tickCount,
			"first_tick":  time.Unix(0, firstTick*int64(time.Millisecond)),
			"last_tick":   time.Unix(0, lastTick*int64(time.Millisecond)),
		})
	}

	metrics["lp_sources"] = lpSources
	metrics["has_simulation"] = hasSimulation
	metrics["timestamp"] = time.Now()

	return metrics, nil
}
