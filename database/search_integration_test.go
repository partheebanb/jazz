package database

import (
	"context"
	"jazz/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchLogs_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := GetTestDB()
	CleanupTestDB(t, db)

	ctx := context.Background()

	project, err := db.CreateProject(ctx, "Test Project")
	require.NoError(t, err)

	now := time.Now()
	logs := []models.LogEntry{
		{ID: uuid.New(), ProjectID: project.ID, Level: "error", Message: "Database connection failed", Timestamp: now},
		{ID: uuid.New(), ProjectID: project.ID, Level: "error", Message: "Payment gateway timeout", Timestamp: now},
		{ID: uuid.New(), ProjectID: project.ID, Level: "info", Message: "User logged in successfully", Timestamp: now},
	}
	err = db.InsertLogsBatch(ctx, logs)
	require.NoError(t, err)

	results, total, err := db.SearchLogs(ctx, project.ID, models.SearchRequest{
		Query: "database",
		Limit: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, int64(1), total)
	assert.Contains(t, results[0].Message, "Database")
	assert.NotNil(t, results[0].Rank)
}

func TestSearchLogs_MultipleWords(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := GetTestDB()
	CleanupTestDB(t, db)

	ctx := context.Background()

	project, err := db.CreateProject(ctx, "Test Project")
	require.NoError(t, err)

	now := time.Now()
	logs := []models.LogEntry{
		{ID: uuid.New(), ProjectID: project.ID, Level: "error", Message: "Database connection timeout", Timestamp: now},
		{ID: uuid.New(), ProjectID: project.ID, Level: "error", Message: "Database query failed", Timestamp: now},
		{ID: uuid.New(), ProjectID: project.ID, Level: "error", Message: "Payment timeout", Timestamp: now},
	}
	err = db.InsertLogsBatch(ctx, logs)
	require.NoError(t, err)

	results, total, err := db.SearchLogs(ctx, project.ID, models.SearchRequest{
		Query: "database timeout",
		Limit: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, int64(1), total)
	assert.Contains(t, results[0].Message, "Database connection timeout")
}

func TestSearchLogs_WithFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := GetTestDB()
	CleanupTestDB(t, db)

	ctx := context.Background()

	project, err := db.CreateProject(ctx, "Test Project")
	require.NoError(t, err)

	now := time.Now()
	logs := []models.LogEntry{
		{ID: uuid.New(), ProjectID: project.ID, Level: "error", Message: "Database error", Source: "backend", Timestamp: now},
		{ID: uuid.New(), ProjectID: project.ID, Level: "info", Message: "Database connected", Source: "backend", Timestamp: now},
		{ID: uuid.New(), ProjectID: project.ID, Level: "error", Message: "Database error", Source: "frontend", Timestamp: now},
	}
	err = db.InsertLogsBatch(ctx, logs)
	require.NoError(t, err)

	results, total, err := db.SearchLogs(ctx, project.ID, models.SearchRequest{
		Query:  "database",
		Level:  "error",
		Source: "backend",
		Limit:  10,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "error", results[0].Level)
	assert.Equal(t, "backend", results[0].Source)
}

func TestSearchLogs_Ranking(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := GetTestDB()
	CleanupTestDB(t, db)

	ctx := context.Background()

	project, err := db.CreateProject(ctx, "Test Project")
	require.NoError(t, err)

	now := time.Now()
	logs := []models.LogEntry{
		{ID: uuid.New(), ProjectID: project.ID, Level: "error", Message: "error", Timestamp: now},             // Low relevance
		{ID: uuid.New(), ProjectID: project.ID, Level: "error", Message: "error error error", Timestamp: now}, // Higher relevance
	}
	err = db.InsertLogsBatch(ctx, logs)
	require.NoError(t, err)

	results, _, err := db.SearchLogs(ctx, project.ID, models.SearchRequest{
		Query: "error",
		Limit: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))

	// First result should have higher rank (more occurrences)
	assert.Greater(t, *results[0].Rank, *results[1].Rank)
}
