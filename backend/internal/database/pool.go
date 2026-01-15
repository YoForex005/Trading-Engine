package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

// InitPool creates application-wide PostgreSQL connection pool
// Call once at application startup with DATABASE_URL from environment
func InitPool(ctx context.Context, connString string) error {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	// Connection pool configuration optimized for trading platform
	// Based on RESEARCH.md recommendations: (CPU cores * 2) + 1 baseline
	config.MaxConns = 20                          // Adjust based on actual CPU count
	config.MinConns = 5                           // Keep connections ready
	config.MaxConnLifetime = 1 * time.Hour        // Stable single-instance database
	config.MaxConnIdleTime = 30 * time.Minute     // Release idle connections
	config.HealthCheckPeriod = 1 * time.Minute    // Periodic health checks

	pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// GetPool returns the singleton connection pool
// Returns nil if InitPool has not been called
func GetPool() *pgxpool.Pool {
	return pool
}

// Close gracefully closes the connection pool
// Call during application shutdown
func Close() {
	if pool != nil {
		pool.Close()
	}
}
