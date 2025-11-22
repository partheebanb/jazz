package main

import (
	"context"
	"jazz/database"
	"jazz/handlers"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	// Create context with timeout for initial connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	r := gin.Default()

	r.GET("/health", handlers.HealthCheck)
	r.POST("/logs", handlers.IngestLogs(db))
	r.GET("/logs", handlers.GetLogs(db))

	log.Println("Server starting on :8080")
	r.Run(":8080")
}
