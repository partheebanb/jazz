package handlers

import (
	"jazz/database"
	"jazz/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
