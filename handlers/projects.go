package handlers

import (
	"jazz/database"
	"jazz/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateProject creates a new project and returns it with a generated API key.
// API key is shown only once - caller must store it securely.
// Project name must be 3-255 characters (validated by binding).
//
// Request body:
//
//	{"name": "My Application"}
//
// Response:
//
//	{
//	  "id": "...",
//	  "name": "My Application",
//	  "api_key": "jazz_...",
//	  "created_at": "...",
//	  "updated_at": "..."
//	}
//
// Returns 201 Created on success, 400 for validation errors, 500 for database errors.
func CreateProject(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateProjectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Bind error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("Creating project: %s", req.Name)

		ctx := c.Request.Context()
		project, err := db.CreateProject(ctx, req.Name)
		if err != nil {
			log.Printf("CreateProject database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to create project",
				"details": err.Error(), // Include details in dev
			})
			return
		}

		log.Printf("Project created: %s", project.ID)
		c.JSON(http.StatusCreated, project)
	}
}

// ListProjects returns all projects ordered by creation date (newest first).
// No authentication required (will be added in Phase 4 with user accounts).
// Response includes projects array and total count.
func ListProjects(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		projects, err := db.ListProjects(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list projects"})
			return
		}

		c.JSON(http.StatusOK, models.ProjectsResponse{
			Projects: projects,
			Total:    len(projects),
		})
	}
}

// GetProject retrieves a single project by ID.
// No authentication required (will be added in Phase 4 with user accounts).
// Returns 404 if project doesn't exist, 500 for database errors.
func GetProject(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectIDStr := c.Param("id")
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
			return
		}

		ctx := c.Request.Context()
		project, err := db.GetProject(ctx, projectID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}

		c.JSON(http.StatusOK, project)
	}
}

// DeleteProject removes a project and all its logs (CASCADE).
func DeleteProject(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectIDStr := c.Param("id")
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
			return
		}

		ctx := c.Request.Context()
		if err := db.DeleteProject(ctx, projectID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete project"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "project deleted"})
	}
}
