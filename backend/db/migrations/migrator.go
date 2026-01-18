package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"time"

	_ "github.com/lib/pq"
)

// Migration represents a database migration
type Migration struct {
	Version     int64
	Name        string
	Up          func(*sql.Tx) error
	Down        func(*sql.Tx) error
	AppliedAt   *time.Time
}

// Migrator handles database migrations
type Migrator struct {
	db         *sql.DB
	migrations []*Migration
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: make([]*Migration, 0),
	}
}

// Register registers a migration
func (m *Migrator) Register(migration *Migration) {
	m.migrations = append(m.migrations, migration)
}

// Init creates the migrations tracking table
func (m *Migrator) Init() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version BIGINT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := m.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	log.Println("[Migrator] Migrations tracking table initialized")
	return nil
}

// GetAppliedMigrations returns list of applied migration versions
func (m *Migrator) GetAppliedMigrations() (map[int64]bool, error) {
	rows, err := m.db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int64]bool)
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, nil
}

// Up runs all pending migrations
func (m *Migrator) Up() error {
	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	// Get applied migrations
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Run pending migrations
	for _, migration := range m.migrations {
		if applied[migration.Version] {
			log.Printf("[Migrator] Migration %d (%s) already applied, skipping", migration.Version, migration.Name)
			continue
		}

		log.Printf("[Migrator] Applying migration %d (%s)...", migration.Version, migration.Name)

		// Begin transaction
		tx, err := m.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Run migration
		if err := migration.Up(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d failed: %w", migration.Version, err)
		}

		// Record migration
		_, err = tx.Exec("INSERT INTO schema_migrations (version, name) VALUES ($1, $2)",
			migration.Version, migration.Name)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration: %w", err)
		}

		// Commit
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration: %w", err)
		}

		log.Printf("[Migrator] Migration %d (%s) applied successfully", migration.Version, migration.Name)
	}

	log.Println("[Migrator] All migrations applied")
	return nil
}

// Down rolls back the last migration
func (m *Migrator) Down() error {
	// Get applied migrations
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	if len(applied) == 0 {
		log.Println("[Migrator] No migrations to rollback")
		return nil
	}

	// Find the latest applied migration
	var latestVersion int64
	for version := range applied {
		if version > latestVersion {
			latestVersion = version
		}
	}

	// Find migration by version
	var targetMigration *Migration
	for _, migration := range m.migrations {
		if migration.Version == latestVersion {
			targetMigration = migration
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration %d not found in registered migrations", latestVersion)
	}

	if targetMigration.Down == nil {
		return fmt.Errorf("migration %d has no down migration", latestVersion)
	}

	log.Printf("[Migrator] Rolling back migration %d (%s)...", targetMigration.Version, targetMigration.Name)

	// Begin transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Run down migration
	if err := targetMigration.Down(tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("rollback failed: %w", err)
	}

	// Remove migration record
	_, err = tx.Exec("DELETE FROM schema_migrations WHERE version = $1", targetMigration.Version)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	log.Printf("[Migrator] Migration %d (%s) rolled back successfully", targetMigration.Version, targetMigration.Name)
	return nil
}

// Status shows migration status
func (m *Migrator) Status() error {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	log.Println("[Migrator] Migration Status:")
	log.Println("=====================================")

	for _, migration := range m.migrations {
		status := "PENDING"
		if applied[migration.Version] {
			status = "APPLIED"
		}
		log.Printf("  %d - %s [%s]", migration.Version, migration.Name, status)
	}

	log.Println("=====================================")
	log.Printf("Total: %d migrations, %d applied, %d pending",
		len(m.migrations), len(applied), len(m.migrations)-len(applied))

	return nil
}

// UpTo runs migrations up to a specific version
func (m *Migrator) UpTo(targetVersion int64) error {
	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	// Get applied migrations
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Run pending migrations up to target
	for _, migration := range m.migrations {
		if migration.Version > targetVersion {
			break
		}

		if applied[migration.Version] {
			continue
		}

		log.Printf("[Migrator] Applying migration %d (%s)...", migration.Version, migration.Name)

		tx, err := m.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if err := migration.Up(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d failed: %w", migration.Version, err)
		}

		_, err = tx.Exec("INSERT INTO schema_migrations (version, name) VALUES ($1, $2)",
			migration.Version, migration.Name)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration: %w", err)
		}

		log.Printf("[Migrator] Migration %d (%s) applied successfully", migration.Version, migration.Name)
	}

	return nil
}
