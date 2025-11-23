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

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		conditions: []string{},
		args:       []interface{}{},
		argCount:   1,
	}
}

func (qb *QueryBuilder) AddCondition(column string, value interface{}) {
	qb.conditions = append(qb.conditions, fmt.Sprintf("%s = $%d", column, qb.argCount))
	qb.args = append(qb.args, value)
	qb.argCount++
}

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

func (qb *QueryBuilder) AddFullTextSearch(searchQuery string) {
	qb.conditions = append(qb.conditions,
		fmt.Sprintf("to_tsvector('english', %s) @@ to_tsquery('english', $%d)", columnMessage, qb.argCount))
	qb.args = append(qb.args, searchQuery)
	qb.argCount++
}

func (qb *QueryBuilder) WhereClause() string {
	if len(qb.conditions) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(qb.conditions, " AND ")
}

func (qb *QueryBuilder) Args() []interface{} {
	return qb.args
}

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
