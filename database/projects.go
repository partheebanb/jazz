package database

import (
	"context"
	"fmt"
	"jazz/models"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetProjectByAPIKey validates an API key and returns the associated project.
// Used by authentication middleware to verify requests.
// Returns error with "invalid API key" message if key not found (safe to expose to client).
// Returns error with technical details if database fails (log server-side only).
func (db *DB) GetProjectByAPIKey(ctx context.Context, apiKey string) (*models.Project, error) {
	query := `
		SELECT id, name, api_key, created_at, updated_at
		FROM projects
		WHERE api_key = $1
	`

	project, err := scanProject(db.Pool.QueryRow(ctx, query, apiKey))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("invalid API key")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

// CreateProject creates a new project with a generated API key.
// API key format: "jazz_" + UUID v4 (e.g., "jazz_a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11").
// API key is returned in plaintext only once - store it securely.
// Returns the created project with all fields populated, including timestamps.
func (db *DB) CreateProject(ctx context.Context, name string) (*models.Project, error) {
	apiKey := generateAPIKey()

	query := `
		INSERT INTO projects (name, api_key)
		VALUES ($1, $2)
		RETURNING id, name, api_key, created_at, updated_at
	`

	project, err := scanProject(db.Pool.QueryRow(ctx, query, name, apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	log.Printf("Created project: %s (ID: %s)", project.Name, project.ID)
	return project, nil
}

// ListProjects returns all projects ordered by creation date (newest first).
// No pagination - suitable for small number of projects (<1000).
// Returns empty slice (not nil) if no projects exist.
func (db *DB) ListProjects(ctx context.Context) ([]models.Project, error) {
	query := `
		SELECT id, name, api_key, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	return scanProjects(rows)
}

// GetProject retrieves a single project by ID.
// Returns error with "project not found" if ID doesn't exist.
// Used for project detail views and validation.
func (db *DB) GetProject(ctx context.Context, projectID uuid.UUID) (*models.Project, error) {
	query := `
		SELECT id, name, api_key, created_at, updated_at
		FROM projects
		WHERE id = $1
	`

	project, err := scanProject(db.Pool.QueryRow(ctx, query, projectID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

// DeleteProject removes a project and all its logs (CASCADE).
// This is a destructive operation that cannot be undone.
// Returns error with "project not found" if ID doesn't exist.
// Logs the deletion for audit trail.
func (db *DB) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1`

	result, err := db.Pool.Exec(ctx, query, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("project not found")
	}

	log.Printf("Deleted project: %s", projectID)
	return nil
}

// Helper functions

func generateAPIKey() string {
	return fmt.Sprintf("jazz_%s", uuid.New().String())
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanProject(row rowScanner) (*models.Project, error) {
	var project models.Project
	err := row.Scan(
		&project.ID,
		&project.Name,
		&project.APIKey,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

type rowsScanner interface {
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
}

func scanProjects(rows rowsScanner) ([]models.Project, error) {
	projects := []models.Project{}
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, *project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating projects: %w", err)
	}

	return projects, nil
}
