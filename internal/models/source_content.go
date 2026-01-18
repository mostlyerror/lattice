package models

import "time"

// SourceContent represents the original video/article/PDF
type SourceContent struct {
	ID          int       `json:"id" db:"id"`
	Type        string    `json:"type" db:"type"` // youtube, pdf, article
	URL         string    `json:"url" db:"url"`
	Title       string    `json:"title" db:"title"`
	Transcript  string    `json:"transcript" db:"transcript"`
	ProcessedAt time.Time `json:"processed_at" db:"processed_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// CreateSourceContentRequest represents the request body for ingesting content
type CreateSourceContentRequest struct {
	Type       string `json:"type" binding:"required,oneof=youtube pdf article"`
	URL        string `json:"url" binding:"required"`
	Title      string `json:"title"`
	Transcript string `json:"transcript"`
}
