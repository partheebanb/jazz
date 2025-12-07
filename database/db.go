// Package database provides PostgreSQL persistence layer for Jazz logging platform.
// It implements connection pooling, CRUD operations for projects and logs,
// and full-text search using PostgreSQL GIN indexes.
package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents a PostgreSQL connection pool with configured limits.
// It wraps pgxpool.Pool and provides methods for all database operations.
// Safe for concurrent use.
type DB struct {
	Pool *pgxpool.Pool
}

// Connect establishes a connection pool to PostgreSQL with production-ready settings.
// Connection pool is configured with:
//   - MaxConns: 25 (prevent overwhelming database)
//   - MinConns: 5 (maintain ready connections)
//   - MaxConnLifetime: 1 hour (prevent stale connections)
//   - MaxConnIdleTime: 30 minutes (release idle connections)
//
// Returns error if unable to parse URL, create pool, or ping database.
func Connect(ctx context.Context, databaseURL string) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established")
	return &DB{Pool: pool}, nil
}

// Close gracefully shuts down the connection pool.
// Waits for all active queries to complete before closing.
// Safe to call multiple times - subsequent calls are no-ops.
func (db *DB) Close() {
	db.Pool.Close()
	log.Println("Database connection closed")
}
