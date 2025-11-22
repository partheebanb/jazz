package middleware

import (
	"jazz/database"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

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
