package database

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateProject(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := GetTestDB()
	CleanupTestDB(t, db)

	ctx := context.Background()
	project, err := db.CreateProject(ctx, "Test Project")

	require.NoError(t, err)
	assert.NotEmpty(t, project.ID)
	assert.Equal(t, "Test Project", project.Name)
	assert.NotEmpty(t, project.APIKey)
	assert.True(t, len(project.APIKey) > 10, "API key should be generated")
	assert.False(t, project.CreatedAt.IsZero())
	assert.False(t, project.UpdatedAt.IsZero())
}

func TestGetProjectByAPIKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := GetTestDB()
	CleanupTestDB(t, db)

	ctx := context.Background()

	created, err := db.CreateProject(ctx, "Test Project")
	require.NoError(t, err)

	retrieved, err := db.GetProjectByAPIKey(ctx, created.APIKey)
	require.NoError(t, err)

	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Name, retrieved.Name)
	assert.Equal(t, created.APIKey, retrieved.APIKey)
}

func TestGetProjectByAPIKey_Invalid(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := GetTestDB()
	CleanupTestDB(t, db)

	ctx := context.Background()

	_, err := db.GetProjectByAPIKey(ctx, "invalid_key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid API key")
}

func TestListProjects(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := GetTestDB()
	CleanupTestDB(t, db)

	ctx := context.Background()

	projects, err := db.ListProjects(ctx)
	require.NoError(t, err)
	assert.Empty(t, projects)

	_, err = db.CreateProject(ctx, "Project 1")
	require.NoError(t, err)
	_, err = db.CreateProject(ctx, "Project 2")
	require.NoError(t, err)
	_, err = db.CreateProject(ctx, "Project 3")
	require.NoError(t, err)

	projects, err = db.ListProjects(ctx)
	require.NoError(t, err)
	assert.Len(t, projects, 3)
}

func TestGetProject(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := GetTestDB()
	CleanupTestDB(t, db)

	ctx := context.Background()

	created, err := db.CreateProject(ctx, "Test Project")
	require.NoError(t, err)

	retrieved, err := db.GetProject(ctx, created.ID)
	require.NoError(t, err)

	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Name, retrieved.Name)
}

func TestGetProject_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := GetTestDB()
	CleanupTestDB(t, db)

	ctx := context.Background()

	_, err := db.GetProject(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteProject(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := GetTestDB()
	CleanupTestDB(t, db)

	ctx := context.Background()

	created, err := db.CreateProject(ctx, "Test Project")
	require.NoError(t, err)

	err = db.DeleteProject(ctx, created.ID)
	require.NoError(t, err)

	_, err = db.GetProject(ctx, created.ID)
	assert.Error(t, err)
}

func TestDeleteProject_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := GetTestDB()
	CleanupTestDB(t, db)

	ctx := context.Background()

	err := db.DeleteProject(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
