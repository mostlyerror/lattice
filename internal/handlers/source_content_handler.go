package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/mostlyerror/lattice/internal/db"
	"github.com/mostlyerror/lattice/internal/models"
	"github.com/mostlyerror/lattice/internal/services"
	"github.com/gin-gonic/gin"
)

var sourceContentService *services.SourceContentService

// InitSourceContentService initializes the source content service
func InitSourceContentService() error {
	var err error
	sourceContentService, err = services.NewSourceContentService()
	if err != nil {
		return err
	}
	return nil
}

// ProcessSourceContent handles POST /api/source-content
// Processes a new YouTube URL through the full pipeline
func ProcessSourceContent(c *gin.Context) {
	var req models.CreateSourceContentRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Currently only support YouTube
	if req.Type != "youtube" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid content type",
			"details": "Only 'youtube' type is currently supported",
		})
		return
	}

	// Process the YouTube URL
	log.Printf("Processing source content request: type=%s, url=%s", req.Type, req.URL)

	result, err := sourceContentService.ProcessYouTubeURL(c.Request.Context(), req.URL)
	if err != nil {
		log.Printf("Error processing source content: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process source content",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Successfully processed source content ID: %d", result.SourceContent.ID)

	// Return the full result
	c.JSON(http.StatusCreated, result)
}

// GetSourceContents handles GET /api/source-content
// Returns all source contents
func GetSourceContents(c *gin.Context) {
	contents, err := db.GetAllSourceContents()
	if err != nil {
		log.Printf("Error getting source contents: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve source contents",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"source_contents": contents,
		"count":           len(contents),
	})
}

// GetSourceContent handles GET /api/source-content/:id
// Returns a specific source content with all related data
func GetSourceContent(c *gin.Context) {
	// Parse ID from URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"details": "ID must be a number",
		})
		return
	}

	// Get source content with related data
	result, err := sourceContentService.GetSourceContentWithRelated(c.Request.Context(), id)
	if err != nil {
		log.Printf("Error getting source content %d: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Source content not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetSourceContentConcepts handles GET /api/source-content/:id/concepts
// Returns all concepts for a source content
func GetSourceContentConcepts(c *gin.Context) {
	// Parse ID from URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"details": "ID must be a number",
		})
		return
	}

	// Get concepts
	concepts, err := db.GetConceptsBySourceContentID(id)
	if err != nil {
		log.Printf("Error getting concepts for source content %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve concepts",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"concepts": concepts,
		"count":    len(concepts),
	})
}

// GetSourceContentQuizzes handles GET /api/source-content/:id/quizzes
// Returns all quizzes for a source content
func GetSourceContentQuizzes(c *gin.Context) {
	// Parse ID from URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"details": "ID must be a number",
		})
		return
	}

	// Get quizzes
	quizzes, err := db.GetQuizzesBySourceContentID(id)
	if err != nil {
		log.Printf("Error getting quizzes for source content %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve quizzes",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"quizzes": quizzes,
		"count":   len(quizzes),
	})
}

// GetSourceContentGeneratedContent handles GET /api/source-content/:id/content
// Returns all generated content for a source content
func GetSourceContentGeneratedContent(c *gin.Context) {
	// Parse ID from URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"details": "ID must be a number",
		})
		return
	}

	// Get concepts first (to get concept IDs)
	concepts, err := db.GetConceptsBySourceContentID(id)
	if err != nil {
		log.Printf("Error getting concepts for source content %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve generated content",
			"details": err.Error(),
		})
		return
	}

	// Get generated content by concept IDs
	var contents []models.GeneratedContent
	if len(concepts) > 0 {
		conceptIDs := make([]int, len(concepts))
		for i, c := range concepts {
			conceptIDs[i] = c.ID
		}

		contents, err = db.GetGeneratedContentByConceptIDs(conceptIDs)
		if err != nil {
			log.Printf("Error getting generated content for source %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve generated content",
				"details": err.Error(),
			})
			return
		}
	} else {
		contents = []models.GeneratedContent{}
	}

	c.JSON(http.StatusOK, gin.H{
		"generated_content": contents,
		"count":             len(contents),
	})
}

// DeleteSourceContent handles DELETE /api/source-content/:id
// Deletes a source content and all related data
func DeleteSourceContent(c *gin.Context) {
	// Parse ID from URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"details": "ID must be a number",
		})
		return
	}

	// Delete source content
	// Note: This should cascade delete related records if foreign keys are set up properly
	err = db.DeleteSourceContent(id)
	if err != nil {
		log.Printf("Error deleting source content %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete source content",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Source content deleted successfully",
		"id":      id,
	})
}
