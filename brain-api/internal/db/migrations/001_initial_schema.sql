-- Initial schema for Brain application
-- Creates tables for concepts, source content, quizzes, learning progress, and generated content

-- Source Contents table (YouTube videos, PDFs, articles)
CREATE TABLE IF NOT EXISTS source_contents (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL CHECK (type IN ('youtube', 'pdf', 'article')),
    url TEXT NOT NULL,
    title TEXT NOT NULL,
    transcript TEXT,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Concepts table (learnable units extracted from content)
CREATE TABLE IF NOT EXISTS concepts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    source_content_id INTEGER REFERENCES source_contents(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Quiz Questions table
CREATE TABLE IF NOT EXISTS quiz_questions (
    id SERIAL PRIMARY KEY,
    concept_id INTEGER NOT NULL REFERENCES concepts(id) ON DELETE CASCADE,
    question TEXT NOT NULL,
    option_a TEXT NOT NULL,
    option_b TEXT NOT NULL,
    option_c TEXT NOT NULL,
    option_d TEXT NOT NULL,
    correct_answer CHAR(1) NOT NULL CHECK (correct_answer IN ('A', 'B', 'C', 'D')),
    explanation TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Quiz Attempts table (tracking user answers)
CREATE TABLE IF NOT EXISTS quiz_attempts (
    id SERIAL PRIMARY KEY,
    question_id INTEGER NOT NULL REFERENCES quiz_questions(id) ON DELETE CASCADE,
    selected_answer CHAR(1) NOT NULL CHECK (selected_answer IN ('A', 'B', 'C', 'D')),
    correct BOOLEAN NOT NULL,
    attempted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Learning Progress table (spaced repetition tracking)
CREATE TABLE IF NOT EXISTS learning_progress (
    id SERIAL PRIMARY KEY,
    concept_id INTEGER UNIQUE NOT NULL REFERENCES concepts(id) ON DELETE CASCADE,
    mastery_level INTEGER DEFAULT 0 CHECK (mastery_level >= 0 AND mastery_level <= 5),
    consecutive_correct INTEGER DEFAULT 0,
    last_reviewed_at TIMESTAMP,
    next_review_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Generated Content table (marketing content created from concepts)
CREATE TABLE IF NOT EXISTS generated_contents (
    id SERIAL PRIMARY KEY,
    platform VARCHAR(50) NOT NULL CHECK (platform IN ('linkedin', 'twitter', 'blog', 'email')),
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    concept_ids JSONB NOT NULL, -- Array of concept IDs
    status VARCHAR(50) DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
    published_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Concept Relationships table (many-to-many relationships between concepts)
CREATE TABLE IF NOT EXISTS concept_relationships (
    id SERIAL PRIMARY KEY,
    from_concept_id INTEGER NOT NULL REFERENCES concepts(id) ON DELETE CASCADE,
    to_concept_id INTEGER NOT NULL REFERENCES concepts(id) ON DELETE CASCADE,
    relationship_type VARCHAR(50) DEFAULT 'related',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(from_concept_id, to_concept_id)
);

-- Publishing Events table (tracking when content was published to platforms)
CREATE TABLE IF NOT EXISTS publishing_events (
    id SERIAL PRIMARY KEY,
    content_id INTEGER NOT NULL REFERENCES generated_contents(id) ON DELETE CASCADE,
    platform VARCHAR(50) NOT NULL,
    platform_post_id VARCHAR(255), -- ID from LinkedIn/Twitter
    published_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    engagement_metrics JSONB -- likes, shares, comments, etc.
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_concepts_source_content ON concepts(source_content_id);
CREATE INDEX IF NOT EXISTS idx_quiz_questions_concept ON quiz_questions(concept_id);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_question ON quiz_attempts(question_id);
CREATE INDEX IF NOT EXISTS idx_learning_progress_concept ON learning_progress(concept_id);
CREATE INDEX IF NOT EXISTS idx_learning_progress_next_review ON learning_progress(next_review_at);
CREATE INDEX IF NOT EXISTS idx_generated_contents_status ON generated_contents(status);
CREATE INDEX IF NOT EXISTS idx_concept_relationships_from ON concept_relationships(from_concept_id);
CREATE INDEX IF NOT EXISTS idx_concept_relationships_to ON concept_relationships(to_concept_id);

-- Create trigger to update updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_concepts_updated_at BEFORE UPDATE ON concepts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_learning_progress_updated_at BEFORE UPDATE ON learning_progress
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_generated_contents_updated_at BEFORE UPDATE ON generated_contents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
