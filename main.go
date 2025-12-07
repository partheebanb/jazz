// Package main provides the Jazz logging service API server.
package main

import (
	"context"
	"jazz/database"
	"jazz/handlers"
	"jazz/middleware"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	r := gin.Default()

	r.GET("/health", handlers.HealthCheck)

	// Public project endpoints
	r.POST("/projects", handlers.CreateProject(db))
	r.GET("/projects", handlers.ListProjects(db))
	r.GET("/projects/:id", handlers.GetProject(db))
	r.DELETE("/projects/:id", handlers.DeleteProject(db))

	// Protected log endpoints (require API key)
	protected := r.Group("")
	protected.Use(middleware.AuthRequired(db))
	{
		protected.POST("/logs", handlers.IngestLogs(db))
		protected.GET("/logs", handlers.GetLogs(db))
		protected.POST("/search", handlers.SearchLogs(db))
	}

	log.Println("Server starting on :8080")
	log.Fatal(r.Run(":8080"))
}
