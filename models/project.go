package models

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" binding:"required,min=3,max=255" db:"name"`
	APIKey    string    `json:"api_key" db:"api_key"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateProjectRequest struct {
	Name string `json:"name" binding:"required,min=3,max=255"`
}

type ProjectsResponse struct {
	Projects []Project `json:"projects"`
	Total    int       `json:"total"`
}
