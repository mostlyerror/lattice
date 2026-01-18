package db

import (
	"database/sql"
	"fmt"

	"github.com/mostlyerror/lattice/internal/models"
)

// GetAllConcepts retrieves all concepts from the database
func GetAllConcepts() ([]models.Concept, error) {
	query := `
		SELECT id, title, description, source_content_id, created_at, updated_at
		FROM concepts
		ORDER BY created_at DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query concepts: %w", err)
	}
	defer rows.Close()

	var concepts []models.Concept
	for rows.Next() {
		var c models.Concept
		err := rows.Scan(
			&c.ID,
			&c.Title,
			&c.Description,
			&c.SourceContentID,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan concept: %w", err)
		}
		concepts = append(concepts, c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating concepts: %w", err)
	}

	return concepts, nil
}

// GetConceptByID retrieves a single concept by ID
func GetConceptByID(id int) (*models.Concept, error) {
	query := `
		SELECT id, title, description, source_content_id, created_at, updated_at
		FROM concepts
		WHERE id = $1
	`

	var c models.Concept
	err := DB.QueryRow(query, id).Scan(
		&c.ID,
		&c.Title,
		&c.Description,
		&c.SourceContentID,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("concept not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query concept: %w", err)
	}

	return &c, nil
}

// CreateConcept creates a new concept in the database
func CreateConcept(req models.CreateConceptRequest) (*models.Concept, error) {
	query := `
		INSERT INTO concepts (title, description, source_content_id)
		VALUES ($1, $2, $3)
		RETURNING id, title, description, source_content_id, created_at, updated_at
	`

	var c models.Concept
	err := DB.QueryRow(
		query,
		req.Title,
		req.Description,
		req.SourceContentID,
	).Scan(
		&c.ID,
		&c.Title,
		&c.Description,
		&c.SourceContentID,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create concept: %w", err)
	}

	return &c, nil
}

// UpdateConcept updates an existing concept
func UpdateConcept(id int, req models.UpdateConceptRequest) (*models.Concept, error) {
	// Build dynamic update query
	query := "UPDATE concepts SET "
	args := []interface{}{}
	argCount := 1

	if req.Title != nil {
		query += fmt.Sprintf("title = $%d, ", argCount)
		args = append(args, *req.Title)
		argCount++
	}

	if req.Description != nil {
		query += fmt.Sprintf("description = $%d, ", argCount)
		args = append(args, *req.Description)
		argCount++
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]

	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, title, description, source_content_id, created_at, updated_at", argCount)
	args = append(args, id)

	var c models.Concept
	err := DB.QueryRow(query, args...).Scan(
		&c.ID,
		&c.Title,
		&c.Description,
		&c.SourceContentID,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("concept not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update concept: %w", err)
	}

	return &c, nil
}

// DeleteConcept deletes a concept by ID
func DeleteConcept(id int) error {
	query := "DELETE FROM concepts WHERE id = $1"

	result, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete concept: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("concept not found")
	}

	return nil
}

// GetConceptsBySourceContentID retrieves all concepts for a source content
func GetConceptsBySourceContentID(sourceContentID int) ([]models.Concept, error) {
	query := `
		SELECT id, title, description, source_content_id, created_at, updated_at
		FROM concepts
		WHERE source_content_id = $1
		ORDER BY created_at DESC
	`

	rows, err := DB.Query(query, sourceContentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query concepts: %w", err)
	}
	defer rows.Close()

	var concepts []models.Concept
	for rows.Next() {
		var c models.Concept
		err := rows.Scan(
			&c.ID,
			&c.Title,
			&c.Description,
			&c.SourceContentID,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan concept: %w", err)
		}
		concepts = append(concepts, c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating concepts: %w", err)
	}

	return concepts, nil
}

// CreateConceptsBatch creates multiple concepts in a single transaction
func CreateConceptsBatch(concepts []models.Concept) ([]models.Concept, error) {
	if len(concepts) == 0 {
		return []models.Concept{}, nil
	}

	// Start transaction
	tx, err := DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	query := `
		INSERT INTO concepts (title, description, source_content_id)
		VALUES ($1, $2, $3)
		RETURNING id, title, description, source_content_id, created_at, updated_at
	`

	createdConcepts := make([]models.Concept, 0, len(concepts))

	for _, concept := range concepts {
		var c models.Concept
		err := tx.QueryRow(
			query,
			concept.Title,
			concept.Description,
			concept.SourceContentID,
		).Scan(
			&c.ID,
			&c.Title,
			&c.Description,
			&c.SourceContentID,
			&c.CreatedAt,
			&c.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create concept: %w", err)
		}

		createdConcepts = append(createdConcepts, c)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return createdConcepts, nil
}
