package main

import (
	"jazz/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/health", handlers.HealthCheck)
	r.POST("/logs", handlers.IngestLogs)
	r.GET("/logs", handlers.GetLogs)

	r.Run(":8080")
}