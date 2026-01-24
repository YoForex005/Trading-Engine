package tickstore

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// retryConfig defines retry behavior for database operations
type retryConfig struct {
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

// defaultRetryConfig for SQLite operations
var defaultRetryConfig = retryConfig{
	maxRetries: 3,
	baseDelay:  10 * time.Millisecond,
	maxDelay:   1 * time.Second,
}

// isBusyError checks if error is retryable (SQLITE_BUSY, SQLITE_LOCKED)
func isBusyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "database is locked") ||
		strings.Contains(errStr, "SQLITE_BUSY") ||
		strings.Contains(errStr, "SQLITE_LOCKED")
}

// retryWithBackoff executes fn with exponential backoff
func retryWithBackoff(cfg retryConfig, fn func() error) (int, error) {
	var lastErr error

	for attempt := 0; attempt < cfg.maxRetries; attempt++ {
		err := fn()
		if err == nil {
			return attempt, nil // Success - return number of retries
		}

		lastErr = err

		// Only retry on busy/locked errors
		if !isBusyError(err) {
			return attempt, fmt.Errorf("non-retryable error: %w", err)
		}

		// Don't sleep after last attempt
		if attempt < cfg.maxRetries-1 {
			// Exponential backoff with jitter
			delay := cfg.baseDelay * time.Duration(1<<uint(attempt))
			if delay > cfg.maxDelay {
				delay = cfg.maxDelay
			}

			// Add jitter (±25%) to prevent thundering herd
			jitter := time.Duration(rand.Int63n(int64(delay / 4)))
			if rand.Intn(2) == 0 {
				delay += jitter
			} else {
				delay -= jitter
			}

			log.Printf("[SQLiteStore] Retry attempt %d/%d after %v (error: %v)",
				attempt+1, cfg.maxRetries, delay, err)
			time.Sleep(delay)
		}
	}

	return cfg.maxRetries, fmt.Errorf("retry exhausted after %d attempts: %w", cfg.maxRetries, lastErr)
}

// ErrorMetrics tracks error statistics for observability
// PERFORMANCE FIX #3: Error metrics & alerting
type ErrorMetrics struct {
	WriteErrors       int64
	LastError         error
	LastErrorTime     time.Time
	ConsecutiveErrors int64
}

// SQLiteStore provides high-performance persistent tick storage using SQLite
// with daily partitioning, WAL mode, and asynchronous batch writes.
type SQLiteStore struct {
	db            *sql.DB
	writeQueue    chan TickWrite
	stopChan      chan struct{}
	wg            sync.WaitGroup
	currentDBPath string
	basePath      string
	mu            sync.RWMutex
	metrics       BatchMetrics
	errorMetrics  ErrorMetrics // PERFORMANCE FIX #3: Track errors for alerting
}

// BatchMetrics tracks batch write performance and failures
type BatchMetrics struct {
	BatchesWritten     int64
	BatchFailures      int64
	TicksWritten       int64
	TicksLost          int64
	RetriesTotal       int64
	LastBatchError     error
	LastBatchErrorTime time.Time
	mu                 sync.RWMutex
}

// TickWrite represents a queued write operation
type TickWrite struct {
	tick Tick
}

// NewSQLiteStore creates a new SQLite-based tick store
func NewSQLiteStore(basePath string) (*SQLiteStore, error) {
	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}

	store := &SQLiteStore{
		basePath:   basePath,
		writeQueue: make(chan TickWrite, 10000), // 10K tick buffer
		stopChan:   make(chan struct{}),
	}

	// Open database for today
	if err := store.rotateDatabaseIfNeeded(); err != nil {
		return nil, fmt.Errorf("failed to open initial database: %w", err)
	}

	// Start async batch writer
	store.wg.Add(1)
	go store.batchWriter()

	// Start daily rotation checker
	store.wg.Add(1)
	go store.rotationWatcher()

	log.Printf("[SQLiteStore] Initialized with base path: %s", basePath)
	return store, nil
}

