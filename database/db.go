package database

import (
	"context"
	"fmt"
	"jazz/models"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func Connect(databaseURL string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) InsertLog(log models.LogEntry) error {
	query := `
		INSERT INTO logs (id, level, message, source, timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := db.Pool.Exec(
		context.Background(),
		query,
		log.ID,
		log.Level,
		log.Message,
		log.Source,
		log.Timestamp,
	)
	return err
}

func (db *DB) QueryLogs(ctx context.Context, params models.QueryParams) ([]models.LogEntry, int64, error) {
	// Build WHERE clause dynamically
	conditions := []string{}
	args := []interface{}{}
	argCount := 1

	if params.Level != "" {
		conditions = append(conditions, fmt.Sprintf("level = $%d", argCount))
		args = append(args, params.Level)
		argCount++
	}

	if params.Source != "" {
		conditions = append(conditions, fmt.Sprintf("source = $%d", argCount))
		args = append(args, params.Source)
		argCount++
	}

	if params.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, params.StartTime)
		if err == nil {
			conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argCount))
			args = append(args, startTime)
			argCount++
		}
	}

	if params.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, params.EndTime)
		if err == nil {
			conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argCount))
			args = append(args, endTime)
			argCount++
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Set defaults
	if params.Limit == 0 {
		params.Limit = 50
	}
	if params.Limit > 1000 {
		params.Limit = 1000
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM logs %s", whereClause)
	var total int64
	err := db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get logs
	query := fmt.Sprintf(`
		SELECT id, level, message, source, timestamp
		FROM logs
		%s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	args = append(args, params.Limit, params.Offset)

	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs := []models.LogEntry{} // Initialize as empty slice, not nil
	for rows.Next() {
		var log models.LogEntry
		err := rows.Scan(&log.ID, &log.Level, &log.Message, &log.Source, &log.Timestamp)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}
func (db *DB) InsertLogsBatch(ctx context.Context, logs []models.LogEntry) error {
	if len(logs) == 0 {
		return nil
	}

	query := `
		INSERT INTO logs (id, level, message, source, timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`

	batch := &pgx.Batch{}
	for _, log := range logs {
		batch.Queue(query, log.ID, log.Level, log.Message, log.Source, log.Timestamp)
	}

	results := db.Pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(logs); i++ {
		_, err := results.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
