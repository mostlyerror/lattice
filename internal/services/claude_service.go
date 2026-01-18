package services

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mostlyerror/lattice/internal/models"
	"github.com/mostlyerror/lattice/pkg/claude"
)

// ClaudeService handles all Claude API interactions
type ClaudeService struct {
	client      *claude.Client
	conceptsMin int
	conceptsMax int
}

// NewClaudeService creates a new Claude service
func NewClaudeService() (*ClaudeService, error) {
	client, err := claude.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Claude client: %w", err)
	}

	// Get config from environment
	conceptsMin := 3
	conceptsMax := 7

	if minStr := os.Getenv("CONCEPTS_MIN"); minStr != "" {
		if min, err := strconv.Atoi(minStr); err == nil {
			conceptsMin = min
		}
	}

	if maxStr := os.Getenv("CONCEPTS_MAX"); maxStr != "" {
		if max, err := strconv.Atoi(maxStr); err == nil {
			conceptsMax = max
		}
	}

	return &ClaudeService{
		client:      client,
		conceptsMin: conceptsMin,
		conceptsMax: conceptsMax,
	}, nil
}

// ExtractConcepts extracts learnable concepts from a transcript
func (s *ClaudeService) ExtractConcepts(ctx context.Context, transcript string, sourceContentID int) ([]models.Concept, error) {
	// Build the prompt
	systemPrompt := "You are an expert educator extracting core learnable concepts from content."

	userPrompt := fmt.Sprintf(`Analyze this transcript and extract %d-%d concepts that someone should learn.

For each concept:
- Title: Clear, concise name (max 100 chars)
- Description: Detailed explanation (2-4 sentences, focus on practical understanding)

Focus on:
- Fundamental ideas and mental models
- Actionable techniques they can apply
- Key insights worth remembering

Return ONLY a JSON array, no markdown formatting, no code blocks:
[{"title": "...", "description": "..."}]

Transcript:
%s`, s.conceptsMin, s.conceptsMax, transcript)

	// Send request to Claude
	responseText, err := s.client.SendMessageWithSystem(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to extract concepts: %w", err)
	}

	// Parse JSON response
	var conceptData []struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := claude.ParseJSONResponse(responseText, &conceptData); err != nil {
		return nil, fmt.Errorf("failed to parse concept JSON: %w", err)
	}

	// Convert to models.Concept
	concepts := make([]models.Concept, 0, len(conceptData))
	for _, c := range conceptData {
		concepts = append(concepts, models.Concept{
			Title:           c.Title,
			Description:     c.Description,
			SourceContentID: &sourceContentID,
		})
	}

	return concepts, nil
}

// GenerateQuiz generates quiz questions for a concept
func (s *ClaudeService) GenerateQuiz(ctx context.Context, concept models.Concept) ([]models.QuizQuestion, error) {
	systemPrompt := "You are an expert educator creating effective quiz questions that test understanding and application, not just recall."

	userPrompt := fmt.Sprintf(`Generate 2-3 quiz questions for this concept to test understanding and application.

Concept:
Title: %s
Description: %s

For each question:
- Question: Tests understanding or application (avoid simple recall)
- 4 options (A, B, C, D) - make them plausible
- Correct answer (A, B, C, or D)
- Explanation: Why correct answer is right and others are wrong (2-3 sentences)

Return ONLY a JSON array, no markdown formatting, no code blocks:
[
  {
    "question": "...",
    "option_a": "...",
    "option_b": "...",
    "option_c": "...",
    "option_d": "...",
    "correct_answer": "B",
    "explanation": "..."
  }
]`, concept.Title, concept.Description)

	// Send request to Claude
	responseText, err := s.client.SendMessageWithSystem(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate quiz: %w", err)
	}

	// Parse JSON response
	var quizData []struct {
		Question      string `json:"question"`
		OptionA       string `json:"option_a"`
		OptionB       string `json:"option_b"`
		OptionC       string `json:"option_c"`
		OptionD       string `json:"option_d"`
		CorrectAnswer string `json:"correct_answer"`
		Explanation   string `json:"explanation"`
	}

	if err := claude.ParseJSONResponse(responseText, &quizData); err != nil {
		return nil, fmt.Errorf("failed to parse quiz JSON: %w", err)
	}

	// Convert to models.QuizQuestion
	questions := make([]models.QuizQuestion, 0, len(quizData))
	for _, q := range quizData {
		questions = append(questions, models.QuizQuestion{
			ConceptID:     concept.ID,
			Question:      q.Question,
			OptionA:       q.OptionA,
			OptionB:       q.OptionB,
			OptionC:       q.OptionC,
			OptionD:       q.OptionD,
			CorrectAnswer: strings.ToUpper(q.CorrectAnswer), // Normalize to uppercase
			Explanation:   q.Explanation,
		})
	}

	return questions, nil
}

