package database

import (
	"context"
	"fmt"
	"jazz/models"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	columnID        = "id"
	columnLevel     = "level"
	columnMessage   = "message"
	columnSource    = "source"
	columnTimestamp = "timestamp"
)

type DB struct {
	Pool *pgxpool.Pool
}

type BatchInsertError struct {
	FailedIndex int
	TotalLogs   int
	Err         error
}

func (e *BatchInsertError) Error() string {
	return fmt.Sprintf("failed to insert log at index %d/%d: %v", e.FailedIndex, e.TotalLogs, e.Err)
}

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

// QueryLogs retrieves logs with filtering and pagination
// Uses COUNT(*) OVER() to get total count in single query
func (db *DB) QueryLogs(ctx context.Context, projectID uuid.UUID, params models.QueryParams) ([]models.LogEntry, int64, error) {
	start := time.Now()
	defer func() {
		log.Printf("QueryLogs: duration=%v project=%s filters=[level=%s source=%s]",
			time.Since(start), projectID, params.Level, params.Source)
	}()

	conditions := []string{"project_id = $1"}
	args := []interface{}{projectID}
	argCount := 2

	if params.Level != "" {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", columnLevel, argCount))
		args = append(args, params.Level)
		argCount++
	}

	if params.Source != "" {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", columnSource, argCount))
		args = append(args, params.Source)
		argCount++
	}

	if params.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, params.StartTime)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid start_time format (expected RFC3339): %w", err)
		}
		conditions = append(conditions, fmt.Sprintf("%s >= $%d", columnTimestamp, argCount))
		args = append(args, startTime)
		argCount++
	}

	if params.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, params.EndTime)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid end_time format (expected RFC3339): %w", err)
		}
		conditions = append(conditions, fmt.Sprintf("%s <= $%d", columnTimestamp, argCount))
		args = append(args, endTime)
		argCount++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	query := fmt.Sprintf(`
		SELECT 
			%s, project_id, %s, %s, %s, %s,
			COUNT(*) OVER() as total_count
		FROM logs
		%s
		ORDER BY %s DESC
		LIMIT $%d OFFSET $%d
	`, columnID, columnLevel, columnMessage, columnSource, columnTimestamp,
		whereClause, columnTimestamp, argCount, argCount+1)

	args = append(args, params.Limit, params.Offset)

	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	logs := []models.LogEntry{}
	var total int64

	for rows.Next() {
		var log models.LogEntry
		err := rows.Scan(
			&log.ID,
			&log.ProjectID,
			&log.Level,
			&log.Message,
			&log.Source,
			&log.Timestamp,
			&total,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan log row: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating rows: %w", err)
	}

	return logs, total, nil
}

func (db *DB) InsertLogsBatch(ctx context.Context, logs []models.LogEntry) error {
	if len(logs) == 0 {
		return nil
	}

	start := time.Now()
	defer func() {
		log.Printf("InsertLogsBatch: duration=%v count=%d", time.Since(start), len(logs))
	}()

	query := `
		INSERT INTO logs (id, project_id, level, message, source, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	batch := &pgx.Batch{}
	for _, log := range logs {
		batch.Queue(query, log.ID, log.ProjectID, log.Level, log.Message, log.Source, log.Timestamp)
	}

	results := db.Pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(logs); i++ {
		_, err := results.Exec()
		if err != nil {
			return &BatchInsertError{
				FailedIndex: i,
				TotalLogs:   len(logs),
				Err:         err,
			}
		}
	}

	return nil
}

func (db *DB) Close() {
	db.Pool.Close()
	log.Println("Database connection closed")
}

// GetProjectByAPIKey validates an API key and returns the project
func (db *DB) GetProjectByAPIKey(ctx context.Context, apiKey string) (*models.Project, error) {
	query := `
		SELECT id, name, api_key, created_at, updated_at
		FROM projects
		WHERE api_key = $1
	`

	var project models.Project
	err := db.Pool.QueryRow(ctx, query, apiKey).Scan(
		&project.ID,
		&project.Name,
		&project.APIKey,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("invalid API key")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// CreateProject creates a new project with a generated API key
func (db *DB) CreateProject(ctx context.Context, name string) (*models.Project, error) {
	apiKey := generateAPIKey()

	query := `
		INSERT INTO projects (name, api_key)
		VALUES ($1, $2)
		RETURNING id, name, api_key, created_at, updated_at
	`

	var project models.Project
	err := db.Pool.QueryRow(ctx, query, name, apiKey).Scan(
		&project.ID,
		&project.Name,
		&project.APIKey,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	log.Printf("Created project: %s (ID: %s)", project.Name, project.ID)
	return &project, nil
}

// ListProjects returns all projects
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

	projects := []models.Project{}
	for rows.Next() {
		var project models.Project
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.APIKey,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating projects: %w", err)
	}

	return projects, nil
}

// GetProject returns a project by ID
func (db *DB) GetProject(ctx context.Context, projectID uuid.UUID) (*models.Project, error) {
	query := `
		SELECT id, name, api_key, created_at, updated_at
		FROM projects
		WHERE id = $1
	`

	var project models.Project
	err := db.Pool.QueryRow(ctx, query, projectID).Scan(
		&project.ID,
		&project.Name,
		&project.APIKey,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// DeleteProject deletes a project (cascades to logs)
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

// generateAPIKey generates a secure random API key
func generateAPIKey() string {
	return fmt.Sprintf("jazz_%s", uuid.New().String())
}
