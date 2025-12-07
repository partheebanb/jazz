package database

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestMain(m *testing.M) {
	// Connect to postgres (not a specific database)
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to postgres: %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure postgres is running:\n")
		fmt.Fprintf(os.Stderr, "  docker-compose up -d postgres\n")
		os.Exit(1)
	}

	// Drop test database if exists (cleanup from previous run)
	_, _ = conn.Exec(ctx, "DROP DATABASE IF EXISTS jazz_test")

	// Create test database
	_, err = conn.Exec(ctx, "CREATE DATABASE jazz_test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create test database: %v\n", err)
		_ = conn.Close(ctx)
		os.Exit(1)
	}

	_ = conn.Close(ctx)

	// Now connect to test database and run migrations
	testDBURL := "postgres://postgres:postgres@localhost:5432/jazz_test?sslmode=disable"
	testDB, err = SetupTestDB(testDBURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup test database: %v\n", err)
		os.Exit(1)
	}

	// Run all tests
	code := m.Run()

	// Cleanup: close connection
	TeardownTestDB(testDB)

	// Drop test database
	conn, err = pgx.Connect(ctx, dbURL)
	if err == nil {
		_, _ = conn.Exec(ctx, "DROP DATABASE IF EXISTS jazz_test")
		_ = conn.Close(ctx)
	}

	os.Exit(code)
}
