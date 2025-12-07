package database

import (
	"fmt"
	"strings"
	"time"
)

const (
	columnID        = "id"
	columnProjectID = "project_id"
	columnLevel     = "level"
	columnMessage   = "message"
	columnSource    = "source"
	columnTimestamp = "timestamp"
)

// QueryBuilder helps build WHERE clauses safely
type QueryBuilder struct {
	conditions []string
	args       []interface{}
	argCount   int
}

// NewQueryBuilder creates a QueryBuilder for constructing safe SQL WHERE clauses.
// All values are parameterized to prevent SQL injection.
// Argument numbering starts at $1 and increments automatically.
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		conditions: []string{},
		args:       []interface{}{},
		argCount:   1,
	}
}

// AddCondition adds an equality condition to the WHERE clause.
// Automatically parameterizes the value (safe from SQL injection).
// Example: AddCondition("level", "error") → "level = $1" with args ["error"]
func (qb *QueryBuilder) AddCondition(column string, value interface{}) {
	qb.conditions = append(qb.conditions, fmt.Sprintf("%s = $%d", column, qb.argCount))
	qb.args = append(qb.args, value)
	qb.argCount++
}

// AddTimeRange adds timestamp range conditions (start and/or end).
// Times must be in RFC3339 format (e.g., "2024-11-22T10:30:00Z").
// Both parameters are optional - empty string skips that bound.
// Returns error if time format is invalid.
//
// Example:
//
//	AddTimeRange("timestamp", "2024-11-01T00:00:00Z", "2024-11-22T23:59:59Z")
//	→ "timestamp >= $1 AND timestamp <= $2"
func (qb *QueryBuilder) AddTimeRange(column, start, end string) error {
	if start != "" {
		startTime, err := parseRFC3339(start)
		if err != nil {
			return fmt.Errorf("invalid start_time: %w", err)
		}
		qb.conditions = append(qb.conditions, fmt.Sprintf("%s >= $%d", column, qb.argCount))
		qb.args = append(qb.args, startTime)
		qb.argCount++
	}

	if end != "" {
		endTime, err := parseRFC3339(end)
		if err != nil {
			return fmt.Errorf("invalid end_time: %w", err)
		}
		qb.conditions = append(qb.conditions, fmt.Sprintf("%s <= $%d", column, qb.argCount))
		qb.args = append(qb.args, endTime)
		qb.argCount++
	}

	return nil
}

// AddFullTextSearch adds PostgreSQL full-text search condition.
// Uses to_tsvector and to_tsquery for GIN index optimization.
// searchQuery must already be in tsquery format (e.g., "hello & world").
// Column is assumed to be 'message' - this is hardcoded for Jazz's use case.
func (qb *QueryBuilder) AddFullTextSearch(searchQuery string) {
	qb.conditions = append(qb.conditions,
		fmt.Sprintf("to_tsvector('english', %s) @@ to_tsquery('english', $%d)", columnMessage, qb.argCount))
	qb.args = append(qb.args, searchQuery)
	qb.argCount++
}

// WhereClause returns the complete WHERE clause with all conditions.
// Conditions are joined with AND.
// Returns empty string if no conditions were added.
// Includes "WHERE " prefix for convenience.
func (qb *QueryBuilder) WhereClause() string {
	if len(qb.conditions) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(qb.conditions, " AND ")
}

// Args returns all parameterized arguments in order.
// Use these as arguments to db.Query() or db.Exec().
// Arguments are in the same order as $1, $2, $3, etc.
func (qb *QueryBuilder) Args() []interface{} {
	return qb.args
}

// NextArgNum returns the next parameter number to use ($N).
// Useful for manually adding LIMIT/OFFSET after building WHERE clause.
// Example: if 3 conditions added, NextArgNum() returns 4.
func (qb *QueryBuilder) NextArgNum() int {
	return qb.argCount
}

// Helper functions

func parseRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func validateLimit(limit, defaultLimit, maxLimit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

func validateOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}
