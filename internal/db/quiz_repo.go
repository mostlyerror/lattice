package db

import (
	"database/sql"
	"fmt"

	"github.com/mostlyerror/lattice/internal/models"
)

// CreateQuizBatch creates multiple quiz questions in a single transaction
func CreateQuizBatch(questions []models.QuizQuestion) ([]models.QuizQuestion, error) {
	if len(questions) == 0 {
		return []models.QuizQuestion{}, nil
	}

	// Start transaction
	tx, err := DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	query := `
		INSERT INTO quiz_questions (
			concept_id, question, option_a, option_b, option_c, option_d,
			correct_answer, explanation
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, concept_id, question, option_a, option_b, option_c, option_d,
			correct_answer, explanation, created_at
	`

	createdQuestions := make([]models.QuizQuestion, 0, len(questions))

	for _, q := range questions {
		var created models.QuizQuestion
		err := tx.QueryRow(
			query,
			q.ConceptID,
			q.Question,
			q.OptionA,
			q.OptionB,
			q.OptionC,
			q.OptionD,
			q.CorrectAnswer,
			q.Explanation,
		).Scan(
			&created.ID,
			&created.ConceptID,
			&created.Question,
			&created.OptionA,
			&created.OptionB,
			&created.OptionC,
			&created.OptionD,
			&created.CorrectAnswer,
			&created.Explanation,
			&created.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create quiz question: %w", err)
		}

		createdQuestions = append(createdQuestions, created)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return createdQuestions, nil
}

// GetQuizzesByConceptID retrieves all quizzes for a concept
func GetQuizzesByConceptID(conceptID int) ([]models.QuizQuestion, error) {
	query := `
		SELECT id, concept_id, question, option_a, option_b, option_c, option_d,
			correct_answer, explanation, created_at
		FROM quiz_questions
		WHERE concept_id = $1
		ORDER BY created_at ASC
	`

	rows, err := DB.Query(query, conceptID)
	if err != nil {
		return nil, fmt.Errorf("failed to query quiz questions: %w", err)
	}
	defer rows.Close()

	var questions []models.QuizQuestion
	for rows.Next() {
		var q models.QuizQuestion
		err := rows.Scan(
			&q.ID,
			&q.ConceptID,
			&q.Question,
			&q.OptionA,
			&q.OptionB,
			&q.OptionC,
			&q.OptionD,
			&q.CorrectAnswer,
			&q.Explanation,
			&q.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quiz question: %w", err)
		}
		questions = append(questions, q)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating quiz questions: %w", err)
	}

	return questions, nil
}

// GetQuizzesBySourceContentID retrieves all quizzes for a source content
func GetQuizzesBySourceContentID(sourceContentID int) ([]models.QuizQuestion, error) {
	query := `
		SELECT q.id, q.concept_id, q.question, q.option_a, q.option_b, q.option_c, q.option_d,
			q.correct_answer, q.explanation, q.created_at
		FROM quiz_questions q
		INNER JOIN concepts c ON q.concept_id = c.id
		WHERE c.source_content_id = $1
		ORDER BY q.created_at ASC
	`

	rows, err := DB.Query(query, sourceContentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query quiz questions: %w", err)
	}
	defer rows.Close()

	var questions []models.QuizQuestion
	for rows.Next() {
		var q models.QuizQuestion
		err := rows.Scan(
			&q.ID,
			&q.ConceptID,
			&q.Question,
			&q.OptionA,
			&q.OptionB,
			&q.OptionC,
			&q.OptionD,
			&q.CorrectAnswer,
			&q.Explanation,
			&q.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quiz question: %w", err)
		}
		questions = append(questions, q)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating quiz questions: %w", err)
	}

	return questions, nil
}

// GetQuizQuestionByID retrieves a single quiz question by ID
func GetQuizQuestionByID(id int) (*models.QuizQuestion, error) {
	query := `
		SELECT id, concept_id, question, option_a, option_b, option_c, option_d,
			correct_answer, explanation, created_at
		FROM quiz_questions
		WHERE id = $1
	`

	var q models.QuizQuestion
	err := DB.QueryRow(query, id).Scan(
		&q.ID,
		&q.ConceptID,
		&q.Question,
		&q.OptionA,
		&q.OptionB,
		&q.OptionC,
		&q.OptionD,
		&q.CorrectAnswer,
		&q.Explanation,
		&q.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("quiz question not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query quiz question: %w", err)
	}

	return &q, nil
}

// DeleteQuizQuestion deletes a quiz question by ID
func DeleteQuizQuestion(id int) error {
	query := "DELETE FROM quiz_questions WHERE id = $1"

	result, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete quiz question: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("quiz question not found")
	}

	return nil
}