// rotateDatabaseIfNeeded opens or creates the database for today
func (s *SQLiteStore) rotateDatabaseIfNeeded() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Construct path for today's database
	now := time.Now().UTC()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("2006-01-02")

	dbDir := filepath.Join(s.basePath, year, month)
	dbPath := filepath.Join(dbDir, fmt.Sprintf("ticks_%s.db", day))

	// Check if we're already using this database
	if s.currentDBPath == dbPath && s.db != nil {
		return nil
	}

	// Close old database if exists
	if s.db != nil {
		log.Printf("[SQLiteStore] Rotating database from %s to %s", filepath.Base(s.currentDBPath), filepath.Base(dbPath))

		// SECURITY: Checkpoint WAL before rotation to ensure all data is persisted
		// FULL mode blocks until all WAL data is moved to main database
		// This prevents data loss during database rotation
		if _, err := s.db.Exec("PRAGMA wal_checkpoint(FULL)"); err != nil {
			log.Printf("[SQLiteStore] WARNING: WAL checkpoint failed before rotation: %v", err)
		} else {
			log.Printf("[SQLiteStore] WAL checkpoint completed before rotation")
		}

		if err := s.db.Close(); err != nil {
			log.Printf("[SQLiteStore] Warning: failed to close old database: %v", err)
		}
	}

	// Create directory
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open SQLite database with optimized settings
	// WAL mode: Write-Ahead Logging for better concurrency
	// NORMAL synchronous: fsync only at checkpoints (balanced durability/performance)
	// busy_timeout: Wait up to 5 seconds if database is locked
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=NORMAL&cache=shared&_cache_size=-64000&_busy_timeout=5000", dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool limits
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	// Initialize schema
	if err := s.initSchema(db); err != nil {
		db.Close()
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	s.db = db
	s.currentDBPath = dbPath
	log.Printf("[SQLiteStore] Opened database: %s", dbPath)

	return nil
}

