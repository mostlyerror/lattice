package db

import (
	"database/sql"
	"fmt"

	"github.com/benjaminpoon/brain-api/internal/models"
)

// CreateSourceContent creates a new source content record
func CreateSourceContent(req models.CreateSourceContentRequest) (*models.SourceContent, error) {
	query := `
		INSERT INTO source_contents (type, url, title, transcript, processed_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, type, url, title, transcript, processed_at, created_at
	`

	var sc models.SourceContent
	err := DB.QueryRow(
		query,
		req.Type,
		req.URL,
		req.Title,
		req.Transcript,
	).Scan(
		&sc.ID,
		&sc.Type,
		&sc.URL,
		&sc.Title,
		&sc.Transcript,
		&sc.ProcessedAt,
		&sc.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create source content: %w", err)
	}

	return &sc, nil
}

// GetSourceContentByURL retrieves source content by URL (for duplicate detection)
func GetSourceContentByURL(url string) (*models.SourceContent, error) {
	query := `
		SELECT id, type, url, title, transcript, processed_at, created_at
		FROM source_contents
		WHERE url = $1
	`

	var sc models.SourceContent
	err := DB.QueryRow(query, url).Scan(
		&sc.ID,
		&sc.Type,
		&sc.URL,
		&sc.Title,
		&sc.Transcript,
		&sc.ProcessedAt,
		&sc.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not an error, just not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query source content: %w", err)
	}

	return &sc, nil
}

// GetAllSourceContents retrieves all source contents
func GetAllSourceContents() ([]models.SourceContent, error) {
	query := `
		SELECT id, type, url, title, transcript, processed_at, created_at
		FROM source_contents
		ORDER BY created_at DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query source contents: %w", err)
	}
	defer rows.Close()

	var contents []models.SourceContent
	for rows.Next() {
		var sc models.SourceContent
		err := rows.Scan(
			&sc.ID,
			&sc.Type,
			&sc.URL,
			&sc.Title,
			&sc.Transcript,
			&sc.ProcessedAt,
			&sc.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan source content: %w", err)
		}
		contents = append(contents, sc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating source contents: %w", err)
	}

	return contents, nil
}

// GetSourceContentByID retrieves a single source content by ID
func GetSourceContentByID(id int) (*models.SourceContent, error) {
	query := `
		SELECT id, type, url, title, transcript, processed_at, created_at
		FROM source_contents
		WHERE id = $1
	`

	var sc models.SourceContent
	err := DB.QueryRow(query, id).Scan(
		&sc.ID,
		&sc.Type,
		&sc.URL,
		&sc.Title,
		&sc.Transcript,
		&sc.ProcessedAt,
		&sc.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("source content not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query source content: %w", err)
	}

	return &sc, nil
}

// DeleteSourceContent deletes a source content by ID
func DeleteSourceContent(id int) error {
	query := "DELETE FROM source_contents WHERE id = $1"

	result, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete source content: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("source content not found")
	}

	return nil
}
