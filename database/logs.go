package database

import (
	"context"
	"fmt"
	"jazz/models"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	defaultLimit = 50
	maxLimit     = 1000
)

// BatchInsertError indicates which log failed during batch insert.
// Contains the index of the failed log and the total batch size for debugging.
type BatchInsertError struct {
	FailedIndex int
	TotalLogs   int
	Err         error
}

func (e *BatchInsertError) Error() string {
	return fmt.Sprintf("failed to insert log at index %d/%d: %v", e.FailedIndex, e.TotalLogs, e.Err)
}

// InsertLogsBatch inserts multiple log entries atomically using pgx batching.
// All logs are inserted in a single network round-trip for performance.
// If any log fails, returns BatchInsertError indicating which log failed.
// Empty slice is a no-op and returns nil.
// All logs must belong to the same project (not enforced, caller's responsibility).
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
	for _, logEntry := range logs {
		batch.Queue(query, logEntry.ID, logEntry.ProjectID, logEntry.Level,
			logEntry.Message, logEntry.Source, logEntry.Timestamp)
	}

	results := db.Pool.SendBatch(ctx, batch)
	defer func() {
		_ = results.Close()
	}()

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

// QueryLogs retrieves logs for a project with optional filtering and pagination.
// If params.Search is provided, delegates to SearchLogs for full-text search.
// Uses COUNT(*) OVER() window function to get total count in single query.
// Returns logs ordered by timestamp DESC (newest first), total count, and any error.
//
// Filters applied:
//   - Level: exact match (e.g., "error", "info")
//   - Source: exact match (e.g., "backend", "frontend")
//   - StartTime/EndTime: inclusive timestamp range (RFC3339 format)
//   - Limit: max results (default 50, max 1000)
//   - Offset: pagination offset (default 0)
//
// Returns empty slice (not nil) if no logs match.
func (db *DB) QueryLogs(ctx context.Context, projectID uuid.UUID, params models.QueryParams) ([]models.LogEntry, int64, error) {
	start := time.Now()
	defer func() {
		log.Printf("QueryLogs: duration=%v project=%s filters=[level=%s source=%s search=%s]",
			time.Since(start), projectID, params.Level, params.Source, params.Search)
	}()

	// If search parameter provided, use SearchLogs instead
	if params.Search != "" {
		searchReq := models.SearchRequest{
			Query:     params.Search,
			Level:     params.Level,
			Source:    params.Source,
			StartTime: params.StartTime,
			EndTime:   params.EndTime,
			Limit:     params.Limit,
			Offset:    params.Offset,
		}
		return db.SearchLogs(ctx, projectID, searchReq)
	}

	// Validate pagination
	limit := validateLimit(params.Limit, defaultLimit, maxLimit)
	offset := validateOffset(params.Offset)

	// Build query
	qb := NewQueryBuilder()
	qb.AddCondition(columnProjectID, projectID)

	if params.Level != "" {
		qb.AddCondition(columnLevel, params.Level)
	}
	if params.Source != "" {
		qb.AddCondition(columnSource, params.Source)
	}
	if err := qb.AddTimeRange(columnTimestamp, params.StartTime, params.EndTime); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT 
			%s, %s, %s, %s, %s, %s,
			COUNT(*) OVER() as total_count
		FROM logs
		%s
		ORDER BY %s DESC
		LIMIT $%d OFFSET $%d
	`, columnID, columnProjectID, columnLevel, columnMessage, columnSource, columnTimestamp,
		qb.WhereClause(), columnTimestamp, qb.NextArgNum(), qb.NextArgNum()+1)

	args := append(qb.Args(), limit, offset)

	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	return scanLogs(rows, false)
}

// Helper functions

func scanLog(row rowScanner, includeRank bool) (*models.LogEntry, int64, error) {
	var log models.LogEntry
	var total int64
	var rank float64

	if includeRank {
		err := row.Scan(
			&log.ID, &log.ProjectID, &log.Level, &log.Message,
			&log.Source, &log.Timestamp, &rank, &total,
		)
		if err != nil {
			return nil, 0, err
		}
		log.Rank = &rank
	} else {
		err := row.Scan(
			&log.ID, &log.ProjectID, &log.Level, &log.Message,
			&log.Source, &log.Timestamp, &total,
		)
		if err != nil {
			return nil, 0, err
		}
	}

	return &log, total, nil
}

func scanLogs(rows rowsScanner, includeRank bool) ([]models.LogEntry, int64, error) {
	logs := []models.LogEntry{}
	var total int64

	for rows.Next() {
		log, t, err := scanLog(rows, includeRank)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan log: %w", err)
		}
		total = t
		logs = append(logs, *log)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating logs: %w", err)
	}

	return logs, total, nil
}
