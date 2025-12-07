package database

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestMain(m *testing.M) {
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		testDBURL = "postgres://postgres:postgres@localhost:5432/jazz_test?sslmode=disable"
	}

	postgresURL := strings.Replace(testDBURL, "/jazz_test", "/postgres", 1)

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, postgresURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to postgres: %v\n", err)
		os.Exit(1)
	}

	_, _ = conn.Exec(ctx, "DROP DATABASE IF EXISTS jazz_test")

	_, err = conn.Exec(ctx, "CREATE DATABASE jazz_test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create test database: %v\n", err)
		_ = conn.Close(ctx)
		os.Exit(1)
	}

	_ = conn.Close(ctx)

	testDB, err = SetupTestDB(testDBURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup test database: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	TeardownTestDB(testDB)

	conn, err = pgx.Connect(ctx, postgresURL)
	if err == nil {
		_, _ = conn.Exec(ctx, "DROP DATABASE IF EXISTS jazz_test")
		_ = conn.Close(ctx)
	}

	os.Exit(code)
}
