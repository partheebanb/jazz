// Package models defines data structures used throughout Jazz.
// All models use JSON tags for API serialization and validation tags for input validation.
package models

import (
	"time"

	"github.com/google/uuid"
)

// LogEntry represents a single log message in the system.
// Used for both API requests/responses and database persistence.
// Timestamp and ID are auto-generated if not provided during ingestion.
type LogEntry struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	Level     string    `json:"level" binding:"required"`
	Message   string    `json:"message" binding:"required"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
	Rank      *float64  `json:"rank,omitempty"` // Only populated for search results
}

// QueryParams defines filtering and pagination options for log queries.
// All fields are optional - empty values are ignored.
// Used with GET /logs endpoint.
type QueryParams struct {
	Level     string `form:"level"`
	Source    string `form:"source"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
	Limit     int    `form:"limit"`
	Offset    int    `form:"offset"`
	Search    string `form:"search"`
}

// SearchRequest defines parameters for full-text search.
// Query field is required and must be at least 3 characters.
// Other fields are optional filters applied after search.
type SearchRequest struct {
	Query     string `json:"query" binding:"required,min=3"`
	Level     string `json:"level"`
	Source    string `json:"source"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

// LogsResponse is the standard response format for log queries.
// Includes pagination metadata to support infinite scroll or pagination UI.
// HasMore indicates if there are additional results beyond current page.
type LogsResponse struct {
	Logs        []LogEntry `json:"logs"`
	Total       int64      `json:"total"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
	HasMore     bool       `json:"has_more"`
	QueryTimeMs *int64     `json:"query_time_ms,omitempty"`
}
