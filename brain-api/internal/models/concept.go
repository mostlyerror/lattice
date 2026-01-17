package models

import "time"

// Concept represents a single learnable unit extracted from content
type Concept struct {
	ID              int       `json:"id" db:"id"`
	Title           string    `json:"title" db:"title"`
	Description     string    `json:"description" db:"description"`
	SourceContentID *int      `json:"source_content_id,omitempty" db:"source_content_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// CreateConceptRequest represents the request body for creating a concept
type CreateConceptRequest struct {
	Title           string `json:"title" binding:"required"`
	Description     string `json:"description" binding:"required"`
	SourceContentID *int   `json:"source_content_id,omitempty"`
}

// UpdateConceptRequest represents the request body for updating a concept
type UpdateConceptRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}
