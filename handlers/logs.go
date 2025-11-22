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

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func IngestLogs(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		log.Printf("ingested %d logs", len(logs))
		c.JSON(http.StatusCreated, gin.H{
			"message": "logs stored",
			"count":   len(logs),
		})
	}
}

func GetLogs(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params models.QueryParams
		if err := c.ShouldBindQuery(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set defaults and limits
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
		logs, total, err := db.QueryLogs(ctx, params)
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
