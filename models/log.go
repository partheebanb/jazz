package models

import (
	"time"

	"github.com/google/uuid"
)

type LogEntry struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	Level     string    `json:"level" binding:"required"`
	Message   string    `json:"message" binding:"required"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
	Rank      *float64  `json:"rank,omitempty"` // Only populated for search results
}

type QueryParams struct {
	Level     string `form:"level"`
	Source    string `form:"source"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
	Limit     int    `form:"limit"`
	Offset    int    `form:"offset"`
	Search    string `form:"search"`
}

type SearchRequest struct {
	Query     string `json:"query" binding:"required,min=3"`
	Level     string `json:"level"`
	Source    string `json:"source"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

type LogsResponse struct {
	Logs        []LogEntry `json:"logs"`
	Total       int64      `json:"total"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
	HasMore     bool       `json:"has_more"`
	QueryTimeMs *int64     `json:"query_time_ms,omitempty"`
}
