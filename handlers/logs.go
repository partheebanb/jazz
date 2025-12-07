// Package handlers implements HTTP request handlers for the Jazz API.
// All handlers follow the Gin framework pattern and require database access.
package handlers

import (
	"fmt"
	"jazz/database"
	"jazz/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxBatchSize  = 1000
	minBatchSize  = 1
	defaultLimit  = 50
	maxLimit      = 1000
	defaultOffset = 0
)

// HealthCheck returns 200 OK with status message.
// Used by load balancers and monitoring systems to verify the API is running.
// Does not check database connectivity (see readiness probe for that).
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// IngestLogs accepts a batch of log entries and stores them in the database.
// Requires valid API key authentication (project_id in context).
// Validates batch size (1-1000 logs) and generates UUIDs/timestamps if missing.
//
// Request body:
//
//	{
//	  "logs": [
//	    {"level": "error", "message": "...", "source": "backend"}
//	  ]
//	}
//
// Returns 201 Created on success, 400 for validation errors, 500 for database errors.
// All logs in batch are inserted atomically - partial failures are not allowed.
func IngestLogs(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get project from auth middleware
		projectID, exists := c.Get("project_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var logs []models.LogEntry
		if err := c.ShouldBindJSON(&logs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if len(logs) < minBatchSize || len(logs) > maxBatchSize {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("batch size must be between %d and %d", minBatchSize, maxBatchSize),
			})
			return
		}

		now := time.Now()
		for i := range logs {
			logs[i].ID = uuid.New()
			logs[i].ProjectID = projectID.(uuid.UUID)
			if logs[i].Timestamp.IsZero() {
				logs[i].Timestamp = now
			}
		}

		ctx := c.Request.Context()
		if err := db.InsertLogsBatch(ctx, logs); err != nil {
			log.Printf("failed to insert logs: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to store logs",
			})
			return
		}

		log.Printf("ingested %d logs for project %s", len(logs), projectID)
		c.JSON(http.StatusCreated, gin.H{
			"message": "logs stored",
			"count":   len(logs),
		})
	}
}

// GetLogs retrieves logs for the authenticated project with optional filtering.
// Supports filtering by level, source, time range, pagination, and full-text search.
// If 'search' parameter is provided, performs full-text search instead of basic query.
//
// Query parameters:
//   - level: filter by log level (exact match)
//   - source: filter by source (exact match)
//   - start_time: RFC3339 timestamp (inclusive)
//   - end_time: RFC3339 timestamp (inclusive)
//   - limit: max results (default 50, max 1000)
//   - offset: pagination offset
//   - search: full-text search query (triggers SearchLogs)
//
// Response includes logs array, total count, and has_more flag for pagination.
func GetLogs(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, exists := c.Get("project_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var params models.QueryParams
		if err := c.ShouldBindQuery(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if params.Limit <= 0 {
			params.Limit = defaultLimit
		}
		if params.Limit > maxLimit {
			params.Limit = maxLimit
		}
		if params.Offset < 0 {
			params.Offset = defaultOffset
		}

		ctx := c.Request.Context()
		logs, total, err := db.QueryLogs(ctx, projectID.(uuid.UUID), params)
		if err != nil {
			log.Printf("failed to query logs: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to retrieve logs",
			})
			return
		}

		response := models.LogsResponse{
			Logs:    logs,
			Total:   total,
			Limit:   params.Limit,
			Offset:  params.Offset,
			HasMore: int64(params.Offset+params.Limit) < total,
		}

		c.JSON(http.StatusOK, response)
	}
}

// SearchLogs performs full-text search on log messages for the authenticated project.
// Uses PostgreSQL GIN indexes for fast search across large datasets.
// Results include relevance ranking and query execution time.
//
// Request body:
//
//	{
//	  "query": "database error",
//	  "level": "error",          // optional
//	  "source": "backend",        // optional
//	  "start_time": "...",        // optional
//	  "end_time": "...",          // optional
//	  "limit": 50,                // optional
//	  "offset": 0                 // optional
//	}
//
// Response includes logs with rank field, total count, and query_time_ms.
// Returns 400 for invalid queries (too short, etc.), 500 for database errors.
func SearchLogs(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, exists := c.Get("project_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req models.SearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set defaults
		if req.Limit == 0 {
			req.Limit = defaultLimit
		}
		if req.Limit > maxLimit {
			req.Limit = maxLimit
		}
		if req.Offset < 0 {
			req.Offset = defaultOffset
		}

		start := time.Now()
		ctx := c.Request.Context()
		logs, total, err := db.SearchLogs(ctx, projectID.(uuid.UUID), req)
		if err != nil {
			log.Printf("search error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "search failed",
			})
			return
		}
		queryTimeMs := time.Since(start).Milliseconds()

		response := models.LogsResponse{
			Logs:        logs,
			Total:       total,
			Limit:       req.Limit,
			Offset:      req.Offset,
			HasMore:     int64(req.Offset+req.Limit) < total,
			QueryTimeMs: &queryTimeMs,
		}

		c.JSON(http.StatusOK, response)
	}
}
