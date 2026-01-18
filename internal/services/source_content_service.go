package services

import (
	"context"
	"fmt"
	"log"

	"github.com/mostlyerror/lattice/internal/db"
	"github.com/mostlyerror/lattice/internal/models"
	"github.com/mostlyerror/lattice/pkg/youtube"
)

// SourceContentService orchestrates the full content processing pipeline
type SourceContentService struct {
	youtubeClient *youtube.Client
	claudeService *ClaudeService
}

// ProcessResult contains the results of processing source content
type ProcessResult struct {
	SourceContent    *models.SourceContent      `json:"source_content"`
	Concepts         []models.Concept           `json:"concepts"`
	Quizzes          []models.QuizQuestion      `json:"quizzes"`
	GeneratedContent []models.GeneratedContent  `json:"generated_content"`
}

// NewSourceContentService creates a new source content service
func NewSourceContentService() (*SourceContentService, error) {
	ytClient, err := youtube.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube client: %w", err)
	}

	claudeService, err := NewClaudeService()
	if err != nil {
		return nil, fmt.Errorf("failed to create Claude service: %w", err)
	}

	return &SourceContentService{
		youtubeClient: ytClient,
		claudeService: claudeService,
	}, nil
}

// ProcessYouTubeURL runs the full workflow for a YouTube video
func (s *SourceContentService) ProcessYouTubeURL(ctx context.Context, url string) (*ProcessResult, error) {
	log.Printf("Processing YouTube URL: %s", url)

	// Step 1: Check for duplicates
	existing, err := db.GetSourceContentByURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}

	if existing != nil {
		log.Printf("URL already processed, returning existing data for source content ID: %d", existing.ID)
		return s.getExistingProcessResult(ctx, existing)
	}

	// Step 2: Fetch YouTube transcript and metadata
	log.Printf("Fetching YouTube video info...")
	videoInfo, err := s.youtubeClient.GetVideoInfo(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch YouTube video: %w", err)
	}

	if videoInfo.Transcript == nil {
		return nil, fmt.Errorf("no transcript available for this video")
	}

	// Step 3: Save source content
	log.Printf("Saving source content...")
	sourceContent, err := db.CreateSourceContent(models.CreateSourceContentRequest{
		Type:       "youtube",
		URL:        url,
		Title:      videoInfo.Metadata.Title,
		Transcript: videoInfo.Transcript.Text,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save source content: %w", err)
	}

	log.Printf("Source content saved with ID: %d", sourceContent.ID)

	// Step 4: Extract concepts via Claude
	log.Printf("Extracting concepts from transcript...")
	concepts, err := s.claudeService.ExtractConcepts(ctx, videoInfo.Transcript.Text, sourceContent.ID)
	if err != nil {
		// Log error but don't fail - we have source content saved
		log.Printf("Warning: Failed to extract concepts: %v", err)
		return &ProcessResult{
			SourceContent:    sourceContent,
			Concepts:         []models.Concept{},
			Quizzes:          []models.QuizQuestion{},
			GeneratedContent: []models.GeneratedContent{},
		}, nil
	}

	// Save concepts to database
	log.Printf("Saving %d concepts to database...", len(concepts))
	savedConcepts, err := db.CreateConceptsBatch(concepts)
	if err != nil {
		log.Printf("Warning: Failed to save concepts: %v", err)
		return &ProcessResult{
			SourceContent:    sourceContent,
			Concepts:         []models.Concept{},
			Quizzes:          []models.QuizQuestion{},
			GeneratedContent: []models.GeneratedContent{},
		}, nil
	}

	log.Printf("Concepts saved successfully")

	// Step 5: Generate quizzes for each concept
	log.Printf("Generating quizzes for concepts...")
	var allQuizzes []models.QuizQuestion

	for _, concept := range savedConcepts {
		quizzes, err := s.claudeService.GenerateQuiz(ctx, concept)
		if err != nil {
			log.Printf("Warning: Failed to generate quiz for concept %d: %v", concept.ID, err)
			continue
		}
		allQuizzes = append(allQuizzes, quizzes...)
	}

	// Save quizzes to database
	if len(allQuizzes) > 0 {
		log.Printf("Saving %d quizzes to database...", len(allQuizzes))
		savedQuizzes, err := db.CreateQuizBatch(allQuizzes)
		if err != nil {
			log.Printf("Warning: Failed to save quizzes: %v", err)
			allQuizzes = []models.QuizQuestion{}
		} else {
			allQuizzes = savedQuizzes
			log.Printf("Quizzes saved successfully")
		}
	}

	// Step 6: Generate content for all platforms
	log.Printf("Generating marketing content...")
	platforms := []string{"linkedin", "twitter", "blog"}
	var generatedContents []models.GeneratedContent

	for _, platform := range platforms {
		content, err := s.claudeService.GenerateContent(ctx, platform, savedConcepts)
		if err != nil {
			log.Printf("Warning: Failed to generate %s content: %v", platform, err)
			continue
		}
		generatedContents = append(generatedContents, *content)
	}

	// Save generated content to database
	if len(generatedContents) > 0 {
		log.Printf("Saving %d generated content pieces to database...", len(generatedContents))
		savedContent, err := db.CreateGeneratedContentBatch(generatedContents)
		if err != nil {
			log.Printf("Warning: Failed to save generated content: %v", err)
			generatedContents = []models.GeneratedContent{}
		} else {
			generatedContents = savedContent
			log.Printf("Generated content saved successfully")
		}
	}

	// Step 7: Return complete result
	log.Printf("Processing complete for source content ID: %d", sourceContent.ID)

	return &ProcessResult{
		SourceContent:    sourceContent,
		Concepts:         savedConcepts,
		Quizzes:          allQuizzes,
		GeneratedContent: generatedContents,
	}, nil
}

// getExistingProcessResult retrieves all related data for an existing source content
func (s *SourceContentService) getExistingProcessResult(ctx context.Context, sourceContent *models.SourceContent) (*ProcessResult, error) {
	// Get concepts
	concepts, err := db.GetConceptsBySourceContentID(sourceContent.ID)
	if err != nil {
		log.Printf("Warning: Failed to get concepts: %v", err)
		concepts = []models.Concept{}
	}

	// Get quizzes
	quizzes, err := db.GetQuizzesBySourceContentID(sourceContent.ID)
	if err != nil {
		log.Printf("Warning: Failed to get quizzes: %v", err)
		quizzes = []models.QuizQuestion{}
	}

	// Get generated content (by concept IDs)
	var generatedContent []models.GeneratedContent
	if len(concepts) > 0 {
		conceptIDs := make([]int, len(concepts))
		for i, c := range concepts {
			conceptIDs[i] = c.ID
		}

		content, err := db.GetGeneratedContentByConceptIDs(conceptIDs)
		if err != nil {
			log.Printf("Warning: Failed to get generated content: %v", err)
			generatedContent = []models.GeneratedContent{}
		} else {
			generatedContent = content
		}
	}

	return &ProcessResult{
		SourceContent:    sourceContent,
		Concepts:         concepts,
		Quizzes:          quizzes,
		GeneratedContent: generatedContent,
	}, nil
}

// GetSourceContentWithRelated retrieves source content and all related data
func (s *SourceContentService) GetSourceContentWithRelated(ctx context.Context, id int) (*ProcessResult, error) {
	sourceContent, err := db.GetSourceContentByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get source content: %w", err)
	}

	return s.getExistingProcessResult(ctx, sourceContent)
}
