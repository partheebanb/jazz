package database

import (
	"context"
	"jazz/models"

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
		INSERT INTO logs (id, level, message, timestamp)
		VALUES ($1, $2, $3, $4)
	`
	_, err := db.Pool.Exec(
		context.Background(),
		query,
		log.ID,
		log.Level,
		log.Message,
		log.Timestamp,
	)
	return err
}

func (db *DB) GetLogs() ([]models.LogEntry, error) {
	query := `
		SELECT id, level, message, timestamp
		FROM logs
		ORDER BY timestamp DESC
		LIMIT 100
	`
	rows, err := db.Pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.LogEntry
	for rows.Next() {
		var log models.LogEntry
		err := rows.Scan(&log.ID, &log.Level, &log.Message, &log.Timestamp)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}