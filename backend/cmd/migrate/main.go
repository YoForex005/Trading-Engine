package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/epic1st/rtx/backend/config"
	"github.com/epic1st/rtx/backend/db/migrations"
	_ "github.com/lib/pq"
)

func main() {
	// Command flags
	upCmd := flag.Bool("up", false, "Run all pending migrations")
	downCmd := flag.Bool("down", false, "Rollback last migration")
	statusCmd := flag.Bool("status", false, "Show migration status")
	initCmd := flag.Bool("init", false, "Initialize migrations table")
	version := flag.Int64("version", 0, "Migrate up to specific version")

	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Build database connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Printf("[Migrate] Connected to database: %s@%s:%s/%s",
		cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	// Create migrator
	migrator := migrations.NewMigrator(db)

	// Register all migrations
	for _, m := range migrations.GetRegisteredMigrations() {
		migrator.Register(m)
	}

	// Execute command
	switch {
	case *initCmd:
		log.Println("[Migrate] Initializing migrations table...")
		if err := migrator.Init(); err != nil {
			log.Fatalf("Failed to initialize: %v", err)
		}
		log.Println("[Migrate] ✅ Migrations table initialized")

	case *upCmd:
		log.Println("[Migrate] Running all pending migrations...")
		if err := migrator.Init(); err != nil {
			log.Fatalf("Failed to initialize: %v", err)
		}
		if err := migrator.Up(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("[Migrate] ✅ All migrations completed successfully")

	case *downCmd:
		log.Println("[Migrate] Rolling back last migration...")
		if err := migrator.Down(); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		log.Println("[Migrate] ✅ Rollback completed successfully")

	case *statusCmd:
		if err := migrator.Init(); err != nil {
			log.Fatalf("Failed to initialize: %v", err)
		}
		if err := migrator.Status(); err != nil {
			log.Fatalf("Failed to get status: %v", err)
		}

	case *version > 0:
		log.Printf("[Migrate] Migrating up to version %d...", *version)
		if err := migrator.Init(); err != nil {
			log.Fatalf("Failed to initialize: %v", err)
		}
		if err := migrator.UpTo(*version); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Printf("[Migrate] ✅ Migrated up to version %d successfully", *version)

	default:
		fmt.Println("RTX Trading Engine - Database Migration Tool")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  migrate -init          Initialize migrations table")
		fmt.Println("  migrate -up            Run all pending migrations")
		fmt.Println("  migrate -down          Rollback last migration")
		fmt.Println("  migrate -status        Show migration status")
		fmt.Println("  migrate -version=N     Migrate up to specific version")
		fmt.Println()
		fmt.Println("Environment variables (or use .env file):")
		fmt.Println("  DB_HOST                Database host (default: localhost)")
		fmt.Println("  DB_PORT                Database port (default: 5432)")
		fmt.Println("  DB_NAME                Database name (default: trading_engine)")
		fmt.Println("  DB_USER                Database user (default: postgres)")
		fmt.Println("  DB_PASSWORD            Database password")
		fmt.Println("  DB_SSL_MODE            SSL mode (default: disable)")
		fmt.Println()
		os.Exit(1)
	}
}