// GenerateContent generates marketing content from concepts
func (s *ClaudeService) GenerateContent(ctx context.Context, platform string, concepts []models.Concept) (*models.GeneratedContent, error) {
	// Build concept summary
	var conceptsText strings.Builder
	for i, c := range concepts {
		conceptsText.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, c.Title, c.Description))
	}

	// Get platform-specific prompt
	systemPrompt, userPrompt := s.getContentPrompts(platform, conceptsText.String())

	// Send request to Claude
	responseText, err := s.client.SendMessageWithSystem(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Parse the response (expecting JSON with title and body)
	var contentData struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}

	if err := claude.ParseJSONResponse(responseText, &contentData); err != nil {
		// If JSON parsing fails, treat the whole response as body and generate a title
		contentData.Body = responseText
		contentData.Title = s.generateTitleFromConcepts(concepts)
	}

	// Extract concept IDs
	conceptIDs := make([]int, len(concepts))
	for i, c := range concepts {
		conceptIDs[i] = c.ID
	}

	return &models.GeneratedContent{
		Platform:   platform,
		Title:      contentData.Title,
		Body:       contentData.Body,
		ConceptIDs: models.IntArray(conceptIDs),
		Status:     "draft",
	}, nil
}

// getContentPrompts returns platform-specific prompts
func (s *ClaudeService) getContentPrompts(platform, conceptsText string) (systemPrompt, userPrompt string) {
	switch platform {
	case "linkedin":
		systemPrompt = "You are a consultant writing a LinkedIn post demonstrating expertise to attract clients."
		userPrompt = fmt.Sprintf(`Create a LinkedIn case study post using these concepts:

%s

Format:
- Hook: Start with a relatable client problem or situation
- Body: Show how you used these concepts to solve it (tell a story)
- Result: Share measurable outcomes or clear benefits
- Call-to-action: Invite discussion or connections

Tone: Professional, credible, approachable (not overly salesy)
Length: 1200-1500 characters

Return as JSON:
{"title": "...", "body": "..."}`, conceptsText)

	case "twitter":
		systemPrompt = "You are a consultant creating an engaging X (Twitter) thread to demonstrate expertise."
		userPrompt = fmt.Sprintf(`Create a 5-tweet thread about these concepts:

%s

Structure:
- Tweet 1: Hook - why this matters (create curiosity)
- Tweets 2-4: Key insights from the concepts (one insight per tweet)
- Tweet 5: Actionable takeaway + CTA

Tone: Casual but authoritative, conversational
Length: Each tweet under 280 characters
Use line breaks for readability

Return as JSON:
{"title": "Thread title", "body": "1/\n[tweet 1]\n\n2/\n[tweet 2]\n\n..."}`, conceptsText)

	case "blog":
		systemPrompt = "You are a consultant writing an educational blog post to demonstrate deep expertise."
		userPrompt = fmt.Sprintf(`Write a comprehensive blog post tutorial using these concepts:

%s

Structure:
- Introduction: Why this matters (set context, create interest)
- Section per concept:
  * Clear explanation
  * How to apply it (with examples)
  * Common mistakes to avoid
- Conclusion: Summary + next steps for the reader

Tone: Teaching, detailed, actionable (position yourself as the expert guide)
Length: 800-1200 words
Use Markdown formatting (headings, lists, etc.)

Return as JSON:
{"title": "...", "body": "..."}`, conceptsText)

	default:
		// Generic email format
		systemPrompt = "You are a consultant creating valuable content to share with your network."
		userPrompt = fmt.Sprintf(`Create an email newsletter about these concepts:

%s

Format:
- Subject line (compelling, specific)
- Introduction (1-2 sentences)
- Key insights (bullet points)
- Conclusion with CTA

Tone: Friendly, professional, valuable
Length: 400-600 words

Return as JSON:
{"title": "Subject line", "body": "Email body"}`, conceptsText)
	}

	return systemPrompt, userPrompt
}

// generateTitleFromConcepts creates a title from concept titles
func (s *ClaudeService) generateTitleFromConcepts(concepts []models.Concept) string {
	if len(concepts) == 0 {
		return "Generated Content"
	}

	if len(concepts) == 1 {
		return concepts[0].Title
	}

	// Use first concept as base
	return fmt.Sprintf("%s and More", concepts[0].Title)
}
