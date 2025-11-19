package models

import (
	"time"

	"github.com/google/uuid"
)

type LogEntry struct {
	ID        uuid.UUID `json:"id"`
	Level     string    `json:"level" binding:"required"`
	Message   string    `json:"message" binding:"required"`
	Timestamp time.Time `json:"timestamp"`
}