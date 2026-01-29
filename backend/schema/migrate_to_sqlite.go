package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Tick represents a single market price update
type Tick struct {
	BrokerID  string    `json:"broker_id"`
	Symbol    string    `json:"symbol"`
	Bid       float64   `json:"bid"`
	Ask       float64   `json:"ask"`
	Spread    float64   `json:"spread"`
	Timestamp time.Time `json:"timestamp"`
	LP        string    `json:"lp"`
}

// MigrationStats tracks migration progress
type MigrationStats struct {
	SymbolsProcessed int
	FilesProcessed   int
	TicksMigrated    int64
	TicksSkipped     int64
	Errors           int
	StartTime        time.Time
	EndTime          time.Time
}

// Config holds migration configuration
type Config struct {
	InputDir   string
	OutputDir  string
	BatchSize  int
	Verbose    bool
	DryRun     bool
	SchemaPath string
}

var schema = `
CREATE TABLE IF NOT EXISTS ticks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol VARCHAR(20) NOT NULL,
    timestamp INTEGER NOT NULL,
    bid REAL NOT NULL,
    ask REAL NOT NULL,
    spread REAL NOT NULL,
    volume INTEGER DEFAULT 0,
    lp_source VARCHAR(50),
    flags INTEGER DEFAULT 0,
    created_at INTEGER DEFAULT (strftime('%s', 'now') * 1000)
);

CREATE INDEX IF NOT EXISTS idx_ticks_symbol_timestamp ON ticks(symbol, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_ticks_timestamp ON ticks(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_ticks_symbol ON ticks(symbol);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ticks_unique ON ticks(symbol, timestamp);

CREATE TABLE IF NOT EXISTS symbols (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol VARCHAR(20) UNIQUE NOT NULL,
    description VARCHAR(255),
    asset_class VARCHAR(50),
    base_currency VARCHAR(10),
    quote_currency VARCHAR(10),
    tick_size REAL,
    contract_size REAL,
    is_active BOOLEAN DEFAULT 1,
    first_tick_at INTEGER,
    last_tick_at INTEGER,
    total_ticks INTEGER DEFAULT 0,
    created_at INTEGER DEFAULT (strftime('%s', 'now') * 1000),
    updated_at INTEGER DEFAULT (strftime('%s', 'now') * 1000)
);

CREATE TABLE IF NOT EXISTS migration_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_file VARCHAR(255) NOT NULL,
    target_db VARCHAR(255) NOT NULL,
    ticks_migrated INTEGER NOT NULL,
    migration_time INTEGER NOT NULL,
    migrated_at INTEGER DEFAULT (strftime('%s', 'now') * 1000)
);
`

func main() {
	cfg := parseFlags()

	log.Printf("=== SQLite Migration Tool ===")
	log.Printf("Input directory: %s", cfg.InputDir)
	log.Printf("Output directory: %s", cfg.OutputDir)
	log.Printf("Batch size: %d", cfg.BatchSize)
	log.Printf("Dry run: %v", cfg.DryRun)
	log.Printf("")

	stats := &MigrationStats{
		StartTime: time.Now(),
	}

	// Scan input directory for symbol directories
	symbols, err := ioutil.ReadDir(cfg.InputDir)
	if err != nil {
		log.Fatalf("Failed to read input directory: %v", err)
	}

	for _, symbolDir := range symbols {
		if !symbolDir.IsDir() {
			continue
		}

		symbol := symbolDir.Name()
		if cfg.Verbose {
			log.Printf("Processing symbol: %s", symbol)
		}

		// Process all JSON files for this symbol
		symbolPath := filepath.Join(cfg.InputDir, symbol)
		files, err := ioutil.ReadDir(symbolPath)
		if err != nil {
			log.Printf("Failed to read symbol directory %s: %v", symbol, err)
			stats.Errors++
			continue
		}

		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".json") {
				continue
			}

			// Extract date from filename (e.g., "2026-01-19.json")
			date := strings.TrimSuffix(file.Name(), ".json")

			filePath := filepath.Join(symbolPath, file.Name())
			if err := migrateFile(cfg, symbol, date, filePath, stats); err != nil {
				log.Printf("Failed to migrate %s: %v", filePath, err)
				stats.Errors++
			} else {
				stats.FilesProcessed++
			}
		}

		stats.SymbolsProcessed++
	}

	stats.EndTime = time.Now()
	printStats(stats)
}

func parseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.InputDir, "input-dir", "data/ticks", "Input directory containing JSON files")
	flag.StringVar(&cfg.OutputDir, "output-dir", "data/ticks/db", "Output directory for SQLite databases")
	flag.IntVar(&cfg.BatchSize, "batch-size", 1000, "Number of ticks per batch insert")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Verbose logging")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Dry run (no writes)")
	flag.StringVar(&cfg.SchemaPath, "schema", "backend/schema/ticks.sql", "Path to schema file")

	flag.Parse()
	return cfg
}

