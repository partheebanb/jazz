package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryBuilder_AddCondition(t *testing.T) {
	qb := NewQueryBuilder()

	qb.AddCondition("level", "error")

	assert.Equal(t, "WHERE level = $1", qb.WhereClause())
	assert.Equal(t, []interface{}{"error"}, qb.Args())
	assert.Equal(t, 2, qb.NextArgNum())
}

func TestQueryBuilder_MultipleConditions(t *testing.T) {
	qb := NewQueryBuilder()

	qb.AddCondition("level", "error")
	qb.AddCondition("source", "backend")
	qb.AddCondition("project_id", "123")

	assert.Equal(t, "WHERE level = $1 AND source = $2 AND project_id = $3", qb.WhereClause())
	assert.Equal(t, []interface{}{"error", "backend", "123"}, qb.Args())
	assert.Equal(t, 4, qb.NextArgNum())
}

func TestQueryBuilder_AddTimeRange(t *testing.T) {
	tests := []struct {
		name           string
		startTime      string
		endTime        string
		wantConditions int
		wantErr        bool
	}{
		{
			name:           "both start and end",
			startTime:      "2024-11-01T00:00:00Z",
			endTime:        "2024-11-22T23:59:59Z",
			wantConditions: 2,
			wantErr:        false,
		},
		{
			name:           "only start",
			startTime:      "2024-11-01T00:00:00Z",
			endTime:        "",
			wantConditions: 1,
			wantErr:        false,
		},
		{
			name:           "only end",
			startTime:      "",
			endTime:        "2024-11-22T23:59:59Z",
			wantConditions: 1,
			wantErr:        false,
		},
		{
			name:           "neither",
			startTime:      "",
			endTime:        "",
			wantConditions: 0,
			wantErr:        false,
		},
		{
			name:      "invalid start time",
			startTime: "not-a-date",
			endTime:   "",
			wantErr:   true,
		},
		{
			name:      "invalid end time",
			startTime: "",
			endTime:   "not-a-date",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder()
			err := qb.AddTimeRange("timestamp", tt.startTime, tt.endTime)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, qb.Args(), tt.wantConditions)
			}
		})
	}
}

func TestQueryBuilder_AddFullTextSearch(t *testing.T) {
	qb := NewQueryBuilder()

	qb.AddFullTextSearch("database & error")

	assert.Contains(t, qb.WhereClause(), "to_tsvector")
	assert.Contains(t, qb.WhereClause(), "to_tsquery")
	assert.Equal(t, []interface{}{"database & error"}, qb.Args())
}

func TestQueryBuilder_WhereClause_Empty(t *testing.T) {
	qb := NewQueryBuilder()

	assert.Equal(t, "", qb.WhereClause())
	assert.Empty(t, qb.Args())
}

func TestQueryBuilder_ComplexQuery(t *testing.T) {
	qb := NewQueryBuilder()

	qb.AddCondition("project_id", "abc-123")
	qb.AddCondition("level", "error")
	err := qb.AddTimeRange("timestamp", "2024-11-01T00:00:00Z", "2024-11-22T23:59:59Z")
	require.NoError(t, err)
	qb.AddFullTextSearch("database & timeout")

	whereClause := qb.WhereClause()

	assert.Contains(t, whereClause, "project_id = $1")
	assert.Contains(t, whereClause, "level = $2")
	assert.Contains(t, whereClause, "timestamp >= $3")
	assert.Contains(t, whereClause, "timestamp <= $4")
	assert.Contains(t, whereClause, "to_tsvector")
	assert.Len(t, qb.Args(), 5)
}
