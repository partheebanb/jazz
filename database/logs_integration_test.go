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

func TestInsertLogsBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := GetTestDB()
	CleanupTestDB(t, db)

	ctx := context.Background()

	project, err := db.CreateProject(ctx, "Test Project")
	require.NoError(t, err)

	logs := []models.LogEntry{
		{
			ID:        uuid.New(),
			ProjectID: project.ID,
			Level:     "error",
			Message:   "Test error 1",
			Source:    "backend",
			Timestamp: time.Now(),
		},
		{
			ID:        uuid.New(),
			ProjectID: project.ID,
			Level:     "info",
			Message:   "Test info 1",
			Source:    "frontend",
			Timestamp: time.Now(),
		},
	}

	err = db.InsertLogsBatch(ctx, logs)
	require.NoError(t, err)

	results, total, err := db.QueryLogs(ctx, project.ID, models.QueryParams{
		Limit: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, int64(2), total)
}

func TestInsertLogsBatch_Empty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := GetTestDB()
	ctx := context.Background()

	err := db.InsertLogsBatch(ctx, []models.LogEntry{})
	assert.NoError(t, err)
}

func TestQueryLogs_Filtering(t *testing.T) {
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
		{ID: uuid.New(), ProjectID: project.ID, Level: "error", Message: "Error 1", Source: "backend", Timestamp: now},
		{ID: uuid.New(), ProjectID: project.ID, Level: "error", Message: "Error 2", Source: "backend", Timestamp: now},
		{ID: uuid.New(), ProjectID: project.ID, Level: "info", Message: "Info 1", Source: "frontend", Timestamp: now},
		{ID: uuid.New(), ProjectID: project.ID, Level: "warning", Message: "Warning 1", Source: "backend", Timestamp: now},
	}
	err = db.InsertLogsBatch(ctx, logs)
	require.NoError(t, err)

	tests := []struct {
		name          string
		params        models.QueryParams
		expectedCount int
	}{
		{
			name:          "no filters",
			params:        models.QueryParams{Limit: 10},
			expectedCount: 4,
		},
		{
			name:          "filter by level",
			params:        models.QueryParams{Level: "error", Limit: 10},
			expectedCount: 2,
		},
		{
			name:          "filter by source",
			params:        models.QueryParams{Source: "backend", Limit: 10},
			expectedCount: 3,
		},
		{
			name:          "filter by level and source",
			params:        models.QueryParams{Level: "error", Source: "backend", Limit: 10},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, _, err := db.QueryLogs(ctx, project.ID, tt.params)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(results))
		})
	}
}

func TestQueryLogs_Pagination(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := GetTestDB()
	CleanupTestDB(t, db)

	ctx := context.Background()

	project, err := db.CreateProject(ctx, "Test Project")
	require.NoError(t, err)

	now := time.Now()
	logs := make([]models.LogEntry, 25)
	for i := 0; i < 25; i++ {
		logs[i] = models.LogEntry{
			ID:        uuid.New(),
			ProjectID: project.ID,
			Level:     "info",
			Message:   "Test log",
			Timestamp: now.Add(time.Duration(i) * time.Second),
		}
	}
	err = db.InsertLogsBatch(ctx, logs)
	require.NoError(t, err)

	page1, total, err := db.QueryLogs(ctx, project.ID, models.QueryParams{
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	assert.Equal(t, 10, len(page1))
	assert.Equal(t, int64(25), total)

	page2, total, err := db.QueryLogs(ctx, project.ID, models.QueryParams{
		Limit:  10,
		Offset: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, 10, len(page2))
	assert.Equal(t, int64(25), total)

	page3, total, err := db.QueryLogs(ctx, project.ID, models.QueryParams{
		Limit:  10,
		Offset: 20,
	})
	require.NoError(t, err)
	assert.Equal(t, 5, len(page3))
	assert.Equal(t, int64(25), total)
}

func TestQueryLogs_ProjectIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := GetTestDB()
	CleanupTestDB(t, db)

	ctx := context.Background()

	project1, err := db.CreateProject(ctx, "Project 1")
	require.NoError(t, err)
	project2, err := db.CreateProject(ctx, "Project 2")
	require.NoError(t, err)

	logs1 := []models.LogEntry{
		{ID: uuid.New(), ProjectID: project1.ID, Level: "error", Message: "P1 Log", Timestamp: time.Now()},
	}
	err = db.InsertLogsBatch(ctx, logs1)
	require.NoError(t, err)

	logs2 := []models.LogEntry{
		{ID: uuid.New(), ProjectID: project2.ID, Level: "error", Message: "P2 Log", Timestamp: time.Now()},
	}
	err = db.InsertLogsBatch(ctx, logs2)
	require.NoError(t, err)

	results, total, err := db.QueryLogs(ctx, project1.ID, models.QueryParams{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "P1 Log", results[0].Message)

	results, total, err = db.QueryLogs(ctx, project2.ID, models.QueryParams{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "P2 Log", results[0].Message)
}
