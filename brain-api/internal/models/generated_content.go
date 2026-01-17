package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// IntArray is a custom type for handling PostgreSQL integer arrays
type IntArray []int

// Scan implements the sql.Scanner interface
func (a *IntArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan IntArray")
	}

	return json.Unmarshal(bytes, a)
}

// Value implements the driver.Valuer interface
func (a IntArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// GeneratedContent represents marketing content created from concepts
type GeneratedContent struct {
	ID          int        `json:"id" db:"id"`
	Platform    string     `json:"platform" db:"platform"` // linkedin, twitter, blog, email
	Title       string     `json:"title" db:"title"`
	Body        string     `json:"body" db:"body"`
	ConceptIDs  IntArray   `json:"concept_ids" db:"concept_ids"` // JSON array of concept IDs
	Status      string     `json:"status" db:"status"`           // draft, published
	PublishedAt *time.Time `json:"published_at,omitempty" db:"published_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// GenerateContentRequest represents the request body for generating content
type GenerateContentRequest struct {
	Platform   string `json:"platform" binding:"required,oneof=linkedin twitter blog email"`
	ConceptIDs []int  `json:"concept_ids" binding:"required,min=1"`
	Tone       string `json:"tone,omitempty"` // professional, casual, technical
}

// UpdateGeneratedContentRequest represents the request body for updating generated content
type UpdateGeneratedContentRequest struct {
	Title  *string `json:"title,omitempty"`
	Body   *string `json:"body,omitempty"`
	Status *string `json:"status,omitempty"`
}
