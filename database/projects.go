package database

import (
	"context"
	"fmt"
	"jazz/models"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

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
