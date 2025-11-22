package handlers

import (
	"jazz/database"
	"jazz/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func IngestLogs(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var logs []models.LogEntry
		if err := c.ShouldBindJSON(&logs); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		for i := range logs {
			logs[i].ID = uuid.New()
			if logs[i].Timestamp.IsZero() {
				logs[i].Timestamp = time.Now()
			}

			if err := db.InsertLog(logs[i]); err != nil {
				c.JSON(500, gin.H{"error": "failed to insert log"})
				return
			}
		}

		c.JSON(201, gin.H{
			"message": "logs stored",
			"count":   len(logs),
		})
	}
}

func GetLogs(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params models.QueryParams
		if err := c.ShouldBindQuery(&params); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		logs, total, err := db.QueryLogs(params)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to query logs"})
			return
		}

		response := models.LogsResponse{
			Logs:    logs,
			Total:   total,
			Limit:   params.Limit,
			Offset:  params.Offset,
			HasMore: int64(params.Offset+params.Limit) < total,
		}

		c.JSON(200, response)
	}
}
