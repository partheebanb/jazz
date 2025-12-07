package models

import (
	"time"

	"github.com/google/uuid"
)

// Project represents a multi-tenant project in Jazz.
// Each project has a unique API key used for authentication.
// All logs belong to exactly one project for data isolation.
type Project struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" binding:"required,min=3,max=255" db:"name"`
	APIKey    string    `json:"api_key" db:"api_key"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateProjectRequest is the payload for creating a new project.
// Name is validated to be 3-255 characters.
type CreateProjectRequest struct {
	Name string `json:"name" binding:"required,min=3,max=255"`
}

// ProjectsResponse is the standard response format for project listings.
// Includes total count for potential pagination in the future.
type ProjectsResponse struct {
	Projects []Project `json:"projects"`
	Total    int       `json:"total"`
}
