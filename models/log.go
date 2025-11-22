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
}

type QueryParams struct {
	Level     string `form:"level"`
	Source    string `form:"source"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
	Limit     int    `form:"limit"`
	Offset    int    `form:"offset"`
}

type LogsResponse struct {
	Logs    []LogEntry `json:"logs"`
	Total   int64      `json:"total"`
	Limit   int        `json:"limit"`
	Offset  int        `json:"offset"`
	HasMore bool       `json:"has_more"`
}