func migrateFile(cfg *Config, symbol, date, filePath string, stats *MigrationStats) error {
	startTime := time.Now()

	// Read JSON file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse ticks
	var ticks []Tick
	if err := json.Unmarshal(data, &ticks); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(ticks) == 0 {
		if cfg.Verbose {
			log.Printf("  Skipping empty file: %s", filePath)
		}
		return nil
	}

	if cfg.Verbose {
		log.Printf("  Migrating %s: %d ticks", filePath, len(ticks))
	}

	if cfg.DryRun {
		stats.TicksMigrated += int64(len(ticks))
		return nil
	}

	// Parse date to determine output database
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date format %s: %w", date, err)
	}

	// Create output directory structure: YYYY/MM/
	dbDir := filepath.Join(cfg.OutputDir, t.Format("2006"), t.Format("01"))
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Open/create SQLite database for this date
	dbPath := filepath.Join(dbDir, fmt.Sprintf("ticks_%s.db", date))
	db, err := openDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Insert ticks in batches
	migrated, skipped, err := insertTicks(db, ticks, cfg.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to insert ticks: %w", err)
	}

	stats.TicksMigrated += migrated
	stats.TicksSkipped += skipped

	// Log migration
	logMigration(db, filePath, dbPath, int(migrated), int(time.Since(startTime).Milliseconds()))

	if cfg.Verbose {
		log.Printf("  ✓ Migrated %d ticks (skipped %d duplicates) in %v", migrated, skipped, time.Since(startTime))
	}

	return nil
}

func openDatabase(dbPath string) (*sql.DB, error) {
	// Check if database exists
	exists := fileExists(dbPath)

	// Open database with WAL mode and performance optimizations
	dsn := dbPath + "?_journal_mode=WAL&_synchronous=NORMAL&cache=shared&_busy_timeout=5000"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	// Create schema if new database
	if !exists {
		if _, err := db.Exec(schema); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create schema: %w", err)
		}

		// Set performance pragmas
		pragmas := []string{
			"PRAGMA cache_size = -64000",       // 64MB cache
			"PRAGMA temp_store = MEMORY",       // In-memory temp tables
			"PRAGMA mmap_size = 268435456",     // 256MB memory-mapped I/O
		}

		for _, pragma := range pragmas {
			if _, err := db.Exec(pragma); err != nil {
				log.Printf("Warning: Failed to set pragma: %s: %v", pragma, err)
			}
		}
	}

	return db, nil
}

func insertTicks(db *sql.DB, ticks []Tick, batchSize int) (int64, int64, error) {
	var migrated, skipped int64

	// Prepare insert statement
	stmt, err := db.Prepare(`
		INSERT OR IGNORE INTO ticks (symbol, timestamp, bid, ask, spread, lp_source)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, 0, err
	}
	defer stmt.Close()

	// Process in batches
	for i := 0; i < len(ticks); i += batchSize {
		end := i + batchSize
		if end > len(ticks) {
			end = len(ticks)
		}

		batch := ticks[i:end]

		// Start transaction
		tx, err := db.Begin()
		if err != nil {
			return migrated, skipped, err
		}

		txStmt := tx.Stmt(stmt)

		for _, tick := range batch {
			result, err := txStmt.Exec(
				tick.Symbol,
				tick.Timestamp.UnixMilli(),
				tick.Bid,
				tick.Ask,
				tick.Spread,
				tick.LP,
			)
			if err != nil {
				tx.Rollback()
				return migrated, skipped, err
			}

			rows, _ := result.RowsAffected()
			if rows > 0 {
				migrated++
			} else {
				skipped++ // Duplicate (already exists)
			}
		}

		if err := tx.Commit(); err != nil {
			return migrated, skipped, err
		}
	}

	return migrated, skipped, nil
}

func logMigration(db *sql.DB, sourceFile, targetDB string, ticksMigrated, migrationTimeMs int) {
	_, err := db.Exec(`
		INSERT INTO migration_log (source_file, target_db, ticks_migrated, migration_time)
		VALUES (?, ?, ?, ?)
	`, sourceFile, targetDB, ticksMigrated, migrationTimeMs)

	if err != nil {
		log.Printf("Warning: Failed to log migration: %v", err)
	}
}

func printStats(stats *MigrationStats) {
	duration := stats.EndTime.Sub(stats.StartTime)
	ticksPerSec := float64(stats.TicksMigrated) / duration.Seconds()

	log.Printf("")
	log.Printf("=== Migration Complete ===")
	log.Printf("Duration: %v", duration)
	log.Printf("Symbols processed: %d", stats.SymbolsProcessed)
	log.Printf("Files processed: %d", stats.FilesProcessed)
	log.Printf("Ticks migrated: %d", stats.TicksMigrated)
	log.Printf("Ticks skipped (duplicates): %d", stats.TicksSkipped)
	log.Printf("Errors: %d", stats.Errors)
	log.Printf("Performance: %.0f ticks/sec", ticksPerSec)
	log.Printf("")

	if stats.Errors > 0 {
		log.Printf("⚠️  Migration completed with %d errors. Review logs above.", stats.Errors)
	} else {
		log.Printf("✅ Migration successful!")
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
