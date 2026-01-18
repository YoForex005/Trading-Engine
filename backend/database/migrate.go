package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// migrationsFS - migrations are loaded from filesystem instead of embed
// due to directory structure (migrations/ is sibling to database/)
var migrationsFS embed.FS // Unused - migrations loaded from ../migrations/ directory

// Migration represents a database migration
type Migration struct {
	Version     int
	Name        string
	Description string
	UpSQL       string
	DownSQL     string
	AppliedAt   *time.Time
}

// Migrator handles database migrations
type Migrator struct {
	db          *sql.DB
	dryRun      bool
	verbose     bool
	migrationsPath string
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB, options ...MigratorOption) *Migrator {
	m := &Migrator{
		db:          db,
		dryRun:      false,
		verbose:     false,
		migrationsPath: "../migrations",
	}

	for _, opt := range options {
		opt(m)
	}

	return m
}

// MigratorOption configures the migrator
type MigratorOption func(*Migrator)

// WithDryRun enables dry-run mode (no actual changes)
func WithDryRun(dryRun bool) MigratorOption {
	return func(m *Migrator) {
		m.dryRun = dryRun
	}
}

// WithVerbose enables verbose logging
func WithVerbose(verbose bool) MigratorOption {
	return func(m *Migrator) {
		m.verbose = verbose
	}
}

// WithMigrationsPath sets custom migrations path
func WithMigrationsPath(path string) MigratorOption {
	return func(m *Migrator) {
		m.migrationsPath = path
	}
}

// Initialize creates the migrations tracking table
func (m *Migrator) Initialize() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		applied_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		execution_time_ms INTEGER,
		checksum VARCHAR(64)
	);

	CREATE INDEX IF NOT EXISTS idx_schema_migrations_applied_at
		ON schema_migrations(applied_at DESC);
	`

	if m.dryRun {
		m.log("DRY RUN: Would create schema_migrations table")
		return nil
	}

	_, err := m.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	m.log("✓ Initialized schema_migrations table")
	return nil
}

// LoadMigrations loads all migration files
func (m *Migrator) LoadMigrations() ([]*Migration, error) {
	var migrations []*Migration

	// Read migrations from embedded filesystem
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		migration, err := m.parseMigrationFile(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to parse migration %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, migration)
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// parseMigrationFile parses a migration file
func (m *Migrator) parseMigrationFile(filename string) (*Migration, error) {
	// Extract version from filename (e.g., "001_init_schema.sql" -> 1)
	parts := strings.Split(filename, "_")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	var version int
	_, err := fmt.Sscanf(parts[0], "%d", &version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse version from filename %s: %w", filename, err)
	}

	// Read file content
	content, err := migrationsFS.ReadFile(filepath.Join("migrations", filename))
	if err != nil {
		return nil, fmt.Errorf("failed to read migration file: %w", err)
	}

	sqlContent := string(content)

	// Extract UP and DOWN migrations
	upSQL, downSQL := m.splitMigrationSQL(sqlContent)

	// Extract name and description from comments
	name := strings.TrimSuffix(filename, ".sql")
	description := m.extractDescription(sqlContent)

	return &Migration{
		Version:     version,
		Name:        name,
		Description: description,
		UpSQL:       upSQL,
		DownSQL:     downSQL,
	}, nil
}

// splitMigrationSQL splits the SQL into UP and DOWN parts
func (m *Migrator) splitMigrationSQL(content string) (up, down string) {
	// Find the DOWN migration section
	downMarker := "-- DOWN Migration"
	downIndex := strings.Index(content, downMarker)

	if downIndex == -1 {
		// No DOWN migration defined
		return content, ""
	}

	up = content[:downIndex]
	down = content[downIndex:]

	// Remove comment markers from DOWN migration
	down = strings.ReplaceAll(down, "/*", "")
	down = strings.ReplaceAll(down, "*/", "")

	return strings.TrimSpace(up), strings.TrimSpace(down)
}

// extractDescription extracts description from migration comments
func (m *Migrator) extractDescription(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-- Description:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "-- Description:"))
		}
	}
	return ""
}

// GetAppliedMigrations returns list of already applied migrations
func (m *Migrator) GetAppliedMigrations() (map[int]*Migration, error) {
	applied := make(map[int]*Migration)

	rows, err := m.db.Query(`
		SELECT version, name, description, applied_at
		FROM schema_migrations
		ORDER BY version
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		migration := &Migration{}
		err := rows.Scan(&migration.Version, &migration.Name, &migration.Description, &migration.AppliedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan migration row: %w", err)
		}
		applied[migration.Version] = migration
	}

	return applied, nil
}

