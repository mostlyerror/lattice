package db

import (
	"database/sql"
	"fmt"

	"github.com/mostlyerror/lattice/internal/models"
)

// CreateGeneratedContent creates a new generated content record
func CreateGeneratedContent(content *models.GeneratedContent) (*models.GeneratedContent, error) {
	query := `
		INSERT INTO generated_contents (platform, title, body, concept_ids, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, platform, title, body, concept_ids, status, published_at, created_at, updated_at
	`

	var gc models.GeneratedContent
	err := DB.QueryRow(
		query,
		content.Platform,
		content.Title,
		content.Body,
		content.ConceptIDs,
		content.Status,
	).Scan(
		&gc.ID,
		&gc.Platform,
		&gc.Title,
		&gc.Body,
		&gc.ConceptIDs,
		&gc.Status,
		&gc.PublishedAt,
		&gc.CreatedAt,
		&gc.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create generated content: %w", err)
	}

	return &gc, nil
}

// CreateGeneratedContentBatch creates multiple generated content records in a transaction
func CreateGeneratedContentBatch(contents []models.GeneratedContent) ([]models.GeneratedContent, error) {
	if len(contents) == 0 {
		return []models.GeneratedContent{}, nil
	}

	// Start transaction
	tx, err := DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO generated_contents (platform, title, body, concept_ids, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, platform, title, body, concept_ids, status, published_at, created_at, updated_at
	`

	createdContents := make([]models.GeneratedContent, 0, len(contents))

	for _, content := range contents {
		var gc models.GeneratedContent
		err := tx.QueryRow(
			query,
			content.Platform,
			content.Title,
			content.Body,
			content.ConceptIDs,
			content.Status,
		).Scan(
			&gc.ID,
			&gc.Platform,
			&gc.Title,
			&gc.Body,
			&gc.ConceptIDs,
			&gc.Status,
			&gc.PublishedAt,
			&gc.CreatedAt,
			&gc.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create generated content: %w", err)
		}

		createdContents = append(createdContents, gc)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return createdContents, nil
}

// GetGeneratedContentByID retrieves a single generated content by ID
func GetGeneratedContentByID(id int) (*models.GeneratedContent, error) {
	query := `
		SELECT id, platform, title, body, concept_ids, status, published_at, created_at, updated_at
		FROM generated_contents
		WHERE id = $1
	`

	var gc models.GeneratedContent
	err := DB.QueryRow(query, id).Scan(
		&gc.ID,
		&gc.Platform,
		&gc.Title,
		&gc.Body,
		&gc.ConceptIDs,
		&gc.Status,
		&gc.PublishedAt,
		&gc.CreatedAt,
		&gc.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("generated content not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query generated content: %w", err)
	}

	return &gc, nil
}

// GetGeneratedContentByConceptIDs retrieves generated content that contains specific concept IDs
func GetGeneratedContentByConceptIDs(conceptIDs []int) ([]models.GeneratedContent, error) {
	// This is a simplified version - in production, you'd want to use PostgreSQL array operators
	// For now, we'll get all and filter in memory
	query := `
		SELECT id, platform, title, body, concept_ids, status, published_at, created_at, updated_at
		FROM generated_contents
		ORDER BY created_at DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query generated contents: %w", err)
	}
	defer rows.Close()

	var contents []models.GeneratedContent
	for rows.Next() {
		var gc models.GeneratedContent
		err := rows.Scan(
			&gc.ID,
			&gc.Platform,
			&gc.Title,
			&gc.Body,
			&gc.ConceptIDs,
			&gc.Status,
			&gc.PublishedAt,
			&gc.CreatedAt,
			&gc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan generated content: %w", err)
		}

		// Filter by concept IDs (check if any match)
		for _, targetID := range conceptIDs {
			for _, contentID := range gc.ConceptIDs {
				if contentID == targetID {
					contents = append(contents, gc)
					break
				}
			}
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating generated contents: %w", err)
	}

	return contents, nil
}

// GetAllGeneratedContents retrieves all generated contents
func GetAllGeneratedContents() ([]models.GeneratedContent, error) {
	query := `
		SELECT id, platform, title, body, concept_ids, status, published_at, created_at, updated_at
		FROM generated_contents
		ORDER BY created_at DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query generated contents: %w", err)
	}
	defer rows.Close()

	var contents []models.GeneratedContent
	for rows.Next() {
		var gc models.GeneratedContent
		err := rows.Scan(
			&gc.ID,
			&gc.Platform,
			&gc.Title,
			&gc.Body,
			&gc.ConceptIDs,
			&gc.Status,
			&gc.PublishedAt,
			&gc.CreatedAt,
			&gc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan generated content: %w", err)
		}
		contents = append(contents, gc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating generated contents: %w", err)
	}

	return contents, nil
}

// UpdateGeneratedContent updates an existing generated content
func UpdateGeneratedContent(id int, req models.UpdateGeneratedContentRequest) (*models.GeneratedContent, error) {
	// Build dynamic update query
	query := "UPDATE generated_contents SET "
	args := []interface{}{}
	argCount := 1

	if req.Title != nil {
		query += fmt.Sprintf("title = $%d, ", argCount)
		args = append(args, *req.Title)
		argCount++
	}

	if req.Body != nil {
		query += fmt.Sprintf("body = $%d, ", argCount)
		args = append(args, *req.Body)
		argCount++
	}

	if req.Status != nil {
		query += fmt.Sprintf("status = $%d, ", argCount)
		args = append(args, *req.Status)
		argCount++
	}

	// Always update updated_at
	query += fmt.Sprintf("updated_at = NOW() WHERE id = $%d ", argCount)
	args = append(args, id)
	argCount++

	query += "RETURNING id, platform, title, body, concept_ids, status, published_at, created_at, updated_at"

	var gc models.GeneratedContent
	err := DB.QueryRow(query, args...).Scan(
		&gc.ID,
		&gc.Platform,
		&gc.Title,
		&gc.Body,
		&gc.ConceptIDs,
		&gc.Status,
		&gc.PublishedAt,
		&gc.CreatedAt,
		&gc.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("generated content not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update generated content: %w", err)
	}

	return &gc, nil
}

// DeleteGeneratedContent deletes a generated content by ID
func DeleteGeneratedContent(id int) error {
	query := "DELETE FROM generated_contents WHERE id = $1"

	result, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete generated content: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("generated content not found")
	}

	return nil
}
