package database

import (
	"context"
	"fmt"
	"jazz/models"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SearchQueryParser struct {
	minLength int
	maxLength int
}

func NewSearchQueryParser() *SearchQueryParser {
	return &SearchQueryParser{
		minLength: 3,
		maxLength: 1000,
	}
}

func (p *SearchQueryParser) Parse(query string) (string, error) {
	query = strings.TrimSpace(query)

	if len(query) < p.minLength {
		return "", fmt.Errorf("search query must be at least %d characters", p.minLength)
	}

	if len(query) > p.maxLength {
		return "", fmt.Errorf("search query too long (max %d characters)", p.maxLength)
	}

	// Clean the query
	query = p.sanitize(query)

	// Split and validate
	words := strings.Fields(query)
	if len(words) == 0 {
		return "", fmt.Errorf("search query is empty")
	}

	// Filter out short words
	validWords := p.filterValidWords(words)
	if len(validWords) == 0 {
		return "", fmt.Errorf("no valid search terms")
	}

	return strings.Join(validWords, " & "), nil
}

func (p *SearchQueryParser) sanitize(query string) string {
	replacements := map[string]string{
		`"`: "",
		"'": "",
		"(": "",
		")": "",
	}

	for old, new := range replacements {
		query = strings.ReplaceAll(query, old, new)
	}

	return query
}

func (p *SearchQueryParser) filterValidWords(words []string) []string {
	valid := []string{}
	for _, word := range words {
		if len(word) >= 2 {
			valid = append(valid, strings.ToLower(word))
		}
	}
	return valid
}

func (db *DB) SearchLogs(ctx context.Context, projectID uuid.UUID, req models.SearchRequest) ([]models.LogEntry, int64, error) {
	start := time.Now()
	defer func() {
		log.Printf("SearchLogs: project=%s query=%q duration=%dms",
			projectID, req.Query, time.Since(start).Milliseconds())
	}()

	// Parse search query
	parser := NewSearchQueryParser()
	tsQuery, err := parser.Parse(req.Query)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid search query: %w", err)
	}

	// Validate pagination
	limit := validateLimit(req.Limit, defaultLimit, maxLimit)
	offset := validateOffset(req.Offset)

	// Build query
	qb := NewQueryBuilder()
	qb.AddCondition(columnProjectID, projectID)
	qb.AddFullTextSearch(tsQuery)

	if req.Level != "" {
		qb.AddCondition(columnLevel, req.Level)
	}
	if req.Source != "" {
		qb.AddCondition(columnSource, req.Source)
	}
	if err := qb.AddTimeRange(columnTimestamp, req.StartTime, req.EndTime); err != nil {
		return nil, 0, err
	}

	// SAFETY: All user input is parameterized. whereClause only contains safe SQL.
	query := fmt.Sprintf(`
		SELECT 
			%s, %s, %s, %s, %s, %s,
			ts_rank(to_tsvector('english', %s), to_tsquery('english', $2)) as rank,
			COUNT(*) OVER() as total_count
		FROM logs
		%s
		ORDER BY rank DESC, %s DESC
		LIMIT $%d OFFSET $%d
	`, columnID, columnProjectID, columnLevel, columnMessage, columnSource, columnTimestamp,
		columnMessage, qb.WhereClause(), columnTimestamp, qb.NextArgNum(), qb.NextArgNum()+1)

	args := append(qb.Args(), limit, offset)

	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search logs: %w", err)
	}
	defer rows.Close()

	return scanLogs(rows, true)
}