// Up runs pending migrations
func (m *Migrator) Up() error {
	migrations, err := m.LoadMigrations()
	if err != nil {
		return err
	}

	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	pendingCount := 0
	for _, migration := range migrations {
		if _, exists := applied[migration.Version]; !exists {
			pendingCount++
		}
	}

	if pendingCount == 0 {
		m.log("✓ Database is up to date. No pending migrations.")
		return nil
	}

	m.log(fmt.Sprintf("Running %d pending migration(s)...\n", pendingCount))

	for _, migration := range migrations {
		if _, exists := applied[migration.Version]; exists {
			m.logVerbose(fmt.Sprintf("⊘ Skipping migration %d (already applied)", migration.Version))
			continue
		}

		if err := m.runMigration(migration, true); err != nil {
			return fmt.Errorf("migration %d failed: %w", migration.Version, err)
		}
	}

	m.log("\n✓ All migrations completed successfully!")
	return nil
}

// Down rolls back the last migration
func (m *Migrator) Down() error {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	if len(applied) == 0 {
		m.log("No migrations to rollback")
		return nil
	}

	// Find the highest version
	maxVersion := 0
	var lastMigration *Migration
	for version, migration := range applied {
		if version > maxVersion {
			maxVersion = version
			lastMigration = migration
		}
	}

	// Load the migration file to get DOWN SQL
	migrations, err := m.LoadMigrations()
	if err != nil {
		return err
	}

	var migrationToRollback *Migration
	for _, m := range migrations {
		if m.Version == maxVersion {
			migrationToRollback = m
			break
		}
	}

	if migrationToRollback == nil {
		return fmt.Errorf("migration file for version %d not found", maxVersion)
	}

	if migrationToRollback.DownSQL == "" {
		return fmt.Errorf("migration %d has no DOWN migration defined", maxVersion)
	}

	m.log(fmt.Sprintf("Rolling back migration %d: %s", maxVersion, lastMigration.Name))

	return m.runMigration(migrationToRollback, false)
}

// runMigration executes a migration (up or down)
func (m *Migrator) runMigration(migration *Migration, up bool) error {
	direction := "UP"
	sql := migration.UpSQL

	if !up {
		direction = "DOWN"
		sql = migration.DownSQL
	}

	m.log(fmt.Sprintf("→ Running migration %d (%s): %s", migration.Version, direction, migration.Name))

	if m.dryRun {
		m.log(fmt.Sprintf("DRY RUN: Would execute migration %d", migration.Version))
		m.logVerbose(fmt.Sprintf("SQL:\n%s", sql))
		return nil
	}

	start := time.Now()

	// Start transaction for atomic migration
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Execute migration SQL
	_, err = tx.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	executionTime := time.Since(start).Milliseconds()

	// Update schema_migrations table
	if up {
		_, err = tx.Exec(`
			INSERT INTO schema_migrations (version, name, description, execution_time_ms)
			VALUES ($1, $2, $3, $4)
		`, migration.Version, migration.Name, migration.Description, executionTime)
	} else {
		_, err = tx.Exec(`
			DELETE FROM schema_migrations WHERE version = $1
		`, migration.Version)
	}

	if err != nil {
		return fmt.Errorf("failed to update schema_migrations: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	m.log(fmt.Sprintf("  ✓ Completed in %dms", executionTime))
	return nil
}

// Status shows migration status
func (m *Migrator) Status() error {
	migrations, err := m.LoadMigrations()
	if err != nil {
		return err
	}

	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	fmt.Println("\n╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    Database Migration Status                     ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝\n")

	fmt.Printf("%-10s %-40s %-12s %-20s\n", "Version", "Name", "Status", "Applied At")
	fmt.Println(strings.Repeat("-", 85))

	for _, migration := range migrations {
		status := "Pending"
		appliedAt := "-"

		if appliedMigration, exists := applied[migration.Version]; exists {
			status = "Applied"
			if appliedMigration.AppliedAt != nil {
				appliedAt = appliedMigration.AppliedAt.Format("2006-01-02 15:04:05")
			}
		}

		fmt.Printf("%-10d %-40s %-12s %-20s\n",
			migration.Version,
			migration.Name,
			status,
			appliedAt,
		)
	}

	pendingCount := len(migrations) - len(applied)
	fmt.Printf("\nTotal migrations: %d | Applied: %d | Pending: %d\n",
		len(migrations), len(applied), pendingCount)

	return nil
}

// log prints a message
func (m *Migrator) log(message string) {
	log.Println(message)
}

// logVerbose prints a message only in verbose mode
func (m *Migrator) logVerbose(message string) {
	if m.verbose {
		log.Println(message)
	}
}

// Connect establishes a database connection
func Connect(connStr string) (*sql.DB, error) {
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

	return db, nil
}

// GetConnectionString builds connection string from environment variables
func GetConnectionString() string {
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
