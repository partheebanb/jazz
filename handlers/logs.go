package handlers

import (
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

func IngestLogs(c *gin.Context) {
	var logs []models.LogEntry
	if err := c.ShouldBindJSON(&logs); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"message": "logs received",
		"count":   len(logs),
	})
}

func GetLogs(c *gin.Context) {
	mockLogs := []models.LogEntry{
		{
			ID:        uuid.New(),
			Level:     "error",
			Message:   "Database connection failed",
			Timestamp: time.Now(),
		},
		{
			ID:        uuid.New(),
			Level:     "info",
			Message:   "Server started successfully",
			Timestamp: time.Now(),
		},
	}

	c.JSON(200, gin.H{
		"logs": mockLogs,
	})
}