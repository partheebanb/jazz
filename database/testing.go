package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	testDB *DB
)

// GetTestDB returns the shared test database connection.
// Available after TestMain has run and SetupTestDB succeeded.
// Returns nil if called before TestMain.
func GetTestDB() *DB {
	return testDB
}

// SetupTestDB creates a test database connection and runs migrations.
// Should be called once in TestMain, not in individual tests.
// Migrations are embedded inline (not read from files) for test isolation.
// Returns error if connection fails or migrations fail.
func SetupTestDB(dbURL string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := Connect(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	if err := runTestMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func runTestMigrations(db *DB) error {
	ctx := context.Background()

	migrations := []string{
		`
		CREATE TABLE IF NOT EXISTS projects (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			api_key VARCHAR(64) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_projects_api_key ON projects(api_key);
		`,
		`
		CREATE TABLE IF NOT EXISTS logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			level VARCHAR(20) NOT NULL,
			message TEXT NOT NULL,
			source VARCHAR(100),
			timestamp TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
		CREATE INDEX IF NOT EXISTS idx_logs_source ON logs(source);
		CREATE INDEX IF NOT EXISTS idx_logs_project_id ON logs(project_id);
		CREATE INDEX IF NOT EXISTS idx_logs_message_search ON logs USING GIN (to_tsvector('english', message));
		`,
	}

	for _, migration := range migrations {
		_, err := db.Pool.Exec(ctx, migration)
		if err != nil {
			return err
		}
	}

	return nil
}

// CleanupTestDB truncates all tables for a fresh test state.
// Call this at the start of each integration test.
// Uses CASCADE to handle foreign key dependencies.
// Fails the test if truncation fails.
func CleanupTestDB(t *testing.T, db *DB) {
	t.Helper()

	ctx := context.Background()
	_, err := db.Pool.Exec(ctx, "TRUNCATE TABLE logs, projects CASCADE")
	require.NoError(t, err)
}

// TeardownTestDB closes the test database connection.
// Should be called once in TestMain after all tests complete.
// Safe to call with nil DB (no-op).
func TeardownTestDB(db *DB) {
	if db != nil {
		db.Close()
	}
}
