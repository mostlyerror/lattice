package handlers

import (
	"net/http"
	"strconv"

	"github.com/mostlyerror/lattice/internal/db"
	"github.com/mostlyerror/lattice/internal/models"
	"github.com/gin-gonic/gin"
)

// GetConcepts handles GET /api/concepts
func GetConcepts(c *gin.Context) {
	concepts, err := db.GetAllConcepts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, concepts)
}

// GetConcept handles GET /api/concepts/:id
func GetConcept(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid concept id"})
		return
	}

	concept, err := db.GetConceptByID(id)
	if err != nil {
		if err.Error() == "concept not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "concept not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, concept)
}

// CreateConcept handles POST /api/concepts
func CreateConcept(c *gin.Context) {
	var req models.CreateConceptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	concept, err := db.CreateConcept(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, concept)
}

// UpdateConcept handles PATCH /api/concepts/:id
func UpdateConcept(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid concept id"})
		return
	}

	var req models.UpdateConceptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	concept, err := db.UpdateConcept(id, req)
	if err != nil {
		if err.Error() == "concept not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "concept not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, concept)
}

// DeleteConcept handles DELETE /api/concepts/:id
func DeleteConcept(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid concept id"})
		return
	}

	err = db.DeleteConcept(id)
	if err != nil {
		if err.Error() == "concept not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "concept not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "concept deleted successfully"})
}
