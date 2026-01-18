package models

import "time"

// QuizQuestion represents a generated quiz question
type QuizQuestion struct {
	ID            int       `json:"id" db:"id"`
	ConceptID     int       `json:"concept_id" db:"concept_id"`
	Question      string    `json:"question" db:"question"`
	OptionA       string    `json:"option_a" db:"option_a"`
	OptionB       string    `json:"option_b" db:"option_b"`
	OptionC       string    `json:"option_c" db:"option_c"`
	OptionD       string    `json:"option_d" db:"option_d"`
	CorrectAnswer string    `json:"correct_answer" db:"correct_answer"` // A, B, C, or D
	Explanation   string    `json:"explanation" db:"explanation"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// QuizAttempt represents a user's quiz answer tracking
type QuizAttempt struct {
	ID             int       `json:"id" db:"id"`
	QuestionID     int       `json:"question_id" db:"question_id"`
	SelectedAnswer string    `json:"selected_answer" db:"selected_answer"`
	Correct        bool      `json:"correct" db:"correct"`
	AttemptedAt    time.Time `json:"attempted_at" db:"attempted_at"`
}

// LearningProgress represents spaced repetition tracking
type LearningProgress struct {
	ID                 int       `json:"id" db:"id"`
	ConceptID          int       `json:"concept_id" db:"concept_id"`
	MasteryLevel       int       `json:"mastery_level" db:"mastery_level"` // 0-5
	ConsecutiveCorrect int       `json:"consecutive_correct" db:"consecutive_correct"`
	LastReviewedAt     time.Time `json:"last_reviewed_at" db:"last_reviewed_at"`
	NextReviewAt       time.Time `json:"next_review_at" db:"next_review_at"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// AnswerQuizRequest represents the request body for answering a quiz question
type AnswerQuizRequest struct {
	QuestionID     int    `json:"question_id" binding:"required"`
	SelectedAnswer string `json:"selected_answer" binding:"required,oneof=A B C D"`
}

// AnswerQuizResponse represents the response after answering a quiz question
type AnswerQuizResponse struct {
	Correct        bool      `json:"correct"`
	CorrectAnswer  string    `json:"correct_answer"`
	Explanation    string    `json:"explanation"`
	NextReviewAt   time.Time `json:"next_review_at"`
	MasteryLevel   int       `json:"mastery_level"`
}