// initSchema creates tables and indexes if they don't exist
func (s *SQLiteStore) initSchema(db *sql.DB) error {
	schema := `
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
		first_tick_at INTEGER,
		last_tick_at INTEGER,
		total_ticks INTEGER DEFAULT 0,
		created_at INTEGER DEFAULT (strftime('%s', 'now') * 1000),
		updated_at INTEGER DEFAULT (strftime('%s', 'now') * 1000)
	);

	CREATE TRIGGER IF NOT EXISTS update_symbol_stats AFTER INSERT ON ticks
	BEGIN
		INSERT OR REPLACE INTO symbols (symbol, first_tick_at, last_tick_at, total_ticks, updated_at)
		VALUES (
			NEW.symbol,
			COALESCE((SELECT first_tick_at FROM symbols WHERE symbol = NEW.symbol), NEW.timestamp),
			NEW.timestamp,
			COALESCE((SELECT total_ticks FROM symbols WHERE symbol = NEW.symbol), 0) + 1,
			strftime('%s', 'now') * 1000
		);
	END;
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	// Optimize database
	if _, err := db.Exec("PRAGMA optimize"); err != nil {
		log.Printf("[SQLiteStore] Warning: failed to optimize: %v", err)
	}

	return nil
}

// StoreTick queues a tick for asynchronous batch writing
// Accepts Tick with time.Time and converts internally to Unix milliseconds
func (s *SQLiteStore) StoreTick(tick Tick) error {
	select {
	case s.writeQueue <- TickWrite{tick: tick}:
		return nil
	default:
		return fmt.Errorf("write queue full")
	}
}

// batchWriter processes queued ticks in batches for optimal performance
func (s *SQLiteStore) batchWriter() {
	defer s.wg.Done()

	batch := make([]Tick, 0, 500)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			// Flush remaining batch before exiting
			if len(batch) > 0 {
				s.writeBatch(batch)
			}
			return

		case write := <-s.writeQueue:
			batch = append(batch, write.tick)
			if len(batch) >= 500 {
				s.writeBatch(batch)
				batch = batch[:0] // Reset batch
			}

		case <-ticker.C:
			// Periodic flush for low-volume periods
			if len(batch) > 0 {
				s.writeBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// writeBatch writes a batch of ticks to SQLite with retry logic
func (s *SQLiteStore) writeBatch(batch []Tick) {
	if len(batch) == 0 {
		return
	}

	// RESILIENCE: Retry transient failures to prevent data loss
	retries, err := retryWithBackoff(defaultRetryConfig, func() error {
		return s.writeBatchOnce(batch)
	})

	// Track retry metrics
	if retries > 0 {
		atomic.AddInt64(&s.metrics.RetriesTotal, int64(retries))
	}

	if err != nil {
		// CRITICAL: Surface retry exhaustion via metrics
		atomic.AddInt64(&s.metrics.BatchFailures, 1)
		atomic.AddInt64(&s.metrics.TicksLost, int64(len(batch)))

		s.metrics.mu.Lock()
		s.metrics.LastBatchError = err
		s.metrics.LastBatchErrorTime = time.Now()
		s.metrics.mu.Unlock()

		log.Printf("ERROR: Batch write failed after retries (lost %d ticks): %v", len(batch), err)

		// Alert if failure rate exceeds threshold
		failures := atomic.LoadInt64(&s.metrics.BatchFailures)
		if failures > 10 {
			// TODO: Integrate with alerting system
			log.Printf("ALERT: High batch failure rate detected (%d failures)", failures)
		}
	} else {
		atomic.AddInt64(&s.metrics.BatchesWritten, 1)
		atomic.AddInt64(&s.metrics.TicksWritten, int64(len(batch)))

		if len(batch) >= 100 {
			log.Printf("[SQLiteStore] Flushed %d ticks to disk (retries: %d)", len(batch), retries)
		}
	}
}

// writeBatchOnce performs single batch write attempt
func (s *SQLiteStore) writeBatchOnce(batch []Tick) error {
	s.mu.RLock()
	db := s.db
	s.mu.RUnlock()

	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Safe to call even after commit

	// Prepare insert statement
	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO ticks (symbol, timestamp, bid, ask, spread, lp_source)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Execute batch inserts
	for _, tick := range batch {
		// Convert time.Time to Unix milliseconds
		timestampMs := tick.Timestamp.UnixMilli()

		_, err := stmt.Exec(
			tick.Symbol,
			timestampMs,
			tick.Bid,
			tick.Ask,
			tick.Spread,
			tick.LP,
		)
		if err != nil {
			// PERFORMANCE FIX #3: Error metrics & alerting
			// Silent failures prevented - now tracked with counters and alerts
			atomic.AddInt64(&s.errorMetrics.WriteErrors, 1)
			atomic.AddInt64(&s.errorMetrics.ConsecutiveErrors, 1)
			s.errorMetrics.LastError = err
			s.errorMetrics.LastErrorTime = time.Now()

			// OBSERVABILITY: Alert on sustained error rates
			if atomic.LoadInt64(&s.errorMetrics.ConsecutiveErrors) >= 100 {
				log.Printf("[SQLiteStore] ALERT: 100+ consecutive tick write failures - check database health")
				// TODO: Integrate with PagerDuty/Slack for production alerting
			}

			return fmt.Errorf("failed to insert tick %s: %w", tick.Symbol, err)
		} else {
			// Reset consecutive error counter on success
			atomic.StoreInt64(&s.errorMetrics.ConsecutiveErrors, 0)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// SECURITY: Periodic WAL checkpoint to prevent WAL file from growing too large
	// PASSIVE mode doesn't block writers, fails immediately if DB is busy
	// Checkpoint every ~10 batches (5000 ticks) to balance durability and performance
	// This ensures WAL doesn't exceed ~5000 rows before being merged into main DB
	if len(batch) >= 450 { // Only checkpoint on substantial batches
		if _, err := db.Exec("PRAGMA wal_checkpoint(PASSIVE)"); err != nil {
			// PASSIVE checkpoint may fail if DB is busy - this is expected and safe
			// WAL will be checkpointed on next batch or during shutdown
			log.Printf("[SQLiteStore] PASSIVE checkpoint deferred (DB busy): %v", err)
		}
	}

	return nil
}

// rotationWatcher checks daily for database rotation at midnight UTC
func (s *SQLiteStore) rotationWatcher() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	lastRotationDate := time.Now().UTC().Format("2006-01-02")

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			currentDate := time.Now().UTC().Format("2006-01-02")
			if currentDate != lastRotationDate {
				log.Printf("[SQLiteStore] Date changed, rotating database...")
				if err := s.rotateDatabaseIfNeeded(); err != nil {
					log.Printf("[SQLiteStore] ERROR: rotation failed: %v", err)
				} else {
					lastRotationDate = currentDate
				}
			}
		}
	}
}

// GetRecentTicks retrieves the N most recent ticks for a symbol
func (s *SQLiteStore) GetRecentTicks(symbol string, limit int) ([]Tick, error) {
	s.mu.RLock()
	db := s.db
	s.mu.RUnlock()

	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := db.Query(`
		SELECT timestamp, bid, ask, spread, lp_source
		FROM ticks
		WHERE symbol = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`, symbol, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ticks []Tick
	for rows.Next() {
		var tick Tick
		var lp sql.NullString

		err := rows.Scan(&tick.Timestamp, &tick.Bid, &tick.Ask, &tick.Spread, &lp)
		if err != nil {
			return nil, err
		}

		tick.Symbol = symbol
		if lp.Valid {
			tick.LP = lp.String
		}

		ticks = append(ticks, tick)
	}

	return ticks, rows.Err()
}

// GetTicksInRange retrieves ticks for a symbol within a time range
func (s *SQLiteStore) GetTicksInRange(symbol string, startTime, endTime int64, limit int) ([]Tick, error) {
	s.mu.RLock()
	db := s.db
	s.mu.RUnlock()

	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := db.Query(`
		SELECT timestamp, bid, ask, spread, lp_source
		FROM ticks
		WHERE symbol = ?
		  AND timestamp >= ?
		  AND timestamp <= ?
		ORDER BY timestamp DESC
		LIMIT ?
	`, symbol, startTime, endTime, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ticks []Tick
	for rows.Next() {
		var tick Tick
		var lp sql.NullString

		err := rows.Scan(&tick.Timestamp, &tick.Bid, &tick.Ask, &tick.Spread, &lp)
		if err != nil {
			return nil, err
		}

		tick.Symbol = symbol
		if lp.Valid {
			tick.LP = lp.String
		}

		ticks = append(ticks, tick)
	}

	return ticks, rows.Err()
}

// GetStats returns storage statistics
func (s *SQLiteStore) GetStats() (map[string]interface{}, error) {
	s.mu.RLock()
	db := s.db
	dbPath := s.currentDBPath
	s.mu.RUnlock()

	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	stats := make(map[string]interface{})

	// Get total tick count
	var tickCount int64
	err := db.QueryRow("SELECT COUNT(*) FROM ticks").Scan(&tickCount)
	if err != nil {
		return nil, err
	}
	stats["total_ticks"] = tickCount

	// Get symbol count
	var symbolCount int64
	err = db.QueryRow("SELECT COUNT(*) FROM symbols").Scan(&symbolCount)
	if err != nil {
		return nil, err
	}
	stats["symbol_count"] = symbolCount

	// Get database file size
	if fileInfo, err := os.Stat(dbPath); err == nil {
		stats["db_size_bytes"] = fileInfo.Size()
		stats["db_size_mb"] = float64(fileInfo.Size()) / (1024 * 1024)
	}

	stats["db_path"] = dbPath
	stats["queue_size"] = len(s.writeQueue)

	// Add batch metrics
	stats["batches_written"] = atomic.LoadInt64(&s.metrics.BatchesWritten)
	stats["batch_failures"] = atomic.LoadInt64(&s.metrics.BatchFailures)
	stats["ticks_written"] = atomic.LoadInt64(&s.metrics.TicksWritten)
	stats["ticks_lost"] = atomic.LoadInt64(&s.metrics.TicksLost)
	stats["retries_total"] = atomic.LoadInt64(&s.metrics.RetriesTotal)

	s.metrics.mu.RLock()
	if s.metrics.LastBatchError != nil {
		stats["last_batch_error"] = s.metrics.LastBatchError.Error()
		stats["last_batch_error_time"] = s.metrics.LastBatchErrorTime.Format(time.RFC3339)
	}
	s.metrics.mu.RUnlock()

	// Calculate success rate
	written := atomic.LoadInt64(&s.metrics.BatchesWritten)
	failed := atomic.LoadInt64(&s.metrics.BatchFailures)
	if written+failed > 0 {
		stats["batch_success_rate"] = float64(written) / float64(written+failed)
	}

	return stats, nil
}

// Stop gracefully shuts down the store
func (s *SQLiteStore) Stop() error {
	log.Println("[SQLiteStore] Stopping...")

	// Signal stop to workers
	close(s.stopChan)

	// Wait for workers to finish (this ensures final batch is written)
	s.wg.Wait()

	// Close database
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		log.Printf("[SQLiteStore] Shutting down gracefully...")

		// SECURITY: Checkpoint WAL before close to ensure durability
		// TRUNCATE mode (most aggressive):
		// 1. Blocks all writers until complete
		// 2. Moves ALL WAL contents to main database file
		// 3. Truncates WAL file to zero bytes
		// This prevents data loss if process crashes after shutdown but before OS flushes WAL
		// Without this, up to 500 ticks (one batch) could be lost in the unflushed WAL
		log.Printf("[SQLiteStore] Checkpointing WAL to main database...")
		if _, err := s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
			log.Printf("WARNING: WAL checkpoint failed during shutdown: %v", err)
			log.Printf("         Data may be recoverable from WAL file on next startup")
			// Continue with shutdown - SQLite will auto-recover WAL on next open
		} else {
			log.Printf("[SQLiteStore] WAL checkpoint completed successfully")
		}

		// Run final optimization to update query planner statistics
		if _, err := s.db.Exec("PRAGMA optimize"); err != nil {
			log.Printf("[SQLiteStore] Warning: final optimization failed: %v", err)
		}

		// Close database connection
		if err := s.db.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
		log.Println("[SQLiteStore] Database closed")
	}

	log.Println("[SQLiteStore] Stopped")
	return nil
}
