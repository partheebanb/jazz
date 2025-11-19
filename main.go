package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LogEntry struct {
	ID        uuid.UUID `json:"id"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	r.POST("/logs", func(c *gin.Context) {
		var logs []LogEntry
		if err := c.ShouldBindJSON(&logs); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(201, gin.H{
			"message": "logs received",
			"count":   len(logs),
		})
	})

	r.GET("/logs", func(c *gin.Context) {
		mockLogs := []LogEntry{
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
	})

	r.Run(":8080")
}