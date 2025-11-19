package main

import (
	"jazz/database"
	"jazz/handlers"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	db, err := database.Connect(databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Pool.Close()

	r := gin.Default()

	r.GET("/health", handlers.HealthCheck)
	r.POST("/logs", handlers.IngestLogs(db))
	r.GET("/logs", handlers.GetLogs(db))

	log.Println("Server starting on :8080")
	r.Run(":8080")
}
