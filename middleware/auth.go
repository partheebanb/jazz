// Package middleware provides HTTP middleware for the Jazz API.
// Currently implements API key authentication and request context enrichment.
package middleware

import (
	"jazz/database"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthRequired validates the API key and enriches the request context.
// Extracts "Authorization: Bearer <api_key>" header and validates against database.
// On success, adds project_id and project to Gin context for use by handlers.
// On failure, returns 401 Unauthorized and aborts the request chain.
//
// Usage:
//
//	protected := router.Group("")
//	protected.Use(middleware.AuthRequired(db))
//	protected.POST("/logs", handlers.IngestLogs(db))
func AuthRequired(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		apiKey := parts[1]

		// Validate API key against database
		ctx := c.Request.Context()
		project, err := db.GetProjectByAPIKey(ctx, apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			c.Abort()
			return
		}

		// Store project in context for handlers to use
		c.Set("project_id", project.ID)
		c.Set("project", project)

		c.Next()
	}
}
