package database

import (
	"context"
	"fmt"
	"jazz/models"
	"log"
	"strings"
	"time"

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
func (db *DB) QueryLogs(ctx context.Context, params models.QueryParams) ([]models.LogEntry, int64, error) {
	start := time.Now()
	defer func() {
		log.Printf("QueryLogs: duration=%v filters=[level=%s source=%s]",
			time.Since(start), params.Level, params.Source)
	}()

	conditions := []string{}
	args := []interface{}{}
	argCount := 1

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

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Single query with COUNT(*) OVER() to avoid N+1 problem
	// SAFETY: All user input is parameterized via $N placeholders.
	// whereClause only contains safe column names and SQL operators.
	query := fmt.Sprintf(`
		SELECT 
			%s, %s, %s, %s, %s,
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

	query := fmt.Sprintf(`
		INSERT INTO logs (%s, %s, %s, %s, %s)
		VALUES ($1, $2, $3, $4, $5)
	`, columnID, columnLevel, columnMessage, columnSource, columnTimestamp)

	batch := &pgx.Batch{}
	for _, log := range logs {
		batch.Queue(query, log.ID, log.Level, log.Message, log.Source, log.Timestamp)
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
